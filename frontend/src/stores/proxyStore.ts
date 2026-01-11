import { create } from 'zustand'
import { proxyApi, ProxyStatus } from '@/api'

interface ProxyStore {
  status: ProxyStatus | null
  isRunning: boolean
  fetchStatus: () => Promise<void>
}

export const useProxyStore = create<ProxyStore>((set) => ({
  status: null,
  isRunning: false,
  
  fetchStatus: async () => {
    try {
      const data = await proxyApi.getStatus()
      set({ status: data, isRunning: data.running })
    } catch {
      set({ status: null, isRunning: false })
    }
  },
}))
