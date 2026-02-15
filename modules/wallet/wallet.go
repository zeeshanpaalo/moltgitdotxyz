// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"code.gitea.io/gitea/modules/setting"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/sha3"
)

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
	// Generate secp256k1 private key using the decred library
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
	}

	// Extract raw private key bytes (32 bytes)
	privKeyBytes := privateKey.Serialize()
	privKeyHex := hex.EncodeToString(privKeyBytes)

	// Get uncompressed public key (04 || X || Y) - 65 bytes
	pubKey := privateKey.PubKey()
	pubKeyBytes := pubKey.SerializeUncompressed()
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
