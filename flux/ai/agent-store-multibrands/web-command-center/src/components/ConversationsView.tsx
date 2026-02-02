'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, formatTimeAgo, getStatusColor } from '@/lib/utils'
import { BRANDS } from '@/types/store'
import { motion } from 'framer-motion'
import {
  MessageSquare,
  Search,
  Filter,
  AlertTriangle,
  Clock,
  User,
  Send,
  MoreVertical,
  Phone,
  ChevronRight,
} from 'lucide-react'
import { useState } from 'react'

export function ConversationsView() {
  const { 
    conversations, 
    selectedBrand,
    selectedConversationId,
    selectConversation,
  } = useStoreAppStore()
  
  const [filter, setFilter] = useState<'all' | 'active' | 'escalated'>('all')
  const [searchTerm, setSearchTerm] = useState('')
  
  const filteredConversations = conversations
    .filter(c => selectedBrand === 'all' || c.brand === selectedBrand)
    .filter(c => {
      if (filter === 'active') return c.state === 'active' || c.state === 'waiting'
      if (filter === 'escalated') return c.state === 'escalated'
      return true
    })
    .filter(c => 
      searchTerm === '' || 
      c.customerPhone.includes(searchTerm) ||
      c.customerName?.toLowerCase().includes(searchTerm.toLowerCase())
    )
  
  const selectedConversation = conversations.find(c => c.id === selectedConversationId)

  return (
    <div className="flex h-full">
      {/* Conversation List */}
      <div className="w-96 border-r border-store-purple/20 flex flex-col">
        {/* Header */}
        <div className="p-4 border-b border-store-purple/20">
          <h2 className="text-lg font-bold text-white mb-4">Conversas</h2>
          
          {/* Search */}
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Buscar por nome ou telefone..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-store-gray/50 border border-store-purple/20 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-store-purple/50"
            />
          </div>
          
          {/* Filters */}
          <div className="flex gap-2">
            {(['all', 'active', 'escalated'] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={cn(
                  "px-3 py-1.5 text-xs rounded-lg transition-all",
                  filter === f
                    ? "bg-store-purple/20 text-store-purple border border-store-purple/50"
                    : "bg-store-gray/30 text-gray-400 hover:text-white"
                )}
              >
                {f === 'all' && 'Todas'}
                {f === 'active' && 'Ativas'}
                {f === 'escalated' && 'Escaladas'}
                <span className="ml-1 text-[10px]">
                  ({conversations.filter(c => {
                    if (f === 'active') return c.state === 'active' || c.state === 'waiting'
                    if (f === 'escalated') return c.state === 'escalated'
                    return true
                  }).length})
                </span>
              </button>
            ))}
          </div>
        </div>
        
        {/* List */}
        <div className="flex-1 overflow-y-auto">
          {filteredConversations.map((conv) => {
            const brand = BRANDS[conv.brand]
            const isSelected = selectedConversationId === conv.id
            
            return (
              <motion.div
                key={conv.id}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                onClick={() => selectConversation(conv.id)}
                className={cn(
                  "p-4 border-b border-store-purple/10 cursor-pointer transition-all",
                  isSelected 
                    ? "bg-store-purple/10 border-l-2 border-l-store-purple"
                    : "hover:bg-store-gray/30"
                )}
              >
                <div className="flex items-start gap-3">
                  <div className="relative">
                    <span className="text-2xl">{brand.sellerAvatar}</span>
                    <div className={cn(
                      "absolute -bottom-1 -right-1 w-3 h-3 rounded-full border-2 border-store-dark",
                      conv.state === 'escalated' ? 'bg-store-red' :
                      conv.state === 'active' ? 'bg-store-green' : 'bg-store-yellow'
                    )} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-white truncate">
                        {conv.customerName || 'Cliente'}
                      </span>
                      {conv.state === 'escalated' && (
                        <AlertTriangle className="w-3 h-3 text-store-red" />
                      )}
                    </div>
                    <p className="text-xs text-gray-500 truncate">
                      {conv.messages[conv.messages.length - 1]?.content || 'Sem mensagens'}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-[10px] text-gray-600">
                        {formatTimeAgo(conv.lastMessageAt)}
                      </span>
                      <span className={cn("text-[10px]", getBrandColor(conv.brand))}>
                        • {brand.name}
                      </span>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-600" />
                </div>
              </motion.div>
            )
          })}
          
          {filteredConversations.length === 0 && (
            <div className="p-8 text-center">
              <MessageSquare className="w-10 h-10 text-gray-600 mx-auto mb-3" />
              <p className="text-sm text-gray-500">Nenhuma conversa encontrada</p>
            </div>
          )}
        </div>
      </div>
      
      {/* Conversation Detail */}
      <div className="flex-1 flex flex-col">
        {selectedConversation ? (
          <>
            {/* Chat Header */}
            <div className="p-4 border-b border-store-purple/20 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{BRANDS[selectedConversation.brand].sellerAvatar}</span>
                <div>
                  <h3 className="text-lg font-bold text-white">
                    {selectedConversation.customerName || 'Cliente'}
                  </h3>
                  <div className="flex items-center gap-2 text-sm text-gray-500">
                    <Phone className="w-3 h-3" />
                    <span>{selectedConversation.customerPhone}</span>
                    <span className={cn(
                      "px-2 py-0.5 text-xs rounded-full",
                      selectedConversation.state === 'escalated'
                        ? 'bg-store-red/10 text-store-red'
                        : selectedConversation.state === 'active'
                        ? 'bg-store-green/10 text-store-green'
                        : 'bg-store-yellow/10 text-store-yellow'
                    )}>
                      {selectedConversation.state}
                    </span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button className="p-2 rounded-lg hover:bg-store-gray/50 text-gray-400 hover:text-white transition-colors">
                  <MoreVertical className="w-5 h-5" />
                </button>
              </div>
            </div>
            
            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {selectedConversation.messages.map((msg) => (
                <div
                  key={msg.id}
                  className={cn(
                    "flex",
                    msg.role === 'customer' ? 'justify-start' : 'justify-end'
                  )}
                >
                  <div className={cn(
                    "max-w-[70%] p-3 rounded-2xl",
                    msg.role === 'customer'
                      ? 'bg-store-gray/50 rounded-bl-sm'
                      : msg.role === 'ai'
                      ? 'bg-store-purple/20 border border-store-purple/30 rounded-br-sm'
                      : 'bg-store-blue/20 border border-store-blue/30 rounded-br-sm'
                  )}>
                    <p className="text-sm text-white">{msg.content}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-[10px] text-gray-500">
                        {new Date(msg.timestamp).toLocaleTimeString('pt-BR', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </span>
                      {msg.role === 'ai' && msg.metadata?.tokensUsed && (
                        <span className="text-[10px] text-gray-600">
                          • {msg.metadata.tokensUsed} tokens
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
            
            {/* Input */}
            <div className="p-4 border-t border-store-purple/20">
              <div className="flex items-center gap-3">
                <input
                  type="text"
                  placeholder="Digite uma mensagem como vendedor humano..."
                  className="flex-1 px-4 py-3 bg-store-gray/50 border border-store-purple/20 rounded-xl text-sm text-white placeholder-gray-500 focus:outline-none focus:border-store-purple/50"
                />
                <button className="p-3 rounded-xl bg-store-purple hover:bg-store-purple/80 transition-colors">
                  <Send className="w-5 h-5 text-white" />
                </button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <MessageSquare className="w-16 h-16 text-gray-700 mx-auto mb-4" />
              <h3 className="text-xl font-bold text-gray-500 mb-2">
                Selecione uma conversa
              </h3>
              <p className="text-sm text-gray-600">
                Escolha uma conversa à esquerda para ver os detalhes
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
