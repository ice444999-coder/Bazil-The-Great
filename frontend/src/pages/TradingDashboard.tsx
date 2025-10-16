import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js';
import Sidebar from '../components/Sidebar';
import StatsCard from '../components/StatsCard';
import { SimpleTradeForm } from '../components/SimpleTradeForm';
import OpenPositionsTable from '../components/OpenPositionsTable';
import { TradeHistoryTable } from '../components/TradeHistoryTable';
import { PlaybookRules } from '../components/PlaybookRules';

// Register Chart.js components
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler);

interface PerformanceData {
  total_profit_loss: number;
  win_rate: number;
  total_trades: number;
  open_positions: number;
  trades: any[];
}

interface Trade {
  id: number;
  trading_pair: string;
  direction: 'BUY' | 'SELL';
  size: number;
  entry_price: number;
  exit_price?: number;
  profit_loss?: number;
  fees: number;
  status: 'OPEN' | 'CLOSED';
  opened_at: string;
  closed_at?: string;
}

interface PlaybookRule {
  rule_id: string;
  confidence: number;
  helpful_count: number;
  harmful_count: number;
  is_active: boolean;
}

export default function TradingDashboard() {
  const navigate = useNavigate();
  const [performance, setPerformance] = useState<PerformanceData | null>(null);
  const [openPositions, setOpenPositions] = useState<Trade[]>([]);
  const [tradeHistory, setTradeHistory] = useState<Trade[]>([]);
  const [rules, setRules] = useState<PlaybookRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [tradeLoading, setTradeLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'manual' | 'solace'>('manual');

  // Auth check
  useEffect(() => {
    const token = localStorage.getItem('ares_token');
    if (!token) {
      navigate('/login');
    }
  }, [navigate]);

  // Fetch all trading data
  const fetchTradingData = async () => {
    const token = localStorage.getItem('ares_token');
    if (!token) return;

    try {
      // 1. Performance stats
      const perfResponse = await fetch('http://localhost:8080/api/v1/trading/performance', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (perfResponse.ok) {
        const perfData = await perfResponse.json();
        setPerformance(perfData);
      }

      // 2. Open positions
      const positionsResponse = await fetch('http://localhost:8080/api/v1/trading/open', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (positionsResponse.ok) {
        const positionsData = await positionsResponse.json();
        setOpenPositions(Array.isArray(positionsData) ? positionsData : positionsData.positions || []);
      }

      // 3. Trade history
      const historyResponse = await fetch('http://localhost:8080/api/v1/trading/history', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (historyResponse.ok) {
        const historyData = await historyResponse.json();
        setTradeHistory(Array.isArray(historyData) ? historyData : historyData.trades || []);
      }

      // 4. Playbook rules (no auth required)
      const rulesResponse = await fetch('http://localhost:8080/api/v1/trading/playbook/');
      if (rulesResponse.ok) {
        const rulesData = await rulesResponse.json();
        setRules(rulesData.rules || []);
      }

      setLoading(false);
    } catch (err: any) {
      console.error('Error fetching trading data:', err);
      setError(err.message);
      setLoading(false);
    }
  };

  // Initial fetch and auto-refresh every 10 seconds
  useEffect(() => {
    fetchTradingData();
    const interval = setInterval(fetchTradingData, 10000);
    return () => clearInterval(interval);
  }, []);

  // Handle navigation
  const handleNavigate = (page: string) => {
    navigate(`/${page}`);
  };

  // Handle logout
  const handleLogout = () => {
    localStorage.removeItem('ares_token');
    localStorage.removeItem('ares_user');
    navigate('/login');
  };

  // Handle trade execution
  const handleTradeSubmit = async (data: { symbol: string; action: string; amount: number }) => {
    const token = localStorage.getItem('ares_token');
    setTradeLoading(true);

    try {
      const response = await fetch('http://localhost:8080/api/v1/trading/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Trade execution failed');
      }

      // Refresh data after successful trade
      setTimeout(fetchTradingData, 1000);
    } finally {
      setTradeLoading(false);
    }
  };

  // Handle close position
  const handleClosePosition = async (id: number) => {
    if (!confirm('Are you sure you want to close this position?')) return;

    const token = localStorage.getItem('ares_token');
    try {
      const response = await fetch('http://localhost:8080/api/v1/trading/close', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ trade_id: id }),
      });

      if (response.ok) {
        alert('Position closed successfully!');
        fetchTradingData();
      } else {
        alert('Failed to close position');
      }
    } catch (error) {
      alert('Network error');
    }
  };

  // Prepare chart data
  const getChartData = () => {
    if (!performance || !performance.trades || performance.trades.length === 0) {
      return {
        labels: [],
        datasets: [
          {
            label: 'Cumulative P&L',
            data: [],
            borderColor: '#667eea',
            backgroundColor: 'rgba(102, 126, 234, 0.1)',
            fill: true,
            tension: 0.4,
          },
        ],
      };
    }

    let cumulativePnL = 0;
    const labels: string[] = [];
    const values: number[] = [];

    performance.trades.forEach((trade, index) => {
      cumulativePnL += trade.profit_loss || 0;
      labels.push(`Trade ${index + 1}`);
      values.push(cumulativePnL);
    });

    return {
      labels,
      datasets: [
        {
          label: 'Cumulative P&L',
          data: values,
          borderColor: '#667eea',
          backgroundColor: 'rgba(102, 126, 234, 0.1)',
          fill: true,
          tension: 0.4,
        },
      ],
    };
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true },
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          callback: (value: any) => '$' + value.toFixed(2),
        },
      },
    },
  };

  if (loading) {
    return (
      <div style={styles.container}>
        <Sidebar activePage="trading" onNavigate={handleNavigate} onLogout={handleLogout} />
        <div style={styles.mainContent}>
          <div style={styles.loading}>Loading...</div>
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <Sidebar activePage="trading" onNavigate={handleNavigate} onLogout={handleLogout} />

      <div style={styles.mainContent}>
        {/* Header */}
        <div style={styles.header}>
          <h2 style={styles.headerTitle}>Trading Dashboard</h2>
          <p style={styles.headerSubtitle}>Execute trades manually or monitor SOLACE's autonomous trading activity</p>
        </div>

        {/* Error message */}
        {error && <div style={styles.errorMessage}>{error}</div>}

        {/* Tab Navigation */}
        <div style={styles.tabContainer}>
          <button
            style={activeTab === 'manual' ? styles.tabActive : styles.tab}
            onClick={() => setActiveTab('manual')}
          >
            📊 Manual Trading
          </button>
          <button
            style={activeTab === 'solace' ? styles.tabActive : styles.tab}
            onClick={() => setActiveTab('solace')}
          >
            🧠 SOLACE Consciousness
          </button>
        </div>

        {/* Manual Trading Tab */}
        {activeTab === 'manual' && (
          <>
            {/* 4 Stat Cards */}
            <div style={styles.statsGrid}>
          <StatsCard
            label="Total P&L"
            value={`$${(performance?.total_profit_loss || 0).toFixed(2)}`}
            positive={performance && performance.total_profit_loss >= 0}
          />
          <StatsCard label="Win Rate" value={`${((performance?.win_rate || 0) * 100).toFixed(1)}%`} />
          <StatsCard label="Total Trades" value={performance?.total_trades || 0} />
          <StatsCard label="Open Positions" value={performance?.open_positions || 0} />
        </div>

        {/* Trade Form + Playbook Rules */}
        <div style={styles.grid}>
          <div style={styles.section}>
            <div style={styles.sectionHeader}>
              <h3>Execute Trade</h3>
            </div>
            <SimpleTradeForm onSubmit={handleTradeSubmit} loading={tradeLoading} />
          </div>

          <div style={styles.section}>
            <div style={styles.sectionHeader}>
              <h3>ACE Playbook Rules</h3>
            </div>
            <PlaybookRules rules={rules} />
          </div>
        </div>

        {/* P&L Chart */}
        <div style={{ ...styles.section, ...styles.gridFull }}>
          <div style={styles.sectionHeader}>
            <h3>Cumulative P&L</h3>
          </div>
          <div style={styles.chartContainer}>
            <Line data={getChartData()} options={chartOptions} />
          </div>
        </div>

        {/* Open Positions Table */}
        <div style={styles.section}>
          <div style={styles.sectionHeader}>
            <h3>Open Positions</h3>
          </div>
          <OpenPositionsTable positions={openPositions} onClose={handleClosePosition} />
        </div>

        {/* Trade History Table */}
        <div style={styles.section}>
          <div style={styles.sectionHeader}>
            <h3>Trade History</h3>
          </div>
          <TradeHistoryTable trades={tradeHistory} />
        </div>
          </>
        )}

        {/* SOLACE Consciousness Tab */}
        {activeTab === 'solace' && (
          <div style={styles.solaceContainer}>
            <h3>🧠 SOLACE Consciousness Trading</h3>
            <p>SOLACE AI integration coming soon...</p>
          </div>
        )}
      </div>
    </div>
  );
}

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    display: 'flex',
    height: '100vh',
    background: '#f5f5f5',
  },
  mainContent: {
    flex: 1,
    marginLeft: '250px',
    overflowY: 'auto',
    padding: '30px',
  },
  header: {
    marginBottom: '30px',
  },
  headerTitle: {
    fontSize: '28px',
    marginBottom: '5px',
    color: '#333',
  },
  headerSubtitle: {
    color: '#666',
    fontSize: '14px',
  },
  statsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '15px',
    marginBottom: '20px',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gap: '20px',
    marginBottom: '20px',
  },
  gridFull: {
    gridColumn: '1 / -1',
  },
  section: {
    background: 'white',
    padding: '25px',
    borderRadius: '10px',
    boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
    marginBottom: '20px',
  },
  sectionHeader: {
    marginBottom: '20px',
    paddingBottom: '15px',
    borderBottom: '2px solid #f0f0f0',
  },
  chartContainer: {
    position: 'relative',
    height: '300px',
  },
  loading: {
    textAlign: 'center',
    padding: '40px',
    color: '#666',
    fontSize: '18px',
  },
  errorMessage: {
    background: '#fee',
    color: '#c33',
    padding: '10px 15px',
    borderRadius: '5px',
    marginBottom: '15px',
  },
  tabContainer: {
    display: 'flex',
    gap: '10px',
    marginBottom: '25px',
    borderBottom: '2px solid #e0e0e0',
  },
  tab: {
    padding: '12px 24px',
    border: 'none',
    background: 'transparent',
    color: '#666',
    fontSize: '15px',
    fontWeight: '500',
    cursor: 'pointer',
    borderBottom: '3px solid transparent',
    transition: 'all 0.3s ease',
  },
  tabActive: {
    padding: '12px 24px',
    border: 'none',
    background: 'transparent',
    color: '#2563eb',
    fontSize: '15px',
    fontWeight: '600',
    cursor: 'pointer',
    borderBottom: '3px solid #2563eb',
  },
  solaceContainer: {
    background: 'white',
    padding: '40px',
    borderRadius: '10px',
    boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
    textAlign: 'center',
  },
};
