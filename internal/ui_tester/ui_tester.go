/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package ui_tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// UITester handles automated UI testing and validation
// Integrates with Playwright/Selenium for browser automation
type UITester struct {
	pythonPath  string
	testScripts []string
	lastRunTime time.Time
	lastResults *TestResults
}

// TestResults represents UI test execution results
type TestResults struct {
	TotalTests  int                    `json:"total_tests"`
	PassedTests int                    `json:"passed_tests"`
	FailedTests int                    `json:"failed_tests"`
	Duration    time.Duration          `json:"duration"`
	Failures    []string               `json:"failures"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details"`
}

// NewUITester initializes the UI testing system
func NewUITester() *UITester {
	return &UITester{
		pythonPath: "python3", // Use "python" on Windows if needed
		testScripts: []string{
			"sentinel_ui_test.py",
			"dashboard_test.py",
			"trading_ui_test.py",
		},
		lastResults: &TestResults{},
	}
}

// RunTests executes all UI test scripts
func (ut *UITester) RunTests() (*TestResults, error) {
	log.Println("🧪 Starting UI tests...")

	results := &TestResults{
		Timestamp: time.Now(),
		Failures:  []string{},
		Details:   make(map[string]interface{}),
	}

	startTime := time.Now()

	for _, script := range ut.testScripts {
		log.Printf("Running test script: %s", script)

		cmd := exec.Command(ut.pythonPath, script)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		results.TotalTests++

		if err != nil {
			results.FailedTests++
			failureMsg := fmt.Sprintf("%s failed: %v\nStderr: %s", script, err, stderr.String())
			results.Failures = append(results.Failures, failureMsg)
			log.Printf("❌ %s", failureMsg)
		} else {
			results.PassedTests++
			log.Printf("✅ %s passed", script)

			// Try to parse JSON output if available
			if stdout.Len() > 0 {
				var testOutput map[string]interface{}
				if err := json.Unmarshal(stdout.Bytes(), &testOutput); err == nil {
					results.Details[script] = testOutput
				}
			}
		}
	}

	results.Duration = time.Since(startTime)
	ut.lastRunTime = time.Now()
	ut.lastResults = results

	log.Printf("🧪 UI Tests Complete: %d/%d passed in %v",
		results.PassedTests, results.TotalTests, results.Duration)

	return results, nil
}

// RunSpecificTest runs a single test script
func (ut *UITester) RunSpecificTest(scriptName string) error {
	log.Printf("🧪 Running specific test: %s", scriptName)

	cmd := exec.Command(ut.pythonPath, scriptName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ Test failed: %s\nOutput: %s", err, string(output))
		return err
	}

	log.Printf("✅ Test passed: %s\nOutput: %s", scriptName, string(output))
	return nil
}

// GetLastResults returns the most recent test results
func (ut *UITester) GetLastResults() *TestResults {
	return ut.lastResults
}

// ShouldRunTests determines if tests should run based on interval
func (ut *UITester) ShouldRunTests(interval time.Duration) bool {
	return time.Since(ut.lastRunTime) > interval
}

// ValidatePage performs quick validation of a specific page
func (ut *UITester) ValidatePage(pageURL string) (bool, error) {
	// Simple curl check or headless browser check
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", pageURL)
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}

	statusCode := string(output)
	if statusCode == "200" {
		log.Printf("✅ Page %s is accessible (HTTP 200)", pageURL)
		return true, nil
	}

	log.Printf("❌ Page %s returned HTTP %s", pageURL, statusCode)
	return false, fmt.Errorf("HTTP %s", statusCode)
}

// GetTestSummary returns a human-readable summary
func (ut *UITester) GetTestSummary() string {
	if ut.lastResults.TotalTests == 0 {
		return "No tests run yet"
	}

	passRate := float64(ut.lastResults.PassedTests) / float64(ut.lastResults.TotalTests) * 100
	return fmt.Sprintf("Pass Rate: %.1f%% (%d/%d) | Last Run: %s ago | Duration: %v",
		passRate,
		ut.lastResults.PassedTests,
		ut.lastResults.TotalTests,
		time.Since(ut.lastRunTime).Round(time.Minute),
		ut.lastResults.Duration,
	)
}
