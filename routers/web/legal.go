// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/services/context"
)

const (
	tplPrivacyPolicy  templates.TplName = "legal/privacy"
	tplTermsOfService templates.TplName = "legal/terms"
)

// PrivacyPolicy renders the privacy policy page
func PrivacyPolicy(ctx *context.Context) {
	ctx.Data["Title"] = "Privacy Policy"
	ctx.HTML(http.StatusOK, tplPrivacyPolicy)
}

// TermsOfService renders the terms of service page
func TermsOfService(ctx *context.Context) {
	ctx.Data["Title"] = "Terms of Service"
	ctx.HTML(http.StatusOK, tplTermsOfService)
}
