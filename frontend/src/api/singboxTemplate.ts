// Sing-Box 配置模板 - 代理组和规则定义
// 参考 singforge-web 项目的规则配置

// 规则集存储路径 (实际路径从后端获取)
export const SINGBOX_RULESET_DIR = '/var/lib/p-box/singbox/ruleset'

// SagerNet 官方规则仓库
// GEO 数据库: https://github.com/SagerNet/sing-geoip, https://github.com/SagerNet/sing-geosite
// Rule-Set: https://github.com/SagerNet/sing-geosite (rule-set 分支)
export const OFFICIAL_GEOSITE_RULESET_URL = 'https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set'
export const OFFICIAL_GEOIP_RULESET_URL = 'https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set'

// Sing-Box 代理组类型
export interface SingBoxProxyGroup {
  tag: string
  type: 'selector' | 'urltest' | 'direct' | 'block'
  name: string
  description: string
  icon: string
  enabled: boolean
  outbounds: string[]
  default?: string
  // urltest 专用
  url?: string
  interval?: string
  tolerance?: number
}

// Sing-Box 规则定义
export interface SingBoxRule {
  type?: 'logical'
  mode?: 'and' | 'or'
  rules?: SingBoxRule[]
  // 匹配条件
  protocol?: string | string[]
  network?: string
  port?: number | number[]
  port_range?: string[]
  domain?: string[]
  domain_suffix?: string[]
  domain_keyword?: string[]
  domain_regex?: string[]
  ip_cidr?: string[]
  source_ip_cidr?: string[]
  rule_set?: string | string[]
  clash_mode?: string
  // 动作
  action?: string
  outbound?: string
}

// Sing-Box 规则集定义
export interface SingBoxRuleSet {
  tag: string
  type: 'local' | 'remote'
  format: 'binary' | 'source'
  path?: string
  url?: string
  download_detour?: string
  update_interval?: string
}

// 默认代理组配置 (与后端 GetSingBoxProxyGroups 保持一致)
export const defaultSingBoxProxyGroups: SingBoxProxyGroup[] = [
  // 1. 自动选择
  {
    tag: 'auto',
    type: 'urltest',
    name: '自动选择',
    description: '自动选择延迟最低的节点',
    icon: '⚡',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  // 2. 故障转移
  {
    tag: 'fallback',
    type: 'urltest',
    name: '故障转移',
    description: '按顺序检测节点可用性',
    icon: '🛡️',
    enabled: true,
    outbounds: ['HongKong', 'Taiwan', 'Japan', 'Singapore', 'America', 'Manual'],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m'
  },
  // 3. 节点选择 (主选择器)
  {
    tag: 'proxy',
    type: 'selector',
    name: '节点选择',
    description: '手动选择代理节点',
    icon: '🚀',
    enabled: true,
    outbounds: ['auto', 'fallback', 'HongKong', 'Taiwan', 'Japan', 'Singapore', 'America', 'Manual', 'Others', 'direct'],
    default: 'auto'
  },
  // 4. 全球直连
  {
    tag: 'DIRECT',
    type: 'selector',
    name: '全球直连',
    description: '直接连接不走代理',
    icon: '🎯',
    enabled: true,
    outbounds: ['direct', 'proxy']
  },
  // 5. 广告拦截
  {
    tag: 'AdBlock',
    type: 'selector',
    name: '广告拦截',
    description: '拦截广告和追踪器',
    icon: '🚫',
    enabled: true,
    outbounds: ['block', 'direct']
  },
  // 6. AI 服务
  {
    tag: 'AI',
    type: 'selector',
    name: 'AI 服务',
    description: 'ChatGPT、Claude、Gemini 等 AI 平台',
    icon: '🤖',
    enabled: true,
    outbounds: ['proxy', 'America', 'Japan', 'Singapore', 'Taiwan', 'Manual', 'auto'],
    default: 'America'
  },
  // 7. 游戏平台
  {
    tag: 'Gaming',
    type: 'selector',
    name: '游戏平台',
    description: 'Steam、Epic 等游戏服务',
    icon: '🎮',
    enabled: true,
    outbounds: ['proxy', 'HongKong', 'Taiwan', 'Japan', 'Manual', 'direct']
  },
  // 8. 国外媒体
  {
    tag: 'Streaming',
    type: 'selector',
    name: '国外媒体',
    description: 'Netflix、Disney+、YouTube、Spotify 等',
    icon: '📺',
    enabled: true,
    outbounds: ['proxy', 'HongKong', 'Taiwan', 'Japan', 'Singapore', 'America', 'Manual', 'auto']
  },
  // 9. 社交媒体
  {
    tag: 'Social',
    type: 'selector',
    name: '社交媒体',
    description: 'Telegram、Twitter、Facebook、Instagram',
    icon: '👥',
    enabled: true,
    outbounds: ['proxy', 'HongKong', 'Taiwan', 'Singapore', 'America', 'Manual', 'auto']
  },
  // 10. 海外聊天
  {
    tag: 'Chat',
    type: 'selector',
    name: '海外聊天',
    description: 'Discord、WhatsApp 等',
    icon: '💬',
    enabled: true,
    outbounds: ['proxy', 'HongKong', 'Taiwan', 'Singapore', 'America', 'Manual', 'auto']
  },
  // 11. 谷歌服务
  {
    tag: 'Google',
    type: 'selector',
    name: 'Google',
    description: 'Google 搜索、Gmail、YouTube 等',
    icon: '🔍',
    enabled: true,
    outbounds: ['proxy', 'HongKong', 'Taiwan', 'Japan', 'America', 'Manual', 'auto']
  },
  // 12. GitHub
  {
    tag: 'GitHub',
    type: 'selector',
    name: 'GitHub',
    description: 'GitHub 代码托管',
    icon: '💻',
    enabled: true,
    outbounds: ['proxy', 'direct', 'Manual', 'auto']
  },
  // 13. 微软服务
  {
    tag: 'Microsoft',
    type: 'selector',
    name: 'Microsoft',
    description: '微软服务',
    icon: '🪟',
    enabled: true,
    outbounds: ['direct', 'proxy', 'HongKong', 'America', 'Manual']
  },
  // 14. 苹果服务
  {
    tag: 'Apple',
    type: 'selector',
    name: 'Apple',
    description: 'Apple 官方服务',
    icon: '🍎',
    enabled: true,
    outbounds: ['direct', 'proxy', 'America', 'Manual']
  },
  // 15. 哔哩哔哩
  {
    tag: 'BiliBili',
    type: 'selector',
    name: '哔哩哔哩',
    description: 'B站港澳台解锁',
    icon: '📺',
    enabled: true,
    outbounds: ['direct', 'HongKong', 'Taiwan', 'Manual']
  },
  // 16. 漏网之鱼
  {
    tag: 'Final',
    type: 'selector',
    name: '漏网之鱼',
    description: '未匹配规则的流量',
    icon: '🌐',
    enabled: true,
    outbounds: ['proxy', 'auto', 'Manual', 'direct']
  },
  // === 地区节点分组 ===
  {
    tag: 'HongKong',
    type: 'urltest',
    name: '香港节点',
    description: '香港节点自动选择',
    icon: '��',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  {
    tag: 'Taiwan',
    type: 'urltest',
    name: '台湾节点',
    description: '台湾节点自动选择',
    icon: '🇹🇼',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  {
    tag: 'Japan',
    type: 'urltest',
    name: '日本节点',
    description: '日本节点自动选择',
    icon: '🇯🇵',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  {
    tag: 'Singapore',
    type: 'urltest',
    name: '新加坡节点',
    description: '新加坡节点自动选择',
    icon: '�🇬',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  {
    tag: 'America',
    type: 'urltest',
    name: '美国节点',
    description: '美国节点自动选择',
    icon: '🇺🇸',
    enabled: true,
    outbounds: [],
    url: 'https://www.gstatic.com/generate_204',
    interval: '5m',
    tolerance: 50
  },
  {
    tag: 'Manual',
    type: 'selector',
    name: '手动节点',
    description: '手动添加的节点',
    icon: '✋',
    enabled: true,
    outbounds: []
  },
  {
    tag: 'Others',
    type: 'selector',
    name: '其他节点',
    description: '其他地区节点',
    icon: '🌍',
    enabled: true,
    outbounds: []
  }
]

// 默认规则集配置 (使用 SagerNet 官方规则仓库)
export const defaultSingBoxRuleSets: SingBoxRuleSet[] = [
  // ==================== 广告拦截 ====================
  {
    tag: 'geosite-category-ads-all',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-category-ads-all.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/geosite-category-ads-all.srs`
  },
  // ==================== AI 服务 ====================
  {
    tag: 'geosite-openai',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-openai.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/openai.srs`
  },
  {
    tag: 'geosite-anthropic',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-anthropic.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/anthropic.srs`
  },
  {
    tag: 'geosite-google-gemini',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-google-gemini.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/google-gemini.srs`
  },
  {
    tag: 'geosite-cursor',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-cursor.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/cursor.srs`
  },
  {
    tag: 'geosite-category-ai-!cn',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-category-ai-!cn.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/category-ai-!cn.srs`
  },
  // ==================== 游戏平台 ====================
  {
    tag: 'geosite-steam',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-steam.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/steam.srs`
  },
  {
    tag: 'geosite-epicgames',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-epicgames.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/epicgames.srs`
  },
  // ==================== 流媒体 ====================
  {
    tag: 'geosite-netflix',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-netflix.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/netflix.srs`
  },
  {
    tag: 'geosite-disney',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-disney.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/disney.srs`
  },
  {
    tag: 'geosite-youtube',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-youtube.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/youtube.srs`
  },
  {
    tag: 'geosite-spotify',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-spotify.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/spotify.srs`
  },
  // ==================== 社交媒体 ====================
  {
    tag: 'geosite-twitter',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-twitter.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/twitter.srs`
  },
  {
    tag: 'geosite-facebook',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-facebook.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/facebook.srs`
  },
  {
    tag: 'geosite-instagram',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-instagram.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/instagram.srs`
  },
  // ==================== 海外聊天 ====================
  {
    tag: 'geosite-telegram',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-telegram.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/telegram.srs`
  },
  {
    tag: 'geosite-whatsapp',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-whatsapp.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/whatsapp.srs`
  },
  {
    tag: 'geosite-discord',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-discord.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/discord.srs`
  },
  // ==================== Google ====================
  {
    tag: 'geosite-google',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-google.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/google.srs`
  },
  // ==================== 开发者 ====================
  {
    tag: 'geosite-github',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-github.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/github.srs`
  },
  // ==================== Microsoft ====================
  {
    tag: 'geosite-microsoft',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-microsoft.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/microsoft.srs`
  },
  // ==================== Apple ====================
  {
    tag: 'geosite-apple',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-apple.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/apple.srs`
  },
  {
    tag: 'geosite-apple-cn',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-apple-cn.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/apple-cn.srs`
  },
  // ==================== 中国直连 ====================
  {
    tag: 'geosite-bilibili',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-bilibili.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/bilibili.srs`
  },
  {
    tag: 'geosite-iqiyi',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-iqiyi.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/iqiyi.srs`
  },
  {
    tag: 'geosite-alibaba',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-alibaba.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/alibaba.srs`
  },
  {
    tag: 'geosite-cn',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-cn.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/cn.srs`
  },
  {
    tag: 'geoip-cn',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geoip-cn.srs`,
    url: `${OFFICIAL_GEOIP_RULESET_URL}/geoip-cn.srs`
  },
  // ==================== 其他海外 ====================
  {
    tag: 'geosite-geolocation-!cn',
    type: 'local',
    format: 'binary',
    path: `${SINGBOX_RULESET_DIR}/geosite-geolocation-!cn.srs`,
    url: `${OFFICIAL_GEOSITE_RULESET_URL}/geolocation-!cn.srs`
  }
]

// 默认路由规则 (参考 singforge-web，按优先级排序)
export const defaultSingBoxRules: SingBoxRule[] = [
  // 1. DNS 劫持
  {
    type: 'logical',
    mode: 'or',
    rules: [
      { port: [53] },
      { protocol: ['dns'] }
    ],
    action: 'hijack-dns'
  },
  // 2. 广告拦截
  {
    rule_set: ['geosite-category-ads-all'],
    outbound: 'ad-block'
  },
  // 3. AI 服务 (ChatGPT、Claude、Gemini、Cursor)
  {
    rule_set: ['geosite-openai', 'geosite-anthropic', 'geosite-google-gemini', 'geosite-cursor', 'geosite-category-ai-!cn'],
    outbound: 'ai-proxy'
  },
  // 4. 游戏平台 (Steam、Epic)
  {
    rule_set: ['geosite-steam', 'geosite-epicgames'],
    outbound: 'game-proxy'
  },
  // 5. 流媒体 (Netflix、Disney+、YouTube、Spotify)
  {
    rule_set: ['geosite-netflix', 'geosite-disney', 'geosite-youtube', 'geosite-spotify'],
    outbound: 'media-proxy'
  },
  // 6. 社交媒体 (Twitter、Facebook、Instagram)
  {
    rule_set: ['geosite-twitter', 'geosite-facebook', 'geosite-instagram'],
    outbound: 'social-proxy'
  },
  // 7. 海外聊天 (Telegram、WhatsApp、Discord)
  {
    rule_set: ['geosite-telegram', 'geosite-whatsapp', 'geosite-discord'],
    outbound: 'chat-proxy'
  },
  // 8. Google
  {
    rule_set: ['geosite-google'],
    outbound: 'google-proxy'
  },
  // 9. 开发者 (GitHub)
  {
    rule_set: ['geosite-github'],
    outbound: 'dev-proxy'
  },
  // 10. Microsoft
  {
    rule_set: ['geosite-microsoft'],
    outbound: 'microsoft-proxy'
  },
  // 11. Apple
  {
    rule_set: ['geosite-apple', 'geosite-apple-cn'],
    outbound: 'apple-proxy'
  },
  // 12. 中国直连 (B站、爱奇艺、阿里巴巴、国内域名/IP)
  {
    rule_set: ['geosite-bilibili', 'geosite-iqiyi', 'geosite-alibaba', 'geosite-cn', 'geoip-cn'],
    outbound: 'cn-direct'
  },
  // 13. 私有地址直连
  {
    ip_cidr: ['10.0.0.0/8', '172.16.0.0/12', '192.168.0.0/16', '127.0.0.0/8', '::1/128', 'fc00::/7'],
    outbound: 'direct'
  },
  // 14. 其他海外
  {
    rule_set: ['geosite-geolocation-!cn'],
    outbound: 'overseas-proxy'
  }
]

// Sing-Box 模板配置
export interface SingBoxTemplate {
  proxyGroups: SingBoxProxyGroup[]
  rules: SingBoxRule[]
  ruleSets: SingBoxRuleSet[]
}

// 默认模板
export const defaultSingBoxTemplate: SingBoxTemplate = {
  proxyGroups: defaultSingBoxProxyGroups,
  rules: defaultSingBoxRules,
  ruleSets: defaultSingBoxRuleSets
}

// API 基础路径
const API_BASE = '/api/proxy'

// 从后端加载模板
export const loadSingBoxTemplate = async (): Promise<SingBoxTemplate> => {
  try {
    const res = await fetch(`${API_BASE}/singbox/template`)
    const data = await res.json()
    if (data.code === 0 && data.data) {
      return data.data as SingBoxTemplate
    }
  } catch (e) {
    console.error('加载 Sing-Box 模板失败:', e)
  }
  return defaultSingBoxTemplate
}

// 保存模板到后端
export const saveSingBoxTemplate = async (template: SingBoxTemplate): Promise<boolean> => {
  try {
    const res = await fetch(`${API_BASE}/singbox/template`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(template)
    })
    const data = await res.json()
    return data.code === 0
  } catch (e) {
    console.error('保存 Sing-Box 模板失败:', e)
    return false
  }
}

// 重置模板为默认值
export const resetSingBoxTemplate = async (): Promise<SingBoxTemplate> => {
  try {
    const res = await fetch(`${API_BASE}/singbox/template/reset`, { method: 'POST' })
    const data = await res.json()
    if (data.code === 0 && data.data) {
      return data.data as SingBoxTemplate
    }
  } catch (e) {
    console.error('重置 Sing-Box 模板失败:', e)
  }
  return defaultSingBoxTemplate
}
