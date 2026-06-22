import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '../store/auth';
import { ErrCode } from './types';

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 8000,
  headers: { 'Content-Type': 'application/json' },
});

// ── Request interceptor: inject Bearer token ──

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().token;
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  if (import.meta.env.DEV) {
    const ts = new Date().toISOString().slice(11, 23);
    console.log(`[axios ${ts}] ${config.method?.toUpperCase()} ${config.baseURL}${config.url}`);
  }
  return config;
});

// ── Token refresh lock (prevents concurrent refresh race) ──

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) reject(error);
    else resolve(token!);
  });
  failedQueue = [];
}

async function refreshTokenRequest(): Promise<string> {
  const currentToken = useAuthStore.getState().token;
  if (!currentToken) throw new Error('No token');

  // Use the configured client (has timeout) instead of raw axios.
  // Raw axios has no timeout — if the proxy hangs, this promise
  // never settles, isRefreshing stays true, and every other 401
  // request queues up forever → app freezes.
  const resp = await client.post('/auth/refresh', { token: currentToken });
  if (resp.data.code !== 0) {
    throw new Error(resp.data.msg || 'Token refresh failed');
  }
  return resp.data.data.token;
}

// ── Proactive refresh: silently refresh when token has < 10 min left ──
// This prevents the "expired token → refresh fails → kicked to login" cycle
// that occurs when the reactive 401 interceptor tries to refresh an already-expired
// token (refresh endpoint itself requires a valid token).
const PROACTIVE_REFRESH_THRESHOLD_SEC = 600; // 10 minutes

function getTokenRemainingSec(): number {
  const token = useAuthStore.getState().token;
  if (!token) return 0;
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')));
    return (payload.exp || 0) - Math.floor(Date.now() / 1000);
  } catch {
    return 0;
  }
}

async function maybeProactiveRefresh() {
  if (isRefreshing) return;
  if (getTokenRemainingSec() > PROACTIVE_REFRESH_THRESHOLD_SEC) return;
  try {
    const newToken = await refreshTokenRequest();
    useAuthStore.getState().setToken(newToken);
    console.log('[auth] Token proactively refreshed');
  } catch {
    // Silently fail — the reactive 401 interceptor will handle actual expiry
  }
}

// ── Response interceptor: unwrap + handle 401 refresh ──

client.interceptors.response.use(
  (response) => {
    if (import.meta.env.DEV) {
      const ts = new Date().toISOString().slice(11, 23);
      console.log(`[axios ${ts}] ← ${response.status} ${response.config.url}`);
    }
    // Proactive refresh: if token has <10 min left, refresh silently.
    maybeProactiveRefresh();
    return response;
  },
  async (error: AxiosError<{ code: number; msg: string }>) => {
    if (import.meta.env.DEV) {
      const ts = new Date().toISOString().slice(11, 23);
      const detail = error.response
        ? `${error.response.status} ${error.response.config.url}`
        : `NETWORK_ERROR ${error.code} ${error.config?.url}: ${error.message}`;
      console.log(`[axios ${ts}] ⚠ ${detail}`);
    }
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    const data = error.response?.data;

    // Only handle 401 with code 2004 (token expired)
    if (error.response?.status === 401 && data?.code === ErrCode.ErrTokenInvalid && !originalRequest._retry) {
      if (isRefreshing) {
        // Queue this request until refresh completes
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${token}`;
              }
              resolve(client(originalRequest));
            },
            reject,
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const newToken = await refreshTokenRequest();
        useAuthStore.getState().setToken(newToken);
        processQueue(null, newToken);
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
        }
        return client(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Other errors: pass through
    return Promise.reject(error);
  },
);

export default client;
