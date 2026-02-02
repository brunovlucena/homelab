'use client'

import { LucideIcon, MessageSquareText } from 'lucide-react'
import { cn } from '@/lib/utils'

interface NavItem {
  id: string
  label: string
  icon: LucideIcon
}

interface SidebarProps {
  navItems: NavItem[]
  currentView: string
  onNavigate: (view: string) => void
}

export function Sidebar({ navItems, currentView, onNavigate }: SidebarProps) {
  return (
    <aside className="w-64 bg-cyber-gray/30 border-r border-cyber-purple/20 flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-cyber-purple/20">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center">
            <MessageSquareText className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="font-bold text-lg gradient-text">AgentChat</h1>
            <p className="text-xs text-gray-500">Command Center</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <button
            key={item.id}
            onClick={() => onNavigate(item.id)}
            className={cn(
              'nav-item w-full',
              currentView === item.id && 'nav-item-active'
            )}
          >
            <item.icon className="w-5 h-5" />
            <span>{item.label}</span>
          </button>
        ))}
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-cyber-purple/20">
        <div className="card p-3">
          <div className="flex items-center gap-2 mb-2">
            <div className="w-2 h-2 rounded-full bg-cyber-green animate-pulse" />
            <span className="text-sm text-gray-400">System Status</span>
          </div>
          <p className="text-xs text-gray-500">All agents operational</p>
        </div>
      </div>
    </aside>
  )
}
