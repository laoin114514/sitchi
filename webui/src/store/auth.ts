import { defineStore } from 'pinia'

interface AuthState {
  accessToken: string
  refreshToken: string
  userId: number | null
  moduleCode: string
  roles: string[]
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    accessToken: localStorage.getItem('access_token') ?? '',
    refreshToken: localStorage.getItem('refresh_token') ?? '',
    userId: Number(localStorage.getItem('user_id') || 0) || null,
    moduleCode: localStorage.getItem('module_code') ?? '',
    roles: JSON.parse(localStorage.getItem('roles') || '[]'),
  }),
  actions: {
    setAuth(payload: {
      accessToken: string
      refreshToken: string
      userId: number
      moduleCode: string
      roles: string[]
    }) {
      this.accessToken = payload.accessToken
      this.refreshToken = payload.refreshToken
      this.userId = payload.userId
      this.moduleCode = payload.moduleCode
      this.roles = payload.roles

      localStorage.setItem('access_token', payload.accessToken)
      localStorage.setItem('refresh_token', payload.refreshToken)
      localStorage.setItem('user_id', String(payload.userId))
      localStorage.setItem('module_code', payload.moduleCode)
      localStorage.setItem('roles', JSON.stringify(payload.roles))
    },
    clearAuth() {
      this.accessToken = ''
      this.refreshToken = ''
      this.userId = null
      this.moduleCode = ''
      this.roles = []

      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user_id')
      localStorage.removeItem('module_code')
      localStorage.removeItem('roles')
    },
  },
})
