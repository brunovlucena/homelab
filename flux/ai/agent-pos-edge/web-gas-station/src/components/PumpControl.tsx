'use client'

import { useGasStationStore, Pump } from '@/store/gasStationStore'
import { cn, formatLiters, formatCurrency } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Gauge, Power, Wrench, AlertCircle, Clock, Droplets } from 'lucide-react'

function PumpCard({ pump }: { pump: Pump }) {
  const { transactions } = useGasStationStore()
  const activeTransaction = transactions.find(t => t.pumpId === pump.id && t.status === 'in_progress')
  
  const statusConfig = {
    idle: { color: 'fuel-amber', label: 'Disponível', icon: Gauge },
    active: { color: 'fuel-green', label: 'Em Uso', icon: Droplets },
    error: { color: 'fuel-red', label: 'Erro', icon: AlertCircle },
    maintenance: { color: 'fuel-blue', label: 'Manutenção', icon: Wrench },
  }
  
  const config = statusConfig[pump.status]

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn(
        'fuel-card p-6 transition-all',
        pump.status === 'active' && 'border-fuel-green/50 shadow-fuel',
        pump.status === 'error' && 'border-fuel-red/50 shadow-alert'
      )}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className={cn(
            'w-16 h-16 rounded-2xl flex items-center justify-center text-2xl font-bold',
            `bg-${config.color}/10 border-2 border-${config.color}/30 text-${config.color}`
          )}>
            {pump.number}
          </div>
          <div>
            <h3 className="text-lg font-bold text-white">Bomba {pump.number}</h3>
            <p className="text-sm text-gray-400 capitalize">{pump.fuelType}</p>
          </div>
        </div>
        <div className={cn('pump-indicator w-4 h-4', `pump-${pump.status === 'active' ? 'active' : pump.status === 'error' ? 'error' : 'idle'}`)} />
      </div>

      {/* Status */}
      <div className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg mb-4',
        `bg-${config.color}/10 border border-${config.color}/30`
      )}>
        <config.icon className={cn('w-4 h-4', `text-${config.color}`)} />
        <span className={cn('text-sm font-medium', `text-${config.color}`)}>{config.label}</span>
        {pump.status === 'active' && (
          <span className="ml-auto text-sm font-mono text-fuel-green animate-pulse">● LIVE</span>
        )}
      </div>

      {/* Active Transaction */}
      {activeTransaction && (
        <div className="p-4 rounded-lg bg-fuel-green/10 border border-fuel-green/30 mb-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">Abastecimento em andamento</span>
            <Clock className="w-4 h-4 text-fuel-green animate-pulse" />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div>
              <p className="text-xs text-gray-500">Litros</p>
              <p className="text-xl font-bold font-mono text-fuel-green">{formatLiters(activeTransaction.liters)}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Valor</p>
              <p className="text-xl font-bold font-mono text-fuel-amber">{formatCurrency(activeTransaction.amount)}</p>
            </div>
          </div>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 mb-4">
        <div>
          <p className="text-xs text-gray-500">Total Vendido</p>
          <p className="text-lg font-mono text-white">{formatLiters(pump.totalDispensed)}</p>
        </div>
        <div>
          <p className="text-xs text-gray-500">Última Manutenção</p>
          <p className="text-lg font-mono text-white">{pump.lastMaintenance}</p>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2 pt-4 border-t border-fuel-gray/50">
        {pump.status === 'idle' && (
          <button className="flex-1 fuel-button text-sm">
            <Power className="w-4 h-4 inline mr-2" />
            Liberar
          </button>
        )}
        {pump.status === 'active' && (
          <button className="flex-1 fuel-button-danger text-sm">
            <Power className="w-4 h-4 inline mr-2" />
            Parar
          </button>
        )}
        <button className="flex-1 px-4 py-2 rounded-lg bg-fuel-gray/50 text-gray-400 hover:bg-fuel-gray transition-colors text-sm">
          <Wrench className="w-4 h-4 inline mr-2" />
          Manutenção
        </button>
      </div>
    </motion.div>
  )
}

export function PumpControl() {
  const { pumps } = useGasStationStore()

  const activePumps = pumps.filter(p => p.status === 'active').length
  const idlePumps = pumps.filter(p => p.status === 'idle').length
  const maintenancePumps = pumps.filter(p => p.status === 'maintenance').length

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Controle de Bombas</h1>
          <p className="text-gray-400">Gerenciamento e monitoramento de bombas</p>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Total de Bombas</p>
          <p className="text-2xl font-bold font-mono text-white">{pumps.length}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-green/30">
          <p className="text-sm text-gray-400">Em Operação</p>
          <p className="text-2xl font-bold font-mono text-fuel-green">{activePumps}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-amber/30">
          <p className="text-sm text-gray-400">Disponíveis</p>
          <p className="text-2xl font-bold font-mono text-fuel-amber">{idlePumps}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-blue/30">
          <p className="text-sm text-gray-400">Em Manutenção</p>
          <p className="text-2xl font-bold font-mono text-fuel-blue">{maintenancePumps}</p>
        </div>
      </div>

      {/* Pumps Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {pumps.map((pump) => (
          <PumpCard key={pump.id} pump={pump} />
        ))}
      </div>
    </div>
  )
}
