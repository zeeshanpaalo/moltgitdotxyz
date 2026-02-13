// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"net/http"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/services/context"
)

const (
	tplSettingsAPIKeys templates.TplName = "user/settings/apikeys"
)

// APIKeys render manage API keys page
func APIKeys(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.api_keys")
	ctx.Data["PageIsSettingsAPIKeys"] = true

	loadAPIKeysData(ctx)

	ctx.HTML(http.StatusOK, tplSettingsAPIKeys)
}

// loadAPIKeysData loads API keys data for the user
func loadAPIKeysData(ctx *context.Context) {
	keys, err := auth_model.ListUserAPIKeys(ctx, ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("ListUserAPIKeys", err)
		return
	}

	ctx.Data["APIKeys"] = keys
	ctx.Data["Environment"] = setting.Service.APIKey.Environment
}

// APIKeysPost creates a new API key
func APIKeysPost(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.api_keys")
	ctx.Data["PageIsSettingsAPIKeys"] = true

	keyName := ctx.FormString("name")
	if keyName == "" {
		ctx.Flash.Error(ctx.Tr("settings.api_key_name_required"))
		ctx.Redirect(setting.AppSubURL + "/user/settings/apikeys")
		return
	}

	// Check if name already exists
	exists, err := auth_model.UserAPIKeyByNameExists(ctx, ctx.Doer.ID, keyName)
	if err != nil {
		ctx.ServerError("UserAPIKeyByNameExists", err)
		return
	}
	if exists {
		ctx.Flash.Error(ctx.Tr("settings.api_key_name_exists"))
		ctx.Redirect(setting.AppSubURL + "/user/settings/apikeys")
		return
	}

	// Generate new API key
	isLive := setting.Service.APIKey.Environment == "production"
	_, fullKey, err := auth_model.GenerateAPIKey(ctx, ctx.Doer.ID, keyName, isLive)
	if err != nil {
		ctx.ServerError("GenerateAPIKey", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("settings.api_key_generated"))
	ctx.Flash.Info(fullKey)

	ctx.Redirect(setting.AppSubURL + "/user/settings/apikeys")
}

// DeleteAPIKey deletes an API key
func DeleteAPIKey(ctx *context.Context) {
	if err := auth_model.DeleteUserAPIKey(ctx, ctx.FormInt64("id"), ctx.Doer.ID); err != nil {
		ctx.Flash.Error("DeleteUserAPIKey: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("settings.api_key_deletion_success"))
	}

	ctx.JSONRedirect(setting.AppSubURL + "/user/settings/apikeys")
}
