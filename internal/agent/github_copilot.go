package agent

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GitHubCopilot wraps GitHub CLI Copilot commands for FORGE apprenticeship learning
type GitHubCopilot struct {
	Enabled  bool
	TestMode bool // When true, returns mock responses instead of calling gh CLI
}

// NewGitHubCopilot creates a new GitHub Copilot wrapper
func NewGitHubCopilot() *GitHubCopilot {
	// Check if TEST_MODE env var is set
	testMode := os.Getenv("FORGE_TEST_MODE") == "true"
	return &GitHubCopilot{
		Enabled:  true,
		TestMode: testMode,
	}
}

// CopilotRequest represents a code generation request
type CopilotRequest struct {
	Prompt   string
	Language string
	Context  string
	TaskType string // "generate", "explain", "test", "refactor"
}

// CopilotResponse represents GitHub Copilot's response
type CopilotResponse struct {
	Code          string
	Explanation   string
	Model         string
	GeneratedAt   time.Time
	ExecutionTime int64 // milliseconds
	Success       bool
	Error         string
}

// GenerateCode asks GitHub Copilot to generate code
func (gc *GitHubCopilot) GenerateCode(prompt string, language string) (*CopilotResponse, error) {
	if !gc.Enabled {
		return nil, fmt.Errorf("GitHub Copilot is not enabled")
	}

	// TEST MODE: Return mock response without calling gh CLI
	if gc.TestMode {
		return &CopilotResponse{
			Code:          fmt.Sprintf("// Mock generated code for: %s\nfunc MockFunction() {\n    // Implementation here\n}", prompt),
			Model:         "github-copilot-mock",
			GeneratedAt:   time.Now(),
			ExecutionTime: 100, // Mock 100ms
			Success:       true,
		}, nil
	}

	startTime := time.Now()

	// Build the command: gh copilot suggest -t shell "create a PostgreSQL index"
	cmd := exec.Command("gh", "copilot", "suggest", "-t", "shell", prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	executionTime := time.Since(startTime).Milliseconds()

	if err != nil {
		return &CopilotResponse{
			Success:       false,
			Error:         fmt.Sprintf("GitHub CLI error: %v, stderr: %s", err, stderr.String()),
			GeneratedAt:   time.Now(),
			ExecutionTime: executionTime,
		}, err
	}

	return &CopilotResponse{
		Code:          stdout.String(),
		Model:         "github-copilot",
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Success:       true,
	}, nil
} // ExplainCode asks GitHub Copilot to explain code
func (gc *GitHubCopilot) ExplainCode(code string) (*CopilotResponse, error) {
	if !gc.Enabled {
		return nil, fmt.Errorf("GitHub Copilot is not enabled")
	}

	startTime := time.Now()

	// Use gh copilot explain with code as input
	cmd := exec.Command("gh", "copilot", "explain", code)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	executionTime := time.Since(startTime).Milliseconds()

	if err != nil {
		return &CopilotResponse{
			Success:       false,
			Error:         fmt.Sprintf("GitHub CLI error: %v, stderr: %s", err, stderr.String()),
			GeneratedAt:   time.Now(),
			ExecutionTime: executionTime,
		}, err
	}

	return &CopilotResponse{
		Explanation:   stdout.String(),
		Model:         "github-copilot",
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Success:       true,
	}, nil
}

// GenerateSQLScript asks GitHub Copilot to generate SQL scripts
func (gc *GitHubCopilot) GenerateSQLScript(description string) (*CopilotResponse, error) {
	prompt := fmt.Sprintf("Generate PostgreSQL SQL script: %s", description)
	return gc.GenerateCode(prompt, "sql")
}

// GenerateGoCode asks GitHub Copilot to generate Go code
func (gc *GitHubCopilot) GenerateGoCode(description string, context string) (*CopilotResponse, error) {
	prompt := description
	if context != "" {
		prompt = fmt.Sprintf("%s (Context: %s)", description, context)
	}
	return gc.GenerateCode(prompt, "go")
}

// GenerateRESTHandler generates a Go REST API handler
func (gc *GitHubCopilot) GenerateRESTHandler(endpoint string, method string, description string) (*CopilotResponse, error) {
	prompt := fmt.Sprintf("Generate Go REST API handler for %s %s: %s", method, endpoint, description)
	return gc.GenerateCode(prompt, "go")
}

// GenerateDatabaseMigration generates a database migration script
func (gc *GitHubCopilot) GenerateDatabaseMigration(description string) (*CopilotResponse, error) {
	prompt := fmt.Sprintf("Generate PostgreSQL migration script to %s", description)
	return gc.GenerateSQLScript(prompt)
}

// GenerateTestCase generates unit test code
func (gc *GitHubCopilot) GenerateTestCase(functionName string, description string) (*CopilotResponse, error) {
	prompt := fmt.Sprintf("Generate Go unit test for function %s: %s", functionName, description)
	return gc.GenerateCode(prompt, "go")
}

// SuggestRefactoring suggests code refactoring improvements
func (gc *GitHubCopilot) SuggestRefactoring(code string, goal string) (*CopilotResponse, error) {
	prompt := fmt.Sprintf("Suggest refactoring for this code to %s:\n%s", goal, code)
	return gc.GenerateCode(prompt, "go")
}

// ExtractPrinciples analyzes generated code and extracts learning principles for FORGE
func (gc *GitHubCopilot) ExtractPrinciples(code string, taskType string) ([]string, error) {
	principles := []string{}

	// Analyze code structure patterns
	if strings.Contains(code, "CREATE INDEX") {
		principles = append(principles, "Use CREATE INDEX CONCURRENTLY to avoid locking tables")
	}
	if strings.Contains(code, "func") && strings.Contains(code, "http.HandlerFunc") {
		principles = append(principles, "REST handlers should use http.HandlerFunc signature")
	}
	if strings.Contains(code, "json.Marshal") {
		principles = append(principles, "Use json.Marshal for JSON responses, handle errors")
	}
	if strings.Contains(code, "defer") {
		principles = append(principles, "Use defer for resource cleanup (Close, Rollback)")
	}
	if strings.Contains(code, "context.Context") {
		principles = append(principles, "Pass context.Context for cancellation and timeouts")
	}

	// Pattern-specific principles
	switch taskType {
	case "database_migration":
		principles = append(principles, "Wrap DDL in transaction if supported, add rollback comments")
	case "rest_api_endpoint":
		principles = append(principles, "Validate input, return proper HTTP status codes, log errors")
	case "unit_test":
		principles = append(principles, "Use table-driven tests, test edge cases, mock dependencies")
	}

	return principles, nil
}

// IsAuthenticated checks if GitHub CLI is authenticated
func (gc *GitHubCopilot) IsAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}

// GetAuthenticationInstructions returns instructions for authenticating GitHub CLI
func (gc *GitHubCopilot) GetAuthenticationInstructions() string {
	return `GitHub CLI is not authenticated. To authenticate:

1. Run: gh auth login
2. Choose: GitHub.com
3. Choose: HTTPS
4. Authenticate via web browser
5. Verify with: gh auth status

Note: Requires GitHub Pro subscription for Copilot access.`
}
