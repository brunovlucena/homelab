// Brand types
export type BrandId = 'fashion' | 'tech' | 'gaming' | 'beauty' | 'home'

export interface Brand {
  id: BrandId
  name: string
  emoji: string
  color: string
  sellerName: string
  sellerAvatar: string
  description: string
}

export const BRANDS: Record<BrandId, Brand> = {
  fashion: {
    id: 'fashion',
    name: 'Fashion',
    emoji: '👗',
    color: 'brand-fashion',
    sellerName: 'Luna',
    sellerAvatar: '👩‍🎨',
    description: 'Moda e estilo para todas as ocasiões',
  },
  tech: {
    id: 'tech',
    name: 'Tech',
    emoji: '💻',
    color: 'brand-tech',
    sellerName: 'Max',
    sellerAvatar: '🤖',
    description: 'Tecnologia e inovação',
  },
  gaming: {
    id: 'gaming',
    name: 'Gaming',
    emoji: '🎮',
    color: 'brand-gaming',
    sellerName: 'Pixel',
    sellerAvatar: '👾',
    description: 'Gaming e entretenimento',
  },
  beauty: {
    id: 'beauty',
    name: 'Beauty',
    emoji: '💄',
    color: 'brand-beauty',
    sellerName: 'Bella',
    sellerAvatar: '💅',
    description: 'Beleza e cuidados pessoais',
  },
  home: {
    id: 'home',
    name: 'Home',
    emoji: '🏠',
    color: 'brand-home',
    sellerName: 'Casa',
    sellerAvatar: '🏡',
    description: 'Casa e decoração',
  },
}

// Seller status
export type SellerStatus = 'online' | 'busy' | 'offline'

export interface Seller {
  id: string
  brand: BrandId
  status: SellerStatus
  activeConversations: number
  messagesHandled: number
  avgResponseTime: number
  satisfaction: number
  lastActive: string
}

// Conversation types
export type ConversationState = 'active' | 'waiting' | 'escalated' | 'closed'

export interface Message {
  id: string
  role: 'customer' | 'ai' | 'human'
  content: string
  timestamp: string
  metadata?: {
    tokensUsed?: number
    responseTime?: number
    sentiment?: number
  }
}

export interface Conversation {
  id: string
  customerId: string
  customerPhone: string
  customerName?: string
  brand: BrandId
  state: ConversationState
  messages: Message[]
  startedAt: string
  lastMessageAt: string
  assignedTo?: string
  escalationReason?: string
}

// Product types
export interface Product {
  id: string
  name: string
  brand: BrandId
  description: string
  price: number
  category: string
  tags: string[]
  stock: number
  images?: string[]
}

// Order types
export type OrderStatus = 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'cancelled'

export interface OrderItem {
  productId: string
  productName: string
  quantity: number
  unitPrice: number
  brand: BrandId
}

export interface Order {
  id: string
  customerId: string
  customerPhone: string
  items: OrderItem[]
  total: number
  status: OrderStatus
  createdAt: string
  updatedAt: string
  sellerId?: string
  brand: BrandId
}

// Metrics types
export interface BrandMetrics {
  brand: BrandId
  messages24h: number
  orders24h: number
  revenue24h: number
  conversionRate: number
  avgResponseTime: number
  escalationRate: number
  satisfaction: number
}

export interface DashboardMetrics {
  totalMessages: number
  totalOrders: number
  totalRevenue: number
  activeConversations: number
  avgResponseTime: number
  escalations: number
  brandMetrics: BrandMetrics[]
}

// Event types
export interface StoreEvent {
  id: string
  type: string
  brand?: BrandId
  timestamp: string
  data: Record<string, unknown>
}
