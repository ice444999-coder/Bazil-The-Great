import { useState, useEffect } from 'react';

interface TakeProfitInputProps {
  currentPrice: number;
  direction: 'long' | 'short';
  onTakeProfitChange: (takeProfit: number | null, takeProfitPercent: number | null) => void;
  disabled?: boolean;
}

export default function TakeProfitInput({
  currentPrice,
  direction,
  onTakeProfitChange,
  disabled = false,
}: TakeProfitInputProps) {
  const [enabled, setEnabled] = useState(false);
  const [takeProfitPrice, setTakeProfitPrice] = useState<string>('');
  const [takeProfitPercent, setTakeProfitPercent] = useState<string>('');
  const [inputMode, setInputMode] = useState<'price' | 'percent'>('percent');

  // Calculate take profit based on direction
  const suggestedTakeProfit = direction === 'long'
    ? currentPrice * 1.10 // 10% above for longs
    : currentPrice * 0.90; // 10% below for shorts

  useEffect(() => {
    if (enabled) {
      // Initialize with suggested values
      if (!takeProfitPrice && !takeProfitPercent) {
        setTakeProfitPercent('10');
        const calculatedPrice = direction === 'long'
          ? currentPrice * 1.10
          : currentPrice * 0.90;
        setTakeProfitPrice(calculatedPrice.toFixed(2));
        onTakeProfitChange(calculatedPrice, 10);
      }
    } else {
      // Clear when disabled
      setTakeProfitPrice('');
      setTakeProfitPercent('');
      onTakeProfitChange(null, null);
    }
  }, [enabled, currentPrice, direction]);

  const handleToggle = () => {
    setEnabled(!enabled);
  };

  const handlePriceChange = (value: string) => {
    setTakeProfitPrice(value);
    setInputMode('price');

    const price = parseFloat(value);
    if (!isNaN(price) && price > 0) {
      // Calculate percentage from price
      const percentDiff = Math.abs(((price - currentPrice) / currentPrice) * 100);
      setTakeProfitPercent(percentDiff.toFixed(2));
      onTakeProfitChange(price, percentDiff);
    } else {
      onTakeProfitChange(null, null);
    }
  };

  const handlePercentChange = (value: string) => {
    setTakeProfitPercent(value);
    setInputMode('percent');

    const percent = parseFloat(value);
    if (!isNaN(percent) && percent > 0 && percent <= 1000) {
      // Calculate price from percentage
      const price = direction === 'long'
        ? currentPrice * (1 + percent / 100)
        : currentPrice * (1 - percent / 100);
      setTakeProfitPrice(price.toFixed(2));
      onTakeProfitChange(price, percent);
    } else {
      onTakeProfitChange(null, null);
    }
  };

  const handleQuickSet = (percent: number) => {
    handlePercentChange(percent.toString());
  };

  // Validate take profit direction
  const isValidTakeProfit = () => {
    const price = parseFloat(takeProfitPrice);
    if (isNaN(price)) return true; // Don't show error if empty

    if (direction === 'long' && price <= currentPrice) {
      return false; // Take profit must be above current price for longs
    }
    if (direction === 'short' && price >= currentPrice) {
      return false; // Take profit must be below current price for shorts
    }
    return true;
  };

  const getEstimatedProfit = () => {
    const price = parseFloat(takeProfitPrice);
    if (isNaN(price)) return 0;

    const percentProfit = Math.abs(((price - currentPrice) / currentPrice) * 100);
    return percentProfit;
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
            Take Profit
          </div>
          <div style={{ fontSize: 11, color: '#848E9C' }}>
            Auto-close position to lock in gains
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
                Target Price (USDT)
              </label>
              <input
                type="number"
                value={takeProfitPrice}
                onChange={(e) => handlePriceChange(e.target.value)}
                placeholder={suggestedTakeProfit.toFixed(2)}
                disabled={disabled}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  background: '#0B0E11',
                  border: `1px solid ${!isValidTakeProfit() ? '#F6465D' : '#2B3139'}`,
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
                Take Profit (%)
              </label>
              <input
                type="number"
                value={takeProfitPercent}
                onChange={(e) => handlePercentChange(e.target.value)}
                placeholder="10"
                min="0"
                max="1000"
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
          {!isValidTakeProfit() && (
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
              ⚠️ Take profit must be {direction === 'long' ? 'above' : 'below'} current price ({currentPrice.toFixed(2)})
            </div>
          )}

          {/* Quick Set Buttons */}
          <div style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 11, color: '#848E9C', marginBottom: 6 }}>Quick Set:</div>
            <div style={{ display: 'flex', gap: 6 }}>
              {[5, 10, 20, 50].map((percent) => (
                <button
                  key={percent}
                  onClick={() => handleQuickSet(percent)}
                  disabled={disabled}
                  style={{
                    flex: 1,
                    padding: '6px 0',
                    background: takeProfitPercent === percent.toString() ? '#0ECB8120' : 'transparent',
                    border: `1px solid ${takeProfitPercent === percent.toString() ? '#0ECB81' : '#2B3139'}`,
                    borderRadius: 4,
                    color: takeProfitPercent === percent.toString() ? '#0ECB81' : '#FFFFFF',
                    fontSize: 12,
                    fontWeight: 600,
                    cursor: disabled ? 'not-allowed' : 'pointer',
                    transition: 'all 0.2s',
                  }}
                  onMouseEnter={(e) => {
                    if (!disabled && takeProfitPercent !== percent.toString()) {
                      e.currentTarget.style.borderColor = '#0ECB81';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (takeProfitPercent !== percent.toString()) {
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
          {isValidTakeProfit() && takeProfitPrice && (
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
                  Estimated Profit
                </div>
                <div style={{ fontSize: 16, fontWeight: 700, color: '#0ECB81' }}>
                  +{getEstimatedProfit().toFixed(2)}%
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontSize: 11, color: '#848E9C', marginBottom: 4 }}>
                  Target Price
                </div>
                <div style={{ fontSize: 16, fontWeight: 700, color: '#FFFFFF' }}>
                  ${parseFloat(takeProfitPrice).toFixed(2)}
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
