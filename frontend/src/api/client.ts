import axios, { AxiosResponse, AxiosError, InternalAxiosRequestConfig } from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - add token
client.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('p-box-token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
client.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    // Check business status code
    if (data && data.code !== undefined && data.code !== 0) {
      const error = new Error(data.message || '请求失败')
      return Promise.reject(error)
    }
    // Return data.data or data
    return data?.data !== undefined ? data.data : data
  },
  (error: AxiosError) => {
    // 401 Unauthorized - redirect to login
    if (error.response?.status === 401) {
      localStorage.removeItem('p-box-token')
      localStorage.removeItem('p-box-user')
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login'
      }
      return Promise.reject(new Error('未授权，请先登录'))
    }

    // Network or server error
    let message = '网络请求失败'
    if (error.response) {
      const data = error.response.data as Record<string, unknown>
      message = (data?.error || data?.message || `服务器错误 (${error.response.status})`) as string
    } else if (error.code === 'ECONNABORTED') {
      message = '请求超时'
    } else if (!navigator.onLine) {
      message = '网络连接已断开'
    }
    return Promise.reject(new Error(message))
  }
)

// Wrapped request methods with correct types
export const api = {
  get: <T>(url: string, config?: object): Promise<T> => 
    client.get(url, config) as Promise<T>,
  post: <T>(url: string, data?: unknown, config?: object): Promise<T> => 
    client.post(url, data, config) as Promise<T>,
  put: <T>(url: string, data?: unknown, config?: object): Promise<T> => 
    client.put(url, data, config) as Promise<T>,
  delete: <T>(url: string, config?: object): Promise<T> => 
    client.delete(url, config) as Promise<T>,
}

export default api
