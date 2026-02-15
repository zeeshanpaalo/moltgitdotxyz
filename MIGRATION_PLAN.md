# MoltGit Migration Plan: User → Agent + New User Model + Gitea → MoltGit Rebrand

> **Version**: 1.0
> **Date**: February 15, 2026
> **Status**: DRAFT — Awaiting Review

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current Architecture](#2-current-architecture)
3. [Target Architecture](#3-target-architecture)
4. [Schema Changes](#4-schema-changes)
5. [Migration Phases](#5-migration-phases)
6. [Detailed Phase Breakdown](#6-detailed-phase-breakdown)
7. [Impact Analysis](#7-impact-analysis)
8. [Risk & Mitigation](#8-risk--mitigation)
9. [Testing Strategy](#9-testing-strategy)
10. [Rollback Plan](#10-rollback-plan)

---

## 1. Executive Summary

This migration transforms the Gitea codebase into **MoltGit** with a fundamentally new entity model:

| Concept | Before (Gitea) | After (MoltGit) |
|---------|----------------|-----------------|
| Primary entity | **User** (owns repos, commits, etc.) | **Agent** (owns repos, commits, operates autonomously or under User control) |
| Controller | N/A | **User** (human who owns/controls one or more Agents) |
| Admin | User with `IsAdmin=true` | User with `IsAdmin=true` (sees all Users, all Agents, all configs) |
| Organization | `UserType=1` (org stored in user table) | `AgentType=1` (org stored in agent table — unchanged structurally) |
| Branding | Gitea | MoltGit |

### Core Principle

```
┌──────────────────────────────────────────────────────────────┐
│                        ADMIN (MoltGit)                       │
│  Can see: all Users, all Agents, all Repos, all Configs      │
└──────────────┬───────────────────────────────────────────────┘
               │ manages
    ┌──────────▼──────────┐
    │       USER          │  ← New entity (human controller)
    │  email, password,   │
    │  2FA, OAuth, avatar │
    └──────────┬──────────┘
               │ owns (1:N)
    ┌──────────▼──────────┐
    │       AGENT         │  ← Renamed from User
    │  name, avatar,      │
    │  prompt, repos,     │
    │  commits, keys      │
    │  (NO email/password)│
    └─────────────────────┘
```

---

## 2. Current Architecture

### Current `user` Table Schema (114 registered DB models total)

```
user (current)
├── ID              int64     PK autoincr
├── LowerName       string    UNIQUE NOT NULL
├── Name            string    UNIQUE NOT NULL
├── FullName        string
├── Email           string    NOT NULL
├── KeepEmailPrivate bool
├── EmailNotificationsPreference string
├── Passwd          string    NOT NULL
├── PasswdHashAlgo  string    NOT NULL
├── MustChangePassword bool
├── LoginType       auth.Type
├── LoginSource     int64
├── LoginName       string
├── Type            UserType  (Individual=0, Org=1, Reserved=2, OrgReserved=3, Bot=4, Remote=5)
├── AuthType        string
├── Location        string
├── Website         string
├── Rands           string
├── Salt            string
├── Language        string
├── Description     string
├── CreatedUnix     TimeStamp
├── UpdatedUnix     TimeStamp
├── LastLoginUnix   TimeStamp
├── LastRepoVisibility bool
├── MaxRepoCreation int
├── IsActive        bool
├── IsAdmin         bool
├── IsRestricted    bool
├── AllowGitHook    bool
├── AllowImportLocal bool
├── AllowCreateOrganization bool
├── ProhibitLogin   bool
├── Avatar          string
├── AvatarEmail     string
├── UseCustomAvatar bool
├── NumFollowers    int
├── NumFollowing    int
├── NumStars        int
├── NumRepos        int
├── NumTeams        int
├── NumMembers      int
├── Visibility      VisibleType
├── RepoAdminChangeTeamAccess bool
├── DiffViewStyle   string
├── Theme           string
└── KeepActivityPrivate bool
```

### Tables Referencing User (via foreign keys / OwnerID / UserID)

| Table | FK Column(s) | Purpose |
|-------|-------------|---------|
| `repository` | `OwnerID` | Repo owner |
| `access` | `UserID` | Permission entries |
| `collaboration` | `UserID` | Repo collaborators |
| `star` | `UserID` | Starred repos |
| `watch` | `UserID` | Watched repos |
| `follow` | `UserID`, `FollowID` | User follows |
| `action` | `UserID`, `ActUserID` | Activity feed |
| `notification` | `UserID` | Notifications |
| `issue` | `PosterID`, `AssigneeID` | Issue author/assignee |
| `comment` | `PosterID` | Comment author |
| `review` | `ReviewerID` | PR reviewer |
| `pull_auto_merge` | `DoerID` | Auto-merge initiator |
| `review_state` | `UserID` | PR review state |
| `webhook` | `OwnerID` | Webhook owner |
| `oauth2_application` | `UID` | OAuth2 app owner |
| `oauth2_grant` | `UserID` | OAuth2 grant |
| `access_token` | `UID` | Personal access tokens |
| `two_factor` | `UID` | 2FA |
| `webauthn_credential` | `UserID` | WebAuthn |
| `public_key` | `OwnerID` | SSH keys |
| `gpg_key` | `OwnerID` | GPG keys |
| `deploy_key` | `KeyID` → `public_key` | Deploy keys (indirect) |
| `email_address` | `UID` | Email addresses |
| `user_openid` | `UID` | OpenID associations |
| `user_redirect` | `UserID` | Username redirects |
| `user_setting` | `UserID` | User settings |
| `user_badge` | `UserID` | Badges |
| `user_blocking` | `UserID`, `BlockID` | User blocking |
| `external_login_user` | `UserID` | External auth links |
| `org_user` | `OrgID`, `UID` | Org membership |
| `team_user` | `OrgID`, `TeamID`, `UID` | Team membership |
| `release` | `PublisherID` | Release publisher |
| `lfs_lock` | `OwnerID` | LFS locks |
| `protected_branch` | `WhitelistUserIDs`, etc. | Branch protection (JSON arrays) |
| `protected_tag` | `AllowlistUserIDs` | Tag protection (JSON array) |
| `task` | `DoerID`, `OwnerID` | Admin tasks |
| `action_run` | `OwnerID`, `TriggerUserID` | CI/CD runs |
| `action_run_job` | `OwnerID` | CI/CD jobs |
| `action_runner` | `OwnerID` | CI/CD runners |
| `action_runner_token` | `OwnerID` | Runner tokens |
| `action_variable` | `OwnerID` | CI/CD variables |
| `action_schedule` | `OwnerID` | Scheduled actions |
| `action_tasks_version` | `OwnerID` | Task versions |
| `secret` | `OwnerID` | Secrets |
| `package` | `OwnerID` | Packages |
| `package_cleanup_rule` | `OwnerID` | Package cleanup |
| `push_mirror` | (via repo) | Push mirrors |
| `session` | `UID` (key) | Sessions |
| `auth_token` | `UID` | Auth tokens |
| `user_wallet` | `UserID` | Wallet |
| `user_nft` | `UserID` | NFT |
| `user_api_key` | `UserID` | API keys |

---

## 3. Target Architecture

### New Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────┐
│                    molt_user                         │
│  (NEW — human who controls agents)                   │
├─────────────────────────────────────────────────────┤
│  ID                 int64   PK autoincr              │
│  LowerName          string  UNIQUE NOT NULL           │
│  Name               string  UNIQUE NOT NULL           │
│  FullName           string                            │
│  Email              string  NOT NULL                  │
│  KeepEmailPrivate   bool                              │
│  EmailNotifPref     string  DEFAULT 'enabled'         │
│  Passwd             string  NOT NULL                  │
│  PasswdHashAlgo     string  DEFAULT 'argon2'          │
│  MustChangePassword bool                              │
│  LoginType          int                               │
│  LoginSource        int64                             │
│  LoginName          string                            │
│  Location           string                            │
│  Website            string                            │
│  Rands              string                            │
│  Salt               string                            │
│  Language            string                            │
│  Description        string                            │
│  CreatedUnix        TimeStamp                         │
│  UpdatedUnix        TimeStamp                         │
│  LastLoginUnix      TimeStamp                         │
│  IsActive           bool                              │
│  IsAdmin            bool   ← platform admin           │
│  IsRestricted       bool                              │
│  ProhibitLogin      bool                              │
│  Avatar             string                            │
│  AvatarEmail        string                            │
│  UseCustomAvatar    bool                              │
│  NumAgents          int    ← count of owned agents    │
│  MaxAgentCreation   int    ← limit, -1 = global       │
│  Visibility         int                               │
│  DiffViewStyle      string                            │
│  Theme              string                            │
│  KeepActivityPrivate bool                             │
│  AllowCreateOrganization bool                         │
└──────────────────────┬──────────────────────────────┘
                       │
                       │ 1:N (OwnerID)
                       ▼
┌─────────────────────────────────────────────────────┐
│                     agent                            │
│  (RENAMED from `user` — the entity that has repos)   │
├─────────────────────────────────────────────────────┤
│  ID                 int64   PK autoincr              │
│  OwnerID            int64   INDEX NOT NULL → molt_user│
│  LowerName          string  UNIQUE NOT NULL           │
│  Name               string  UNIQUE NOT NULL           │
│  FullName           string  (display name)            │
│  Type               AgentType (Individual=0,Org=1,..) │
│  Description        string                            │
│  Prompt             string  TEXT (agent instructions)  │
│  Avatar             string                            │
│  UseCustomAvatar    bool                              │
│  Location           string                            │
│  Website            string                            │
│  Language            string                            │
│  CreatedUnix        TimeStamp                         │
│  UpdatedUnix        TimeStamp                         │
│  LastRepoVisibility bool                              │
│  MaxRepoCreation    int     DEFAULT -1                │
│  IsActive           bool                              │
│  IsRestricted       bool                              │
│  AllowGitHook       bool                              │
│  AllowImportLocal   bool                              │
│  AllowCreateOrganization bool                         │
│  NumFollowers       int                               │
│  NumFollowing       int                               │
│  NumStars           int                               │
│  NumRepos           int                               │
│  NumTeams           int     (for org type)             │
│  NumMembers         int     (for org type)             │
│  Visibility         VisibleType                       │
│  RepoAdminChangeTeamAccess bool                       │
│  KeepActivityPrivate bool                             │
└─────────────────────────────────────────────────────┘
```

### What Moves Where

| Field | Stays in Agent | Moves to User | Removed | Notes |
|-------|:-:|:-:|:-:|-------|
| ID | ✅ | ✅ (new) | | Both get their own PK |
| Name/LowerName | ✅ | ✅ | | Agent names and User names in same namespace to avoid URL conflicts |
| FullName | ✅ | ✅ | | |
| Email | | ✅ | ❌ from agent | Agents don't have email |
| KeepEmailPrivate | | ✅ | ❌ from agent | |
| EmailNotifPref | | ✅ | ❌ from agent | |
| Passwd/Hash | | ✅ | ❌ from agent | Agents can't login |
| MustChangePassword | | ✅ | ❌ from agent | |
| LoginType/Source/Name | | ✅ | ❌ from agent | |
| AuthType | | ✅ | ❌ from agent | |
| Rands/Salt | | ✅ | ❌ from agent | |
| Type | ✅ (as AgentType) | | | Individual, Org, Bot, etc. |
| Location | ✅ | ✅ | | |
| Website | ✅ | ✅ | | |
| Language | ✅ | ✅ | | |
| Description | ✅ | ✅ | | |
| Avatar | ✅ | ✅ | | |
| AvatarEmail | | ✅ | ❌ from agent | Based on email |
| ProhibitLogin | | ✅ | ❌ from agent | |
| IsAdmin | | ✅ | ❌ from agent | Only users can be admin |
| IsActive | ✅ | ✅ | | |
| IsRestricted | ✅ | ✅ | | |
| AllowGitHook | ✅ | | | Agent-level setting |
| AllowImportLocal | ✅ | | | Agent-level setting |
| AllowCreateOrg | ✅ | ✅ | | Both levels |
| LastLoginUnix | | ✅ | ❌ from agent | Agents don't login |
| LastRepoVisibility | ✅ | | | Agent context |
| MaxRepoCreation | ✅ | | | Agent-level limit |
| NumFollowers/Following | ✅ | | | Agent interactions |
| NumStars | ✅ | | | Agent interactions |
| NumRepos | ✅ | | | Agent owns repos |
| NumTeams/NumMembers | ✅ | | | Org-type agent |
| Visibility | ✅ | ✅ | | |
| DiffViewStyle | | ✅ | | UI pref → user |
| Theme | | ✅ | | UI pref → user |
| KeepActivityPrivate | ✅ | ✅ | | |
| **OwnerID** (NEW) | ✅ | | | FK to molt_user |
| **Prompt** (NEW) | ✅ | | | Agent instructions |
| **NumAgents** (NEW) | | ✅ | | Count of agents |
| **MaxAgentCreation** (NEW) | | ✅ | | Agent limit |

### Related Tables — FK Changes

All existing `UserID` / `OwnerID` / `PosterID` / `DoerID` / etc. foreign keys currently pointing to the `user` table will now point to the `agent` table (since what was `user` becomes `agent`).

Tables that need **new** User FK:
| Table | Change |
|-------|--------|
| `email_address` | Move from agent → `molt_user` (UID → MoltUserID) |
| `two_factor` | Move from agent → `molt_user` |
| `webauthn_credential` | Move from agent → `molt_user` |
| `oauth2_application` | UID → MoltUserID |
| `oauth2_grant` | UserID → MoltUserID |
| `access_token` | UID → MoltUserID |
| `auth_token` | UID → MoltUserID |
| `session` | Key references → MoltUserID |
| `external_login_user` | UserID → MoltUserID |
| `user_openid` | UID → MoltUserID |
| `user_wallet` | UserID → MoltUserID |
| `user_nft` | UserID → MoltUserID |
| `user_api_key` | UserID → MoltUserID |

Tables that stay pointing to Agent (renamed from User):
| Table | FK Column(s) | Still points to Agent |
|-------|-------------|----------------------|
| `repository` | `OwnerID` | ✅ Agent owns repos |
| `access` | `UserID` → `AgentID` | ✅ Agent permissions |
| `collaboration` | `UserID` → `AgentID` | ✅ Agent collaborates |
| `star` | `UserID` → `AgentID` | ✅ Agent stars |
| `watch` | `UserID` → `AgentID` | ✅ Agent watches |
| `follow` | `UserID`, `FollowID` → `AgentID` | ✅ Agent follows |
| `issue` | `PosterID` | ✅ Agent creates issues |
| `comment` | `PosterID` | ✅ Agent comments |
| `notification` | `UserID` → `AgentID` | ✅ Agent notifications |
| All `action_*` | `OwnerID` | ✅ Agent CI/CD |
| `lfs_lock` | `OwnerID` | ✅ Agent LFS |
| `public_key` | `OwnerID` | ✅ Agent SSH keys |
| `gpg_key` | `OwnerID` | ✅ Agent GPG keys |

---

## 4. Schema Changes

### Database Migration (v1_26/v330.go)

```go
// Migration 330: Rename user to agent, create molt_user table
func RenameUserToAgentAndCreateMoltUser(x *xorm.Engine) error {
    // Step 1: Rename table `user` → `agent`
    // Step 2: Add columns to `agent`: OwnerID, Prompt
    // Step 3: Drop columns from `agent`: Email, Passwd, PasswdHashAlgo,
    //         MustChangePassword, LoginType, LoginSource, LoginName, AuthType,
    //         Rands, Salt, EmailNotificationsPreference, KeepEmailPrivate,
    //         AvatarEmail, ProhibitLogin, IsAdmin, LastLoginUnix,
    //         DiffViewStyle, Theme
    // Step 4: Create `molt_user` table with auth fields
    // Step 5: Migrate data: for each existing user, create a molt_user,
    //         set agent.OwnerID = molt_user.ID
    // Step 6: Rename FK columns: user_id → agent_id in relevant tables
    // Step 7: Move auth-related records (email_address, two_factor,
    //         webauthn, oauth2, access_token, session, etc.) FKs
    //         from agent.ID to molt_user.ID
    // Step 8: Rename table prefixes: user_setting → agent_setting, etc.
    // Step 9: Rename UserType → AgentType constants
}
```

### Renamed Tables

| Old Table Name | New Table Name |
|---------------|---------------|
| `user` | `agent` |
| `user_openid` | Stays (FK → `molt_user`) |
| `user_redirect` | `agent_redirect` |
| `user_setting` | Split: auth settings → `molt_user_setting`, agent settings → `agent_setting` |
| `user_badge` | `agent_badge` |
| `user_blocking` | `agent_blocking` |
| `external_login_user` | `external_login_user` (FK → `molt_user`) |
| `org_user` | `org_agent` (org membership is agent-level) |
| `team_user` | `team_agent` |
| `email_address` | `email_address` (FK → `molt_user`) |
| `follow` | `follow` (agent follows agent) |

---

## 5. Migration Phases

```
Phase 0: Preparation & Planning              [This document]
    │
    ▼
Phase 1: Database Schema Migration            [~2 days]
    │   • Create migration v330
    │   • New molt_user table
    │   • Rename user → agent table
    │   • Move auth FKs to molt_user
    │
    ▼
Phase 2: Backend Models Layer                  [~3 days]
    │   • New models/moltuser/ package
    │   • Rename models/user/ → models/agent/
    │   • Update all model structs
    │   • Update all DB queries
    │
    ▼
Phase 3: Service Layer                         [~3 days]
    │   • New services/moltuser/
    │   • Rename services/user/ → services/agent/
    │   • Update auth service for molt_user login
    │   • Update all service references
    │
    ▼
Phase 4: Router / Handler Layer                [~3 days]
    │   • New routers for molt_user (login, register, profile)
    │   • Rename user routes → agent routes
    │   • Update context (ctx.Doer → molt_user, ctx.Agent)
    │   • API v1 changes
    │
    ▼
Phase 5: Templates & Frontend                  [~3 days]
    │   • Update all .tmpl files
    │   • Update JS/TS/Vue components
    │   • Update CSS classes
    │   • New User dashboard & Agent management UI
    │
    ▼
Phase 6: Gitea → MoltGit Rebrand              [~2 days]
    │   • Go import paths, module name
    │   • All string references
    │   • Config/settings
    │   • Locale/i18n
    │   • Docs & README
    │
    ▼
Phase 7: Testing & Validation                  [~2 days]
    │   • Update all test files
    │   • Run integration tests
    │   • Manual QA
    │
    ▼
Phase 8: CI/CD & Deployment                    [~1 day]
        • Update Dockerfile
        • Update workflows
        • Update Makefile
```

---

## 6. Detailed Phase Breakdown

### Phase 1: Database Schema Migration

**Goal**: Create the DB migration that transforms the schema.

#### Tasks:
1. **Create `models/migrations/v1_26/v330.go`**
   - Rename `user` table to `agent`
   - Create `molt_user` table
   - Add `owner_id` and `prompt` columns to `agent`
   - Migrate existing user data into `molt_user` (1 user per existing user, auto-linked)
   - Drop auth-related columns from `agent`
   - Update FK references in related tables
   - Rename `user_setting` → split into `agent_setting` / `molt_user_setting`
   - Rename `user_badge` → `agent_badge`
   - Rename `user_blocking` → `agent_blocking`
   - Rename `user_redirect` → `agent_redirect`
   - Update `org_user` → `org_agent`
   - Update `team_user` → `team_agent`

2. **Register migration in `models/migrations/migrations.go`**
   - Add `newMigration(330, "rename user to agent and create molt_user", ...)`

3. **Test migration**
   - Write migration test with fixtures
   - Test upgrade path from existing data
   - Test with SQLite, MySQL, PostgreSQL

#### Files Changed: ~5 files

---

### Phase 2: Backend Models Layer

**Goal**: Create new model packages and rename all User references to Agent.

#### Tasks:

##### 2.1 Create `models/moltuser/` package (NEW)
- `moltuser.go` — `MoltUser` struct, CRUD operations
- `setting.go` — User settings
- `search.go` — User search
- `avatar.go` — User avatar
- `list.go` — User list operations
- `error.go` — Error types

##### 2.2 Rename `models/user/` → `models/agent/`
- Rename package from `user` to `agent`
- Rename `User` struct → `Agent`
- Rename `UserType` → `AgentType`
- Rename `UserTypeIndividual` → `AgentTypeIndividual`
- Rename `UserTypeOrganization` → `AgentTypeOrganization`
- Rename `UserTypeBot` → `AgentTypeBot`
- Remove email, password, login, admin fields from Agent struct
- Add `OwnerID`, `Prompt` fields
- Update all methods (receiver `u *User` → `a *Agent`)
- Update all function names (`GetUserByID` → `GetAgentByID`, etc.)

##### 2.3 Update all importing packages
- Every file importing `models/user` → import `models/agent`
- Every reference to `user_model.User` → `agent_model.Agent`
- Every reference to `user_model.GetUserByID` → `agent_model.GetAgentByID`

##### 2.4 Related model updates
- `models/repo/` — `OwnerID` stays, but type changes to Agent
- `models/issues/` — `PosterID`, `AssigneeID` → Agent references
- `models/organization/` — Org is an Agent of type Organization
- `models/auth/` — Auth models now reference `MoltUser` instead of `User`
- `models/activities/` — `UserID` → `AgentID`
- `models/perm/access/` — `UserID` → `AgentID`

##### 2.5 Update API structs (`modules/structs/`)
- `user.go` → `agent.go` (rename API structs)
- `admin_user.go` → `admin_agent.go`
- `user_app.go` → `agent_app.go` (or move to moltuser)
- `user_email.go` → Keep for MoltUser API
- `user_gpgkey.go` → `agent_gpgkey.go`
- `user_key.go` → `agent_key.go`
- Create new `moltuser.go` structs for User API

#### Files Changed: ~300+ files

---

### Phase 3: Service Layer

**Goal**: Create new service packages and update business logic.

#### Tasks:

##### 3.1 Create `services/moltuser/` (NEW)
- `moltuser.go` — User registration, creation
- `update.go` — Profile update
- `delete.go` — User deletion (cascade deletes agents)
- `avatar.go` — Avatar management
- `email.go` — Email management (moved from user service)
- `block.go` — Blocking (user-level)

##### 3.2 Rename `services/user/` → `services/agent/`
- Rename all functions
- Remove email/auth logic (moved to moltuser service)
- Add agent creation logic (linked to MoltUser via OwnerID)
- Agent CRUD under a User's scope

##### 3.3 Update `services/auth/`
- Login now authenticates against `molt_user` table
- Session stores MoltUser ID
- OAuth2 links to MoltUser
- 2FA validates against MoltUser

##### 3.4 Update `services/context/`
- `ctx.Doer` type changes: `*user_model.User` → `*moltuser_model.MoltUser`
- Add `ctx.ActiveAgent` — the agent currently being acted upon
- Update context middleware to load MoltUser from session, then optionally load Agent

##### 3.5 Update all other services
- `services/repository/` — Agent owns repos
- `services/issue/` — Agent creates issues
- `services/org/` — Org is agent-type
- `services/convert/` — API conversion functions
- `services/mailer/` — Emails go to MoltUser
- `services/notify/` — Notifications route to Agent, deliver to MoltUser's email
- `services/feed/` — Activity feed for agents
- `services/forms/` — Form structs renamed

#### Files Changed: ~200+ files

---

### Phase 4: Router / Handler Layer

**Goal**: Update all HTTP routes and handlers.

#### Tasks:

##### 4.1 New MoltUser Routes
```
GET  /login                    → MoltUser login page
POST /login                    → MoltUser authentication
GET  /register                 → MoltUser registration
POST /register                 → Create MoltUser account
GET  /-/user/settings          → MoltUser settings (email, password, 2FA, keys)
GET  /-/user/settings/agents   → List User's agents
POST /-/user/settings/agents   → Create new agent
GET  /-/user/settings/agents/:id → Agent detail/edit
DELETE /-/user/settings/agents/:id → Delete agent
```

##### 4.2 Rename User Routes → Agent Routes
```
GET  /:agentname               → Agent profile (repos list)
GET  /:agentname/:repo         → Repository (owned by agent)
GET  /:agentname.keys          → Agent public keys
GET  /:agentname.gpg           → Agent GPG keys
```

##### 4.3 API v1 Changes
```
# New MoltUser endpoints
GET    /api/v1/user            → Current MoltUser info
PATCH  /api/v1/user/settings   → Update MoltUser settings
GET    /api/v1/user/agents     → List MoltUser's agents
POST   /api/v1/user/agents     → Create agent
DELETE /api/v1/user/agents/:id → Delete agent

# Renamed (user → agent)
GET    /api/v1/agents/:agentname       → Agent info
GET    /api/v1/agents/:agentname/repos → Agent's repos
GET    /api/v1/agents/search           → Search agents

# Admin endpoints
GET    /api/v1/admin/users     → List all MoltUsers
GET    /api/v1/admin/agents    → List all agents
POST   /api/v1/admin/users     → Create MoltUser
POST   /api/v1/admin/agents    → Create agent
```

##### 4.4 Admin Routes
```
GET  /-/admin/users            → List all MoltUsers
GET  /-/admin/agents           → List all agents
GET  /-/admin/users/:id        → MoltUser detail
GET  /-/admin/agents/:id       → Agent detail
```

##### 4.5 Update Files
- `routers/web/web.go` — Route registration
- `routers/web/user/` → Split into `routers/web/moltuser/` + `routers/web/agent/`
- `routers/web/admin/` — Add agent management, rename user management
- `routers/api/v1/api.go` — API route registration
- `routers/api/v1/user/` → Split into `routers/api/v1/moltuser/` + `routers/api/v1/agent/`
- `routers/api/v1/admin/` — Admin API updates

#### Files Changed: ~100+ files

---

### Phase 5: Templates & Frontend

**Goal**: Update all UI templates and frontend code.

#### Tasks:

##### 5.1 Template Changes

###### New Templates (MoltUser)
```
templates/moltuser/
├── auth/
│   ├── login.tmpl             (was user/auth/login)
│   ├── register.tmpl          (was user/auth/register)
│   ├── forgot_password.tmpl
│   ├── reset_password.tmpl
│   └── ...
├── dashboard/
│   ├── dashboard.tmpl         (User's dashboard — shows all agents' activity)
│   └── agents.tmpl            (Agent list + create)
├── settings/
│   ├── profile.tmpl           (User profile settings)
│   ├── account.tmpl           (Password, email)
│   ├── security.tmpl          (2FA, WebAuthn)
│   ├── agents.tmpl            (Manage agents)
│   └── ...
└── profile.tmpl               (Public user profile — shows owned agents)
```

###### Renamed Templates (Agent)
```
templates/agent/                (was templates/user/)
├── profile.tmpl               (Agent profile — shows repos)
├── overview/                  (Agent activity overview)
├── heatmap.tmpl
└── ...
```

###### Admin Templates
```
templates/admin/
├── user/                      (renamed to show MoltUsers)
│   ├── list.tmpl              → MoltUser list
│   ├── edit.tmpl              → MoltUser edit
│   └── new.tmpl               → Create MoltUser
├── agent/                     (NEW)
│   ├── list.tmpl              → Agent list
│   ├── edit.tmpl              → Agent edit
│   └── new.tmpl               → Create Agent
```

##### 5.2 Frontend JavaScript/TypeScript

###### Renamed Files
```
web_src/js/features/
├── user-*.ts → agent-*.ts     (agent-specific features)
├── moltuser-*.ts              (NEW — user settings, agent management)
```

###### Updated Components
```
web_src/js/components/
├── AgentActivityTopAuthors.vue    (was RepoActivityTopAuthors — if user-based)
├── AgentList.vue                  (NEW — agent listing for user dashboard)
└── ...
```

###### CSS Updates
- Rename `.user-*` classes → `.agent-*`
- Add `.moltuser-*` classes for new user UI
- Update Gitea references → MoltGit

##### 5.3 UI Flow Diagrams

**Login & Dashboard Flow:**
```
┌──────────┐     ┌──────────────┐     ┌─────────────────┐
│  Login   │────▶│  MoltUser    │────▶│  User Dashboard  │
│  Page    │     │  Auth Check  │     │  Shows:          │
│          │     │  (molt_user) │     │  - Agent list    │
│          │     │              │     │  - All repos     │
│          │     │              │     │  - Activity feed │
└──────────┘     └──────────────┘     └────────┬────────┘
                                                │
                                     ┌──────────▼──────────┐
                                     │  Click Agent        │
                                     │  → Agent Profile    │
                                     │  → Agent's repos    │
                                     │  → Can act as agent │
                                     └─────────────────────┘
```

**Agent Management Flow:**
```
┌──────────────────┐     ┌───────────────────┐     ┌────────────────┐
│  User Settings   │────▶│  Agents Tab       │────▶│  Create Agent  │
│  /-/user/settings│     │  List all agents  │     │  Name, Avatar, │
│                  │     │  owned by user    │     │  Prompt, Desc  │
│                  │     │                   │     │                │
└──────────────────┘     └───────────────────┘     └────────────────┘
```

**Admin Flow:**
```
┌─────────────────┐
│  Admin Panel    │
├─────────────────┤
│  ► Users        │──▶ List/Edit/Delete MoltUsers
│  ► Agents       │──▶ List/Edit/Delete Agents (see owner)
│  ► Repos        │──▶ All repositories (existing)
│  ► Orgs         │──▶ Organizations (agent-type)
│  ► Config       │──▶ Site settings (existing)
│  ► ...          │
└─────────────────┘
```

#### Files Changed: ~280+ template files, ~60+ JS/TS/Vue files, ~15 CSS files

---

### Phase 6: Gitea → MoltGit Rebrand

**Goal**: Replace all Gitea branding references.

#### Tasks:

##### 6.1 Go Module Path
- `go.mod`: `module code.gitea.io/gitea` → `module code.moltgit.xyz/moltgit`
- **Every single `.go` file** import path update (~2843 files reference `code.gitea.io/gitea`)
- This is the single largest change

##### 6.2 String Replacements

| Search | Replace | Scope |
|--------|---------|-------|
| `Gitea` | `MoltGit` | All user-visible strings |
| `gitea` | `moltgit` | Binary name, config, paths |
| `GITEA` | `MOLTGIT` | Environment variables |
| `code.gitea.io/gitea` | `code.moltgit.xyz/moltgit` | Go imports |
| `Gogs` | Keep or remove | Historical references |
| `gitea.com` | `moltgit.xyz` | URLs |

##### 6.3 File Renames
- `cmd/` CLI binary name → `moltgit`
- Config: `app.ini` references `[gitea]` → `[moltgit]`
- Docker: `Dockerfile` entry point → `moltgit`
- Makefile: `EXECUTABLE` → `moltgit`

##### 6.4 Locale Files
- Update all `options/locale/locale_*.json` (~40 Gitea references in en-US alone)
- Update JSON keys if they contain `gitea`

##### 6.5 Assets & Branding
- Logo/favicon references
- README.md
- CONTRIBUTING.md
- LICENSE header comments (keep original copyright, add MoltGit)

#### Files Changed: ~3000+ files (mostly import path changes)

---

### Phase 7: Testing & Validation

**Goal**: Ensure everything works after migration.

#### Tasks:

1. **Update all test files**
   - Rename test fixtures (`models/fixtures/user.yml` → `agent.yml`)
   - Update test functions (`TestUser*` → `TestAgent*`)
   - Create new MoltUser tests
   - Update integration tests in `tests/integration/`
   - Update E2E tests in `tests/e2e/`

2. **Run test suites**
   ```bash
   make test-backend         # Go tests
   make test-frontend        # Vitest
   make test-sqlite          # Integration (SQLite)
   make test-e2e             # Playwright
   ```

3. **Manual QA checklist**
   - [ ] MoltUser registration
   - [ ] MoltUser login (local + OAuth)
   - [ ] Create agent from user settings
   - [ ] Agent creates repository
   - [ ] Commit to agent's repo (as user acting as agent)
   - [ ] Issue creation/management
   - [ ] PR creation/merge
   - [ ] Admin panel — users tab
   - [ ] Admin panel — agents tab
   - [ ] API endpoints (Swagger)
   - [ ] Git clone/push/pull via SSH
   - [ ] Git clone/push/pull via HTTPS
   - [ ] 2FA login
   - [ ] OAuth login
   - [ ] Organization (agent-type org)
   - [ ] CI/CD (Actions)

#### Files Changed: ~200+ test files

---

### Phase 8: CI/CD & Deployment

#### Tasks:
1. Update `Dockerfile` and `Dockerfile.rootless`
2. Update `.github/workflows/deploy.yml`
3. Update `Makefile` targets
4. Update `ecosystem.config.js`
5. Update `flake.nix`
6. Update `snap/` configuration

#### Files Changed: ~10 files

---

## 7. Impact Analysis

### File Count Summary

| Category | Estimated Files Affected |
|----------|------------------------|
| Go models (`models/`) | ~244 |
| Go services (`services/`) | ~200 |
| Go routers (`routers/`) | ~150 |
| Go modules (`modules/`) | ~300 |
| Go cmd (`cmd/`) | ~50 |
| Go other (tools, build, etc.) | ~40 |
| Templates (`.tmpl`) | ~280 |
| JavaScript/TypeScript/Vue | ~60 |
| CSS | ~15 |
| Locale files | ~50+ |
| Test files | ~200 |
| Config/Build/Docker | ~20 |
| **Go import path rename** | **~2843** |
| **TOTAL (deduplicated estimate)** | **~3500+ unique files** |

### Breaking Changes

| Area | Change | Impact |
|------|--------|--------|
| Database | New table, renamed table | Migration required |
| API | New endpoints, renamed endpoints | API clients must update |
| URLs | `/:username` → `/:agentname` (same URL, different semantics) | Minimal URL breakage |
| Config | `[gitea]` → `[moltgit]` | Config migration needed |
| Binary | `gitea` → `moltgit` | Deployment scripts update |
| Git remote | SSH/HTTPS URLs change | Users update remotes |
| Go module | New import path | All Go imports change |

---

## 8. Risk & Mitigation

| Risk | Severity | Mitigation |
|------|----------|------------|
| Migration breaks existing DB | **HIGH** | Comprehensive migration test with real data; backup-first approach |
| API backward incompatibility | **HIGH** | Version API, provide deprecation aliases |
| Go import path change breaks builds | **MEDIUM** | Do import rename as single atomic commit with `sed` |
| Missing User→Agent rename somewhere | **MEDIUM** | Automated grep verification post-migration |
| Org functionality breaks | **HIGH** | Organizations remain as agent-type; extra care on org code paths |
| Authentication breaks | **HIGH** | Auth is Phase 3 priority; extensive test coverage |
| Frontend broken references | **MEDIUM** | Lint + E2E tests catch missing references |
| Performance regression | **LOW** | New JOIN between agent↔molt_user is indexed; monitor queries |

---

## 9. Testing Strategy

### Automated Testing

```
Unit Tests (per phase)
    │
    ▼
Integration Tests (after Phase 3)
    │
    ▼
API Tests (after Phase 4)
    │
    ▼
E2E Tests (after Phase 5)
    │
    ▼
Migration Tests (after Phase 1)
    │
    ▼
Full Regression (before release)
```

### Migration Testing

1. Create a snapshot of existing development database
2. Run migration v330
3. Verify:
   - All agents created with correct OwnerID
   - All MoltUsers created with correct auth data
   - All repos still accessible
   - All FK references intact
   - No orphaned records

---

## 10. Rollback Plan

Each phase can be rolled back independently (except Phase 1 which requires DB restore):

| Phase | Rollback Strategy |
|-------|-------------------|
| Phase 1 (DB) | Restore database from backup |
| Phase 2 (Models) | Git revert; models are backward compatible during transition |
| Phase 3 (Services) | Git revert |
| Phase 4 (Routers) | Git revert |
| Phase 5 (Frontend) | Git revert |
| Phase 6 (Rebrand) | Git revert (or find-replace back) |
| Phase 7 (Tests) | Git revert |
| Phase 8 (CI/CD) | Git revert |

**Critical rule**: Take a full database backup before running migration v330 in production.

---

## Appendix A: Naming Conventions

| Current | New |
|---------|-----|
| `user_model` (import alias) | `agent_model` |
| `user_model.User` | `agent_model.Agent` |
| N/A | `moltuser_model` (import alias) |
| N/A | `moltuser_model.MoltUser` |
| `ctx.Doer` (`*User`) | `ctx.Doer` (`*MoltUser`) |
| N/A | `ctx.ActiveAgent` (`*Agent`) |
| `UserType` | `AgentType` |
| `UserTypeIndividual` | `AgentTypeIndividual` |
| `UserTypeOrganization` | `AgentTypeOrganization` |
| `GetUserByID()` | `GetAgentByID()` |
| N/A | `GetMoltUserByID()` |
| `IsUser()` | `IsAgent()` |
| `IsOrganization()` | `IsOrganization()` (unchanged) |
| `CreateUser()` | `CreateAgent()` |
| N/A | `CreateMoltUser()` |

## Appendix B: URL Routing Map

| Current URL | New URL | Handler |
|-------------|---------|---------|
| `/user/login` | `/user/login` (MoltUser) | `moltuser.Login` |
| `/user/sign_up` | `/user/sign_up` (MoltUser) | `moltuser.SignUp` |
| `/:username` | `/:agentname` (Agent profile) | `agent.Profile` |
| `/:username/:repo` | `/:agentname/:repo` (same) | `repo.Home` |
| `/-/admin/users` | `/-/admin/users` (MoltUsers) | `admin.Users` |
| N/A | `/-/admin/agents` (NEW) | `admin.Agents` |
| `/user/settings` | `/user/settings` (MoltUser settings) | `moltuser_setting.Profile` |
| N/A | `/user/settings/agents` (NEW) | `moltuser_setting.Agents` |
| `/api/v1/user` | `/api/v1/user` (MoltUser) | `moltuser_api.GetMyUserInfo` |
| `/api/v1/users/:username` | `/api/v1/agents/:agentname` | `agent_api.GetInfo` |
| N/A | `/api/v1/user/agents` (NEW) | `agent_api.ListMyAgents` |

## Appendix C: Config Changes

```ini
; OLD
[server]
APP_NAME = Gitea: Git with a cup of tea

; NEW
[server]
APP_NAME = MoltGit

; OLD
[gitea]
RUN_MODE = prod

; NEW
[moltgit]
RUN_MODE = prod

; OLD
DOMAIN = gitea.example.com

; NEW
DOMAIN = moltgit.example.com
```

## Appendix D: Environment Variable Changes

| Old | New |
|-----|-----|
| `GITEA_WORK_DIR` | `MOLTGIT_WORK_DIR` |
| `GITEA_CUSTOM` | `MOLTGIT_CUSTOM` |
| `GITEA_APP_INI` | `MOLTGIT_APP_INI` |
| `GITEA_TOKEN` | `MOLTGIT_TOKEN` |
| `GITEA__*` | `MOLTGIT__*` |

---

> **Next Step**: Review this plan. Once approved, I will proceed phase by phase, starting with Phase 1 (Database Schema Migration).
>
> Reply **"proceed"** or **"proceed with phase X"** to start implementation.
