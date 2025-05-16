import { defineConfig } from 'vitest/config';
import { fileURLToPath } from 'url';

// Helper to create a URL for the current directory
const __dirname = fileURLToPath(new URL('.', import.meta.url));

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom', // Use jsdom for browser-like environment
    setupFiles: './test/setup.js',
    testTimeout: 10000, // 10 seconds
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        '**/node_modules/**',
        '**/test/**',
        '**/*.config.js',
      ],
      all: true,
      include: ['**/*.{js,jsx,ts,tsx}'],
    },
    // Watch mode is on by default for 'npm test', off for 'npm run test:coverage'
    watch: !process.env.CI,
    include: ['**/*.test.js'],
    exclude: ['**/node_modules/**', '**/dist/**'],
  },
  resolve: {
    alias: {
      // Add any path aliases here if needed
      '@': __dirname,
    },
  },
});
