import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Download, Globe, Shield, Loader2, Check, X, Clock, Settings, Copy } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { 
  loadSingBoxTemplate, 
  defaultSingBoxRuleSets
} from '@/api/singboxTemplate'

interface RuleSet {
  tag: string
  type: 'remote' | 'local'
  format: 'source' | 'binary'
  url?: string
  path?: string
  downloadDetour?: string
  updateInterval?: string
  size: number
  updatedAt: string
  status: 'pending' | 'downloading' | 'completed' | 'failed'
  exists: boolean  // 本地文件是否存在
}

interface GeoResource {
  name: string
  type: 'geoip' | 'geosite'
  url: string
  path: string
  downloadDetour?: string
  updateInterval?: string
  size: number
  updatedAt: string
  status: 'pending' | 'downloading' | 'completed' | 'failed'
  exists: boolean  // 本地文件是否存在
}

interface SingBoxRuleConfig {
  autoUpdate: boolean
  updateInterval: number
  lastUpdate: string
  githubProxy: string      // 当前使用的 GitHub 代理
  githubProxies: string[]  // 可用的代理列表
  customProxies: string[]  // 用户自定义代理
  rulesetDir: string       // 实际存储路径
}

// API 基础路径
const API_BASE = '/api'

export default function SingBoxRulesetPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [geoResources, setGeoResources] = useState<GeoResource[]>([])
  const [ruleSets, setRuleSets] = useState<RuleSet[]>([])
  const [config, setConfig] = useState<SingBoxRuleConfig>({
    autoUpdate: true,
    updateInterval: 1,
    lastUpdate: '',
    githubProxy: '',
    githubProxies: [],
    customProxies: [],
    rulesetDir: ''
  })
  const [newProxy, setNewProxy] = useState('')  // 新代理输入
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState(false)
  const [saving, setSaving] = useState(false)
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null)

  // 复制 URL 并显示视觉反馈
  const copyUrl = (url: string) => {
    navigator.clipboard.writeText(url)
    setCopiedUrl(url)
    setTimeout(() => setCopiedUrl(null), 2000)
  }

  useEffect(() => {
    loadData()
    const interval = setInterval(checkStatus, 3000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const [geoRes, ruleSetRes, configRes] = await Promise.all([
        fetch(`${API_BASE}/singbox/ruleset/geo`).then(r => r.json()),
        fetch(`${API_BASE}/singbox/ruleset/rules`).then(r => r.json()),
        fetch(`${API_BASE}/singbox/ruleset/config`).then(r => r.json())
      ])
      setGeoResources(geoRes.data || getDefaultGeoResources())
      setRuleSets(ruleSetRes.data || await getRuleSetsFromTemplate())
      setConfig(configRes.data || { 
        autoUpdate: true, 
        updateInterval: 1, 
        lastUpdate: '', 
        githubProxy: '',
        githubProxies: [],
        customProxies: [],
        rulesetDir: ''
      })
    } catch {
      // 使用默认值
      setGeoResources(getDefaultGeoResources())
      const defaultRuleSets = await getRuleSetsFromTemplate()
      setRuleSets(defaultRuleSets)
    } finally {
      setLoading(false)
    }
  }

  const checkStatus = async () => {
    try {
      const res = await fetch(`${API_BASE}/singbox/ruleset/status`)
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

  // 全部更新 - 调用后端 API 执行下载
  const handleUpdateAll = async () => {
    try {
      setUpdating(true)
      
      // 调用后端 API，传递 GitHub 代理配置
      const response = await fetch(`${API_BASE}/singbox/ruleset/update`, { 
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          githubProxy: config.githubProxy || ''
        })
      })
      
      if (!response.ok) {
        throw new Error(t('singboxRuleset.updateFailed'))
      }
      
      // 开始轮询状态，获取后端下载进度
      const pollInterval = setInterval(async () => {
        await loadData()
      }, 2000)
      
      // 5分钟后停止轮询
      setTimeout(() => {
        clearInterval(pollInterval)
        setUpdating(false)
        loadData()
      }, 5 * 60 * 1000)
      
    } catch (err) {
      alert(t('singboxRuleset.startUpdateFailed') + ': ' + (err as Error).message)
      setUpdating(false)
      loadData()
    }
  }
  
  // 下载单个规则集 - 调用后端 API 执行下载
  const handleDownloadSingle = async (tag: string, isGeo: boolean) => {
    try {
      const response = await fetch(`${API_BASE}/singbox/ruleset/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          tag,
          isGeo,
          githubProxy: config.githubProxy || ''
        })
      })
      
      if (!response.ok) {
        throw new Error(t('singboxRuleset.downloadRequestFailed'))
      }
      
      // 轮询状态
      const pollStatus = setInterval(async () => {
        await loadData()
      }, 1000)
      
      // 2分钟后停止轮询
      setTimeout(() => {
        clearInterval(pollStatus)
        loadData()
      }, 2 * 60 * 1000)
      
    } catch (err) {
      alert(t('singboxRuleset.downloadFailed') + ': ' + (err as Error).message)
      loadData()
    }
  }

  const handleSaveConfig = async () => {
    try {
      setSaving(true)
      await fetch(`${API_BASE}/singbox/ruleset/config`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      alert(t('singboxRuleset.configSaved'))
    } catch {
      alert(t('singboxRuleset.saveFailed'))
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

  // 从配置模板同步规则集 (后备方案，实际数据从后端获取)
  const getRuleSetsFromTemplate = async (): Promise<RuleSet[]> => {
    // 优先从后端获取模板，否则使用默认模板
    const template = await loadSingBoxTemplate()
    const ruleSetList = template?.ruleSets || defaultSingBoxRuleSets
    
    return ruleSetList.map((rs: { tag: string; type: string; format: string; url?: string; path?: string }) => ({
      tag: rs.tag,
      type: rs.type as 'local' | 'remote',
      format: rs.format as 'binary' | 'source',
      url: rs.url || '',
      path: rs.path || '',  // 实际路径从后端获取
      size: 0,
      updatedAt: '',
      status: 'pending' as const,
      exists: false
    }))
  }

  // 默认 GEO 资源 (后备方案，实际数据从后端获取)
  const getDefaultGeoResources = (): GeoResource[] => [
    {
      name: 'geoip.db',
      type: 'geoip',
      url: 'https://github.com/SagerNet/sing-geoip/releases/latest/download/geoip.db',
      path: '',  // 实际路径从后端获取
      downloadDetour: 'direct',
      updateInterval: '7d',
      size: 0,
      updatedAt: '',
      status: 'pending',
      exists: false
    },
    {
      name: 'geosite.db',
      type: 'geosite',
      url: 'https://github.com/SagerNet/sing-geosite/releases/latest/download/geosite.db',
      path: '',  // 实际路径从后端获取
      downloadDetour: 'direct',
      updateInterval: '7d',
      size: 0,
      updatedAt: '',
      status: 'pending',
      exists: false
    }
  ]

  const completedGeo = geoResources.filter(f => f.status === 'completed').length
  const completedRules = ruleSets.filter(f => f.status === 'completed').length
  const totalCount = geoResources.length + ruleSets.length
  const completedCount = completedGeo + completedRules

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    )
  }

  // 计算下载中的数量
  const downloadingCount = [...geoResources, ...ruleSets].filter(
    item => item.status === 'downloading'
  ).length
  const failedCount = [...geoResources, ...ruleSets].filter(
    item => item.status === 'failed'
  ).length
  const progressPercent = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('singboxRuleset.title')}</h2>
          <p className={cn(
            'text-sm mt-1',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>
            {t('singboxRuleset.downloaded')} {completedCount}/{totalCount} {t('singboxRuleset.ruleFiles')}
            {downloadingCount > 0 && ` · ${t('singboxRuleset.downloading')} ${downloadingCount}`}
            {failedCount > 0 && ` · ${failedCount} ${t('singboxRuleset.failed')}`}
          </p>
          {/* 进度条 */}
          {updating && (
            <div className="mt-2 w-64">
              <div className={cn(
                'h-1.5 rounded-full overflow-hidden',
                themeStyle === 'apple-glass' ? 'bg-slate-200' : 'bg-slate-700'
              )}>
                <div 
                  className="h-full bg-gradient-to-r from-cyan-500 to-blue-500 transition-all duration-300 relative"
                  style={{ width: `${Math.max(progressPercent, 5)}%` }}
                >
                  <div className="absolute inset-0 bg-white/30 animate-pulse" />
                </div>
              </div>
            </div>
          )}
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
          {updating ? t('singboxRuleset.updating') : t('singboxRuleset.updateAll')}
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
          )}>{t('singboxRuleset.updateSettings')}</span>
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
              'text-sm whitespace-nowrap',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('singboxRuleset.autoUpdate')}</span>
          </label>
          <div className="flex items-center gap-2">
            <span className={cn(
              'text-sm whitespace-nowrap',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>{t('singboxRuleset.updateInterval')}:</span>
            <select
              value={config.updateInterval}
              onChange={(e) => setConfig({ ...config, updateInterval: Number(e.target.value) })}
              className="form-input text-sm py-1.5 w-24"
            >
              <option value={1}>1 {t('singboxRuleset.days')}</option>
              <option value={2}>2 {t('singboxRuleset.days')}</option>
              <option value={3}>3 {t('singboxRuleset.days')}</option>
              <option value={5}>5 {t('singboxRuleset.days')}</option>
              <option value={7}>7 {t('singboxRuleset.days')}</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <span className={cn(
              'text-sm whitespace-nowrap',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>{t('singboxRuleset.githubProxy')}:</span>
            {config.githubProxy === '__custom__' ? (
              // 自定义代理输入模式 - 替换下拉框
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
                  {t('singboxRuleset.confirm')}
                </button>
                <button
                  onClick={() => {
                    setConfig({ ...config, githubProxy: '' })
                    setNewProxy('')
                  }}
                  className="control-btn secondary text-xs px-2 whitespace-nowrap"
                >
                  {t('singboxRuleset.cancel')}
                </button>
              </div>
            ) : (
              // 下拉选择模式
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
                <option value="">{t('singboxRuleset.directConnect')}</option>
                {config.githubProxies?.filter(p => p).map(proxy => (
                  <option key={proxy} value={proxy}>{proxy.replace('https://', '')}</option>
                ))}
                {config.customProxies?.map(proxy => (
                  <option key={proxy} value={proxy}>★ {proxy.replace('https://', '')}</option>
                ))}
                <option value="__custom__">{t('singboxRuleset.addCustom')}</option>
              </select>
            )}
          </div>
          <button
            onClick={handleSaveConfig}
            disabled={saving}
            className="control-btn secondary text-xs whitespace-nowrap"
          >
            {saving ? <Loader2 className="w-3 h-3 animate-spin" /> : null}
            {t('singboxRuleset.saveSettings')}
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
          )}>{t('singboxRuleset.geoDatabase')}</span>
          <span className={cn(
            'text-xs',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>({geoResources.length})</span>
        </div>
        <div className={cn(
          'divide-y',
          themeStyle === 'apple-glass' ? 'divide-black/5' : 'divide-white/5'
        )}>
          {geoResources.map((file) => (
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
                  <div 
                    className={cn(
                      'text-xs truncate max-w-md cursor-pointer transition-all duration-200',
                      copiedUrl === file.url 
                        ? 'text-green-500 font-medium' 
                        : themeStyle === 'apple-glass' 
                          ? 'text-slate-500 hover:text-blue-500' 
                          : 'text-slate-400 hover:text-blue-400'
                    )}
                    onClick={() => copyUrl(file.url)}
                    title="点击复制 URL"
                  >
                    {copiedUrl === file.url ? '✓ 已复制!' : file.url}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className={cn(
                  'text-right text-sm',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )}>
                  <div>{formatSize(file.size)}</div>
                  {file.updatedAt && (
                    <div className="text-xs">{new Date(file.updatedAt).toLocaleDateString()}</div>
                  )}
                </div>
                <button
                  onClick={() => handleDownloadSingle(file.name, true)}
                  disabled={file.status === 'downloading'}
                  className={cn(
                    'p-1.5 rounded-lg transition-all',
                    file.status === 'downloading' 
                      ? 'opacity-50 cursor-not-allowed' 
                      : 'hover:bg-blue-500/20 text-blue-500'
                  )}
                  title={t('singboxRuleset.downloadUpdate')}
                >
                  {file.status === 'downloading' ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <Download className="w-4 h-4" />
                  )}
                </button>
              </div>
            </div>
          ))}
          {geoResources.length === 0 && (
            <div className={cn(
              'px-4 py-8 text-center text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              {t('singboxRuleset.noGeoFiles')}
            </div>
          )}
        </div>
      </div>

      {/* 规则集 */}
      <div className="glass-card overflow-hidden">
        <div className={cn(
          'px-4 py-3 border-b flex items-center justify-between',
          themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
        )}>
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-green-500" />
            <span className={cn(
              'font-medium',
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{t('singboxRuleset.ruleSet')}</span>
            <span className={cn(
              'text-xs',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>({ruleSets.length})</span>
          </div>
        </div>
        <div className={cn(
          'grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0',
          themeStyle === 'apple-glass' ? 'divide-black/5' : 'divide-white/5'
        )}>
          {ruleSets.map((ruleSet, index) => (
            <div 
              key={ruleSet.tag} 
              className={cn(
                'px-4 py-3 flex items-center justify-between',
                themeStyle === 'apple-glass' ? 'hover:bg-black/5' : 'hover:bg-white/5',
                index % 2 === 0 && themeStyle === 'apple-glass' ? 'md:border-r md:border-black/5' : '',
                index % 2 === 0 && themeStyle !== 'apple-glass' ? 'md:border-r md:border-white/5' : '',
                index >= 2 && themeStyle === 'apple-glass' ? 'md:border-t md:border-black/5' : '',
                index >= 2 && themeStyle !== 'apple-glass' ? 'md:border-t md:border-white/5' : ''
              )}
            >
              <div className="flex items-center gap-3 min-w-0 flex-1">
                {getStatusIcon(ruleSet.status)}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <div className={cn(
                      'font-medium text-sm truncate',
                      themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                    )}>{ruleSet.tag}</div>
                    {ruleSet.url && (
                      <button
                        onClick={() => copyUrl(ruleSet.url!)}
                        className={cn(
                          'p-1 rounded transition-all',
                          copiedUrl === ruleSet.url
                            ? 'text-green-500'
                            : themeStyle === 'apple-glass'
                              ? 'text-slate-400 hover:text-blue-500 hover:bg-black/5'
                              : 'text-slate-500 hover:text-blue-400 hover:bg-white/10'
                        )}
                        title="复制 URL"
                      >
                        {copiedUrl === ruleSet.url ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                      </button>
                    )}
                  </div>
                  <div className={cn(
                    'text-xs',
                    themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                  )}>
                    {ruleSet.format === 'binary' ? t('singboxRuleset.srsFormat') : t('singboxRuleset.jsonFormat')}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2 flex-shrink-0">
                <div className={cn(
                  'text-right text-sm',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )}>
                  {formatSize(ruleSet.size)}
                </div>
                <button
                  onClick={() => handleDownloadSingle(ruleSet.tag, false)}
                  disabled={ruleSet.status === 'downloading' || !ruleSet.url}
                  className={cn(
                    'p-1 rounded transition-all',
                    ruleSet.status === 'downloading' || !ruleSet.url
                      ? 'opacity-50 cursor-not-allowed' 
                      : 'hover:bg-green-500/20 text-green-500'
                  )}
                  title={t('singboxRuleset.downloadUpdate')}
                >
                  {ruleSet.status === 'downloading' ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Download className="w-3.5 h-3.5" />
                  )}
                </button>
              </div>
            </div>
          ))}
          {ruleSets.length === 0 && (
            <div className={cn(
              'col-span-2 px-4 py-8 text-center text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              {t('singboxRuleset.noRuleSets')}
            </div>
          )}
        </div>
      </div>

      {/* 说明 */}
      <div className={cn(
        'text-xs p-4 rounded-lg',
        themeStyle === 'apple-glass' ? 'bg-blue-50 text-blue-700' : 'bg-blue-500/10 text-blue-400'
      )}>
        <p className="font-medium mb-2">{t('singboxRuleset.description')}:</p>
        <ul className="list-disc list-inside space-y-1">
          <li>{t('singboxRuleset.descSync')}</li>
          <li><strong>{t('singboxRuleset.descPath')}</strong>: {config.rulesetDir || t('singboxRuleset.loading')}</li>
          <li>{t('singboxRuleset.descGeo')}</li>
          <li>{t('singboxRuleset.descRuleSet')}</li>
          <li>{t('singboxRuleset.descGenerate')}</li>
          <li>{t('singboxRuleset.descUpdateAll')}</li>
        </ul>
      </div>
    </div>
  )
}
