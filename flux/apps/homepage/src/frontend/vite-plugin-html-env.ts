import { Plugin } from 'vite'

/**
 * Vite plugin to replace environment variable placeholders in HTML files
 * This allows us to inject environment variables into index.html at build time
 */
export function htmlEnvPlugin(): Plugin {
  return {
    name: 'html-env-plugin',
    transformIndexHtml(html) {
      // Replace __VITE_*__ placeholders with actual environment variable values
      return html.replace(/__VITE_(\w+)__/g, (match, key) => {
        const envKey = `VITE_${key}`
        const value = process.env[envKey]
        
        // If environment variable is not set, keep the placeholder as a warning
        if (!value) {
          console.warn(`⚠️  Warning: Environment variable ${envKey} is not set. Using placeholder.`)
          return match
        }
        
        return value
      })
    }
  }
}
