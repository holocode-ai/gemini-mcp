# Build stage
FROM golang:1.23-alpine AS builder

# Install git (needed for go modules with private repos)
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser for security
RUN adduser -D -g '' -u 10001 appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
ARG TARGETARCH=amd64

# Build the application for target architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.version=${VERSION} -X 'main.buildTime=${BUILD_TIME}' -X main.gitCommit=${GIT_COMMIT}" \
    -o gemini-mcp main.go

# Final stage - use distroless for smaller image with basic utilities
FROM gcr.io/distroless/static:nonroot

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/gemini-mcp /app/gemini-mcp

# Set working directory
WORKDIR /app

# Environment variables
ENV OUTPUT_DIR=/app/output
ENV TRANSPORT=stdio
ENV PORT=8080

# Expose HTTP port (for HTTP/SSE transport mode)
EXPOSE 8080

# Default command
ENTRYPOINT ["/app/gemini-mcp"]