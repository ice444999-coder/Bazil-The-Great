// ============================================
// FILE 2: src/components/TradeHistoryTable.tsx
// ============================================

interface TradeHistory {
  id: number;
  trading_pair: string;
  direction: 'BUY' | 'SELL';
  size: number;
  entry_price: number;
  status: 'OPEN' | 'CLOSED';
  opened_at: string;
}

interface TradeHistoryTableProps {
  trades: TradeHistory[];
}

export function TradeHistoryTable({ trades }: TradeHistoryTableProps) {
  if (trades.length === 0) {
    return <div style={styles.emptyState}>No trade history yet</div>;
  }

  // Show only last 20 trades
  const recentTrades = trades.slice(0, 20);

  return (
    <div style={styles.tableContainer}>
      <table style={styles.table}>
        <thead style={styles.thead}>
          <tr>
            <th style={styles.th}>ID</th>
            <th style={styles.th}>Pair</th>
            <th style={styles.th}>Direction</th>
            <th style={styles.th}>Size</th>
            <th style={styles.th}>Entry Price</th>
            <th style={styles.th}>Status</th>
            <th style={styles.th}>Executed</th>
          </tr>
        </thead>
        <tbody>
          {recentTrades.map((trade) => (
            <tr key={trade.id} style={styles.tr}>
              <td style={styles.td}>
                <strong>#{trade.id}</strong>
              </td>
              <td style={styles.td}>
                <strong>{trade.trading_pair}</strong>
              </td>
              <td style={{ ...styles.td, color: trade.direction === 'BUY' ? '#10b981' : '#ef4444' }}>
                {trade.direction}
              </td>
              <td style={styles.td}>${trade.size.toFixed(2)}</td>
              <td style={styles.td}>${trade.entry_price.toFixed(4)}</td>
              <td style={styles.td}>
                <span
                  style={{
                    ...styles.badge,
                    ...(trade.status === 'OPEN' ? styles.badgeSuccess : styles.badgeInfo),
                  }}
                >
                  {trade.status}
                </span>
              </td>
              <td style={styles.td}>{new Date(trade.opened_at).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export type { TradeHistoryTableProps, TradeHistory };

const styles: { [key: string]: React.CSSProperties } = {
  tableContainer: {
    overflowX: 'auto',
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse',
  },
  thead: {
    background: '#f9fafb',
  },
  th: {
    textAlign: 'left',
    padding: '12px',
    fontWeight: 600,
    fontSize: '13px',
    color: '#666',
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  },
  tr: {
    transition: 'background 0.2s',
    cursor: 'pointer',
  },
  td: {
    padding: '12px',
    borderTop: '1px solid #f0f0f0',
  },
  badge: {
    padding: '4px 12px',
    borderRadius: '12px',
    fontSize: '12px',
    fontWeight: 600,
  },
  badgeSuccess: {
    background: '#d1fae5',
    color: '#065f46',
  },
  badgeInfo: {
    background: '#dbeafe',
    color: '#1e40af',
  },
  emptyState: {
    textAlign: 'center',
    padding: '40px',
    color: '#999',
  },
};
