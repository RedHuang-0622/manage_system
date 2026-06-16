import client from './client';
import type { ApiResponse, LoginReq, LoginResp, RoleInfo, UserInfo } from './types';

/** POST /auth/login */
export async function login(req: LoginReq): Promise<ApiResponse<LoginResp>> {
  const { data } = await client.post('/auth/login', req);
  return data;
}

/** POST /auth/logout — pass token explicitly to avoid stale Zustand state */
export async function logout(token?: string): Promise<ApiResponse<null>> {
  const headers: Record<string, string> = {};
  if (token) headers.Authorization = `Bearer ${token}`;
  const { data } = await client.post('/auth/logout', null, { headers });
  return data;
}

/** POST /auth/refresh */
export async function refreshToken(currentToken: string): Promise<ApiResponse<LoginResp>> {
  const { data } = await client.post('/auth/refresh', { token: currentToken });
  return data;
}

/** GET /roles */
export async function listRoles(): Promise<ApiResponse<RoleInfo[]>> {
  const { data } = await client.get('/roles');
  return data;
}

/** Convert base64url to standard base64 (JWT uses base64url, atob needs base64) */
function base64urlDecode(str: string): string {
  // Replace URL-safe chars and add padding
  const base64 = str.replace(/-/g, '+').replace(/_/g, '/');
  return atob(base64);
}

/** GET /auth/me (decoded from JWT) — utility for getting current user info */
export function decodeToken(token: string): UserInfo | null {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(base64urlDecode(payload));
    return {
      user_id: decoded.user_id,
      username: decoded.username,
      role_id: decoded.role_id,
      role_name: decoded.role_name,
    };
  } catch {
    return null;
  }
}
