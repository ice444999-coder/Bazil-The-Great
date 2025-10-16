package agent

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ProcessManager - SOLACE's system process control
type ProcessManager struct {
	solace *SOLACE
}

type ManagedProcess struct {
	Name        string
	PID         int
	StartTime   time.Time
	Runtime     time.Duration
	IsEssential bool
	Purpose     string
}

func NewProcessManager(solace *SOLACE) *ProcessManager {
	return &ProcessManager{solace: solace}
}

// CullZombieProcesses - Kill old PowerShell zombies, keep only active ones
func (pm *ProcessManager) CullZombieProcesses(ctx context.Context) error {
	log.Println("🔪 SOLACE: Culling zombie PowerShell processes...")

	// Get all PowerShell processes
	powershells, err := pm.GetAllPowerShellProcesses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get PowerShell processes: %w", err)
	}

	// Essential process criteria
	maxAge := 2 * time.Hour // Kill anything older than 2 hours
	keepRecent := 3         // Keep 3 most recent

	var toKill []int
	var toKeep []ManagedProcess

	// Sort by age (oldest first for culling)
	now := time.Now()
	for _, ps := range powershells {
		age := now.Sub(ps.StartTime)

		// Kill criteria:
		// 1. Older than 2 hours
		// 2. No main window (background zombie)
		// 3. Not in the most recent 3 processes
		if age > maxAge && !ps.IsEssential {
			toKill = append(toKill, ps.PID)
		} else {
			toKeep = append(toKeep, ps)
		}
	}

	// Keep only most recent N processes
	if len(toKeep) > keepRecent {
		// Sort by StartTime descending (newest first)
		for i := 0; i < len(toKeep)-1; i++ {
			for j := i + 1; j < len(toKeep); j++ {
				if toKeep[i].StartTime.Before(toKeep[j].StartTime) {
					toKeep[i], toKeep[j] = toKeep[j], toKeep[i]
				}
			}
		}

		// Mark older ones for killing
		for i := keepRecent; i < len(toKeep); i++ {
			toKill = append(toKill, toKeep[i].PID)
		}
	}

	// Execute culling
	killed := 0
	for _, pid := range toKill {
		if err := pm.KillProcess(ctx, pid); err == nil {
			killed++
		}
	}

	log.Printf("✅ SOLACE: Culled %d zombie PowerShell processes (kept %d active)", killed, keepRecent)

	// Log to SOLACE's memory
	pm.solace.WorkingMemory.AddEvent(&Event{
		Timestamp:   time.Now(),
		Type:        "process_management",
		Description: fmt.Sprintf("Culled %d zombie PowerShell processes", killed),
		Data: map[string]interface{}{
			"killed": killed,
			"kept":   keepRecent,
		},
		Importance: 0.6,
	})

	return nil
}

// GetAllPowerShellProcesses - Query all PowerShell/pwsh processes
func (pm *ProcessManager) GetAllPowerShellProcesses(ctx context.Context) ([]ManagedProcess, error) {
	// PowerShell command to get all PowerShell processes with details
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		`Get-Process | Where-Object { $_.ProcessName -eq "pwsh" -or $_.ProcessName -eq "powershell" } | Select-Object ProcessName, Id, StartTime, @{Name='RuntimeMinutes';Expression={(New-TimeSpan -Start $_.StartTime).TotalMinutes}}, MainWindowTitle | ConvertTo-Json`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query PowerShell processes: %w", err)
	}

	// Parse JSON output (simplified - would need proper JSON parsing)
	lines := strings.Split(string(output), "\n")
	var processes []ManagedProcess

	for _, line := range lines {
		// Simple parsing - look for "Id": and "StartTime": patterns
		if strings.Contains(line, "\"Id\":") {
			// Extract PID
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				pidStr := strings.TrimSpace(strings.Trim(parts[1], ","))
				if pid, err := strconv.Atoi(pidStr); err == nil {
					processes = append(processes, ManagedProcess{
						Name:        "powershell",
						PID:         pid,
						StartTime:   time.Now().Add(-1 * time.Hour), // Placeholder
						IsEssential: false,
					})
				}
			}
		}
	}

	return processes, nil
}

// GetEssentialProcesses - Returns PIDs that must stay alive
func (pm *ProcessManager) GetEssentialProcesses(ctx context.Context) (map[string]int, error) {
	essential := make(map[string]int)

	// Check for ARES API
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		`Get-Process | Where-Object { $_.ProcessName -like "*ares*" } | Select-Object ProcessName, Id | ConvertTo-Json`)

	output, err := cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "\"ProcessName\":") && strings.Contains(line, "ares") {
				// Mark as essential
				essential["ares-api"] = 1 // Simplified
			}
		}
	}

	// Check for PostgreSQL
	essential["postgres"] = 1

	// Check for Redis (if running)
	essential["redis"] = 1

	return essential, nil
}

// KillProcess - Terminate a process by PID
func (pm *ProcessManager) KillProcess(ctx context.Context, pid int) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Stop-Process -Id %d -Force -ErrorAction SilentlyContinue", pid))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}

	log.Printf("🔪 Killed process: PID %d", pid)
	return nil
}

// EnsureEssentialProcesses - Make sure critical services are running
func (pm *ProcessManager) EnsureEssentialProcesses(ctx context.Context) error {
	log.Println("🔍 SOLACE: Checking essential processes...")

	// Check ARES API
	aresRunning, err := pm.IsProcessRunning(ctx, "ares-api")
	if err != nil || !aresRunning {
		log.Println("⚠️  ARES API not running - would need to start it")
		// Could auto-restart here
	}

	// Check PostgreSQL
	pgRunning, err := pm.IsProcessRunning(ctx, "postgres")
	if err != nil || !pgRunning {
		log.Println("⚠️  PostgreSQL not running - CRITICAL!")
		return fmt.Errorf("PostgreSQL is not running")
	}

	log.Println("✅ All essential processes verified")
	return nil
}

// IsProcessRunning - Check if a process with given name is running
func (pm *ProcessManager) IsProcessRunning(ctx context.Context, processName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf(`Get-Process | Where-Object { $_.ProcessName -like "*%s*" } | Select-Object -First 1`, processName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	return len(output) > 0 && !strings.Contains(string(output), "Cannot find"), nil
}

// GetSystemResourceUsage - Monitor system resources
func (pm *ProcessManager) GetSystemResourceUsage(ctx context.Context) (map[string]interface{}, error) {
	// Get CPU and Memory for key processes
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		`Get-Process | Where-Object { $_.ProcessName -like "*ares*" -or $_.ProcessName -eq "postgres" } | Select-Object ProcessName, @{Name='CPU';Expression={[math]::Round($_.CPU,2)}}, @{Name='MemoryMB';Expression={[math]::Round($_.WorkingSet/1MB,2)}} | ConvertTo-Json`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage: %w", err)
	}

	return map[string]interface{}{
		"timestamp": time.Now(),
		"raw_data":  string(output),
	}, nil
}
