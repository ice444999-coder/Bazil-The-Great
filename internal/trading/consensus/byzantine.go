package consensus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ByzantineConsensus implements Practical Byzantine Fault Tolerance (PBFT)
// for multi-agent trading decision consensus
// Fault tolerance: f = (n-1)/3 where n = total agents, f = max faulty agents
type ByzantineConsensus struct {
	// Configuration
	nodeID      int // This node's ID (0 to n-1)
	totalNodes  int // Total number of nodes in the system
	faultyNodes int // Maximum faulty nodes (f = (n-1)/3)
	currentView int // Current view number
	primaryNode int // Current primary node ID

	// State
	sequenceNumber int64                         // Monotonically increasing sequence number
	log            map[string]*ConsensusLogEntry // Sequence -> Log entry
	prepared       map[string]bool               // Sequence -> Prepared status
	committed      map[string]bool               // Sequence -> Committed status

	// Message buffers
	prePrepareMsgs map[string][]*PrePrepareMessage
	prepareMsgs    map[string]map[int]*PrepareMessage // Sequence -> NodeID -> Message
	commitMsgs     map[string]map[int]*CommitMessage  // Sequence -> NodeID -> Message

	// Callbacks
	onConsensusReached func(*ConsensusDecision) // Called when consensus is reached

	// Thread safety
	mutex sync.RWMutex

	// Network simulation (in real implementation, this would be network messaging)
	messageQueue chan ConsensusMessage
	nodes        []*ConsensusNode // All nodes in the system
}

// ConsensusNode represents a node in the Byzantine consensus system
type ConsensusNode struct {
	ID            int
	IsFaulty      bool
	Consensus     *ByzantineConsensus
	MessageBuffer []ConsensusMessage
}

// ConsensusMessage represents any message in the PBFT protocol
type ConsensusMessage interface {
	GetType() string
	GetSequence() int64
	GetNodeID() int
	GetView() int
}

// ConsensusDecision represents a trading decision that reached consensus
type ConsensusDecision struct {
	SequenceNumber int64
	View           int
	Decision       string // "buy", "sell", "hold"
	Symbol         string
	Confidence     float64
	Reasoning      string
	Timestamp      time.Time
	Votes          int // Number of nodes that agreed
}

// ConsensusLogEntry represents an entry in the consensus log
type ConsensusLogEntry struct {
	SequenceNumber int64
	View           int
	Request        *ConsensusRequest
	PrePrepare     *PrePrepareMessage
	PrepareCount   int
	CommitCount    int
	Decision       *ConsensusDecision
	Status         string // "pre-prepared", "prepared", "committed"
}

// ConsensusRequest represents a trading decision request to be consensus-ed
type ConsensusRequest struct {
	ClientID   string    `json:"client_id"`
	Timestamp  time.Time `json:"timestamp"`
	Symbol     string    `json:"symbol"`
	Action     string    `json:"action"` // "buy", "sell", "hold"
	Confidence float64   `json:"confidence"`
	Reasoning  string    `json:"reasoning"`
	Strategy   string    `json:"strategy"`
}

// PrePrepareMessage represents the pre-prepare phase message
type PrePrepareMessage struct {
	Type           string            `json:"type"`
	View           int               `json:"view"`
	SequenceNumber int64             `json:"sequence_number"`
	RequestDigest  string            `json:"request_digest"`
	NodeID         int               `json:"node_id"`
	Request        *ConsensusRequest `json:"request"`
}

func (m *PrePrepareMessage) GetType() string    { return "pre-prepare" }
func (m *PrePrepareMessage) GetSequence() int64 { return m.SequenceNumber }
func (m *PrePrepareMessage) GetNodeID() int     { return m.NodeID }
func (m *PrePrepareMessage) GetView() int       { return m.View }

// PrepareMessage represents the prepare phase message
type PrepareMessage struct {
	Type           string `json:"type"`
	View           int    `json:"view"`
	SequenceNumber int64  `json:"sequence_number"`
	RequestDigest  string `json:"request_digest"`
	NodeID         int    `json:"node_id"`
}

func (m *PrepareMessage) GetType() string    { return "prepare" }
func (m *PrepareMessage) GetSequence() int64 { return m.SequenceNumber }
func (m *PrepareMessage) GetNodeID() int     { return m.NodeID }
func (m *PrepareMessage) GetView() int       { return m.View }

// CommitMessage represents the commit phase message
type CommitMessage struct {
	Type           string `json:"type"`
	View           int    `json:"view"`
	SequenceNumber int64  `json:"sequence_number"`
	RequestDigest  string `json:"request_digest"`
	NodeID         int    `json:"node_id"`
}

func (m *CommitMessage) GetType() string    { return "commit" }
func (m *CommitMessage) GetSequence() int64 { return m.SequenceNumber }
func (m *CommitMessage) GetNodeID() int     { return m.NodeID }
func (m *CommitMessage) GetView() int       { return m.View }

// NewByzantineConsensus creates a new Byzantine consensus instance
func NewByzantineConsensus(nodeID, totalNodes int) *ByzantineConsensus {
	faultyNodes := (totalNodes - 1) / 3

	bc := &ByzantineConsensus{
		nodeID:         nodeID,
		totalNodes:     totalNodes,
		faultyNodes:    faultyNodes,
		currentView:    0,
		primaryNode:    0, // Primary is always node 0 in view 0
		sequenceNumber: 0,
		log:            make(map[string]*ConsensusLogEntry),
		prepared:       make(map[string]bool),
		committed:      make(map[string]bool),
		prePrepareMsgs: make(map[string][]*PrePrepareMessage),
		prepareMsgs:    make(map[string]map[int]*PrepareMessage),
		commitMsgs:     make(map[string]map[int]*CommitMessage),
		messageQueue:   make(chan ConsensusMessage, 1000),
		nodes:          make([]*ConsensusNode, totalNodes),
	}

	// Initialize nodes
	for i := 0; i < totalNodes; i++ {
		bc.nodes[i] = &ConsensusNode{
			ID:            i,
			IsFaulty:      false, // Can be set later for testing
			Consensus:     bc,
			MessageBuffer: make([]ConsensusMessage, 0),
		}
	}

	return bc
}

// Start begins the consensus protocol
func (bc *ByzantineConsensus) Start() {
	go bc.messageProcessor()
}

// Stop halts the consensus protocol
func (bc *ByzantineConsensus) Stop() {
	close(bc.messageQueue)
}

// ProposeDecision proposes a trading decision for consensus
func (bc *ByzantineConsensus) ProposeDecision(request *ConsensusRequest) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Only primary can propose
	if bc.nodeID != bc.primaryNode {
		return fmt.Errorf("only primary node can propose decisions")
	}

	// Increment sequence number
	bc.sequenceNumber++

	// Create digest of the request
	digest := bc.hashRequest(request)

	// Create pre-prepare message
	prePrepare := &PrePrepareMessage{
		Type:           "pre-prepare",
		View:           bc.currentView,
		SequenceNumber: bc.sequenceNumber,
		RequestDigest:  digest,
		NodeID:         bc.nodeID,
		Request:        request,
	}

	// Log the entry
	key := bc.makeLogKey(bc.sequenceNumber)
	bc.log[key] = &ConsensusLogEntry{
		SequenceNumber: bc.sequenceNumber,
		View:           bc.currentView,
		Request:        request,
		PrePrepare:     prePrepare,
		Status:         "pre-prepared",
	}

	// Broadcast pre-prepare message to all nodes
	bc.broadcastMessage(prePrepare)

	return nil
}

// HandleMessage processes incoming consensus messages
func (bc *ByzantineConsensus) HandleMessage(msg ConsensusMessage) {
	bc.messageQueue <- msg
}

// messageProcessor processes messages from the queue
func (bc *ByzantineConsensus) messageProcessor() {
	for msg := range bc.messageQueue {
		bc.processMessage(msg)
	}
}

// processMessage handles different types of consensus messages
func (bc *ByzantineConsensus) processMessage(msg ConsensusMessage) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	switch m := msg.(type) {
	case *PrePrepareMessage:
		bc.handlePrePrepare(m)
	case *PrepareMessage:
		bc.handlePrepare(m)
	case *CommitMessage:
		bc.handleCommit(m)
	}
}

// handlePrePrepare processes pre-prepare messages
func (bc *ByzantineConsensus) handlePrePrepare(msg *PrePrepareMessage) {
	// Verify the message is from primary
	if msg.NodeID != bc.primaryNode {
		return // Ignore messages not from primary
	}

	// Verify digest
	if msg.RequestDigest != bc.hashRequest(msg.Request) {
		return // Invalid digest
	}

	key := bc.makeLogKey(msg.SequenceNumber)

	// Create log entry if it doesn't exist
	if _, exists := bc.log[key]; !exists {
		bc.log[key] = &ConsensusLogEntry{
			SequenceNumber: msg.SequenceNumber,
			View:           msg.View,
			Request:        msg.Request,
			PrePrepare:     msg,
			Status:         "pre-prepared",
		}
	}

	// Store pre-prepare message
	seqKey := fmt.Sprintf("%d", msg.SequenceNumber)
	bc.prePrepareMsgs[seqKey] = append(bc.prePrepareMsgs[seqKey], msg)

	// Send prepare message
	prepareMsg := &PrepareMessage{
		Type:           "prepare",
		View:           msg.View,
		SequenceNumber: msg.SequenceNumber,
		RequestDigest:  msg.RequestDigest,
		NodeID:         bc.nodeID,
	}

	bc.broadcastMessage(prepareMsg)
}

// handlePrepare processes prepare messages
func (bc *ByzantineConsensus) handlePrepare(msg *PrepareMessage) {
	seqKey := fmt.Sprintf("%d", msg.SequenceNumber)

	// Initialize prepare messages map for this sequence
	if bc.prepareMsgs[seqKey] == nil {
		bc.prepareMsgs[seqKey] = make(map[int]*PrepareMessage)
	}

	// Store prepare message
	bc.prepareMsgs[seqKey][msg.NodeID] = msg

	// Check if we have 2f + 1 prepare messages (including our own)
	prepareCount := len(bc.prepareMsgs[seqKey])
	if prepareCount >= 2*bc.faultyNodes+1 {
		key := bc.makeLogKey(msg.SequenceNumber)
		if entry, exists := bc.log[key]; exists {
			entry.PrepareCount = prepareCount
			entry.Status = "prepared"
			bc.prepared[key] = true

			// Send commit message
			commitMsg := &CommitMessage{
				Type:           "commit",
				View:           msg.View,
				SequenceNumber: msg.SequenceNumber,
				RequestDigest:  msg.RequestDigest,
				NodeID:         bc.nodeID,
			}

			bc.broadcastMessage(commitMsg)
		}
	}
}

// handleCommit processes commit messages
func (bc *ByzantineConsensus) handleCommit(msg *CommitMessage) {
	seqKey := fmt.Sprintf("%d", msg.SequenceNumber)

	// Initialize commit messages map for this sequence
	if bc.commitMsgs[seqKey] == nil {
		bc.commitMsgs[seqKey] = make(map[int]*CommitMessage)
	}

	// Store commit message
	bc.commitMsgs[seqKey][msg.NodeID] = msg

	// Check if we have 2f + 1 commit messages (including our own)
	commitCount := len(bc.commitMsgs[seqKey])
	if commitCount >= 2*bc.faultyNodes+1 {
		key := bc.makeLogKey(msg.SequenceNumber)
		if entry, exists := bc.log[key]; exists {
			entry.CommitCount = commitCount
			entry.Status = "committed"
			bc.committed[key] = true

			// Consensus reached! Create decision
			decision := &ConsensusDecision{
				SequenceNumber: entry.SequenceNumber,
				View:           entry.View,
				Decision:       entry.Request.Action,
				Symbol:         entry.Request.Symbol,
				Confidence:     entry.Request.Confidence,
				Reasoning:      entry.Request.Reasoning,
				Timestamp:      time.Now(),
				Votes:          commitCount,
			}

			entry.Decision = decision

			// Notify callback
			if bc.onConsensusReached != nil {
				go bc.onConsensusReached(decision)
			}
		}
	}
}

// broadcastMessage sends a message to all nodes
func (bc *ByzantineConsensus) broadcastMessage(msg ConsensusMessage) {
	// In a real distributed system, this would send over network
	// For simulation, we deliver to all local nodes
	for _, node := range bc.nodes {
		if node.ID != bc.nodeID { // Don't send to self
			node.MessageBuffer = append(node.MessageBuffer, msg)
		}
	}
}

// hashRequest creates a SHA256 hash of the request
func (bc *ByzantineConsensus) hashRequest(request *ConsensusRequest) string {
	data, _ := json.Marshal(request)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// makeLogKey creates a unique key for log entries
func (bc *ByzantineConsensus) makeLogKey(sequence int64) string {
	return fmt.Sprintf("%d-%d", bc.currentView, sequence)
}

// SetOnConsensusReached sets the callback for when consensus is reached
func (bc *ByzantineConsensus) SetOnConsensusReached(callback func(*ConsensusDecision)) {
	bc.onConsensusReached = callback
}

// GetStatus returns the current status of the consensus system
func (bc *ByzantineConsensus) GetStatus() map[string]interface{} {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return map[string]interface{}{
		"node_id":         bc.nodeID,
		"total_nodes":     bc.totalNodes,
		"faulty_nodes":    bc.faultyNodes,
		"current_view":    bc.currentView,
		"primary_node":    bc.primaryNode,
		"sequence_number": bc.sequenceNumber,
		"log_entries":     len(bc.log),
		"prepared_count":  len(bc.prepared),
		"committed_count": len(bc.committed),
	}
}

// SimulateFaultyNode marks a node as faulty for testing
func (bc *ByzantineConsensus) SimulateFaultyNode(nodeID int, faulty bool) {
	if nodeID >= 0 && nodeID < len(bc.nodes) {
		bc.nodes[nodeID].IsFaulty = faulty
	}
}
