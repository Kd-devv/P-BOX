import api from './client'

export interface SystemConfig {
  autoStart: boolean
  ipForward: boolean
  bbrEnabled: boolean
  tunOptimized: boolean
}

export interface SystemInfo {
  name: string
  version: string
  buildTime: string
}

export interface SystemResources {
  os: string
  platform: string
  kernel: string
  arch: string
  cpuModel: string
  cpuCores: number
  cpuUsage: number
  memoryTotal: number
  memoryUsed: number
  memoryPercent: number
  diskTotal: number
  diskUsed: number
  diskPercent: number
  uptime: number
}

export interface GeoIPInfo {
  ip: string
  country: string
  countryCode: string
  region: string
  city: string
  isp: string
  org: string
}

export const systemApi = {
  // Get system info (version etc)
  getInfo: () => api.get<SystemInfo>('/system/info'),
  
  // Get system resources (CPU, memory, disk)
  getResources: () => api.get<SystemResources>('/system/resources'),
  
  // Get system config
  getConfig: () => api.get<SystemConfig>('/system/config'),
  
  // Set auto start
  setAutoStart: (enabled: boolean) => api.put('/system/autostart', { enabled }),
  
  // Set IP forward
  setIPForward: (enabled: boolean) => api.put('/system/ipforward', { enabled }),
  
  // Set BBR
  setBBR: (enabled: boolean) => api.put('/system/bbr', { enabled }),
  
  // Set TUN optimize
  setTUNOptimize: (enabled: boolean) => api.put('/system/tunoptimize', { enabled }),
  
  // Optimize all
  optimizeAll: () => api.post('/system/optimize-all', {}),
  
  // 系统代理
  enableSystemProxy: (host: string, port: number) => api.post('/system/proxy/enable', { host, port }),
  disableSystemProxy: () => api.post('/system/proxy/disable', {}),
  getSystemProxyStatus: () => api.get<{ enabled: boolean; host: string; port: number }>('/system/proxy/status'),
  
  // 出口 IP 信息
  getGeoIP: (lang: string = 'zh') => api.get<GeoIPInfo>(`/system/geoip?lang=${lang}`),
}
