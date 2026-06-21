import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import http from 'node:http';

// ── HTTP Agent: connection pool toward backend ──
// Without keepAlive, Vite creates a new TCP connection per proxy request.
// Paired with a Go backend that lacks IdleTimeout (CLOSE_WAIT leak),
// this exhausts the OS ephemeral port range within minutes, causing
// all proxy requests to hang until timeout.
//
// keepAlive + maxSockets bounds give Vite a managed connection pool:
//   - keepAlive: true → reuse TCP connections (fewer handshakes)
//   - maxSockets: 10  → cap concurrent connections, no explosion
//   - maxFreeSockets: 2 → only 2 idle connections kept warm
//   - timeout: 12000  → kill idle sockets before Go's IdleTimeout (30s)
const agent = new http.Agent({
  keepAlive: true,
  keepAliveMsecs: 30000,
  maxSockets: 10,
  maxFreeSockets: 2,
  timeout: 12000,
});

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
        agent,
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
