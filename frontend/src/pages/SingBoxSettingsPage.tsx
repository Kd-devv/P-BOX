import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { 
  Globe, Zap, RotateCcw, Layers, ArrowRightLeft,
  ChevronDown, ChevronUp, Network, Server, Monitor
} from 'lucide-react'
import { singboxApi, SingBoxSettings, defaultSingBoxSettings, defaultTunSettings, defaultSystemSettings, detectPlatform, getPlatformDefaultSettings } from '@/api/singbox'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'

// 设置分组组件
function SettingsSection({ 
  title, 
  icon: Icon, 
  children, 
  defaultOpen = true,
  themeStyle
}: { 
  title: string
  icon: React.ElementType
  children: React.ReactNode
  defaultOpen?: boolean
  themeStyle: string
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen)

  return (
    <div className="glass-card overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'w-full flex items-center justify-between p-4 transition-colors',
          themeStyle === 'apple-glass' 
            ? 'hover:bg-black/5' 
            : 'hover:bg-white/5'
        )}
      >
        <div className="flex items-center gap-3">
          <div className={cn(
            'w-8 h-8 rounded-lg flex items-center justify-center',
            themeStyle === 'apple-glass'
              ? 'bg-purple-500/10 text-purple-600'
              : 'bg-purple-500/20 text-purple-400'
          )}>
            <Icon className="w-4 h-4" />
          </div>
          <span className={cn(
            'font-medium text-sm',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{title}</span>
        </div>
        {isOpen ? <ChevronUp className="w-4 h-4 text-muted-foreground" /> : <ChevronDown className="w-4 h-4 text-muted-foreground" />}
      </button>
      {isOpen && (
        <div className={cn(
          'p-4 pt-0 space-y-4',
          themeStyle === 'apple-glass' ? 'border-t border-black/5' : 'border-t border-white/5'
        )}>
          {children}
        </div>
      )}
    </div>
  )
}

// 表单控件组件
function FormField({ 
  label, 
  description, 
  children,
  themeStyle
}: { 
  label: React.ReactNode
  description?: string
  children: React.ReactNode
  themeStyle: string
}) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 sm:gap-4 py-2">
      <div className="flex-1 min-w-0">
        <div className={cn(
          'text-sm font-medium',
          themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-200'
        )}>{label}</div>
        {description && (
          <div className={cn(
            'text-xs mt-0.5',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>{description}</div>
        )}
      </div>
      <div className="flex-shrink-0">{children}</div>
    </div>
  )
}

// 开关组件
function Toggle({ 
  value, 
  onChange, 
  disabled,
  themeStyle 
}: { 
  value: boolean
  onChange: (v: boolean) => void
  disabled?: boolean
  themeStyle: string
}) {
  return (
    <button
      onClick={() => !disabled && onChange(!value)}
      disabled={disabled}
      className={cn(
        'w-12 h-6 rounded-full transition-all relative',
        value 
          ? (themeStyle === 'apple-glass' ? 'bg-purple-500' : 'bg-purple-500') 
          : (themeStyle === 'apple-glass' ? 'bg-slate-300' : 'bg-slate-600'),
        disabled && 'opacity-50 cursor-not-allowed'
      )}
    >
      <span className={cn(
        'absolute top-1 w-4 h-4 rounded-full bg-white transition-all shadow-sm',
        value ? 'left-7' : 'left-1'
      )} />
    </button>
  )
}

// 选择框组件
function Select<T extends string>({ 
  value, 
  onChange, 
  options,
  themeStyle
}: { 
  value: T
  onChange: (v: T) => void
  options: { value: T; label: string }[]
  themeStyle: string
}) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value as T)}
      className={cn(
        'px-3 py-1.5 rounded-lg text-sm border transition-colors appearance-none cursor-pointer min-w-[140px]',
        themeStyle === 'apple-glass'
          ? 'bg-white border-slate-200 text-slate-700 focus:border-purple-500'
          : 'bg-white/10 border-white/10 text-white focus:border-purple-500'
      )}
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value} className="bg-slate-800 text-white">
          {opt.label}
        </option>
      ))}
    </select>
  )
}

// 数字输入组件
function NumberInput({ 
  value, 
  onChange, 
  min, 
  max,
  themeStyle
}: { 
  value: number
  onChange: (v: number) => void
  min?: number
  max?: number
  themeStyle: string
}) {
  return (
    <input
      type="number"
      value={value}
      onChange={(e) => onChange(parseInt(e.target.value) || 0)}
      min={min}
      max={max}
      className={cn(
        'px-3 py-1.5 rounded-lg text-sm border transition-colors w-24 text-center',
        themeStyle === 'apple-glass'
          ? 'bg-white border-slate-200 text-slate-700 focus:border-purple-500'
          : 'bg-white/10 border-white/10 text-white focus:border-purple-500'
      )}
    />
  )
}

// 文本输入组件
function TextInput({ 
  value, 
  onChange, 
  placeholder,
  themeStyle
}: { 
  value: string
  onChange: (v: string) => void
  placeholder?: string
  themeStyle: string
}) {
  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className={cn(
        'px-3 py-1.5 rounded-lg text-sm border transition-colors w-48',
        themeStyle === 'apple-glass'
          ? 'bg-white border-slate-200 text-slate-700 focus:border-purple-500 placeholder:text-slate-400'
          : 'bg-white/10 border-white/10 text-white focus:border-purple-500 placeholder:text-slate-500'
      )}
    />
  )
}

export default function SingBoxSettingsPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [settings, setSettings] = useState<SingBoxSettings>(defaultSingBoxSettings)
  
  // 加载设置
  useEffect(() => {
    setSettings(singboxApi.loadSettings())
  }, [])

  // 更新设置
  const updateSettings = <K extends keyof SingBoxSettings>(key: K, value: SingBoxSettings[K]) => {
    const newSettings = { ...settings, [key]: value }
    setSettings(newSettings)
    singboxApi.saveSettings(newSettings)
  }

  // 切换模式并应用最优配置
  const handleModeChange = (mode: 'tun' | 'system') => {
    const optimalSettings = mode === 'tun' 
      ? { ...settings, ...defaultTunSettings }
      : { ...settings, ...defaultSystemSettings }
    setSettings(optimalSettings)
    singboxApi.saveSettings(optimalSettings)
  }

  // 重置设置 (使用平台默认)
  const handleReset = () => {
    const platformSettings = getPlatformDefaultSettings()
    setSettings(platformSettings)
    singboxApi.saveSettings(platformSettings)
  }

  // 获取平台名称
  const getPlatformName = () => {
    const platform = detectPlatform()
    switch (platform) {
      case 'macos': return 'macOS'
      case 'windows': return 'Windows'
      case 'linux': return 'Linux'
      default: return t('singboxSettings.unknown')
    }
  }

  return (
    <div className="space-y-4 sm:space-y-6 pb-20">
      {/* 顶部操作栏 - 与 Mihomo 设置页面一致 */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>{t('singboxSettings.title')}</h2>
          <p className={cn(
            'text-sm mt-1',
            themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
          )}>
            {t('singboxSettings.platform')}: {getPlatformName()} · {t('singboxSettings.recommend')}: {detectPlatform() === 'macos' ? t('singboxSettings.systemProxy') : t('singboxSettings.tunMode')}
          </p>
        </div>
        
        <div className="flex items-center gap-3">
          {/* 模式选择器 - 扁平化苹果风格 */}
          <div className={cn(
            'flex items-center gap-0.5 p-1 rounded-lg border h-8',
            themeStyle === 'apple-glass'
              ? 'bg-white/40 border-black/10'
              : 'bg-white/5 border-white/10'
          )}>
            <button
              onClick={() => handleModeChange('tun')}
              className={cn(
                'flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium transition-all',
                settings.mode === 'tun'
                  ? (themeStyle === 'apple-glass' ? 'bg-purple-500 text-white shadow-sm' : 'bg-purple-500 text-white')
                  : (themeStyle === 'apple-glass' ? 'text-slate-600 hover:bg-black/5' : 'text-slate-400 hover:bg-white/10')
              )}
              title="TUN - 虚拟网卡模式"
            >
              <Layers className="w-3 h-3" />
              <span className="hidden sm:inline">TUN</span>
            </button>
            <button
              onClick={() => handleModeChange('system')}
              className={cn(
                'flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium transition-all',
                settings.mode === 'system'
                  ? (themeStyle === 'apple-glass' ? 'bg-slate-500 text-white shadow-sm' : 'bg-slate-500 text-white')
                  : (themeStyle === 'apple-glass' ? 'text-slate-600 hover:bg-black/5' : 'text-slate-400 hover:bg-white/10')
              )}
              title={t('singboxSettings.systemProxy')}
            >
              <ArrowRightLeft className="w-3 h-3" />
              <span className="hidden sm:inline">{t('singboxSettings.systemProxy')}</span>
            </button>
          </div>
          
          <button
            onClick={handleReset}
            className={cn(
              'flex items-center gap-1.5 px-3 rounded-lg border h-8 text-xs font-medium transition-all',
              themeStyle === 'apple-glass'
                ? 'bg-white/40 border-black/10 text-slate-600 hover:bg-white/60'
                : 'bg-white/5 border-white/10 text-slate-400 hover:bg-white/10'
            )}
          >
            <RotateCcw className="w-3 h-3" />
            <span className="hidden sm:inline">{t('singboxSettings.reset')}</span>
          </button>
        </div>
      </div>

      
      {/* TUN 设置 - 只在 TUN 模式下显示 */}
      {settings.mode === 'tun' && (
        <SettingsSection title={t('singboxSettings.tunSettings')} icon={Layers} themeStyle={themeStyle}>
          <FormField
            label={t('singboxSettings.tunStack')}
            description={t('singboxSettings.tunStackDesc')}
            themeStyle={themeStyle}
          >
            <Select
              value={settings.tunStack}
              onChange={(v) => updateSettings('tunStack', v)}
              options={[
                { value: 'system', label: 'System' },
                { value: 'gvisor', label: 'gVisor' },
                { value: 'mixed', label: 'Mixed' },
              ]}
              themeStyle={themeStyle}
            />
          </FormField>
          <FormField
            label={t('singboxSettings.tunMtu')}
            description={t('singboxSettings.tunMtuDesc')}
            themeStyle={themeStyle}
          >
            <NumberInput
              value={settings.tunMtu}
              onChange={(v) => updateSettings('tunMtu', v)}
              min={1280}
              max={65535}
              themeStyle={themeStyle}
            />
          </FormField>
          <FormField
            label={t('singboxSettings.strictRoute')}
            description={t('singboxSettings.strictRouteDesc')}
            themeStyle={themeStyle}
          >
            <Toggle
              value={settings.strictRoute}
              onChange={(v) => updateSettings('strictRoute', v)}
              themeStyle={themeStyle}
            />
          </FormField>
          <FormField
            label={t('singboxSettings.autoRedirect')}
            description={t('singboxSettings.autoRedirectDesc')}
            themeStyle={themeStyle}
          >
            <Toggle
              value={settings.autoRedirect}
              onChange={(v) => updateSettings('autoRedirect', v)}
              themeStyle={themeStyle}
            />
          </FormField>
        </SettingsSection>
      )}

      {/* 性能优化 */}
      <SettingsSection title={t('singboxSettings.performance')} icon={Zap} themeStyle={themeStyle}>
        <FormField
          label={t('singboxSettings.sniff')}
          description={t('singboxSettings.sniffDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.sniff}
            onChange={(v) => updateSettings('sniff', v)}
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.overrideDest')}
          description={t('singboxSettings.overrideDestDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.sniffOverrideDestination}
            onChange={(v) => updateSettings('sniffOverrideDestination', v)}
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.tcpFastOpen')}
          description={t('singboxSettings.tcpFastOpenDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.tcpFastOpen}
            onChange={(v) => updateSettings('tcpFastOpen', v)}
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.tcpMultiPath')}
          description={t('singboxSettings.tcpMultiPathDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.tcpMultiPath}
            onChange={(v) => updateSettings('tcpMultiPath', v)}
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.udpFragment')}
          description={t('singboxSettings.udpFragmentDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.udpFragment}
            onChange={(v) => updateSettings('udpFragment', v)}
            themeStyle={themeStyle}
          />
        </FormField>
      </SettingsSection>

      {/* 端口设置 */}
      <SettingsSection title={t('singboxSettings.portSettings')} icon={Network} themeStyle={themeStyle}>
        <FormField
          label={t('singboxSettings.mixedPort')}
          description={t('singboxSettings.mixedPortDesc')}
          themeStyle={themeStyle}
        >
          <NumberInput
            value={settings.mixedPort}
            onChange={(v) => updateSettings('mixedPort', v)}
            min={1}
            max={65535}
            themeStyle={themeStyle}
          />
        </FormField>
        {settings.mode === 'system' && (
          <>
            <FormField
              label={t('singboxSettings.httpPort')}
              themeStyle={themeStyle}
            >
              <NumberInput
                value={settings.httpPort}
                onChange={(v) => updateSettings('httpPort', v)}
                min={1}
                max={65535}
                themeStyle={themeStyle}
              />
            </FormField>
            <FormField
              label={t('singboxSettings.socksPort')}
              themeStyle={themeStyle}
            >
              <NumberInput
                value={settings.socksPort}
                onChange={(v) => updateSettings('socksPort', v)}
                min={1}
                max={65535}
                themeStyle={themeStyle}
              />
            </FormField>
          </>
        )}
      </SettingsSection>

      {/* DNS 设置 */}
      <SettingsSection title={t('singboxSettings.dnsSettings')} icon={Globe} themeStyle={themeStyle}>
        <FormField
          label={t('singboxSettings.fakeip')}
          description={t('singboxSettings.fakeipDesc')}
          themeStyle={themeStyle}
        >
          <Toggle
            value={settings.fakeip}
            onChange={(v) => updateSettings('fakeip', v)}
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.dnsStrategy')}
          description={t('singboxSettings.dnsStrategyDesc')}
          themeStyle={themeStyle}
        >
          <Select
            value={settings.dnsStrategy}
            onChange={(v) => updateSettings('dnsStrategy', v)}
            options={[
              { value: 'prefer_ipv4', label: t('singboxSettings.preferIpv4') },
              { value: 'prefer_ipv6', label: t('singboxSettings.preferIpv6') },
              { value: 'ipv4_only', label: t('singboxSettings.ipv4Only') },
              { value: 'ipv6_only', label: t('singboxSettings.ipv6Only') },
            ]}
            themeStyle={themeStyle}
          />
        </FormField>
      </SettingsSection>

      {/* Clash API */}
      <SettingsSection title={t('singboxSettings.clashApi')} icon={Server} defaultOpen={false} themeStyle={themeStyle}>
        <FormField
          label={t('singboxSettings.apiAddr')}
          description={t('singboxSettings.apiAddrDesc')}
          themeStyle={themeStyle}
        >
          <TextInput
            value={settings.clashApiAddr}
            onChange={(v) => updateSettings('clashApiAddr', v)}
            placeholder="127.0.0.1:9090"
            themeStyle={themeStyle}
          />
        </FormField>
        <FormField
          label={t('singboxSettings.apiSecret')}
          description={t('singboxSettings.apiSecretDesc')}
          themeStyle={themeStyle}
        >
          <TextInput
            value={settings.clashApiSecret}
            onChange={(v) => updateSettings('clashApiSecret', v)}
            placeholder={t('singboxSettings.apiSecretPlaceholder')}
            themeStyle={themeStyle}
          />
        </FormField>
      </SettingsSection>

      {/* 日志设置 */}
      <SettingsSection title={t('singboxSettings.logSettings')} icon={Monitor} defaultOpen={false} themeStyle={themeStyle}>
        <FormField
          label={t('singboxSettings.logLevel')}
          themeStyle={themeStyle}
        >
          <Select
            value={settings.logLevel}
            onChange={(v) => updateSettings('logLevel', v)}
            options={[
              { value: 'trace', label: 'Trace' },
              { value: 'debug', label: 'Debug' },
              { value: 'info', label: 'Info' },
              { value: 'warn', label: 'Warn' },
              { value: 'error', label: 'Error' },
              { value: 'fatal', label: 'Fatal' },
              { value: 'panic', label: 'Panic' },
            ]}
            themeStyle={themeStyle}
          />
        </FormField>
      </SettingsSection>
    </div>
  )
}
