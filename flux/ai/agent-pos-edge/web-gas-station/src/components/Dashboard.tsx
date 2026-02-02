'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { cn, formatCurrency, formatLiters, formatPercent, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import {
  Fuel,
  TrendingUp,
  DollarSign,
  Gauge,
  Database,
  AlertTriangle,
  Clock,
  Activity,
  ArrowUp,
  ArrowDown,
  Droplets,
  Camera,
  Brain,
  Bot,
  Eye,
  Scan
} from 'lucide-react'

export function Dashboard() {
  const { tanks, pumps, transactions, alerts, agents, cameras, detections, setActiveView } = useGasStationStore()

  const activePumps = pumps.filter(p => p.status === 'active').length
  const idlePumps = pumps.filter(p => p.status === 'idle').length
  const totalRevenue = transactions.filter(t => t.status === 'completed').reduce((sum, t) => sum + t.amount, 0)
  const totalLiters = transactions.filter(t => t.status === 'completed').reduce((sum, t) => sum + t.liters, 0)
  const criticalAlerts = alerts.filter(a => a.type === 'critical' && !a.acknowledged).length
  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'security')
  const activeAgents = visionAgents.filter(a => a.status === 'online' || a.status === 'processing').length
  const onlineCameras = cameras.filter(c => c.status !== 'offline').length
  const todayDetections = detections.length

  const stats = [
    { label: 'Bombas Ativas', value: `${activePumps}/${pumps.length}`, icon: Gauge, color: 'fuel-green', trend: '+2' },
    { label: 'Vendas Hoje', value: formatCurrency(totalRevenue), icon: DollarSign, color: 'fuel-amber', trend: '+15%' },
    { label: 'AI Agents', value: `${activeAgents}/${visionAgents.length}`, icon: Brain, color: 'fuel-blue', trend: 'Online' },
    { label: 'Alertas', value: criticalAlerts.toString(), icon: AlertTriangle, color: criticalAlerts > 0 ? 'fuel-red' : 'fuel-green', trend: criticalAlerts > 0 ? 'Atenção' : 'OK' },
  ]

  const recentTransactions = transactions.slice(0, 5)
  const activeTransactions = transactions.filter(t => t.status === 'in_progress')

  return (
    <div className="p-6 space-y-6">
      {/* Hero Section */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-fuel-green/20 via-fuel-dark to-fuel-lime/10 border border-fuel-green/30 p-8"
      >
        <div className="absolute top-0 right-0 w-96 h-96 bg-fuel-green/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-64 h-64 bg-fuel-lime/10 rounded-full blur-3xl" />
        
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-4">
            <div className="relative">
              <Fuel className="w-10 h-10 text-fuel-green" />
              <div className="absolute inset-0 animate-ping">
                <Fuel className="w-10 h-10 text-fuel-green opacity-50" />
              </div>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-white">
                Gas Station <span className="text-fuel-green">Command Center</span>
              </h1>
              <p className="text-gray-400">Monitoramento em tempo real de operações</p>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-4 mt-6">
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-fuel-dark/50 border border-fuel-green/30">
              <Activity className="w-4 h-4 text-fuel-green animate-pulse" />
              <span className="text-sm text-fuel-green font-mono">{activePumps} bombas operando</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-fuel-dark/50 border border-fuel-blue/30">
              <Camera className="w-4 h-4 text-fuel-blue" />
              <span className="text-sm text-fuel-blue font-mono">{onlineCameras} câmeras online</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-fuel-dark/50 border border-fuel-cyan/30">
              <Brain className="w-4 h-4 text-fuel-cyan animate-pulse" />
              <span className="text-sm text-fuel-cyan font-mono">{activeAgents} AI agents ativos</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-fuel-dark/50 border border-fuel-amber/30">
              <Scan className="w-4 h-4 text-fuel-amber" />
              <span className="text-sm text-fuel-amber font-mono">{todayDetections} detecções hoje</span>
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
            className="fuel-card p-5 hover:border-fuel-green/50 transition-colors group"
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-gray-400 mb-1">{stat.label}</p>
                <p className="text-2xl font-bold text-white font-mono">{stat.value}</p>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className={cn('w-3 h-3', `text-${stat.color}`)} />
                  <span className={cn('text-xs font-mono', `text-${stat.color}`)}>{stat.trend}</span>
                </div>
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
        {/* Tank Overview */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
          className="lg:col-span-2 fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">Níveis dos Tanques</h2>
            <button onClick={() => setActiveView('tanks')} className="text-sm text-fuel-green hover:underline">
              Ver Todos →
            </button>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {tanks.map((tank) => {
              const percent = formatPercent(tank.currentLevel, tank.capacity)
              const color = tank.status === 'critical' ? 'fuel-red' : tank.status === 'low' ? 'fuel-amber' : 'fuel-green'
              
              return (
                <div key={tank.id} className="relative">
                  <div className="tank-container h-40 flex flex-col justify-end p-2">
                    <div 
                      className={cn('tank-fill rounded-lg', `bg-gradient-to-t from-${color} to-${color}/50`)}
                      style={{ height: `${percent}%` }}
                    />
                  </div>
                  <div className="mt-2 text-center">
                    <p className="text-sm font-medium text-white">{tank.name}</p>
                    <p className={cn('text-lg font-bold font-mono', `text-${color}`)}>{percent}%</p>
                    <p className="text-xs text-gray-500 capitalize">{tank.fuelType}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </motion.div>

        {/* Active Transactions */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">Abastecimentos Ativos</h2>
            <span className="px-2 py-1 rounded-full bg-fuel-green/20 text-fuel-green text-xs font-bold">
              {activeTransactions.length}
            </span>
          </div>
          <div className="space-y-3">
            {activeTransactions.length === 0 ? (
              <div className="text-center py-8">
                <Gauge className="w-10 h-10 text-gray-600 mx-auto mb-3" />
                <p className="text-sm text-gray-500">Nenhum abastecimento ativo</p>
              </div>
            ) : (
              activeTransactions.map((txn) => (
                <div key={txn.id} className="p-3 rounded-lg bg-fuel-gray/30 border border-fuel-green/20">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-bold text-white">Bomba {pumps.find(p => p.id === txn.pumpId)?.number}</span>
                    <div className="pump-indicator pump-active" />
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-400 capitalize">{txn.fuelType}</span>
                    <span className="text-fuel-green font-mono">{formatLiters(txn.liters)}</span>
                  </div>
                  <div className="flex items-center justify-between text-sm mt-1">
                    <span className="text-gray-500">{getTimeSince(txn.startTime)}</span>
                    <span className="text-fuel-amber font-mono">{formatCurrency(txn.amount)}</span>
                  </div>
                </div>
              ))
            )}
          </div>
        </motion.div>
      </div>

      {/* AI Agents & Cameras Quick View */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* AI Vision Agents */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.35 }}
          className="fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">AI Vision Agents</h2>
            <button onClick={() => setActiveView('cameras')} className="text-sm text-fuel-green hover:underline">
              Ver Câmeras →
            </button>
          </div>
          <div className="space-y-3">
            {visionAgents.slice(0, 4).map(agent => (
              <div key={agent.id} className="flex items-center justify-between p-3 rounded-lg bg-fuel-gray/30">
                <div className="flex items-center gap-3">
                  <div className={cn(
                    'p-2 rounded-lg',
                    agent.status === 'processing' ? 'bg-fuel-amber/20' : 'bg-fuel-green/20'
                  )}>
                    <Brain className={cn(
                      'w-4 h-4',
                      agent.status === 'processing' ? 'text-fuel-amber animate-pulse' : 'text-fuel-green'
                    )} />
                  </div>
                  <div>
                    <p className="text-sm text-white">{agent.name}</p>
                    <p className="text-xs text-gray-500">{agent.processedToday} processados</p>
                  </div>
                </div>
                <span className={cn(
                  'px-2 py-1 rounded text-xs',
                  agent.status === 'online' ? 'bg-fuel-green/20 text-fuel-green' :
                  agent.status === 'processing' ? 'bg-fuel-amber/20 text-fuel-amber' :
                  'bg-fuel-red/20 text-fuel-red'
                )}>
                  {agent.status}
                </span>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Recent AI Detections */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="lg:col-span-2 fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">Últimas Detecções de IA</h2>
            <button onClick={() => setActiveView('cameras')} className="text-sm text-fuel-green hover:underline">
              Ver Todas →
            </button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {detections.slice(0, 4).map(det => {
              const camera = cameras.find(c => c.id === det.cameraId)
              const agent = agents.find(a => a.id === det.agentId)
              return (
                <div key={det.id} className="flex items-center justify-between p-3 rounded-lg bg-fuel-gray/30 border border-fuel-gray/50">
                  <div className="flex items-center gap-3">
                    <div className={cn(
                      'p-2 rounded-lg',
                      det.type === 'suspicious' ? 'bg-fuel-red/20' : 'bg-fuel-green/20'
                    )}>
                      <Eye className={cn(
                        'w-4 h-4',
                        det.type === 'suspicious' ? 'text-fuel-red' : 'text-fuel-green'
                      )} />
                    </div>
                    <div>
                      <p className="text-sm text-white capitalize">{det.type.replace('-', ' ')}</p>
                      <p className="text-xs text-gray-500">{camera?.name} • {agent?.name}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-mono text-fuel-green">{(det.confidence * 100).toFixed(0)}%</p>
                    <p className="text-xs text-gray-500">{getTimeSince(det.timestamp)}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </motion.div>
      </div>

      {/* Recent Transactions & Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Transactions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">Transações Recentes</h2>
            <button onClick={() => setActiveView('transactions')} className="text-sm text-fuel-green hover:underline">
              Ver Todas →
            </button>
          </div>
          <div className="space-y-2">
            {recentTransactions.map((txn) => (
              <div key={txn.id} className="flex items-center justify-between p-3 rounded-lg bg-fuel-gray/30 hover:bg-fuel-gray/50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className={cn(
                    'w-2 h-2 rounded-full',
                    txn.status === 'completed' ? 'bg-fuel-green' : txn.status === 'in_progress' ? 'bg-fuel-amber animate-pulse' : 'bg-fuel-red'
                  )} />
                  <div>
                    <p className="text-sm text-white">Bomba {pumps.find(p => p.id === txn.pumpId)?.number} - {txn.fuelType}</p>
                    <p className="text-xs text-gray-500">{getTimeSince(txn.startTime)}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-mono text-fuel-green">{formatCurrency(txn.amount)}</p>
                  <p className="text-xs text-gray-500 font-mono">{formatLiters(txn.liters)}</p>
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Recent Alerts */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="fuel-card p-5"
        >
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-white">Alertas Recentes</h2>
            <button onClick={() => setActiveView('alerts')} className="text-sm text-fuel-green hover:underline">
              Ver Todos →
            </button>
          </div>
          <div className="space-y-2">
            {alerts.slice(0, 5).map((alert) => (
              <div 
                key={alert.id} 
                className={cn(
                  'p-3 rounded-lg border transition-colors',
                  alert.type === 'critical' ? 'bg-fuel-red/10 border-fuel-red/30' :
                  alert.type === 'warning' ? 'bg-fuel-amber/10 border-fuel-amber/30' :
                  'bg-fuel-blue/10 border-fuel-blue/30',
                  alert.acknowledged && 'opacity-50'
                )}
              >
                <div className="flex items-start gap-3">
                  <AlertTriangle className={cn(
                    'w-5 h-5 flex-shrink-0',
                    alert.type === 'critical' ? 'text-fuel-red' :
                    alert.type === 'warning' ? 'text-fuel-amber' : 'text-fuel-blue'
                  )} />
                  <div>
                    <p className="text-sm font-medium text-white">{alert.title}</p>
                    <p className="text-xs text-gray-400 mt-0.5">{alert.message}</p>
                    <p className="text-xs text-gray-500 mt-1">{getTimeSince(alert.timestamp)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </div>
  )
}
