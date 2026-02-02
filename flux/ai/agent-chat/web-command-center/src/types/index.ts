// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📱 AgentChat Command Center - Type Definitions
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ════════════════════════════════════════════════════════════════════════════
// User Types
// ════════════════════════════════════════════════════════════════════════════

export interface User {
  id: string
  username: string
  displayName: string
  email: string
  avatar?: string
  status: UserStatus
  createdAt: string
  lastActiveAt: string
  settings: UserSettings
  agentAssistantId?: string  // Their personal assistant agent
}

export type UserStatus = 'active' | 'inactive' | 'suspended' | 'pending'

export interface UserSettings {
  voiceCloneEnabled: boolean
  voiceCloneId?: string
  locationSharingEnabled: boolean
  proximityAlertRadius: number  // km
  contactsCanSeeLocation: boolean
  notificationsEnabled: boolean
}

// ════════════════════════════════════════════════════════════════════════════
// Agent Types
// ════════════════════════════════════════════════════════════════════════════

export interface Agent {
  id: string
  name: string
  displayName: string
  description: string
  role: AgentRole
  status: AgentStatus
  avatar: string
  color: string
  namespace: string
  metrics: AgentMetrics
  capabilities: string[]
  lastActivity?: string
  endpoint?: string
}

export type AgentRole = 
  | 'core'           // messaging-hub
  | 'capability'     // voice, media, location
  | 'assistant'      // per-user assistant
  | 'orchestrator'   // command-center

export type AgentStatus = 'online' | 'offline' | 'scaling' | 'error' | 'deploying'

export interface AgentMetrics {
  eventsProcessed: number
  successRate: number
  avgResponseTime: number
  uptime: number
  activeConnections: number
  queueDepth: number
  lastError?: string
  cpuUsage: number
  memoryUsage: number
}

// ════════════════════════════════════════════════════════════════════════════
// Chat Types
// ════════════════════════════════════════════════════════════════════════════

export interface Chat {
  id: string
  type: ChatType
  participants: ChatParticipant[]
  lastMessage?: Message
  createdAt: string
  updatedAt: string
  unreadCount: number
}

export type ChatType = 'direct' | 'assistant' | 'group'

export interface ChatParticipant {
  userId: string
  displayName: string
  avatar?: string
  role: 'owner' | 'member'
  joinedAt: string
}

export interface Message {
  id: string
  chatId: string
  senderId: string
  senderType: 'user' | 'agent'
  content: MessageContent
  status: MessageStatus
  timestamp: string
  metadata?: MessageMetadata
}

export type MessageStatus = 'sending' | 'sent' | 'delivered' | 'read' | 'failed'

export interface MessageContent {
  type: MessageContentType
  text?: string
  mediaUrl?: string
  mediaType?: string
  voiceUrl?: string
  voiceDuration?: number
  location?: LocationData
  generatedBy?: string  // agent that generated content
}

export type MessageContentType = 'text' | 'image' | 'video' | 'voice' | 'location' | 'system'

export interface MessageMetadata {
  deviceId?: string
  appVersion?: string
  eventType?: string
  traceId?: string
}

// ════════════════════════════════════════════════════════════════════════════
// Location Types
// ════════════════════════════════════════════════════════════════════════════

export interface LocationData {
  latitude: number
  longitude: number
  accuracy?: number
  altitude?: number
  city?: string
  country?: string
  timestamp: string
}

export interface ProximityAlert {
  id: string
  userId: string
  contactId: string
  distance: number
  city: string
  timestamp: string
  acknowledged: boolean
}

// ════════════════════════════════════════════════════════════════════════════
// Voice Types
// ════════════════════════════════════════════════════════════════════════════

export interface VoiceClone {
  id: string
  userId: string
  status: VoiceCloneStatus
  sampleCount: number
  totalDuration: number
  createdAt: string
  lastUsedAt?: string
}

export type VoiceCloneStatus = 'pending' | 'processing' | 'ready' | 'failed'

// ════════════════════════════════════════════════════════════════════════════
// Alert Types
// ════════════════════════════════════════════════════════════════════════════

export interface Alert {
  id: string
  type: AlertType
  severity: AlertSeverity
  title: string
  message: string
  source: string
  timestamp: string
  acknowledged: boolean
  acknowledgedBy?: string
  acknowledgedAt?: string
  metadata?: Record<string, unknown>
}

export type AlertType = 
  | 'system'
  | 'security'
  | 'performance'
  | 'user'
  | 'agent'

export type AlertSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info'

// ════════════════════════════════════════════════════════════════════════════
// Analytics Types
// ════════════════════════════════════════════════════════════════════════════

export interface DashboardMetrics {
  totalUsers: number
  activeUsers: number
  totalMessages: number
  messagesLast24h: number
  totalAgents: number
  agentsOnline: number
  voiceClonesActive: number
  imagesGenerated: number
  locationAlerts: number
  systemHealth: SystemHealth
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'critical'
  uptime: number
  lastIncident?: string
  services: ServiceHealth[]
}

export interface ServiceHealth {
  name: string
  status: 'up' | 'down' | 'degraded'
  latency: number
  errorRate: number
}

export interface TimeSeriesData {
  timestamp: string
  value: number
}

export interface AnalyticsData {
  messagesOverTime: TimeSeriesData[]
  usersOverTime: TimeSeriesData[]
  agentResponseTimes: TimeSeriesData[]
  featureUsage: {
    voice: number
    images: number
    videos: number
    location: number
  }
}

// ════════════════════════════════════════════════════════════════════════════
// CloudEvent Types
// ════════════════════════════════════════════════════════════════════════════

export interface AgentChatCloudEvent {
  specversion: '1.0'
  id: string
  source: string
  type: string
  subject?: string
  time: string
  datacontenttype: string
  data: unknown
}

// ════════════════════════════════════════════════════════════════════════════
// API Types
// ════════════════════════════════════════════════════════════════════════════

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, unknown>
}
