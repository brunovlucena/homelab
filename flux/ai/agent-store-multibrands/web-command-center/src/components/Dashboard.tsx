'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, getBrandBgColor, formatCurrency, formatPercent, formatTimeAgo } from '@/lib/utils'
import { BRANDS, BrandId } from '@/types/store'
import { motion } from 'framer-motion'
import {
  MessageSquare,
  ShoppingCart,
  DollarSign,
  TrendingUp,
  Clock,
  AlertTriangle,
  Users,
  Zap,
  Activity,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react'

const statCards = [
  {
    label: 'Mensagens Hoje',
    key: 'totalMessages',
    icon: MessageSquare,
    color: 'store-purple',
    format: (v: number) => v.toLocaleString('pt-BR'),
    change: '+23%',
    trend: 'up',
  },
  {
    label: 'Pedidos Hoje',
    key: 'totalOrders',
    icon: ShoppingCart,
    color: 'store-blue',
    format: (v: number) => v.toString(),
    change: '+12%',
    trend: 'up',
  },
  {
    label: 'Receita Hoje',
    key: 'totalRevenue',
    icon: DollarSign,
    color: 'store-green',
    format: (v: number) => formatCurrency(v),
    change: '+18%',
    trend: 'up',
  },
  {
    label: 'Tempo Resposta',
    key: 'avgResponseTime',
    icon: Clock,
    color: 'store-yellow',
    format: (v: number) => `${v.toFixed(1)}s`,
    change: '-5%',
    trend: 'up',
  },
]

export function Dashboard() {
  const { 
    metrics, 
    conversations, 
    orders, 
    events,
    sellers,
    setActiveView,
    selectedBrand,
  } = useStoreAppStore()
  
  const filteredConversations = selectedBrand === 'all' 
    ? conversations 
    : conversations.filter(c => c.brand === selectedBrand)
  
  const activeConversations = filteredConversations.filter(c => c.state === 'active' || c.state === 'waiting')
  const escalatedConversations = filteredConversations.filter(c => c.state === 'escalated')
  const recentOrders = orders.slice(0, 5)
  const recentEvents = events.slice(0, 6)

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Hero Section */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-store-purple/20 via-store-dark to-store-pink/10 border border-store-purple/30 p-8"
      >
        <div className="absolute top-0 right-0 w-96 h-96 bg-store-purple/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-64 h-64 bg-store-pink/10 rounded-full blur-3xl" />
        
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-4">
            <div className="relative">
              <span className="text-4xl">🏪</span>
              <div className="absolute inset-0 animate-ping">
                <span className="text-4xl opacity-30">🏪</span>
              </div>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-white">
                Bem-vindo ao <span className="text-gradient">Command Center</span>
              </h1>
              <p className="text-gray-400">
                Seus vendedores AI estão prontos para atender
              </p>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-4 mt-6">
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-store-dark/50 border border-store-green/30">
              <Users className="w-4 h-4 text-store-green" />
              <span className="text-sm text-store-green font-mono">
                {sellers.filter(s => s.status === 'online').length} vendedores online
              </span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-store-dark/50 border border-store-blue/30">
              <Activity className="w-4 h-4 text-store-blue" />
              <span className="text-sm text-store-blue font-mono">
                {activeConversations.length} conversas ativas
              </span>
            </div>
            {escalatedConversations.length > 0 && (
              <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-store-dark/50 border border-store-red/30">
                <AlertTriangle className="w-4 h-4 text-store-red" />
                <span className="text-sm text-store-red font-mono">
                  {escalatedConversations.length} escalações pendentes
                </span>
              </div>
            )}
          </div>
        </div>
      </motion.div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, index) => {
          const value = metrics[stat.key as keyof typeof metrics]
          const numValue = typeof value === 'number' ? value : 0
          
          return (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="store-card store-card-hover p-5"
            >
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-400 mb-1">{stat.label}</p>
                  <p className="text-3xl font-bold text-white">{stat.format(numValue)}</p>
                  <div className="flex items-center gap-1 mt-2">
                    {stat.trend === 'up' ? (
                      <ArrowUpRight className="w-3 h-3 text-store-green" />
                    ) : (
                      <ArrowDownRight className="w-3 h-3 text-store-red" />
                    )}
                    <span className={cn(
                      "text-xs font-mono",
                      stat.trend === 'up' ? 'text-store-green' : 'text-store-red'
                    )}>
                      {stat.change}
                    </span>
                  </div>
                </div>
                <div className={cn(
                  "p-3 rounded-xl",
                  `bg-${stat.color}/10 border border-${stat.color}/30`
                )}>
                  <stat.icon className={cn("w-6 h-6", `text-${stat.color}`)} />
                </div>
              </div>
            </motion.div>
          )
        })}
      </div>

      {/* Brand Performance */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="store-card p-5"
      >
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-white">Performance por Marca</h2>
          <button
            onClick={() => setActiveView('analytics')}
            className="text-sm text-store-purple hover:text-store-pink transition-colors"
          >
            Ver Detalhes →
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          {metrics.brandMetrics.map((brandMetric) => {
            const brand = BRANDS[brandMetric.brand]
            return (
              <div
                key={brandMetric.brand}
                className={cn(
                  "p-4 rounded-xl bg-store-dark/50 border transition-all hover:scale-105 cursor-pointer",
                  `border-${brand.color}/30 hover:border-${brand.color}/60`
                )}
              >
                <div className="flex items-center gap-2 mb-3">
                  <span className="text-2xl">{brand.emoji}</span>
                  <span className={cn("font-bold", getBrandColor(brandMetric.brand))}>
                    {brand.name}
                  </span>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Mensagens</span>
                    <span className="text-white font-mono">{brandMetric.messages24h}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Pedidos</span>
                    <span className="text-white font-mono">{brandMetric.orders24h}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Receita</span>
                    <span className="text-store-green font-mono">
                      {formatCurrency(brandMetric.revenue24h)}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Conversão</span>
                    <span className="text-store-blue font-mono">
                      {formatPercent(brandMetric.conversionRate)}
                    </span>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Conversations */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="lg:col-span-2 space-y-4"
        >
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-bold text-white">Conversas Ativas</h2>
            <button
              onClick={() => setActiveView('conversations')}
              className="text-sm text-store-purple hover:text-store-pink transition-colors"
            >
              Ver Todas →
            </button>
          </div>
          <div className="store-card divide-y divide-store-purple/10">
            {activeConversations.slice(0, 5).map((conv) => {
              const brand = BRANDS[conv.brand]
              return (
                <div
                  key={conv.id}
                  className="p-4 hover:bg-store-purple/5 transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{brand.sellerAvatar}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-white">
                          {conv.customerName || conv.customerPhone}
                        </span>
                        <span className={cn(
                          "px-2 py-0.5 text-xs rounded-full",
                          conv.state === 'active' 
                            ? 'bg-store-green/10 text-store-green'
                            : 'bg-store-yellow/10 text-store-yellow'
                        )}>
                          {conv.state === 'active' ? 'Ativo' : 'Aguardando'}
                        </span>
                      </div>
                      <p className="text-sm text-gray-500 truncate">
                        {conv.messages[conv.messages.length - 1]?.content}
                      </p>
                    </div>
                    <div className="text-right">
                      <span className={cn("text-xs", getBrandColor(conv.brand))}>
                        {brand.name}
                      </span>
                      <p className="text-xs text-gray-600">
                        {formatTimeAgo(conv.lastMessageAt)}
                      </p>
                    </div>
                  </div>
                </div>
              )
            })}
            {activeConversations.length === 0 && (
              <div className="p-8 text-center">
                <MessageSquare className="w-10 h-10 text-gray-600 mx-auto mb-3" />
                <p className="text-sm text-gray-500">Nenhuma conversa ativa no momento</p>
              </div>
            )}
          </div>
        </motion.div>

        {/* Recent Orders & Events */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.4 }}
          className="space-y-6"
        >
          {/* Recent Orders */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-bold text-white">Pedidos Recentes</h2>
              <button
                onClick={() => setActiveView('orders')}
                className="text-sm text-store-purple hover:text-store-pink transition-colors"
              >
                Ver Todos →
              </button>
            </div>
            <div className="store-card divide-y divide-store-purple/10">
              {recentOrders.map((order) => {
                const brand = BRANDS[order.brand]
                return (
                  <div key={order.id} className="p-3 hover:bg-store-purple/5 transition-colors">
                    <div className="flex items-center gap-3">
                      <span className="text-xl">{brand.emoji}</span>
                      <div className="flex-1">
                        <p className="text-sm font-mono text-white">{order.id}</p>
                        <p className="text-xs text-gray-500">{order.items.length} item(s)</p>
                      </div>
                      <div className="text-right">
                        <p className="text-sm font-bold text-store-green">
                          {formatCurrency(order.total)}
                        </p>
                        <p className="text-xs text-gray-600">
                          {formatTimeAgo(order.createdAt)}
                        </p>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Live Events */}
          <div>
            <h2 className="text-lg font-bold text-white mb-4">Eventos em Tempo Real</h2>
            <div className="store-card p-4 space-y-2">
              {recentEvents.map((event) => (
                <div
                  key={event.id}
                  className="flex items-center gap-2 p-2 rounded-lg bg-store-dark/50 animate-slide-in"
                >
                  <Zap className="w-4 h-4 text-store-purple" />
                  <span className="text-xs font-mono text-gray-400 truncate flex-1">
                    {event.type.split('.').slice(-2).join('.')}
                  </span>
                  {event.brand && (
                    <span className="text-sm">{BRANDS[event.brand].emoji}</span>
                  )}
                </div>
              ))}
              {recentEvents.length === 0 && (
                <div className="text-center py-4">
                  <p className="text-sm text-gray-500">Aguardando eventos...</p>
                </div>
              )}
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}
