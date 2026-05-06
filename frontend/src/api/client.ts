import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器：添加认证头
api.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.accessToken) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：处理错误和令牌刷新
api.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data.code !== undefined && data.code !== 0) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return response
  },
  async (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      // 尝试刷新令牌
      if (authStore.refreshToken && !error.config._retry) {
        error.config._retry = true
        try {
          await authStore.refresh()
          error.config.headers.Authorization = `Bearer ${authStore.accessToken}`
          return api(error.config)
        } catch {
          authStore.logout()
          router.push({ name: 'login' })
        }
      } else {
        authStore.logout()
        router.push({ name: 'login' })
      }
    }
    return Promise.reject(error)
  }
)

export default api
