'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn } from '@/lib/utils'
import { BRANDS } from '@/types/store'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Command,
  Search,
  LayoutDashboard,
  MessageSquare,
  ShoppingCart,
  Users,
  BarChart3,
  X,
} from 'lucide-react'
import { useEffect, useState, useCallback } from 'react'

const commands = [
  { id: 'dashboard', label: 'Ir para Dashboard', icon: LayoutDashboard, shortcut: '1' },
  { id: 'conversations', label: 'Ver Conversas', icon: MessageSquare, shortcut: '2' },
  { id: 'orders', label: 'Ver Pedidos', icon: ShoppingCart, shortcut: '3' },
  { id: 'sellers', label: 'Ver Vendedores', icon: Users, shortcut: '4' },
  { id: 'analytics', label: 'Ver Analytics', icon: BarChart3, shortcut: '5' },
] as const

export function CommandPalette() {
  const { 
    commandPaletteOpen, 
    toggleCommandPalette, 
    setActiveView,
    setSelectedBrand,
  } = useStoreAppStore()
  
  const [search, setSearch] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  
  const allItems = [
    ...commands.map(c => ({ ...c, type: 'command' as const })),
    ...Object.values(BRANDS).map(b => ({
      id: `brand-${b.id}`,
      label: `Filtrar: ${b.name}`,
      icon: () => <span className="text-lg">{b.emoji}</span>,
      shortcut: '',
      type: 'brand' as const,
      brandId: b.id,
    })),
  ]
  
  const filteredItems = allItems.filter(item =>
    item.label.toLowerCase().includes(search.toLowerCase())
  )

  const handleSelect = useCallback((item: typeof allItems[0]) => {
    if (item.type === 'command') {
      setActiveView(item.id as any)
    } else if (item.type === 'brand') {
      setSelectedBrand((item as any).brandId)
    }
    toggleCommandPalette()
    setSearch('')
    setSelectedIndex(0)
  }, [setActiveView, setSelectedBrand, toggleCommandPalette])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Open palette with Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        toggleCommandPalette()
      }
      
      // Close with Escape
      if (e.key === 'Escape' && commandPaletteOpen) {
        toggleCommandPalette()
        setSearch('')
        setSelectedIndex(0)
      }
      
      // Navigate with arrows
      if (commandPaletteOpen) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          setSelectedIndex(i => Math.min(i + 1, filteredItems.length - 1))
        }
        if (e.key === 'ArrowUp') {
          e.preventDefault()
          setSelectedIndex(i => Math.max(i - 1, 0))
        }
        if (e.key === 'Enter' && filteredItems[selectedIndex]) {
          e.preventDefault()
          handleSelect(filteredItems[selectedIndex])
        }
      }
    }
    
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [commandPaletteOpen, toggleCommandPalette, filteredItems, selectedIndex, handleSelect])

  // Reset selection when search changes
  useEffect(() => {
    setSelectedIndex(0)
  }, [search])

  return (
    <AnimatePresence>
      {commandPaletteOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={toggleCommandPalette}
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50"
          />
          
          {/* Palette */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: -20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: -20 }}
            className="fixed top-[20%] left-1/2 -translate-x-1/2 w-full max-w-lg z-50"
          >
            <div className="bg-store-dark border border-store-purple/30 rounded-xl shadow-2xl overflow-hidden">
              {/* Search Input */}
              <div className="flex items-center gap-3 p-4 border-b border-store-purple/20">
                <Search className="w-5 h-5 text-gray-500" />
                <input
                  type="text"
                  placeholder="Buscar comandos..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  autoFocus
                  className="flex-1 bg-transparent text-white placeholder-gray-500 focus:outline-none"
                />
                <button
                  onClick={toggleCommandPalette}
                  className="p-1 rounded hover:bg-store-gray/50 transition-colors"
                >
                  <X className="w-4 h-4 text-gray-500" />
                </button>
              </div>
              
              {/* Commands List */}
              <div className="max-h-80 overflow-y-auto p-2">
                {filteredItems.length === 0 ? (
                  <div className="p-4 text-center text-gray-500">
                    Nenhum resultado encontrado
                  </div>
                ) : (
                  <>
                    {/* Commands Section */}
                    <div className="mb-2">
                      <p className="px-3 py-1 text-xs text-gray-600 uppercase">Navegação</p>
                      {filteredItems.filter(i => i.type === 'command').map((item, index) => {
                        const Icon = item.icon
                        const globalIndex = filteredItems.indexOf(item)
                        
                        return (
                          <button
                            key={item.id}
                            onClick={() => handleSelect(item)}
                            className={cn(
                              "w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors",
                              globalIndex === selectedIndex
                                ? "bg-store-purple/20 text-white"
                                : "text-gray-400 hover:bg-store-gray/50 hover:text-white"
                            )}
                          >
                            <Icon className="w-4 h-4" />
                            <span className="flex-1 text-left text-sm">{item.label}</span>
                            {item.shortcut && (
                              <kbd className="px-2 py-0.5 text-xs bg-store-gray rounded">
                                {item.shortcut}
                              </kbd>
                            )}
                          </button>
                        )
                      })}
                    </div>
                    
                    {/* Brands Section */}
                    {filteredItems.filter(i => i.type === 'brand').length > 0 && (
                      <div>
                        <p className="px-3 py-1 text-xs text-gray-600 uppercase">Marcas</p>
                        {filteredItems.filter(i => i.type === 'brand').map((item) => {
                          const Icon = item.icon
                          const globalIndex = filteredItems.indexOf(item)
                          
                          return (
                            <button
                              key={item.id}
                              onClick={() => handleSelect(item)}
                              className={cn(
                                "w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors",
                                globalIndex === selectedIndex
                                  ? "bg-store-purple/20 text-white"
                                  : "text-gray-400 hover:bg-store-gray/50 hover:text-white"
                              )}
                            >
                              <Icon />
                              <span className="flex-1 text-left text-sm">{item.label}</span>
                            </button>
                          )
                        })}
                      </div>
                    )}
                  </>
                )}
              </div>
              
              {/* Footer */}
              <div className="flex items-center justify-between px-4 py-2 border-t border-store-purple/20 text-xs text-gray-600">
                <div className="flex items-center gap-4">
                  <span className="flex items-center gap-1">
                    <kbd className="px-1.5 py-0.5 bg-store-gray rounded">↑</kbd>
                    <kbd className="px-1.5 py-0.5 bg-store-gray rounded">↓</kbd>
                    navegar
                  </span>
                  <span className="flex items-center gap-1">
                    <kbd className="px-1.5 py-0.5 bg-store-gray rounded">Enter</kbd>
                    selecionar
                  </span>
                </div>
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-store-gray rounded">Esc</kbd>
                  fechar
                </span>
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}
