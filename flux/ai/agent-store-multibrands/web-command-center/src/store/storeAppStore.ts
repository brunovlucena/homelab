import { create } from 'zustand'
import { 
  BrandId, 
  Seller, 
  Conversation, 
  Order, 
  StoreEvent, 
  DashboardMetrics,
  BRANDS 
} from '@/types/store'

// Initial sellers
const INITIAL_SELLERS: Seller[] = Object.values(BRANDS).map((brand) => ({
  id: `seller-${brand.id}`,
  brand: brand.id,
  status: 'online' as const,
  activeConversations: Math.floor(Math.random() * 5),
  messagesHandled: Math.floor(Math.random() * 100) + 50,
  avgResponseTime: Math.random() * 2 + 1,
  satisfaction: Math.random() * 0.2 + 0.8,
  lastActive: new Date().toISOString(),
}))

// Mock metrics
const INITIAL_METRICS: DashboardMetrics = {
  totalMessages: 1247,
  totalOrders: 89,
  totalRevenue: 45890.50,
  activeConversations: 23,
  avgResponseTime: 1.8,
  escalations: 5,
  brandMetrics: Object.values(BRANDS).map((brand) => ({
    brand: brand.id,
    messages24h: Math.floor(Math.random() * 300) + 100,
    orders24h: Math.floor(Math.random() * 30) + 10,
    revenue24h: Math.random() * 15000 + 5000,
    conversionRate: Math.random() * 0.2 + 0.05,
    avgResponseTime: Math.random() * 2 + 1,
    escalationRate: Math.random() * 0.1,
    satisfaction: Math.random() * 0.2 + 0.8,
  })),
}

interface StoreAppState {
  // UI State
  sidebarOpen: boolean
  activeView: 'dashboard' | 'conversations' | 'orders' | 'products' | 'sellers' | 'analytics'
  selectedBrand: BrandId | 'all'
  commandPaletteOpen: boolean
  
  // Data
  sellers: Seller[]
  conversations: Conversation[]
  orders: Order[]
  events: StoreEvent[]
  metrics: DashboardMetrics
  
  // Selected items
  selectedConversationId: string | null
  selectedOrderId: string | null
  
  // Actions
  toggleSidebar: () => void
  setActiveView: (view: StoreAppState['activeView']) => void
  setSelectedBrand: (brand: BrandId | 'all') => void
  toggleCommandPalette: () => void
  
  selectConversation: (id: string | null) => void
  selectOrder: (id: string | null) => void
  
  addEvent: (event: StoreEvent) => void
  addConversation: (conversation: Conversation) => void
  updateConversation: (id: string, updates: Partial<Conversation>) => void
  addMessageToConversation: (conversationId: string, message: Conversation['messages'][0]) => void
  
  addOrder: (order: Order) => void
  updateOrderStatus: (orderId: string, status: Order['status']) => void
  
  updateSellerStatus: (sellerId: string, status: Seller['status']) => void
  updateMetrics: (metrics: Partial<DashboardMetrics>) => void
}

export const useStoreAppStore = create<StoreAppState>((set, get) => ({
  // Initial UI state
  sidebarOpen: true,
  activeView: 'dashboard',
  selectedBrand: 'all',
  commandPaletteOpen: false,
  
  // Initial data
  sellers: INITIAL_SELLERS,
  conversations: [],
  orders: [],
  events: [],
  metrics: INITIAL_METRICS,
  
  // Selected items
  selectedConversationId: null,
  selectedOrderId: null,
  
  // Actions
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  
  setActiveView: (view) => set({ activeView: view }),
  
  setSelectedBrand: (brand) => set({ selectedBrand: brand }),
  
  toggleCommandPalette: () => set((state) => ({ 
    commandPaletteOpen: !state.commandPaletteOpen 
  })),
  
  selectConversation: (id) => set({ selectedConversationId: id }),
  
  selectOrder: (id) => set({ selectedOrderId: id }),
  
  addEvent: (event) => set((state) => ({
    events: [event, ...state.events.slice(0, 99)],
  })),
  
  addConversation: (conversation) => set((state) => ({
    conversations: [conversation, ...state.conversations],
  })),
  
  updateConversation: (id, updates) => set((state) => ({
    conversations: state.conversations.map((conv) =>
      conv.id === id ? { ...conv, ...updates } : conv
    ),
  })),
  
  addMessageToConversation: (conversationId, message) => set((state) => ({
    conversations: state.conversations.map((conv) =>
      conv.id === conversationId
        ? { 
            ...conv, 
            messages: [...conv.messages, message],
            lastMessageAt: message.timestamp,
          }
        : conv
    ),
  })),
  
  addOrder: (order) => set((state) => ({
    orders: [order, ...state.orders],
  })),
  
  updateOrderStatus: (orderId, status) => set((state) => ({
    orders: state.orders.map((order) =>
      order.id === orderId 
        ? { ...order, status, updatedAt: new Date().toISOString() }
        : order
    ),
  })),
  
  updateSellerStatus: (sellerId, status) => set((state) => ({
    sellers: state.sellers.map((seller) =>
      seller.id === sellerId ? { ...seller, status } : seller
    ),
  })),
  
  updateMetrics: (newMetrics) => set((state) => ({
    metrics: { ...state.metrics, ...newMetrics },
  })),
}))

// Selectors
export const selectSellerByBrand = (state: StoreAppState, brand: BrandId) =>
  state.sellers.find((s) => s.brand === brand)

export const selectConversationsByBrand = (state: StoreAppState, brand: BrandId | 'all') =>
  brand === 'all' 
    ? state.conversations 
    : state.conversations.filter((c) => c.brand === brand)

export const selectOrdersByBrand = (state: StoreAppState, brand: BrandId | 'all') =>
  brand === 'all'
    ? state.orders
    : state.orders.filter((o) => o.brand === brand)

export const selectActiveConversations = (state: StoreAppState) =>
  state.conversations.filter((c) => c.state === 'active' || c.state === 'waiting')

export const selectEscalatedConversations = (state: StoreAppState) =>
  state.conversations.filter((c) => c.state === 'escalated')
