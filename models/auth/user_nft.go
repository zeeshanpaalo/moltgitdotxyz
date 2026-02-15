// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"fmt"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/modules/util"
)

// NFT mint status
const (
	NFTStatusPending = "pending"
	NFTStatusMinted  = "minted"
	NFTStatusFailed  = "failed"
)

// ErrUserNFTNotExist represents a "UserNFTNotExist" kind of error.
type ErrUserNFTNotExist struct {
	UID int64
}

func (err ErrUserNFTNotExist) Error() string {
	return fmt.Sprintf("user NFT does not exist [uid: %d]", err.UID)
}

func (err ErrUserNFTNotExist) Unwrap() error {
	return util.ErrNotExist
}

// IsErrUserNFTNotExist checks if an error is ErrUserNFTNotExist.
func IsErrUserNFTNotExist(err error) bool {
	_, ok := err.(ErrUserNFTNotExist)
	return ok
}

// UserNFT represents a minted NFT record for a user.
type UserNFT struct {
	ID              int64              `xorm:"pk autoincr"`
	UID             int64              `xorm:"INDEX NOT NULL"`
	WalletAddress   string             `xorm:"VARCHAR(42) NOT NULL"` // User's wallet address the NFT was minted to
	ContractAddress string             `xorm:"VARCHAR(42) NOT NULL"` // NFT contract address
	TokenID         int64              `xorm:"NOT NULL DEFAULT -1"`  // Token ID from the contract (-1 = not yet known)
	TxHash          string             `xorm:"VARCHAR(66)"`          // Transaction hash (0x-prefixed)
	ChainID         int64              `xorm:"NOT NULL"`             // Blockchain chain ID
	Status          string             `xorm:"VARCHAR(20) NOT NULL"` // "pending", "minted", "failed"
	ErrorMsg        string             `xorm:"TEXT"`                 // Error message if failed
	CreatedUnix     timeutil.TimeStamp `xorm:"INDEX created"`
	UpdatedUnix     timeutil.TimeStamp `xorm:"INDEX updated"`
}

func init() {
	db.RegisterModel(new(UserNFT))
}

// CreateUserNFT creates a new NFT record in the database.
func CreateUserNFT(ctx context.Context, nft *UserNFT) error {
	_, err := db.GetEngine(ctx).Insert(nft)
	return err
}

// GetUserNFT returns the NFT record for a given user ID and contract address.
func GetUserNFT(ctx context.Context, uid int64, contractAddress string) (*UserNFT, error) {
	nft := &UserNFT{}
	has, err := db.GetEngine(ctx).Where("uid = ? AND contract_address = ?", uid, contractAddress).Get(nft)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrUserNFTNotExist{UID: uid}
	}
	return nft, nil
}

// GetUserNFTs returns all NFT records for a given user ID.
func GetUserNFTs(ctx context.Context, uid int64) ([]*UserNFT, error) {
	nfts := make([]*UserNFT, 0)
	err := db.GetEngine(ctx).Where("uid = ?", uid).Find(&nfts)
	return nfts, err
}

// UpdateUserNFT updates an existing NFT record.
func UpdateUserNFT(ctx context.Context, nft *UserNFT) error {
	_, err := db.GetEngine(ctx).ID(nft.ID).Cols("token_id", "tx_hash", "status", "error_msg", "updated_unix").Update(nft)
	return err
}

// GetPendingNFTs returns all NFT records with "pending" status (for retry logic).
func GetPendingNFTs(ctx context.Context) ([]*UserNFT, error) {
	nfts := make([]*UserNFT, 0)
	err := db.GetEngine(ctx).Where("status = ?", NFTStatusPending).Find(&nfts)
	return nfts, err
}

// HasUserNFTForContract checks if a user already has an NFT record for a given contract.
func HasUserNFTForContract(ctx context.Context, uid int64, contractAddress string) (bool, error) {
	return db.GetEngine(ctx).Where("uid = ? AND contract_address = ? AND status = ?", uid, contractAddress, NFTStatusMinted).Exist(&UserNFT{})
}
