/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package llm_test

import (
	"ares_api/pkg/llm"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFileAccessTools_ReadFile verifies file reading with security
func TestFileAccessTools_ReadFile(t *testing.T) {
	// Create temporary workspace for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Write test file
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create file tools with allowed path
	ft := &llm.FileAccessTools{
		WorkspaceRoot: tempDir,
		AllowedPaths:  []string{tempDir},
	}

	ctx := context.Background()

	// Test reading file
	result, err := ft.ReadFile(ctx, testFile, 0)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if result.Content != content {
		t.Errorf("Content mismatch: expected %q, got %q", content, result.Content)
	}

	if result.Lines != 5 {
		t.Errorf("Expected 5 lines, got %d", result.Lines)
	}

	t.Logf("✅ ReadFile: %d lines, %d KB", result.Lines, result.SizeKB)
}

// TestFileAccessTools_Security verifies path restrictions
func TestFileAccessTools_Security(t *testing.T) {
	tempDir := t.TempDir()
	
	ft := &llm.FileAccessTools{
		WorkspaceRoot: tempDir,
		AllowedPaths:  []string{tempDir},
	}

	ctx := context.Background()

	// Try to read outside allowed path
	forbiddenPath := "C:/Windows/System32/config/SAM"
	result, err := ft.ReadFile(ctx, forbiddenPath, 0)
	
	if err == nil {
		t.Error("Expected security error for forbidden path")
	}

	if result.Error != "Path not allowed - outside workspace" {
		t.Errorf("Expected security error message, got: %s", result.Error)
	}

	t.Logf("✅ Security check passed: %v", err)
}

// TestFileAccessTools_ListDirectory verifies directory listing
func TestFileAccessTools_ListDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test structure
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.go"), []byte("package main"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subfolder"), 0755)

	ft := &llm.FileAccessTools{
		WorkspaceRoot: tempDir,
		AllowedPaths:  []string{tempDir},
	}

	ctx := context.Background()

	result, err := ft.ListDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ListDirectory failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d: %v", len(result.Files), result.Files)
	}

	if len(result.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d: %v", len(result.Folders), result.Folders)
	}

	t.Logf("✅ ListDirectory: %d files, %d folders", len(result.Files), len(result.Folders))
}

// TestFileAccessTools_SearchCode verifies code search
func TestFileAccessTools_SearchCode(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files
	goFile := filepath.Join(tempDir, "main.go")
	txtFile := filepath.Join(tempDir, "readme.txt")
	
	os.WriteFile(goFile, []byte("package main\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"), 0644)
	os.WriteFile(txtFile, []byte("This is a readme\nHello World"), 0644)

	ft := &llm.FileAccessTools{
		WorkspaceRoot: tempDir,
		AllowedPaths:  []string{tempDir},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Search for "Hello" in .go files only
	results, err := ft.SearchCode(ctx, "Hello", tempDir, []string{".go"}, 10)
	if err != nil {
		t.Errorf("SearchCode failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result (.go file), got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].Line != 3 {
			t.Errorf("Expected line 3, got line %d", results[0].Line)
		}
		t.Logf("✅ SearchCode: Found '%s' at %s:%d", results[0].Match, results[0].Path, results[0].Line)
	}
}

// TestFileAccessTools_MaxLines verifies line limiting
func TestFileAccessTools_MaxLines(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.txt")
	
	// Create file with 100 lines
	content := ""
	for i := 1; i <= 100; i++ {
		content += "Line " + string(rune(i)) + "\n"
	}
	os.WriteFile(testFile, []byte(content), 0644)

	ft := &llm.FileAccessTools{
		WorkspaceRoot: tempDir,
		AllowedPaths:  []string{tempDir},
	}

	ctx := context.Background()

	// Read only first 10 lines
	result, err := ft.ReadFile(ctx, testFile, 10)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if result.Lines != 10 {
		t.Errorf("Expected 10 lines (truncated), got %d", result.Lines)
	}

	t.Logf("✅ MaxLines: Truncated to %d lines", result.Lines)
}
