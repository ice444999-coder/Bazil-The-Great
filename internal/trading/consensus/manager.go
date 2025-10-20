package consensus

import (
	"fmt"
	"sync"
	"time"

	"ares_api/internal/trading/strategies"
)

// ConsensusManager coordinates between trading strategies and Byzantine consensus
type ConsensusManager struct {
	// Consensus system
	byzantineConsensus *ByzantineConsensus

	// Trading integration
	strategyManager *strategies.StrategyManager

	// Configuration
	minConsensusThreshold float64 // Minimum confidence for consensus proposal
	consensusTimeout      time.Duration
	maxConcurrentTrades   int

	// State
	activeConsensus  map[int64]*ConsensusSession
	pendingDecisions []*ConsensusRequest
	executedTrades   []*ConsensusDecision

	// Callbacks
	onTradeExecuted func(*ConsensusDecision)

	// Thread safety
	mutex sync.RWMutex
}

// ConsensusSession represents an active consensus session
type ConsensusSession struct {
	Request   *ConsensusRequest
	StartTime time.Time
	Timeout   time.Duration
	Status    string // "pending", "consensus_reached", "timeout", "executed"
	Decision  *ConsensusDecision
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager(strategyManager *strategies.StrategyManager, totalNodes int) *ConsensusManager {
	cm := &ConsensusManager{
		byzantineConsensus:    NewByzantineConsensus(0, totalNodes), // Node 0 is primary
		strategyManager:       strategyManager,
		minConsensusThreshold: 0.7, // 70% confidence minimum
		consensusTimeout:      30 * time.Second,
		maxConcurrentTrades:   5,
		activeConsensus:       make(map[int64]*ConsensusSession),
		pendingDecisions:      make([]*ConsensusRequest, 0),
		executedTrades:        make([]*ConsensusDecision, 0),
	}

	// Set consensus callback
	cm.byzantineConsensus.SetOnConsensusReached(cm.handleConsensusReached)

	// Start the consensus system
	cm.byzantineConsensus.Start()

	return cm
}

// ProposeTradeDecision proposes a trading decision for consensus
func (cm *ConsensusManager) ProposeTradeDecision(signal *strategies.TradeSignal) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if we have too many concurrent trades
	if len(cm.activeConsensus) >= cm.maxConcurrentTrades {
		return fmt.Errorf("maximum concurrent trades reached (%d)", cm.maxConcurrentTrades)
	}

	// Check minimum confidence threshold
	if signal.Confidence < cm.minConsensusThreshold {
		return fmt.Errorf("signal confidence %.2f below threshold %.2f",
			signal.Confidence, cm.minConsensusThreshold)
	}

	// Create consensus request
	request := &ConsensusRequest{
		ClientID:   fmt.Sprintf("strategy-%s", signal.Strategy),
		Timestamp:  signal.Timestamp,
		Symbol:     signal.Symbol,
		Action:     signal.Action,
		Confidence: signal.Confidence,
		Reasoning:  signal.Reasoning,
		Strategy:   signal.Strategy,
	}

	// Create consensus session
	session := &ConsensusSession{
		Request:   request,
		StartTime: time.Now(),
		Timeout:   cm.consensusTimeout,
		Status:    "pending",
	}

	// Store session
	cm.activeConsensus[cm.byzantineConsensus.sequenceNumber+1] = session

	// Propose to Byzantine consensus
	err := cm.byzantineConsensus.ProposeDecision(request)
	if err != nil {
		delete(cm.activeConsensus, cm.byzantineConsensus.sequenceNumber+1)
		return fmt.Errorf("failed to propose decision: %v", err)
	}

	return nil
}

// handleConsensusReached is called when Byzantine consensus is reached
func (cm *ConsensusManager) handleConsensusReached(decision *ConsensusDecision) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Find the session
	session, exists := cm.activeConsensus[decision.SequenceNumber]
	if !exists {
		return
	}

	// Update session
	session.Status = "consensus_reached"
	session.Decision = decision

	// Execute the trade
	cm.executeTrade(decision)

	// Move to executed trades
	cm.executedTrades = append(cm.executedTrades, decision)

	// Remove from active
	delete(cm.activeConsensus, decision.SequenceNumber)

	// Notify callback
	if cm.onTradeExecuted != nil {
		go cm.onTradeExecuted(decision)
	}
}

// executeTrade executes a consensus-approved trade
func (cm *ConsensusManager) executeTrade(decision *ConsensusDecision) {
	// In a real implementation, this would:
	// 1. Check risk limits
	// 2. Verify sufficient balance
	// 3. Execute via trading API (Jupiter DEX, etc.)
	// 4. Update position tracking
	// 5. Log the trade

	fmt.Printf("EXECUTING CONSENSUS TRADE: %s %s (Confidence: %.2f, Votes: %d)\n",
		decision.Decision, decision.Symbol, decision.Confidence, decision.Votes)

	// For now, just log the execution
	// TODO: Integrate with actual trading execution
}

// CheckTimeouts checks for timed-out consensus sessions
func (cm *ConsensusManager) CheckTimeouts() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	for seq, session := range cm.activeConsensus {
		if now.Sub(session.StartTime) > session.Timeout {
			session.Status = "timeout"
			delete(cm.activeConsensus, seq)
		}
	}
}

// GetActiveConsensus returns active consensus sessions
func (cm *ConsensusManager) GetActiveConsensus() map[int64]*ConsensusSession {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[int64]*ConsensusSession)
	for k, v := range cm.activeConsensus {
		result[k] = v
	}
	return result
}

// GetExecutedTrades returns executed trades
func (cm *ConsensusManager) GetExecutedTrades() []*ConsensusDecision {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make([]*ConsensusDecision, len(cm.executedTrades))
	copy(result, cm.executedTrades)
	return result
}

// GetConsensusStats returns consensus system statistics
func (cm *ConsensusManager) GetConsensusStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return map[string]interface{}{
		"active_consensus":  len(cm.activeConsensus),
		"executed_trades":   len(cm.executedTrades),
		"pending_decisions": len(cm.pendingDecisions),
		"consensus_system":  cm.byzantineConsensus.GetStatus(),
		"min_threshold":     cm.minConsensusThreshold,
		"consensus_timeout": cm.consensusTimeout.String(),
		"max_concurrent":    cm.maxConcurrentTrades,
	}
}

// SetOnTradeExecuted sets the callback for trade execution
func (cm *ConsensusManager) SetOnTradeExecuted(callback func(*ConsensusDecision)) {
	cm.onTradeExecuted = callback
}

// UpdateConfiguration updates consensus manager configuration
func (cm *ConsensusManager) UpdateConfiguration(config map[string]interface{}) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if threshold, ok := config["min_consensus_threshold"].(float64); ok {
		cm.minConsensusThreshold = threshold
	}

	if timeout, ok := config["consensus_timeout_seconds"].(float64); ok {
		cm.consensusTimeout = time.Duration(timeout) * time.Second
	}

	if maxConcurrent, ok := config["max_concurrent_trades"].(float64); ok {
		cm.maxConcurrentTrades = int(maxConcurrent)
	}
}

// SimulateNetworkDelay simulates network delays for testing
func (cm *ConsensusManager) SimulateNetworkDelay(delay time.Duration) {
	// In a real system, this would affect message delivery
	// For simulation purposes, we could add delays to message processing
}

// Stop stops the consensus manager
func (cm *ConsensusManager) Stop() {
	cm.byzantineConsensus.Stop()
}
