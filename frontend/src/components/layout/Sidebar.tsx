import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useSidebar } from './Layout'
import { useThemeStore } from '@/stores/themeStore'
import { useProxyStore } from '@/stores/proxyStore'
import { useCoreStore } from '@/stores/coreStore'
import { systemApi } from '@/api/system'
import { coreApi } from '@/api/core'
import {
  LayoutDashboard,
  Globe,
  ListTree,
  Link2,
  FileCode,
  Settings,
  ArrowLeftRight,
  FileText,
  Database,
  X,
  Cpu,
  Github,
  SlidersHorizontal,
  Network,
  Send,
} from 'lucide-react'

// 根据核心类型动态获取主导航项
const getMainNavItems = (activeCore: 'mihomo' | 'singbox') => [
  { path: '/', icon: LayoutDashboard, labelKey: 'nav.dashboard', color: 'blue' },
  { path: '/proxy-switch', icon: ArrowLeftRight, labelKey: 'nav.proxySwitch', color: 'purple' },
  { path: '/nodes', icon: Globe, labelKey: 'nav.nodes', color: 'cyan' },
  { path: '/connections', icon: Link2, labelKey: 'nav.connections', color: 'teal' },
  { path: '/subscriptions', icon: ListTree, labelKey: 'nav.subscriptions', color: 'green' },
  { path: '/config-generator', icon: FileCode, labelKey: 'nav.configGenerator', color: 'orange' },
  // 根据核心类型显示不同的规则集页面
  activeCore === 'singbox'
    ? { path: '/singbox-ruleset', icon: Database, labelKey: 'nav.singboxRuleset', color: 'pink' }
    : { path: '/ruleset', icon: Database, labelKey: 'nav.ruleset', color: 'pink' },
]

// 根据核心类型动态获取系统导航项
const getSystemNavItems = (activeCore: 'mihomo' | 'singbox') => [
  { path: '/core-manage', icon: Cpu, labelKey: 'nav.coreManage', color: 'red' },
  activeCore === 'singbox'
    ? { path: '/singbox-settings', icon: SlidersHorizontal, labelKey: 'nav.singboxSettings', color: 'purple' }
    : { path: '/proxy-settings', icon: SlidersHorizontal, labelKey: 'nav.proxySettings', color: 'indigo' },
  { path: '/wireguard', icon: Network, labelKey: 'nav.wireguard', color: 'cyan' },
  { path: '/logs', icon: FileText, labelKey: 'nav.logs', color: 'yellow' },
  { path: '/settings', icon: Settings, labelKey: 'nav.settings', color: 'rose' },
]

const GITHUB_URL = 'https://github.com/star8618/P-BOX'
const TELEGRAM_URL = 'https://t.me/+8d9PNOt-w6BkNzU1'

export default function Sidebar() {
  const location = useLocation()
  const { t } = useTranslation()
  const { isOpen, close } = useSidebar()
  const { themeStyle } = useThemeStore()
  const { isRunning, fetchStatus } = useProxyStore()
  const { activeCore, setActiveCore } = useCoreStore()
  const [appVersion, setAppVersion] = useState('0.0.0')

  // 获取代理状态
  useEffect(() => {
    fetchStatus()
    const interval = setInterval(fetchStatus, 5000)
    return () => clearInterval(interval)
  }, [fetchStatus])

  // 初始化：从后端同步当前核心状态
  useEffect(() => {
    coreApi.getStatus().then(status => {
      if (status?.currentCore && status.currentCore !== activeCore) {
        setActiveCore(status.currentCore)
      }
    }).catch(() => {})
  }, []) // 只在首次加载时执行

  // 获取应用版本号
  useEffect(() => {
    systemApi.getInfo().then(info => {
      if (info?.version) {
        setAppVersion(info.version)
      }
    }).catch(() => {})
  }, [])

  // 根据核心类型获取主导航项
  const mainNavItems = getMainNavItems(activeCore)
  
  // 根据代理状态过滤导航项
  const filteredMainNavItems = mainNavItems.filter(item => {
    // 代理切换只有在运行时显示
    if (item.path === '/proxy-switch') return isRunning
    return true
  })

  const renderNavItems = (items: typeof mainNavItems) => (
    <div className="space-y-0.5">
      {items.map((item) => {
        const Icon = item.icon
        const isActive = location.pathname === item.path

        return (
          <Link
            key={item.path}
            to={item.path}
            onClick={close}
            className={cn('nav-item group', isActive && 'active')}
          >
            <div className={cn('app-icon', item.color, 'scale-90 group-hover:scale-100 transition-transform')}>
              <Icon className="w-4 h-4" />
            </div>
            <span className="text-xs font-medium truncate">{t(item.labelKey)}</span>
          </Link>
        )
      })}
    </div>
  )

  return (
    <aside className={cn(
      'sidebar w-64 flex-shrink-0 flex flex-col justify-between z-50',
      'fixed inset-y-0 left-0 transform transition-transform duration-300 lg:relative lg:translate-x-0',
      isOpen ? 'translate-x-0' : '-translate-x-full'
    )}>
      <div>
        {/* Logo */}
        <div className={cn(
          'p-6 border-b',
          themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
        )}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={cn(
                'w-10 h-10 rounded-[10px] flex items-center justify-center shadow-lg overflow-hidden',
                themeStyle === 'apple-glass' 
                  ? 'bg-gradient-to-br from-slate-100 to-white border border-black/5' 
                  : 'bg-black border border-purple-500/20'
              )}>
                <img src="/p-box-logo.png" alt="P-BOX" className="w-8 h-8 object-contain" />
              </div>
              <div>
                <h1 className={cn(
                  'text-lg font-bold tracking-widest leading-none',
                  themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                )}>
                  P-BOX
                  <span className={cn(
                    'text-xs font-normal ml-1',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
                  )}>PRO</span>
                </h1>
              </div>
            </div>
            {/* Mobile close button */}
            <button 
              onClick={close}
              className={cn(
                'p-2 rounded-lg lg:hidden',
                themeStyle === 'apple-glass' 
                  ? 'hover:bg-black/5 text-slate-500' 
                  : 'hover:bg-white/10 text-slate-400'
              )}
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Navigation menu */}
        <div className="px-3 py-4 space-y-6 overflow-y-auto max-h-[calc(100vh-200px)] custom-scrollbar">
          <div>
            <div className="px-2 pb-2 text-[10px] font-mono text-slate-500 uppercase tracking-wider">Main Module</div>
            {renderNavItems(filteredMainNavItems)}
          </div>
          
          <div>
            <div className="px-2 pb-2 text-[10px] font-mono text-slate-500 uppercase tracking-wider">System</div>
            {renderNavItems(getSystemNavItems(activeCore))}
          </div>
        </div>
      </div>

      {/* Version & GitHub */}
      <div className={cn(
        'p-4 border-t',
        themeStyle === 'apple-glass' ? 'border-black/5 bg-white/20' : 'border-white/5 bg-black/20'
      )}>
        <div className={cn(
          'rounded-lg p-3 border',
          themeStyle === 'apple-glass' ? 'bg-white/40 border-white/20' : 'bg-white/5 border-white/5'
        )}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className={cn(
                'w-8 h-8 rounded-lg flex items-center justify-center overflow-hidden',
                themeStyle === 'apple-glass' 
                  ? 'bg-gradient-to-br from-slate-100 to-white border border-black/5' 
                  : 'bg-black border border-purple-500/20'
              )}>
                <img src="/p-box-logo.png" alt="P-BOX" className="w-6 h-6 object-contain" />
              </div>
              <div>
                <div className={cn(
                  'text-xs font-semibold',
                  themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-white'
                )}>P-BOX</div>
                <div className="text-[10px] text-slate-500 font-mono">v{appVersion}</div>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <a
                href={TELEGRAM_URL}
                target="_blank"
                rel="noopener noreferrer"
                className={cn(
                  'flex items-center gap-1 p-1.5 rounded-md text-xs transition-colors',
                  themeStyle === 'apple-glass'
                    ? 'hover:bg-black/5 text-slate-600'
                    : 'hover:bg-white/10 text-slate-400'
                )}
                title="Join Telegram Group"
              >
                <Send className="w-4 h-4" />
              </a>
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className={cn(
                  'flex items-center gap-1 p-1.5 rounded-md text-xs transition-colors',
                  themeStyle === 'apple-glass'
                    ? 'hover:bg-black/5 text-slate-600'
                    : 'hover:bg-white/10 text-slate-400'
                )}
                title="GitHub"
              >
                <Github className="w-4 h-4" />
              </a>
            </div>
          </div>
        </div>
      </div>
    </aside>
  )
}
