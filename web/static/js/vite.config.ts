import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  build: {
    outDir: '../',
    emptyOutDir: false,
    sourcemap: true,
    rollupOptions: {
      input: {
        'main': resolve(__dirname, 'src/main.ts'),
        'toast': resolve(__dirname, 'src/toast.ts'),
        'coverage': resolve(__dirname, 'src/coverage.ts'),
        'settings': resolve(__dirname, 'src/settings.ts'),
        'websocket': resolve(__dirname, 'src/websocket.ts')
      },
      output: {
        entryFileNames: '[name].js',
        chunkFileNames: '[name]-[hash].js',
        assetFileNames: '[name].[ext]'
      }
    }
  },
  server: {
    port: 3000,
    open: true,
  },
});
