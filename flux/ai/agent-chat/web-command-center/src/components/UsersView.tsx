'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Search, 
  Plus, 
  MoreVertical, 
  Mic, 
  MapPin, 
  Bot,
  CheckCircle,
  XCircle,
  Clock
} from 'lucide-react'
import type { User, UserStatus } from '@/types'

// Mock data
const mockUsers: User[] = [
  {
    id: 'user-001',
    username: 'bruno_lucena',
    displayName: 'Bruno Lucena',
    email: 'bruno@example.com',
    avatar: '👨‍💻',
    status: 'active',
    createdAt: '2025-01-15T10:00:00Z',
    lastActiveAt: '2025-12-10T09:30:00Z',
    agentAssistantId: 'agent-assistant-001',
    settings: {
      voiceCloneEnabled: true,
      voiceCloneId: 'vc-001',
      locationSharingEnabled: true,
      proximityAlertRadius: 5,
      contactsCanSeeLocation: true,
      notificationsEnabled: true,
    }
  },
  {
    id: 'user-002',
    username: 'maria_garcia',
    displayName: 'Maria Garcia',
    email: 'maria@example.com',
    avatar: '👩‍🎨',
    status: 'active',
    createdAt: '2025-02-20T14:00:00Z',
    lastActiveAt: '2025-12-10T08:45:00Z',
    agentAssistantId: 'agent-assistant-002',
    settings: {
      voiceCloneEnabled: true,
      voiceCloneId: 'vc-002',
      locationSharingEnabled: false,
      proximityAlertRadius: 10,
      contactsCanSeeLocation: false,
      notificationsEnabled: true,
    }
  },
  {
    id: 'user-003',
    username: 'john_doe',
    displayName: 'John Doe',
    email: 'john@example.com',
    avatar: '🧑‍💼',
    status: 'pending',
    createdAt: '2025-12-09T16:00:00Z',
    lastActiveAt: '2025-12-09T16:00:00Z',
    settings: {
      voiceCloneEnabled: false,
      locationSharingEnabled: false,
      proximityAlertRadius: 5,
      contactsCanSeeLocation: false,
      notificationsEnabled: true,
    }
  },
]

export function UsersView() {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<UserStatus | 'all'>('all')

  const filteredUsers = mockUsers.filter(user => {
    const matchesSearch = user.displayName.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         user.email.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = selectedStatus === 'all' || user.status === selectedStatus
    return matchesSearch && matchesStatus
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Search users..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="input-field pl-10 w-80"
            />
          </div>
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value as UserStatus | 'all')}
            className="input-field"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="pending">Pending</option>
            <option value="inactive">Inactive</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
        <button className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add User
        </button>
      </div>

      {/* Users Table */}
      <div className="card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-cyber-purple/20">
              <th className="text-left p-4 text-sm font-medium text-gray-400">User</th>
              <th className="text-left p-4 text-sm font-medium text-gray-400">Status</th>
              <th className="text-left p-4 text-sm font-medium text-gray-400">Features</th>
              <th className="text-left p-4 text-sm font-medium text-gray-400">Agent</th>
              <th className="text-left p-4 text-sm font-medium text-gray-400">Last Active</th>
              <th className="text-right p-4 text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.map((user) => (
              <motion.tr
                key={user.id}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="border-b border-cyber-purple/10 hover:bg-cyber-purple/5 transition-colors"
              >
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center text-lg">
                      {user.avatar}
                    </div>
                    <div>
                      <p className="font-medium">{user.displayName}</p>
                      <p className="text-sm text-gray-500">@{user.username}</p>
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <StatusBadge status={user.status} />
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    <FeatureIcon 
                      enabled={user.settings.voiceCloneEnabled} 
                      icon={Mic} 
                      label="Voice Clone"
                    />
                    <FeatureIcon 
                      enabled={user.settings.locationSharingEnabled} 
                      icon={MapPin} 
                      label="Location"
                    />
                  </div>
                </td>
                <td className="p-4">
                  {user.agentAssistantId ? (
                    <div className="flex items-center gap-2 text-cyber-green">
                      <Bot className="w-4 h-4" />
                      <span className="text-sm">Deployed</span>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2 text-gray-500">
                      <Clock className="w-4 h-4" />
                      <span className="text-sm">Pending</span>
                    </div>
                  )}
                </td>
                <td className="p-4 text-sm text-gray-400">
                  {formatRelativeTime(user.lastActiveAt)}
                </td>
                <td className="p-4 text-right">
                  <button className="p-2 rounded-lg hover:bg-cyber-purple/20 text-gray-400 hover:text-white transition-colors">
                    <MoreVertical className="w-4 h-4" />
                  </button>
                </td>
              </motion.tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: UserStatus }) {
  const styles: Record<UserStatus, { bg: string; text: string; icon: React.ElementType }> = {
    active: { bg: 'bg-cyber-green/20', text: 'text-cyber-green', icon: CheckCircle },
    inactive: { bg: 'bg-gray-500/20', text: 'text-gray-400', icon: XCircle },
    pending: { bg: 'bg-cyber-yellow/20', text: 'text-cyber-yellow', icon: Clock },
    suspended: { bg: 'bg-cyber-red/20', text: 'text-cyber-red', icon: XCircle },
  }

  const { bg, text, icon: Icon } = styles[status]

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${bg} ${text}`}>
      <Icon className="w-3 h-3" />
      {status}
    </span>
  )
}

function FeatureIcon({ enabled, icon: Icon, label }: { enabled: boolean; icon: React.ElementType; label: string }) {
  return (
    <div 
      className={`p-1.5 rounded-lg ${enabled ? 'bg-cyber-purple/20 text-cyber-purple' : 'bg-gray-500/20 text-gray-500'}`}
      title={`${label}: ${enabled ? 'Enabled' : 'Disabled'}`}
    >
      <Icon className="w-4 h-4" />
    </div>
  )
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  return `${Math.floor(diffMins / 1440)}d ago`
}
