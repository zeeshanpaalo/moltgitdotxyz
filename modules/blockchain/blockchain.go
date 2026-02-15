// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TestAgent ABI - only the mint function we need
const testAgentMintABI = `[{"inputs":[{"internalType":"address","name":"agent","type":"address"}],"name":"mint","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"totalMinted","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`

// MintResult contains the result of an NFT mint operation.
type MintResult struct {
	TxHash  string
	TokenID int64
	Success bool
	Error   error
}

// MintNFT mints an NFT to the given wallet address using the server's configured wallet.
// It sends a transaction to the TestAgent contract's mint(address) function.
func MintNFT(ctx context.Context, toAddress string) (*MintResult, error) {
	if !setting.Service.Blockchain.Enabled {
		return nil, fmt.Errorf("blockchain features are not enabled")
	}

	rpcURL := setting.Service.Blockchain.RPCURL
	privateKeyHex := setting.Service.Blockchain.PrivateKey
	contractAddr := setting.Service.Blockchain.ContractAddress
	chainID := setting.Service.Blockchain.ChainID

	// Connect to the RPC endpoint
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// Parse the server's private key (strip 0x prefix if present)
	privKeyHex := strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return &MintResult{Success: false, Error: fmt.Errorf("cannot cast public key")}, fmt.Errorf("cannot cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Info("Blockchain: Minting NFT from %s to %s on chain %d", fromAddress.Hex(), toAddress, chainID)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Parse contract ABI
	parsedABI, err := abi.JSON(strings.NewReader(testAgentMintABI))
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Target address for the NFT mint
	to := common.HexToAddress(toAddress)

	// Create the transaction
	contract := common.HexToAddress(contractAddr)
	chainIDBig := big.NewInt(chainID)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainIDBig)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to create transactor: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(200000) // Gas limit for mint
	auth.GasPrice = gasPrice

	// Use the bound contract approach for cleaner transaction handling
	boundContract := bind.NewBoundContract(contract, parsedABI, client, client, client)

	tx, err := boundContract.Transact(auth, "mint", to)
	if err != nil {
		return &MintResult{Success: false, Error: err}, fmt.Errorf("failed to send mint transaction: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Info("Blockchain: Mint transaction sent: %s", txHash)

	// Wait for the transaction receipt with a timeout
	receiptCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	receipt, err := bind.WaitMined(receiptCtx, client, tx)
	if err != nil {
		log.Warn("Blockchain: Failed to wait for mint receipt (tx: %s): %v", txHash, err)
		// Transaction was sent but we couldn't get receipt - return pending
		return &MintResult{
			TxHash:  txHash,
			TokenID: -1,
			Success: false,
			Error:   fmt.Errorf("transaction sent but receipt not confirmed: %w", err),
		}, nil
	}

	if receipt.Status == 0 {
		log.Error("Blockchain: Mint transaction failed (tx: %s)", txHash)
		return &MintResult{
			TxHash:  txHash,
			TokenID: -1,
			Success: false,
			Error:   fmt.Errorf("transaction reverted"),
		}, nil
	}

	// Try to extract tokenID from the Transfer event logs
	tokenID := int64(-1)
	transferEventSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	for _, vLog := range receipt.Logs {
		if len(vLog.Topics) >= 4 && vLog.Topics[0] == transferEventSig {
			tokenID = new(big.Int).SetBytes(vLog.Topics[3].Bytes()).Int64()
			break
		}
	}

	log.Info("Blockchain: NFT minted successfully! TX: %s, TokenID: %d", txHash, tokenID)

	return &MintResult{
		TxHash:  txHash,
		TokenID: tokenID,
		Success: true,
	}, nil
}

// IsEnabled returns whether blockchain features are enabled.
func IsEnabled() bool {
	return setting.Service.Blockchain.Enabled
}
