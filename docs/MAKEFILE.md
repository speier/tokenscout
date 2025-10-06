# Makefile Targets

Quick reference for all make commands.

## Development

```bash
# Run with debug logging
make dev

# Build binary (with version info)
make build

# Build with custom version
make build VERSION=v1.2.3

# Run tests
make test

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```

## Version Information

All builds automatically inject version information:

```bash
# Build with version info
make build

# Binary includes:
# - Version: from VERSION variable (default: v0.1.0)
# - Commit: git commit SHA (short)
# - Date: build timestamp (UTC)

# Check version
./tokenscout version
# Output:
# TokenScout v0.1.0
#   Commit:      abc1234
#   Built:       2025-10-06T16:01:39Z
#   Go version:  go1.21.0
#   OS/Arch:     darwin/arm64
```

## Release Management

```bash
# Create and push release (auto-increments minor version)
make release
# This runs: bump-minor → test → build → commit → tag → push

# Manual version control:
make bump-patch    # 0.1.0 → 0.1.1
make bump-minor    # 0.1.0 → 0.2.0
make bump-major    # 0.1.0 → 1.0.0

# Release without auto-bump (uses current VERSION)
make release-manual

# Test release locally (doesn't push to GitHub)
make release-test

# Check GoReleaser config
make release-check

# Build for all platforms manually
make build-all
```

**Version File:** Version is stored in `VERSION` file (e.g., `0.1.0`)  
**Auto-increment:** `make release` automatically bumps minor version  
**Manual control:** Use `bump-*` targets or edit VERSION file directly

## Dependencies

```bash
# Download and tidy Go modules
make deps

# Initialize config file
make init
```

## Release Workflow Example

```bash
# 1. Make your changes
git add .
git commit -m "feat: add new feature"
git push

# 2. Create release
make release VERSION=v1.0.0

# Output:
# Creating release v1.0.0
# ... creates tag v1.0.0 ...
# ... git push origin v1.0.0 ...
# ✓ Release v1.0.0 pushed. GitHub Actions will build and publish.

# 3. GitHub Actions automatically builds and publishes
# Check: https://github.com/speier/tokenscout/releases
```

## VERSION Variable

The `VERSION` variable defaults to `v0.1.0` but can be overridden:

```bash
# Use default version
make release

# Override version
make release VERSION=v2.1.5

# Pre-release versions
make release VERSION=v1.0.0-beta.1
make release VERSION=v2.0.0-rc.1
```

## Testing Releases

Before creating a real release, test locally:

```bash
# Check GoReleaser config is valid
make release-check

# Build release artifacts locally (creates ./dist/)
make release-test

# Inspect what would be released
ls -lh dist/
```

This creates a snapshot release in `./dist/` without pushing to GitHub.

## CI/CD Integration

The Makefile works alongside GitHub Actions:

- **`make release`** → Creates tag → Triggers `.github/workflows/release.yml`
- **Push to main** → Triggers `.github/workflows/ci.yml` (builds & tests)
- **Pull requests** → Triggers `.github/workflows/ci.yml` & `.github/workflows/test-release.yml`

## Tips

**Automate version bumping:**
```bash
# Create a script to auto-increment version
./scripts/bump-version.sh patch  # v1.0.0 → v1.0.1
./scripts/bump-version.sh minor  # v1.0.0 → v1.1.0
./scripts/bump-version.sh major  # v1.0.0 → v2.0.0
make release VERSION=$(cat VERSION)
```

**Check what would be released:**
```bash
make release-test
tree dist/
```

**Multiple releases:**
```bash
make release VERSION=v1.0.0
# Wait for GitHub Actions to complete
make release VERSION=v1.0.1
```
