/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        // Flag-inspired palette. Contrast (WCAG) against white noted; gold is never used on white.
        flag: "#EE1C25", // brand mark and large type only (4.35:1)
        crimson: "#C8102E", // buttons, links, selected chips (5.88:1)
        maroon: "#A50E1F", // hero and card-back surfaces (7.83:1 with white)
        gold: "#FFDE00", // accents on red or ink only
        ink: "#1C1917",
        canvas: "#F5F3F0",
        line: "#E7E2DC",
        muted: "#6B6560", // secondary text (5.6:1 on white)
        tone1: "#C8102E",
        tone2: "#B45309",
        tone3: "#15803D",
        tone4: "#1D4ED8",
        tone5: "#6B7280",
      },
      fontFamily: {
        sans: ["Outfit", "Noto Sans SC", "ui-sans-serif", "system-ui", "sans-serif"],
        hanzi: ["Noto Sans SC", "PingFang SC", "Hiragino Sans GB", "sans-serif"],
      },
      boxShadow: {
        card: "0 1px 0 #E7E2DC, 0 6px 16px rgba(28,25,23,.06)",
        lift: "0 2px 0 #E7E2DC, 0 18px 40px rgba(28,25,23,.10)",
        float: "0 2px 0 #E7E2DC, 0 24px 50px rgba(28,25,23,.14)",
      },
      backgroundImage: {
        hero: "radial-gradient(120% 140% at 100% 0%, #C8102E 0%, #A50E1F 55%, #7d0a17 100%)",
      },
    },
  },
  plugins: [],
};
