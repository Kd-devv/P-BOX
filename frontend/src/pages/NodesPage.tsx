import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, RefreshCw, Trash2, Zap, Copy, Loader2, Globe, Server, Search, X, Filter, ChevronDown, Link } from 'lucide-react'
import { nodeApi, Node } from '@/api/node'
import { subscriptionApi, Subscription } from '@/api/subscription'
import { cn, getLatencyColor, formatLatency } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { AddNodeDialog } from '@/components/AddNodeDialog'

// 国家/地区关键词映射
const COUNTRY_KEYWORDS: Record<string, string[]> = {
  '香港': ['香港', 'HK', 'Hong Kong', 'HongKong'],
  '台湾': ['台湾', 'TW', 'Taiwan'],
  '日本': ['日本', 'JP', 'Japan', '东京', 'Tokyo'],
  '新加坡': ['新加坡', 'SG', 'Singapore'],
  '美国': ['美国', 'US', 'USA', 'United States', '洛杉矶', '硅谷'],
  '韩国': ['韩国', 'KR', 'Korea', '首尔'],
  '英国': ['英国', 'UK', 'Britain', '伦敦'],
  '德国': ['德国', 'DE', 'Germany'],
  '澳大利亚': ['澳大利亚', 'AU', 'Australia'],
  '加拿大': ['加拿大', 'CA', 'Canada'],
}

// 检测节点国家
function detectCountry(nodeName: string): string {
  for (const [country, keywords] of Object.entries(COUNTRY_KEYWORDS)) {
    if (keywords.some(k => nodeName.toLowerCase().includes(k.toLowerCase()))) {
      return country
    }
  }
  return '其他'
}

export default function NodesPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [nodes, setNodes] = useState<Node[]>([])
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [loading, setLoading] = useState(true)
  const [testing, setTesting] = useState<string | null>(null)
  const [showImportModal, setShowImportModal] = useState(false)
  const [showAddNodeDialog, setShowAddNodeDialog] = useState(false)
  const [importUrl, setImportUrl] = useState('')
  
  // 搜索和过滤状态
  const [searchTerm, setSearchTerm] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [filters, setFilters] = useState({
    source: 'all' as 'all' | 'manual' | string,
    protocol: 'all' as string,
    country: 'all' as string,
    delay: 'all' as 'all' | 'fast' | 'medium' | 'slow' | 'timeout' | 'untested',
  })

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      setLoading(true)
      const [nodesData, subsData] = await Promise.all([
        nodeApi.list(),
        subscriptionApi.list()
      ])
      setNodes(nodesData || [])
      setSubscriptions(subsData || [])
    } catch {
      // Ignore errors
    } finally {
      setLoading(false)
    }
  }

  const handleImport = async () => {
    if (!importUrl) return
    try {
      await nodeApi.importUrl(importUrl)
      setShowImportModal(false)
      setImportUrl('')
      await fetchData()
    } catch {
      // Ignore errors
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('common.confirm') + '?')) return
    try {
      await nodeApi.delete(id)
      await fetchData()
    } catch {
      // Ignore errors
    }
  }

  const handleTest = async (node: Node) => {
    try {
      setTesting(node.id)
      const result = await nodeApi.testDelay(node.id, node.server, node.serverPort)
      setNodes(prev => prev.map(n => 
        n.id === node.id ? { ...n, delay: result.delay } : n
      ))
    } catch {
      setNodes(prev => prev.map(n => 
        n.id === node.id ? { ...n, delay: 0 } : n
      ))
    } finally {
      setTesting(null)
    }
  }

  const handleTestAll = async () => {
    const nodeIds = nodes.map(n => n.id)
    try {
      setTesting('all')
      const results = await nodeApi.testDelayBatch(nodeIds)
      setNodes(prev => prev.map(n => ({
        ...n,
        delay: results[n.id] ?? n.delay
      })))
    } catch {
      // Ignore errors
    } finally {
      setTesting(null)
    }
  }

  const handleCopyShare = async (id: string) => {
    try {
      const result = await nodeApi.getShareUrl(id)
      await navigator.clipboard.writeText(result.url)
    } catch {
      // Ignore errors
    }
  }

  // 获取可用的过滤选项
  const filterOptions = useMemo(() => {
    const protocols = [...new Set(nodes.map(n => n.type.toUpperCase()))]
    const countries = [...new Set(nodes.map(n => detectCountry(n.name)))]
    return { protocols, countries }
  }, [nodes])

  // 过滤节点
  const filteredNodes = useMemo(() => {
    return nodes.filter(node => {
      // 搜索过滤
      const matchSearch = !searchTerm || 
        node.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        node.server.toLowerCase().includes(searchTerm.toLowerCase()) ||
        node.type.toLowerCase().includes(searchTerm.toLowerCase())
      
      if (!matchSearch) return false

      // 来源过滤
      if (filters.source !== 'all') {
        if (filters.source === 'manual' && !node.isManual) return false
        if (filters.source !== 'manual' && node.subscriptionId !== filters.source) return false
      }

      // 协议过滤
      if (filters.protocol !== 'all' && node.type.toUpperCase() !== filters.protocol) {
        return false
      }

      // 国家过滤
      if (filters.country !== 'all' && detectCountry(node.name) !== filters.country) {
        return false
      }

      // 延迟过滤
      const delay = node.delay
      if (filters.delay !== 'all') {
        if (filters.delay === 'untested' && delay !== -1) return false
        if (filters.delay === 'fast' && (delay <= 0 || delay >= 100)) return false
        if (filters.delay === 'medium' && (delay < 100 || delay >= 200)) return false
        if (filters.delay === 'slow' && (delay < 200 || delay === 0)) return false
        if (filters.delay === 'timeout' && delay !== 0) return false
      }

      return true
    })
  }, [nodes, searchTerm, filters])

  // Group nodes by subscription
  const manualNodes = filteredNodes.filter(n => n.isManual)
  const subscriptionNodes = filteredNodes.filter(n => !n.isManual)

  const hasActiveFilters = searchTerm || filters.source !== 'all' || 
    filters.protocol !== 'all' || filters.country !== 'all' || filters.delay !== 'all'

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
        )}>{t('nav.nodes')}</h2>
        <div className="flex gap-2">
          <button
            onClick={handleTestAll}
            disabled={testing === 'all'}
            className="control-btn secondary text-xs"
          >
            <Zap className={cn('w-3 h-3', testing === 'all' && 'animate-pulse')} />
            {t('proxy.testAll')}
          </button>
          <button
            onClick={() => setShowImportModal(true)}
            className="control-btn secondary text-xs"
          >
            <Link className="w-3 h-3" />
            {t('nodes.importLink')}
          </button>
          <button
            onClick={() => setShowAddNodeDialog(true)}
            className="control-btn primary text-xs"
          >
            <Plus className="w-3 h-3" />
            {t('common.add')}
          </button>
        </div>
      </div>

      {/* 添加节点对话框 */}
      <AddNodeDialog
        open={showAddNodeDialog}
        onOpenChange={setShowAddNodeDialog}
        onSuccess={fetchData}
      />

      {/* 搜索和过滤 */}
      <div className="space-y-3">
        <div className="flex gap-2">
          {/* 搜索框 */}
          <div className="relative flex-1">
            <Search className={cn(
              'absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none z-10',
              themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
            )} />
            <input
              type="text"
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              placeholder={t('common.search')}
              className="form-input !pl-10 pr-10"
            />
            {searchTerm && (
              <button
                onClick={() => setSearchTerm('')}
                className="absolute right-3 top-1/2 -translate-y-1/2"
              >
                <X className={cn(
                  'w-4 h-4',
                  themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                )} />
              </button>
            )}
          </div>
          {/* 过滤按钮 */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              'px-4 py-2 rounded-lg border transition-colors flex items-center gap-2',
              showFilters 
                ? themeStyle === 'apple-glass'
                  ? 'bg-blue-500 text-white border-blue-500'
                  : 'bg-cyan-500 text-white border-cyan-500'
                : themeStyle === 'apple-glass'
                  ? 'border-black/10 hover:bg-black/5'
                  : 'border-white/10 hover:bg-white/5'
            )}
          >
            <Filter className="w-4 h-4" />
            <span className="hidden sm:inline">{t('common.filter') || '过滤'}</span>
            <ChevronDown className={cn('w-4 h-4 transition-transform', showFilters && 'rotate-180')} />
          </button>
        </div>

        {/* 过滤选项 */}
        {showFilters && (
          <div className={cn(
            'grid grid-cols-2 md:grid-cols-4 gap-3 p-4 rounded-lg border',
            themeStyle === 'apple-glass'
              ? 'bg-white/50 border-black/10'
              : 'bg-white/5 border-white/10'
          )}>
            {/* 来源 */}
            <div>
              <label className={cn(
                'text-xs font-medium mb-1 block',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>{t('nodes.source') || '来源'}</label>
              <select
                value={filters.source}
                onChange={e => setFilters({ ...filters, source: e.target.value })}
                className="form-input text-sm py-1.5"
              >
                <option value="all">{t('common.all') || '全部'}</option>
                <option value="manual">{t('nodes.manual') || '手动添加'}</option>
                {subscriptions.map(sub => (
                  <option key={sub.id} value={sub.id}>{sub.name}</option>
                ))}
              </select>
            </div>

            {/* 协议 */}
            <div>
              <label className={cn(
                'text-xs font-medium mb-1 block',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>{t('nodes.protocol') || '协议'}</label>
              <select
                value={filters.protocol}
                onChange={e => setFilters({ ...filters, protocol: e.target.value })}
                className="form-input text-sm py-1.5"
              >
                <option value="all">{t('common.all') || '全部'}</option>
                {filterOptions.protocols.map(p => (
                  <option key={p} value={p}>{p}</option>
                ))}
              </select>
            </div>

            {/* 国家/地区 */}
            <div>
              <label className={cn(
                'text-xs font-medium mb-1 block',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>{t('nodes.country') || '国家/地区'}</label>
              <select
                value={filters.country}
                onChange={e => setFilters({ ...filters, country: e.target.value })}
                className="form-input text-sm py-1.5"
              >
                <option value="all">{t('common.all') || '全部'}</option>
                {filterOptions.countries.map(c => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>

            {/* 延迟 */}
            <div>
              <label className={cn(
                'text-xs font-medium mb-1 block',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>{t('nodes.delay') || '延迟'}</label>
              <select
                value={filters.delay}
                onChange={e => setFilters({ ...filters, delay: e.target.value as typeof filters.delay })}
                className="form-input text-sm py-1.5"
              >
                <option value="all">{t('common.all') || '全部'}</option>
                <option value="fast">快速 (&lt;100ms)</option>
                <option value="medium">中等 (100-200ms)</option>
                <option value="slow">较慢 (&gt;200ms)</option>
                <option value="timeout">超时</option>
                <option value="untested">未测试</option>
              </select>
            </div>

            {/* 重置按钮 */}
            <div className="col-span-2 md:col-span-4 flex justify-end">
              <button
                onClick={() => setFilters({ source: 'all', protocol: 'all', country: 'all', delay: 'all' })}
                className={cn(
                  'text-sm hover:underline',
                  themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
                )}
              >
                {t('common.reset') || '重置过滤'}
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Stats */}
      <div className={cn(
        'text-sm',
        themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-400'
      )}>
        共 {filteredNodes.length} 个节点
        {hasActiveFilters && ` (已过滤，总共 ${nodes.length} 个)`}
        <span className="mx-3">|</span>
        <span className="text-green-500">{filteredNodes.filter(n => n.delay > 0).length}</span> 在线
        <span className="mx-1">·</span>
        <span className="text-red-500">{filteredNodes.filter(n => n.delay === 0).length}</span> 超时
      </div>

      {/* Manual Nodes */}
      {manualNodes.length > 0 && (
        <div className="glass-card p-4">
          <h3 className={cn(
            'text-sm font-medium mb-3 flex items-center gap-2',
            themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
          )}>
            <Server className="w-4 h-4" />
            {t('nodes.manual')} ({manualNodes.length})
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {manualNodes.map(node => (
              <NodeCard 
                key={node.id}
                node={node}
                testing={testing === node.id}
                themeStyle={themeStyle}
                onTest={() => handleTest(node)}
                onDelete={() => handleDelete(node.id)}
                onCopy={() => handleCopyShare(node.id)}
              />
            ))}
          </div>
        </div>
      )}

      {/* Subscription Nodes */}
      <div className="glass-card p-4">
        <h3 className={cn(
          'text-sm font-medium mb-3 flex items-center gap-2',
          themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
        )}>
          <Globe className="w-4 h-4" />
          {t('subscriptions.nodes')} ({subscriptionNodes.length})
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {subscriptionNodes.map(node => (
            <NodeCard 
              key={node.id}
              node={node}
              testing={testing === node.id}
              themeStyle={themeStyle}
              onTest={() => handleTest(node)}
              onDelete={node.isManual ? () => handleDelete(node.id) : undefined}
              onCopy={() => handleCopyShare(node.id)}
            />
          ))}
        </div>
        
        {subscriptionNodes.length === 0 && (
          <div className={cn(
            'text-center py-8 text-sm',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
          )}>
            {t('common.loading')}
          </div>
        )}
      </div>

      {/* Import Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className={cn(
            'w-full max-w-md rounded-2xl p-6',
            themeStyle === 'apple-glass'
              ? 'bg-white/90 backdrop-blur-xl'
              : 'bg-neutral-900 border border-white/10'
          )}>
            <h3 className={cn(
              'text-lg font-semibold mb-4',
              themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
            )}>{t('nodes.importUrl')}</h3>
            <input
              type="text"
              value={importUrl}
              onChange={(e) => setImportUrl(e.target.value)}
              className="form-input mb-4"
              placeholder="vmess://... or ss://..."
            />
            <div className="flex gap-3">
              <button
                onClick={() => setShowImportModal(false)}
                className="flex-1 py-2.5 rounded-lg border border-white/10 text-sm"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleImport}
                className={cn(
                  'flex-1 py-2.5 rounded-lg text-sm font-medium text-white',
                  themeStyle === 'apple-glass'
                    ? 'bg-blue-500 hover:bg-blue-600'
                    : 'bg-indigo-500 hover:bg-indigo-600'
                )}
              >
                {t('common.add')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// Node Card Component
function NodeCard({ 
  node, 
  testing, 
  themeStyle, 
  onTest, 
  onDelete, 
  onCopy 
}: { 
  node: Node
  testing: boolean
  themeStyle: string
  onTest: () => void
  onDelete?: () => void
  onCopy: () => void
}) {
  return (
    <div className={cn(
      'p-3 rounded-lg border transition-all group',
      themeStyle === 'apple-glass'
        ? 'bg-white/50 border-black/5 hover:bg-white'
        : 'bg-white/5 border-white/5 hover:bg-white/10'
    )}>
      <div className="flex items-center justify-between mb-2">
        <span className={cn(
          'text-[10px] font-mono uppercase px-1.5 py-0.5 rounded',
          themeStyle === 'apple-glass' ? 'bg-black/5' : 'bg-white/10'
        )}>
          {node.type}
        </span>
        <span className={cn('text-xs font-mono', getLatencyColor(node.delay))}>
          {formatLatency(node.delay)}
        </span>
      </div>
      
      <div className={cn(
        'font-medium text-sm truncate mb-1',
        themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
      )}>
        {node.name}
      </div>
      
      <div className={cn(
        'text-xs truncate mb-3',
        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
      )}>
        {node.server}:{node.serverPort}
      </div>

      <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={onTest}
          disabled={testing}
          className="flex-1 py-1 rounded text-xs bg-white/10 hover:bg-white/20"
        >
          <RefreshCw className={cn('w-3 h-3 mx-auto', testing && 'animate-spin')} />
        </button>
        <button
          onClick={onCopy}
          className="flex-1 py-1 rounded text-xs bg-white/10 hover:bg-white/20"
        >
          <Copy className="w-3 h-3 mx-auto" />
        </button>
        {onDelete && (
          <button
            onClick={onDelete}
            className="flex-1 py-1 rounded text-xs bg-red-500/10 hover:bg-red-500/20"
          >
            <Trash2 className="w-3 h-3 mx-auto text-red-400" />
          </button>
        )}
      </div>
    </div>
  )
}
