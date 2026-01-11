import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type CoreType = 'mihomo' | 'singbox'

interface CoreState {
  // 当前激活的核心类型
  activeCore: CoreType
  // 设置当前核心
  setActiveCore: (core: CoreType) => void
}

export const useCoreStore = create<CoreState>()(
  persist(
    (set) => ({
      activeCore: 'mihomo',
      setActiveCore: (core) => set({ activeCore: core }),
    }),
    {
      name: 'core-storage',
    }
  )
)
