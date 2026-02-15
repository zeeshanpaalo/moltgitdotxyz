// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"code.gitea.io/gitea/modules/timeutil"

	"xorm.io/xorm"
)

func AddUserNFTTable(x *xorm.Engine) error {
	type UserNFT struct {
		ID              int64              `xorm:"pk autoincr"`
		UID             int64              `xorm:"INDEX NOT NULL"`
		WalletAddress   string             `xorm:"VARCHAR(42) NOT NULL"`
		ContractAddress string             `xorm:"VARCHAR(42) NOT NULL"`
		TokenID         int64              `xorm:"NOT NULL DEFAULT -1"`
		TxHash          string             `xorm:"VARCHAR(66)"`
		ChainID         int64              `xorm:"NOT NULL"`
		Status          string             `xorm:"VARCHAR(20) NOT NULL"`
		ErrorMsg        string             `xorm:"TEXT"`
		CreatedUnix     timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix     timeutil.TimeStamp `xorm:"INDEX updated"`
	}
	return x.Sync(new(UserNFT))
}
