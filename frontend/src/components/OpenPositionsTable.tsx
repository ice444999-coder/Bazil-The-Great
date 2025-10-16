import styles from './OpenPositionsTable.module.css';

interface Trade {
  id: number;
  trading_pair: string;
  direction: 'BUY' | 'SELL';
  size: number;
  entry_price: number;
  fees: number;
  opened_at: string;
}

interface OpenPositionsTableProps {
  positions: Trade[];
  onClose: (id: number) => void;
  loading?: boolean;
}

export default function OpenPositionsTable({ positions, onClose, loading }: OpenPositionsTableProps) {
  if (positions.length === 0) {
    return <div className={styles.emptyState}>No open positions</div>;
  }

  return (
    <div className={styles.tableContainer}>
      <table className={styles.table}>
        <thead className={styles.thead}>
          <tr>
            <th className={styles.th}>ID</th>
            <th className={styles.th}>Pair</th>
            <th className={styles.th}>Direction</th>
            <th className={styles.th}>Size</th>
            <th className={styles.th}>Entry Price</th>
            <th className={styles.th}>Fees</th>
            <th className={styles.th}>Opened</th>
            <th className={styles.th}>Action</th>
          </tr>
        </thead>
        <tbody className={styles.tbody}>
          {positions.map((pos) => (
            <tr key={pos.id} className={styles.tr}>
              <td className={styles.td}>
                <strong>#{pos.id}</strong>
              </td>
              <td className={styles.td}>
                <strong>{pos.trading_pair}</strong>
              </td>
              <td className={pos.direction === 'BUY' ? styles.tdLong : styles.tdShort}>
                {pos.direction}
              </td>
              <td className={styles.td}>${pos.size.toFixed(2)}</td>
              <td className={styles.td}>${pos.entry_price.toFixed(4)}</td>
              <td className={styles.td}>${pos.fees.toFixed(2)}</td>
              <td className={styles.td}>{new Date(pos.opened_at).toLocaleString()}</td>
              <td className={styles.td}>
                <button
                  className={styles.closeButton}
                  onClick={() => onClose(pos.id)}
                  disabled={loading}
                >
                  Close
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export type { OpenPositionsTableProps, Trade };
