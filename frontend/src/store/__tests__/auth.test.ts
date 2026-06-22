import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore } from '../auth';

// Helper to create a valid-looking JWT token
function makeToken(payload: Record<string, unknown>): string {
  const body = btoa(JSON.stringify(payload));
  const bodyUrlSafe = body.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `eyJhbGciOiJIUzI1NiJ9.${bodyUrlSafe}.fake-sig`;
}

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({ token: null, user: null });
    // Clear localStorage to avoid persist interference
    localStorage.clear();
  });

  describe('setToken', () => {
    it('sets token and decodes user info from JWT', () => {
      const token = makeToken({
        user_id: 1,
        username: 'admin',
        role_id: 1,
        role_name: 'super_admin',
      });

      useAuthStore.getState().setToken(token);

      const state = useAuthStore.getState();
      expect(state.token).toBe(token);
      expect(state.user).toEqual({
        user_id: 1,
        username: 'admin',
        role_id: 1,
        role_name: 'super_admin',
      });
    });

    it('sets user to null for an invalid token', () => {
      useAuthStore.getState().setToken('invalid.jwt.token');
      const state = useAuthStore.getState();
      expect(state.token).toBe('invalid.jwt.token');
      expect(state.user).toBeNull();
    });
  });

  describe('logout', () => {
    it('clears token and user', () => {
      const token = makeToken({ user_id: 1, username: 'test', role_id: 1, role_name: 'admin' });
      useAuthStore.getState().setToken(token);

      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
    });
  });

  describe('isLoggedIn', () => {
    it('returns false when token is null', () => {
      expect(useAuthStore.getState().isLoggedIn()).toBe(false);
    });

    it('returns false for an expired token', () => {
      const expiredToken = makeToken({
        user_id: 1,
        username: 'admin',
        role_id: 1,
        role_name: 'super_admin',
        exp: Math.floor(Date.now() / 1000) - 3600, // 1 hour ago
      });

      useAuthStore.getState().setToken(expiredToken);
      expect(useAuthStore.getState().isLoggedIn()).toBe(false);
    });

    it('returns true for a valid non-expired token', () => {
      const validToken = makeToken({
        user_id: 1,
        username: 'admin',
        role_id: 1,
        role_name: 'super_admin',
        exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
      });

      useAuthStore.getState().setToken(validToken);
      expect(useAuthStore.getState().isLoggedIn()).toBe(true);
    });

    it('returns false when exp is not a number', () => {
      const token = makeToken({
        user_id: 1,
        username: 'admin',
        role_id: 1,
        role_name: 'super_admin',
        exp: 'not-a-number',
      });

      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isLoggedIn()).toBe(false);
    });
  });

  describe('isAdmin', () => {
    it('returns true for super_admin role', () => {
      const token = makeToken({
        user_id: 1, username: 'admin', role_id: 1, role_name: 'super_admin',
        exp: Math.floor(Date.now() / 1000) + 3600,
      });
      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isAdmin()).toBe(true);
    });

    it('returns true for lab_admin role', () => {
      const token = makeToken({
        user_id: 2, username: 'lab', role_id: 2, role_name: 'lab_admin',
        exp: Math.floor(Date.now() / 1000) + 3600,
      });
      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isAdmin()).toBe(true);
    });

    it('returns false for member role', () => {
      const token = makeToken({
        user_id: 3, username: 'user', role_id: 4, role_name: 'member',
        exp: Math.floor(Date.now() / 1000) + 3600,
      });
      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isAdmin()).toBe(false);
    });

    it('returns false when user is null', () => {
      expect(useAuthStore.getState().isAdmin()).toBe(false);
    });
  });

  describe('isSuperAdmin', () => {
    it('returns true only for super_admin', () => {
      const token = makeToken({
        user_id: 1, username: 'admin', role_id: 1, role_name: 'super_admin',
        exp: Math.floor(Date.now() / 1000) + 3600,
      });
      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isSuperAdmin()).toBe(true);
    });

    it('returns false for lab_admin', () => {
      const token = makeToken({
        user_id: 2, username: 'lab', role_id: 2, role_name: 'lab_admin',
        exp: Math.floor(Date.now() / 1000) + 3600,
      });
      useAuthStore.getState().setToken(token);
      expect(useAuthStore.getState().isSuperAdmin()).toBe(false);
    });
  });
});
