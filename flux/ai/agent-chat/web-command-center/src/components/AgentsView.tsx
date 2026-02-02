'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Search, 
  RefreshCw, 
  MoreVertical, 
  Activity,
  Cpu,
  HardDrive,
  Zap,
  Clock,
  TrendingUp,
  AlertTriangle
} from 'lucide-react'
import type { Agent, AgentStatus } from '@/types'

// Mock data
const mockAgents: Agent[] = [
  {
    id: 'messaging-hub',
    name: 'messaging-hub',
    displayName: 'Messaging Hub',
    description: 'Central message routing and delivery',
    role: 'core',
    status: 'online',
    avatar: '💬',
    color: 'from-cyber-purple to-cyber-pink',
    namespace: 'agent-chat',
    capabilities: ['Message routing', 'WebSocket', 'Presence', 'History'],
    metrics: {
      eventsProcessed: 45832,
      successRate: 99.9,
      avgResponseTime: 12,
      uptime: 99.99,
      activeConnections: 89,
      queueDepth: 23,
      cpuUsage: 35,
      memoryUsage: 42,
    }
  },
  {
    id: 'voice-agent',
    name: 'voice-agent',
    displayName: 'Voice Agent',
    description: 'Voice cloning, TTS, and speech recognition',
    role: 'capability',
    status: 'online',
    avatar: '🗣️',
    color: 'from-cyber-green to-emerald-400',
    namespace: 'agent-chat',
    capabilities: ['Voice cloning', 'TTS', 'STT', 'Voice ID'],
    metrics: {
      eventsProcessed: 2341,
      successRate: 98.5,
      avgResponseTime: 450,
      uptime: 99.9,
      activeConnections: 12,
      queueDepth: 5,
      cpuUsage: 65,
      memoryUsage: 78,
    }
  },
  {
    id: 'media-agent',
    name: 'media-agent',
    displayName: 'Media Agent',
    description: 'AI-powered image and video generation',
    role: 'capability',
    status: 'online',
    avatar: '🎨',
    color: 'from-cyber-pink to-rose-400',
    namespace: 'agent-chat',
    capabilities: ['Image gen', 'Video gen', 'Style transfer', 'Analysis'],
    metrics: {
      eventsProcessed: 892,
      successRate: 97.2,
      avgResponseTime: 3500,
      uptime: 99.5,
      activeConnections: 3,
      queueDepth: 8,
      cpuUsage: 82,
      memoryUsage: 85,
    }
  },
  {
    id: 'location-agent',
    name: 'location-agent',
    displayName: 'Location Agent',
    description: 'Location tracking and proximity alerts',
    role: 'capability',
    status: 'online',
    avatar: '📍',
    color: 'from-cyber-blue to-cyan-400',
    namespace: 'agent-chat',
    capabilities: ['Tracking', 'Geofencing', 'Proximity', 'Geocoding'],
    metrics: {
      eventsProcessed: 15623,
      successRate: 99.8,
      avgResponseTime: 8,
      uptime: 99.99,
      activeConnections: 156,
      queueDepth: 0,
      cpuUsage: 22,
      memoryUsage: 35,
    }
  },
  {
    id: 'command-center',
    name: 'command-center',
    displayName: 'Command Center',
    description: 'Platform orchestration and admin',
    role: 'orchestrator',
    status: 'online',
    avatar: '🎛️',
    color: 'from-cyber-yellow to-orange-400',
    namespace: 'agent-chat',
    capabilities: ['Monitoring', 'User mgmt', 'Agent deploy', 'Analytics'],
    metrics: {
      eventsProcessed: 3456,
      successRate: 99.7,
      avgResponseTime: 45,
      uptime: 99.99,
      activeConnections: 5,
      queueDepth: 2,
      cpuUsage: 28,
      memoryUsage: 45,
    }
  },
]

export function AgentsView() {
  const [searchQuery, setSearchQuery] = useState('')

  const filteredAgents = mockAgents.filter(agent =>
    agent.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    agent.displayName.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search agents..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input-field pl-10 w-80"
          />
        </div>
        <button className="btn-secondary flex items-center gap-2">
          <RefreshCw className="w-4 h-4" />
          Refresh Status
        </button>
      </div>

      {/* Agents Grid */}
      <div className="grid grid-cols-2 gap-6">
        {filteredAgents.map((agent) => (
          <AgentCard key={agent.id} agent={agent} />
        ))}
      </div>
    </div>
  )
}

function AgentCard({ agent }: { agent: Agent }) {
  return (
    <motion.div
      className="card card-hover p-6"
      whileHover={{ scale: 1.01 }}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-4">
          <div className={`w-14 h-14 rounded-xl bg-gradient-to-br ${agent.color} flex items-center justify-center text-2xl shadow-lg`}>
            {agent.avatar}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-bold text-lg">{agent.displayName}</h3>
              <StatusBadge status={agent.status} />
            </div>
            <p className="text-sm text-gray-400">{agent.description}</p>
          </div>
        </div>
        <button className="p-2 rounded-lg hover:bg-cyber-purple/20 text-gray-400">
          <MoreVertical className="w-4 h-4" />
        </button>
      </div>

      {/* Capabilities */}
      <div className="flex flex-wrap gap-2 mb-4">
        {agent.capabilities.map((cap) => (
          <span 
            key={cap} 
            className="px-2 py-1 text-xs bg-cyber-dark/50 rounded-lg text-gray-400"
          >
            {cap}
          </span>
        ))}
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-4 gap-3 mb-4">
        <MetricBox 
          icon={Activity} 
          label="Success" 
          value={`${agent.metrics.successRate}%`}
          color={agent.metrics.successRate >= 99 ? 'text-cyber-green' : 'text-cyber-yellow'}
        />
        <MetricBox 
          icon={Zap} 
          label="Latency" 
          value={`${agent.metrics.avgResponseTime}ms`}
          color={agent.metrics.avgResponseTime < 100 ? 'text-cyber-green' : 'text-cyber-yellow'}
        />
        <MetricBox 
          icon={TrendingUp} 
          label="Events" 
          value={formatNumber(agent.metrics.eventsProcessed)}
          color="text-cyber-purple"
        />
        <MetricBox 
          icon={Clock} 
          label="Uptime" 
          value={`${agent.metrics.uptime}%`}
          color="text-cyber-blue"
        />
      </div>

      {/* Resource Usage */}
      <div className="grid grid-cols-2 gap-4">
        <ResourceBar 
          icon={Cpu} 
          label="CPU" 
          value={agent.metrics.cpuUsage}
        />
        <ResourceBar 
          icon={HardDrive} 
          label="Memory" 
          value={agent.metrics.memoryUsage}
        />
      </div>
    </motion.div>
  )
}

function StatusBadge({ status }: { status: AgentStatus }) {
  const styles: Record<AgentStatus, { bg: string; text: string }> = {
    online: { bg: 'bg-cyber-green/20', text: 'text-cyber-green' },
    offline: { bg: 'bg-cyber-red/20', text: 'text-cyber-red' },
    scaling: { bg: 'bg-cyber-blue/20', text: 'text-cyber-blue' },
    error: { bg: 'bg-cyber-red/20', text: 'text-cyber-red' },
    deploying: { bg: 'bg-cyber-yellow/20', text: 'text-cyber-yellow' },
  }

  const { bg, text } = styles[status]

  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${bg} ${text} flex items-center gap-1`}>
      <span className="w-1.5 h-1.5 rounded-full bg-current animate-pulse" />
      {status}
    </span>
  )
}

function MetricBox({ icon: Icon, label, value, color }: { 
  icon: React.ElementType
  label: string
  value: string
  color: string
}) {
  return (
    <div className="bg-cyber-dark/50 rounded-lg p-2 text-center">
      <Icon className={`w-4 h-4 ${color} mx-auto mb-1`} />
      <p className={`text-sm font-bold ${color}`}>{value}</p>
      <p className="text-xs text-gray-500">{label}</p>
    </div>
  )
}

function ResourceBar({ icon: Icon, label, value }: {
  icon: React.ElementType
  label: string
  value: number
}) {
  const color = value < 50 ? 'bg-cyber-green' : value < 80 ? 'bg-cyber-yellow' : 'bg-cyber-red'
  
  return (
    <div className="bg-cyber-dark/50 rounded-lg p-3">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Icon className="w-4 h-4 text-gray-400" />
          <span className="text-sm text-gray-400">{label}</span>
        </div>
        <span className="text-sm font-medium">{value}%</span>
      </div>
      <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
        <div 
          className={`h-full ${color} transition-all duration-500`}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  )
}

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}
