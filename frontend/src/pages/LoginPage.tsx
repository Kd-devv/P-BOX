import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Lock, Loader2, User } from 'lucide-react'
import { authApi } from '@/api/auth'
import { useThemeStore } from '@/stores/themeStore'
import { cn } from '@/lib/utils'

export default function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { themeStyle } = useThemeStore()
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const isGlass = themeStyle === 'apple-glass'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const res = await authApi.login({ username, password })
      localStorage.setItem('p-box-token', res.token)
      navigate('/')
    } catch (err) {
      setError((err as Error).message || t('login.error'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={cn(
      'min-h-screen flex items-center justify-center p-4',
      isGlass 
        ? "bg-[url('https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2564&auto=format&fit=crop')] bg-cover bg-center"
        : 'bg-neutral-950'
    )}>
      <div className={cn(
        'w-full max-w-sm p-8 rounded-3xl shadow-2xl',
        isGlass
          ? 'bg-white/20 backdrop-blur-2xl border border-white/30'
          : 'bg-neutral-900/95 backdrop-blur-xl border border-neutral-800'
      )}>
        {/* Logo */}
        <div className="text-center mb-8">
          <div className={cn(
            'w-20 h-20 rounded-[22px] mx-auto mb-4 flex items-center justify-center shadow-lg overflow-hidden',
            isGlass
              ? 'bg-gradient-to-br from-slate-100 to-white border border-black/5'
              : 'bg-black border border-purple-500/20'
          )}>
            <img src="/p-box-logo.png" alt="P-BOX" className="w-16 h-16 object-contain" />
          </div>
          <h1 className={cn(
            'text-2xl font-bold',
            isGlass ? 'text-slate-800' : 'text-white'
          )}>P-BOX</h1>
          <p className={cn(
            'text-sm mt-1',
            isGlass ? 'text-slate-600' : 'text-slate-400'
          )}>{t('login.title')}</p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* 用户名 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              isGlass ? 'text-slate-700' : 'text-slate-300'
            )}>{t('login.username')}</label>
            <div className="relative">
              <User className={cn(
                'absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5',
                isGlass ? 'text-slate-400' : 'text-slate-500'
              )} />
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder={t('login.usernamePlaceholder')}
                className={cn(
                  'w-full pl-10 pr-4 py-3 rounded-xl border transition-all',
                  isGlass
                    ? 'bg-white/50 border-white/40 text-slate-800 placeholder:text-slate-400 focus:bg-white/70 focus:border-blue-400'
                    : 'bg-neutral-800 border-neutral-700 text-white placeholder:text-slate-500 focus:border-cyan-500'
                )}
              />
            </div>
          </div>

          {/* 密码 */}
          <div>
            <label className={cn(
              'block text-sm font-medium mb-1.5',
              isGlass ? 'text-slate-700' : 'text-slate-300'
            )}>{t('login.password')}</label>
            <div className="relative">
              <Lock className={cn(
                'absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5',
                isGlass ? 'text-slate-400' : 'text-slate-500'
              )} />
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t('login.passwordPlaceholder')}
                className={cn(
                  'w-full pl-10 pr-4 py-3 rounded-xl border transition-all',
                  isGlass
                    ? 'bg-white/50 border-white/40 text-slate-800 placeholder:text-slate-400 focus:bg-white/70 focus:border-blue-400'
                    : 'bg-neutral-800 border-neutral-700 text-white placeholder:text-slate-500 focus:border-cyan-500'
                )}
                autoFocus
              />
            </div>
          </div>

          {error && (
            <div className={cn(
              'text-sm text-center py-2 px-3 rounded-lg',
              isGlass ? 'bg-red-100/80 text-red-600' : 'bg-red-500/20 text-red-400'
            )}>{error}</div>
          )}

          <button
            type="submit"
            disabled={loading || !password}
            className={cn(
              'w-full py-3.5 rounded-xl font-semibold transition-all duration-200 mt-6',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              isGlass
                ? 'bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white shadow-lg'
                : 'bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-600 hover:to-blue-700 text-white'
            )}
          >
            {loading ? (
              <Loader2 className="w-5 h-5 animate-spin mx-auto" />
            ) : (
              t('login.submit')
            )}
          </button>
        </form>

        {/* 底部版权 */}
        <p className={cn(
          'text-xs text-center mt-6',
          isGlass ? 'text-slate-500' : 'text-slate-600'
        )}>
          © 2024 P-BOX. All rights reserved.
        </p>
      </div>
    </div>
  )
}
