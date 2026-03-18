import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        background: '#F9FAFB', // gray-50
        foreground: '#111827', // gray-900
        card: {
          DEFAULT: '#FFFFFF',
          foreground: '#111827',
        },
        muted: {
          DEFAULT: '#F3F4F6', // gray-100
          foreground: '#6B7280', // gray-500
        },
        primary: {
          DEFAULT: '#F43F5E',
          foreground: '#FFFFFF',
          50: '#FFF1F2',
          100: '#FFE4E6',
          200: '#FECDD3',
          300: '#FDA4AF',
          400: '#FB7185',
          500: '#F43F5E',
          600: '#E11D48',
          700: '#BE123C',
          800: '#9F1239',
          900: '#881337',
        },
        trust: {
          high: '#22C55E',
          medium: '#F59E0B',
          low: '#EF4444',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

export default config;
