import { useState, useEffect } from 'react';
import SimpleChart from '../components/SimpleChart';

// Binance dark theme colors
const COLORS = {
  background: '#0B0E11',
  cardBg: '#1E2329',
  border: '#2B3139',
  text: '#EAECEF',
  textSecondary: '#848E9C',
  gold: '#F0B90B',
  green: '#0ECB81',
  red: '#F6465D',
};

interface SystemService {
  name: string;
  status: 'online' | 'offline' | 'degraded';
  port?: number;
}

interface Decision {
  timestamp: string;
  type: 'scanning' | 'analysis' | 'reasoning' | 'decision' | 'execution';
  message: string;
  data?: any;
}

export const SOLACEConsciousnessTradingUI: React.FC = () => {
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [isLoading, setIsLoading] = useState(true);
  const [systemHealth, setSystemHealth] = useState<SystemService[]>([
    { name: 'ARES API', status: 'online', port: 8080 },
    { name: 'PostgreSQL', status: 'online', port: 5432 },
    { name: 'SOLACE Agent', status: 'online' },
  ]);
  const [performance, setPerformance] = useState({
    winRate: 45.2,
    totalTrades: 127,
    balance: 10500.00,
    pnl: 500.00
  });
  const [decisions, setDecisions] = useState<Decision[]>([
    { timestamp: new Date().toISOString(), type: 'scanning', message: 'Scanning market for opportunities...', data: { pairs: 5 } },
    { timestamp: new Date().toISOString(), type: 'analysis', message: 'Analyzing BTC/USDT signals', data: { confidence: 0.72 } },
  ]);

  // Mark as loaded after initial render
  useEffect(() => {
    const timer = setTimeout(() => setIsLoading(false), 500);
    return () => clearTimeout(timer);
  }, []);

  // Simulate new decisions for demo (remove this when backend is connected)
  useEffect(() => {
    const interval = setInterval(() => {
      const newDecision: Decision = {
        timestamp: new Date().toISOString(),
        type: 'scanning',
        message: `Checking ${symbol} market conditions...`,
        data: { symbol }
      };
      setDecisions(prev => [...prev.slice(-20), newDecision]);
    }, 30000); // Every 30 seconds

    return () => clearInterval(interval);
  }, [symbol]);

  useEffect(() => {
    // Fetch system health every 10 seconds
    const healthInterval = setInterval(async () => {
      try {
        const response = await fetch('http://localhost:8080/health/services');
        const data = await response.json();
        if (data.services) {
          setSystemHealth(data.services);
        }
      } catch (err) {
        console.error('Failed to fetch system health:', err);
      }
    }, 10000);

    // Fetch performance every 5 seconds
    const perfInterval = setInterval(async () => {
      try {
        const response = await fetch('http://localhost:8080/api/v1/trading/performance');
        if (!response.ok) {
          console.warn('Performance API returned:', response.status);
          return; // Keep existing performance data
        }
        const data = await response.json();
        // Validate data has required fields before updating state
        if (data && typeof data.winRate === 'number') {
          setPerformance(data);
        } else {
          console.warn('Invalid performance data received:', data);
        }
      } catch (err) {
        console.error('Failed to fetch performance:', err);
        // Keep existing state on error
      }
    }, 5000);

    return () => {
      clearInterval(healthInterval);
      clearInterval(perfInterval);
    };
  }, []);

  const getStatusColor = (status: string) => {
    switch(status) {
      case 'online': return COLORS.green;
      case 'degraded': return COLORS.gold;
      case 'offline': return COLORS.red;
      default: return COLORS.textSecondary;
    }
  };

  const getStatusEmoji = (status: string) => {
    switch(status) {
      case 'online': return '🟢';
      case 'degraded': return '🟡';
      case 'offline': return '🔴';
      default: return '⚪';
    }
  };

  const getDecisionEmoji = (type: Decision['type']) => {
    switch(type) {
      case 'scanning': return '🔍';
      case 'analysis': return '📊';
      case 'reasoning': return '💭';
      case 'decision': return '🎯';
      case 'execution': return '✅';
      default: return '•';
    }
  };

  const progress = (performance.winRate / 60) * 100;
  const progressColor = performance.winRate >= 60 ? COLORS.green :
                        performance.winRate >= 55 ? COLORS.gold :
                        performance.winRate >= 50 ? '#FF9500' : COLORS.red;

  // Show loading screen while initializing
  if (isLoading) {
    return (
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        backgroundColor: COLORS.background,
        color: COLORS.text,
        flexDirection: 'column',
        gap: '20px'
      }}>
        <div style={{
          width: '60px',
          height: '60px',
          border: '4px solid #2B3139',
          borderTopColor: COLORS.gold,
          borderRadius: '50%',
          animation: 'spin 1s linear infinite'
        }} />
        <div style={{ fontSize: '18px', fontWeight: '600' }}>
          Loading SOLACE Consciousness Interface...
        </div>
        <style>{`
          @keyframes spin {
            to { transform: rotate(360deg); }
          }
        `}</style>
      </div>
    );
  }

  return (
    <div style={{ 
      minHeight: '100vh',
      backgroundColor: COLORS.background,
      color: COLORS.text,
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
    }}>
      {/* System Health Bar */}
      <div style={{
        backgroundColor: COLORS.cardBg,
        borderBottom: `1px solid ${COLORS.border}`,
        padding: '12px 20px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <div style={{ fontWeight: 'bold', fontSize: '16px' }}>
          🧠 SOLACE Consciousness Trading Platform
        </div>

        <div style={{ 
          display: 'flex', 
          gap: '24px',
          fontSize: '13px'
        }}>
          {systemHealth.map((service, idx) => (
            <div 
              key={idx}
              style={{ display: 'flex', alignItems: 'center', gap: '6px' }}
            >
              <span>{getStatusEmoji(service.status)}</span>
              <span style={{ color: COLORS.text }}>{service.name}:</span>
              <span style={{ 
                color: getStatusColor(service.status),
                fontWeight: '600'
              }}>
                {service.status.toUpperCase()}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Main Content Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: '70% 30%',
        gap: '16px',
        padding: '20px',
        minHeight: 'calc(100vh - 60px)'
      }}>
        {/* Left Column: TradingView Chart */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {/* Symbol Selector */}
          <div style={{
            backgroundColor: COLORS.cardBg,
            padding: '12px 16px',
            borderRadius: '8px',
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            border: `1px solid ${COLORS.border}`
          }}>
            <span style={{ fontWeight: '600' }}>Trading Pair:</span>
            <select
              value={symbol}
              onChange={(e) => setSymbol(e.target.value)}
              style={{
                backgroundColor: COLORS.background,
                color: COLORS.text,
                border: `1px solid ${COLORS.border}`,
                padding: '6px 12px',
                borderRadius: '4px',
                fontSize: '14px',
                cursor: 'pointer'
              }}
            >
              <option value="BTCUSDT">BTC/USDT</option>
              <option value="ETHUSDT">ETH/USDT</option>
              <option value="SOLUSDT">SOL/USDT</option>
              <option value="ADAUSDT">ADA/USDT</option>
              <option value="DOTUSDT">DOT/USDT</option>
            </select>
            <div style={{ 
              marginLeft: 'auto',
              color: COLORS.textSecondary,
              fontSize: '13px'
            }}>
              Professional TradingView Chart with RSI, MACD, EMA
            </div>
          </div>

          {/* TradingView Chart */}
          <div style={{
            backgroundColor: COLORS.cardBg,
            padding: '8px',
            borderRadius: '8px',
            border: `1px solid ${COLORS.border}`
          }}>
            <SimpleChart 
              symbol={symbol}
              height={550}
            />
          </div>
        </div>

        {/* Right Column: SOLACE Brain Panel */}
        <div style={{ 
          display: 'flex', 
          flexDirection: 'column', 
          gap: '16px'
        }}>
          {/* Progress to 60% Win Rate */}
          <div style={{
            backgroundColor: COLORS.cardBg,
            border: `1px solid ${COLORS.border}`,
            borderRadius: '8px',
            padding: '20px'
          }}>
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              marginBottom: '12px'
            }}>
              <span style={{ fontWeight: 'bold' }}>
                🎯 Progress to Autonomous Trading
              </span>
              <span style={{ color: COLORS.textSecondary, fontSize: '13px' }}>
                {performance.totalTrades} trades
              </span>
            </div>

            {/* Progress Bar */}
            <div style={{
              width: '100%',
              height: '32px',
              backgroundColor: COLORS.background,
              borderRadius: '6px',
              overflow: 'hidden',
              position: 'relative',
              marginBottom: '16px'
            }}>
              <div style={{
                width: `${Math.min(progress, 100)}%`,
                height: '100%',
                backgroundColor: progressColor,
                transition: 'width 0.5s ease',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}>
                <span style={{ 
                  fontWeight: 'bold',
                  color: 'white',
                  fontSize: '14px',
                  position: 'absolute'
                }}>
                  {performance.winRate?.toFixed(1) ?? '0.0'}%
                </span>
              </div>
            </div>

            {/* Status */}
            <div style={{ textAlign: 'center' }}>
              {performance.winRate >= 60 ? (
                <div style={{
                  color: COLORS.green,
                  fontWeight: 'bold',
                  fontSize: '16px'
                }}>
                  🎉 TARGET ACHIEVED! READY FOR LIVE TRADING
                </div>
              ) : (
                <div>
                  <div style={{ fontSize: '16px', marginBottom: '8px' }}>
                    <span style={{ color: progressColor, fontWeight: 'bold' }}>
                      {(60 - (performance.winRate ?? 0)).toFixed(1)}%
                    </span>
                    <span style={{ color: COLORS.textSecondary }}> away from target</span>
                  </div>
                  <div style={{ fontSize: '12px', color: COLORS.textSecondary }}>
                    Target: 60% win rate for autonomous approval
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Live Decision Stream */}
          <div style={{
            backgroundColor: COLORS.cardBg,
            border: `1px solid ${COLORS.border}`,
            borderRadius: '8px',
            padding: '16px',
            height: '300px',
            overflowY: 'auto',
            fontFamily: 'monospace',
            fontSize: '13px'
          }}>
            <div style={{ 
              fontWeight: 'bold',
              marginBottom: '12px',
              position: 'sticky',
              top: 0,
              backgroundColor: COLORS.cardBg,
              paddingBottom: '8px'
            }}>
              🧠 SOLACE Live Thoughts
            </div>

            {decisions.map((decision, idx) => (
              <div 
                key={idx}
                style={{ 
                  marginBottom: '8px',
                  color: COLORS.textSecondary,
                  opacity: idx === decisions.length - 1 ? 1 : 0.7
                }}
              >
                <span style={{ color: COLORS.gold }}>
                  {new Date(decision.timestamp).toLocaleTimeString()}
                </span>
                {' '}
                <span>{getDecisionEmoji(decision.type)}</span>
                {' '}
                {decision.message}
                {decision.data && (
                  <span style={{ color: '#667eea', marginLeft: '8px' }}>
                    {JSON.stringify(decision.data)}
                  </span>
                )}
              </div>
            ))}

            {decisions.length === 0 && (
              <div style={{ color: COLORS.textSecondary, textAlign: 'center', marginTop: '40px' }}>
                Waiting for SOLACE to start thinking...
              </div>
            )}
          </div>

          {/* Approval Queue Placeholder */}
          <div style={{
            backgroundColor: COLORS.cardBg,
            border: `1px solid ${COLORS.border}`,
            borderRadius: '8px',
            padding: '32px',
            textAlign: 'center',
            color: COLORS.textSecondary
          }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>⏳</div>
            <div>No pending trade proposals</div>
            <div style={{ fontSize: '12px', marginTop: '8px' }}>
              SOLACE will notify when confident enough to propose
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SOLACEConsciousnessTradingUI;
