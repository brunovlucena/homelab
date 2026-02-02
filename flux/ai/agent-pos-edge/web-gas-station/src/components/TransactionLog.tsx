'use client'

import { useGasStationStore } from '@/store/gasStationStore'
import { cn, formatCurrency, formatLiters, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Receipt, Search, Filter, Download, CreditCard, Banknote, Smartphone } from 'lucide-react'
import { useState } from 'react'

export function TransactionLog() {
  const { transactions, pumps } = useGasStationStore()
  const [filter, setFilter] = useState<'all' | 'completed' | 'in_progress'>('all')
  const [search, setSearch] = useState('')

  const filteredTransactions = transactions.filter(t => {
    if (filter !== 'all' && t.status !== filter) return false
    if (search && !t.id.includes(search) && !t.fuelType.includes(search.toLowerCase())) return false
    return true
  })

  const totalRevenue = transactions.filter(t => t.status === 'completed').reduce((sum, t) => sum + t.amount, 0)
  const totalLiters = transactions.filter(t => t.status === 'completed').reduce((sum, t) => sum + t.liters, 0)
  const avgTransaction = totalRevenue / transactions.filter(t => t.status === 'completed').length || 0

  const paymentIcon = (method?: string) => {
    switch (method) {
      case 'card': return <CreditCard className="w-4 h-4" />
      case 'cash': return <Banknote className="w-4 h-4" />
      case 'app': return <Smartphone className="w-4 h-4" />
      default: return null
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Log de Transações</h1>
          <p className="text-gray-400">Histórico de vendas e abastecimentos</p>
        </div>
        <button className="fuel-button">
          <Download className="w-4 h-4 inline mr-2" />
          Exportar
        </button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="fuel-card p-4">
          <p className="text-sm text-gray-400">Total Transações</p>
          <p className="text-2xl font-bold font-mono text-white">{transactions.length}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-green/30">
          <p className="text-sm text-gray-400">Receita Total</p>
          <p className="text-2xl font-bold font-mono text-fuel-green">{formatCurrency(totalRevenue)}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-blue/30">
          <p className="text-sm text-gray-400">Litros Vendidos</p>
          <p className="text-2xl font-bold font-mono text-fuel-blue">{formatLiters(totalLiters)}</p>
        </div>
        <div className="fuel-card p-4 border-fuel-amber/30">
          <p className="text-sm text-gray-400">Ticket Médio</p>
          <p className="text-2xl font-bold font-mono text-fuel-amber">{formatCurrency(avgTransaction)}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col md:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Buscar por ID ou combustível..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-fuel-gray/50 border border-fuel-gray rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-fuel-green/50"
          />
        </div>
        <div className="flex gap-2">
          {['all', 'completed', 'in_progress'].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f as typeof filter)}
              className={cn(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === f 
                  ? 'bg-fuel-green/20 text-fuel-green border border-fuel-green/30' 
                  : 'bg-fuel-gray/50 text-gray-400 hover:bg-fuel-gray'
              )}
            >
              {f === 'all' ? 'Todas' : f === 'completed' ? 'Concluídas' : 'Em Andamento'}
            </button>
          ))}
        </div>
      </div>

      {/* Transactions Table */}
      <div className="fuel-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-fuel-gray/50">
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Bomba</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Combustível</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Litros</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Valor</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Pagamento</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase tracking-wider">Horário</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-fuel-gray/30">
              {filteredTransactions.map((txn, index) => (
                <motion.tr 
                  key={txn.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.05 }}
                  className="hover:bg-fuel-gray/30 transition-colors"
                >
                  <td className="px-4 py-3 text-sm font-mono text-white">{txn.id}</td>
                  <td className="px-4 py-3 text-sm text-white">
                    Bomba {pumps.find(p => p.id === txn.pumpId)?.number}
                  </td>
                  <td className="px-4 py-3 text-sm text-white capitalize">{txn.fuelType}</td>
                  <td className="px-4 py-3 text-sm font-mono text-fuel-blue">{formatLiters(txn.liters)}</td>
                  <td className="px-4 py-3 text-sm font-mono text-fuel-green">{formatCurrency(txn.amount)}</td>
                  <td className="px-4 py-3">
                    {txn.paymentMethod ? (
                      <div className="flex items-center gap-2 text-gray-400">
                        {paymentIcon(txn.paymentMethod)}
                        <span className="text-sm capitalize">{txn.paymentMethod}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-gray-500">-</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <span className={cn(
                      'status-badge',
                      txn.status === 'completed' ? 'status-online' : 
                      txn.status === 'in_progress' ? 'status-warning' : 'status-offline'
                    )}>
                      {txn.status === 'completed' ? 'Concluída' : 
                       txn.status === 'in_progress' ? 'Em andamento' : 'Cancelada'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-400">{getTimeSince(txn.startTime)}</td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
