'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { Header } from '@/components/Header'
import { Sidebar } from '@/components/Sidebar'
import { Dashboard } from '@/components/Dashboard'
import { TankMonitor } from '@/components/TankMonitor'
import { PumpControl } from '@/components/PumpControl'
import { TransactionLog } from '@/components/TransactionLog'
import { AlertPanel } from '@/components/AlertPanel'
import { AgentStatus } from '@/components/AgentStatus'
import { SmartCameras } from '@/components/SmartCameras'

export default function Home() {
  const { activeView, sidebarOpen } = useGasStationStore()

  const renderView = () => {
    switch (activeView) {
      case 'dashboard':
        return <Dashboard />
      case 'cameras':
        return <SmartCameras />
      case 'tanks':
        return <TankMonitor />
      case 'pumps':
        return <PumpControl />
      case 'transactions':
        return <TransactionLog />
      case 'alerts':
        return <AlertPanel />
      case 'agents':
        return <AgentStatus />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-fuel-black">
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
