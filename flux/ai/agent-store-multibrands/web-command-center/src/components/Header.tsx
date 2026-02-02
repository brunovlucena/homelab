'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, getBrandBgColor } from '@/lib/utils'
import { BRANDS, BrandId } from '@/types/store'
import { 
  Menu, 
  Command, 
  Bell, 
  Search,
  Activity,
  ChevronDown
} from 'lucide-react'
import { motion } from 'framer-motion'
import { useState } from 'react'

export function Header() {
  const { 
    toggleSidebar, 
    toggleCommandPalette,
    selectedBrand,
    setSelectedBrand,
    metrics,
    events,
  } = useStoreAppStore()
  
  const [brandDropdownOpen, setBrandDropdownOpen] = useState(false)
  
  const unreadEvents = events.filter(e => 
    new Date(e.timestamp).getTime() > Date.now() - 3600000
  ).length

  return (
    <header className="fixed top-0 left-0 right-0 h-16 z-50 bg-store-darker/95 backdrop-blur-xl border-b border-store-purple/20">
      <div className="flex items-center justify-between h-full px-4">
        {/* Left Section */}
        <div className="flex items-center gap-4">
          <button
            onClick={toggleSidebar}
            className="p-2 rounded-lg hover:bg-store-gray/50 transition-colors"
          >
            <Menu className="w-5 h-5 text-gray-400 hover:text-white" />
          </button>
          
          <div className="flex items-center gap-3">
            <div className="relative">
              <span className="text-2xl">🏪</span>
              <div className="absolute -top-1 -right-1 w-3 h-3 bg-store-green rounded-full animate-pulse" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-white">
                Store <span className="text-gradient">Command Center</span>
              </h1>
              <p className="text-xs text-gray-500 font-mono">MultiBrands AI Sellers</p>
            </div>
          </div>
        </div>

        {/* Center Section - Brand Selector */}
        <div className="relative">
          <button
            onClick={() => setBrandDropdownOpen(!brandDropdownOpen)}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-store-gray/50 border border-store-purple/20 hover:border-store-purple/40 transition-all"
          >
            {selectedBrand === 'all' ? (
              <>
                <span className="text-lg">🌐</span>
                <span className="text-sm text-gray-300">Todas as Marcas</span>
              </>
            ) : (
              <>
                <span className="text-lg">{BRANDS[selectedBrand].emoji}</span>
                <span className={cn("text-sm", getBrandColor(selectedBrand))}>
                  {BRANDS[selectedBrand].name}
                </span>
              </>
            )}
            <ChevronDown className="w-4 h-4 text-gray-500" />
          </button>
          
          {brandDropdownOpen && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="absolute top-full mt-2 left-0 w-48 py-2 rounded-lg bg-store-gray border border-store-purple/30 shadow-xl"
            >
              <button
                onClick={() => {
                  setSelectedBrand('all')
                  setBrandDropdownOpen(false)
                }}
                className="w-full flex items-center gap-2 px-4 py-2 hover:bg-store-purple/10 transition-colors"
              >
                <span className="text-lg">🌐</span>
                <span className="text-sm text-gray-300">Todas as Marcas</span>
              </button>
              <div className="h-px bg-store-purple/20 my-1" />
              {Object.values(BRANDS).map((brand) => (
                <button
                  key={brand.id}
                  onClick={() => {
                    setSelectedBrand(brand.id)
                    setBrandDropdownOpen(false)
                  }}
                  className="w-full flex items-center gap-2 px-4 py-2 hover:bg-store-purple/10 transition-colors"
                >
                  <span className="text-lg">{brand.emoji}</span>
                  <span className={cn("text-sm", getBrandColor(brand.id))}>
                    {brand.name}
                  </span>
                  <span className="text-xs text-gray-500 ml-auto">
                    {brand.sellerName}
                  </span>
                </button>
              ))}
            </motion.div>
          )}
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-3">
          {/* Live Stats */}
          <div className="hidden md:flex items-center gap-4 px-4 py-1.5 rounded-lg bg-store-gray/30 border border-store-purple/10">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-store-green" />
              <span className="text-xs font-mono text-gray-400">
                {metrics.activeConversations} ativos
              </span>
            </div>
            <div className="h-4 w-px bg-store-purple/20" />
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono text-store-green">
                R$ {metrics.totalRevenue.toLocaleString('pt-BR', { minimumFractionDigits: 0 })}
              </span>
            </div>
          </div>
          
          {/* Search */}
          <button
            onClick={toggleCommandPalette}
            className="flex items-center gap-2 px-3 py-2 rounded-lg bg-store-gray/50 border border-store-purple/20 hover:border-store-purple/40 transition-all"
          >
            <Search className="w-4 h-4 text-gray-500" />
            <span className="text-sm text-gray-500 hidden sm:inline">Buscar...</span>
            <kbd className="hidden sm:flex items-center gap-1 px-2 py-0.5 rounded bg-store-darker text-xs text-gray-600">
              <Command className="w-3 h-3" /> K
            </kbd>
          </button>

          {/* Notifications */}
          <button className="relative p-2 rounded-lg hover:bg-store-gray/50 transition-colors">
            <Bell className="w-5 h-5 text-gray-400 hover:text-white" />
            {unreadEvents > 0 && (
              <span className="absolute top-1 right-1 w-4 h-4 flex items-center justify-center bg-store-red rounded-full text-[10px] font-bold">
                {unreadEvents > 9 ? '9+' : unreadEvents}
              </span>
            )}
          </button>

          {/* User Avatar */}
          <div className="w-9 h-9 rounded-full bg-gradient-to-br from-store-purple to-store-pink flex items-center justify-center text-sm font-bold cursor-pointer hover:shadow-lg hover:shadow-store-purple/30 transition-all">
            B
          </div>
        </div>
      </div>
    </header>
  )
}
