'use client'

import { Bell, Search, User } from 'lucide-react'

interface HeaderProps {
  title: string
}

export function Header({ title }: HeaderProps) {
  return (
    <header className="h-16 bg-cyber-gray/30 border-b border-cyber-purple/20 flex items-center justify-between px-6">
      <h1 className="text-xl font-bold">{title}</h1>
      
      <div className="flex items-center gap-4">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search..."
            className="input-field pl-10 w-64 text-sm"
          />
        </div>

        {/* Notifications */}
        <button className="relative p-2 rounded-lg hover:bg-cyber-purple/10 transition-colors">
          <Bell className="w-5 h-5 text-gray-400" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-cyber-red rounded-full" />
        </button>

        {/* User Menu */}
        <button className="flex items-center gap-2 p-2 rounded-lg hover:bg-cyber-purple/10 transition-colors">
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center">
            <User className="w-4 h-4 text-white" />
          </div>
          <span className="text-sm font-medium">Admin</span>
        </button>
      </div>
    </header>
  )
}
