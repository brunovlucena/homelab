import axios, { AxiosInstance, AxiosResponse } from 'axios'

// =============================================================================
// 📋 TYPES
// =============================================================================

export interface ChatMessage {
  message: string
  timestamp?: string
}

export interface ChatResponse {
  response: string
  timestamp: string
  model?: string
  sources?: string[]
  mode?: 'direct' | 'mcp'
}

export interface AgentStatus {
  status: string
  service: string
  timestamp: string
  mcp_server?: {
    status: string
    url: string
  }
}

export interface LogAnalysisRequest {
  logs: string
  context?: string
}

export interface LogAnalysisResponse {
  analysis: string
  severity?: string
  recommendations?: string[]
  timestamp: string
}

export interface LLMStatus {
  status: 'healthy' | 'error'
  error?: string
  agent_status?: AgentStatus
}

// =============================================================================
// 🤖 CHATBOT SERVICE
// =============================================================================

class ChatbotService {
  private client: AxiosInstance
  private agentBaseUrl: string
  private initialized: boolean = false

  constructor() {
    // Determine agent URL based on environment
    // In production, use internal cluster service
    // In development, use NodePort or localhost proxy
    const isProduction = process.env.NODE_ENV === 'production'
    const apiUrl = process.env.VITE_API_URL || '/api/v1'
    
    // Use API proxy for agent-sre in production, direct NodePort in dev
    this.agentBaseUrl = isProduction 
      ? `${apiUrl}/agent-sre`  // Proxy through homepage API
      : 'http://localhost:31081' // Direct NodePort access

    this.client = axios.create({
      baseURL: this.agentBaseUrl,
      timeout: 30000, // 30 seconds for AI responses
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        console.log(`🤖 [ChatbotService] Request to ${config.url}`)
        return config
      },
      (error) => {
        console.error('🤖 [ChatbotService] Request error:', error)
        return Promise.reject(error)
      }
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        console.log('🤖 [ChatbotService] Response received:', response.status)
        return response
      },
      (error) => {
        console.error('🤖 [ChatbotService] Response error:', error)
        if (error.response?.status === 503) {
          console.warn('🤖 [ChatbotService] Agent-SRE service unavailable')
        }
        return Promise.reject(error)
      }
    )
  }

  // =============================================================================
  // 🚀 INITIALIZATION
  // =============================================================================

  initialize(): void {
    if (this.initialized) {
      console.log('🤖 [ChatbotService] Already initialized')
      return
    }

    console.log('🤖 [ChatbotService] Initializing...')
    console.log(`🤖 [ChatbotService] Agent URL: ${this.agentBaseUrl}`)
    this.initialized = true
  }

  // =============================================================================
  // 💬 CHAT METHODS
  // =============================================================================

  /**
   * Send a chat message using direct agent communication
   */
  async chat(message: string): Promise<ChatResponse> {
    const request: ChatMessage = {
      message,
      timestamp: new Date().toISOString(),
    }

    const response: AxiosResponse<ChatResponse> = await this.client.post('/chat', request)
    return {
      ...response.data,
      mode: 'direct',
    }
  }

  /**
   * Send a chat message using MCP protocol
   */
  async mcpChat(message: string): Promise<ChatResponse> {
    const request: ChatMessage = {
      message,
      timestamp: new Date().toISOString(),
    }

    const response: AxiosResponse<ChatResponse> = await this.client.post('/mcp/chat', request)
    return {
      ...response.data,
      mode: 'mcp',
    }
  }

  /**
   * Process a chat message with fallback strategy
   * Tries MCP first, falls back to direct if MCP fails
   */
  async processMessage(message: string): Promise<{ text: string; sources?: string[] }> {
    try {
      // Try MCP chat first (preferred for complex queries)
      console.log('🤖 [ChatbotService] Attempting MCP chat...')
      const mcpResponse = await this.mcpChat(message)
      console.log('🤖 [ChatbotService] MCP chat successful')
      
      return {
        text: mcpResponse.response,
        sources: mcpResponse.sources || ['Agent-SRE (MCP)'],
      }
    } catch (mcpError) {
      console.warn('🤖 [ChatbotService] MCP chat failed, falling back to direct chat:', mcpError)
      
      try {
        // Fallback to direct chat
        console.log('🤖 [ChatbotService] Attempting direct chat...')
        const directResponse = await this.chat(message)
        console.log('🤖 [ChatbotService] Direct chat successful')
        
        return {
          text: directResponse.response,
          sources: directResponse.sources || ['Agent-SRE (Direct)'],
        }
      } catch (directError) {
        console.error('🤖 [ChatbotService] All chat methods failed:', directError)
        
        // Return a helpful error message
        return {
          text: 'Sorry, I\'m currently unavailable. The SRE agent service might be down. Please try again later.',
          sources: ['Error Handler'],
        }
      }
    }
  }

  // =============================================================================
  // 📊 LOG ANALYSIS
  // =============================================================================

  /**
   * Analyze logs using direct agent communication
   */
  async analyzeLogsDirect(logs: string, context?: string): Promise<LogAnalysisResponse> {
    const request: LogAnalysisRequest = {
      logs,
      context,
    }

    const response: AxiosResponse<LogAnalysisResponse> = await this.client.post('/analyze-logs', request)
    return response.data
  }

  /**
   * Analyze logs using MCP protocol
   */
  async analyzeLogsMCP(logs: string, context?: string): Promise<LogAnalysisResponse> {
    const request: LogAnalysisRequest = {
      logs,
      context,
    }

    const response: AxiosResponse<LogAnalysisResponse> = await this.client.post('/mcp/analyze-logs', request)
    return response.data
  }

  // =============================================================================
  // 🏥 HEALTH & STATUS
  // =============================================================================

  /**
   * Check agent health
   */
  async healthCheck(): Promise<{ status: string }> {
    const response: AxiosResponse<{ status: string }> = await this.client.get('/health')
    return response.data
  }

  /**
   * Check agent readiness
   */
  async readyCheck(): Promise<{ status: string }> {
    const response: AxiosResponse<{ status: string }> = await this.client.get('/ready')
    return response.data
  }

  /**
   * Get detailed agent status
   */
  async getStatus(): Promise<AgentStatus> {
    const response: AxiosResponse<AgentStatus> = await this.client.get('/status')
    return response.data
  }

  /**
   * Get LLM status for UI display
   */
  async getLLMStatus(): Promise<LLMStatus> {
    try {
      const status = await this.getStatus()
      return {
        status: 'healthy',
        agent_status: status,
      }
    } catch (error) {
      console.error('🤖 [ChatbotService] Failed to get LLM status:', error)
      return {
        status: 'error',
        error: 'Agent-SRE service is unavailable',
      }
    }
  }

  // =============================================================================
  // 🎯 CONVENIENCE METHODS
  // =============================================================================

  /**
   * Check if agent is available
   */
  async isAvailable(): Promise<boolean> {
    try {
      await this.healthCheck()
      return true
    } catch (error) {
      return false
    }
  }

  /**
   * Get agent info for debugging
   */
  getAgentInfo(): { baseUrl: string; initialized: boolean } {
    return {
      baseUrl: this.agentBaseUrl,
      initialized: this.initialized,
    }
  }
}

// =============================================================================
// 📤 EXPORTS
// =============================================================================

const chatbotService = new ChatbotService()
export default chatbotService
export { ChatbotService }

