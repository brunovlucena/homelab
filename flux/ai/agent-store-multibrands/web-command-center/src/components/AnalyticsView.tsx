'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, formatCurrency, formatPercent } from '@/lib/utils'
import { BRANDS } from '@/types/store'
import { motion } from 'framer-motion'
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Users,
  MessageSquare,
  ShoppingCart,
  Clock,
  Target,
} from 'lucide-react'

// Simple bar chart component
function BarChart({ data, maxValue, color }: { 
  data: { label: string; value: number; color: string }[]
  maxValue: number
  color?: string 
}) {
  return (
    <div className="space-y-3">
      {data.map((item, i) => (
        <div key={i} className="space-y-1">
          <div className="flex justify-between text-sm">
            <span className={cn("text-gray-400", item.color)}>{item.label}</span>
            <span className="text-white font-mono">{item.value}</span>
          </div>
          <div className="h-2 bg-store-gray rounded-full overflow-hidden">
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${(item.value / maxValue) * 100}%` }}
              transition={{ duration: 0.8, delay: i * 0.1 }}
              className={cn("h-full rounded-full", item.color.replace('text-', 'bg-'))}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

export function AnalyticsView() {
  const { metrics, orders, conversations } = useStoreAppStore()
  
  const brandData = metrics.brandMetrics.map(bm => ({
    label: BRANDS[bm.brand].name,
    value: bm.messages24h,
    color: getBrandColor(bm.brand),
  }))
  
  const maxMessages = Math.max(...brandData.map(d => d.value))
  
  const revenueData = metrics.brandMetrics.map(bm => ({
    label: BRANDS[bm.brand].name,
    value: bm.revenue24h,
    color: getBrandColor(bm.brand),
  }))
  
  const maxRevenue = Math.max(...revenueData.map(d => d.value))

  // Calculate hourly distribution (mock data)
  const hourlyData = Array.from({ length: 24 }, (_, i) => ({
    hour: i,
    messages: Math.floor(Math.random() * 50 + 10),
    orders: Math.floor(Math.random() * 10),
  }))

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Analytics</h1>
          <p className="text-gray-500">Métricas e performance da loja</p>
        </div>
        <div className="flex gap-2">
          {['Hoje', '7 dias', '30 dias'].map((period) => (
            <button
              key={period}
              className={cn(
                "px-4 py-2 text-sm rounded-lg transition-all",
                period === 'Hoje'
                  ? "bg-store-purple/20 text-store-purple border border-store-purple/50"
                  : "bg-store-gray/30 text-gray-400 hover:text-white"
              )}
            >
              {period}
            </button>
          ))}
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { 
            label: 'Receita Total', 
            value: formatCurrency(metrics.totalRevenue), 
            change: '+18%', 
            trend: 'up',
            icon: DollarSign,
            color: 'store-green'
          },
          { 
            label: 'Taxa de Conversão', 
            value: '12.5%', 
            change: '+2.3%', 
            trend: 'up',
            icon: Target,
            color: 'store-blue'
          },
          { 
            label: 'Tempo Médio Resposta', 
            value: `${metrics.avgResponseTime.toFixed(1)}s`, 
            change: '-0.3s', 
            trend: 'up',
            icon: Clock,
            color: 'store-purple'
          },
          { 
            label: 'Taxa de Escalação', 
            value: '4.2%', 
            change: '-1.1%', 
            trend: 'up',
            icon: Users,
            color: 'store-yellow'
          },
        ].map((kpi, i) => (
          <motion.div
            key={kpi.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="store-card p-5"
          >
            <div className="flex items-start justify-between mb-3">
              <div className={cn("p-2 rounded-lg", `bg-${kpi.color}/10`)}>
                <kpi.icon className={cn("w-5 h-5", `text-${kpi.color}`)} />
              </div>
              <div className={cn(
                "flex items-center gap-1 text-xs font-mono",
                kpi.trend === 'up' ? 'text-store-green' : 'text-store-red'
              )}>
                {kpi.trend === 'up' ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                {kpi.change}
              </div>
            </div>
            <p className="text-2xl font-bold text-white">{kpi.value}</p>
            <p className="text-sm text-gray-500">{kpi.label}</p>
          </motion.div>
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-2 gap-6">
        {/* Messages by Brand */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="store-card p-5"
        >
          <div className="flex items-center gap-2 mb-4">
            <MessageSquare className="w-5 h-5 text-store-purple" />
            <h3 className="text-lg font-bold text-white">Mensagens por Marca</h3>
          </div>
          <BarChart data={brandData} maxValue={maxMessages} />
        </motion.div>

        {/* Revenue by Brand */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="store-card p-5"
        >
          <div className="flex items-center gap-2 mb-4">
            <DollarSign className="w-5 h-5 text-store-green" />
            <h3 className="text-lg font-bold text-white">Receita por Marca</h3>
          </div>
          <div className="space-y-3">
            {revenueData.map((item, i) => (
              <div key={i} className="space-y-1">
                <div className="flex justify-between text-sm">
                  <span className={cn("text-gray-400", item.color)}>{item.label}</span>
                  <span className="text-store-green font-mono">{formatCurrency(item.value)}</span>
                </div>
                <div className="h-2 bg-store-gray rounded-full overflow-hidden">
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${(item.value / maxRevenue) * 100}%` }}
                    transition={{ duration: 0.8, delay: i * 0.1 }}
                    className="h-full rounded-full bg-store-green"
                  />
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>

      {/* Hourly Activity */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="store-card p-5"
      >
        <div className="flex items-center gap-2 mb-4">
          <BarChart3 className="w-5 h-5 text-store-blue" />
          <h3 className="text-lg font-bold text-white">Atividade por Hora</h3>
        </div>
        <div className="flex items-end gap-1 h-40">
          {hourlyData.map((hour, i) => {
            const maxMsg = Math.max(...hourlyData.map(h => h.messages))
            const height = (hour.messages / maxMsg) * 100
            
            return (
              <motion.div
                key={hour.hour}
                initial={{ height: 0 }}
                animate={{ height: `${height}%` }}
                transition={{ duration: 0.5, delay: i * 0.02 }}
                className="flex-1 bg-store-purple/30 hover:bg-store-purple/50 rounded-t transition-colors cursor-pointer group relative"
              >
                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-store-dark rounded text-xs text-white opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                  {hour.hour}h: {hour.messages} msgs
                </div>
              </motion.div>
            )
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-600">
          <span>0h</span>
          <span>6h</span>
          <span>12h</span>
          <span>18h</span>
          <span>23h</span>
        </div>
      </motion.div>

      {/* Brand Performance Table */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="store-card overflow-hidden"
      >
        <div className="p-5 border-b border-store-purple/20">
          <h3 className="text-lg font-bold text-white">Performance Detalhada por Marca</h3>
        </div>
        <table className="w-full">
          <thead>
            <tr className="border-b border-store-purple/10">
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Marca</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Mensagens</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Pedidos</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Receita</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Conversão</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Tempo Resp.</th>
              <th className="px-5 py-3 text-left text-xs font-bold text-gray-500 uppercase">Satisfação</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-store-purple/10">
            {metrics.brandMetrics.map((bm) => {
              const brand = BRANDS[bm.brand]
              return (
                <tr key={bm.brand} className="hover:bg-store-purple/5">
                  <td className="px-5 py-4">
                    <div className="flex items-center gap-2">
                      <span className="text-xl">{brand.emoji}</span>
                      <span className={cn("font-medium", getBrandColor(bm.brand))}>
                        {brand.name}
                      </span>
                    </div>
                  </td>
                  <td className="px-5 py-4 font-mono text-white">{bm.messages24h}</td>
                  <td className="px-5 py-4 font-mono text-white">{bm.orders24h}</td>
                  <td className="px-5 py-4 font-mono text-store-green">{formatCurrency(bm.revenue24h)}</td>
                  <td className="px-5 py-4 font-mono text-store-blue">{formatPercent(bm.conversionRate)}</td>
                  <td className="px-5 py-4 font-mono text-white">{bm.avgResponseTime.toFixed(1)}s</td>
                  <td className="px-5 py-4">
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 bg-store-gray rounded-full overflow-hidden">
                        <div 
                          className="h-full bg-store-green rounded-full"
                          style={{ width: `${bm.satisfaction * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-mono text-white">{formatPercent(bm.satisfaction)}</span>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </motion.div>
    </div>
  )
}
