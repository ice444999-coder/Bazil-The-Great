import React, { useState, useEffect } from 'react';
import TradingChart from '../components/TradingChart';
import OrderBook from '../components/OrderBook';
import RecentTrades from '../components/RecentTrades';
import AdvancedOrderForm from '../components/AdvancedOrderForm';
import TimeframeSelector from '../components/TimeframeSelector';
import IndicatorSelector from '../components/IndicatorSelector';
import DrawingToolsBar from '../components/DrawingToolsBar';

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

interface OrderData {
  symbol: string;
  orderType: 'market' | 'limit';
  direction: 'long' | 'short';
  amount: number;
  price?: number;
  stopLoss?: number;
  stopLossPercent?: number;
  takeProfit?: number;
  takeProfitPercent?: number;
}

const BinanceTrading: React.FC = () => {
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [interval, setInterval] = useState('1h');
  const [currentPrice, setCurrentPrice] = useState(66500.00);
  const [indicators, setIndicators] = useState<Record<string, boolean>>({});

  // Simulate price updates
  useEffect(() => {
    const priceInterval = setInterval(() => {
      setCurrentPrice(prev => prev + (Math.random() - 0.5) * 100);
    }, 3000);
    return () => clearInterval(priceInterval);
  }, []);

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      backgroundColor: COLORS.background,
      color: COLORS.text,
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    }}>
      {/* Top Header Bar */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        padding: '12px 16px',
        backgroundColor: COLORS.cardBg,
        borderBottom: `1px solid ${COLORS.border}`,
        gap: '24px',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span style={{ fontSize: '16px', fontWeight: 600 }}>{symbol}</span>
          <select
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            style={{
              backgroundColor: COLORS.background,
              color: COLORS.text,
              border: `1px solid ${COLORS.border}`,
              padding: '4px 8px',
              borderRadius: '4px',
              fontSize: '14px',
            }}
          >
            <option value="BTCUSDT">BTC/USDT</option>
            <option value="ETHUSDT">ETH/USDT</option>
            <option value="SOLUSDT">SOL/USDT</option>
            <option value="ADAUSDT">ADA/USDT</option>
            <option value="DOTUSDT">DOT/USDT</option>
          </select>
        </div>

        <div style={{ display: 'flex', gap: '16px', fontSize: '14px' }}>
          <div>
            <span style={{ color: COLORS.textSecondary }}>Price: </span>
            <span style={{ fontWeight: 600 }}>${currentPrice.toFixed(2)}</span>
          </div>
          <div>
            <span style={{ color: COLORS.textSecondary }}>24h Change: </span>
            <span style={{ color: COLORS.green, fontWeight: 600 }}>+2.45%</span>
          </div>
          <div>
            <span style={{ color: COLORS.textSecondary }}>24h High: </span>
            <span>$67,234.56</span>
          </div>
          <div>
            <span style={{ color: COLORS.textSecondary }}>24h Low: </span>
            <span>$65,123.45</span>
          </div>
          <div>
            <span style={{ color: COLORS.textSecondary }}>24h Volume: </span>
            <span>23,456 BTC</span>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Left Column - Chart + Controls */}
        <div style={{ 
          flex: 1, 
          display: 'flex', 
          flexDirection: 'column',
          borderRight: `1px solid ${COLORS.border}`,
        }}>
          {/* Chart Controls */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            padding: '8px 12px',
            backgroundColor: COLORS.cardBg,
            borderBottom: `1px solid ${COLORS.border}`,
            gap: '12px',
          }}>
            <TimeframeSelector
              onTimeframeChange={(timeframe: string, intervalValue: string) => {
                setInterval(intervalValue);
              }}
              defaultTimeframe={interval}
            />
            <div style={{ width: '1px', height: '20px', backgroundColor: COLORS.border }} />
            <IndicatorSelector
              onIndicatorToggle={(indicatorId: string, enabled: boolean) => {
                setIndicators(prev => ({ ...prev, [indicatorId]: enabled }));
              }}
            />
            <div style={{ width: '1px', height: '20px', backgroundColor: COLORS.border }} />
            <DrawingToolsBar onToolSelect={(tool: string) => console.log('Drawing tool:', tool)} />
          </div>

          {/* Trading Chart */}
          <div style={{ flex: 1, backgroundColor: COLORS.background }}>
            <TradingChart
              symbol={symbol}
            />
          </div>

          {/* Bottom Panel - Recent Trades */}
          <div style={{
            height: '200px',
            backgroundColor: COLORS.cardBg,
            borderTop: `1px solid ${COLORS.border}`,
          }}>
            <RecentTrades symbol={symbol} />
          </div>
        </div>

        {/* Right Column - OrderBook + Order Entry */}
        <div style={{
          width: '380px',
          display: 'flex',
          flexDirection: 'column',
          backgroundColor: COLORS.cardBg,
        }}>
          {/* Order Book */}
          <div style={{
            flex: 1,
            borderBottom: `1px solid ${COLORS.border}`,
            overflow: 'hidden',
          }}>
            <OrderBook symbol={symbol} />
          </div>

          {/* Order Entry Form */}
          <div style={{
            height: '450px',
            padding: '16px',
            overflowY: 'auto',
          }}>
            <AdvancedOrderForm
              symbol={symbol}
              currentPrice={currentPrice}
              onSubmit={(order: OrderData) => {
                console.log('Order submitted:', order);
                alert(`Order submitted: ${order.orderType} ${order.direction} ${order.amount} ${symbol}`);
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default BinanceTrading;
