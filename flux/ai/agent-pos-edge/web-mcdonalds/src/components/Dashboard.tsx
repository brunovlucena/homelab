'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { cn, formatCurrency, getTimeSince, getOrderTypeLabel, getOrderTypeColor } from '@/lib/utils'
import { motion } from 'framer-motion'
import {
  ChefHat,
  TrendingUp,
  DollarSign,
  Clock,
  Users,
  Utensils,
  Car,
  Bike,
  ShoppingBag,
  Activity,
  Timer,
  Camera,
  Brain,
  Bot,
  Eye,
  Sparkles
} from 'lucide-react'

export function Dashboard() {
  const { orders, stations, staff, agents, cameras, detections, setActiveView } = useMcdonaldsStore()

  const newOrders = orders.filter(o => o.status === 'new').length
  const preparingOrders = orders.filter(o => o.status === 'preparing').length
  const readyOrders = orders.filter(o => o.status === 'ready').length
  const completedOrders = orders.filter(o => o.status === 'delivered').length
  const totalRevenue = orders.filter(o => o.status === 'delivered').reduce((sum, o) => sum + o.total, 0)
  const avgPrepTime = Math.round(orders.filter(o => o.completedAt).reduce((sum, o) => {
    const start = new Date(o.startedAt!).getTime()
    const end = new Date(o.completedAt!).getTime()
    return sum + (end - start) / 1000
  }, 0) / (orders.filter(o => o.completedAt).length || 1))
  
  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'quality' || a.type === 'customer')
  const activeAgents = visionAgents.filter(a => a.status === 'online' || a.status === 'processing').length
  const onlineCameras = cameras.filter(c => c.status !== 'offline').length

  const orderTypeIcon = (type: string) => {
    switch (type) {
      case 'dine-in': return Utensils
      case 'drive-thru': return Car
      case 'delivery': return Bike
      default: return ShoppingBag
    }
  }

  const stats = [
    { label: 'Pedidos em Fila', value: newOrders + preparingOrders, icon: ChefHat, color: 'mc-red' },
    { label: 'Prontos p/ Entrega', value: readyOrders, icon: Activity, color: 'mc-green' },
    { label: 'AI Agents', value: `${activeAgents}/${visionAgents.length}`, icon: Brain, color: 'mc-gold' },
    { label: 'Tempo Médio', value: `${Math.floor(avgPrepTime / 60)}:${(avgPrepTime % 60).toString().padStart(2, '0')}`, icon: Timer, color: 'mc-blue' },
  ]

  return (
    <div className="p-6 space-y-6">
      {/* Hero Section */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-mc-red/20 via-mc-dark to-mc-gold/10 border border-mc-red/30 p-8"
      >
        <div className="absolute top-0 right-0 w-96 h-96 bg-mc-red/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-64 h-64 bg-mc-gold/10 rounded-full blur-3xl" />
        
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-xl golden-gradient flex items-center justify-center shadow-mc-gold">
              <span className="text-mc-red font-brand font-extrabold text-2xl">M</span>
            </div>
            <div>
              <h1 className="text-3xl font-brand font-bold text-white">
                Kitchen <span className="text-mc-gold">Command Center</span>
              </h1>
              <p className="text-gray-400">Gerenciamento de cozinha em tempo real</p>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-4 mt-6">
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-mc-dark/50 border border-mc-red/30">
              <ChefHat className="w-4 h-4 text-mc-red" />
              <span className="text-sm text-mc-red font-mono">{preparingOrders} em preparo</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-mc-dark/50 border border-mc-green/30">
              <Activity className="w-4 h-4 text-mc-green" />
              <span className="text-sm text-mc-green font-mono">{readyOrders} prontos</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-mc-dark/50 border border-mc-blue/30">
              <Camera className="w-4 h-4 text-mc-blue" />
              <span className="text-sm text-mc-blue font-mono">{onlineCameras} câmeras</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-mc-dark/50 border border-mc-gold/30">
              <Brain className="w-4 h-4 text-mc-gold animate-pulse" />
              <span className="text-sm text-mc-gold font-mono">{activeAgents} AI agents</span>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className={cn(
              'mc-card p-5 hover:border-mc-gold/50 transition-colors group',
              `border-${stat.color}/30`
            )}
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-gray-400 mb-1">{stat.label}</p>
                <p className="text-2xl font-bold text-white font-mono">{stat.value}</p>
              </div>
              <div className={cn(
                'p-3 rounded-xl transition-transform group-hover:scale-110',
                `bg-${stat.color}/10 border border-${stat.color}/30`
              )}>
                <stat.icon className={cn('w-6 h-6', `text-${stat.color}`)} />
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Orders */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
          className="lg:col-span-2 mc-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-brand font-bold text-white">Pedidos Ativos</h2>
            <button onClick={() => setActiveView('orders')} className="text-sm text-mc-gold hover:underline">
              Ver Todos →
            </button>
          </div>
          <div className="space-y-3">
            {orders.filter(o => o.status !== 'delivered' && o.status !== 'cancelled').slice(0, 5).map((order) => {
              const Icon = orderTypeIcon(order.type)
              const typeColor = getOrderTypeColor(order.type)
              
              return (
                <div 
                  key={order.id}
                  className={cn(
                    'order-card p-4',
                    order.status === 'new' ? 'order-new' : 
                    order.status === 'preparing' ? 'order-preparing' : 'order-ready'
                  )}
                >
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <span className="text-2xl font-brand font-bold text-white">#{order.orderNumber}</span>
                      <div className={cn('flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium', `bg-${typeColor}/20 text-${typeColor}`)}>
                        <Icon className="w-3 h-3" />
                        {getOrderTypeLabel(order.type)}
                      </div>
                      {order.priority === 'rush' && (
                        <span className="px-2 py-1 rounded-full bg-mc-red/20 text-mc-red text-xs font-bold animate-pulse">
                          RUSH
                        </span>
                      )}
                      {order.priority === 'vip' && (
                        <span className="px-2 py-1 rounded-full bg-mc-gold/20 text-mc-gold text-xs font-bold">
                          VIP
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <Clock className="w-4 h-4 text-gray-500" />
                      <span className="text-sm font-mono text-gray-400">{getTimeSince(order.createdAt)}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 flex-wrap">
                    {order.items.map((item, idx) => (
                      <span key={idx} className="px-2 py-1 rounded bg-mc-gray/50 text-sm text-gray-300">
                        {item.quantity}x {item.menuItem.name}
                      </span>
                    ))}
                  </div>
                  <div className="flex items-center justify-between mt-3 pt-3 border-t border-mc-gray/30">
                    <span className="text-lg font-bold text-mc-gold">{formatCurrency(order.total)}</span>
                    <span className={cn(
                      'px-3 py-1 rounded-full text-xs font-bold uppercase',
                      order.status === 'new' ? 'bg-status-new/20 text-status-new' :
                      order.status === 'preparing' ? 'bg-status-preparing/20 text-status-preparing' :
                      'bg-status-ready/20 text-status-ready'
                    )}>
                      {order.status === 'new' ? 'Novo' : order.status === 'preparing' ? 'Preparando' : 'Pronto'}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        </motion.div>

        {/* Kitchen Stations */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="mc-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-brand font-bold text-white">Estações</h2>
            <button onClick={() => setActiveView('kitchen')} className="text-sm text-mc-gold hover:underline">
              Ver Cozinha →
            </button>
          </div>
          <div className="space-y-3">
            {stations.map((station) => (
              <div 
                key={station.id}
                className={cn(
                  'p-3 rounded-lg border transition-colors',
                  station.status === 'active' ? 'bg-mc-green/10 border-mc-green/30' :
                  station.status === 'busy' ? 'bg-mc-orange/10 border-mc-orange/30' :
                  'bg-mc-red/10 border-mc-red/30'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-white">{station.name}</span>
                  <span className={cn(
                    'px-2 py-0.5 rounded-full text-xs font-medium',
                    station.status === 'active' ? 'bg-mc-green/20 text-mc-green' :
                    station.status === 'busy' ? 'bg-mc-orange/20 text-mc-orange' :
                    'bg-mc-red/20 text-mc-red'
                  )}>
                    {station.status === 'active' ? 'Ativo' : station.status === 'busy' ? 'Ocupado' : 'Offline'}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 bg-mc-gray/50 rounded-full overflow-hidden">
                    <div 
                      className={cn(
                        'h-full rounded-full transition-all',
                        station.status === 'active' ? 'bg-mc-green' :
                        station.status === 'busy' ? 'bg-mc-orange' : 'bg-mc-red'
                      )}
                      style={{ width: `${(station.activeOrders / station.capacity) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs text-gray-400 font-mono">{station.activeOrders}/{station.capacity}</span>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </div>
  )
}
