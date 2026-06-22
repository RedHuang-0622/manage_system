import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { usePermission } from '../usePermission';
import { useAuthStore } from '../../store/auth';

function makeToken(payload: Record<string, unknown>): string {
  const body = btoa(JSON.stringify(payload));
  const bodyUrlSafe = body.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `eyJhbGciOiJIUzI1NiJ9.${bodyUrlSafe}.fake-sig`;
}

function loginAs(role_name: string) {
  const token = makeToken({
    user_id: 1, username: 'test', role_id: 1, role_name,
    exp: Math.floor(Date.now() / 1000) + 3600,
  });
  useAuthStore.getState().setToken(token);
}

describe('usePermission', () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, user: null });
    localStorage.clear();
  });

  it('returns roleName undefined when not logged in', () => {
    const { result } = renderHook(() => usePermission());
    expect(result.current.roleName).toBeUndefined();
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isSuperAdmin).toBe(false);
  });

  describe('hasRole', () => {
    it('returns true when role matches', () => {
      loginAs('member');
      const { result } = renderHook(() => usePermission());
      expect(result.current.hasRole(['member'])).toBe(true);
      expect(result.current.hasRole(['member', 'viewer'])).toBe(true);
    });

    it('returns false when role does not match', () => {
      loginAs('member');
      const { result } = renderHook(() => usePermission());
      expect(result.current.hasRole(['super_admin'])).toBe(false);
      expect(result.current.hasRole(['super_admin', 'lab_admin'])).toBe(false);
    });
  });

  describe('role-based checks', () => {
    it('super_admin has full permissions', () => {
      loginAs('super_admin');
      const { result } = renderHook(() => usePermission());
      expect(result.current.isAdmin).toBe(true);
      expect(result.current.isSuperAdmin).toBe(true);
      expect(result.current.isEquipManager).toBe(true);
      expect(result.current.canManageUsers).toBe(true);
    });

    it('lab_admin has admin + equip permissions', () => {
      loginAs('lab_admin');
      const { result } = renderHook(() => usePermission());
      expect(result.current.isAdmin).toBe(true);
      expect(result.current.isSuperAdmin).toBe(false);
      expect(result.current.isEquipManager).toBe(true);
      expect(result.current.canManageUsers).toBe(true);
    });

    it('equipment_manager has equip but not admin', () => {
      loginAs('equipment_manager');
      const { result } = renderHook(() => usePermission());
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isSuperAdmin).toBe(false);
      expect(result.current.isEquipManager).toBe(true);
      expect(result.current.canManageUsers).toBe(false);
    });

    it('member has no special permissions', () => {
      loginAs('member');
      const { result } = renderHook(() => usePermission());
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isSuperAdmin).toBe(false);
      expect(result.current.isEquipManager).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
    });

    it('viewer has no special permissions', () => {
      loginAs('viewer');
      const { result } = renderHook(() => usePermission());
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isSuperAdmin).toBe(false);
      expect(result.current.isEquipManager).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
    });
  });
});
