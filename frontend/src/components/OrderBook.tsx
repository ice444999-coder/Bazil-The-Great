import { useEffect, useState, useRef } from 'react';

interface OrderBookProps {
  symbol: string;
  height?: number;
  depth?: number;
}

interface OrderBookLevel {
  price: number;
  quantity: number;
  total: number;
}

interface OrderBookData {
  bids: OrderBookLevel[];
  asks: OrderBookLevel[];
  maxTotal: number;
}

export default function OrderBook({ symbol, height = 600, depth = 15 }: OrderBookProps) {
  const [orderBook, setOrderBook] = useState<OrderBookData>({
    bids: [],
    asks: [],
    maxTotal: 0,
  });
  const [lastUpdate, setLastUpdate] = useState<number>(Date.now());
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const binanceSymbol = symbol.replace('/', '').toLowerCase();
    let isSubscribed = true;

    // Fetch initial order book snapshot
    fetchOrderBookSnapshot(binanceSymbol, depth)
      .then((snapshot) => {
        if (isSubscribed) {
          setOrderBook(snapshot);
          setError(null);
        }
      })
      .catch((err) => {
        if (isSubscribed) {
          setError(err.message || 'Failed to load order book');
        }
      });

    // Connect to WebSocket for real-time updates
    const ws = new WebSocket(`wss://stream.binance.com:9443/ws/${binanceSymbol}@depth@100ms`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log(`OrderBook WebSocket connected for ${symbol}`);
    };

    ws.onmessage = (event) => {
      if (!isSubscribed) return;

      try {
        const data = JSON.parse(event.data);
        
        // Update order book with incremental data
        setOrderBook((prev) => {
          const newBids = updateLevels(prev.bids, data.b || []);
          const newAsks = updateLevels(prev.asks, data.a || []);
          
          return processOrderBook(newBids, newAsks, depth);
        });

        setLastUpdate(Date.now());
        setError(null);
      } catch (err) {
        console.error('Error parsing WebSocket data:', err);
      }
    };

    ws.onerror = (err) => {
      console.error('OrderBook WebSocket error:', err);
      if (isSubscribed) {
        setError('WebSocket connection error');
      }
    };

    ws.onclose = () => {
      console.log(`OrderBook WebSocket closed for ${symbol}`);
    };

    // Cleanup
    return () => {
      isSubscribed = false;
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [symbol, depth]);

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
          Order Book
        </div>
        <div style={{ color: '#848E9C', fontSize: 12 }}>
          {symbol}
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
        <div style={{ textAlign: 'right' }}>Total</div>
      </div>

      {/* Order Book Content */}
      <div style={{ flex: 1, overflow: 'auto' }}>
        {/* Asks (Sell Orders) - Red, reverse order */}
        <div>
          {orderBook.asks
            .slice()
            .reverse()
            .map((ask, index) => (
              <OrderBookRow
                key={`ask-${index}`}
                price={ask.price}
                quantity={ask.quantity}
                total={ask.total}
                percentage={(ask.total / orderBook.maxTotal) * 100}
                type="ask"
              />
            ))}
        </div>

        {/* Spread Display */}
        {orderBook.bids.length > 0 && orderBook.asks.length > 0 && (
          <div
            style={{
              padding: '12px 20px',
              background: '#1E2329',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              borderTop: '1px solid #2B3139',
              borderBottom: '1px solid #2B3139',
            }}
          >
            <div>
              <div style={{ color: '#848E9C', fontSize: 11 }}>Spread</div>
              <div style={{ color: '#F0B90B', fontSize: 14, fontWeight: 600 }}>
                {(orderBook.asks[0].price - orderBook.bids[0].price).toFixed(2)}
              </div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ color: '#848E9C', fontSize: 11 }}>Mid Price</div>
              <div style={{ color: '#FFFFFF', fontSize: 14, fontWeight: 600 }}>
                {((orderBook.asks[0].price + orderBook.bids[0].price) / 2).toFixed(2)}
              </div>
            </div>
          </div>
        )}

        {/* Bids (Buy Orders) - Green */}
        <div>
          {orderBook.bids.map((bid, index) => (
            <OrderBookRow
              key={`bid-${index}`}
              price={bid.price}
              quantity={bid.quantity}
              total={bid.total}
              percentage={(bid.total / orderBook.maxTotal) * 100}
              type="bid"
            />
          ))}
        </div>
      </div>
    </div>
  );
}

// Order Book Row Component
function OrderBookRow({
  price,
  quantity,
  total,
  percentage,
  type,
}: {
  price: number;
  quantity: number;
  total: number;
  percentage: number;
  type: 'bid' | 'ask';
}) {
  const bgColor = type === 'bid' ? '#0ECB8110' : '#F6465D10';
  const textColor = type === 'bid' ? '#0ECB81' : '#F6465D';

  return (
    <div
      style={{
        position: 'relative',
        display: 'grid',
        gridTemplateColumns: '1fr 1fr 1fr',
        padding: '6px 20px',
        fontSize: 13,
        cursor: 'pointer',
        transition: 'background 0.2s',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = '#1E2329';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent';
      }}
    >
      {/* Depth Bar */}
      <div
        style={{
          position: 'absolute',
          right: 0,
          top: 0,
          bottom: 0,
          width: `${percentage}%`,
          background: bgColor,
          zIndex: 0,
        }}
      />

      {/* Content */}
      <div style={{ color: textColor, fontWeight: 600, position: 'relative', zIndex: 1 }}>
        {price.toFixed(2)}
      </div>
      <div style={{ textAlign: 'right', color: '#FFFFFF', position: 'relative', zIndex: 1 }}>
        {quantity.toFixed(4)}
      </div>
      <div style={{ textAlign: 'right', color: '#848E9C', position: 'relative', zIndex: 1 }}>
        {total.toFixed(4)}
      </div>
    </div>
  );
}

// Fetch initial order book snapshot from Binance REST API
async function fetchOrderBookSnapshot(symbol: string, limit: number): Promise<OrderBookData> {
  const response = await fetch(
    `https://api.binance.com/api/v3/depth?symbol=${symbol.toUpperCase()}&limit=${limit}`
  );

  if (!response.ok) {
    throw new Error(`Binance API error: ${response.status}`);
  }

  const data = await response.json();

  const bids: OrderBookLevel[] = data.bids.map((bid: string[]) => ({
    price: parseFloat(bid[0]),
    quantity: parseFloat(bid[1]),
    total: 0,
  }));

  const asks: OrderBookLevel[] = data.asks.map((ask: string[]) => ({
    price: parseFloat(ask[0]),
    quantity: parseFloat(ask[1]),
    total: 0,
  }));

  return processOrderBook(bids, asks, limit);
}

// Update order book levels with WebSocket incremental data
function updateLevels(
  currentLevels: OrderBookLevel[],
  updates: string[][]
): OrderBookLevel[] {
  const levelMap = new Map<number, number>();

  // Add current levels to map
  currentLevels.forEach((level) => {
    levelMap.set(level.price, level.quantity);
  });

  // Apply updates
  updates.forEach((update) => {
    const price = parseFloat(update[0]);
    const quantity = parseFloat(update[1]);

    if (quantity === 0) {
      levelMap.delete(price);
    } else {
      levelMap.set(price, quantity);
    }
  });

  // Convert back to array
  return Array.from(levelMap.entries()).map(([price, quantity]) => ({
    price,
    quantity,
    total: 0,
  }));
}

// Process order book (calculate totals, sort, limit depth)
function processOrderBook(
  bids: OrderBookLevel[],
  asks: OrderBookLevel[],
  limit: number
): OrderBookData {
  // Sort and limit
  const sortedBids = bids.sort((a, b) => b.price - a.price).slice(0, limit);
  const sortedAsks = asks.sort((a, b) => a.price - b.price).slice(0, limit);

  // Calculate cumulative totals
  let bidTotal = 0;
  sortedBids.forEach((bid) => {
    bidTotal += bid.quantity;
    bid.total = bidTotal;
  });

  let askTotal = 0;
  sortedAsks.forEach((ask) => {
    askTotal += ask.quantity;
    ask.total = askTotal;
  });

  const maxTotal = Math.max(bidTotal, askTotal);

  return {
    bids: sortedBids,
    asks: sortedAsks,
    maxTotal,
  };
}
