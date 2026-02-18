package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestGenerateImage(t *testing.T) {
	// Ensure the input file exists
	inputImage := "assets/thispersondoesnotexist.jpg"
	if _, err := os.Stat(inputImage); os.IsNotExist(err) {
		t.Fatalf("Input image not found: %s", inputImage)
	}

	// Construct the command: go run . -p prompts/speaking-prompt.md -i assets/thispersondoesnotexist.jpg -s 1024x1024
	cmd := exec.Command("go", "run", ".",
		"-p", "prompts/speaking-prompt.md",
		"-i", inputImage,
		"-s", "1024x1024",
	)

	// Pipe stdout and stderr to the test output so we can see what happens
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	t.Logf("Running command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify the output file was created
	outputFile := "assets/thispersondoesnotexist_generated.jpg"
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file %s was not created", outputFile)
	} else {
		t.Logf("Successfully generated %s", outputFile)
		// Clean up the generated file after test (optional, commented out for now so you can inspect it)
		// os.Remove(outputFile)
	}
}
