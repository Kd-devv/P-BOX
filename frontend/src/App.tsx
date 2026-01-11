import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { Toaster } from 'sonner'
import { useState, useEffect, ReactNode } from 'react'
import { Loader2 } from 'lucide-react'
import { Layout } from '@/components/layout'
import DashboardPage from '@/pages/DashboardPage'
import ProxySwitchPage from '@/pages/ProxySwitchPage'
import NodesPage from '@/pages/NodesPage'
import SubscriptionsPage from '@/pages/SubscriptionsPage'
import ConnectionsPage from '@/pages/ConnectionsPage'
import LogsPage from '@/pages/LogsPage'
import RulesetPage from '@/pages/RulesetPage'
import SettingsPage from '@/pages/SettingsPage'
import ToolsPage from '@/pages/ToolsPage'
import ConfigGeneratorPage from '@/pages/ConfigGeneratorPage'
import CoreManagePage from '@/pages/CoreManagePage'
import ProxySettingsPage from '@/pages/ProxySettingsPage'
import LoginPage from '@/pages/LoginPage'
import WireGuardPage from '@/pages/WireGuardPage'
import LegalPage from '@/pages/LegalPage'
import SingBoxSettingsPage from '@/pages/SingBoxSettingsPage'
import SingBoxRulesetPage from '@/pages/SingBoxRulesetPage'
import { authApi } from '@/api/auth'

// Auth guard component
function AuthGuard({ children }: { children: ReactNode }) {
  const location = useLocation()
  const [checking, setChecking] = useState(true)
  const [needLogin, setNeedLogin] = useState(false)

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const status = await authApi.check()
        if (status.enabled && !status.authenticated) {
          setNeedLogin(true)
        }
      } catch {
        // Ignore errors
      } finally {
        setChecking(false)
      }
    }
    checkAuth()
  }, [location.pathname])

  if (checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    )
  }

  if (needLogin) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/*" element={
          <AuthGuard>
            <Layout>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/proxy-switch" element={<ProxySwitchPage />} />
                <Route path="/nodes" element={<NodesPage />} />
                <Route path="/subscriptions" element={<SubscriptionsPage />} />
                <Route path="/connections" element={<ConnectionsPage />} />
                <Route path="/logs" element={<LogsPage />} />
                <Route path="/ruleset" element={<RulesetPage />} />
                <Route path="/tools" element={<ToolsPage />} />
                <Route path="/config-generator" element={<ConfigGeneratorPage />} />
                <Route path="/core-manage" element={<CoreManagePage />} />
                <Route path="/proxy-settings" element={<ProxySettingsPage />} />
                <Route path="/singbox-settings" element={<SingBoxSettingsPage />} />
                <Route path="/singbox-ruleset" element={<SingBoxRulesetPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/wireguard" element={<WireGuardPage />} />
                <Route path="/legal" element={<LegalPage />} />
              </Routes>
            </Layout>
          </AuthGuard>
        } />
      </Routes>
      <Toaster position="bottom-right" richColors />
    </>
  )
}

export default App
