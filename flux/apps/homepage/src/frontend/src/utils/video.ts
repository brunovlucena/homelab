// =============================================================================
// 🎥 VIDEO URL UTILITIES
// =============================================================================

/**
 * Supported video platform domains
 */
const VIDEO_DOMAINS = [
  'youtube.com',
  'youtu.be',
  'vimeo.com',
  'dailymotion.com',
  'twitch.tv'
] as const

/**
 * Check if a URL is a video URL from a supported platform
 * 
 * @param url - The URL to check
 * @returns true if the URL is from a supported video platform
 * 
 * @example
 * isVideoUrl('https://youtube.com/watch?v=abc123') // true
 * isVideoUrl('https://example.com') // false
 */
export function isVideoUrl(url: string): boolean {
  try {
    const urlObj = new URL(url)
    return VIDEO_DOMAINS.some(domain => urlObj.hostname.includes(domain))
  } catch {
    return false
  }
}

/**
 * Check if a URL is a YouTube URL
 * 
 * @param url - The URL to check
 * @returns true if the URL is from YouTube
 * 
 * @example
 * isYouTubeUrl('https://youtube.com/watch?v=abc123') // true
 * isYouTubeUrl('https://youtu.be/abc123') // true
 */
export function isYouTubeUrl(url: string): boolean {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')
  } catch {
    return false
  }
}

/**
 * Extract YouTube video ID from a URL
 * 
 * @param url - YouTube URL (youtube.com or youtu.be)
 * @returns Video ID if found, null otherwise
 * 
 * @example
 * getYouTubeVideoId('https://youtube.com/watch?v=abc123') // 'abc123'
 * getYouTubeVideoId('https://youtu.be/abc123') // 'abc123'
 */
export function getYouTubeVideoId(url: string): string | null {
  try {
    const urlObj = new URL(url)
    if (urlObj.hostname.includes('youtube.com')) {
      return urlObj.searchParams.get('v')
    }
    if (urlObj.hostname.includes('youtu.be')) {
      return urlObj.pathname.slice(1)
    }
    return null
  } catch {
    return null
  }
}

/**
 * Get the embed URL for a video from various platforms
 * 
 * @param url - The video URL
 * @returns Embed URL for the video, or original URL if conversion fails
 * 
 * @example
 * getVideoEmbedUrl('https://youtube.com/watch?v=abc123') 
 * // Returns 'https://www.youtube.com/embed/abc123'
 */
export function getVideoEmbedUrl(url: string): string {
  try {
    const urlObj = new URL(url)
    
    // YouTube - handled by lite-youtube-embed
    if (urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')) {
      const videoId = urlObj.searchParams.get('v') || urlObj.pathname.slice(1)
      return `https://www.youtube.com/embed/${videoId}`
    }
    
    // Vimeo
    if (urlObj.hostname.includes('vimeo.com')) {
      const videoId = urlObj.pathname.slice(1)
      return `https://player.vimeo.com/video/${videoId}`
    }
    
    // Dailymotion
    if (urlObj.hostname.includes('dailymotion.com')) {
      const videoId = urlObj.pathname.split('/').pop() || ''
      return `https://www.dailymotion.com/embed/video/${videoId}`
    }
    
    // Twitch
    if (urlObj.hostname.includes('twitch.tv')) {
      const videoId = urlObj.pathname.split('/').pop() || ''
      return `https://player.twitch.tv/?video=v${videoId}&parent=${window.location.hostname}`
    }
    
    return url
  } catch {
    return url
  }
}
