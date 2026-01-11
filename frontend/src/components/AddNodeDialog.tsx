import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { nodeApi, PROTOCOL_LIST, ProtocolField } from '@/api/node'

interface AddNodeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export function AddNodeDialog({ open, onOpenChange, onSuccess }: AddNodeDialogProps) {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [protocol, setProtocol] = useState('vmess')
  const [name, setName] = useState('')
  const [server, setServer] = useState('')
  const [serverPort, setServerPort] = useState(443)
  const [config, setConfig] = useState<Record<string, unknown>>({})
  const [fields, setFields] = useState<ProtocolField[]>([])
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  // 加载协议字段定义
  useEffect(() => {
    if (protocol && open) {
      loadProtocolFields(protocol)
    }
  }, [protocol, open])

  // 重置表单
  useEffect(() => {
    if (!open) {
      setName('')
      setServer('')
      setServerPort(443)
      setConfig({})
    }
  }, [open])

  const loadProtocolFields = async (proto: string) => {
    setLoading(true)
    setConfig({}) // 先清空旧配置
    try {
      const data = await nodeApi.getProtocolFields(proto)
      const fieldDefs = data.fields || []
      setFields(fieldDefs)
      
      // 设置默认值
      const defaultConfig: Record<string, unknown> = {}
      fieldDefs.forEach(field => {
        if (field.default !== undefined) {
          defaultConfig[field.name] = field.default
        }
      })
      setConfig(defaultConfig)
    } catch (e) {
      console.error('加载协议字段失败:', e)
      setFields([])
      setConfig({})
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async () => {
    if (!name || !server || !serverPort) {
      alert(t('nodes.fillRequiredFields'))
      return
    }

    setSubmitting(true)
    try {
      await nodeApi.addManualAdvanced({
        name,
        type: protocol,
        server,
        server_port: serverPort,
        config
      })
      onSuccess()
      onOpenChange(false)
    } catch (e) {
      console.error('添加节点失败:', e)
      alert(t('nodes.addFailed'))
    } finally {
      setSubmitting(false)
    }
  }

  const updateConfig = (fieldName: string, value: unknown) => {
    setConfig(prev => ({ ...prev, [fieldName]: value }))
  }

  // 检查字段是否应该显示（基于依赖）
  const shouldShowField = (field: ProtocolField): boolean => {
    if (!field.depends_on) return true
    return config[field.depends_on] === field.depends_value
  }

  // 渲染字段
  const renderField = (field: ProtocolField) => {
    if (!shouldShowField(field)) return null

    const inputClass = cn(
      'w-full px-3 py-2 rounded-lg text-sm',
      themeStyle === 'apple-glass'
        ? 'bg-white/60 border border-black/10 text-slate-900 placeholder:text-slate-400'
        : 'bg-white/5 border border-white/10 text-slate-200 placeholder:text-slate-500'
    )

    switch (field.type) {
      case 'select':
        return (
          <select
            value={String(config[field.name] ?? field.default ?? '')}
            onChange={(e) => updateConfig(field.name, e.target.value)}
            className={inputClass}
          >
            {field.options?.map(opt => (
              <option key={String(opt.value)} value={String(opt.value)}>
                {opt.label}
              </option>
            ))}
          </select>
        )

      case 'boolean':
        return (
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={Boolean(config[field.name])}
              onChange={(e) => updateConfig(field.name, e.target.checked)}
              className="w-4 h-4 rounded"
            />
            <span className={cn(
              'text-sm',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>
              {field.description || field.label}
            </span>
          </label>
        )

      case 'number':
        return (
          <input
            type="number"
            value={Number(config[field.name] ?? field.default ?? 0)}
            onChange={(e) => updateConfig(field.name, Number(e.target.value))}
            min={field.min}
            max={field.max}
            placeholder={field.placeholder}
            className={inputClass}
          />
        )

      case 'textarea':
        return (
          <textarea
            value={String(config[field.name] ?? '')}
            onChange={(e) => updateConfig(field.name, e.target.value)}
            placeholder={field.placeholder}
            rows={3}
            className={inputClass}
          />
        )

      case 'password':
        return (
          <input
            type="password"
            value={String(config[field.name] ?? '')}
            onChange={(e) => updateConfig(field.name, e.target.value)}
            placeholder={field.placeholder}
            className={inputClass}
          />
        )

      default:
        return (
          <input
            type="text"
            value={String(config[field.name] ?? '')}
            onChange={(e) => updateConfig(field.name, e.target.value)}
            placeholder={field.placeholder}
            className={inputClass}
          />
        )
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className={cn(
        'w-full max-w-lg max-h-[85vh] overflow-hidden rounded-2xl shadow-2xl',
        themeStyle === 'apple-glass'
          ? 'bg-white/90 backdrop-blur-xl'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        {/* Header */}
        <div className={cn(
          'flex items-center justify-between px-6 py-4 border-b',
          themeStyle === 'apple-glass' ? 'border-black/10' : 'border-white/10'
        )}>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-900' : 'text-white'
          )}>
            {t('nodes.addNode')}
          </h2>
          <button
            onClick={() => onOpenChange(false)}
            className={cn(
              'p-1 rounded-lg transition-colors',
              themeStyle === 'apple-glass'
                ? 'hover:bg-black/5 text-slate-600'
                : 'hover:bg-white/10 text-slate-400'
            )}
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-6 overflow-y-auto max-h-[calc(85vh-140px)] space-y-4">
          {/* 协议选择 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>
              {t('nodes.protocolType')} *
            </label>
            <select
              value={protocol}
              onChange={(e) => setProtocol(e.target.value)}
              className={cn(
                'w-full px-3 py-2 rounded-lg text-sm',
                themeStyle === 'apple-glass'
                  ? 'bg-white/60 border border-black/10 text-slate-900'
                  : 'bg-white/5 border border-white/10 text-slate-200'
              )}
            >
              {PROTOCOL_LIST.map(p => (
                <option key={p.value} value={p.value}>{p.label}</option>
              ))}
            </select>
          </div>

          {/* 基础字段 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>
                {t('nodes.nodeName')} *
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('nodes.nodeNamePlaceholder')}
                className={cn(
                  'w-full px-3 py-2 rounded-lg text-sm',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/60 border border-black/10 text-slate-900 placeholder:text-slate-400'
                    : 'bg-white/5 border border-white/10 text-slate-200 placeholder:text-slate-500'
                )}
              />
            </div>
            <div>
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>
                {t('nodes.serverAddress')} *
              </label>
              <input
                type="text"
                value={server}
                onChange={(e) => setServer(e.target.value)}
                placeholder="example.com"
                className={cn(
                  'w-full px-3 py-2 rounded-lg text-sm',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/60 border border-black/10 text-slate-900 placeholder:text-slate-400'
                    : 'bg-white/5 border border-white/10 text-slate-200 placeholder:text-slate-500'
                )}
              />
            </div>
            <div>
              <label className={cn(
                'block text-sm font-medium mb-1.5',
                themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
              )}>
                {t('nodes.port')} *
              </label>
              <input
                type="number"
                value={serverPort}
                onChange={(e) => setServerPort(Number(e.target.value))}
                min={1}
                max={65535}
                className={cn(
                  'w-full px-3 py-2 rounded-lg text-sm',
                  themeStyle === 'apple-glass'
                    ? 'bg-white/60 border border-black/10 text-slate-900'
                    : 'bg-white/5 border border-white/10 text-slate-200'
                )}
              />
            </div>
          </div>

          {/* 协议特定字段 */}
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className={cn(
                'w-6 h-6 animate-spin',
                themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400'
              )} />
            </div>
          ) : fields.length > 0 && (
            <div className="space-y-4">
              <div className={cn(
                'text-xs font-medium uppercase tracking-wider',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
              )}>
                {t('nodes.protocolConfig')}
              </div>
              {fields.map(field => {
                if (!shouldShowField(field)) return null
                return (
                  <div key={field.name}>
                    {field.type !== 'boolean' && (
                      <label className={cn(
                        'block text-sm font-medium mb-1.5',
                        themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
                      )}>
                        {field.label} {field.required && '*'}
                      </label>
                    )}
                    {renderField(field)}
                    {field.description && field.type !== 'boolean' && (
                      <p className={cn(
                        'mt-1 text-xs',
                        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-500'
                      )}>
                        {field.description}
                      </p>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className={cn(
          'flex justify-end gap-3 px-6 py-4 border-t',
          themeStyle === 'apple-glass' ? 'border-black/10' : 'border-white/10'
        )}>
          <button
            onClick={() => onOpenChange(false)}
            disabled={submitting}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              themeStyle === 'apple-glass'
                ? 'bg-black/5 text-slate-700 hover:bg-black/10'
                : 'bg-white/5 text-slate-300 hover:bg-white/10'
            )}
          >
            {t('common.cancel')}
          </button>
          <button
            onClick={handleSubmit}
            disabled={submitting || !name || !server}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors flex items-center gap-2',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 hover:bg-blue-600 disabled:opacity-50'
                : 'bg-cyan-500 hover:bg-cyan-600 disabled:opacity-50'
            )}
          >
            {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
            {t('nodes.addNode')}
          </button>
        </div>
      </div>
    </div>
  )
}
