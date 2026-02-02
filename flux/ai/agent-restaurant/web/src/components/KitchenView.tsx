'use client'

import { useRestaurantStore, KitchenTicket } from '@/store/restaurantStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Clock, AlertTriangle, CheckCircle, Flame, Timer } from 'lucide-react'

export function KitchenView() {
  const { tickets, updateTicket } = useRestaurantStore()
  
  const pendingItems = tickets.flatMap(t => t.items).filter(i => i.status === 'pending').length
  const preparingItems = tickets.flatMap(t => t.items).filter(i => i.status === 'preparing').length
  const readyItems = tickets.flatMap(t => t.items).filter(i => i.status === 'ready').length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-serif text-2xl font-bold text-wine-900">Kitchen Display</h1>
          <p className="text-wood-500">Chef Marco's domain</p>
        </div>
        
        {/* Quick Stats */}
        <div className="flex gap-4">
          <StatBadge icon={Clock} label="Pending" value={pendingItems} color="wine" />
          <StatBadge icon={Flame} label="Cooking" value={preparingItems} color="gold" />
          <StatBadge icon={CheckCircle} label="Ready" value={readyItems} color="emerald" pulse />
        </div>
      </div>
      
      {/* Kitchen Stations */}
      <div className="grid grid-cols-4 gap-4">
        <StationCard name="Garde Manger" emoji="🥗" items={['Cold apps', 'Salads', 'Carpaccio']} />
        <StationCard name="Sauté" emoji="🍳" items={['Risotto', 'Pasta', 'Sauces']} />
        <StationCard name="Grill" emoji="🔥" items={['Fish', 'Meat', 'Vegetables']} />
        <StationCard name="Pastry" emoji="🍰" items={['Desserts', 'Bread', 'Pastries']} />
      </div>
      
      {/* Tickets */}
      <div className="grid grid-cols-3 gap-6">
        {/* Pending Queue */}
        <TicketColumn 
          title="🔴 FIRE" 
          tickets={tickets.filter(t => t.items.some(i => i.status === 'pending'))}
          color="wine"
        />
        
        {/* In Progress */}
        <TicketColumn 
          title="🟡 COOKING" 
          tickets={tickets.filter(t => t.items.some(i => i.status === 'preparing') && !t.items.some(i => i.status === 'pending'))}
          color="gold"
        />
        
        {/* Ready */}
        <TicketColumn 
          title="🟢 READY" 
          tickets={tickets.filter(t => t.items.every(i => i.status === 'ready' || i.status === 'served'))}
          color="emerald"
        />
      </div>
    </div>
  )
}

function StatBadge({ 
  icon: Icon, 
  label, 
  value, 
  color,
  pulse = false 
}: { 
  icon: typeof Clock
  label: string
  value: number
  color: 'wine' | 'gold' | 'emerald'
  pulse?: boolean
}) {
  const colors = {
    wine: 'bg-wine-100 text-wine-700 border-wine-200',
    gold: 'bg-gold-100 text-gold-700 border-gold-200',
    emerald: 'bg-emerald-100 text-emerald-700 border-emerald-200',
  }

  return (
    <div className={cn(
      "flex items-center gap-2 px-4 py-2 rounded-lg border",
      colors[color],
      pulse && value > 0 && "animate-pulse"
    )}>
      <Icon className="w-4 h-4" />
      <span className="font-medium">{value}</span>
      <span className="text-sm opacity-75">{label}</span>
    </div>
  )
}

function StationCard({ name, emoji, items }: { name: string; emoji: string; items: string[] }) {
  return (
    <div className="elegant-card p-4">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-2xl">{emoji}</span>
        <span className="font-semibold text-wine-900">{name}</span>
      </div>
      <div className="text-xs text-wood-500">
        {items.join(' • ')}
      </div>
    </div>
  )
}

function TicketColumn({ 
  title, 
  tickets, 
  color 
}: { 
  title: string
  tickets: KitchenTicket[]
  color: 'wine' | 'gold' | 'emerald'
}) {
  const borderColors = {
    wine: 'border-wine-300',
    gold: 'border-gold-300',
    emerald: 'border-emerald-300',
  }

  return (
    <div className={cn(
      "elegant-card p-4 border-t-4",
      borderColors[color]
    )}>
      <h3 className="font-mono font-bold text-lg mb-4">{title}</h3>
      
      <div className="space-y-4">
        {tickets.length === 0 ? (
          <div className="text-center py-8 text-wood-400">
            No tickets
          </div>
        ) : (
          tickets.map((ticket) => (
            <motion.div
              key={ticket.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className={cn(
                "kitchen-ticket",
                ticket.priority === 'vip' && "border-gold-400 bg-gold-50"
              )}
            >
              {/* Ticket Header */}
              <div className="flex items-center justify-between mb-3 pb-2 border-b border-dashed border-wood-300">
                <div className="flex items-center gap-2">
                  <span className="font-mono font-bold text-lg">
                    {ticket.tableId.replace('table-', 'T')}
                  </span>
                  {ticket.priority === 'vip' && (
                    <span className="px-2 py-0.5 bg-gold-500 text-white text-xs rounded font-bold">
                      VIP
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-1 text-xs text-wood-500">
                  <Timer className="w-3 h-3" />
                  {getTimeSince(ticket.createdAt)}
                </div>
              </div>
              
              {/* Items */}
              <div className="space-y-2">
                {ticket.items.map((item, idx) => (
                  <div 
                    key={idx}
                    className={cn(
                      "flex items-center justify-between py-1",
                      item.status === 'served' && "opacity-50 line-through"
                    )}
                  >
                    <div className="flex items-center gap-2">
                      <span className={cn(
                        "w-3 h-3 rounded-full",
                        item.status === 'pending' && "bg-wine-500",
                        item.status === 'preparing' && "bg-gold-500 animate-pulse",
                        item.status === 'ready' && "bg-emerald-500",
                        item.status === 'served' && "bg-gray-400",
                      )} />
                      <span className="font-medium">{item.quantity}x</span>
                      <span>{item.dish}</span>
                    </div>
                    <span className="text-xs text-wood-400 uppercase">
                      {item.station}
                    </span>
                  </div>
                ))}
              </div>
              
              {/* Special Requests */}
              {ticket.items.some(i => i.specialRequests) && (
                <div className="mt-3 pt-2 border-t border-dashed border-wood-300">
                  <div className="flex items-start gap-1 text-xs text-wine-600">
                    <AlertTriangle className="w-3 h-3 mt-0.5 flex-shrink-0" />
                    <span>
                      {ticket.items.find(i => i.specialRequests)?.specialRequests}
                    </span>
                  </div>
                </div>
              )}
            </motion.div>
          ))
        )}
      </div>
    </div>
  )
}
