import { useState, useEffect } from 'react';

interface SimpleTradeFormProps {
  onSubmit: (data: { symbol: string; action: string; amount: number }) => Promise<void>;
  loading: boolean;
}

export function SimpleTradeForm({ onSubmit, loading }: SimpleTradeFormProps) {
  const [symbol, setSymbol] = useState('BTC');
  const [action, setAction] = useState('BUY');
  const [amount, setAmount] = useState('');
  const [currentPrice, setCurrentPrice] = useState('0.00');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Fetch current price when symbol changes
  useEffect(() => {
    const fetchPrice = async () => {
      try {
        const response = await fetch(
          `https://api.coingecko.com/api/v3/simple/price?ids=${getCoingeckoId(symbol)}&vs_currencies=usd`
        );
        const data = await response.json();
        const id = getCoingeckoId(symbol);
        if (data[id]) {
          setCurrentPrice(data[id].usd.toFixed(2));
        }
      } catch (err) {
        console.error('Failed to fetch price:', err);
      }
    };

    fetchPrice();
    const interval = setInterval(fetchPrice, 10000); // Update every 10 seconds
    return () => clearInterval(interval);
  }, [symbol]);

  const getCoingeckoId = (sym: string) => {
    const map: { [key: string]: string } = {
      BTC: 'bitcoin',
      ETH: 'ethereum',
      SOL: 'solana',
      ADA: 'cardano',
      DOT: 'polkadot',
    };
    return map[sym] || 'bitcoin';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      setError('Please enter a valid amount');
      return;
    }

    try {
      await onSubmit({ symbol, action, amount: amountNum });
      setSuccess(`Trade executed successfully! ${action} ${symbol} for $${amountNum}`);
      setAmount('');
      setTimeout(() => setSuccess(''), 5000);
    } catch (err: any) {
      setError(err.message || 'Trade execution failed');
    }
  };

  return (
    <form onSubmit={handleSubmit} style={styles.form}>
      {error && <div style={styles.errorMessage}>{error}</div>}
      {success && <div style={styles.successMessage}>{success}</div>}

      <div style={styles.formRow}>
        <div style={styles.formGroup}>
          <label style={styles.label}>Symbol</label>
          <select
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            style={styles.select}
            disabled={loading}
          >
            <option value="BTC">Bitcoin (BTC)</option>
            <option value="ETH">Ethereum (ETH)</option>
            <option value="SOL">Solana (SOL)</option>
            <option value="ADA">Cardano (ADA)</option>
            <option value="DOT">Polkadot (DOT)</option>
          </select>
        </div>

        <div style={styles.formGroup}>
          <label style={styles.label}>Action</label>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            style={styles.select}
            disabled={loading}
          >
            <option value="BUY">Buy</option>
            <option value="SELL">Sell</option>
          </select>
        </div>
      </div>

      <div style={styles.formRow}>
        <div style={styles.formGroup}>
          <label style={styles.label}>Amount ($)</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="Enter amount in USD"
            style={styles.input}
            disabled={loading}
            min="1"
            step="0.01"
          />
        </div>

        <div style={styles.formGroup}>
          <label style={styles.label}>Current Price</label>
          <input
            type="text"
            value={`$${currentPrice}`}
            readOnly
            style={styles.input}
            disabled
          />
        </div>
      </div>

      <button
        type="submit"
        style={{
          ...styles.button,
          ...(loading ? styles.buttonDisabled : {}),
        }}
        disabled={loading}
      >
        {loading ? 'Executing...' : 'Execute Trade'}
      </button>
    </form>
  );
}

const styles: { [key: string]: React.CSSProperties } = {
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '15px',
  },
  formRow: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '15px',
  },
  formGroup: {
    display: 'flex',
    flexDirection: 'column',
  },
  label: {
    display: 'block',
    marginBottom: '8px',
    fontWeight: 500,
    fontSize: '14px',
    color: '#333',
  },
  input: {
    width: '100%',
    padding: '10px 15px',
    border: '2px solid #e0e0e0',
    borderRadius: '5px',
    fontSize: '14px',
    transition: 'border-color 0.3s',
  },
  select: {
    width: '100%',
    padding: '10px 15px',
    border: '2px solid #e0e0e0',
    borderRadius: '5px',
    fontSize: '14px',
    transition: 'border-color 0.3s',
    background: 'white',
    cursor: 'pointer',
  },
  button: {
    padding: '12px 20px',
    background: '#10b981',
    color: 'white',
    border: 'none',
    borderRadius: '5px',
    fontSize: '14px',
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'background 0.3s',
  },
  buttonDisabled: {
    background: '#ccc',
    cursor: 'not-allowed',
  },
  errorMessage: {
    background: '#fee',
    color: '#c33',
    padding: '10px 15px',
    borderRadius: '5px',
  },
  successMessage: {
    background: '#efe',
    color: '#3c3',
    padding: '10px 15px',
    borderRadius: '5px',
  },
};

export type { SimpleTradeFormProps };
