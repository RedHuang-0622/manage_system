import client from './client';
import type { ApiResponse, PageData, Equipment, ListEquipReq, CreateEquipReq, UpdateEquipReq } from './types';

/** GET /equipments */
export async function listEquipments(params: ListEquipReq): Promise<ApiResponse<PageData<Equipment>>> {
  const { data } = await client.get('/equipments', { params });
  return data;
}

/** GET /equipments/:id */
export async function getEquipment(id: number): Promise<ApiResponse<Equipment>> {
  const { data } = await client.get(`/equipments/${id}`);
  return data;
}

/** POST /equipments */
export async function createEquipment(req: CreateEquipReq): Promise<ApiResponse<Equipment>> {
  const { data } = await client.post('/equipments', req);
  return data;
}

/** PUT /equipments/:id */
export async function updateEquipment(id: number, req: UpdateEquipReq): Promise<ApiResponse<null>> {
  const { data } = await client.put(`/equipments/${id}`, req);
  return data;
}

/** DELETE /equipments/:id */
export async function disableEquipment(id: number): Promise<ApiResponse<null>> {
  const { data } = await client.delete(`/equipments/${id}`);
  return data;
}
