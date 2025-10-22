package agent

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// SelfHealUI - Solace directs detection (UI tests + Bazil sniff), Forge patching, and application
func (s *SOLACE) SelfHealUI() {
	log.Println("🔍 SOLACE: Starting self-healing UI scan...")

	faultType := "ui_wiring_display"

	// Step 1: Use new Bazil sniffer to analyze UI code
	findings, sniffErr := s.bazil.SniffCode("./web")
	if sniffErr != nil {
		log.Printf("❌ Bazil sniff error: %v\n", sniffErr)
		return
	}

	if len(findings) == 0 {
		log.Println("✅ No UI faults detected - system healthy")
		return
	}

	log.Printf("🔍 Bazil found %d potential UI issues\n", len(findings))

	// Step 2: Analyze findings for specific UI issues (navigation, wiring, etc.)
	needsHealing := false
	issueDescription := ""

	// Check for navigation issues from findings
	for _, finding := range findings {
		if finding.FaultType == "pattern_match" && finding.Confidence >= 0.5 {
			needsHealing = true
			issueDescription = fmt.Sprintf("UI issue detected: %s at %s:%d",
				finding.Description, finding.FilePath, finding.LineNumber)
			log.Printf("⚠️ %s\n", issueDescription)
			break
		}
	}

	// If faults detected, direct Forge to generate patch
	if needsHealing {
		log.Printf("🛠️ Solace directing Forge to generate patch for: %s\n", faultType)

		patch, patchErr := s.forge.GenerateUIPatch(issueDescription)
		if patchErr != nil {
			log.Printf("❌ Forge patch generation failed: %v\n", patchErr)
			return
		}

		// Step 3: Apply patch safely via git apply
		patchFile := "/tmp/ui_patch.diff" // Temp file (Windows: use os.TempDir())
		if os.PathSeparator == '\\' {
			patchFile = os.TempDir() + "\\ui_patch.diff"
		}

		err := os.WriteFile(patchFile, []byte(patch), 0644)
		if err != nil {
			log.Printf("❌ Failed to write patch file: %v\n", err)
			return
		}
		defer os.Remove(patchFile) // Clean up

		// Dry-run first
		cmd := exec.Command("git", "apply", "--check", patchFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			log.Printf("❌ Patch dry-run failed: %v\nStderr: %s\n", err, stderr.String())
			log.Println("💡 Patch may conflict with current code. Manual review recommended.")
			return
		}

		// Apply patch
		log.Println("✅ Patch dry-run successful, applying patch...")
		cmd = exec.Command("git", "apply", patchFile)
		cmd.Stderr = &stderr

		output, applyErr := cmd.Output()
		if applyErr != nil {
			log.Printf("❌ Patch apply failed: %v\nStderr: %s\n", applyErr, stderr.String())
		} else {
			log.Printf("✅ Patch applied successfully!\nOutput: %s\n", string(output))

			// Reward Bazil
			if err := s.bazil.TrackReward(faultType, 1); err != nil {
				log.Printf("⚠️ Failed to track reward: %v\n", err)
			}

			log.Printf("🏆 Bazil rewarded for fixing %s\n", faultType)

			// Note: In production, you'd restart services via event bus
			// For now, just log
			log.Println("🔄 Note: Server restart may be needed for UI changes to take effect")
		}
	} else {
		log.Println("✅ No UI faults detected - system healthy")
	}
}

// Helper function to check if string contains substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
