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
        // Check if token is expired (basic client-side check)
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const now = Math.floor(Date.now() / 1000);
          return payload.exp > now;
        } catch {
          return false;
        }
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
