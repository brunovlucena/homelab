'use client'

import { useMcdonaldsStore, ViewType } from '@/store/mcdonaldsStore'
import { cn } from '@/lib/utils'
import { 
  LayoutDashboard, 
  ChefHat, 
  ClipboardList, 
  UtensilsCrossed, 
  Users,
  BarChart3,
  ChevronRight,
  Camera,
  Bot,
  Brain
} from 'lucide-react'

const navItems: { id: ViewType; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'kitchen', label: 'Cozinha', icon: ChefHat },
  { id: 'orders', label: 'Pedidos', icon: ClipboardList },
  { id: 'cameras', label: 'Câmeras & AI', icon: Camera },
  { id: 'agents', label: 'AI Agents', icon: Bot },
  { id: 'menu', label: 'Cardápio', icon: UtensilsCrossed },
  { id: 'staff', label: 'Equipe', icon: Users },
  { id: 'analytics', label: 'Relatórios', icon: BarChart3 },
]

export function Sidebar() {
  const { activeView, setActiveView, sidebarOpen, orders, stations, cameras, agents } = useMcdonaldsStore()

  const newOrders = orders.filter(o => o.status === 'new').length
  const preparingOrders = orders.filter(o => o.status === 'preparing').length
  const readyOrders = orders.filter(o => o.status === 'ready').length
  const busyStations = stations.filter(s => s.status === 'busy').length
  const onlineCameras = cameras.filter(c => c.status !== 'offline').length
  const activeAgents = agents.filter(a => a.status === 'online' || a.status === 'processing').length

  if (!sidebarOpen) return null

  return (
    <aside className="fixed left-0 top-16 bottom-0 w-64 bg-mc-dark/95 backdrop-blur-md border-r border-mc-gray/50 z-40">
      <div className="flex flex-col h-full">
        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = activeView === item.id
            let badge = null
            
            if (item.id === 'kitchen' && preparingOrders > 0) {
              badge = <span className="px-2 py-0.5 rounded-full bg-mc-orange/20 text-mc-orange text-xs font-bold">{preparingOrders}</span>
            }
            if (item.id === 'orders' && newOrders > 0) {
              badge = <span className="px-2 py-0.5 rounded-full bg-mc-red/20 text-mc-red text-xs font-bold animate-pulse">{newOrders}</span>
            }
            if (item.id === 'cameras') {
              badge = <span className="px-2 py-0.5 rounded-full bg-mc-green/20 text-mc-green text-xs font-bold">{onlineCameras}</span>
            }
            if (item.id === 'agents') {
              badge = <span className="px-2 py-0.5 rounded-full bg-mc-gold/20 text-mc-gold text-xs font-bold">{activeAgents}</span>
            }
            
            return (
              <button
                key={item.id}
                onClick={() => setActiveView(item.id)}
                className={cn(
                  'nav-item w-full',
                  isActive && 'nav-item-active'
                )}
              >
                <item.icon className={cn('w-5 h-5', isActive ? 'text-mc-gold' : 'text-gray-400')} />
                <span className="flex-1 text-left">{item.label}</span>
                {badge}
                {isActive && <ChevronRight className="w-4 h-4 text-mc-gold" />}
              </button>
            )
          })}
        </nav>

        {/* Quick Stats */}
        <div className="p-4 border-t border-mc-gray/50">
          <h3 className="text-xs uppercase tracking-wider text-gray-500 mb-3">Status Rápido</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-400">Novos</span>
              <span className="text-sm font-mono text-status-new">{newOrders}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-400">Preparando</span>
              <span className="text-sm font-mono text-status-preparing">{preparingOrders}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-400">Prontos</span>
              <span className="text-sm font-mono text-status-ready">{readyOrders}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-400">AI Agents</span>
              <span className="text-sm font-mono text-mc-gold">{activeAgents}/{agents.length}</span>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-mc-gray/50">
          <div className="text-center">
            <p className="text-xs text-gray-500">Kitchen Agent + AI Vision</p>
            <p className="text-xs text-mc-gold font-mono">v0.2.0</p>
          </div>
        </div>
      </div>
    </aside>
  )
}
