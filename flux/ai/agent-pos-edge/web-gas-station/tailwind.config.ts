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
        // Gas Station Industrial Theme
        'fuel-black': '#0d1117',
        'fuel-dark': '#161b22',
        'fuel-gray': '#21262d',
        'fuel-green': '#22c55e',
        'fuel-lime': '#84cc16',
        'fuel-orange': '#f97316',
        'fuel-amber': '#f59e0b',
        'fuel-red': '#ef4444',
        'fuel-blue': '#3b82f6',
        'fuel-cyan': '#06b6d4',
        'fuel-yellow': '#eab308',
        'diesel': '#f59e0b',
        'gasoline': '#22c55e',
        'premium': '#3b82f6',
        'ethanol': '#84cc16',
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        display: ['Inter', 'system-ui', 'sans-serif'],
        industrial: ['Roboto Condensed', 'sans-serif'],
      },
      backgroundImage: {
        'fuel-grid': `
          linear-gradient(to right, rgba(34, 197, 94, 0.05) 1px, transparent 1px),
          linear-gradient(to bottom, rgba(34, 197, 94, 0.05) 1px, transparent 1px)
        `,
        'tank-gradient': 'linear-gradient(180deg, rgba(34, 197, 94, 0.2) 0%, rgba(34, 197, 94, 0.05) 100%)',
      },
      backgroundSize: {
        'grid': '24px 24px',
      },
      animation: {
        'pulse-fuel': 'pulse-fuel 2s ease-in-out infinite',
        'fill-up': 'fill-up 1s ease-out',
        'pump': 'pump 0.5s ease-in-out infinite',
        'flow': 'flow 2s linear infinite',
      },
      keyframes: {
        'pulse-fuel': {
          '0%, 100%': { boxShadow: '0 0 15px rgba(34, 197, 94, 0.4)' },
          '50%': { boxShadow: '0 0 30px rgba(34, 197, 94, 0.7)' },
        },
        'fill-up': {
          '0%': { height: '0%' },
          '100%': { height: 'var(--fill-level)' },
        },
        'pump': {
          '0%, 100%': { transform: 'scaleY(1)' },
          '50%': { transform: 'scaleY(0.95)' },
        },
        'flow': {
          '0%': { backgroundPosition: '0 0' },
          '100%': { backgroundPosition: '100% 0' },
        },
      },
      boxShadow: {
        'fuel': '0 0 20px rgba(34, 197, 94, 0.3)',
        'fuel-lg': '0 0 40px rgba(34, 197, 94, 0.4)',
        'tank': 'inset 0 0 30px rgba(34, 197, 94, 0.2)',
        'alert': '0 0 20px rgba(239, 68, 68, 0.5)',
      },
    },
  },
  plugins: [],
}

export default config
