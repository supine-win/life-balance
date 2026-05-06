import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type UserInfo } from '@/api'

const TOKEN_KEY = 'lr_access_token'
const REFRESH_KEY = 'lr_refresh_token'
const USER_KEY = 'lr_user'

export const useAuthStore = defineStore('auth', () => {
  // State
  const accessToken = ref(localStorage.getItem(TOKEN_KEY) || '')
  const refreshToken = ref(localStorage.getItem(REFRESH_KEY) || '')
  const user = ref<UserInfo | null>(null)
  const loading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.roles?.includes('admin') || false)
  const username = computed(() => user.value?.nickname || user.value?.username || '')

  function hasPermission(permission: string): boolean {
    if (!user.value) return false
    if (user.value.roles?.includes('admin')) return true
    return user.value.permissions?.includes(permission) || false
  }

  // Actions
  async function login(usernameInput: string, password: string) {
    loading.value = true
    try {
      const { data } = await authApi.login({ username: usernameInput, password })
      const result = data.data || data as any
      setTokens(result.access_token, result.refresh_token)
      user.value = result.user
      localStorage.setItem(USER_KEY, JSON.stringify(result.user))
      return result
    } finally {
      loading.value = false
    }
  }

  async function register(input: { username: string; email: string; password: string; nickname?: string }) {
    loading.value = true
    try {
      const { data } = await authApi.register(input)
      const result = data.data || data as any
      setTokens(result.access_token, result.refresh_token)
      user.value = result.user
      localStorage.setItem(USER_KEY, JSON.stringify(result.user))
      return result
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    if (!refreshToken.value) throw new Error('No refresh token')
    const { data } = await authApi.refresh(refreshToken.value)
    const result = data.data || data as any
    setTokens(result.access_token, result.refresh_token)
  }

  async function fetchUser() {
    try {
      const { data } = await authApi.me()
      user.value = data.data || data as any
      localStorage.setItem(USER_KEY, JSON.stringify(user.value))
    } catch {
      logout()
    }
  }

  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem(USER_KEY)
  }

  function restoreSession() {
    accessToken.value = localStorage.getItem(TOKEN_KEY) || ''
    refreshToken.value = localStorage.getItem(REFRESH_KEY) || ''
    const userStr = localStorage.getItem(USER_KEY)
    if (userStr) {
      try {
        user.value = JSON.parse(userStr)
      } catch {
        logout()
      }
    }
  }

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    localStorage.setItem(TOKEN_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
  }

  return {
    accessToken,
    refreshToken,
    user,
    loading,
    isAuthenticated,
    isAdmin,
    username,
    hasPermission,
    login,
    register,
    refresh,
    fetchUser,
    logout,
    restoreSession,
  }
})
