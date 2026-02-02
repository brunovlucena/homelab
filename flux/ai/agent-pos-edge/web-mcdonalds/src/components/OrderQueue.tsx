'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { cn, formatCurrency, getTimeSince, getOrderTypeLabel, getOrderTypeColor } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Search, Filter, Utensils, Car, Bike, ShoppingBag, Clock } from 'lucide-react'
import { useState } from 'react'

export function OrderQueue() {
  const { orders, updateOrderStatus } = useMcdonaldsStore()
  const [filter, setFilter] = useState<'all' | 'new' | 'preparing' | 'ready'>('all')
  const [search, setSearch] = useState('')

  const filteredOrders = orders.filter(o => {
    if (filter !== 'all' && o.status !== filter) return false
    if (search && !o.orderNumber.toString().includes(search)) return false
    return true
  })

  const orderTypeIcon = (type: string) => {
    switch (type) {
      case 'dine-in': return Utensils
      case 'drive-thru': return Car
      case 'delivery': return Bike
      default: return ShoppingBag
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">Fila de Pedidos</h1>
          <p className="text-gray-400">Todos os pedidos do turno</p>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="mc-card p-4">
          <p className="text-sm text-gray-400">Total</p>
          <p className="text-2xl font-bold font-mono text-white">{orders.length}</p>
        </div>
        <div className="mc-card p-4 border-status-new/30">
          <p className="text-sm text-gray-400">Novos</p>
          <p className="text-2xl font-bold font-mono text-status-new">{orders.filter(o => o.status === 'new').length}</p>
        </div>
        <div className="mc-card p-4 border-status-preparing/30">
          <p className="text-sm text-gray-400">Preparando</p>
          <p className="text-2xl font-bold font-mono text-status-preparing">{orders.filter(o => o.status === 'preparing').length}</p>
        </div>
        <div className="mc-card p-4 border-status-ready/30">
          <p className="text-sm text-gray-400">Prontos</p>
          <p className="text-2xl font-bold font-mono text-status-ready">{orders.filter(o => o.status === 'ready').length}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col md:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Buscar por número do pedido..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-mc-gray/50 border border-mc-gray rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-mc-gold/50"
          />
        </div>
        <div className="flex gap-2">
          {[
            { id: 'all', label: 'Todos' },
            { id: 'new', label: 'Novos', color: 'status-new' },
            { id: 'preparing', label: 'Preparando', color: 'status-preparing' },
            { id: 'ready', label: 'Prontos', color: 'status-ready' },
          ].map((f) => (
            <button
              key={f.id}
              onClick={() => setFilter(f.id as typeof filter)}
              className={cn(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === f.id 
                  ? f.color ? `bg-${f.color}/20 text-${f.color} border border-${f.color}/30` : 'bg-mc-gold/20 text-mc-gold border border-mc-gold/30'
                  : 'bg-mc-gray/50 text-gray-400 hover:bg-mc-gray'
              )}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {/* Orders Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredOrders.map((order, index) => {
          const Icon = orderTypeIcon(order.type)
          const typeColor = getOrderTypeColor(order.type)
          
          return (
            <motion.div
              key={order.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className={cn(
                'order-card p-4',
                order.status === 'new' ? 'order-new' : 
                order.status === 'preparing' ? 'order-preparing' : 'order-ready'
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <span className="text-2xl font-brand font-bold text-white">#{order.orderNumber}</span>
                <div className={cn('flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium', `bg-${typeColor}/20 text-${typeColor}`)}>
                  <Icon className="w-3 h-3" />
                  {getOrderTypeLabel(order.type)}
                </div>
              </div>

              <div className="space-y-1 mb-3">
                {order.items.slice(0, 3).map((item, idx) => (
                  <p key={idx} className="text-sm text-gray-300">
                    {item.quantity}x {item.menuItem.name}
                  </p>
                ))}
                {order.items.length > 3 && (
                  <p className="text-sm text-gray-500">+{order.items.length - 3} mais itens</p>
                )}
              </div>

              <div className="flex items-center justify-between pt-3 border-t border-mc-gray/30">
                <div className="flex items-center gap-2 text-gray-500 text-sm">
                  <Clock className="w-4 h-4" />
                  {getTimeSince(order.createdAt)}
                </div>
                <span className="text-lg font-bold text-mc-gold">{formatCurrency(order.total)}</span>
              </div>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
