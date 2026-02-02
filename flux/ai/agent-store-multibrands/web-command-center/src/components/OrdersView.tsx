'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, formatCurrency, formatTimeAgo, getOrderStatusColor } from '@/lib/utils'
import { BRANDS, OrderStatus } from '@/types/store'
import { motion } from 'framer-motion'
import {
  ShoppingCart,
  Search,
  Package,
  Truck,
  CheckCircle,
  XCircle,
  Clock,
  ChevronDown,
} from 'lucide-react'
import { useState } from 'react'

const statusIcons: Record<OrderStatus, typeof Package> = {
  pending: Clock,
  confirmed: CheckCircle,
  processing: Package,
  shipped: Truck,
  delivered: CheckCircle,
  cancelled: XCircle,
}

const statusLabels: Record<OrderStatus, string> = {
  pending: 'Pendente',
  confirmed: 'Confirmado',
  processing: 'Processando',
  shipped: 'Enviado',
  delivered: 'Entregue',
  cancelled: 'Cancelado',
}

export function OrdersView() {
  const { orders, selectedBrand, updateOrderStatus } = useStoreAppStore()
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<OrderStatus | 'all'>('all')
  
  const filteredOrders = orders
    .filter(o => selectedBrand === 'all' || o.brand === selectedBrand)
    .filter(o => statusFilter === 'all' || o.status === statusFilter)
    .filter(o => 
      searchTerm === '' ||
      o.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      o.customerPhone.includes(searchTerm)
    )

  const stats = {
    pending: orders.filter(o => o.status === 'pending').length,
    processing: orders.filter(o => o.status === 'processing' || o.status === 'confirmed').length,
    shipped: orders.filter(o => o.status === 'shipped').length,
    delivered: orders.filter(o => o.status === 'delivered').length,
  }

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Pedidos</h1>
          <p className="text-gray-500">Gerencie todos os pedidos da loja</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: 'Pendentes', value: stats.pending, color: 'store-yellow', icon: Clock },
          { label: 'Em Processo', value: stats.processing, color: 'store-blue', icon: Package },
          { label: 'Enviados', value: stats.shipped, color: 'store-orange', icon: Truck },
          { label: 'Entregues', value: stats.delivered, color: 'store-green', icon: CheckCircle },
        ].map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="store-card p-4"
          >
            <div className="flex items-center gap-3">
              <div className={cn("p-2 rounded-lg", `bg-${stat.color}/10`)}>
                <stat.icon className={cn("w-5 h-5", `text-${stat.color}`)} />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stat.value}</p>
                <p className="text-sm text-gray-500">{stat.label}</p>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Buscar por ID ou telefone..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-store-gray/50 border border-store-purple/20 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-store-purple/50"
          />
        </div>
        
        <div className="flex gap-2">
          {(['all', 'pending', 'processing', 'shipped', 'delivered'] as const).map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={cn(
                "px-3 py-2 text-xs rounded-lg transition-all",
                statusFilter === status
                  ? "bg-store-purple/20 text-store-purple border border-store-purple/50"
                  : "bg-store-gray/30 text-gray-400 hover:text-white"
              )}
            >
              {status === 'all' ? 'Todos' : statusLabels[status]}
            </button>
          ))}
        </div>
      </div>

      {/* Orders Table */}
      <div className="store-card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-store-purple/20">
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Pedido</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Cliente</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Marca</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Items</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Total</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-gray-500 uppercase">Data</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-store-purple/10">
            {filteredOrders.map((order) => {
              const brand = BRANDS[order.brand]
              const StatusIcon = statusIcons[order.status]
              
              return (
                <motion.tr
                  key={order.id}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="hover:bg-store-purple/5 transition-colors"
                >
                  <td className="px-4 py-3">
                    <span className="font-mono text-sm text-white">{order.id}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-sm text-gray-400">{order.customerPhone}</span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{brand.emoji}</span>
                      <span className={cn("text-sm", getBrandColor(order.brand))}>
                        {brand.name}
                      </span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-sm text-gray-400">
                      {order.items.length} item(s)
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-sm font-bold text-store-green">
                      {formatCurrency(order.total)}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="relative inline-block">
                      <select
                        value={order.status}
                        onChange={(e) => updateOrderStatus(order.id, e.target.value as OrderStatus)}
                        className={cn(
                          "appearance-none pl-8 pr-8 py-1.5 rounded-lg text-xs font-medium cursor-pointer",
                          getOrderStatusColor(order.status),
                          "bg-transparent border-0 focus:outline-none focus:ring-2 focus:ring-store-purple/50"
                        )}
                      >
                        {Object.entries(statusLabels).map(([value, label]) => (
                          <option key={value} value={value} className="bg-store-dark">
                            {label}
                          </option>
                        ))}
                      </select>
                      <StatusIcon className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none" />
                      <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 w-3 h-3 pointer-events-none opacity-50" />
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-xs text-gray-500">
                      {formatTimeAgo(order.createdAt)}
                    </span>
                  </td>
                </motion.tr>
              )
            })}
          </tbody>
        </table>
        
        {filteredOrders.length === 0 && (
          <div className="p-8 text-center">
            <ShoppingCart className="w-10 h-10 text-gray-600 mx-auto mb-3" />
            <p className="text-sm text-gray-500">Nenhum pedido encontrado</p>
          </div>
        )}
      </div>
    </div>
  )
}
