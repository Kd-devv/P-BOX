import api from './client'

export interface AuthStatus {
  enabled: boolean
  authenticated: boolean
}

export interface AuthConfig {
  enabled: boolean
  username: string
}

export interface LoginRequest {
  username?: string
  password: string
}

export interface LoginResponse {
  token: string
}

export const authApi = {
  // Check auth status
  check: () => api.get<AuthStatus>('/auth/check'),
  
  // Get auth config
  getConfig: () => api.get<AuthConfig>('/auth/config'),
  
  // Set auth enabled
  setEnabled: (enabled: boolean) => api.put('/auth/enabled', { enabled }),
  
  // Update username
  updateUsername: (username: string) => api.put('/auth/username', { username }),
  
  // Login
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  
  // Logout
  logout: () => api.post('/auth/logout'),
  
  // Change password
  changePassword: (oldPassword: string, newPassword: string) => 
    api.put('/auth/password', { oldPassword, newPassword }),
}

// Clear auth token
export function clearAuth() {
  localStorage.removeItem('token')
}
