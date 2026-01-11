import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { RefreshCw, Trash2, Globe, Loader2 } from 'lucide-react'
import { mihomoApi } from '@/api/mihomo'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

export default function ToolsPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [loading, setLoading] = useState<string | null>(null)

  const handleReloadConfig = async () => {
    try {
      setLoading('reload')
      await mihomoApi.reloadConfig()
      alert(t('tools.reloadSuccess') || '重载成功')
    } catch (e: unknown) {
      alert((e as Error)?.message || '重载失败')
    } finally {
      setLoading(null)
    }
  }

  const handleFlushDns = async () => {
    try {
      setLoading('dns')
      await mihomoApi.flushDns()
      alert(t('tools.flushDnsSuccess') || 'DNS 缓存已刷新')
    } catch (e: unknown) {
      alert((e as Error)?.message || '刷新失败')
    } finally {
      setLoading(null)
    }
  }

  const handleUpdateGeo = async () => {
    try {
      setLoading('geo')
      await mihomoApi.updateGeo()
      alert(t('tools.updateGeoSuccess') || 'GeoIP 已更新')
    } catch (e: unknown) {
      alert((e as Error)?.message || '更新失败')
    } finally {
      setLoading(null)
    }
  }

  const quickControls = [
    {
      id: 'reload',
      icon: RefreshCw,
      label: t('tools.reloadCore') || '重载核心',
      description: t('tools.reloadCoreDesc') || '重新加载配置文件',
      color: 'orange',
      onClick: handleReloadConfig,
    },
    {
      id: 'dns',
      icon: Trash2,
      label: t('tools.flushDns') || '刷新 DNS',
      description: t('tools.flushDnsDesc') || '清空 DNS 解析缓存',
      color: 'pink',
      onClick: handleFlushDns,
    },
    {
      id: 'geo',
      icon: Globe,
      label: t('tools.updateGeo') || '更新 GeoIP',
      description: t('tools.updateGeoDesc') || '更新 GeoIP/GeoSite 数据库',
      color: 'blue',
      onClick: handleUpdateGeo,
    },
  ]

  const getColorClasses = (color: string) => {
    const colors: Record<string, string> = {
      orange: themeStyle === 'apple-glass' 
        ? 'bg-orange-500 text-white' 
        : 'bg-orange-500/20 text-orange-400',
      pink: themeStyle === 'apple-glass'
        ? 'bg-pink-500 text-white'
        : 'bg-pink-500/20 text-pink-400',
      blue: themeStyle === 'apple-glass'
        ? 'bg-blue-500 text-white'
        : 'bg-blue-500/20 text-blue-400',
    }
    return colors[color] || colors.blue
  }

  return (
    <div className="space-y-6">
      {/* 快速控制 */}
      <div className="glass-card p-6">
        <h2 className={cn(
          'text-lg font-semibold mb-4',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>
          {t('tools.quickControl') || '快速控制'}
        </h2>
        <div className="space-y-3">
          {quickControls.map((control) => {
            const Icon = control.icon
            const isLoading = loading === control.id
            return (
              <button
                key={control.id}
                onClick={control.onClick}
                disabled={loading !== null}
                className={cn(
                  'w-full flex items-center gap-4 p-4 rounded-xl transition-all',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/60 hover:bg-white/80 border border-black/5'
                    : 'bg-white/5 hover:bg-white/10 border border-white/5',
                  loading !== null && 'opacity-50 cursor-not-allowed'
                )}
              >
                <div className={cn(
                  'w-12 h-12 rounded-xl flex items-center justify-center',
                  getColorClasses(control.color)
                )}>
                  {isLoading ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    <Icon className="w-5 h-5" />
                  )}
                </div>
                <div className="text-left">
                  <div className={cn(
                    'font-medium',
                    themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                  )}>
                    {control.label}
                  </div>
                  <div className={cn(
                    'text-sm',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                  )}>
                    {control.description}
                  </div>
                </div>
              </button>
            )
          })}
        </div>
      </div>

      {/* 更多工具（占位） */}
      <div className="glass-card p-6">
        <h2 className={cn(
          'text-lg font-semibold mb-4',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>
          {t('tools.moreTools') || '更多工具'}
        </h2>
        <p className={cn(
          'text-sm',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          {t('tools.comingSoon') || '更多功能开发中...'}
        </p>
      </div>
    </div>
  )
}
