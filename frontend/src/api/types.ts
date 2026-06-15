// ── Unified Response (maps to backend pkg/response/response.go) ──

export interface ApiResponse<T = unknown> {
  code: number;
  msg: string;
  data: T;
}

export interface PageData<T> {
  total: number;
  page: number;
  page_size: number;
  list: T[];
}

// ── Pagination ──

export interface PageReq {
  page?: number;
  page_size?: number;
}

// ── Error Codes (maps to backend pkg/errcode/errcode.go) ──

export const ErrCode = {
  Success: 0,
  ErrInvalidParam: 1001,
  ErrNotFound: 1002,
  ErrConflict: 1003,
  ErrInternal: 1004,
  ErrAuthFailed: 2001,
  ErrAccountDisabled: 2002,
  ErrTokenMissing: 2003,
  ErrTokenInvalid: 2004,
  ErrPermissionDenied: 2005,
  ErrRateLimited: 2006,
  ErrUserExists: 3001,
  ErrUserNotFound: 3002,
  ErrEquipmentNotFound: 3003,
  ErrStockInsufficient: 3004,
  ErrDuplicateBorrow: 3005,
  ErrBorrowApproveFailed: 3006,
  ErrBorrowStatusInvalid: 3007,
  ErrInternalServer: 5000,
} as const;

// ── JWT / Auth ──

export interface LoginReq {
  username: string;
  password: string;
}

export interface LoginResp {
  token: string;
  expires_in: number;
}

export interface UserInfo {
  user_id: number;
  username: string;
  role_id: number;
  role_name: 'super_admin' | 'lab_admin' | 'member';
}

// ── Role ──

export interface RoleInfo {
  id: number;
  role_name: string;
  description: string;
}

// ── User ──

export interface User {
  id: number;
  username: string;
  real_name: string;
  email: string;
  phone: string;
  role_id: number;
  role?: RoleInfo;
  status: number; // 1=enabled 0=disabled
  created_at: string;
  updated_at: string;
}

export interface CreateUserReq {
  username: string;
  password: string;
  real_name: string;
  email?: string;
  phone?: string;
  role_id: number;
}

export interface UpdateUserReq {
  real_name?: string;
  email?: string;
  phone?: string;
  role_id?: number;
  status?: number;
}

export interface ChangePasswordReq {
  old_password: string;
  new_password: string;
}

export interface ListUserReq extends PageReq {
  keyword?: string;
  status?: number;
  role_id?: number;
}

// ── Equipment ──

export interface Equipment {
  id: number;
  name: string;
  model: string;
  category: string;
  total_stock: number;
  available_stock: number;
  location: string;
  description: string;
  status: number; // 1=online 0=offline
  created_at: string;
  updated_at: string;
}

export interface CreateEquipReq {
  name: string;
  model?: string;
  category?: string;
  total_stock: number;
  location?: string;
  description?: string;
}

export interface UpdateEquipReq {
  name?: string;
  model?: string;
  category?: string;
  total_stock?: number;
  location?: string;
  description?: string;
  status?: number;
}

export interface ListEquipReq extends PageReq {
  keyword?: string;
  category?: string;
  status?: number;
  only_available?: number;
}

// ── Borrow ──

export type BorrowStatus = '申请中' | '已借出' | '已归还' | '被拒绝';

export interface UserBrief {
  id: number;
  username: string;
  real_name: string;
}

export interface EquipmentBrief {
  id: number;
  name: string;
  model: string;
}

export interface BorrowRecord {
  id: number;
  user_id: number;
  equipment_id: number;
  quantity: number;
  status: BorrowStatus;
  apply_note: string;
  approve_note: string;
  approver_id: number;
  apply_at: string;
  approve_at: string | null;
  return_at: string | null;
  user?: UserBrief;
  equipment?: EquipmentBrief;
  approver?: UserBrief;
}

export interface ApplyBorrowReq {
  equipment_id: number;
  quantity: number;
  apply_note?: string;
}

export interface ApproveBorrowReq {
  approve: boolean;
  approve_note?: string;
}

export interface ListBorrowReq extends PageReq {
  status?: string;
  user_id?: number;
  equipment_id?: number;
}
