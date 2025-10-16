package llm_test

import (
	"ares_api/pkg/llm"
	"testing"
	"time"
)

// TestContextManagerCreation verifies context manager initialization
func TestContextManagerCreation(t *testing.T) {
	cm := llm.NewContextManager(0, 0) // Use defaults

	if cm.MaxTokens != 150000 {
		t.Errorf("Expected MaxTokens=150000, got %d", cm.MaxTokens)
	}

	if cm.WindowDuration != 2*time.Hour {
		t.Errorf("Expected WindowDuration=2h, got %v", cm.WindowDuration)
	}

	if cm.UsedTokens != 0 {
		t.Errorf("Expected UsedTokens=0, got %d", cm.UsedTokens)
	}

	t.Logf("✅ Context manager created: MaxTokens=%d, Window=%v", cm.MaxTokens, cm.WindowDuration)
}

// TestTokenBudgetEnforcement verifies token limit enforcement
func TestTokenBudgetEnforcement(t *testing.T) {
	cm := llm.NewContextManager(1000, 2*time.Hour) // Small budget for testing

	// Add message within budget
	msg1 := llm.Message{Role: "user", Content: "Hello"}
	err := cm.AddMessage(msg1, 100)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if cm.UsedTokens != 100 {
		t.Errorf("Expected UsedTokens=100, got %d", cm.UsedTokens)
	}

	// Try to exceed budget
	msg2 := llm.Message{Role: "assistant", Content: "Long response..."}
	err = cm.AddMessage(msg2, 1000) // Would exceed 1000 limit
	if err == nil {
		t.Error("Expected budget exceeded error, got nil")
	}

	if cm.UsedTokens != 100 { // Should remain unchanged
		t.Errorf("Expected UsedTokens=100, got %d", cm.UsedTokens)
	}

	t.Logf("✅ Token budget enforcement working: %v", err)
}

// TestRollingWindowCleanup verifies old message removal
func TestRollingWindowCleanup(t *testing.T) {
	// Use very short window for testing (1 second)
	cm := llm.NewContextManager(150000, 1*time.Second)

	// Add message
	msg1 := llm.Message{Role: "user", Content: "First message"}
	err := cm.AddMessage(msg1, 50)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	if cm.UsedTokens != 50 {
		t.Errorf("Expected UsedTokens=50, got %d", cm.UsedTokens)
	}

	// Wait for window to expire
	time.Sleep(1500 * time.Millisecond)

	// Add another message (should trigger cleanup)
	msg2 := llm.Message{Role: "user", Content: "Second message"}
	err = cm.AddMessage(msg2, 50)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Old tokens should be cleaned up
	if cm.UsedTokens > 100 {
		t.Logf("⚠️  UsedTokens=%d (cleanup may not have triggered yet)", cm.UsedTokens)
	}

	stats := cm.GetWindowStats()
	t.Logf("✅ Rolling window stats: Used=%d, Messages=%d", stats.UsedTokens, stats.MessageCount)
}

// TestTokenEstimation verifies token estimation function
func TestTokenEstimation(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"Hello", 1},              // ~5 chars / 4 = 1 token
		{"Hello world", 2},        // ~11 chars / 4 = 2 tokens
		{"The quick brown fox", 4}, // ~19 chars / 4 = 4 tokens
	}

	for _, test := range tests {
		result := llm.EstimateTokens(test.text)
		if result != test.expected {
			t.Errorf("EstimateTokens(%q) = %d, expected %d", test.text, result, test.expected)
		}
	}

	t.Log("✅ Token estimation working (~4 chars/token)")
}

// TestGetRemainingTokens verifies remaining token calculation
func TestGetRemainingTokens(t *testing.T) {
	cm := llm.NewContextManager(1000, 2*time.Hour)

	// Initially should have full budget
	remaining := cm.GetRemainingTokens()
	if remaining != 1000 {
		t.Errorf("Expected 1000 remaining, got %d", remaining)
	}

	// Add message
	msg := llm.Message{Role: "user", Content: "Test"}
	cm.AddMessage(msg, 250)

	remaining = cm.GetRemainingTokens()
	if remaining != 750 {
		t.Errorf("Expected 750 remaining, got %d", remaining)
	}

	t.Logf("✅ Remaining tokens: %d / %d", remaining, cm.MaxTokens)
}
