import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        timeout: 10000,       // 10s proxy timeout (prevents Vite hang)
        proxyTimeout: 12000,  // 12s upstream timeout
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
