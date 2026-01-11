import api from './client'

// DNS 设置
export interface DNSSettings {
  enable: boolean
  listen: string
  preferH3: boolean
  cacheAlgorithm: string
  ipv6: boolean
  useHosts: boolean
  useSystemHosts: boolean
  respectRules: boolean
  enhancedMode: string
  fakeIpRange: string
  fakeIpRange6: string
  fakeIpFilterMode: string
  fakeIpFilter: string[]
  defaultNameserver: string[]
  nameserver: string[]
  fallback: string[]
  proxyServerNameserver: string[]
  directNameserver: string[]
}

// TUN 设置
export interface TUNSettings {
  enable: boolean
  device: string
  stack: string
  mtu: number
  gso: boolean
  gsoMaxSize: number
  autoRoute: boolean
  autoRedirect: boolean
  autoDetectInterface: boolean
  strictRoute: boolean
  udpTimeout: number
  dnsHijack: string[]
  endpointIndependentNat: boolean
  routeAddress: string[]
  routeExcludeAddress: string[]
  iproute2TableIndex: number
  iproute2RuleIndex: number
}

// 嗅探设置
export interface SnifferSettings {
  enable: boolean
  forceDnsMapping: boolean
  parsePureIp: boolean
  overrideDest: boolean
  sniffHttp: boolean
  sniffTls: boolean
  sniffQuic: boolean
  skipDomain: string[]
}

// 认证用户
export interface AuthUser {
  username: string
  password: string
  enabled: boolean
}

// 代理设置
export interface ProxySettings {
  // 端口设置
  mixedPortEnabled: boolean
  mixedPort: number
  socksPortEnabled: boolean
  socksPort: number
  httpPortEnabled: boolean
  httpPort: number
  redirPortEnabled: boolean
  redirPort: number
  tproxyPortEnabled: boolean
  tproxyPort: number

  // 认证设置 (启用的账号自动开启认证)
  authentication: AuthUser[]

  // 基础设置
  allowLan: boolean
  bindAddress: string
  autoStart: boolean
  autoStartDelay: number

  // 运行模式
  mode: string
  logLevel: string
  ipv6: boolean

  // 性能优化
  unifiedDelay: boolean
  tcpConcurrent: boolean
  findProcessMode: string

  // TCP Keep-Alive
  keepAliveInterval: number
  keepAliveIdle: number
  disableKeepAlive: boolean

  // TLS
  globalClientFingerprint: string
  skipCertVerify: boolean

  // GEO 数据
  geodataMode: boolean
  geodataLoader: string
  geositeMatcher: string
  geoAutoUpdate: boolean
  geoUpdateInterval: number

  // 外部资源
  globalUa: string
  etagSupport: boolean

  // 网络接口
  interfaceName: string
  routingMark: number

  // 子设置
  dns: DNSSettings
  tun: TUNSettings
  sniffer: SnifferSettings
}

// 预设
export interface SettingsPreset {
  id: string
  name: string
  description: string
  icon: string
}

// API
export const proxySettingsApi = {
  // 获取设置
  getSettings: (): Promise<ProxySettings> => {
    return api.get('/proxy/settings')
  },

  // 更新设置
  updateSettings: (settings: ProxySettings): Promise<void> => {
    return api.put('/proxy/settings', settings)
  },

  // 重置设置
  resetSettings: (): Promise<ProxySettings> => {
    return api.post('/proxy/settings/reset')
  },

  // 获取预设列表
  getPresets: (): Promise<SettingsPreset[]> => {
    return api.get('/proxy/settings/presets')
  },

  // 应用预设
  applyPreset: (presetId: string): Promise<ProxySettings> => {
    return api.post('/proxy/settings/apply-preset', { presetId })
  },
}
