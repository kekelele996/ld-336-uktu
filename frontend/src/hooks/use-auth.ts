import { authStore } from '../stores/auth.store';

// 认证工具：返回登录状态与角色判断（供路由守卫与按钮显隐复用）。
export function useAuth() {
  return {
    isLoggedIn: () => authStore.isLoggedIn(),
    user: () => authStore.user(),
    token: () => authStore.token(),
    hasRole: (...roles: string[]) => {
      const u = authStore.user();
      return !!u && roles.includes(u.role);
    },
  };
}
