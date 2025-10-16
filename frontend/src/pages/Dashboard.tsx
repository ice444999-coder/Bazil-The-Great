import { useState, useEffect } from 'react'
import { tradingApi } from '../lib/api'
import { useAuthStore } from '../stores/authStore'

interface Trade {
  id: number
  trading_pair: string
  direction: string
  size: number
  entry_price: number
  exit_price?: number
  profit_loss?: number
  status: string
  reasoning?: string
  confidence_score?: number
  created_at: string
}

export default function Dashboard() {
  const { clearToken } = useAuthStore()
  const [trades, setTrades] = useState<Trade[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const loadTrades = async () => {
      try {
        const data = await tradingApi.getHistory()
        setTrades(data)
      } catch (err: any) {
        setError(err.message || 'Failed to load trades')
      } finally {
        setLoading(false)
      }
    }

    loadTrades()
  }, [])

  return (
    <div style={{
      minHeight: '100vh',
      background: '#0B0E11',
      color: '#FFFFFF',
    }}>
      {/* Top Navigation */}
      <nav style={{
        background: '#1E2329',
        padding: '16px 32px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        borderBottom: '1px solid #2B3139',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
          <h1 style={{ color: '#F0B90B', fontSize: '24px', fontWeight: 'bold' }}>
            ARES
          </h1>
          <div style={{ display: 'flex', gap: '16px' }}>
            <a href="#" style={{ color: '#FFFFFF', textDecoration: 'none' }}>Dashboard</a>
            <a href="#" style={{ color: '#848E9C', textDecoration: 'none' }}>Trading</a>
            <a href="#" style={{ color: '#848E9C', textDecoration: 'none' }}>Analytics</a>
          </div>
        </div>
        <button
          onClick={clearToken}
          style={{
            padding: '8px 16px',
            background: 'transparent',
            border: '1px solid #2B3139',
            borderRadius: '4px',
            color: '#FFFFFF',
            cursor: 'pointer',
          }}
        >
          Logout
        </button>
      </nav>

      {/* Main Content */}
      <div style={{ padding: '32px' }}>
        <div style={{ marginBottom: '32px' }}>
          <h2 style={{ fontSize: '24px', marginBottom: '8px' }}>
            Trading Dashboard
          </h2>
          <p style={{ color: '#848E9C', fontSize: '14px' }}>
            SOLACE Autonomous Trading Engine
          </p>
        </div>

        {/* Stats Cards */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '16px',
          marginBottom: '32px',
        }}>
          <div style={{
            background: '#1E2329',
            padding: '24px',
            borderRadius: '8px',
            border: '1px solid #2B3139',
          }}>
            <p style={{ color: '#848E9C', fontSize: '14px', marginBottom: '8px' }}>
              Total Trades
            </p>
            <p style={{ fontSize: '32px', fontWeight: 'bold', color: '#F0B90B' }}>
              {trades.length}
            </p>
          </div>
          
          <div style={{
            background: '#1E2329',
            padding: '24px',
            borderRadius: '8px',
            border: '1px solid #2B3139',
          }}>
            <p style={{ color: '#848E9C', fontSize: '14px', marginBottom: '8px' }}>
              Open Positions
            </p>
            <p style={{ fontSize: '32px', fontWeight: 'bold', color: '#0ECB81' }}>
              {trades.filter(t => t.status === 'OPEN').length}
            </p>
          </div>

          <div style={{
            background: '#1E2329',
            padding: '24px',
            borderRadius: '8px',
            border: '1px solid #2B3139',
          }}>
            <p style={{ color: '#848E9C', fontSize: '14px', marginBottom: '8px' }}>
              Win Rate
            </p>
            <p style={{ fontSize: '32px', fontWeight: 'bold' }}>
              {trades.length > 0 ? '—' : '0%'}
            </p>
          </div>

          <div style={{
            background: '#1E2329',
            padding: '24px',
            borderRadius: '8px',
            border: '1px solid #2B3139',
          }}>
            <p style={{ color: '#848E9C', fontSize: '14px', marginBottom: '8px' }}>
              Total P&L
            </p>
            <p style={{ fontSize: '32px', fontWeight: 'bold' }}>
              $—
            </p>
          </div>
        </div>

        {/* Trades Table */}
        <div style={{
          background: '#1E2329',
          borderRadius: '8px',
          border: '1px solid #2B3139',
          overflow: 'hidden',
        }}>
          <div style={{ padding: '24px', borderBottom: '1px solid #2B3139' }}>
            <h3 style={{ fontSize: '18px' }}>Recent Trades</h3>
            <p style={{ color: '#848E9C', fontSize: '14px', marginTop: '4px' }}>
              SOLACE autonomous trade execution history
            </p>
          </div>

          {loading && (
            <div style={{ padding: '48px', textAlign: 'center', color: '#848E9C' }}>
              Loading trades...
            </div>
          )}

          {error && (
            <div style={{ padding: '48px', textAlign: 'center', color: '#F6465D' }}>
              {error}
            </div>
          )}

          {!loading && !error && trades.length === 0 && (
            <div style={{ padding: '48px', textAlign: 'center', color: '#848E9C' }}>
              No trades executed yet
            </div>
          )}

          {!loading && !error && trades.length > 0 && (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ background: '#0B0E11', color: '#848E9C', fontSize: '14px' }}>
                    <th style={{ padding: '16px', textAlign: 'left' }}>Pair</th>
                    <th style={{ padding: '16px', textAlign: 'left' }}>Direction</th>
                    <th style={{ padding: '16px', textAlign: 'right' }}>Size</th>
                    <th style={{ padding: '16px', textAlign: 'right' }}>Entry</th>
                    <th style={{ padding: '16px', textAlign: 'right' }}>Exit</th>
                    <th style={{ padding: '16px', textAlign: 'left' }}>Status</th>
                    <th style={{ padding: '16px', textAlign: 'left' }}>Reasoning</th>
                  </tr>
                </thead>
                <tbody>
                  {trades.map((trade) => (
                    <tr key={trade.id} style={{ borderTop: '1px solid #2B3139' }}>
                      <td style={{ padding: '16px', fontWeight: 'bold' }}>
                        {trade.trading_pair}
                      </td>
                      <td style={{ padding: '16px' }}>
                        <span style={{
                          color: trade.direction === 'BUY' ? '#0ECB81' : '#F6465D',
                          fontWeight: 'bold',
                        }}>
                          {trade.direction}
                        </span>
                      </td>
                      <td style={{ padding: '16px', textAlign: 'right' }}>
                        {trade.size.toFixed(4)}
                      </td>
                      <td style={{ padding: '16px', textAlign: 'right' }}>
                        ${trade.entry_price.toLocaleString()}
                      </td>
                      <td style={{ padding: '16px', textAlign: 'right' }}>
                        {trade.exit_price ? `$${trade.exit_price.toLocaleString()}` : '—'}
                      </td>
                      <td style={{ padding: '16px' }}>
                        <span style={{
                          padding: '4px 8px',
                          borderRadius: '4px',
                          fontSize: '12px',
                          background: trade.status === 'OPEN' ? '#0ECB8120' : '#848E9C20',
                          color: trade.status === 'OPEN' ? '#0ECB81' : '#848E9C',
                        }}>
                          {trade.status}
                        </span>
                      </td>
                      <td style={{ padding: '16px', color: '#848E9C', fontSize: '14px' }}>
                        {trade.reasoning || 'No reasoning provided'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
