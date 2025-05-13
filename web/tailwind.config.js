module.exports = {
  safelist: [
    // Critical classes used in templates - explicitly included
    'bg-white', 'rounded', 'shadow', 'p-4', 'm-4', 'text-2xl', 'font-bold', 'mb-2', 'flex-1', 'flex-col', 'min-h-screen',
    // Patterns for other classes
    {
    pattern: /^bg-/,
    variants: ['hover', 'focus']
  }, {
    pattern: /^text-/,
    variants: ['hover', 'focus']
  }, {
    pattern: /^m-/
  }, {
    pattern: /^p-/
  }, {
    pattern: /^flex/
  }, {
    pattern: /^rounded/
  }, {
    pattern: /^shadow/
  }, {
    pattern: /^font-/
  }, {
    pattern: /^border/
  }],
  content: [
    "./templates/**/*.{html,hbs,js,ts}",
    "./static/js/**/*.{js,ts}"
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
