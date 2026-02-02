import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { htmlEnvPlugin } from './vite-plugin-html-env'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    htmlEnvPlugin(),
  ],
  // Use /tmp for cache directory to avoid permission issues in containers
  // This is especially important when using volume mounts with Telepresence
  // Always use /tmp in containers (when NODE_ENV is not production or when in Docker)
  cacheDir: process.env.NODE_ENV === 'production' ? undefined : '/tmp/.vite',
  // Force Vite to use /tmp for temp files during config loading
  // This prevents EACCES errors when node_modules is mounted as volume
  optimizeDeps: {
    // Store optimized deps in /tmp to avoid permission issues
    ...(process.env.NODE_ENV !== 'production' ? { cacheDir: '/tmp/.vite-optimize' } : {}),
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      buffer: 'buffer',
    },
  },
  define: {
    'global': 'globalThis',
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
    },
  },
  server: {
    port: 8080,
    host: '0.0.0.0',
    watch: {
      usePolling: true,
      interval: 1000,
    },
    proxy: {
      '/api': {
        // Port-forward mode: use localhost:8081
        // Normal dev mode: use K8s service DNS
        target: process.env.USE_LOCALHOST_API === 'true'
          ? 'http://localhost:8081'
          : 'http://homepage-api.homepage.svc:8080',
        changeOrigin: true,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('proxy error', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('Sending Request to the Target:', req.method, req.url);
          });
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
          });
        },
      },
    },
    headers: {
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      'X-XSS-Protection': '1; mode=block',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false, // Disable in production for smaller bundle
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom']
        }
      }
    }
  },
}) 