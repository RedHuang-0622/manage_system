import { useAuthStore } from '../store/auth';

export type RoleName = 'super_admin' | 'lab_admin' | 'equipment_manager' | 'member' | 'viewer';

export function usePermission() {
  const user = useAuthStore((s) => s.user);
  const roleName = user?.role_name as RoleName | undefined;

  const hasRole = (roles: RoleName[]): boolean => {
    if (!roleName) return false;
    return roles.includes(roleName);
  };

  const isAdmin = roleName === 'super_admin' || roleName === 'lab_admin';
  const isSuperAdmin = roleName === 'super_admin';
  const isEquipManager = roleName === 'super_admin' || roleName === 'lab_admin' || roleName === 'equipment_manager';
  const canManageUsers = roleName === 'super_admin' || roleName === 'lab_admin';

  return { roleName, hasRole, isAdmin, isSuperAdmin, isEquipManager, canManageUsers };
}
