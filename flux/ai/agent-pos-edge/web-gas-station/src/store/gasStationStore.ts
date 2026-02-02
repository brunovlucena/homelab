import { create } from 'zustand'

export type ViewType = 'dashboard' | 'tanks' | 'pumps' | 'transactions' | 'alerts' | 'agents' | 'cameras'
export type PumpStatus = 'idle' | 'active' | 'error' | 'maintenance'
export type TankStatus = 'normal' | 'low' | 'critical' | 'refilling'
export type FuelType = 'gasoline' | 'diesel' | 'premium' | 'ethanol'
export type AgentType = 'pump' | 'command-center' | 'pos-edge' | 'vision' | 'security' | 'analytics'
export type CameraStatus = 'online' | 'offline' | 'recording' | 'analyzing'
export type DetectionType = 'vehicle' | 'person' | 'license-plate' | 'safety-violation' | 'spill' | 'fire' | 'suspicious'

export interface Tank {
  id: string
  name: string
  fuelType: FuelType
  capacity: number
  currentLevel: number
  status: TankStatus
  lastReading: string
  temperature: number
}

export interface Pump {
  id: string
  number: number
  status: PumpStatus
  connectedTank: string
  fuelType: FuelType
  currentTransaction?: Transaction
  totalDispensed: number
  lastMaintenance: string
  cameraId?: string // Associated camera
}

export interface Transaction {
  id: string
  pumpId: string
  fuelType: FuelType
  liters: number
  amount: number
  startTime: string
  endTime?: string
  status: 'in_progress' | 'completed' | 'cancelled'
  paymentMethod?: 'cash' | 'card' | 'app'
  licensePlate?: string // AI detected
  vehicleType?: string // AI detected
}

export interface Alert {
  id: string
  type: 'warning' | 'critical' | 'info'
  title: string
  message: string
  source: string
  timestamp: string
  acknowledged: boolean
  aiGenerated?: boolean
  detectionId?: string
}

export interface Camera {
  id: string
  name: string
  location: string
  type: 'pump' | 'entrance' | 'store' | 'perimeter' | 'tank'
  status: CameraStatus
  resolution: string
  fps: number
  nightVision: boolean
  ptzEnabled: boolean
  connectedAgents: string[]
  lastFrame?: string
  streamUrl?: string
}

export interface AIDetection {
  id: string
  cameraId: string
  type: DetectionType
  confidence: number
  timestamp: string
  boundingBox?: { x: number; y: number; width: number; height: number }
  metadata: Record<string, string>
  processed: boolean
  agentId: string
  screenshot?: string
}

export interface Agent {
  id: string
  name: string
  type: AgentType
  status: 'online' | 'offline' | 'degraded' | 'processing'
  lastHeartbeat: string
  version: string
  description: string
  capabilities: string[]
  assignedCameras?: string[]
  processedToday: number
  metrics: {
    cpu: number
    memory: number
    requests: number
    inferenceTime?: number
  }
}

interface GasStationState {
  activeView: ViewType
  sidebarOpen: boolean
  tanks: Tank[]
  pumps: Pump[]
  transactions: Transaction[]
  alerts: Alert[]
  agents: Agent[]
  cameras: Camera[]
  detections: AIDetection[]
  stationName: string
  stationId: string
  
  // Data source tracking
  dataSource: 'live' | 'mock' | 'error'
  isLoading: boolean
  errorMessage: string | null
  lastFetched: string | null
  
  setActiveView: (view: ViewType) => void
  toggleSidebar: () => void
  updateTank: (id: string, data: Partial<Tank>) => void
  updatePump: (id: string, data: Partial<Pump>) => void
  addTransaction: (transaction: Transaction) => void
  addAlert: (alert: Alert) => void
  acknowledgeAlert: (id: string) => void
  addDetection: (detection: AIDetection) => void
  updateCamera: (id: string, data: Partial<Camera>) => void
  fetchLiveData: () => Promise<void>
  setDataSource: (source: 'live' | 'mock' | 'error', message?: string) => void
}

// Initial mock data
const initialTanks: Tank[] = [
  { id: 'tank-1', name: 'Tank A', fuelType: 'gasoline', capacity: 30000, currentLevel: 18500, status: 'normal', lastReading: new Date().toISOString(), temperature: 22.5 },
  { id: 'tank-2', name: 'Tank B', fuelType: 'diesel', capacity: 25000, currentLevel: 6200, status: 'low', lastReading: new Date().toISOString(), temperature: 21.8 },
  { id: 'tank-3', name: 'Tank C', fuelType: 'premium', capacity: 20000, currentLevel: 15800, status: 'normal', lastReading: new Date().toISOString(), temperature: 23.1 },
  { id: 'tank-4', name: 'Tank D', fuelType: 'ethanol', capacity: 15000, currentLevel: 2100, status: 'critical', lastReading: new Date().toISOString(), temperature: 22.0 },
]

const initialPumps: Pump[] = [
  { id: 'pump-1', number: 1, status: 'active', connectedTank: 'tank-1', fuelType: 'gasoline', totalDispensed: 45230, lastMaintenance: '2024-12-01', cameraId: 'cam-pump-1' },
  { id: 'pump-2', number: 2, status: 'idle', connectedTank: 'tank-1', fuelType: 'gasoline', totalDispensed: 38920, lastMaintenance: '2024-12-01', cameraId: 'cam-pump-2' },
  { id: 'pump-3', number: 3, status: 'active', connectedTank: 'tank-2', fuelType: 'diesel', totalDispensed: 62150, lastMaintenance: '2024-11-28', cameraId: 'cam-pump-3' },
  { id: 'pump-4', number: 4, status: 'idle', connectedTank: 'tank-2', fuelType: 'diesel', totalDispensed: 55420, lastMaintenance: '2024-11-28', cameraId: 'cam-pump-4' },
  { id: 'pump-5', number: 5, status: 'maintenance', connectedTank: 'tank-3', fuelType: 'premium', totalDispensed: 28760, lastMaintenance: '2024-12-05', cameraId: 'cam-pump-5' },
  { id: 'pump-6', number: 6, status: 'idle', connectedTank: 'tank-4', fuelType: 'ethanol', totalDispensed: 19340, lastMaintenance: '2024-12-03', cameraId: 'cam-pump-6' },
]

const initialTransactions: Transaction[] = [
  { id: 'txn-1', pumpId: 'pump-1', fuelType: 'gasoline', liters: 45.2, amount: 248.60, startTime: new Date(Date.now() - 120000).toISOString(), status: 'in_progress', licensePlate: 'ABC-1234', vehicleType: 'SUV' },
  { id: 'txn-2', pumpId: 'pump-3', fuelType: 'diesel', liters: 82.5, amount: 412.50, startTime: new Date(Date.now() - 300000).toISOString(), status: 'in_progress', licensePlate: 'XYZ-5678', vehicleType: 'Truck' },
  { id: 'txn-3', pumpId: 'pump-2', fuelType: 'gasoline', liters: 30.0, amount: 165.00, startTime: new Date(Date.now() - 600000).toISOString(), endTime: new Date(Date.now() - 540000).toISOString(), status: 'completed', paymentMethod: 'card', licensePlate: 'DEF-9012', vehicleType: 'Sedan' },
  { id: 'txn-4', pumpId: 'pump-4', fuelType: 'diesel', liters: 150.0, amount: 750.00, startTime: new Date(Date.now() - 900000).toISOString(), endTime: new Date(Date.now() - 780000).toISOString(), status: 'completed', paymentMethod: 'cash', licensePlate: 'GHI-3456', vehicleType: 'Van' },
]

const initialAlerts: Alert[] = [
  { id: 'alert-1', type: 'critical', title: 'Tank D Low Level', message: 'Ethanol tank is at 14% capacity. Schedule refill immediately.', source: 'tank-4', timestamp: new Date(Date.now() - 1800000).toISOString(), acknowledged: false },
  { id: 'alert-2', type: 'warning', title: 'Tank B Low Level', message: 'Diesel tank is at 25% capacity. Consider scheduling refill.', source: 'tank-2', timestamp: new Date(Date.now() - 3600000).toISOString(), acknowledged: false },
  { id: 'alert-3', type: 'warning', title: 'Pump 5 Maintenance', message: 'Pump 5 is under scheduled maintenance.', source: 'pump-5', timestamp: new Date(Date.now() - 7200000).toISOString(), acknowledged: true },
  { id: 'alert-4', type: 'info', title: 'Suspicious Activity Detected', message: 'Vision AI detected unusual behavior near pump 3. Review camera footage.', source: 'cam-pump-3', timestamp: new Date(Date.now() - 300000).toISOString(), acknowledged: false, aiGenerated: true, detectionId: 'det-3' },
]

const initialCameras: Camera[] = [
  { id: 'cam-pump-1', name: 'Pump 1 Camera', location: 'Pump Island 1', type: 'pump', status: 'recording', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1', 'agent-security'] },
  { id: 'cam-pump-2', name: 'Pump 2 Camera', location: 'Pump Island 1', type: 'pump', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1'] },
  { id: 'cam-pump-3', name: 'Pump 3 Camera', location: 'Pump Island 2', type: 'pump', status: 'analyzing', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1', 'agent-security'] },
  { id: 'cam-pump-4', name: 'Pump 4 Camera', location: 'Pump Island 2', type: 'pump', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1'] },
  { id: 'cam-pump-5', name: 'Pump 5 Camera', location: 'Pump Island 3', type: 'pump', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1'] },
  { id: 'cam-pump-6', name: 'Pump 6 Camera', location: 'Pump Island 3', type: 'pump', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-vision-1'] },
  { id: 'cam-entrance', name: 'Entrance Camera', location: 'Main Entrance', type: 'entrance', status: 'recording', resolution: '4K', fps: 60, nightVision: true, ptzEnabled: true, connectedAgents: ['agent-vision-2', 'agent-lpr'] },
  { id: 'cam-store', name: 'Store Camera', location: 'Convenience Store', type: 'store', status: 'recording', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: true, connectedAgents: ['agent-security'] },
  { id: 'cam-perimeter-1', name: 'Perimeter North', location: 'North Fence', type: 'perimeter', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: true, connectedAgents: ['agent-security'] },
  { id: 'cam-perimeter-2', name: 'Perimeter South', location: 'South Fence', type: 'perimeter', status: 'online', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: true, connectedAgents: ['agent-security'] },
  { id: 'cam-tanks', name: 'Tank Area Camera', location: 'Underground Tanks', type: 'tank', status: 'recording', resolution: '1080p', fps: 30, nightVision: true, ptzEnabled: false, connectedAgents: ['agent-safety'] },
]

const initialDetections: AIDetection[] = [
  { id: 'det-1', cameraId: 'cam-entrance', type: 'vehicle', confidence: 0.97, timestamp: new Date(Date.now() - 60000).toISOString(), metadata: { make: 'Toyota', model: 'Corolla', color: 'Silver' }, processed: true, agentId: 'agent-vision-2' },
  { id: 'det-2', cameraId: 'cam-entrance', type: 'license-plate', confidence: 0.95, timestamp: new Date(Date.now() - 60000).toISOString(), metadata: { plate: 'ABC-1234', state: 'SP' }, processed: true, agentId: 'agent-lpr' },
  { id: 'det-3', cameraId: 'cam-pump-3', type: 'suspicious', confidence: 0.78, timestamp: new Date(Date.now() - 300000).toISOString(), metadata: { description: 'Person lingering without vehicle', duration: '5min' }, processed: true, agentId: 'agent-security' },
  { id: 'det-4', cameraId: 'cam-pump-1', type: 'vehicle', confidence: 0.99, timestamp: new Date(Date.now() - 120000).toISOString(), metadata: { make: 'Honda', model: 'CR-V', color: 'Black', type: 'SUV' }, processed: true, agentId: 'agent-vision-1' },
  { id: 'det-5', cameraId: 'cam-tanks', type: 'person', confidence: 0.92, timestamp: new Date(Date.now() - 180000).toISOString(), metadata: { description: 'Authorized personnel', badge: 'STAFF-001' }, processed: true, agentId: 'agent-safety' },
]

const initialAgents: Agent[] = [
  { 
    id: 'agent-1', 
    name: 'pump-agent', 
    type: 'pump', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.1.0', 
    description: 'Monitors pump operations and transactions',
    capabilities: ['transaction-monitoring', 'pump-control', 'flow-analysis'],
    processedToday: 156,
    metrics: { cpu: 12, memory: 45, requests: 1250 } 
  },
  { 
    id: 'agent-2', 
    name: 'command-center', 
    type: 'command-center', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.1.0',
    description: 'Central coordination and alerting system',
    capabilities: ['alert-aggregation', 'reporting', 'notification'],
    processedToday: 892,
    metrics: { cpu: 8, memory: 38, requests: 890 } 
  },
  { 
    id: 'agent-3', 
    name: 'pos-edge', 
    type: 'pos-edge', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.1.0',
    description: 'Point-of-sale integration and payment processing',
    capabilities: ['payment-processing', 'receipt-generation', 'inventory-sync'],
    processedToday: 234,
    metrics: { cpu: 15, memory: 52, requests: 2100 } 
  },
  { 
    id: 'agent-vision-1', 
    name: 'vision-agent-pumps', 
    type: 'vision', 
    status: 'processing', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.2.0',
    description: 'AI vision analysis for pump area cameras',
    capabilities: ['vehicle-detection', 'person-detection', 'behavior-analysis', 'plate-recognition'],
    assignedCameras: ['cam-pump-1', 'cam-pump-2', 'cam-pump-3', 'cam-pump-4', 'cam-pump-5', 'cam-pump-6'],
    processedToday: 4521,
    metrics: { cpu: 78, memory: 85, requests: 4521, inferenceTime: 45 } 
  },
  { 
    id: 'agent-vision-2', 
    name: 'vision-agent-entrance', 
    type: 'vision', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.2.0',
    description: 'AI vision for entrance monitoring',
    capabilities: ['vehicle-detection', 'traffic-counting', 'queue-detection'],
    assignedCameras: ['cam-entrance'],
    processedToday: 1823,
    metrics: { cpu: 65, memory: 72, requests: 1823, inferenceTime: 38 } 
  },
  { 
    id: 'agent-lpr', 
    name: 'license-plate-agent', 
    type: 'vision', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v1.0.0',
    description: 'Automatic License Plate Recognition (ALPR)',
    capabilities: ['plate-recognition', 'vehicle-tracking', 'database-lookup'],
    assignedCameras: ['cam-entrance'],
    processedToday: 1823,
    metrics: { cpu: 45, memory: 60, requests: 1823, inferenceTime: 25 } 
  },
  { 
    id: 'agent-security', 
    name: 'security-agent', 
    type: 'security', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.3.0',
    description: 'Security monitoring and threat detection',
    capabilities: ['anomaly-detection', 'intrusion-detection', 'suspicious-behavior', 'alert-generation'],
    assignedCameras: ['cam-pump-1', 'cam-pump-3', 'cam-store', 'cam-perimeter-1', 'cam-perimeter-2'],
    processedToday: 3245,
    metrics: { cpu: 55, memory: 68, requests: 3245, inferenceTime: 52 } 
  },
  { 
    id: 'agent-safety', 
    name: 'safety-agent', 
    type: 'security', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.1.0',
    description: 'Safety compliance and hazard detection',
    capabilities: ['spill-detection', 'fire-detection', 'ppe-compliance', 'restricted-area'],
    assignedCameras: ['cam-tanks'],
    processedToday: 892,
    metrics: { cpu: 32, memory: 45, requests: 892, inferenceTime: 60 } 
  },
  { 
    id: 'agent-analytics', 
    name: 'analytics-agent', 
    type: 'analytics', 
    status: 'online', 
    lastHeartbeat: new Date().toISOString(), 
    version: 'v0.1.0',
    description: 'Business analytics and insights generation',
    capabilities: ['traffic-analysis', 'sales-prediction', 'customer-behavior', 'reporting'],
    processedToday: 24,
    metrics: { cpu: 25, memory: 55, requests: 24 } 
  },
]

export const useGasStationStore = create<GasStationState>((set, get) => ({
  activeView: 'dashboard',
  sidebarOpen: true,
  tanks: initialTanks,
  pumps: initialPumps,
  transactions: initialTransactions,
  alerts: initialAlerts,
  agents: initialAgents,
  cameras: initialCameras,
  detections: initialDetections,
  stationName: 'Shell Station #1247',
  stationId: 'GS-BR-SP-1247',
  
  // ⚠️ Data source tracking - defaults to MOCK with warning
  dataSource: 'mock',
  isLoading: false,
  errorMessage: '⚠️ Using MOCK data. Configure PROMETHEUS_URL and KUBERNETES_API_URL for live metrics.',
  lastFetched: null,
  
  setActiveView: (view) => set({ activeView: view }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  
  updateTank: (id, data) => set((state) => ({
    tanks: state.tanks.map(t => t.id === id ? { ...t, ...data } : t)
  })),
  
  updatePump: (id, data) => set((state) => ({
    pumps: state.pumps.map(p => p.id === id ? { ...p, ...data } : p)
  })),
  
  addTransaction: (transaction) => set((state) => ({
    transactions: [transaction, ...state.transactions]
  })),
  
  addAlert: (alert) => set((state) => ({
    alerts: [alert, ...state.alerts]
  })),
  
  acknowledgeAlert: (id) => set((state) => ({
    alerts: state.alerts.map(a => a.id === id ? { ...a, acknowledged: true } : a)
  })),

  addDetection: (detection) => set((state) => ({
    detections: [detection, ...state.detections]
  })),

  updateCamera: (id, data) => set((state) => ({
    cameras: state.cameras.map(c => c.id === id ? { ...c, ...data } : c)
  })),
  
  setDataSource: (source, message) => set({
    dataSource: source,
    errorMessage: message || null,
    lastFetched: new Date().toISOString(),
  }),
  
  // Fetch live data from APIs
  fetchLiveData: async () => {
    set({ isLoading: true })
    try {
      const response = await fetch('/api/metrics')
      const data = await response.json()
      
      if (data.success && data.agents?.length > 0) {
        // Map live agents to our format
        const liveAgents = data.agents.map((a: any, index: number) => ({
          ...initialAgents[index % initialAgents.length],
          name: a.name,
          status: a.status === 'online' ? 'online' : 'offline',
          version: a.version || 'unknown',
          lastHeartbeat: new Date().toISOString(),
        }))
        
        set({
          agents: liveAgents.length > 0 ? liveAgents : initialAgents,
          dataSource: 'live',
          errorMessage: null,
          lastFetched: new Date().toISOString(),
        })
      } else {
        set({
          dataSource: 'mock',
          errorMessage: data.message || '⚠️ Could not fetch live data - using MOCK data',
        })
      }
    } catch (error) {
      set({
        dataSource: 'error',
        errorMessage: `Failed to connect: ${error}`,
      })
    } finally {
      set({ isLoading: false })
    }
  },
}))
