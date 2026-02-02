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
        // McDonald's Fast Food Theme
        'mc-black': '#1a1a1a',
        'mc-dark': '#242424',
        'mc-gray': '#333333',
        'mc-red': '#da291c',
        'mc-red-dark': '#bf0811',
        'mc-gold': '#ffc72c',
        'mc-yellow': '#ffbc0d',
        'mc-green': '#27ae60',
        'mc-orange': '#f39c12',
        'mc-blue': '#3498db',
        'mc-white': '#ffffff',
        // Order statuses
        'status-new': '#3498db',
        'status-preparing': '#f39c12',
        'status-ready': '#27ae60',
        'status-delivered': '#9b59b6',
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        display: ['Inter', 'system-ui', 'sans-serif'],
        brand: ['Poppins', 'sans-serif'],
      },
      backgroundImage: {
        'mc-pattern': `
          radial-gradient(circle at 20% 80%, rgba(218, 41, 28, 0.1) 0%, transparent 50%),
          radial-gradient(circle at 80% 20%, rgba(255, 199, 44, 0.1) 0%, transparent 50%)
        `,
        'order-gradient': 'linear-gradient(135deg, rgba(218, 41, 28, 0.1) 0%, rgba(255, 199, 44, 0.05) 100%)',
      },
      animation: {
        'order-pulse': 'order-pulse 2s ease-in-out infinite',
        'timer-tick': 'timer-tick 1s steps(1) infinite',
        'slide-in': 'slide-in 0.3s ease-out',
        'bell-ring': 'bell-ring 0.5s ease-in-out',
      },
      keyframes: {
        'order-pulse': {
          '0%, 100%': { boxShadow: '0 0 15px rgba(218, 41, 28, 0.3)' },
          '50%': { boxShadow: '0 0 30px rgba(218, 41, 28, 0.6)' },
        },
        'timer-tick': {
          '0%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
        'slide-in': {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        'bell-ring': {
          '0%, 100%': { transform: 'rotate(0deg)' },
          '25%': { transform: 'rotate(15deg)' },
          '75%': { transform: 'rotate(-15deg)' },
        },
      },
      boxShadow: {
        'mc-red': '0 0 20px rgba(218, 41, 28, 0.4)',
        'mc-gold': '0 0 20px rgba(255, 199, 44, 0.4)',
        'order': '0 4px 20px rgba(0, 0, 0, 0.3)',
        'card': '0 2px 10px rgba(0, 0, 0, 0.2)',
      },
    },
  },
  plugins: [],
}

export default config
