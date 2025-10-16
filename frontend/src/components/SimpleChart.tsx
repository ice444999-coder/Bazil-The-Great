interface SimpleChartProps {
  symbol: string;
  height?: number;
}

export const SimpleChart: React.FC<SimpleChartProps> = ({ symbol, height = 600 }) => {
  return (
    <div style={{
      height: `${height}px`,
      backgroundColor: '#0B0E11',
      borderRadius: '8px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      flexDirection: 'column',
      gap: '20px',
      border: '1px solid #2B3139'
    }}>
      <div style={{ fontSize: '48px' }}>📈</div>
      <div style={{ fontSize: '24px', color: '#EAECEF', fontWeight: '600' }}>
        {symbol} Chart
      </div>
      <div style={{ fontSize: '14px', color: '#848E9C' }}>
        Professional chart loading...
      </div>
      <div style={{
        padding: '12px 24px',
        backgroundColor: '#1E2329',
        borderRadius: '6px',
        color: '#F0B90B',
        fontSize: '13px',
        fontFamily: 'monospace'
      }}>
        Chart component will render here when TradingChart is fully loaded
      </div>
    </div>
  );
};

export default SimpleChart;
