/**
 * MainTradingDashboard - Phase 1 Integration
 * 
 * Wires together existing components with backend API:
 * - OpenPositionsTable (displays open trades)
 * - StatsCard (shows P&L metrics)
 * - Auto-refresh every 5 seconds
 * - Live price updates for unrealized P&L
 */

import { useState, useEffect } from 'react';
import OpenPositionsTable from './OpenPositionsTable';
import StatsCard from './StatsCard';
import { tradingApi } from '../utils/api';
import type { SandboxTrade, TradePerformance, CryptoPrice } from '../types/trading';
import { formatCurrency, formatPercentage, calculateUnrealizedPnL } from '../types/trading';

export default function MainTradingDashboard() {
  // State
  const [positions, setPositions] = useState<SandboxTrade[]>([]);
  const [performance, setPerformance] = useState<TradePerformance | null>(null);
  const [prices, setPrices] = useState<Record<string, CryptoPrice>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  /**
   * Fetch all dashboard data
   */
  const fetchDashboardData = async () => {
    try {
      setError(null);
      
      // Fetch in parallel for better performance
      const [positionsData, performanceData, pricesData] = await Promise.all([
        tradingApi.getOpenPositions(),
        tradingApi.getPerformance(),
        tradingApi.getPrices(),
      ]);

      setPositions(positionsData);
      setPerformance(performanceData);
      setPrices(pricesData);
      setLastUpdate(new Date());
      
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Close a specific position
   */
  const handleClosePosition = async (id: number) => {
    try {
      setLoading(true);
      const result = await tradingApi.closeTrade(id);
      
      if (result.success) {
        // Refresh data immediately after closing
        await fetchDashboardData();
      }
    } catch (err) {
      console.error(`Failed to close trade #${id}:`, err);
      setError(err instanceof Error ? err.message : 'Failed to close trade');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Emergency close all positions
   */
  const handleCloseAll = async () => {
    if (!confirm('⚠️ Close ALL open positions? This cannot be undone!')) {
      return;
    }

    try {
      setLoading(true);
      const result = await tradingApi.closeAllTrades();
      
      if (result.success) {
        alert(`✅ Closed ${result.closed_count} positions`);
        await fetchDashboardData();
      }
    } catch (err) {
      console.error('Failed to close all trades:', err);
      setError(err instanceof Error ? err.message : 'Failed to close all trades');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Calculate total unrealized P&L for open positions
   */
  const calculateTotalUnrealizedPnL = (): number => {
    return positions.reduce((total, position) => {
      const symbol = position.trading_pair.replace('/', '').toUpperCase(); // BTC/USDT → BTCUSDT
      const currentPrice = prices[symbol]?.price;
      
      if (!currentPrice) return total;
      
      const unrealizedPnL = calculateUnrealizedPnL(
        position.entry_price,
        currentPrice,
        position.size,
        position.direction
      );
      
      return total + unrealizedPnL;
    }, 0);
  };

  /**
   * Initial load and polling setup
   */
  useEffect(() => {
    // Initial fetch
    fetchDashboardData();

    // Poll every 5 seconds for updates
    const interval = setInterval(fetchDashboardData, 5000);

    // Cleanup on unmount
    return () => clearInterval(interval);
  }, []);

  // Loading state
  if (loading && !performance) {
    return (
      <div style={styles.container}>
        <div style={styles.loadingState}>
          <div style={styles.spinner}></div>
          <p>Loading dashboard...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error && !performance) {
    return (
      <div style={styles.container}>
        <div style={styles.errorState}>
          <h3>⚠️ Error Loading Dashboard</h3>
          <p>{error}</p>
          <button style={styles.retryButton} onClick={() => fetchDashboardData()}>
            Retry
          </button>
        </div>
      </div>
    );
  }

  const totalUnrealizedPnL = calculateTotalUnrealizedPnL();
  const totalPnL = (performance?.total_pnl || 0) + totalUnrealizedPnL;

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <h1 style={styles.title}>SOLACE Trading Dashboard</h1>
        <div style={styles.headerActions}>
          <span style={styles.lastUpdate}>
            Last update: {lastUpdate.toLocaleTimeString()}
          </span>
          <button
            style={styles.refreshButton}
            onClick={() => fetchDashboardData()}
            disabled={loading}
          >
            🔄 Refresh
          </button>
          {positions.length > 0 && (
            <button
              style={styles.closeAllButton}
              onClick={handleCloseAll}
              disabled={loading}
            >
              🚨 Close All
            </button>
          )}
        </div>
      </div>

      {/* Error banner (if any, but still showing data) */}
      {error && (
        <div style={styles.errorBanner}>
          ⚠️ {error}
        </div>
      )}

      {/* Stats Grid */}
      <div style={styles.statsGrid}>
        <StatsCard
          label="Total P&L"
          value={formatCurrency(totalPnL)}
          positive={totalPnL >= 0}
        />
        <StatsCard
          label="Realized P&L"
          value={formatCurrency(performance?.total_pnl)}
          positive={(performance?.total_pnl || 0) >= 0}
        />
        <StatsCard
          label="Unrealized P&L"
          value={formatCurrency(totalUnrealizedPnL)}
          positive={totalUnrealizedPnL >= 0}
        />
        <StatsCard
          label="Win Rate"
          value={formatPercentage(performance?.win_rate)}
        />
        <StatsCard
          label="Open Positions"
          value={positions.length}
        />
        <StatsCard
          label="Total Trades"
          value={performance?.total_trades || 0}
        />
        <StatsCard
          label="Daily P&L"
          value={formatCurrency(performance?.daily_pnl)}
          positive={(performance?.daily_pnl || 0) >= 0}
        />
        <StatsCard
          label="Sharpe Ratio"
          value={performance?.sharpe_ratio?.toFixed(2) || 'N/A'}
        />
      </div>

      {/* Open Positions Table */}
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>
          Open Positions ({positions.length})
        </h2>
        
        {positions.length === 0 ? (
          <div style={styles.emptyState}>
            <p>No open positions</p>
            <p style={styles.emptyStateHint}>
              Execute a trade to see it here
            </p>
          </div>
        ) : (
          <OpenPositionsTable
            positions={positions.map(pos => ({
              id: pos.id,
              trading_pair: pos.trading_pair,
              direction: pos.direction === 'LONG' ? 'BUY' : 'SELL',
              size: pos.size,
              entry_price: pos.entry_price,
              fees: 0, // TODO: Add fees to backend
              opened_at: pos.opened_at,
            }))}
            onClose={handleClosePosition}
            loading={loading}
          />
        )}
      </div>

      {/* Additional Info Section */}
      <div style={styles.infoSection}>
        <div style={styles.infoCard}>
          <h3 style={styles.infoTitle}>Performance Metrics</h3>
          <div style={styles.infoGrid}>
            <div>
              <span style={styles.infoLabel}>Best Trade:</span>
              <span style={styles.infoValue}>
                {formatCurrency(performance?.best_trade)}
              </span>
            </div>
            <div>
              <span style={styles.infoLabel}>Worst Trade:</span>
              <span style={styles.infoValue}>
                {formatCurrency(performance?.worst_trade)}
              </span>
            </div>
            <div>
              <span style={styles.infoLabel}>Avg Profit:</span>
              <span style={styles.infoValue}>
                {formatCurrency(performance?.avg_profit)}
              </span>
            </div>
            <div>
              <span style={styles.infoLabel}>Avg Loss:</span>
              <span style={styles.infoValue}>
                {formatCurrency(performance?.avg_loss)}
              </span>
            </div>
            <div>
              <span style={styles.infoLabel}>Max Drawdown:</span>
              <span style={styles.infoValue}>
                {formatCurrency(performance?.max_drawdown)}
              </span>
            </div>
            <div>
              <span style={styles.infoLabel}>Avg Hold Time:</span>
              <span style={styles.infoValue}>
                {performance?.avg_hold_time || 'N/A'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    padding: '20px',
    maxWidth: '1400px',
    margin: '0 auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px',
  },
  title: {
    fontSize: '32px',
    fontWeight: 700,
    color: '#333',
    margin: 0,
  },
  headerActions: {
    display: 'flex',
    gap: '12px',
    alignItems: 'center',
  },
  lastUpdate: {
    fontSize: '14px',
    color: '#666',
  },
  refreshButton: {
    padding: '8px 16px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    fontSize: '14px',
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'background 0.3s',
  },
  closeAllButton: {
    padding: '8px 16px',
    background: '#ef4444',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    fontSize: '14px',
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'background 0.3s',
  },
  errorBanner: {
    background: '#fee',
    border: '1px solid #fcc',
    borderRadius: '6px',
    padding: '12px',
    marginBottom: '20px',
    color: '#c33',
  },
  statsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '16px',
    marginBottom: '30px',
  },
  section: {
    marginBottom: '30px',
  },
  sectionTitle: {
    fontSize: '20px',
    fontWeight: 600,
    color: '#333',
    marginBottom: '16px',
  },
  emptyState: {
    textAlign: 'center',
    padding: '60px 20px',
    background: '#f9fafb',
    borderRadius: '8px',
  },
  emptyStateHint: {
    color: '#999',
    fontSize: '14px',
    marginTop: '8px',
  },
  infoSection: {
    marginTop: '30px',
  },
  infoCard: {
    background: '#f9fafb',
    padding: '20px',
    borderRadius: '8px',
  },
  infoTitle: {
    fontSize: '18px',
    fontWeight: 600,
    color: '#333',
    marginBottom: '16px',
  },
  infoGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '16px',
  },
  infoLabel: {
    fontSize: '14px',
    color: '#666',
    marginRight: '8px',
  },
  infoValue: {
    fontSize: '16px',
    fontWeight: 600,
    color: '#333',
  },
  loadingState: {
    textAlign: 'center',
    padding: '100px 20px',
  },
  spinner: {
    width: '50px',
    height: '50px',
    border: '4px solid #f3f4f6',
    borderTop: '4px solid #667eea',
    borderRadius: '50%',
    animation: 'spin 1s linear infinite',
    margin: '0 auto 20px',
  },
  errorState: {
    textAlign: 'center',
    padding: '100px 20px',
  },
  retryButton: {
    marginTop: '20px',
    padding: '10px 20px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    fontSize: '16px',
    fontWeight: 600,
    cursor: 'pointer',
  },
};
