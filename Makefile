.PHONY: build run test clean init release

# Version for releases (override with: make release VERSION=v1.0.0)
VERSION ?= v0.1.0

# Build flags
LDFLAGS := -X main.version=$(VERSION)
LDFLAGS += -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS += -X main.date=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o tokenscout .

# Run in development
run:
	go run . start --dry-run

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f tokenscout tokenscout-*
	rm -f *.db *.db-shm *.db-wal

# Initialize config
init:
	go run . init

# Run in development with debug logs
dev:
	go run . start --dry-run --log-level debug

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o tokenscout-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o tokenscout-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o tokenscout-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o tokenscout-windows-amd64.exe .

# Create and push a release tag (validates first)
release: test build
	@echo "Creating release $(VERSION)"
	@echo "Validation passed ✓"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✓ Release $(VERSION) pushed. GitHub Actions will build and publish."

# Test release process locally (doesn't push)
release-test:
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed. Install: brew install goreleaser"; exit 1; }
	goreleaser release --snapshot --clean
	@echo "✓ Test release complete. Check ./dist/ folder"

# Check GoReleaser config
release-check:
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed. Install: brew install goreleaser"; exit 1; }
	goreleaser check
