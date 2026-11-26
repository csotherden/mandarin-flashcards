/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paper: "#f4efe4",
        ink: "#1c1915",
        cinnabar: "#9f2d20",
        moss: "#1f6b45",
        amberglow: "#a16207",
      },
      fontFamily: {
        sans: ["Outfit", "ui-sans-serif", "system-ui", "sans-serif"],
        hanzi: ["Noto Sans SC", "PingFang SC", "Hiragino Sans GB", "sans-serif"],
      },
    },
  },
  plugins: [],
};
