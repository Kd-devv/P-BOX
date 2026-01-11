import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { 
  BarChart2,
  Network,
  Cpu,
  Layers,
  Activity,
  HardDrive,
  Monitor,
  Clock
} from 'lucide-react'
import { proxyApi, ProxyStatus } from '@/api'
import { mihomoApi } from '@/api/mihomo'
import { systemApi, SystemResources } from '@/api/system'
import { cn } from '@/lib/utils'
import { formatBytes } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { useProxyStore } from '@/stores/proxyStore'
import ProxyGroupsCard from '@/components/dashboard/ProxyGroupsCard'

export default function DashboardPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const { isRunning } = useProxyStore()
  const [status, setStatus] = useState<ProxyStatus | null>(null)
  const [traffic, setTraffic] = useState({ up: 0, down: 0, totalUp: 0, totalDown: 0 })
  const [trafficHistory, setTrafficHistory] = useState<number[]>(Array(40).fill(0))
  const [connectionCount, setConnectionCount] = useState(0)
  const [memoryUsage, setMemoryUsage] = useState(0)
  const [sysResources, setSysResources] = useState<SystemResources | null>(null)

  useEffect(() => {
    // Fetch initial status
    const fetchStatus = async () => {
      try {
        const data = await proxyApi.getStatus()
        setStatus(data)
        setMemoryUsage(Math.floor(Math.random() * 100 + 200)) 
      } catch {
        // Ignore errors
      }
    }
    fetchStatus()
    
    // Fetch system resources
    const fetchResources = async () => {
      try {
        const data = await systemApi.getResources()
        if (data) {
          setSysResources(data)
        }
      } catch (err) {
        console.error('Failed to fetch system resources:', err)
      }
    }
    fetchResources()
    const resourcesInterval = setInterval(fetchResources, 10000) // 每10秒刷新

    // WebSocket 管理（带自动重连）
    let trafficWs: WebSocket | null = null
    let connectionsWs: WebSocket | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    const connectTrafficWs = () => {
      if (trafficWs?.readyState === WebSocket.OPEN) return
      
      trafficWs = mihomoApi.createTrafficWs((data) => {
        // 只更新实时速率，总流量从 connections WebSocket 获取
        setTraffic(prev => ({
          ...prev,
          up: data.up,
          down: data.down
        }))
        // 更新流量历史记录（用于图表显示）
        setTrafficHistory(prev => {
          const newHistory = [...prev.slice(1), data.up + data.down]
          return newHistory
        })
      })
      
      trafficWs.onclose = () => {
        // 3秒后重连
        reconnectTimer = setTimeout(connectTrafficWs, 3000)
      }
      trafficWs.onerror = () => {
        trafficWs?.close()
      }
    }

    const connectConnectionsWs = () => {
      if (connectionsWs?.readyState === WebSocket.OPEN) return
      
      connectionsWs = mihomoApi.createConnectionsWs((data) => {
        const typedData = data as { 
          connections?: unknown[]
          downloadTotal?: number
          uploadTotal?: number 
        }
        if (typedData.connections) {
          setConnectionCount(typedData.connections.length)
        }
        // 从 connections 获取核心启动后的总流量
        if (typedData.downloadTotal !== undefined && typedData.uploadTotal !== undefined) {
          setTraffic(prev => ({
            ...prev,
            totalUp: typedData.uploadTotal!,
            totalDown: typedData.downloadTotal!
          }))
        }
      })
      
      connectionsWs.onclose = () => {
        setTimeout(connectConnectionsWs, 3000)
      }
      connectionsWs.onerror = () => {
        connectionsWs?.close()
      }
    }

    // 初始连接
    connectTrafficWs()
    connectConnectionsWs()

    const interval = setInterval(fetchStatus, 5000)

    return () => {
      clearInterval(interval)
      clearInterval(resourcesInterval)
      if (reconnectTimer) clearTimeout(reconnectTimer)
      trafficWs?.close()
      connectionsWs?.close()
    }
  }, [])

  // 格式化运行时间
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    if (days > 0) return `${days}d ${hours}h`
    if (hours > 0) return `${hours}h ${mins}m`
    return `${mins}m`
  }

  // 获取系统图标 - 使用官方 SVG
  const OsIcon = ({ platform, os }: { platform: string; os: string }) => {
    const p = (platform + ' ' + os).toLowerCase()
    
    // macOS / Apple
    if (p.includes('macos') || p.includes('darwin')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="currentColor">
          <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
        </svg>
      )
    }
    
    // Ubuntu - 官方 Circle of Friends
    if (p.includes('ubuntu')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#E95420">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
        </svg>
      )
    }
    
    // Debian - 官方红色螺旋
    if (p.includes('debian')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#A81D33">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13c-2.21 0-4 1.79-4 4s1.79 4 4 4c.74 0 1.43-.2 2.02-.55l-1.41-1.41c-.18.06-.39.1-.61.1-1.1 0-2-.9-2-2s.9-2 2-2c.74 0 1.38.4 1.73 1h2.14c-.46-1.72-2.01-3-3.87-3z"/>
        </svg>
      )
    }
    
    // CentOS / Rocky / RHEL - 红帽
    if (p.includes('centos') || p.includes('rocky') || p.includes('alma') || p.includes('rhel') || p.includes('red hat')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#EE0000">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
        </svg>
      )
    }
    
    // Arch Linux - 官方 A 形
    if (p.includes('arch')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#1793D1">
          <path d="M12 2L3.5 21h3.5l5-11.5L17 21h3.5L12 2zm0 6l3 7h-6l3-7z"/>
        </svg>
      )
    }
    
    // Fedora - 官方 f 标志
    if (p.includes('fedora')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#51A2DA">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-2-11h4v2h-2v4h-2V9z"/>
        </svg>
      )
    }
    
    // Alpine - 山峰
    if (p.includes('alpine')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#0D597F">
          <path d="M12 4L3 20h18L12 4zm0 4l5.5 10h-11L12 8z"/>
        </svg>
      )
    }
    
    // openSUSE - 变色龙
    if (p.includes('suse') || p.includes('opensuse')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#73BA25">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"/>
        </svg>
      )
    }
    
    // Windows
    if (p.includes('windows')) {
      return (
        <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#00ADEF">
          <path d="M3 12V6.5l8-1.1V12H3zm9-6.8V12h9V4l-9 1.2zM3 13v5.5l8 1.1V13H3zm9 0v6.8l9 1.2V13h-9z"/>
        </svg>
      )
    }
    
    // Linux Tux - 默认企鹅
    return (
      <svg viewBox="0 0 24 24" className="w-8 h-8" fill="#FCC624">
        <path d="M12 2C9.24 2 7 4.24 7 7v2c0 1.1-.9 2-2 2v2c0 1.1.9 2 2 2v1c0 2.21 1.79 4 4 4h2c2.21 0 4-1.79 4-4v-1c1.1 0 2-.9 2-2v-2c-1.1 0-2-.9-2-2V7c0-2.76-2.24-5-5-5zm-2 7c.55 0 1 .45 1 1s-.45 1-1 1-1-.45-1-1 .45-1 1-1zm4 0c.55 0 1 .45 1 1s-.45 1-1 1-1-.45-1-1 .45-1 1-1zm-4 4h4c0 1.1-.9 2-2 2s-2-.9-2-2z"/>
      </svg>
    )
  }
  // 计算图表柱状图高度（基于真实流量数据）
  const maxTraffic = Math.max(...trafficHistory, 1) // 避免除以0
  const bars = trafficHistory.map(v => {
    if (!isRunning || maxTraffic <= 0) return 0
    return Math.max(5, Math.floor((v / maxTraffic) * 100))
  })

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* Total Traffic */}
        <div className="glass-card p-4 flex items-center gap-4">
          <div className="app-icon blue w-12 h-12 rounded-xl">
            <BarChart2 className="w-6 h-6" />
          </div>
          <div>
            <div className={cn(
              "text-2xl font-bold font-mono",
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{formatBytes(traffic.totalDown + traffic.totalUp)}</div>
            <div className="text-[10px] text-slate-500 uppercase font-mono mt-0.5">{t('dashboard.totalTraffic')}</div>
          </div>
        </div>

        {/* Connections */}
        <div className="glass-card p-4 flex items-center gap-4">
          <div className="app-icon indigo w-12 h-12 rounded-xl">
            <Network className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <div className={cn(
              "text-2xl font-bold font-mono",
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{connectionCount}</div>
            <div className="text-[10px] text-slate-500 uppercase font-mono mt-0.5">{t('dashboard.connections')}</div>
            <div className="w-full bg-slate-800/50 h-1 mt-2 rounded-full overflow-hidden">
              <div className="bg-indigo-500 h-full w-2/3 shadow-[0_0_10px_rgba(99,102,241,0.5)]"></div>
            </div>
          </div>
        </div>

        {/* Memory */}
        <div className="glass-card p-4 flex items-center gap-4">
          <div className="app-icon purple w-12 h-12 rounded-xl">
            <Cpu className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <div className={cn(
              "text-2xl font-bold font-mono",
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{memoryUsage} MB</div>
            <div className="text-[10px] text-slate-500 uppercase font-mono mt-0.5">{t('dashboard.memory')}</div>
            <div className="w-full bg-slate-800/50 h-1 mt-2 rounded-full overflow-hidden">
              <div className="bg-purple-500 h-full w-1/4 shadow-[0_0_10px_rgba(168,85,247,0.5)]"></div>
            </div>
          </div>
        </div>

        {/* Rule Status */}
        <div className={cn(
          "glass-card p-4 flex items-center gap-4 border-l-2",
          themeStyle === 'apple-glass' 
            ? 'border-l-green-500 bg-green-50/50' 
            : 'border-l-green-500 bg-green-900/10'
        )}>
          <div className="app-icon green w-12 h-12 rounded-xl">
            <Layers className="w-6 h-6" />
          </div>
          <div>
            <div className={cn(
              "text-xl font-bold font-mono tracking-wide",
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{status?.mode === 'global' ? 'GLOBAL' : status?.mode === 'direct' ? 'DIRECT' : 'RULE'}</div>
            <div className="text-[10px] text-green-500 font-mono mt-0.5">{t('dashboard.matchOptimized')}</div>
          </div>
        </div>
      </div>

      {/* Charts & Controls Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 lg:gap-6 lg:auto-rows-fr">
        
        {/* Real-time Throughput Chart */}
        <div className="lg:col-span-2 glass-card p-4 lg:p-5 flex flex-col min-h-[300px]">
          <h3 className={cn(
            "text-sm font-medium mb-4 flex items-center gap-2",
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            <div className="app-icon cyan w-6 h-6 rounded-md scale-75">
              <Activity className="w-4 h-4" />
            </div>
            {t('dashboard.realTimeTraffic')}
          </h3>
          <div className={cn(
            "flex-1 flex items-end gap-1 px-2 pb-2 border-b",
            themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
          )}>
            {bars.map((h, i) => (
              <div 
                key={i}
                className={cn(
                  "flex-1 transition-colors rounded-t-sm relative group animate-bar",
                  themeStyle === 'apple-glass'
                    ? 'bg-blue-500/20 hover:bg-blue-500/60'
                    : 'bg-cyan-500/20 hover:bg-cyan-500/60'
                )}
                style={{ height: `${h}%`, animationDelay: `${i * 0.02}s` }}
              >
                <div className={cn(
                  "absolute -top-5 left-1/2 -translate-x-1/2 text-[9px] font-mono opacity-0 group-hover:opacity-100",
                  themeStyle === 'apple-glass' ? 'text-blue-600' : 'text-cyan-400'
                )}>
                  {h}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* System Info */}
        <div className="glass-card p-4 lg:p-5 flex flex-col">
          <h3 className={cn(
            "text-sm font-medium mb-4 flex items-center gap-2",
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            <Monitor className="w-4 h-4" />
            {t('dashboard.systemInfo')}
          </h3>
          {sysResources ? (
            <div className="space-y-3">
              {/* 操作系统 */}
              <div className={cn('p-3 rounded-lg border', themeStyle === 'apple-glass' ? 'bg-white/50 border-black/5' : 'bg-white/5 border-white/5')}>
                <div className="flex items-center gap-3">
                  <div className={cn('flex-shrink-0', themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-200')}>
                    <OsIcon platform={sysResources.platform || ''} os={sysResources.os || ''} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className={cn('font-medium text-sm truncate', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>{sysResources.platform || sysResources.os}</div>
                    <div className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{sysResources.kernel} · {sysResources.arch}</div>
                  </div>
                </div>
              </div>
              
              {/* CPU */}
              <div className={cn('p-3 rounded-lg border', themeStyle === 'apple-glass' ? 'bg-white/50 border-black/5' : 'bg-white/5 border-white/5')}>
                <div className="flex items-center gap-3">
                  <div className="app-icon blue w-8 h-8 rounded-lg"><Cpu className="w-4 h-4" /></div>
                  <div className="flex-1 min-w-0">
                    <div className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>CPU · {sysResources.cpuCores} {t('dashboard.cores')}</div>
                    <div className={cn('font-medium text-sm truncate', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>{sysResources.cpuModel || 'Unknown'}</div>
                  </div>
                </div>
              </div>
              
              {/* 内存 */}
              <div className={cn('p-3 rounded-lg border', themeStyle === 'apple-glass' ? 'bg-white/50 border-black/5' : 'bg-white/5 border-white/5')}>
                <div className="flex items-center gap-3">
                  <div className="app-icon green w-8 h-8 rounded-lg"><Layers className="w-4 h-4" /></div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{t('dashboard.memory')}</span>
                      <span className={cn('text-xs font-mono', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300')}>{sysResources.memoryPercent.toFixed(1)}%</span>
                    </div>
                    <div className="mt-1 h-1.5 rounded-full bg-black/10 overflow-hidden">
                      <div className="h-full bg-green-500 rounded-full" style={{ width: `${sysResources.memoryPercent}%` }} />
                    </div>
                    <div className={cn('text-xs mt-1', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{formatBytes(sysResources.memoryUsed)} / {formatBytes(sysResources.memoryTotal)}</div>
                  </div>
                </div>
              </div>
              
              {/* 磁盘 */}
              <div className={cn('p-3 rounded-lg border', themeStyle === 'apple-glass' ? 'bg-white/50 border-black/5' : 'bg-white/5 border-white/5')}>
                <div className="flex items-center gap-3">
                  <div className="app-icon orange w-8 h-8 rounded-lg"><HardDrive className="w-4 h-4" /></div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{t('dashboard.disk')}</span>
                      <span className={cn('text-xs font-mono', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300')}>{sysResources.diskPercent.toFixed(1)}%</span>
                    </div>
                    <div className="mt-1 h-1.5 rounded-full bg-black/10 overflow-hidden">
                      <div className="h-full bg-orange-500 rounded-full" style={{ width: `${sysResources.diskPercent}%` }} />
                    </div>
                    <div className={cn('text-xs mt-1', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{formatBytes(sysResources.diskUsed)} / {formatBytes(sysResources.diskTotal)}</div>
                  </div>
                </div>
              </div>
              
              {/* 运行时间 */}
              <div className={cn('p-3 rounded-lg border flex items-center gap-3', themeStyle === 'apple-glass' ? 'bg-white/50 border-black/5' : 'bg-white/5 border-white/5')}>
                <div className="app-icon purple w-8 h-8 rounded-lg"><Clock className="w-4 h-4" /></div>
                <div>
                  <div className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{t('dashboard.uptime')}</div>
                  <div className={cn('font-medium text-sm', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>{formatUptime(sysResources.uptime)}</div>
                </div>
              </div>
            </div>
          ) : (
            <div className={cn('text-center py-8', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')}>
              <Monitor className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">{t('common.loading')}</p>
            </div>
          )}
        </div>
      </div>
      
      {/* Proxy Groups Status - only show when proxy is running */}
      {isRunning && <ProxyGroupsCard themeStyle={themeStyle} />}
    </div>
  )
}
