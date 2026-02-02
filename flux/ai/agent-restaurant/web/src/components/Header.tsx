'use client'

import { useRestaurantStore } from '@/store/restaurantStore'
import { Bell, Settings, User, Clock } from 'lucide-react'
import { useState, useEffect } from 'react'

export function Header() {
  const { stats } = useRestaurantStore()
  const [currentTime, setCurrentTime] = useState(new Date())
  
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  return (
    <header className="fixed top-0 left-64 right-0 h-16 bg-white border-b border-cream-200 z-40">
      <div className="flex items-center justify-between h-full px-6">
        {/* Left: Time and Status */}
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2 text-wood-600">
            <Clock className="w-5 h-5" />
            <span className="font-medium">
              {currentTime.toLocaleTimeString('en-US', {
                hour: '2-digit',
                minute: '2-digit',
              })}
            </span>
          </div>
          <div className="h-6 w-px bg-cream-300" />
          <div className="flex items-center gap-4 text-sm">
            <span className="text-wood-500">Tonight:</span>
            <span className="font-semibold text-wine-900">{stats.guestsTonight} guests</span>
            <span className="text-gold-600 font-semibold">${stats.revenue.toLocaleString()}</span>
          </div>
        </div>
        
        {/* Right: Actions */}
        <div className="flex items-center gap-4">
          {/* Satisfaction Score */}
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-50 border border-emerald-200">
            <span className="text-lg">⭐</span>
            <span className="font-semibold text-emerald-700">{stats.satisfaction}</span>
          </div>
          
          {/* Notifications */}
          <button className="relative p-2 rounded-lg hover:bg-cream-100 transition-colors">
            <Bell className="w-5 h-5 text-wood-600" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-wine-600 rounded-full" />
          </button>
          
          {/* Settings */}
          <button className="p-2 rounded-lg hover:bg-cream-100 transition-colors">
            <Settings className="w-5 h-5 text-wood-600" />
          </button>
          
          {/* Profile */}
          <button className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-cream-100 transition-colors">
            <div className="w-8 h-8 rounded-full bg-wine-900 flex items-center justify-center text-white text-sm font-medium">
              GM
            </div>
            <span className="text-sm font-medium text-wood-700">Manager</span>
          </button>
        </div>
      </div>
    </header>
  )
}
