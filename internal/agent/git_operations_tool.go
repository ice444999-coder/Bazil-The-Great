package agent

import (
	"ares_api/pkg/llm"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Git Operations Tool - Allows SOLACE to commit and push changes autonomously
func gitOperationsTool() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        "git_commit_and_push",
			Description: "🔧 Commit staged changes to Git with structured message and optionally push to GitHub. Use this to persist improvements, fixes, and new features. ALWAYS check git_status first to see what's changed.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"files": map[string]interface{}{
						"type": "array",
						"items": map[string]string{
							"type": "string",
						},
						"description": "Files to stage and commit (relative to C:/ARES_Workspace). Example: ['ARES_API/internal/agent/solace_tools.go', '.env']",
					},
					"commit_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"feat", "fix", "docs", "refactor", "test", "chore", "perf"},
						"description": "Conventional commit type: feat (new feature), fix (bug fix), docs (documentation), refactor (code restructure), test (tests), chore (maintenance), perf (performance)",
					},
					"scope": map[string]string{
						"type":        "string",
						"description": "Component being modified (e.g., 'solace', 'api', 'database', 'git-ops')",
					},
					"message": map[string]string{
						"type":        "string",
						"description": "Short commit summary (50 chars or less). Example: 'Add autonomous git operations capability'",
					},
					"body": map[string]string{
						"type":        "string",
						"description": "Detailed commit description (optional). Explain what changed and why.",
					},
					"push": map[string]interface{}{
						"type":        "boolean",
						"description": "Push to remote GitHub repository after committing (default: false for safety)",
						"default":     false,
					},
				},
				"required": []string{"files", "commit_type", "message"},
			},
		},
	}
}

func gitStatusTool() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        "git_status",
			Description: "📊 Check current Git repository status: branch name, uncommitted files, last commit info. Use this BEFORE git_commit_and_push to see what's changed.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func gitLogTool() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        "git_log",
			Description: "📜 View recent Git commit history (last 10 commits). Shows commit hashes, authors, dates, and messages.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of commits to show (default: 10, max: 50)",
						"default":     10,
					},
				},
			},
		},
	}
}

// Handler for git_commit_and_push
func (s *SOLACE) handleGitCommitAndPush(args map[string]interface{}) (string, error) {
	// Git repo is in C:/ARES_Workspace/ARES_API
	workspaceRoot := "C:/ARES_Workspace/ARES_API"

	// Parse arguments
	filesRaw, ok := args["files"].([]interface{})
	if !ok {
		return "", fmt.Errorf("files parameter is required and must be an array")
	}

	var files []string
	for _, f := range filesRaw {
		if fStr, ok := f.(string); ok {
			files = append(files, fStr)
		}
	}

	if len(files) == 0 {
		return "", fmt.Errorf("at least one file must be specified")
	}

	commitType, ok := args["commit_type"].(string)
	if !ok {
		return "", fmt.Errorf("commit_type is required")
	}

	message, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("message is required")
	}

	scope := ""
	if s, ok := args["scope"].(string); ok {
		scope = s
	}

	body := ""
	if b, ok := args["body"].(string); ok {
		body = b
	}

	push := false
	if p, ok := args["push"].(bool); ok {
		push = p
	}

	// Change to workspace directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspaceRoot); err != nil {
		return "", fmt.Errorf("failed to change to workspace directory: %v", err)
	}

	// Stage files
	var stagedFiles []string
	for _, file := range files {
		cmd := exec.Command("git", "add", file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to stage %s: %v\n%s", file, err, string(output))
		}
		stagedFiles = append(stagedFiles, file)
	}

	// Build conventional commit message
	commitMsg := fmt.Sprintf("%s", commitType)
	if scope != "" {
		commitMsg += fmt.Sprintf("(%s)", scope)
	}
	commitMsg += fmt.Sprintf(": %s", message)

	if body != "" {
		commitMsg += "\n\n" + body
	}

	// Add metadata
	commitMsg += fmt.Sprintf("\n\n🤖 Auto-committed by SOLACE")
	commitMsg += fmt.Sprintf("\nTimestamp: %s", time.Now().Format("2006-01-02 15:04:05"))
	commitMsg += fmt.Sprintf("\nFiles: %s", strings.Join(stagedFiles, ", "))

	// Commit
	cmd := exec.Command("git", "commit", "-m", commitMsg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to commit: %v\n%s", err, string(output))
	}

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	hashOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %v", err)
	}
	commitHash := strings.TrimSpace(string(hashOutput))

	result := fmt.Sprintf("✅ Commit successful!\n\n")
	result += fmt.Sprintf("Commit Hash: %s\n", commitHash[:8])
	result += fmt.Sprintf("Type: %s\n", commitType)
	if scope != "" {
		result += fmt.Sprintf("Scope: %s\n", scope)
	}
	result += fmt.Sprintf("Message: %s\n", message)
	result += fmt.Sprintf("Files staged: %d\n", len(stagedFiles))
	result += fmt.Sprintf("  - %s\n", strings.Join(stagedFiles, "\n  - "))

	// Push if requested
	if push {
		// Get current branch
		cmd = exec.Command("git", "branch", "--show-current")
		branchOutput, err := cmd.Output()
		if err != nil {
			return result + fmt.Sprintf("\n⚠️ Warning: Failed to get current branch for push: %v", err), nil
		}
		branch := strings.TrimSpace(string(branchOutput))

		// Push to remote
		cmd = exec.Command("git", "push", "origin", branch)
		pushOutput, err := cmd.CombinedOutput()
		if err != nil {
			return result + fmt.Sprintf("\n⚠️ Warning: Failed to push to remote: %v\n%s", err, string(pushOutput)), nil
		}

		result += fmt.Sprintf("\n🚀 Pushed to origin/%s successfully!", branch)
	} else {
		result += "\n\n💡 Tip: Use push=true to automatically push to GitHub"
	}

	return result, nil
}

// Handler for git_status
func (s *SOLACE) handleGitStatus(args map[string]interface{}) (string, error) {
	workspaceRoot := "C:/ARES_Workspace/ARES_API"

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspaceRoot); err != nil {
		return "", fmt.Errorf("failed to change to workspace directory: %v", err)
	}

	// Get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	branchOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get status
	cmd = exec.Command("git", "status", "--porcelain")
	statusOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %v", err)
	}

	// Get last commit
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%h - %s (%cr) <%an>")
	lastCommitOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit: %v", err)
	}
	lastCommit := strings.TrimSpace(string(lastCommitOutput))

	result := fmt.Sprintf("📊 Git Status\n\n")
	result += fmt.Sprintf("Branch: %s\n", branch)
	result += fmt.Sprintf("Last commit: %s\n\n", lastCommit)

	statusStr := strings.TrimSpace(string(statusOutput))
	if statusStr == "" {
		result += "✅ Working tree clean - no uncommitted changes\n"
	} else {
		result += "📝 Uncommitted changes:\n\n"
		lines := strings.Split(statusStr, "\n")
		for _, line := range lines {
			if len(line) >= 3 {
				status := line[:2]
				file := strings.TrimSpace(line[3:])

				switch status {
				case "M ", " M":
					result += fmt.Sprintf("  🔧 Modified: %s\n", file)
				case "A ", " A":
					result += fmt.Sprintf("  ➕ Added: %s\n", file)
				case "D ", " D":
					result += fmt.Sprintf("  ➖ Deleted: %s\n", file)
				case "??":
					result += fmt.Sprintf("  ❓ Untracked: %s\n", file)
				case "R ":
					result += fmt.Sprintf("  🔄 Renamed: %s\n", file)
				default:
					result += fmt.Sprintf("  %s %s\n", status, file)
				}
			}
		}
	}

	return result, nil
}

// Handler for git_log
func (s *SOLACE) handleGitLog(args map[string]interface{}) (string, error) {
	workspaceRoot := "C:/ARES_Workspace/ARES_API"

	count := 10
	if c, ok := args["count"].(float64); ok {
		count = int(c)
		if count > 50 {
			count = 50
		}
		if count < 1 {
			count = 1
		}
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspaceRoot); err != nil {
		return "", fmt.Errorf("failed to change to workspace directory: %v", err)
	}

	// Get log
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--pretty=format:%h|%s|%cr|%an")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git log: %v", err)
	}

	result := fmt.Sprintf("📜 Git History (last %d commits)\n\n", count)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			hash := parts[0]
			msg := parts[1]
			when := parts[2]
			author := parts[3]

			result += fmt.Sprintf("%d. [%s] %s\n", i+1, hash, msg)
			result += fmt.Sprintf("   👤 %s • ⏰ %s\n\n", author, when)
		}
	}

	return result, nil
}

// Git operation handlers are called from executeTool() switch statement in solace_tools.go
