package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gemini-mcp/internal/common"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

var (
	transport   = flag.String("transport", "", "Transport type (stdio, http, or sse)")
	showVersion = flag.Bool("version", false, "Show version information")
)

// Version information - these will be set during build
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

const (
	serviceName = "gemini-mcp"
)

type Server struct {
	config *common.Config
	client *genai.Client
}

// Input types for tools
type GeminiImageGenerationInput struct {
	Prompt          string   `json:"prompt" jsonschema:"description:Detailed text prompt describing what you want to visualize. Be specific about style, composition, colors, mood, and any particular elements you want included in the image."`
	Model           string   `json:"model,omitempty" jsonschema:"description:Image generation model to use. Supported models: 'gemini-3-pro-image-preview' (default - Gemini 3 Pro with native image generation aka Nano Banana Pro), 'gemini-2.5-flash-image' (fast Gemini image model aka Nano Banana), 'imagen-4.0-generate-001' (Imagen 4.0), 'imagen-3.0-generate-001' (Imagen 3.0).,default:gemini-3-pro-image-preview"`
	Style           string   `json:"style,omitempty" jsonschema:"description:Image style preference such as 'photorealistic', 'artistic', 'cartoon', 'sketch', 'oil painting', 'watercolor', etc."`
	AspectRatio     string   `json:"aspect_ratio,omitempty" jsonschema:"description:Preferred aspect ratio for the image. Supported ratios: '1:1' (square), '3:4', '4:3', '9:16' (portrait), '16:9' (landscape)"`
	Quality         string   `json:"quality,omitempty" jsonschema:"description:Image quality preference: 'high' (2K resolution), 'medium' (1K), 'draft' (1K). Higher quality may take longer to generate.,default:high"`
	SafetyLevel     string   `json:"safety_level,omitempty" jsonschema:"description:Content safety level: 'strict', 'moderate', 'permissive'. Controls content filtering.,default:moderate"`
	Language        string   `json:"language,omitempty" jsonschema:"description:Language for prompt processing. Supported: 'en' (English), 'es-MX' (Spanish Mexico), 'ja' (Japanese), 'zh' (Chinese), 'hi' (Hindi),default:en"`
	IncludeText     bool     `json:"include_text,omitempty" jsonschema:"description:Whether to include high-fidelity text rendering in the image. Enable for images that need clear text elements.,default:false"`
	Tags            []string `json:"tags,omitempty" jsonschema:"description:Optional tags to help categorize or describe the generated image"`
	OutputDirectory string   `json:"output_directory,omitempty" jsonschema:"description:Optional. Local directory path where the generated image and metadata will be saved. If not provided, files will be saved to the default output directory."`
}

type GeminiImageGenerationOutput struct {
	Description   string            `json:"description"`
	Model         string            `json:"model"`
	Style         string            `json:"style,omitempty"`
	AspectRatio   string            `json:"aspect_ratio,omitempty"`
	Quality       string            `json:"quality,omitempty"`
	Language      string            `json:"language,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	SavedFiles    []string          `json:"saved_files,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	GeneratedAt   string            `json:"generated_at"`
	ImagesCreated int               `json:"images_created"`
}

type GeminiImageEditInput struct {
	InputImagePath  string `json:"input_image_path" jsonschema:"description:Path to the input image file to edit (PNG, JPEG, WebP supported)"`
	EditPrompt      string `json:"edit_prompt" jsonschema:"description:Detailed description of how to edit the image. Be specific about what changes to make."`
	Model           string `json:"model,omitempty" jsonschema:"description:Gemini model to use for image editing,default:gemini-3-pro-image-preview"`
	AspectRatio     string `json:"aspect_ratio,omitempty" jsonschema:"description:Preferred aspect ratio for the edited image. Common ratios: '1:1' (square), '16:9' (landscape), '9:16' (portrait), '4:3', '3:4'"`
	PreserveStyle   bool   `json:"preserve_style,omitempty" jsonschema:"description:Whether to preserve the original image style during editing,default:true"`
	EditType        string `json:"edit_type,omitempty" jsonschema:"description:Type of edit: 'modify' (change elements), 'add' (add new elements), 'remove' (remove elements), 'style' (change style),default:modify"`
	MaskArea        string `json:"mask_area,omitempty" jsonschema:"description:Specific area to focus edits on (e.g., 'background', 'foreground', 'top-left', 'center')"`
	OutputDirectory string `json:"output_directory,omitempty" jsonschema:"description:Optional. Local directory path where the edited image will be saved."`
}

type GeminiImageEditOutput struct {
	OriginalImage string            `json:"original_image"`
	EditedImage   string            `json:"edited_image,omitempty"`
	EditType      string            `json:"edit_type"`
	AspectRatio   string            `json:"aspect_ratio,omitempty"`
	Model         string            `json:"model"`
	SavedFiles    []string          `json:"saved_files,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	GeneratedAt   string            `json:"generated_at"`
}

type GeminiMultiImageInput struct {
	InputImagePaths []string `json:"input_image_paths" jsonschema:"description:Paths to input image files to combine (2-3 images recommended)"`
	CombinePrompt   string   `json:"combine_prompt" jsonschema:"description:Description of how to combine or blend the images"`
	Model           string   `json:"model,omitempty" jsonschema:"description:Gemini model to use for multi-image processing,default:gemini-3-pro-image-preview"`
	AspectRatio     string   `json:"aspect_ratio,omitempty" jsonschema:"description:Preferred aspect ratio for the combined image. Common ratios: '1:1' (square), '16:9' (landscape), '9:16' (portrait), '4:3', '3:4'"`
	BlendMode       string   `json:"blend_mode,omitempty" jsonschema:"description:How to blend images: 'merge', 'collage', 'overlay', 'sequence',default:merge"`
	OutputStyle     string   `json:"output_style,omitempty" jsonschema:"description:Style for the combined image: 'photorealistic', 'artistic', 'seamless'"`
	OutputDirectory string   `json:"output_directory,omitempty" jsonschema:"description:Optional. Local directory path where the combined image will be saved."`
}

type GeminiMultiImageOutput struct {
	InputImages     []string          `json:"input_images"`
	CombinedImage   string            `json:"combined_image,omitempty"`
	BlendMode       string            `json:"blend_mode"`
	AspectRatio     string            `json:"aspect_ratio,omitempty"`
	Model           string            `json:"model"`
	SavedFiles      []string          `json:"saved_files,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	GeneratedAt     string            `json:"generated_at"`
	ImagesProcessed int               `json:"images_processed"`
}

// Text-to-Video Generation
type VeoTextToVideoInput struct {
	Prompt          string `json:"prompt" jsonschema:"description:Detailed text prompt describing the video content (max 1024 tokens). Be specific about scenes, actions, camera movements, visual style, and any audio elements you want included."`
	NegativePrompt  string `json:"negative_prompt,omitempty" jsonschema:"description:Description of what should NOT appear in the video. Use to avoid unwanted content or styles."`
	AspectRatio     string `json:"aspect_ratio,omitempty" jsonschema:"description:Video width-to-height ratio,default:16:9,enum:16:9,enum:9:16"`
	Resolution      string `json:"resolution,omitempty" jsonschema:"description:Video resolution. Note: 1080p only supported for 16:9 aspect ratio,default:720p,enum:720p,enum:1080p"`
	Model           string `json:"model,omitempty" jsonschema:"description:Veo model version to use,default:veo-3.0-generate-001,enum:veo-3.0-generate-001,enum:veo-3.0-fast-generate-001,enum:veo-2.0-generate-001"`
	Seed            int    `json:"seed,omitempty" jsonschema:"description:Optional seed value for slight reproducibility in generation"`
	OutputDirectory string `json:"output_directory,omitempty" jsonschema:"description:Local directory path where the 8-second MP4 video will be saved. Videos have 2-day retention on server and include SynthID watermark."`
}

// Image-to-Video Generation
type VeoImageToVideoInput struct {
	ImagePath       string `json:"image_path" jsonschema:"description:Path to the initial image file to animate as the starting frame of the video. Supports JPEG, PNG formats."`
	Prompt          string `json:"prompt" jsonschema:"description:Text prompt describing how the image should be animated and what should happen in the video (max 1024 tokens)."`
	NegativePrompt  string `json:"negative_prompt,omitempty" jsonschema:"description:Description of what should NOT happen in the animation or appear in the video."`
	AspectRatio     string `json:"aspect_ratio,omitempty" jsonschema:"description:Video width-to-height ratio,default:16:9,enum:16:9,enum:9:16"`
	Resolution      string `json:"resolution,omitempty" jsonschema:"description:Video resolution. Note: 1080p only supported for 16:9 aspect ratio,default:720p,enum:720p,enum:1080p"`
	Model           string `json:"model,omitempty" jsonschema:"description:Veo model version to use,default:veo-3.0-generate-001,enum:veo-3.0-generate-001,enum:veo-3.0-fast-generate-001,enum:veo-2.0-generate-001"`
	Seed            int    `json:"seed,omitempty" jsonschema:"description:Optional seed value for slight reproducibility in generation"`
	OutputDirectory string `json:"output_directory,omitempty" jsonschema:"description:Local directory path where the 8-second MP4 video will be saved. Videos have 2-day retention on server and include SynthID watermark."`
}

// Legacy input type for backward compatibility
type VeoGenerationInput struct {
	Prompt          string `json:"prompt" jsonschema:"description:Detailed text prompt describing the video content (max 1024 tokens). Be specific about scenes, actions, camera movements, visual style, and any audio elements you want included."`
	NegativePrompt  string `json:"negative_prompt,omitempty" jsonschema:"description:Description of what should NOT appear in the video. Use to avoid unwanted content or styles."`
	AspectRatio     string `json:"aspect_ratio,omitempty" jsonschema:"description:Video width-to-height ratio,default:16:9,enum:16:9,enum:9:16"`
	Resolution      string `json:"resolution,omitempty" jsonschema:"description:Video resolution. Note: 1080p only supported for 16:9 aspect ratio,default:720p,enum:720p,enum:1080p"`
	Model           string `json:"model,omitempty" jsonschema:"description:Veo model version to use,default:veo-3.0-generate-001,enum:veo-3.0-generate-001,enum:veo-3.0-fast-generate-001,enum:veo-2.0-generate-001"`
	ImagePath       string `json:"image_path,omitempty" jsonschema:"description:Optional path to initial image file to animate as the starting frame of the video"`
	Seed            int    `json:"seed,omitempty" jsonschema:"description:Optional seed value for slight reproducibility in generation"`
	OutputDirectory string `json:"output_directory,omitempty" jsonschema:"description:Local directory path where the 8-second MP4 video will be saved. Videos have 2-day retention on server and include SynthID watermark."`
}

type VeoGenerationOutput struct {
	OperationID     string            `json:"operation_id,omitempty"`
	Status          string            `json:"status"`
	VideoURL        string            `json:"video_url,omitempty"`
	SavedFiles      []string          `json:"saved_files,omitempty"`
	Model           string            `json:"model"`
	AspectRatio     string            `json:"aspect_ratio"`
	Resolution      string            `json:"resolution"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	GeneratedAt     string            `json:"generated_at"`
	EstimatedLength string            `json:"estimated_length"`
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", serviceName, version)
		fmt.Println("A Model Context Protocol server for Google Gemini AI services")
		fmt.Printf("Built: %s\n", buildTime)
		fmt.Printf("Commit: %s\n", gitCommit)
		return
	}

	// Load configuration
	config := common.LoadConfig()
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Override transport if specified via flag
	if *transport != "" {
		config.Transport = *transport
	}

	// Create Gemini client
	ctx := context.Background()
	clientConfig := &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	server := &Server{
		config: config,
		client: client,
	}

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    serviceName,
		Version: version,
	}, nil)

	// Register tools
	server.registerTools(mcpServer)

	log.Printf("Starting %s v%s (Transport: %s)", serviceName, version, config.Transport)

	// Run server with stdio transport
	if err := mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func (s *Server) registerTools(server *mcp.Server) {
	// Register gemini_image_generation tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gemini_image_generation",
		Description: "Generate high-quality images using Google's latest Gemini image generation models. Supports text-to-image generation with advanced style control, quality settings, and multi-language prompts. Features include customizable aspect ratios, artistic styles, content safety levels, and high-fidelity text rendering.",
	}, s.handleGeminiImageGeneration)

	// Register gemini_image_edit tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gemini_image_edit",
		Description: "Edit existing images using Google's Gemini AI models. Supports targeted image modifications, style transfers, object addition/removal, and background changes. Provides precise control over edit types and can preserve original image characteristics while making specific alterations.",
	}, s.handleGeminiImageEdit)

	// Register gemini_multi_image tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gemini_multi_image",
		Description: "Combine and blend multiple images using Google's Gemini AI models. Supports merging 2-3 images into cohesive compositions, creating collages, overlays, and seamless blends. Ideal for character consistency across scenes, style unification, and creative image compositions.",
	}, s.handleGeminiMultiImage)

	// Register veo_text_to_video tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "veo_text_to_video",
		Description: "Generate 8-second videos from text prompts using Google's Veo 3.0 models. Create videos with detailed scene descriptions, camera movements, and realistic physics. Supports 16:9/9:16 aspect ratios, 720p/1080p resolution, negative prompts, and includes SynthID watermarking.",
	}, s.handleVeoTextToVideo)

	// Register veo_image_to_video tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "veo_image_to_video",
		Description: "Animate static images into 8-second videos using Google's Veo 3.0 models. Transform photos into dynamic scenes with natural motion, camera movements, and realistic physics. Input image becomes the starting frame of the generated video.",
	}, s.handleVeoImageToVideo)

	// Register veo_generate_video tool (legacy)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "veo_generate_video",
		Description: "Generate high-quality 8-second videos using Google's Veo 3.0 video generation models. Supports both text-to-video and image-to-video creation with advanced scene composition, camera movements, and realistic physics. Features include 16:9 and 9:16 aspect ratios, 720p/1080p resolution, negative prompts for content exclusion, and automatic operation polling with video URL retrieval.",
	}, s.handleVeoGeneration)

}

func (s *Server) handleGeminiImageGeneration(ctx context.Context, req *mcp.CallToolRequest, input GeminiImageGenerationInput) (*mcp.CallToolResult, GeminiImageGenerationOutput, error) {
	if input.Prompt == "" {
		return nil, GeminiImageGenerationOutput{}, fmt.Errorf("prompt is required")
	}

	// Set defaults
	model := input.Model
	if model == "" {
		model = "gemini-3-pro-image-preview" // Default to Gemini 3 Pro Image for native image generation
	}

	style := input.Style
	if style == "" {
		style = "photorealistic"
	}

	quality := input.Quality
	if quality == "" {
		quality = "high"
	}

	language := input.Language
	if language == "" {
		language = "en"
	}

	log.Printf("Generating image with model %s for prompt: %s (style: %s, quality: %s)", model, input.Prompt, style, quality)

	// Build enhanced prompt with style and parameters
	var promptParts []string
	promptParts = append(promptParts, input.Prompt)

	if style != "" && style != "photorealistic" {
		promptParts = append(promptParts, fmt.Sprintf("in %s style", style))
	}

	if input.IncludeText {
		promptParts = append(promptParts, "with high-fidelity text rendering")
	}

	if quality == "high" {
		promptParts = append(promptParts, "highly detailed")
	}

	promptText := strings.Join(promptParts, ", ")

	// Map quality to image size
	// Quality: "high" -> "2K", "medium" -> "1K", "draft" -> "1K"
	imageSize := "2K"
	if quality == "medium" || quality == "draft" {
		imageSize = "1K"
	}

	var savedFiles []string
	timestamp := time.Now().Format("20060102_150405")
	var imagesCreated int

	// Determine output directory
	outputDir := input.OutputDirectory
	if outputDir == "" {
		outputDir = s.config.OutputDir
	}

	// Check if using Gemini native image generation or Imagen
	isGeminiModel := strings.HasPrefix(model, "gemini-")

	if isGeminiModel {
		// Use GenerateContent for Gemini native image generation models
		log.Printf("Using GenerateContent API for Gemini model: %s", model)

		// Create content with text prompt
		contents := []*genai.Content{
			genai.NewContentFromText(promptText, genai.RoleUser),
		}

		// Configure for image generation
		config := &genai.GenerateContentConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
		}

		// Generate content
		response, err := s.client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			return nil, GeminiImageGenerationOutput{}, fmt.Errorf("error generating image: %v", err)
		}

		if response == nil || len(response.Candidates) == 0 {
			return nil, GeminiImageGenerationOutput{}, fmt.Errorf("no image was generated")
		}

		// Process response and extract images
		if outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Printf("Error creating output directory: %v", err)
			}
		}

		for _, candidate := range response.Candidates {
			if candidate.Content == nil {
				continue
			}

			for i, part := range candidate.Content.Parts {
				if part.InlineData != nil && len(part.InlineData.Data) > 0 {
					imagesCreated++
					if outputDir != "" {
						filename := fmt.Sprintf("gemini_generated_%s_%s_%d.png", style, timestamp, i)
						outputPath := filepath.Join(outputDir, filename)

						if err := os.WriteFile(outputPath, part.InlineData.Data, 0644); err == nil {
							savedFiles = append(savedFiles, outputPath)
							log.Printf("Saved generated image to: %s", outputPath)
						} else {
							log.Printf("Error saving image: %v", err)
						}
					}
				}
			}
		}

		if imagesCreated == 0 {
			return nil, GeminiImageGenerationOutput{}, fmt.Errorf("no images were generated in response")
		}

	} else {
		// Use GenerateImages for Imagen models
		log.Printf("Using GenerateImages API for Imagen model: %s", model)

		// Configure GenerateImagesConfig
		config := &genai.GenerateImagesConfig{
			NumberOfImages: 1,
		}

		// Set aspect ratio if provided
		if input.AspectRatio != "" {
			config.AspectRatio = input.AspectRatio
		}

		config.ImageSize = imageSize

		// Set safety level
		if input.SafetyLevel != "" {
			switch input.SafetyLevel {
			case "strict":
				config.SafetyFilterLevel = genai.SafetyFilterLevelBlockLowAndAbove
			case "permissive":
				config.SafetyFilterLevel = genai.SafetyFilterLevelBlockOnlyHigh
			default:
				config.SafetyFilterLevel = genai.SafetyFilterLevelBlockMediumAndAbove
			}
		}

		// Set language if supported
		if language != "en" {
			switch language {
			case "zh":
				config.Language = genai.ImagePromptLanguageZh
			case "hi":
				config.Language = genai.ImagePromptLanguageHi
			case "ja":
				config.Language = genai.ImagePromptLanguageJa
			case "es-MX", "es":
				config.Language = genai.ImagePromptLanguageEs
			}
		}

		// Generate images using the dedicated GenerateImages method
		response, err := s.client.Models.GenerateImages(ctx, model, promptText, config)
		if err != nil {
			return nil, GeminiImageGenerationOutput{}, fmt.Errorf("error generating images: %v", err)
		}

		if response == nil || len(response.GeneratedImages) == 0 {
			return nil, GeminiImageGenerationOutput{}, fmt.Errorf("no images were generated")
		}

		imagesCreated = len(response.GeneratedImages)

		// Save generated images to local directory
		if outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err == nil {
				for i, genImage := range response.GeneratedImages {
					if genImage.Image != nil && len(genImage.Image.ImageBytes) > 0 {
						filename := fmt.Sprintf("imagen_generated_%s_%s_%d.png", style, timestamp, i)
						outputPath := filepath.Join(outputDir, filename)

						if err := os.WriteFile(outputPath, genImage.Image.ImageBytes, 0644); err == nil {
							savedFiles = append(savedFiles, outputPath)
							log.Printf("Saved generated image to: %s", outputPath)
						} else {
							log.Printf("Error saving image: %v", err)
						}
					}
				}
			} else {
				log.Printf("Error creating output directory: %v", err)
			}
		}
	}

	// Create result description
	resultText := fmt.Sprintf("Successfully generated %d image(s) using %s", imagesCreated, model)

	// Create metadata
	metadata := map[string]string{
		"original_prompt": input.Prompt,
		"enhanced_prompt": promptText,
		"quality":         quality,
		"safety_level":    input.SafetyLevel,
		"image_size":      imageSize,
	}

	// Save metadata if output directory is specified
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err == nil {
			filename := fmt.Sprintf("image_metadata_%s.json", timestamp)
			outputPath := filepath.Join(outputDir, filename)

			metadataContent := map[string]interface{}{
				"model":           model,
				"prompt":          input.Prompt,
				"enhanced_prompt": promptText,
				"style":           style,
				"aspect_ratio":    input.AspectRatio,
				"quality":         quality,
				"image_size":      imageSize,
				"language":        language,
				"include_text":    input.IncludeText,
				"tags":            input.Tags,
				"generated_at":    timestamp,
				"images_created":  imagesCreated,
			}

			if jsonData, err := json.MarshalIndent(metadataContent, "", "  "); err == nil {
				if err := os.WriteFile(outputPath, jsonData, 0644); err == nil {
					savedFiles = append(savedFiles, outputPath)
				}
			}
		}
	}

	return nil, GeminiImageGenerationOutput{
		Description:   resultText,
		Model:         model,
		Style:         style,
		AspectRatio:   input.AspectRatio,
		Quality:       quality,
		Language:      language,
		Tags:          input.Tags,
		SavedFiles:    savedFiles,
		Metadata:      metadata,
		GeneratedAt:   timestamp,
		ImagesCreated: imagesCreated,
	}, nil
}

func (s *Server) handleGeminiImageEdit(ctx context.Context, req *mcp.CallToolRequest, input GeminiImageEditInput) (*mcp.CallToolResult, GeminiImageEditOutput, error) {
	if input.InputImagePath == "" {
		return nil, GeminiImageEditOutput{}, fmt.Errorf("input_image_path is required")
	}
	if input.EditPrompt == "" {
		return nil, GeminiImageEditOutput{}, fmt.Errorf("edit_prompt is required")
	}

	model := input.Model
	if model == "" {
		model = "gemini-3-pro-image-preview"
	}

	editType := input.EditType
	if editType == "" {
		editType = "modify"
	}

	log.Printf("Editing image %s with model %s: %s", input.InputImagePath, model, input.EditPrompt)

	// Read input image
	imgData, err := os.ReadFile(input.InputImagePath)
	if err != nil {
		return nil, GeminiImageEditOutput{}, fmt.Errorf("failed to read input image: %v", err)
	}

	// Build edit prompt with instructions
	var promptParts []string
	promptParts = append(promptParts, input.EditPrompt)

	if input.AspectRatio != "" {
		promptParts = append(promptParts, fmt.Sprintf("Aspect ratio: %s", input.AspectRatio))
	}

	if input.PreserveStyle {
		promptParts = append(promptParts, "Preserve the original image style and characteristics")
	}

	if input.MaskArea != "" {
		promptParts = append(promptParts, fmt.Sprintf("Focus changes on the %s area", input.MaskArea))
	}

	switch editType {
	case "add":
		promptParts = append(promptParts, "Add the requested elements to the image")
	case "remove":
		promptParts = append(promptParts, "Remove the specified elements from the image")
	case "style":
		promptParts = append(promptParts, "Change the style while keeping the subject matter")
	default:
		promptParts = append(promptParts, "Modify the image as requested")
	}

	promptText := strings.Join(promptParts, ". ")

	// Create content parts with image and text
	parts := []*genai.Part{
		genai.NewPartFromText(promptText),
		&genai.Part{
			InlineData: &genai.Blob{
				MIMEType: "image/png",
				Data:     imgData,
			},
		},
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	response, err := s.client.Models.GenerateContent(ctx, model, contents, nil)
	if err != nil {
		return nil, GeminiImageEditOutput{}, fmt.Errorf("error editing image: %v", err)
	}

	if response == nil || len(response.Candidates) == 0 {
		return nil, GeminiImageEditOutput{}, fmt.Errorf("no edited content was generated")
	}

	// Process response
	var savedFiles []string
	timestamp := time.Now().Format("20060102_150405")
	var editedImagePath string

	for _, candidate := range response.Candidates {
		if candidate.Content == nil {
			continue
		}

		for i, part := range candidate.Content.Parts {
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				// Save edited image
				outputDir := input.OutputDirectory
				if outputDir == "" {
					outputDir = s.config.OutputDir
				}

				if outputDir != "" {
					if err := os.MkdirAll(outputDir, 0755); err == nil {
						filename := fmt.Sprintf("gemini_edited_%s_%s_%d.png", editType, timestamp, i)
						outputPath := filepath.Join(outputDir, filename)

						if err := os.WriteFile(outputPath, part.InlineData.Data, 0644); err == nil {
							savedFiles = append(savedFiles, outputPath)
							editedImagePath = outputPath
							log.Printf("Saved edited image to: %s", outputPath)
						}
					}
				}
			}
		}
	}

	// Create metadata
	metadata := map[string]string{
		"original_image": input.InputImagePath,
		"edit_prompt":    input.EditPrompt,
		"edit_type":      editType,
		"aspect_ratio":   input.AspectRatio,
		"preserve_style": fmt.Sprintf("%t", input.PreserveStyle),
		"mask_area":      input.MaskArea,
	}

	return nil, GeminiImageEditOutput{
		OriginalImage: input.InputImagePath,
		EditedImage:   editedImagePath,
		EditType:      editType,
		AspectRatio:   input.AspectRatio,
		Model:         model,
		SavedFiles:    savedFiles,
		Metadata:      metadata,
		GeneratedAt:   timestamp,
	}, nil
}

func (s *Server) handleGeminiMultiImage(ctx context.Context, req *mcp.CallToolRequest, input GeminiMultiImageInput) (*mcp.CallToolResult, GeminiMultiImageOutput, error) {
	if len(input.InputImagePaths) < 2 {
		return nil, GeminiMultiImageOutput{}, fmt.Errorf("at least 2 input images are required")
	}
	if len(input.InputImagePaths) > 3 {
		return nil, GeminiMultiImageOutput{}, fmt.Errorf("maximum 3 input images supported")
	}
	if input.CombinePrompt == "" {
		return nil, GeminiMultiImageOutput{}, fmt.Errorf("combine_prompt is required")
	}

	model := input.Model
	if model == "" {
		model = "gemini-3-pro-image-preview"
	}

	blendMode := input.BlendMode
	if blendMode == "" {
		blendMode = "merge"
	}

	log.Printf("Combining %d images with model %s: %s", len(input.InputImagePaths), model, input.CombinePrompt)

	// Build parts array starting with text prompt
	var promptParts []string
	promptParts = append(promptParts, input.CombinePrompt)

	if input.AspectRatio != "" {
		promptParts = append(promptParts, fmt.Sprintf("Aspect ratio: %s", input.AspectRatio))
	}

	switch blendMode {
	case "collage":
		promptParts = append(promptParts, "Create a collage arrangement of the images")
	case "overlay":
		promptParts = append(promptParts, "Overlay the images with artistic blending")
	case "sequence":
		promptParts = append(promptParts, "Arrange the images in a sequence or timeline")
	default:
		promptParts = append(promptParts, "Seamlessly merge the images into a cohesive composition")
	}

	if input.OutputStyle != "" {
		promptParts = append(promptParts, fmt.Sprintf("Output style: %s", input.OutputStyle))
	}

	promptText := strings.Join(promptParts, ". ")
	parts := []*genai.Part{genai.NewPartFromText(promptText)}

	// Add all input images to parts
	for i, imagePath := range input.InputImagePaths {
		imgData, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, GeminiMultiImageOutput{}, fmt.Errorf("failed to read image %d (%s): %v", i+1, imagePath, err)
		}

		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: "image/png",
				Data:     imgData,
			},
		})
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	response, err := s.client.Models.GenerateContent(ctx, model, contents, nil)
	if err != nil {
		return nil, GeminiMultiImageOutput{}, fmt.Errorf("error combining images: %v", err)
	}

	if response == nil || len(response.Candidates) == 0 {
		return nil, GeminiMultiImageOutput{}, fmt.Errorf("no combined content was generated")
	}

	// Process response
	var savedFiles []string
	timestamp := time.Now().Format("20060102_150405")
	var combinedImagePath string

	for _, candidate := range response.Candidates {
		if candidate.Content == nil {
			continue
		}

		for i, part := range candidate.Content.Parts {
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				// Save combined image
				outputDir := input.OutputDirectory
				if outputDir == "" {
					outputDir = s.config.OutputDir
				}

				if outputDir != "" {
					if err := os.MkdirAll(outputDir, 0755); err == nil {
						filename := fmt.Sprintf("gemini_combined_%s_%s_%d.png", blendMode, timestamp, i)
						outputPath := filepath.Join(outputDir, filename)

						if err := os.WriteFile(outputPath, part.InlineData.Data, 0644); err == nil {
							savedFiles = append(savedFiles, outputPath)
							combinedImagePath = outputPath
							log.Printf("Saved combined image to: %s", outputPath)
						}
					}
				}
			}
		}
	}

	// Create metadata
	metadata := map[string]string{
		"combine_prompt": input.CombinePrompt,
		"blend_mode":     blendMode,
		"aspect_ratio":   input.AspectRatio,
		"output_style":   input.OutputStyle,
		"images_count":   fmt.Sprintf("%d", len(input.InputImagePaths)),
	}

	return nil, GeminiMultiImageOutput{
		InputImages:     input.InputImagePaths,
		CombinedImage:   combinedImagePath,
		BlendMode:       blendMode,
		AspectRatio:     input.AspectRatio,
		Model:           model,
		SavedFiles:      savedFiles,
		Metadata:        metadata,
		GeneratedAt:     timestamp,
		ImagesProcessed: len(input.InputImagePaths),
	}, nil
}

func (s *Server) handleVeoGeneration(ctx context.Context, req *mcp.CallToolRequest, input VeoGenerationInput) (*mcp.CallToolResult, VeoGenerationOutput, error) {
	if input.Prompt == "" {
		return nil, VeoGenerationOutput{}, fmt.Errorf("prompt is required")
	}

	// Set defaults
	aspectRatio := input.AspectRatio
	if aspectRatio == "" {
		aspectRatio = "16:9"
	}

	resolution := input.Resolution
	if resolution == "" {
		resolution = "720p"
	}

	model := input.Model
	if model == "" {
		model = "veo-3.0-generate-001"
	}

	log.Printf("Generating video with model %s for prompt: %s (aspect: %s, resolution: %s)", model, input.Prompt, aspectRatio, resolution)

	timestamp := time.Now().Format("20060102_150405")

	// Build prompt with negative prompt if specified
	promptText := input.Prompt
	if input.NegativePrompt != "" {
		promptText = fmt.Sprintf("%s. Avoid: %s", input.Prompt, input.NegativePrompt)
	}

	// Generate video using Gemini API - correct signature from documentation
	operation, err := s.client.Models.GenerateVideos(
		ctx,
		model,
		promptText,
		nil, // image parameter (nil for text-only)
		nil, // config parameter (nil to use defaults)
	)
	if err != nil {
		return nil, VeoGenerationOutput{}, fmt.Errorf("error starting video generation: %v", err)
	}

	operationID := operation.Name
	log.Printf("Video generation started with operation ID: %s", operationID)

	// Poll operation status until completion
	maxAttempts := 60 // 10 minutes max
	for i := 0; i < maxAttempts && !operation.Done; i++ {
		log.Printf("Waiting for video generation to complete... (attempt %d/%d)", i+1, maxAttempts)
		time.Sleep(10 * time.Second)
		operation, err = s.client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			log.Printf("Error checking operation status: %v", err)
			break
		}
	}

	var savedFiles []string
	var videoURL string
	status := "generating"

	if operation.Done {
		if operation.Error != nil {
			status = "failed"
			log.Printf("Video generation failed: %v", operation.Error)
		} else if len(operation.Response.GeneratedVideos) > 0 {
			status = "completed"
			video := operation.Response.GeneratedVideos[0]
			log.Printf("Video generation completed successfully")

			// Download and save video if output directory is specified
			if input.OutputDirectory != "" {
				if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
					filename := fmt.Sprintf("veo_video_%s.mp4", timestamp)
					outputPath := filepath.Join(input.OutputDirectory, filename)

					// Download the video file following official documentation pattern
					s.client.Files.Download(ctx, video.Video, nil)

					// Save the video bytes to file
					err = os.WriteFile(outputPath, video.Video.VideoBytes, 0644)
					if err != nil {
						log.Printf("Error saving video file: %v", err)
					} else {
						savedFiles = append(savedFiles, outputPath)
						videoURL = outputPath
						log.Printf("Video saved to: %s", outputPath)
					}
				}
			}
		}
	} else {
		status = "timeout"
		log.Printf("Video generation timed out after 10 minutes")
	}

	// Save metadata
	metadata := map[string]string{
		"original_prompt": input.Prompt,
		"negative_prompt": input.NegativePrompt,
		"operation_id":    operationID,
	}

	if input.OutputDirectory != "" {
		if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
			filename := fmt.Sprintf("veo_metadata_%s.json", timestamp)
			outputPath := filepath.Join(input.OutputDirectory, filename)

			metadataContent := map[string]interface{}{
				"model":            model,
				"prompt":           input.Prompt,
				"negative_prompt":  input.NegativePrompt,
				"aspect_ratio":     aspectRatio,
				"resolution":       resolution,
				"operation_id":     operationID,
				"video_url":        videoURL,
				"status":           status,
				"generated_at":     timestamp,
				"estimated_length": "8 seconds",
			}

			if jsonData, err := json.MarshalIndent(metadataContent, "", "  "); err == nil {
				if err := os.WriteFile(outputPath, jsonData, 0644); err == nil {
					savedFiles = append(savedFiles, outputPath)
				}
			}
		}
	}

	return nil, VeoGenerationOutput{
		OperationID:     operationID,
		Status:          status,
		VideoURL:        videoURL,
		SavedFiles:      savedFiles,
		Model:           model,
		AspectRatio:     aspectRatio,
		Resolution:      resolution,
		Metadata:        metadata,
		GeneratedAt:     timestamp,
		EstimatedLength: "8 seconds",
	}, nil
}

func (s *Server) handleVeoTextToVideo(ctx context.Context, req *mcp.CallToolRequest, input VeoTextToVideoInput) (*mcp.CallToolResult, VeoGenerationOutput, error) {
	if input.Prompt == "" {
		return nil, VeoGenerationOutput{}, fmt.Errorf("prompt is required")
	}

	// Set defaults
	aspectRatio := input.AspectRatio
	if aspectRatio == "" {
		aspectRatio = "16:9"
	}

	resolution := input.Resolution
	if resolution == "" {
		resolution = "720p"
	}

	model := input.Model
	if model == "" {
		model = "veo-3.0-generate-001"
	}

	log.Printf("Generating text-to-video with model %s for prompt: %s (aspect: %s, resolution: %s)", model, input.Prompt, aspectRatio, resolution)

	timestamp := time.Now().Format("20060102_150405")

	// Build prompt with negative prompt if specified
	promptText := input.Prompt
	if input.NegativePrompt != "" {
		promptText = fmt.Sprintf("%s. Avoid: %s", input.Prompt, input.NegativePrompt)
	}

	// Generate video using Gemini API - text-to-video (no image)
	operation, err := s.client.Models.GenerateVideos(
		ctx,
		model,
		promptText,
		nil, // No image for text-to-video
		nil, // Use default config
	)
	if err != nil {
		return nil, VeoGenerationOutput{}, fmt.Errorf("error starting text-to-video generation: %v", err)
	}

	operationID := operation.Name
	log.Printf("Text-to-video generation started with operation ID: %s", operationID)

	// Poll operation status until completion
	maxAttempts := 60 // 10 minutes max
	for i := 0; i < maxAttempts && !operation.Done; i++ {
		log.Printf("Waiting for text-to-video generation to complete... (attempt %d/%d)", i+1, maxAttempts)
		time.Sleep(10 * time.Second)
		operation, err = s.client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			log.Printf("Error checking operation status: %v", err)
			break
		}
	}

	var savedFiles []string
	var videoURL string
	status := "generating"

	if operation.Done {
		if operation.Error != nil {
			status = "failed"
			log.Printf("Text-to-video generation failed: %v", operation.Error)
		} else if len(operation.Response.GeneratedVideos) > 0 {
			status = "completed"
			video := operation.Response.GeneratedVideos[0]
			log.Printf("Text-to-video generation completed successfully")

			// Download and save video if output directory is specified
			if input.OutputDirectory != "" {
				if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
					filename := fmt.Sprintf("veo_text_to_video_%s.mp4", timestamp)
					outputPath := filepath.Join(input.OutputDirectory, filename)

					// Download the video file following official documentation pattern
					s.client.Files.Download(ctx, video.Video, nil)

					// Save the video bytes to file
					err = os.WriteFile(outputPath, video.Video.VideoBytes, 0644)
					if err != nil {
						log.Printf("Error saving video file: %v", err)
					} else {
						savedFiles = append(savedFiles, outputPath)
						videoURL = outputPath
						log.Printf("Text-to-video saved to: %s", outputPath)
					}
				}
			}
		}
	} else {
		status = "timeout"
		log.Printf("Text-to-video generation timed out after 10 minutes")
	}

	// Save metadata
	metadata := map[string]string{
		"generation_type": "text-to-video",
		"original_prompt": input.Prompt,
		"negative_prompt": input.NegativePrompt,
		"operation_id":    operationID,
	}

	if input.Seed > 0 {
		metadata["seed"] = fmt.Sprintf("%d", input.Seed)
	}

	if input.OutputDirectory != "" {
		if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
			filename := fmt.Sprintf("veo_text_to_video_metadata_%s.json", timestamp)
			outputPath := filepath.Join(input.OutputDirectory, filename)

			metadataContent := map[string]interface{}{
				"generation_type":  "text-to-video",
				"model":            model,
				"prompt":           input.Prompt,
				"negative_prompt":  input.NegativePrompt,
				"aspect_ratio":     aspectRatio,
				"resolution":       resolution,
				"seed":             input.Seed,
				"operation_id":     operationID,
				"video_url":        videoURL,
				"status":           status,
				"generated_at":     timestamp,
				"estimated_length": "8 seconds",
			}

			if jsonData, err := json.MarshalIndent(metadataContent, "", "  "); err == nil {
				if err := os.WriteFile(outputPath, jsonData, 0644); err == nil {
					savedFiles = append(savedFiles, outputPath)
				}
			}
		}
	}

	return nil, VeoGenerationOutput{
		OperationID:     operationID,
		Status:          status,
		VideoURL:        videoURL,
		SavedFiles:      savedFiles,
		Model:           model,
		AspectRatio:     aspectRatio,
		Resolution:      resolution,
		Metadata:        metadata,
		GeneratedAt:     timestamp,
		EstimatedLength: "8 seconds",
	}, nil
}

func (s *Server) handleVeoImageToVideo(ctx context.Context, req *mcp.CallToolRequest, input VeoImageToVideoInput) (*mcp.CallToolResult, VeoGenerationOutput, error) {
	if input.ImagePath == "" {
		return nil, VeoGenerationOutput{}, fmt.Errorf("image_path is required")
	}
	if input.Prompt == "" {
		return nil, VeoGenerationOutput{}, fmt.Errorf("prompt is required")
	}

	// Check if image file exists
	if _, err := os.Stat(input.ImagePath); os.IsNotExist(err) {
		return nil, VeoGenerationOutput{}, fmt.Errorf("image file not found: %s", input.ImagePath)
	}

	// Set defaults
	aspectRatio := input.AspectRatio
	if aspectRatio == "" {
		aspectRatio = "16:9"
	}

	resolution := input.Resolution
	if resolution == "" {
		resolution = "720p"
	}

	model := input.Model
	if model == "" {
		model = "veo-3.0-generate-001"
	}

	log.Printf("Generating image-to-video with model %s for image: %s, prompt: %s (aspect: %s, resolution: %s)",
		model, input.ImagePath, input.Prompt, aspectRatio, resolution)

	timestamp := time.Now().Format("20060102_150405")

	// Read the input image file
	imageData, err := os.ReadFile(input.ImagePath)
	if err != nil {
		return nil, VeoGenerationOutput{}, fmt.Errorf("failed to read input image: %v", err)
	}

	// Create Image object from the file data
	inputImage := &genai.Image{
		ImageBytes: imageData,
	}

	// Build prompt with negative prompt if specified
	promptText := input.Prompt
	if input.NegativePrompt != "" {
		promptText = fmt.Sprintf("%s. Avoid: %s", input.Prompt, input.NegativePrompt)
	}

	// Generate video using Gemini API - image-to-video
	operation, err := s.client.Models.GenerateVideos(
		ctx,
		model,
		promptText,
		inputImage, // Pass the processed image
		nil,        // Use default config
	)
	if err != nil {
		return nil, VeoGenerationOutput{}, fmt.Errorf("error starting image-to-video generation: %v", err)
	}

	operationID := operation.Name
	log.Printf("Image-to-video generation started with operation ID: %s", operationID)

	// Poll operation status until completion
	maxAttempts := 60 // 10 minutes max
	for i := 0; i < maxAttempts && !operation.Done; i++ {
		log.Printf("Waiting for image-to-video generation to complete... (attempt %d/%d)", i+1, maxAttempts)
		time.Sleep(10 * time.Second)
		operation, err = s.client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			log.Printf("Error checking operation status: %v", err)
			break
		}
	}

	var savedFiles []string
	var videoURL string
	status := "generating"

	if operation.Done {
		if operation.Error != nil {
			status = "failed"
			log.Printf("Image-to-video generation failed: %v", operation.Error)
		} else if len(operation.Response.GeneratedVideos) > 0 {
			status = "completed"
			video := operation.Response.GeneratedVideos[0]
			log.Printf("Image-to-video generation completed successfully")

			// Download and save video if output directory is specified
			if input.OutputDirectory != "" {
				if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
					filename := fmt.Sprintf("veo_image_to_video_%s.mp4", timestamp)
					outputPath := filepath.Join(input.OutputDirectory, filename)

					// Download the video file following official documentation pattern
					s.client.Files.Download(ctx, video.Video, nil)

					// Save the video bytes to file
					err = os.WriteFile(outputPath, video.Video.VideoBytes, 0644)
					if err != nil {
						log.Printf("Error saving video file: %v", err)
					} else {
						savedFiles = append(savedFiles, outputPath)
						videoURL = outputPath
						log.Printf("Image-to-video saved to: %s", outputPath)
					}
				}
			}
		}
	} else {
		status = "timeout"
		log.Printf("Image-to-video generation timed out after 10 minutes")
	}

	// Save metadata
	metadata := map[string]string{
		"generation_type": "image-to-video",
		"input_image":     input.ImagePath,
		"original_prompt": input.Prompt,
		"negative_prompt": input.NegativePrompt,
		"operation_id":    operationID,
	}

	if input.Seed > 0 {
		metadata["seed"] = fmt.Sprintf("%d", input.Seed)
	}

	if input.OutputDirectory != "" {
		if err := os.MkdirAll(input.OutputDirectory, 0755); err == nil {
			filename := fmt.Sprintf("veo_image_to_video_metadata_%s.json", timestamp)
			outputPath := filepath.Join(input.OutputDirectory, filename)

			metadataContent := map[string]interface{}{
				"generation_type":  "image-to-video",
				"model":            model,
				"input_image":      input.ImagePath,
				"prompt":           input.Prompt,
				"negative_prompt":  input.NegativePrompt,
				"aspect_ratio":     aspectRatio,
				"resolution":       resolution,
				"seed":             input.Seed,
				"operation_id":     operationID,
				"video_url":        videoURL,
				"status":           status,
				"generated_at":     timestamp,
				"estimated_length": "8 seconds",
			}

			if jsonData, err := json.MarshalIndent(metadataContent, "", "  "); err == nil {
				if err := os.WriteFile(outputPath, jsonData, 0644); err == nil {
					savedFiles = append(savedFiles, outputPath)
				}
			}
		}
	}

	return nil, VeoGenerationOutput{
		OperationID:     operationID,
		Status:          status,
		VideoURL:        videoURL,
		SavedFiles:      savedFiles,
		Model:           model,
		AspectRatio:     aspectRatio,
		Resolution:      resolution,
		Metadata:        metadata,
		GeneratedAt:     timestamp,
		EstimatedLength: "8 seconds",
	}, nil
}
