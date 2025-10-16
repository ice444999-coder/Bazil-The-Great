import { useState, useEffect } from 'react';

interface StopLossInputProps {
  currentPrice: number;
  direction: 'long' | 'short';
  onStopLossChange: (stopLoss: number | null, stopLossPercent: number | null) => void;
  disabled?: boolean;
}

export default function StopLossInput({
  currentPrice,
  direction,
  onStopLossChange,
  disabled = false,
}: StopLossInputProps) {
  const [enabled, setEnabled] = useState(false);
  const [stopLossPrice, setStopLossPrice] = useState<string>('');
  const [stopLossPercent, setStopLossPercent] = useState<string>('');
  const [inputMode, setInputMode] = useState<'price' | 'percent'>('percent');

  // Calculate stop loss based on direction
  const suggestedStopLoss = direction === 'long'
    ? currentPrice * 0.95 // 5% below for longs
    : currentPrice * 1.05; // 5% above for shorts

  useEffect(() => {
    if (enabled) {
      // Initialize with suggested values
      if (!stopLossPrice && !stopLossPercent) {
        setStopLossPercent('5');
        const calculatedPrice = direction === 'long'
          ? currentPrice * 0.95
          : currentPrice * 1.05;
        setStopLossPrice(calculatedPrice.toFixed(2));
        onStopLossChange(calculatedPrice, 5);
      }
    } else {
      // Clear when disabled
      setStopLossPrice('');
      setStopLossPercent('');
      onStopLossChange(null, null);
    }
  }, [enabled, currentPrice, direction]);

  const handleToggle = () => {
    setEnabled(!enabled);
  };

  const handlePriceChange = (value: string) => {
    setStopLossPrice(value);
    setInputMode('price');

    const price = parseFloat(value);
    if (!isNaN(price) && price > 0) {
      // Calculate percentage from price
      const percentDiff = Math.abs(((price - currentPrice) / currentPrice) * 100);
      setStopLossPercent(percentDiff.toFixed(2));
      onStopLossChange(price, percentDiff);
    } else {
      onStopLossChange(null, null);
    }
  };

  const handlePercentChange = (value: string) => {
    setStopLossPercent(value);
    setInputMode('percent');

    const percent = parseFloat(value);
    if (!isNaN(percent) && percent > 0 && percent <= 100) {
      // Calculate price from percentage
      const price = direction === 'long'
        ? currentPrice * (1 - percent / 100)
        : currentPrice * (1 + percent / 100);
      setStopLossPrice(price.toFixed(2));
      onStopLossChange(price, percent);
    } else {
      onStopLossChange(null, null);
    }
  };

  const handleQuickSet = (percent: number) => {
    handlePercentChange(percent.toString());
  };

  // Validate stop loss direction
  const isValidStopLoss = () => {
    const price = parseFloat(stopLossPrice);
    if (isNaN(price)) return true; // Don't show error if empty

    if (direction === 'long' && price >= currentPrice) {
      return false; // Stop loss must be below current price for longs
    }
    if (direction === 'short' && price <= currentPrice) {
      return false; // Stop loss must be above current price for shorts
    }
    return true;
  };

  const getEstimatedLoss = () => {
    const price = parseFloat(stopLossPrice);
    if (isNaN(price)) return 0;

    const percentLoss = Math.abs(((price - currentPrice) / currentPrice) * 100);
    return percentLoss;
  };

  return (
    <div
      style={{
        background: '#1E2329',
        border: '1px solid #2B3139',
        borderRadius: 8,
        padding: 16,
      }}
    >
      {/* Header with Toggle */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
        }}
      >
        <div>
          <div style={{ fontSize: 14, fontWeight: 600, color: '#FFFFFF', marginBottom: 4 }}>
            Stop Loss
          </div>
          <div style={{ fontSize: 11, color: '#848E9C' }}>
            Auto-close position to limit losses
          </div>
        </div>

        {/* Toggle Switch */}
        <div
          onClick={!disabled ? handleToggle : undefined}
          style={{
            width: 44,
            height: 24,
            borderRadius: 12,
            background: enabled ? '#0ECB81' : '#2B3139',
            cursor: disabled ? 'not-allowed' : 'pointer',
            position: 'relative',
            transition: 'background 0.2s',
            opacity: disabled ? 0.5 : 1,
          }}
        >
          <div
            style={{
              position: 'absolute',
              top: 2,
              left: enabled ? 22 : 2,
              width: 20,
              height: 20,
              borderRadius: '50%',
              background: '#FFFFFF',
              transition: 'left 0.2s',
            }}
          />
        </div>
      </div>

      {enabled && (
        <>
          {/* Input Fields */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 12 }}>
            {/* Price Input */}
            <div>
              <label style={{ display: 'block', fontSize: 12, color: '#848E9C', marginBottom: 6 }}>
                Stop Price (USDT)
              </label>
              <input
                type="number"
                value={stopLossPrice}
                onChange={(e) => handlePriceChange(e.target.value)}
                placeholder={suggestedStopLoss.toFixed(2)}
                disabled={disabled}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  background: '#0B0E11',
                  border: `1px solid ${!isValidStopLoss() ? '#F6465D' : '#2B3139'}`,
                  borderRadius: 6,
                  color: '#FFFFFF',
                  fontSize: 14,
                  outline: 'none',
                }}
              />
            </div>

            {/* Percent Input */}
            <div>
              <label style={{ display: 'block', fontSize: 12, color: '#848E9C', marginBottom: 6 }}>
                Stop Loss (%)
              </label>
              <input
                type="number"
                value={stopLossPercent}
                onChange={(e) => handlePercentChange(e.target.value)}
                placeholder="5"
                min="0"
                max="100"
                step="0.1"
                disabled={disabled}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  background: '#0B0E11',
                  border: '1px solid #2B3139',
                  borderRadius: 6,
                  color: '#FFFFFF',
                  fontSize: 14,
                  outline: 'none',
                }}
              />
            </div>
          </div>

          {/* Validation Error */}
          {!isValidStopLoss() && (
            <div
              style={{
                padding: '8px 12px',
                background: '#F6465D20',
                border: '1px solid #F6465D',
                borderRadius: 6,
                fontSize: 12,
                color: '#F6465D',
                marginBottom: 12,
              }}
            >
              ⚠️ Stop loss must be {direction === 'long' ? 'below' : 'above'} current price ({currentPrice.toFixed(2)})
            </div>
          )}

          {/* Quick Set Buttons */}
          <div style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 11, color: '#848E9C', marginBottom: 6 }}>Quick Set:</div>
            <div style={{ display: 'flex', gap: 6 }}>
              {[2, 5, 10, 15].map((percent) => (
                <button
                  key={percent}
                  onClick={() => handleQuickSet(percent)}
                  disabled={disabled}
                  style={{
                    flex: 1,
                    padding: '6px 0',
                    background: stopLossPercent === percent.toString() ? '#F0B90B20' : 'transparent',
                    border: `1px solid ${stopLossPercent === percent.toString() ? '#F0B90B' : '#2B3139'}`,
                    borderRadius: 4,
                    color: stopLossPercent === percent.toString() ? '#F0B90B' : '#FFFFFF',
                    fontSize: 12,
                    fontWeight: 600,
                    cursor: disabled ? 'not-allowed' : 'pointer',
                    transition: 'all 0.2s',
                  }}
                  onMouseEnter={(e) => {
                    if (!disabled && stopLossPercent !== percent.toString()) {
                      e.currentTarget.style.borderColor = '#F0B90B';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (stopLossPercent !== percent.toString()) {
                      e.currentTarget.style.borderColor = '#2B3139';
                    }
                  }}
                >
                  {percent}%
                </button>
              ))}
            </div>
          </div>

          {/* Info Display */}
          {isValidStopLoss() && stopLossPrice && (
            <div
              style={{
                padding: '12px',
                background: '#0B0E11',
                borderRadius: 6,
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div>
                <div style={{ fontSize: 11, color: '#848E9C', marginBottom: 4 }}>
                  Estimated Max Loss
                </div>
                <div style={{ fontSize: 16, fontWeight: 700, color: '#F6465D' }}>
                  -{getEstimatedLoss().toFixed(2)}%
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontSize: 11, color: '#848E9C', marginBottom: 4 }}>
                  Trigger Price
                </div>
                <div style={{ fontSize: 16, fontWeight: 700, color: '#FFFFFF' }}>
                  ${parseFloat(stopLossPrice).toFixed(2)}
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
