'use client'

import { useGasStationStore, ViewType } from '@/store/gasStationStore'
import { cn } from '@/lib/utils'
import { 
  LayoutDashboard, 
  Database, 
  Gauge, 
  Receipt, 
  AlertTriangle, 
  Bot,
  ChevronRight,
  Droplets,
  Camera,
  Brain
} from 'lucide-react'

const navItems: { id: ViewType; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'cameras', label: 'Câmeras & IA', icon: Camera },
  { id: 'tanks', label: 'Tanques', icon: Database },
  { id: 'pumps', label: 'Bombas', icon: Gauge },
  { id: 'transactions', label: 'Transações', icon: Receipt },
  { id: 'alerts', label: 'Alertas', icon: AlertTriangle },
  { id: 'agents', label: 'Agentes', icon: Bot },
]

export function Sidebar() {
  const { activeView, setActiveView, sidebarOpen, tanks, pumps, alerts, cameras, agents } = useGasStationStore()

  const activePumps = pumps.filter(p => p.status === 'active').length
  const criticalTanks = tanks.filter(t => t.status === 'critical').length
  const unackedAlerts = alerts.filter(a => !a.acknowledged).length
  const analyzingCameras = cameras.filter(c => c.status === 'analyzing' || c.status === 'recording').length
  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'security').length

  if (!sidebarOpen) return null

  return (
    <aside className="fixed left-0 top-16 bottom-0 w-64 bg-fuel-dark/95 backdrop-blur-md border-r border-fuel-gray/50 z-40">
      <div className="flex flex-col h-full">
        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = activeView === item.id
            let badge = null
            
            if (item.id === 'pumps' && activePumps > 0) {
              badge = <span className="px-2 py-0.5 rounded-full bg-fuel-green/20 text-fuel-green text-xs font-bold">{activePumps}</span>
            }
            if (item.id === 'tanks' && criticalTanks > 0) {
              badge = <span className="px-2 py-0.5 rounded-full bg-fuel-red/20 text-fuel-red text-xs font-bold">{criticalTanks}</span>
            }
            if (item.id === 'alerts' && unackedAlerts > 0) {
              badge = <span className="px-2 py-0.5 rounded-full bg-fuel-amber/20 text-fuel-amber text-xs font-bold">{unackedAlerts}</span>
            }
            if (item.id === 'cameras') {
              badge = <span className="px-2 py-0.5 rounded-full bg-fuel-blue/20 text-fuel-blue text-xs font-bold">{analyzingCameras}</span>
            }
            if (item.id === 'agents') {
              badge = <span className="px-2 py-0.5 rounded-full bg-fuel-cyan/20 text-fuel-cyan text-xs font-bold">{visionAgents}</span>
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
                <item.icon className={cn('w-5 h-5', isActive ? 'text-fuel-green' : 'text-gray-400')} />
                <span className="flex-1 text-left">{item.label}</span>
                {badge}
                {isActive && <ChevronRight className="w-4 h-4 text-fuel-green" />}
              </button>
            )
          })}
        </nav>

        {/* Quick Stats */}
        <div className="p-4 border-t border-fuel-gray/50">
          <h3 className="text-xs uppercase tracking-wider text-gray-500 mb-3">Resumo Rápido</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplets className="w-4 h-4 text-fuel-green" />
                <span className="text-sm text-gray-400">Gasolina</span>
              </div>
              <span className="text-sm font-mono text-fuel-green">62%</span>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplets className="w-4 h-4 text-diesel" />
                <span className="text-sm text-gray-400">Diesel</span>
              </div>
              <span className="text-sm font-mono text-fuel-amber">25%</span>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplets className="w-4 h-4 text-premium" />
                <span className="text-sm text-gray-400">Premium</span>
              </div>
              <span className="text-sm font-mono text-fuel-green">79%</span>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplets className="w-4 h-4 text-fuel-lime" />
                <span className="text-sm text-gray-400">Etanol</span>
              </div>
              <span className="text-sm font-mono text-fuel-red">14%</span>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-fuel-gray/50">
          <div className="text-center">
            <p className="text-xs text-gray-500">POS Edge Agent</p>
            <p className="text-xs text-fuel-green font-mono">v0.1.0</p>
          </div>
        </div>
      </div>
    </aside>
  )
}
