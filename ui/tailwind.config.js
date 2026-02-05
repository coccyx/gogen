/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Legacy cribl colors (can be removed later)
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
        // Terminal-style dark theme colors
        'term-bg': '#0d1117',
        'term-bg-elevated': '#161b22',
        'term-bg-muted': '#21262d',
        'term-border': '#30363d',
        'term-text': '#e6edf3',
        'term-text-muted': '#8b949e',
        'term-green': '#3fb950',
        'term-cyan': '#58a6ff',
        'term-red': '#f85149',
      },
      fontFamily: {
        'mono': ['JetBrains Mono', 'ui-monospace', 'SFMono-Regular', 'monospace'],
      },
      borderRadius: {
        'sm': '4px',
        DEFAULT: '4px',
      },
    },
  },
  plugins: [],
} 