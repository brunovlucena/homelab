'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  LayoutDashboard, 
  Users, 
  FileText, 
  Activity,
  Shield,
  Bell, 
  Settings,
  Stethoscope,
  ClipboardList,
  AlertCircle
} from 'lucide-react'

// Components
import { Sidebar } from '@/components/Sidebar'
import { Header } from '@/components/Header'
import { DashboardView } from '@/components/DashboardView'
import { PatientsView } from '@/components/PatientsView'
import { RecordsView } from '@/components/RecordsView'
import { ComplianceView } from '@/components/ComplianceView'
import { AlertsView } from '@/components/AlertsView'

type View = 'dashboard' | 'patients' | 'records' | 'compliance' | 'alerts' | 'settings'

const navItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'patients', label: 'Patients', icon: Users },
  { id: 'records', label: 'Medical Records', icon: FileText },
  { id: 'compliance', label: 'HIPAA Compliance', icon: Shield },
  { id: 'alerts', label: 'Alerts', icon: Bell },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export default function MedicalCommandCenter() {
  const [currentView, setCurrentView] = useState<View>('dashboard')

  const renderView = () => {
    switch (currentView) {
      case 'dashboard':
        return <DashboardView />
      case 'patients':
        return <PatientsView />
      case 'records':
        return <RecordsView />
      case 'compliance':
        return <ComplianceView />
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
