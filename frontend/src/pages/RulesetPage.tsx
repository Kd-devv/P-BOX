import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Download, Globe, Shield, Loader2, Check, X, Clock, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

interface RuleFile {
  name: string
  url: string
  path: string
  description: string
  size: number
  updatedAt: string
  status: 'pending' | 'downloading' | 'completed' | 'failed'
}

interface RuleSetConfig {
  autoUpdate: boolean
  updateInterval: number
  lastUpdate: string
  githubProxy: string
  githubProxies: string[]
  customProxies: string[]
}

// 默认 GitHub 代理列表
const defaultGitHubProxies = [
  'https://ghfast.top',
  'https://ghproxy.link',
  'https://gh-proxy.com',
  'https://ghps.cc',
]

// API 基础路径
const API_BASE = '/api'

export default function RulesetPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [geoFiles, setGeoFiles] = useState<RuleFile[]>([])
  const [providerFiles, setProviderFiles] = useState<RuleFile[]>([])
  const [config, setConfig] = useState<RuleSetConfig>({
    autoUpdate: true,
    updateInterval: 1,
    lastUpdate: '',
    githubProxy: '',
    githubProxies: defaultGitHubProxies,
    customProxies: [],
  })
  const [newProxy, setNewProxy] = useState('')
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadData()
    const interval = setInterval(checkStatus, 3000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const [geoRes, providerRes, configRes] = await Promise.all([
        fetch(`${API_BASE}/ruleset/geo`).then(r => r.json()),
        fetch(`${API_BASE}/ruleset/providers`).then(r => r.json()),
        fetch(`${API_BASE}/ruleset/config`).then(r => r.json())
      ])
      setGeoFiles(geoRes.data || [])
      setProviderFiles(providerRes.data || [])
      const loadedConfig = configRes.data || {}
      setConfig({
        autoUpdate: loadedConfig.autoUpdate ?? true,
        updateInterval: loadedConfig.updateInterval ?? 1,
        lastUpdate: loadedConfig.lastUpdate ?? '',
        githubProxy: loadedConfig.githubProxy ?? '',
        githubProxies: loadedConfig.githubProxies?.length ? loadedConfig.githubProxies : defaultGitHubProxies,
        customProxies: loadedConfig.customProxies ?? [],
      })
    } catch {
      // Ignore errors
    } finally {
      setLoading(false)
    }
  }

  const checkStatus = async () => {
    try {
      const res = await fetch(`${API_BASE}/ruleset/status`)
      const data = await res.json()
      const status = data.data || {}
      if (!status.updating && updating) {
        loadData()
      }
      setUpdating(status.updating)
    } catch {
      // Ignore
    }
  }

  const handleUpdateAll = async () => {
    try {
      setUpdating(true)
      await fetch(`${API_BASE}/ruleset/update`, { 
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ githubProxy: config.githubProxy })
      })
      alert(t('ruleset.updateStarted') || '开始更新规则文件')
    } catch {
      alert(t('common.error') || '启动更新失败')
      setUpdating(false)
    }
  }

  const handleSaveConfig = async () => {
    try {
      setSaving(true)
      await fetch(`${API_BASE}/ruleset/config`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      alert(t('common.success') || '配置已保存')
    } catch {
      alert(t('common.error') || '保存配置失败')
    } finally {
      setSaving(false)
    }
  }

  const formatSize = (bytes: number) => {
    if (!bytes || bytes === 0) return '-'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <Check className="h-4 w-4 text-green-500" />
      case 'downloading':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />
      case 'failed':
        return <X className="h-4 w-4 text-red-500" />
      default:
        return <Clock className="h-4 w-4 text-gray-400" />
    }
  }

  const completedCount = [...geoFiles, ...providerFiles].filter(f => f.status === 'completed').length
  const totalCount = geoFiles.length + providerFiles.length

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('ruleset.title') || '规则集'}</h2>
          <p className={cn(
            'text-sm mt-1',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>
            已下载 {completedCount}/{totalCount} 个规则文件
            {config.lastUpdate && ` · 最后更新: ${config.lastUpdate}`}
          </p>
        </div>
        <button
          onClick={handleUpdateAll}
          disabled={updating}
          className="control-btn primary text-xs"
        >
          {updating ? (
            <Loader2 className="w-3 h-3 animate-spin" />
          ) : (
            <Download className="w-3 h-3" />
          )}
          {updating ? (t('common.updating') || '更新中...') : (t('ruleset.updateAll') || '全部更新')}
        </button>
      </div>

      {/* 配置区域 */}
      <div className="glass-card p-4">
        <div className="flex items-center gap-2 mb-4">
          <Settings className={cn(
            'h-5 w-5',
            themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
          )} />
          <span className={cn(
            'font-medium',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('ruleset.settings') || '更新设置'}</span>
        </div>
        <div className="flex flex-wrap items-center gap-6">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={config.autoUpdate}
              onChange={(e) => setConfig({ ...config, autoUpdate: e.target.checked })}
              className="w-4 h-4 rounded"
            />
            <span className={cn(
              'text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('ruleset.autoUpdate') || '自动更新'}</span>
          </label>
          <div className="flex items-center gap-2">
            <span className={cn(
              'text-sm whitespace-nowrap',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>{t('ruleset.interval') || '更新间隔'}:</span>
            <select
              value={config.updateInterval}
              onChange={(e) => setConfig({ ...config, updateInterval: Number(e.target.value) })}
              className="form-input text-sm py-1.5 w-24"
            >
              <option value={1}>1 天</option>
              <option value={2}>2 天</option>
              <option value={3}>3 天</option>
              <option value={5}>5 天</option>
              <option value={7}>7 天</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <span className={cn(
              'text-sm whitespace-nowrap',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>GitHub代理:</span>
            {config.githubProxy === '__custom__' ? (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={newProxy}
                  onChange={(e) => setNewProxy(e.target.value)}
                  placeholder="https://proxy.example.com"
                  className="form-input text-sm py-1.5 w-48"
                  autoFocus
                />
                <button
                  onClick={() => {
                    if (newProxy && newProxy.startsWith('https://')) {
                      const proxies = [...(config.customProxies || []), newProxy]
                      setConfig({ ...config, customProxies: proxies, githubProxy: newProxy })
                      setNewProxy('')
                    }
                  }}
                  disabled={!newProxy || !newProxy.startsWith('https://')}
                  className={cn(
                    'control-btn text-xs px-3 whitespace-nowrap',
                    newProxy && newProxy.startsWith('https://') ? 'primary' : 'secondary opacity-50'
                  )}
                >
                  确定
                </button>
                <button
                  onClick={() => {
                    setConfig({ ...config, githubProxy: '' })
                    setNewProxy('')
                  }}
                  className="control-btn secondary text-xs px-2 whitespace-nowrap"
                >
                  取消
                </button>
              </div>
            ) : (
              <select
                value={config.githubProxy}
                onChange={(e) => {
                  if (e.target.value === '__custom__') {
                    setConfig({ ...config, githubProxy: '__custom__' })
                  } else {
                    setConfig({ ...config, githubProxy: e.target.value })
                    setNewProxy('')
                  }
                }}
                className="form-input text-sm py-1.5 w-48"
              >
                <option value="">直连(无代理)</option>
                {config.githubProxies?.filter(p => p).map(proxy => (
                  <option key={proxy} value={proxy}>{proxy.replace('https://', '')}</option>
                ))}
                {config.customProxies?.map(proxy => (
                  <option key={proxy} value={proxy}>★ {proxy.replace('https://', '')}</option>
                ))}
                <option value="__custom__">+ 添加自定义...</option>
              </select>
            )}
          </div>
          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="control-btn secondary text-xs whitespace-nowrap"
          >
            {saving ? <Loader2 className="w-3 h-3 animate-spin" /> : null}
            {t('common.save') || '保存设置'}
          </button>
        </div>
      </div>

      {/* GEO 数据库 */}
      <div className="glass-card overflow-hidden">
        <div className={cn(
          'px-4 py-3 border-b flex items-center gap-2',
          themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
        )}>
          <Globe className="h-5 w-5 text-blue-500" />
          <span className={cn(
            'font-medium',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>GEO 数据库</span>
          <span className={cn(
            'text-xs',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>({geoFiles.length})</span>
        </div>
        <div className={cn(
          'divide-y',
          themeStyle === 'apple-glass' ? 'divide-black/5' : 'divide-white/5'
        )}>
          {geoFiles.map((file) => (
            <div key={file.name} className={cn(
              'px-4 py-3 flex items-center justify-between',
              themeStyle === 'apple-glass' ? 'hover:bg-black/5' : 'hover:bg-white/5'
            )}>
              <div className="flex items-center gap-3">
                {getStatusIcon(file.status)}
                <div>
                  <div className={cn(
                    'font-medium text-sm',
                    themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                  )}>{file.name}</div>
                  <div className={cn(
                    'text-xs',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                  )}>{file.description}</div>
                </div>
              </div>
              <div className={cn(
                'text-right text-sm',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>
                <div>{formatSize(file.size)}</div>
                {file.updatedAt && (
                  <div className="text-xs">{file.updatedAt}</div>
                )}
              </div>
            </div>
          ))}
          {geoFiles.length === 0 && (
            <div className={cn(
              'px-4 py-8 text-center text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              {t('ruleset.noGeoFiles') || '暂无 GEO 数据文件'}
            </div>
          )}
        </div>
      </div>

      {/* 规则提供者 */}
      <div className="glass-card overflow-hidden">
        <div className={cn(
          'px-4 py-3 border-b flex items-center gap-2',
          themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
        )}>
          <Shield className="h-5 w-5 text-green-500" />
          <span className={cn(
            'font-medium',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('ruleset.providers') || '规则提供者'}</span>
          <span className={cn(
            'text-xs',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>({providerFiles.length})</span>
        </div>
        <div className={cn(
          'grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0',
          themeStyle === 'apple-glass' ? 'divide-black/5' : 'divide-white/5'
        )}>
          {providerFiles.map((file, index) => (
            <div 
              key={file.name} 
              className={cn(
                'px-4 py-3 flex items-center justify-between',
                themeStyle === 'apple-glass' ? 'hover:bg-black/5' : 'hover:bg-white/5',
                index % 2 === 0 && themeStyle === 'apple-glass' ? 'md:border-r md:border-black/5' : '',
                index % 2 === 0 && themeStyle !== 'apple-glass' ? 'md:border-r md:border-white/5' : '',
                index >= 2 && themeStyle === 'apple-glass' ? 'md:border-t md:border-black/5' : '',
                index >= 2 && themeStyle !== 'apple-glass' ? 'md:border-t md:border-white/5' : ''
              )}
            >
              <div className="flex items-center gap-3 min-w-0">
                {getStatusIcon(file.status)}
                <div className="min-w-0">
                  <div className={cn(
                    'font-medium text-sm truncate',
                    themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                  )}>{file.name}</div>
                  <div className={cn(
                    'text-xs truncate',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                  )}>{file.description}</div>
                </div>
              </div>
              <div className={cn(
                'text-right text-sm flex-shrink-0 ml-2',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>
                {formatSize(file.size)}
              </div>
            </div>
          ))}
          {providerFiles.length === 0 && (
            <div className={cn(
              'col-span-2 px-4 py-8 text-center text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              {t('ruleset.noProviders') || '暂无规则提供者'}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
