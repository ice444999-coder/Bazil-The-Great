package strategies

import "time"

// MarketData contains current market state and historical data
type MarketData struct {
	Symbol        string
	CurrentPrice  float64
	PriceHistory  []float64 // Recent prices (e.g., last 100 candles)
	VolumeHistory []float64 // Recent volumes
	CurrentVolume float64
	BidAskSpread  float64
	OrderBookBids []OrderBookEntry
	OrderBookAsks []OrderBookEntry
	Timestamp     time.Time
	TimeFrame     string // "1m", "5m", "15m", "1h", etc.
}

// OrderBookEntry represents a single level in the order book
type OrderBookEntry struct {
	Price  float64
	Amount float64
}

// TradeSignal represents a trading opportunity identified by a strategy
type TradeSignal struct {
	Action      string // "buy", "sell", "hold"
	Symbol      string
	Confidence  float64 // 0.0 to 1.0
	Reasoning   string
	Strategy    string
	Timestamp   time.Time
	TargetGain  float64 // Expected gain as decimal (e.g., 0.05 = 5%)
	StopLoss    float64 // Stop loss as decimal (e.g., -0.02 = -2%)
	MaxHoldTime int     // Maximum hold time in seconds
	Priority    int     // 1-10, higher = more urgent
}

// StrategyAnalysis contains strategy evaluation of market conditions
type StrategyAnalysis struct {
	StrategyName   string
	Score          float64 // 0.0 to 1.0, how favorable are conditions
	Indicators     map[string]float64
	Recommendation string // "STRONG_SIGNAL", "MODERATE_SIGNAL", "WEAK_SIGNAL", "NO_SIGNAL"
	Timestamp      time.Time
}

// Strategy interface that all trading strategies must implement
type Strategy interface {
	// Generate creates a trade signal based on current market data
	Generate(marketData *MarketData) (*TradeSignal, error)

	// Analyze evaluates market conditions without generating a signal
	Analyze(marketData *MarketData) *StrategyAnalysis

	// GetConfig returns current strategy configuration
	GetConfig() map[string]interface{}

	// UpdateConfig updates strategy parameters (for optimization)
	UpdateConfig(params map[string]interface{}) error
}

// StrategyPerformance tracks how well a strategy is performing
type StrategyPerformance struct {
	StrategyName    string
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	WinRate         float64
	TotalProfitLoss float64
	AvgProfit       float64
	AvgLoss         float64
	SharpeRatio     float64
	MaxDrawdown     float64
	ProfitFactor    float64 // Gross profit / Gross loss
	LastUpdated     time.Time
}

// WhaleTransaction represents a large transaction detected on-chain
type WhaleTransaction struct {
	TxHash        string
	WalletAddress string
	Symbol        string
	Amount        float64 // In USD
	Direction     string  // "buy" or "sell"
	Timestamp     time.Time
	Exchange      string  // If known
	Confidence    float64 // How confident we are this is a whale move
}

// ConsensusVote represents an agent's vote on a trade proposal
type ConsensusVote struct {
	AgentName  string
	TradeID    string
	Vote       string // "approve", "reject", "abstain"
	Confidence float64
	Reasoning  string
	Timestamp  time.Time
}

// CapitalAllocation represents how capital is distributed among strategies
type CapitalAllocation struct {
	StrategyName      string
	AllocatedCapital  float64
	AllocationPercent float64
	LastRebalance     time.Time
	PerformanceScore  float64 // Used for game-theoretic allocation
}

// FaultRecoveryLog tracks retry attempts for failed trades
type FaultRecoveryLog struct {
	TradeID       string
	AttemptNumber int
	ErrorType     string
	ErrorMessage  string
	BackoffDelay  int // Seconds waited before retry
	Success       bool
	Timestamp     time.Time
}

// VectorClock represents a logical timestamp for distributed event ordering
type VectorClock struct {
	AgentName string
	Clock     map[string]int64 // Agent name -> timestamp
	EventID   string
	Timestamp time.Time
}
