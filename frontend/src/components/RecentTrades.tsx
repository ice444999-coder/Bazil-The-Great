import { useEffect, useState, useRef } from 'react';

interface RecentTradesProps {
  symbol: string;
  height?: number;
  maxTrades?: number;
}

interface Trade {
  id: string;
  price: number;
  quantity: number;
  time: number;
  isBuyerMaker: boolean;
}

export default function RecentTrades({ symbol, height = 600, maxTrades = 50 }: RecentTradesProps) {
  const [trades, setTrades] = useState<Trade[]>([]);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const tradesContainerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);

  useEffect(() => {
    const binanceSymbol = symbol.replace('/', '').toLowerCase();
    let isSubscribed = true;

    // Fetch initial recent trades
    fetchRecentTrades(binanceSymbol, maxTrades)
      .then((initialTrades) => {
        if (isSubscribed) {
          setTrades(initialTrades);
          setError(null);
        }
      })
      .catch((err) => {
        if (isSubscribed) {
          setError(err.message || 'Failed to load trades');
        }
      });

    // Connect to WebSocket for real-time trade stream
    const ws = new WebSocket(`wss://stream.binance.com:9443/ws/${binanceSymbol}@trade`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log(`RecentTrades WebSocket connected for ${symbol}`);
    };

    ws.onmessage = (event) => {
      if (!isSubscribed) return;

      try {
        const data = JSON.parse(event.data);
        
        const newTrade: Trade = {
          id: data.t.toString(),
          price: parseFloat(data.p),
          quantity: parseFloat(data.q),
          time: data.T,
          isBuyerMaker: data.m,
        };

        setTrades((prevTrades) => {
          const updated = [newTrade, ...prevTrades];
          return updated.slice(0, maxTrades);
        });

        setError(null);
      } catch (err) {
        console.error('Error parsing trade data:', err);
      }
    };

    ws.onerror = (err) => {
      console.error('RecentTrades WebSocket error:', err);
      if (isSubscribed) {
        setError('WebSocket connection error');
      }
    };

    ws.onclose = () => {
      console.log(`RecentTrades WebSocket closed for ${symbol}`);
    };

    // Cleanup
    return () => {
      isSubscribed = false;
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [symbol, maxTrades]);

  // Auto-scroll to top when new trades arrive (if enabled)
  useEffect(() => {
    if (autoScroll && tradesContainerRef.current) {
      tradesContainerRef.current.scrollTop = 0;
    }
  }, [trades, autoScroll]);

  // Handle manual scroll - disable auto-scroll if user scrolls down
  const handleScroll = () => {
    if (tradesContainerRef.current) {
      const isAtTop = tradesContainerRef.current.scrollTop < 10;
      setAutoScroll(isAtTop);
    }
  };

  if (error) {
    return (
      <div
        style={{
          height: height,
          background: '#0B0E11',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#F6465D',
          padding: 20,
          borderRadius: 8,
        }}
      >
        <div>
          <div style={{ fontSize: 18, marginBottom: 8 }}>⚠️ Error</div>
          <div>{error}</div>
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        height: height,
        background: '#0B0E11',
        borderRadius: 8,
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '16px 20px',
          borderBottom: '1px solid #2B3139',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div style={{ color: '#FFFFFF', fontSize: 16, fontWeight: 600 }}>
          Recent Trades
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{ color: '#848E9C', fontSize: 12 }}>
            {symbol}
          </div>
          {!autoScroll && (
            <button
              onClick={() => setAutoScroll(true)}
              style={{
                padding: '4px 8px',
                background: '#F0B90B20',
                border: '1px solid #F0B90B',
                borderRadius: 4,
                color: '#F0B90B',
                fontSize: 11,
                cursor: 'pointer',
                fontWeight: 600,
              }}
            >
              Resume Auto-Scroll
            </button>
          )}
        </div>
      </div>

      {/* Column Headers */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr 1fr',
          padding: '12px 20px',
          borderBottom: '1px solid #2B3139',
          fontSize: 12,
          color: '#848E9C',
          fontWeight: 600,
        }}
      >
        <div>Price (USDT)</div>
        <div style={{ textAlign: 'right' }}>Amount</div>
        <div style={{ textAlign: 'right' }}>Time</div>
      </div>

      {/* Trades List */}
      <div
        ref={tradesContainerRef}
        onScroll={handleScroll}
        style={{
          flex: 1,
          overflow: 'auto',
        }}
      >
        {trades.length === 0 ? (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              color: '#848E9C',
            }}
          >
            Waiting for trades...
          </div>
        ) : (
          trades.map((trade, index) => (
            <TradeRow key={trade.id} trade={trade} isNew={index === 0} />
          ))
        )}
      </div>

      {/* Footer Stats */}
      <div
        style={{
          padding: '12px 20px',
          borderTop: '1px solid #2B3139',
          background: '#1E2329',
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: 12,
        }}
      >
        <div>
          <span style={{ color: '#848E9C' }}>Total Trades: </span>
          <span style={{ color: '#FFFFFF', fontWeight: 600 }}>{trades.length}</span>
        </div>
        <div>
          <span style={{ color: '#848E9C' }}>Live Stream </span>
          <span
            style={{
              display: 'inline-block',
              width: 8,
              height: 8,
              borderRadius: '50%',
              background: '#0ECB81',
              marginLeft: 4,
              animation: 'pulse 2s infinite',
            }}
          />
        </div>
      </div>

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.3; }
        }
        @keyframes slideIn {
          from { 
            transform: translateX(-10px);
            opacity: 0;
          }
          to { 
            transform: translateX(0);
            opacity: 1;
          }
        }
      `}</style>
    </div>
  );
}

// Trade Row Component
function TradeRow({ trade, isNew }: { trade: Trade; isNew: boolean }) {
  const isBuy = !trade.isBuyerMaker; // Taker bought = bullish
  const color = isBuy ? '#0ECB81' : '#F6465D';
  const time = new Date(trade.time);
  const timeStr = `${time.getHours().toString().padStart(2, '0')}:${time
    .getMinutes()
    .toString()
    .padStart(2, '0')}:${time.getSeconds().toString().padStart(2, '0')}`;

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr 1fr',
        padding: '8px 20px',
        fontSize: 13,
        cursor: 'pointer',
        transition: 'background 0.2s',
        animation: isNew ? 'slideIn 0.3s ease-out' : 'none',
        background: isNew ? (isBuy ? '#0ECB8108' : '#F6465D08') : 'transparent',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = '#1E2329';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent';
      }}
    >
      <div style={{ color: color, fontWeight: 600 }}>
        {trade.price.toFixed(2)}
      </div>
      <div style={{ textAlign: 'right', color: '#FFFFFF' }}>
        {trade.quantity.toFixed(6)}
      </div>
      <div style={{ textAlign: 'right', color: '#848E9C' }}>
        {timeStr}
      </div>
    </div>
  );
}

// Fetch initial recent trades from Binance REST API
async function fetchRecentTrades(symbol: string, limit: number): Promise<Trade[]> {
  const response = await fetch(
    `https://api.binance.com/api/v3/trades?symbol=${symbol.toUpperCase()}&limit=${limit}`
  );

  if (!response.ok) {
    throw new Error(`Binance API error: ${response.status}`);
  }

  const data = await response.json();

  return data.map((trade: any) => ({
    id: trade.id.toString(),
    price: parseFloat(trade.price),
    quantity: parseFloat(trade.qty),
    time: trade.time,
    isBuyerMaker: trade.isBuyerMaker,
  }));
}
