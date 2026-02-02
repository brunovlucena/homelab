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
        'cyber-dark': '#0a0e1a',
        'cyber-purple': '#a855f7',
        'cyber-pink': '#ec4899',
        'cyber-blue': '#3b82f6',
        'cyber-cyan': '#06b6d4',
        'cyber-green': '#10b981',
        'cyber-yellow': '#f59e0b',
        'cyber-red': '#ef4444',
        'medical-blue': '#1e40af',
        'medical-green': '#059669',
        'medical-red': '#dc2626',
      },
      fontFamily: {
        sans: ['var(--font-inter)'],
        mono: ['var(--font-mono)'],
      },
    },
  },
  plugins: [],
}
export default config
