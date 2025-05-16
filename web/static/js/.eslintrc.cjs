module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
    'vitest-globals/env': true,
  },
  extends: [
    'eslint:recommended',
    'plugin:vitest/recommended',
    'prettier', // Must be last in the list
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  rules: {
    // Add custom rules here
    'no-console': 'warn',
    'no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
  },
  overrides: [
    {
      files: ['**/*.test.js'],
      env: {
        'vitest-globals/env': true,
      },
    },
  ],
};
