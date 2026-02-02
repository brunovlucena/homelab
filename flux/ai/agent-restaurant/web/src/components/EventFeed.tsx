'use client'

import { useRestaurantStore } from '@/store/restaurantStore'
import { cn, formatTime } from '@/lib/utils'
import { motion, AnimatePresence } from 'framer-motion'
import { Radio, Filter, Search } from 'lucide-react'
import { useState, useEffect } from 'react'

// Mock events for demo
const mockEvents = [
  {
    id: '1',
    type: 'restaurant.guest.seated',
    source: '/agent-restaurant/host/maximilian',
    data: { tableId: 'table-01', guestName: 'Mr. Santos', partySize: 2 },
    timestamp: new Date(Date.now() - 5000).toISOString(),
  },
  {
    id: '2',
    type: 'restaurant.order.created',
    source: '/agent-restaurant/waiter/pierre',
    data: { tableId: 'table-01', items: ['Risotto ai Porcini', 'Branzino'] },
    timestamp: new Date(Date.now() - 120000).toISOString(),
  },
  {
    id: '3',
    type: 'restaurant.kitchen.dish.started',
    source: '/agent-restaurant/kitchen/marco',
    data: { dish: 'Risotto ai Porcini', station: 'saute' },
    timestamp: new Date(Date.now() - 180000).toISOString(),
  },
  {
    id: '4',
    type: 'restaurant.sommelier.pairing',
    source: '/agent-restaurant/sommelier/isabella',
    data: { tableId: 'table-01', wine: 'Gavi di Gavi 2020', dish: 'Risotto ai Porcini' },
    timestamp: new Date(Date.now() - 240000).toISOString(),
  },
  {
    id: '5',
    type: 'restaurant.service.presentation',
    source: '/agent-restaurant/waiter/pierre',
    data: { tableId: 'table-01', dish: 'Risotto ai Porcini', narrative: 'Chef Marco\'s signature dish...' },
    timestamp: new Date(Date.now() - 300000).toISOString(),
  },
]

const eventTypeInfo: Record<string, { emoji: string; color: string; label: string }> = {
  'restaurant.guest.arrived': { emoji: '🚪', color: 'wine', label: 'Guest Arrived' },
  'restaurant.guest.seated': { emoji: '🪑', color: 'emerald', label: 'Guest Seated' },
  'restaurant.guest.departed': { emoji: '👋', color: 'wood', label: 'Guest Left' },
  'restaurant.order.created': { emoji: '📝', color: 'gold', label: 'New Order' },
  'restaurant.order.served': { emoji: '🍽️', color: 'emerald', label: 'Order Served' },
  'restaurant.kitchen.dish.started': { emoji: '🔥', color: 'gold', label: 'Cooking Started' },
  'restaurant.kitchen.dish.ready': { emoji: '✅', color: 'emerald', label: 'Dish Ready' },
  'restaurant.sommelier.pairing': { emoji: '🍷', color: 'wine', label: 'Wine Pairing' },
  'restaurant.service.presentation': { emoji: '✨', color: 'gold', label: 'Dish Presented' },
}

export function EventFeed() {
  const { events, addEvent } = useRestaurantStore()
  const [filter, setFilter] = useState<string>('all')
  const [search, setSearch] = useState('')
  
  // Initialize with mock events
  useEffect(() => {
    if (events.length === 0) {
      mockEvents.forEach((event, idx) => {
        setTimeout(() => addEvent(event), idx * 100)
      })
    }
  }, [])
  
  // Simulate new events
  useEffect(() => {
    const interval = setInterval(() => {
      const randomEvent = mockEvents[Math.floor(Math.random() * mockEvents.length)]
      addEvent({
        ...randomEvent,
        id: `evt-${Date.now()}`,
        timestamp: new Date().toISOString(),
      })
    }, 15000)
    
    return () => clearInterval(interval)
  }, [addEvent])
  
  const displayEvents = events.length > 0 ? events : mockEvents
  
  const filteredEvents = displayEvents.filter(event => {
    if (filter !== 'all' && !event.type.includes(filter)) return false
    if (search && !JSON.stringify(event).toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-serif text-2xl font-bold text-wine-900">Event Feed</h1>
          <p className="text-wood-500">Real-time CloudEvents from all agents</p>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
          <span className="text-sm text-wood-500">Live</span>
        </div>
      </div>
      
      {/* Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-wood-400" />
          <input
            type="text"
            placeholder="Search events..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-cream-300 rounded-lg"
          />
        </div>
        
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="px-4 py-2 border border-cream-300 rounded-lg bg-white"
        >
          <option value="all">All Events</option>
          <option value="guest">Guest Events</option>
          <option value="order">Order Events</option>
          <option value="kitchen">Kitchen Events</option>
          <option value="service">Service Events</option>
          <option value="sommelier">Wine Events</option>
        </select>
      </div>
      
      {/* Event List */}
      <div className="elegant-card p-4 max-h-[calc(100vh-300px)] overflow-y-auto">
        <AnimatePresence mode="popLayout">
          {filteredEvents.length === 0 ? (
            <div className="text-center py-12 text-wood-400">
              <Radio className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No events to display</p>
            </div>
          ) : (
            <div className="space-y-2">
              {filteredEvents.map((event, index) => {
                const info = eventTypeInfo[event.type] || { 
                  emoji: '📡', 
                  color: 'wood', 
                  label: event.type.split('.').pop() 
                }
                
                return (
                  <motion.div
                    key={event.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 20 }}
                    transition={{ delay: index * 0.02 }}
                    className={cn(
                      "flex items-start gap-4 p-4 rounded-lg bg-cream-50 border-l-4 hover:bg-cream-100 transition-colors",
                      info.color === 'wine' && "border-wine-500",
                      info.color === 'gold' && "border-gold-500",
                      info.color === 'emerald' && "border-emerald-500",
                      info.color === 'wood' && "border-wood-500",
                    )}
                  >
                    <div className="text-2xl">{info.emoji}</div>
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-wine-900">{info.label}</span>
                        <span className="text-xs text-wood-400 font-mono">
                          {event.type}
                        </span>
                      </div>
                      
                      <p className="text-sm text-wood-600 mt-1">
                        {formatEventData(event.type, event.data)}
                      </p>
                      
                      <div className="flex items-center gap-3 mt-2 text-xs text-wood-400">
                        <span>Source: {event.source.split('/').pop()}</span>
                      </div>
                    </div>
                    
                    <div className="text-xs text-wood-400 whitespace-nowrap">
                      {formatTime(event.timestamp)}
                    </div>
                  </motion.div>
                )
              })}
            </div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}

function formatEventData(type: string, data: Record<string, unknown>): string {
  switch (type) {
    case 'restaurant.guest.seated':
      return `${data.guestName} seated at ${data.tableId} (party of ${data.partySize})`
    case 'restaurant.order.created':
      return `New order at ${data.tableId}: ${(data.items as string[]).join(', ')}`
    case 'restaurant.kitchen.dish.started':
      return `Started cooking ${data.dish} at ${data.station} station`
    case 'restaurant.sommelier.pairing':
      return `Suggested ${data.wine} to pair with ${data.dish} for ${data.tableId}`
    case 'restaurant.service.presentation':
      return `Presented ${data.dish} at ${data.tableId}`
    default:
      return JSON.stringify(data)
  }
}
