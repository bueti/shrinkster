/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/html/*.html", "./ui/html/pages/*.html", "./ui/html/partials/*.html"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Iosevka Aile Iaso", "sans-serif"],
        mono: ["Iosevka Curly Iaso", "monospace"],
        serif: ["Iosevka Etoile Iaso", "serif"],
      },
    },
  },
  plugins: [],
};
