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

// ErrUserWalletNotExist represents a "UserWalletNotExist" kind of error.
type ErrUserWalletNotExist struct {
	UID int64
}

func (err ErrUserWalletNotExist) Error() string {
	return fmt.Sprintf("user wallet does not exist [uid: %d]", err.UID)
}

func (err ErrUserWalletNotExist) Unwrap() error {
	return util.ErrNotExist
}

// IsErrUserWalletNotExist checks if an error is ErrUserWalletNotExist.
func IsErrUserWalletNotExist(err error) bool {
	_, ok := err.(ErrUserWalletNotExist)
	return ok
}

// UserWallet represents a user's Web3 wallet stored in the database.
type UserWallet struct {
	ID                  int64              `xorm:"pk autoincr"`
	UID                 int64              `xorm:"INDEX NOT NULL"`
	Address             string             `xorm:"VARCHAR(42) NOT NULL"`  // 0x-prefixed EVM address
	PublicKey           string             `xorm:"VARCHAR(130) NOT NULL"` // Hex-encoded uncompressed public key
	EncryptedPrivateKey string             `xorm:"TEXT NOT NULL"`         // AES-256-GCM encrypted private key
	Network             string             `xorm:"VARCHAR(20) NOT NULL"`  // "mainnet" or "testnet"
	CreatedUnix         timeutil.TimeStamp `xorm:"INDEX created"`
	UpdatedUnix         timeutil.TimeStamp `xorm:"INDEX updated"`
}

func init() {
	db.RegisterModel(new(UserWallet))
}

// CreateUserWallet creates a new wallet record in the database.
func CreateUserWallet(ctx context.Context, wallet *UserWallet) error {
	_, err := db.GetEngine(ctx).Insert(wallet)
	return err
}

// GetUserWallet returns the wallet for a given user ID and network.
func GetUserWallet(ctx context.Context, uid int64, network string) (*UserWallet, error) {
	wallet := &UserWallet{UID: uid, Network: network}
	has, err := db.GetEngine(ctx).Where("uid = ? AND network = ?", uid, network).Get(wallet)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrUserWalletNotExist{UID: uid}
	}
	return wallet, nil
}

// GetUserWallets returns all wallets for a given user ID.
func GetUserWallets(ctx context.Context, uid int64) ([]*UserWallet, error) {
	wallets := make([]*UserWallet, 0, 2)
	return wallets, db.GetEngine(ctx).Where("uid = ?", uid).Find(&wallets)
}

// DeleteUserWallets deletes all wallets for a given user ID.
func DeleteUserWallets(ctx context.Context, uid int64) error {
	_, err := db.GetEngine(ctx).Where("uid = ?", uid).Delete(&UserWallet{})
	return err
}
