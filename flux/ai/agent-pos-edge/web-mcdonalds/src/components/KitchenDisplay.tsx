'use client'

import { useMcdonaldsStore, Order } from '@/store/mcdonaldsStore'
import { cn, getTimeSince, getOrderTypeLabel, getOrderTypeColor, getElapsedSeconds, formatTime } from '@/lib/utils'
import { motion, AnimatePresence } from 'framer-motion'
import { Clock, Utensils, Car, Bike, ShoppingBag, ChefHat, CheckCircle, AlertTriangle } from 'lucide-react'
import { useState, useEffect } from 'react'

function OrderTimer({ startTime, estimatedTime }: { startTime: string, estimatedTime: number }) {
  const [elapsed, setElapsed] = useState(getElapsedSeconds(startTime))
  
  useEffect(() => {
    const interval = setInterval(() => {
      setElapsed(getElapsedSeconds(startTime))
    }, 1000)
    return () => clearInterval(interval)
  }, [startTime])
  
  const remaining = estimatedTime - elapsed
  const isLate = remaining < 0
  const isUrgent = remaining < 60 && remaining >= 0
  
  return (
    <div className={cn(
      'flex items-center gap-2 px-3 py-1.5 rounded-lg font-mono text-lg',
      isLate ? 'bg-mc-red/20 text-mc-red timer-blink' :
      isUrgent ? 'bg-mc-orange/20 text-mc-orange' :
      'bg-mc-green/20 text-mc-green'
    )}>
      <Clock className="w-5 h-5" />
      {isLate ? (
        <span>+{formatTime(Math.abs(remaining))}</span>
      ) : (
        <span>{formatTime(remaining)}</span>
      )}
    </div>
  )
}

function KitchenOrderCard({ order, onStatusChange }: { order: Order, onStatusChange: (status: Order['status']) => void }) {
  const orderTypeIcon = (type: string) => {
    switch (type) {
      case 'dine-in': return Utensils
      case 'drive-thru': return Car
      case 'delivery': return Bike
      default: return ShoppingBag
    }
  }
  
  const Icon = orderTypeIcon(order.type)
  const typeColor = getOrderTypeColor(order.type)

  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.9, y: 20 }}
      animate={{ opacity: 1, scale: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.9, y: -20 }}
      className={cn(
        'order-card p-4',
        order.status === 'new' ? 'order-new' : 
        order.status === 'preparing' ? 'order-preparing' : 'order-ready'
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-3xl font-brand font-extrabold text-white">#{order.orderNumber}</span>
          {order.priority === 'rush' && (
            <span className="px-2 py-1 rounded bg-mc-red text-white text-xs font-bold animate-pulse flex items-center gap-1">
              <AlertTriangle className="w-3 h-3" />
              RUSH
            </span>
          )}
          {order.priority === 'vip' && (
            <span className="px-2 py-1 rounded golden-gradient text-mc-black text-xs font-bold">VIP</span>
          )}
        </div>
        <div className={cn('flex items-center gap-1 px-3 py-1.5 rounded-lg', `bg-${typeColor}/20`)}>
          <Icon className={cn('w-4 h-4', `text-${typeColor}`)} />
          <span className={cn('text-sm font-medium', `text-${typeColor}`)}>{getOrderTypeLabel(order.type)}</span>
        </div>
      </div>

      {/* Timer */}
      {order.startedAt && (
        <div className="mb-3">
          <OrderTimer startTime={order.startedAt} estimatedTime={order.estimatedTime} />
        </div>
      )}

      {/* Items */}
      <div className="space-y-2 mb-4">
        {order.items.map((item, idx) => (
          <div key={idx} className="flex items-center justify-between p-2 rounded-lg bg-mc-gray/30">
            <div className="flex items-center gap-2">
              <span className="quantity-badge">{item.quantity}</span>
              <span className="text-white font-medium">{item.menuItem.name}</span>
            </div>
            {item.customizations && item.customizations.length > 0 && (
              <span className="text-xs text-mc-orange">+{item.customizations.length} mod</span>
            )}
          </div>
        ))}
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        {order.status === 'new' && (
          <button
            onClick={() => onStatusChange('preparing')}
            className="flex-1 mc-button flex items-center justify-center gap-2"
          >
            <ChefHat className="w-4 h-4" />
            Iniciar Preparo
          </button>
        )}
        {order.status === 'preparing' && (
          <button
            onClick={() => onStatusChange('ready')}
            className="flex-1 mc-button-gold flex items-center justify-center gap-2"
          >
            <CheckCircle className="w-4 h-4" />
            Marcar Pronto
          </button>
        )}
        {order.status === 'ready' && (
          <button
            onClick={() => onStatusChange('delivered')}
            className="flex-1 mc-button flex items-center justify-center gap-2"
          >
            <CheckCircle className="w-4 h-4" />
            Entregar
          </button>
        )}
      </div>
    </motion.div>
  )
}

export function KitchenDisplay() {
  const { orders, updateOrderStatus } = useMcdonaldsStore()

  const newOrders = orders.filter(o => o.status === 'new')
  const preparingOrders = orders.filter(o => o.status === 'preparing')
  const readyOrders = orders.filter(o => o.status === 'ready')

  return (
    <div className="p-6 h-full">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-brand font-bold text-white">Kitchen Display System</h1>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-status-new/20">
            <span className="status-dot status-dot-new" />
            <span className="text-sm text-status-new font-mono">{newOrders.length} novos</span>
          </div>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-status-preparing/20">
            <span className="status-dot status-dot-preparing" />
            <span className="text-sm text-status-preparing font-mono">{preparingOrders.length} preparando</span>
          </div>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-status-ready/20">
            <span className="status-dot status-dot-ready" />
            <span className="text-sm text-status-ready font-mono">{readyOrders.length} prontos</span>
          </div>
        </div>
      </div>

      {/* 3-Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[calc(100%-80px)]">
        {/* New Orders Column */}
        <div className="flex flex-col">
          <div className="flex items-center gap-2 mb-4 p-3 rounded-lg bg-status-new/10 border border-status-new/30">
            <div className="w-3 h-3 rounded-full bg-status-new animate-pulse" />
            <h2 className="text-lg font-bold text-status-new">NOVOS</h2>
            <span className="ml-auto text-2xl font-brand font-bold text-status-new">{newOrders.length}</span>
          </div>
          <div className="flex-1 space-y-4 overflow-y-auto pr-2">
            <AnimatePresence mode="popLayout">
              {newOrders.map(order => (
                <KitchenOrderCard 
                  key={order.id} 
                  order={order} 
                  onStatusChange={(status) => updateOrderStatus(order.id, status)}
                />
              ))}
            </AnimatePresence>
            {newOrders.length === 0 && (
              <div className="text-center py-12 text-gray-500">
                <ChefHat className="w-12 h-12 mx-auto mb-3 opacity-30" />
                <p>Nenhum pedido novo</p>
              </div>
            )}
          </div>
        </div>

        {/* Preparing Column */}
        <div className="flex flex-col">
          <div className="flex items-center gap-2 mb-4 p-3 rounded-lg bg-status-preparing/10 border border-status-preparing/30">
            <div className="w-3 h-3 rounded-full bg-status-preparing animate-pulse" />
            <h2 className="text-lg font-bold text-status-preparing">PREPARANDO</h2>
            <span className="ml-auto text-2xl font-brand font-bold text-status-preparing">{preparingOrders.length}</span>
          </div>
          <div className="flex-1 space-y-4 overflow-y-auto pr-2">
            <AnimatePresence mode="popLayout">
              {preparingOrders.map(order => (
                <KitchenOrderCard 
                  key={order.id} 
                  order={order} 
                  onStatusChange={(status) => updateOrderStatus(order.id, status)}
                />
              ))}
            </AnimatePresence>
            {preparingOrders.length === 0 && (
              <div className="text-center py-12 text-gray-500">
                <Clock className="w-12 h-12 mx-auto mb-3 opacity-30" />
                <p>Nenhum pedido em preparo</p>
              </div>
            )}
          </div>
        </div>

        {/* Ready Column */}
        <div className="flex flex-col">
          <div className="flex items-center gap-2 mb-4 p-3 rounded-lg bg-status-ready/10 border border-status-ready/30">
            <div className="w-3 h-3 rounded-full bg-status-ready animate-pulse" />
            <h2 className="text-lg font-bold text-status-ready">PRONTOS</h2>
            <span className="ml-auto text-2xl font-brand font-bold text-status-ready">{readyOrders.length}</span>
          </div>
          <div className="flex-1 space-y-4 overflow-y-auto pr-2">
            <AnimatePresence mode="popLayout">
              {readyOrders.map(order => (
                <KitchenOrderCard 
                  key={order.id} 
                  order={order} 
                  onStatusChange={(status) => updateOrderStatus(order.id, status)}
                />
              ))}
            </AnimatePresence>
            {readyOrders.length === 0 && (
              <div className="text-center py-12 text-gray-500">
                <CheckCircle className="w-12 h-12 mx-auto mb-3 opacity-30" />
                <p>Nenhum pedido pronto</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
