import type { Config } from "tailwindcss";

const config: Config = {
    content: [
        "./pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./components/**/*.{js,ts,jsx,tsx,mdx}",
        "./app/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: {
                    50: '#FAF5FF',
                    100: '#F3E8FF',
                    200: '#E9D5FF',
                    300: '#D8B4FE',
                    400: '#C084FC',
                    500: '#A855F7',
                    600: '#9333EA',
                    700: '#7C3AED',
                    800: '#6D28D9',
                    900: '#5B21B6',
                    950: '#4C1D95',
                },
                purple: {
                    light: '#EDE9FE',
                    DEFAULT: '#6D28D9',
                    hover: '#5B21B6',
                    dark: '#4C1D95',
                },
                success: {
                    light: '#D1FAE5',
                    DEFAULT: '#059669',
                    dark: '#047857',
                },
                warning: {
                    light: '#FEF3C7',
                    DEFAULT: '#D97706',
                    dark: '#B45309',
                },
                error: {
                    light: '#FEE2E2',
                    DEFAULT: '#DC2626',
                    dark: '#B91C1C',
                },
                info: {
                    light: '#E0F2FE',
                    DEFAULT: '#0284C7',
                    dark: '#0369A1',
                },
                gray: {
                    50: '#F9FAFB',
                    100: '#F3F4F6',
                    200: '#E5E7EB',
                    300: '#D1D5DB',
                    400: '#9CA3AF',
                    500: '#6B7280',
                    600: '#4B5563',
                    700: '#374151',
                    800: '#1F2937',
                    900: '#111827',
                    950: '#030712',
                },
            },
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
                mono: ['Fira Code', 'monospace'],
            },
            fontSize: {
                'h1': ['2rem', { lineHeight: '2.5rem', fontWeight: '600' }],
                'h2': ['1.5rem', { lineHeight: '2rem', fontWeight: '600' }],
                'h3': ['1.25rem', { lineHeight: '1.75rem', fontWeight: '600' }],
                'h4': ['1.125rem', { lineHeight: '1.5rem', fontWeight: '500' }],
                'body': ['1rem', { lineHeight: '1.5rem', fontWeight: '400' }],
                'small': ['0.875rem', { lineHeight: '1.25rem', fontWeight: '400' }],
                'tiny': ['0.75rem', { lineHeight: '1rem', fontWeight: '400' }],
            },
            spacing: {
                'xs': '0.25rem',   // 4px
                'sm': '0.5rem',    // 8px
                'md': '1rem',      // 16px
                'lg': '1.5rem',    // 24px
                'xl': '2rem',      // 32px
                '2xl': '3rem',     // 48px
                '3xl': '4rem',     // 64px
            },
            borderRadius: {
                'sm': '0.25rem',
                'md': '0.375rem',
                'lg': '0.5rem',
                'xl': '0.75rem',
                '2xl': '1rem',
            },
            boxShadow: {
                'subtle': '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
                'card': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                'elevated': '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
            },
        },
    },
    plugins: [],
};

export default config;
