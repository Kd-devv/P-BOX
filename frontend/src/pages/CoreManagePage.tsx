import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Download, Check, RefreshCw, Cpu, ArrowRight, Loader2 } from 'lucide-react'
import { coreApi, CoreStatus } from '@/api/core'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { useCoreStore, CoreType } from '@/stores/coreStore'

export default function CoreManagePage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [status, setStatus] = useState<CoreStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [switching, setSwitching] = useState(false)
  const [downloading, setDownloading] = useState<string | null>(null)
  const [downloadProgress, setDownloadProgress] = useState(0)
  const [refreshing, setRefreshing] = useState(false)
  const [platform, setPlatform] = useState<{ os: string; arch: string } | null>(null)

  useEffect(() => {
    fetchStatus()
    fetchPlatform()
  }, [])

  const fetchPlatform = async () => {
    try {
      const info = await coreApi.getPlatformInfo()
      setPlatform(info)
    } catch {
      // Ignore errors
    }
  }

  const fetchStatus = async () => {
    try {
      setLoading(true)
      const data = await coreApi.getStatus()
      setStatus(data)
    } catch {
      // Ignore errors
    } finally {
      setLoading(false)
    }
  }

  const { activeCore, setActiveCore } = useCoreStore()

  // 获取状态后同步 coreStore
  useEffect(() => {
    if (status?.currentCore && status.currentCore !== activeCore) {
      setActiveCore(status.currentCore)
    }
  }, [status?.currentCore, activeCore, setActiveCore])

  const handleSwitch = async (coreType: string) => {
    if (switching) return
    try {
      setSwitching(true)
      await coreApi.switchCore(coreType)
      // 同步更新 coreStore
      setActiveCore(coreType as CoreType)
      await fetchStatus()
    } catch {
      // Ignore errors
    } finally {
      setSwitching(false)
    }
  }

  const handleRefresh = async () => {
    if (refreshing) return
    try {
      setRefreshing(true)
      await coreApi.refreshVersions()
      await fetchStatus()
    } catch {
      // Ignore errors
    } finally {
      setRefreshing(false)
    }
  }

  const handleDownload = async (coreType: string) => {
    if (downloading) return
    try {
      setDownloading(coreType)
      setDownloadProgress(0)
      await coreApi.downloadCore(coreType)
      
      // Poll for progress
      const pollProgress = setInterval(async () => {
        try {
          const progress = await coreApi.getDownloadProgress(coreType)
          setDownloadProgress(progress.progress)
          if (!progress.downloading) {
            clearInterval(pollProgress)
            setDownloading(null)
            await fetchStatus()
          }
        } catch {
          clearInterval(pollProgress)
          setDownloading(null)
        }
      }, 500)
    } catch {
      setDownloading(null)
    }
  }

  const cores = [
    { 
      id: 'mihomo', 
      name: 'Mihomo', 
      description: t('coreManage.mihomoDesc'),
      color: 'indigo'
    },
    { 
      id: 'singbox', 
      name: 'Sing-box', 
      description: t('coreManage.singboxDesc'),
      color: 'purple'
    },
  ]

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    )
  }

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Current Core Info */}
      <div className="glass-card p-5">
        <h3 className={cn(
          'text-sm font-medium mb-4 flex items-center gap-2',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>
          <Cpu className={cn(
            'w-4 h-4',
            themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
          )} />
          {t('coreManage.currentCore')}
        </h3>
        <div className={cn(
          'flex items-center gap-4 p-4 rounded-xl border',
          themeStyle === 'apple-glass'
            ? 'bg-blue-50/50 border-blue-200'
            : 'bg-indigo-500/10 border-indigo-500/30'
        )}>
          <div className="app-icon indigo w-12 h-12 rounded-xl">
            <Cpu className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <div className={cn(
              'text-lg font-bold',
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>
              {status?.currentCore === 'mihomo' ? 'Mihomo' : 'Sing-box'}
            </div>
            <div className={cn(
              'text-sm font-mono',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              v{status?.cores[status.currentCore]?.version || 'Unknown'}
            </div>
          </div>
          <div className={cn(
            'px-3 py-1 rounded-full text-xs font-medium',
            themeStyle === 'apple-glass'
              ? 'bg-green-100 text-green-700'
              : 'bg-green-500/20 text-green-400'
          )}>
            {t('common.running')}
          </div>
        </div>
      </div>

      {/* Available Cores */}
      <div className="glass-card p-5">
        <h3 className={cn(
          'text-sm font-medium mb-4',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>
          {t('coreManage.availableCores')}
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {cores.map((core) => {
            const coreInfo = status?.cores[core.id]
            const isCurrent = status?.currentCore === core.id
            const isDownloading = downloading === core.id

            return (
              <div
                key={core.id}
                className={cn(
                  'p-5 rounded-xl border-2 transition-all',
                  isCurrent
                    ? themeStyle === 'apple-glass'
                      ? 'border-blue-500/50 bg-blue-50/30'
                      : 'border-indigo-500/50 bg-indigo-500/10'
                    : themeStyle === 'apple-glass'
                      ? 'border-transparent bg-black/[0.02] hover:bg-black/[0.04]'
                      : 'border-transparent bg-white/5 hover:bg-white/10'
                )}
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className={cn('app-icon w-10 h-10 rounded-xl', core.color)}>
                      <Cpu className="w-5 h-5" />
                    </div>
                    <div>
                      <div className={cn(
                        'font-bold',
                        themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                      )}>{core.name}</div>
                      <div className={cn(
                        'text-xs',
                        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                      )}>{core.description}</div>
                    </div>
                  </div>
                  {isCurrent && (
                    <Check className={cn(
                      'w-5 h-5',
                      themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
                    )} />
                  )}
                </div>

                <div className={cn(
                  'text-xs space-y-1 mb-4',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )}>
                  <div className="flex justify-between">
                    <span>{t('coreManage.installed')}:</span>
                    <span className="font-mono">
                      {coreInfo?.installed ? `v${coreInfo.version}` : t('coreManage.notInstalled')}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span>{t('coreManage.latest')}:</span>
                    <span className="font-mono">
                      {coreInfo?.latestVersion || t('common.unknown')}
                    </span>
                  </div>
                </div>

                {/* Download progress */}
                {isDownloading && (
                  <div className="mb-4">
                    <div className="w-full bg-slate-800/50 h-2 rounded-full overflow-hidden">
                      <div 
                        className="bg-indigo-500 h-full transition-all duration-300"
                        style={{ width: `${downloadProgress}%` }}
                      />
                    </div>
                    <div className="text-xs text-center mt-1 text-slate-400">
                      {Math.round(downloadProgress)}%
                    </div>
                  </div>
                )}

                <div className="flex gap-2">
                  {!coreInfo?.installed || coreInfo.version !== coreInfo.latestVersion ? (
                    <button
                      onClick={() => handleDownload(core.id)}
                      disabled={isDownloading}
                      className={cn(
                        'flex-1 flex items-center justify-center gap-2 py-2 rounded-lg text-xs font-medium transition-colors',
                        themeStyle === 'apple-glass'
                          ? 'bg-blue-500 hover:bg-blue-600 text-white'
                          : 'bg-indigo-500 hover:bg-indigo-600 text-white',
                        isDownloading && 'opacity-50'
                      )}
                    >
                      {isDownloading ? (
                        <RefreshCw className="w-3 h-3 animate-spin" />
                      ) : (
                        <Download className="w-3 h-3" />
                      )}
                      {coreInfo?.installed ? t('common.update') : t('coreManage.install')}
                    </button>
                  ) : null}

                  {coreInfo?.installed && !isCurrent && (
                    <button
                      onClick={() => handleSwitch(core.id)}
                      disabled={switching}
                      className={cn(
                        'flex-1 flex items-center justify-center gap-2 py-2 rounded-lg text-xs font-medium transition-colors border',
                        themeStyle === 'apple-glass'
                          ? 'bg-white/50 hover:bg-white border-black/10 text-slate-700'
                          : 'bg-white/5 hover:bg-white/10 border-white/10 text-white'
                      )}
                    >
                      <ArrowRight className="w-3 h-3" />
                      {t('coreManage.switchTo')}
                    </button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Platform Info & Refresh button */}
      <div className="flex items-center justify-between">
        {platform && (
          <div className={cn(
            'text-xs font-mono',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>
            {t('coreManage.platform')}: {platform.os}/{platform.arch}
          </div>
        )}
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            themeStyle === 'apple-glass'
              ? 'bg-black/5 hover:bg-black/10 text-slate-600'
              : 'bg-white/5 hover:bg-white/10 text-slate-300',
            refreshing && 'opacity-50'
          )}
        >
          <RefreshCw className={cn('w-4 h-4', refreshing && 'animate-spin')} />
          {refreshing ? t('common.refreshing') : t('common.refresh')}
        </button>
      </div>
    </div>
  )
}
