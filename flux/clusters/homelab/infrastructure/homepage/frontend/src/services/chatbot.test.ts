// Mock axios before importing anything
jest.mock('axios', () => ({
  create: jest.fn(() => ({
    post: jest.fn(),
    get: jest.fn(),
    interceptors: {
      request: { use: jest.fn() },
      response: { use: jest.fn() },
    },
  })),
}))

import axios from 'axios'
import chatbotService, { ChatbotService } from './chatbot'

const mockedAxios = axios as jest.Mocked<typeof axios>

describe('ChatbotService', () => {
  let service: ChatbotService
  let mockAxiosInstance: any

  beforeAll(() => {
    // Set up mock axios instance
    mockAxiosInstance = {
      post: jest.fn(),
      get: jest.fn(),
      interceptors: {
        request: { use: jest.fn() },
        response: { use: jest.fn() },
      },
    }
    mockedAxios.create.mockReturnValue(mockAxiosInstance)
  })

  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks()
    mockAxiosInstance.post.mockClear()
    mockAxiosInstance.get.mockClear()
    
    // Create a fresh instance for each test
    service = new ChatbotService()
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
    it('should send a chat message successfully', async () => {
      const mockResponse = {
        data: {
          response: 'To debug a pod, use kubectl logs',
          timestamp: '2025-10-08T12:00:00Z',
          model: 'bruno-sre:latest',
        },
      }

      mockAxiosInstance.post.mockResolvedValue(mockResponse)

      const result = await service.chat('How do I debug a pod?')

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/chat', {
        message: 'How do I debug a pod?',
        timestamp: expect.any(String),
      })
      expect(result.response).toBe('To debug a pod, use kubectl logs')
      // Note: 'mode' field no longer exists in Jamie implementation
    })

    it('should handle errors in chat', async () => {
      mockAxiosInstance.post.mockRejectedValue(new Error('Network error'))

      await expect(service.chat('test')).rejects.toThrow('Network error')
    })
  })

  describe('processMessage', () => {
    it('should process message successfully', async () => {
      const mockResponse = {
        data: {
          response: 'Agent Bruno response',
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      mockAxiosInstance.post.mockResolvedValue(mockResponse)

      const result = await service.processMessage('test question')

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/chat', expect.any(Object))
      expect(result.text).toBe('Agent Bruno response')
      expect(result.sources).toContain('Agent Bruno')
    })

    it('should return error message when chat fails', async () => {
      mockAxiosInstance.post.mockRejectedValue(new Error('Service unavailable'))

      const result = await service.processMessage('test question')

      expect(result.text).toContain('currently unavailable')
      expect(result.text).toContain('Agent Bruno')
      expect(result.sources).toContain('Error Handler')
    })
  })

  describe('analyzeLogs', () => {
    it('should analyze logs successfully', async () => {
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

      mockAxiosInstance.post.mockResolvedValue(mockResponse)

      const result = await service.analyzeLogs(
        'ERROR: Connection timeout',
        'Production API'
      )

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/analyze-logs', {
        logs: 'ERROR: Connection timeout',
        context: 'Production API',
      })
      expect(result.severity).toBe('high')
      expect(result.recommendations).toHaveLength(2)
    })

    it('should handle log analysis errors', async () => {
      mockAxiosInstance.post.mockRejectedValue(new Error('Analysis failed'))

      await expect(service.analyzeLogs('WARN: High memory usage')).rejects.toThrow('Analysis failed')
    })
  })

  describe('healthCheck', () => {
    it('should return healthy status', async () => {
      const mockResponse = {
        data: { status: 'healthy' },
      }

      mockAxiosInstance.get.mockResolvedValue(mockResponse)

      const result = await service.healthCheck()

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/health')
      expect(result.status).toBe('healthy')
    })

    it('should handle health check errors', async () => {
      mockAxiosInstance.get.mockRejectedValue(new Error('Service down'))

      await expect(service.healthCheck()).rejects.toThrow('Service down')
    })
  })

  describe('getStatus', () => {
    it('should return detailed agent status', async () => {
      const mockResponse = {
        data: {
          status: 'healthy',
          service: 'agent-bruno',
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      mockAxiosInstance.get.mockResolvedValue(mockResponse)

      const result = await service.getStatus()

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/status')
      expect(result.status).toBe('healthy')
      expect(result.service).toBe('agent-bruno')
    })
  })

  describe('getLLMStatus', () => {
    it('should return healthy LLM status', async () => {
      const mockResponse = {
        data: {
          status: 'healthy',
          service: 'agent-bruno',
          timestamp: '2025-10-08T12:00:00Z',
        },
      }

      mockAxiosInstance.get.mockResolvedValue(mockResponse)

      const result = await service.getLLMStatus()

      expect(result.status).toBe('healthy')
      expect(result.agent_status).toBeDefined()
    })

    it('should return error status when service is down', async () => {
      mockAxiosInstance.get.mockRejectedValue(new Error('Connection refused'))

      const result = await service.getLLMStatus()

      expect(result.status).toBe('error')
      expect(result.error).toBe('Agent Bruno service is unavailable')
    })
  })

  describe('isAvailable', () => {
    it('should return true when service is available', async () => {
      mockAxiosInstance.get.mockResolvedValue({ data: { status: 'healthy' } })

      const result = await service.isAvailable()

      expect(result).toBe(true)
    })

    it('should return false when service is unavailable', async () => {
      mockAxiosInstance.get.mockRejectedValue(new Error('Service down'))

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

    it('should include Agent Bruno endpoint in baseUrl', () => {
      const info = service.getAgentInfo()
      
      expect(info.baseUrl).toContain('agent-bruno')
    })
  })
})
