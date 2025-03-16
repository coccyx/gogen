/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'cribl-primary': '#261926',
        'cribl-cyan': '#00F9BB',
        'cribl-purple': '#7f66ff',
        'cribl-pink': '#ff3399',
        'cribl-red': '#f25e65',
        'cribl-white': '#ffffff',
        'cribl-gray': '#f1f1f1',
        'cribl-dark': '#1A1A1A',
        'cribl-blue': '#3A85F7',
        'cribl-orange': '#FF6B35',
        'cribl-light-gray': '#F5F5F5',
      },
    },
  },
  plugins: [],
} 