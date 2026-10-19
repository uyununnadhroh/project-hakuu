/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./pages/**/*.{js,jsx}",
    "./components/**/*.{js,jsx}",
    "./utils/**/*.{js,jsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#4c0c13',
        'primary-hover': '#3a0a0f',
        'primary-light': '#fcebc8',
        'secondary': '#daa278',
        'bg-base': '#faf6ee',
        'border-base': '#eadfd0',
        'text-dark': '#2b1d16',
        'text-muted': '#8a7768',
        'text-warm': '#7e4722',
        'success': '#16a34a',
        'warning': '#d97706',
        'error': '#dc2626',
      },
      fontFamily: {
        inter: ['Inter', 'sans-serif'],
      },
      borderRadius: {
        'card': '12px',
      },
      boxShadow: {
        'card': '0px 3px 16px 0px rgba(89,56,33,0.03)',
        'modal': '0px 8px 32px 0px rgba(89,56,33,0.12)',
      },
    },
  },
  plugins: [],
}
