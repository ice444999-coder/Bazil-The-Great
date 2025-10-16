import { useState } from 'react'
import { authApi } from '../lib/api'
import { useAuthStore } from '../stores/authStore'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setToken } = useAuthStore()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const token = await authApi.login(username, password)
      setToken(token)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      background: '#0B0E11',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    }}>
      <div style={{
        background: '#1E2329',
        padding: '48px',
        borderRadius: '8px',
        width: '400px',
        maxWidth: '90%',
      }}>
        <h1 style={{
          color: '#F0B90B',
          fontSize: '32px',
          fontWeight: 'bold',
          marginBottom: '8px',
          textAlign: 'center',
        }}>
          ARES
        </h1>
        <p style={{
          color: '#848E9C',
          fontSize: '14px',
          marginBottom: '32px',
          textAlign: 'center',
        }}>
          Autonomous Trading System
        </p>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: '24px' }}>
            <label style={{
              display: 'block',
              color: '#FFFFFF',
              fontSize: '14px',
              marginBottom: '8px',
            }}>
              Username
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              style={{
                width: '100%',
                padding: '12px',
                background: '#0B0E11',
                border: '1px solid #2B3139',
                borderRadius: '4px',
                color: '#FFFFFF',
                fontSize: '14px',
              }}
              required
            />
          </div>

          <div style={{ marginBottom: '24px' }}>
            <label style={{
              display: 'block',
              color: '#FFFFFF',
              fontSize: '14px',
              marginBottom: '8px',
            }}>
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={{
                width: '100%',
                padding: '12px',
                background: '#0B0E11',
                border: '1px solid #2B3139',
                borderRadius: '4px',
                color: '#FFFFFF',
                fontSize: '14px',
              }}
              required
            />
          </div>

          {error && (
            <div style={{
              padding: '12px',
              background: '#F6465D20',
              border: '1px solid #F6465D',
              color: '#F6465D',
              borderRadius: '4px',
              fontSize: '14px',
              marginBottom: '24px',
            }}>
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%',
              padding: '14px',
              background: loading ? '#6B7280' : '#F0B90B',
              color: '#0B0E11',
              border: 'none',
              borderRadius: '4px',
              fontSize: '16px',
              fontWeight: 'bold',
              cursor: loading ? 'not-allowed' : 'pointer',
            }}
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <p style={{
          color: '#848E9C',
          fontSize: '12px',
          marginTop: '24px',
          textAlign: 'center',
        }}>
          SOLACE Δ3-2 Autonomous Agent
        </p>
      </div>
    </div>
  )
}
