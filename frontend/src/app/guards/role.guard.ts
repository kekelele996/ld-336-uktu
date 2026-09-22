import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthStore } from '../../stores/auth.store';

// 角色守卫：仅允许指定角色访问。
export function roleGuard(...roles: string[]): CanActivateFn {
  return () => {
    const auth = inject(AuthStore);
    const router = inject(Router);
    if (auth.hasRole(...roles)) {
      return true;
    }
    return router.createUrlTree(['/dashboard']);
  };
}
