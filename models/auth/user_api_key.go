// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/modules/util"

	lru "github.com/hashicorp/golang-lru/v2"
)

// ErrUserAPIKeyNotExist represents a "UserAPIKeyNotExist" kind of error.
type ErrUserAPIKeyNotExist struct {
	Key string
}

func (err ErrUserAPIKeyNotExist) Error() string {
	return fmt.Sprintf("user API key does not exist [key: %s]", err.Key)
}

func (err ErrUserAPIKeyNotExist) Unwrap() error {
	return util.ErrNotExist
}

// IsErrUserAPIKeyNotExist checks if an error is ErrUserAPIKeyNotExist.
func IsErrUserAPIKeyNotExist(err error) bool {
	_, ok := err.(ErrUserAPIKeyNotExist)
	return ok
}

var successfulAPIKeyCache *lru.Cache[string, any]

// UserAPIKey represents a user's API key for authentication.
type UserAPIKey struct {
	ID           int64              `xorm:"pk autoincr"`
	UID          int64              `xorm:"INDEX NOT NULL"`
	Name         string             `xorm:"NOT NULL"`
	KeyHash      string             `xorm:"UNIQUE NOT NULL"` // sha256 of key
	KeySalt      string             `xorm:"NOT NULL"`
	KeyLastEight string             `xorm:"INDEX NOT NULL"`
	KeyPrefix    string             `xorm:"NOT NULL"` // mk_live_ or mk_test_
	IsLive       bool               `xorm:"NOT NULL DEFAULT false"`
	CreatedUnix  timeutil.TimeStamp `xorm:"INDEX created"`
	UpdatedUnix  timeutil.TimeStamp `xorm:"INDEX updated"`
	LastUsedUnix timeutil.TimeStamp `xorm:"INDEX"`
}

func init() {
	db.RegisterModel(new(UserAPIKey), func() error {
		if setting.SuccessfulTokensCacheSize > 0 {
			var err error
			successfulAPIKeyCache, err = lru.New[string, any](setting.SuccessfulTokensCacheSize)
			if err != nil {
				return fmt.Errorf("unable to allocate UserAPIKey cache: %w", err)
			}
		} else {
			successfulAPIKeyCache = nil
		}
		return nil
	})
}

// HashAPIKey hashes the API key with salt
func HashAPIKey(key, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateAPIKey generates a new API key with the appropriate prefix
func GenerateAPIKey(ctx context.Context, uid int64, name string, isLive bool) (*UserAPIKey, string, error) {
	// Generate random bytes for the key
	keyBytes, err := util.CryptoRandomBytes(32)
	if err != nil {
		return nil, "", err
	}

	// Generate salt
	salt, err := util.CryptoRandomString(10)
	if err != nil {
		return nil, "", err
	}

	// Encode key as hex
	keyHex := hex.EncodeToString(keyBytes)

	// Determine prefix based on environment
	prefix := "mk_test_"
	if isLive {
		prefix = "mk_live_"
	}

	// Full API key is prefix + hex key
	fullKey := prefix + keyHex

	apiKey := &UserAPIKey{
		UID:          uid,
		Name:         name,
		KeySalt:      salt,
		KeyHash:      HashAPIKey(fullKey, salt),
		KeyLastEight: fullKey[len(fullKey)-8:],
		KeyPrefix:    prefix,
		IsLive:       isLive,
	}

	if _, err := db.GetEngine(ctx).Insert(apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, fullKey, nil
}

// GetUserAPIKeyByKey returns the API key by key string
func GetUserAPIKeyByKey(ctx context.Context, key string) (*UserAPIKey, error) {
	if key == "" {
		return nil, ErrUserAPIKeyNotExist{Key: key}
	}

	// API keys should start with mk_live_ or mk_test_ (both are 8 characters)
	if len(key) < 16 || (key[:8] != "mk_live_" && key[:8] != "mk_test_") {
		return nil, ErrUserAPIKeyNotExist{Key: key}
	}

	lastEight := key[len(key)-8:]

	// Check cache first
	if id := getAPIKeyIDFromCache(key); id > 0 {
		apiKey := &UserAPIKey{KeyLastEight: lastEight}
		has, err := db.GetEngine(ctx).ID(id).Get(apiKey)
		if err != nil {
			return nil, err
		}
		if has {
			return apiKey, nil
		}
		successfulAPIKeyCache.Remove(key)
	}

	// Query from database
	var keys []UserAPIKey
	err := db.GetEngine(ctx).Where("key_last_eight = ?", lastEight).Find(&keys)
	if err != nil {
		return nil, err
	} else if len(keys) == 0 {
		return nil, ErrUserAPIKeyNotExist{Key: key}
	}

	// Check hash
	for _, k := range keys {
		tempHash := HashAPIKey(key, k.KeySalt)
		if subtle.ConstantTimeCompare([]byte(k.KeyHash), []byte(tempHash)) == 1 {
			if successfulAPIKeyCache != nil {
				successfulAPIKeyCache.Add(key, k.ID)
			}
			// Update last used time
			k.LastUsedUnix = timeutil.TimeStampNow()
			_, _ = db.GetEngine(ctx).ID(k.ID).Cols("last_used_unix").Update(&k)
			return &k, nil
		}
	}

	return nil, ErrUserAPIKeyNotExist{Key: key}
}

func getAPIKeyIDFromCache(key string) int64 {
	if successfulAPIKeyCache == nil {
		return 0
	}
	idInterface, ok := successfulAPIKeyCache.Get(key)
	if !ok {
		return 0
	}
	id, ok := idInterface.(int64)
	if !ok {
		return 0
	}
	return id
}

// ListUserAPIKeys returns all API keys for a user
func ListUserAPIKeys(ctx context.Context, uid int64) ([]*UserAPIKey, error) {
	keys := make([]*UserAPIKey, 0)
	return keys, db.GetEngine(ctx).Where("uid = ?", uid).Find(&keys)
}

// DeleteUserAPIKey deletes an API key
func DeleteUserAPIKey(ctx context.Context, id, uid int64) error {
	_, err := db.GetEngine(ctx).Where("id = ? AND uid = ?", id, uid).Delete(&UserAPIKey{})
	return err
}

// UserAPIKeyByNameExists checks if an API key name already exists for a user
func UserAPIKeyByNameExists(ctx context.Context, uid int64, name string) (bool, error) {
	return db.GetEngine(ctx).Where("uid = ? AND name = ?", uid, name).Exist(&UserAPIKey{})
}
