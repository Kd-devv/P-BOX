import { 
  Zap, Globe, Rocket, Target, Bot, Tv, Film, Cloud, Plane, Music,
  MessageCircle, Search, Twitter, Gamepad2, Apple, Github, Ban, 
  Fish, Flag, Shield, Server, Wifi, Radio, type LucideIcon
} from 'lucide-react'

// 关键词 -> 图标映射 (用于智能匹配)
const iconKeywords: [string[], LucideIcon][] = [
  [['自动', 'auto', 'url-test', 'urltest'], Zap],
  [['故障', 'fallback', '备用'], Shield],
  [['节点选择', 'select', 'proxy', '手动'], Rocket],
  [['直连', 'direct', '国内'], Target],
  [['ai', 'gpt', 'openai', 'claude', 'gemini', 'copilot'], Bot],
  [['媒体', 'media', 'stream', '流媒体'], Globe],
  [['netflix', 'nf', '奈飞'], Film],
  [['disney', '迪士尼', 'hbo', 'youtube', 'ytb'], Tv],
  [['telegram', 'tg', '电报'], MessageCircle],
  [['google', '谷歌', 'gcp'], Search],
  [['twitter', 'x', '推特'], Twitter],
  [['facebook', 'fb', '脸书', 'meta', 'instagram'], Globe],
  [['game', '游戏', 'steam', 'playstation', 'xbox', 'switch'], Gamepad2],
  [['bilibili', 'bili', '哔哩', 'b站'], Tv],
  [['microsoft', '微软', 'azure', 'bing', 'office'], Cloud],
  [['apple', '苹果', 'icloud'], Apple],
  [['github', 'gitlab', 'dev'], Github],
  [['广告', 'ad', 'block', '拦截', 'reject'], Ban],
  [['漏网', 'final', '兜底', '其他'], Fish],
  [['香港', 'hk', 'hong'], Flag],
  [['台湾', 'tw', 'taiwan'], Flag],
  [['日本', 'jp', 'japan'], Flag],
  [['新加坡', 'sg', 'singapore'], Flag],
  [['美国', 'us', 'usa', 'america'], Flag],
  [['韩国', 'kr', 'korea'], Flag],
  [['英国', 'uk', 'britain'], Flag],
  [['德国', 'de', 'germany'], Flag],
  [['加拿大', 'ca', 'canada'], Flag],
  [['澳大利亚', 'au', 'australia'], Flag],
  [['emby', 'plex', 'jellyfin'], Server],
  [['spotify', '音乐', 'music'], Music],
  [['tiktok', '抖音', 'douyin'], Radio],
  [['speedtest', '测速'], Wifi],
  [['机场', 'airport', '订阅'], Plane],
]

// 关键词 -> 颜色映射
const colorKeywords: [string[], string][] = [
  [['自动', 'auto', 'url-test'], 'bg-amber-500'],
  [['故障', 'fallback'], 'bg-rose-500'],
  [['节点选择', 'select', 'proxy'], 'bg-blue-500'],
  [['直连', 'direct'], 'bg-emerald-500'],
  [['ai', 'gpt', 'openai', 'claude'], 'bg-purple-500'],
  [['媒体', 'media', 'stream'], 'bg-pink-500'],
  [['netflix', 'nf'], 'bg-red-600'],
  [['disney', 'hbo'], 'bg-blue-600'],
  [['telegram', 'tg', '电报'], 'bg-sky-500'],
  [['google', '谷歌'], 'bg-blue-600'],
  [['twitter', '推特'], 'bg-sky-400'],
  [['facebook', 'meta'], 'bg-indigo-600'],
  [['game', '游戏'], 'bg-violet-500'],
  [['bilibili', 'bili', '哔哩'], 'bg-pink-400'],
  [['microsoft', '微软'], 'bg-cyan-600'],
  [['apple', '苹果'], 'bg-slate-700'],
  [['github'], 'bg-neutral-800'],
  [['广告', 'ad', 'block', 'reject'], 'bg-red-500'],
  [['漏网', 'final'], 'bg-teal-500'],
  [['香港', 'hk'], 'bg-rose-600'],
  [['台湾', 'tw'], 'bg-blue-700'],
  [['日本', 'jp'], 'bg-red-700'],
  [['新加坡', 'sg'], 'bg-red-500'],
  [['美国', 'us'], 'bg-blue-800'],
  [['韩国', 'kr'], 'bg-indigo-500'],
  [['英国', 'uk'], 'bg-blue-700'],
  [['手动', 'manual'], 'bg-orange-500'],
]

// 不显示在仪表盘的分组类型/关键词
export const hiddenGroupTypes = ['DIRECT', 'REJECT', 'Direct', 'Reject', '全球直连', '广告拦截']

// 智能匹配图标
export function getGroupIcon(name: string): LucideIcon {
  const lowerName = name.toLowerCase()
  for (const [keywords, icon] of iconKeywords) {
    if (keywords.some(k => lowerName.includes(k.toLowerCase()))) {
      return icon
    }
  }
  return Globe
}

// 智能匹配颜色
export function getGroupIconColor(name: string): string {
  const lowerName = name.toLowerCase()
  for (const [keywords, color] of colorKeywords) {
    if (keywords.some(k => lowerName.includes(k.toLowerCase()))) {
      return color
    }
  }
  return 'bg-slate-500'
}

// 分组排序（未知分组排后面）
export function getGroupOrder(name: string): number {
  const lowerName = name.toLowerCase()
  // 优先级排序
  const priorities = [
    ['自动', 'auto', 'url-test'],
    ['故障', 'fallback'],
    ['节点选择', 'select', 'proxy'],
    ['ai', 'gpt', 'openai'],
    ['媒体', 'media', 'netflix'],
    ['telegram', 'tg', '电报'],
    ['google', '谷歌'],
    ['game', '游戏'],
  ]
  for (let i = 0; i < priorities.length; i++) {
    if (priorities[i].some(k => lowerName.includes(k))) {
      return i
    }
  }
  return 999
}

// 国家代码转国旗 emoji
export function countryCodeToFlag(code: string): string {
  if (!code || code.length !== 2) return '🌐'
  // 台湾使用中国国旗
  const normalizedCode = code.toUpperCase() === 'TW' ? 'CN' : code.toUpperCase()
  const codePoints = normalizedCode.split('').map(char => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

// 延迟颜色
export function getDelayColorClass(delay: number): string {
  if (delay === 0) return 'bg-slate-500/20 text-slate-400'
  if (delay < 100) return 'bg-emerald-500/20 text-emerald-400'
  if (delay < 200) return 'bg-amber-500/20 text-amber-400'
  return 'bg-rose-500/20 text-rose-400'
}
