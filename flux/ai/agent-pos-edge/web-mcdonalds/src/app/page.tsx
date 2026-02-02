'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { Header } from '@/components/Header'
import { Sidebar } from '@/components/Sidebar'
import { Dashboard } from '@/components/Dashboard'
import { KitchenDisplay } from '@/components/KitchenDisplay'
import { OrderQueue } from '@/components/OrderQueue'
import { MenuManagement } from '@/components/MenuManagement'
import { StaffPanel } from '@/components/StaffPanel'
import { Analytics } from '@/components/Analytics'
import { SmartCameras } from '@/components/SmartCameras'
import { AgentPanel } from '@/components/AgentPanel'

export default function Home() {
  const { activeView, sidebarOpen } = useMcdonaldsStore()

  const renderView = () => {
    switch (activeView) {
      case 'dashboard':
        return <Dashboard />
      case 'kitchen':
        return <KitchenDisplay />
      case 'orders':
        return <OrderQueue />
      case 'cameras':
        return <SmartCameras />
      case 'agents':
        return <AgentPanel />
      case 'menu':
        return <MenuManagement />
      case 'staff':
        return <StaffPanel />
      case 'analytics':
        return <Analytics />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-mc-black">
      <Header />
      <Sidebar />
      <main className={`pt-16 min-h-screen transition-all duration-300 ${sidebarOpen ? 'pl-64' : 'pl-0'}`}>
        <div className="h-[calc(100vh-64px)] overflow-y-auto">
          {renderView()}
        </div>
      </main>
    </div>
  )
}
