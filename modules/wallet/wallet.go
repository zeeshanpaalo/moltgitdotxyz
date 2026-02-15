// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"code.gitea.io/gitea/modules/setting"

	"golang.org/x/crypto/sha3"
)

// secp256k1 curve parameters (used by Ethereum/Monad/EVM chains)
var secp256k1Curve = newSecp256k1()

type secp256k1CurveParams struct {
	*elliptic.CurveParams
}

func newSecp256k1() elliptic.Curve {
	params := &elliptic.CurveParams{
		Name:    "secp256k1",
		BitSize: 256,
	}
	params.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	params.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	params.B, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000007", 16)
	params.Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	params.Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	return &secp256k1CurveParams{params}
}

// Params returns the curve parameters.
func (curve *secp256k1CurveParams) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

// WalletInfo contains the generated wallet information.
type WalletInfo struct {
	PrivateKeyHex       string // Hex-encoded private key (32 bytes)
	PublicKeyHex        string // Hex-encoded uncompressed public key (65 bytes with 04 prefix)
	Address             string // EVM-compatible address (0x-prefixed, 20 bytes)
	EncryptedPrivateKey string // AES-256-GCM encrypted private key (hex-encoded)
}

// GenerateWallet creates a new secp256k1 keypair and derives an EVM-compatible address.
// The private key is also encrypted using AES-256-GCM with a key derived from the server's SecretKey.
func GenerateWallet() (*WalletInfo, error) {
	// Generate secp256k1 private key
	privateKey, err := ecdsa.GenerateKey(secp256k1Curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
	}

	// Extract raw private key bytes (32 bytes)
	privKeyBytes := privateKey.D.Bytes()
	// Ensure it's 32 bytes (left-pad with zeros if necessary)
	if len(privKeyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(privKeyBytes):], privKeyBytes)
		privKeyBytes = padded
	}
	privKeyHex := hex.EncodeToString(privKeyBytes)

	// Get uncompressed public key (04 || X || Y)
	pubKeyBytes := elliptic.Marshal(secp256k1Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	// Derive EVM address: Keccak-256 of public key (without 04 prefix), take last 20 bytes
	address := pubKeyToAddress(pubKeyBytes)

	// Encrypt private key
	encryptedPrivKey, err := encryptPrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	return &WalletInfo{
		PrivateKeyHex:       privKeyHex,
		PublicKeyHex:        pubKeyHex,
		Address:             address,
		EncryptedPrivateKey: encryptedPrivKey,
	}, nil
}

// pubKeyToAddress derives an EVM-compatible address from an uncompressed public key.
func pubKeyToAddress(pubKeyBytes []byte) string {
	// Remove the 04 prefix (uncompressed point marker)
	pubKeyNoPrefix := pubKeyBytes[1:]

	// Keccak-256 hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pubKeyNoPrefix)
	hash := hasher.Sum(nil)

	// Take last 20 bytes
	addressBytes := hash[len(hash)-20:]
	return "0x" + hex.EncodeToString(addressBytes)
}

// deriveEncryptionKey derives a 32-byte AES key from the server's SecretKey using SHA-256.
func deriveEncryptionKey() []byte {
	h := sha256.Sum256([]byte("wallet-encryption:" + setting.SecretKey))
	return h[:]
}

// encryptPrivateKey encrypts the private key using AES-256-GCM.
// Format: hex(nonce || ciphertext)
func encryptPrivateKey(privKeyBytes []byte) (string, error) {
	key := deriveEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, privKeyBytes, nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey decrypts an encrypted private key.
func DecryptPrivateKey(encryptedHex string) ([]byte, error) {
	data, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	key := deriveEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return plaintext, nil
}
