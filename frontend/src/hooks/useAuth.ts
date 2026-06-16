import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { message } from 'antd';
import { useAuthStore } from '../store/auth';
import * as authApi from '../api/auth';
import { ErrCode } from '../api/types';

export function useAuth() {
  const { setToken, logout, isLoggedIn, isAdmin, isSuperAdmin, token, user } = useAuthStore();
  const navigate = useNavigate();

  const login = useCallback(
    async (username: string, password: string) => {
      const resp = await authApi.login({ username, password });
      if (resp.code !== 0) {
        const errMsg =
          resp.code === ErrCode.ErrAuthFailed
            ? '用户名或密码错误'
            : resp.code === ErrCode.ErrAccountDisabled
              ? '账号已被禁用'
              : resp.msg || '登录失败';
        message.error(errMsg);
        return false;
      }
      setToken(resp.data.token);
      message.success('登录成功');
      navigate('/', { replace: true });
      return true;
    },
    [setToken, navigate],
  );

  const doLogout = useCallback(() => {
    // Capture token before clearing state (needed for API call)
    const currentToken = useAuthStore.getState().token;
    // Clear local state immediately for instant UI response
    logout();
    navigate('/login', { replace: true });
    // Send token to backend for blacklisting (fire-and-forget, explicit token)
    if (currentToken) {
      authApi.logout(currentToken).catch(() => {});
    }
  }, [logout, navigate]);

  return {
    login,
    logout: doLogout,
    isLoggedIn: isLoggedIn(),
    isAdmin: isAdmin(),
    isSuperAdmin: isSuperAdmin(),
    token,
    user,
  };
}
