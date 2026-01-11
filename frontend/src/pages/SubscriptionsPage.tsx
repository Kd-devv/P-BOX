import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, RefreshCw, Trash2, Edit2, X, Loader2, Link, ChevronDown } from 'lucide-react'
import { subscriptionApi, Subscription, AddSubscriptionRequest } from '@/api/subscription'
import { cn, formatBytes } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

// 默认过滤关键词列表
const DEFAULT_KEYWORDS = [
  '过期时间', '剩余流量', 'QQ群', '官网', '到期', '公告', '更新', '距离',
  '套餐', '流量', '过期', '剩余', '倍率', '异常', 'QQ', '官方'
]

// 请求头预设模板
const HEADER_PRESETS: Record<string, { name: string; headers: Record<string, string> }> = {
  'default': { name: '默认', headers: { 'User-Agent': 'P-BOX/1.0' } },
  'clash': { name: 'Clash', headers: { 'User-Agent': 'clash.meta' } },
  'v2rayn': { name: 'v2rayN', headers: { 'User-Agent': 'v2rayN' } },
  'quantumult': { name: 'Quantumult X', headers: { 'User-Agent': 'Quantumult%20X' } },
}

export default function SubscriptionsPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingSub, setEditingSub] = useState<Subscription | null>(null)
  const [showAdvanced, setShowAdvanced] = useState(false)

  // Form state
  const [formData, setFormData] = useState<AddSubscriptionRequest>({
    name: '',
    url: '',
    autoUpdate: true,
    updateInterval: 86400,
    filterKeywords: [],
    filterMode: 'exclude',
    customHeaders: HEADER_PRESETS['default'].headers,
  })
  const [keywords, setKeywords] = useState<string[]>([])
  const [newKeyword, setNewKeyword] = useState('')
  const [headerPreset, setHeaderPreset] = useState('default')

  useEffect(() => {
    fetchSubscriptions()
  }, [])

  const fetchSubscriptions = async () => {
    try {
      setLoading(true)
      const data = await subscriptionApi.list()
      setSubscriptions(data || [])
    } catch {
      // Ignore errors
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      url: '',
      autoUpdate: true,
      updateInterval: 86400,
      filterKeywords: [],
      filterMode: 'exclude',
      customHeaders: HEADER_PRESETS['default'].headers,
    })
    setKeywords([])
    setNewKeyword('')
    setHeaderPreset('default')
    setShowAdvanced(false)
  }

  const addKeyword = (kw: string) => {
    if (kw && !keywords.includes(kw)) {
      setKeywords([...keywords, kw])
    }
    setNewKeyword('')
  }

  const removeKeyword = (kw: string) => {
    setKeywords(keywords.filter(k => k !== kw))
  }

  const useDefaultKeywords = () => {
    setKeywords([...new Set([...keywords, ...DEFAULT_KEYWORDS])])
  }

  const handlePresetChange = (preset: string) => {
    setHeaderPreset(preset)
    setFormData({ ...formData, customHeaders: HEADER_PRESETS[preset]?.headers || {} })
  }

  const handleAdd = async () => {
    if (!formData.name || !formData.url) return
    try {
      await subscriptionApi.add({ ...formData, filterKeywords: keywords })
      setShowAddModal(false)
      resetForm()
      await fetchSubscriptions()
    } catch {
      // Ignore errors
    }
  }

  const handleUpdate = async () => {
    if (!editingSub || !formData.name || !formData.url) return
    try {
      await subscriptionApi.update(editingSub.id, { ...formData, filterKeywords: keywords })
      setEditingSub(null)
      resetForm()
      await fetchSubscriptions()
    } catch {
      // Ignore errors
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('common.confirm') + '?')) return
    try {
      await subscriptionApi.delete(id)
      await fetchSubscriptions()
    } catch {
      // Ignore errors
    }
  }

  const handleRefresh = async (id: string) => {
    try {
      setRefreshing(id)
      await subscriptionApi.refresh(id)
      await fetchSubscriptions()
    } catch {
      // Ignore errors
    } finally {
      setRefreshing(null)
    }
  }

  const handleRefreshAll = async () => {
    try {
      setRefreshing('all')
      await subscriptionApi.refreshAll()
      await fetchSubscriptions()
    } catch {
      // Ignore errors
    } finally {
      setRefreshing(null)
    }
  }

  const openEditModal = (sub: Subscription) => {
    setEditingSub(sub)
    setFormData({
      name: sub.name,
      url: sub.url,
      autoUpdate: sub.autoUpdate,
      updateInterval: sub.updateInterval,
      filterKeywords: sub.filterKeywords || [],
      filterMode: sub.filterMode || 'exclude',
      customHeaders: sub.customHeaders || HEADER_PRESETS['default'].headers,
    })
    // 加载过滤关键词
    setKeywords(sub.filterKeywords || [])
    // 检测请求头预设
    const currentHeaders = sub.customHeaders || {}
    const matchedPreset = Object.entries(HEADER_PRESETS).find(([, preset]) => 
      JSON.stringify(preset.headers) === JSON.stringify(currentHeaders)
    )
    setHeaderPreset(matchedPreset ? matchedPreset[0] : 'default')
    // 如果有自定义设置，展开高级选项
    if ((sub.filterKeywords && sub.filterKeywords.length > 0) || sub.customHeaders) {
      setShowAdvanced(true)
    }
  }

  const formatExpireTime = (expireTime?: string) => {
    if (!expireTime) return t('common.unknown')
    const date = new Date(expireTime)
    const now = new Date()
    const diff = date.getTime() - now.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    if (days < 0) return t('subscriptions.expired')
    return `${days} ${t('subscriptions.daysLeft')}`
  }

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
        <h2 className={cn(
          'text-lg font-semibold',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>{t('subscriptions.title')}</h2>
        <div className="flex gap-2">
          <button
            onClick={handleRefreshAll}
            disabled={refreshing === 'all'}
            className="control-btn secondary text-xs"
          >
            <RefreshCw className={cn('w-3 h-3', refreshing === 'all' && 'animate-spin')} />
            {t('common.refresh')}
          </button>
          <button
            onClick={() => setShowAddModal(true)}
            className="control-btn primary text-xs"
          >
            <Plus className="w-3 h-3" />
            {t('subscriptions.addSubscription')}
          </button>
        </div>
      </div>

      {/* Subscriptions Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {subscriptions.map((sub) => (
          <div
            key={sub.id}
            className={cn(
              'glass-card p-5 relative group',
              'border-l-2',
              sub.expireTime && new Date(sub.expireTime) < new Date()
                ? 'border-l-red-500'
                : 'border-l-green-500'
            )}
          >
            {/* Actions */}
            <div className="absolute top-3 right-3 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                onClick={() => handleRefresh(sub.id)}
                disabled={refreshing === sub.id}
                className="p-1.5 rounded hover:bg-white/10"
              >
                <RefreshCw className={cn(
                  'w-3.5 h-3.5',
                  refreshing === sub.id && 'animate-spin',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )} />
              </button>
              <button
                onClick={() => openEditModal(sub)}
                className="p-1.5 rounded hover:bg-white/10"
              >
                <Edit2 className={cn(
                  'w-3.5 h-3.5',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )} />
              </button>
              <button
                onClick={() => handleDelete(sub.id)}
                className="p-1.5 rounded hover:bg-red-500/20"
              >
                <Trash2 className="w-3.5 h-3.5 text-red-400" />
              </button>
            </div>

            {/* Content */}
            <div className="flex items-start gap-3 mb-4">
              <div className="app-icon purple w-10 h-10 rounded-lg">
                <Link className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <h3 className={cn(
                  'font-semibold truncate',
                  themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                )}>{sub.name}</h3>
                <p className={cn(
                  'text-xs truncate',
                  themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                )}>{sub.url}</p>
              </div>
            </div>

            {/* Stats */}
            <div className={cn(
              'space-y-2 text-xs',
              themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
            )}>
              <div className="flex justify-between">
                <span>{t('subscriptions.nodes')}:</span>
                <span className="font-mono">{sub.filteredNodeCount} / {sub.nodeCount}</span>
              </div>
              <div className="flex justify-between items-center">
                <span>{t('subscriptions.status')}:</span>
                {sub.lastUpdateStatus === 'failed' ? (
                  <span className="text-red-400 flex items-center gap-1" title={sub.lastError}>
                    <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
                    {t('subscriptions.updateFailed')}
                  </span>
                ) : sub.lastUpdateStatus === 'success' ? (
                  <span className="text-green-400 flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                    {t('subscriptions.updateSuccess')}
                  </span>
                ) : (
                  <span className="text-slate-400">
                    {t('subscriptions.notUpdated')}
                  </span>
                )}
              </div>
              {sub.traffic && (
                <div className="flex justify-between">
                  <span>{t('subscriptions.usage')}:</span>
                  <span className="font-mono">
                    {formatBytes(sub.traffic.download)} / {formatBytes(sub.traffic.total)}
                  </span>
                </div>
              )}
              {sub.traffic && (
                <div className="w-full bg-slate-800/50 h-1.5 rounded-full overflow-hidden">
                  <div 
                    className="bg-purple-500 h-full"
                    style={{ width: `${(sub.traffic.download / sub.traffic.total) * 100}%` }}
                  />
                </div>
              )}
              <div className="flex justify-between">
                <span>{t('subscriptions.expire')}:</span>
                <span className={cn(
                  'font-mono',
                  sub.expireTime && new Date(sub.expireTime) < new Date() ? 'text-red-400' : ''
                )}>
                  {formatExpireTime(sub.expireTime)}
                </span>
              </div>
            </div>
          </div>
        ))}

        {/* Add Card */}
        <button
          onClick={() => setShowAddModal(true)}
          className={cn(
            'glass-card p-5 border-2 border-dashed flex flex-col items-center justify-center gap-3 min-h-[180px] transition-all',
            themeStyle === 'apple-glass'
              ? 'border-black/10 hover:border-blue-500/50 hover:bg-blue-50/50'
              : 'border-white/10 hover:border-indigo-500/50 hover:bg-indigo-500/5'
          )}
        >
          <Plus className={cn(
            'w-8 h-8',
            themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
          )} />
          <span className={cn(
            'text-sm',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>{t('subscriptions.addSubscription')}</span>
        </button>
      </div>

      {/* Add/Edit Modal */}
      {(showAddModal || editingSub) && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className={cn(
            'w-full max-w-md rounded-2xl p-6',
            themeStyle === 'apple-glass'
              ? 'bg-white/90 backdrop-blur-xl'
              : 'bg-neutral-900 border border-white/10'
          )}>
            <div className="flex items-center justify-between mb-6">
              <h3 className={cn(
                'text-lg font-semibold',
                themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
              )}>
                {editingSub ? t('common.edit') : t('subscriptions.addSubscription')}
              </h3>
              <button
                onClick={() => {
                  setShowAddModal(false)
                  setEditingSub(null)
                  resetForm()
                }}
                className="p-2 rounded-lg hover:bg-white/10"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className={cn(
                  'block text-sm mb-2',
                  themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
                )}>{t('subscriptions.name')}</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="form-input"
                  placeholder="My Subscription"
                />
              </div>
              <div>
                <label className={cn(
                  'block text-sm mb-2',
                  themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
                )}>{t('subscriptions.url')}</label>
                <input
                  type="text"
                  value={formData.url}
                  onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                  className="form-input"
                  placeholder="https://..."
                />
              </div>
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="autoUpdate"
                  checked={formData.autoUpdate}
                  onChange={(e) => setFormData({ ...formData, autoUpdate: e.target.checked })}
                  className="w-4 h-4"
                />
                <label htmlFor="autoUpdate" className={cn(
                  'text-sm',
                  themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
                )}>
                  {t('subscriptions.updateInterval')} (24h)
                </label>
              </div>

              {/* 高级选项折叠按钮 */}
              <button
                type="button"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className={cn(
                  'flex items-center gap-2 text-sm',
                  themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
                )}
              >
                <ChevronDown className={cn('w-4 h-4 transition-transform', showAdvanced && 'rotate-180')} />
                {t('subscriptions.advancedOptions') || '高级选项'}
              </button>

              {/* 高级选项内容 */}
              {showAdvanced && (
                <div className={cn(
                  'space-y-4 p-4 rounded-lg',
                  themeStyle === 'apple-glass' ? 'bg-black/5' : 'bg-white/5'
                )}>
                  {/* 过滤关键词 */}
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <label className={cn(
                        'text-sm font-medium',
                        themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
                      )}>{t('subscriptions.filterKeywords') || '过滤关键词'}</label>
                      <button
                        type="button"
                        onClick={useDefaultKeywords}
                        className={cn(
                          'text-xs hover:underline',
                          themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
                        )}
                      >
                        {t('subscriptions.useDefault') || '使用默认关键词'}
                      </button>
                    </div>
                    
                    {/* 已添加的关键词 */}
                    {keywords.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mb-2">
                        {keywords.map(kw => (
                          <span key={kw} className={cn(
                            'inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs',
                            themeStyle === 'apple-glass' ? 'bg-slate-200 text-slate-700' : 'bg-white/10 text-slate-300'
                          )}>
                            {kw}
                            <button type="button" onClick={() => removeKeyword(kw)} className="hover:text-red-400">
                              <X className="w-3 h-3" />
                            </button>
                          </span>
                        ))}
                      </div>
                    )}
                    
                    {/* 添加关键词输入 */}
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={newKeyword}
                        onChange={e => setNewKeyword(e.target.value)}
                        onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addKeyword(newKeyword))}
                        placeholder={t('subscriptions.filterKeywordsHint') || '输入关键词后回车添加'}
                        className="form-input flex-1 text-sm"
                      />
                      <button
                        type="button"
                        onClick={() => addKeyword(newKeyword)}
                        className="control-btn secondary text-xs"
                      >
                        {t('common.add') || '添加'}
                      </button>
                    </div>

                    {/* 过滤模式 */}
                    <div className="flex gap-4 mt-2">
                      <label className="flex items-center gap-2 text-xs cursor-pointer">
                        <input
                          type="radio"
                          name="filterMode"
                          checked={formData.filterMode === 'exclude'}
                          onChange={() => setFormData({ ...formData, filterMode: 'exclude' })}
                        />
                        <span className={themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'}>
                          {t('subscriptions.filterExclude') || '排除包含关键词的节点'}
                        </span>
                      </label>
                      <label className="flex items-center gap-2 text-xs cursor-pointer">
                        <input
                          type="radio"
                          name="filterMode"
                          checked={formData.filterMode === 'include'}
                          onChange={() => setFormData({ ...formData, filterMode: 'include' })}
                        />
                        <span className={themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'}>
                          {t('subscriptions.filterInclude') || '仅保留包含关键词的节点'}
                        </span>
                      </label>
                    </div>
                  </div>

                  {/* 请求头预设 */}
                  <div>
                    <label className={cn(
                      'block text-sm font-medium mb-2',
                      themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
                    )}>{t('subscriptions.customHeaders') || '请求头预设'}</label>
                    <select
                      value={headerPreset}
                      onChange={e => handlePresetChange(e.target.value)}
                      className="form-input text-sm"
                    >
                      {Object.entries(HEADER_PRESETS).map(([key, val]) => (
                        <option key={key} value={key}>{val.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
              )}
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setShowAddModal(false)
                  setEditingSub(null)
                  resetForm()
                }}
                className="flex-1 py-2.5 rounded-lg border border-white/10 text-sm"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={editingSub ? handleUpdate : handleAdd}
                className={cn(
                  'flex-1 py-2.5 rounded-lg text-sm font-medium text-white',
                  themeStyle === 'apple-glass'
                    ? 'bg-blue-500 hover:bg-blue-600'
                    : 'bg-indigo-500 hover:bg-indigo-600'
                )}
              >
                {editingSub ? t('common.save') : t('common.add')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
