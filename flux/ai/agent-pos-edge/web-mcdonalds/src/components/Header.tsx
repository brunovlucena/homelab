'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { Menu, Bell, Clock, ChefHat, Activity } from 'lucide-react'
import { useState, useEffect } from 'react'

export function Header() {
  const { storeName, storeId, toggleSidebar, orders } = useMcdonaldsStore()
  const [currentTime, setCurrentTime] = useState(new Date())
  
  const pendingOrders = orders.filter(o => o.status === 'new' || o.status === 'preparing').length
  const readyOrders = orders.filter(o => o.status === 'ready').length

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  return (
    <header className="fixed top-0 left-0 right-0 h-16 z-50 bg-mc-dark/95 backdrop-blur-md border-b border-mc-gray/50">
      <div className="h-full px-4 flex items-center justify-between">
        {/* Left Section */}
        <div className="flex items-center gap-4">
          <button
            onClick={toggleSidebar}
            className="p-2 rounded-lg hover:bg-mc-gray/50 transition-colors"
          >
            <Menu className="w-5 h-5 text-gray-400" />
          </button>
          
          <div className="flex items-center gap-3">
            {/* McDonald's Logo */}
            <div className="w-10 h-10 rounded-xl golden-gradient flex items-center justify-center shadow-mc-gold">
              <span className="text-mc-red font-brand font-extrabold text-xl">M</span>
            </div>
            <div>
              <h1 className="font-brand font-bold text-white text-lg leading-tight">{storeName}</h1>
              <p className="text-xs text-gray-500 font-mono">{storeId}</p>
            </div>
          </div>
        </div>

        {/* Center Section - Live Status */}
        <div className="hidden md:flex items-center gap-6">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-mc-red/10 border border-mc-red/30">
            <ChefHat className="w-4 h-4 text-mc-red" />
            <span className="text-sm font-mono text-mc-red">{pendingOrders} em preparo</span>
          </div>
          
          {readyOrders > 0 && (
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-mc-green/10 border border-mc-green/30 animate-pulse">
              <Activity className="w-4 h-4 text-mc-green" />
              <span className="text-sm font-mono text-mc-green">{readyOrders} prontos</span>
            </div>
          )}
          
          <div className="flex items-center gap-2 text-gray-400">
            <Clock className="w-4 h-4" />
            <span className="font-mono text-sm">
              {currentTime.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            </span>
          </div>
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-3">
          <button className="relative p-2 rounded-lg hover:bg-mc-gray/50 transition-colors">
            <Bell className="w-5 h-5 text-gray-400" />
            {readyOrders > 0 && (
              <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-mc-green text-white text-xs flex items-center justify-center font-bold animate-bounce">
                {readyOrders}
              </span>
            )}
          </button>
          
          <div className="w-px h-8 bg-mc-gray/50" />
          
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-mc-red to-mc-gold flex items-center justify-center">
              <span className="text-sm font-bold text-white">G</span>
            </div>
            <span className="hidden md:block text-sm text-gray-300">Gerente</span>
          </div>
        </div>
      </div>
    </header>
  )
}
