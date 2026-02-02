import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { getAssetUrl } from './index'

describe('getAssetUrl', () => {
  beforeEach(() => {
    // Reset environment variables before each test
    vi.stubEnv('VITE_CDN_BASE_URL', '')
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('should return relative path when CDN is not configured', () => {
    const result = getAssetUrl('assets/eu.webp')
    expect(result).toBe('./assets/eu.webp')
  })

  it('should handle paths with leading ./', () => {
    const result = getAssetUrl('./assets/eu.png')
    expect(result).toBe('./assets/eu.png')
  })

  it('should handle paths with leading /', () => {
    const result = getAssetUrl('/assets/eu.png')
    expect(result).toBe('./assets/eu.png')
  })

  it('should return CDN URL when VITE_CDN_BASE_URL is set', () => {
    vi.stubEnv('VITE_CDN_BASE_URL', 'https://storage.googleapis.com/my-bucket')
    const result = getAssetUrl('assets/eu.webp')
    expect(result).toBe('https://storage.googleapis.com/my-bucket/assets/eu.webp')
  })

  it('should handle CDN URL with trailing slash', () => {
    vi.stubEnv('VITE_CDN_BASE_URL', 'https://storage.googleapis.com/my-bucket/')
    const result = getAssetUrl('assets/eu.png')
    expect(result).toBe('https://storage.googleapis.com/my-bucket/assets/eu.png')
  })

  it('should handle custom CDN domain', () => {
    vi.stubEnv('VITE_CDN_BASE_URL', 'https://cdn.example.com')
    const result = getAssetUrl('assets/image.jpg')
    expect(result).toBe('https://cdn.example.com/assets/image.jpg')
  })

  it('should normalize various path formats to same CDN URL', () => {
    vi.stubEnv('VITE_CDN_BASE_URL', 'https://storage.googleapis.com/my-bucket')
    
    const path1 = getAssetUrl('assets/eu.webp')
    const path2 = getAssetUrl('./assets/eu.webp')
    const path3 = getAssetUrl('/assets/eu.webp')
    
    expect(path1).toBe(path2)
    expect(path2).toBe(path3)
    expect(path1).toBe('https://storage.googleapis.com/my-bucket/assets/eu.webp')
  })

  it('should handle nested asset paths', () => {
    vi.stubEnv('VITE_CDN_BASE_URL', 'https://cdn.example.com')
    const result = getAssetUrl('assets/logos/logo.png')
    expect(result).toBe('https://cdn.example.com/assets/logos/logo.png')
  })
})
