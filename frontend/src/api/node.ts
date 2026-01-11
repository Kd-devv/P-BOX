import api from './client'

export interface Node {
  id: string
  name: string
  type: string
  server: string
  serverPort: number
  subscriptionId?: string
  isManual: boolean
  enabled: boolean
  delay: number      // Latency in ms, 0=timeout, -1=not tested
  lastTest: number   // Last test timestamp
  config: string
  shareUrl?: string
}

// 协议字段定义
export interface ProtocolField {
  name: string
  label: string
  type: 'text' | 'number' | 'password' | 'select' | 'boolean' | 'textarea'
  required: boolean
  default?: unknown
  placeholder?: string
  description?: string
  options?: Array<{ label: string; value: unknown }>
  min?: number
  max?: number
  depends_on?: string
  depends_value?: unknown
}

// 手动添加节点请求
export interface ManualNodeRequest {
  name: string
  type: string
  server: string
  server_port: number
  config: Record<string, unknown>
}

// 支持的协议列表
export const PROTOCOL_LIST = [
  { value: 'vmess', label: 'VMess' },
  { value: 'vless', label: 'VLESS' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'shadowsocks', label: 'Shadowsocks' },
  { value: 'socks', label: 'SOCKS5' },
  { value: 'hysteria', label: 'Hysteria' },
  { value: 'hysteria2', label: 'Hysteria2' },
  { value: 'tuic', label: 'TUIC' },
  { value: 'wireguard', label: 'WireGuard' },
  { value: 'ssh', label: 'SSH' },
]

export const nodeApi = {
  // Get all nodes
  list: () => api.get<Node[]>('/nodes'),

  // Import node from URL
  importUrl: (url: string) => 
    api.post<Node>('/nodes/import', { url }),

  // Add manual node (simple)
  addManual: (data: { name: string; type: string; server: string; port: number; config?: string }) =>
    api.post<Node>('/nodes/manual', data),

  // Add manual node (advanced with full config)
  addManualAdvanced: (data: ManualNodeRequest) =>
    api.post<Node>('/nodes/manual/advanced', data),

  // Delete manual node
  delete: (id: string) => 
    api.delete(`/nodes/${id}`),

  // Test single node delay
  testDelay: (nodeId: string, server: string, port: number, timeout?: number) =>
    api.post<{ delay: number }>('/nodes/test', { nodeId, server, port, timeout }),

  // Batch test delays
  testDelayBatch: (nodeIds: string[], timeout?: number) =>
    api.post<Record<string, number>>('/nodes/test-batch', { nodeIds, timeout }),

  // Get share URL
  getShareUrl: (id: string) =>
    api.get<{ url: string }>(`/nodes/${id}/share`),

  // Get protocol field definitions
  getProtocolFields: (protocol: string) =>
    api.get<{ protocol: string; fields: ProtocolField[] }>(`/nodes/protocols/${protocol}/fields`),
}
