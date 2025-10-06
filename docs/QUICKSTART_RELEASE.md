# Quick Release Guide

Ultra-simple guide to creating a release.

## The One-Command Release

```bash
make release VERSION=v1.0.0
```

That's it! This will:
1. Run tests (validates code)
2. Build locally (ensures it compiles)
3. Create git tag `v1.0.0`
4. Push tag to GitHub
5. GitHub Actions builds for all platforms
6. Creates GitHub Release with binaries

## Step-by-Step Example

```bash
# 1. Make sure all changes are committed
git status
# Should show: "nothing to commit, working tree clean"

# 2. Create release
make release VERSION=v1.0.0

# Output:
# Creating release v1.0.0
# ... git tag -a v1.0.0 -m "Release v1.0.0" ...
# ... git push origin v1.0.0 ...
# ✓ Release v1.0.0 pushed. GitHub Actions will build and publish.

# 3. Watch GitHub Actions (optional)
# Visit: https://github.com/speier/tokenscout/actions

# 4. Download binaries (after ~2-3 minutes)
# Visit: https://github.com/speier/tokenscout/releases
```

## Before Your First Release

```bash
# Test locally first
make release-test

# This builds everything without pushing to GitHub
# Check ./dist/ folder to see what would be released
ls -lh dist/
```

## Version Numbers

Follow semantic versioning:

```bash
# Initial release
make release VERSION=v1.0.0

# Bug fix
make release VERSION=v1.0.1

# New feature
make release VERSION=v1.1.0

# Breaking change
make release VERSION=v2.0.0

# Pre-release
make release VERSION=v1.0.0-beta.1
```

## What Gets Released

Each release includes:
- Linux (amd64, arm64)
- macOS Intel (amd64)
- macOS Apple Silicon (arm64)
- Windows (amd64)

Plus:
- Checksums file
- Auto-generated changelog

## Troubleshooting

**Error: tag already exists**
```bash
# Delete local and remote tag
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Try again
make release VERSION=v1.0.0
```

**Need to undo a release?**
```bash
# Delete from GitHub (via web UI)
# Then delete tag locally
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

## Advanced: Dry Run

Want to see what would happen without pushing?

```bash
# Shows commands without executing
make -n release VERSION=v1.0.0
```

## CI/CD Workflow

```
Your Machine                 GitHub
     │                          │
     │  make release            │
     │  ─────────────────────>  │
     │                          │
     │                          │  GitHub Actions
     │                          │  triggered
     │                          │     │
     │                          │     ├─ Build Linux
     │                          │     ├─ Build macOS
     │                          │     ├─ Build Windows
     │                          │     ├─ Create archives
     │                          │     ├─ Generate changelog
     │                          │     └─ Create release
     │                          │
     │  Download binaries       │
     │  <─────────────────────  │
```

Done! 🚀
