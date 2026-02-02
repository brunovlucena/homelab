'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { cn, formatCurrency } from '@/lib/utils'
import { motion } from 'framer-motion'
import { TrendingUp, DollarSign, Clock, Users, BarChart3, PieChart, Activity } from 'lucide-react'

export function Analytics() {
  const { orders, staff, stations } = useMcdonaldsStore()

  const completedOrders = orders.filter(o => o.status === 'delivered')
  const totalRevenue = completedOrders.reduce((sum, o) => sum + o.total, 0)
  const avgTicket = totalRevenue / (completedOrders.length || 1)
  const totalItems = orders.reduce((sum, o) => sum + o.items.reduce((s, i) => s + i.quantity, 0), 0)

  // Calculate avg prep time
  const ordersWithTime = orders.filter(o => o.startedAt && o.completedAt)
  const avgPrepTime = ordersWithTime.length > 0 
    ? Math.round(ordersWithTime.reduce((sum, o) => {
        const start = new Date(o.startedAt!).getTime()
        const end = new Date(o.completedAt!).getTime()
        return sum + (end - start) / 1000
      }, 0) / ordersWithTime.length)
    : 0

  // Order type distribution
  const orderTypes = {
    'dine-in': orders.filter(o => o.type === 'dine-in').length,
    'drive-thru': orders.filter(o => o.type === 'drive-thru').length,
    'delivery': orders.filter(o => o.type === 'delivery').length,
    'takeaway': orders.filter(o => o.type === 'takeaway').length,
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">Relatórios e Analytics</h1>
          <p className="text-gray-400">Métricas de desempenho em tempo real</p>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mc-card p-5"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Vendas Hoje</p>
              <p className="text-2xl font-bold text-mc-gold">{formatCurrency(totalRevenue)}</p>
              <p className="text-xs text-mc-green mt-1">+12% vs ontem</p>
            </div>
            <div className="p-3 rounded-xl bg-mc-gold/10 border border-mc-gold/30">
              <DollarSign className="w-6 h-6 text-mc-gold" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="mc-card p-5"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Pedidos</p>
              <p className="text-2xl font-bold text-white">{orders.length}</p>
              <p className="text-xs text-mc-green mt-1">+8% vs ontem</p>
            </div>
            <div className="p-3 rounded-xl bg-mc-red/10 border border-mc-red/30">
              <BarChart3 className="w-6 h-6 text-mc-red" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="mc-card p-5"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Ticket Médio</p>
              <p className="text-2xl font-bold text-white">{formatCurrency(avgTicket)}</p>
              <p className="text-xs text-mc-green mt-1">+5% vs ontem</p>
            </div>
            <div className="p-3 rounded-xl bg-mc-blue/10 border border-mc-blue/30">
              <TrendingUp className="w-6 h-6 text-mc-blue" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="mc-card p-5"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Tempo Médio</p>
              <p className="text-2xl font-bold text-white">{Math.floor(avgPrepTime / 60)}:{(avgPrepTime % 60).toString().padStart(2, '0')}</p>
              <p className="text-xs text-mc-orange mt-1">Meta: 3:00</p>
            </div>
            <div className="p-3 rounded-xl bg-mc-green/10 border border-mc-green/30">
              <Clock className="w-6 h-6 text-mc-green" />
            </div>
          </div>
        </motion.div>
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Order Types */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.4 }}
          className="mc-card p-5"
        >
          <h2 className="text-lg font-brand font-bold text-white mb-4">Distribuição por Canal</h2>
          <div className="space-y-4">
            {Object.entries(orderTypes).map(([type, count]) => {
              const total = Object.values(orderTypes).reduce((a, b) => a + b, 0)
              const percent = total > 0 ? (count / total) * 100 : 0
              const labels: Record<string, string> = {
                'dine-in': 'Salão',
                'drive-thru': 'Drive-Thru',
                'delivery': 'Delivery',
                'takeaway': 'Viagem'
              }
              const colors: Record<string, string> = {
                'dine-in': 'mc-blue',
                'drive-thru': 'mc-gold',
                'delivery': 'mc-orange',
                'takeaway': 'mc-green'
              }
              
              return (
                <div key={type}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">{labels[type]}</span>
                    <span className="text-sm font-mono text-white">{count} ({percent.toFixed(0)}%)</span>
                  </div>
                  <div className="h-2 bg-mc-gray/50 rounded-full overflow-hidden">
                    <div 
                      className={cn('h-full rounded-full transition-all', `bg-${colors[type]}`)}
                      style={{ width: `${percent}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </motion.div>

        {/* Station Performance */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.5 }}
          className="mc-card p-5"
        >
          <h2 className="text-lg font-brand font-bold text-white mb-4">Desempenho das Estações</h2>
          <div className="space-y-4">
            {stations.map((station) => {
              const utilization = (station.activeOrders / station.capacity) * 100
              
              return (
                <div key={station.id}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">{station.name}</span>
                    <span className={cn(
                      'text-sm font-mono',
                      station.status === 'active' ? 'text-mc-green' :
                      station.status === 'busy' ? 'text-mc-orange' : 'text-mc-red'
                    )}>
                      {station.activeOrders}/{station.capacity} ({utilization.toFixed(0)}%)
                    </span>
                  </div>
                  <div className="h-2 bg-mc-gray/50 rounded-full overflow-hidden">
                    <div 
                      className={cn(
                        'h-full rounded-full transition-all',
                        station.status === 'active' ? 'bg-mc-green' :
                        station.status === 'busy' ? 'bg-mc-orange' : 'bg-mc-red'
                      )}
                      style={{ width: `${utilization}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </motion.div>
      </div>

      {/* Staff Performance */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="mc-card p-5"
      >
        <h2 className="text-lg font-brand font-bold text-white mb-4">Performance da Equipe</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4">
          {staff.map((member) => (
            <div key={member.id} className="text-center p-3 rounded-lg bg-mc-gray/30">
              <div className={cn(
                'w-10 h-10 mx-auto rounded-full flex items-center justify-center text-sm font-bold mb-2',
                member.status === 'active' ? 'bg-mc-green/20 text-mc-green' :
                member.status === 'break' ? 'bg-mc-orange/20 text-mc-orange' : 'bg-mc-red/20 text-mc-red'
              )}>
                {member.name.split(' ').map(n => n[0]).join('')}
              </div>
              <p className="text-xs text-gray-400 truncate">{member.name.split(' ')[0]}</p>
              <p className="text-lg font-bold text-mc-gold">{member.ordersCompleted}</p>
              <p className="text-xs text-gray-500">pedidos</p>
            </div>
          ))}
        </div>
      </motion.div>
    </div>
  )
}
