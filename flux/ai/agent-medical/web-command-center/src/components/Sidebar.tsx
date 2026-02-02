'use client'

import { motion } from 'framer-motion'
import { Stethoscope } from 'lucide-react'

interface NavItem {
  id: string
  label: string
  icon: React.ElementType
}

interface SidebarProps {
  navItems: NavItem[]
  currentView: string
  onNavigate: (view: string) => void
}

export function Sidebar({ navItems, currentView, onNavigate }: SidebarProps) {
  return (
    <div className="w-64 bg-gradient-to-b from-gray-900 to-gray-900/50 border-r border-gray-800 flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-gray-800">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-gradient-to-br from-medical-blue to-medical-green rounded-lg">
            <Stethoscope className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold bg-gradient-to-r from-medical-blue to-medical-green bg-clip-text text-transparent">
              Medical Agent
            </h1>
            <p className="text-xs text-gray-500">Command Center</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-2">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = currentView === item.id
          
          return (
            <motion.button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={`
                w-full flex items-center gap-3 px-4 py-3 rounded-lg
                transition-all text-left
                ${isActive 
                  ? 'bg-gradient-to-r from-medical-blue/20 to-medical-green/20 border border-medical-blue/50 text-white' 
                  : 'text-gray-400 hover:bg-gray-800/50 hover:text-white'
                }
              `}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              <Icon className={`w-5 h-5 ${isActive ? 'text-medical-blue' : ''}`} />
              <span className="font-medium">{item.label}</span>
            </motion.button>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-gray-800">
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-2 h-2 bg-medical-green rounded-full animate-pulse" />
          <span>HIPAA Compliant</span>
        </div>
        <p className="text-xs text-gray-600 mt-1">v1.0.0</p>
      </div>
    </div>
  )
}
