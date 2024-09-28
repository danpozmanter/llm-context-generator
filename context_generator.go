package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
)

// Config struct to hold command line arguments
type Config struct {
	SourceDir  string
	OutputFile string
	Patterns   map[string]struct{}
	Excludes   map[string]struct{}
	Clipboard  bool
}

// parseArgs parses the command line arguments
func parseArgs() Config {
	sourceDir := flag.String("s", "", "Source directory to scan")
	outputFile := flag.String("o", "", "Output file path (use -c to copy to clipboard)")
	patterns := flag.String("p", "", "File extension patterns separated by ','")
	excludes := flag.String("e", "", "File path patterns to exclude, separated by ','")
	clipboardFlag := flag.Bool("c", false, "Copy output to clipboard instead of saving to a file")
	flag.Parse()

	if *sourceDir == "" || *patterns == "" || (*outputFile == "" && !*clipboardFlag) {
		fmt.Println("Usage: -s <source directory> -o <output file> -p <patterns> -e <excludes> [-c to copy to clipboard]")
		os.Exit(1)
	}

	toMap := func(list string) map[string]struct{} {
		m := make(map[string]struct{})
		for _, item := range strings.Split(list, ",") {
			if item != "" {
				m[item] = struct{}{}
			}
		}
		return m
	}

	return Config{
		SourceDir:  *sourceDir,
		OutputFile: *outputFile,
		Patterns:   toMap(*patterns),
		Excludes:   toMap(*excludes),
		Clipboard:  *clipboardFlag,
	}
}

// walkDir recursively walks through the source directory and collects matching files
func walkDir(sourceDir string, patterns, excludes map[string]struct{}) ([]string, error) {
	var files []string
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if shouldBeExcluded(path, excludes) || !isPatternMatched(info.Name(), patterns) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// Helper functions for exclusion and pattern matching
func shouldBeExcluded(path string, excludes map[string]struct{}) bool {
	for exclude := range excludes {
		if strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}

func isPatternMatched(fileName string, patterns map[string]struct{}) bool {
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	_, matched := patterns[ext]
	return matched
}

// generateOutputString creates the output string from the collected files
func generateOutputString(files []string) (string, error) {
	var sb strings.Builder
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("=%s=\n%s\n=/%s=\n", filepath.Base(file), string(content), filepath.Base(file)))
	}
	return sb.String(), nil
}

// writeOutput handles both file and clipboard output
func writeOutput(outputFile string, outputString string, toClipboard bool) error {
	if toClipboard {
		return clipboard.WriteAll(outputString)
	}
	return os.WriteFile(outputFile, []byte(outputString), 0644)
}

// printPatterns prints patterns from a map with a given message
func printPatterns(message string, patterns map[string]struct{}) {
	fmt.Printf("%s: ", message)
	for pattern := range patterns {
		fmt.Printf("%s ", pattern)
	}
	fmt.Println()
}

func main() {
	config := parseArgs()

	// Echo back the choices by directly iterating over the map keys
	fmt.Printf("Scanning source directory %s\n", config.SourceDir)
	printPatterns("Matching patterns", config.Patterns)
	if len(config.Excludes) > 0 {
		printPatterns("Excluding patterns", config.Excludes)
	}

	files, err := walkDir(config.SourceDir, config.Patterns, config.Excludes)
	if err != nil {
		fmt.Println("Error walking the directory:", err)
		os.Exit(1)
	}

	outputString, err := generateOutputString(files)
	if err != nil {
		fmt.Println("Error generating output string:", err)
		os.Exit(1)
	}

	if err := writeOutput(config.OutputFile, outputString, config.Clipboard); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	target := "clipboard"
	if !config.Clipboard {
		target = config.OutputFile
	}
	fmt.Printf("Output successfully written to: %s\n", target)
}
