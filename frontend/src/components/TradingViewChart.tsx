import { useEffect, useRef } from 'react';

declare global {
  interface Window {
    TradingView: any;
  }
}

interface TradingViewChartProps {
  symbol?: string;
  interval?: string;
  theme?: 'light' | 'dark';
  height?: number;
}

export const TradingViewChart: React.FC<TradingViewChartProps> = ({
  symbol = 'BTCUSDT',
  interval = '60',
  theme = 'dark',
  height = 600
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetRef = useRef<any>(null);

  useEffect(() => {
    // Clean up previous widget
    if (widgetRef.current) {
      try {
        widgetRef.current.remove();
      } catch (e) {
        console.log('Widget cleanup skipped');
      }
    }

    // Load TradingView widget script if not already loaded
    const existingScript = document.getElementById('tradingview-widget-script');
    
    const initWidget = () => {
      if (containerRef.current && window.TradingView) {
        try {
          widgetRef.current = new window.TradingView.widget({
            autosize: true,
            symbol: `BINANCE:${symbol}`,
            interval: interval,
            timezone: 'Etc/UTC',
            theme: theme,
            style: '1',
            locale: 'en',
            toolbar_bg: '#0B0E11',
            enable_publishing: false,
            backgroundColor: '#0B0E11',
            gridColor: '#1E2329',
            hide_top_toolbar: false,
            hide_legend: false,
            save_image: false,
            container_id: 'tradingview_chart',
            studies: [
              'RSI@tv-basicstudies',
              'MACD@tv-basicstudies',
              'MAExp@tv-basicstudies'
            ]
          });
        } catch (error) {
          console.error('Error initializing TradingView widget:', error);
        }
      }
    };

    if (!existingScript) {
      const script = document.createElement('script');
      script.id = 'tradingview-widget-script';
      script.src = 'https://s3.tradingview.com/tv.js';
      script.async = true;
      script.onload = initWidget;
      script.onerror = () => {
        console.error('Failed to load TradingView script');
      };
      document.head.appendChild(script);
    } else if (window.TradingView) {
      initWidget();
    }

    return () => {
      if (widgetRef.current) {
        try {
          widgetRef.current.remove();
        } catch (e) {
          // Ignore cleanup errors
        }
      }
    };
  }, [symbol, interval, theme]);

  return (
    <div 
      ref={containerRef}
      id="tradingview_chart"
      style={{ 
        height: `${height}px`,
        backgroundColor: '#0B0E11',
        borderRadius: '8px',
        overflow: 'hidden'
      }}
    />
  );
};

export default TradingViewChart;
