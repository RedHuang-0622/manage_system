import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '../store/auth';
import { ErrCode } from './types';

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

// ── Request interceptor: inject Bearer token ──

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().token;
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
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

  const resp = await axios.post('/api/v1/auth/refresh', { token: currentToken }, {
    headers: { Authorization: `Bearer ${currentToken}` },
  });
  if (resp.data.code !== 0) {
    throw new Error(resp.data.msg || 'Token refresh failed');
  }
  return resp.data.data.token;
}

// ── Response interceptor: unwrap + handle 401 refresh ──

client.interceptors.response.use(
  (response) => {
    // Unwrap: return {code, msg, data} directly
    return response;
  },
  async (error: AxiosError<{ code: number; msg: string }>) => {
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
