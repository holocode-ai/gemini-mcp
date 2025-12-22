# Gemini MCP Server

A comprehensive Model Context Protocol (MCP) server for Google Gemini AI services, providing advanced multimodal generation capabilities including image generation, image editing, and video creation through Google's state-of-the-art AI models.

## 🚀 Features

### **Multimodal AI Services**
- **🖼️ Image Generation**: High-quality image creation using Gemini 3.0 Pro models
- **✏️ Image Editing**: Advanced image modification and enhancement using Gemini AI models
- **🔀 Multi-Image Composition**: Seamless blending and combining of multiple images
- **🎬 Video Generation**: Cinematic video creation using Google's Veo 3.1 models with native audio (text-to-video and image-to-video)

### **Advanced Model Support**
- **Gemini Models**: `gemini-3-pro-image-preview` (default - Gemini 3 Pro with native image generation), `gemini-2.5-flash-image`
- **Veo Models**: `veo-3.1-generate-preview` (default - latest with native audio), `veo-3.1-fast-generate-preview`, `veo-3.0-generate-preview`, `veo-3.0-fast-generate-001`

### **MCP Protocol Features**
- **Stdio Transport**: Direct integration with MCP clients
- **Comprehensive Tool Descriptions**: Detailed parameter documentation and usage examples
- **File Output Management**: Configurable output directories with metadata
- **Error Handling**: Robust error handling with informative responses

## 📋 Prerequisites

### For Pre-built Binaries
- **Google API Key** with Gemini API access (required)
- **Optional**: Google Cloud Project ID for advanced features

### For Building from Source
- **Go 1.23+** (required for building)
- **Google API Key** with Gemini API access (required)
- **Optional**: Google Cloud Project ID for advanced features

## 🛠️ Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/YOUR_USERNAME/gemini-mcp/releases):

**Linux (x86_64)**
```bash
# Download and extract
wget https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-linux-amd64.tar.gz
tar -xzf gemini-mcp-VERSION-linux-amd64.tar.gz

# Make executable and move to PATH
chmod +x gemini-mcp-VERSION-linux-amd64
sudo mv gemini-mcp-VERSION-linux-amd64 /usr/local/bin/gemini-mcp
```

**Linux (ARM64)**
```bash
wget https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-linux-arm64.tar.gz
tar -xzf gemini-mcp-VERSION-linux-arm64.tar.gz
chmod +x gemini-mcp-VERSION-linux-arm64
sudo mv gemini-mcp-VERSION-linux-arm64 /usr/local/bin/gemini-mcp
```

**macOS (Intel)**
```bash
# Download and extract
curl -LO https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-darwin-amd64.tar.gz
tar -xzf gemini-mcp-VERSION-darwin-amd64.tar.gz

# Make executable and move to PATH
chmod +x gemini-mcp-VERSION-darwin-amd64
sudo mv gemini-mcp-VERSION-darwin-amd64 /usr/local/bin/gemini-mcp
```

**macOS (Apple Silicon)**
```bash
curl -LO https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-darwin-arm64.tar.gz
tar -xzf gemini-mcp-VERSION-darwin-arm64.tar.gz
chmod +x gemini-mcp-VERSION-darwin-arm64
sudo mv gemini-mcp-VERSION-darwin-arm64 /usr/local/bin/gemini-mcp
```

**Windows**
```powershell
# Download the zip file from releases page
# Extract gemini-mcp-VERSION-windows-amd64.zip
# Add the extracted .exe to your PATH
```

**Verify installation:**
```bash
gemini-mcp -version
```

### Option 2: Build from Source

1. **Clone and build**:
```bash
git clone <repository-url>
cd gemini-mcp
go build -o gemini-mcp main.go
```

2. **Set up API key**:
```bash
export GOOGLE_API_KEY="your_google_api_key_here"
```

3. **Test the installation**:
```bash
./gemini-mcp -version
```

### Option 3: Using Makefile

1. **Install dependencies**:
```bash
make deps
```

2. **Build application**:
```bash
make build
```

3. **Set up environment**:
```bash
cp .env.example .env
# Edit .env with your API key
```

## 🎯 Usage

### Command Line Interface

```bash
./gemini-mcp [options]

Options:
  -transport string    Transport type: stdio (default)
  -version            Show version information
```

### Stdio Mode (MCP Integration)

Run the server for direct MCP client integration:
```bash
./gemini-mcp
```

### Testing MCP Protocol

```bash
# Test basic connectivity
./test_mcp.sh

# Manual testing
echo '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' | ./gemini-mcp
```

## 🛠️ Available Tools

### 1. **gemini_image_generation**
Generate high-quality images using Google's latest Gemini image generation models with advanced style control and quality settings.

**Key Features:**
- Advanced style control and artistic options
- Multi-language prompt support
- Customizable aspect ratios and quality settings
- Content safety levels and text rendering options

**Parameters:**
- `prompt` (required): Detailed description of desired image
- `model`: Gemini model variant (default: `gemini-3-pro-preview`)
- `output_directory`: Local save path

### 2. **gemini_image_edit**
Edit existing images using Google's Gemini AI models with targeted modifications.

**Key Features:**
- Targeted image modifications and style transfers
- Object addition/removal capabilities
- Background changes while preserving original characteristics
- Precise control over edit types

**Parameters:**
- `prompt` (required): Description of desired edits
- `image_path`: Path to the image to edit
- `edit_type`: Type of edit operation
- `output_directory`: Local save path

### 3. **gemini_multi_image**
Combine and blend multiple images using Google's Gemini AI models.

**Key Features:**
- Merge 2-3 images into cohesive compositions
- Create collages, overlays, and seamless blends
- Character consistency across scenes
- Style unification for creative compositions

**Parameters:**
- `prompt` (required): Description of desired composition
- `image_paths`: Array of image paths to combine
- `blend_mode`: How to combine the images
- `output_directory`: Local save path

### 4. **veo_text_to_video**
Generate 4-8 second videos from text prompts using Google's Veo 3.1 models with native audio.

**Key Features:**
- Detailed scene descriptions with camera movements
- Realistic physics and natural motion
- Native audio generation (dialogue, sounds, music)
- Support for 16:9/9:16 aspect ratios
- 720p/1080p resolution options
- Flexible duration: 4, 6, or 8 seconds
- SynthID watermarking

**Parameters:**
- `prompt` (required): Detailed video scene description
- `negative_prompt`: Content to avoid in the video
- `aspect_ratio`: Video ratio (`16:9`, `9:16`)
- `resolution`: Video quality (`720p`, `1080p`)
- `model`: Veo variant (default: `veo-3.1-generate-preview`)
- `seed`: Optional seed for reproducibility
- `output_directory`: Local save path

### 6. **veo_image_to_video**
Animate static images into 4-8 second videos using Google's Veo 3.1 models with native audio.

**Key Features:**
- Transform photos into dynamic scenes
- Natural motion and camera movements
- Input image becomes the starting frame
- Realistic physics simulation

**Parameters:**
- `prompt` (required): Description of desired animation
- `image_path`: Path to input image
- `negative_prompt`: Content to avoid
- `aspect_ratio`: Video ratio (`16:9`, `9:16`)
- `resolution`: Video quality (`720p`, `1080p`)
- `model`: Veo variant (default: `veo-3.1-generate-preview`)
- `output_directory`: Local save path

### 7. **veo_generate_video** (Legacy)
General video generation tool supporting both text-to-video and image-to-video creation.

**Key Features:**
- Backward compatibility with existing workflows
- Supports both text and image inputs
- Advanced scene composition
- Automatic operation polling

**Parameters:**
- `prompt` (required): Video description
- `image_path`: Optional input image for image-to-video
- `aspect_ratio`: Video ratio
- `resolution`: Video quality
- `negative_prompt`: Content exclusion
- `output_directory`: Local save path

## 🔧 Environment Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GOOGLE_API_KEY` | Gemini API authentication key | - | ✅ Yes |
| `GOOGLE_PROJECT_ID` | Google Cloud Project ID | - | ❌ Optional |
| `GOOGLE_LOCATION` | Google Cloud region | `us-central1` | ❌ Optional |
| `OUTPUT_DIR` | File output directory | `./output` | ❌ Optional |
| `TRANSPORT` | MCP transport protocol | `stdio` | ❌ Optional |

## 🔌 MCP Client Integration

### Claude Desktop Configuration
```json
{
  "mcpServers": {
    "gemini": {
      "command": "gemini-mcp",
      "env": {
        "GOOGLE_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

**Note:** If you installed the binary to a custom location, use the full path:
```json
{
  "mcpServers": {
    "gemini": {
      "command": "/path/to/gemini-mcp",
      "env": {
        "GOOGLE_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

### Cline VSCode Extension
```json
{
  "cline.mcp.servers": [
    {
      "name": "gemini",
      "command": "gemini-mcp",
      "env": {
        "GOOGLE_API_KEY": "your_api_key_here"
      }
    }
  ]
}
```

## 🧪 Development

### Building from Source
```bash
go mod tidy
go build -o gemini-mcp main.go
```

### Testing
```bash
make test
./test_mcp.sh
```

### Code Quality
```bash
make fmt    # Format code
make clean  # Clean artifacts
```

## 📝 Implementation Notes

- **Gemini Integration**: Uses `google.golang.org/genai` with Gemini API backend
- **Protocol Compliance**: Implements MCP 2024-11-05 specification
- **Image Generation**: Full implementation with Gemini 3.0 Pro models
- **Video Generation**: Complete Veo 3.1 integration with native audio, operation polling, and proper file downloads
- **File Management**: Generated content saved with metadata and timestamps
- **Error Handling**: Comprehensive error responses with helpful messages
- **Multi-modal Support**: Supports text-to-image, image-to-image, text-to-video, and image-to-video workflows

## 🤝 Contributing

This project is designed to be a comprehensive MCP server for Google's AI services. Contributions are welcome for:
- Additional model support
- Transport protocol enhancements
- Full implementation of placeholder services
- Documentation improvements

## 📄 License

MIT License - see LICENSE file for details.