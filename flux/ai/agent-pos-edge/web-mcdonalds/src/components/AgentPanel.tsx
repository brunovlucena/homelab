'use client'

import { useMcdonaldsStore, Agent } from '@/store/mcdonaldsStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { 
  Bot, 
  Brain, 
  Cpu, 
  HardDrive, 
  Activity, 
  Clock, 
  RefreshCw, 
  Camera,
  Utensils,
  Users,
  Sparkles,
  Package,
  ShieldCheck
} from 'lucide-react'

function AgentCard({ agent }: { agent: Agent }) {
  const { cameras } = useMcdonaldsStore()
  const assignedCameraData = cameras.filter(c => agent.assignedCameras?.includes(c.id))

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'vision': return Brain
      case 'quality': return Sparkles
      case 'customer': return Users
      case 'inventory': return Package
      case 'order': return Utensils
      default: return Bot
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'vision': return 'mc-gold'
      case 'quality': return 'mc-green'
      case 'customer': return 'mc-blue'
      case 'inventory': return 'mc-orange'
      case 'order': return 'mc-red'
      default: return 'gray-400'
    }
  }

  const Icon = getTypeIcon(agent.type)
  const color = getTypeColor(agent.type)

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn(
        'mc-card p-5 transition-all',
        agent.status === 'processing' && `border-${color}/50`
      )}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className={cn(
            'p-3 rounded-xl',
            `bg-${color}/10 border border-${color}/30`
          )}>
            <Icon className={cn('w-6 h-6', `text-${color}`)} />
          </div>
          <div>
            <h3 className="font-brand font-bold text-white">{agent.name}</h3>
            <p className="text-xs text-gray-500">{agent.description}</p>
          </div>
        </div>
        <span className={cn(
          'px-2 py-1 rounded text-xs font-medium',
          agent.status === 'online' ? 'bg-mc-green/20 text-mc-green' :
          agent.status === 'processing' ? 'bg-mc-gold/20 text-mc-gold animate-pulse' :
          'bg-mc-red/20 text-mc-red'
        )}>
          {agent.status}
        </span>
      </div>

      {/* Capabilities */}
      <div className="flex flex-wrap gap-1 mb-4">
        {agent.capabilities.slice(0, 4).map(cap => (
          <span key={cap} className="px-2 py-0.5 rounded bg-mc-gray/50 text-xs text-gray-400">
            {cap}
          </span>
        ))}
        {agent.capabilities.length > 4 && (
          <span className="px-2 py-0.5 rounded bg-mc-gray/50 text-xs text-gray-400">
            +{agent.capabilities.length - 4}
          </span>
        )}
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-3 gap-3 mb-4">
        <div className="p-3 rounded-lg bg-mc-gray/30">
          <div className="flex items-center gap-2 mb-1">
            <Cpu className="w-3 h-3 text-gray-500" />
            <span className="text-xs text-gray-500">CPU</span>
          </div>
          <p className="text-lg font-bold font-mono text-white">{agent.metrics.cpu}%</p>
          <div className="mt-1 h-1.5 bg-mc-gray rounded-full overflow-hidden">
            <div 
              className={cn(
                'h-full rounded-full transition-all',
                agent.metrics.cpu > 80 ? 'bg-mc-red' : agent.metrics.cpu > 50 ? 'bg-mc-gold' : 'bg-mc-green'
              )}
              style={{ width: `${agent.metrics.cpu}%` }}
            />
          </div>
        </div>
        <div className="p-3 rounded-lg bg-mc-gray/30">
          <div className="flex items-center gap-2 mb-1">
            <HardDrive className="w-3 h-3 text-gray-500" />
            <span className="text-xs text-gray-500">Memory</span>
          </div>
          <p className="text-lg font-bold font-mono text-white">{agent.metrics.memory}%</p>
          <div className="mt-1 h-1.5 bg-mc-gray rounded-full overflow-hidden">
            <div 
              className="h-full bg-mc-blue rounded-full transition-all"
              style={{ width: `${agent.metrics.memory}%` }}
            />
          </div>
        </div>
        <div className="p-3 rounded-lg bg-mc-gray/30">
          <div className="flex items-center gap-2 mb-1">
            <Activity className="w-3 h-3 text-gray-500" />
            <span className="text-xs text-gray-500">Processed</span>
          </div>
          <p className="text-lg font-bold font-mono text-mc-gold">{agent.processedToday}</p>
          <p className="text-xs text-gray-500 mt-1">today</p>
        </div>
      </div>

      {/* Inference Time */}
      {agent.metrics.inferenceTime && (
        <div className="flex items-center justify-between p-2 rounded bg-mc-gray/30 mb-4">
          <span className="text-xs text-gray-400">Inference Time</span>
          <span className="text-sm font-mono text-mc-green">{agent.metrics.inferenceTime}ms</span>
        </div>
      )}

      {/* Assigned Cameras */}
      {assignedCameraData.length > 0 && (
        <div className="pt-4 border-t border-mc-gray/30">
          <p className="text-xs text-gray-500 mb-2 flex items-center gap-1">
            <Camera className="w-3 h-3" />
            Monitored Cameras ({assignedCameraData.length})
          </p>
          <div className="flex flex-wrap gap-2">
            {assignedCameraData.map(cam => (
              <span key={cam.id} className="px-2 py-1 rounded bg-mc-gray/50 text-xs text-gray-400">
                {cam.name}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between mt-4 pt-4 border-t border-mc-gray/30 text-xs text-gray-500">
        <span className="font-mono">{agent.version}</span>
        <div className="flex items-center gap-1">
          <Clock className="w-3 h-3" />
          {getTimeSince(agent.lastHeartbeat)}
        </div>
      </div>
    </motion.div>
  )
}

export function AgentPanel() {
  const { agents, cameras } = useMcdonaldsStore()

  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'quality' || a.type === 'customer')
  const operationalAgents = agents.filter(a => a.type === 'order' || a.type === 'inventory')
  const activeAgents = agents.filter(a => a.status === 'online' || a.status === 'processing').length
  const totalProcessed = agents.reduce((sum, a) => sum + a.processedToday, 0)

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">AI Agents</h1>
          <p className="text-gray-400">Monitoramento de agentes inteligentes</p>
        </div>
        <button className="mc-button">
          <RefreshCw className="w-4 h-4 inline mr-2" />
          Atualizar
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="mc-card p-4">
          <p className="text-sm text-gray-400">Total Agents</p>
          <p className="text-2xl font-bold font-mono text-white">{agents.length}</p>
        </div>
        <div className="mc-card p-4 border-mc-green/30">
          <p className="text-sm text-gray-400">Online</p>
          <p className="text-2xl font-bold font-mono text-mc-green">{activeAgents}</p>
        </div>
        <div className="mc-card p-4 border-mc-gold/30">
          <p className="text-sm text-gray-400">Processing</p>
          <p className="text-2xl font-bold font-mono text-mc-gold">{agents.filter(a => a.status === 'processing').length}</p>
        </div>
        <div className="mc-card p-4 border-mc-blue/30">
          <p className="text-sm text-gray-400">Total Processed</p>
          <p className="text-2xl font-bold font-mono text-mc-blue">{totalProcessed}</p>
        </div>
      </div>

      {/* Vision & Quality Agents */}
      <div>
        <h2 className="text-lg font-brand font-bold text-white mb-4 flex items-center gap-2">
          <Brain className="w-5 h-5 text-mc-gold" />
          Vision & Quality Agents
        </h2>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {visionAgents.map(agent => (
            <AgentCard key={agent.id} agent={agent} />
          ))}
        </div>
      </div>

      {/* Operational Agents */}
      <div>
        <h2 className="text-lg font-brand font-bold text-white mb-4 flex items-center gap-2">
          <Utensils className="w-5 h-5 text-mc-red" />
          Operational Agents
        </h2>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {operationalAgents.map(agent => (
            <AgentCard key={agent.id} agent={agent} />
          ))}
        </div>
      </div>
    </div>
  )
}
