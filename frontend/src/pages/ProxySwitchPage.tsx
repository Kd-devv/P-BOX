import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { RefreshCw, Check, GripVertical, Zap } from 'lucide-react'
import { DndContext, closestCenter, DragEndEvent } from '@dnd-kit/core'
import { SortableContext, useSortable, verticalListSortingStrategy, arrayMove } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { mihomoApi, ProxyNode } from '@/api/mihomo'
import { cn, getLatencyColor, formatLatency } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { getGroupIcon, getGroupIconColor, getGroupOrder } from '@/lib/proxyGroups'

interface ProxyGroup {
  name: string
  type: string
  now?: string
  all: string[]
}

// 默认代理组名称翻译映射 (中文 <-> 英文)
const GROUP_NAME_ZH_TO_EN: Record<string, string> = {
  '自动选择': 'Auto Select',
  '故障转移': 'Failover',
  '节点选择': 'Node Select',
  '全球直连': 'Direct',
  'AI服务': 'AI Services',
  '国外媒体': 'Streaming',
  '电报消息': 'Telegram',
  '谷歌服务': 'Google',
  '推特消息': 'Twitter',
  '脸书服务': 'Facebook',
  '游戏平台': 'Gaming',
  '哔哩哔哩': 'Bilibili',
  '微软服务': 'Microsoft',
  '苹果服务': 'Apple',
  '广告拦截': 'Ad Block',
  '漏网之鱼': 'Final',
  '香港节点': 'Hong Kong',
  '台湾节点': 'Taiwan',
  '日本节点': 'Japan',
  '新加坡节点': 'Singapore',
  '美国节点': 'United States',
  '手动节点': 'Manual',
  '其他节点': 'Others',
}

// 创建反向映射 (英文 -> 中文)
const GROUP_NAME_EN_TO_ZH: Record<string, string> = Object.fromEntries(
  Object.entries(GROUP_NAME_ZH_TO_EN).map(([zh, en]) => [en, zh])
)

// 翻译分组名称
const translateGroupName = (name: string, lang: string): string => {
  if (lang.startsWith('zh')) {
    // 中文界面：英文名 -> 中文名
    return GROUP_NAME_EN_TO_ZH[name] || name
  } else {
    // 英文界面：中文名 -> 英文名
    return GROUP_NAME_ZH_TO_EN[name] || name
  }
}

// 翻译分组类型
const getGroupTypeName = (type: string, t: (key: string) => string): string => {
  const lowerType = type.toLowerCase()
  const typeMap: Record<string, string> = {
    'selector': t('proxy.selector'),
    'urltest': t('proxy.urlTest'),
    'fallback': t('proxy.fallback'),
    'loadbalance': t('proxy.loadBalance'),
    'load-balance': t('proxy.loadBalance'),
  }
  return typeMap[lowerType] || type
}

// 可拖拽的分组项组件
function SortableGroupItem({ 
  group, 
  isSelected, 
  onClick, 
  themeStyle 
}: { 
  group: ProxyGroup
  isSelected: boolean
  onClick: () => void
  themeStyle: string
}) {
  const { t, i18n } = useTranslation()
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: group.name })
  const Icon = getGroupIcon(group.name)
  const displayName = translateGroupName(group.name, i18n.language)
  
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    zIndex: isDragging ? 1000 : 'auto',
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'flex items-center gap-2 px-2 py-2 rounded-xl text-sm transition-all duration-200 cursor-pointer',
        isSelected
          ? themeStyle === 'apple-glass'
            ? 'bg-blue-500/15 border border-blue-500/30'
            : 'bg-cyan-500/20 border border-cyan-500/30'
          : themeStyle === 'apple-glass'
            ? 'hover:bg-black/5 border border-transparent'
            : 'hover:bg-white/5 border border-transparent'
      )}
      onClick={onClick}
    >
      {/* 拖拽手柄 */}
      <div
        {...attributes}
        {...listeners}
        className={cn(
          'p-1 rounded cursor-grab active:cursor-grabbing',
          themeStyle === 'apple-glass' ? 'text-slate-400 hover:text-slate-600' : 'text-slate-500 hover:text-slate-300'
        )}
      >
        <GripVertical className="w-4 h-4" />
      </div>
      
      {/* 扁平化彩色图标 */}
      <div className={cn(
        'w-8 h-8 rounded-xl flex items-center justify-center flex-shrink-0 text-white shadow-sm',
        getGroupIconColor(group.name)
      )}>
        <Icon className="w-4 h-4" />
      </div>
      
      <div className="flex-1 min-w-0">
        <div className={cn(
          'font-medium truncate text-sm',
          isSelected
            ? themeStyle === 'apple-glass' ? 'text-blue-600' : 'text-cyan-400'
            : themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-200'
        )}>{displayName}</div>
        <div className={cn(
          'text-xs truncate',
          themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
        )}>
          {getGroupTypeName(group.type, t)} · {group.all.length}
        </div>
      </div>
    </div>
  )
}

export default function ProxySwitchPage() {
  const { t, i18n } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [groups, setGroups] = useState<ProxyGroup[]>([])
  const [nodes, setNodes] = useState<Record<string, ProxyNode>>({})
  const [selectedGroup, setSelectedGroup] = useState<string>('')
  const [testing, setTesting] = useState<string | null>(null)
  const [delays, setDelays] = useState<Record<string, number>>({})

  useEffect(() => {
    fetchProxies()
  }, [])

  const fetchProxies = async () => {
    try {
      const proxies = await mihomoApi.getProxies()
      setNodes(proxies)
      
      // Extract groups and sort by predefined order
      const groupList: ProxyGroup[] = []
      for (const [name, node] of Object.entries(proxies)) {
        if (node.all && node.all.length > 0) {
          groupList.push({
            name,
            type: node.type,
            now: node.now,
            all: node.all
          })
        }
      }
      // Sort by predefined order
      groupList.sort((a, b) => getGroupOrder(a.name) - getGroupOrder(b.name))
      setGroups(groupList)
      if (groupList.length > 0 && !selectedGroup) {
        setSelectedGroup(groupList[0].name)
      }
    } catch {
      // Ignore errors
    }
  }

  const handleSelect = async (groupName: string, nodeName: string) => {
    try {
      await mihomoApi.selectProxy(groupName, nodeName)
      await fetchProxies()
    } catch {
      // Ignore errors
    }
  }

  const handleTest = async (nodeName: string) => {
    setTesting(nodeName)
    try {
      const delay = await mihomoApi.testDelay(nodeName)
      setDelays(prev => ({ ...prev, [nodeName]: delay }))
    } finally {
      setTesting(null)
    }
  }

  const handleTestAll = async () => {
    const group = groups.find(g => g.name === selectedGroup)
    if (!group) return

    for (const nodeName of group.all) {
      await handleTest(nodeName)
    }
  }

  const currentGroup = groups.find(g => g.name === selectedGroup)

  // 拖拽结束处理
  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (over && active.id !== over.id) {
      setGroups((items) => {
        const oldIndex = items.findIndex((i) => i.name === active.id)
        const newIndex = items.findIndex((i) => i.name === over.id)
        return arrayMove(items, oldIndex, newIndex)
      })
    }
  }

  return (
    <div className="flex flex-col lg:flex-row gap-4 h-[calc(100vh-8rem)]">
      {/* Mobile: Group selector dropdown */}
      <div className="lg:hidden glass-card p-3">
        <select
          value={selectedGroup}
          onChange={(e) => setSelectedGroup(e.target.value)}
          className={cn(
            'w-full px-3 py-2 rounded-lg text-sm font-medium appearance-none cursor-pointer',
            themeStyle === 'apple-glass'
              ? 'bg-white/60 border border-black/10 text-slate-700'
              : 'bg-white/10 border border-white/10 text-white'
          )}
        >
          {groups.map((group) => (
            <option key={group.name} value={group.name} className="bg-slate-800">
              {translateGroupName(group.name, i18n.language)} ({group.all.length})
            </option>
          ))}
        </select>
      </div>

      {/* Desktop: Groups list */}
      <div className="hidden lg:block w-72 glass-card p-3 overflow-auto flex-shrink-0">
        <div className={cn(
          "text-xs uppercase tracking-wider mb-3 px-2",
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
        )}>
          {t('proxy.groups')}
        </div>
        <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
          <SortableContext items={groups.map(g => g.name)} strategy={verticalListSortingStrategy}>
            <div className="space-y-1">
              {groups.map((group) => (
                <SortableGroupItem
                  key={group.name}
                  group={group}
                  isSelected={selectedGroup === group.name}
                  onClick={() => setSelectedGroup(group.name)}
                  themeStyle={themeStyle}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      </div>

      {/* Nodes grid */}
      <div className="flex-1 glass-card p-4 overflow-auto">
        <div className="flex items-center justify-between mb-4">
          <div className={cn(
            "text-sm font-medium",
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {currentGroup ? translateGroupName(currentGroup.name, i18n.language) : ''} ({currentGroup?.all.length || 0} {t('proxy.nodes')})
          </div>
          <button
            onClick={handleTestAll}
            className="control-btn secondary text-xs"
          >
            <Zap className="w-3 h-3" />
            {t('proxy.testAll')}
          </button>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-2 sm:gap-3">
          {currentGroup?.all.map((nodeName) => {
            const node = nodes[nodeName]
            const isSelected = currentGroup.now === nodeName
            const delay = delays[nodeName] ?? node?.history?.[0]?.delay ?? 0

            return (
              <div
                key={nodeName}
                onClick={() => handleSelect(selectedGroup, nodeName)}
                className={cn(
                  'p-3 rounded-xl cursor-pointer transition-all duration-200 group',
                  isSelected
                    ? 'bg-primary/20 border-2 border-primary/50'
                    : 'bg-white/5 border-2 border-transparent hover:bg-white/10 hover:border-white/20'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-[10px] font-mono text-muted-foreground uppercase bg-white/5 px-1.5 py-0.5 rounded">
                    {node?.type || 'Unknown'}
                  </span>
                  {isSelected && (
                    <Check className="w-4 h-4 text-primary" />
                  )}
                </div>
                <div className="font-medium text-foreground text-sm truncate mb-2">
                  {nodeName}
                </div>
                <div className="flex items-center justify-between">
                  <span className={cn('text-xs font-mono', getLatencyColor(delay))}>
                    {formatLatency(delay)}
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleTest(nodeName)
                    }}
                    className="opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <RefreshCw className={cn(
                      'w-3.5 h-3.5 text-muted-foreground hover:text-foreground',
                      testing === nodeName && 'animate-spin'
                    )} />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
