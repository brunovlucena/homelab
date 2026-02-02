import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Restaurant elegant palette
        'wine': {
          50: '#fdf2f4',
          100: '#fce7ea',
          200: '#f9d2d9',
          300: '#f4adb9',
          400: '#ec7d93',
          500: '#e04d6f',
          600: '#cc3057',
          700: '#ab2347',
          800: '#8f2040',
          900: '#722038', // Primary burgundy
          950: '#450c1c',
        },
        'gold': {
          50: '#fefce8',
          100: '#fef9c3',
          200: '#fef08a',
          300: '#fde047',
          400: '#facc15',
          500: '#d4a012', // Primary gold
          600: '#a17c0a',
          700: '#7c5e0d',
          800: '#664c12',
          900: '#563f15',
        },
        'cream': {
          50: '#fefdfb',
          100: '#fdf8f3', // Primary cream
          200: '#f9efe3',
          300: '#f3e0c8',
          400: '#e8c79c',
          500: '#dcab70',
        },
        'wood': {
          50: '#f9f6f3',
          100: '#f1ebe3',
          200: '#e1d4c5',
          300: '#cdb8a0',
          400: '#b69778',
          500: '#a5805c',
          600: '#8f6a4a', // Dark wood
          700: '#775643',
          800: '#63483b',
          900: '#533d34',
        },
      },
      fontFamily: {
        'serif': ['Playfair Display', 'Georgia', 'serif'],
        'sans': ['Inter', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'restaurant-pattern': "url('/pattern.svg')",
      },
    },
  },
  plugins: [],
}

export default config
