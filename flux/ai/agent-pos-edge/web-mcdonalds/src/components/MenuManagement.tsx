'use client'

import { useMcdonaldsStore } from '@/store/mcdonaldsStore'
import { cn, formatCurrency } from '@/lib/utils'
import { motion } from 'framer-motion'
import { UtensilsCrossed, Coffee, Cookie, Sunrise, Gift, Search, ToggleLeft, ToggleRight } from 'lucide-react'
import { useState } from 'react'

const categoryIcons = {
  'burgers': UtensilsCrossed,
  'sides': Cookie,
  'drinks': Coffee,
  'desserts': Cookie,
  'breakfast': Sunrise,
  'happy-meal': Gift,
}

const categoryLabels = {
  'burgers': 'Sanduíches',
  'sides': 'Acompanhamentos',
  'drinks': 'Bebidas',
  'desserts': 'Sobremesas',
  'breakfast': 'Café da Manhã',
  'happy-meal': 'McLanche Feliz',
}

export function MenuManagement() {
  const { menuItems } = useMcdonaldsStore()
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [search, setSearch] = useState('')

  const categories = ['all', ...Array.from(new Set(menuItems.map(item => item.category)))]
  
  const filteredItems = menuItems.filter(item => {
    if (selectedCategory !== 'all' && item.category !== selectedCategory) return false
    if (search && !item.name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">Gerenciamento do Cardápio</h1>
          <p className="text-gray-400">Controle de itens e disponibilidade</p>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col md:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Buscar item..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-mc-gray/50 border border-mc-gray rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-mc-gold/50"
          />
        </div>
      </div>

      {/* Category Tabs */}
      <div className="flex gap-2 overflow-x-auto pb-2">
        {categories.map((cat) => {
          const Icon = cat === 'all' ? UtensilsCrossed : categoryIcons[cat as keyof typeof categoryIcons] || UtensilsCrossed
          const label = cat === 'all' ? 'Todos' : categoryLabels[cat as keyof typeof categoryLabels] || cat
          
          return (
            <button
              key={cat}
              onClick={() => setSelectedCategory(cat)}
              className={cn(
                'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors',
                selectedCategory === cat 
                  ? 'bg-mc-red text-white' 
                  : 'bg-mc-gray/50 text-gray-400 hover:bg-mc-gray'
              )}
            >
              <Icon className="w-4 h-4" />
              {label}
            </button>
          )
        })}
      </div>

      {/* Menu Items Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredItems.map((item, index) => {
          const CategoryIcon = categoryIcons[item.category as keyof typeof categoryIcons] || UtensilsCrossed
          
          return (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className={cn(
                'mc-card p-4 transition-all',
                !item.available && 'opacity-50'
              )}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-mc-red/10 border border-mc-red/30">
                    <CategoryIcon className="w-5 h-5 text-mc-red" />
                  </div>
                  <div>
                    <h3 className="font-bold text-white">{item.name}</h3>
                    <p className="text-xs text-gray-500 capitalize">{categoryLabels[item.category as keyof typeof categoryLabels]}</p>
                  </div>
                </div>
                <button className={cn(
                  'transition-colors',
                  item.available ? 'text-mc-green' : 'text-mc-red'
                )}>
                  {item.available ? (
                    <ToggleRight className="w-8 h-8" />
                  ) : (
                    <ToggleLeft className="w-8 h-8" />
                  )}
                </button>
              </div>

              <div className="flex items-center justify-between pt-3 border-t border-mc-gray/30">
                <div>
                  <p className="text-xs text-gray-500">Preço</p>
                  <p className="text-lg font-bold text-mc-gold">{formatCurrency(item.price)}</p>
                </div>
                <div className="text-right">
                  <p className="text-xs text-gray-500">Tempo Preparo</p>
                  <p className="text-sm font-mono text-white">{Math.floor(item.prepTime / 60)}:{(item.prepTime % 60).toString().padStart(2, '0')}</p>
                </div>
              </div>

              <div className="mt-3">
                <span className={cn(
                  'px-2 py-1 rounded-full text-xs font-medium',
                  item.available ? 'bg-mc-green/20 text-mc-green' : 'bg-mc-red/20 text-mc-red'
                )}>
                  {item.available ? 'Disponível' : 'Indisponível'}
                </span>
              </div>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
