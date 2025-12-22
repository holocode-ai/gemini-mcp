# Gemini MCP 服务器

适用于 Google Gemini AI 服务的综合模型上下文协议 (MCP) 服务器，通过 Google 最先进的 AI 模型提供高级多模态生成功能，包括图像生成、图像编辑和视频创作。

## 🚀 功能特性

### **多模态 AI 服务**
- **🖼️ 图像生成**：使用 Gemini 3.0 Pro 模型进行高质量图像创作
- **✏️ 图像编辑**：使用 Gemini AI 模型进行高级图像修改和增强
- **🔀 多图像合成**：无缝混合和组合多张图像
- **🎬 视频生成**：使用 Google 的 Veo 3.1 模型进行电影级视频创作，支持原生音频（文本生成视频和图像生成视频）

### **先进模型支持**
- **Gemini 模型**：`gemini-3-pro-image-preview`（默认 - Gemini 3 Pro 原生图像生成）、`gemini-2.5-flash-image`
- **Veo 模型**：`veo-3.1-generate-preview`（默认 - 最新版本，支持原生音频）、`veo-3.1-fast-generate-preview`、`veo-3.0-generate-preview`、`veo-3.0-fast-generate-001`

### **MCP 协议功能**
- **Stdio 传输**：直接与 MCP 客户端集成
- **全面的工具描述**：详细的参数文档和使用示例
- **文件输出管理**：可配置的输出目录和元数据
- **错误处理**：强大的错误处理机制，提供有用的响应信息

## 📋 先决条件

### 使用预编译二进制
- **Google API Key**，需要 Gemini API 访问权限（必需）
- **可选**：Google Cloud 项目 ID 用于高级功能

### 从源码构建
- **Go 1.23+**（构建必需）
- **Google API Key**，需要 Gemini API 访问权限（必需）
- **可选**：Google Cloud 项目 ID 用于高级功能

## 🛠️ 安装

### 方式一：下载预编译二进制（推荐）

从 [GitHub Releases](https://github.com/YOUR_USERNAME/gemini-mcp/releases) 下载适合您平台的最新版本：

**Linux (x86_64)**
```bash
# 下载并解压
wget https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-linux-amd64.tar.gz
tar -xzf gemini-mcp-VERSION-linux-amd64.tar.gz

# 添加执行权限并移动到 PATH
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
# 下载并解压
curl -LO https://github.com/YOUR_USERNAME/gemini-mcp/releases/latest/download/gemini-mcp-VERSION-darwin-amd64.tar.gz
tar -xzf gemini-mcp-VERSION-darwin-amd64.tar.gz

# 添加执行权限并移动到 PATH
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
# 从 releases 页面下载 zip 文件
# 解压 gemini-mcp-VERSION-windows-amd64.zip
# 将解压的 .exe 文件添加到系统 PATH
```

**验证安装：**
```bash
gemini-mcp -version
```

### 方式二：从源码构建

1. **克隆和构建**：
```bash
git clone <repository-url>
cd gemini-mcp
go build -o gemini-mcp main.go
```

2. **设置 API 密钥**：
```bash
export GOOGLE_API_KEY="your_google_api_key_here"
```

3. **测试安装**：
```bash
./gemini-mcp -version
```

### 方式三：使用 Makefile

1. **安装依赖**：
```bash
make deps
```

2. **构建应用**：
```bash
make build
```

3. **设置环境**：
```bash
cp .env.example .env
# 编辑 .env 文件，添加您的 API 密钥
```

## 🎯 使用方法

### 命令行界面

```bash
./gemini-mcp [选项]

选项:
  -transport string    传输类型：stdio（默认）
  -version            显示版本信息
```

### Stdio 模式（MCP 集成）

运行服务器以直接与 MCP 客户端集成：
```bash
./gemini-mcp
```

### 测试 MCP 协议

```bash
# 测试基本连接
./test_mcp.sh

# 手动测试
echo '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' | ./gemini-mcp
```

## 🛠️ 可用工具

### 1. **gemini_image_generation**
使用 Google 最新的 Gemini 图像生成模型生成高质量图像，具备先进的风格控制和质量设置。

**主要功能：**
- 高级风格控制和艺术选项
- 多语言提示支持
- 可定制的纵横比和质量设置
- 内容安全级别和文本渲染选项

**参数：**
- `prompt`（必需）：所需图像的详细描述
- `model`：Gemini 模型变体（默认：`gemini-3-pro-preview`）
- `output_directory`：本地保存路径

### 2. **gemini_image_edit**
使用 Google 的 Gemini AI 模型编辑现有图像，进行有针对性的修改。

**主要功能：**
- 有针对性的图像修改和风格转换
- 对象添加/删除功能
- 背景更改，同时保留原始特征
- 对编辑类型的精确控制

**参数：**
- `prompt`（必需）：所需编辑的描述
- `image_path`：要编辑的图像路径
- `edit_type`：编辑操作类型
- `output_directory`：本地保存路径

### 3. **gemini_multi_image**
使用 Google 的 Gemini AI 模型组合和混合多张图像。

**主要功能：**
- 将 2-3 张图像合并为统一的构图
- 创建拼贴、叠加和无缝混合
- 跨场景的角色一致性
- 创意构图的风格统一

**参数：**
- `prompt`（必需）：所需构图的描述
- `image_paths`：要组合的图像路径数组
- `blend_mode`：如何组合图像
- `output_directory`：本地保存路径

### 4. **veo_text_to_video**
使用 Google 的 Veo 3.1 模型从文本提示生成 4-8 秒视频，支持原生音频。

**主要功能：**
- 详细的场景描述和摄像机运动
- 真实的物理效果和自然运动
- 原生音频生成（对话、音效、音乐）
- 支持 16:9/9:16 纵横比
- 720p/1080p 分辨率选项
- 灵活时长：4、6 或 8 秒
- SynthID 水印

**参数：**
- `prompt`（必需）：详细的视频场景描述
- `negative_prompt`：视频中要避免的内容
- `aspect_ratio`：视频比例（`16:9`、`9:16`）
- `resolution`：视频质量（`720p`、`1080p`）
- `model`：Veo 变体（默认：`veo-3.1-generate-preview`）
- `seed`：可选的种子值用于可重现性
- `output_directory`：本地保存路径

### 6. **veo_image_to_video**
使用 Google 的 Veo 3.1 模型将静态图像动画化为 4-8 秒视频，支持原生音频。

**主要功能：**
- 将照片转换为动态场景
- 自然运动和摄像机运动
- 输入图像成为起始帧
- 真实的物理模拟

**参数：**
- `prompt`（必需）：所需动画的描述
- `image_path`：输入图像路径
- `negative_prompt`：要避免的内容
- `aspect_ratio`：视频比例（`16:9`、`9:16`）
- `resolution`：视频质量（`720p`、`1080p`）
- `model`：Veo 变体（默认：`veo-3.1-generate-preview`）
- `output_directory`：本地保存路径

### 7. **veo_generate_video**（旧版）
通用视频生成工具，支持文本生成视频和图像生成视频创作。

**主要功能：**
- 与现有工作流程的向后兼容性
- 支持文本和图像输入
- 高级场景构图
- 自动操作轮询

**参数：**
- `prompt`（必需）：视频描述
- `image_path`：可选的输入图像（用于图像生成视频）
- `aspect_ratio`：视频比例
- `resolution`：视频质量
- `negative_prompt`：内容排除
- `output_directory`：本地保存路径

## 🔧 环境配置

| 变量 | 描述 | 默认值 | 必需 |
|----------|-------------|---------|----------|
| `GOOGLE_API_KEY` | Gemini API 认证密钥 | - | ✅ 是 |
| `GOOGLE_PROJECT_ID` | Google Cloud 项目 ID | - | ❌ 可选 |
| `GOOGLE_LOCATION` | Google Cloud 区域 | `us-central1` | ❌ 可选 |
| `OUTPUT_DIR` | 文件输出目录 | `./output` | ❌ 可选 |
| `TRANSPORT` | MCP 传输协议 | `stdio` | ❌ 可选 |

## 🔌 MCP 客户端集成

### Claude Desktop 配置
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

**注意：** 如果您将二进制文件安装到自定义位置，请使用完整路径：
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

### Cline VSCode 扩展
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

## 🧪 开发

### 从源码构建
```bash
go mod tidy
go build -o gemini-mcp main.go
```

### 测试
```bash
make test
./test_mcp.sh
```

### 代码质量
```bash
make fmt    # 格式化代码
make clean  # 清理构建产物
```

## 📝 实现说明

- **Gemini 集成**：使用 `google.golang.org/genai` 与 Gemini API 后端集成
- **协议合规性**：实现 MCP 2024-11-05 规范
- **图像生成**：完全实现 Gemini 3.0 Pro 模型
- **视频生成**：完整的 Veo 3.1 集成，支持原生音频、操作轮询和正确的文件下载
- **文件管理**：生成的内容保存时包含元数据和时间戳
- **错误处理**：全面的错误响应机制，提供有用的错误信息
- **多模态支持**：支持文本生成图像、图像生成图像、文本生成视频和图像生成视频工作流程

## 🤝 贡献

本项目旨在成为 Google AI 服务的综合 MCP 服务器。欢迎为以下方面做出贡献：
- 额外的模型支持
- 传输协议增强
- 占位服务的完整实现
- 文档改进

## 📄 许可证

MIT 许可证 - 详情请参见 LICENSE 文件。