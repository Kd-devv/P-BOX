import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Palette, Info, Check, Server, Power, Zap, Rocket, Gauge, ArrowUpDown, Loader2, Lock, Shield, ChevronRight } from 'lucide-react'
import { useThemeStore, ThemeStyle } from '@/stores/themeStore'
import { systemApi, SystemConfig } from '@/api/system'
import { authApi } from '@/api/auth'
import { cn } from '@/lib/utils'

export default function SettingsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { themeStyle, setThemeStyle } = useThemeStore()
  const [sysConfig, setSysConfig] = useState<SystemConfig | null>(null)
  const [sysLoading, setSysLoading] = useState(true)
  const [authEnabled, setAuthEnabled] = useState(false)

  useEffect(() => { fetchSysConfig(); fetchAuthConfig() }, [])

  const fetchAuthConfig = async () => { try { const cfg = await authApi.getConfig(); setAuthEnabled(cfg.enabled) } catch {} }
  const handleAuthToggle = async () => { try { await authApi.setEnabled(!authEnabled); setAuthEnabled(!authEnabled) } catch {} }
  const fetchSysConfig = async () => { try { setSysLoading(true); setSysConfig(await systemApi.getConfig()) } catch {} finally { setSysLoading(false) } }

  const handleSysToggle = async (key: keyof SystemConfig, setter: (e: boolean) => Promise<unknown>) => {
    if (!sysConfig) return
    try { await setter(!sysConfig[key]); setSysConfig({ ...sysConfig, [key]: !sysConfig[key] }) } catch {}
  }

  const handleOptimizeAll = async () => { try { await systemApi.optimizeAll(); await fetchSysConfig() } catch {} }

  const themeStyles: { id: ThemeStyle; label: string; description: string }[] = [
    { id: 'apple-glass', label: t('settings.appleGlass'), description: t('settings.glassDescription') },
    { id: 'apple-pro-dark', label: t('settings.appleProDark'), description: t('settings.proDarkDescription') },
  ]

  const Toggle = ({ value, onChange }: { value: boolean; onChange: () => void }) => (
    <button onClick={onChange} className={cn('w-12 h-6 rounded-full transition-colors relative', value ? (themeStyle === 'apple-glass' ? 'bg-blue-500' : 'bg-cyan-500') : 'bg-slate-600')}>
      <span className={cn('absolute top-1 w-4 h-4 rounded-full bg-white transition-all', value ? 'left-7' : 'left-1')} />
    </button>
  )

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      {/* 主题 */}
      <div className="glass-card p-5">
        <h3 className={cn('text-sm font-medium mb-4 flex items-center gap-2', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}><Palette className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />{t('settings.themeStyle')}</h3>
        <div className="grid grid-cols-2 gap-3">
          {themeStyles.map((theme) => (
            <button key={theme.id} onClick={() => setThemeStyle(theme.id)} className={cn('relative p-3 rounded-xl text-left transition-all border-2', themeStyle === theme.id ? (themeStyle === 'apple-glass' ? 'bg-blue-500/10 border-blue-500/50' : 'bg-cyan-500/10 border-cyan-500/50') : (themeStyle === 'apple-glass' ? 'bg-black/[0.02] border-transparent' : 'bg-white/5 border-transparent'))}>
              {themeStyle === theme.id && <div className="absolute top-2 right-2"><Check className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} /></div>}
              <div className={cn('font-medium text-sm', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>{theme.label}</div>
              <div className={cn('text-xs mt-0.5', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-neutral-400')}>{theme.description}</div>
            </button>
          ))}
        </div>
      </div>

      {/* 安全 */}
      <div className="glass-card p-5">
        <h3 className={cn('text-sm font-medium mb-4 flex items-center gap-2', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}><Lock className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />{t('settings.security')}</h3>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div><span className={cn('text-sm', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300')}>{t('settings.enableAuth')}</span><p className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')}>{t('settings.enableAuthDesc')}</p></div>
            <Toggle value={authEnabled} onChange={handleAuthToggle} />
          </div>
          {/* 安全与隐私政策 */}
          <button
            onClick={() => navigate('/legal')}
            className={cn(
              'w-full flex items-center justify-between p-3 rounded-xl transition-all',
              themeStyle === 'apple-glass' 
                ? 'bg-gradient-to-r from-amber-500/10 to-orange-500/10 hover:from-amber-500/20 hover:to-orange-500/20 border border-amber-500/20' 
                : 'bg-gradient-to-r from-amber-500/10 to-orange-500/10 hover:from-amber-500/20 hover:to-orange-500/20 border border-amber-500/20'
            )}
          >
            <div className="flex items-center gap-3">
              <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center', themeStyle === 'apple-glass' ? 'bg-amber-500/20' : 'bg-amber-500/20')}>
                <Shield className="w-4 h-4 text-amber-500" />
              </div>
              <div className="text-left">
                <div className={cn('text-sm font-medium', themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-white')}>
                  {t('legal.securityPolicy')}
                </div>
                <div className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>
                  {t('legal.securityPolicyDesc')}
                </div>
              </div>
            </div>
            <ChevronRight className={cn('w-5 h-5', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')} />
          </button>
        </div>
      </div>

      {/* 系统 */}
      <div className="glass-card p-5 lg:col-span-2">
        <h3 className={cn('text-sm font-medium mb-4 flex items-center gap-2', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}><Server className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />{t('settings.systemSettings')}</h3>
        {sysLoading ? (<div className="flex items-center justify-center py-6"><Loader2 className={cn('w-6 h-6 animate-spin', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} /></div>) : sysConfig ? (
          <div className="space-y-4">
            <div className={cn('flex items-center justify-between p-4 rounded-xl', themeStyle === 'apple-glass' ? 'bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20' : 'bg-gradient-to-r from-cyan-500/10 to-purple-500/10 border border-cyan-500/20')}>
              <div className="flex items-center gap-3"><div className={cn('w-10 h-10 rounded-full flex items-center justify-center', themeStyle === 'apple-glass' ? 'bg-blue-500/20' : 'bg-cyan-500/20')}><Rocket className={cn('w-5 h-5', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} /></div><div><div className={cn('font-medium', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>{t('settings.oneClickOptimize')}</div><div className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')}>{t('settings.optimizeDesc')}</div></div></div>
              <button onClick={handleOptimizeAll} className={cn('flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white', themeStyle === 'apple-glass' ? 'bg-blue-500 hover:bg-blue-600' : 'bg-cyan-500 hover:bg-cyan-600')}><Zap className="w-4 h-4" />{t('settings.optimizeNow')}</button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {[{ key: 'autoStart' as const, icon: Power, label: t('settings.autoStart'), desc: t('settings.autoStartDesc'), setter: systemApi.setAutoStart },{ key: 'ipForward' as const, icon: ArrowUpDown, label: t('settings.ipForward'), desc: t('settings.ipForwardDesc'), setter: systemApi.setIPForward },{ key: 'bbrEnabled' as const, icon: Gauge, label: t('settings.bbr'), desc: t('settings.bbrDesc'), setter: systemApi.setBBR },{ key: 'tunOptimized' as const, icon: Rocket, label: t('settings.tunOptimize'), desc: t('settings.tunOptimizeDesc'), setter: systemApi.setTUNOptimize }].map(({ key, icon: Icon, label, desc, setter }) => (
                <div key={key} className="flex items-center justify-between"><div className="flex items-center gap-3"><Icon className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400')} /><div><span className={cn('text-sm', themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300')}>{label}</span><p className={cn('text-xs', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')}>{desc}</p></div></div><Toggle value={sysConfig[key]} onChange={() => handleSysToggle(key, setter)} /></div>
              ))}
            </div>
            <p className={cn('text-xs text-center', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')}>{t('settings.sysNote')}</p>
          </div>
        ) : (<div className={cn('text-center py-6', themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500')}><Server className="w-10 h-10 mx-auto mb-3 opacity-50" /><p className="text-sm">{t('settings.sysNotAvailable')}</p></div>)}
      </div>

      {/* 关于 */}
      <div className="glass-card p-5 lg:col-span-2">
        <h3 className={cn('text-sm font-medium mb-4 flex items-center gap-2', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}><Info className={cn('w-4 h-4', themeStyle === 'apple-glass' ? 'text-blue-500' : 'text-cyan-400')} />{t('settings.about')}</h3>
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div><span className={themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-neutral-500'}>{t('settings.version')}</span><p className={cn('font-mono', themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white')}>v1.0</p></div>
          <div><span className={themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-neutral-500'}>{t('settings.frontend')}</span><p className={themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'}>React + TS</p></div>
          <div><span className={themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-neutral-500'}>{t('settings.uiLibrary')}</span><p className={themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'}>Tailwind</p></div>
        </div>
      </div>
    </div>
  )
}
