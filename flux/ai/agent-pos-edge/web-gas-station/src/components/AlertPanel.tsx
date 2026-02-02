'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { AlertTriangle, AlertCircle, Info, Check, Bell, Filter } from 'lucide-react'
import { useState } from 'react'

export function AlertPanel() {
  const { alerts, acknowledgeAlert } = useGasStationStore()
  const [filter, setFilter] = useState<'all' | 'critical' | 'warning' | 'info'>('all')

  const filteredAlerts = alerts.filter(a => filter === 'all' || a.type === filter)
  
  const criticalCount = alerts.filter(a => a.type === 'critical' && !a.acknowledged).length
  const warningCount = alerts.filter(a => a.type === 'warning' && !a.acknowledged).length
  const infoCount = alerts.filter(a => a.type === 'info' && !a.acknowledged).length

  const getAlertIcon = (type: string) => {
    switch (type) {
      case 'critical': return AlertTriangle
      case 'warning': return AlertCircle
      default: return Info
    }
  }

  const getAlertColor = (type: string) => {
    switch (type) {
      case 'critical': return 'fuel-red'
      case 'warning': return 'fuel-amber'
      default: return 'fuel-blue'
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Central de Alertas</h1>
          <p className="text-gray-400">Monitoramento de eventos e notificações</p>
        </div>
        <button className="fuel-button">
          <Check className="w-4 h-4 inline mr-2" />
          Confirmar Todos
        </button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Total de Alertas</p>
          <p className="text-2xl font-bold font-mono text-white">{alerts.length}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-red/30">
          <p className="text-sm text-gray-400">Críticos</p>
          <p className="text-2xl font-bold font-mono text-fuel-red">{criticalCount}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-amber/30">
          <p className="text-sm text-gray-400">Avisos</p>
          <p className="text-2xl font-bold font-mono text-fuel-amber">{warningCount}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-blue/30">
          <p className="text-sm text-gray-400">Informativos</p>
          <p className="text-2xl font-bold font-mono text-fuel-blue">{infoCount}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        {[
          { id: 'all', label: 'Todos' },
          { id: 'critical', label: 'Críticos', color: 'fuel-red' },
          { id: 'warning', label: 'Avisos', color: 'fuel-amber' },
          { id: 'info', label: 'Info', color: 'fuel-blue' },
        ].map((f) => (
          <button
            key={f.id}
            onClick={() => setFilter(f.id as typeof filter)}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              filter === f.id 
                ? f.color ? `bg-${f.color}/20 text-${f.color} border border-${f.color}/30` : 'bg-fuel-green/20 text-fuel-green border border-fuel-green/30'
                : 'bg-fuel-gray/50 text-gray-400 hover:bg-fuel-gray'
            )}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Alerts List */}
      <div className="space-y-3">
        {filteredAlerts.length === 0 ? (
          <div className="fuel-card p-12 text-center">
            <Bell className="w-12 h-12 text-gray-600 mx-auto mb-4" />
            <p className="text-gray-400">Nenhum alerta encontrado</p>
          </div>
        ) : (
          filteredAlerts.map((alert, index) => {
            const Icon = getAlertIcon(alert.type)
            const color = getAlertColor(alert.type)
            
            return (
              <motion.div
                key={alert.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.05 }}
                className={cn(
                  'fuel-card p-4 transition-all',
                  `border-${color}/30`,
                  alert.acknowledged && 'opacity-50'
                )}
              >
                <div className="flex items-start gap-4">
                  <div className={cn(
                    'p-3 rounded-xl',
                    `bg-${color}/10 border border-${color}/30`
                  )}>
                    <Icon className={cn('w-6 h-6', `text-${color}`)} />
                  </div>
                  
                  <div className="flex-1">
                    <div className="flex items-start justify-between">
                      <div>
                        <h3 className="font-bold text-white">{alert.title}</h3>
                        <p className="text-sm text-gray-400 mt-1">{alert.message}</p>
                      </div>
                      <span className={cn('status-badge', `status-${alert.type === 'critical' ? 'offline' : alert.type === 'warning' ? 'warning' : 'online'}`)}>
                        {alert.type}
                      </span>
                    </div>
                    
                    <div className="flex items-center justify-between mt-4">
                      <div className="flex items-center gap-4 text-sm text-gray-500">
                        <span>Fonte: {alert.source}</span>
                        <span>{getTimeSince(alert.timestamp)}</span>
                      </div>
                      
                      {!alert.acknowledged && (
                        <button
                          onClick={() => acknowledgeAlert(alert.id)}
                          className="fuel-button text-sm"
                        >
                          <Check className="w-4 h-4 inline mr-2" />
                          Confirmar
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </motion.div>
            )
          })
        )}
      </div>
    </div>
  )
}
