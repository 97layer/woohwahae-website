import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // 우화해 브랜드 컬러 (미니멀 & 고급스러움)
        'brand': {
          black: '#0A0A0A',
          white: '#FAFAFA',
          gold: '#D4AF37',
          gray: {
            50: '#F8F8F8',
            100: '#E8E8E8',
            200: '#D1D1D1',
            300: '#B0B0B0',
            400: '#888888',
            500: '#6D6D6D',
            600: '#5D5D5D',
            700: '#4F4F4F',
            800: '#454545',
            900: '#3A3A3A',
          }
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        display: ['Inter', 'system-ui', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
export default config
