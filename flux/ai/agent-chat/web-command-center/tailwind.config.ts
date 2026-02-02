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
        'cyber-dark': '#0a0a0f',
        'cyber-gray': '#1a1a2e',
        'cyber-purple': '#8b5cf6',
        'cyber-pink': '#ec4899',
        'cyber-blue': '#3b82f6',
        'cyber-cyan': '#06b6d4',
        'cyber-green': '#10b981',
        'cyber-yellow': '#f59e0b',
        'cyber-red': '#ef4444',
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
        mono: ['var(--font-mono)', 'monospace'],
      },
      boxShadow: {
        'cyber': '0 0 20px rgba(139, 92, 246, 0.3)',
        'cyber-lg': '0 0 40px rgba(139, 92, 246, 0.4)',
      },
    },
  },
  plugins: [],
}

export default config
