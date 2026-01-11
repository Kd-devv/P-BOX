import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// Theme types: apple-glass (毛玻璃风格) and apple-pro-dark (专业深色)
export type ThemeStyle = 'apple-glass' | 'apple-pro-dark'
export type ColorScheme = 'light' | 'dark' | 'system'

interface ThemeState {
  themeStyle: ThemeStyle
  colorScheme: ColorScheme
  setThemeStyle: (style: ThemeStyle) => void
  setColorScheme: (scheme: ColorScheme) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      themeStyle: 'apple-glass',
      colorScheme: 'dark',
      setThemeStyle: (themeStyle) => {
        set({ themeStyle })
        applyTheme(themeStyle, useThemeStore.getState().colorScheme)
      },
      setColorScheme: (colorScheme) => {
        set({ colorScheme })
        applyTheme(useThemeStore.getState().themeStyle, colorScheme)
      },
    }),
    { name: 'p-box-theme-v2' }
  )
)

function applyTheme(style: ThemeStyle, scheme: ColorScheme) {
  const root = document.documentElement
  
  // Determine if dark mode
  const isDark =
    scheme === 'dark' ||
    (scheme === 'system' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)
  
  // Apply color scheme
  root.classList.toggle('dark', isDark)
  
  // Apply theme style
  root.classList.remove('theme-apple-glass', 'theme-apple-pro-dark')
  root.classList.add(`theme-${style}`)
  
  // Set data attribute for CSS
  root.setAttribute('data-theme-style', style)
}

export function initTheme() {
  const { themeStyle, colorScheme } = useThemeStore.getState()
  applyTheme(themeStyle, colorScheme)

  // Listen for system theme changes
  window
    .matchMedia('(prefers-color-scheme: dark)')
    .addEventListener('change', () => {
      if (useThemeStore.getState().colorScheme === 'system') {
        applyTheme(useThemeStore.getState().themeStyle, 'system')
      }
    })
}
