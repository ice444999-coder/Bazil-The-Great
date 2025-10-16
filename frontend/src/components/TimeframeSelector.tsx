import { useState } from 'react';

interface TimeframeSelectorProps {
  onTimeframeChange: (timeframe: string, interval: string) => void;
  defaultTimeframe?: string;
}

interface Timeframe {
  label: string;
  value: string;
  interval: string; // Binance API interval format
}

const TIMEFRAMES: Timeframe[] = [
  { label: '1m', value: '1m', interval: '1m' },
  { label: '3m', value: '3m', interval: '3m' },
  { label: '5m', value: '5m', interval: '5m' },
  { label: '15m', value: '15m', interval: '15m' },
  { label: '30m', value: '30m', interval: '30m' },
  { label: '1H', value: '1h', interval: '1h' },
  { label: '2H', value: '2h', interval: '2h' },
  { label: '4H', value: '4h', interval: '4h' },
  { label: '6H', value: '6h', interval: '6h' },
  { label: '12H', value: '12h', interval: '12h' },
  { label: '1D', value: '1d', interval: '1d' },
  { label: '3D', value: '3d', interval: '3d' },
  { label: '1W', value: '1w', interval: '1w' },
  { label: '1M', value: '1M', interval: '1M' },
];

export default function TimeframeSelector({ 
  onTimeframeChange, 
  defaultTimeframe = '1h' 
}: TimeframeSelectorProps) {
  const [selectedTimeframe, setSelectedTimeframe] = useState(defaultTimeframe);

  const handleTimeframeClick = (timeframe: Timeframe) => {
    setSelectedTimeframe(timeframe.value);
    onTimeframeChange(timeframe.value, timeframe.interval);
  };

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        background: '#1E2329',
        padding: '8px 12px',
        borderRadius: 6,
        border: '1px solid #2B3139',
      }}
    >
      {/* Label */}
      <div
        style={{
          fontSize: 12,
          color: '#848E9C',
          fontWeight: 600,
          marginRight: 4,
        }}
      >
        Time:
      </div>

      {/* Timeframe Buttons */}
      <div
        style={{
          display: 'flex',
          gap: 4,
          flexWrap: 'wrap',
        }}
      >
        {TIMEFRAMES.map((timeframe) => {
          const isSelected = selectedTimeframe === timeframe.value;
          
          return (
            <button
              key={timeframe.value}
              onClick={() => handleTimeframeClick(timeframe)}
              style={{
                padding: '6px 12px',
                background: isSelected ? '#F0B90B' : 'transparent',
                color: isSelected ? '#0B0E11' : '#FFFFFF',
                border: isSelected ? 'none' : '1px solid #2B3139',
                borderRadius: 4,
                fontSize: 12,
                fontWeight: 600,
                cursor: 'pointer',
                transition: 'all 0.2s',
                minWidth: 42,
              }}
              onMouseEnter={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.background = '#2B3139';
                  e.currentTarget.style.borderColor = '#F0B90B';
                }
              }}
              onMouseLeave={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.background = 'transparent';
                  e.currentTarget.style.borderColor = '#2B3139';
                }
              }}
            >
              {timeframe.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

// Export timeframes for use in other components
export { TIMEFRAMES };
export type { Timeframe };
