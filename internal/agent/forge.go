package agent

import (
	"ares_api/internal/models"
	"ares_api/pkg/llm"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Forge - AI apprentice for code generation and patching
type Forge struct {
	db  *gorm.DB
	llm *llm.Client
}

// NewForge creates a new Forge instance
func NewForge(db *gorm.DB, llmClient *llm.Client) *Forge {
	return &Forge{
		db:  db,
		llm: llmClient,
	}
}

// GenerateUIPatch - Forge generates git diff patch under Solace direction
func (f *Forge) GenerateUIPatch(issue string) (string, error) {
	prompt := fmt.Sprintf(`Generate a git diff patch to fix UI issue: %s.

Requirements:
- Output ONLY valid git diff format
- Include proper file headers (diff --git, index, ---, +++)
- Use context lines (@@ markers)
- Keep changes minimal and targeted
- No explanations or comments outside the diff

Example format:
diff --git a/web/dashboard.html b/web/dashboard.html
index abc123..def456 100644
--- a/web/dashboard.html
+++ b/web/dashboard.html
@@ -100,1 +100,1 @@
-<old code>
+<new code>

Now generate the patch for: %s`, issue, issue)

	// Create chat messages for LLM
	messages := []llm.Message{
		{Role: "system", Content: "You are Forge, an expert code generator. Generate ONLY valid git diff patches without explanations."},
		{Role: "user", Content: prompt},
	}

	ctx := context.Background()
	response, err := f.llm.Generate(ctx, messages, 0.3)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Validate diff format
	if !strings.HasPrefix(strings.TrimSpace(response), "diff --git") {
		return "", errors.New("invalid patch format from LLM - missing diff header")
	}

	log.Printf("🛠️ Forge generated patch for issue: %s\n", issue)

	// Log preview
	previewLen := 500
	if len(response) < previewLen {
		previewLen = len(response)
	}
	log.Printf("Patch preview:\n%s\n", response[:previewLen])

	return response, nil
}

// ================================================================================
// SELF-HEALING PATCH METHODS - Safe Git-Based Patching
// ================================================================================

// GeneratePatch creates a repair patch for code issues with human oversight
func (f *Forge) GeneratePatch(findings []models.BazilFinding, targetDir string) (*models.BazilPatchApproval, error) {
	log.Printf("🛠️  Forge analyzing %d findings for patch generation...", len(findings))

	// Build detailed prompt with AST-aware context
	prompt := f.buildPatchPrompt(findings)

	// Generate patch content via LLM
	messages := []llm.Message{
		{Role: "system", Content: "You are Forge, an expert code repair AI. Generate ONLY valid git diff patches. Include file paths, line numbers, and minimal changes."},
		{Role: "user", Content: prompt},
	}

	ctx := context.Background()
	response, err := f.llm.Generate(ctx, messages, 0.2) // Low temp for determinism
	if err != nil {
		return nil, fmt.Errorf("patch generation failed: %v", err)
	}

	// Validate patch format
	if !strings.Contains(response, "diff --git") {
		return nil, errors.New("invalid patch format - missing diff headers")
	}

	// Create patch approval record
	patchID := uuid.New().String()
	branchName := fmt.Sprintf("bazil-patch-%s", patchID[:8])

	findingIDs := []string{}
	for _, f := range findings {
		findingIDs = append(findingIDs, f.UUID.String())
	}

	patch := &models.BazilPatchApproval{
		PatchID:      patchID,
		FindingIDs:   strings.Join(findingIDs, ","),
		PatchContent: response,
		Status:       "pending",
		BranchName:   branchName,
	}

	// Save to database for human review
	if err := f.db.Create(patch).Error; err != nil {
		return nil, fmt.Errorf("failed to save patch: %v", err)
	}

	log.Printf("✅ Patch %s generated and ready for approval", patchID)
	return patch, nil
}

// buildPatchPrompt constructs detailed prompt from findings
func (f *Forge) buildPatchPrompt(findings []models.BazilFinding) string {
	var sb strings.Builder
	sb.WriteString("Generate a git diff patch to fix the following code issues:\n\n")

	for i, finding := range findings {
		sb.WriteString(fmt.Sprintf("Issue #%d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Type: %s\n", finding.FaultType))
		sb.WriteString(fmt.Sprintf("  File: %s:%d\n", finding.FilePath, finding.LineNumber))
		sb.WriteString(fmt.Sprintf("  Description: %s\n", finding.Description))
		sb.WriteString(fmt.Sprintf("  Confidence: %.2f\n\n", finding.Confidence))
	}

	sb.WriteString("\nRequirements:\n")
	sb.WriteString("1. Output ONLY valid git diff format (diff --git, index, ---, +++, @@)\n")
	sb.WriteString("2. Include 3 lines of context before/after changes\n")
	sb.WriteString("3. Make minimal, targeted fixes only\n")
	sb.WriteString("4. Preserve existing code style and formatting\n")
	sb.WriteString("5. DO NOT add explanatory comments\n")

	return sb.String()
}

// ApplyPatch applies the approved patch on a new git branch
func (f *Forge) ApplyPatch(patch *models.BazilPatchApproval) error {
	log.Printf("🔧 Applying patch %s on branch %s...", patch.PatchID, patch.BranchName)

	// Safety check: Only apply approved patches
	if patch.Status != "approved" {
		return fmt.Errorf("patch not approved (status: %s)", patch.Status)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get cwd: %v", err)
	}

	// Create new branch
	cmd := exec.Command("git", "checkout", "-b", patch.BranchName)
	cmd.Dir = cwd
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch: %v - %s", err, string(output))
	}

	// Write patch to temp file
	patchFile := filepath.Join(os.TempDir(), fmt.Sprintf("patch_%s.diff", patch.PatchID))
	if err := os.WriteFile(patchFile, []byte(patch.PatchContent), 0644); err != nil {
		return fmt.Errorf("failed to write patch file: %v", err)
	}
	defer os.Remove(patchFile)

	// Apply patch
	cmd = exec.Command("git", "apply", patchFile)
	cmd.Dir = cwd
	if output, err := cmd.CombinedOutput(); err != nil {
		// Rollback branch on failure
		exec.Command("git", "checkout", "main").Run()
		exec.Command("git", "branch", "-D", patch.BranchName).Run()
		return fmt.Errorf("failed to apply patch: %v - %s", err, string(output))
	}

	// Commit changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = cwd
	cmd.Run()

	commitMsg := fmt.Sprintf("Bazil self-heal: %s", patch.PatchID)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = cwd
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %v - %s", err, string(output))
	}

	// Update patch status
	patch.Status = "applied"
	f.db.Save(patch)

	log.Println("✅ Patch applied successfully")
	return nil
}

// RollbackPatch discards a failed patch branch
func (f *Forge) RollbackPatch(branchName string) error {
	log.Printf("⏪ Rolling back branch %s...", branchName)

	// Switch back to main
	cmd := exec.Command("git", "checkout", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout main: %v - %s", err, string(output))
	}

	// Delete branch
	cmd = exec.Command("git", "branch", "-D", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete branch: %v - %s", err, string(output))
	}

	log.Println("✅ Branch rolled back successfully")
	return nil
}

// MergePatch merges an approved and tested patch into main
func (f *Forge) MergePatch(branchName string) error {
	log.Printf("🔀 Merging branch %s into main...", branchName)

	// Switch to main
	cmd := exec.Command("git", "checkout", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout main: %v - %s", err, string(output))
	}

	// Merge branch
	cmd = exec.Command("git", "merge", "--no-ff", branchName, "-m", fmt.Sprintf("Merge %s", branchName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to merge: %v - %s", err, string(output))
	}

	// Delete branch
	cmd = exec.Command("git", "branch", "-d", branchName)
	cmd.Run() // Best effort cleanup

	log.Println("✅ Patch merged successfully")
	return nil
}
