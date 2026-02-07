/** @type {import('tailwindcss').Config} */
export default {
    content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
    theme: {
        extend: {
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
                mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
            },
        },
    },
    plugins: [
        require('daisyui'),
    ],
    daisyui: {
        themes: [
            {
                pouch: {
                    "primary": "#6366f1",          // Indigo
                    "primary-content": "#ffffff",
                    "secondary": "#8b5cf6",        // Violet
                    "secondary-content": "#ffffff",
                    "accent": "#06b6d4",           // Cyan
                    "accent-content": "#ffffff",
                    "neutral": "#1e1e2e",          // Dark slate
                    "neutral-content": "#cdd6f4",
                    "base-100": "#1e1e2e",         // Background
                    "base-200": "#181825",         // Darker
                    "base-300": "#11111b",         // Darkest
                    "base-content": "#cdd6f4",     // Light text
                    "info": "#89b4fa",
                    "info-content": "#1e1e2e",
                    "success": "#a6e3a1",
                    "success-content": "#1e1e2e",
                    "warning": "#f9e2af",
                    "warning-content": "#1e1e2e",
                    "error": "#f38ba8",
                    "error-content": "#1e1e2e",
                },
            },
            "light",
            "dark",
        ],
    }
}
