package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

// Config struct to hold command line arguments
type Config struct {
	SourceDir  string
	OutputFile string
	Console    bool
	Patterns   []*regexp.Regexp
	Excludes   []*regexp.Regexp
	Clipboard  bool
}

type ClipboardWriter func(string) error

// Helper function to convert comma-separated patterns into regex slices
func compilePatterns(patterns string) ([]*regexp.Regexp, error) {
	var result []*regexp.Regexp
	if patterns == "" {
		return result, nil
	}

	for _, pattern := range strings.Split(patterns, ",") {
		if pattern == "" {
			continue
		}
		// Convert glob patterns to regex
		// Replace * with .* and escape dots
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = "^" + pattern + "$" // Ensure full match

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %v", pattern, err)
		}
		result = append(result, re)
	}
	return result, nil
}

// parseArgs parses the command line arguments
func parseArgs() (Config, error) {
	sourceDir := flag.String("s", ".", "Source directory to scan (default: current directory)")
	outputFile := flag.String("o", "", "Output file path (if specified, output will not go to the clipboard or console)")
	patterns := flag.String("p", "", "File extension patterns separated by ',' (e.g., '*.go,*.md')")
	excludes := flag.String("e", "", "File path patterns to exclude, separated by ',' (e.g., '*.test.go,vendor/*')")
	consoleFlag := flag.Bool("c", false, "Write output to the console instead of the clipboard (default: false)")
	clipboardFlag := true

	flag.Parse()

	// Check for required arguments
	if *patterns == "" {
		return Config{}, fmt.Errorf("missing required flag: -p <patterns>")
	}

	compiledPatterns, err := compilePatterns(*patterns)
	if err != nil {
		return Config{}, err
	}

	compiledExcludes, err := compilePatterns(*excludes)
	if err != nil {
		return Config{}, err
	}

	if *outputFile != "" || *consoleFlag {
		clipboardFlag = false
	}

	return Config{
		SourceDir:  *sourceDir,
		OutputFile: *outputFile,
		Patterns:   compiledPatterns,
		Excludes:   compiledExcludes,
		Console:    *consoleFlag,
		Clipboard:  clipboardFlag,
	}, nil
}

// walkDir recursively walks through the source directory and collects matching files
func walkDir(sourceDir string, patterns []*regexp.Regexp, excludes []*regexp.Regexp) ([]string, error) {
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
func shouldBeExcluded(path string, excludes []*regexp.Regexp) bool {
	for _, exclude := range excludes {
		if exclude.MatchString(path) {
			return true
		}
	}
	return false
}

func isPatternMatched(fileName string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(fileName) {
			return true
		}
	}
	return len(patterns) == 0 // If no patterns specified, match all
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
func writeOutput(outputString string, outputFile string, toConsole bool, toClipboard bool, writeToClipboard ClipboardWriter) error {
	if toConsole {
		fmt.Println(outputString)
		return nil
	}
	if toClipboard {
		return writeToClipboard(outputString)
	}
	return os.WriteFile(outputFile, []byte(outputString), 0644)
}

// printPatterns prints patterns from a slice with a given message
func printPatterns(message string, patterns []*regexp.Regexp) {
	patternStrs := make([]string, len(patterns))
	for i, p := range patterns {
		patternStrs[i] = p.String()
	}
	fmt.Printf("%s: %s\n", message, strings.Join(patternStrs, ", "))
}

func main() {
	config, err := parseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		fmt.Println("Usage: -s <source directory> -o <output file> -p <patterns> -e <excludes> [-c to write to console]")
		os.Exit(1)
	}

	// Echo back the choices
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

	if err := writeOutput(outputString, config.OutputFile, config.Console, config.Clipboard, clipboard.WriteAll); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	target := "clipboard"
	switch {
	case config.Console:
		target = "console"
	case config.OutputFile != "":
		target = config.OutputFile
	}
	fmt.Printf("Output successfully written to: %s\n", target)
}
