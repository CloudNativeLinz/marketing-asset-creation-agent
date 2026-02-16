// Command generate is a CLI tool that generates/edits images using the
// Azure OpenAI image-edits endpoint.
//
// Usage:
//
//	go run ./cmd/generate -p PROMPT -i IMAGE [-b BACKGROUND] [-s SIZE]
//
// Required:
//
//	-p  The text prompt for image generation/editing
//	-i  The input (foreground) image file
//
// Optional:
//
//	-b  Background image file. When provided, both images are sent to the
//	    API so the model can composite the foreground onto the background.
//	-s  Image size (default: 1024x1024)
//
// The output file is auto-generated from the foreground image name with
// "_generated" appended.
// Example: input "assets/cat.png" → output "assets/cat_generated.png"
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Azure resource configuration – keep in sync with the shell script.
const (
	resourceHost = "cloudnativelinz-poland-resource.openai.azure.com"
	deployment   = "gpt-image-1.5"
	apiVersion   = "2025-04-01-preview"
)

// imageEditResponse represents the JSON payload returned by the Azure
// OpenAI /images/edits endpoint.
type imageEditResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func main() {
	prompt := flag.String("p", "", "The text prompt for image generation/editing (required)")
	inputImage := flag.String("i", "", "The foreground input image file (required)")
	bgImage := flag.String("b", "", "Background image file (optional)")
	size := flag.String("s", "1024x1024", "Image size")
	flag.Parse()

	if *prompt == "" {
		fatalf("Error: Prompt is required (-p)\nUsage: generate -p PROMPT -i IMAGE [-b BACKGROUND] [-s SIZE]")
	}
	if *inputImage == "" {
		fatalf("Error: Input image is required (-i)\nUsage: generate -p PROMPT -i IMAGE [-b BACKGROUND] [-s SIZE]")
	}
	if _, err := os.Stat(*inputImage); os.IsNotExist(err) {
		fatalf("Error: Input image not found: %s", *inputImage)
	}
	if *bgImage != "" {
		if _, err := os.Stat(*bgImage); os.IsNotExist(err) {
			fatalf("Error: Background image not found: %s", *bgImage)
		}
	}

	outputFile := deriveOutputPath(*inputImage)

	fmt.Println("Generating image with Azure OpenAI...")
	fmt.Printf("Prompt:  %s\n", *prompt)
	if *bgImage != "" {
		fmt.Printf("Background: %s\n", *bgImage)
	}
	fmt.Printf("Foreground: %s\n", *inputImage)
	fmt.Printf("Output:  %s\n", outputFile)
	fmt.Printf("Size:    %s\n\n", *size)

	// Ensure the output directory exists.
	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		fatalf("Error creating output directory: %v", err)
	}

	token, err := getAzureToken()
	if err != nil {
		fatalf("Error obtaining Azure access token: %v", err)
	}

	endpoint := fmt.Sprintf("https://%s/openai/deployments/%s/images/edits?api-version=%s",
		resourceHost, deployment, apiVersion)

	fmt.Printf("Using image edits endpoint (api-version: %s)...\n", apiVersion)

	// Collect image paths: background first (if provided), then foreground.
	images := []string{}
	if *bgImage != "" {
		images = append(images, *bgImage)
	}
	images = append(images, *inputImage)

	body, contentType, err := buildMultipartBody(images, *prompt, *size)
	if err != nil {
		fatalf("Error building request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		fatalf("Error creating HTTP request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fatalf("Error reading response: %v", err)
	}

	var result imageEditResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		fatalf("Error parsing response JSON: %v\nRaw (first 500 bytes): %s", err, truncate(string(respBytes), 500))
	}

	if result.Error != nil {
		fatalf("Image edit failed: %s (type=%s, code=%s)", result.Error.Message, result.Error.Type, result.Error.Code)
	}

	if len(result.Data) == 0 || result.Data[0].B64JSON == "" {
		fatalf("Image edit failed: no image data in response")
	}

	imageBytes, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
	if err != nil {
		fatalf("Error decoding base64 image: %v", err)
	}

	if err := os.WriteFile(outputFile, imageBytes, 0o644); err != nil {
		fatalf("Error writing output file: %v", err)
	}

	fmt.Println("✅ Image edit successful")
	fmt.Printf("Image saved to %s\n", outputFile)
	fmt.Printf("Size: %.2f MB\n", float64(len(imageBytes))/(1024*1024))
}

// deriveOutputPath returns the output file path by appending "_generated"
// before the file extension.
// Example: "assets/cat.png" → "assets/cat_generated.png"
func deriveOutputPath(input string) string {
	dir := filepath.Dir(input)
	ext := filepath.Ext(input)
	name := strings.TrimSuffix(filepath.Base(input), ext)
	return filepath.Join(dir, name+"_generated"+ext)
}

// getAzureToken shells out to `az account get-access-token` to retrieve a
// bearer token for the Cognitive Services resource.
func getAzureToken() (string, error) {
	cmd := exec.Command("az", "account", "get-access-token",
		"--resource", "https://cognitiveservices.azure.com",
		"--query", "accessToken",
		"-o", "tsv",
	)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("az CLI failed: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// buildMultipartBody creates the multipart/form-data payload expected by the
// Azure OpenAI /images/edits endpoint. Multiple images are sent as
// repeated "image[]" fields.
func buildMultipartBody(imagePaths []string, prompt, size string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Add each image file. Use "image[]" when there are multiple images so
	// the API receives them as an array; use "image" for a single file to
	// stay compatible with the original behaviour.
	fieldName := "image"
	if len(imagePaths) > 1 {
		fieldName = "image[]"
	}

	for _, imagePath := range imagePaths {
		if err := addImagePart(w, fieldName, imagePath); err != nil {
			return nil, "", err
		}
	}

	// Add the remaining text fields.
	for k, v := range map[string]string{
		"prompt": prompt,
		"n":      "1",
		"size":   size,
	} {
		if err := w.WriteField(k, v); err != nil {
			return nil, "", fmt.Errorf("writing field %s: %w", k, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("closing multipart writer: %w", err)
	}

	return &buf, w.FormDataContentType(), nil
}

// addImagePart adds a single image file as a multipart form part with the
// correct MIME type.
func addImagePart(w *multipart.Writer, fieldName, imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("opening image %s: %w", imagePath, err)
	}
	defer file.Close()

	// Use the correct MIME type so the API doesn't reject with
	// "unsupported mimetype ('application/octet-stream')".
	mimeType := mime.TypeByExtension(filepath.Ext(imagePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filepath.Base(imagePath)))
	partHeader.Set("Content-Type", mimeType)
	part, err := w.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("creating form file for %s: %w", imagePath, err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copying image data for %s: %w", imagePath, err)
	}
	return nil
}

// truncate returns at most n bytes of s.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// fatalf prints a message to stderr and exits with code 1.
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
	os.Exit(1)
}
