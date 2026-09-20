/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Single accent. Everything else is neutral — no secondary violet/purple.
        brand: {
          50: '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
          950: '#1e1b4b',
        },
      },
      // The app was written with text-xs as its body size (12px, too small to read
      // comfortably). Nudging the scale here fixes every call site at once.
      fontSize: {
        '2xs': ['0.6875rem', { lineHeight: '1rem' }],      // 11px — badges, meta only
        xs: ['0.8125rem', { lineHeight: '1.25rem' }],      // 13px
        sm: ['0.875rem', { lineHeight: '1.375rem' }],      // 14px — body
        base: ['0.9375rem', { lineHeight: '1.5rem' }],     // 15px
        lg: ['1.0625rem', { lineHeight: '1.5rem' }],       // 17px
        xl: ['1.25rem', { lineHeight: '1.75rem' }],        // 20px
        '2xl': ['1.5rem', { lineHeight: '2rem' }],         // 24px
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],    // 30px
      },
      borderRadius: {
        // Flatten the 8px/12px/16px/24px sprawl to two real steps.
        lg: '0.5rem',
        xl: '0.625rem',
        '2xl': '0.625rem',
        '3xl': '0.75rem',
      },
    },
  },
  plugins: [],
}
