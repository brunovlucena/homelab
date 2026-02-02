'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { Users, UserCheck, Coffee, UserX, ChefHat, CreditCard, Wrench } from 'lucide-react'

const roleIcons = {
  'manager': Users,
  'kitchen': ChefHat,
  'crew': Wrench,
  'cashier': CreditCard,
}

const roleLabels = {
  'manager': 'Gerente',
  'kitchen': 'Cozinha',
  'crew': 'Equipe',
  'cashier': 'Caixa',
}

export function StaffPanel() {
  const { staff, stations } = useMcdonaldsStore()

  const activeStaff = staff.filter(s => s.status === 'active').length
  const onBreak = staff.filter(s => s.status === 'break').length
  const totalOrders = staff.reduce((sum, s) => sum + s.ordersCompleted, 0)

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">Painel da Equipe</h1>
          <p className="text-gray-400">Gerenciamento de funcionários</p>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="mc-card p-4">
          <p className="text-sm text-gray-400">Total</p>
          <p className="text-2xl font-bold font-mono text-white">{staff.length}</p>
        </div>
        <div className="mc-card p-4 border-mc-green/30">
          <p className="text-sm text-gray-400">Ativos</p>
          <p className="text-2xl font-bold font-mono text-mc-green">{activeStaff}</p>
        </div>
        <div className="mc-card p-4 border-mc-orange/30">
          <p className="text-sm text-gray-400">Em Pausa</p>
          <p className="text-2xl font-bold font-mono text-mc-orange">{onBreak}</p>
        </div>
        <div className="mc-card p-4 border-mc-gold/30">
          <p className="text-sm text-gray-400">Pedidos Hoje</p>
          <p className="text-2xl font-bold font-mono text-mc-gold">{totalOrders}</p>
        </div>
      </div>

      {/* Staff Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {staff.map((member, index) => {
          const RoleIcon = roleIcons[member.role as keyof typeof roleIcons] || Users
          const station = stations.find(s => s.id === member.station)
          
          return (
            <motion.div
              key={member.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className={cn(
                'mc-card p-4 transition-all',
                member.status === 'active' && 'border-mc-green/30',
                member.status === 'break' && 'border-mc-orange/30',
                member.status === 'offline' && 'border-mc-red/30 opacity-50'
              )}
            >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className={cn(
                    'w-12 h-12 rounded-full flex items-center justify-center text-lg font-bold',
                    member.status === 'active' ? 'bg-mc-green/20 text-mc-green' :
                    member.status === 'break' ? 'bg-mc-orange/20 text-mc-orange' :
                    'bg-mc-red/20 text-mc-red'
                  )}>
                    {member.name.split(' ').map(n => n[0]).join('')}
                  </div>
                  <div>
                    <h3 className="font-bold text-white">{member.name}</h3>
                    <div className="flex items-center gap-1 text-gray-400 text-sm">
                      <RoleIcon className="w-3 h-3" />
                      {roleLabels[member.role as keyof typeof roleLabels]}
                    </div>
                  </div>
                </div>
                <span className={cn(
                  'px-2 py-1 rounded-full text-xs font-medium',
                  member.status === 'active' ? 'bg-mc-green/20 text-mc-green' :
                  member.status === 'break' ? 'bg-mc-orange/20 text-mc-orange' :
                  'bg-mc-red/20 text-mc-red'
                )}>
                  {member.status === 'active' ? 'Ativo' : member.status === 'break' ? 'Pausa' : 'Offline'}
                </span>
              </div>

              <div className="space-y-2">
                {station && (
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-500">Estação</span>
                    <span className="text-white">{station.name}</span>
                  </div>
                )}
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">Pedidos Completados</span>
                  <span className="text-mc-gold font-mono">{member.ordersCompleted}</span>
                </div>
              </div>

              <div className="flex gap-2 mt-4 pt-4 border-t border-mc-gray/30">
                {member.status === 'active' ? (
                  <button className="flex-1 mc-button-outline text-sm flex items-center justify-center gap-2">
                    <Coffee className="w-4 h-4" />
                    Dar Pausa
                  </button>
                ) : member.status === 'break' ? (
                  <button className="flex-1 mc-button text-sm flex items-center justify-center gap-2">
                    <UserCheck className="w-4 h-4" />
                    Retornar
                  </button>
                ) : null}
              </div>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
