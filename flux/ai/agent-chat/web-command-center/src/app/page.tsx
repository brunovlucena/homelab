'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  LayoutDashboard, 
  Users, 
  Bot, 
  MessageSquare, 
  MapPin, 
  Mic, 
  Image,
  Bell, 
  Settings,
  Shield,
  Activity,
  Zap,
  TrendingUp
} from 'lucide-react'

// Components
import { Sidebar } from '@/components/Sidebar'
import { Header } from '@/components/Header'
import { DashboardView } from '@/components/DashboardView'
import { UsersView } from '@/components/UsersView'
import { AgentsView } from '@/components/AgentsView'
import { ChatsView } from '@/components/ChatsView'
import { AlertsView } from '@/components/AlertsView'

type View = 'dashboard' | 'users' | 'agents' | 'chats' | 'alerts' | 'settings'

const navItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'users', label: 'Users', icon: Users },
  { id: 'agents', label: 'Agents', icon: Bot },
  { id: 'chats', label: 'Conversations', icon: MessageSquare },
  { id: 'alerts', label: 'Alerts', icon: Bell },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export default function CommandCenter() {
  const [currentView, setCurrentView] = useState<View>('dashboard')

  const renderView = () => {
    switch (currentView) {
      case 'dashboard':
        return <DashboardView />
      case 'users':
        return <UsersView />
      case 'agents':
        return <AgentsView />
      case 'chats':
        return <ChatsView />
      case 'alerts':
        return <AlertsView />
      case 'settings':
        return <SettingsPlaceholder />
      default:
        return <DashboardView />
    }
  }

  return (
    <div className="flex h-screen bg-cyber-dark">
      {/* Sidebar */}
      <Sidebar 
        navItems={navItems}
        currentView={currentView}
        onNavigate={(view) => setCurrentView(view as View)}
      />
      
      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header title={navItems.find(n => n.id === currentView)?.label || 'Dashboard'} />
        
        <main className="flex-1 overflow-auto p-6">
          <motion.div
            key={currentView}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
          >
            {renderView()}
          </motion.div>
        </main>
      </div>
    </div>
  )
}

function SettingsPlaceholder() {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <Settings className="w-16 h-16 text-cyber-purple/50 mx-auto mb-4" />
        <h2 className="text-xl font-bold text-gray-400">Settings</h2>
        <p className="text-gray-500 mt-2">System configuration coming soon</p>
      </div>
    </div>
  )
}
