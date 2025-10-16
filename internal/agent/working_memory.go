package agent

import (
	"time"
)

// NewWorkingMemory creates a new working memory buffer
func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{
		RecentEvents:    make([]*Event, 0, 100),
		RecentDecisions: make([]*Decision, 0, 100),
		ActiveContext:   make(map[string]interface{}),
	}
}

// AddEvent stores a new event in working memory
func (wm *WorkingMemory) AddEvent(event *Event) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.RecentEvents = append(wm.RecentEvents, event)

	// Keep only last 2 hours
	cutoff := time.Now().Add(-2 * time.Hour)
	for i, e := range wm.RecentEvents {
		if e.Timestamp.After(cutoff) {
			wm.RecentEvents = wm.RecentEvents[i:]
			break
		}
	}
}

// AddDecision stores a new decision in working memory
func (wm *WorkingMemory) AddDecision(decision *Decision) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.RecentDecisions = append(wm.RecentDecisions, decision)

	// Keep only last 50 decisions
	if len(wm.RecentDecisions) > 50 {
		wm.RecentDecisions = wm.RecentDecisions[len(wm.RecentDecisions)-50:]
	}
}

// GetRecentDecisions returns last N decisions
func (wm *WorkingMemory) GetRecentDecisions(n int) []*Decision {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if n > len(wm.RecentDecisions) {
		n = len(wm.RecentDecisions)
	}

	if n == 0 {
		return []*Decision{}
	}

	return wm.RecentDecisions[len(wm.RecentDecisions)-n:]
}

// GetRecentEvents returns last N events
func (wm *WorkingMemory) GetRecentEvents(n int) []*Event {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if n > len(wm.RecentEvents) {
		n = len(wm.RecentEvents)
	}

	if n == 0 {
		return []*Event{}
	}

	return wm.RecentEvents[len(wm.RecentEvents)-n:]
}

// SetLastPrice stores the last known price for a symbol
func (wm *WorkingMemory) SetLastPrice(symbol string, price float64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.ActiveContext == nil {
		wm.ActiveContext = make(map[string]interface{})
	}

	wm.ActiveContext["last_price_"+symbol] = price
}

// GetLastPrice retrieves the last known price for a symbol
func (wm *WorkingMemory) GetLastPrice(symbol string) float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.ActiveContext == nil {
		return 0
	}

	if price, ok := wm.ActiveContext["last_price_"+symbol].(float64); ok {
		return price
	}

	return 0
}

// Summary returns a human-readable summary of working memory
func (wm *WorkingMemory) Summary() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	summary := "Recent context:\n"

	// Show recent events
	if len(wm.RecentEvents) > 0 {
		summary += "Recent events: "
		count := len(wm.RecentEvents)
		if count > 3 {
			count = 3
		}
		for i := len(wm.RecentEvents) - count; i < len(wm.RecentEvents); i++ {
			summary += wm.RecentEvents[i].Description + "; "
		}
		summary += "\n"
	}

	// Show recent decisions
	if len(wm.RecentDecisions) > 0 {
		summary += "Recent decisions: "
		count := len(wm.RecentDecisions)
		if count > 3 {
			count = 3
		}
		for i := len(wm.RecentDecisions) - count; i < len(wm.RecentDecisions); i++ {
			summary += string(wm.RecentDecisions[i].Action.Type) + "; "
		}
	}

	return summary
}

// Clear removes old memories beyond retention window
func (wm *WorkingMemory) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	cutoff := time.Now().Add(-2 * time.Hour)

	// Clear old events
	newEvents := make([]*Event, 0)
	for _, e := range wm.RecentEvents {
		if e.Timestamp.After(cutoff) {
			newEvents = append(newEvents, e)
		}
	}
	wm.RecentEvents = newEvents

	// Keep last 50 decisions regardless of time
	if len(wm.RecentDecisions) > 50 {
		wm.RecentDecisions = wm.RecentDecisions[len(wm.RecentDecisions)-50:]
	}
}
