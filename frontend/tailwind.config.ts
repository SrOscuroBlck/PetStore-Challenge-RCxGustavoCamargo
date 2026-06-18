import type { Config } from 'tailwindcss';

// Colors resolve to CSS variables (src/styles/tokens.css) so themes switch without
// recompiling Tailwind. See docs/DESIGN.md for the full system.
const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: 'rgb(var(--color-bg) / <alpha-value>)',
        fg: 'rgb(var(--color-fg) / <alpha-value>)',
        card: 'rgb(var(--color-card) / <alpha-value>)',
        border: 'rgb(var(--color-border) / <alpha-value>)',
        muted: 'rgb(var(--color-muted) / <alpha-value>)',
        ink: {
          DEFAULT: 'rgb(var(--color-ink) / <alpha-value>)',
          fg: 'rgb(var(--color-ink-fg) / <alpha-value>)',
        },
        primary: {
          DEFAULT: 'rgb(var(--color-primary) / <alpha-value>)',
          strong: 'rgb(var(--color-primary-strong) / <alpha-value>)',
          fg: 'rgb(var(--color-primary-fg) / <alpha-value>)',
        },
        accent: 'rgb(var(--color-accent) / <alpha-value>)',
        cat: 'rgb(var(--color-cat) / <alpha-value>)',
        dog: 'rgb(var(--color-dog) / <alpha-value>)',
        frog: 'rgb(var(--color-frog) / <alpha-value>)',
        danger: 'rgb(var(--color-danger) / <alpha-value>)',
        success: 'rgb(var(--color-success) / <alpha-value>)',
      },
      borderRadius: {
        DEFAULT: 'var(--radius)',
        lg: 'calc(var(--radius) + 0.25rem)',
        xl: 'calc(var(--radius) + 0.5rem)',
        '2xl': 'calc(var(--radius) + 1rem)',
      },
      fontFamily: {
        // Body/UI — warm + rounded + legible. Display — friendly rounded face for character.
        sans: ['Nunito', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        display: ['Fredoka', 'ui-rounded', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      letterSpacing: {
        eyebrow: '0.14em',
      },
      boxShadow: {
        // Soft, warm elevation — pet-commerce friendliness over hard tech borders.
        soft: '0 2px 8px -2px rgb(38 33 30 / 0.08), 0 6px 20px -8px rgb(38 33 30 / 0.10)',
        lift: '0 8px 28px -10px rgb(242 101 34 / 0.28), 0 4px 12px -6px rgb(38 33 30 / 0.12)',
      },
      keyframes: {
        'fade-in-up': {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'pop-in': {
          '0%': { opacity: '0', transform: 'scale(0.96)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        wiggle: {
          '0%, 100%': { transform: 'rotate(-7deg)' },
          '50%': { transform: 'rotate(7deg)' },
        },
      },
      animation: {
        'fade-in-up': 'fade-in-up 0.35s ease-out both',
        'pop-in': 'pop-in 0.3s ease-out both',
        wiggle: 'wiggle 0.5s ease-in-out',
      },
    },
  },
  plugins: [],
};

export default config;
