import { ReactNode, useState, createContext, useContext } from 'react'
import Sidebar from './Sidebar'
import Header from './Header'

interface LayoutProps {
  children: ReactNode
}

// Sidebar state context
interface SidebarContextType {
  isOpen: boolean
  toggle: () => void
  close: () => void
}

const SidebarContext = createContext<SidebarContextType>({
  isOpen: false,
  toggle: () => {},
  close: () => {},
})

export const useSidebar = () => useContext(SidebarContext)

export default function Layout({ children }: LayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const sidebarContext = {
    isOpen: sidebarOpen,
    toggle: () => setSidebarOpen(!sidebarOpen),
    close: () => setSidebarOpen(false),
  }

  return (
    <SidebarContext.Provider value={sidebarContext}>
      <div className="flex h-screen overflow-hidden">
        {/* Mobile overlay */}
        {sidebarOpen && (
          <div 
            className="fixed inset-0 bg-black/50 z-40 lg:hidden backdrop-blur-sm"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        {/* Sidebar */}
        <Sidebar />

        {/* Main content area - with glass effect for Apple Glass theme */}
        <div className="main-content flex-1 flex flex-col overflow-hidden min-w-0">
          {/* Header */}
          <div className="relative z-50">
            <Header />
          </div>

          {/* Page content */}
          <main className="flex-1 overflow-auto p-4 lg:p-6 relative z-10">
            <div className="animate-fadeIn">{children}</div>
          </main>
        </div>
      </div>
    </SidebarContext.Provider>
  )
}
