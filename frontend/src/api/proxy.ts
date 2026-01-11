import api from './client'

// Transparent proxy mode types
export type TransparentMode = 'off' | 'tun' | 'tproxy' | 'redirect'

export interface ProxyStatus {
  running: boolean
  coreType: string
  coreVersion: string
  mode: 'rule' | 'global' | 'direct'
  mixedPort: number
  socksPort: number
  allowLan: boolean
  tunEnabled: boolean
  transparentMode: TransparentMode
  uptime: number
}

export interface ProxyConfig {
  mixedPort: number
  socksPort: number
  redirPort: number
  tproxyPort: number
  allowLan: boolean
  ipv6: boolean
  mode: string
  logLevel: string
  tunEnabled: boolean
  tunStack: string
  transparentMode: TransparentMode
  autoStart: boolean
  autoStartDelay: number
}

export const proxyApi = {
  getStatus: () => api.get<ProxyStatus>('/proxy/status'),
  start: () => api.post('/proxy/start'),
  stop: () => api.post('/proxy/stop'),
  restart: () => api.post('/proxy/restart'),
  setMode: (mode: string) => api.put('/proxy/mode', { mode }),
  setTunMode: (enabled: boolean) => api.put('/proxy/tun', { enabled }),
  setTransparentMode: (mode: TransparentMode) => api.put('/proxy/transparent', { mode }),
  getConfig: () => api.get<ProxyConfig>('/proxy/config'),
  updateConfig: (config: ProxyConfig) => api.put('/proxy/config', config),
}
