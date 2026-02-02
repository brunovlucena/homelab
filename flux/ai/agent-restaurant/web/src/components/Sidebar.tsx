'use client'

import { useRestaurantStore, ViewType } from '@/store/restaurantStore'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  MapPin,
  ChefHat,
  Users,
  UtensilsCrossed,
  Activity,
  Wine,
} from 'lucide-react'

const menuItems: { id: ViewType; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'floor', label: 'Floor Plan', icon: MapPin },
  { id: 'kitchen', label: 'Kitchen', icon: ChefHat },
  { id: 'agents', label: 'AI Agents', icon: Users },
  { id: 'menu', label: 'Menu', icon: UtensilsCrossed },
  { id: 'events', label: 'Events', icon: Activity },
]

export function Sidebar() {
  const { activeView, setActiveView, agents } = useRestaurantStore()
  
  const activeAgents = agents.filter(a => a.status !== 'offline').length

  return (
    <aside className="fixed left-0 top-0 w-64 h-screen bg-wine-900 text-white z-50">
      {/* Logo */}
      <div className="h-16 flex items-center justify-center border-b border-wine-800">
        <div className="flex items-center gap-3">
          <Wine className="w-8 h-8 text-gold-500" />
          <div>
            <h1 className="font-serif text-xl font-bold">Ristorante</h1>
            <p className="text-xs text-wine-300 -mt-1">Stellare</p>
          </div>
        </div>
      </div>
      
      {/* Navigation */}
      <nav className="p-4 space-y-2">
        {menuItems.map((item) => {
          const Icon = item.icon
          const isActive = activeView === item.id
          
          return (
            <button
              key={item.id}
              onClick={() => setActiveView(item.id)}
              className={cn(
                "w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all",
                isActive
                  ? "bg-gold-500 text-wine-900 font-medium"
                  : "text-wine-200 hover:bg-wine-800 hover:text-white"
              )}
            >
              <Icon className="w-5 h-5" />
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>
      
      {/* Agent Status */}
      <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-wine-800">
        <div className="bg-wine-800 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm text-wine-300">AI Agents</span>
            <span className="text-xs px-2 py-1 rounded-full bg-emerald-500/20 text-emerald-400">
              {activeAgents} active
            </span>
          </div>
          <div className="flex -space-x-2">
            {agents.map((agent) => (
              <div
                key={agent.id}
                className={cn(
                  "w-10 h-10 rounded-full flex items-center justify-center text-lg border-2",
                  agent.status === 'offline'
                    ? "bg-gray-600 border-gray-500 opacity-50"
                    : agent.status === 'busy'
                    ? "bg-gold-500 border-gold-400"
                    : "bg-wine-700 border-wine-600"
                )}
                title={`${agent.name} - ${agent.status}`}
              >
                {agent.avatar}
              </div>
            ))}
          </div>
        </div>
      </div>
    </aside>
  )
}
