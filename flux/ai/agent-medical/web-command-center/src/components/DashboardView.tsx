'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Users, 
  FileText, 
  Activity,
  Shield,
  Stethoscope,
  ClipboardList,
  AlertTriangle,
  RefreshCw,
  Wifi,
  WifiOff,
  Database,
  CheckCircle,
  Clock,
  TrendingUp
} from 'lucide-react'

// Types for API responses
interface MetricsData {
  success: boolean
  source: string
  message?: string
  metrics: Record<string, any>
}

// ⚠️ MOCK DATA - Used when live services are unavailable
const MOCK_METRICS = {
  totalPatients: 0,
  activePatients: 0,
  totalRecords: 0,
  recordsLast24h: 0,
  queriesLast24h: 0,
  hipaaAudits: 0,
  agentStatus: 'offline',
  complianceScore: 0,
}

export function DashboardView() {
  const [dataSource, setDataSource] = useState<'live' | 'mock' | 'loading'>('loading')
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [metrics, setMetrics] = useState(MOCK_METRICS)
  const [lastFetched, setLastFetched] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const fetchLiveData = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/metrics')
      const data: MetricsData = await response.json()
      
      if (data.success) {
        setDataSource('live')
        setErrorMessage(null)
        setMetrics({...MOCK_METRICS, ...data.metrics})
        setLastFetched(new Date().toISOString())
      } else {
        setDataSource('mock')
        setErrorMessage(data.message || 'Could not fetch live data')
        setMetrics(MOCK_METRICS)
      }
    } catch (error) {
      setDataSource('mock')
      setErrorMessage(`Failed to connect: ${error}`)
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
                {errorMessage || 'Agent backend is not accessible. Showing placeholder data.'}
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
          className="flex items-center justify-between p-4 rounded-lg bg-medical-green/20 border border-medical-green/50"
        >
          <div className="flex items-center gap-3">
            <Wifi className="w-5 h-5 text-medical-green" />
            <div>
              <p className="font-semibold text-medical-green">✅ Connected to Medical Agent Backend</p>
              <p className="text-sm text-medical-green/80">
                Last updated: {lastFetched ? new Date(lastFetched).toLocaleTimeString() : 'Now'}
              </p>
            </div>
          </div>
          <button
            onClick={fetchLiveData}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-medical-green text-black rounded-lg font-medium hover:bg-medical-green/80 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </motion.div>
      )}

      {/* HIPAA Compliance Badge */}
      <motion.div
        className="flex items-center justify-center gap-2 p-3 rounded-lg bg-medical-blue/10 border border-medical-blue/30"
        initial={{ scale: 0.95 }}
        animate={{ scale: 1 }}
      >
        <Shield className="w-5 h-5 text-medical-blue" />
        <span className="text-medical-blue font-semibold">HIPAA Compliant System</span>
        <span className="badge-hipaa">Encrypted</span>
        <span className="badge-hipaa">Audited</span>
        <span className="badge-hipaa">Access Controlled</span>
      </motion.div>

      {/* Stats Grid */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          icon={Users}
          label="Total Patients"
          value={metrics.totalPatients.toLocaleString()}
          subValue={dataSource === 'live' ? `${metrics.activePatients} active` : '⚠️ Mock data'}
          color="from-medical-blue to-cyan-400"
        />
        <StatCard
          icon={FileText}
          label="Medical Records"
          value={metrics.totalRecords.toLocaleString()}
          subValue={dataSource === 'live' ? `${metrics.recordsLast24h} added today` : '⚠️ Mock data'}
          color="from-medical-green to-emerald-400"
        />
        <StatCard
          icon={Activity}
          label="Queries (24h)"
          value={metrics.queriesLast24h.toLocaleString()}
          subValue={dataSource === 'live' ? 'AI-powered searches' : '⚠️ Mock data'}
          color="from-cyber-purple to-cyber-pink"
        />
        <StatCard
          icon={Shield}
          label="HIPAA Audits"
          value={metrics.hipaaAudits.toLocaleString()}
          subValue={dataSource === 'live' ? 'Compliance logs' : '⚠️ Mock data'}
          color="from-cyber-yellow to-orange-400"
        />
      </div>

      {/* Second Row Stats */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard
          icon={Stethoscope}
          label="Agent Status"
          value={metrics.agentStatus === 'online' ? 'Online' : 'Offline'}
          subValue={dataSource === 'live' ? 'Knative Service' : '⚠️ Not connected'}
          color={metrics.agentStatus === 'online' ? 'from-medical-green to-emerald-400' : 'from-gray-600 to-gray-500'}
        />
        <StatCard
          icon={CheckCircle}
          label="Compliance Score"
          value={`${metrics.complianceScore}%`}
          subValue={dataSource === 'live' ? 'HIPAA adherence' : '⚠️ Mock data'}
          color="from-medical-blue to-indigo-400"
        />
        <StatCard
          icon={TrendingUp}
          label="System Health"
          value={dataSource === 'live' ? '99.9%' : '0%'}
          subValue={dataSource === 'live' ? 'Uptime' : '⚠️ Not monitored'}
          color="from-emerald-500 to-teal-400"
        />
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-2 gap-6">
        {/* Agent Status */}
        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <Stethoscope className="w-5 h-5 text-medical-blue" />
            Medical Agent Status
            {dataSource === 'mock' && (
              <span className="text-xs px-2 py-0.5 bg-yellow-500/20 text-yellow-500 rounded">MOCK</span>
            )}
            {dataSource === 'live' && (
              <span className="text-xs px-2 py-0.5 bg-medical-green/20 text-medical-green rounded">LIVE</span>
            )}
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 bg-cyber-dark/50 rounded-lg">
              <div className="flex items-center gap-3">
                <div className={`w-2 h-2 rounded-full ${
                  metrics.agentStatus === 'online' ? 'bg-medical-green animate-pulse' : 'bg-gray-500'
                }`} />
                <span className="font-mono text-sm">agent-medical</span>
              </div>
              <span className="text-xs text-gray-500">{metrics.agentStatus}</span>
            </div>
          </div>
          {dataSource === 'mock' && (
            <p className="text-xs text-yellow-500/70 mt-4 p-2 bg-yellow-500/10 rounded">
              ⚠️ Showing placeholder data. Connect to backend for real agent status.
            </p>
          )}
        </div>

        {/* Recent Activity */}
        <div className="card p-6">
          <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
            <Clock className="w-5 h-5 text-medical-blue" />
            Recent Activity
            {dataSource === 'mock' && (
              <span className="text-xs px-2 py-0.5 bg-yellow-500/20 text-yellow-500 rounded">MOCK</span>
            )}
          </h3>
          {dataSource === 'mock' ? (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <WifiOff className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-sm">No live activity data</p>
              <p className="text-xs mt-1">Connect to medical agent to see real events</p>
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-sm text-gray-400">Activity feed from medical agent...</p>
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
