// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"html/template"
	"net/http"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/models/db"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/auth/password"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/session"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/modules/wallet"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/modules/web/middleware"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/forms"
	"code.gitea.io/gitea/services/mailer"
	user_service "code.gitea.io/gitea/services/user"
)

const (
	tplSignUpAPIKey        templates.TplName = "user/auth/signup_apikey"
	tplSignUpAPIKeySuccess templates.TplName = "user/auth/signup_apikey_success"
	tplSignInAPIKey        templates.TplName = "user/auth/signin_apikey"
)

// SignUpAPIKey renders the API key based signup page
func SignUpAPIKey(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_up")
	ctx.Data["SignUpLink"] = setting.AppSubURL + "/user/sign_up/new"

	hasUsers, _ := user_model.HasUsers(ctx)
	ctx.Data["IsFirstTimeRegistration"] = !hasUsers.HasAnyUser

	ctx.Data["PageIsSignUp"] = true

	// Show Disabled Registration message if DisableRegistration or AllowOnlyExternalRegistration options are true
	ctx.Data["DisableRegistration"] = setting.Service.DisableRegistration || setting.Service.AllowOnlyExternalRegistration

	rememberAuthRedirectLink(ctx)

	ctx.HTML(http.StatusOK, tplSignUpAPIKey)
}

// SignUpAPIKeyPost handles API key based registration
func SignUpAPIKeyPost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.RegisterForm)
	ctx.Data["Title"] = ctx.Tr("sign_up")
	ctx.Data["SignUpLink"] = setting.AppSubURL + "/user/sign_up/new"
	ctx.Data["PageIsSignUp"] = true

	wantJSON := ctx.FormBool("jsondata")

	// Set default password if none provided
	// if form.Password == "" {
	// 	form.Password = "hunza123"
	// 	form.Retype = "hunza123"
	// }

	// Permission denied if DisableRegistration or AllowOnlyExternalRegistration options are true
	if setting.Service.DisableRegistration || setting.Service.AllowOnlyExternalRegistration {
		if wantJSON {
			ctx.JSON(http.StatusForbidden, map[string]any{"error": "registration is disabled"})
			return
		}
		ctx.HTTPError(http.StatusForbidden)
		return
	}

	if ctx.HasError() {
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": ctx.GetErrMsg()})
			return
		}
		ctx.HTML(http.StatusOK, tplSignUpAPIKey)
		return
	}

	if !wantJSON {
		context.VerifyCaptcha(ctx, tplSignUpAPIKey, form)
		if ctx.Written() {
			return
		}
	}

	if !form.IsEmailDomainAllowed() {
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": ctx.Tr("auth.email_domain_blacklisted")})
			return
		}
		ctx.RenderWithErr(ctx.Tr("auth.email_domain_blacklisted"), tplSignUpAPIKey, &form)
		return
	}

	if form.Password != form.Retype {
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": ctx.Tr("form.password_not_match")})
			return
		}
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(ctx.Tr("form.password_not_match"), tplSignUpAPIKey, &form)
		return
	}
	if len(form.Password) < setting.MinPasswordLength {
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": ctx.Tr("auth.password_too_short", setting.MinPasswordLength)})
			return
		}
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(ctx.Tr("auth.password_too_short", setting.MinPasswordLength), tplSignUpAPIKey, &form)
		return
	}
	if !password.IsComplexEnough(form.Password) {
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": password.BuildComplexityError(ctx.Locale)})
			return
		}
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(password.BuildComplexityError(ctx.Locale), tplSignUpAPIKey, &form)
		return
	}
	if err := password.IsPwned(ctx, form.Password); err != nil {
		errMsg := ctx.Tr("auth.password_pwned", "https://haveibeenpwned.com/Passwords")
		if password.IsErrIsPwnedRequest(err) {
			log.Error(err.Error())
			errMsg = ctx.Tr("auth.password_pwned_err")
		}
		if wantJSON {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": errMsg})
			return
		}
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		return
	}

	u := &user_model.User{
		Name:     form.UserName,
		Email:    form.Email,
		Passwd:   form.Password,
		AuthType: "both", // Users can use both API key and token
	}

	meta := &user_model.Meta{
		InitialIP:        ctx.RemoteAddr(),
		InitialUserAgent: ctx.Req.UserAgent(),
	}

	if err := user_model.CreateUser(ctx, u, meta, nil); err != nil {
		var errMsg template.HTML
		switch {
		case user_model.IsErrUserAlreadyExist(err):
			errMsg = ctx.Tr("form.username_been_taken")
			if wantJSON {
				ctx.JSON(http.StatusConflict, map[string]any{"error": errMsg, "field": "username"})
				return
			}
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		case user_model.IsErrEmailAlreadyUsed(err):
			errMsg = ctx.Tr("form.email_been_used")
			if wantJSON {
				ctx.JSON(http.StatusConflict, map[string]any{"error": errMsg, "field": "email"})
				return
			}
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		case user_model.IsErrEmailCharIsNotSupported(err), user_model.IsErrEmailInvalid(err):
			errMsg = ctx.Tr("form.email_invalid")
			if wantJSON {
				ctx.JSON(http.StatusBadRequest, map[string]any{"error": errMsg, "field": "email"})
				return
			}
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		case db.IsErrNameReserved(err):
			errMsg = ctx.Tr("user.form.name_reserved", err.(db.ErrNameReserved).Name)
			if wantJSON {
				ctx.JSON(http.StatusBadRequest, map[string]any{"error": errMsg, "field": "username"})
				return
			}
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		case db.IsErrNameCharsNotAllowed(err):
			errMsg = ctx.Tr("user.form.name_chars_not_allowed", err.(db.ErrNameCharsNotAllowed).Name)
			if wantJSON {
				ctx.JSON(http.StatusBadRequest, map[string]any{"error": errMsg, "field": "username"})
				return
			}
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		case db.IsErrNamePatternNotAllowed(err):
			errMsg = ctx.Tr("user.form.name_pattern_not_allowed", err.(db.ErrNamePatternNotAllowed).Pattern)
			if wantJSON {
				ctx.JSON(http.StatusBadRequest, map[string]any{"error": errMsg, "field": "username"})
				return
			}
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(errMsg, tplSignUpAPIKey, &form)
		default:
			if wantJSON {
				ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "internal server error"})
				return
			}
			ctx.ServerError("CreateUser", err)
		}
		return
	}

	// Generate API key for the user
	isLive := setting.Service.APIKey.Environment == "production"
	apiKey, fullKey, err := auth_model.GenerateAPIKey(ctx, u.ID, "Default API Key", isLive)
	if err != nil {
		log.Error("Failed to generate API key for user %s: %v", u.Name, err)
		if wantJSON {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate API key"})
			return
		}
		ctx.ServerError("GenerateAPIKey", err)
		return
	}

	log.Info("Generated API key for user %s (ID: %d), Key ID: %d", u.Name, u.ID, apiKey.ID)

	// Generate Web3 wallet for the user
	walletInfo, err := wallet.GenerateWallet()
	if err != nil {
		log.Error("Failed to generate wallet for user %s: %v", u.Name, err)
		if wantJSON {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate wallet"})
			return
		}
		ctx.ServerError("GenerateWallet", err)
		return
	}

	// Determine wallet network based on API key environment
	walletNetwork := "testnet"
	if isLive {
		walletNetwork = "mainnet"
	}

	// Store wallet in database
	userWallet := &auth_model.UserWallet{
		UID:                 u.ID,
		Address:             walletInfo.Address,
		PublicKey:           walletInfo.PublicKeyHex,
		EncryptedPrivateKey: walletInfo.EncryptedPrivateKey,
		Network:             walletNetwork,
	}
	if err := auth_model.CreateUserWallet(ctx, userWallet); err != nil {
		log.Error("Failed to store wallet for user %s: %v", u.Name, err)
		if wantJSON {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to store wallet"})
			return
		}
		ctx.ServerError("CreateUserWallet", err)
		return
	}

	log.Info("Generated wallet for user %s (ID: %d), Address: %s, Network: %s", u.Name, u.ID, walletInfo.Address, walletNetwork)

	// Send activation email if required
	if !u.IsActive && setting.Service.RegisterEmailConfirm {
		mailer.SendActivateAccountMail(ctx.Locale, u)

		if wantJSON {
			ctx.JSON(http.StatusOK, map[string]any{
				"status":              "activation_required",
				"username":            u.Name,
				"email":               u.Email,
				"api_key":             fullKey,
				"wallet_address":      walletInfo.Address,
				"wallet_public_key":   walletInfo.PublicKeyHex,
				"wallet_network":      walletNetwork,
				"activation_required": true,
				"message":             "Please check your email to activate your account.",
			})
			return
		}

		ctx.Data["IsSendRegisterMail"] = true
		ctx.Data["Email"] = u.Email
		ctx.Data["ActiveCodeLives"] = timeutil.MinutesToFriendly(setting.Service.ActiveCodeLives, ctx.Locale)
		ctx.HTML(http.StatusOK, TplActivate)

		if err := ctx.Cache.Put("MailResendLimit_"+u.LowerName, u.LowerName, 180); err != nil {
			log.Error("Set cache(MailResendLimit) fail: %v", err)
		}
		return
	}

	// Return JSON response if requested
	if wantJSON {
		ctx.JSON(http.StatusOK, map[string]any{
			"status":            "success",
			"username":          u.Name,
			"email":             u.Email,
			"api_key":           fullKey,
			"wallet_address":    walletInfo.Address,
			"wallet_public_key": walletInfo.PublicKeyHex,
			"wallet_network":    walletNetwork,
		})
		return
	}

	// Display the API key and wallet info to the user (IMPORTANT: Show before sign-in)
	ctx.Data["Title"] = ctx.Tr("auth.sign_up_successful")
	ctx.Data["APIKey"] = fullKey
	ctx.Data["UserName"] = u.Name
	ctx.Data["WalletAddress"] = walletInfo.Address
	ctx.Data["WalletPublicKey"] = walletInfo.PublicKeyHex
	ctx.Data["WalletNetwork"] = walletNetwork
	ctx.Data["PageIsSignUp"] = true

	// DO NOT auto sign in - user needs to save their API key first
	ctx.HTML(http.StatusOK, tplSignUpAPIKeySuccess)
}

// SignInAPIKey renders the API key based login page
func SignInAPIKey(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_in")
	ctx.Data["SignInLink"] = setting.AppSubURL + "/user/login/new"
	ctx.Data["PageIsSignIn"] = true
	ctx.Data["PageIsLogin"] = true

	rememberAuthRedirectLink(ctx)

	ctx.HTML(http.StatusOK, tplSignInAPIKey)
}

// SignInAPIKeyPost handles API key based login
func SignInAPIKeyPost(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	apiKey := ctx.FormString("api_key")

	if apiKey == "" {
		ctx.RenderWithErr(ctx.Tr("form.api_key_empty"), tplSignInAPIKey, nil)
		return
	}

	// Verify the API key
	userAPIKey, err := auth_model.GetUserAPIKeyByKey(ctx, apiKey)
	if err != nil {
		if auth_model.IsErrUserAPIKeyNotExist(err) {
			ctx.RenderWithErr(ctx.Tr("form.api_key_invalid"), tplSignInAPIKey, nil)
		} else {
			ctx.ServerError("GetUserAPIKeyByKey", err)
		}
		return
	}

	// Get the user
	u, err := user_model.GetUserByID(ctx, userAPIKey.UID)
	if err != nil {
		if user_model.IsErrUserNotExist(err) {
			ctx.RenderWithErr(ctx.Tr("form.api_key_invalid"), tplSignInAPIKey, nil)
		} else {
			ctx.ServerError("GetUserByID", err)
		}
		return
	}

	// Check if user is active
	if !u.IsActive {
		if setting.Service.RegisterEmailConfirm {
			ctx.Data["ResendLimitExceeded"] = true
			ctx.HTML(http.StatusOK, TplActivate)
			return
		}
		ctx.RenderWithErr(ctx.Tr("auth.prohibit_login"), tplSignInAPIKey, nil)
		return
	}

	// Check if login is prohibited
	if u.ProhibitLogin {
		ctx.RenderWithErr(ctx.Tr("auth.prohibit_login"), tplSignInAPIKey, nil)
		return
	}

	// Successful login with API key - store in session
	handleAPIKeySignIn(ctx, u, apiKey)
}

// handleAPIKeySignIn handles sign-in using API key (stores API key in session)
func handleAPIKeySignIn(ctx *context.Context, u *user_model.User, apiKey string) {
	// Update session with user info and API key
	if err := updateSession(ctx, []string{
		// Delete auth-related data
		"openid_verified_uri",
		"openid_signin_remember",
		"openid_determined_email",
		"openid_determined_username",
		"twofaUid",
		"twofaRemember",
		"linkAccount",
		"linkAccountData",
	}, map[string]any{
		session.KeyUID:   u.ID,
		session.KeyUname: u.Name,
		"api_key":        apiKey, // Store API key in session
		"is_api_key":     true,   // Flag to indicate API key authentication
	}); err != nil {
		ctx.ServerError("RegenerateSession", err)
		return
	}

	// Set locale
	middleware.SetLocaleCookie(ctx.Resp, u.Language, 0)

	if u.Language != "" && ctx.Locale.Language() != u.Language {
		ctx.Locale = middleware.Locale(ctx.Resp, ctx.Req)
	}

	// Register last login
	if err := user_service.UpdateUser(ctx, u, &user_service.UpdateOptions{SetLastLogin: true}); err != nil {
		log.Error("UpdateUser: %v", err)
	}

	// Redirect to redirect_to or home page
	redirectAfterAuth(ctx)
}
