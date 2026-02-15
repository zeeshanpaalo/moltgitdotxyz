// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"code.gitea.io/gitea/modules/timeutil"

	"xorm.io/xorm"
)

func AddUserWalletTable(x *xorm.Engine) error {
	type UserWallet struct {
		ID                  int64              `xorm:"pk autoincr"`
		UID                 int64              `xorm:"INDEX NOT NULL"`
		Address             string             `xorm:"VARCHAR(42) NOT NULL"`
		PublicKey           string             `xorm:"VARCHAR(130) NOT NULL"`
		EncryptedPrivateKey string             `xorm:"TEXT NOT NULL"`
		Network             string             `xorm:"VARCHAR(20) NOT NULL"`
		CreatedUnix         timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix         timeutil.TimeStamp `xorm:"INDEX updated"`
	}

	return x.Sync(new(UserWallet))
}
