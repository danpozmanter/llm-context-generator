package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config struct to hold command line arguments
type Config struct {
	SourceDir  string
	OutputFile string
	Patterns   []string
	Excludes   []string
}

// parseArgs parses the command line arguments
func parseArgs() Config {
	sourceDir := flag.String("s", "", "Source directory to scan")
	outputFile := flag.String("o", "", "Output file path")
	patterns := flag.String("p", "", "File extension patterns separated by ','")
	excludes := flag.String("e", "", "File path patterns to exclude, separated by ','")
	flag.Parse()

	if *sourceDir == "" || *outputFile == "" || *patterns == "" {
		fmt.Println("Usage: -s <source directory> -o <output file> -p <patterns> -e <excludes>")
		os.Exit(1)
	}

	excludeList := []string{}
	if *excludes != "" {
		excludeList = strings.Split(*excludes, ",")
	}

	return Config{
		SourceDir:  *sourceDir,
		OutputFile: *outputFile,
		Patterns:   strings.Split(*patterns, ","),
		Excludes:   excludeList,
	}
}

// walkDir recursively walks through the source directory and collects matching files
func walkDir(sourceDir string, patterns []string, excludes []string) ([]string, error) {
	var files []string

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, exclude := range excludes {
			if strings.Contains(path, exclude) {
				return nil
			}
		}

		for _, pattern := range patterns {
			if strings.HasSuffix(info.Name(), "."+pattern) {
				files = append(files, path)
				break
			}
		}
		return nil
	})

	return files, err
}

// generateOutputString creates the output string from the collected files
func generateOutputString(files []string) (string, error) {
	var outputString strings.Builder

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		filename := filepath.Base(file)
		outputString.WriteString(fmt.Sprintf("=%s=\n%s\n=/%s=\n", filename, string(content), filename))
	}

	return outputString.String(), nil
}

// writeOutputToFile writes the generated output string to the specified output file
func writeOutputToFile(outputFile string, outputString string) error {
	return os.WriteFile(outputFile, []byte(outputString), 0644)
}

func main() {
	// Parse command line arguments
	config := parseArgs()

	// Echo back the choices
	fmt.Printf("Scanning source directory %s\n", config.SourceDir)
	fmt.Printf("Matching patterns %s\n", strings.Join(config.Patterns, ", "))
	if len(config.Excludes) > 0 {
		fmt.Printf("Excluding patterns %s\n", strings.Join(config.Excludes, ", "))
	}

	// Walk through the source directory
	files, err := walkDir(config.SourceDir, config.Patterns, config.Excludes)
	if err != nil {
		fmt.Println("Error walking the directory:", err)
		os.Exit(1)
	}

	// Generate the output string
	outputString, err := generateOutputString(files)
	if err != nil {
		fmt.Println("Error generating output string:", err)
		os.Exit(1)
	}

	// Write the output string to the output file
	err = writeOutputToFile(config.OutputFile, outputString)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		os.Exit(1)
	}

	fmt.Println("Output written to", config.OutputFile)
}
