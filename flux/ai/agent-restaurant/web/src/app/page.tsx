'use client'

import { useState } from 'react'
import { Header } from '@/components/Header'
import { Sidebar } from '@/components/Sidebar'
import { Dashboard } from '@/components/Dashboard'
import { FloorPlan } from '@/components/FloorPlan'
import { KitchenView } from '@/components/KitchenView'
import { AgentPanel } from '@/components/AgentPanel'
import { MenuManager } from '@/components/MenuManager'
import { EventFeed } from '@/components/EventFeed'
import { useRestaurantStore } from '@/store/restaurantStore'

export default function Home() {
  const { activeView } = useRestaurantStore()

  const renderView = () => {
    switch (activeView) {
      case 'dashboard':
        return <Dashboard />
      case 'floor':
        return <FloorPlan />
      case 'kitchen':
        return <KitchenView />
      case 'agents':
        return <AgentPanel />
      case 'menu':
        return <MenuManager />
      case 'events':
        return <EventFeed />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-cream-100">
      <Header />
      <Sidebar />
      <main className="pl-64 pt-16 min-h-screen">
        <div className="p-6">
          {renderView()}
        </div>
      </main>
    </div>
  )
}
