'use client'

import { useEffect } from 'react'
import { useRestaurantStore } from '@/store/restaurantStore'
import { cn, formatCurrency, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import {
  Users,
  DollarSign,
  Clock,
  Star,
  ChefHat,
  Wine,
  UtensilsCrossed,
  TrendingUp,
  AlertCircle,
  AlertTriangle,
  RefreshCw,
  Database,
  Wifi,
  WifiOff,
} from 'lucide-react'

export function Dashboard() {
  const { 
    stats, tables, tickets, agents, setActiveView,
    dataSource, errorMessage, isLoading, lastFetched, fetchLiveData 
  } = useRestaurantStore()
  
  // Try to fetch live data on mount
  useEffect(() => {
    fetchLiveData()
  }, [fetchLiveData])
  
  const occupiedTables = tables.filter(t => t.status === 'occupied').length
  const pendingOrders = tickets.filter(t => 
    t.items.some(i => i.status === 'pending' || i.status === 'preparing')
  ).length
  const readyDishes = tickets.flatMap(t => t.items).filter(i => i.status === 'ready').length
  const onlineAgents = agents.filter(a => a.status === 'active' || a.status === 'busy').length

  return (
    <div className="space-y-6">
      {/* ⚠️ Data Source Warning Banner */}
      {dataSource !== 'live' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className={cn(
            "flex items-center justify-between p-4 rounded-lg border",
            dataSource === 'mock' 
              ? "bg-amber-50 border-amber-300 text-amber-900"
              : "bg-red-50 border-red-300 text-red-900"
          )}
        >
          <div className="flex items-center gap-3">
            <AlertTriangle className="w-5 h-5" />
            <div>
              <p className="font-semibold">
                {dataSource === 'mock' ? '⚠️ Using MOCK Data' : '❌ Connection Error'}
              </p>
              <p className="text-sm opacity-80">
                {errorMessage || 'Configure environment variables for live data'}
              </p>
            </div>
          </div>
          <button
            onClick={() => fetchLiveData()}
            disabled={isLoading}
            className={cn(
              "flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors",
              dataSource === 'mock'
                ? "bg-amber-600 text-white hover:bg-amber-700"
                : "bg-red-600 text-white hover:bg-red-700",
              isLoading && "opacity-50 cursor-not-allowed"
            )}
          >
            <RefreshCw className={cn("w-4 h-4", isLoading && "animate-spin")} />
            {isLoading ? 'Connecting...' : 'Retry Connection'}
          </button>
        </motion.div>
      )}
      
      {/* Live Data Success Banner */}
      {dataSource === 'live' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center justify-between p-4 rounded-lg bg-emerald-50 border border-emerald-300 text-emerald-900"
        >
          <div className="flex items-center gap-3">
            <Wifi className="w-5 h-5" />
            <div>
              <p className="font-semibold">✅ Connected to Live Backend</p>
              <p className="text-sm opacity-80">
                Last updated: {lastFetched ? new Date(lastFetched).toLocaleTimeString() : 'Now'}
              </p>
            </div>
          </div>
          <button
            onClick={() => fetchLiveData()}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg font-medium hover:bg-emerald-700 transition-colors"
          >
            <RefreshCw className={cn("w-4 h-4", isLoading && "animate-spin")} />
            Refresh
          </button>
        </motion.div>
      )}

      {/* Welcome Section */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-wine-900 via-wine-800 to-wine-900 p-8 text-white"
      >
        <div className="absolute top-0 right-0 w-64 h-64 bg-gold-500/10 rounded-full blur-3xl" />
        <div className="relative z-10">
          <h1 className="font-serif text-3xl font-bold mb-2">
            Buonasera, <span className="text-gold-400">Manager</span>
          </h1>
          <p className="text-wine-200">
            {dataSource === 'live' 
              ? `Live monitoring active. ${occupiedTables} tables occupied.`
              : `⚠️ Showing demo data. ${occupiedTables} tables shown.`
            }
          </p>
          
          <div className="flex gap-4 mt-6">
            <div className={cn(
              "flex items-center gap-2 px-4 py-2 rounded-lg",
              onlineAgents > 0 ? "bg-white/10" : "bg-red-500/20"
            )}>
              {onlineAgents > 0 ? (
                <>
                  <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span className="text-sm">{onlineAgents}/{agents.length} agents online</span>
                </>
              ) : (
                <>
                  <WifiOff className="w-4 h-4 text-red-400" />
                  <span className="text-sm">No agents connected</span>
                </>
              )}
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-white/10">
              <ChefHat className="w-4 h-4 text-gold-400" />
              <span className="text-sm">{pendingOrders} orders in kitchen</span>
            </div>
            {readyDishes > 0 && (
              <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-gold-500/20 pulse-gold">
                <AlertCircle className="w-4 h-4 text-gold-400" />
                <span className="text-sm font-medium">{readyDishes} dishes ready to serve!</span>
              </div>
            )}
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-white/5">
              <Database className="w-4 h-4 text-wine-300" />
              <span className="text-xs text-wine-300">
                Source: {dataSource.toUpperCase()}
              </span>
            </div>
          </div>
        </div>
      </motion.div>
      
      {/* Stats Grid */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          icon={Users}
          label="Guests Tonight"
          value={stats.guestsTonight.toString()}
          change="+12%"
          color="wine"
        />
        <StatCard
          icon={DollarSign}
          label="Revenue"
          value={formatCurrency(stats.revenue)}
          change="+8%"
          color="gold"
        />
        <StatCard
          icon={Clock}
          label="Avg Wait Time"
          value={`${stats.avgWaitTime}m`}
          change="-2m"
          positive
          color="emerald"
        />
        <StatCard
          icon={Star}
          label="Satisfaction"
          value={stats.satisfaction.toString()}
          change="+0.1"
          color="amber"
        />
      </div>
      
      {/* Main Content Grid */}
      <div className="grid grid-cols-3 gap-6">
        {/* Agent Activity */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          className="col-span-2 elegant-card p-6"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-serif text-xl font-semibold text-wine-900">Agent Activity</h2>
            <button 
              onClick={() => setActiveView('agents')}
              className="text-sm text-wine-600 hover:text-wine-800"
            >
              View All →
            </button>
          </div>
          
          <div className="space-y-4">
            {agents.map((agent) => (
              <div
                key={agent.id}
                className="flex items-center gap-4 p-4 rounded-xl bg-cream-50 border border-cream-200"
              >
                <div className={cn(
                  "w-12 h-12 rounded-full flex items-center justify-center text-2xl",
                  agent.status === 'busy' ? "bg-gold-100" : "bg-wine-100"
                )}>
                  {agent.avatar}
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-wood-900">{agent.name}</span>
                    <span className={cn(
                      "text-xs px-2 py-0.5 rounded-full",
                      agent.status === 'busy' 
                        ? "bg-gold-100 text-gold-700" 
                        : "bg-emerald-100 text-emerald-700"
                    )}>
                      {agent.status}
                    </span>
                  </div>
                  <p className="text-sm text-wood-500">{agent.currentTask}</p>
                </div>
                <div className="text-right">
                  <span className="text-xs text-wood-400 capitalize">{agent.role}</span>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
        
        {/* Table Overview */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="elegant-card p-6"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-serif text-xl font-semibold text-wine-900">Tables</h2>
            <button 
              onClick={() => setActiveView('floor')}
              className="text-sm text-wine-600 hover:text-wine-800"
            >
              Floor Plan →
            </button>
          </div>
          
          <div className="grid grid-cols-3 gap-2">
            {tables.map((table) => (
              <div
                key={table.id}
                className={cn(
                  "aspect-square rounded-lg flex flex-col items-center justify-center text-center p-2 transition-all cursor-pointer hover:scale-105",
                  table.status === 'available' && "bg-emerald-100 border-2 border-emerald-300",
                  table.status === 'occupied' && "bg-wine-100 border-2 border-wine-300",
                  table.status === 'reserved' && "bg-gold-100 border-2 border-gold-300",
                  table.status === 'cleaning' && "bg-gray-100 border-2 border-gray-300",
                )}
              >
                <span className="text-xs font-medium">{table.id.replace('table-', 'T')}</span>
                <span className="text-lg">{table.capacity}👤</span>
                {table.guestName && (
                  <span className="text-[10px] truncate w-full">{table.guestName}</span>
                )}
              </div>
            ))}
          </div>
          
          {/* Legend */}
          <div className="flex flex-wrap gap-2 mt-4 pt-4 border-t border-cream-200">
            <span className="flex items-center gap-1 text-xs">
              <div className="w-3 h-3 rounded bg-emerald-300" /> Available
            </span>
            <span className="flex items-center gap-1 text-xs">
              <div className="w-3 h-3 rounded bg-wine-300" /> Occupied
            </span>
            <span className="flex items-center gap-1 text-xs">
              <div className="w-3 h-3 rounded bg-gold-300" /> Reserved
            </span>
          </div>
        </motion.div>
      </div>
      
      {/* Kitchen Queue */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="elegant-card p-6"
      >
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-serif text-xl font-semibold text-wine-900">Kitchen Queue</h2>
          <button 
            onClick={() => setActiveView('kitchen')}
            className="text-sm text-wine-600 hover:text-wine-800"
          >
            Full Kitchen View →
          </button>
        </div>
        
        <div className="grid grid-cols-3 gap-4">
          {tickets.slice(0, 3).map((ticket) => (
            <div
              key={ticket.id}
              className={cn(
                "kitchen-ticket",
                ticket.priority === 'vip' && "border-gold-400 bg-gold-50"
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-bold">{ticket.tableId.replace('table-', 'TABLE ')}</span>
                {ticket.priority === 'vip' && (
                  <span className="text-xs px-2 py-0.5 bg-gold-500 text-white rounded">VIP</span>
                )}
              </div>
              <div className="text-xs text-wood-500 mb-2">
                {getTimeSince(ticket.createdAt)}
              </div>
              <div className="space-y-1">
                {ticket.items.map((item, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <span className={cn(
                      "w-2 h-2 rounded-full",
                      item.status === 'served' && "bg-gray-400",
                      item.status === 'ready' && "bg-emerald-500 animate-pulse",
                      item.status === 'preparing' && "bg-gold-500",
                      item.status === 'pending' && "bg-wine-500",
                    )} />
                    <span className={cn(
                      "text-sm",
                      item.status === 'served' && "line-through text-gray-400"
                    )}>
                      {item.quantity}x {item.dish}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </motion.div>
    </div>
  )
}

function StatCard({
  icon: Icon,
  label,
  value,
  change,
  positive = true,
  color,
}: {
  icon: typeof Users
  label: string
  value: string
  change: string
  positive?: boolean
  color: 'wine' | 'gold' | 'emerald' | 'amber'
}) {
  const colorClasses = {
    wine: 'bg-wine-100 text-wine-700',
    gold: 'bg-gold-100 text-gold-700',
    emerald: 'bg-emerald-100 text-emerald-700',
    amber: 'bg-amber-100 text-amber-700',
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="elegant-card p-5"
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-wood-500 mb-1">{label}</p>
          <p className="text-2xl font-serif font-bold text-wine-900">{value}</p>
          <div className="flex items-center gap-1 mt-1">
            <TrendingUp className={cn(
              "w-3 h-3",
              positive ? "text-emerald-600" : "text-wine-600"
            )} />
            <span className={cn(
              "text-xs font-medium",
              positive ? "text-emerald-600" : "text-wine-600"
            )}>
              {change}
            </span>
          </div>
        </div>
        <div className={cn("p-3 rounded-xl", colorClasses[color])}>
          <Icon className="w-5 h-5" />
        </div>
      </div>
    </motion.div>
  )
}
