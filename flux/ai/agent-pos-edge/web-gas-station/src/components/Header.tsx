'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { cn } from '@/lib/utils'
import { Menu, Bell, Fuel, Activity, Clock, Settings } from 'lucide-react'
import { useState, useEffect } from 'react'

export function Header() {
  const { stationName, stationId, toggleSidebar, alerts } = useGasStationStore()
  const [currentTime, setCurrentTime] = useState(new Date())
  
  const unacknowledgedAlerts = alerts.filter(a => !a.acknowledged).length

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  return (
    <header className="fixed top-0 left-0 right-0 h-16 z-50 bg-fuel-dark/95 backdrop-blur-md border-b border-fuel-gray/50">
      <div className="h-full px-4 flex items-center justify-between">
        {/* Left Section */}
        <div className="flex items-center gap-4">
          <button
            onClick={toggleSidebar}
            className="p-2 rounded-lg hover:bg-fuel-gray/50 transition-colors"
          >
            <Menu className="w-5 h-5 text-gray-400" />
          </button>
          
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-fuel-green/10 border border-fuel-green/30">
              <Fuel className="w-6 h-6 text-fuel-green" />
            </div>
            <div>
              <h1 className="font-bold text-white text-lg leading-tight">{stationName}</h1>
              <p className="text-xs text-gray-500 font-mono">{stationId}</p>
            </div>
          </div>
        </div>

        {/* Center Section - Live Status */}
        <div className="hidden md:flex items-center gap-6">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-fuel-green/10 border border-fuel-green/30">
            <Activity className="w-4 h-4 text-fuel-green animate-pulse" />
            <span className="text-sm font-mono text-fuel-green">ONLINE</span>
          </div>
          
          <div className="flex items-center gap-2 text-gray-400">
            <Clock className="w-4 h-4" />
            <span className="font-mono text-sm">
              {currentTime.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            </span>
          </div>
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-3">
          <button className="relative p-2 rounded-lg hover:bg-fuel-gray/50 transition-colors">
            <Bell className="w-5 h-5 text-gray-400" />
            {unacknowledgedAlerts > 0 && (
              <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-fuel-red text-white text-xs flex items-center justify-center font-bold animate-pulse">
                {unacknowledgedAlerts}
              </span>
            )}
          </button>
          
          <button className="p-2 rounded-lg hover:bg-fuel-gray/50 transition-colors">
            <Settings className="w-5 h-5 text-gray-400" />
          </button>
          
          <div className="w-px h-8 bg-fuel-gray/50" />
          
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-fuel-green to-fuel-lime flex items-center justify-center">
              <span className="text-sm font-bold text-fuel-black">OP</span>
            </div>
            <span className="hidden md:block text-sm text-gray-300">Operador</span>
          </div>
        </div>
      </div>
    </header>
  )
}
