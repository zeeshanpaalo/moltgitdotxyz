// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"xorm.io/xorm"
)

func AddAuthTypeToUser(x *xorm.Engine) error {
	type User struct {
		ID       int64  `xorm:"pk autoincr"`
		AuthType string `xorm:"VARCHAR(20) NOT NULL DEFAULT 'token'"` // 'token', 'apikey', or 'both'
	}

	return x.Sync(new(User))
}
