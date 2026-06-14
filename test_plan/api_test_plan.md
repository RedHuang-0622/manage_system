# 实验室设备管理系统 — API 接口测试方案

> **Base URL**: `http://127.0.0.1:8080`
> **认证方式**: `Authorization: Bearer <token>` (JWT)
> **种子管理员**: `admin` / `admin123` — 角色 `super_admin`
> **统一响应格式**:
> ```json
> { "code": 0, "msg": "success", "data": ... }
> ```

## 目录

1. [公开接口](#1-公开接口无需认证)
2. [认证管理](#2-认证管理需-token)
3. [角色管理](#3-角色管理需-token)
4. [用户管理](#4-用户管理需-token)
5. [设备管理](#5-设备管理需-token)
6. [借阅工单](#6-借阅工单需-token)
7. [错误码速查](#7-错误码速查)

---

## 1. 公开接口（无需认证）

### 1.1 健康检查

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/health` |
| **Auth** | 无 |

**Response** `200`
```json
{ "status": "ok" }
```

---

### 1.2 登录

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/auth/login` |
| **Content-Type** | `application/json` |
| **Auth** | 无 |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `username` | string | ✅ | 3-32 字符 | 用户名 |
| `password` | string | ✅ | 6-64 字符 | 密码 |

```json
{
    "username": "admin",
    "password": "admin123"
}
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 1781190184
    }
}
```

> **注意**: 后续所有受保护接口必须在 Header 中携带 `Authorization: Bearer {token}`

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 用户名或密码错误 | 400 | 2001 | 用户名或密码错误 |
| 账号已被禁用 | 401 | 2002 | 账号已被禁用 |
| 参数校验失败 | 400 | 1001 | 请求参数错误 |

---

## 2. 认证管理（需 Token）

### 2.1 登出

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/auth/logout` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |
| **Body** | 无（空 body 或 `{}`） |

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "登出成功" }
}
```

> **说明**: 登出后将当前 Token 加入 Redis 黑名单，该 Token 不可再用于 Refresh。

---

### 2.2 刷新 Token

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/auth/refresh` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token（当前有效的 Token） |
| **Body** | 无（空 body 或 `{}`） |

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 1781190238
    }
}
```

> **说明**: 旧 Token 加入黑名单，签发新 Token。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| Token 已失效（黑名单） | 401 | 2004 | Token已失效 |
| Token 无效或过期 | 401 | 2004 | Token无效或已过期 |

---

## 3. 角色管理（需 Token）

### 3.1 角色列表

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/roles` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": [
        { "id": 1, "role_name": "super_admin", "description": "超级管理员（指导老师）" },
        { "id": 2, "role_name": "lab_admin",  "description": "实验室负责人" },
        { "id": 3, "role_name": "member",     "description": "普通成员" }
    ]
}
```

| 角色 | 权限范围 |
|------|----------|
| `super_admin` | `全部接口` |
| `lab_admin` | `用户管理` / `设备管理` / `借阅管理` / `角色(只读)` |
| `member` | `设备(只读)` / `借阅申请` / `我的借阅` / `归还` / `取消` / `改自己密码` |

---

## 4. 用户管理（需 Token）

### 4.1 用户列表（分页）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/users` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Query Parameters**

| 字段 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 10 | 每页条数 |
| `keyword` | string | 否 | - | 搜索关键字（用户名/真实姓名） |
| `status` | int | 否 | - | 状态筛选：1=启用 0=禁用 |
| `role_id` | uint | 否 | - | 角色ID筛选 |

```
GET /api/v1/users?page=1&page_size=5&keyword=admin
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "total": 1,
        "page": 1,
        "page_size": 5,
        "list": [
            {
                "id": 1,
                "username": "admin",
                "real_name": "系统管理员",
                "email": "",
                "phone": "",
                "role_id": 1,
                "status": 1,
                "created_at": "2026-06-11 21:00:14",
                "updated_at": "2026-06-11 21:00:14",
                "role": { "id": 1, "role_name": "super_admin", "description": "超级管理员（指导老师）" }
            }
        ]
    }
}
```

---

### 4.2 用户详情

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/users/:id` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户ID |

```
GET /api/v1/users/1
```

**Response** `200` — 结构同 `4.1 用户列表` 中的 `list` 单项。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 用户不存在 | 404 | 3002 | 用户不存在 |
| 无效ID | 400 | 1001 | 无效的用户ID |
| 越权查看他人 | 403 | 2005 | 权限不足 |

---

### 4.3 创建用户

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/users` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `username` | string | ✅ | 3-32 字符，字母数字 | 用户名（唯一） |
| `password` | string | ✅ | 6-64 字符 | 密码 |
| `real_name` | string | ✅ | 2-32 字符 | 真实姓名 |
| `email` | string | 否 | 合法邮箱 | 邮箱 |
| `phone` | string | 否 | 纯数字，≤16位 | 手机号 |
| `role_id` | uint | ✅ | &gt;0 | 角色ID |

```json
{
    "username": "zhangsan",
    "password": "pass123456",
    "real_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "role_id": 3
}
```

**Response** `201`
```json
{
    "code": 0,
    "msg": "创建成功",
    "data": {
        "id": 2,
        "username": "zhangsan",
        "real_name": "张三",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "role_id": 3,
        "status": 1,
        "created_at": "2026-06-11 21:03:07",
        "updated_at": "2026-06-11 21:03:07"
    }
}
```

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 用户名已存在 | 409 | 3001 | 用户名已存在 |
| 角色不存在 | 400 | 1001 | 角色不存在 |
| 参数校验失败 | 400 | 1001 | 请求参数错误 |

---

### 4.4 更新用户

| 属性 | 值 |
|------|-----|
| **Method** | `PUT` |
| **URL** | `/api/v1/users/:id` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户ID |

**Request Body (JSON)** — 所有字段可选，只更新非空字段

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `real_name` | string | 否 | 2-32 字符 | 真实姓名 |
| `email` | string | 否 | 合法邮箱 | 邮箱 |
| `phone` | string | 否 | 纯数字，≤16位 | 手机号 |
| `role_id` | uint | 否 | &gt;0 | 角色ID |
| `status` | int8 | 否 | 0 或 1 | 启用状态 |

```json
{
    "real_name": "张三疯",
    "email": "zhangsan_new@example.com"
}
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "更新成功" }
}
```

---

### 4.5 禁用用户

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/users/:id/disable` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |
| **Body** | 无（空 body 或 `{}`） |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 用户ID |

```
POST /api/v1/users/2/disable
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "用户已禁用" }
}
```

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 不能禁用自己 | 400 | 1001 | 不能禁用自己 |
| 用户不存在 | 404 | 3002 | 用户不存在 |

---

### 4.6 修改密码

| 属性 | 值 |
|------|-----|
| **Method** | `PUT` |
| **URL** | `/api/v1/users/:id/password` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 目标用户ID |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `old_password` | string | 非管理员必填 | - | 旧密码 |
| `new_password` | string | ✅ | 6-64 字符 | 新密码 |

```json
{
    "old_password": "oldpass123",
    "new_password": "newpass456"
}
```

> **规则**:
> - **管理员** (`super_admin` / `lab_admin`) 修改他人密码：无需 `old_password`
> - **普通成员** 修改自己密码：必须提供 `old_password`
> - **普通成员** 修改他人密码：403 权限不足

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "密码修改成功" }
}
```

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 越权修改他人 | 403 | 2005 | 权限不足 |
| 旧密码为空 | 400 | 1001 | 旧密码不能为空 |
| 旧密码错误 | 400 | 1001 | 旧密码错误 |

---

## 5. 设备管理（需 Token）

### 5.1 设备列表（分页 + Redis 缓存）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/equipments` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Query Parameters**

| 字段 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 12 | 每页条数 |
| `keyword` | string | 否 | - | 搜索关键字（设备名） |
| `category` | string | 否 | - | 分类筛选 |
| `status` | int | 否 | - | 状态：1=上架 0=下架 |
| `only_available` | int | 否 | - | 仅看有库存：1=是 |

```
GET /api/v1/equipments?page=1&page_size=12&category=电子测量&only_available=1
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "total": 1,
        "page": 1,
        "page_size": 12,
        "list": [
            {
                "id": 1,
                "name": "示波器",
                "model": "TDS2014C",
                "category": "电子测量",
                "total_stock": 10,
                "available_stock": 8,
                "location": "实验室A-101",
                "description": "泰克四通道数字示波器",
                "status": 1,
                "created_at": "2026-06-11 21:03:08",
                "updated_at": "2026-06-11 21:03:08"
            }
        ]
    }
}
```

> **缓存策略**: 设备列表缓存 30 秒；创建/更新/下架后自动失效相关缓存。

---

### 5.2 设备详情（Redis 缓存）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/equipments/:id` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 设备ID |

```
GET /api/v1/equipments/1
```

**Response** `200` — 结构同 `5.1 设备列表` 中的 `list` 单项。

> **缓存策略**: 设备详情缓存 60 秒。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 设备不存在 | 404 | 3003 | 设备不存在 |
| 无效ID | 400 | 1001 | 无效的设备ID |

---

### 5.3 创建设备

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/equipments` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `name` | string | ✅ | 1-128 字符 | 设备名称 |
| `model` | string | 否 | ≤64 字符 | 型号 |
| `category` | string | 否 | ≤32 字符 | 分类 |
| `total_stock` | uint | 否 | ≥0 | 库存总量（默认0），可用库存=总量 |
| `location` | string | 否 | ≤64 字符 | 存放位置 |
| `description` | string | 否 | ≤1024 字符 | 描述 |

```json
{
    "name": "示波器",
    "model": "TDS2014C",
    "category": "电子测量",
    "total_stock": 10,
    "location": "实验室A-101",
    "description": "泰克四通道数字示波器"
}
```

**Response** `201`
```json
{
    "code": 0,
    "msg": "创建成功",
    "data": {
        "id": 1,
        "name": "示波器",
        "model": "TDS2014C",
        "category": "电子测量",
        "total_stock": 10,
        "available_stock": 10,
        "location": "实验室A-101",
        "description": "泰克四通道数字示波器",
        "status": 1,
        "created_at": "2026-06-11 21:03:08",
        "updated_at": "2026-06-11 21:03:08"
    }
}
```

---

### 5.4 更新设备

| 属性 | 值 |
|------|-----|
| **Method** | `PUT` |
| **URL** | `/api/v1/equipments/:id` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 设备ID |

**Request Body (JSON)** — 所有字段可选

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `name` | string | 否 | 1-128 字符 | 设备名称 |
| `model` | string | 否 | ≤64 字符 | 型号 |
| `category` | string | 否 | ≤32 字符 | 分类 |
| `total_stock` | uint | 否 | ≥0 | 库存总量（可用库存联动增减） |
| `location` | string | 否 | ≤64 字符 | 存放位置 |
| `description` | string | 否 | ≤1024 字符 | 描述 |
| `status` | int8 | 否 | 0 或 1 | 上架状态 |

```json
{
    "location": "实验室B-202",
    "total_stock": 15,
    "description": "已搬迁至B栋，库存补充5台"
}
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "更新成功" }
}
```

> **库存联动**: `delta = newTotal - oldTotal`，可用库存 `available_stock += delta`，如果结果 <0 则拒绝。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 库存不足（下调用量） | 400 | 3004 | 库存不足，当前可用库存为 X，无法下调至 Y |
| 设备不存在 | 404 | 3003 | 设备不存在 |

---

### 5.5 下架设备（事务）

| 属性 | 值 |
|------|-----|
| **Method** | `DELETE` |
| **URL** | `/api/v1/equipments/:id` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 设备ID |

```
DELETE /api/v1/equipments/1
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "设备已下架" }
}
```

> **事务保护**: `SELECT ... FOR UPDATE` 锁定设备行 → 检查是否有"已借出"状态的借阅 → 有则拒绝，无则下架。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 设备有未归还借阅 | 409 | 1003 | 设备有未归还借阅，无法下架 |
| 设备不存在 | 404 | 3003 | 设备不存在 |

---

## 6. 借阅工单（需 Token）

### 状态机

```
申请中 ──审批通过──→ 已借出 ──归还──→ 已归还
  │                    ↑
  ├──审批拒绝──→ 被拒绝   │
  └──取消──→ 已取消       │
                          
终态: 已归还 / 被拒绝 / 已取消
```

### 6.1 申请借阅

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/borrows/apply` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `equipment_id` | uint | ✅ | &gt;0 | 设备ID |
| `quantity` | uint | ✅ | &gt;0 | 借阅数量 |
| `apply_note` | string | 否 | ≤256 字符 | 申请备注 |

```json
{
    "equipment_id": 1,
    "quantity": 2,
    "apply_note": "实验课需要使用"
}
```

**Response** `201`
```json
{
    "code": 0,
    "msg": "创建成功",
    "data": {
        "id": 1,
        "user_id": 1,
        "equipment_id": 1,
        "quantity": 2,
        "status": "申请中",
        "apply_note": "实验课需要使用",
        "approve_note": "",
        "approver_id": null,
        "apply_at": "2026-06-11 21:09:26",
        "approve_at": null,
        "return_at": null
    }
}
```

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 设备不存在或已下架 | 404 | 3003 | 设备不存在或已下架 |
| 库存不足 | 400 | 3004 | 设备库存不足 |
| 已有该设备未归还借阅 | 409 | 3005 | 您已有该设备的未归还借阅 |

---

### 6.2 审批借阅（事务 + 库存行锁）

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/borrows/:id/approve` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 工单ID |

**Request Body (JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `approve` | bool | ✅ | true=通过, false=拒绝 |
| `approve_note` | string | 否 | 审批备注（≤256字符） |

```json
{
    "approve": true,
    "approve_note": "同意借用"
}
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1,
        "user_id": 1,
        "equipment_id": 1,
        "quantity": 2,
        "status": "已借出",
        "apply_note": "实验课需要使用",
        "approve_note": "同意借用",
        "approver_id": 1,
        "apply_at": "2026-06-11 21:09:26",
        "approve_at": "2026-06-11 21:10:15",
        "return_at": null
    }
}
```

> **事务保护**: ① `FOR UPDATE` 锁定工单行 ② `FOR UPDATE` 锁定设备行 ③ 通过时扣减 `available_stock` ④ 更新工单状态。
> 
> **拒绝时**: 不扣库存，工单状态 → `被拒绝`。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 工单不存在 | 404 | 1002 | 工单不存在 |
| 工单状态异常（非申请中） | 400 | 3007 | 工单状态异常，无法审批 |
| 库存不足 | 400 | 3006 | 库存不足，无法借出 |

---

### 6.3 归还借阅（事务 + 库存行锁）

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/borrows/:id/return` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |
| **Body** | 无（空 body 或 `{}`） |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 工单ID |

```
POST /api/v1/borrows/1/return
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1,
        "status": "已归还",
        "return_at": "2026-06-11 21:10:15",
        "...": "..."
    }
}
```

> **权限规则**: 管理员可归还任意已借出工单；普通成员只能归还自己的。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 工单状态异常（非已借出） | 400 | 3007 | 工单状态异常，无法归还 |
| 越权归还他人借阅 | 403 | 2005 | 权限不足 |
| 工单不存在 | 404 | 1002 | 工单不存在 |

---

### 6.4 取消申请

| 属性 | 值 |
|------|-----|
| **Method** | `POST` |
| **URL** | `/api/v1/borrows/:id/cancel` |
| **Content-Type** | `application/json` |
| **Auth** | Bearer Token |
| **Body** | 无（空 body 或 `{}`） |

**Path Parameters**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 工单ID |

```
POST /api/v1/borrows/2/cancel
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": { "msg": "已取消申请" }
}
```

> **规则**: 仅工单归属人可取消；仅 `申请中` 状态可取消。

**错误场景**

| 场景 | 响应码 | 错误码 | 消息 |
|------|--------|--------|------|
| 越权取消他人 | 403 | 2005 | 权限不足 |
| 工单状态异常 | 400 | 3007 | 工单状态异常，无法取消 |

---

### 6.5 我的借阅记录（分页）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/borrows/my` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Query Parameters**

| 字段 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 10 | 每页条数 |
| `status` | string | 否 | - | 状态筛选：`申请中`/`已借出`/`已归还`/`被拒绝`/`已取消` |

```
GET /api/v1/borrows/my?page=1&page_size=10&status=已借出
```

**Response** `200`
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "total": 1,
        "page": 1,
        "page_size": 10,
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "equipment_id": 1,
                "quantity": 2,
                "status": "已借出",
                "apply_note": "实验需要",
                "approve_note": "同意借用",
                "approver_id": 1,
                "apply_at": "2026-06-11 21:09:26",
                "approve_at": "2026-06-11 21:10:15",
                "return_at": null,
                "user": { "id": 1, "username": "admin", "real_name": "系统管理员" },
                "equipment": { "id": 1, "name": "示波器" }
            }
        ]
    }
}
```

---

### 6.6 待审批列表（分页）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/borrows/pending` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Query Parameters**

| 字段 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 10 | 每页条数 |

```
GET /api/v1/borrows/pending?page=1&page_size=10
```

**Response** `200` — 结构同 `6.5`，固定筛选 `status=申请中`。

---

### 6.7 全部借阅记录（分页）

| 属性 | 值 |
|------|-----|
| **Method** | `GET` |
| **URL** | `/api/v1/borrows` |
| **Auth** | Bearer Token |
| **Body** | 无 |

**Query Parameters**

| 字段 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 10 | 每页条数 |
| `user_id` | uint | 否 | - | 按用户筛选 |
| `equipment_id` | uint | 否 | - | 按设备筛选 |
| `status` | string | 否 | - | 状态筛选 |

```
GET /api/v1/borrows?page=1&page_size=10&status=已归还&user_id=1
```

**Response** `200` — 结构同 `6.5`，含分页+关联数据。

---

## 7. 错误码速查

| 分组 | 码值 | 常量 | 消息 |
|------|------|------|------|
| 通用 | 1001 | `ErrInvalidParam` | 请求参数错误 |
| 通用 | 1002 | `ErrNotFound` | 资源不存在 |
| 通用 | 1003 | `ErrConflict` | 资源冲突 |
| 通用 | 1004 | `ErrInternal` | 内部错误 |
| 认证 | 2001 | `ErrAuthFailed` | 用户名或密码错误 |
| 认证 | 2002 | `ErrAccountDisabled` | 账号已被禁用 |
| 认证 | 2003 | `ErrTokenMissing` | 未提供认证Token |
| 认证 | 2004 | `ErrTokenInvalid` | Token无效或已过期 |
| 认证 | 2005 | `ErrPermissionDenied` | 权限不足 |
| 业务 | 3001 | `ErrUserExists` | 用户名已存在 |
| 业务 | 3002 | `ErrUserNotFound` | 用户不存在 |
| 业务 | 3003 | `ErrEquipmentNotFound` | 设备不存在 |
| 业务 | 3004 | `ErrStockInsufficient` | 设备库存不足 |
| 业务 | 3005 | `ErrDuplicateBorrow` | 已有该设备的未归还借阅 |
| 业务 | 3006 | `ErrBorrowApproveFailed` | 审批失败，库存不足 |
| 业务 | 3007 | `ErrBorrowStatusInvalid` | 工单状态异常 |
| 系统 | 5000 | `ErrInternalServer` | 系统内部错误 |

---

## 8. 全局说明

### 请求头
```
Content-Type: application/json
Authorization: Bearer <token>    ← 除 /health 和 /auth/login 外所有接口必须携带
```

### 统一响应格式
```json
{
    "code": 0,          // 0=成功, 其他=错误（见错误码表）
    "msg": "success",   // 提示信息
    "data": {}          // 响应数据（可能为 null）
}
```

### 分页响应格式（列表接口统一使用）
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "total": 100,       // 总记录数
        "page": 1,          // 当前页码
        "page_size": 10,    // 每页条数
        "list": []          // 数据列表
    }
}
```

### 测试流程建议

```
① GET  /api/v1/health                    → 验证服务可用
② POST /api/v1/auth/login                → 获取 Token         ⭐ 保存 token
③ GET  /api/v1/roles                     → 查看角色列表
④ POST /api/v1/users                     → 创建成员用户
⑤ GET  /api/v1/users?page=1&page_size=10 → 查看用户列表
⑥ GET  /api/v1/users/2                   → 查看用户详情  
⑦ PUT  /api/v1/users/2                   → 更新用户信息
⑧ PUT  /api/v1/users/2/password          → 修改密码
⑨ POST /api/v1/equipments                → 创建设备
⑩ GET  /api/v1/equipments                → 查看设备列表
⑪ GET  /api/v1/equipments/1              → 查看设备详情
⑫ PUT  /api/v1/equipments/1              → 更新设备
⑬ POST /api/v1/borrows/apply             → 申请借阅
⑭ GET  /api/v1/borrows/pending           → 查看待审批
⑮ POST /api/v1/borrows/1/approve         → 审批通过         ⭐ approve:true
⑯ GET  /api/v1/borrows/my                → 我的借阅记录
⑰ POST /api/v1/borrows/1/return          → 归还借阅
⑱ POST /api/v1/borrows/apply             → 再申请一条
⑲ POST /api/v1/borrows/2/cancel          → 取消申请
⑳ POST /api/v1/users/2/disable           → 禁用用户
㉑ DELETE /api/v1/equipments/1            → 下架设备
㉒ POST /api/v1/auth/refresh              → 刷新 Token        ⭐ 旧 token 失效
㉓ POST /api/v1/auth/logout               → 登出              ⭐ token 入黑名单
```
