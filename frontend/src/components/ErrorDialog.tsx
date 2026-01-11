import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { X, Copy, Check, AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

interface ErrorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  error: string
}

export function ErrorDialog({ open, onOpenChange, title, error }: ErrorDialogProps) {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(error)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (e) {
      console.error('复制失败:', e)
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className={cn(
        'w-full max-w-2xl max-h-[80vh] overflow-hidden rounded-2xl shadow-2xl',
        themeStyle === 'apple-glass'
          ? 'bg-white/95 backdrop-blur-xl'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        {/* Header */}
        <div className={cn(
          'flex items-center justify-between px-6 py-4 border-b',
          themeStyle === 'apple-glass' ? 'border-black/10' : 'border-white/10'
        )}>
          <div className="flex items-center gap-3">
            <div className={cn(
              'p-2 rounded-lg',
              themeStyle === 'apple-glass' ? 'bg-red-100' : 'bg-red-500/20'
            )}>
              <AlertTriangle className={cn(
                'w-5 h-5',
                themeStyle === 'apple-glass' ? 'text-red-600' : 'text-red-400'
              )} />
            </div>
            <h2 className={cn(
              'text-lg font-semibold',
              themeStyle === 'apple-glass' ? 'text-slate-900' : 'text-white'
            )}>
              {title}
            </h2>
          </div>
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
        <div className="p-6">
          <div className={cn(
            'relative rounded-lg p-4 font-mono text-sm overflow-auto max-h-[50vh]',
            themeStyle === 'apple-glass'
              ? 'bg-slate-100 text-slate-800'
              : 'bg-black/30 text-slate-200'
          )}>
            <pre className="whitespace-pre-wrap break-words">{error}</pre>
          </div>
        </div>

        {/* Footer */}
        <div className={cn(
          'flex justify-end gap-3 px-6 py-4 border-t',
          themeStyle === 'apple-glass' ? 'border-black/10' : 'border-white/10'
        )}>
          <button
            onClick={handleCopy}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              copied
                ? themeStyle === 'apple-glass'
                  ? 'bg-green-100 text-green-700'
                  : 'bg-green-500/20 text-green-400'
                : themeStyle === 'apple-glass'
                  ? 'bg-black/5 text-slate-700 hover:bg-black/10'
                  : 'bg-white/5 text-slate-300 hover:bg-white/10'
            )}
          >
            {copied ? (
              <>
                <Check className="w-4 h-4" />
                {t('common.copied') || '已复制'}
              </>
            ) : (
              <>
                <Copy className="w-4 h-4" />
                {t('common.copy')}
              </>
            )}
          </button>
          <button
            onClick={() => onOpenChange(false)}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 hover:bg-blue-600'
                : 'bg-cyan-500 hover:bg-cyan-600'
            )}
          >
            {t('common.close')}
          </button>
        </div>
      </div>
    </div>
  )
}
