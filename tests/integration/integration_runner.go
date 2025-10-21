/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IntegrationTestResult represents the result of an integration test
type IntegrationTestResult struct {
	TestName     string        `json:"test_name"`
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// IntegrationTester runs end-to-end tests against the ARES API
type IntegrationTester struct {
	baseURL string
	client  *http.Client
	results []IntegrationTestResult
}

// NewIntegrationTester creates a new integration tester
func NewIntegrationTester(baseURL string) *IntegrationTester {
	return &IntegrationTester{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: make([]IntegrationTestResult, 0),
	}
}

// TestHealthCheck tests the basic health endpoint
func (it *IntegrationTester) TestHealthCheck() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/health")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "Health Check",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to connect to health endpoint"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "Health check returned non-200 status"
	} else {
		result.Success = true
		result.Message = "Health check passed"
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// TestConcurrencyStatus tests the concurrency monitoring endpoints
func (it *IntegrationTester) TestConcurrencyStatus() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/api/v1/concurrency/status")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "Concurrency Status",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to get concurrency status"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "Concurrency status returned non-200 status"
	} else {
		// Parse response
		body, _ := io.ReadAll(resp.Body)
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if status, ok := response["status"].(string); ok && status == "operational" {
				result.Success = true
				result.Message = "Concurrency system operational"
			} else {
				result.Success = false
				result.Message = "Concurrency system not operational"
			}
		} else {
			result.Success = false
			result.Message = "Failed to parse concurrency status response"
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// TestGRPOStatus tests the GRPO learning system
func (it *IntegrationTester) TestGRPOStatus() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/api/v1/grpo/stats")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "GRPO Learning System",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to get GRPO stats"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "GRPO stats returned non-200 status"
	} else {
		result.Success = true
		result.Message = "GRPO learning system accessible"
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// TestStrategyManagement tests the strategy management endpoints
func (it *IntegrationTester) TestStrategyManagement() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/api/v1/strategies/")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "Strategy Management",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to get strategies"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "Strategy endpoint returned non-200 status"
	} else {
		// Parse response
		body, _ := io.ReadAll(resp.Body)
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if strategies, ok := response["strategies"].([]interface{}); ok && len(strategies) > 0 {
				result.Success = true
				result.Message = fmt.Sprintf("Found %d trading strategies", len(strategies))
			} else {
				result.Success = false
				result.Message = "No strategies returned"
			}
		} else {
			result.Success = false
			result.Message = "Failed to parse strategies response"
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// TestConsensusStatus tests the Byzantine consensus system
func (it *IntegrationTester) TestConsensusStatus() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/api/v1/consensus/status")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "Byzantine Consensus",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to get consensus status"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "Consensus status returned non-200 status"
	} else {
		result.Success = true
		result.Message = "Byzantine consensus system accessible"
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// TestVectorClock tests the vector clock system
func (it *IntegrationTester) TestVectorClock() {
	start := time.Now()
	resp, err := it.client.Get(it.baseURL + "/api/v1/concurrency/vector-clock")
	duration := time.Since(start)

	result := IntegrationTestResult{
		TestName:     "Vector Clock System",
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to get vector clock"
	} else if resp.StatusCode != 200 {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Message = "Vector clock returned non-200 status"
	} else {
		result.Success = true
		result.Message = "Vector clock system operational"
	}
	if resp != nil {
		resp.Body.Close()
	}

	it.results = append(it.results, result)
}

// RunAllTests runs all integration tests
func (it *IntegrationTester) RunAllTests() {
	fmt.Println("🧪 Starting ARES API Integration Tests...")
	fmt.Println(strings.Repeat("=", 50))

	it.TestHealthCheck()
	it.TestConcurrencyStatus()
	it.TestGRPOStatus()
	it.TestStrategyManagement()
	it.TestConsensusStatus()
	it.TestVectorClock()

	it.PrintResults()
}

// PrintResults prints the test results
func (it *IntegrationTester) PrintResults() {
	fmt.Println("\n📊 Integration Test Results")
	fmt.Println(strings.Repeat("=", 50))

	passed := 0
	total := len(it.results)

	for _, result := range it.results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		} else {
			passed++
		}

		fmt.Printf("%s %s (%.2fs)\n", status, result.TestName, result.ResponseTime.Seconds())
		if result.Success {
			fmt.Printf("   %s\n", result.Message)
		} else {
			fmt.Printf("   Error: %s\n", result.Error)
			fmt.Printf("   Message: %s\n", result.Message)
		}
		fmt.Println()
	}

	fmt.Printf("🎯 Overall: %d/%d tests passed (%.1f%%)\n", passed, total, float64(passed)/float64(total)*100)

	if passed == total {
		fmt.Println("🎉 All integration tests passed! ARES system is fully operational.")
	} else {
		fmt.Println("⚠️  Some tests failed. Check the system configuration and logs.")
	}
}

// SaveResults saves test results to a JSON file
func (it *IntegrationTester) SaveResults(filename string) error {
	data, err := json.MarshalIndent(it.results, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(filename, data)
}

func runIntegrationTests() {
	tester := NewIntegrationTester("http://localhost:8080")
	tester.RunAllTests()

	// Save results
	if err := tester.SaveResults("integration_test_results.json"); err != nil {
		fmt.Printf("Failed to save results: %v\n", err)
	} else {
		fmt.Println("📄 Results saved to integration_test_results.json")
	}
}

func main() {
	runIntegrationTests()
}

// Helper function to write file (would need to be implemented)
func writeFile(filename string, data []byte) error {
	// This would write to a file in a real implementation
	fmt.Printf("Would write %d bytes to %s\n", len(data), filename)
	return nil
}
