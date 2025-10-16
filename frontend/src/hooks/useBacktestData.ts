import { useState } from 'react';

interface BacktestParams {
  symbol: string;
  interval: string; // '1m', '5m', '15m', '1h', '4h', '1d'
  startDate: string; // ISO date string
  endDate: string; // ISO date string
  strategy: BacktestStrategy;
}

interface BacktestStrategy {
  type: 'sma_crossover' | 'rsi_oversold' | 'macd_crossover' | 'bollinger_bounce' | 'solace_mimicry';
  params?: {
    fastPeriod?: number;
    slowPeriod?: number;
    rsiPeriod?: number;
    rsiOversold?: number;
    rsiOverbought?: number;
    macdFast?: number;
    macdSlow?: number;
    macdSignal?: number;
    bbPeriod?: number;
    bbStdDev?: number;
    stopLossPercent?: number;
    takeProfitPercent?: number;
  };
}

interface BacktestTrade {
  entryTime: number;
  entryPrice: number;
  exitTime: number;
  exitPrice: number;
  direction: 'long' | 'short';
  profitLoss: number;
  profitLossPercent: number;
  reason: string;
}

interface BacktestResults {
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRate: number;
  totalProfitLoss: number;
  totalProfitLossPercent: number;
  averageWin: number;
  averageLoss: number;
  largestWin: number;
  largestLoss: number;
  profitFactor: number;
  sharpeRatio: number;
  maxDrawdown: number;
  maxDrawdownPercent: number;
  trades: BacktestTrade[];
  equityCurve: Array<{ time: number; equity: number }>;
  startingCapital: number;
  endingCapital: number;
}

interface UseBacktestDataReturn {
  runBacktest: (params: BacktestParams) => Promise<void>;
  results: BacktestResults | null;
  loading: boolean;
  error: string | null;
  progress: number;
}

export default function useBacktestData(): UseBacktestDataReturn {
  const [results, setResults] = useState<BacktestResults | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState(0);

  const runBacktest = async (params: BacktestParams): Promise<void> => {
    setLoading(true);
    setError(null);
    setProgress(0);
    setResults(null);

    try {
      // Step 1: Fetch historical data from Binance (20%)
      setProgress(20);
      const historicalData = await fetchHistoricalData(
        params.symbol,
        params.interval,
        params.startDate,
        params.endDate
      );

      if (historicalData.length === 0) {
        throw new Error('No historical data available for selected period');
      }

      // Step 2: Apply strategy and simulate trades (60%)
      setProgress(60);
      const backtestResults = await simulateTrades(historicalData, params.strategy);

      // Step 3: Calculate metrics (80%)
      setProgress(80);
      const metrics = calculateMetrics(backtestResults, 10000); // Starting capital $10,000

      // Step 4: Complete (100%)
      setProgress(100);
      setResults(metrics);
    } catch (err: any) {
      console.error('Backtest error:', err);
      setError(err.message || 'Failed to run backtest');
    } finally {
      setLoading(false);
    }
  };

  return { runBacktest, results, loading, error, progress };
}

// Fetch historical OHLCV data from Binance
async function fetchHistoricalData(
  symbol: string,
  interval: string,
  startDate: string,
  endDate: string
): Promise<Array<{ time: number; open: number; high: number; low: number; close: number; volume: number }>> {
  const binanceSymbol = symbol.replace('/', '').toUpperCase();
  const startTime = new Date(startDate).getTime();
  const endTime = new Date(endDate).getTime();

  // Binance API limits to 1000 candles per request
  const limit = 1000;
  let allData: any[] = [];
  let currentTime = startTime;

  while (currentTime < endTime) {
    const url = `https://api.binance.com/api/v3/klines?symbol=${binanceSymbol}&interval=${interval}&startTime=${currentTime}&limit=${limit}`;
    
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Binance API error: ${response.status}`);
    }

    const data = await response.json();
    if (data.length === 0) break;

    allData = allData.concat(data);

    // Move to next batch
    currentTime = data[data.length - 1][0] + 1;

    // Prevent fetching beyond end date
    if (currentTime >= endTime) break;

    // Rate limiting: wait 100ms between requests
    await new Promise((resolve) => setTimeout(resolve, 100));
  }

  // Parse to readable format
  return allData
    .filter((candle) => candle[0] <= endTime)
    .map((candle) => ({
      time: candle[0],
      open: parseFloat(candle[1]),
      high: parseFloat(candle[2]),
      low: parseFloat(candle[3]),
      close: parseFloat(candle[4]),
      volume: parseFloat(candle[5]),
    }));
}

// Simulate trades based on strategy
async function simulateTrades(
  data: Array<{ time: number; open: number; high: number; low: number; close: number; volume: number }>,
  strategy: BacktestStrategy
): Promise<BacktestTrade[]> {
  const trades: BacktestTrade[] = [];

  // Simple SMA Crossover strategy example
  if (strategy.type === 'sma_crossover') {
    const fastPeriod = strategy.params?.fastPeriod || 10;
    const slowPeriod = strategy.params?.slowPeriod || 30;
    const stopLoss = strategy.params?.stopLossPercent || 5;
    const takeProfit = strategy.params?.takeProfitPercent || 10;

    let position: { entry: number; entryPrice: number; direction: 'long' | 'short' } | null = null;

    for (let i = slowPeriod; i < data.length; i++) {
      const fastSMA = calculateSMA(data, i, fastPeriod);
      const slowSMA = calculateSMA(data, i, slowPeriod);
      const prevFastSMA = calculateSMA(data, i - 1, fastPeriod);
      const prevSlowSMA = calculateSMA(data, i - 1, slowPeriod);

      // Entry signal: Fast SMA crosses above Slow SMA
      if (!position && prevFastSMA <= prevSlowSMA && fastSMA > slowSMA) {
        position = {
          entry: data[i].time,
          entryPrice: data[i].close,
          direction: 'long',
        };
      }

      // Exit signal: Stop loss, take profit, or cross back
      if (position) {
        const currentPrice = data[i].close;
        const pnlPercent = ((currentPrice - position.entryPrice) / position.entryPrice) * 100;

        let exitReason = '';
        let shouldExit = false;

        if (pnlPercent <= -stopLoss) {
          exitReason = 'Stop Loss';
          shouldExit = true;
        } else if (pnlPercent >= takeProfit) {
          exitReason = 'Take Profit';
          shouldExit = true;
        } else if (prevFastSMA >= prevSlowSMA && fastSMA < slowSMA) {
          exitReason = 'SMA Cross Exit';
          shouldExit = true;
        }

        if (shouldExit) {
          trades.push({
            entryTime: position.entry,
            entryPrice: position.entryPrice,
            exitTime: data[i].time,
            exitPrice: currentPrice,
            direction: position.direction,
            profitLoss: currentPrice - position.entryPrice,
            profitLossPercent: pnlPercent,
            reason: exitReason,
          });
          position = null;
        }
      }
    }

    // Close open position at end if still open
    if (position) {
      const lastCandle = data[data.length - 1];
      const pnlPercent = ((lastCandle.close - position.entryPrice) / position.entryPrice) * 100;
      trades.push({
        entryTime: position.entry,
        entryPrice: position.entryPrice,
        exitTime: lastCandle.time,
        exitPrice: lastCandle.close,
        direction: position.direction,
        profitLoss: lastCandle.close - position.entryPrice,
        profitLossPercent: pnlPercent,
        reason: 'End of Period',
      });
    }
  }

  return trades;
}

// Calculate Simple Moving Average
function calculateSMA(
  data: Array<{ close: number }>,
  index: number,
  period: number
): number {
  if (index < period - 1) return 0;
  
  let sum = 0;
  for (let i = 0; i < period; i++) {
    sum += data[index - i].close;
  }
  return sum / period;
}

// Calculate backtest metrics
function calculateMetrics(trades: BacktestTrade[], startingCapital: number): BacktestResults {
  if (trades.length === 0) {
    return {
      totalTrades: 0,
      winningTrades: 0,
      losingTrades: 0,
      winRate: 0,
      totalProfitLoss: 0,
      totalProfitLossPercent: 0,
      averageWin: 0,
      averageLoss: 0,
      largestWin: 0,
      largestLoss: 0,
      profitFactor: 0,
      sharpeRatio: 0,
      maxDrawdown: 0,
      maxDrawdownPercent: 0,
      trades: [],
      equityCurve: [],
      startingCapital,
      endingCapital: startingCapital,
    };
  }

  const winningTrades = trades.filter((t) => t.profitLoss > 0);
  const losingTrades = trades.filter((t) => t.profitLoss <= 0);

  const totalProfit = winningTrades.reduce((sum, t) => sum + t.profitLoss, 0);
  const totalLoss = Math.abs(losingTrades.reduce((sum, t) => sum + t.profitLoss, 0));

  const totalProfitLoss = totalProfit - totalLoss;
  const totalProfitLossPercent = (totalProfitLoss / startingCapital) * 100;

  // Equity curve
  let equity = startingCapital;
  const equityCurve = trades.map((trade) => {
    equity += trade.profitLoss;
    return { time: trade.exitTime, equity };
  });

  // Max drawdown
  let peak = startingCapital;
  let maxDrawdown = 0;
  equityCurve.forEach((point) => {
    if (point.equity > peak) peak = point.equity;
    const drawdown = peak - point.equity;
    if (drawdown > maxDrawdown) maxDrawdown = drawdown;
  });

  return {
    totalTrades: trades.length,
    winningTrades: winningTrades.length,
    losingTrades: losingTrades.length,
    winRate: (winningTrades.length / trades.length) * 100,
    totalProfitLoss,
    totalProfitLossPercent,
    averageWin: winningTrades.length > 0 ? totalProfit / winningTrades.length : 0,
    averageLoss: losingTrades.length > 0 ? totalLoss / losingTrades.length : 0,
    largestWin: winningTrades.length > 0 ? Math.max(...winningTrades.map((t) => t.profitLoss)) : 0,
    largestLoss: losingTrades.length > 0 ? Math.min(...losingTrades.map((t) => t.profitLoss)) : 0,
    profitFactor: totalLoss > 0 ? totalProfit / totalLoss : 0,
    sharpeRatio: 0, // Simplified - would need daily returns for accurate calc
    maxDrawdown,
    maxDrawdownPercent: (maxDrawdown / startingCapital) * 100,
    trades,
    equityCurve,
    startingCapital,
    endingCapital: startingCapital + totalProfitLoss,
  };
}

export type { BacktestParams, BacktestStrategy, BacktestResults, BacktestTrade, UseBacktestDataReturn };
