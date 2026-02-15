// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"encoding/hex"
	"net/http"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/modules/wallet"
	"code.gitea.io/gitea/services/context"
)

const (
	tplSettingsWallet templates.TplName = "user/settings/wallet"
)

// Wallet renders the wallet settings page
func Wallet(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.wallet")
	ctx.Data["PageIsSettingsWallet"] = true

	// Use request context so DB uses the same context as the rest of the request
	wallets, err := auth_model.GetUserWallets(ctx.Req.Context(), ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("GetUserWallets", err)
		return
	}

	if wallets == nil {
		wallets = []*auth_model.UserWallet{}
	}
	ctx.Data["Wallets"] = wallets

	// Decrypt private keys for display (same order as Wallets; empty string if decryption fails)
	walletPrivateKeys := make([]string, len(wallets))
	for i, w := range wallets {
		decrypted, err := wallet.DecryptPrivateKey(w.EncryptedPrivateKey)
		if err == nil {
			walletPrivateKeys[i] = hex.EncodeToString(decrypted)
		}
	}
	ctx.Data["WalletPrivateKeys"] = walletPrivateKeys

	// Fetch NFTs for the user
	nfts, err := auth_model.GetUserNFTs(ctx.Req.Context(), ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("GetUserNFTs", err)
		return
	}
	ctx.Data["NFTs"] = nfts

	ctx.HTML(http.StatusOK, tplSettingsWallet)
}
