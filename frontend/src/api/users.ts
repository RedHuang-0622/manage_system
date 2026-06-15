import client from './client';
import type { ApiResponse, PageData, User, ListUserReq, CreateUserReq, UpdateUserReq, ChangePasswordReq } from './types';

/** GET /users */
export async function listUsers(params: ListUserReq): Promise<ApiResponse<PageData<User>>> {
  const { data } = await client.get('/users', { params });
  return data;
}

/** GET /users/:id */
export async function getUser(id: number): Promise<ApiResponse<User>> {
  const { data } = await client.get(`/users/${id}`);
  return data;
}

/** POST /users */
export async function createUser(req: CreateUserReq): Promise<ApiResponse<User>> {
  const { data } = await client.post('/users', req);
  return data;
}

/** PUT /users/:id */
export async function updateUser(id: number, req: UpdateUserReq): Promise<ApiResponse<null>> {
  const { data } = await client.put(`/users/${id}`, req);
  return data;
}

/** POST /users/:id/disable */
export async function disableUser(id: number): Promise<ApiResponse<null>> {
  const { data } = await client.post(`/users/${id}/disable`);
  return data;
}

/** PUT /users/:id/password */
export async function changePassword(id: number, req: ChangePasswordReq): Promise<ApiResponse<null>> {
  const { data } = await client.put(`/users/${id}/password`, req);
  return data;
}
