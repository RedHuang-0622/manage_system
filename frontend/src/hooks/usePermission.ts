import { useAuthStore } from '../store/auth';

type RoleName = 'super_admin' | 'lab_admin' | 'member';

export function usePermission() {
  const user = useAuthStore((s) => s.user);
  const roleName = user?.role_name as RoleName | undefined;

  const hasRole = (roles: RoleName[]): boolean => {
    if (!roleName) return false;
    return roles.includes(roleName);
  };

  const isAdmin = roleName === 'super_admin' || roleName === 'lab_admin';
  const isSuperAdmin = roleName === 'super_admin';

  return { roleName, hasRole, isAdmin, isSuperAdmin };
}
