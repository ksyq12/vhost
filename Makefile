.PHONY: build test clean install lint

# Build variables
BINARY_NAME=vhost
BUILD_DIR=./build
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/vhost

# Build for multiple platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/vhost
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/vhost
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/vhost
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/vhost

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install to /usr/local/bin
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Development helpers
dev: build
	./$(BINARY_NAME)

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install to /usr/local/bin"
	@echo "  uninstall    - Uninstall from /usr/local/bin"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  tidy         - Tidy dependencies"
