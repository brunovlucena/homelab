import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Store brand colors
        'store-dark': '#0a0a0f',
        'store-darker': '#050508',
        'store-gray': '#1a1a24',
        'store-purple': '#8b5cf6',
        'store-pink': '#ec4899',
        'store-blue': '#3b82f6',
        'store-green': '#10b981',
        'store-yellow': '#f59e0b',
        'store-red': '#ef4444',
        'store-orange': '#f97316',
        // Brand colors
        'brand-fashion': '#ec4899',
        'brand-tech': '#3b82f6',
        'brand-gaming': '#8b5cf6',
        'brand-beauty': '#f472b6',
        'brand-home': '#10b981',
      },
      fontFamily: {
        'display': ['Inter', 'system-ui', 'sans-serif'],
        'mono': ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
      },
      keyframes: {
        glow: {
          '0%': { boxShadow: '0 0 5px rgb(139 92 246 / 0.5)' },
          '100%': { boxShadow: '0 0 20px rgb(139 92 246 / 0.8)' },
        },
      },
    },
  },
  plugins: [],
}

export default config
