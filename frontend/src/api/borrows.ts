import client from './client';
import type { ApiResponse, PageData, BorrowRecord, ListBorrowReq, ApplyBorrowReq, ApproveBorrowReq } from './types';

/** POST /borrows/apply */
export async function applyBorrow(req: ApplyBorrowReq): Promise<ApiResponse<BorrowRecord>> {
  const { data } = await client.post('/borrows/apply', req);
  return data;
}

/** POST /borrows/:id/approve */
export async function approveBorrow(id: number, req: ApproveBorrowReq): Promise<ApiResponse<BorrowRecord>> {
  const { data } = await client.post(`/borrows/${id}/approve`, req);
  return data;
}

/** POST /borrows/:id/return */
export async function returnBorrow(id: number): Promise<ApiResponse<BorrowRecord>> {
  const { data } = await client.post(`/borrows/${id}/return`);
  return data;
}

/** POST /borrows/:id/cancel */
export async function cancelBorrow(id: number): Promise<ApiResponse<null>> {
  const { data } = await client.post(`/borrows/${id}/cancel`);
  return data;
}

/** GET /borrows/my */
export async function listMyRecords(params: ListBorrowReq): Promise<ApiResponse<PageData<BorrowRecord>>> {
  const { data } = await client.get('/borrows/my', { params });
  return data;
}

/** GET /borrows/pending */
export async function listPendingRecords(params: ListBorrowReq): Promise<ApiResponse<PageData<BorrowRecord>>> {
  const { data } = await client.get('/borrows/pending', { params });
  return data;
}

/** GET /borrows — all records (admin) */
export async function listAllRecords(params: ListBorrowReq): Promise<ApiResponse<PageData<BorrowRecord>>> {
  const { data } = await client.get('/borrows', { params });
  return data;
}
