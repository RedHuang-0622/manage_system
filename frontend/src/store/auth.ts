import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { UserInfo } from '../api/types';
import { decodeToken } from '../api/auth';

interface AuthState {
  token: string | null;
  user: UserInfo | null;

  // Actions
  setToken: (token: string) => void;
  setUser: (user: UserInfo) => void;
  logout: () => void;

  // Computed
  isLoggedIn: () => boolean;
  isAdmin: () => boolean;
  isSuperAdmin: () => boolean;
}

/** Decode JWT payload (handles base64url → base64 conversion for atob) */
function parseJWTPayload(token: string): Record<string, unknown> | null {
  try {
    const base64url = token.split('.')[1];
    const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
    return JSON.parse(atob(base64));
  } catch {
    return null;
  }
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,

      setToken: (token: string) => {
        const user = decodeToken(token);
        set({ token, user });
      },

      setUser: (user: UserInfo) => {
        set({ user });
      },

      logout: () => {
        set({ token: null, user: null });
      },

      isLoggedIn: () => {
        const { token } = get();
        if (!token) return false;
        const payload = parseJWTPayload(token);
        if (!payload) return false;
        const now = Math.floor(Date.now() / 1000);
        return typeof payload.exp === 'number' && payload.exp > now;
      },

      isAdmin: () => {
        const { user } = get();
        return user?.role_name === 'super_admin' || user?.role_name === 'lab_admin';
      },

      isSuperAdmin: () => {
        const { user } = get();
        return user?.role_name === 'super_admin';
      },
    }),
    {
      name: 'lab-auth-storage',
      partialize: (state) => ({ token: state.token }),
    },
  ),
);
