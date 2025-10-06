# Version Information

## Checking Version

```bash
# Detailed version info
tokenscout version

# Short version
tokenscout --version
tokenscout -v
```

## Version Command Output

```
TokenScout v1.0.0
  Commit:      abc1234
  Built:       2025-10-06T16:01:39Z
  Go version:  go1.21.0
  OS/Arch:     darwin/arm64
  VCS dirty:   yes
```

**Fields:**
- **Version**: Release version (v1.0.0) or "dev" for development builds
- **Commit**: Git commit SHA (short format)
- **Built**: UTC timestamp when binary was built
- **Go version**: Version of Go used to build
- **OS/Arch**: Target operating system and architecture
- **VCS dirty**: Shown if built from modified sources (uncommitted changes)

## How Version Info is Set

### During Build

The Makefile automatically injects version information via ldflags:

```bash
make build
# Sets:
# - version: from VERSION variable (default v0.1.0)
# - commit: current git commit SHA
# - date: current UTC timestamp
```

### During Release

GoReleaser automatically sets version from git tag:

```bash
make release VERSION=v1.0.0
# Creates tag v1.0.0
# GoReleaser builds with:
# - version: v1.0.0 (from tag)
# - commit: git commit SHA
# - date: release timestamp
```

## Custom Version Build

```bash
# Build with custom version
make build VERSION=v1.2.3

# Or directly with go build
go build -ldflags "-X main.version=v1.2.3 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" -o tokenscout .
```

## Development Builds

When built without ldflags (e.g., `go run .`):

```
TokenScout dev
  Commit:      none
  Built:       unknown
  Go version:  go1.21.0
  OS/Arch:     darwin/arm64
```

## VCS Information

Go 1.18+ includes VCS (Version Control System) information automatically:

- **vcs.revision**: Git commit hash
- **vcs.time**: Commit timestamp  
- **vcs.modified**: Whether source had uncommitted changes

This info is embedded in the binary and shown by `version` command.

## Version Flags

```bash
# Long form
./tokenscout --version

# Short form
./tokenscout -v

# Detailed info
./tokenscout version
```

All three work the same way:
- `--version` and `-v`: Show version string only
- `version` command: Show detailed build info

## Implementation

Version information is injected at compile time:

**main.go:**
```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

**Makefile:**
```makefile
LDFLAGS := -X main.version=$(VERSION)
LDFLAGS += -X main.commit=$(shell git rev-parse --short HEAD)
LDFLAGS += -X main.date=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "$(LDFLAGS)" -o tokenscout .
```

These ldflags replace the default values at build time.

## Release vs Development

**Development build:**
```
go run .
# version = "dev"
```

**Make build:**
```
make build
# version = "v0.1.0" (from Makefile)
```

**Release build:**
```
make release VERSION=v1.0.0
# version = "v1.0.0" (from tag)
```

## Checking in Scripts

```bash
# Get version programmatically
VERSION=$(./tokenscout --version)
echo "Running TokenScout $VERSION"

# Check if version is dev
if ./tokenscout --version | grep -q "dev"; then
    echo "Warning: Using development build"
fi
```
