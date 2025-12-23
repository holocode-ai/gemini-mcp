.PHONY: build build-all build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64 clean run test install deps release docker

# Variables
BINARY_NAME=gemini-mcp
BUILD_DIR=build
OUTPUT_DIR=output

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -X 'main.buildTime=${BUILD_TIME}' -X main.gitCommit=${GIT_COMMIT} -s -w"

# Build the application (current platform)
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Build for macOS ARM (Apple Silicon)
build-darwin-arm64:
	@echo "Building $(BINARY_NAME) v$(VERSION) for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go

# Build for macOS Intel
build-darwin-amd64:
	@echo "Building $(BINARY_NAME) v$(VERSION) for macOS Intel..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go

# Build for Linux x86_64
build-linux-amd64:
	@echo "Building $(BINARY_NAME) v$(VERSION) for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go

# Build for Linux ARM64
build-linux-arm64:
	@echo "Building $(BINARY_NAME) v$(VERSION) for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go

# Build all platforms
build-all: build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64
	@echo "All platform builds complete!"

# Build for multiple platforms (with version suffix, for releases)
release:
	@echo "Building $(BINARY_NAME) v$(VERSION) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) \
				-o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-$$os-$$arch$$ext main.go; \
		done; \
	done

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)

# Run the application in stdio mode
run: build
	@echo "Running $(BINARY_NAME) in stdio mode..."
	@mkdir -p $(OUTPUT_DIR)
	$(BUILD_DIR)/$(BINARY_NAME)

# Run the application in HTTP mode
run-http: build
	@echo "Running $(BINARY_NAME) in HTTP mode..."
	@mkdir -p $(OUTPUT_DIR)
	TRANSPORT=http PORT=8080 $(BUILD_DIR)/$(BINARY_NAME)

# Run the application in SSE mode (alias for run-http)
run-sse: run-http

# Test the application
test:
	@echo "Running tests..."
	go test -v ./...

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Docker targets
docker:
	@echo "Building Docker image $(BINARY_NAME):$(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest .

docker-multiarch:
	@echo "Building multi-architecture Docker image $(BINARY_NAME):$(VERSION)..."
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest .

docker-run:
	@echo "Running Docker container in stdio mode..."
	docker run --rm -it \
		-e GOOGLE_API_KEY \
		-v $(PWD)/output:/app/output \
		$(BINARY_NAME):latest

docker-run-http:
	@echo "Running Docker container in HTTP mode..."
	docker run --rm -it \
		-e GOOGLE_API_KEY \
		-e TRANSPORT=http \
		-e PORT=8080 \
		-e SERVICE_TOKENS \
		-p 8080:8080 \
		-v $(PWD)/output:/app/output \
		$(BINARY_NAME):latest

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build              - Build for current platform"
	@echo "  build-darwin-arm64 - Build for macOS ARM (Apple Silicon)"
	@echo "  build-darwin-amd64 - Build for macOS Intel"
	@echo "  build-linux-amd64  - Build for Linux x86_64"
	@echo "  build-linux-arm64  - Build for Linux ARM64"
	@echo "  build-all          - Build all platforms"
	@echo "  release            - Build for release (with version suffix)"
	@echo ""
	@echo "Run:"
	@echo "  run                - Run in stdio mode"
	@echo "  run-http           - Run in HTTP mode (port 8080)"
	@echo "  run-sse            - Alias for run-http"
	@echo ""
	@echo "Docker:"
	@echo "  docker             - Build Docker image"
	@echo "  docker-multiarch   - Build multi-arch Docker image (amd64/arm64)"
	@echo "  docker-run         - Run Docker container in stdio mode"
	@echo "  docker-run-http    - Run Docker container in HTTP mode"
	@echo ""
	@echo "Development:"
	@echo "  deps               - Install dependencies"
	@echo "  test               - Run tests"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  install            - Install binary to GOPATH/bin"
	@echo "  clean              - Clean build artifacts"
	@echo ""
	@echo "Info:"
	@echo "  version            - Show version information"
	@echo "  help               - Show this help"