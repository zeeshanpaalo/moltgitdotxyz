// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"net/http"
	"strings"

	auth_model "code.gitea.io/gitea/models/auth"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/auth/httpauth"
	"code.gitea.io/gitea/modules/log"
)

var _ Method = &APIKey{}

// APIKey implements the Auth interface and authenticates requests
// by looking for an API key in the "Authorization" header with Bearer scheme.
// API keys start with mk_live_ or mk_test_ prefix.
type APIKey struct{}

// Name represents the name of auth method
func (a *APIKey) Name() string {
	return "apikey"
}

// parseAPIKey returns the API key from request
func parseAPIKey(req *http.Request) (string, bool) {
	// Check header for Bearer token
	if auHead := req.Header.Get("Authorization"); auHead != "" {
		parsed, ok := httpauth.ParseAuthorizationHeader(auHead)
		if ok && parsed.BearerToken != nil {
			token := parsed.BearerToken.Token
			// Check if it's an API key (starts with mk_live_ or mk_test_)
			if strings.HasPrefix(token, "mk_live_") || strings.HasPrefix(token, "mk_test_") {
				return token, true
			}
		}
	}
	return "", false
}

// Verify extracts the API key from the "Authorization" header or session and returns
// the corresponding user object for that key.
// If verification is successful returns an existing user object.
// Returns nil if verification fails.
func (a *APIKey) Verify(req *http.Request, w http.ResponseWriter, store DataStore, sess SessionStore) (*user_model.User, error) {
	var apiKey string
	var fromSession bool

	// First check if API key is in session (for browser-based API key authentication)
	if sess != nil {
		if isAPIKey, _ := sess.Get("is_api_key").(bool); isAPIKey {
			if sessionKey, ok := sess.Get("api_key").(string); ok && sessionKey != "" {
				apiKey = sessionKey
				fromSession = true
			}
		}
	}

	// If not in session, check header for Bearer token
	if apiKey == "" {
		var ok bool
		apiKey, ok = parseAPIKey(req)
		if !ok {
			return nil, nil
		}
		
		// For header-based auth, check if this is an API path
		detector := newAuthPathDetector(req)
		if !detector.isAPIPath() && !detector.isAttachmentDownload() &&
			!detector.isGitRawOrAttachPath() && !detector.isArchivePath() {
			return nil, nil
		}
	}

	// Verify the API key
	userAPIKey, err := auth_model.GetUserAPIKeyByKey(req.Context(), apiKey)
	if err != nil {
		if auth_model.IsErrUserAPIKeyNotExist(err) {
			// If from session and key doesn't exist, clear session
			if fromSession && sess != nil {
				sess.Delete("api_key")
				sess.Delete("is_api_key")
			}
			return nil, nil
		}
		log.Error("GetUserAPIKeyByKey: %v", err)
		return nil, err
	}

	// Get the user
	user, err := user_model.GetUserByID(req.Context(), userAPIKey.UID)
	if err != nil {
		if user_model.IsErrUserNotExist(err) {
			return nil, nil
		}
		log.Error("GetUserByID for API key: %v", err)
		return nil, err
	}

	// Mark this as an API key authentication
	store.GetData()["IsApiKey"] = true
	store.GetData()["ApiKeyID"] = userAPIKey.ID

	log.Trace("API Key Authentication: Valid API key for user[%s] (from %s)", user.Name, map[bool]string{true: "session", false: "header"}[fromSession])

	return user, nil
}
