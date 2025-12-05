# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Model Context Protocol (MCP) server for Google Gemini AI services. It provides tools for image generation and text generation using the Google Gemini API, with a direct MCP protocol implementation.

## Development Commands

### Building and Running
- `make build` - Build the application to `build/gemini-mcp`
- `go build -o gemini-mcp main.go` - Direct build command
- `./gemini-mcp -version` - Show version information
- `./gemini-mcp` - Run in stdio mode (default MCP transport)
- `./test_mcp.sh` - Run MCP protocol tests

### Testing and Quality
- `make test` - Run all tests
- `make fmt` - Format Go code
- `make clean` - Clean build artifacts
- `go mod tidy` - Clean up dependencies

### Environment Setup
- Required: `GOOGLE_API_KEY` environment variable for Gemini API access
- Optional: `GOOGLE_PROJECT_ID`, `GOOGLE_LOCATION`, `OUTPUT_DIR`, `TRANSPORT`
- Copy `.env.example` to `.env` and configure environment variables

## Architecture

### Project Structure
- `main.go` - Main application with MCP server implementation
- `internal/common/config.go` - Configuration management
- `pkg/types/mcp.go` - MCP protocol type definitions
- `require.md` - Project requirements documentation
- `test_mcp.sh` - MCP protocol test script

### Key Components

1. **MCP Server**: Direct JSON-RPC protocol implementation in main.go
2. **Transport**: Currently supports stdio transport for MCP client integration
3. **Configuration**: Environment-based configuration with validation
4. **Tools**: Integrated tool handlers for Gemini services

### MCP Protocol Implementation
- Follows MCP 2024-11-05 specification
- Implements initialize, tools/list, and tools/call methods
- Direct JSON-RPC message handling over stdin/stdout
- Tools generate files to configurable output directory

### Current Transport Support
- **stdio**: JSON-RPC over stdin/stdout (implemented)
- **HTTP/SSE**: Planned for future implementation

## AI Service Integration

Uses `google.golang.org/genai` library with Gemini API backend:

### Currently Implemented
- **gemini_image_generation**: Image generation using Gemini 3.0 Pro models
- **gemini_image_edit**: Image editing using Gemini AI models
- **gemini_multi_image**: Multi-image composition using Gemini AI models
- **veo_text_to_video**: Video generation from text using Veo 3.0 models
- **veo_image_to_video**: Video generation from images using Veo 3.0 models

### Implementation Details
- Uses `client.Models.GenerateContent()` for image generation and editing
- Uses `client.Models.GenerateVideos()` for video generation
- Supports file output to local directories
- Proper error handling and response formatting

### Environment Configuration
- `GOOGLE_API_KEY`: Required Gemini API key
- `GOOGLE_PROJECT_ID`: Optional project ID
- `GOOGLE_LOCATION`: Region (default: us-central1)
- `OUTPUT_DIR`: Local output directory (default: ./output)
- `TRANSPORT`: Transport protocol (default: stdio)

## Development Notes

- Built for Gemini API (not Vertex AI)
- Architecture inspired by mcp-genmedia project but simplified for Gemini API
- Direct MCP protocol implementation without external libraries
- Focus on reliable stdio transport for MCP client compatibility