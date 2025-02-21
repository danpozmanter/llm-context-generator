package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// --- Tests for compilePatterns ---

func TestCompilePatterns_Empty(t *testing.T) {
	pats, err := compilePatterns("")
	if err != nil {
		t.Fatalf("Expected no error for empty patterns, got: %v", err)
	}
	if len(pats) != 0 {
		t.Errorf("Expected 0 patterns, got: %d", len(pats))
	}
}

func TestCompilePatterns_Invalid(t *testing.T) {
	// Use an invalid pattern (unbalanced bracket)
	_, err := compilePatterns("*.go,[invalid")
	if err == nil {
		t.Error("Expected error for invalid pattern, got none")
	}
}

func TestCompilePatterns_SkipEmptyToken(t *testing.T) {
	// Provide a comma with an empty token in between.
	pats, err := compilePatterns("*.go,,*.md")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(pats) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(pats))
	}
}

// --- Tests for isPatternMatched and shouldBeExcluded ---

func TestIsPatternMatched_NoPatterns(t *testing.T) {
	// No patterns means match everything.
	if !isPatternMatched("anyfile.txt", []*regexp.Regexp{}) {
		t.Error("Expected match when no patterns are provided")
	}
}

func TestIsPatternMatched_WithPatterns(t *testing.T) {
	cases := []struct {
		pattern  string
		filename string
		expected bool
	}{
		{"*.go", "file.go", true},
		{"*.go", "file.txt", false},
		{"*.go,*.md", "README.md", true},
		{"test_*.go", "test_main.go", true},
		{"test_*.go", "main_test.go", false},
	}
	for _, tc := range cases {
		pats, err := compilePatterns(tc.pattern)
		if err != nil {
			t.Fatalf("Error compiling pattern %q: %v", tc.pattern, err)
		}
		res := isPatternMatched(tc.filename, pats)
		if res != tc.expected {
			t.Errorf("For pattern %q on %q, expected %v, got %v", tc.pattern, tc.filename, tc.expected, res)
		}
	}
}

func TestShouldBeExcluded(t *testing.T) {
	cases := []struct {
		exclude  string
		path     string
		expected bool
	}{
		{"*_test.go", "main_test.go", true},
		{"*_test.go", "main.go", false},
		{"vendor/*", "vendor/pkg/file.go", true},
		{"vendor/*", "src/pkg/file.go", false},
	}
	for _, tc := range cases {
		excludes, err := compilePatterns(tc.exclude)
		if err != nil {
			t.Fatalf("Error compiling exclude pattern %q: %v", tc.exclude, err)
		}
		res := shouldBeExcluded(tc.path, excludes)
		if res != tc.expected {
			t.Errorf("For exclude %q on %q, expected %v, got %v", tc.exclude, tc.path, tc.expected, res)
		}
	}
}

// --- Tests for walkDir ---

func TestWalkDir_Filters(t *testing.T) {
	tempDir := t.TempDir()
	// Create one matching and one non-matching file.
	matchFile := filepath.Join(tempDir, "file.go")
	nonMatchFile := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(matchFile, []byte("match"), 0644); err != nil {
		t.Fatalf("Failed to write matching file: %v", err)
	}
	if err := os.WriteFile(nonMatchFile, []byte("no match"), 0644); err != nil {
		t.Fatalf("Failed to write non-matching file: %v", err)
	}
	pats, err := compilePatterns("*.go")
	if err != nil {
		t.Fatalf("Error compiling patterns: %v", err)
	}
	files, err := walkDir(tempDir, pats, []*regexp.Regexp{})
	if err != nil {
		t.Fatalf("Error walking directory: %v", err)
	}
	if len(files) != 1 || !strings.HasSuffix(files[0], "file.go") {
		t.Errorf("Expected only file.go, got: %v", files)
	}
}

func TestWalkDir_Excludes(t *testing.T) {
	tempDir := t.TempDir()
	incFile := filepath.Join(tempDir, "include.go")
	excFile := filepath.Join(tempDir, "exclude_test.go")
	if err := os.WriteFile(incFile, []byte("include"), 0644); err != nil {
		t.Fatalf("Failed to write include file: %v", err)
	}
	if err := os.WriteFile(excFile, []byte("exclude"), 0644); err != nil {
		t.Fatalf("Failed to write exclude file: %v", err)
	}
	pats, _ := compilePatterns("*.go")
	exc, _ := compilePatterns("*_test.go")
	files, err := walkDir(tempDir, pats, exc)
	if err != nil {
		t.Fatalf("Error walking directory: %v", err)
	}
	if len(files) != 1 || !strings.HasSuffix(files[0], "include.go") {
		t.Errorf("Expected only include.go, got: %v", files)
	}
}

func TestWalkDir_NonExistent(t *testing.T) {
	_, err := walkDir("non_existent_dir", []*regexp.Regexp{}, []*regexp.Regexp{})
	if err == nil {
		t.Error("Expected error for non-existent directory, got none")
	}
}

// --- Tests for generateOutputString ---

func TestGenerateOutputString_Valid(t *testing.T) {
	tempDir := t.TempDir()
	fName := "test.txt"
	fPath := filepath.Join(tempDir, fName)
	content := "hello world"
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("Error writing temp file: %v", err)
	}
	out, err := generateOutputString([]string{fPath})
	if err != nil {
		t.Fatalf("Error generating output: %v", err)
	}
	expected := fmt.Sprintf("=%s=\n%s\n=/%s=\n", fName, content, fName)
	if out != expected {
		t.Errorf("Output mismatch.\nGot:\n%s\nWant:\n%s", out, expected)
	}
}

func TestGenerateOutputString_FileError(t *testing.T) {
	_, err := generateOutputString([]string{"non_existent.txt"})
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

// --- Tests for writeOutput ---

func TestWriteOutput_File(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "out.txt")
	testStr := "file output"
	if err := writeOutput(testStr, outPath, false, false, nil); err != nil {
		t.Fatalf("writeOutput returned error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if string(data) != testStr {
		t.Errorf("Expected %q, got %q", testStr, string(data))
	}
}

func TestWriteOutput_Console(t *testing.T) {
	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	testStr := "console output"
	if err := writeOutput(testStr, "", true, false, nil); err != nil {
		t.Fatalf("Error writing to console: %v", err)
	}
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), testStr) {
		t.Errorf("Expected console output to contain %q, got %q", testStr, string(out))
	}
}

func TestWriteOutput_ClipboardError(t *testing.T) {
	// Simulate clipboard error.
	clipErr := func(s string) error { return fmt.Errorf("clipboard failed") }
	err := writeOutput("dummy", "", false, true, clipErr)
	if err == nil || !strings.Contains(err.Error(), "clipboard failed") {
		t.Errorf("Expected clipboard error, got: %v", err)
	}
}

// --- Tests for printPatterns ---

func TestPrintPatterns_NonEmpty(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	pats, _ := compilePatterns("*.go,*.md")
	printPatterns("TestPrefix", pats)
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "TestPrefix:") {
		t.Errorf("Expected output to contain 'TestPrefix:', got %q", string(out))
	}
}

func TestPrintPatterns_Empty(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printPatterns("Empty", []*regexp.Regexp{})
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "Empty:") {
		t.Errorf("Expected output to contain 'Empty:', got %q", string(out))
	}
}

// --- Tests for parseArgs ---

func TestParseArgs_MissingPatterns(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"cmd", "-s", "."} // Missing -p flag.
	_, err := parseArgs()
	if err == nil {
		t.Error("Expected error when -p flag is missing")
	}
}

func TestParseArgs_WithOutputFile(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"cmd", "-s", ".", "-p", "*.go", "-o", "dummy.txt"}
	cfg, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.OutputFile != "dummy.txt" || cfg.Console || cfg.Clipboard {
		t.Errorf("Parsed config mismatch: %+v", cfg)
	}
}

func TestParseArgs_WithConsoleFlag(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"cmd", "-s", ".", "-p", "*.go", "-c"}
	cfg, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !cfg.Console || cfg.Clipboard {
		t.Errorf("Parsed config mismatch: %+v", cfg)
	}
}

// --- Tests for main() ---

func TestMain_FileOutput(t *testing.T) {
	// Create a temporary file in a temp directory.
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "main_out.txt")
	// Create a dummy file that should be included.
	dummy := filepath.Join(tempDir, "dummy.go")
	if err := os.WriteFile(dummy, []byte("package dummy"), 0644); err != nil {
		t.Fatalf("Error writing dummy file: %v", err)
	}
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"cmd", "-s", tempDir, "-o", outPath, "-p", "*.go"}
	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "Output successfully written to:") {
		t.Errorf("Expected success message, got: %q", string(out))
	}
	// Verify file output.
	data, err := os.ReadFile(outPath)
	if err != nil || !strings.Contains(string(data), "dummy.go") {
		t.Errorf("File output does not contain expected dummy file, err: %v", err)
	}
}

func TestMain_ConsoleOutput(t *testing.T) {
	// Create a temporary directory with a dummy file.
	tempDir := t.TempDir()
	dummy := filepath.Join(tempDir, "dummy.go")
	if err := os.WriteFile(dummy, []byte("package dummy"), 0644); err != nil {
		t.Fatalf("Error writing dummy file: %v", err)
	}
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"cmd", "-s", tempDir, "-p", "*.go", "-c"}
	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "Output successfully written to: console") {
		t.Errorf("Expected console target message, got: %q", string(out))
	}
}
