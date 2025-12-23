package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// UploadResult is the JSON response from the server
type UploadResult struct {
	ObjectKey   string `json:"object_key"`
	DownloadURL string `json:"download_url,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	MIMEType    string `json:"mime_type"`
	Size        int64  `json:"size"`
	Message     string `json:"message"`
	UploadedAt  string `json:"uploaded_at"`
}

// ErrorResult is the JSON response for errors
type ErrorResult struct {
	Error string `json:"error"`
}

func main() {
	// Define flags
	serverURL := flag.String("server", "", "Server upload URL (e.g., http://localhost:8080/upload)")
	token := flag.String("token", "", "One-time authentication token")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("upload_media version %s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Git commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	// Get file path from remaining arguments
	args := flag.Args()
	if len(args) != 1 {
		printUsage()
		os.Exit(1)
	}

	filePath := args[0]

	// Validate required flags
	if *serverURL == "" {
		outputError("--server is required")
		os.Exit(1)
	}
	if *token == "" {
		outputError("--token is required")
		os.Exit(1)
	}

	// Validate absolute path
	if !filepath.IsAbs(filePath) {
		outputError("File path must be absolute: " + filePath)
		os.Exit(1)
	}

	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			outputError("File not found: " + filePath)
		} else {
			outputError(fmt.Sprintf("Cannot access file: %v", err))
		}
		os.Exit(1)
	}

	if info.IsDir() {
		outputError("Path is a directory, not a file: " + filePath)
		os.Exit(1)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		outputError(fmt.Sprintf("Failed to open file: %v", err))
		os.Exit(1)
	}
	defer file.Close()

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file to form
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		outputError(fmt.Sprintf("Failed to create form file: %v", err))
		os.Exit(1)
	}

	if _, err := io.Copy(part, file); err != nil {
		outputError(fmt.Sprintf("Failed to copy file data: %v", err))
		os.Exit(1)
	}

	writer.Close()

	// Create HTTP request
	req, err := http.NewRequest("POST", *serverURL, &body)
	if err != nil {
		outputError(fmt.Sprintf("Failed to create request: %v", err))
		os.Exit(1)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+*token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		outputError(fmt.Sprintf("Failed to send request: %v", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		outputError(fmt.Sprintf("Failed to read response: %v", err))
		os.Exit(1)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResult ErrorResult
		if json.Unmarshal(respBody, &errResult) == nil && errResult.Error != "" {
			outputError(errResult.Error)
		} else {
			outputError(fmt.Sprintf("Server returned status %d: %s", resp.StatusCode, string(respBody)))
		}
		os.Exit(1)
	}

	// Output server response (already JSON formatted)
	fmt.Println(string(respBody))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `upload_media - Upload files to gemini-mcp server

Usage:
  upload_media --server <url> --token <token> <file_path>
  upload_media -version
  upload_media -help

Required Flags:
  --server    Server upload URL (e.g., "http://localhost:8080/upload")
  --token     One-time authentication token (provided by upload_media MCP tool)

Arguments:
  <file_path>    Absolute path to the file to upload

Examples:
  upload_media --server "http://localhost:8080/upload" --token "abc123..." /Users/example/photo.png

Output:
  JSON object with object_key, download_url, mime_type, size, etc.
  Use the object_key with gemini_image_edit, gemini_multi_image, or veo_image_to_video.

Note:
  The token is ONE-TIME USE only. Get a new token by calling the upload_media MCP tool.
`)
}

func outputError(msg string) {
	result := ErrorResult{Error: msg}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(os.Stderr, string(jsonBytes))
}
