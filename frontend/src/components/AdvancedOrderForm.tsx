import { useState } from 'react';
import StopLossInput from './StopLossInput';
import TakeProfitInput from './TakeProfitInput';
import styles from './AdvancedOrderForm.module.css';

interface AdvancedOrderFormProps {
  symbol: string;
  currentPrice: number;
  onSubmit: (order: OrderData) => void;
  loading?: boolean;
}

interface OrderData {
  symbol: string;
  orderType: 'market' | 'limit';
  direction: 'long' | 'short';
  amount: number;
  price?: number;
  stopLoss?: number;
  stopLossPercent?: number;
  takeProfit?: number;
  takeProfitPercent?: number;
}

export default function AdvancedOrderForm({
  symbol,
  currentPrice,
  onSubmit,
  loading = false,
}: AdvancedOrderFormProps) {
  const [orderType, setOrderType] = useState<'market' | 'limit'>('market');
  const [direction, setDirection] = useState<'long' | 'short'>('long');
  const [amount, setAmount] = useState<string>('');
  const [limitPrice, setLimitPrice] = useState<string>('');
  const [stopLoss, setStopLoss] = useState<{ price: number | null; percent: number | null }>({
    price: null,
    percent: null,
  });
  const [takeProfit, setTakeProfit] = useState<{ price: number | null; percent: number | null }>({
    price: null,
    percent: null,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      alert('Please enter a valid amount');
      return;
    }

    if (orderType === 'limit') {
      const limitPriceNum = parseFloat(limitPrice);
      if (isNaN(limitPriceNum) || limitPriceNum <= 0) {
        alert('Please enter a valid limit price');
        return;
      }
    }

    const orderData: OrderData = {
      symbol,
      orderType,
      direction,
      amount: amountNum,
      price: orderType === 'limit' ? parseFloat(limitPrice) : undefined,
      stopLoss: stopLoss.price || undefined,
      stopLossPercent: stopLoss.percent || undefined,
      takeProfit: takeProfit.price || undefined,
      takeProfitPercent: takeProfit.percent || undefined,
    };

    onSubmit(orderData);
  };

  const getEstimatedCost = () => {
    const amountNum = parseFloat(amount);
    if (isNaN(amountNum)) return 0;

    const price = orderType === 'limit' ? parseFloat(limitPrice) : currentPrice;
    return amountNum * price;
  };

  const getEstimatedPnL = () => {
    if (!takeProfit.price && !stopLoss.price) return null;

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum)) return null;

    const entryPrice = orderType === 'limit' ? parseFloat(limitPrice) : currentPrice;

    let maxProfit = 0;
    let maxLoss = 0;

    if (takeProfit.price) {
      maxProfit = direction === 'long'
        ? (takeProfit.price - entryPrice) * amountNum
        : (entryPrice - takeProfit.price) * amountNum;
    }

    if (stopLoss.price) {
      maxLoss = direction === 'long'
        ? (stopLoss.price - entryPrice) * amountNum
        : (entryPrice - stopLoss.price) * amountNum;
    }

    return { maxProfit, maxLoss };
  };

  const pnl = getEstimatedPnL();

  return (
    <form onSubmit={handleSubmit}>
      <div className={styles.container}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.title}>Place Order</div>
          <div className={styles.subtitle}>
            {symbol} • Current Price: ${currentPrice.toFixed(2)}
          </div>
        </div>

        {/* Order Type Selector */}
        <div className={styles.formGroup}>
          <label className={styles.label}>Order Type</label>
          <div className={styles.buttonGroup}>
            {(['market', 'limit'] as const).map((type) => (
              <button
                key={type}
                type="button"
                onClick={() => setOrderType(type)}
                className={orderType === type ? styles.typeButtonActive : styles.typeButton}
              >
                {type}
              </button>
            ))}
          </div>
        </div>

        {/* Direction Selector */}
        <div className={styles.formGroup}>
          <label className={styles.label}>Direction</label>
          <div className={styles.buttonGroup}>
            <button
              type="button"
              onClick={() => setDirection('long')}
              className={`${styles.sideButtonBuy} ${direction === 'long' ? styles.sideButtonActive : ''}`}
            >
              BUY / LONG
            </button>
            <button
              type="button"
              onClick={() => setDirection('short')}
              className={`${styles.sideButtonSell} ${direction === 'short' ? styles.sideButtonActive : ''}`}
            >
              SELL / SHORT
            </button>
          </div>
        </div>

        {/* Limit Price */}
        {orderType === 'limit' && (
          <div className={styles.formGroup}>
            <label className={styles.label}>Limit Price (USDT)</label>
            <input
              type="number"
              value={limitPrice}
              onChange={(e) => setLimitPrice(e.target.value)}
              placeholder={currentPrice.toFixed(2)}
              step="0.01"
              className={styles.input}
            />
          </div>
        )}

        {/* Amount Input */}
        <div className={styles.formGroup}>
          <label className={styles.label}>Amount (USDT)</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="100.00"
            min="0"
            step="0.01"
            className={styles.input}
          />
          {amount && (
            <div className={styles.helperText}>
              Estimated Cost: ${getEstimatedCost().toFixed(2)} USDT
            </div>
          )}
        </div>

        {/* Stop Loss */}
        <div className={styles.formGroup}>
          <StopLossInput
            currentPrice={orderType === 'limit' ? parseFloat(limitPrice) || currentPrice : currentPrice}
            direction={direction}
            onStopLossChange={(price, percent) => setStopLoss({ price, percent })}
            disabled={loading}
          />
        </div>

        {/* Take Profit */}
        <div className={styles.riskSummary}>
          <TakeProfitInput
            currentPrice={orderType === 'limit' ? parseFloat(limitPrice) || currentPrice : currentPrice}
            direction={direction}
            onTakeProfitChange={(price, percent) => setTakeProfit({ price, percent })}
            disabled={loading}
          />
        </div>

        {/* PnL Summary */}
        {pnl && (pnl.maxProfit > 0 || pnl.maxLoss < 0) && (
          <div className={styles.riskBox}>
            <div className={styles.riskLabel}>Risk/Reward Summary</div>
            <div className={styles.riskMetrics}>
              {pnl.maxLoss < 0 && (
                <div className={styles.riskMetric}>
                  <div className={styles.metricLabel}>Max Loss</div>
                  <div className={styles.metricValueLoss}>
                    ${Math.abs(pnl.maxLoss).toFixed(2)}
                  </div>
                  <div className={styles.metricSubtextLoss}>
                    {stopLoss.percent?.toFixed(2)}%
                  </div>
                </div>
              )}
              {pnl.maxProfit > 0 && (
                <div className={styles.riskMetricRight}>
                  <div className={styles.metricLabel}>Max Profit</div>
                  <div className={styles.metricValueProfit}>
                    ${pnl.maxProfit.toFixed(2)}
                  </div>
                  <div className={styles.metricSubtextProfit}>
                    {takeProfit.percent?.toFixed(2)}%
                  </div>
                </div>
              )}
            </div>
            {pnl.maxProfit > 0 && pnl.maxLoss < 0 && (
              <div className={styles.marginInfo}>
                <div className={styles.marginLabel}>Risk/Reward Ratio</div>
                <div className={styles.marginValue}>
                  1 : {(Math.abs(pnl.maxProfit / pnl.maxLoss)).toFixed(2)}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          disabled={loading || !amount}
          className={styles.submitButton}
        >
          {loading ? 'Placing Order...' : `Place ${direction} Order`}
        </button>
      </div>
    </form>
  );
}

export type { OrderData };
