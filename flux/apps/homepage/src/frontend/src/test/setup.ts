import '@testing-library/jest-dom'
import { afterEach, vi, beforeEach } from 'vitest'
import { cleanup } from '@testing-library/react'

// Mock fetch globally
global.fetch = vi.fn()

// Setup mocks before each test
beforeEach(() => {
  // Reset fetch mock
  vi.mocked(global.fetch).mockReset()
  
  // Default mock for fetch - handle API calls gracefully
  vi.mocked(global.fetch).mockImplementation((url: string | URL | Request) => {
    const urlString = typeof url === 'string' ? url : url.toString()
    
    // Mock health check endpoint
    if (urlString.includes('/api/chat/health')) {
      return Promise.resolve({
        ok: true,
        json: async () => ({
          status: 'healthy',
          model: 'gemma3n:e4b',
          provider: 'ollama'
        })
      } as Response)
    }
    
    // Mock chat endpoint
    if (urlString.includes('/api/chat')) {
      return Promise.resolve({
        ok: true,
        json: async () => ({
          response: 'Test response',
          model: 'gemma3n:e4b',
          timestamp: new Date().toISOString()
        })
      } as Response)
    }
    
    // Default: return a successful empty response to prevent network errors
    return Promise.resolve({
      ok: true,
      json: async () => ({}),
      status: 200,
      statusText: 'OK'
    } as Response)
  })
})

// Cleanup after each test
afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

