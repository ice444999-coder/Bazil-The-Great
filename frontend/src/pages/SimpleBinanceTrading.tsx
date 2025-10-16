import React, { useState } from 'react';
import TradingChart from '../components/TradingChart';

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

const SimpleBinanceTrading: React.FC = () => {
  const [symbol, setSymbol] = useState('BTCUSDT');
  const currentPrice = 66500.00;

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
        {/* Left Column - Chart Placeholder */}
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
            <div style={{ display: 'flex', gap: '4px' }}>
              {['1m', '5m', '15m', '30m', '1H', '4H', '1D', '1W'].map(tf => (
                <button key={tf} style={{
                  backgroundColor: tf === '1H' ? COLORS.gold : 'transparent',
                  color: tf === '1H' ? COLORS.background : COLORS.text,
                  border: 'none',
                  padding: '4px 8px',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  fontSize: '12px',
                  fontWeight: 500,
                }}>
                  {tf}
                </button>
              ))}
            </div>
          </div>

          {/* Chart Area */}
          <div style={{ 
            flex: 1, 
            backgroundColor: COLORS.background,
            padding: '8px',
          }}>
            <TradingChart symbol={symbol.replace('USDT', '/USDT')} height={500} />
          </div>

          {/* Bottom Panel - Recent Trades */}
          <div style={{
            height: '200px',
            backgroundColor: COLORS.cardBg,
            borderTop: `1px solid ${COLORS.border}`,
            padding: '12px',
          }}>
            <div style={{ fontWeight: 600, marginBottom: '8px', fontSize: '14px' }}>
              Recent Trades
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', fontSize: '12px' }}>
              {[
                { price: '66,523.45', amount: '0.125', time: '14:32:12', type: 'buy' },
                { price: '66,520.00', amount: '0.085', time: '14:32:10', type: 'sell' },
                { price: '66,525.12', amount: '0.250', time: '14:32:08', type: 'buy' },
                { price: '66,518.90', amount: '0.175', time: '14:32:05', type: 'sell' },
              ].map((trade, i) => (
                <div key={i} style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: trade.type === 'buy' ? COLORS.green : COLORS.red }}>
                    ${trade.price}
                  </span>
                  <span style={{ color: COLORS.textSecondary }}>{trade.amount}</span>
                  <span style={{ color: COLORS.textSecondary }}>{trade.time}</span>
                </div>
              ))}
            </div>
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
            padding: '12px',
            overflow: 'hidden',
          }}>
            <div style={{ fontWeight: 600, marginBottom: '8px', fontSize: '14px' }}>
              Order Book
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: COLORS.textSecondary, marginBottom: '8px' }}>
              <span>Price (USDT)</span>
              <span>Amount (BTC)</span>
              <span>Total</span>
            </div>
            {/* Asks (red - sell orders) */}
            {[
              { price: '66,528.50', amount: '0.125', total: '8,316.06' },
              { price: '66,527.00', amount: '0.085', total: '5,654.80' },
              { price: '66,526.20', amount: '0.250', total: '16,631.55' },
            ].map((order, i) => (
              <div key={`ask-${i}`} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '4px' }}>
                <span style={{ color: COLORS.red }}>{order.price}</span>
                <span>{order.amount}</span>
                <span style={{ color: COLORS.textSecondary }}>{order.total}</span>
              </div>
            ))}
            <div style={{ textAlign: 'center', margin: '12px 0', fontSize: '14px', fontWeight: 600, color: COLORS.green }}>
              ${currentPrice.toFixed(2)}
            </div>
            {/* Bids (green - buy orders) */}
            {[
              { price: '66,499.00', amount: '0.175', total: '11,637.33' },
              { price: '66,498.50', amount: '0.090', total: '5,984.87' },
              { price: '66,497.20', amount: '0.300', total: '19,949.16' },
            ].map((order, i) => (
              <div key={`bid-${i}`} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '4px' }}>
                <span style={{ color: COLORS.green }}>{order.price}</span>
                <span>{order.amount}</span>
                <span style={{ color: COLORS.textSecondary }}>{order.total}</span>
              </div>
            ))}
          </div>

          {/* Order Entry Form */}
          <div style={{
            height: '450px',
            padding: '16px',
            overflowY: 'auto',
          }}>
            <div style={{ fontWeight: 600, marginBottom: '12px', fontSize: '14px' }}>
              Spot Trading
            </div>
            
            {/* Buy/Sell Tabs */}
            <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
              <button style={{
                flex: 1,
                backgroundColor: COLORS.green,
                color: 'white',
                border: 'none',
                padding: '8px',
                borderRadius: '4px',
                cursor: 'pointer',
                fontWeight: 600,
              }}>
                Buy
              </button>
              <button style={{
                flex: 1,
                backgroundColor: 'transparent',
                color: COLORS.text,
                border: `1px solid ${COLORS.border}`,
                padding: '8px',
                borderRadius: '4px',
                cursor: 'pointer',
              }}>
                Sell
              </button>
            </div>

            {/* Order Type */}
            <div style={{ marginBottom: '12px' }}>
              <label style={{ fontSize: '12px', color: COLORS.textSecondary, display: 'block', marginBottom: '4px' }}>
                Order Type
              </label>
              <select style={{
                width: '100%',
                backgroundColor: COLORS.background,
                color: COLORS.text,
                border: `1px solid ${COLORS.border}`,
                padding: '8px',
                borderRadius: '4px',
                fontSize: '14px',
              }}>
                <option>Market</option>
                <option>Limit</option>
                <option>Stop Limit</option>
              </select>
            </div>

            {/* Amount */}
            <div style={{ marginBottom: '12px' }}>
              <label style={{ fontSize: '12px', color: COLORS.textSecondary, display: 'block', marginBottom: '4px' }}>
                Amount (BTC)
              </label>
              <input 
                type="text" 
                placeholder="0.00"
                style={{
                  width: '100%',
                  backgroundColor: COLORS.background,
                  color: COLORS.text,
                  border: `1px solid ${COLORS.border}`,
                  padding: '8px',
                  borderRadius: '4px',
                  fontSize: '14px',
                }}
              />
            </div>

            {/* Total */}
            <div style={{ marginBottom: '16px' }}>
              <label style={{ fontSize: '12px', color: COLORS.textSecondary, display: 'block', marginBottom: '4px' }}>
                Total (USDT)
              </label>
              <input 
                type="text" 
                placeholder="0.00"
                style={{
                  width: '100%',
                  backgroundColor: COLORS.background,
                  color: COLORS.text,
                  border: `1px solid ${COLORS.border}`,
                  padding: '8px',
                  borderRadius: '4px',
                  fontSize: '14px',
                }}
              />
            </div>

            {/* Submit Button */}
            <button style={{
              width: '100%',
              backgroundColor: COLORS.green,
              color: 'white',
              border: 'none',
              padding: '12px',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 600,
              fontSize: '14px',
            }}>
              Buy {symbol.replace('USDT', '')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SimpleBinanceTrading;
