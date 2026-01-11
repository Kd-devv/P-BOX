import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Trash2, Pause, Play, ArrowDown, AlertCircle, AlertTriangle, Info, Bug, Search } from 'lucide-react'
import { mihomoApi, LogEntry } from '@/api/mihomo'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

type LogLevel = 'debug' | 'info' | 'warning' | 'error'

interface LogItem extends LogEntry {
  id: number
  time: string
}

export default function LogsPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [logs, setLogs] = useState<LogItem[]>([])
  const [level, setLevel] = useState<LogLevel>('info')
  const [paused, setPaused] = useState(false)
  const [filter, setFilter] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const logsEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const idRef = useRef(0)
  const pausedRef = useRef(paused)

  // 保持 pausedRef 同步
  useEffect(() => {
    pausedRef.current = paused
  }, [paused])

  // WebSocket 连接
  useEffect(() => {
    const connectWs = () => {
      wsRef.current?.close()
      // 切换级别时清空日志
      setLogs([])
      wsRef.current = mihomoApi.createLogsWs((data) => {
        if (pausedRef.current) return
        const logItem: LogItem = {
          ...data,
          id: idRef.current++,
          time: new Date().toLocaleTimeString()
        }
        setLogs(prev => [...prev.slice(-500), logItem])
      }, level)
    }

    connectWs()
    return () => {
      wsRef.current?.close()
    }
  }, [level])

  // 检测用户是否在底部
  const handleScroll = useCallback(() => {
    const container = logsContainerRef.current
    if (!container) return
    
    const { scrollTop, scrollHeight, clientHeight } = container
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
    setAutoScroll(isAtBottom)
  }, [])

  // 只有在 autoScroll 为 true 时才自动滚动
  useEffect(() => {
    if (autoScroll && !paused) {
      logsEndRef.current?.scrollIntoView({ behavior: 'auto' })
    }
  }, [logs, autoScroll, paused])

  // 获取日志图标
  const getLogIcon = (type: string) => {
    switch (type) {
      case 'error': return <AlertCircle className="w-3.5 h-3.5 text-red-500" />
      case 'warning': return <AlertTriangle className="w-3.5 h-3.5 text-yellow-500" />
      case 'info': return <Info className="w-3.5 h-3.5 text-blue-500" />
      default: return <Bug className="w-3.5 h-3.5 text-slate-400" />
    }
  }

  const handleClear = () => {
    setLogs([])
  }

  const getLevelColor = (type: string) => {
    switch (type) {
      case 'error': return 'text-red-400'
      case 'warning': return 'text-yellow-400'
      case 'debug': return 'text-slate-500'
      default: return themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
    }
  }

  const filteredLogs = filter 
    ? logs.filter(log => log.payload.toLowerCase().includes(filter.toLowerCase()))
    : logs

  const levels: LogLevel[] = ['debug', 'info', 'warning', 'error']

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
        <h2 className={cn(
          'text-lg font-semibold',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>{t('logs.title')}</h2>
        <div className="flex flex-wrap gap-2">
          {/* Level selector */}
          <div className="flex rounded-lg overflow-hidden border border-white/10">
            {levels.map(l => (
              <button
                key={l}
                onClick={() => setLevel(l)}
                className={cn(
                  'px-2 sm:px-3 py-1.5 text-xs font-medium transition-colors',
                  level === l
                    ? themeStyle === 'apple-glass'
                      ? 'bg-blue-500 text-white'
                      : 'bg-indigo-500 text-white'
                    : themeStyle === 'apple-glass'
                      ? 'bg-white/50 text-slate-600 hover:bg-white'
                      : 'bg-white/5 text-slate-400 hover:bg-white/10'
                )}
              >
                {t(`logs.${l}`)}
              </button>
            ))}
          </div>

          <button
            onClick={() => setPaused(!paused)}
            className={cn(
              'control-btn text-xs',
              paused ? 'primary' : 'secondary'
            )}
          >
            {paused ? <Play className="w-3 h-3" /> : <Pause className="w-3 h-3" />}
          </button>

          <button
            onClick={handleClear}
            className="control-btn danger text-xs"
          >
            <Trash2 className="w-3 h-3" />
            {t('logs.clear')}
          </button>
        </div>
      </div>

      {/* Filter */}
      <div className="relative mb-4">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none z-10" />
        <input
          type="text"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder={t('logs.filter') + '...'}
          className="form-input !pl-10"
        />
      </div>

      {/* Logs container */}
      <div className="relative flex-1">
        <div 
          ref={logsContainerRef}
          onScroll={handleScroll}
          className={cn(
            'absolute inset-0 glass-card p-4 overflow-auto font-mono text-xs',
            themeStyle === 'apple-glass' ? 'bg-white/60' : 'bg-black/30'
          )}
        >
          {filteredLogs.length === 0 ? (
            <div className={cn(
              'text-center py-8',
              themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
            )}>
              {logs.length === 0 ? t('logs.noLogs') || '等待日志...' : t('logs.noMatch') || '无匹配日志'}
            </div>
          ) : (
            <div className="space-y-1">
              {filteredLogs.map(log => (
                <div
                  key={log.id}
                  className={cn(
                    'flex items-start gap-2 py-1.5 px-2 rounded',
                    'hover:bg-white/5 transition-colors',
                    log.type === 'error' && 'bg-red-500/5'
                  )}
                >
                  <span className="flex-shrink-0 mt-0.5">
                    {getLogIcon(log.type)}
                  </span>
                  <span className={cn(
                    'text-[10px] opacity-60 w-16 flex-shrink-0',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                  )}>
                    {log.time}
                  </span>
                  <span className={cn(
                    'flex-1 break-all',
                    getLevelColor(log.type)
                  )}>
                    {log.payload}
                  </span>
                </div>
              ))}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>

        {/* 回到底部按钮 */}
        {!autoScroll && (
          <button
            onClick={() => {
              setAutoScroll(true)
              logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
            }}
            className={cn(
              'absolute bottom-4 right-4 p-2 rounded-full shadow-lg transition-all',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 text-white hover:bg-blue-600'
                : 'bg-indigo-500 text-white hover:bg-indigo-600'
            )}
            title="回到底部"
          >
            <ArrowDown className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Status bar */}
      <div className={cn(
        'flex items-center justify-between mt-2 text-xs',
        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
      )}>
        <span>{filteredLogs.length} {t('logs.title')}</span>
        <span className="flex items-center gap-2">
          <span className={cn(
            'w-2 h-2 rounded-full',
            paused ? 'bg-yellow-500' : 'bg-green-500 animate-pulse'
          )} />
          {paused ? 'Paused' : 'Streaming'}
        </span>
      </div>
    </div>
  )
}
