import axios from 'axios'
import chatbotService, { ChatbotService } from './chatbot'

// Mock axios
jest.mock('axios')
const mockedAxios = axios as jest.Mocked<typeof axios>

describe('ChatbotService', () => {
  let service: ChatbotService

  beforeEach(() => {
    // Create a fresh instance for each test
    service = new ChatbotService()
    
    // Mock axios.create to return a mocked instance
    const mockAxiosInstance = {
      post: jest.fn(),
      get: jest.fn(),
      interceptors: {
        request: { use: jest.fn() },
        response: { use: jest.fn() },
      },
    }
    mockedAxios.create.mockReturnValue(mockAxiosInstance as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  describe('initialize', () => {
    it('should initialize the service', () => {
      service.initialize()
      expect(service.getAgentInfo().initialized).toBe(true)
    })

    it('should not initialize twice', () => {
      service.initialize()
      service.initialize()
      expect(service.getAgentInfo().initialized).toBe(true)
    })
  })

  describe('chat', () => {
    it('should send a direct chat message successfully', async () => {
      const mockResponse = {
        data: {
          response: 'To debug a pod, use kubectl logs',
          timestamp: '2025-10-08T12:00:00Z',
          model: 'bruno-sre:latest',
        },
      }

      const mockPost = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.post = mockPost

      const result = await service.chat('How do I debug a pod?')

      expect(mockPost).toHaveBeenCalledWith('/chat', {
        message: 'How do I debug a pod?',
        timestamp: expect.any(String),
      })
      expect(result.response).toBe('To debug a pod, use kubectl logs')
      expect(result.mode).toBe('direct')
    })

    it('should handle errors in direct chat', async () => {
      const mockPost = jest.fn().mockRejectedValue(new Error('Network error'))
      ;(service as any).client.post = mockPost

      await expect(service.chat('test')).rejects.toThrow('Network error')
    })
  })

  describe('mcpChat', () => {
    it('should send an MCP chat message successfully', async () => {
      const mockResponse = {
        data: {
          response: 'Monitoring best practices include...',
          timestamp: '2025-10-08T12:00:00Z',
          sources: ['MCP Server', 'Knowledge Base'],
        },
      }

      const mockPost = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.post = mockPost

      const result = await service.mcpChat('Tell me about monitoring')

      expect(mockPost).toHaveBeenCalledWith('/mcp/chat', {
        message: 'Tell me about monitoring',
        timestamp: expect.any(String),
      })
      expect(result.response).toBe('Monitoring best practices include...')
      expect(result.mode).toBe('mcp')
      expect(result.sources).toContain('MCP Server')
    })
  })

  describe('processMessage', () => {
    it('should use MCP chat when available', async () => {
      const mockResponse = {
        data: {
          response: 'MCP response',
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      const mockPost = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.post = mockPost

      const result = await service.processMessage('test question')

      expect(mockPost).toHaveBeenCalledWith('/mcp/chat', expect.any(Object))
      expect(result.text).toBe('MCP response')
      expect(result.sources).toContain('Agent-SRE (MCP)')
    })

    it('should fallback to direct chat when MCP fails', async () => {
      const mockPost = jest.fn()
        .mockRejectedValueOnce(new Error('MCP unavailable'))
        .mockResolvedValueOnce({
          data: {
            response: 'Direct response',
            timestamp: '2025-10-08T12:00:00Z',
          },
        })
      ;(service as any).client.post = mockPost

      const result = await service.processMessage('test question')

      expect(mockPost).toHaveBeenCalledTimes(2)
      expect(mockPost).toHaveBeenNthCalledWith(1, '/mcp/chat', expect.any(Object))
      expect(mockPost).toHaveBeenNthCalledWith(2, '/chat', expect.any(Object))
      expect(result.text).toBe('Direct response')
      expect(result.sources).toContain('Agent-SRE (Direct)')
    })

    it('should return error message when both MCP and direct fail', async () => {
      const mockPost = jest.fn()
        .mockRejectedValueOnce(new Error('MCP unavailable'))
        .mockRejectedValueOnce(new Error('Direct unavailable'))
      ;(service as any).client.post = mockPost

      const result = await service.processMessage('test question')

      expect(result.text).toContain('currently unavailable')
      expect(result.sources).toContain('Error Handler')
    })
  })

  describe('analyzeLogsDirect', () => {
    it('should analyze logs using direct mode', async () => {
      const mockResponse = {
        data: {
          analysis: 'The error indicates a database connection issue',
          severity: 'high',
          recommendations: [
            'Check database connectivity',
            'Verify credentials',
          ],
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      const mockPost = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.post = mockPost

      const result = await service.analyzeLogsDirect(
        'ERROR: Connection timeout',
        'Production API'
      )

      expect(mockPost).toHaveBeenCalledWith('/analyze-logs', {
        logs: 'ERROR: Connection timeout',
        context: 'Production API',
      })
      expect(result.severity).toBe('high')
      expect(result.recommendations).toHaveLength(2)
    })
  })

  describe('analyzeLogsMCP', () => {
    it('should analyze logs using MCP mode', async () => {
      const mockResponse = {
        data: {
          analysis: 'Memory pressure detected',
          severity: 'medium',
          recommendations: ['Increase memory limits'],
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      const mockPost = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.post = mockPost

      const result = await service.analyzeLogsMCP('WARN: High memory usage')

      expect(mockPost).toHaveBeenCalledWith('/mcp/analyze-logs', {
        logs: 'WARN: High memory usage',
        context: undefined,
      })
      expect(result.analysis).toBe('Memory pressure detected')
    })
  })

  describe('healthCheck', () => {
    it('should return healthy status', async () => {
      const mockResponse = {
        data: { status: 'healthy' },
      }

      const mockGet = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.get = mockGet

      const result = await service.healthCheck()

      expect(mockGet).toHaveBeenCalledWith('/health')
      expect(result.status).toBe('healthy')
    })

    it('should handle health check errors', async () => {
      const mockGet = jest.fn().mockRejectedValue(new Error('Service down'))
      ;(service as any).client.get = mockGet

      await expect(service.healthCheck()).rejects.toThrow('Service down')
    })
  })

  describe('getStatus', () => {
    it('should return detailed agent status', async () => {
      const mockResponse = {
        data: {
          status: 'healthy',
          service: 'sre-agent',
          timestamp: '2025-10-08T12:00:00Z',
          mcp_server: {
            status: 'healthy',
            url: 'http://mcp-server:30120',
          },
        },
      }

      const mockGet = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.get = mockGet

      const result = await service.getStatus()

      expect(mockGet).toHaveBeenCalledWith('/status')
      expect(result.status).toBe('healthy')
      expect(result.mcp_server?.status).toBe('healthy')
    })
  })

  describe('getLLMStatus', () => {
    it('should return healthy LLM status', async () => {
      const mockResponse = {
        data: {
          status: 'healthy',
          service: 'sre-agent',
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      const mockGet = jest.fn().mockResolvedValue(mockResponse)
      ;(service as any).client.get = mockGet

      const result = await service.getLLMStatus()

      expect(result.status).toBe('healthy')
      expect(result.agent_status).toBeDefined()
    })

    it('should return error status when service is down', async () => {
      const mockGet = jest.fn().mockRejectedValue(new Error('Connection refused'))
      ;(service as any).client.get = mockGet

      const result = await service.getLLMStatus()

      expect(result.status).toBe('error')
      expect(result.error).toBe('Agent-SRE service is unavailable')
    })
  })

  describe('isAvailable', () => {
    it('should return true when service is available', async () => {
      const mockGet = jest.fn().mockResolvedValue({ data: { status: 'healthy' } })
      ;(service as any).client.get = mockGet

      const result = await service.isAvailable()

      expect(result).toBe(true)
    })

    it('should return false when service is unavailable', async () => {
      const mockGet = jest.fn().mockRejectedValue(new Error('Service down'))
      ;(service as any).client.get = mockGet

      const result = await service.isAvailable()

      expect(result).toBe(false)
    })
  })

  describe('getAgentInfo', () => {
    it('should return agent info', () => {
      const info = service.getAgentInfo()

      expect(info).toHaveProperty('baseUrl')
      expect(info).toHaveProperty('initialized')
      expect(typeof info.baseUrl).toBe('string')
      expect(typeof info.initialized).toBe('boolean')
    })
  })
})

