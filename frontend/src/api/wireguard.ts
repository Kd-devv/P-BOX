import apiClient from './client'

export interface WireGuardClient {
  id: string
  name: string
  private_key: string
  public_key: string
  preshared_key: string
  allowed_ips: string
  dns: string
  enabled: boolean
  description: string
  created_at: string
}

export interface WireGuardServer {
  id: string
  name: string
  tag: string
  enabled: boolean
  auto_start: boolean
  endpoint: string
  listen_port: number
  private_key: string
  public_key: string
  address: string
  mtu: number
  dns: string
  description: string
  created_at: string
  updated_at: string
  clients: WireGuardClient[]
}

export interface WireGuardStatus {
  linux: boolean
  installed: boolean
}

// 获取系统状态
export const getSystemStatus = () =>
  apiClient.get<WireGuardStatus>('/wireguard/status')

// 安装 WireGuard
export const install = () =>
  apiClient.post<void>('/wireguard/install')

// 获取默认 DNS（本机内网 IP）
export const getDefaultDNS = () =>
  apiClient.get<{ dns: string[]; local_ip: string }>('/wireguard/default-dns')

// 获取所有服务器
export const getServers = () =>
  apiClient.get<WireGuardServer[]>('/wireguard/servers')

// 获取单个服务器
export const getServer = (id: string) =>
  apiClient.get<WireGuardServer>(`/wireguard/servers/${id}`)

// 创建服务器
export const createServer = (data: Partial<WireGuardServer>) =>
  apiClient.post<WireGuardServer>('/wireguard/servers', data)

// 更新服务器
export const updateServer = (id: string, data: { name?: string; endpoint?: string; auto_start?: boolean; description?: string }) =>
  apiClient.put<WireGuardServer>(`/wireguard/servers/${id}`, data)

// 删除服务器
export const deleteServer = (id: string) =>
  apiClient.delete<void>(`/wireguard/servers/${id}`)

// 启动服务器
export const applyConfig = (id: string) =>
  apiClient.post<void>(`/wireguard/servers/${id}/apply`)

// 停止服务器
export const stopServer = (id: string) =>
  apiClient.post<void>(`/wireguard/servers/${id}/stop`)

// 获取服务器状态
export const getServerStatus = (id: string) =>
  apiClient.get<{ running: boolean; output: string }>(`/wireguard/servers/${id}/status`)

// 添加客户端
export const addClient = (serverId: string, data: Partial<WireGuardClient>) =>
  apiClient.post<WireGuardClient>(`/wireguard/servers/${serverId}/clients`, data)

// 更新客户端
export const updateClient = (serverId: string, clientId: string, data: { name?: string; description?: string; enabled?: boolean }) =>
  apiClient.put<WireGuardClient>(`/wireguard/servers/${serverId}/clients/${clientId}`, data)

// 删除客户端
export const deleteClient = (serverId: string, clientId: string) =>
  apiClient.delete<void>(`/wireguard/servers/${serverId}/clients/${clientId}`)

// 获取客户端配置
export const getClientConfig = (serverId: string, clientId: string, endpoint?: string) =>
  apiClient.get<string>(`/wireguard/servers/${serverId}/clients/${clientId}/config`, {
    params: { endpoint }
  })

export default {
  getSystemStatus,
  install,
  getDefaultDNS,
  getServers,
  getServer,
  createServer,
  updateServer,
  deleteServer,
  applyConfig,
  stopServer,
  getServerStatus,
  addClient,
  updateClient,
  deleteClient,
  getClientConfig
}
