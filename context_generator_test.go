package main

import (
	"os"
	"strings"
	"testing"
)

func TestLLMContextGenerator(t *testing.T) {
	// Define the arguments for the test
	sourceDir := "."
	outputFile := "test_output.txt"
	patterns := "go,md"
	excludes := "context_generator_test.go"

	// Prepare the command line arguments
	os.Args = []string{"cmd", "-s", sourceDir, "-o", outputFile, "-p", patterns, "-e", excludes}

	// Run the main function
	main()

	// Check if the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file %s was not created", outputFile)
	}

	// Read the output file
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file %s: %v", outputFile, err)
	}

	// Check if context_generator.go is included in the output
	if !strings.Contains(string(outputContent), "context_generator.go") {
		t.Errorf("context_generator.go was not included in the output")
	}

	// Check if context_generator_test.go is excluded from the output
	if strings.Contains(string(outputContent), "context_generator_test.go") {
		t.Errorf("context_generator_test.go was not excluded from the output")
	}

	// Check if README.md is included in the output
	if !strings.Contains(string(outputContent), "README.md") {
		t.Errorf("README.md was not included in the output")
	}

	// Clean up the output file after the test
	if err := os.Remove(outputFile); err != nil {
		t.Fatalf("Failed to remove output file %s: %v", outputFile, err)
	}
}
