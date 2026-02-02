import { create } from 'zustand'

export type ViewType = 'dashboard' | 'floor' | 'kitchen' | 'agents' | 'menu' | 'events'
export type TableStatus = 'available' | 'occupied' | 'reserved' | 'cleaning'
export type OrderStatus = 'pending' | 'preparing' | 'ready' | 'served'

export interface Table {
  id: string
  capacity: number
  status: TableStatus
  guestName?: string
  partySize?: number
  server?: string
  seatedAt?: string
  currentOrder?: string
}

export interface KitchenTicket {
  id: string
  tableId: string
  items: {
    dish: string
    quantity: number
    specialRequests?: string
    status: OrderStatus
    station: string
  }[]
  priority: 'normal' | 'rush' | 'vip'
  createdAt: string
  estimatedReady?: string
}

export interface Agent {
  id: string
  name: string
  role: 'host' | 'waiter' | 'chef' | 'sommelier'
  avatar: string
  status: 'active' | 'busy' | 'offline'
  currentTask?: string
  lastActivity?: string
  metrics?: {
    cpu: number
    memory: number
    requests: number
    latency: number
  }
}

export interface RestaurantEvent {
  id: string
  type: string
  source: string
  data: Record<string, unknown>
  timestamp: string
}

interface RestaurantState {
  activeView: ViewType
  setActiveView: (view: ViewType) => void
  
  // Data loading state
  isLoading: boolean
  dataSource: 'live' | 'mock' | 'error'
  lastFetched: string | null
  errorMessage: string | null
  
  // Tables
  tables: Table[]
  updateTable: (id: string, updates: Partial<Table>) => void
  setTables: (tables: Table[]) => void
  
  // Kitchen
  tickets: KitchenTicket[]
  addTicket: (ticket: KitchenTicket) => void
  updateTicket: (id: string, updates: Partial<KitchenTicket>) => void
  setTickets: (tickets: KitchenTicket[]) => void
  
  // Agents
  agents: Agent[]
  updateAgent: (id: string, updates: Partial<Agent>) => void
  setAgents: (agents: Agent[]) => void
  
  // Events
  events: RestaurantEvent[]
  addEvent: (event: RestaurantEvent) => void
  
  // Stats
  stats: {
    guestsTonight: number
    revenue: number
    avgWaitTime: number
    satisfaction: number
  }
  setStats: (stats: RestaurantState['stats']) => void
  
  // Actions
  fetchLiveData: () => Promise<void>
  setDataSource: (source: 'live' | 'mock' | 'error', message?: string) => void
}

// ⚠️ WARNING: This is MOCK DATA - used as fallback when live services are unavailable
// The data below is NOT real and should be replaced with actual API calls
const MOCK_TABLES: Table[] = [
  { id: 'table-01', capacity: 2, status: 'occupied', guestName: 'Mr. Santos', partySize: 2, server: 'Pierre' },
  { id: 'table-02', capacity: 2, status: 'available' },
  { id: 'table-03', capacity: 4, status: 'reserved', guestName: 'Johnson Party' },
  { id: 'table-04', capacity: 4, status: 'occupied', guestName: 'Ms. Chen', partySize: 3, server: 'Pierre' },
  { id: 'table-05', capacity: 6, status: 'available' },
  { id: 'table-06', capacity: 8, status: 'occupied', guestName: 'Anniversary Dinner', partySize: 6, server: 'Pierre' },
]

// ⚠️ MOCK AGENTS - Replace with real agent status from Kubernetes/Prometheus
const MOCK_AGENTS: Agent[] = [
  { id: 'host-maximilian', name: 'Maximilian', role: 'host', avatar: '🎩', status: 'offline', currentTask: '⚠️ MOCK DATA - Agent not connected' },
  { id: 'waiter-pierre', name: 'Pierre', role: 'waiter', avatar: '👔', status: 'offline', currentTask: '⚠️ MOCK DATA - Agent not connected' },
  { id: 'chef-marco', name: 'Marco', role: 'chef', avatar: '👨‍🍳', status: 'offline', currentTask: '⚠️ MOCK DATA - Agent not connected' },
  { id: 'sommelier-isabella', name: 'Isabella', role: 'sommelier', avatar: '🍷', status: 'offline', currentTask: '⚠️ MOCK DATA - Agent not connected' },
]

// ⚠️ MOCK TICKETS - Replace with real kitchen orders
const MOCK_TICKETS: KitchenTicket[] = [
  {
    id: 'ticket-001',
    tableId: 'table-01',
    items: [
      { dish: 'Risotto ai Porcini', quantity: 1, status: 'ready', station: 'saute' },
      { dish: 'Branzino alla Griglia', quantity: 1, status: 'preparing', station: 'grill' },
    ],
    priority: 'normal',
    createdAt: new Date(Date.now() - 15 * 60000).toISOString(),
  },
  {
    id: 'ticket-002',
    tableId: 'table-04',
    items: [
      { dish: 'Bruschetta Trio', quantity: 1, status: 'served', station: 'garde_manger' },
      { dish: 'Ossobuco alla Milanese', quantity: 2, status: 'preparing', station: 'braise' },
    ],
    priority: 'normal',
    createdAt: new Date(Date.now() - 25 * 60000).toISOString(),
  },
  {
    id: 'ticket-003',
    tableId: 'table-06',
    items: [
      { dish: 'Carpaccio di Manzo', quantity: 3, status: 'ready', station: 'garde_manger' },
      { dish: 'Tagliatelle al Ragù', quantity: 2, status: 'pending', station: 'pasta' },
      { dish: 'Risotto ai Porcini', quantity: 2, status: 'pending', station: 'saute' },
    ],
    priority: 'vip',
    createdAt: new Date(Date.now() - 10 * 60000).toISOString(),
  },
]

export const useRestaurantStore = create<RestaurantState>((set, get) => ({
  activeView: 'dashboard',
  setActiveView: (view) => set({ activeView: view }),
  
  // Data state - defaults to mock with warning
  isLoading: false,
  dataSource: 'mock',
  lastFetched: null,
  errorMessage: '⚠️ Using MOCK data. Configure PROMETHEUS_URL and KUBERNETES_API_URL to enable live data.',
  
  tables: MOCK_TABLES,
  setTables: (tables) => set({ tables }),
  updateTable: (id, updates) =>
    set((state) => ({
      tables: state.tables.map((t) =>
        t.id === id ? { ...t, ...updates } : t
      ),
    })),
  
  tickets: MOCK_TICKETS,
  setTickets: (tickets) => set({ tickets }),
  addTicket: (ticket) =>
    set((state) => ({ tickets: [ticket, ...state.tickets] })),
  updateTicket: (id, updates) =>
    set((state) => ({
      tickets: state.tickets.map((t) =>
        t.id === id ? { ...t, ...updates } : t
      ),
    })),
  
  agents: MOCK_AGENTS,
  setAgents: (agents) => set({ agents }),
  updateAgent: (id, updates) =>
    set((state) => ({
      agents: state.agents.map((a) =>
        a.id === id ? { ...a, ...updates } : a
      ),
    })),
  
  events: [],
  addEvent: (event) =>
    set((state) => ({ events: [event, ...state.events].slice(0, 100) })),
  
  // ⚠️ MOCK STATS - Replace with real metrics from Prometheus
  stats: {
    guestsTonight: 0,  // Changed from fake 47 to 0
    revenue: 0,        // Changed from fake 4250 to 0
    avgWaitTime: 0,    // Changed from fake 12 to 0
    satisfaction: 0,   // Changed from fake 4.8 to 0
  },
  setStats: (stats) => set({ stats }),
  
  setDataSource: (source, message) => set({ 
    dataSource: source, 
    errorMessage: message || null,
    lastFetched: new Date().toISOString(),
  }),
  
  // Fetch live data from APIs
  fetchLiveData: async () => {
    set({ isLoading: true })
    
    try {
      // Fetch agents from API
      const agentsResponse = await fetch('/api/agents')
      const agentsData = await agentsResponse.json()
      
      // Fetch metrics from API
      const metricsResponse = await fetch('/api/metrics')
      const metricsData = await metricsResponse.json()
      
      if (agentsData.success && agentsData.agents.length > 0) {
        // Map Kubernetes agents to our format
        const liveAgents: Agent[] = agentsData.agents.map((a: any, index: number) => ({
          id: a.name,
          name: a.name.split('-').map((w: string) => w.charAt(0).toUpperCase() + w.slice(1)).join(' '),
          role: a.role || ['host', 'waiter', 'chef', 'sommelier'][index % 4],
          avatar: ['🎩', '👔', '👨‍🍳', '🍷'][index % 4],
          status: a.status === 'online' ? 'active' : 'offline',
          currentTask: a.status === 'online' ? 'Connected to live backend' : 'Service unavailable',
          lastActivity: a.lastHeartbeat,
          metrics: a.metrics,
        }))
        
        set({
          agents: liveAgents,
          dataSource: 'live',
          errorMessage: null,
          lastFetched: new Date().toISOString(),
        })
      } else {
        // Fall back to mock data with warning
        set({
          agents: MOCK_AGENTS,
          dataSource: 'mock',
          errorMessage: agentsData.message || 'Could not fetch live agent data',
        })
      }
      
      // Use real metrics if available
      if (metricsData.success) {
        // TODO: Parse and set real stats from Prometheus
        console.log('Live metrics available:', metricsData)
      }
      
    } catch (error) {
      console.error('Failed to fetch live data:', error)
      set({
        dataSource: 'error',
        errorMessage: `Failed to connect to backend: ${error}`,
      })
    } finally {
      set({ isLoading: false })
    }
  },
}))
