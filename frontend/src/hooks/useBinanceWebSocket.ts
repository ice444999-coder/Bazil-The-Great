import { useEffect, useRef, useState } from 'react';

interface UseBinanceWebSocketProps {
  symbol: string;
  streamType: 'trade' | 'ticker' | 'kline' | 'depth';
  interval?: string; // For kline only (e.g., '1m', '1h')
  enabled?: boolean;
}

interface TradeData {
  eventType: 'trade';
  eventTime: number;
  symbol: string;
  tradeId: number;
  price: number;
  quantity: number;
  buyerOrderId: number;
  sellerOrderId: number;
  tradeTime: number;
  isBuyerMaker: boolean;
}

interface TickerData {
  eventType: 'ticker';
  eventTime: number;
  symbol: string;
  priceChange: number;
  priceChangePercent: number;
  weightedAvgPrice: number;
  prevClosePrice: number;
  lastPrice: number;
  lastQuantity: number;
  bestBidPrice: number;
  bestBidQuantity: number;
  bestAskPrice: number;
  bestAskQuantity: number;
  openPrice: number;
  highPrice: number;
  lowPrice: number;
  totalTradedBaseVolume: number;
  totalTradedQuoteVolume: number;
  openTime: number;
  closeTime: number;
  firstTradeId: number;
  lastTradeId: number;
  totalTrades: number;
}

interface KlineData {
  eventType: 'kline';
  eventTime: number;
  symbol: string;
  kline: {
    startTime: number;
    closeTime: number;
    symbol: string;
    interval: string;
    firstTradeId: number;
    lastTradeId: number;
    open: number;
    close: number;
    high: number;
    low: number;
    baseAssetVolume: number;
    numberOfTrades: number;
    isClosed: boolean;
    quoteAssetVolume: number;
    takerBuyBaseAssetVolume: number;
    takerBuyQuoteAssetVolume: number;
  };
}

interface DepthData {
  eventType: 'depth';
  eventTime: number;
  symbol: string;
  firstUpdateId: number;
  finalUpdateId: number;
  bids: Array<[string, string]>; // [price, quantity]
  asks: Array<[string, string]>; // [price, quantity]
}

type WebSocketData = TradeData | TickerData | KlineData | DepthData;

interface WebSocketState {
  data: WebSocketData | null;
  isConnected: boolean;
  error: string | null;
  reconnectAttempts: number;
}

export default function useBinanceWebSocket({
  symbol,
  streamType,
  interval,
  enabled = true,
}: UseBinanceWebSocketProps): WebSocketState {
  const [state, setState] = useState<WebSocketState>({
    data: null,
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);

  useEffect(() => {
    if (!enabled) {
      // Cleanup if disabled
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      return;
    }

    const binanceSymbol = symbol.replace('/', '').toLowerCase();
    let streamUrl = '';

    // Build WebSocket URL based on stream type
    switch (streamType) {
      case 'trade':
        streamUrl = `wss://stream.binance.com:9443/ws/${binanceSymbol}@trade`;
        break;
      case 'ticker':
        streamUrl = `wss://stream.binance.com:9443/ws/${binanceSymbol}@ticker`;
        break;
      case 'kline':
        if (!interval) {
          console.error('Interval required for kline stream');
          return;
        }
        streamUrl = `wss://stream.binance.com:9443/ws/${binanceSymbol}@kline_${interval}`;
        break;
      case 'depth':
        streamUrl = `wss://stream.binance.com:9443/ws/${binanceSymbol}@depth@100ms`;
        break;
      default:
        console.error('Invalid stream type');
        return;
    }

    const connect = () => {
      try {
        const ws = new WebSocket(streamUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log(`[WebSocket] Connected to ${streamType} stream for ${symbol}`);
          setState((prev) => ({
            ...prev,
            isConnected: true,
            error: null,
            reconnectAttempts: 0,
          }));
          reconnectAttemptsRef.current = 0;
        };

        ws.onmessage = (event) => {
          try {
            const rawData = JSON.parse(event.data);
            const parsedData = parseWebSocketData(rawData, streamType);
            
            setState((prev) => ({
              ...prev,
              data: parsedData,
              error: null,
            }));
          } catch (err) {
            console.error('[WebSocket] Parse error:', err);
            setState((prev) => ({
              ...prev,
              error: 'Failed to parse message',
            }));
          }
        };

        ws.onerror = (error) => {
          console.error('[WebSocket] Error:', error);
          setState((prev) => ({
            ...prev,
            error: 'Connection error',
            isConnected: false,
          }));
        };

        ws.onclose = (event) => {
          console.log(`[WebSocket] Closed: ${event.code} ${event.reason}`);
          setState((prev) => ({
            ...prev,
            isConnected: false,
          }));

          // Auto-reconnect with exponential backoff
          if (enabled && reconnectAttemptsRef.current < 10) {
            const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
            console.log(`[WebSocket] Reconnecting in ${delay}ms...`);
            
            reconnectTimeoutRef.current = setTimeout(() => {
              reconnectAttemptsRef.current += 1;
              setState((prev) => ({
                ...prev,
                reconnectAttempts: reconnectAttemptsRef.current,
              }));
              connect();
            }, delay);
          }
        };
      } catch (err) {
        console.error('[WebSocket] Connection failed:', err);
        setState((prev) => ({
          ...prev,
          error: 'Failed to connect',
          isConnected: false,
        }));
      }
    };

    connect();

    // Cleanup
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [symbol, streamType, interval, enabled]);

  return state;
}

// Parse raw WebSocket data based on stream type
function parseWebSocketData(data: any, streamType: string): WebSocketData {
  switch (streamType) {
    case 'trade':
      return {
        eventType: 'trade',
        eventTime: data.E,
        symbol: data.s,
        tradeId: data.t,
        price: parseFloat(data.p),
        quantity: parseFloat(data.q),
        buyerOrderId: data.b,
        sellerOrderId: data.a,
        tradeTime: data.T,
        isBuyerMaker: data.m,
      } as TradeData;

    case 'ticker':
      return {
        eventType: 'ticker',
        eventTime: data.E,
        symbol: data.s,
        priceChange: parseFloat(data.p),
        priceChangePercent: parseFloat(data.P),
        weightedAvgPrice: parseFloat(data.w),
        prevClosePrice: parseFloat(data.x),
        lastPrice: parseFloat(data.c),
        lastQuantity: parseFloat(data.Q),
        bestBidPrice: parseFloat(data.b),
        bestBidQuantity: parseFloat(data.B),
        bestAskPrice: parseFloat(data.a),
        bestAskQuantity: parseFloat(data.A),
        openPrice: parseFloat(data.o),
        highPrice: parseFloat(data.h),
        lowPrice: parseFloat(data.l),
        totalTradedBaseVolume: parseFloat(data.v),
        totalTradedQuoteVolume: parseFloat(data.q),
        openTime: data.O,
        closeTime: data.C,
        firstTradeId: data.F,
        lastTradeId: data.L,
        totalTrades: data.n,
      } as TickerData;

    case 'kline':
      return {
        eventType: 'kline',
        eventTime: data.E,
        symbol: data.s,
        kline: {
          startTime: data.k.t,
          closeTime: data.k.T,
          symbol: data.k.s,
          interval: data.k.i,
          firstTradeId: data.k.f,
          lastTradeId: data.k.L,
          open: parseFloat(data.k.o),
          close: parseFloat(data.k.c),
          high: parseFloat(data.k.h),
          low: parseFloat(data.k.l),
          baseAssetVolume: parseFloat(data.k.v),
          numberOfTrades: data.k.n,
          isClosed: data.k.x,
          quoteAssetVolume: parseFloat(data.k.q),
          takerBuyBaseAssetVolume: parseFloat(data.k.V),
          takerBuyQuoteAssetVolume: parseFloat(data.k.Q),
        },
      } as KlineData;

    case 'depth':
      return {
        eventType: 'depth',
        eventTime: data.E,
        symbol: data.s,
        firstUpdateId: data.U,
        finalUpdateId: data.u,
        bids: data.b || [],
        asks: data.a || [],
      } as DepthData;

    default:
      throw new Error(`Unknown stream type: ${streamType}`);
  }
}

// Export types for use in components
export type {
  WebSocketData,
  TradeData,
  TickerData,
  KlineData,
  DepthData,
  WebSocketState,
};
