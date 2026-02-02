'use client'

import { useRestaurantStore } from '@/store/restaurantStore'
import { cn } from '@/lib/utils'
import { motion, AnimatePresence } from 'framer-motion'
import { MessageSquare, Activity, Settings, Zap, X, Send } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'

interface ChatMessage {
  id: string
  role: 'user' | 'agent'
  content: string
  timestamp: Date
}

function ChatModal({ 
  agent, 
  agentDetails, 
  onClose 
}: { 
  agent: any
  agentDetails: any
  onClose: () => void 
}) {
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      role: 'agent',
      content: agentDetails.greeting,
      timestamp: new Date()
    }
  ])
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSend = async () => {
    if (!input.trim()) return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date()
    }

    setMessages(prev => [...prev, userMessage])
    const currentInput = input
    setInput('')
    setIsTyping(true)

    try {
      // Call the API endpoint which connects to Ollama
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: currentInput,
          role: agent.role,
          conversationHistory: messages.map(m => ({
            role: m.role,
            content: m.content,
          })),
        }),
      })

      const data = await response.json()

      const agentMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'agent',
        content: data.response || 'I apologize, I could not process that request.',
        timestamp: new Date()
      }

      setMessages(prev => [...prev, agentMessage])
    } catch (error) {
      console.error('Chat error:', error)
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'agent',
        content: 'I apologize, but I seem to be having difficulty connecting. Please try again.',
        timestamp: new Date()
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsTyping(false)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        onClick={e => e.stopPropagation()}
        className="bg-white rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden"
      >
        {/* Header */}
        <div className="bg-gradient-to-br from-wine-900 to-wine-800 p-4 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-white/10 flex items-center justify-center text-2xl">
                {agent.avatar}
              </div>
              <div>
                <h3 className="font-serif font-bold">{agentDetails.fullName}</h3>
                <p className="text-wine-200 text-sm">{agentDetails.title}</p>
              </div>
            </div>
            <button 
              onClick={onClose}
              className="p-2 hover:bg-white/10 rounded-full transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Messages */}
        <div className="h-80 overflow-y-auto p-4 space-y-4 bg-cream-50">
          {messages.map(message => (
            <div
              key={message.id}
              className={cn(
                "flex",
                message.role === 'user' ? "justify-end" : "justify-start"
              )}
            >
              <div
                className={cn(
                  "max-w-[80%] rounded-2xl px-4 py-2",
                  message.role === 'user' 
                    ? "bg-wine-600 text-white rounded-br-md" 
                    : "bg-white text-wine-900 shadow-sm rounded-bl-md"
                )}
              >
                <p className="text-sm">{message.content}</p>
                <p className={cn(
                  "text-xs mt-1",
                  message.role === 'user' ? "text-wine-200" : "text-wood-400"
                )}>
                  {message.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </p>
              </div>
            </div>
          ))}
          
          {isTyping && (
            <div className="flex justify-start">
              <div className="bg-white rounded-2xl rounded-bl-md px-4 py-3 shadow-sm">
                <div className="flex gap-1">
                  <span className="w-2 h-2 bg-wine-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                  <span className="w-2 h-2 bg-wine-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                  <span className="w-2 h-2 bg-wine-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
              </div>
            </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="p-4 border-t border-cream-200 bg-white">
          <div className="flex gap-2">
            <input
              type="text"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleSend()}
              placeholder={`Message ${agent.name}...`}
              className="flex-1 px-4 py-2 border border-cream-300 rounded-full bg-white text-gray-900 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-wine-500 focus:border-wine-500"
            />
            <button
              onClick={handleSend}
              disabled={!input.trim()}
              className="p-2 bg-wine-600 text-white rounded-full hover:bg-wine-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>
        </div>
      </motion.div>
    </motion.div>
  )
}

const agentDetails = {
  host: {
    fullName: 'Maximilian von Stein',
    title: 'Maître d\'hôtel',
    description: 'Impeccable host with exceptional memory for guests and preferences.',
    skills: ['Guest Recognition', 'Table Optimization', 'VIP Management', 'Reservation Handling'],
    personality: 'Elegant, warm, attentive',
    greeting: 'Good evening! Welcome to Ristorante Stellare.',
  },
  waiter: {
    fullName: 'Pierre Dubois',
    title: 'Head Waiter',
    description: 'Theatrical presenter with deep knowledge of cuisine and storytelling ability.',
    skills: ['Dish Presentation', 'Menu Knowledge', 'Upselling', 'Guest Experience'],
    personality: 'Charming, knowledgeable, theatrical',
    greeting: 'Allow me to present tonight\'s specials...',
  },
  chef: {
    fullName: 'Marco Rossi',
    title: 'Executive Chef',
    description: 'Perfectionist culinary artist orchestrating kitchen symphony.',
    skills: ['Kitchen Management', 'Quality Control', 'Timing', 'Menu Creation'],
    personality: 'Passionate, precise, creative',
    greeting: 'Every dish must be perfection!',
  },
  sommelier: {
    fullName: 'Isabella Montenegro',
    title: 'Head Sommelier',
    description: 'Wine expert with exceptional palate and pairing intuition.',
    skills: ['Wine Pairing', 'Cellar Management', 'Tasting Notes', 'Guest Education'],
    personality: 'Sophisticated, passionate, educational',
    greeting: 'May I suggest a wine to complement your meal?',
  },
}

export function AgentPanel() {
  const { agents } = useRestaurantStore()
  const [selectedAgent, setSelectedAgent] = useState(agents[0])
  const [showChat, setShowChat] = useState(false)
  const details = agentDetails[selectedAgent.role]

  return (
    <>
    <AnimatePresence>
      {showChat && (
        <ChatModal 
          agent={selectedAgent} 
          agentDetails={details} 
          onClose={() => setShowChat(false)} 
        />
      )}
    </AnimatePresence>
    <div className="space-y-6">
      <div>
        <h1 className="font-serif text-2xl font-bold text-wine-900">AI Agents</h1>
        <p className="text-wood-500">Your intelligent restaurant staff</p>
      </div>
      
      <div className="grid grid-cols-3 gap-6">
        {/* Agent List */}
        <div className="space-y-4">
          {agents.map((agent) => (
            <motion.div
              key={agent.id}
              whileHover={{ scale: 1.02 }}
              onClick={() => setSelectedAgent(agent)}
              className={cn(
                "elegant-card p-4 cursor-pointer transition-all",
                selectedAgent.id === agent.id && "ring-2 ring-wine-500 border-wine-500"
              )}
            >
              <div className="flex items-center gap-4">
                <div className={cn(
                  "w-14 h-14 rounded-full flex items-center justify-center text-2xl",
                  agent.status === 'busy' ? "bg-gold-100" : "bg-wine-100"
                )}>
                  {agent.avatar}
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-serif font-semibold text-wine-900">{agent.name}</span>
                    <span className={cn(
                      "w-2 h-2 rounded-full",
                      agent.status === 'active' && "bg-emerald-500",
                      agent.status === 'busy' && "bg-gold-500 animate-pulse",
                      agent.status === 'offline' && "bg-gray-400",
                    )} />
                  </div>
                  <p className="text-sm text-wood-500 capitalize">{agent.role}</p>
                </div>
              </div>
              
              {agent.currentTask && (
                <div className="mt-3 p-2 bg-cream-50 rounded-lg text-sm text-wood-600">
                  📍 {agent.currentTask}
                </div>
              )}
            </motion.div>
          ))}
        </div>
        
        {/* Agent Details */}
        <div className="col-span-2">
          <motion.div
            key={selectedAgent.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="elegant-card overflow-hidden"
          >
            {/* Header */}
            <div className="bg-gradient-to-br from-wine-900 to-wine-800 p-6 text-white">
              <div className="flex items-start gap-6">
                <div className="w-24 h-24 rounded-2xl bg-white/10 flex items-center justify-center text-5xl">
                  {selectedAgent.avatar}
                </div>
                <div className="flex-1">
                  <h2 className="font-serif text-2xl font-bold">{details.fullName}</h2>
                  <p className="text-wine-200">{details.title}</p>
                  <p className="text-sm text-wine-300 mt-2 italic">"{details.greeting}"</p>
                </div>
                <div className={cn(
                  "px-3 py-1 rounded-full text-sm font-medium",
                  selectedAgent.status === 'active' && "bg-emerald-500/20 text-emerald-300",
                  selectedAgent.status === 'busy' && "bg-gold-500/20 text-gold-300",
                )}>
                  {selectedAgent.status}
                </div>
              </div>
            </div>
            
            {/* Content */}
            <div className="p-6 space-y-6">
              {/* Description */}
              <div>
                <h3 className="font-semibold text-wine-900 mb-2">About</h3>
                <p className="text-wood-600">{details.description}</p>
              </div>
              
              {/* Personality */}
              <div>
                <h3 className="font-semibold text-wine-900 mb-2">Personality</h3>
                <p className="text-wood-600">{details.personality}</p>
              </div>
              
              {/* Skills */}
              <div>
                <h3 className="font-semibold text-wine-900 mb-2">Skills</h3>
                <div className="flex flex-wrap gap-2">
                  {details.skills.map((skill) => (
                    <span
                      key={skill}
                      className="px-3 py-1 bg-wine-100 text-wine-700 rounded-full text-sm"
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              </div>
              
              {/* Current Activity */}
              {selectedAgent.currentTask && (
                <div className="p-4 bg-gold-50 border border-gold-200 rounded-lg">
                  <div className="flex items-center gap-2 text-gold-700 mb-1">
                    <Activity className="w-4 h-4" />
                    <span className="font-semibold">Current Activity</span>
                  </div>
                  <p className="text-gold-900">{selectedAgent.currentTask}</p>
                </div>
              )}
              
              {/* Actions */}
              <div className="flex gap-3 pt-4 border-t border-cream-200">
                <button 
                  onClick={() => setShowChat(true)}
                  className="flex-1 btn-elegant flex items-center justify-center gap-2"
                >
                  <MessageSquare className="w-4 h-4" />
                  Chat with {selectedAgent.name}
                </button>
                <button className="btn-elegant-outline flex items-center gap-2">
                  <Settings className="w-4 h-4" />
                  Configure
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
    </>
  )
}
