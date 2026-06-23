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
    host: '127.0.0.1',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        // proxyTimeout: max time to wait for backend to send a response.
        // If the backend hangs, we fail in 7s instead of waiting for
        // the browser's 8s axios timeout — and we return a 504, so the
        // error interceptor gets an HTTP error rather than ECONNABORTED.
        proxyTimeout: 7000,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            console.error('[vite:proxy] backend error:', err.message);
            // If response hasn't been sent yet, send 502
            if (res && !res.headersSent) {
              res.writeHead(502, { 'Content-Type': 'application/json' });
              res.end(JSON.stringify({ code: 9502, msg: 'Backend unreachable' }));
            }
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
