import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number, decimals = 2): string {
  // 处理无效值
  if (bytes === undefined || bytes === null || isNaN(bytes) || bytes <= 0) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  // 防止 i 超出 sizes 数组范围
  const index = Math.min(i, sizes.length - 1)
  return parseFloat((bytes / Math.pow(k, index)).toFixed(dm)) + ' ' + sizes[index]
}

export function formatDuration(seconds: number): string {
  const secs = seconds % 60
  if (seconds < 60) return `${secs}s`
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60)
    return `${mins}m ${secs}s`
  }
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  return `${hours}h ${mins}m ${secs}s`
}

export function getLatencyColor(latency: number): string {
  if (latency === 0 || latency > 5000) return 'latency-timeout'
  if (latency < 200) return 'latency-good'
  if (latency < 500) return 'latency-medium'
  return 'latency-bad'
}

export function formatLatency(latency: number): string {
  if (latency === 0 || latency > 5000) return 'Timeout'
  return `${latency}ms`
}
