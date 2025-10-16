/**
 * TypeScript interfaces matching backend sandbox_trades schema
 * Based on migration: 004_autonomous_trading_system.sql
 */

/**
 * Core trade data matching sandbox_trades table
 */
export interface SandboxTrade {
  id: number;
  user_id: number;
  session_id: string;
  trading_pair: string;          // e.g., "BTC/USDT", "ETH/USDT"
  direction: 'LONG' | 'SHORT';   // Trade direction
  size: number;                  // Position size
  entry_price: number;           // Entry price
  exit_price: number | null;     // Exit price (null if still open)
  stop_loss: number | null;      // Stop-loss price
  take_profit: number | null;    // Take-profit price
  profit_loss: number | null;    // Realized P&L
  profit_loss_pct: number | null; // P&L percentage
  status: 'OPEN' | 'CLOSED' | 'CANCELLED';
  strategy_type: string | null;  // e.g., "momentum", "mean_reversion"
  indicators_used: string[] | null; // e.g., ["RSI", "MACD", "EMA"]
  reasoning: string;             // SOLACE's explanation
  confidence_score: number;      // 0.0 to 1.0
  sentiment_score: number | null; // Market sentiment
  market_regime: string | null;  // e.g., "BULL", "BEAR", "CHOP", "VOLATILITY_SPIKE"
  regime_confidence: number | null; // Confidence in regime detection
  trade_hash: string;            // SHA256 audit trail
  lineage_trail: any | null;     // JSONB - parent trade references
  solace_override: boolean;      // Human override flag
  override_reason: string | null; // Reason for override
  opened_at: string;             // ISO timestamp
  closed_at: string | null;      // ISO timestamp (null if open)
  created_at: string;            // ISO timestamp
  updated_at: string;            // ISO timestamp
}

/**
 * Trading performance metrics
 */
export interface TradePerformance {
  total_trades: number;
  open_trades: number;
  closed_trades: number;
  total_pnl: number;
  total_pnl_pct: number;
  win_rate: number;              // Percentage (0-100)
  avg_profit: number;
  avg_loss: number;
  sharpe_ratio: number | null;
  max_drawdown: number | null;
  best_trade: number | null;
  worst_trade: number | null;
  avg_hold_time: string | null;  // e.g., "2h 15m"
  daily_pnl: number;
  weekly_pnl: number;
  monthly_pnl: number;
}

/**
 * Live crypto price data
 */
export interface CryptoPrice {
  symbol: string;
  price: number;
  change_24h: number;            // Percentage change
  volume_24h: number;
  market_cap: number;
  last_updated: string;          // ISO timestamp
}

/**
 * Request to execute a new trade
 */
export interface ExecuteTradeRequest {
  trading_pair: string;          // e.g., "BTC/USDT"
  direction: 'LONG' | 'SHORT';
  size: number;
  stop_loss?: number;
  take_profit?: number;
  reasoning?: string;            // Optional manual reasoning
  strategy_type?: string;        // Optional strategy identifier
}

/**
 * Response after executing a trade
 */
export interface ExecuteTradeResponse {
  success: boolean;
  trade_id: number;
  message: string;
  trade: SandboxTrade;
}

/**
 * Decision trace data (Glass Box system)
 */
export interface DecisionTrace {
  id: number;
  trace_id: string;
  trade_id: number | null;
  decision_type: string;         // e.g., "trade_entry", "trade_exit"
  reasoning: string;
  confidence_score: number;
  data_sources: any;             // JSONB
  created_at: string;
}

/**
 * GRPO bias data
 */
export interface GRPOBias {
  id: number;
  token: string;
  bias_value: number;
  sample_count: number;
  avg_reward: number;
  last_updated: string;
}

/**
 * GRPO statistics
 */
export interface GRPOStats {
  total_biases: number;
  total_rewards: number;
  avg_reward: number;
  learning_rate: number;
  last_checkpoint: string | null;
}

/**
 * Chart data point for TradingChart component
 */
export interface ChartDataPoint {
  time: number;                  // Unix timestamp
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

/**
 * Binance kline data (raw format)
 */
export type BinanceKline = [
  number,  // Open time
  string,  // Open
  string,  // High
  string,  // Low
  string,  // Close
  string,  // Volume
  number,  // Close time
  string,  // Quote asset volume
  number,  // Number of trades
  string,  // Taker buy base asset volume
  string,  // Taker buy quote asset volume
  string   // Ignore
];

/**
 * Helper to convert Binance kline to ChartDataPoint
 */
export function binanceKlineToChartData(kline: BinanceKline): ChartDataPoint {
  return {
    time: kline[0],
    open: parseFloat(kline[1]),
    high: parseFloat(kline[2]),
    low: parseFloat(kline[3]),
    close: parseFloat(kline[4]),
    volume: parseFloat(kline[5]),
  };
}

/**
 * Format currency for display
 */
export function formatCurrency(value: number | null | undefined): string {
  if (value === null || value === undefined) return 'N/A';
  return `$${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

/**
 * Format percentage for display
 */
export function formatPercentage(value: number | null | undefined): string {
  if (value === null || value === undefined) return 'N/A';
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}%`;
}

/**
 * Format timestamp to readable date
 */
export function formatDate(timestamp: string | null | undefined): string {
  if (!timestamp) return 'N/A';
  return new Date(timestamp).toLocaleString();
}

/**
 * Calculate unrealized P&L for open position
 */
export function calculateUnrealizedPnL(
  entryPrice: number,
  currentPrice: number,
  size: number,
  direction: 'LONG' | 'SHORT'
): number {
  if (direction === 'LONG') {
    return (currentPrice - entryPrice) * size;
  } else {
    return (entryPrice - currentPrice) * size;
  }
}

/**
 * Get color for P&L display
 */
export function getPnLColor(value: number | null | undefined): string {
  if (value === null || value === undefined) return '#666';
  return value >= 0 ? '#10b981' : '#ef4444';
}
