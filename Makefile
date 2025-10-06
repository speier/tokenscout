.PHONY: build run test clean init release bump-patch bump-minor bump-major

# Read version from VERSION file (can be overridden: make build VERSION=v1.2.3)
VERSION ?= v$(shell cat VERSION 2>/dev/null || echo "0.0.0")

# Build the binary
build:
	@COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	echo "Building $(VERSION) (commit: $$COMMIT)..."; \
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$$COMMIT -X main.date=$$DATE" -o tokenscout .

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
	@COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	LDFLAGS="-X main.version=$(VERSION) -X main.commit=$$COMMIT -X main.date=$$DATE"; \
	GOOS=linux GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o tokenscout-linux-amd64 .; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o tokenscout-darwin-amd64 .; \
	GOOS=darwin GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o tokenscout-darwin-arm64 .; \
	GOOS=windows GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o tokenscout-windows-amd64.exe .

# Bump version (patch by default, use: make bump-minor or make bump-major)
bump-patch:
	@echo "Bumping patch version..."
	@current=$$(cat VERSION); \
	major=$$(echo $$current | cut -d. -f1); \
	minor=$$(echo $$current | cut -d. -f2); \
	patch=$$(echo $$current | cut -d. -f3); \
	patch=$$((patch + 1)); \
	echo "$$major.$$minor.$$patch" > VERSION
	@echo "Version bumped: $$(cat VERSION)"

bump-minor:
	@echo "Bumping minor version..."
	@current=$$(cat VERSION); \
	major=$$(echo $$current | cut -d. -f1); \
	minor=$$(echo $$current | cut -d. -f2); \
	minor=$$((minor + 1)); \
	echo "$$major.$$minor.0" > VERSION
	@echo "Version bumped: $$(cat VERSION)"

bump-major:
	@echo "Bumping major version..."
	@current=$$(cat VERSION); \
	major=$$(echo $$current | cut -d. -f1); \
	major=$$((major + 1)); \
	echo "$$major.0.0" > VERSION
	@echo "Version bumped: $$(cat VERSION)"

# Create and push a release tag (auto-increments patch version)
release: bump-patch
	@$(MAKE) test
	@NEW_VERSION=v$$(cat VERSION); \
	$(MAKE) build VERSION=$$NEW_VERSION; \
	echo "Creating release $$NEW_VERSION"; \
	echo "Validation passed ✓"; \
	git add VERSION; \
	git commit -m "chore: bump version to $$NEW_VERSION"; \
	git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
	git push origin main; \
	git push origin $$NEW_VERSION; \
	echo "✓ Release $$NEW_VERSION pushed. GitHub Actions will build and publish."

# Create release without auto-bump (manual version control)
release-manual: test build
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
