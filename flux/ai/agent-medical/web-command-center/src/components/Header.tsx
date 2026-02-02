'use client'

import { Bell, Search, User } from 'lucide-react'

interface HeaderProps {
  title: string
}

export function Header({ title }: HeaderProps) {
  return (
    <header className="h-16 border-b border-gray-800 bg-gray-900/50 backdrop-blur-sm flex items-center justify-between px-6">
      <div>
        <h2 className="text-xl font-bold">{title}</h2>
        <p className="text-xs text-gray-500">HIPAA-Compliant Medical Records System</p>
      </div>

      <div className="flex items-center gap-4">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search patients..."
            className="pl-10 pr-4 py-2 bg-gray-800/50 border border-gray-700 rounded-lg text-sm focus:outline-none focus:border-medical-blue transition-colors w-64"
          />
        </div>

        {/* Notifications */}
        <button className="relative p-2 rounded-lg hover:bg-gray-800/50 transition-colors">
          <Bell className="w-5 h-5 text-gray-400" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-medical-red rounded-full"></span>
        </button>

        {/* User */}
        <button className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-800/50 transition-colors">
          <div className="w-8 h-8 bg-gradient-to-br from-medical-blue to-medical-green rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-white" />
          </div>
          <div className="text-left">
            <p className="text-sm font-medium">Dr. Admin</p>
            <p className="text-xs text-gray-500">Doctor</p>
          </div>
        </button>
      </div>
    </header>
  )
}
