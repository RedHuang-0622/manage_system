import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig(({ mode }) => ({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  define:
    mode !== 'production'
      ? { 'import.meta.env.VITE_API_BASE_URL': JSON.stringify('http://localhost:8080/api/v1') }
      : {},
  server: {
    port: 5173,
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
}));
