/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#00D4AA',
        secondary: '#7B61FF',
        accent: '#FFD700',
        dark: '#1A1A2E',
        darker: '#0F0F1A',
      },
    },
  },
  plugins: [],
};

