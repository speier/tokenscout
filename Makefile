.PHONY: build run test clean init

# Build the binary
build:
	go build -o tokenscout .

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
	GOOS=linux GOARCH=amd64 go build -o tokenscout-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o tokenscout-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o tokenscout-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o tokenscout-windows-amd64.exe .
