'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Users, 
  MessageSquare, 
  Bot, 
  Mic, 
  Image, 
  MapPin,
  Activity,
  TrendingUp,
  Zap,
  Clock,
  AlertTriangle,
  RefreshCw,
  Wifi,
  WifiOff,
  Database,
} from 'lucide-react'

// Types for API responses
interface MetricsData {
  success: boolean
  source: string
  message?: string
  metrics: Record<string, any>
  agents: Array<{ name: string; status: string; version: string }>
}

interface AgentStatus {
  name: string
  status: 'online' | 'offline'
  requests: string
  metrics?: { cpu: number; memory: number }
}

// ⚠️ MOCK DATA - Used when live services are unavailable
const MOCK_METRICS = {
  totalUsers: 0,
  activeUsers: 0,
  totalMessages: 0,
  messagesLast24h: 0,
  totalAgents: 5,
  agentsOnline: 0,
  voiceClonesActive: 0,
  imagesGenerated: 0,
  locationAlerts: 0,
}

const MOCK_AGENT_STATUS: AgentStatus[] = [
  { name: 'messaging-hub', status: 'offline', requests: '0/min' },
  { name: 'voice-agent', status: 'offline', requests: '0/min' },
  { name: 'media-agent', status: 'offline', requests: '0/min' },
  { name: 'location-agent', status: 'offline', requests: '0/min' },
  { name: 'command-center', status: 'offline', requests: '0/min' },
]

export function DashboardView() {
  const [dataSource, setDataSource] = useState<'live' | 'mock' | 'loading'>('loading')
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [metrics, setMetrics] = useState(MOCK_METRICS)
  const [agentStatus, setAgentStatus] = useState<AgentStatus[]>(MOCK_AGENT_STATUS)
  const [lastFetched, setLastFetched] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const fetchLiveData = async () => {
    setIsLoading(true)
    try {
      const [metricsRes, agentsRes] = await Promise.all([
        fetch('/api/metrics'),
        fetch('/api/agents'),
      ])
      
      const metricsData: MetricsData = await metricsRes.json()
      const agentsData = await agentsRes.json()
      
      if (metricsData.success || agentsData.success) {
        setDataSource('live')
        setErrorMessage(null)
        
        // Update agents from live data
        if (agentsData.success && agentsData.agents?.length > 0) {
          const liveAgents: AgentStatus[] = agentsData.agents.map((a: any) => ({
            name: a.name,
            status: a.status === 'online' ? 'online' : 'offline',
            requests: a.metrics?.requests ? `${a.metrics.requests}/min` : '0/min',
            metrics: a.metrics,
          }))
          setAgentStatus(liveAgents)
          setMetrics(prev => ({
            ...prev,
            totalAgents: liveAgents.length,
            agentsOnline: liveAgents.filter(a => a.status === 'online').length,
          }))
        }
        
        setLastFetched(new Date().toISOString())
      } else {
        setDataSource('mock')
        setErrorMessage(metricsData.message || agentsData.message || 'Could not fetch live data')
        setAgentStatus(MOCK_AGENT_STATUS)
        setMetrics(MOCK_METRICS)
      }
    } catch (error) {
      setDataSource('mock')
      setErrorMessage(`Failed to connect: ${error}`)
      setAgentStatus(MOCK_AGENT_STATUS)
      setMetrics(MOCK_METRICS)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchLiveData()
    // Refresh every 30 seconds
    const interval = setInterval(fetchLiveData, 30000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="space-y-6">
      {/* ⚠️ Data Source Warning Banner */}
      {dataSource === 'mock' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center justify-between p-4 rounded-lg bg-yellow-500/20 border border-yellow-500/50"
        >
          <div className="flex items-center gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-500" />
            <div>
              <p className="font-semibold text-yellow-500">⚠️ Using MOCK Data - Not Connected to Live Services</p>
              <p className="text-sm text-yellow-500/80">
                {errorMessage || 'Configure PROMETHEUS_URL and KUBERNETES_API_URL for real metrics'}
              </p>
            </div>
          </div>
          <button
            onClick={fetchLiveData}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-yellow-500 text-black rounded-lg font-medium hover:bg-yellow-400 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            {isLoading ? 'Connecting...' : 'Retry'}
          </button>
        </motion.div>
      )}
      
      {dataSource === 'live' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center justify-between p-4 rounded-lg bg-cyber-green/20 border border-cyber-green/50"
        >
          <div className="flex items-center gap-3">
            <Wifi className="w-5 h-5 text-cyber-green" />
            <div>
              <p className="font-semibold text-cyber-green">✅ Connected to Live Backend</p>
              <p className="text-sm text-cyber-green/80">
                Last updated: {lastFetched ? new Date(lastFetched).toLocaleTimeString() : 'Now'}
              </p>
            </div>
          </div>
          <button
            onClick={fetchLiveData}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-cyber-green text-black rounded-lg font-medium hover:bg-cyber-green/80 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </motion.div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          icon={Users}
          label="Total Users"
          value={metrics.totalUsers.toLocaleString()}
          subValue={dataSource === 'live' ? `${metrics.activeUsers} active now` : '⚠️ Mock data'}
          color="from-cyber-purple to-cyber-pink"
        />
        <StatCard
          icon={MessageSquare}
          label="Messages (24h)"
          value={metrics.messagesLast24h.toLocaleString()}
          subValue={dataSource === 'live' ? `${metrics.totalMessages.toLocaleString()} total` : '⚠️ Mock data'}
          color="from-cyber-blue to-cyber-cyan"
        />
        <StatCard
          icon={Bot}
          label="Agents Online"
          value={`${metrics.agentsOnline}/${metrics.totalAgents}`}
          subValue={dataSource === 'live' ? 'From Kubernetes' : '⚠️ Not connected'}
          color="from-cyber-green to-emerald-400"
        />
        <StatCard
          icon={Zap}
          label="Voice Clones"
          value={metrics.voiceClonesActive.toLocaleString()}
          subValue={dataSource === 'live' ? 'Active profiles' : '⚠️ Mock data'}
          color="from-cyber-yellow to-orange-400"
        />
      </div>

      {/* Second Row Stats */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard
          icon={Image}
          label="Images Generated"
          value={metrics.imagesGenerated.toLocaleString()}
          subValue={dataSource === 'live' ? 'Via Media Agent' : '⚠️ Mock data'}
          color="from-pink-500 to-rose-400"
        />
        <StatCard
          icon={MapPin}
          label="Location Alerts"
          value={metrics.locationAlerts.toLocaleString()}
          subValue={dataSource === 'live' ? 'Proximity notifications' : '⚠️ Mock data'}
          color="from-indigo-500 to-purple-400"
        />
        <StatCard
          icon={Activity}
          label="System Health"
          value={dataSource === 'live' ? '99.9%' : '0%'}
          subValue={dataSource === 'live' ? 'From Prometheus' : '⚠️ Not monitored'}
          color="from-emerald-500 to-teal-400"
        />
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-2 gap-6">
        {/* Agent Status */}
        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <Bot className="w-5 h-5 text-cyber-purple" />
            Agent Status
            {dataSource === 'mock' && (
              <span className="text-xs px-2 py-0.5 bg-yellow-500/20 text-yellow-500 rounded">MOCK</span>
            )}
            {dataSource === 'live' && (
              <span className="text-xs px-2 py-0.5 bg-cyber-green/20 text-cyber-green rounded">LIVE</span>
            )}
          </h3>
          <div className="space-y-3">
            {agentStatus.map((agent) => (
              <div 
                key={agent.name}
                className="flex items-center justify-between p-3 bg-cyber-dark/50 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <div className={`w-2 h-2 rounded-full ${
                    agent.status === 'online' ? 'bg-cyber-green animate-pulse' : 'bg-cyber-red'
                  }`} />
                  <span className="font-mono text-sm">{agent.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  {agent.metrics && dataSource === 'live' && (
                    <span className="text-xs text-gray-400">
                      CPU: {agent.metrics.cpu}% | Mem: {agent.metrics.memory}MB
                    </span>
                  )}
                  <span className="text-xs text-gray-500">{agent.requests}</span>
                </div>
              </div>
            ))}
          </div>
          {dataSource === 'mock' && (
            <p className="text-xs text-yellow-500/70 mt-4 p-2 bg-yellow-500/10 rounded">
              ⚠️ Showing placeholder data. Connect to Kubernetes for real agent status.
            </p>
          )}
        </div>

        {/* Recent Activity */}
        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <Clock className="w-5 h-5 text-cyber-purple" />
            Recent Activity
            {dataSource === 'mock' && (
              <span className="text-xs px-2 py-0.5 bg-yellow-500/20 text-yellow-500 rounded">MOCK</span>
            )}
          </h3>
          {dataSource === 'mock' ? (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <WifiOff className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-sm">No live activity data</p>
              <p className="text-xs mt-1">Connect to backend services to see real events</p>
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-sm text-gray-400">Activity feed from connected services...</p>
            </div>
          )}
        </div>
      </div>
      
      {/* Data Source Footer */}
      <div className="flex items-center justify-center gap-2 text-xs text-gray-500 py-4">
        <Database className="w-4 h-4" />
        <span>Data Source: {dataSource.toUpperCase()}</span>
        {lastFetched && <span>| Last Updated: {new Date(lastFetched).toLocaleTimeString()}</span>}
      </div>
    </div>
  )
}

interface StatCardProps {
  icon: React.ElementType
  label: string
  value: string
  subValue: string
  color: string
}

function StatCard({ icon: Icon, label, value, subValue, color }: StatCardProps) {
  return (
    <motion.div 
      className="card card-hover p-5"
      whileHover={{ scale: 1.02 }}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-gray-400 uppercase tracking-wide">{label}</p>
          <p className={`text-3xl font-bold mt-1 bg-gradient-to-r ${color} bg-clip-text text-transparent`}>
            {value}
          </p>
          <p className="text-xs text-gray-500 mt-1">{subValue}</p>
        </div>
        <div className={`p-3 rounded-xl bg-gradient-to-br ${color} bg-opacity-20`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </motion.div>
  )
}

function ActivityIcon({ type }: { type: string }) {
  const icons: Record<string, { icon: React.ElementType; color: string }> = {
    user: { icon: Users, color: 'text-cyber-purple' },
    voice: { icon: Mic, color: 'text-cyber-green' },
    image: { icon: Image, color: 'text-cyber-pink' },
    location: { icon: MapPin, color: 'text-cyber-blue' },
    agent: { icon: Bot, color: 'text-cyber-yellow' },
  }
  
  const { icon: Icon, color } = icons[type] || icons.user
  return <Icon className={`w-4 h-4 ${color} mt-0.5`} />
}
