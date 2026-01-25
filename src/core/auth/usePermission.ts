import { useAuthStore } from '@/modules/auth/store';
import type { UserRole } from '@/types';

export type PermissionAction = 'create' | 'edit' | 'delete' | 'admin_tools' | 'view';

/**
 * Permission matrix: Maps roles to their allowed actions
 * Easy to extend for new roles or permissions
 */
const PERMISSION_MATRIX: Record<UserRole, Set<PermissionAction>> = {
  RoleAdmin: new Set(['create', 'edit', 'delete', 'admin_tools', 'view']),
  RoleUser: new Set(['view']),
};

export function usePermission() {
  const auth = useAuthStore();

  function can(action: PermissionAction): boolean {
    if (!auth.user) return false;
    
    const allowedActions = PERMISSION_MATRIX[auth.user.role];
    return allowedActions?.has(action) ?? false;
  }

  function canAny(...actions: PermissionAction[]): boolean {
    return actions.some((action) => can(action));
  }

  function canAll(...actions: PermissionAction[]): boolean {
    return actions.every((action) => can(action));
  }

  return { can, canAny, canAll };
}
