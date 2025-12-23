.PHONY: build build-all build-gemini-mcp build-upload-media clean run test install deps release docker help

# Variables
GEMINI_MCP_BINARY=gemini-mcp
UPLOAD_MEDIA_BINARY=upload_media
BUILD_DIR=build
OUTPUT_DIR=output

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -X 'main.buildTime=${BUILD_TIME}' -X main.gitCommit=${GIT_COMMIT} -s -w"

# Build both binaries (default target)
build: build-gemini-mcp build-upload-media
	@echo "Build complete: $(BUILD_DIR)/$(GEMINI_MCP_BINARY), $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)"

# Build gemini-mcp only
build-gemini-mcp:
	@echo "Building $(GEMINI_MCP_BINARY) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GEMINI_MCP_BINARY) main.go

# Build upload_media CLI only
build-upload-media:
	@echo "Building $(UPLOAD_MEDIA_BINARY) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY) ./cmd/upload_media

# Build for macOS ARM (Apple Silicon)
build-darwin-arm64: build-gemini-mcp-darwin-arm64 build-upload-media-darwin-arm64

build-gemini-mcp-darwin-arm64:
	@echo "Building $(GEMINI_MCP_BINARY) v$(VERSION) for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GEMINI_MCP_BINARY)-darwin-arm64 main.go

build-upload-media-darwin-arm64:
	@echo "Building $(UPLOAD_MEDIA_BINARY) v$(VERSION) for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)-darwin-arm64 ./cmd/upload_media

# Build for macOS Intel
build-darwin-amd64: build-gemini-mcp-darwin-amd64 build-upload-media-darwin-amd64

build-gemini-mcp-darwin-amd64:
	@echo "Building $(GEMINI_MCP_BINARY) v$(VERSION) for macOS Intel..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GEMINI_MCP_BINARY)-darwin-amd64 main.go

build-upload-media-darwin-amd64:
	@echo "Building $(UPLOAD_MEDIA_BINARY) v$(VERSION) for macOS Intel..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)-darwin-amd64 ./cmd/upload_media

# Build for Linux x86_64
build-linux-amd64: build-gemini-mcp-linux-amd64 build-upload-media-linux-amd64

build-gemini-mcp-linux-amd64:
	@echo "Building $(GEMINI_MCP_BINARY) v$(VERSION) for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GEMINI_MCP_BINARY)-linux-amd64 main.go

build-upload-media-linux-amd64:
	@echo "Building $(UPLOAD_MEDIA_BINARY) v$(VERSION) for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)-linux-amd64 ./cmd/upload_media

# Build for Linux ARM64
build-linux-arm64: build-gemini-mcp-linux-arm64 build-upload-media-linux-arm64

build-gemini-mcp-linux-arm64:
	@echo "Building $(GEMINI_MCP_BINARY) v$(VERSION) for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GEMINI_MCP_BINARY)-linux-arm64 main.go

build-upload-media-linux-arm64:
	@echo "Building $(UPLOAD_MEDIA_BINARY) v$(VERSION) for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)-linux-arm64 ./cmd/upload_media

# Build all platforms
build-all: build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64
	@echo "All platform builds complete!"

# Build for multiple platforms (with version suffix, for releases)
release:
	@echo "Building v$(VERSION) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			echo "Building $(GEMINI_MCP_BINARY) for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) \
				-o $(BUILD_DIR)/$(GEMINI_MCP_BINARY)-$(VERSION)-$$os-$$arch$$ext main.go; \
			echo "Building $(UPLOAD_MEDIA_BINARY) for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) \
				-o $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY)-$(VERSION)-$$os-$$arch$$ext ./cmd/upload_media; \
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
run: build-gemini-mcp
	@echo "Running $(GEMINI_MCP_BINARY) in stdio mode..."
	@mkdir -p $(OUTPUT_DIR)
	$(BUILD_DIR)/$(GEMINI_MCP_BINARY)

# Run the application in HTTP mode
run-http: build-gemini-mcp
	@echo "Running $(GEMINI_MCP_BINARY) in HTTP mode..."
	@mkdir -p $(OUTPUT_DIR)
	TRANSPORT=http PORT=8080 $(BUILD_DIR)/$(GEMINI_MCP_BINARY)

# Run the application in SSE mode (alias for run-http)
run-sse: run-http

# Test the application
test:
	@echo "Running tests..."
	go test -v ./...

# Install both binaries to GOPATH/bin
install: build
	@echo "Installing binaries..."
	cp $(BUILD_DIR)/$(GEMINI_MCP_BINARY) $(GOPATH)/bin/
	cp $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY) $(GOPATH)/bin/

# Install gemini-mcp only
install-gemini-mcp: build-gemini-mcp
	@echo "Installing $(GEMINI_MCP_BINARY)..."
	cp $(BUILD_DIR)/$(GEMINI_MCP_BINARY) $(GOPATH)/bin/

# Install upload_media only
install-upload-media: build-upload-media
	@echo "Installing $(UPLOAD_MEDIA_BINARY)..."
	cp $(BUILD_DIR)/$(UPLOAD_MEDIA_BINARY) $(GOPATH)/bin/

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
	@echo "Building Docker image $(GEMINI_MCP_BINARY):$(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(GEMINI_MCP_BINARY):$(VERSION) \
		-t $(GEMINI_MCP_BINARY):latest .

docker-multiarch:
	@echo "Building multi-architecture Docker image $(GEMINI_MCP_BINARY):$(VERSION)..."
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(GEMINI_MCP_BINARY):$(VERSION) \
		-t $(GEMINI_MCP_BINARY):latest .

docker-run:
	@echo "Running Docker container in stdio mode..."
	docker run --rm -it \
		-e GOOGLE_API_KEY \
		-v $(PWD)/output:/app/output \
		$(GEMINI_MCP_BINARY):latest

docker-run-http:
	@echo "Running Docker container in HTTP mode..."
	docker run --rm -it \
		-e GOOGLE_API_KEY \
		-e TRANSPORT=http \
		-e PORT=8080 \
		-e SERVICE_TOKENS \
		-p 8080:8080 \
		-v $(PWD)/output:/app/output \
		$(GEMINI_MCP_BINARY):latest

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
	@echo "  build                    - Build both gemini-mcp and upload_media"
	@echo "  build-gemini-mcp         - Build gemini-mcp only"
	@echo "  build-upload-media       - Build upload_media CLI only"
	@echo "  build-darwin-arm64       - Build both for macOS ARM (Apple Silicon)"
	@echo "  build-darwin-amd64       - Build both for macOS Intel"
	@echo "  build-linux-amd64        - Build both for Linux x86_64"
	@echo "  build-linux-arm64        - Build both for Linux ARM64"
	@echo "  build-all                - Build both for all platforms"
	@echo "  release                  - Build for release (with version suffix)"
	@echo ""
	@echo "Run:"
	@echo "  run                      - Run gemini-mcp in stdio mode"
	@echo "  run-http                 - Run gemini-mcp in HTTP mode (port 8080)"
	@echo "  run-sse                  - Alias for run-http"
	@echo ""
	@echo "Docker:"
	@echo "  docker                   - Build Docker image"
	@echo "  docker-multiarch         - Build multi-arch Docker image (amd64/arm64)"
	@echo "  docker-run               - Run Docker container in stdio mode"
	@echo "  docker-run-http          - Run Docker container in HTTP mode"
	@echo ""
	@echo "Development:"
	@echo "  deps                     - Install dependencies"
	@echo "  test                     - Run tests"
	@echo "  fmt                      - Format code"
	@echo "  lint                     - Lint code"
	@echo "  install                  - Install both binaries to GOPATH/bin"
	@echo "  install-gemini-mcp       - Install gemini-mcp to GOPATH/bin"
	@echo "  install-upload-media     - Install upload_media to GOPATH/bin"
	@echo "  clean                    - Clean build artifacts"
	@echo ""
	@echo "Info:"
	@echo "  version                  - Show version information"
	@echo "  help                     - Show this help"
