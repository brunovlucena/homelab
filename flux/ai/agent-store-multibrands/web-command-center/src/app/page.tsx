'use client'

import { useStoreAppStore } from '@/store/storeAppStore'
import { 
  Header, 
  Sidebar, 
  Dashboard, 
  ConversationsView,
  OrdersView,
  SellersView,
  AnalyticsView,
  CommandPalette 
} from '@/components'
import { cn } from '@/lib/utils'
import { useEffect } from 'react'
import { BrandId, BRANDS, Conversation, Order, StoreEvent } from '@/types/store'
import { generateId } from '@/lib/utils'

// Generate mock conversations
function generateMockConversations(): Conversation[] {
  const conversations: Conversation[] = []
  const brands = Object.keys(BRANDS) as BrandId[]
  const states = ['active', 'waiting', 'escalated'] as const
  
  for (let i = 0; i < 15; i++) {
    const brand = brands[Math.floor(Math.random() * brands.length)]
    const state = states[Math.floor(Math.random() * states.length)]
    
    conversations.push({
      id: `conv-${i}`,
      customerId: `cust-${i}`,
      customerPhone: `5511${Math.floor(Math.random() * 900000000 + 100000000)}`,
      customerName: ['Maria', 'João', 'Ana', 'Pedro', 'Carla'][Math.floor(Math.random() * 5)],
      brand,
      state,
      messages: [
        {
          id: `msg-${i}-1`,
          role: 'customer',
          content: 'Olá! Gostaria de saber mais sobre os produtos.',
          timestamp: new Date(Date.now() - Math.random() * 3600000).toISOString(),
        },
        {
          id: `msg-${i}-2`,
          role: 'ai',
          content: `Olá! Sou ${BRANDS[brand].sellerName} e estou aqui para ajudar! O que você procura?`,
          timestamp: new Date(Date.now() - Math.random() * 1800000).toISOString(),
          metadata: { tokensUsed: 85, responseTime: 1.2 },
        },
      ],
      startedAt: new Date(Date.now() - Math.random() * 7200000).toISOString(),
      lastMessageAt: new Date(Date.now() - Math.random() * 600000).toISOString(),
      escalationReason: state === 'escalated' ? 'Cliente solicitou atendente humano' : undefined,
    })
  }
  
  return conversations
}

// Generate mock orders
function generateMockOrders(): Order[] {
  const orders: Order[] = []
  const brands = Object.keys(BRANDS) as BrandId[]
  const statuses = ['pending', 'confirmed', 'processing', 'shipped', 'delivered'] as const
  
  for (let i = 0; i < 20; i++) {
    const brand = brands[Math.floor(Math.random() * brands.length)]
    const status = statuses[Math.floor(Math.random() * statuses.length)]
    const itemCount = Math.floor(Math.random() * 3) + 1
    const items = Array.from({ length: itemCount }, (_, j) => ({
      productId: `${brand}-00${j + 1}`,
      productName: `Produto ${brand} ${j + 1}`,
      quantity: Math.floor(Math.random() * 3) + 1,
      unitPrice: Math.random() * 500 + 50,
      brand,
    }))
    
    orders.push({
      id: `ORD-${String(i + 1).padStart(5, '0')}`,
      customerId: `cust-${i}`,
      customerPhone: `5511${Math.floor(Math.random() * 900000000 + 100000000)}`,
      items,
      total: items.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0),
      status,
      createdAt: new Date(Date.now() - Math.random() * 86400000 * 7).toISOString(),
      updatedAt: new Date(Date.now() - Math.random() * 86400000).toISOString(),
      brand,
    })
  }
  
  return orders.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
}

export default function Home() {
  const { 
    activeView, 
    sidebarOpen,
    addConversation,
    addOrder,
    addEvent,
    conversations,
    orders,
  } = useStoreAppStore()

  // Initialize mock data
  useEffect(() => {
    if (conversations.length === 0) {
      generateMockConversations().forEach(addConversation)
    }
    if (orders.length === 0) {
      generateMockOrders().forEach(addOrder)
    }
    
    // Simulate real-time events
    const interval = setInterval(() => {
      const eventTypes = [
        'store.chat.message.new',
        'store.order.created',
        'store.sales.escalate',
        'store.product.viewed',
      ]
      const brands = Object.keys(BRANDS) as BrandId[]
      
      const event: StoreEvent = {
        id: generateId(),
        type: eventTypes[Math.floor(Math.random() * eventTypes.length)],
        brand: brands[Math.floor(Math.random() * brands.length)],
        timestamp: new Date().toISOString(),
        data: { mock: true },
      }
      
      addEvent(event)
    }, 5000)
    
    return () => clearInterval(interval)
  }, [addConversation, addOrder, addEvent, conversations.length, orders.length])

  const renderView = () => {
    switch (activeView) {
      case 'dashboard':
        return <Dashboard />
      case 'conversations':
        return <ConversationsView />
      case 'orders':
        return <OrdersView />
      case 'sellers':
        return <SellersView />
      case 'analytics':
        return <AnalyticsView />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-store-dark">
      {/* Header */}
      <Header />

      {/* Sidebar */}
      <Sidebar />

      {/* Main Content */}
      <main className={cn(
        "pt-16 min-h-screen transition-all duration-300",
        sidebarOpen ? "pl-64" : "pl-0"
      )}>
        <div className="h-[calc(100vh-64px)]">
          {renderView()}
        </div>
      </main>

      {/* Command Palette */}
      <CommandPalette />
    </div>
  )
}
