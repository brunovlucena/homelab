/**
 * 🖼️ Asset URL Helper
 * 
 * Provides utility functions to get asset URLs that are proxied through the API
 * This allows the browser to fetch assets from MinIO without direct internet access to MinIO
 */

/**
 * Get the base API URL from environment or default to '/api/v1'
 */
const getApiBaseUrl = (): string => {
  return import.meta.env.VITE_API_URL || '/api/v1'
}

/**
 * Get a proxied asset URL
 * 
 * @param assetPath - The path to the asset in MinIO (e.g., 'eu.webp', 'logos/bruno-logo.png')
 * @returns The full URL to the proxied asset through the API
 * 
 * @example
 * getAssetUrl('eu.webp') // Returns: '/api/v1/assets/eu.webp'
 * getAssetUrl('logos/bruno-logo.png') // Returns: '/api/v1/assets/logos/bruno-logo.png'
 */
export const getAssetUrl = (assetPath: string): string => {
  // Remove leading slash if present
  const cleanPath = assetPath.startsWith('/') ? assetPath.slice(1) : assetPath
  
  // Construct the proxied asset URL
  return `${getApiBaseUrl()}/assets/${cleanPath}`
}

/**
 * Check if we're in development mode
 */
export const isDevelopment = (): boolean => {
  return import.meta.env.DEV || import.meta.env.MODE === 'development'
}

/**
 * Get the full URL for an asset (useful for preloading)
 * 
 * @param assetPath - The path to the asset in MinIO
 * @returns The full URL including hostname
 */
export const getFullAssetUrl = (assetPath: string): string => {
  const assetUrl = getAssetUrl(assetPath)
  
  // If it's already a full URL, return as-is
  if (assetUrl.startsWith('http://') || assetUrl.startsWith('https://')) {
    return assetUrl
  }
  
  // Construct full URL with current origin
  return `${window.location.origin}${assetUrl}`
}

