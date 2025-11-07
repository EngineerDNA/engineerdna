# Contributing to EngineerDNA

Thank you for your interest in contributing to EngineerDNA!

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Quality Standards](#code-quality-standards)
- [Testing](#testing)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Release Process](#release-process)

## Code of Conduct

This project follows a professional and inclusive code of conduct. Please be respectful and constructive in all interactions.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Node.js 20 or later
- Make
- Git

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/engineerdna/engineerdna.git
cd engineerdna

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Build the project
make build

# Initialize the database
./bin/engineerdna init

# Start the development server
./scripts/restart-dev-server.sh --rebuild --frontend
```

### Development Workflow

#### Backend Development

```bash
# Make changes to Go code
# ...

# Run checks
go vet ./...
staticcheck ./...
go test ./...

# Rebuild and restart
./scripts/restart-dev-server.sh --rebuild
```

#### Frontend Development

```bash
cd frontend

# Make changes to React code
# ...

# Run checks
npm run lint
npm run typecheck
npm run build

# Start dev server with hot reload
npm run dev
```

#### Using the Restart Script

The `restart-dev-server.sh` script is the recommended way to manage development servers:

```bash
# Basic restart (just restart the server)
./scripts/restart-dev-server.sh

# Rebuild and restart
./scripts/restart-dev-server.sh --rebuild

# Include frontend dev server (Vite on port 5173)
./scripts/restart-dev-server.sh --frontend

# Rebuild everything and start both servers
./scripts/restart-dev-server.sh --rebuild --frontend
```

Logs are written to:
- Backend: `/tmp/engineerdna-backend.log`
- Frontend: `/tmp/engineerdna-frontend.log`

## Making Changes

### Branch Strategy

- `main` - Production-ready code
- Feature branches - Use descriptive names (e.g., `feature/plugin-scheduler`, `fix/memory-leak`)

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, no logic changes)
- `refactor` - Code refactoring (no functional changes)
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

Examples:
```
feat(plugins): add CSV import plugin
fix(api): resolve race condition in event sync
docs(readme): update installation instructions
```

### File Naming Conventions

#### Go Files
- Use snake_case for file names: `plugin_loader.go`, `event_repository.go`
- Use lowercase for package names: `package plugin`
- Use PascalCase for type names in code: `type PluginLoader struct`

#### TypeScript/React Files
- Use kebab-case for file names: `metric-card.tsx`, `dashboard-layout.tsx`
- Use PascalCase for component names in code: `export function MetricCard()`

## Code Quality Standards

### Go Code

All Go code must pass these checks before submission:

```bash
# Formatting
gofmt -l .
goimports -l .

# Linting
go vet ./...
staticcheck ./...
golangci-lint run ./...

# Build
go build -v -o bin/engineerdna .

# Tests
go test -v -race ./...
```

### Frontend Code

All frontend code must pass these checks:

```bash
cd frontend

# Type checking
npm run typecheck

# Linting
npm run lint

# Formatting
npx prettier --check 'src/**/*.{ts,tsx,js,jsx,css,json}'

# Build
npm run build

# Tests
npm test -- --run
```

### Security

- No hardcoded secrets or API keys
- All secrets must use encryption (AES-256-GCM)
- External API calls require anonymization
- Localhost-only binding (127.0.0.1:3847)
- No `.env` files should be committed

## Testing

### Unit Tests

```bash
# Backend
go test ./...

# Frontend
cd frontend && npm test
```

### Integration Tests

```bash
# Test plugin system
go test ./internal/plugin/...

# Test API endpoints
go test ./internal/api/...
```

### Manual Testing

```bash
# Start server
./scripts/restart-dev-server.sh --rebuild --frontend

# Test in browser
open http://localhost:3847

# Test CLI commands
./bin/engineerdna version
./bin/engineerdna init
./bin/engineerdna sync
```

## Submitting Pull Requests

### Before Submitting

1. **Update CHANGELOG.md** - Add an entry describing your changes
2. **Bump version** - Update version in CHANGELOG.md if needed
3. **Run all checks** - Ensure all quality checks pass
4. **Test thoroughly** - Verify your changes work as expected
5. **Update documentation** - Update README.md or other docs if needed

### PR Process

1. **Create a branch** from `main`
2. **Make your changes** with clear commits
3. **Update CHANGELOG.md** with a new version and your changes
4. **Run quality checks** - All checks must pass
5. **Push your branch** to GitHub
6. **Create a pull request** with a clear description
7. **Wait for review** - Address any feedback
8. **Merge** - Once approved, your PR will be merged

### PR Requirements

All PRs must:
- Include CHANGELOG.md update with version bump
- Pass all quality checks (Go vet, lint, tests)
- Include tests for new functionality
- Have clear commit messages
- Be approved by at least one maintainer

## Release Process

EngineerDNA uses **git tags** as the single source of truth for versioning, with automated release management via GoReleaser.

### Creating a Release

Releases are triggered by pushing git tags to the repository.

#### Step 1: Update CHANGELOG.md

Add a new version section at the top of `CHANGELOG.md`:

```markdown
## [1.2.0] - 2025-11-05

### Added
- CSV import plugin with intelligent column mapping
- Identity resolution API with confidence scoring

### Fixed
- Memory leak in event sync process
- Race condition in plugin loader

### Security
- Improved input validation for API endpoints
```

#### Step 2: Commit and Push

```bash
# Commit CHANGELOG update
git add CHANGELOG.md
git commit -m "Release v1.2.0"
git push origin main
```

#### Step 3: Create Git Tag

```bash
# Create annotated tag
git tag -a v1.2.0 -m "Release v1.2.0"

# Push tag to trigger release
git push origin v1.2.0
```

#### Step 4: Automated Release (GitHub Actions)

When the tag is pushed, GitHub Actions automatically:

1. **Builds Frontend** - Compiles React app and embeds in Go binary
2. **Cross-Compiles Go** - Builds for 6 platforms (Linux, macOS, Windows × amd64/arm64)
3. **Injects Version** - Sets `main.Version = "v1.2.0"` in binary
4. **Creates Archives** - `.tar.gz` for Unix, `.zip` for Windows
5. **Generates Checksums** - SHA256 for all binaries
6. **Creates GitHub Release** - With release notes from CHANGELOG.md
7. **Uploads Binaries** - Attaches all artifacts to release

### Release Assets

Each release includes:

- **Source code** (ZIP and tar.gz)
- **Binaries** for 6 platforms:
  - `engineerdna-v1.2.0-linux-amd64.tar.gz`
  - `engineerdna-v1.2.0-linux-arm64.tar.gz`
  - `engineerdna-v1.2.0-darwin-amd64.tar.gz`
  - `engineerdna-v1.2.0-darwin-arm64.tar.gz`
  - `engineerdna-v1.2.0-windows-amd64.zip`
  - `engineerdna-v1.2.0-windows-arm64.zip`
- **Checksums** - `checksums.txt` with SHA256 hashes

### Version Management

**Single Source of Truth:** Git Tags

Version comes from (in priority order):
1. **Git tags** (production) - e.g., `v1.2.0`
2. **CHANGELOG.md** (fallback) - When no git tag exists
3. **`main.Version`** (injected at build time)

**Local Development:**
```bash
# Build without tags (uses CHANGELOG.md)
make build
# → Version: 1.0.0

# After creating a tag
git tag v1.1.0
make build
# → Version: v1.1.0

# Check version
./bin/engineerdna version
# → EngineerDNA version v1.1.0
```

### Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0) - Incompatible API changes
- **MINOR** version (0.X.0) - New functionality, backwards compatible
- **PATCH** version (0.0.X) - Bug fixes, backwards compatible

Examples:
- `1.0.0` � `1.0.1` - Bug fix
- `1.0.1` � `1.1.0` - New feature
- `1.1.0` � `2.0.0` - Breaking change

### Hotfix Releases

For urgent fixes to production:

```bash
# Create hotfix branch from tag
git checkout -b hotfix/1.1.1 v1.1.0

# Make fix and update CHANGELOG
vim CHANGELOG.md
git commit -am "Fix critical bug"

# Tag and push
git tag v1.1.1
git push origin v1.1.1

# Merge back to main
git checkout main
git merge hotfix/1.1.1
git push origin main
```

### Pre-Release Versions

For beta/RC releases, use pre-release suffixes:

```bash
# Create pre-release tag
git tag v1.2.0-beta.1
git push origin v1.2.0-beta.1

# GoReleaser auto-detects and marks as pre-release
```

### Version Validation

The `.claude/hooks/validate_version_bump.py` hook enforces:

1. **Code changes require CHANGELOG update** - Go, TypeScript, migrations, build configs
2. **CHANGELOG update requires version bump** - Version must increment
3. **No hardcoded versions** - Blocks hardcoded versions in code

**Example blocked commit:**
```bash
git commit -m "Add feature"
# [ERROR] Code changed - version bump required!
#
# Current version: 1.0.0
# Suggested: 1.1.0 (minor bump)
```

### Troubleshooting Releases

#### Release workflow fails

1. Check [GitHub Actions](https://github.com/EngineerDNA/engineerdna/actions)
2. Look for GoReleaser errors in workflow logs
3. Common issues:
   - Frontend build failure - Run `cd frontend && npm ci && npm run build`
   - Go build failure - Run `go build -v .`
   - Tag format incorrect - Must be `vX.Y.Z` (e.g., `v1.2.0`)

#### Version shows "dev" locally

```bash
# Check for git tags
git tag

# Create tag if missing
git tag v1.0.0
make build
```

#### Test GoReleaser locally

```bash
# Install GoReleaser
brew install goreleaser  # macOS
# OR
go install github.com/goreleaser/goreleaser@latest

# Test release (without pushing)
goreleaser release --snapshot --clean

# Check output
ls -la dist/
```

## Questions?

If you have questions about contributing:
- Open a [GitHub Discussion](https://github.com/engineerdna/engineerdna/discussions)
- Open a [GitHub Issue](https://github.com/engineerdna/engineerdna/issues)
- Check existing documentation in the repository

## License

By contributing to EngineerDNA, you agree that your contributions will be licensed under the Apache License 2.0.
