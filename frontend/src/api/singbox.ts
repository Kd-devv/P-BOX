import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// Sing-Box 配置生成选项
export interface SingBoxGenerateOptions {
  mode: 'tun' | 'system'       // TUN 模式 或 系统代理模式
  fakeip: boolean              // 启用 FakeIP
  mixedPort: number            // 混合代理端口
  httpPort?: number            // HTTP 代理端口
  socksPort?: number           // SOCKS5 代理端口
  clashApiAddr: string         // Clash API 地址
  clashApiSecret?: string      // Clash API 密钥
  tunStack: 'system' | 'gvisor' | 'mixed'  // TUN 栈类型
  tunMtu?: number              // TUN MTU
  dnsStrategy: 'prefer_ipv4' | 'prefer_ipv6' | 'ipv4_only' | 'ipv6_only'
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal' | 'panic'
}

// 生成结果
export interface SingBoxGenerateResult {
  configPath: string
  nodeCount: number
  mode: string
  validationError?: string  // 配置验证错误信息
}

// Sing-Box 设置 (持久化保存)
export interface SingBoxSettings {
  mode: 'tun' | 'system'
  // 端口设置
  mixedPort: number
  httpPort: number
  socksPort: number
  // Clash API
  clashApiAddr: string
  clashApiSecret: string
  // TUN 设置
  tunStack: 'system' | 'gvisor' | 'mixed'
  tunMtu: number
  autoRedirect: boolean      // Linux nftables 性能优化
  strictRoute: boolean       // 严格路由
  // DNS 设置
  fakeip: boolean
  dnsStrategy: 'prefer_ipv4' | 'prefer_ipv6' | 'ipv4_only' | 'ipv6_only'
  // 性能优化
  tcpFastOpen: boolean
  tcpMultiPath: boolean
  udpFragment: boolean
  sniff: boolean
  sniffOverrideDestination: boolean
  // 日志
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal' | 'panic'
}

// TUN 模式默认最优配置
export const defaultTunSettings: Partial<SingBoxSettings> = {
  mode: 'tun',
  mixedPort: 7890,
  tunStack: 'system',        // 性能最好
  tunMtu: 9000,
  autoRedirect: true,        // Linux 性能优化
  strictRoute: true,         // 严格路由
  fakeip: true,              // TUN 模式推荐 FakeIP
  dnsStrategy: 'prefer_ipv4',
  tcpFastOpen: true,
  tcpMultiPath: false,       // 需要 Go 1.21+，谨慎开启
  udpFragment: true,
  sniff: true,
  sniffOverrideDestination: true,
  logLevel: 'info',
}

// 系统代理模式默认最优配置
export const defaultSystemSettings: Partial<SingBoxSettings> = {
  mode: 'system',
  mixedPort: 7890,
  httpPort: 7891,
  socksPort: 7892,
  fakeip: false,             // 系统代理不推荐 FakeIP
  dnsStrategy: 'prefer_ipv4',
  tcpFastOpen: true,
  tcpMultiPath: false,
  udpFragment: true,
  sniff: true,
  sniffOverrideDestination: true,
  logLevel: 'info',
}

// 检测平台
export const detectPlatform = (): 'macos' | 'windows' | 'linux' | 'unknown' => {
  const userAgent = navigator.userAgent.toLowerCase()
  if (userAgent.includes('mac')) return 'macos'
  if (userAgent.includes('win')) return 'windows'
  if (userAgent.includes('linux')) return 'linux'
  return 'unknown'
}

// 根据平台获取默认模式
export const getDefaultMode = (): 'tun' | 'system' => {
  const platform = detectPlatform()
  // macOS 默认使用系统代理模式
  if (platform === 'macos') return 'system'
  // Linux/Windows 默认使用 TUN 模式
  return 'tun'
}

// 获取平台对应的默认设置
export const getPlatformDefaultSettings = (): SingBoxSettings => {
  const mode = getDefaultMode()
  const baseSettings: SingBoxSettings = {
    mode,
    mixedPort: 7890,
    httpPort: 7891,
    socksPort: 7892,
    clashApiAddr: '127.0.0.1:9090',
    clashApiSecret: '',
    tunStack: 'system',
    tunMtu: 9000,
    autoRedirect: true,
    strictRoute: true,
    fakeip: mode === 'tun', // TUN 模式推荐 FakeIP
    dnsStrategy: 'prefer_ipv4',
    tcpFastOpen: true,
    tcpMultiPath: false,
    udpFragment: true,
    sniff: true,
    sniffOverrideDestination: true,
    logLevel: 'info',
  }
  return baseSettings
}

// 默认设置 (根据平台自动选择)
export const defaultSingBoxSettings: SingBoxSettings = getPlatformDefaultSettings()

// 生成配置响应
export interface SingBoxGenerateResponse {
  code: number
  message: string
  data: SingBoxGenerateResult
}

export const singboxApi = {
  // 生成 Sing-Box 配置
  generateConfig: async (options: Partial<SingBoxGenerateOptions>): Promise<SingBoxGenerateResponse> => {
    const res = await client.post('/proxy/singbox/generate', options)
    return res.data
  },

  // 获取配置预览
  getConfigPreview: async (): Promise<string> => {
    const res = await client.get('/proxy/singbox/preview')
    return res.data.data.content
  },

  // 获取配置下载 URL
  getDownloadUrl: (): string => {
    return '/api/proxy/singbox/download'
  },

  // 保存设置到本地存储
  saveSettings: (settings: SingBoxSettings): void => {
    localStorage.setItem('singbox-settings', JSON.stringify(settings))
  },

  // 从本地存储加载设置
  loadSettings: (): SingBoxSettings => {
    const saved = localStorage.getItem('singbox-settings')
    if (saved) {
      try {
        return { ...defaultSingBoxSettings, ...JSON.parse(saved) }
      } catch {
        return defaultSingBoxSettings
      }
    }
    return defaultSingBoxSettings
  },
}
