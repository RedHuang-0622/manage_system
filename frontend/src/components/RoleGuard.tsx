import { Navigate } from 'react-router-dom';
import { usePermission } from '../hooks/usePermission';

interface RoleGuardProps {
  roles: Array<'super_admin' | 'lab_admin' | 'equipment_manager' | 'member' | 'viewer'>;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/** Route-level role guard: redirects to "/" if role not allowed */
export default function RoleGuard({ roles, children, fallback }: RoleGuardProps) {
  const { hasRole } = usePermission();

  if (!hasRole(roles)) {
    if (fallback) return <>{fallback}</>;
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}
