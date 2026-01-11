// Mihomo API client - communicates with core API through backend proxy

// Use backend proxy (avoid CORS issues)
const getProxyApiBase = () => '/api/proxy/mihomo'

// Direct access to Mihomo API (only for WebSocket)
const getDirectApiBase = () => {
  const host = window.location.hostname || '127.0.0.1'
  return `http://${host}:9090`
}

export interface ProxyNode {
  name: string
  type: string
  now?: string
  all?: string[]
  history?: { delay: number }[]
}

export interface ProxyGroup {
  name: string
  type: string
  now?: string
  all?: string[]
}

export interface MihomoConfig {
  mode: string
  'mixed-port': number
  'allow-lan': boolean
}

export const mihomoApi = {
  // Check if core is running (via backend API)
  async isRunning(): Promise<boolean> {
    try {
      const res = await fetch(`${getProxyApiBase()}/proxies`, { 
        signal: AbortSignal.timeout(2000) 
      })
      return res.ok
    } catch {
      return false
    }
  },

  // Get version (via backend)
  async getVersion(): Promise<string> {
    const res = await fetch(`${getDirectApiBase()}/version`)
    const data = await res.json()
    return data.version
  },

  // Get configs
  async getConfigs(): Promise<MihomoConfig> {
    const res = await fetch(`${getDirectApiBase()}/configs`)
    return res.json()
  },

  // Update configs
  async patchConfigs(config: Partial<MihomoConfig>): Promise<void> {
    await fetch(`${getDirectApiBase()}/configs`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    })
  },

  // Get all proxies (via backend proxy)
  async getProxies(): Promise<Record<string, ProxyNode>> {
    const res = await fetch(`${getProxyApiBase()}/proxies`)
    const data = await res.json()
    return data.proxies
  },

  // Get single proxy group (via backend proxy)
  async getProxy(name: string): Promise<ProxyGroup> {
    const res = await fetch(`${getProxyApiBase()}/proxies/${encodeURIComponent(name)}`)
    return res.json()
  },

  // Select proxy node (via backend proxy)
  async selectProxy(group: string, name: string): Promise<void> {
    const res = await fetch(`${getProxyApiBase()}/proxies/${encodeURIComponent(group)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) {
      throw new Error(`切换失败: ${res.statusText}`)
    }
  },

  // Test node delay (via backend proxy)
  async testDelay(name: string, url = 'http://www.gstatic.com/generate_204', timeout = 5000): Promise<number> {
    try {
      const res = await fetch(
        `${getProxyApiBase()}/proxies/${encodeURIComponent(name)}/delay?url=${encodeURIComponent(url)}&timeout=${timeout}`
      )
      if (!res.ok) {
        return 0
      }
      const data = await res.json()
      return data.delay || 0
    } catch {
      return 0
    }
  },

  // Get connections (direct access)
  async getConnections(): Promise<{ downloadTotal: number; uploadTotal: number; connections: unknown[] }> {
    const res = await fetch(`${getDirectApiBase()}/connections`)
    return res.json()
  },

  // Close all connections
  async closeAllConnections(): Promise<void> {
    await fetch(`${getDirectApiBase()}/connections`, { method: 'DELETE' })
  },

  // 快速控制 API
  // 重载配置（重载核心）
  async reloadConfig(): Promise<void> {
    const res = await fetch(`${getDirectApiBase()}/configs?force=true`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    })
    if (!res.ok) {
      throw new Error('重载配置失败')
    }
  },

  // 刷新 DNS 缓存
  async flushDns(): Promise<void> {
    const res = await fetch(`${getDirectApiBase()}/cache/flushdns`, {
      method: 'POST',
    })
    if (!res.ok) {
      throw new Error('刷新 DNS 失败')
    }
  },

  // 更新 GeoIP/GeoSite 数据库
  async updateGeo(): Promise<void> {
    const res = await fetch(`${getDirectApiBase()}/upgrade/geo`, {
      method: 'POST',
    })
    if (!res.ok) {
      throw new Error('更新 GeoIP 失败')
    }
  },

  // Close single connection
  async closeConnection(id: string): Promise<void> {
    await fetch(`${getDirectApiBase()}/connections/${id}`, { method: 'DELETE' })
  },

  // Connections real-time update WebSocket (via backend WebSocket proxy)
  createConnectionsWs(onMessage: (data: unknown) => void): WebSocket {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const ws = new WebSocket(`${protocol}//${host}/ws/connections`)
    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        onMessage(data)
      } catch {
        // ignore parse errors
      }
    }
    return ws
  },

  // Traffic stats (via backend WebSocket proxy)
  createTrafficWs(onMessage: (data: { up: number; down: number }) => void): WebSocket {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const ws = new WebSocket(`${protocol}//${host}/ws/traffic`)
    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        onMessage(data)
      } catch {
        // ignore parse errors
      }
    }
    return ws
  },

  // Logs (via backend WebSocket proxy)
  createLogsWs(onMessage: (data: LogEntry) => void, level = 'info'): WebSocket {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const ws = new WebSocket(`${protocol}//${host}/ws/logs?level=${level}`)
    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        onMessage(data)
      } catch {
        // ignore parse errors
      }
    }
    return ws
  },
}

export interface LogEntry {
  type: 'info' | 'warning' | 'error' | 'debug'
  payload: string
}
