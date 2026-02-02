'use client'

import { useGasStationStore, Camera, AIDetection, Agent } from '@/store/gasStationStore'
import { cn, getTimeSince } from '@/lib/utils'
import { motion, AnimatePresence } from 'framer-motion'
import { 
  Camera as CameraIcon, 
  Video, 
  Eye, 
  AlertTriangle, 
  Bot, 
  Play, 
  Pause, 
  Maximize2,
  Settings,
  Wifi,
  WifiOff,
  Scan,
  Car,
  User,
  Shield,
  Flame,
  Droplets,
  CircleAlert,
  Brain,
  Activity
} from 'lucide-react'
import { useState } from 'react'

function CameraFeed({ camera, detections }: { camera: Camera, detections: AIDetection[] }) {
  const { agents } = useGasStationStore()
  const [expanded, setExpanded] = useState(false)
  const cameraDetections = detections.filter(d => d.cameraId === camera.id).slice(0, 5)
  const connectedAgentData = agents.filter(a => camera.connectedAgents.includes(a.id))
  
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'recording': return 'fuel-red'
      case 'analyzing': return 'fuel-amber'
      case 'online': return 'fuel-green'
      default: return 'fuel-gray'
    }
  }

  const getDetectionIcon = (type: string) => {
    switch (type) {
      case 'vehicle': return Car
      case 'person': return User
      case 'license-plate': return Scan
      case 'suspicious': return CircleAlert
      case 'fire': return Flame
      case 'spill': return Droplets
      default: return Eye
    }
  }

  return (
    <motion.div
      layout
      className={cn(
        'fuel-card overflow-hidden transition-all',
        expanded ? 'col-span-2 row-span-2' : '',
        camera.status === 'analyzing' && 'border-fuel-amber/50'
      )}
    >
      {/* Camera Preview */}
      <div className="relative aspect-video bg-fuel-gray/50 overflow-hidden">
        {/* Simulated video feed */}
        <div className="absolute inset-0 bg-gradient-to-br from-fuel-gray/80 to-fuel-black flex items-center justify-center">
          <div className="text-center">
            <Video className="w-12 h-12 text-fuel-green/30 mx-auto mb-2" />
            <p className="text-xs text-gray-500 font-mono">{camera.resolution} @ {camera.fps}fps</p>
          </div>
        </div>
        
        {/* Overlay info */}
        <div className="absolute top-2 left-2 right-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className={cn(
              'flex items-center gap-1 px-2 py-1 rounded bg-fuel-black/80 backdrop-blur-sm',
            )}>
              <div className={cn(
                'w-2 h-2 rounded-full',
                camera.status === 'recording' && 'bg-fuel-red animate-pulse',
                camera.status === 'analyzing' && 'bg-fuel-amber animate-pulse',
                camera.status === 'online' && 'bg-fuel-green',
                camera.status === 'offline' && 'bg-fuel-red'
              )} />
              <span className="text-xs text-white font-mono uppercase">{camera.status}</span>
            </div>
            {camera.nightVision && (
              <div className="px-2 py-1 rounded bg-fuel-black/80 backdrop-blur-sm">
                <span className="text-xs text-fuel-green">NV</span>
              </div>
            )}
          </div>
          <button 
            onClick={() => setExpanded(!expanded)}
            className="p-1.5 rounded bg-fuel-black/80 backdrop-blur-sm hover:bg-fuel-gray/80 transition-colors"
          >
            <Maximize2 className="w-4 h-4 text-white" />
          </button>
        </div>

        {/* AI Detection overlay */}
        {camera.status === 'analyzing' && (
          <div className="absolute bottom-2 left-2 right-2">
            <div className="flex items-center gap-2 px-3 py-2 rounded bg-fuel-amber/20 backdrop-blur-sm border border-fuel-amber/30">
              <Brain className="w-4 h-4 text-fuel-amber animate-pulse" />
              <span className="text-xs text-fuel-amber">AI Analyzing...</span>
            </div>
          </div>
        )}

        {/* Recent detections badges */}
        {cameraDetections.length > 0 && (
          <div className="absolute bottom-2 right-2 flex gap-1">
            {cameraDetections.slice(0, 3).map((det) => {
              const Icon = getDetectionIcon(det.type)
              return (
                <div 
                  key={det.id}
                  className="p-1.5 rounded bg-fuel-black/80 backdrop-blur-sm"
                  title={`${det.type}: ${det.confidence * 100}%`}
                >
                  <Icon className="w-3 h-3 text-fuel-green" />
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
            <h3 className="font-bold text-white">{camera.name}</h3>
            <p className="text-xs text-gray-500">{camera.location}</p>
          </div>
          <span className={cn(
            'px-2 py-1 rounded text-xs font-medium capitalize',
            `bg-${getStatusColor(camera.status)}/20 text-${getStatusColor(camera.status)}`
          )}>
            {camera.type}
          </span>
        </div>

        {/* Connected AI Agents */}
        <div className="mb-3">
          <p className="text-xs text-gray-500 mb-2">Connected AI Agents:</p>
          <div className="flex flex-wrap gap-2">
            {connectedAgentData.map(agent => (
              <div 
                key={agent.id}
                className={cn(
                  'flex items-center gap-1 px-2 py-1 rounded text-xs',
                  agent.status === 'online' ? 'bg-fuel-green/10 text-fuel-green' :
                  agent.status === 'processing' ? 'bg-fuel-amber/10 text-fuel-amber' :
                  'bg-fuel-red/10 text-fuel-red'
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
            <p className="text-xs text-gray-500 mb-2">Recent AI Detections:</p>
            <div className="space-y-1">
              {cameraDetections.slice(0, 3).map(det => {
                const Icon = getDetectionIcon(det.type)
                return (
                  <div key={det.id} className="flex items-center justify-between p-2 rounded bg-fuel-gray/30 text-xs">
                    <div className="flex items-center gap-2">
                      <Icon className="w-3 h-3 text-fuel-green" />
                      <span className="text-white capitalize">{det.type.replace('-', ' ')}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-fuel-green font-mono">{(det.confidence * 100).toFixed(0)}%</span>
                      <span className="text-gray-500">{getTimeSince(det.timestamp)}</span>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Controls */}
        <div className="flex gap-2 mt-3 pt-3 border-t border-fuel-gray/30">
          <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded bg-fuel-gray/50 hover:bg-fuel-gray transition-colors text-xs text-gray-400">
            <Play className="w-3 h-3" />
            Live
          </button>
          <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded bg-fuel-gray/50 hover:bg-fuel-gray transition-colors text-xs text-gray-400">
            <Video className="w-3 h-3" />
            Playback
          </button>
          <button className="px-3 py-2 rounded bg-fuel-gray/50 hover:bg-fuel-gray transition-colors">
            <Settings className="w-3 h-3 text-gray-400" />
          </button>
        </div>
      </div>
    </motion.div>
  )
}

function AIAgentCard({ agent }: { agent: Agent }) {
  const { cameras } = useGasStationStore()
  const assignedCameraData = cameras.filter(c => agent.assignedCameras?.includes(c.id))

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'vision': return 'fuel-blue'
      case 'security': return 'fuel-red'
      case 'analytics': return 'fuel-amber'
      default: return 'fuel-green'
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="fuel-card p-4"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className={cn(
            'p-2 rounded-xl',
            `bg-${getTypeColor(agent.type)}/10 border border-${getTypeColor(agent.type)}/30`
          )}>
            <Brain className={cn('w-5 h-5', `text-${getTypeColor(agent.type)}`)} />
          </div>
          <div>
            <h3 className="font-bold text-white">{agent.name}</h3>
            <p className="text-xs text-gray-500">{agent.description}</p>
          </div>
        </div>
        <span className={cn(
          'px-2 py-1 rounded text-xs font-medium',
          agent.status === 'online' ? 'bg-fuel-green/20 text-fuel-green' :
          agent.status === 'processing' ? 'bg-fuel-amber/20 text-fuel-amber animate-pulse' :
          'bg-fuel-red/20 text-fuel-red'
        )}>
          {agent.status}
        </span>
      </div>

      {/* Capabilities */}
      <div className="flex flex-wrap gap-1 mb-3">
        {agent.capabilities.slice(0, 4).map(cap => (
          <span key={cap} className="px-2 py-0.5 rounded bg-fuel-gray/50 text-xs text-gray-400">
            {cap}
          </span>
        ))}
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-3 gap-2 mb-3">
        <div className="p-2 rounded bg-fuel-gray/30 text-center">
          <p className="text-lg font-bold text-white">{agent.processedToday}</p>
          <p className="text-xs text-gray-500">Processed</p>
        </div>
        <div className="p-2 rounded bg-fuel-gray/30 text-center">
          <p className="text-lg font-bold text-white">{agent.metrics.cpu}%</p>
          <p className="text-xs text-gray-500">CPU</p>
        </div>
        {agent.metrics.inferenceTime && (
          <div className="p-2 rounded bg-fuel-gray/30 text-center">
            <p className="text-lg font-bold text-white">{agent.metrics.inferenceTime}ms</p>
            <p className="text-xs text-gray-500">Inference</p>
          </div>
        )}
      </div>

      {/* Assigned Cameras */}
      {assignedCameraData.length > 0 && (
        <div>
          <p className="text-xs text-gray-500 mb-2">Monitoring {assignedCameraData.length} camera(s)</p>
          <div className="flex flex-wrap gap-1">
            {assignedCameraData.slice(0, 4).map(cam => (
              <span key={cam.id} className="px-2 py-1 rounded bg-fuel-gray/50 text-xs text-gray-400">
                {cam.name.replace(' Camera', '')}
              </span>
            ))}
            {assignedCameraData.length > 4 && (
              <span className="px-2 py-1 rounded bg-fuel-gray/50 text-xs text-gray-400">
                +{assignedCameraData.length - 4}
              </span>
            )}
          </div>
        </div>
      )}
    </motion.div>
  )
}

export function SmartCameras() {
  const { cameras, detections, agents, setActiveView } = useGasStationStore()
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [filter, setFilter] = useState<'all' | 'pump' | 'entrance' | 'perimeter'>('all')

  const filteredCameras = cameras.filter(c => filter === 'all' || c.type === filter)
  const visionAgents = agents.filter(a => a.type === 'vision' || a.type === 'security')
  const onlineCameras = cameras.filter(c => c.status !== 'offline').length
  const totalDetections = detections.length
  const activeAgents = visionAgents.filter(a => a.status === 'online' || a.status === 'processing').length

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Smart Cameras & AI Vision</h1>
          <p className="text-gray-400">Real-time video analytics powered by AI agents</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="fuel-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-fuel-green/10 border border-fuel-green/30">
              <CameraIcon className="w-5 h-5 text-fuel-green" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Cameras Online</p>
              <p className="text-2xl font-bold text-white">{onlineCameras}/{cameras.length}</p>
            </div>
          </div>
        </motion.div>
        
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="fuel-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-fuel-blue/10 border border-fuel-blue/30">
              <Brain className="w-5 h-5 text-fuel-blue" />
            </div>
            <div>
              <p className="text-sm text-gray-400">AI Agents Active</p>
              <p className="text-2xl font-bold text-white">{activeAgents}/{visionAgents.length}</p>
            </div>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="fuel-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-fuel-amber/10 border border-fuel-amber/30">
              <Scan className="w-5 h-5 text-fuel-amber" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Detections Today</p>
              <p className="text-2xl font-bold text-white">{totalDetections}</p>
            </div>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }} className="fuel-card p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-fuel-red/10 border border-fuel-red/30">
              <Shield className="w-5 h-5 text-fuel-red" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Security Alerts</p>
              <p className="text-2xl font-bold text-white">{detections.filter(d => d.type === 'suspicious').length}</p>
            </div>
          </div>
        </motion.div>
      </div>

      {/* AI Agents Section */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-white">AI Vision Agents</h2>
          <button onClick={() => setActiveView('agents')} className="text-sm text-fuel-green hover:underline">
            View All Agents →
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {visionAgents.slice(0, 6).map(agent => (
            <AIAgentCard key={agent.id} agent={agent} />
          ))}
        </div>
      </div>

      {/* Camera Filters */}
      <div className="flex items-center gap-4">
        <div className="flex gap-2">
          {['all', 'pump', 'entrance', 'perimeter'].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f as typeof filter)}
              className={cn(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === f 
                  ? 'bg-fuel-green/20 text-fuel-green border border-fuel-green/30' 
                  : 'bg-fuel-gray/50 text-gray-400 hover:bg-fuel-gray'
              )}
            >
              {f === 'all' ? 'All Cameras' : f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Camera Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredCameras.map((camera) => (
          <CameraFeed key={camera.id} camera={camera} detections={detections} />
        ))}
      </div>

      {/* Recent AI Detections */}
      <div className="fuel-card p-5">
        <h2 className="text-lg font-bold text-white mb-4">Recent AI Detections</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-fuel-gray/30">
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Type</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Camera</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Agent</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Confidence</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Details</th>
                <th className="px-4 py-3 text-left text-xs text-gray-500 uppercase">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-fuel-gray/20">
              {detections.slice(0, 10).map(det => {
                const camera = cameras.find(c => c.id === det.cameraId)
                const agent = agents.find(a => a.id === det.agentId)
                const Icon = det.type === 'vehicle' ? Car : det.type === 'person' ? User : det.type === 'license-plate' ? Scan : CircleAlert
                
                return (
                  <tr key={det.id} className="hover:bg-fuel-gray/20">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Icon className="w-4 h-4 text-fuel-green" />
                        <span className="text-sm text-white capitalize">{det.type.replace('-', ' ')}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-400">{camera?.name}</td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 rounded bg-fuel-blue/10 text-fuel-blue text-xs">{agent?.name}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={cn(
                        'font-mono text-sm',
                        det.confidence >= 0.9 ? 'text-fuel-green' : det.confidence >= 0.7 ? 'text-fuel-amber' : 'text-fuel-red'
                      )}>
                        {(det.confidence * 100).toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-400">
                      {Object.entries(det.metadata).slice(0, 2).map(([k, v]) => `${k}: ${v}`).join(', ')}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">{getTimeSince(det.timestamp)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
