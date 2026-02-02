'use client'

import { useGasStationStore, Tank } from '@/store/gasStationStore'
import { cn, formatLiters, formatPercent } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Database, Thermometer, Clock, AlertTriangle, RefreshCw, TrendingDown } from 'lucide-react'

function TankCard({ tank }: { tank: Tank }) {
  const percent = formatPercent(tank.currentLevel, tank.capacity)
  const color = tank.status === 'critical' ? 'fuel-red' : tank.status === 'low' ? 'fuel-amber' : 'fuel-green'
  
  const fuelColorMap: Record<string, string> = {
    gasoline: 'from-fuel-green to-fuel-green/60',
    diesel: 'from-diesel to-diesel/60',
    premium: 'from-premium to-premium/60',
    ethanol: 'from-fuel-lime to-fuel-lime/60',
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="fuel-card p-6 hover:border-fuel-green/50 transition-all"
    >
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-xl font-bold text-white">{tank.name}</h3>
          <p className="text-sm text-gray-400 capitalize">{tank.fuelType}</p>
        </div>
        <span className={cn('status-badge', `status-${tank.status === 'normal' ? 'online' : tank.status === 'low' ? 'warning' : 'offline'}`)}>
          {tank.status}
        </span>
      </div>

      {/* Tank Visualization */}
      <div className="relative h-48 mb-4">
        <div className="absolute inset-0 rounded-xl bg-fuel-gray/30 border-2 border-fuel-gray overflow-hidden">
          <div 
            className={cn(
              'absolute bottom-0 left-0 right-0 transition-all duration-1000',
              `bg-gradient-to-t ${fuelColorMap[tank.fuelType]}`
            )}
            style={{ height: `${percent}%` }}
          >
            {/* Animated waves */}
            <div className="absolute top-0 left-0 right-0 h-2 opacity-50">
              <div className="h-full bg-white/20 animate-pulse rounded-t-full" />
            </div>
          </div>
          
          {/* Level markers */}
          {[25, 50, 75].map((level) => (
            <div
              key={level}
              className="absolute left-0 right-0 border-t border-dashed border-gray-600/50"
              style={{ bottom: `${level}%` }}
            >
              <span className="absolute -top-2.5 right-2 text-xs text-gray-500">{level}%</span>
            </div>
          ))}
        </div>
        
        {/* Percentage overlay */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <p className={cn('text-4xl font-bold font-mono', `text-${color}`)}>{percent}%</p>
            <p className="text-sm text-gray-400">{formatLiters(tank.currentLevel)}</p>
          </div>
        </div>
      </div>

      {/* Tank Info */}
      <div className="grid grid-cols-2 gap-4">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-gray-500" />
          <div>
            <p className="text-xs text-gray-500">Capacidade</p>
            <p className="text-sm font-mono text-white">{formatLiters(tank.capacity)}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Thermometer className="w-4 h-4 text-gray-500" />
          <div>
            <p className="text-xs text-gray-500">Temperatura</p>
            <p className="text-sm font-mono text-white">{tank.temperature}°C</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <TrendingDown className="w-4 h-4 text-gray-500" />
          <div>
            <p className="text-xs text-gray-500">Consumo/h</p>
            <p className="text-sm font-mono text-fuel-amber">~120L</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-gray-500" />
          <div>
            <p className="text-xs text-gray-500">Autonomia</p>
            <p className="text-sm font-mono text-white">{Math.round(tank.currentLevel / 120)}h</p>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2 mt-4 pt-4 border-t border-fuel-gray/50">
        <button className="flex-1 fuel-button text-sm">
          <RefreshCw className="w-4 h-4 inline mr-2" />
          Atualizar
        </button>
        {tank.status !== 'normal' && (
          <button className="flex-1 fuel-button-danger text-sm">
            <AlertTriangle className="w-4 h-4 inline mr-2" />
            Agendar Reabastecimento
          </button>
        )}
      </div>
    </motion.div>
  )
}

export function TankMonitor() {
  const { tanks } = useGasStationStore()

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Monitoramento de Tanques</h1>
          <p className="text-gray-400">Níveis e status de todos os tanques</p>
        </div>
        <button className="fuel-button">
          <RefreshCw className="w-4 h-4 inline mr-2" />
          Atualizar Todos
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Total Armazenado</p>
          <p className="text-2xl font-bold font-mono text-white">
            {formatLiters(tanks.reduce((sum, t) => sum + t.currentLevel, 0))}
          </p>
        </div>
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Capacidade Total</p>
          <p className="text-2xl font-bold font-mono text-white">
            {formatLiters(tanks.reduce((sum, t) => sum + t.capacity, 0))}
          </p>
        </div>
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Tanques OK</p>
          <p className="text-2xl font-bold font-mono text-fuel-green">
            {tanks.filter(t => t.status === 'normal').length}/{tanks.length}
          </p>
        </div>
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Alertas Ativos</p>
          <p className="text-2xl font-bold font-mono text-fuel-red">
            {tanks.filter(t => t.status === 'critical' || t.status === 'low').length}
          </p>
        </div>
      </div>

      {/* Tank Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {tanks.map((tank) => (
          <TankCard key={tank.id} tank={tank} />
        ))}
      </div>
    </div>
  )
}
