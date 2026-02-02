import { create } from 'zustand'

export type ViewType = 'dashboard' | 'kitchen' | 'orders' | 'menu' | 'staff' | 'analytics' | 'cameras' | 'agents'
export type OrderStatus = 'new' | 'preparing' | 'ready' | 'delivered' | 'cancelled'
export type OrderType = 'dine-in' | 'drive-thru' | 'delivery' | 'takeaway'
export type AgentType = 'kitchen' | 'order' | 'vision' | 'quality' | 'inventory' | 'customer'
export type CameraStatus = 'online' | 'offline' | 'recording' | 'analyzing'
export type DetectionType = 'customer' | 'queue' | 'food-quality' | 'cleanliness' | 'safety' | 'drive-thru-vehicle'

export interface MenuItem {
  id: string
  name: string
  category: 'burgers' | 'sides' | 'drinks' | 'desserts' | 'breakfast' | 'happy-meal'
  price: number
  prepTime: number // in seconds
  image?: string
  available: boolean
}

export interface OrderItem {
  menuItem: MenuItem
  quantity: number
  customizations?: string[]
}

export interface Order {
  id: string
  orderNumber: number
  type: OrderType
  items: OrderItem[]
  status: OrderStatus
  total: number
  createdAt: string
  startedAt?: string
  completedAt?: string
  estimatedTime: number // in seconds
  station?: string
  priority: 'normal' | 'rush' | 'vip'
  customerAnalysis?: {
    sentiment: 'happy' | 'neutral' | 'frustrated'
    waitTime: number
  }
}

export interface KitchenStation {
  id: string
  name: string
  type: 'grill' | 'fryer' | 'assembly' | 'drinks' | 'desserts'
  activeOrders: number
  capacity: number
  status: 'active' | 'busy' | 'offline'
  cameraId?: string
}

export interface Staff {
  id: string
  name: string
  role: 'manager' | 'crew' | 'kitchen' | 'cashier'
  station?: string
  status: 'active' | 'break' | 'offline'
  ordersCompleted: number
}

export interface Camera {
  id: string
  name: string
  location: string
  type: 'kitchen' | 'counter' | 'drive-thru' | 'dining' | 'entrance'
  status: CameraStatus
  resolution: string
  fps: number
  connectedAgents: string[]
}

export interface AIDetection {
  id: string
  cameraId: string
  type: DetectionType
  confidence: number
  timestamp: string
  metadata: Record<string, string>
  processed: boolean
  agentId: string
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

interface McdonaldsState {
  activeView: ViewType
  sidebarOpen: boolean
  orders: Order[]
  menuItems: MenuItem[]
  stations: KitchenStation[]
  staff: Staff[]
  cameras: Camera[]
  agents: Agent[]
  detections: AIDetection[]
  storeName: string
  storeId: string
  setActiveView: (view: ViewType) => void
  toggleSidebar: () => void
  updateOrderStatus: (id: string, status: OrderStatus) => void
  addOrder: (order: Order) => void
  addDetection: (detection: AIDetection) => void
}

// Initial mock data
const initialMenuItems: MenuItem[] = [
  { id: 'big-mac', name: 'Big Mac', category: 'burgers', price: 25.90, prepTime: 180, available: true },
  { id: 'quarter', name: 'Quarteirão', category: 'burgers', price: 27.90, prepTime: 200, available: true },
  { id: 'mcchicken', name: 'McChicken', category: 'burgers', price: 19.90, prepTime: 150, available: true },
  { id: 'cheddar', name: 'McChicken Cheddar', category: 'burgers', price: 22.90, prepTime: 160, available: true },
  { id: 'fries-m', name: 'McFritas Média', category: 'sides', price: 12.90, prepTime: 120, available: true },
  { id: 'fries-g', name: 'McFritas Grande', category: 'sides', price: 15.90, prepTime: 120, available: true },
  { id: 'nuggets-6', name: 'McNuggets 6un', category: 'sides', price: 16.90, prepTime: 150, available: true },
  { id: 'coke-m', name: 'Coca-Cola Média', category: 'drinks', price: 9.90, prepTime: 30, available: true },
  { id: 'mcshake', name: 'McShake Ovomaltine', category: 'desserts', price: 14.90, prepTime: 60, available: true },
  { id: 'sundae', name: 'Sundae', category: 'desserts', price: 8.90, prepTime: 45, available: true },
]

const initialOrders: Order[] = [
  {
    id: 'ord-001', orderNumber: 147, type: 'dine-in',
    items: [{ menuItem: initialMenuItems[0], quantity: 2 }, { menuItem: initialMenuItems[4], quantity: 2 }, { menuItem: initialMenuItems[7], quantity: 2 }],
    status: 'preparing', total: 97.40, createdAt: new Date(Date.now() - 180000).toISOString(),
    startedAt: new Date(Date.now() - 120000).toISOString(), estimatedTime: 180, station: 'grill-1', priority: 'normal',
    customerAnalysis: { sentiment: 'neutral', waitTime: 120 }
  },
  {
    id: 'ord-002', orderNumber: 148, type: 'drive-thru',
    items: [{ menuItem: initialMenuItems[1], quantity: 1 }, { menuItem: initialMenuItems[5], quantity: 1 }, { menuItem: initialMenuItems[6], quantity: 1 }],
    status: 'preparing', total: 60.70, createdAt: new Date(Date.now() - 120000).toISOString(),
    startedAt: new Date(Date.now() - 90000).toISOString(), estimatedTime: 200, station: 'grill-2', priority: 'rush'
  },
  {
    id: 'ord-003', orderNumber: 149, type: 'delivery',
    items: [{ menuItem: initialMenuItems[2], quantity: 3 }, { menuItem: initialMenuItems[4], quantity: 3 }],
    status: 'new', total: 98.40, createdAt: new Date(Date.now() - 60000).toISOString(), estimatedTime: 150, priority: 'normal'
  },
  {
    id: 'ord-004', orderNumber: 150, type: 'takeaway',
    items: [{ menuItem: initialMenuItems[8], quantity: 2 }, { menuItem: initialMenuItems[9], quantity: 1 }],
    status: 'ready', total: 38.70, createdAt: new Date(Date.now() - 300000).toISOString(),
    startedAt: new Date(Date.now() - 240000).toISOString(), completedAt: new Date(Date.now() - 60000).toISOString(),
    estimatedTime: 60, priority: 'normal'
  },
  {
    id: 'ord-005', orderNumber: 151, type: 'dine-in',
    items: [{ menuItem: initialMenuItems[3], quantity: 1 }, { menuItem: initialMenuItems[4], quantity: 1 }, { menuItem: initialMenuItems[7], quantity: 1 }],
    status: 'new', total: 45.70, createdAt: new Date(Date.now() - 30000).toISOString(), estimatedTime: 160, priority: 'vip',
    customerAnalysis: { sentiment: 'happy', waitTime: 30 }
  },
]

const initialStations: KitchenStation[] = [
  { id: 'grill-1', name: 'Grill 1', type: 'grill', activeOrders: 2, capacity: 4, status: 'busy', cameraId: 'cam-kitchen-1' },
  { id: 'grill-2', name: 'Grill 2', type: 'grill', activeOrders: 1, capacity: 4, status: 'active', cameraId: 'cam-kitchen-1' },
  { id: 'fryer-1', name: 'Fritadeira 1', type: 'fryer', activeOrders: 3, capacity: 6, status: 'busy', cameraId: 'cam-kitchen-2' },
  { id: 'assembly', name: 'Montagem', type: 'assembly', activeOrders: 2, capacity: 8, status: 'active', cameraId: 'cam-kitchen-2' },
  { id: 'drinks', name: 'Bebidas', type: 'drinks', activeOrders: 1, capacity: 10, status: 'active', cameraId: 'cam-counter' },
  { id: 'desserts', name: 'Sobremesas', type: 'desserts', activeOrders: 1, capacity: 5, status: 'active', cameraId: 'cam-counter' },
]

const initialStaff: Staff[] = [
  { id: 'staff-1', name: 'Carlos Silva', role: 'manager', status: 'active', ordersCompleted: 0 },
  { id: 'staff-2', name: 'Ana Santos', role: 'kitchen', station: 'grill-1', status: 'active', ordersCompleted: 23 },
  { id: 'staff-3', name: 'Pedro Oliveira', role: 'kitchen', station: 'grill-2', status: 'active', ordersCompleted: 18 },
  { id: 'staff-4', name: 'Maria Costa', role: 'kitchen', station: 'fryer-1', status: 'active', ordersCompleted: 31 },
  { id: 'staff-5', name: 'João Lima', role: 'crew', station: 'assembly', status: 'active', ordersCompleted: 45 },
  { id: 'staff-6', name: 'Lucia Ferreira', role: 'cashier', status: 'active', ordersCompleted: 67 },
  { id: 'staff-7', name: 'Rafael Souza', role: 'crew', station: 'drinks', status: 'break', ordersCompleted: 28 },
]

const initialCameras: Camera[] = [
  { id: 'cam-kitchen-1', name: 'Cozinha Principal', location: 'Grill Area', type: 'kitchen', status: 'recording', resolution: '1080p', fps: 30, connectedAgents: ['agent-kitchen-vision', 'agent-quality'] },
  { id: 'cam-kitchen-2', name: 'Cozinha Secundária', location: 'Fryer & Assembly', type: 'kitchen', status: 'analyzing', resolution: '1080p', fps: 30, connectedAgents: ['agent-kitchen-vision', 'agent-safety'] },
  { id: 'cam-counter', name: 'Balcão', location: 'Front Counter', type: 'counter', status: 'recording', resolution: '1080p', fps: 30, connectedAgents: ['agent-customer', 'agent-queue'] },
  { id: 'cam-drive-1', name: 'Drive-Thru Menu', location: 'Order Point', type: 'drive-thru', status: 'recording', resolution: '4K', fps: 60, connectedAgents: ['agent-drive-thru', 'agent-customer'] },
  { id: 'cam-drive-2', name: 'Drive-Thru Window', location: 'Pickup Window', type: 'drive-thru', status: 'online', resolution: '1080p', fps: 30, connectedAgents: ['agent-drive-thru'] },
  { id: 'cam-dining', name: 'Salão', location: 'Dining Area', type: 'dining', status: 'online', resolution: '1080p', fps: 30, connectedAgents: ['agent-customer', 'agent-cleanliness'] },
  { id: 'cam-entrance', name: 'Entrada', location: 'Main Entrance', type: 'entrance', status: 'recording', resolution: '1080p', fps: 30, connectedAgents: ['agent-customer'] },
]

const initialAgents: Agent[] = [
  {
    id: 'agent-kitchen-vision', name: 'kitchen-vision-agent', type: 'vision', status: 'processing',
    lastHeartbeat: new Date().toISOString(), version: 'v0.2.0',
    description: 'Monitors kitchen operations and food preparation',
    capabilities: ['food-tracking', 'prep-time-estimation', 'station-monitoring', 'hygiene-check'],
    assignedCameras: ['cam-kitchen-1', 'cam-kitchen-2'], processedToday: 3421,
    metrics: { cpu: 72, memory: 78, requests: 3421, inferenceTime: 42 }
  },
  {
    id: 'agent-quality', name: 'food-quality-agent', type: 'quality', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v1.0.0',
    description: 'AI food quality control and presentation analysis',
    capabilities: ['food-quality-check', 'presentation-score', 'temperature-estimation', 'portion-verification'],
    assignedCameras: ['cam-kitchen-1'], processedToday: 892,
    metrics: { cpu: 45, memory: 55, requests: 892, inferenceTime: 85 }
  },
  {
    id: 'agent-customer', name: 'customer-experience-agent', type: 'customer', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v0.3.0',
    description: 'Customer sentiment and experience analysis',
    capabilities: ['sentiment-analysis', 'wait-time-tracking', 'satisfaction-prediction', 'service-quality'],
    assignedCameras: ['cam-counter', 'cam-drive-1', 'cam-dining', 'cam-entrance'], processedToday: 1567,
    metrics: { cpu: 38, memory: 62, requests: 1567, inferenceTime: 65 }
  },
  {
    id: 'agent-queue', name: 'queue-management-agent', type: 'vision', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v0.1.0',
    description: 'Queue length detection and wait time prediction',
    capabilities: ['queue-counting', 'wait-time-prediction', 'peak-detection', 'staff-alert'],
    assignedCameras: ['cam-counter'], processedToday: 2341,
    metrics: { cpu: 28, memory: 45, requests: 2341, inferenceTime: 35 }
  },
  {
    id: 'agent-drive-thru', name: 'drive-thru-agent', type: 'vision', status: 'processing',
    lastHeartbeat: new Date().toISOString(), version: 'v0.2.0',
    description: 'Drive-thru vehicle detection and service optimization',
    capabilities: ['vehicle-detection', 'license-plate-recognition', 'service-time-tracking', 'order-association'],
    assignedCameras: ['cam-drive-1', 'cam-drive-2'], processedToday: 1823,
    metrics: { cpu: 65, memory: 72, requests: 1823, inferenceTime: 48 }
  },
  {
    id: 'agent-safety', name: 'safety-compliance-agent', type: 'quality', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v0.1.0',
    description: 'Kitchen safety and compliance monitoring',
    capabilities: ['ppe-detection', 'safety-violation', 'temperature-monitoring', 'hygiene-compliance'],
    assignedCameras: ['cam-kitchen-2'], processedToday: 567,
    metrics: { cpu: 32, memory: 48, requests: 567, inferenceTime: 55 }
  },
  {
    id: 'agent-cleanliness', name: 'cleanliness-agent', type: 'quality', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v0.1.0',
    description: 'Restaurant cleanliness and maintenance monitoring',
    capabilities: ['cleanliness-score', 'spill-detection', 'table-status', 'trash-monitoring'],
    assignedCameras: ['cam-dining'], processedToday: 234,
    metrics: { cpu: 22, memory: 38, requests: 234, inferenceTime: 45 }
  },
  {
    id: 'agent-inventory', name: 'inventory-agent', type: 'inventory', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v0.1.0',
    description: 'Real-time inventory tracking and prediction',
    capabilities: ['stock-monitoring', 'usage-prediction', 'reorder-alerts', 'waste-tracking'],
    processedToday: 156,
    metrics: { cpu: 18, memory: 35, requests: 156 }
  },
  {
    id: 'agent-order', name: 'order-orchestrator-agent', type: 'order', status: 'online',
    lastHeartbeat: new Date().toISOString(), version: 'v1.0.0',
    description: 'Order routing and kitchen coordination',
    capabilities: ['order-routing', 'load-balancing', 'priority-management', 'timing-optimization'],
    processedToday: 892,
    metrics: { cpu: 25, memory: 42, requests: 892 }
  },
]

const initialDetections: AIDetection[] = [
  { id: 'det-1', cameraId: 'cam-counter', type: 'queue', confidence: 0.95, timestamp: new Date(Date.now() - 30000).toISOString(), metadata: { queueLength: '4', estimatedWait: '3min' }, processed: true, agentId: 'agent-queue' },
  { id: 'det-2', cameraId: 'cam-kitchen-1', type: 'food-quality', confidence: 0.92, timestamp: new Date(Date.now() - 60000).toISOString(), metadata: { item: 'Big Mac', score: '95', presentation: 'excellent' }, processed: true, agentId: 'agent-quality' },
  { id: 'det-3', cameraId: 'cam-drive-1', type: 'drive-thru-vehicle', confidence: 0.98, timestamp: new Date(Date.now() - 90000).toISOString(), metadata: { vehicleType: 'sedan', plate: 'ABC-1234', waitTime: '45s' }, processed: true, agentId: 'agent-drive-thru' },
  { id: 'det-4', cameraId: 'cam-dining', type: 'customer', confidence: 0.88, timestamp: new Date(Date.now() - 120000).toISOString(), metadata: { sentiment: 'happy', tableOccupancy: '8/12' }, processed: true, agentId: 'agent-customer' },
  { id: 'det-5', cameraId: 'cam-kitchen-2', type: 'safety', confidence: 0.96, timestamp: new Date(Date.now() - 150000).toISOString(), metadata: { status: 'compliant', ppeScore: '100%' }, processed: true, agentId: 'agent-safety' },
  { id: 'det-6', cameraId: 'cam-dining', type: 'cleanliness', confidence: 0.91, timestamp: new Date(Date.now() - 180000).toISOString(), metadata: { score: '92', tablesClean: '10/12' }, processed: true, agentId: 'agent-cleanliness' },
]

export const useMcdonaldsStore = create<McdonaldsState>((set) => ({
  activeView: 'dashboard',
  sidebarOpen: true,
  orders: initialOrders,
  menuItems: initialMenuItems,
  stations: initialStations,
  staff: initialStaff,
  cameras: initialCameras,
  agents: initialAgents,
  detections: initialDetections,
  storeName: "McDonald's Paulista",
  storeId: 'MC-BR-SP-0247',
  
  setActiveView: (view) => set({ activeView: view }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  
  updateOrderStatus: (id, status) => set((state) => ({
    orders: state.orders.map(o => o.id === id ? { 
      ...o, 
      status,
      startedAt: status === 'preparing' && !o.startedAt ? new Date().toISOString() : o.startedAt,
      completedAt: status === 'ready' ? new Date().toISOString() : o.completedAt
    } : o)
  })),
  
  addOrder: (order) => set((state) => ({
    orders: [order, ...state.orders]
  })),

  addDetection: (detection) => set((state) => ({
    detections: [detection, ...state.detections]
  })),
}))
