import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2, Play, Square, Users, Copy, Download, Server, AlertCircle, RefreshCw, QrCode, Check, X, Loader2, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import wireguardApi, { WireGuardServer, WireGuardClient, WireGuardStatus } from '@/api/wireguard'

export default function WireGuardPage() {
  const { t } = useTranslation()
  const { themeStyle } = useThemeStore()
  const [status, setStatus] = useState<WireGuardStatus | null>(null)
  const [servers, setServers] = useState<WireGuardServer[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedServer, setSelectedServer] = useState<WireGuardServer | null>(null)
  const [showModal, setShowModal] = useState<'create' | 'client' | 'config' | 'delete' | 'editServer' | 'editClient' | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'server' | 'client'; id: string; name: string } | null>(null)
  const [serverForm, setServerForm] = useState({ name: '', tag: 'wg0', listen: '0.0.0.0', listen_port: 51820, address: '10.0.1.1/24', mtu: 1420, dns: '', description: '', endpoint: '', auto_start: false })
  const [editServerForm, setEditServerForm] = useState({ name: '', endpoint: '', auto_start: false, description: '' })
  const [editClientForm, setEditClientForm] = useState({ name: '', description: '', enabled: true })
  const [editingServer, setEditingServer] = useState<WireGuardServer | null>(null)
  const [editingClient, setEditingClient] = useState<WireGuardClient | null>(null)
  const [localIP, setLocalIP] = useState('')
  const [defaultDNS, setDefaultDNS] = useState('')
  const [clientForm, setClientForm] = useState({ name: '', description: '', allowed_ips: '', dns: '', preshared_key: '' })
  const [clientConfig, setClientConfig] = useState('')
  const [selectedClient, setSelectedClient] = useState<WireGuardClient | null>(null)
  const [qrCodeURL, setQrCodeURL] = useState('')
  const [copied, setCopied] = useState(false)
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [installing, setInstalling] = useState(false)

  const showMessage = (type: 'success' | 'error', text: string) => { setMsg({ type, text }); setTimeout(() => setMsg(null), 3000) }

  const handleInstall = async () => {
    try {
      setInstalling(true)
      await wireguardApi.install()
      showMessage('success', t('wg.installSuccess'))
      loadData() // 刷新状态
    } catch (e: any) { showMessage('error', e.message || t('wg.installFailed')) }
    finally { setInstalling(false) }
  }

  const loadData = async () => {
    try {
      setLoading(true)
      const [statusRes, serversRes, dnsRes] = await Promise.all([
        wireguardApi.getSystemStatus(), 
        wireguardApi.getServers(),
        wireguardApi.getDefaultDNS()
      ])
      setStatus(statusRes)
      setServers(serversRes || [])
      // 同步更新 selectedServer（保持数据一致性）
      if (selectedServer && serversRes) {
        const updated = serversRes.find(s => s.id === selectedServer.id)
        if (updated) setSelectedServer(updated)
        else setSelectedServer(null) // 服务器被删除了
      }
      if (dnsRes?.local_ip) {
        setLocalIP(dnsRes.local_ip)
        const dns = dnsRes.dns?.join(',') || `${dnsRes.local_ip},1.1.1.1`
        setDefaultDNS(dns)
        setServerForm(prev => ({ ...prev, dns }))
      }
    } catch { /* ignore */ } finally { setLoading(false) }
  }

  useEffect(() => { loadData() }, [])

  const handleCreateServer = async () => {
    try {
      await wireguardApi.createServer(serverForm)
      showMessage('success', t('wg.createSuccess'))
      setShowModal(null)
      setServerForm({ name: '', tag: 'wg0', listen: '0.0.0.0', listen_port: 51820, address: '10.0.1.1/24', mtu: 1420, dns: defaultDNS || '1.1.1.1', description: '', endpoint: '', auto_start: false })
      loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.createFailed')) }
  }

  const handleToggleServer = async (server: WireGuardServer) => {
    try {
      if (server.enabled) { await wireguardApi.stopServer(server.id); showMessage('success', t('wg.stopped')) }
      else { await wireguardApi.applyConfig(server.id); showMessage('success', t('wg.started')) }
      loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.operationFailed')) }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      if (deleteTarget.type === 'server') {
        await wireguardApi.deleteServer(deleteTarget.id)
        if (selectedServer?.id === deleteTarget.id) setSelectedServer(null)
      } else if (selectedServer) {
        await wireguardApi.deleteClient(selectedServer.id, deleteTarget.id)
        const res = await wireguardApi.getServer(selectedServer.id)
        setSelectedServer(res)
      }
      showMessage('success', t('wg.deleteSuccess'))
      setShowModal(null)
      setDeleteTarget(null)
      loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.deleteFailed')) }
  }

  // 打开编辑服务器弹窗
  const openEditServerModal = (server: WireGuardServer) => {
    setEditingServer(server)
    setEditServerForm({ name: server.name, endpoint: server.endpoint || '', auto_start: server.auto_start || false, description: server.description || '' })
    setShowModal('editServer')
  }

  // 保存服务器编辑
  const handleUpdateServer = async () => {
    if (!editingServer) return
    try {
      await wireguardApi.updateServer(editingServer.id, editServerForm)
      showMessage('success', t('wg.updateSuccess'))
      setShowModal(null)
      setEditingServer(null)
      // 先获取更新后的服务器数据
      const updatedServer = await wireguardApi.getServer(editingServer.id)
      // 更新选中的服务器
      if (selectedServer?.id === editingServer.id) {
        setSelectedServer(updatedServer)
      }
      // 刷新列表
      await loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.updateFailed')) }
  }

  // 打开编辑客户端弹窗
  const openEditClientModal = (client: WireGuardClient) => {
    setEditingClient(client)
    setEditClientForm({ name: client.name, description: client.description || '', enabled: client.enabled })
    setShowModal('editClient')
  }

  // 保存客户端编辑
  const handleUpdateClient = async () => {
    if (!editingClient || !selectedServer) return
    try {
      await wireguardApi.updateClient(selectedServer.id, editingClient.id, editClientForm)
      showMessage('success', t('wg.updateSuccess'))
      setShowModal(null)
      setEditingClient(null)
      const res = await wireguardApi.getServer(selectedServer.id)
      setSelectedServer(res)
      loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.updateFailed')) }
  }

  // 获取下一个可用的客户端名称
  const getNextClientName = () => {
    if (!selectedServer?.clients?.length) return '客户端 1'
    const usedNumbers = new Set<number>()
    selectedServer.clients.forEach(c => {
      const match = c.name.match(/客户端\s*(\d+)/)
      if (match) usedNumbers.add(parseInt(match[1]))
    })
    for (let i = 1; i <= 999; i++) {
      if (!usedNumbers.has(i)) return `客户端 ${i}`
    }
    return `客户端 ${Date.now() % 1000}`
  }

  // 打开添加客户端弹窗
  const openAddClientModal = () => {
    setClientForm({ name: getNextClientName(), description: '', allowed_ips: '', dns: '', preshared_key: '' })
    setShowModal('client')
  }

  const handleAddClient = async () => {
    if (!selectedServer) return
    try {
      await wireguardApi.addClient(selectedServer.id, clientForm)
      showMessage('success', t('wg.addSuccess'))
      setShowModal(null)
      setClientForm({ name: '', description: '', allowed_ips: '', dns: '', preshared_key: '' })
      const res = await wireguardApi.getServer(selectedServer.id)
      setSelectedServer(res)
      loadData()
    } catch (e: any) { showMessage('error', e.response?.data?.message || t('wg.addFailed')) }
  }

  const handleGetClientConfig = async (client: WireGuardClient) => {
    if (!selectedServer) return
    try {
      const res = await wireguardApi.getClientConfig(selectedServer.id, client.id, selectedServer.endpoint || undefined)
      setClientConfig(res)
      setSelectedClient(client)
      // 动态生成二维码
      const QRCode = await import('qrcode')
      const qrDataURL = await QRCode.toDataURL(res, { width: 256, margin: 2, color: { dark: '#000000', light: '#ffffff' } })
      setQrCodeURL(qrDataURL)
      setShowModal('config')
    } catch (e: any) { showMessage('error', e.response?.data?.message || '获取配置失败') }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(clientConfig)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
      showMessage('success', '已复制到剪贴板')
    } catch { showMessage('error', '复制失败') }
  }

  const handleDownloadConfig = () => {
    const blob = new Blob([clientConfig], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedClient?.name || 'wireguard'}.conf`
    a.click()
    URL.revokeObjectURL(url)
    showMessage('success', t('wg.downloadConfig'))
  }

  const handleDownloadQR = () => {
    const a = document.createElement('a')
    a.href = qrCodeURL
    a.download = `${selectedClient?.name || 'wireguard'}-qr.png`
    a.click()
    showMessage('success', t('wg.downloadQr'))
  }

  const isGlass = themeStyle === 'apple-glass'
  const textMain = isGlass ? 'text-slate-800' : 'text-white'
  const textSub = isGlass ? 'text-slate-500' : 'text-slate-400'
  const accent = isGlass ? 'bg-blue-500' : 'bg-cyan-500'

  if (status && !status.linux) {
    return (
      <div className="glass-card p-8 text-center">
        <AlertCircle className="w-16 h-16 mx-auto mb-4 text-yellow-500" />
        <h2 className={cn('text-xl font-semibold mb-2', textMain)}>{t('wg.linuxOnly')}</h2>
        <p className={textSub}>{t('wg.linuxOnlyDesc')}</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {msg && <div className={cn('fixed top-4 right-4 z-50 px-4 py-2 rounded-lg text-white text-sm', msg.type === 'success' ? 'bg-green-500' : 'bg-red-500')}>{msg.text}</div>}
      <div className="flex items-center justify-between">
        <div>
          <h1 className={cn('text-xl font-bold', textMain)}>{t('wg.title')}</h1>
          <p className={cn('text-sm', textSub)}>{t('wg.desc')}</p>
        </div>
        <div className="flex gap-2">
          <button onClick={loadData} disabled={loading} className={cn('flex items-center gap-2 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}><RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />{t('common.refresh')}</button>
          <button onClick={() => setShowModal('create')} className={cn('flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-white', accent)}><Plus className="w-4 h-4" />{t('wg.newServer')}</button>
        </div>
      </div>

      {status && !status.installed && (
        <div className="glass-card p-4 flex items-center justify-between border border-yellow-500/30 bg-yellow-500/10">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-yellow-500" />
            <span className={textMain}>{t('wg.notInstalled')}</span>
          </div>
          <button 
            onClick={handleInstall} 
            disabled={installing}
            className={cn('flex items-center gap-2 px-4 py-2 rounded-lg text-sm text-white', accent, installing && 'opacity-50')}
          >
            {installing ? <><Loader2 className="w-4 h-4 animate-spin" />{t('wg.installing')}</> : t('wg.clickInstall')}
          </button>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="space-y-3">
          <h2 className={cn('text-sm font-medium flex items-center gap-2', textMain)}><Server className="w-4 h-4" />{t('wg.serverList')}</h2>
          {servers.length === 0 ? <div className="glass-card p-8 text-center"><Server className={cn('w-10 h-10 mx-auto mb-2', textSub)} /><p className={textSub}>{t('wg.noServers')}</p></div> : servers.map(server => (
            <div key={server.id} onClick={() => setSelectedServer(server)} className={cn('glass-card p-4 cursor-pointer min-h-[160px] flex flex-col', selectedServer?.id === server.id && (isGlass ? 'ring-2 ring-blue-500' : 'ring-2 ring-cyan-500'))}>
              <div className="flex items-center justify-between mb-2">
                <span className={cn('font-medium', textMain)}>{server.name}</span>
                <span className={cn('text-xs px-2 py-0.5 rounded', server.enabled ? 'bg-green-500/20 text-green-500' : 'bg-slate-500/20 text-slate-400')}>{server.enabled ? t('wg.running') : t('wg.stopped')}</span>
              </div>
              <div className={cn('text-xs space-y-1 flex-1', textSub)}>
                <div className="flex justify-between"><span>{t('wg.port')}</span><span>{server.listen_port}</span></div>
                <div className="flex justify-between"><span>{t('wg.clients')}</span><span>{server.clients?.length || 0}</span></div>
                {server.endpoint && <div className="flex justify-between"><span>{t('wg.endpoint')}</span><span className="truncate max-w-[100px]">{server.endpoint}</span></div>}
              </div>
              <div className="flex gap-2 mt-3">
                <button onClick={(e) => { e.stopPropagation(); handleToggleServer(server) }} className={cn('flex-1 py-1.5 rounded text-xs text-white', server.enabled ? 'bg-red-500' : accent)}>{server.enabled ? <><Square className="w-3 h-3 inline mr-1" />{t('wg.stop')}</> : <><Play className="w-3 h-3 inline mr-1" />{t('wg.start')}</>}</button>
                <button onClick={(e) => { e.stopPropagation(); openEditServerModal(server) }} className={cn('px-3 py-1.5 rounded text-xs', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}><Settings className="w-3 h-3" /></button>
                <button onClick={(e) => { e.stopPropagation(); setDeleteTarget({ type: 'server', id: server.id, name: server.name }); setShowModal('delete') }} className="px-3 py-1.5 rounded text-xs bg-red-500/20 text-red-400"><Trash2 className="w-3 h-3" /></button>
              </div>
            </div>
          ))}
        </div>

        <div className="lg:col-span-2 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className={cn('text-sm font-medium flex items-center gap-2', textMain)}><Users className="w-4 h-4" />{t('wg.clientList')}{selectedServer && <span className={textSub}>- {selectedServer.name}</span>}</h2>
            {selectedServer && <button onClick={openAddClientModal} className={cn('px-3 py-1.5 rounded-lg text-sm text-white', accent)}><Plus className="w-4 h-4" /></button>}
          </div>
          {!selectedServer ? <div className="glass-card p-12 text-center"><p className={textSub}>{t('wg.selectServer')}</p></div> : !selectedServer.clients?.length ? (
            <div className="glass-card p-12 text-center">
              <Users className={cn('w-12 h-12 mx-auto mb-3', textSub)} />
              <p className={cn('mb-4', textSub)}>{t('wg.noClients')}</p>
              <button onClick={openAddClientModal} className={cn('px-4 py-2 rounded-lg text-sm text-white', accent)}>
                <Plus className="w-4 h-4 inline mr-1" />{t('wg.addClient')}
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {selectedServer.clients.map(client => (
                <div key={client.id} className="glass-card p-4 min-h-[140px] flex flex-col">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center', isGlass ? 'bg-cyan-500/20' : 'bg-cyan-500/30')}>
                        <Users className="w-4 h-4 text-cyan-500" />
                      </div>
                      <div>
                        <div className={cn('font-medium', textMain)}>{client.name}</div>
                        <div className={cn('text-xs font-mono', textSub)}>{client.allowed_ips}</div>
                      </div>
                    </div>
                    <span className={cn('text-xs px-2 py-0.5 rounded', client.enabled ? 'bg-green-500/20 text-green-500' : 'bg-slate-500/20 text-slate-400')}>
                      {client.enabled ? t('wg.enabled') : t('wg.disabled')}
                    </span>
                  </div>
                  {client.description && <div className={cn('text-xs mb-3 flex-1', textSub)}>{client.description}</div>}
                  <div className="flex gap-2 mt-auto">
                    <button onClick={() => handleGetClientConfig(client)} className={cn('flex-1 flex items-center justify-center gap-1 py-2 rounded-lg text-xs', isGlass ? 'bg-black/5 hover:bg-black/10' : 'bg-white/10 hover:bg-white/15', textMain)}>
                      <QrCode className="w-3 h-3" />{t('wg.qrCode')}
                    </button>
                    <button onClick={() => handleGetClientConfig(client)} className={cn('flex-1 flex items-center justify-center gap-1 py-2 rounded-lg text-xs', isGlass ? 'bg-black/5 hover:bg-black/10' : 'bg-white/10 hover:bg-white/15', textMain)}>
                      <Download className="w-3 h-3" />{t('wg.download')}
                    </button>
                    <button onClick={() => openEditClientModal(client)} className={cn('px-3 py-2 rounded-lg text-xs', isGlass ? 'bg-black/5 hover:bg-black/10' : 'bg-white/10 hover:bg-white/15', textMain)}>
                      <Settings className="w-3 h-3" />
                    </button>
                    <button onClick={() => { setDeleteTarget({ type: 'client', id: client.id, name: client.name }); setShowModal('delete') }} className="px-3 py-2 rounded-lg text-xs bg-red-500/20 text-red-400 hover:bg-red-500/30">
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setShowModal(null)}>
          <div onClick={(e) => e.stopPropagation()} className={cn('w-full max-w-md rounded-2xl p-6', isGlass ? 'bg-white/95 backdrop-blur-xl' : 'bg-slate-900')}>
            {showModal === 'create' && (<>
              <h3 className={cn('text-lg font-semibold mb-4', textMain)}>{t('wg.createServer')}</h3>
              <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
                {/* 基本信息 */}
                <div className={cn('text-xs font-medium', textSub)}>{t('wg.basicInfo')}</div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.name')} *</label>
                    <input value={serverForm.name} onChange={(e) => setServerForm({...serverForm, name: e.target.value})} placeholder={t('wg.namePlaceholder')} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  </div>
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.interfaceName')} *</label>
                    <input value={serverForm.tag} onChange={(e) => setServerForm({...serverForm, tag: e.target.value})} placeholder="wg0" className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  </div>
                </div>
                
                {/* 网络配置 */}
                <div className={cn('text-xs font-medium pt-2', textSub)}>{t('wg.networkConfig')}</div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.listenAddr')}</label>
                    <input value={serverForm.listen} onChange={(e) => setServerForm({...serverForm, listen: e.target.value})} placeholder="0.0.0.0" className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  </div>
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.listenPort')}</label>
                    <input type="number" value={serverForm.listen_port} onChange={(e) => setServerForm({...serverForm, listen_port: parseInt(e.target.value) || 51820})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  </div>
                </div>
                <div>
                  <label className={cn('text-xs', textSub)}>{t('wg.serverAddr')}</label>
                  <input value={serverForm.address} onChange={(e) => setServerForm({...serverForm, address: e.target.value})} placeholder="10.0.1.1/24" className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.mtu')}</label>
                    <input type="number" value={serverForm.mtu} onChange={(e) => setServerForm({...serverForm, mtu: parseInt(e.target.value) || 1420})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  </div>
                  <div>
                    <label className={cn('text-xs', textSub)}>{t('wg.dnsServer')}</label>
                    <input value={serverForm.dns} onChange={(e) => setServerForm({...serverForm, dns: e.target.value})} placeholder={localIP ? `${localIP},1.1.1.1` : '1.1.1.1,8.8.8.8'} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                    {localIP && <p className={cn('text-xs mt-1', textSub)}>💡 {t('wg.dnsDetected', { ip: localIP })}</p>}
                  </div>
                </div>
                <div>
                  <label className={cn('text-xs', textSub)}>{t('wg.descOptional')}</label>
                  <input value={serverForm.description} onChange={(e) => setServerForm({...serverForm, description: e.target.value})} placeholder={t('wg.descPlaceholder')} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                
                {/* 提示 */}
                <div className={cn('p-3 rounded-lg text-xs', isGlass ? 'bg-blue-50 border border-blue-200' : 'bg-blue-900/20 border border-blue-800')}>
                  <div className={cn('font-medium mb-1', isGlass ? 'text-blue-800' : 'text-blue-400')}>💡 {t('wg.autoConfig')}</div>
                  <ul className={cn('space-y-0.5', isGlass ? 'text-blue-700' : 'text-blue-300')}>
                    <li>• {t('wg.autoKeyGen')}</li>
                    <li>• {t('wg.autoDns')}</li>
                    <li>• {t('wg.autoTun')}</li>
                  </ul>
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button onClick={() => setShowModal(null)} className={cn('flex-1 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>{t('wg.cancel')}</button>
                <button onClick={handleCreateServer} disabled={!serverForm.name || !serverForm.tag} className={cn('flex-1 py-2 rounded-lg text-sm text-white', accent, (!serverForm.name || !serverForm.tag) && 'opacity-50')}>{t('wg.createService')}</button>
              </div>
            </>)}
            {showModal === 'client' && (<>
              <h3 className={cn('text-lg font-semibold mb-4', textMain)}>{t('wg.addClient')}</h3>
              <div className="space-y-4">
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.clientName')} *</label>
                  <input 
                    value={clientForm.name} 
                    onChange={(e) => setClientForm({...clientForm, name: e.target.value})} 
                    placeholder={t('wg.clientNamePlaceholder')}
                    className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} 
                  />
                </div>
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.clientDesc')}</label>
                  <input 
                    value={clientForm.description} 
                    onChange={(e) => setClientForm({...clientForm, description: e.target.value})} 
                    placeholder={t('wg.clientDescPlaceholder')}
                    className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} 
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={cn('text-xs font-medium', textSub)}>{t('wg.ipAddr')}</label>
                    <input 
                      value={clientForm.allowed_ips} 
                      onChange={(e) => setClientForm({...clientForm, allowed_ips: e.target.value})} 
                      placeholder={t('wg.ipAutoAssign')}
                      className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} 
                    />
                  </div>
                  <div>
                    <label className={cn('text-xs font-medium', textSub)}>{t('wg.dnsOptional')}</label>
                    <input 
                      value={clientForm.dns} 
                      onChange={(e) => setClientForm({...clientForm, dns: e.target.value})} 
                      placeholder={t('wg.dnsInherit')}
                      className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} 
                    />
                  </div>
                </div>
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.pskOptional')}</label>
                  <input 
                    value={clientForm.preshared_key} 
                    onChange={(e) => setClientForm({...clientForm, preshared_key: e.target.value})} 
                    placeholder={t('wg.pskAutoGen')}
                    className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm font-mono', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} 
                  />
                </div>
                <div className={cn('p-3 rounded-lg text-xs', isGlass ? 'bg-blue-50 border border-blue-200' : 'bg-blue-900/20 border border-blue-800')}>
                  <div className={cn('font-medium mb-1', isGlass ? 'text-blue-800' : 'text-blue-400')}>💡 {t('wg.tips')}</div>
                  <ul className={cn('space-y-0.5', isGlass ? 'text-blue-700' : 'text-blue-300')}>
                    <li>• {t('wg.tipKeyAutoGen')}</li>
                    <li>• {t('wg.tipIpAutoAssign')}</li>
                    <li>• {t('wg.tipDnsInherit')}</li>
                  </ul>
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button onClick={() => setShowModal(null)} className={cn('flex-1 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>{t('wg.cancel')}</button>
                <button onClick={handleAddClient} disabled={!clientForm.name} className={cn('flex-1 py-2 rounded-lg text-sm text-white', accent, !clientForm.name && 'opacity-50')}>{t('wg.addClient')}</button>
              </div>
            </>)}
            {showModal === 'config' && (<>
              <div className="flex items-center justify-between mb-4">
                <h3 className={cn('text-lg font-semibold', textMain)}>{t('wg.clientConfig')} {selectedClient && <span className={textSub}>- {selectedClient.name}</span>}</h3>
                <button onClick={() => setShowModal(null)} className={cn('p-1 rounded', isGlass ? 'hover:bg-black/10' : 'hover:bg-white/10')}><X className="w-5 h-5" /></button>
              </div>
              
              {/* 二维码 */}
              {qrCodeURL && (
                <div className="flex flex-col items-center mb-4">
                  <div className={cn('p-3 rounded-lg mb-2', isGlass ? 'bg-white border border-slate-200' : 'bg-white')}>
                    <img src={qrCodeURL} alt="QR Code" className="w-48 h-48" />
                  </div>
                  <button onClick={handleDownloadQR} className={cn('flex items-center gap-1 text-xs px-3 py-1.5 rounded', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>
                    <QrCode className="w-3 h-3" />{t('wg.downloadQr')}
                  </button>
                </div>
              )}
              
              {/* 配置文本 */}
              <div className="mb-4">
                <div className="flex items-center justify-between mb-2">
                  <span className={cn('text-sm', textSub)}>{t('wg.configText')}</span>
                  <div className="flex gap-2">
                    <button onClick={handleCopy} className={cn('flex items-center gap-1 text-xs px-3 py-1.5 rounded', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>
                      {copied ? <><Check className="w-3 h-3 text-green-500" />{t('wg.copied')}</> : <><Copy className="w-3 h-3" />{t('wg.copyConfig')}</>}
                    </button>
                    <button onClick={handleDownloadConfig} className={cn('flex items-center gap-1 text-xs px-3 py-1.5 rounded', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>
                      <Download className="w-3 h-3" />{t('wg.download')}
                    </button>
                  </div>
                </div>
                <pre className={cn('p-3 rounded-lg text-xs font-mono overflow-auto max-h-48', isGlass ? 'bg-black/5' : 'bg-black/30', textMain)}>{clientConfig}</pre>
              </div>
              
              {/* 使用说明 */}
              <div className={cn('p-3 rounded-lg text-xs space-y-1', isGlass ? 'bg-green-50 border border-green-200' : 'bg-green-900/20 border border-green-800')}>
                <div className={cn('font-medium', isGlass ? 'text-green-800' : 'text-green-400')}>✅ {t('wg.usage')}</div>
                <ol className={cn('list-decimal list-inside space-y-0.5', isGlass ? 'text-green-700' : 'text-green-300')}>
                  <li>{t('wg.usageStep1')}</li>
                  <li>{t('wg.usageStep2')}</li>
                  <li>{t('wg.usageStep3')}</li>
                </ol>
              </div>
              
              {/* 服务器信息 */}
              {selectedServer && (
                <div className={cn('mt-4 p-3 rounded-lg grid grid-cols-2 gap-3', isGlass ? 'bg-black/5' : 'bg-white/5')}>
                  <div><span className={cn('text-xs', textSub)}>{t('wg.listenPort')}</span><div className={cn('font-mono text-sm', textMain)}>{selectedServer.listen_port}</div></div>
                  <div><span className={cn('text-xs', textSub)}>{t('wg.ipAddr')}</span><div className={cn('font-mono text-sm', textMain)}>{selectedClient?.allowed_ips}</div></div>
                </div>
              )}
            </>)}
            {showModal === 'delete' && deleteTarget && (<>
              <h3 className={cn('text-lg font-semibold mb-2', textMain)}>{t('wg.confirm')}</h3>
              <p className={cn('text-sm mb-6', textSub)}>{t('wg.confirmDelete', { name: deleteTarget.name })}</p>
              <div className="flex gap-2">
                <button onClick={() => { setShowModal(null); setDeleteTarget(null) }} className={cn('flex-1 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>{t('wg.cancel')}</button>
                <button onClick={handleDelete} className="flex-1 py-2 rounded-lg text-sm text-white bg-red-500">{t('wg.delete')}</button>
              </div>
            </>)}
            {showModal === 'editServer' && editingServer && (<>
              <h3 className={cn('text-lg font-semibold mb-4', textMain)}>{t('wg.editServer')}</h3>
              <div className="space-y-4">
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.serverName')}</label>
                  <input value={editServerForm.name} onChange={(e) => setEditServerForm({...editServerForm, name: e.target.value})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.endpoint')} *</label>
                  <input value={editServerForm.endpoint} onChange={(e) => setEditServerForm({...editServerForm, endpoint: e.target.value})} placeholder={t('wg.endpointPlaceholder')} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                  <p className={cn('text-xs mt-1', textSub)}>{t('wg.endpointHint')}</p>
                </div>
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.desc')}</label>
                  <input value={editServerForm.description} onChange={(e) => setEditServerForm({...editServerForm, description: e.target.value})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                <div className="flex items-center justify-between">
                  <label className={cn('text-sm', textMain)}>{t('wg.autoStart')}</label>
                  <button onClick={() => setEditServerForm({...editServerForm, auto_start: !editServerForm.auto_start})} className={cn('w-12 h-6 rounded-full transition-colors', editServerForm.auto_start ? accent : (isGlass ? 'bg-black/10' : 'bg-white/20'))}>
                    <div className={cn('w-5 h-5 rounded-full bg-white shadow transition-transform', editServerForm.auto_start ? 'translate-x-6' : 'translate-x-0.5')} />
                  </button>
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button onClick={() => { setShowModal(null); setEditingServer(null) }} className={cn('flex-1 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>{t('wg.cancel')}</button>
                <button onClick={handleUpdateServer} className={cn('flex-1 py-2 rounded-lg text-sm text-white', accent)}>{t('wg.save')}</button>
              </div>
            </>)}
            {showModal === 'editClient' && editingClient && (<>
              <h3 className={cn('text-lg font-semibold mb-4', textMain)}>{t('wg.editClient')}</h3>
              <div className="space-y-4">
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.clientName')}</label>
                  <input value={editClientForm.name} onChange={(e) => setEditClientForm({...editClientForm, name: e.target.value})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                <div>
                  <label className={cn('text-xs font-medium', textSub)}>{t('wg.clientDesc')}</label>
                  <input value={editClientForm.description} onChange={(e) => setEditClientForm({...editClientForm, description: e.target.value})} className={cn('w-full mt-1 px-3 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)} />
                </div>
                <div className="flex items-center justify-between">
                  <label className={cn('text-sm', textMain)}>{t('wg.clientEnabled')}</label>
                  <button onClick={() => setEditClientForm({...editClientForm, enabled: !editClientForm.enabled})} className={cn('w-12 h-6 rounded-full transition-colors', editClientForm.enabled ? 'bg-green-500' : (isGlass ? 'bg-black/10' : 'bg-white/20'))}>
                    <div className={cn('w-5 h-5 rounded-full bg-white shadow transition-transform', editClientForm.enabled ? 'translate-x-6' : 'translate-x-0.5')} />
                  </button>
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button onClick={() => { setShowModal(null); setEditingClient(null) }} className={cn('flex-1 py-2 rounded-lg text-sm', isGlass ? 'bg-black/5' : 'bg-white/10', textMain)}>{t('wg.cancel')}</button>
                <button onClick={handleUpdateClient} className={cn('flex-1 py-2 rounded-lg text-sm text-white', accent)}>{t('wg.save')}</button>
              </div>
            </>)}
          </div>
        </div>
      )}
    </div>
  )
}
