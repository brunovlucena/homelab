'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Bot, Cpu, HardDrive, Activity, Clock, RefreshCw, Terminal } from 'lucide-react'

export function AgentStatus() {
  const { agents } = useGasStationStore()

  const onlineAgents = agents.filter(a => a.status === 'online').length

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Status dos Agentes</h1>
          <p className="text-gray-400">Monitoramento de agentes POS Edge</p>
        </div>
        <button className="fuel-button">
          <RefreshCw className="w-4 h-4 inline mr-2" />
          Atualizar
        </button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Total de Agentes</p>
          <p className="text-2xl font-bold font-mono text-white">{agents.length}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-green/30">
          <p className="text-sm text-gray-400">Online</p>
          <p className="text-2xl font-bold font-mono text-fuel-green">{onlineAgents}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-red/30">
          <p className="text-sm text-gray-400">Offline</p>
          <p className="text-2xl font-bold font-mono text-fuel-red">{agents.length - onlineAgents}</p>
        </div>
      </div>

      {/* Agents Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {agents.map((agent, index) => (
          <motion.div
            key={agent.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className={cn(
              'fuel-card p-6 transition-all',
              agent.status === 'online' && 'border-fuel-green/30',
              agent.status === 'offline' && 'border-fuel-red/30',
              agent.status === 'degraded' && 'border-fuel-amber/30'
            )}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={cn(
                  'p-3 rounded-xl',
                  agent.status === 'online' ? 'bg-fuel-green/10 border border-fuel-green/30' :
                  agent.status === 'degraded' ? 'bg-fuel-amber/10 border border-fuel-amber/30' :
                  'bg-fuel-red/10 border border-fuel-red/30'
                )}>
                  <Bot className={cn(
                    'w-6 h-6',
                    agent.status === 'online' ? 'text-fuel-green' :
                    agent.status === 'degraded' ? 'text-fuel-amber' : 'text-fuel-red'
                  )} />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">{agent.name}</h3>
                  <p className="text-sm text-gray-400 capitalize">{agent.type.replace('-', ' ')}</p>
                </div>
              </div>
              <span className={cn(
                'status-badge',
                agent.status === 'online' ? 'status-online' :
                agent.status === 'degraded' ? 'status-warning' : 'status-offline'
              )}>
                {agent.status}
              </span>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-3 gap-4 mb-4">
              <div className="p-3 rounded-lg bg-fuel-gray/30">
                <div className="flex items-center gap-2 mb-2">
                  <Cpu className="w-4 h-4 text-gray-500" />
                  <span className="text-xs text-gray-500">CPU</span>
                </div>
                <p className="text-xl font-bold font-mono text-white">{agent.metrics.cpu}%</p>
                <div className="mt-2 h-1.5 bg-fuel-gray rounded-full overflow-hidden">
                  <div 
                    className="h-full bg-fuel-green rounded-full transition-all"
                    style={{ width: `${agent.metrics.cpu}%` }}
                  />
                </div>
              </div>
              <div className="p-3 rounded-lg bg-fuel-gray/30">
                <div className="flex items-center gap-2 mb-2">
                  <HardDrive className="w-4 h-4 text-gray-500" />
                  <span className="text-xs text-gray-500">Memory</span>
                </div>
                <p className="text-xl font-bold font-mono text-white">{agent.metrics.memory}%</p>
                <div className="mt-2 h-1.5 bg-fuel-gray rounded-full overflow-hidden">
                  <div 
                    className="h-full bg-fuel-blue rounded-full transition-all"
                    style={{ width: `${agent.metrics.memory}%` }}
                  />
                </div>
              </div>
              <div className="p-3 rounded-lg bg-fuel-gray/30">
                <div className="flex items-center gap-2 mb-2">
                  <Activity className="w-4 h-4 text-gray-500" />
                  <span className="text-xs text-gray-500">Requests</span>
                </div>
                <p className="text-xl font-bold font-mono text-white">{agent.metrics.requests}</p>
                <p className="text-xs text-gray-500 mt-1">last hour</p>
              </div>
            </div>

            {/* Info */}
            <div className="flex items-center justify-between pt-4 border-t border-fuel-gray/50 text-sm">
              <div className="flex items-center gap-2 text-gray-400">
                <Terminal className="w-4 h-4" />
                <span className="font-mono">{agent.version}</span>
              </div>
              <div className="flex items-center gap-2 text-gray-400">
                <Clock className="w-4 h-4" />
                <span>Heartbeat: {getTimeSince(agent.lastHeartbeat)}</span>
              </div>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  )
}
