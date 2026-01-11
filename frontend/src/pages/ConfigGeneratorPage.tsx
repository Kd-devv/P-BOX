import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  Users,
  List,
  Eye,
  Plus,
  Trash2,
  Edit2,
  Save,
  RotateCcw,
  ChevronDown,
  ChevronRight,
  Database,
  GripVertical,
  Globe,
  Zap,
  Rocket,
  Copy,
  Check,
  X,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { useCoreStore } from '@/stores/coreStore'
import { api } from '@/api/client'
import { singboxApi } from '@/api/singbox'
import { 
  loadSingBoxTemplate, 
  resetSingBoxTemplate,
  defaultSingBoxTemplate,
  type SingBoxTemplate
} from '@/api/singboxTemplate'
import { ErrorDialog } from '@/components/ErrorDialog'

// 默认代理组名称翻译映射 (中文 -> 英文)
const GROUP_NAME_MAP: Record<string, string> = {
  // 主要分组
  '自动选择': 'Auto Select',
  '故障转移': 'Failover',
  '节点选择': 'Node Select',
  '全球直连': 'Direct',
  // 服务分组
  'AI服务': 'AI Services',
  '国外媒体': 'Streaming',
  'Netflix': 'Netflix',
  '电报消息': 'Telegram',
  '谷歌服务': 'Google',
  '推特消息': 'Twitter',
  '脸书服务': 'Facebook',
  '游戏平台': 'Gaming',
  '哔哩哔哩': 'Bilibili',
  '微软服务': 'Microsoft',
  '苹果服务': 'Apple',
  'GitHub': 'GitHub',
  '广告拦截': 'Ad Block',
  '漏网之鱼': 'Final',
  // 地区节点
  '香港节点': 'Hong Kong',
  '台湾节点': 'Taiwan',
  '日本节点': 'Japan',
  '新加坡节点': 'Singapore',
  '美国节点': 'United States',
  '手动节点': 'Manual',
  '其他节点': 'Others',
}

// 翻译代理组名称 (当语言为英文时)
const translateGroupName = (name: string, lang: string): string => {
  // lang 可能是 'zh', 'zh-CN', 'zh-TW' 等，统一检查前缀
  if (lang.startsWith('zh')) return name
  return GROUP_NAME_MAP[name] || name
}

// 类型定义
interface ProxyGroup {
  name: string
  type: string
  icon: string
  description: string
  enabled: boolean
  proxies: string[]
  url?: string
  interval?: number
  tolerance?: number
  lazy?: boolean
  filter?: string
  useAll?: boolean
}

interface Rule {
  type: string
  payload: string
  proxy: string
  noResolve: boolean
  description: string
}

interface RuleProvider {
  name: string
  type: string
  behavior: string
  url: string
  path: string
  interval: number
  format: string
  description: string
}

interface ConfigTemplate {
  proxyGroups: ProxyGroup[]
  rules: Rule[]
  ruleProviders: RuleProvider[]
}

export default function ConfigGeneratorPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const { activeCore } = useCoreStore()
  const [activeTab, setActiveTab] = useState('groups')
  // Mihomo 模板
  const [template, setTemplate] = useState<ConfigTemplate | null>(null)
  // Sing-Box 原始模板 (用于保存和重置)
  const [, setSingboxTemplate] = useState<SingBoxTemplate | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  // 错误弹窗状态
  const [showError, setShowError] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')

  // 两种核心类型都显示相同的 tabs
  const tabs = [
    { id: 'groups', icon: Users, label: t('configGenerator.proxyGroups') || '代理组' },
    { id: 'rules', icon: List, label: t('configGenerator.rules') || '规则' },
    { id: 'providers', icon: Database, label: t('configGenerator.rulesets') || '规则集' },
    { id: 'preview', icon: Eye, label: t('configGenerator.preview') || '预览' },
  ]

  // 加载配置模板 - 根据核心类型
  const loadTemplate = async () => {
    try {
      setLoading(true)
      if (activeCore === 'singbox') {
        // 加载 Sing-Box 模板 (从后端 API)
        const sbTemplate = await loadSingBoxTemplate()
        setSingboxTemplate(sbTemplate)
        // 同时设置一个兼容的 template 用于 UI
        setTemplate({
          proxyGroups: (sbTemplate.proxyGroups || []).map(g => ({
            name: g.name || g.tag,  // 使用中文名称显示，如果没有则用 tag
            type: g.type === 'urltest' ? 'url-test' : g.type,
            icon: g.icon || '',
            description: `${g.description || ''} (${g.tag})`,  // 在描述中显示 tag
            enabled: g.enabled ?? true,
            proxies: g.outbounds || [],
            useAll: false
          })),
          rules: (sbTemplate.rules || []).map((r) => {
            // rule_set 可能是 string 或 string[]
            let payload = ''
            if (r.rule_set) {
              payload = Array.isArray(r.rule_set) ? r.rule_set.join(',') : String(r.rule_set)
            } else if (r.domain_suffix) {
              payload = r.domain_suffix.join(',')
            } else if (r.ip_cidr) {
              payload = r.ip_cidr.join(',')
            }
            return {
              type: r.rule_set ? 'RULE-SET' : r.domain_suffix ? 'DOMAIN-SUFFIX' : r.ip_cidr ? 'IP-CIDR' : 'MATCH',
              payload,
              proxy: r.outbound || '',
              noResolve: false,
              description: r.action || ''
            }
          }),
          ruleProviders: (sbTemplate.ruleSets || []).map(rs => ({
            name: rs.tag,
            type: rs.type,
            behavior: rs.format,
            url: rs.url || '',
            path: rs.path || '',
            interval: 86400,
            format: rs.format,
            description: ''
          }))
        })
      } else {
        // 加载 Mihomo 模板
        const data = await api.get<ConfigTemplate>('/proxy/template')
        setTemplate(data)
      }
    } catch {
      // 使用默认模板
      if (activeCore === 'singbox') {
        setSingboxTemplate(defaultSingBoxTemplate)
        setTemplate({
          proxyGroups: [],
          rules: [],
          ruleProviders: []
        })
      } else {
        setTemplate({
          proxyGroups: [],
          rules: [],
          ruleProviders: []
        })
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTemplate()
  }, [activeCore])

  // 重置为默认
  const resetTemplate = async () => {
    if (!confirm(t('configGenerator.confirmReset') || '确定要重置为默认配置吗？')) return
    try {
      setSaving(true)
      if (activeCore === 'singbox') {
        // 重置 Sing-Box 模板 (从后端获取默认值)
        const newTemplate = await resetSingBoxTemplate()
        setSingboxTemplate(newTemplate)
      } else {
        await api.post('/proxy/template/reset', {})
      }
      await loadTemplate()
    } catch {
      // Ignore
    } finally {
      setSaving(false)
    }
  }

  // 生成配置 - 根据核心类型
  const generateConfig = async () => {
    try {
      setSaving(true)
      
      if (activeCore === 'singbox') {
        // 生成 Sing-Box 配置
        const settings = singboxApi.loadSettings()
        const result = await singboxApi.generateConfig(settings)
        
        // 检查是否有验证错误 (code === 2 表示验证失败)
        if (result.code === 2 && result.data?.validationError) {
          setErrorMessage(result.data.validationError)
          setShowError(true)
          return
        }
        
        // 检查其他错误
        if (result.code !== 0) {
          setErrorMessage(result.message || '生成配置失败')
          setShowError(true)
          return
        }
        
        // 验证成功提示
        toast.success(t('configGenerator.generateSuccess') || '配置生成成功', {
          description: `${t('configGenerator.validationPassed') || '配置验证通过'} - ${result.data?.nodeCount || 0} ${t('nodes.title') || '节点'}`
        })
      } else {
        // 生成 Mihomo 配置
        await api.post('/proxy/generate', { nodes: [] })
        // 刷新模板以获取最新配置
        await loadTemplate()
        
        // 成功提示
        toast.success(t('configGenerator.generateSuccess') || '配置生成成功')
      }
      
      // 切换到预览 tab
      setActiveTab('preview')
    } catch (err) {
      console.error('生成配置失败:', err)
      setErrorMessage(err instanceof Error ? err.message : '生成配置失败')
      setShowError(true)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className={cn(
          'w-8 h-8 animate-spin',
          themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
        )} />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* 顶部标题和核心类型 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('configGenerator.title') || '配置生成'}</h2>
          <p className={cn(
            'text-sm mt-1',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>
            当前核心: <span className={cn(
              'font-medium px-2 py-0.5 rounded',
              activeCore === 'singbox'
                ? 'bg-purple-500/20 text-purple-500'
                : 'bg-cyan-500/20 text-cyan-500'
            )}>{activeCore === 'singbox' ? 'Sing-Box' : 'Mihomo'}</span>
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={generateConfig}
            disabled={saving}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
              activeCore === 'singbox'
                ? 'bg-purple-500 text-white hover:bg-purple-600'
                : themeStyle === 'apple-glass'
                  ? 'bg-blue-500 text-white hover:bg-blue-600'
                  : 'bg-cyan-500 text-white hover:bg-cyan-600'
            )}
          >
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
            {saving ? t('configGenerator.generating') : t('configGenerator.generate')}
          </button>
        <button
          onClick={resetTemplate}
          disabled={saving}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
            themeStyle === 'apple-glass'
              ? 'bg-white/60 border border-black/10 text-slate-700 hover:bg-white/80'
              : 'bg-white/5 border border-white/10 text-slate-300 hover:bg-white/10'
          )}
        >
          <RotateCcw className="w-4 h-4" />
          {t('configGenerator.resetDefault') || '重置默认'}
        </button>
        </div>
      </div>

      {/* 标签页 */}
      <div className={cn(
        'flex gap-1 p-1 rounded-xl',
        themeStyle === 'apple-glass' ? 'bg-black/5' : 'bg-white/5'
      )}>
        {tabs.map((tab) => {
          const Icon = tab.icon
          const count = template ? (
            tab.id === 'groups' ? template.proxyGroups.length :
            tab.id === 'rules' ? template.rules.length :
            tab.id === 'providers' ? template.ruleProviders.length : 0
          ) : 0
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex-1 flex items-center justify-center gap-1 sm:gap-2 px-2 sm:px-4 py-2 sm:py-2.5 rounded-lg text-xs sm:text-sm font-medium transition-all',
                activeTab === tab.id
                  ? themeStyle === 'apple-glass'
                    ? 'bg-white shadow-sm text-slate-800'
                    : 'bg-white/10 text-white'
                  : themeStyle === 'apple-glass'
                    ? 'text-slate-500 hover:text-slate-700'
                    : 'text-slate-400 hover:text-white'
              )}
            >
              <Icon className="w-4 h-4" />
              <span className="hidden sm:inline">{tab.label}</span>
              {count > 0 && (
                <span className={cn(
                  'text-xs px-1.5 py-0.5 rounded-full',
                  themeStyle === 'apple-glass' ? 'bg-blue-100 text-blue-600' : 'bg-cyan-500/20 text-cyan-400'
                )}>{count}</span>
              )}
            </button>
          )
        })}
      </div>

      {/* 内容区 */}
      {template && (
        <div className="glass-card p-3 sm:p-5">
          {activeTab === 'groups' && (
            <ProxyGroupsTab template={template} setTemplate={setTemplate} />
          )}
          {activeTab === 'rules' && (
            <RulesTab template={template} setTemplate={setTemplate} />
          )}
          {activeTab === 'providers' && (
            <ProvidersTab template={template} setTemplate={setTemplate} />
          )}
          {activeTab === 'preview' && (
            <PreviewTab />
          )}
        </div>
      )}

      {/* 错误弹窗 */}
      <ErrorDialog
        open={showError}
        onOpenChange={setShowError}
        title={t('configGenerator.validationError') || '配置验证失败'}
        error={errorMessage}
      />
    </div>
  )
}

// 代理组管理 Tab
function ProxyGroupsTab({ 
  template, 
  setTemplate 
}: { 
  template: ConfigTemplate
  setTemplate: (tpl: ConfigTemplate) => void 
}) {
  const { t, i18n } = useTranslation()
  const lang = i18n.language
  const { themeStyle } = useThemeStore()
  const [editingGroup, setEditingGroup] = useState<ProxyGroup | null>(null)
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  const saveGroups = async (groups: ProxyGroup[]) => {
    try {
      await api.put('/proxy/template/groups', groups)
      setTemplate({ ...template, proxyGroups: groups })
    } catch {
      // Ignore
    }
  }

  const deleteGroup = (name: string) => {
    if (!confirm(t('configGen.confirmDelete'))) return
    const newGroups = template.proxyGroups.filter(g => g.name !== name)
    saveGroups(newGroups)
  }

  const toggleEnabled = (name: string) => {
    const newGroups = template.proxyGroups.map(g => 
      g.name === name ? { ...g, enabled: !g.enabled } : g
    )
    saveGroups(newGroups)
  }

  const addGroup = () => {
    setEditingGroup({
      name: '',
      type: 'select',
      icon: 'globe',
      description: '',
      enabled: true,
      proxies: ['节点选择', 'DIRECT'],
      useAll: false,
    })
  }

  const saveEditingGroup = () => {
    if (!editingGroup) return
    if (!editingGroup.name.trim()) return

    const existingIndex = template.proxyGroups.findIndex(g => g.name === editingGroup.name)
    let newGroups: ProxyGroup[]
    
    if (existingIndex >= 0) {
      newGroups = [...template.proxyGroups]
      newGroups[existingIndex] = editingGroup
    } else {
      newGroups = [...template.proxyGroups, editingGroup]
    }

    saveGroups(newGroups)
    setEditingGroup(null)
  }

  const toggleExpand = (name: string) => {
    const newExpanded = new Set(expandedGroups)
    if (newExpanded.has(name)) {
      newExpanded.delete(name)
    } else {
      newExpanded.add(name)
    }
    setExpandedGroups(newExpanded)
  }

  // 拖拽处理函数
  const handleDragStart = (e: React.DragEvent, index: number) => {
    setDraggedIndex(index)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', index.toString())
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverIndex(index)
  }

  const handleDragLeave = () => {
    setDragOverIndex(null)
  }

  const handleDrop = (e: React.DragEvent, dropIndex: number) => {
    e.preventDefault()
    const dragIndex = parseInt(e.dataTransfer.getData('text/plain'))
    
    if (dragIndex === dropIndex) {
      setDraggedIndex(null)
      setDragOverIndex(null)
      return
    }

    // 重新排序
    const newGroups = [...template.proxyGroups]
    const [draggedItem] = newGroups.splice(dragIndex, 1)
    newGroups.splice(dropIndex, 0, draggedItem)
    
    saveGroups(newGroups)
    setDraggedIndex(null)
    setDragOverIndex(null)
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
    setDragOverIndex(null)
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <p className={cn(
          'text-sm',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          {t('configGen.groupsDescription')}
        </p>
        <button
          onClick={addGroup}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
            themeStyle === 'apple-glass'
              ? 'bg-blue-500 text-white hover:bg-blue-600'
              : 'bg-cyan-500 text-white hover:bg-cyan-600'
          )}
        >
          <Plus className="w-4 h-4" />
          {t('configGen.addGroup')}
        </button>
      </div>

      {/* 分组列表 */}
      <div className="space-y-2">
        {template.proxyGroups.length === 0 ? (
          <div className={cn(
            'text-center py-12',
            themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
          )}>
            {t('configGen.noGroups')}
          </div>
        ) : (
          template.proxyGroups.map((group, index) => {
            const isExpanded = expandedGroups.has(group.name)
            const isEnabled = group.enabled !== false
            const isDragging = draggedIndex === index
            const isDragOver = dragOverIndex === index
            
            return (
              <div
                key={group.name}
                draggable
                onDragStart={(e) => handleDragStart(e, index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragLeave={handleDragLeave}
                onDrop={(e) => handleDrop(e, index)}
                onDragEnd={handleDragEnd}
                className={cn(
                  'rounded-xl border transition-all',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/40 border-white/30'
                    : 'bg-white/5 border-white/10',
                  !isEnabled && 'opacity-50',
                  isDragging && 'opacity-50 scale-[0.98]',
                  isDragOver && 'border-blue-500 border-2'
                )}
              >
                <div className="flex items-center gap-3 p-3">
                  <GripVertical className={cn(
                    'w-4 h-4 cursor-grab active:cursor-grabbing',
                    themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                  )} />
                  
                  <button onClick={() => toggleExpand(group.name)} className="p-1">
                    {isExpanded ? (
                      <ChevronDown className="w-4 h-4" />
                    ) : (
                      <ChevronRight className="w-4 h-4" />
                    )}
                  </button>

                  <div className={cn(
                    'w-8 h-8 rounded-lg flex items-center justify-center',
                    themeStyle === 'apple-glass'
                      ? 'bg-blue-100 text-blue-600'
                      : 'bg-cyan-500/20 text-cyan-400'
                  )}>
                    {group.type === 'url-test' ? <Zap className="w-4 h-4" /> :
                     group.type === 'fallback' ? <Rocket className="w-4 h-4" /> :
                     <Globe className="w-4 h-4" />}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className={cn(
                      'font-medium truncate',
                      themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                    )}>{translateGroupName(group.name, lang)}</div>
                    <div className={cn(
                      'text-xs',
                      themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                    )}>
                      {group.type} · {group.useAll ? t('configGen.allNodes') : `${group.proxies.length} ${t('configGen.proxiesCount')}`}
                    </div>
                  </div>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => toggleEnabled(group.name)}
                      className={cn(
                        'p-2 rounded-lg transition-colors',
                        isEnabled
                          ? 'text-green-500 hover:bg-green-500/10'
                          : 'text-slate-400 hover:bg-slate-500/10'
                      )}
                    >
                      <Check className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => setEditingGroup(group)}
                      className={cn(
                        'p-2 rounded-lg transition-colors',
                        themeStyle === 'apple-glass'
                          ? 'text-slate-600 hover:bg-black/5'
                          : 'text-slate-400 hover:bg-white/10'
                      )}
                    >
                      <Edit2 className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => deleteGroup(group.name)}
                      className="p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                {isExpanded && (
                  <div className={cn(
                    'px-4 pb-3 pt-1 border-t',
                    themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
                  )}>
                    <div className={cn(
                      'text-xs mb-2',
                      themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                    )}>
                      {group.description || '无描述'}
                    </div>
                    {group.filter && (
                      <div className={cn(
                        'text-xs font-mono mb-2',
                        themeStyle === 'apple-glass' ? 'text-blue-600' : 'text-cyan-400'
                      )}>
                        过滤: {group.filter}
                      </div>
                    )}
                    <div className="flex flex-wrap gap-1">
                      {(group.useAll ? ['全部节点'] : group.proxies).map((proxy, i) => (
                        <span
                          key={i}
                          className={cn(
                            'text-xs px-2 py-0.5 rounded',
                            themeStyle === 'apple-glass'
                              ? 'bg-black/5 text-slate-600'
                              : 'bg-white/10 text-slate-300'
                          )}
                        >
                          {proxy}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>

      {/* 编辑对话框 - 使用 Portal 渲染到 body */}
      {editingGroup && createPortal(
        <EditGroupDialog
          group={editingGroup}
          onChange={setEditingGroup}
          onSave={saveEditingGroup}
          onCancel={() => setEditingGroup(null)}
          isNew={!template.proxyGroups.find(g => g.name === editingGroup.name)}
        />,
        document.body
      )}
    </div>
  )
}

// 规则管理 Tab
function RulesTab({ 
  template, 
  setTemplate 
}: { 
  template: ConfigTemplate
  setTemplate: (tpl: ConfigTemplate) => void 
}) {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [editingRule, setEditingRule] = useState<Rule | null>(null)
  const [editingIndex, setEditingIndex] = useState<number>(-1)

  const saveRules = async (rules: Rule[]) => {
    try {
      await api.put('/proxy/template/rules', rules)
      setTemplate({ ...template, rules })
    } catch {
      // Ignore
    }
  }

  const deleteRule = (index: number) => {
    const newRules = template.rules.filter((_, i) => i !== index)
    saveRules(newRules)
  }

  const addRule = () => {
    setEditingRule({
      type: 'DOMAIN-SUFFIX',
      payload: '',
      proxy: template.proxyGroups[0]?.name || 'DIRECT',
      noResolve: false,
      description: ''
    })
    setEditingIndex(-1)
  }

  const saveEditingRule = () => {
    if (!editingRule) return
    let newRules: Rule[]
    
    // 检查是否有批量输入（多行）
    const payloadLines = editingRule.payload.split('\n').filter(l => l.trim())
    
    if (editingIndex >= 0) {
      // 编辑模式：直接更新单条规则
      newRules = [...template.rules]
      newRules[editingIndex] = { ...editingRule, payload: payloadLines[0] || '' }
    } else if (payloadLines.length > 1 && editingRule.type !== 'RULE-SET') {
      // 新增模式 + 批量输入：生成多条规则
      const batchRules: Rule[] = payloadLines.map(payload => ({
        ...editingRule,
        payload: payload.trim()
      }))
      newRules = [...template.rules, ...batchRules]
    } else {
      // 新增单条规则
      newRules = [...template.rules, { ...editingRule, payload: payloadLines[0] || '' }]
    }

    saveRules(newRules)
    setEditingRule(null)
    setEditingIndex(-1)
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <p className={cn(
          'text-sm',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          {t('configGen.rulesDescription')}
        </p>
        <button
          onClick={addRule}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
            themeStyle === 'apple-glass'
              ? 'bg-blue-500 text-white hover:bg-blue-600'
              : 'bg-cyan-500 text-white hover:bg-cyan-600'
          )}
        >
          <Plus className="w-4 h-4" />
          {t('configGen.addRule')}
        </button>
      </div>

      <div className="space-y-2">
        {template.rules.length === 0 ? (
          <div className={cn(
            'text-center py-12',
            themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
          )}>
            {t('configGen.noRules')}
          </div>
        ) : (
          template.rules.map((rule, index) => (
            <div
              key={index}
              className={cn(
                'flex items-center gap-3 p-3 rounded-xl border',
                themeStyle === 'apple-glass'
                  ? 'bg-white/40 border-white/30'
                  : 'bg-white/5 border-white/10'
              )}
            >
              <div className={cn(
                'px-2 py-1 rounded text-xs font-mono',
                themeStyle === 'apple-glass'
                  ? 'bg-purple-100 text-purple-600'
                  : 'bg-purple-500/20 text-purple-400'
              )}>
                {rule.type}
              </div>
              <div className="flex-1 min-w-0">
                <div className={cn(
                  'font-mono text-sm truncate',
                  themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-200'
                )}>
                  {rule.payload || `(${t('common.empty')})`}
                </div>
                {rule.description && (
                  <div className={cn(
                    'text-xs truncate',
                    themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                  )}>
                    {rule.description}
                  </div>
                )}
              </div>
              <div className={cn(
                'px-2 py-1 rounded text-xs',
                themeStyle === 'apple-glass'
                  ? 'bg-green-100 text-green-600'
                  : 'bg-green-500/20 text-green-400'
              )}>
                → {rule.proxy}
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => { setEditingRule(rule); setEditingIndex(index) }}
                  className={cn(
                    'p-2 rounded-lg transition-colors',
                    themeStyle === 'apple-glass'
                      ? 'text-slate-600 hover:bg-black/5'
                      : 'text-slate-400 hover:bg-white/10'
                  )}
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => deleteRule(index)}
                  className="p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {editingRule && createPortal(
        <EditRuleDialog
          rule={editingRule}
          proxyGroups={template.proxyGroups}
          onChange={setEditingRule}
          onSave={saveEditingRule}
          onCancel={() => { setEditingRule(null); setEditingIndex(-1) }}
          isNew={editingIndex < 0}
        />,
        document.body
      )}
    </div>
  )
}

// 规则集 Tab
function ProvidersTab({ template, setTemplate }: { template: ConfigTemplate, setTemplate: (tpl: ConfigTemplate) => void }) {
  const { themeStyle } = useThemeStore()
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null)
  const [updating, setUpdating] = useState(false)
  const [updatedProviders, setUpdatedProviders] = useState<Set<string>>(new Set())
  const [currentUpdating, setCurrentUpdating] = useState<string | null>(null)
  const [editingProvider, setEditingProvider] = useState<string | null>(null)
  const [editingUrl, setEditingUrl] = useState('')

  const handleEditUrl = (providerName: string, currentUrl: string) => {
    setEditingProvider(providerName)
    setEditingUrl(currentUrl)
  }

  const handleSaveUrl = async () => {
    if (!editingProvider) return
    const newProviders = template.ruleProviders.map(p => 
      p.name === editingProvider ? { ...p, url: editingUrl } : p
    )
    try {
      await api.put('/proxy/template/providers', newProviders)
      setTemplate({ ...template, ruleProviders: newProviders })
    } catch {
      // Ignore
    }
    setEditingProvider(null)
  }

  const copyUrl = (url: string) => {
    navigator.clipboard.writeText(url)
    setCopiedUrl(url)
    setTimeout(() => setCopiedUrl(null), 2000)
  }

  const handleUpdateAll = async () => {
    if (updating || template.ruleProviders.length === 0) return
    
    setUpdating(true)
    setUpdatedProviders(new Set())
    
    // 逐个更新规则集，显示进度
    for (const provider of template.ruleProviders) {
      setCurrentUpdating(provider.name)
      try {
        // 调用后端强制更新规则集
        await api.post(`/proxy/providers/rules/${provider.name}`)
        // 标记为已更新
        setUpdatedProviders(prev => new Set(prev).add(provider.name))
      } catch {
        // 忽略单个失败，继续更新其他
      }
      // 短暂延迟，让用户看到进度
      await new Promise(r => setTimeout(r, 300))
    }
    
    setCurrentUpdating(null)
    setUpdating(false)
    
    // 5秒后清除完成状态
    setTimeout(() => setUpdatedProviders(new Set()), 5000)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className={cn(
          'text-sm',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          规则集由系统自动管理，可在规则页面引用
        </p>
        {template.ruleProviders.length > 0 && (
          <button
            onClick={handleUpdateAll}
            disabled={updating}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors',
              updating
                ? 'opacity-50 cursor-not-allowed'
                : themeStyle === 'apple-glass'
                  ? 'bg-blue-500 text-white hover:bg-blue-600'
                  : 'bg-cyan-500 text-white hover:bg-cyan-600'
            )}
          >
            {updating ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <RefreshCw className="w-3.5 h-3.5" />
            )}
            {updating ? '更新中...' : '批量更新'}
          </button>
        )}
      </div>

      <div className="space-y-2">
        {template.ruleProviders.length === 0 ? (
          <div className={cn(
            'text-center py-12',
            themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
          )}>
            暂无规则集
          </div>
        ) : (
          template.ruleProviders.map((provider) => {
            const isUpdating = currentUpdating === provider.name
            const isUpdated = updatedProviders.has(provider.name)
            
            return (
              <div
                key={provider.name}
                className={cn(
                  'p-3 rounded-xl border group transition-colors',
                  isUpdated && 'ring-1 ring-green-500/50',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/40 border-white/30'
                    : 'bg-white/5 border-white/10'
                )}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    {/* 状态图标 */}
                    <div className="w-5 h-5 flex-shrink-0 flex items-center justify-center">
                      {isUpdating ? (
                        <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
                      ) : isUpdated ? (
                        <Check className="w-5 h-5 text-green-500" />
                      ) : (
                        <Database className={cn(
                          'w-5 h-5',
                          themeStyle === 'apple-glass' ? 'text-orange-500' : 'text-orange-400'
                        )} />
                      )}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className={cn(
                        'font-medium',
                        themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                      )}>{provider.name}</div>
                      <div className={cn(
                        'text-xs',
                        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
                      )}>
                        {provider.behavior} · {provider.type}
                      </div>
                      {/* URL 显示/编辑 */}
                      {editingProvider === provider.name ? (
                        <div className="flex items-center gap-2 mt-1">
                          <input
                            type="text"
                            value={editingUrl}
                            onChange={(e) => setEditingUrl(e.target.value)}
                            className={cn(
                              'flex-1 text-xs px-2 py-1 rounded border',
                              themeStyle === 'apple-glass'
                                ? 'bg-white border-slate-200 text-slate-700'
                                : 'bg-neutral-800 border-neutral-600 text-slate-200'
                            )}
                            autoFocus
                          />
                          <button
                            onClick={handleSaveUrl}
                            className="p-1 rounded text-green-500 hover:bg-green-500/10"
                          >
                            <Check className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => setEditingProvider(null)}
                            className="p-1 rounded text-red-500 hover:bg-red-500/10"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        </div>
                      ) : provider.url && (
                        <div 
                          className={cn(
                            'text-[10px] truncate mt-1 cursor-pointer transition-all duration-200',
                            copiedUrl === provider.url 
                              ? 'text-green-500 font-medium' 
                              : themeStyle === 'apple-glass' 
                                ? 'text-slate-400 hover:text-blue-500' 
                                : 'text-slate-500 hover:text-blue-400'
                          )} 
                          onClick={() => copyUrl(provider.url)}
                          title="点击复制 URL"
                        >
                          {copiedUrl === provider.url ? '✓ 已复制!' : provider.url}
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 flex-shrink-0">
                    {/* 编辑 URL 按钮 */}
                    <button
                      onClick={() => handleEditUrl(provider.name, provider.url || '')}
                      className={cn(
                        'p-1.5 rounded-lg transition-colors',
                        themeStyle === 'apple-glass'
                          ? 'text-slate-400 hover:text-slate-600 hover:bg-black/5'
                          : 'text-slate-500 hover:text-slate-300 hover:bg-white/10'
                      )}
                      title="编辑 URL"
                    >
                      <Edit2 className="w-4 h-4" />
                    </button>
                    {/* 复制 URL 按钮 */}
                    {provider.url && (
                      <button
                        onClick={() => copyUrl(provider.url)}
                        className={cn(
                          'p-1.5 rounded-lg transition-colors',
                          copiedUrl === provider.url
                            ? 'text-green-500'
                            : themeStyle === 'apple-glass'
                              ? 'text-slate-400 hover:text-slate-600 hover:bg-black/5'
                              : 'text-slate-500 hover:text-slate-300 hover:bg-white/10'
                        )}
                        title="复制 URL"
                      >
                        {copiedUrl === provider.url ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <Copy className="w-4 h-4" />
                        )}
                      </button>
                    )}
                    <div className={cn(
                      'text-xs',
                      isUpdated ? 'text-green-500' : themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                    )}>
                      {isUpdated ? '已更新' : `${provider.interval / 3600}h 更新`}
                    </div>
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

// 预览 Tab - 根据核心类型加载配置
function PreviewTab() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const { activeCore } = useCoreStore()
  const [copied, setCopied] = useState(false)
  const [configContent, setConfigContent] = useState<string>('')
  const [loading, setLoading] = useState(true)

  // 加载生成的配置文件
  const loadConfig = async () => {
    try {
      setLoading(true)
      
      if (activeCore === 'singbox') {
        // 加载 Sing-Box 配置
        try {
          const content = await singboxApi.getConfigPreview()
          setConfigContent(content)
        } catch {
          setConfigContent('// Sing-Box 配置文件未生成\n// 请先点击上方的「生成配置」按钮')
        }
      } else {
        // 加载 Mihomo 配置
        const data = await api.get<{ content: string }>('/proxy/config/preview')
        if (data?.content) {
          setConfigContent(data.content)
        } else {
          setConfigContent('# 配置文件未生成\n# 请先点击上方的「生成配置」按钮')
        }
      }
    } catch {
      const emptyMsg = activeCore === 'singbox' 
        ? '// Sing-Box 配置文件未生成\n// 请先点击上方的「生成配置」按钮'
        : '# 配置文件未生成\n# 请先点击上方的「生成配置」按钮'
      setConfigContent(emptyMsg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadConfig()
  }, [activeCore])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(configContent)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleRefresh = () => {
    loadConfig()
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <p className={cn(
          'text-sm',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          {activeCore === 'singbox' 
            ? '预览生成的 Sing-Box 配置文件 (JSON 格式)'
            : (t('configGenerator.previewDescription') || '预览生成的 config.yaml 配置文件')}
        </p>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-all',
              themeStyle === 'apple-glass'
                ? 'bg-black/5 text-slate-600 hover:bg-black/10'
                : 'bg-white/10 text-slate-300 hover:bg-white/20'
            )}
          >
            <RefreshCw className="w-4 h-4" />
            {t('common.refresh') || '刷新'}
          </button>
          <button
            onClick={handleCopy}
            disabled={!configContent}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-all',
              copied
                ? 'bg-green-500/20 text-green-500'
                : themeStyle === 'apple-glass'
                  ? 'bg-black/5 text-slate-600 hover:bg-black/10'
                  : 'bg-white/10 text-slate-300 hover:bg-white/20'
            )}
          >
            {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
            {copied ? t('common.copied') || '已复制' : t('common.copy') || '复制'}
          </button>
        </div>
      </div>

      <pre className={cn(
        'p-4 rounded-xl overflow-auto text-sm font-mono max-h-[600px]',
        themeStyle === 'apple-glass'
          ? 'bg-slate-100 text-slate-700 border border-black/10'
          : 'bg-black/30 text-green-400 border border-white/10'
      )}>
        {loading ? t('common.loading') || '加载中...' : configContent}
      </pre>
    </div>
  )
}

// 编辑代理组对话框
function EditGroupDialog({
  group,
  onChange,
  onSave,
  onCancel,
  isNew,
}: {
  group: ProxyGroup
  onChange: (g: ProxyGroup) => void
  onSave: () => void
  onCancel: () => void
  isNew: boolean
}) {
  const { themeStyle } = useThemeStore()

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className={cn(
        'w-full max-w-lg max-h-[80vh] overflow-y-auto rounded-2xl p-6',
        themeStyle === 'apple-glass'
          ? 'bg-white/90 backdrop-blur-xl border border-white/50'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        <div className="flex items-center justify-between mb-6">
          <h3 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{isNew ? '添加' : '编辑'}代理组</h3>
          <button onClick={onCancel} className="p-2 rounded-lg hover:bg-black/5">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>名称 *</label>
            <input
              type="text"
              value={group.name}
              onChange={(e) => onChange({ ...group, name: e.target.value })}
              className="form-input"
              placeholder="例如：🚀 节点选择"
            />
          </div>

          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>类型</label>
            <select
              value={group.type}
              onChange={(e) => onChange({ ...group, type: e.target.value })}
              className="form-input"
            >
              <option value="select">select - 手动选择</option>
              <option value="url-test">url-test - 自动测速</option>
              <option value="fallback">fallback - 故障转移</option>
              <option value="load-balance">load-balance - 负载均衡</option>
            </select>
          </div>

          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>说明</label>
            <input
              type="text"
              value={group.description}
              onChange={(e) => onChange({ ...group, description: e.target.value })}
              className="form-input"
              placeholder="描述这个分组的用途"
            />
          </div>

          {(group.type === 'url-test' || group.type === 'fallback') && (
            <div>
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>节点过滤正则</label>
              <input
                type="text"
                value={group.filter || ''}
                onChange={(e) => onChange({ ...group, filter: e.target.value })}
                className="form-input font-mono"
                placeholder="(?i)港|HK|Hong"
              />
            </div>
          )}

          <div>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={group.useAll || false}
                onChange={(e) => onChange({ ...group, useAll: e.target.checked })}
                className="rounded"
              />
              <span className={cn(
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>使用全部订阅节点</span>
            </label>
          </div>

          {!group.useAll && (
            <div>
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>代理列表（每行一个）</label>
              
              {/* 快捷分组按钮 */}
              <div className="flex flex-wrap gap-1.5 mb-2">
                {[
                  { label: '节点选择', value: '节点选择' },
                  { label: '自动选择', value: '自动选择' },
                  { label: '故障转移', value: '故障转移' },
                  { label: '直连', value: '直连' },
                  { label: '香港节点', value: '香港节点' },
                  { label: '台湾节点', value: '台湾节点' },
                  { label: '日本节点', value: '日本节点' },
                  { label: '新加坡节点', value: '新加坡节点' },
                  { label: '美国节点', value: '美国节点' },
                  { label: '手动节点', value: '手动节点' },
                  { label: '其他节点', value: '其他节点' },
                ].map(item => (
                  <button
                    key={item.value}
                    type="button"
                    onClick={() => {
                      const current = group.proxies || []
                      if (!current.includes(item.value)) {
                        onChange({ ...group, proxies: [...current, item.value] })
                      }
                    }}
                    className={cn(
                      'px-2 py-1 text-xs rounded-md transition-colors',
                      group.proxies?.includes(item.value)
                        ? themeStyle === 'apple-glass'
                          ? 'bg-blue-500 text-white'
                          : 'bg-cyan-500 text-white'
                        : themeStyle === 'apple-glass'
                          ? 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                          : 'bg-white/10 text-slate-300 hover:bg-white/20'
                    )}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
              
              <textarea
                value={group.proxies.join('\n')}
                onChange={(e) => onChange({ 
                  ...group, 
                  proxies: e.target.value.split('\n').filter(p => p.trim()) 
                })}
                className="form-input h-32 font-mono text-sm"
                placeholder="节点选择&#10;自动选择&#10;直连"
              />
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 mt-6">
          <button 
            onClick={onCancel} 
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium',
              themeStyle === 'apple-glass'
                ? 'hover:bg-black/5 text-slate-600'
                : 'hover:bg-white/10 text-slate-400'
            )}
          >
            取消
          </button>
          <button
            onClick={onSave}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 hover:bg-blue-600'
                : 'bg-cyan-500 hover:bg-cyan-600'
            )}
          >
            <Save className="w-4 h-4" />
            保存
          </button>
        </div>
      </div>
    </div>
  )
}

// 编辑规则对话框
function EditRuleDialog({
  rule,
  proxyGroups,
  onChange,
  onSave,
  onCancel,
  isNew,
}: {
  rule: Rule
  proxyGroups: ProxyGroup[]
  onChange: (r: Rule) => void
  onSave: () => void
  onCancel: () => void
  isNew: boolean
}) {
  const { themeStyle } = useThemeStore()

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className={cn(
        'w-full max-w-lg rounded-2xl p-6',
        themeStyle === 'apple-glass'
          ? 'bg-white/90 backdrop-blur-xl border border-white/50'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        <div className="flex items-center justify-between mb-6">
          <h3 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{isNew ? '添加' : '编辑'}规则</h3>
          <button onClick={onCancel} className="p-2 rounded-lg hover:bg-black/5">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>规则类型</label>
            <select
              value={rule.type}
              onChange={(e) => onChange({ ...rule, type: e.target.value })}
              className="form-input"
            >
              <option value="DOMAIN">DOMAIN - 完整域名</option>
              <option value="DOMAIN-SUFFIX">DOMAIN-SUFFIX - 域名后缀</option>
              <option value="DOMAIN-KEYWORD">DOMAIN-KEYWORD - 域名关键字</option>
              <option value="IP-CIDR">IP-CIDR - IP 地址段</option>
              <option value="GEOIP">GEOIP - 地理 IP</option>
              <option value="RULE-SET">RULE-SET - 规则集</option>
              <option value="MATCH">MATCH - 兜底规则</option>
            </select>
          </div>

          {rule.type !== 'MATCH' && (
            <div>
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>
                规则内容 * 
                <span className={cn(
                  'text-xs ml-2 font-normal',
                  themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                )}>
                  {rule.type === 'RULE-SET' ? '规则集名称' : '(批量输入：一行一条)'}
                </span>
              </label>
              <textarea
                value={rule.payload}
                onChange={(e) => onChange({ ...rule, payload: e.target.value })}
                className="form-input font-mono min-h-[100px] resize-y"
                placeholder={
                  rule.type === 'DOMAIN' ? 'www.google.com\nwww.example.com' :
                  rule.type === 'DOMAIN-SUFFIX' ? 'google.com\nexample.com' :
                  rule.type === 'DOMAIN-KEYWORD' ? 'google\nexample' :
                  rule.type === 'IP-CIDR' ? '192.168.0.0/16\n10.0.0.0/8' :
                  rule.type === 'GEOIP' ? 'CN' :
                  rule.type === 'RULE-SET' ? 'google-domain' : ''
                }
                rows={rule.type === 'RULE-SET' ? 1 : 4}
              />
              {rule.type !== 'RULE-SET' && rule.payload && rule.payload.includes('\n') && (
                <div className={cn(
                  'text-xs mt-1',
                  themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                )}>
                  将生成 {rule.payload.split('\n').filter(l => l.trim()).length} 条规则
                </div>
              )}
            </div>
          )}

          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>目标代理组</label>
            <select
              value={rule.proxy}
              onChange={(e) => onChange({ ...rule, proxy: e.target.value })}
              className="form-input"
            >
              <option value="DIRECT">DIRECT</option>
              <option value="REJECT">REJECT</option>
              {proxyGroups.map(g => (
                <option key={g.name} value={g.name}>{g.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>说明</label>
            <input
              type="text"
              value={rule.description}
              onChange={(e) => onChange({ ...rule, description: e.target.value })}
              className="form-input"
              placeholder="规则用途描述"
            />
          </div>

          {(rule.type === 'IP-CIDR' || rule.type === 'GEOIP' || rule.type === 'RULE-SET') && (
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={rule.noResolve}
                onChange={(e) => onChange({ ...rule, noResolve: e.target.checked })}
                className="rounded"
              />
              <span className={cn(
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>no-resolve（不解析域名）</span>
            </label>
          )}
        </div>

        <div className="flex justify-end gap-2 mt-6">
          <button 
            onClick={onCancel} 
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium',
              themeStyle === 'apple-glass'
                ? 'hover:bg-black/5 text-slate-600'
                : 'hover:bg-white/10 text-slate-400'
            )}
          >
            取消
          </button>
          <button
            onClick={onSave}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 hover:bg-blue-600'
                : 'bg-cyan-500 hover:bg-cyan-600'
            )}
          >
            <Save className="w-4 h-4" />
            保存
          </button>
        </div>
      </div>
    </div>
  )
}
