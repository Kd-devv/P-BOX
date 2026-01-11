import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSidebar } from './Layout'
import { useThemeStore } from '@/stores/themeStore'
import { Menu, ArrowDown, ArrowUp, Power, RefreshCw, Loader2, Globe, User, LogOut, X, Key, Eye, EyeOff, Zap, Navigation, Unplug } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useState, useEffect, useRef } from 'react'
import { proxyApi, ProxyStatus } from '@/api'
import { mihomoApi } from '@/api/mihomo'
import { authApi, clearAuth } from '@/api/auth'
import { formatBytes, formatDuration } from '@/lib/utils'

// Map paths to nav keys
const pathToNavKey: Record<string, string> = {
  '/': 'nav.dashboard',
  '/proxy-switch': 'nav.proxySwitch',
  '/nodes': 'nav.nodes',
  '/subscriptions': 'nav.subscriptions',
  '/connections': 'nav.connections',
  '/logs': 'nav.logs',
  '/ruleset': 'nav.ruleset',
  '/config-generator': 'nav.configGenerator',
  '/core-manage': 'nav.coreManage',
  '/tools': 'nav.tools',
  '/settings': 'nav.settings',
}

export default function Header() {
  const location = useLocation()
  const navigate = useNavigate()
  const { t, i18n } = useTranslation()
  const { toggle } = useSidebar()
  const { themeStyle } = useThemeStore()
  
  const [status, setStatus] = useState<ProxyStatus | null>(null)
  const [traffic, setTraffic] = useState({ up: 0, down: 0 })
  const [loading, setLoading] = useState<'start' | 'stop' | 'restart' | null>(null)
  const [displayUptime, setDisplayUptime] = useState(0)
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [authEnabled, setAuthEnabled] = useState(false)
  const [username, setUsername] = useState('admin')
  const [showPasswordDialog, setShowPasswordDialog] = useState(false)
  const [showUsernameDialog, setShowUsernameDialog] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)

  const fetchStatus = async () => {
    try {
      const data = await proxyApi.getStatus()
      setStatus(data)
    } catch {
      // Ignore errors
    }
  }

  const handleStart = async () => {
    try {
      setLoading('start')
      await proxyApi.start()
      await fetchStatus()
    } catch {
      // Ignore errors
    } finally {
      setLoading(null)
    }
  }

  const handleStop = async () => {
    try {
      setLoading('stop')
      await proxyApi.stop()
      await fetchStatus()
    } catch {
      // Ignore errors
    } finally {
      setLoading(null)
    }
  }

  const handleRestart = async () => {
    try {
      setLoading('restart')
      await proxyApi.restart()
      await fetchStatus()
    } catch {
      // Ignore errors
    } finally {
      setLoading(null)
    }
  }

  const handleModeChange = async (mode: 'rule' | 'global' | 'direct') => {
    try {
      await proxyApi.setMode(mode)
      await fetchStatus()
    } catch {
      // Ignore errors
    }
  }

  useEffect(() => {
    fetchStatus()
    checkAuthStatus()
    const interval = setInterval(fetchStatus, 5000)

    // Setup traffic WebSocket
    const trafficWs = mihomoApi.createTrafficWs((data) => {
      setTraffic({ up: data.up, down: data.down })
    })

    // Close user menu on outside click
    const handleClickOutside = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setShowUserMenu(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      clearInterval(interval)
      trafficWs.close()
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  // 每秒更新运行时间显示
  useEffect(() => {
    if (status?.uptime !== undefined && status.uptime > 0 && status.running) {
      setDisplayUptime(status.uptime)
      const timer = setInterval(() => {
        setDisplayUptime(prev => prev + 1)
      }, 1000)
      return () => clearInterval(timer)
    } else {
      setDisplayUptime(0)
    }
  }, [status?.uptime, status?.running])

  const checkAuthStatus = async () => {
    try {
      const status = await authApi.check()
      setAuthEnabled(status.enabled)
      if (status.enabled) {
        const config = await authApi.getConfig()
        setUsername(config.username || 'admin')
      }
    } catch {
      // Ignore
    }
  }

  const handleLogout = async () => {
    try {
      await authApi.logout()
    } catch {
      // Ignore
    }
    clearAuth()
    navigate('/login')
  }


  const pageTitle = t(pathToNavKey[location.pathname] || 'nav.dashboard')

  return (
    <header className="h-14 lg:h-16 flex items-center justify-between px-4 lg:px-6 relative z-[100]">
      {/* Left: Menu button (mobile) + Title + Status */}
      <div className="flex items-center gap-4">
        <button 
          onClick={toggle}
          className={cn(
            'p-2 rounded-lg lg:hidden transition-colors',
            themeStyle === 'apple-glass'
              ? 'hover:bg-black/5 text-slate-600'
              : 'hover:bg-white/10 text-neutral-400'
          )}
        >
          <Menu className="w-5 h-5" />
        </button>
        <h1 className={cn(
          'hidden sm:block text-lg lg:text-xl font-semibold',
          themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
        )}>{pageTitle}</h1>
        
        {/* Core status badge - desktop only */}
        <div className={cn(
          'hidden lg:flex items-center gap-2 text-xs font-mono',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          <div className="h-4 w-px bg-current opacity-30" />
          <span className={status?.running ? 'text-green-500' : 'text-red-500'}>●</span>
          <span>{status?.running ? t('common.running') : t('common.stopped')}</span>
          {displayUptime > 0 && (
            <>
              <span className="opacity-50">|</span>
              <span>{t('header.uptime')}: {formatDuration(displayUptime)}</span>
            </>
          )}
        </div>

        {/* Speed indicators - next to status */}
        <div className={cn(
          'hidden sm:flex items-center gap-3 text-xs font-mono',
          themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
        )}>
          <div className="h-4 w-px bg-current opacity-30" />
          <div className="flex items-center gap-1">
            <ArrowDown className="w-3 h-3 text-green-500" />
            <span>{formatBytes(traffic.down)}/s</span>
          </div>
          <div className="flex items-center gap-1">
            <ArrowUp className="w-3 h-3 text-blue-500" />
            <span>{formatBytes(traffic.up)}/s</span>
          </div>
        </div>
      </div>

      {/* Right: Controls + Speed indicators */}
      <div className="flex items-center gap-1 sm:gap-2 lg:gap-3">
        {/* Mobile: Simple status + control */}
        <div className={cn(
          'flex lg:hidden items-center gap-2 px-2 py-1 rounded-full border',
          themeStyle === 'apple-glass'
            ? 'border-black/10 bg-white/40'
            : 'border-white/10 bg-white/5'
        )}>
          <span className={cn(
            'w-2 h-2 rounded-full',
            status?.running ? 'bg-green-500' : 'bg-red-500'
          )} />
          {status?.running ? (
            <button
              onClick={handleStop}
              disabled={loading !== null}
              className={cn(
                'p-1.5 rounded-md transition-all',
                themeStyle === 'apple-glass'
                  ? 'text-red-500 hover:bg-red-100'
                  : 'text-red-400 hover:bg-red-500/20'
              )}
            >
              {loading === 'stop' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Power className="w-4 h-4" />}
            </button>
          ) : (
            <button
              onClick={handleStart}
              disabled={loading !== null}
              className={cn(
                'p-1.5 rounded-md transition-all',
                themeStyle === 'apple-glass'
                  ? 'text-green-500 hover:bg-green-100'
                  : 'text-green-400 hover:bg-green-500/20'
              )}
            >
              {loading === 'start' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Power className="w-4 h-4" />}
            </button>
          )}
        </div>

        {/* Desktop: Control buttons */}
        <div className={cn(
          'hidden lg:flex items-center gap-2 px-3 py-1 rounded-full border h-8',
          themeStyle === 'apple-glass'
            ? 'border-black/10 bg-white/40'
            : 'border-white/10 bg-white/5'
        )}>
          {/* Status indicator */}
          <div className="flex items-center gap-2 pr-2 border-r border-current/10">
            <span className={cn(
              'w-2 h-2 rounded-full',
              status?.running ? 'bg-green-500' : 'bg-red-500'
            )} />
            <span className={cn(
              'text-xs uppercase font-medium',
              status?.running 
                ? 'text-green-500' 
                : 'text-red-500'
            )}>
              {status?.running ? 'ONLINE' : 'OFFLINE'}
            </span>
            {status?.coreVersion && (
              <span className={cn(
                'text-xs font-mono',
                themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
              )}>v{status.coreVersion}</span>
            )}
          </div>

          {/* Start/Stop button */}
          {status?.running ? (
            <button
              onClick={handleStop}
              disabled={loading !== null}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-medium transition-all',
                themeStyle === 'apple-glass'
                  ? 'bg-red-100 text-red-600 hover:bg-red-200'
                  : 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
              )}
            >
              {loading === 'stop' ? (
                <Loader2 className="w-3 h-3 animate-spin" />
              ) : (
                <Power className="w-3 h-3" />
              )}
              Stop
            </button>
          ) : (
            <button
              onClick={handleStart}
              disabled={loading !== null}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-medium transition-all',
                themeStyle === 'apple-glass'
                  ? 'bg-green-100 text-green-600 hover:bg-green-200'
                  : 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
              )}
            >
              {loading === 'start' ? (
                <Loader2 className="w-3 h-3 animate-spin" />
              ) : (
                <Power className="w-3 h-3" />
              )}
              Start
            </button>
          )}

          {/* Restart button */}
          <button
            onClick={handleRestart}
            disabled={loading !== null || !status?.running}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-medium transition-all',
              !status?.running && 'opacity-50 cursor-not-allowed',
              themeStyle === 'apple-glass'
                ? 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                : 'bg-white/10 text-slate-300 hover:bg-white/20'
            )}
          >
            {loading === 'restart' ? (
              <Loader2 className="w-3 h-3 animate-spin" />
            ) : (
              <RefreshCw className="w-3 h-3" />
            )}
            Restart
          </button>
        </div>

        {/* Mode Switcher - 未运行时可选择，运行时禁用 */}
        <div className={cn(
          'flex items-center gap-0.5 sm:gap-1 px-1 sm:px-2 py-1 rounded-full border h-8',
          themeStyle === 'apple-glass'
            ? 'border-black/10 bg-white/40'
            : 'border-white/10 bg-white/5',
          status?.running && 'opacity-50'
        )}>
          <button
            onClick={() => handleModeChange('rule')}
            disabled={status?.running}
            className={cn(
              'flex items-center gap-0.5 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-md text-[10px] sm:text-xs font-medium transition-all',
              status?.running && 'cursor-not-allowed',
              (status?.mode === 'rule' || (!status && true))
                ? (themeStyle === 'apple-glass' ? 'bg-blue-500 text-white' : 'bg-cyan-500 text-white')
                : (themeStyle === 'apple-glass' ? 'text-slate-600 hover:bg-slate-100' : 'text-slate-400 hover:bg-white/10')
            )}
            title={status?.running ? t('header.modeDisabled') : t('proxySettings.modeRule')}
          >
            <Navigation className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
            <span className="hidden sm:inline">{t('header.rule')}</span>
          </button>
          <button
            onClick={() => handleModeChange('global')}
            disabled={status?.running}
            className={cn(
              'flex items-center gap-0.5 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-md text-[10px] sm:text-xs font-medium transition-all',
              status?.running && 'cursor-not-allowed',
              status?.mode === 'global'
                ? (themeStyle === 'apple-glass' ? 'bg-orange-500 text-white' : 'bg-orange-500 text-white')
                : (themeStyle === 'apple-glass' ? 'text-slate-600 hover:bg-slate-100' : 'text-slate-400 hover:bg-white/10')
            )}
            title={status?.running ? t('header.modeDisabled') : t('proxySettings.modeGlobal')}
          >
            <Zap className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
            <span className="hidden sm:inline">{t('header.global')}</span>
          </button>
          <button
            onClick={() => handleModeChange('direct')}
            disabled={status?.running}
            className={cn(
              'flex items-center gap-0.5 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-md text-[10px] sm:text-xs font-medium transition-all',
              status?.running && 'cursor-not-allowed',
              status?.mode === 'direct'
                ? (themeStyle === 'apple-glass' ? 'bg-slate-500 text-white' : 'bg-slate-500 text-white')
                : (themeStyle === 'apple-glass' ? 'text-slate-600 hover:bg-slate-100' : 'text-slate-400 hover:bg-white/10')
            )}
            title={status?.running ? t('header.modeDisabled') : t('proxySettings.modeDirect')}
          >
            <Unplug className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
            <span className="hidden sm:inline">{t('header.direct')}</span>
          </button>
        </div>

        {/* Language Switcher - responsive */}
        <button
          onClick={() => {
            const newLang = i18n.language.startsWith('zh') ? 'en' : 'zh'
            i18n.changeLanguage(newLang)
          }}
          className={cn(
            'flex items-center gap-1 sm:gap-1.5 px-2 sm:px-3 py-1 rounded-full border transition-all h-8',
            themeStyle === 'apple-glass'
              ? 'border-black/10 bg-white/50 hover:bg-white/70 text-slate-700'
              : 'border-white/10 bg-white/5 hover:bg-white/10 text-slate-300'
          )}
          title={i18n.language.startsWith('zh') ? 'Switch to English' : '切换到中文'}
        >
          <Globe className="w-4 h-4" />
          <span className="text-xs font-medium hidden sm:inline">
            {i18n.language.startsWith('zh') ? '中文' : 'EN'}
          </span>
        </button>

        {/* User Menu */}
        {authEnabled && (
          <div className="relative" ref={userMenuRef}>
            <button
              onClick={() => setShowUserMenu(!showUserMenu)}
              className={cn(
                'flex items-center gap-2 px-3 py-1.5 rounded-full border transition-all',
                themeStyle === 'apple-glass'
                  ? 'border-black/10 bg-white/50 hover:bg-white/70 text-slate-700'
                  : 'border-white/10 bg-white/5 hover:bg-white/10 text-slate-300'
              )}
            >
              {/* 用户头像 */}
              <div className={cn(
                'w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold text-white',
                'bg-gradient-to-br from-blue-500 to-purple-600'
              )}>
                {username.charAt(0).toUpperCase()}
              </div>
              <span className="text-xs sm:text-sm font-medium hidden sm:inline">{username}</span>
            </button>

            {showUserMenu && (
              <div className={cn(
                'absolute right-0 top-full mt-2 w-48 rounded-xl overflow-hidden shadow-xl z-[999] border',
                themeStyle === 'apple-glass'
                  ? 'bg-white/95 backdrop-blur-xl border-black/10'
                  : 'bg-slate-900/95 backdrop-blur-xl border-white/10'
              )}>
                {/* 用户信息头部 */}
                <div className={cn(
                  'px-4 py-3 border-b',
                  themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
                )}>
                  <div className="flex items-center gap-3">
                    <div className={cn(
                      'w-10 h-10 rounded-full flex items-center justify-center text-lg font-bold text-white',
                      'bg-gradient-to-br from-blue-500 to-purple-600'
                    )}>
                      {username.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className={cn(
                        'font-medium',
                        themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
                      )}>{username}</div>
                      <div className={cn(
                        'text-xs',
                        themeStyle === 'apple-glass' ? 'text-slate-400' : 'text-slate-500'
                      )}>{t('auth.administrator')}</div>
                    </div>
                  </div>
                </div>
                
                {/* 菜单选项 */}
                <div className="py-1">
                  <button
                    onClick={() => {
                      setShowUserMenu(false)
                      setShowUsernameDialog(true)
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors',
                      themeStyle === 'apple-glass'
                        ? 'text-slate-700 hover:bg-black/5'
                        : 'text-slate-300 hover:bg-white/5'
                    )}
                  >
                    <User className="w-4 h-4" />
                    <span>{t('auth.changeUsername')}</span>
                  </button>
                  <button
                    onClick={() => {
                      setShowUserMenu(false)
                      setShowPasswordDialog(true)
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors',
                      themeStyle === 'apple-glass'
                        ? 'text-slate-700 hover:bg-black/5'
                        : 'text-slate-300 hover:bg-white/5'
                    )}
                  >
                    <Key className="w-4 h-4" />
                    <span>{t('auth.changePassword')}</span>
                  </button>
                  <button
                    onClick={() => {
                      setShowUserMenu(false)
                      navigate('/settings')
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors',
                      themeStyle === 'apple-glass'
                        ? 'text-slate-700 hover:bg-black/5'
                        : 'text-slate-300 hover:bg-white/5'
                    )}
                  >
                    <Globe className="w-4 h-4" />
                    <span>{t('settings.title')}</span>
                  </button>
                </div>
                
                {/* 退出 */}
                <div className={cn(
                  'border-t py-1',
                  themeStyle === 'apple-glass' ? 'border-black/5' : 'border-white/5'
                )}>
                  <button
                    onClick={() => {
                      setShowUserMenu(false)
                      handleLogout()
                    }}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors text-red-500 hover:bg-red-500/10"
                  >
                    <LogOut className="w-4 h-4" />
                    <span>{t('auth.logout')}</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 修改密码弹窗 */}
      {showPasswordDialog && (
        <PasswordDialog 
          onClose={() => setShowPasswordDialog(false)} 
          themeStyle={themeStyle}
        />
      )}

      {/* 修改用户名弹窗 */}
      {showUsernameDialog && (
        <UsernameDialog 
          onClose={() => setShowUsernameDialog(false)} 
          themeStyle={themeStyle}
          currentUsername={username}
          onSuccess={(newUsername) => setUsername(newUsername)}
        />
      )}
    </header>
  )
}

// 修改密码弹窗
function PasswordDialog({ onClose, themeStyle }: { onClose: () => void, themeStyle: string }) {
  const { t } = useTranslation()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showOld, setShowOld] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    if (newPassword !== confirmPassword) {
      setError(t('auth.passwordMismatch'))
      return
    }
    if (newPassword.length < 4) {
      setError(t('auth.passwordTooShort'))
      return
    }
    
    setLoading(true)
    setError('')
    try {
      await authApi.changePassword(oldPassword, newPassword)
      onClose()
    } catch (e: unknown) {
      setError((e as Error).message || t('auth.changeFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-[9999] p-4">
      <div className={cn(
        'w-full max-w-sm rounded-2xl p-6',
        themeStyle === 'apple-glass'
          ? 'bg-white/95 backdrop-blur-xl border border-white/50'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        <div className="flex items-center justify-between mb-6">
          <h3 className={cn(
            'text-lg font-semibold flex items-center gap-2',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            <Key className="w-5 h-5" />
            {t('auth.changePassword')}
          </h3>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-black/5">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-4">
          {/* 旧密码 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('auth.oldPassword')}</label>
            <div className="relative">
              <input
                type={showOld ? 'text' : 'password'}
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                className="form-input pr-10"
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowOld(!showOld)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400"
              >
                {showOld ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          {/* 新密码 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('auth.newPassword')}</label>
            <div className="relative">
              <input
                type={showNew ? 'text' : 'password'}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="form-input pr-10"
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowNew(!showNew)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400"
              >
                {showNew ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          {/* 确认密码 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('auth.confirmPassword')}</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="form-input"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <div className="text-sm text-red-500 text-center py-2">{error}</div>
          )}

          <button
            onClick={handleSubmit}
            disabled={loading || !oldPassword || !newPassword || !confirmPassword}
            className={cn(
              'w-full py-3 rounded-xl font-medium transition-all',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 text-white hover:bg-blue-600'
                : 'bg-cyan-500 text-white hover:bg-cyan-600'
            )}
          >
            {loading ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> : t('common.save')}
          </button>
        </div>
      </div>
    </div>
  )
}

// 修改用户名弹窗
function UsernameDialog({ 
  onClose, 
  themeStyle, 
  currentUsername, 
  onSuccess 
}: { 
  onClose: () => void
  themeStyle: string
  currentUsername: string
  onSuccess: (newUsername: string) => void 
}) {
  const { t } = useTranslation()
  const [newUsername, setNewUsername] = useState(currentUsername)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    if (newUsername.length < 2) {
      setError(t('auth.usernameTooShort'))
      return
    }
    
    setLoading(true)
    setError('')
    try {
      await authApi.updateUsername(newUsername)
      onSuccess(newUsername)
      onClose()
    } catch (e: unknown) {
      setError((e as Error).message || t('auth.changeFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-[9999] p-4">
      <div className={cn(
        'w-full max-w-sm rounded-2xl p-6',
        themeStyle === 'apple-glass'
          ? 'bg-white/95 backdrop-blur-xl border border-white/50'
          : 'bg-slate-900/95 backdrop-blur-xl border border-white/10'
      )}>
        <div className="flex items-center justify-between mb-6">
          <h3 className={cn(
            'text-lg font-semibold flex items-center gap-2',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            <User className="w-5 h-5" />
            {t('auth.changeUsername')}
          </h3>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-black/5">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>{t('auth.newUsername')}</label>
            <input
              type="text"
              value={newUsername}
              onChange={(e) => setNewUsername(e.target.value)}
              className="form-input"
              placeholder={t('auth.enterUsername')}
            />
          </div>

          {error && (
            <div className="text-sm text-red-500 text-center py-2">{error}</div>
          )}

          <button
            onClick={handleSubmit}
            disabled={loading || !newUsername || newUsername === currentUsername}
            className={cn(
              'w-full py-3 rounded-xl font-medium transition-all',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              themeStyle === 'apple-glass'
                ? 'bg-blue-500 text-white hover:bg-blue-600'
                : 'bg-cyan-500 text-white hover:bg-cyan-600'
            )}
          >
            {loading ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> : t('common.save')}
          </button>
        </div>
      </div>
    </div>
  )
}
