import axios, { AxiosInstance, AxiosResponse } from 'axios'
import { env } from '../utils/env'

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
}

export interface AgentStatus {
  status: string
  service: string
  timestamp: string
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
    // 🤖 Use the API path for Agent Bruno (Homepage chatbot and knowledge assistant)
    const apiUrl = env.API_URL
    this.agentBaseUrl = `${apiUrl}/agent-bruno`

    this.client = axios.create({
      baseURL: this.agentBaseUrl,
      timeout: 60000, // 60 seconds for AI responses
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        console.log(`🤖 [ChatbotService] Request to Agent Bruno: ${config.url}`)
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
        console.log('🤖 [ChatbotService] Response from Agent Bruno:', response.status)
        return response
      },
      (error) => {
        console.error('🤖 [ChatbotService] Response error:', error)
        if (error.response?.status === 503) {
          console.warn('🤖 [ChatbotService] Agent Bruno service unavailable')
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

    console.log('🤖 [ChatbotService] Initializing Agent Bruno connection...')
    console.log(`🤖 [ChatbotService] Agent Bruno URL: ${this.agentBaseUrl}`)
    this.initialized = true
  }

  // =============================================================================
  // 💬 CHAT METHODS
  // =============================================================================

  /**
   * Send a chat message to the agent
   */
  async chat(message: string): Promise<ChatResponse> {
    const request: ChatMessage = {
      message,
      timestamp: new Date().toISOString(),
    }

    const response: AxiosResponse<ChatResponse> = await this.client.post('/chat', request)
    return response.data
  }

  /**
   * Process a chat message
   */
  async processMessage(message: string): Promise<{ text: string; sources?: string[] }> {
    try {
      console.log('🤖 [ChatbotService] Sending message to Agent Bruno...')
      const response = await this.chat(message)
      console.log('🤖 [ChatbotService] Message sent successfully')
      
      return {
        text: response.response,
        sources: response.sources || ['Agent Bruno'],
      }
    } catch (error) {
      console.error('🤖 [ChatbotService] Chat failed:', error)
      
      return {
        text: 'Sorry, I\'m currently unavailable. Agent Bruno might be temporarily down. Please try again later.',
        sources: ['Error Handler'],
      }
    }
  }

  // =============================================================================
  // 📊 LOG ANALYSIS
  // =============================================================================

  /**
   * Analyze logs
   */
  async analyzeLogs(logs: string, context?: string): Promise<LogAnalysisResponse> {
    const request: LogAnalysisRequest = {
      logs,
      context,
    }

    const response: AxiosResponse<LogAnalysisResponse> = await this.client.post('/analyze-logs', request)
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
      console.error('🤖 [ChatbotService] Failed to get Agent Bruno status:', error)
      return {
        status: 'error',
        error: 'Agent Bruno service is unavailable',
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

