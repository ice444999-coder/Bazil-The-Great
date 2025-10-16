import { useEffect, useRef, useState } from 'react';
import { createChart, IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts';

interface TradingChartProps {
  symbol: string;
  height?: number;
}

interface CandleData {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export default function TradingChart({ symbol, height = 600 }: TradingChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    // Create chart with Binance dark theme
    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: height,
      layout: {
        background: { color: '#0B0E11' },
        textColor: '#848E9C',
      },
      grid: {
        vertLines: { color: '#1E2329' },
        horzLines: { color: '#1E2329' },
      },
      crosshair: {
        mode: 1,
      },
      rightPriceScale: {
        borderColor: '#2B3139',
      },
      timeScale: {
        borderColor: '#2B3139',
        timeVisible: true,
        secondsVisible: false,
      },
    });

    chartRef.current = chart;

    // Add candlestick series
    const candlestickSeries = (chart as any).addCandlestickSeries({
      upColor: '#0ECB81',
      downColor: '#F6465D',
      borderUpColor: '#0ECB81',
      borderDownColor: '#F6465D',
      wickUpColor: '#0ECB81',
      wickDownColor: '#F6465D',
    });

    candlestickSeriesRef.current = candlestickSeries;

    // Add volume series (bottom pane)
    const volumeSeries = (chart as any).addHistogramSeries({
      color: '#26a69a',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '',
      scaleMargins: {
        top: 0.8,
        bottom: 0,
      },
    });

    volumeSeriesRef.current = volumeSeries;

    // Fetch initial data
    fetchCandleData(symbol)
      .then((data) => {
        if (data.length === 0) {
          setError('No data available');
          setLoading(false);
          return;
        }

        const candleData: CandlestickData<Time>[] = data.map((candle) => ({
          time: (candle.time / 1000) as Time,
          open: candle.open,
          high: candle.high,
          low: candle.low,
          close: candle.close,
        }));

        const volumeData = data.map((candle) => ({
          time: (candle.time / 1000) as Time,
          value: candle.volume,
          color: candle.close >= candle.open ? '#0ECB8150' : '#F6465D50',
        }));

        candlestickSeries.setData(candleData);
        volumeSeries.setData(volumeData);

        // Fit content to view
        chart.timeScale().fitContent();
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message || 'Failed to load chart data');
        setLoading(false);
      });

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
        });
      }
    };

    window.addEventListener('resize', handleResize);

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, [symbol, height]);

  return (
    <div style={{ position: 'relative', width: '100%', height: height }}>
      {loading && (
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: '#0B0E11',
            color: '#848E9C',
            zIndex: 10,
          }}
        >
          <div>
            <div
              style={{
                width: 40,
                height: 40,
                border: '3px solid #2B3139',
                borderTopColor: '#F0B90B',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite',
                margin: '0 auto 10px',
              }}
            />
            <div>Loading {symbol} chart...</div>
          </div>
        </div>
      )}

      {error && (
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: '#0B0E11',
            color: '#F6465D',
            zIndex: 10,
          }}
        >
          <div>
            <div style={{ fontSize: 18, marginBottom: 8 }}>⚠️ Error</div>
            <div>{error}</div>
          </div>
        </div>
      )}

      <div ref={chartContainerRef} style={{ width: '100%', height: '100%' }} />

      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}

// Fetch candle data from Binance API
async function fetchCandleData(symbol: string): Promise<CandleData[]> {
  try {
    // Binance API expects symbols without slash (e.g., BTCUSDT not BTC/USDT)
    const binanceSymbol = symbol.replace('/', '').toUpperCase();
    
    // Fetch 1-hour candles for last 7 days (168 candles)
    const response = await fetch(
      `https://api.binance.com/api/v3/klines?symbol=${binanceSymbol}&interval=1h&limit=168`
    );

    if (!response.ok) {
      throw new Error(`Binance API error: ${response.status}`);
    }

    const data = await response.json();

    // Binance klines format: [time, open, high, low, close, volume, ...]
    return data.map((candle: any[]) => ({
      time: candle[0], // Timestamp in milliseconds
      open: parseFloat(candle[1]),
      high: parseFloat(candle[2]),
      low: parseFloat(candle[3]),
      close: parseFloat(candle[4]),
      volume: parseFloat(candle[5]),
    }));
  } catch (error) {
    console.error('Error fetching candle data:', error);
    throw error;
  }
}
