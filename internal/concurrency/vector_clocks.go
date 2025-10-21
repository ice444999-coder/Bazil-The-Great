/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package concurrency

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// ============================================
// ELASTIC VECTOR CLOCKS
// Distributed tracing and causality tracking
// ============================================

// VectorClock represents a logical timestamp for distributed event ordering
type VectorClock struct {
	AgentName string            `json:"agent_name"`
	Clock     map[string]int64  `json:"clock"` // Agent name -> timestamp
	EventID   string            `json:"event_id"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewVectorClock creates a new vector clock for an agent
func NewVectorClock(agentName string) *VectorClock {
	return &VectorClock{
		AgentName: agentName,
		Clock:     make(map[string]int64),
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// Increment increments the clock for this agent
func (vc *VectorClock) Increment() {
	vc.Clock[vc.AgentName]++
	vc.Timestamp = time.Now()
}

// Update merges another vector clock into this one
func (vc *VectorClock) Update(other *VectorClock) {
	for agent, timestamp := range other.Clock {
		if current, exists := vc.Clock[agent]; !exists || timestamp > current {
			vc.Clock[agent] = timestamp
		}
	}
	vc.Timestamp = time.Now()
}

// HappensBefore checks if this clock happened before another clock
func (vc *VectorClock) HappensBefore(other *VectorClock) bool {
	// Check if all timestamps in vc are <= corresponding timestamps in other
	// AND at least one timestamp in vc is < corresponding timestamp in other

	hasLess := false
	for agent, timestamp := range vc.Clock {
		otherTimestamp, exists := other.Clock[agent]
		if !exists {
			otherTimestamp = 0
		}

		if timestamp > otherTimestamp {
			return false
		}
		if timestamp < otherTimestamp {
			hasLess = true
		}
	}

	// Check for agents in other that don't exist in vc
	for agent, otherTimestamp := range other.Clock {
		if _, exists := vc.Clock[agent]; !exists && otherTimestamp > 0 {
			hasLess = true
		}
	}

	return hasLess
}

// Concurrent checks if two clocks are concurrent (neither happened before the other)
func (vc *VectorClock) Concurrent(other *VectorClock) bool {
	return !vc.HappensBefore(other) && !other.HappensBefore(vc)
}

// Equals checks if two vector clocks are identical
func (vc *VectorClock) Equals(other *VectorClock) bool {
	if len(vc.Clock) != len(other.Clock) {
		return false
	}

	for agent, timestamp := range vc.Clock {
		if other.Clock[agent] != timestamp {
			return false
		}
	}

	return true
}

// Copy creates a deep copy of the vector clock
func (vc *VectorClock) Copy() *VectorClock {
	clockCopy := make(map[string]int64)
	for k, v := range vc.Clock {
		clockCopy[k] = v
	}

	metadataCopy := make(map[string]string)
	for k, v := range vc.Metadata {
		metadataCopy[k] = v
	}

	return &VectorClock{
		AgentName: vc.AgentName,
		Clock:     clockCopy,
		EventID:   vc.EventID,
		Timestamp: vc.Timestamp,
		Metadata:  metadataCopy,
	}
}

// String returns a string representation of the vector clock
func (vc *VectorClock) String() string {
	var parts []string
	for agent, ts := range vc.Clock {
		parts = append(parts, fmt.Sprintf("%s:%d", agent, ts))
	}
	sort.Strings(parts)
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// ToJSON serializes the vector clock to JSON
func (vc *VectorClock) ToJSON() ([]byte, error) {
	return json.Marshal(vc)
}

// FromJSON deserializes a vector clock from JSON
func VectorClockFromJSON(data []byte) (*VectorClock, error) {
	var vc VectorClock
	err := json.Unmarshal(data, &vc)
	if err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if vc.Clock == nil {
		vc.Clock = make(map[string]int64)
	}
	if vc.Metadata == nil {
		vc.Metadata = make(map[string]string)
	}

	return &vc, nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// DistributedTrace represents a distributed trace with vector clocks
type DistributedTrace struct {
	TraceID   string            `json:"trace_id"`
	RootEvent *VectorClock      `json:"root_event"`
	Events    []*VectorClock    `json:"events"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Status    TraceStatus       `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TraceStatus represents the status of a distributed trace
type TraceStatus int

const (
	TraceActive TraceStatus = iota
	TraceCompleted
	TraceFailed
)

// NewDistributedTrace creates a new distributed trace
func NewDistributedTrace(traceID string, rootAgent string) *DistributedTrace {
	rootEvent := NewVectorClock(rootAgent)
	rootEvent.Increment()

	return &DistributedTrace{
		TraceID:   traceID,
		RootEvent: rootEvent,
		Events:    []*VectorClock{rootEvent},
		StartTime: time.Now(),
		Status:    TraceActive,
		Metadata:  make(map[string]string),
	}
}

// AddEvent adds a new event to the trace
func (dt *DistributedTrace) AddEvent(agentName string, parentClock *VectorClock) *VectorClock {
	event := NewVectorClock(agentName)
	if parentClock != nil {
		event.Update(parentClock)
	}
	event.Increment()

	dt.Events = append(dt.Events, event)
	return event
}

// Complete marks the trace as completed
func (dt *DistributedTrace) Complete() {
	now := time.Now()
	dt.EndTime = &now
	dt.Status = TraceCompleted
}

// Fail marks the trace as failed
func (dt *DistributedTrace) Fail() {
	now := time.Now()
	dt.EndTime = &now
	dt.Status = TraceFailed
}

// Duration returns the duration of the trace
func (dt *DistributedTrace) Duration() time.Duration {
	if dt.EndTime == nil {
		return time.Since(dt.StartTime)
	}
	return dt.EndTime.Sub(dt.StartTime)
}

// CausalityGraph represents the causality relationships between events
type CausalityGraph struct {
	Events   []*VectorClock         `json:"events"`
	Edges    []CausalityEdge        `json:"edges"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CausalityEdge represents a causality relationship between two events
type CausalityEdge struct {
	FromEventID string `json:"from_event_id"`
	ToEventID   string `json:"to_event_id"`
	Type        string `json:"type"` // "happens_before", "concurrent", "causes"
}

// BuildCausalityGraph builds a causality graph from a set of vector clocks
func BuildCausalityGraph(events []*VectorClock) *CausalityGraph {
	graph := &CausalityGraph{
		Events:   events,
		Edges:    []CausalityEdge{},
		Metadata: make(map[string]interface{}),
	}

	// Build edges based on causality relationships
	for i, event1 := range events {
		for j, event2 := range events {
			if i == j {
				continue
			}

			var edgeType string
			if event1.HappensBefore(event2) {
				edgeType = "happens_before"
			} else if event2.HappensBefore(event1) {
				edgeType = "happens_before"
				// Swap to maintain direction
				event1, event2 = event2, event1
			} else if event1.Concurrent(event2) {
				edgeType = "concurrent"
			} else {
				continue // Same event or identical
			}

			edge := CausalityEdge{
				FromEventID: event1.EventID,
				ToEventID:   event2.EventID,
				Type:        edgeType,
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	return graph
}

// LamportClock implements Lamport logical clocks for simpler causality
type LamportClock struct {
	counter *AtomicCounter
}

// NewLamportClock creates a new Lamport clock
func NewLamportClock() *LamportClock {
	return &LamportClock{
		counter: NewAtomicCounter(0),
	}
}

// Tick increments the clock and returns the new value
func (lc *LamportClock) Tick() int64 {
	return lc.counter.Increment()
}

// Update updates the clock with a received timestamp
func (lc *LamportClock) Update(receivedTimestamp int64) int64 {
	for {
		current := lc.counter.Load()
		newValue := current
		if receivedTimestamp > current {
			newValue = receivedTimestamp
		}
		newValue++ // Always increment after update

		if lc.counter.CompareAndSwap(current, newValue) {
			return newValue
		}
	}
}

// Current returns the current clock value
func (lc *LamportClock) Current() int64 {
	return lc.counter.Load()
}

// VersionVector implements version vectors for conflict detection
type VersionVector struct {
	versions map[string]int64
}

// NewVersionVector creates a new version vector
func NewVersionVector() *VersionVector {
	return &VersionVector{
		versions: make(map[string]int64),
	}
}

// Increment increments the version for a replica
func (vv *VersionVector) Increment(replica string) {
	vv.versions[replica]++
}

// Update merges another version vector
func (vv *VersionVector) Update(other *VersionVector) {
	for replica, version := range other.versions {
		if current, exists := vv.versions[replica]; !exists || version > current {
			vv.versions[replica] = version
		}
	}
}

// Compare compares this version vector with another
func (vv *VersionVector) Compare(other *VersionVector) VersionComparison {
	thisDominates := true
	otherDominates := true

	for replica, version := range vv.versions {
		otherVersion, exists := other.versions[replica]
		if !exists {
			otherVersion = 0
		}

		if version > otherVersion {
			otherDominates = false
		} else if version < otherVersion {
			thisDominates = false
		}
	}

	for replica, otherVersion := range other.versions {
		if _, exists := vv.versions[replica]; !exists {
			if otherVersion > 0 {
				thisDominates = false
			}
		}
	}

	if thisDominates && otherDominates {
		return VersionEqual
	} else if thisDominates {
		return VersionDominates
	} else if otherDominates {
		return VersionDominated
	} else {
		return VersionConcurrent
	}
}

// VersionComparison represents the result of comparing two version vectors
type VersionComparison int

const (
	VersionEqual      VersionComparison = iota
	VersionDominates                    // This version vector dominates the other
	VersionDominated                    // This version vector is dominated by the other
	VersionConcurrent                   // Version vectors are concurrent
)

// String returns a string representation of the version vector
func (vv *VersionVector) String() string {
	var parts []string
	for replica, version := range vv.versions {
		parts = append(parts, fmt.Sprintf("%s:%d", replica, version))
	}
	sort.Strings(parts)
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
