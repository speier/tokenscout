# Release Process

## Automated Releases

Releases are fully automated via GitHub Actions and GoReleaser.

## Creating a Release

1. **Update version in code** (optional - GoReleaser uses git tags)
2. **Commit and push changes**
   ```bash
   git add .
   git commit -m "chore: prepare for release v1.0.0"
   git push
   ```

3. **Create and push a tag**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

4. **GitHub Actions automatically:**
   - Builds for Linux, macOS (Intel + ARM), Windows
   - Creates release archives (.tar.gz, .zip)
   - Generates checksums
   - Creates GitHub Release with changelog
   - Uploads all binaries

## What Gets Built

### Platforms
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Archive Contents
Each release archive includes:
- `tokenscout` binary
- `README.md`
- `LICENSE`
- `docs/` folder
- `config.example.yaml`

## Release Workflow

```
1. Tag pushed (v*)
   ↓
2. GitHub Actions triggered
   ↓
3. GoReleaser builds all platforms
   ↓
4. Archives created
   ↓
5. Checksums generated
   ↓
6. GitHub Release created
   ↓
7. Binaries uploaded
   ↓
8. Done! 🎉
```

## Version Numbering

Use semantic versioning: `vMAJOR.MINOR.PATCH`

- `v1.0.0` - Initial release
- `v1.1.0` - New features
- `v1.0.1` - Bug fixes
- `v2.0.0` - Breaking changes

## Pre-releases

For beta/alpha releases, add suffix:
```bash
git tag -a v1.0.0-beta.1 -m "Beta release"
git push origin v1.0.0-beta.1
```

GoReleaser automatically marks these as pre-releases.

## Testing Releases Locally

```bash
# Test GoReleaser config
goreleaser check

# Build without releasing (snapshot)
goreleaser build --snapshot --clean

# Test full release locally
goreleaser release --snapshot --clean
```

## CI/CD Workflows

### `.github/workflows/ci.yml`
- Runs on every push/PR
- Tests and builds
- Ensures code compiles

### `.github/workflows/release.yml`
- Runs on tag push (v*)
- Creates actual GitHub Release
- Uploads binaries

### `.github/workflows/test-release.yml`
- Runs on every push/PR
- Tests GoReleaser config
- Creates snapshot builds (not published)

## Troubleshooting

**Release failed?**
- Check GitHub Actions logs
- Verify `.goreleaser.yml` syntax: `goreleaser check`
- Ensure tag follows `v*` pattern

**Binary not working?**
- Check build logs in GitHub Actions
- Test locally with `goreleaser build --snapshot`

**Missing platforms?**
- Update `goos` and `goarch` in `.goreleaser.yml`
