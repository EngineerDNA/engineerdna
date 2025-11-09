# Contributing to EngineerDNA

Thank you for your interest in contributing to EngineerDNA!

## Code of Conduct

This project follows a professional and inclusive code of conduct. Please be respectful and constructive in all interactions.

## Development Setup

### Prerequisites

- Go 1.21+, Node.js 20+, Make, Git

### Initial Setup

```bash
git clone https://github.com/engineerdna/engineerdna.git
cd engineerdna
go mod download
cd frontend && npm install && cd ..
make build
./bin/engineerdna init
./scripts/restart-dev-server.sh --rebuild --frontend
```

### Development Workflow

**Backend**:
```bash
# Make changes, then run checks
go vet ./... && staticcheck ./... && go test ./...
./scripts/restart-dev-server.sh --rebuild
```

**Frontend**:
```bash
cd frontend
npm run lint && npm run typecheck && npm run build
npm run dev  # Hot reload dev server
```

**Logs**: `/tmp/engineerdna-backend.log`, `/tmp/engineerdna-frontend.log`

## Making Changes

### Branch Strategy

- `main` - Production-ready code
- Feature branches - Descriptive names (`feature/plugin-scheduler`, `fix/memory-leak`)

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Examples**:
```
feat(plugins): add CSV import plugin
fix(api): resolve race condition in event sync
docs(readme): update installation instructions
```

### File Naming

- **Go**: snake_case (`plugin_loader.go`), PascalCase types (`type PluginLoader struct`)
- **TypeScript/React**: kebab-case (`metric-card.tsx`), PascalCase components (`export function MetricCard()`)

## Code Quality Standards

### Go Code

Must pass before submission:
```bash
gofmt -l .
goimports -l .
go vet ./...
staticcheck ./...
golangci-lint run ./...
go test -v -race ./...
```

### Frontend Code

Must pass before submission:
```bash
cd frontend
npm run typecheck
npm run lint
npx prettier --check 'src/**/*.{ts,tsx,js,jsx,css,json}'
npm run build
npm test -- --run
```

### Security

- No hardcoded secrets or API keys
- All user inputs must be validated and sanitized
- Use parameterized queries (prevent SQL injection)
- Encrypt secrets at rest with AES-256-GCM
- Anonymize data before sending to external APIs

### Performance

- Files under 500 lines (enforced by hook)
- Functions under 50 lines
- Cyclomatic complexity under 15
- No N+1 queries in database operations
- Use connection pooling and caching where appropriate

### Documentation

- All public functions and types must have godoc comments
- Complex logic requires inline comments explaining "why" not "what"
- Update CHANGELOG.md for all user-facing changes
- Update relevant CLAUDE.md files for architectural changes

## Testing

### Backend Tests

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -v -race ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Organization**:
- Unit tests: `*_test.go` in same package
- Integration tests: `*_integration_test.go`
- Table-driven tests for multiple scenarios

**Example**:
```go
func TestEventStore_Create(t *testing.T) {
    tests := []struct {
        name    string
        event   *models.Event
        wantErr bool
    }{
        {"valid event", validEvent, false},
        {"duplicate event", duplicateEvent, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := store.Create(tt.event)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Frontend Tests

```bash
cd frontend
npm test                    # Run all tests
npm test -- --coverage      # With coverage
npm test -- EventList.test  # Specific test
```

**Test Organization**:
- Unit tests: `*.test.tsx` next to component
- Integration tests: `*.integration.test.tsx`

## Submitting Pull Requests

### Before Submitting

1. **Run all checks**:
   ```bash
   go vet ./... && staticcheck ./... && go test ./... && make build
   cd frontend && npm run lint && npm run typecheck && npm run build && cd ..
   ```

2. **Update documentation**:
   - Add to CHANGELOG.md under "Unreleased"
   - Update relevant CLAUDE.md files
   - Update API documentation if adding endpoints

3. **Write descriptive PR description**:
   - What: What does this PR do?
   - Why: Why is this change needed?
   - How: How does it work?
   - Testing: How was this tested?

### PR Template

```markdown
## Summary
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Added unit tests
- [ ] Added integration tests
- [ ] Manually tested

## Checklist
- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

### Review Process

1. Automated checks must pass (GitHub Actions)
2. At least one maintainer approval required
3. No merge conflicts with main
4. All review comments addressed

## Release Process

### Version Numbering

Follow Semantic Versioning (semver):
- `MAJOR.MINOR.PATCH`
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

### Release Steps

1. **Update version**:
   ```bash
   # Update CHANGELOG.md
   # Move "Unreleased" section to new version
   # Example: ## [1.2.0] - 2025-11-08
   ```

2. **Create release commit**:
   ```bash
   git add CHANGELOG.md
   git commit -m "Release v1.2.0"
   git push origin main
   ```

3. **Create GitHub release**:
   - GitHub Actions automatically builds and publishes release
   - Binaries for Linux, macOS, Windows (amd64 + arm64)
   - Docker images pushed to registry

## Common Issues

**Build fails with "cannot find package"**:
```bash
go mod download
go mod tidy
```

**Frontend dev server won't start**:
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run dev
```

**Port 3847 already in use**:
```bash
./scripts/restart-dev-server.sh  # Automatically kills old process
# Or manually:
lsof -ti:3847 | xargs kill -9
```

**Database locked error**:
```bash
pkill engineerdna
rm ~/.engineerdna/engineerdna.db-shm ~/.engineerdna/engineerdna.db-wal
```

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Join discussions for questions and ideas
- Read CLAUDE.md files for architecture details

## Recognition

Contributors are recognized in:
- CHANGELOG.md (for each release)
- GitHub contributors page
- Release notes

Thank you for contributing to EngineerDNA!
