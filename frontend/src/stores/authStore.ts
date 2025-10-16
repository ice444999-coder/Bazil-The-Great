import { create } from 'zustand'

interface AuthState {
  token: string | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  clearToken: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('ares_token'),
  isAuthenticated: !!localStorage.getItem('ares_token'),
  setToken: (token: string) => {
    localStorage.setItem('ares_token', token)
    set({ token, isAuthenticated: true })
  },
  clearToken: () => {
    localStorage.removeItem('ares_token')
    set({ token: null, isAuthenticated: false })
  },
}))
