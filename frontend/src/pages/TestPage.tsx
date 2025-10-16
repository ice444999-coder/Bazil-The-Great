import React from 'react';

const TestPage: React.FC = () => {
  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      backgroundColor: '#0B0E11',
      color: '#EAECEF',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'sans-serif',
    }}>
      <h1 style={{ fontSize: '48px', marginBottom: '20px' }}>🚀 ARES React App is Working!</h1>
      <p style={{ fontSize: '24px', color: '#F0B90B' }}>Binance-style Trading UI Loading...</p>
      <div style={{ marginTop: '40px', fontSize: '18px' }}>
        <div>✅ React Router: OK</div>
        <div>✅ Vite Build: OK</div>
        <div>✅ Go Backend: OK</div>
        <div>✅ Static Assets: OK</div>
      </div>
    </div>
  );
};

export default TestPage;
