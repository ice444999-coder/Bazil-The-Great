import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - add token to all requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('ares_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor - handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('ares_token')
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

export default api

// API functions
export const authApi = {
  login: async (username: string, password: string) => {
    const response = await api.post('/users/login', { username, password })
    return response.data.access_token
  },
}

export const tradingApi = {
  getHistory: async () => {
    const response = await api.get('/trading/history')
    return Array.isArray(response.data) ? response.data : []
  },
  getOpen: async () => {
    const response = await api.get('/trading/open')
    return response.data
  },
  getPerformance: async () => {
    const response = await api.get('/trading/performance')
    return response.data
  },
}

export const solaceApi = {
  getMemoryLog: async () => {
    const response = await api.get('/masterplan/memory/logs')
    return response.data
  },
  getGlassBox: async () => {
    const response = await api.get('/glass-box/logs')
    return response.data
  },
}
