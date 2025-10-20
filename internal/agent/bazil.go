package agent

import (
	"ares_api/internal/models"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BazilSniffer performs god-tier code analysis with human oversight
type BazilSniffer struct {
	db          *gorm.DB
	dryRun      bool
	maxFindings int
}

// NewBazilSniffer creates a new code analysis agent
func NewBazilSniffer(db *gorm.DB) *BazilSniffer {
	dryRun := os.Getenv("DRY_RUN") != "false" // Default true
	return &BazilSniffer{
		db:          db,
		dryRun:      dryRun,
		maxFindings: 100, // Limit to prevent overwhelming results
	}
}

// SniffCode performs comprehensive code analysis with AST + linters
func (b *BazilSniffer) SniffCode(dir string) ([]models.BazilFinding, error) {
	log.Printf("🔍 Bazil starting code analysis on: %s (dry_run=%v)", dir, b.dryRun)

	findings := []models.BazilFinding{}

	// Check if analysis should run based on toggles
	if os.Getenv("BAZIL_ENABLED") == "false" {
		log.Println("⏸️  Bazil analysis disabled via toggle")
		return findings, nil
	}

	// Phase 1: Grep for basic patterns
	if os.Getenv("BAZIL_GREP_ENABLED") != "false" {
		grepFindings, err := b.runGrepAnalysis(dir)
		if err != nil {
			log.Printf("⚠️  Grep analysis error: %v", err)
		} else {
			findings = append(findings, grepFindings...)
			log.Printf("✅ Grep analysis found %d issues", len(grepFindings))
		}
	}

	// Phase 2: AST parsing for semantic analysis
	if os.Getenv("BAZIL_AST_ENABLED") != "false" {
		astFindings, err := b.runASTAnalysis(dir)
		if err != nil {
			log.Printf("⚠️  AST analysis error: %v", err)
		} else {
			findings = append(findings, astFindings...)
			log.Printf("✅ AST analysis found %d issues", len(astFindings))
		}
	}

	// Phase 3: Run linters for bugs/perf/security
	if os.Getenv("BAZIL_LINTERS_ENABLED") != "false" {
		linterFindings, err := b.runLinterAnalysis(dir)
		if err != nil {
			log.Printf("⚠️  Linter analysis error: %v", err)
		} else {
			findings = append(findings, linterFindings...)
			log.Printf("✅ Linter analysis found %d issues", len(linterFindings))
		}
	}

	// Limit findings if too many
	if len(findings) > b.maxFindings {
		log.Printf("⚠️  Found %d issues, limiting to top %d by confidence", len(findings), b.maxFindings)
		findings = b.sortAndLimitByConfidence(findings)
	}

	// Save findings to database
	if !b.dryRun {
		for i := range findings {
			findings[i].Status = "pending"
			if err := b.db.Create(&findings[i]).Error; err != nil {
				log.Printf("❌ Error saving finding: %v", err)
			}
		}
		log.Printf("💾 Saved %d findings to database", len(findings))
	} else {
		log.Printf("🔍 DRY RUN: Would save %d findings (not persisted)", len(findings))
	}

	return findings, nil
}

// runGrepAnalysis performs pattern matching
func (b *BazilSniffer) runGrepAnalysis(dir string) ([]models.BazilFinding, error) {
	findings := []models.BazilFinding{}

	// Check if grep is available
	if _, err := exec.LookPath("grep"); err != nil {
		return findings, fmt.Errorf("grep not found: %v", err)
	}

	patterns := []string{
		"TODO|FIXME|HACK|XXX",
		"console\\.log|fmt\\.Println.*debug",
		"panic\\(|recover\\(\\)",
		"nav-item.*active",
	}

	for _, pattern := range patterns {
		cmd := exec.Command("grep", "-r", "-n", "-i", "-E", pattern, dir,
			"--include=*.go", "--include=*.html", "--include=*.js")
		output, err := cmd.Output()
		if err != nil {
			// Grep returns non-zero if no matches, which is OK
			continue
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				findings = append(findings, models.BazilFinding{
					FaultType:   "pattern_match",
					Description: fmt.Sprintf("Pattern '%s': %s", pattern, strings.TrimSpace(parts[2])),
					FilePath:    parts[0],
					LineNumber:  atoi(parts[1]),
					Confidence:  0.5,
					UUID:        uuid.New(),
				})
			}
		}
	}

	return findings, nil
}

// runASTAnalysis performs semantic code analysis
func (b *BazilSniffer) runASTAnalysis(dir string) ([]models.BazilFinding, error) {
	findings := []models.BazilFinding{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}

		// Skip vendor and test files for performance
		if strings.Contains(path, "/vendor/") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if parseErr != nil {
			log.Printf("⚠️  Parse error in %s: %v", path, parseErr)
			return nil // Continue with other files
		}

		// Check for unused imports
		usedPackages := make(map[string]bool)

		// First pass: identify used packages
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if ident, ok := x.X.(*ast.Ident); ok {
					usedPackages[ident.Name] = true
				}
			}
			return true
		})

		// Second pass: check imports
		for _, imp := range node.Imports {
			if imp.Name != nil && imp.Name.Name == "_" {
				continue // Blank imports are intentional
			}

			pkgPath := strings.Trim(imp.Path.Value, `"`)
			pkgName := filepath.Base(pkgPath)
			if imp.Name != nil {
				pkgName = imp.Name.Name
			}

			if !usedPackages[pkgName] {
				findings = append(findings, models.BazilFinding{
					FaultType:   "unused_import",
					Description: fmt.Sprintf("Unused import: %s", pkgPath),
					FilePath:    path,
					LineNumber:  fset.Position(imp.Pos()).Line,
					Confidence:  0.8,
					UUID:        uuid.New(),
				})
			}
		}

		// Check for potential issues in function calls
		ast.Inspect(node, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				// Check for calls that might return errors but aren't checked
				if ident, ok := call.Fun.(*ast.Ident); ok {
					if strings.HasSuffix(ident.Name, "Error") && len(call.Args) == 0 {
						findings = append(findings, models.BazilFinding{
							FaultType:   "potential_issue",
							Description: fmt.Sprintf("Call to %s with no args - verify intent", ident.Name),
							FilePath:    path,
							LineNumber:  fset.Position(call.Pos()).Line,
							Confidence:  0.6,
							UUID:        uuid.New(),
						})
					}
				}
			}
			return true
		})

		return nil
	})

	return findings, err
}

// runLinterAnalysis runs external linters if available
func (b *BazilSniffer) runLinterAnalysis(dir string) ([]models.BazilFinding, error) {
	findings := []models.BazilFinding{}

	linters := []struct {
		name      string
		cmd       []string
		faultType string
		confBase  float64
	}{
		{"golangci-lint", []string{"golangci-lint", "run", "--no-config", "--fast", dir}, "aggregate_lint", 0.9},
		{"staticcheck", []string{"staticcheck", dir}, "perf_bug", 0.85},
		{"gosec", []string{"gosec", "-quiet", dir}, "security", 0.95},
	}

	for _, l := range linters {
		// Check if linter is available
		if _, err := exec.LookPath(l.cmd[0]); err != nil {
			log.Printf("⏭️  Skipping %s (not installed): %v", l.name, err)
			continue
		}

		cmd := exec.Command(l.cmd[0], l.cmd[1:]...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			log.Printf("⚠️  %s returned error (may have findings): %v", l.name, err)
		}

		output := stdout.String()
		if output == "" {
			output = stderr.String()
		}

		for _, line := range strings.Split(output, "\n") {
			if line == "" || strings.Contains(line, "No issues found") {
				continue
			}

			// Parse output format: file:line:col: message
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 3 {
				lineNum := atoi(parts[1])
				desc := line
				if len(parts) >= 4 {
					desc = strings.TrimSpace(parts[3])
				}

				findings = append(findings, models.BazilFinding{
					FaultType:   l.faultType,
					Description: desc,
					FilePath:    parts[0],
					LineNumber:  lineNum,
					Confidence:  l.confBase,
					UUID:        uuid.New(),
				})
			}
		}
	}

	return findings, nil
}

// sortAndLimitByConfidence sorts findings and returns top N
func (b *BazilSniffer) sortAndLimitByConfidence(findings []models.BazilFinding) []models.BazilFinding {
	// Simple bubble sort by confidence (descending)
	for i := 0; i < len(findings)-1; i++ {
		for j := 0; j < len(findings)-i-1; j++ {
			if findings[j].Confidence < findings[j+1].Confidence {
				findings[j], findings[j+1] = findings[j+1], findings[j]
			}
		}
	}

	if len(findings) > b.maxFindings {
		return findings[:b.maxFindings]
	}
	return findings
}

// TrackReward saves points for successful findings
func (b *BazilSniffer) TrackReward(faultType string, points int) error {
	if b.dryRun {
		log.Printf("🔍 DRY RUN: Would award %d points for %s", points, faultType)
		return nil
	}

	var reward models.BazilReward
	if err := b.db.Where("fault_type = ?", faultType).FirstOrCreate(&reward, models.BazilReward{
		FaultType: faultType,
		Points:    0,
	}).Error; err != nil {
		return fmt.Errorf("failed to query reward: %v", err)
	}

	reward.Points += points
	if err := b.db.Save(&reward).Error; err != nil {
		return fmt.Errorf("failed to save reward: %v", err)
	}

	log.Printf("🏆 Awarded %d points for %s (total: %d)", points, faultType, reward.Points)
	return nil
}

// GetPendingFindings returns findings awaiting human review
func (b *BazilSniffer) GetPendingFindings() ([]models.BazilFinding, error) {
	var findings []models.BazilFinding
	err := b.db.Where("status = ?", "pending").Order("confidence DESC").Find(&findings).Error
	return findings, err
}

// GetRewards retrieves all reward records
func (b *BazilSniffer) GetRewards() (map[string]int, error) {
	var rewards []models.BazilReward
	if err := b.db.Find(&rewards).Error; err != nil {
		return nil, err
	}
	rewardMap := make(map[string]int)
	for _, r := range rewards {
		rewardMap[r.FaultType] = r.Points
	}
	return rewardMap, nil
}

// Helper: atoi with zero fallback
func atoi(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return i
}
