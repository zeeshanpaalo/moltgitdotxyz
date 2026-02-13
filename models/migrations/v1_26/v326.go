// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"code.gitea.io/gitea/modules/timeutil"

	"xorm.io/xorm"
)

func AddUserAPIKeyTable(x *xorm.Engine) error {
	type UserAPIKey struct {
		ID             int64              `xorm:"pk autoincr"`
		UID            int64              `xorm:"INDEX NOT NULL"`
		Name           string             `xorm:"NOT NULL"`
		KeyHash        string             `xorm:"UNIQUE NOT NULL"` // sha256 of key
		KeySalt        string             `xorm:"NOT NULL"`
		KeyLastEight   string             `xorm:"INDEX NOT NULL"`
		KeyPrefix      string             `xorm:"NOT NULL"` // mk_live_ or mk_test_
		IsLive         bool               `xorm:"NOT NULL DEFAULT false"`
		CreatedUnix    timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix    timeutil.TimeStamp `xorm:"INDEX updated"`
		LastUsedUnix   timeutil.TimeStamp `xorm:"INDEX"`
	}

	return x.Sync(new(UserAPIKey))
}
