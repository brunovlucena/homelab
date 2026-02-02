'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { cn, getBrandColor, getStatusColor } from '@/lib/utils'
import { BRANDS } from '@/types/store'
import { 
  LayoutDashboard, 
  MessageSquare, 
  ShoppingCart, 
  Users,
  BarChart3,
  ChevronRight,
  Circle,
  Package
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

const navItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'conversations', label: 'Conversas', icon: MessageSquare },
  { id: 'orders', label: 'Pedidos', icon: ShoppingCart },
  { id: 'sellers', label: 'Vendedores AI', icon: Users },
  { id: 'analytics', label: 'Analytics', icon: BarChart3 },
] as const

export function Sidebar() {
  const { 
    sidebarOpen, 
    activeView, 
    setActiveView,
    sellers,
    selectedBrand,
    setSelectedBrand,
    conversations,
  } = useStoreAppStore()

  const escalatedCount = conversations.filter(c => c.state === 'escalated').length
  const activeCount = conversations.filter(c => c.state === 'active').length

  return (
    <AnimatePresence>
      {sidebarOpen && (
        <motion.aside
          initial={{ x: -280, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          exit={{ x: -280, opacity: 0 }}
          transition={{ type: 'spring', damping: 25, stiffness: 200 }}
          className="fixed left-0 top-16 bottom-0 w-64 z-40 bg-store-darker/95 backdrop-blur-xl border-r border-store-purple/20"
        >
          <div className="flex flex-col h-full">
            {/* Navigation */}
            <nav className="p-4 space-y-1">
              <p className="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3 px-3">
                Navegação
              </p>
              {navItems.map((item) => {
                const Icon = item.icon
                const isActive = activeView === item.id
                const badge = item.id === 'conversations' && escalatedCount > 0 
                  ? escalatedCount 
                  : undefined
                
                return (
                  <button
                    key={item.id}
                    onClick={() => setActiveView(item.id)}
                    className={cn(
                      "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all group",
                      isActive
                        ? "bg-store-purple/20 border border-store-purple/50 text-white"
                        : "text-gray-400 hover:bg-store-gray/50 hover:text-white"
                    )}
                  >
                    <Icon className={cn(
                      "w-5 h-5 transition-colors",
                      isActive ? "text-store-purple" : "text-gray-500 group-hover:text-store-purple"
                    )} />
                    <span className="text-sm">{item.label}</span>
                    {badge && (
                      <span className="ml-auto px-2 py-0.5 text-xs font-bold bg-store-red/20 text-store-red rounded-full">
                        {badge}
                      </span>
                    )}
                    {isActive && (
                      <ChevronRight className="w-4 h-4 ml-auto text-store-purple" />
                    )}
                  </button>
                )
              })}
            </nav>

            {/* Brand Quick Access */}
            <div className="flex-1 p-4 overflow-y-auto border-t border-store-purple/10">
              <p className="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3 px-3">
                Vendedores AI
              </p>
              <div className="space-y-1">
                {sellers.map((seller) => {
                  const brand = BRANDS[seller.brand]
                  const isSelected = selectedBrand === seller.brand
                  
                  return (
                    <button
                      key={seller.id}
                      onClick={() => {
                        setSelectedBrand(isSelected ? 'all' : seller.brand)
                      }}
                      className={cn(
                        "w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-all group",
                        isSelected
                          ? "bg-store-purple/20 border border-store-purple/40"
                          : "hover:bg-store-gray/50"
                      )}
                    >
                      <div className="relative">
                        <span className="text-xl">{brand.sellerAvatar}</span>
                        <Circle 
                          className={cn(
                            "absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5",
                            getStatusColor(seller.status)
                          )} 
                          fill="currentColor"
                        />
                      </div>
                      <div className="flex-1 text-left">
                        <p className={cn(
                          "text-sm group-hover:text-white truncate",
                          getBrandColor(seller.brand)
                        )}>
                          {brand.sellerName}
                        </p>
                        <p className="text-xs text-gray-500">
                          {seller.activeConversations} conversas
                        </p>
                      </div>
                      <span className="text-lg">{brand.emoji}</span>
                    </button>
                  )
                })}
              </div>
            </div>

            {/* Footer Stats */}
            <div className="p-4 border-t border-store-purple/20">
              <div className="grid grid-cols-2 gap-2">
                <div className="p-3 rounded-lg bg-store-gray/30 border border-store-green/20">
                  <p className="text-xs text-gray-500">Ativos</p>
                  <p className="text-lg font-bold text-store-green">{activeCount}</p>
                </div>
                <div className="p-3 rounded-lg bg-store-gray/30 border border-store-red/20">
                  <p className="text-xs text-gray-500">Escalados</p>
                  <p className="text-lg font-bold text-store-red">{escalatedCount}</p>
                </div>
              </div>
            </div>
          </div>
        </motion.aside>
      )}
    </AnimatePresence>
  )
}
