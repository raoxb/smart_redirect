import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/api'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  
  setToken: (token: string) => void
  setUser: (user: User) => void
  logout: () => void
  checkAuth: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,

      setToken: (token: string) => {
        localStorage.setItem('token', token)
        set({ token, isAuthenticated: true })
      },

      setUser: (user: User) => {
        set({ user })
      },

      logout: () => {
        localStorage.removeItem('token')
        set({ 
          token: null, 
          user: null, 
          isAuthenticated: false 
        })
      },

      checkAuth: () => {
        const token = localStorage.getItem('token')
        if (token) {
          set({ token, isAuthenticated: true })
          return true
        }
        return false
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated
      }),
    }
  )
)