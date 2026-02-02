'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Search, Eye, MoreVertical, MessageSquare, Bot, User, Image, Mic, MapPin } from 'lucide-react'

interface ChatPreview {
  id: string
  participants: { name: string; avatar: string; isAgent: boolean }[]
  lastMessage: string
  lastMessageType: string
  timestamp: string
  messageCount: number
}

// Mock data
const mockChats: ChatPreview[] = [
  {
    id: 'chat-001',
    participants: [
      { name: 'Bruno Lucena', avatar: '👨‍💻', isAgent: false },
      { name: 'Agent Assistant', avatar: '🤖', isAgent: true },
    ],
    lastMessage: 'Sure! I can generate that sunset image for you. Here it is...',
    lastMessageType: 'image',
    timestamp: '2025-12-10T09:30:00Z',
    messageCount: 156,
  },
  {
    id: 'chat-002',
    participants: [
      { name: 'Maria Garcia', avatar: '👩‍🎨', isAgent: false },
      { name: 'Agent Assistant', avatar: '🤖', isAgent: true },
    ],
    lastMessage: 'Your voice message has been sent to your contact list.',
    lastMessageType: 'voice',
    timestamp: '2025-12-10T08:45:00Z',
    messageCount: 89,
  },
  {
    id: 'chat-003',
    participants: [
      { name: 'John Doe', avatar: '🧑‍💼', isAgent: false },
      { name: 'Agent Assistant', avatar: '🤖', isAgent: true },
    ],
    lastMessage: 'Hi! I\'m your personal AI assistant. How can I help you today?',
    lastMessageType: 'text',
    timestamp: '2025-12-09T16:00:00Z',
    messageCount: 3,
  },
]

export function ChatsView() {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedChat, setSelectedChat] = useState<string | null>(null)

  const filteredChats = mockChats.filter(chat =>
    chat.participants.some(p => 
      p.name.toLowerCase().includes(searchQuery.toLowerCase())
    ) ||
    chat.lastMessage.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="h-[calc(100vh-12rem)] flex gap-6">
      {/* Chat List */}
      <div className="w-96 flex flex-col">
        <div className="relative mb-4">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input-field pl-10 w-full"
          />
        </div>
        
        <div className="card flex-1 overflow-auto">
          {filteredChats.map((chat) => (
            <motion.div
              key={chat.id}
              className={`p-4 border-b border-cyber-purple/10 cursor-pointer transition-colors ${
                selectedChat === chat.id ? 'bg-cyber-purple/10' : 'hover:bg-cyber-purple/5'
              }`}
              onClick={() => setSelectedChat(chat.id)}
              whileHover={{ x: 2 }}
            >
              <div className="flex items-start gap-3">
                {/* Avatars */}
                <div className="relative">
                  <div className="w-12 h-12 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center text-lg">
                    {chat.participants[0].avatar}
                  </div>
                  <div className="absolute -bottom-1 -right-1 w-6 h-6 rounded-full bg-cyber-dark border-2 border-cyber-gray flex items-center justify-center text-xs">
                    {chat.participants[1].avatar}
                  </div>
                </div>
                
                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-medium truncate">{chat.participants[0].name}</span>
                    <span className="text-xs text-gray-500">{formatTime(chat.timestamp)}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <MessageTypeIcon type={chat.lastMessageType} />
                    <p className="text-sm text-gray-400 truncate">{chat.lastMessage}</p>
                  </div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>

      {/* Chat Detail */}
      <div className="flex-1 card flex flex-col">
        {selectedChat ? (
          <ChatDetail chat={mockChats.find(c => c.id === selectedChat)!} />
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <MessageSquare className="w-16 h-16 text-cyber-purple/30 mx-auto mb-4" />
              <p className="text-gray-400">Select a conversation to view details</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function ChatDetail({ chat }: { chat: ChatPreview }) {
  return (
    <>
      {/* Header */}
      <div className="p-4 border-b border-cyber-purple/20 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center">
            {chat.participants[0].avatar}
          </div>
          <div>
            <h3 className="font-medium">{chat.participants[0].name}</h3>
            <p className="text-sm text-gray-500">{chat.messageCount} messages</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-secondary flex items-center gap-2 text-sm">
            <Eye className="w-4 h-4" />
            View Full
          </button>
          <button className="p-2 rounded-lg hover:bg-cyber-purple/20 text-gray-400">
            <MoreVertical className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Message Preview */}
      <div className="flex-1 p-4 overflow-auto">
        <div className="space-y-4">
          {/* User Message */}
          <div className="flex gap-3">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-pink flex items-center justify-center text-sm flex-shrink-0">
              {chat.participants[0].avatar}
            </div>
            <div className="bg-cyber-dark/50 rounded-2xl rounded-tl-sm px-4 py-2 max-w-md">
              <p className="text-sm">Can you generate a beautiful sunset image for me?</p>
              <span className="text-xs text-gray-500 mt-1">9:28 AM</span>
            </div>
          </div>

          {/* Agent Response */}
          <div className="flex gap-3 flex-row-reverse">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-cyber-green to-emerald-400 flex items-center justify-center text-sm flex-shrink-0">
              🤖
            </div>
            <div className="bg-cyber-purple/20 rounded-2xl rounded-tr-sm px-4 py-2 max-w-md">
              <p className="text-sm">{chat.lastMessage}</p>
              <div className="mt-2 bg-cyber-dark/50 rounded-lg p-2 flex items-center gap-2">
                <Image className="w-4 h-4 text-cyber-pink" />
                <span className="text-xs text-gray-400">Generated image attached</span>
              </div>
              <span className="text-xs text-gray-500 mt-1">9:30 AM</span>
            </div>
          </div>
        </div>
      </div>

      {/* Privacy Notice */}
      <div className="p-4 border-t border-cyber-purple/20 bg-cyber-dark/30">
        <p className="text-xs text-gray-500 text-center">
          🔒 Conversation monitoring is for admin purposes only. User data is encrypted.
        </p>
      </div>
    </>
  )
}

function MessageTypeIcon({ type }: { type: string }) {
  const icons: Record<string, { icon: React.ElementType; color: string }> = {
    text: { icon: MessageSquare, color: 'text-gray-400' },
    image: { icon: Image, color: 'text-cyber-pink' },
    voice: { icon: Mic, color: 'text-cyber-green' },
    location: { icon: MapPin, color: 'text-cyber-blue' },
  }
  
  const { icon: Icon, color } = icons[type] || icons.text
  return <Icon className={`w-3 h-3 ${color} flex-shrink-0`} />
}

function formatTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000)
  
  if (diffDays === 0) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } else if (diffDays === 1) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
  }
}
