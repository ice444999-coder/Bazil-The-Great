import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import ErrorBoundary from './components/ErrorBoundary'
import Login from './pages/Login'
import TradingDashboard from './pages/TradingDashboard'
import SimpleBinanceTrading from './pages/SimpleBinanceTrading'
import SOLACEConsciousnessTrading from './pages/SOLACEConsciousnessTrading'
import TestPage from './pages/TestPage'

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/test" element={<TestPage />} />
          <Route path="/login" element={<Login />} />
          <Route path="/dashboard" element={<TradingDashboard />} />
          <Route path="/trading" element={<SimpleBinanceTrading />} />
          <Route path="/solace" element={<SOLACEConsciousnessTrading />} />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App
