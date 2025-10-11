package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	serverPort = "8080"
	maxRetries = 30
	retryDelay = 1 * time.Second
)

var (
	serverCmd  *exec.Cmd
	desktopCmd *exec.Cmd
)

func main() {
	// Hide console window on Windows
	hideConsoleWindow()

	log.SetOutput(os.Stdout)
	log.Println("🚀 ARES Launcher Starting...")

	// Get the directory where the launcher is running
	exePath, err := os.Executable()
	if err != nil {
		showError("Failed to get executable path: " + err.Error())
		return
	}
	baseDir := filepath.Dir(exePath)

	// Find the API server executable
	apiServerPath := findAPIServer(baseDir)
	if apiServerPath == "" {
		showError("Could not find ARES API server.\nExpected 'ares_api.exe' in same directory")
		return
	}

	// Find the Desktop UI executable
	desktopUIPath := findDesktopUI(baseDir)
	if desktopUIPath == "" {
		showError("Could not find ARES Desktop UI.\nExpected ARESDesktop.exe in ARES_UI folder")
		return
	}

	// Setup signal handlers for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the API server
	log.Println("Starting API Server...")
	if err := startAPIServer(apiServerPath); err != nil {
		showError("Failed to start API server: " + err.Error())
		return
	}

	// Wait for server to be ready
	if !waitForServer() {
		showError("API Server failed to start within timeout period")
		stopAll()
		return
	}

	log.Println("✓ ARES API Server is ready")

	// Start the Desktop UI
	log.Println("Starting Desktop UI...")
	if err := startDesktopUI(desktopUIPath); err != nil {
		log.Printf("Warning: Failed to start desktop UI: %v\n", err)
		stopAll()
		return
	}

	log.Println("✓ ARES Desktop is running")
	log.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal or desktop UI to close
	select {
	case <-sigChan:
		log.Println("\n🛑 Shutting down ARES...")
	case <-waitForDesktopUI():
		log.Println("\n🛑 Desktop UI closed, shutting down...")
	}

	stopAll()
	log.Println("✓ ARES stopped gracefully")
}

// findAPIServer locates the API server executable
func findAPIServer(baseDir string) string {
	// Try same directory first
	exePath := filepath.Join(baseDir, "ares_api.exe")
	if fileExists(exePath) {
		return exePath
	}

	// Try ../ARES_API directory
	exePath = filepath.Join(filepath.Dir(baseDir), "ARES_API", "ares_api.exe")
	if fileExists(exePath) {
		return exePath
	}

	return ""
}

// findDesktopUI locates the Desktop UI executable
func findDesktopUI(baseDir string) string {
	// Try ../ARES_UI/ARESDesktop/bin/Release/net8.0/ARESDesktop.exe
	uiPath := filepath.Join(filepath.Dir(baseDir), "ARES_UI", "ARESDesktop", "bin", "Release", "net8.0", "ARESDesktop.exe")
	if fileExists(uiPath) {
		return uiPath
	}

	// Try same directory
	uiPath = filepath.Join(baseDir, "ARESDesktop.exe")
	if fileExists(uiPath) {
		return uiPath
	}

	return ""
}

// startAPIServer starts the API server process
func startAPIServer(serverPath string) error {
	cmd := exec.Command(serverPath)

	// Set working directory
	cmd.Dir = filepath.Dir(serverPath)

	// Hide window on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	// Redirect output to log file
	logFile, err := os.Create(filepath.Join(cmd.Dir, "ares_server.log"))
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	serverCmd = cmd
	log.Printf("✓ API Server started (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// startDesktopUI starts the Desktop UI process
func startDesktopUI(uiPath string) error {
	cmd := exec.Command(uiPath)

	// Set working directory
	cmd.Dir = filepath.Dir(uiPath)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start desktop UI: %w", err)
	}

	desktopCmd = cmd
	log.Printf("✓ Desktop UI started (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// waitForServer waits for the API server to become ready
func waitForServer() bool {
	log.Println("⏳ Waiting for API server to be ready...")

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get("http://localhost:8080/swagger/index.html")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(retryDelay)
		if i%5 == 0 && i > 0 {
			log.Printf("  Still waiting... (%d/%d)\n", i, maxRetries)
		}
	}

	return false
}

// waitForDesktopUI returns a channel that closes when desktop UI exits
func waitForDesktopUI() <-chan struct{} {
	ch := make(chan struct{})
	if desktopCmd != nil {
		go func() {
			desktopCmd.Wait()
			close(ch)
		}()
	}
	return ch
}

// stopAll gracefully stops all processes
func stopAll() {
	stopDesktopUI()
	stopAPIServer()
}

// stopAPIServer gracefully stops the API server
func stopAPIServer() {
	if serverCmd == nil || serverCmd.Process == nil {
		return
	}

	log.Printf("Stopping API server (PID: %d)...\n", serverCmd.Process.Pid)

	// Try graceful shutdown first
	if err := serverCmd.Process.Signal(os.Interrupt); err != nil {
		// Force kill if graceful shutdown fails
		serverCmd.Process.Kill()
	}

	// Wait for process to exit
	serverCmd.Wait()
}

// stopDesktopUI gracefully stops the Desktop UI
func stopDesktopUI() {
	if desktopCmd == nil || desktopCmd.Process == nil {
		return
	}

	log.Printf("Stopping Desktop UI (PID: %d)...\n", desktopCmd.Process.Pid)

	// Try graceful shutdown first
	if err := desktopCmd.Process.Signal(os.Interrupt); err != nil {
		// Force kill if graceful shutdown fails
		desktopCmd.Process.Kill()
	}

	// Wait for process to exit
	desktopCmd.Wait()
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// showError shows an error message to the user
func showError(message string) {
	log.Printf("ERROR: %s\n", message)

	// Show Windows message box
	cmd := exec.Command("mshta", "vbscript:Execute(\"msgbox \"\""+message+"\"\",16,\"\"ARES Launcher Error\"\":close\")")
	cmd.Run()
}

// hideConsoleWindow hides the console window on Windows
func hideConsoleWindow() {
	// This will be called at startup
	// The actual hiding is handled by build flags
}
