package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ThoughtJournal logs SOLACE's internal reasoning to a file
// This provides transparency into decision-making
type ThoughtJournal struct {
	FilePath string
	mu       sync.Mutex
}

// NewThoughtJournal creates a new thought journal
func NewThoughtJournal(workspaceRoot string) *ThoughtJournal {
	// Create journal directory if it doesn't exist
	journalDir := filepath.Join(workspaceRoot, "SOLACE_Journal")
	os.MkdirAll(journalDir, 0755)
	
	// Create daily journal file
	today := time.Now().Format("2006-01-02")
	journalPath := filepath.Join(journalDir, fmt.Sprintf("SOLACE_Thoughts_%s.log", today))
	
	tj := &ThoughtJournal{
		FilePath: journalPath,
	}
	
	// Write header if new file
	if _, err := os.Stat(journalPath); os.IsNotExist(err) {
		tj.Write("=" + string(make([]byte, 78)) + "=")
		tj.Write(fmt.Sprintf(" SOLACE Thought Journal - %s", today))
		tj.Write("=" + string(make([]byte, 78)) + "=")
		tj.Write("")
	}
	
	return tj
}

// Write logs a thought to the journal
func (tj *ThoughtJournal) Write(thought string) {
	tj.mu.Lock()
	defer tj.mu.Unlock()
	
	timestamp := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] %s\n", timestamp, thought)
	
	f, err := os.OpenFile(tj.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	
	f.WriteString(entry)
}

// WriteSection writes a section header
func (tj *ThoughtJournal) WriteSection(title string) {
	tj.mu.Lock()
	defer tj.mu.Unlock()
	
	timestamp := time.Now().Format("15:04:05")
	
	f, err := os.OpenFile(tj.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	
	f.WriteString("\n")
	f.WriteString(fmt.Sprintf("[%s] ─────────────────────────────────────────────────────────────────────\n", timestamp))
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, title))
	f.WriteString(fmt.Sprintf("[%s] ─────────────────────────────────────────────────────────────────────\n", timestamp))
}

// GetTodaysThoughts returns the contents of today's journal
func (tj *ThoughtJournal) GetTodaysThoughts() (string, error) {
	data, err := os.ReadFile(tj.FilePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
