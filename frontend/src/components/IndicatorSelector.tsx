import { useState } from 'react';

interface IndicatorSelectorProps {
  onIndicatorToggle: (indicatorId: string, enabled: boolean, settings?: any) => void;
  availableIndicators?: Indicator[];
}

interface Indicator {
  id: string;
  name: string;
  description: string;
  category: 'trend' | 'momentum' | 'volatility' | 'volume';
  defaultEnabled: boolean;
  settings?: {
    period?: number;
    color?: string;
    [key: string]: any;
  };
}

const DEFAULT_INDICATORS: Indicator[] = [
  {
    id: 'sma',
    name: 'SMA',
    description: 'Simple Moving Average',
    category: 'trend',
    defaultEnabled: false,
    settings: { period: 20, color: '#F0B90B' },
  },
  {
    id: 'ema',
    name: 'EMA',
    description: 'Exponential Moving Average',
    category: 'trend',
    defaultEnabled: false,
    settings: { period: 12, color: '#0ECB81' },
  },
  {
    id: 'rsi',
    name: 'RSI',
    description: 'Relative Strength Index',
    category: 'momentum',
    defaultEnabled: false,
    settings: { period: 14, overbought: 70, oversold: 30 },
  },
  {
    id: 'macd',
    name: 'MACD',
    description: 'Moving Average Convergence Divergence',
    category: 'momentum',
    defaultEnabled: false,
    settings: { fast: 12, slow: 26, signal: 9 },
  },
  {
    id: 'bb',
    name: 'BB',
    description: 'Bollinger Bands',
    category: 'volatility',
    defaultEnabled: false,
    settings: { period: 20, stdDev: 2, color: '#A020F0' },
  },
  {
    id: 'atr',
    name: 'ATR',
    description: 'Average True Range',
    category: 'volatility',
    defaultEnabled: false,
    settings: { period: 14 },
  },
  {
    id: 'volume',
    name: 'Volume',
    description: 'Trading Volume',
    category: 'volume',
    defaultEnabled: true,
    settings: {},
  },
  {
    id: 'stoch',
    name: 'Stochastic',
    description: 'Stochastic Oscillator',
    category: 'momentum',
    defaultEnabled: false,
    settings: { kPeriod: 14, dPeriod: 3 },
  },
];

export default function IndicatorSelector({ 
  onIndicatorToggle,
  availableIndicators = DEFAULT_INDICATORS 
}: IndicatorSelectorProps) {
  const [enabledIndicators, setEnabledIndicators] = useState<Set<string>>(
    new Set(availableIndicators.filter((ind) => ind.defaultEnabled).map((ind) => ind.id))
  );
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState<string | null>(null);

  const handleToggle = (indicator: Indicator) => {
    const newEnabled = new Set(enabledIndicators);
    const willBeEnabled = !enabledIndicators.has(indicator.id);

    if (willBeEnabled) {
      newEnabled.add(indicator.id);
    } else {
      newEnabled.delete(indicator.id);
    }

    setEnabledIndicators(newEnabled);
    onIndicatorToggle(indicator.id, willBeEnabled, indicator.settings);
  };

  const categorizeIndicators = () => {
    const categories: { [key: string]: Indicator[] } = {
      trend: [],
      momentum: [],
      volatility: [],
      volume: [],
    };

    availableIndicators.forEach((indicator) => {
      categories[indicator.category].push(indicator);
    });

    return categories;
  };

  const categories = categorizeIndicators();
  const activeCount = enabledIndicators.size;

  return (
    <div style={{ position: 'relative' }}>
      {/* Main Button */}
      <button
        onClick={() => setDropdownOpen(!dropdownOpen)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '8px 16px',
          background: '#1E2329',
          border: '1px solid #2B3139',
          borderRadius: 6,
          color: '#FFFFFF',
          fontSize: 13,
          fontWeight: 600,
          cursor: 'pointer',
          transition: 'all 0.2s',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.borderColor = '#F0B90B';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = '#2B3139';
        }}
      >
        <span>📊</span>
        <span>Indicators</span>
        {activeCount > 0 && (
          <span
            style={{
              padding: '2px 6px',
              background: '#F0B90B',
              color: '#0B0E11',
              borderRadius: 10,
              fontSize: 11,
              fontWeight: 700,
            }}
          >
            {activeCount}
          </span>
        )}
        <span style={{ marginLeft: 4 }}>▼</span>
      </button>

      {/* Dropdown Menu */}
      {dropdownOpen && (
        <>
          {/* Backdrop */}
          <div
            onClick={() => setDropdownOpen(false)}
            style={{
              position: 'fixed',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              zIndex: 999,
            }}
          />

          {/* Dropdown Content */}
          <div
            style={{
              position: 'absolute',
              top: '100%',
              left: 0,
              marginTop: 8,
              background: '#1E2329',
              border: '1px solid #2B3139',
              borderRadius: 8,
              padding: 12,
              minWidth: 320,
              maxHeight: 480,
              overflowY: 'auto',
              zIndex: 1000,
              boxShadow: '0 8px 24px rgba(0, 0, 0, 0.5)',
            }}
          >
            {/* Header */}
            <div
              style={{
                marginBottom: 12,
                paddingBottom: 12,
                borderBottom: '1px solid #2B3139',
              }}
            >
              <div style={{ fontSize: 14, fontWeight: 600, color: '#FFFFFF', marginBottom: 4 }}>
                Technical Indicators
              </div>
              <div style={{ fontSize: 11, color: '#848E9C' }}>
                Select indicators to display on chart
              </div>
            </div>

            {/* Categories */}
            {Object.entries(categories).map(([category, indicators]) => (
              <div key={category} style={{ marginBottom: 16 }}>
                {/* Category Header */}
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: '#848E9C',
                    textTransform: 'uppercase',
                    letterSpacing: '0.5px',
                    marginBottom: 8,
                  }}
                >
                  {category}
                </div>

                {/* Indicators in Category */}
                {indicators.map((indicator) => {
                  const isEnabled = enabledIndicators.has(indicator.id);

                  return (
                    <div
                      key={indicator.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '8px 12px',
                        marginBottom: 4,
                        background: isEnabled ? '#F0B90B10' : 'transparent',
                        border: `1px solid ${isEnabled ? '#F0B90B40' : 'transparent'}`,
                        borderRadius: 6,
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                      }}
                      onClick={() => handleToggle(indicator)}
                      onMouseEnter={(e) => {
                        if (!isEnabled) {
                          e.currentTarget.style.background = '#2B3139';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (!isEnabled) {
                          e.currentTarget.style.background = 'transparent';
                        }
                      }}
                    >
                      <div style={{ flex: 1 }}>
                        <div
                          style={{
                            fontSize: 13,
                            fontWeight: 600,
                            color: isEnabled ? '#F0B90B' : '#FFFFFF',
                            marginBottom: 2,
                          }}
                        >
                          {indicator.name}
                        </div>
                        <div style={{ fontSize: 11, color: '#848E9C' }}>
                          {indicator.description}
                        </div>
                      </div>

                      {/* Toggle Checkbox */}
                      <div
                        style={{
                          width: 20,
                          height: 20,
                          borderRadius: 4,
                          border: `2px solid ${isEnabled ? '#F0B90B' : '#2B3139'}`,
                          background: isEnabled ? '#F0B90B' : 'transparent',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all 0.2s',
                        }}
                      >
                        {isEnabled && (
                          <span style={{ color: '#0B0E11', fontSize: 14, fontWeight: 700 }}>
                            ✓
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}

            {/* Footer */}
            <div
              style={{
                marginTop: 12,
                paddingTop: 12,
                borderTop: '1px solid #2B3139',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div style={{ fontSize: 11, color: '#848E9C' }}>
                {activeCount} {activeCount === 1 ? 'indicator' : 'indicators'} active
              </div>
              <button
                onClick={() => setDropdownOpen(false)}
                style={{
                  padding: '6px 12px',
                  background: '#F0B90B',
                  border: 'none',
                  borderRadius: 4,
                  color: '#0B0E11',
                  fontSize: 12,
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                Done
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export { DEFAULT_INDICATORS };
export type { Indicator };
