'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, getBrandBgColor, formatPercent, formatDuration } from '@/lib/utils'
import { BRANDS, SellerStatus } from '@/types/store'
import { motion } from 'framer-motion'
import {
  Users,
  MessageSquare,
  Clock,
  Star,
  Activity,
  TrendingUp,
  Circle,
  Settings,
  RefreshCw,
} from 'lucide-react'

export function SellersView() {
  const { sellers, updateSellerStatus, conversations } = useStoreAppStore()

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Vendedores AI</h1>
          <p className="text-gray-500">Monitore e gerencie seus agentes de vendas</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 rounded-lg bg-store-purple/20 text-store-purple hover:bg-store-purple/30 transition-colors">
          <RefreshCw className="w-4 h-4" />
          <span className="text-sm">Atualizar Status</span>
        </button>
      </div>

      {/* Seller Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {sellers.map((seller, index) => {
          const brand = BRANDS[seller.brand]
          const brandConversations = conversations.filter(c => c.brand === seller.brand)
          const activeConvs = brandConversations.filter(c => c.state === 'active' || c.state === 'waiting').length
          const escalatedConvs = brandConversations.filter(c => c.state === 'escalated').length
          
          return (
            <motion.div
              key={seller.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className={cn(
                "store-card store-card-hover overflow-hidden",
                `glow-${seller.brand}`
              )}
            >
              {/* Header */}
              <div className={cn(
                "p-4 bg-gradient-to-r",
                `from-${brand.color}/20 to-transparent`
              )}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <span className="text-4xl">{brand.sellerAvatar}</span>
                      <Circle 
                        className={cn(
                          "absolute -bottom-1 -right-1 w-4 h-4 border-2 border-store-dark",
                          seller.status === 'online' ? 'text-store-green' :
                          seller.status === 'busy' ? 'text-store-yellow' : 'text-gray-500'
                        )}
                        fill="currentColor"
                      />
                    </div>
                    <div>
                      <h3 className={cn("text-xl font-bold", getBrandColor(seller.brand))}>
                        {brand.sellerName}
                      </h3>
                      <p className="text-sm text-gray-500">{brand.name} Expert</p>
                    </div>
                  </div>
                  <span className="text-3xl">{brand.emoji}</span>
                </div>
              </div>
              
              {/* Stats */}
              <div className="p-4 space-y-4">
                <div className="grid grid-cols-2 gap-3">
                  <div className="p-3 rounded-lg bg-store-dark/50">
                    <div className="flex items-center gap-2 mb-1">
                      <MessageSquare className="w-4 h-4 text-store-purple" />
                      <span className="text-xs text-gray-500">Conversas</span>
                    </div>
                    <p className="text-xl font-bold text-white">{activeConvs}</p>
                  </div>
                  <div className="p-3 rounded-lg bg-store-dark/50">
                    <div className="flex items-center gap-2 mb-1">
                      <Clock className="w-4 h-4 text-store-blue" />
                      <span className="text-xs text-gray-500">Tempo Resp.</span>
                    </div>
                    <p className="text-xl font-bold text-white">
                      {seller.avgResponseTime.toFixed(1)}s
                    </p>
                  </div>
                  <div className="p-3 rounded-lg bg-store-dark/50">
                    <div className="flex items-center gap-2 mb-1">
                      <TrendingUp className="w-4 h-4 text-store-green" />
                      <span className="text-xs text-gray-500">Mensagens</span>
                    </div>
                    <p className="text-xl font-bold text-white">{seller.messagesHandled}</p>
                  </div>
                  <div className="p-3 rounded-lg bg-store-dark/50">
                    <div className="flex items-center gap-2 mb-1">
                      <Star className="w-4 h-4 text-store-yellow" />
                      <span className="text-xs text-gray-500">Satisfação</span>
                    </div>
                    <p className="text-xl font-bold text-white">
                      {formatPercent(seller.satisfaction)}
                    </p>
                  </div>
                </div>
                
                {/* Escalations Alert */}
                {escalatedConvs > 0 && (
                  <div className="p-3 rounded-lg bg-store-red/10 border border-store-red/30">
                    <div className="flex items-center gap-2">
                      <Activity className="w-4 h-4 text-store-red" />
                      <span className="text-sm text-store-red">
                        {escalatedConvs} escalação(ões) pendente(s)
                      </span>
                    </div>
                  </div>
                )}
                
                {/* Status Toggle */}
                <div className="flex items-center justify-between pt-2 border-t border-store-purple/10">
                  <span className="text-sm text-gray-500">Status</span>
                  <div className="flex gap-2">
                    {(['online', 'busy', 'offline'] as SellerStatus[]).map((status) => (
                      <button
                        key={status}
                        onClick={() => updateSellerStatus(seller.id, status)}
                        className={cn(
                          "px-3 py-1 text-xs rounded-lg transition-all",
                          seller.status === status
                            ? status === 'online' ? 'bg-store-green/20 text-store-green border border-store-green/50'
                            : status === 'busy' ? 'bg-store-yellow/20 text-store-yellow border border-store-yellow/50'
                            : 'bg-gray-500/20 text-gray-400 border border-gray-500/50'
                            : 'bg-store-gray/30 text-gray-500 hover:text-white'
                        )}
                      >
                        {status === 'online' && 'Online'}
                        {status === 'busy' && 'Ocupado'}
                        {status === 'offline' && 'Offline'}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </motion.div>
          )
        })}
      </div>

      {/* AI Performance Summary */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="store-card p-6"
      >
        <h2 className="text-lg font-bold text-white mb-4">Performance Geral dos Agentes</h2>
        <div className="grid grid-cols-4 gap-6">
          <div>
            <p className="text-sm text-gray-500 mb-1">Total de Mensagens Hoje</p>
            <p className="text-3xl font-bold text-white">
              {sellers.reduce((sum, s) => sum + s.messagesHandled, 0)}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Tempo Médio de Resposta</p>
            <p className="text-3xl font-bold text-store-blue">
              {(sellers.reduce((sum, s) => sum + s.avgResponseTime, 0) / sellers.length).toFixed(1)}s
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Taxa de Satisfação</p>
            <p className="text-3xl font-bold text-store-green">
              {formatPercent(sellers.reduce((sum, s) => sum + s.satisfaction, 0) / sellers.length)}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Agentes Online</p>
            <p className="text-3xl font-bold text-store-purple">
              {sellers.filter(s => s.status === 'online').length}/{sellers.length}
            </p>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
