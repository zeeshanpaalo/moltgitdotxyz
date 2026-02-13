# API Key Authentication Feature

## Overview

This feature adds support for API key-based authentication in Gitea, similar to services like Stripe. Users can sign up and authenticate using API keys with prefixes that identify the environment (production or development).

## Features

1. **Dual Authentication System**: Users can authenticate using either:
   - Traditional session-based authentication (tokens)
   - API key-based authentication (Bearer tokens with mk_ prefix)

2. **Environment-Specific Keys**:
   - Production keys: `mk_live_...`
   - Development keys: `mk_test_...`

3. **New Registration Flow**: 
   - Route: `http://localhost:3000/user/sign_up/new`
   - Generates API key on signup
   - Shows key once (must be saved by user)

4. **API Key Login**:
   - Route: `http://localhost:3000/user/login/new`
   - Login using API key directly

5. **API Key Management**:
   - View all API keys in user settings
   - Generate new keys
   - Delete existing keys
   - See last used timestamps

## Configuration

Add the following to your `app.ini` file:

```ini
[apikey]
; Environment for API key generation: "production" or "development"
; production generates keys with mk_live_ prefix
; development generates keys with mk_test_ prefix
ENVIRONMENT = development
```

## Usage

### 1. Sign Up with API Key

Navigate to `http://localhost:3000/user/sign_up/new` and register:
- Username
- Email
- Password
- Confirm Password

Upon successful registration, an API key will be generated and displayed once. Save it securely!

### 2. Using API Keys for Authentication

API keys can be used in the `Authorization` header for API requests:

```bash
curl -H "Authorization: Bearer mk_test_abc123..." https://your-gitea-instance/api/v1/user/repos
```

### 3. Login with API Key

For web-based login using API key:
1. Navigate to `http://localhost:3000/user/login/new`
2. Enter your API key
3. Submit to authenticate

### 4. Managing API Keys

Go to User Settings → API Keys (`/user/settings/apikeys`) to:
- View all your API keys
- Generate new keys
- Delete old keys
- See usage statistics

## Database Schema

### New Table: `user_api_key`

Stores API keys for users:
- `id`: Primary key
- `uid`: User ID (foreign key)
- `name`: Key name/description
- `key_hash`: SHA256 hash of the key
- `key_salt`: Salt for hashing
- `key_last_eight`: Last 8 characters (for identification)
- `key_prefix`: Prefix (mk_live_ or mk_test_)
- `is_live`: Boolean indicating production key
- `created_unix`: Creation timestamp
- `updated_unix`: Update timestamp
- `last_used_unix`: Last usage timestamp

### Updated Table: `user`

Added field:
- `auth_type`: VARCHAR(20) - Can be 'token', 'apikey', or 'both'

## API Routes

### Public Routes
- `GET /user/sign_up/new` - API key signup page
- `POST /user/sign_up/new` - Process API key signup
- `GET /user/login/new` - API key login page
- `POST /user/login/new` - Process API key login

### Protected Routes (requires authentication)
- `GET /user/settings/apikeys` - View API keys
- `POST /user/settings/apikeys` - Generate new API key
- `POST /user/settings/apikeys/delete` - Delete API key

## Security Considerations

1. **Key Storage**: API keys are hashed with SHA256 + salt before storage
2. **One-Time Display**: Keys are shown only once upon generation
3. **Rate Limiting**: Same rate limits apply as for token-based auth
4. **HTTPS Required**: Always use HTTPS in production when transmitting API keys
5. **Key Rotation**: Users can generate multiple keys and rotate them regularly

## Authentication Flow

The authentication middleware checks in this order:
1. **API Key** (Bearer token with mk_ prefix)
2. **OAuth2**
3. **Basic Auth**
4. **Reverse Proxy** (if enabled)
5. **Session**
6. **SSPI** (if enabled on Windows)

## Migration

Run Gitea after adding these files, and the database migrations will execute automatically:
- Migration 326: Create `user_api_key` table
- Migration 327: Add `auth_type` column to `user` table

## Testing

### Test API Key Generation
```bash
# Sign up via new route
curl -X POST http://localhost:3000/user/sign_up/new \
  -d "user_name=testuser" \
  -d "email=test@example.com" \
  -d "password=Test123!" \
  -d "retype=Test123!"
```

### Test API Key Authentication
```bash
# Use API key in Authorization header
curl -H "Authorization: Bearer mk_test_your_key_here" \
  http://localhost:3000/api/v1/user
```

## Files Added/Modified

### New Files
- `models/auth/user_api_key.go` - API key model
- `models/migrations/v1_26/v326.go` - Migration for api key table
- `models/migrations/v1_26/v327.go` - Migration for auth type
- `routers/web/auth/auth_apikey.go` - API key auth handlers
- `routers/web/user/setting/apikeys.go` - Settings handlers
- `services/auth/apikey.go` - API key authentication method
- `templates/user/auth/signup_apikey.tmpl` - Signup page
- `templates/user/auth/signup_apikey_inner.tmpl` - Signup form
- `templates/user/auth/signin_apikey.tmpl` - Login page
- `templates/user/auth/signin_apikey_inner.tmpl` - Login form
- `templates/user/settings/apikeys.tmpl` - Settings page

### Modified Files
- `models/migrations/migrations.go` - Added new migrations
- `models/user/user.go` - Added AuthType field
- `modules/setting/service.go` - Added API key configuration
- `routers/web/web.go` - Added routes for API key auth
- `templates/user/settings/navbar.tmpl` - Added API keys link

## Notes

- Users registered via standard signup (`/user/sign_up`) will have `auth_type = 'token'`
- Users registered via API key signup (`/user/sign_up/new`) will have `auth_type = 'both'`
- Users can manage multiple API keys simultaneously
- Each API key can have a descriptive name for easy identification
- API keys persist across sessions and don't expire unless deleted
