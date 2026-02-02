'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Bell, 
  AlertTriangle, 
  AlertCircle, 
  Info, 
  CheckCircle,
  Filter,
  Check
} from 'lucide-react'
import type { Alert, AlertSeverity } from '@/types'

// Mock data
const mockAlerts: Alert[] = [
  {
    id: 'alert-001',
    type: 'system',
    severity: 'high',
    title: 'High Memory Usage',
    message: 'Media Agent memory usage exceeded 85% threshold',
    source: 'media-agent',
    timestamp: '2025-12-10T09:15:00Z',
    acknowledged: false,
  },
  {
    id: 'alert-002',
    type: 'security',
    severity: 'medium',
    title: 'Unusual Login Pattern',
    message: 'User bruno_lucena logged in from new location: Tokyo, Japan',
    source: 'auth-service',
    timestamp: '2025-12-10T08:30:00Z',
    acknowledged: false,
  },
  {
    id: 'alert-003',
    type: 'performance',
    severity: 'low',
    title: 'Voice Clone Queue Building',
    message: 'Voice Agent queue depth increased to 15 requests',
    source: 'voice-agent',
    timestamp: '2025-12-10T07:45:00Z',
    acknowledged: true,
    acknowledgedBy: 'admin',
    acknowledgedAt: '2025-12-10T08:00:00Z',
  },
  {
    id: 'alert-004',
    type: 'agent',
    severity: 'info',
    title: 'New Agent Deployed',
    message: 'Agent assistant deployed for user john_doe',
    source: 'command-center',
    timestamp: '2025-12-09T16:05:00Z',
    acknowledged: true,
    acknowledgedBy: 'system',
    acknowledgedAt: '2025-12-09T16:05:00Z',
  },
]

export function AlertsView() {
  const [selectedSeverity, setSelectedSeverity] = useState<AlertSeverity | 'all'>('all')
  const [showAcknowledged, setShowAcknowledged] = useState(true)

  const filteredAlerts = mockAlerts.filter(alert => {
    const matchesSeverity = selectedSeverity === 'all' || alert.severity === selectedSeverity
    const matchesAck = showAcknowledged || !alert.acknowledged
    return matchesSeverity && matchesAck
  })

  const unacknowledgedCount = mockAlerts.filter(a => !a.acknowledged).length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-gray-400" />
            <select
              value={selectedSeverity}
              onChange={(e) => setSelectedSeverity(e.target.value as AlertSeverity | 'all')}
              className="input-field"
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
              <option value="info">Info</option>
            </select>
          </div>
          <label className="flex items-center gap-2 text-sm text-gray-400">
            <input
              type="checkbox"
              checked={showAcknowledged}
              onChange={(e) => setShowAcknowledged(e.target.checked)}
              className="rounded border-gray-600 bg-cyber-dark text-cyber-purple focus:ring-cyber-purple"
            />
            Show acknowledged
          </label>
        </div>
        {unacknowledgedCount > 0 && (
          <button className="btn-primary flex items-center gap-2">
            <Check className="w-4 h-4" />
            Acknowledge All ({unacknowledgedCount})
          </button>
        )}
      </div>

      {/* Alert Stats */}
      <div className="grid grid-cols-5 gap-4">
        <AlertStatCard severity="critical" count={mockAlerts.filter(a => a.severity === 'critical').length} />
        <AlertStatCard severity="high" count={mockAlerts.filter(a => a.severity === 'high').length} />
        <AlertStatCard severity="medium" count={mockAlerts.filter(a => a.severity === 'medium').length} />
        <AlertStatCard severity="low" count={mockAlerts.filter(a => a.severity === 'low').length} />
        <AlertStatCard severity="info" count={mockAlerts.filter(a => a.severity === 'info').length} />
      </div>

      {/* Alerts List */}
      <div className="space-y-3">
        {filteredAlerts.map((alert) => (
          <AlertCard key={alert.id} alert={alert} />
        ))}
      </div>
    </div>
  )
}

function AlertStatCard({ severity, count }: { severity: AlertSeverity; count: number }) {
  const config: Record<AlertSeverity, { bg: string; icon: React.ElementType; label: string }> = {
    critical: { bg: 'from-red-600 to-red-800', icon: AlertCircle, label: 'Critical' },
    high: { bg: 'from-orange-500 to-orange-700', icon: AlertTriangle, label: 'High' },
    medium: { bg: 'from-yellow-500 to-yellow-700', icon: AlertTriangle, label: 'Medium' },
    low: { bg: 'from-blue-500 to-blue-700', icon: Info, label: 'Low' },
    info: { bg: 'from-gray-500 to-gray-700', icon: Info, label: 'Info' },
  }

  const { bg, icon: Icon, label } = config[severity]

  return (
    <div className="card p-4 text-center">
      <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${bg} flex items-center justify-center mx-auto mb-2`}>
        <Icon className="w-5 h-5 text-white" />
      </div>
      <p className="text-2xl font-bold">{count}</p>
      <p className="text-xs text-gray-400 uppercase">{label}</p>
    </div>
  )
}

function AlertCard({ alert }: { alert: Alert }) {
  const severityConfig: Record<AlertSeverity, { border: string; icon: React.ElementType; iconColor: string }> = {
    critical: { border: 'border-l-red-500', icon: AlertCircle, iconColor: 'text-red-500' },
    high: { border: 'border-l-orange-500', icon: AlertTriangle, iconColor: 'text-orange-500' },
    medium: { border: 'border-l-yellow-500', icon: AlertTriangle, iconColor: 'text-yellow-500' },
    low: { border: 'border-l-blue-500', icon: Info, iconColor: 'text-blue-500' },
    info: { border: 'border-l-gray-500', icon: Info, iconColor: 'text-gray-400' },
  }

  const { border, icon: Icon, iconColor } = severityConfig[alert.severity]

  return (
    <motion.div
      className={`card border-l-4 ${border} p-4 ${alert.acknowledged ? 'opacity-60' : ''}`}
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: alert.acknowledged ? 0.6 : 1, x: 0 }}
    >
      <div className="flex items-start gap-4">
        <Icon className={`w-5 h-5 ${iconColor} mt-0.5 flex-shrink-0`} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-1">
            <h4 className="font-medium">{alert.title}</h4>
            <span className="text-xs text-gray-500">{formatRelativeTime(alert.timestamp)}</span>
          </div>
          <p className="text-sm text-gray-400 mb-2">{alert.message}</p>
          <div className="flex items-center gap-4 text-xs text-gray-500">
            <span>Source: <span className="font-mono text-cyber-purple">{alert.source}</span></span>
            {alert.acknowledged && (
              <span className="flex items-center gap-1 text-cyber-green">
                <CheckCircle className="w-3 h-3" />
                Acknowledged by {alert.acknowledgedBy}
              </span>
            )}
          </div>
        </div>
        {!alert.acknowledged && (
          <button className="btn-secondary text-sm px-3 py-1">
            Acknowledge
          </button>
        )}
      </div>
    </motion.div>
  )
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  return `${Math.floor(diffMins / 1440)}d ago`
}
