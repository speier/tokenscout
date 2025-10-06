# Quick Release Guide

Ultra-simple guide to creating a release.

## The One-Command Release

```bash
make release
```

That's it! This will:
1. Auto-increment patch version (0.1.0 → 0.1.1)
2. Run tests (validates code)
3. Build locally (ensures it compiles)
4. Commit VERSION file
5. Create git tag (e.g., `v0.1.1`)
6. Push to GitHub
7. GitHub Actions builds for all platforms
8. Creates GitHub Release with binaries

## Step-by-Step Example

```bash
# 1. Make sure all changes are committed
git status
# Should show: "nothing to commit, working tree clean"

# 2. Create release (auto-increments version)
make release

# Output:
# Bumping patch version...
# Version bumped: 0.1.1
# ... runs tests ...
# ... builds binary ...
# Creating release v0.1.1
# ... git commit -m "chore: bump version to v0.1.1" ...
# ... git tag -a v0.1.1 -m "Release v0.1.1" ...
# ... git push origin main ...
# ... git push origin v0.1.1 ...
# ✓ Release v0.1.1 pushed. GitHub Actions will build and publish.

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

## Version Control

Version is stored in the `VERSION` file and auto-increments:

```bash
# Default: auto-increment patch version (0.1.0 → 0.1.1)
make release

# Manual version bumps for bigger changes:
make bump-minor    # 0.1.0 → 0.2.0 (new features)
make bump-major    # 0.1.0 → 1.0.0 (breaking changes)
make release-manual

# Or bump then release:
make bump-patch    # 0.1.0 → 0.1.1
make release-manual

# Or edit VERSION file directly
echo "1.0.0" > VERSION
make release-manual
```

**Semantic Versioning:**
- **Patch** (0.0.1): Bug fixes, no new features
- **Minor** (0.1.0): New features, backward compatible
- **Major** (1.0.0): Breaking changes

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
