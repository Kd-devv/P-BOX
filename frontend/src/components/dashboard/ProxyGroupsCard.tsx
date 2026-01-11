import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { LayoutGrid, Settings, Check, ChevronDown, Loader2, X, RefreshCw } from 'lucide-react'
import { mihomoApi, ProxyGroup } from '@/api/mihomo'
import { systemApi, GeoIPInfo } from '@/api/system'
import { cn } from '@/lib/utils'
import { getGroupIcon, getGroupIconColor, getDelayColorClass, hiddenGroupTypes, countryCodeToFlag } from '@/lib/proxyGroups'

interface ProxyGroupsCardProps {
  themeStyle: string
}

const STORAGE_KEY = 'dashboard_visible_groups'

export default function ProxyGroupsCard({ themeStyle }: ProxyGroupsCardProps) {
  const { t, i18n } = useTranslation()
  const [groups, setGroups] = useState<ProxyGroup[]>([])
  const [geoInfo, setGeoInfo] = useState<GeoIPInfo | null>(null)
  const [geoLoading, setGeoLoading] = useState(false)
  const [loading, setLoading] = useState(true)
  const [showSettings, setShowSettings] = useState(false)
  const [visibleGroups, setVisibleGroups] = useState<string[]>([])
  const [selectingGroup, setSelectingGroup] = useState<ProxyGroup | null>(null)
  const [switching, setSwitching] = useState(false)

  // 加载可见分组设置
  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      try {
        setVisibleGroups(JSON.parse(saved))
      } catch {
        setVisibleGroups([])
      }
    }
  }, [])

  // 获取分组列表
  const fetchGroups = useCallback(async () => {
    try {
      const proxies = await mihomoApi.getProxies()
      // 过滤出可切换的分组，排除直连/拒绝类型
      const proxyGroups = Object.values(proxies)
        .filter((p): p is ProxyGroup => 
          'all' in p && 
          Array.isArray(p.all) &&
          p.all.length > 0 &&
          !hiddenGroupTypes.includes(p.name) &&
          !hiddenGroupTypes.includes(p.type)
        )
      setGroups(proxyGroups)
      
      // 如果没有保存的设置，默认显示前5个
      if (visibleGroups.length === 0 && proxyGroups.length > 0) {
        const defaultVisible = proxyGroups.slice(0, 5).map(g => g.name)
        setVisibleGroups(defaultVisible)
        localStorage.setItem(STORAGE_KEY, JSON.stringify(defaultVisible))
      }
    } catch (err) {
      console.error('Failed to fetch groups:', err)
    } finally {
      setLoading(false)
    }
  }, [visibleGroups.length])

  // 获取出口IP信息（通过后端API）
  const fetchGeoInfo = useCallback(async () => {
    setGeoLoading(true)
    try {
      const lang = i18n.language.startsWith('zh') ? 'zh' : 'en'
      const data = await systemApi.getGeoIP(lang)
      if (data) {
        setGeoInfo(data)
      }
    } catch (err) {
      console.error('Failed to fetch geo info:', err)
    } finally {
      setGeoLoading(false)
    }
  }, [i18n.language])

  useEffect(() => {
    fetchGroups()
    fetchGeoInfo()
    const interval = setInterval(fetchGroups, 10000)
    return () => clearInterval(interval)
  }, [fetchGroups, fetchGeoInfo])

  // 切换分组可见性
  const toggleGroupVisibility = (name: string) => {
    const newVisible = visibleGroups.includes(name)
      ? visibleGroups.filter(n => n !== name)
      : [...visibleGroups, name]
    setVisibleGroups(newVisible)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newVisible))
  }

  // 切换节点
  const switchNode = async (groupName: string, nodeName: string) => {
    setSwitching(true)
    try {
      await mihomoApi.selectProxy(groupName, nodeName)
      await fetchGroups()
      setSelectingGroup(null)
    } catch (err) {
      console.error('Failed to switch node:', err)
    } finally {
      setSwitching(false)
    }
  }

  const displayedGroups = groups.filter(g => visibleGroups.includes(g.name))

  return (
    <div className="glass-card p-4 relative">
      {/* 标题栏 */}
      <div className="flex items-center justify-between mb-4 gap-3">
        <div className="flex items-center gap-2 shrink-0">
          <LayoutGrid className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />
          <span className={cn('text-sm font-medium', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>
            {t('dashboard.proxyGroups')}
          </span>
        </div>
        
        {/* 出口信息 */}
        <div className="flex items-center gap-1.5 text-xs flex-1 justify-center min-w-0 overflow-hidden">
          {geoLoading ? (
            <Loader2 className="w-3 h-3 animate-spin text-slate-500" />
          ) : geoInfo ? (
            <>
              <span className="shrink-0">{countryCodeToFlag(geoInfo.countryCode)}</span>
              <span className={cn('truncate', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300')}>{geoInfo.country}</span>
              <span className="text-slate-500 shrink-0">·</span>
              <span className="font-mono text-[11px] text-slate-500 truncate">{geoInfo.ip}</span>
              <span className="text-slate-600 shrink-0 hidden sm:inline">·</span>
              <span className="text-slate-500 text-[11px] truncate hidden sm:inline">{geoInfo.isp}</span>
            </>
          ) : (
            <span className="text-slate-500">-</span>
          )}
          <button onClick={fetchGeoInfo} className="p-1 hover:bg-white/10 rounded shrink-0 ml-1" title={t('common.refresh')}>
            <RefreshCw className={cn('w-3 h-3 text-slate-500', geoLoading && 'animate-spin')} />
          </button>
        </div>
        
        {/* 设置按钮 */}
        <button 
          onClick={() => setShowSettings(!showSettings)} 
          className={cn('p-1.5 rounded-lg transition-colors shrink-0', themeStyle === 'apple-glass' ? 'hover:bg-black/5 text-slate-500' : 'hover:bg-white/10 text-slate-400')}
        >
          <Settings className="w-4 h-4" />
        </button>
      </div>

      {/* 设置弹窗 */}
      {showSettings && (
        <div className={cn(
          'absolute top-14 right-4 w-52 rounded-xl shadow-2xl p-3 z-30 border backdrop-blur-xl',
          themeStyle === 'apple-glass' ? 'bg-white/95 border-black/10' : 'bg-neutral-900 border-white/10'
        )}>
          <div className="flex items-center justify-between mb-2">
            <span className={cn('text-xs font-medium', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400')}>
              {t('dashboard.selectGroups')}
            </span>
            <button onClick={() => setShowSettings(false)} className="p-0.5 hover:bg-white/10 rounded">
              <X className="w-3.5 h-3.5 text-slate-500" />
            </button>
          </div>
          <div className="space-y-0.5 max-h-56 overflow-y-auto">
            {groups.map(g => (
              <label 
                key={g.name}
                className={cn(
                  'flex items-center justify-between p-2 rounded-lg cursor-pointer text-xs transition-colors',
                  themeStyle === 'apple-glass' ? 'hover:bg-black/5' : 'hover:bg-white/5'
                )}
              >
                <div className="flex items-center gap-2 min-w-0">
                  {(() => { const Icon = getGroupIcon(g.name); return <Icon className="w-3.5 h-3.5 text-slate-500 shrink-0" /> })()}
                  <span className={cn('truncate', themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300')}>{g.name}</span>
                </div>
                <input 
                  type="checkbox" 
                  checked={visibleGroups.includes(g.name)}
                  onChange={() => toggleGroupVisibility(g.name)}
                  className="w-3.5 h-3.5 rounded border-slate-600 text-cyan-500 cursor-pointer"
                />
              </label>
            ))}
          </div>
        </div>
      )}

      {/* 分组卡片网格 */}
      {loading ? (
        <div className="flex items-center justify-center py-8">
          <Loader2 className={cn('w-6 h-6 animate-spin', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />
        </div>
      ) : displayedGroups.length === 0 ? (
        <div className={cn('text-center py-8 text-sm', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500')}>
          {t('dashboard.noGroupsSelected')}
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-2">
          {displayedGroups.map(group => {
            const Icon = getGroupIcon(group.name)
            const delay = (group as any).delay || 0
            
            return (
              <div 
                key={group.name}
                className={cn(
                  'p-3 rounded-xl border transition-all cursor-pointer group/card',
                  themeStyle === 'apple-glass' 
                    ? 'bg-white/50 border-black/5 hover:bg-white/70' 
                    : 'bg-white/[0.03] border-white/5 hover:bg-white/[0.06]'
                )}
                onClick={() => setSelectingGroup(group)}
              >
                {/* 分组名+延迟 */}
                <div className="flex items-center justify-between gap-1 mb-2">
                  <div className="flex items-center gap-1.5 min-w-0">
                    <div className={cn('w-5 h-5 rounded-md flex items-center justify-center shrink-0 text-white', getGroupIconColor(group.name))}>
                      <Icon className="w-3 h-3" />
                    </div>
                    <span className={cn('text-[11px] font-medium truncate', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-cyan-400')}>
                      {group.name}
                    </span>
                  </div>
                  {delay > 0 && (
                    <span className={cn('text-[10px] px-1.5 py-0.5 rounded font-mono shrink-0', getDelayColorClass(delay))}>
                      {delay}
                    </span>
                  )}
                </div>
                
                {/* 当前节点 */}
                <div className="flex items-center gap-1">
                  <span className={cn(
                    'text-xs truncate flex-1',
                    themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-200'
                  )}>
                    {group.now || '-'}
                  </span>
                  <ChevronDown className={cn(
                    'w-3 h-3 shrink-0 opacity-0 group-hover/card:opacity-60 transition-opacity',
                    themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                  )} />
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* 节点选择弹窗 */}
      {selectingGroup && (
        <div 
          className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4"
          onClick={() => !switching && setSelectingGroup(null)}
        >
          <div 
            className={cn(
              'w-full max-w-sm max-h-[70vh] rounded-2xl flex flex-col border shadow-2xl',
              themeStyle === 'apple-glass' ? 'bg-white border-black/10' : 'bg-neutral-900 border-white/10'
            )}
            onClick={e => e.stopPropagation()}
          >
            <div className={cn('p-4 border-b flex items-center justify-between', themeStyle === 'apple-glass' ? 'border-black/10' : 'border-white/10')}>
              <div className="flex items-center gap-2">
                {(() => { const Icon = getGroupIcon(selectingGroup.name); return (
                  <div className={cn('w-6 h-6 rounded-lg flex items-center justify-center text-white', getGroupIconColor(selectingGroup.name))}>
                    <Icon className="w-3.5 h-3.5" />
                  </div>
                )})()}
                <span className={cn('font-medium text-sm', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>
                  {selectingGroup.name}
                </span>
              </div>
              <button 
                onClick={() => setSelectingGroup(null)} 
                disabled={switching}
                className={cn('p-1 rounded-lg', themeStyle === 'apple-glass' ? 'hover:bg-black/5' : 'hover:bg-white/10')}
              >
                <X className="w-4 h-4 text-slate-500" />
              </button>
            </div>
            <div className="p-2 overflow-y-auto flex-1">
              {(selectingGroup.all || []).map(nodeName => {
                const isActive = nodeName === selectingGroup.now
                return (
                  <button
                    key={nodeName}
                    onClick={() => switchNode(selectingGroup.name, nodeName)}
                    disabled={switching}
                    className={cn(
                      'w-full text-left px-3 py-2.5 rounded-lg text-sm transition-colors flex items-center justify-between',
                      isActive 
                        ? (themeStyle === 'apple-glass' ? 'bg-blue-500/10 text-blue-600' : 'bg-cyan-500/20 text-cyan-400')
                        : (themeStyle === 'apple-glass' ? 'hover:bg-black/5 text-slate-700' : 'hover:bg-white/5 text-slate-300'),
                      switching && 'opacity-50 cursor-not-allowed'
                    )}
                  >
                    <span className="truncate">{nodeName}</span>
                    {isActive && <Check className="w-4 h-4 shrink-0" />}
                  </button>
                )
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
