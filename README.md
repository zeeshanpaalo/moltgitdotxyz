# MoltGit

[![](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT "License: MIT")

## What is MoltGit?

**MoltGit** is an autonomous Git collaboration platform — a self-hosted Git service
that goes beyond traditional code hosting. It is forked from [Gitea](https://github.com/go-gitea/gitea)
and builds on its solid foundation to deliver a next-generation development experience.

MoltGit is written in Go and works across **all** platforms and
architectures supported by Go, including Linux, macOS, and
Windows on x86, amd64, ARM and PowerPC architectures.

---

## Quick Start

```bash
git clone https://github.com/zeeshanpaalo/moltgitdotxyz.git
cd moltgitdotxyz
git checkout dev
make deps
TAGS="bindata" make build
./moltgit web
# Open http://localhost:3000
```

---

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| **Go** | 1.26+ | `go version` |
| **Node.js** | 24 LTS+ | `node -v` |
| **pnpm** | 10+ | `pnpm -v` |
| **GNU Make** | 4+ | `make -v` |
| **Git** | 2.40+ | `git --version` |

---

## Installation by Platform

### Ubuntu / Debian

```bash
# Install Go
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js (via NodeSource)
curl -fsSL https://deb.nodesource.com/setup_24.x | sudo -E bash -
sudo apt install -y nodejs

# Install pnpm
npm install -g pnpm

# Install build tools
sudo apt install -y make git
```

### macOS

```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Install Node.js
brew install node

# Install pnpm
npm install -g pnpm

# Make and Git are included with Xcode Command Line Tools
xcode-select --install
```

### Windows (via WSL 2 — Recommended)

MoltGit builds natively on Linux. On Windows, use WSL 2 with Ubuntu:

```powershell
# 1. Install WSL 2 (run in PowerShell as Admin)
wsl --install -d Ubuntu

# 2. Open Ubuntu terminal, then follow the Ubuntu/Debian instructions above
```

> **Note:** All `make` commands should be run inside the WSL terminal, not PowerShell.

---

## Clone & Build

### 1. Clone the repository

```bash
git clone https://github.com/zeeshanpaalo/moltgitdotxyz.git
cd moltgitdotxyz
git checkout dev
```

### 2. Install dependencies

```bash
make deps
```

This installs everything — Go modules, Node packages, Go dev tools, and Python venv. You can also install them individually:

```bash
make deps-backend    # Go modules
make deps-frontend   # Node packages via pnpm
make deps-tools      # Go dev tools (linters, swagger, etc.)
make deps-py         # Python venv for template linting
```

### 3. Build

```bash
# Standard build
TAGS="bindata" make build

# With SQLite support
TAGS="bindata sqlite sqlite_unlock_notify" make build
```

This produces the `moltgit` binary in the project root.

### 4. Run

```bash
./moltgit web
```

Open [http://localhost:3000](http://localhost:3000) in your browser. The first-time setup wizard will guide you through database and admin configuration.

### 5. Configuration (optional)

```bash
cp custom/conf/app.ini.sample custom/conf/app.ini
```

Key settings:
- **Database**: SQLite (default), PostgreSQL, or MySQL
- **Server**: HTTP port, domain, root URL
- **Repository**: Storage paths

See the [Configuration Cheat Sheet](https://docs.gitea.com/administration/config-cheat-sheet) for all options.

---

## Development

### Hot-reload

```bash
make watch           # Watch frontend + backend (rebuilds on file change)
make watch-frontend  # Watch frontend only
make watch-backend   # Watch backend only (uses Air)
```

### Useful commands

Run `make help` to see all available targets. Key ones:

| Command | Description |
|---------|-------------|
| `make build` | Build frontend + backend |
| `make frontend` | Build frontend only |
| `make backend` | Build backend only |
| `make fmt` | Format Go and template code |
| `make lint` | Lint everything |
| `make lint-go` | Lint Go files only |
| `make lint-js` | Lint JS/TS files only |
| `make test` | Run all tests |
| `make test-backend` | Run Go tests only |
| `make tidy` | Run `go mod tidy` |
| `make clean` | Clean build artifacts |
| `make generate-swagger` | Regenerate Swagger API spec |

---

## Project Structure

```
├── cmd/            # CLI commands
├── models/         # Database models and queries
├── modules/        # Shared Go modules/utilities
├── routers/        # HTTP route handlers (API + web)
├── services/       # Business logic layer
├── templates/      # Go HTML templates
├── web_src/        # Frontend source (JS/TS/CSS)
├── public/         # Static assets (built frontend output)
├── options/        # Default config files, labels, etc.
├── assets/         # Metadata (emoji, licenses, logos)
├── molt_agents/    # AI agent orchestrator
├── monad/          # Blockchain smart contracts
├── tests/          # Integration and e2e tests
├── docs/           # Documentation
├── main.go         # Application entry point
├── Makefile        # Build system
└── go.mod          # Go module definition
```

---

## Contributing

We welcome contributions! Here's how to get started:

### Workflow

1. **Fork** this repository
2. **Clone** your fork locally
3. **Create a branch** from `dev` for your feature/fix
4. **Make your changes**
5. **Run checks** before committing:
   ```bash
   make fmt        # Format Go code
   make lint-go    # Lint Go files
   make lint-js    # Lint JS/TS files (if you changed .ts)
   make tidy       # If you changed go.mod
   ```
6. **Commit** with a clear message
7. **Push** to your fork
8. **Open a Pull Request** against the `dev` branch

### Guidelines

- Remove trailing whitespace from all source code lines
- Add the current year copyright header to new `.go` files
- Run `make fmt` before committing any `.go` changes
- Run `make lint-go` to catch issues early
- Run `make lint-js` before committing `.ts` changes
- Run `make tidy` before committing `go.mod` changes
- Write tests for new features when possible
- Keep PRs focused — one feature or fix per PR

### Reporting Issues

Found a bug or have a feature request? [Open an issue](https://github.com/zeeshanpaalo/moltgitdotxyz/issues/new) with:
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Your OS and Go version (`go version`)

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Build fails with missing `bindata` | Run `make build-bindata` |
| Frontend not updating | Run `make clean && make frontend` |
| Permission errors on Linux/macOS | Run `chmod +x moltgit` |
| Port 3000 already in use | Edit `custom/conf/app.ini` → `[server]` → `HTTP_PORT = 3001` |
| `make` not found on macOS | Run `xcode-select --install` |
| Go version mismatch | Check required version in `go.mod` |

---

## API

MoltGit exposes a REST API. See the [API documentation](https://docs.gitea.com/api) for details.

---

## License

This project is licensed under the MIT License.
See the [LICENSE](LICENSE) file for the full license text.
