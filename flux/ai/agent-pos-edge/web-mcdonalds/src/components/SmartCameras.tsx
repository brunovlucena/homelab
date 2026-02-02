'use client'

import { useMcdonaldsStore, Camera, AIDetection } from '@/store/mcdonaldsStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion } from 'framer-motion'
import { 
  Camera as CameraIcon, 
  Video, 
  Eye, 
  Bot, 
  Play, 
  Maximize2,
  Settings,
  Brain,
  Users,
  Car,
  Utensils,
  Sparkles,
  ShieldCheck,
  Activity
} from 'lucide-react'
import { useState } from 'react'

function CameraFeed({ camera, detections }: { camera: Camera, detections: AIDetection[] }) {
  const { agents } = useMcdonaldsStore()
  const [expanded, setExpanded] = useState(false)
  const cameraDetections = detections.filter(d => d.cameraId === camera.id).slice(0, 5)
  const connectedAgentData = agents.filter(a => camera.connectedAgents.includes(a.id))
  
  const getDetectionIcon = (type: string) => {
    switch (type) {
      case 'customer': return Users
      case 'queue': return Users
      case 'food-quality': return Sparkles
      case 'cleanliness': return Sparkles
      case 'safety': return ShieldCheck
      case 'drive-thru-vehicle': return Car
      default: return Eye
    }
  }

  return (
    <motion.div
      layout
      className={cn(
        'mc-card overflow-hidden transition-all',
        expanded ? 'col-span-2 row-span-2' : '',
        camera.status === 'analyzing' && 'border-mc-gold/50'
      )}
    >
      {/* Camera Preview */}
      <div className="relative aspect-video bg-mc-gray/50 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-mc-gray/80 to-mc-black flex items-center justify-center">
          <div className="text-center">
            <Video className="w-12 h-12 text-mc-red/30 mx-auto mb-2" />
            <p className="text-xs text-gray-500 font-mono">{camera.resolution} @ {camera.fps}fps</p>
          </div>
        </div>
        
        {/* Overlay info */}
        <div className="absolute top-2 left-2 right-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1 px-2 py-1 rounded bg-mc-black/80 backdrop-blur-sm">
              <div className={cn(
                'w-2 h-2 rounded-full',
                camera.status === 'recording' && 'bg-mc-red animate-pulse',
                camera.status === 'analyzing' && 'bg-mc-gold animate-pulse',
                camera.status === 'online' && 'bg-mc-green',
                camera.status === 'offline' && 'bg-mc-red'
              )} />
              <span className="text-xs text-white font-mono uppercase">{camera.status}</span>
            </div>
          </div>
          <button 
            onClick={() => setExpanded(!expanded)}
            className="p-1.5 rounded bg-mc-black/80 backdrop-blur-sm hover:bg-mc-gray/80 transition-colors"
          >
            <Maximize2 className="w-4 h-4 text-white" />
          </button>
        </div>

        {/* AI Detection overlay */}
        {camera.status === 'analyzing' && (
          <div className="absolute bottom-2 left-2 right-2">
            <div className="flex items-center gap-2 px-3 py-2 rounded bg-mc-gold/20 backdrop-blur-sm border border-mc-gold/30">
              <Brain className="w-4 h-4 text-mc-gold animate-pulse" />
              <span className="text-xs text-mc-gold">AI Processing...</span>
            </div>
          </div>
        )}

        {cameraDetections.length > 0 && (
          <div className="absolute bottom-2 right-2 flex gap-1">
            {cameraDetections.slice(0, 3).map((det) => {
              const Icon = getDetectionIcon(det.type)
              return (
                <div 
                  key={det.id}
                  className="p-1.5 rounded bg-mc-black/80 backdrop-blur-sm"
                  title={`${det.type}: ${det.confidence * 100}%`}
                >
                  <Icon className="w-3 h-3 text-mc-green" />
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Camera Info */}
      <div className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h3 className="font-brand font-bold text-white">{camera.name}</h3>
            <p className="text-xs text-gray-500">{camera.location}</p>
          </div>
          <span className={cn(
            'px-2 py-1 rounded text-xs font-medium capitalize',
            camera.type === 'kitchen' ? 'bg-mc-orange/20 text-mc-orange' :
            camera.type === 'drive-thru' ? 'bg-mc-gold/20 text-mc-gold' :
            'bg-mc-blue/20 text-mc-blue'
          )}>
            {camera.type}
          </span>
        </div>

        {/* Connected AI Agents */}
        <div className="mb-3">
          <p className="text-xs text-gray-500 mb-2">AI Agents:</p>
          <div className="flex flex-wrap gap-2">
            {connectedAgentData.map(agent => (
              <div 
                key={agent.id}
                className={cn(
                  'flex items-center gap-1 px-2 py-1 rounded text-xs',
                  agent.status === 'online' ? 'bg-mc-green/10 text-mc-green' :
                  agent.status === 'processing' ? 'bg-mc-gold/10 text-mc-gold' :
                  'bg-mc-red/10 text-mc-red'
                )}
              >
                <Bot className="w-3 h-3" />
                {agent.name.replace('-agent', '').replace('agent-', '')}
              </div>
            ))}
          </div>
        </div>

        {/* Recent Detections */}
        {cameraDetections.length > 0 && (
          <div>
            <p className="text-xs text-gray-500 mb-2">Detecções Recentes:</p>
            <div className="space-y-1">
              {cameraDetections.slice(0, 3).map(det => {
                const Icon = getDetectionIcon(det.type)
                return (
                  <div key={det.id} className="flex items-center justify-between p-2 rounded bg-mc-gray/30 text-xs">
                    <div className="flex items-center gap-2">
                      <Icon className="w-3 h-3 text-mc-gold" />
                      <span className="text-white capitalize">{det.type.replace('-', ' ')}</span>
                    </div>
                    <span className="text-mc-green font-mono">{(det.confidence * 100).toFixed(0)}%</span>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Controls */}
        <div className="flex gap-2 mt-3 pt-3 border-t border-mc-gray/30">
          <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded bg-mc-gray/50 hover:bg-mc-gray transition-colors text-xs text-gray-400">
            <Play className="w-3 h-3" />
            Ao Vivo
          </button>
          <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded bg-mc-gray/50 hover:bg-mc-gray transition-colors text-xs text-gray-400">
            <Video className="w-3 h-3" />
            Gravações
          </button>
        </div>
      </div>
    </motion.div>
  )
}

export function SmartCameras() {
  const { cameras, detections, agents, setActiveView } = useMcdonaldsStore()
  const [filter, setFilter] = useState<'all' | 'kitchen' | 'counter' | 'drive-thru'>('all')

  const filteredCameras = cameras.filter(c => filter === 'all' || c.type === filter)
  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'quality' || a.type === 'customer')
  const onlineCameras = cameras.filter(c => c.status !== 'offline').length
  const activeAgents = visionAgents.filter(a => a.status === 'online' || a.status === 'processing').length

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-brand font-bold text-white">Câmeras & AI Vision</h1>
          <p className="text-gray-400">Monitoramento inteligente da operação</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="mc-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-mc-green/10 border border-mc-green/30">
              <CameraIcon className="w-5 h-5 text-mc-green" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Câmeras Online</p>
              <p className="text-2xl font-bold text-white">{onlineCameras}/{cameras.length}</p>
            </div>
          </div>
        </motion.div>
        
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="mc-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-mc-gold/10 border border-mc-gold/30">
              <Brain className="w-5 h-5 text-mc-gold" />
            </div>
            <div>
              <p className="text-sm text-gray-400">AI Agents</p>
              <p className="text-2xl font-bold text-white">{activeAgents}/{visionAgents.length}</p>
            </div>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="mc-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-mc-blue/10 border border-mc-blue/30">
              <Activity className="w-5 h-5 text-mc-blue" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Detecções Hoje</p>
              <p className="text-2xl font-bold text-white">{detections.length}</p>
            </div>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }} className="mc-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-mc-red/10 border border-mc-red/30">
              <Sparkles className="w-5 h-5 text-mc-red" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Quality Score</p>
              <p className="text-2xl font-bold text-white">95%</p>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Filter */}
      <div className="flex gap-2">
        {['all', 'kitchen', 'counter', 'drive-thru'].map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f as typeof filter)}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              filter === f 
                ? 'bg-mc-red text-white' 
                : 'bg-mc-gray/50 text-gray-400 hover:bg-mc-gray'
            )}
          >
            {f === 'all' ? 'Todas' : f === 'drive-thru' ? 'Drive-Thru' : f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
      </div>

      {/* Camera Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredCameras.map((camera) => (
          <CameraFeed key={camera.id} camera={camera} detections={detections} />
        ))}
      </div>
    </div>
  )
}
