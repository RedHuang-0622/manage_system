# 02 — 前端 REST API

> **Base URL:** `http://localhost:8080/api/v1`  
> **Content-Type:** `application/json`  
> **认证方式:** `Authorization: Bearer <jwt_token>`

---

## 1 认证接口

### 1.1 登录

```
POST /auth/login
```

| 项目 | 内容 |
|------|------|
| 权限 | 无 |
| 请求 | `{"username":"string","password":"string"}` |
| 成功 | 200 `{"code":0,"msg":"success","data":{"token":"eyJ...","expires_in":7200}}` |
| 失败 | 400 参数校验 / 2001 用户名或密码错误 / 2002 账号禁用 |

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 1.2 登出

```
POST /auth/logout
```

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 请求 | 无 Body |
| 成功 | 200 `{"code":0,"msg":"登出成功","data":null}` |
| 失败 | 401 Token 无效 |

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <token>"
```

### 1.3 刷新 Token

```
POST /auth/refresh
```

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 请求 | `{"token":"<current_token>"}` |
| 成功 | 200 `{"code":0,"msg":"success","data":{"token":"<new>","expires_in":7200}}` |
| 失败 | 2004 Token 无效/过期/已登出 |

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <old_token>" \
  -d '{"token":"<old_token>"}'
```

---

## 2 用户管理接口

### 2.1 角色列表

```
GET /roles
```

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 成功 | 200 `{"code":0,"msg":"success","data":[{"id":1,"role_name":"super_admin","description":"..."},...]}` |

**curl:**
```bash
curl http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <token>"
```

### 2.2 创建用户

```
POST /users
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 请求 | `{"username":"s","password":"s","real_name":"s","email":"s","phone":"s","role_id":n}` |
| 成功 | 201 `{"code":0,"msg":"创建成功","data":{"id":n,"username":"...","real_name":"...",...}}` |
| 失败 | 400 参数校验 / 3001 用户名已存在 / 403 权限不足 |

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "username":"zhangsan","password":"123456","real_name":"张三",
    "email":"zhangsan@lab.edu.cn","phone":"13800138000","role_id":3
  }'
```

### 2.3 用户列表（分页）

```
GET /users?page=1&page_size=10&keyword=&status=-1&role_id=0
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页条数（最大100） |
| keyword | string | "" | 模糊搜索 username / real_name |
| status | int | -1 | 1=启用, 0=禁用, -1=全部 |
| role_id | int | 0 | 0=全部 |

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 成功 | 200 `{"code":0,"msg":"success","data":{"total":n,"page":1,"page_size":10,"list":[...]}}` |

**curl:**
```bash
curl "http://localhost:8080/api/v1/users?page=1&page_size=10&keyword=张&status=1" \
  -H "Authorization: Bearer <admin_token>"
```

### 2.4 用户详情

```
GET /users/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin，或本人 |
| 成功 | 200 `{"code":0,"msg":"success","data":{"id":n,"username":"...","role":{...},...}}` |
| 失败 | 3002 用户不存在 |

### 2.5 更新用户

```
PUT /users/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 请求 | `{"real_name":"s","email":"s","phone":"s","role_id":n,"status":n}` (全部可选) |
| 成功 | 200 `{"code":0,"msg":"更新成功","data":{...}}` |

**约束：** 不允许修改 username / password（password 另有独立接口）

### 2.6 禁用用户

```
DELETE /users/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin |
| 成功 | 200 `{"code":0,"msg":"禁用成功","data":null}` |
| 失败 | 403 权限不足 / 不允许禁用自己 / 有未归还借阅 |

### 2.7 修改密码

```
PUT /users/{id}/password
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin 可改任意用户；其他用户仅可改本人 |
| 请求 | `{"old_password":"s","new_password":"s"}` |
| 成功 | 200 `{"code":0,"msg":"密码修改成功","data":null}` |
| 失败 | 2001 原密码错误 / 403 权限不足 |

---

## 3 设备管理接口

### 3.1 设备大厅（列表）

```
GET /equipments?page=1&page_size=12&keyword=&category=&status=-1&only_available=0
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 12 | 每页条数 |
| keyword | string | "" | 模糊搜索名称 |
| category | string | "" | 分类筛选 |
| status | int | -1 | 1=上架, 0=下架, -1=全部 |
| only_available | int | 0 | 1=仅显示有库存 |

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 成功 | 200 `{"code":0,"msg":"success","data":{"total":n,"list":[...]}}` |

**curl:**
```bash
curl "http://localhost:8080/api/v1/equipments?page=1&page_size=12&category=服务器&only_available=1" \
  -H "Authorization: Bearer <token>"
```

**缓存行为：** 首次查询写 Redis（TTL 30s），后续命中缓存直接返回；写操作后缓存自动失效。

### 3.2 设备详情

```
GET /equipments/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 成功 | 200 `{"code":0,"msg":"success","data":{...}}` |
| 失败 | 3003 设备不存在 |

**缓存行为：** 首次查 DB 写 Redis（TTL 60s）；更新/下架时缓存失效。

### 3.3 创建设备（入库）

```
POST /equipments
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 请求 | `{"name":"s","model":"s","category":"s","total_stock":n,"location":"s","description":"s"}` |
| 成功 | 201 `{"code":0,"msg":"创建成功","data":{...}}` |

**副作用：** available_stock = total_stock，批量失效列表缓存。

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/equipments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"name":"GPU A100","model":"DGX-A100","category":"服务器","total_stock":4,"location":"A301"}'
```

### 3.4 更新设备

```
PUT /equipments/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 请求 | `{"name":"s","model":"s","category":"s","total_stock":n,"location":"s","description":"s","status":n}` (全部可选) |
| 成功 | 200 `{"code":0,"msg":"更新成功","data":{...}}` |

**库存联动：** available = old_available + (new_total - old_total)，必须 ≥0。  
**副作用：** 失效详情缓存 + 批量失效列表缓存。

### 3.5 下架设备

```
DELETE /equipments/{id}
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin |
| 请求 | 无 |
| 成功 | 200 `{"code":0,"msg":"下架成功","data":null}` |
| 失败 | 有未归还借阅时拒绝 |

---

## 4 借阅管理接口

### 4.1 发起借阅申请

```
POST /borrows/apply
```

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户 |
| 请求 | `{"equipment_id":n,"quantity":n,"apply_note":"s"}` |
| 成功 | 201 `{"code":0,"msg":"申请已提交，等待审批","data":{"id":n,"status":"申请中","apply_at":"..."}}` |
| 失败 | 3003 设备不存在 / 3004 库存不足 / 3005 已有未归还借阅 |

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/borrows/apply \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <member_token>" \
  -d '{"equipment_id":1,"quantity":1,"apply_note":"实验需要"}'
```

### 4.2 审批借阅

```
POST /borrows/{id}/approve
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 请求 | `{"approve":true/false,"approve_note":"s"}` |
| 成功 | 200 `{"code":0,...}` |
| 失败 | 3006 库存不足 / 3007 状态异常 / 403 权限不足 |

**实现：** DB 事务内锁定 borrow_record + lab_equipment 行，扣库存并更新状态。

**curl (通过):**
```bash
curl -X POST http://localhost:8080/api/v1/borrows/100/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"approve":true,"approve_note":"同意"}'
```

**curl (拒绝):**
```bash
curl -X POST http://localhost:8080/api/v1/borrows/100/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"approve":false,"approve_note":"设备已预约，请改天"}'
```

### 4.3 归还设备

```
POST /borrows/{id}/return
```

| 项目 | 内容 |
|------|------|
| 权限 | 借用人本人 / lab_admin / super_admin |
| 请求 | 无 |
| 成功 | 200 `{"code":0,"msg":"归还成功","data":{"id":n,"status":"已归还","return_at":"..."}}` |

**实现：** DB 事务内锁定记录 + 设备行，恢复库存。

**curl:**
```bash
curl -X POST http://localhost:8080/api/v1/borrows/100/return \
  -H "Authorization: Bearer <token>"
```

### 4.4 我的借阅记录

```
GET /borrows/my?page=1&page_size=10&status=
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页条数 |
| status | string | "" | 筛选：申请中/已借出/已归还/被拒绝 |

| 项目 | 内容 |
|------|------|
| 权限 | 登录用户（只能查看自己） |
| 成功 | 200 `{"code":0,"msg":"success","data":{"total":n,"list":[...]}}` |

**响应中每条记录含 `equipment` 摘要（id, name, model）**

### 4.5 待审批列表

```
GET /borrows/pending?page=1&page_size=10
```

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |
| 成功 | 200 `{"code":0,"msg":"success","data":{"total":n,"list":[...]}}` |

**返回所有 status='申请中' 的工单，含 user 和 equipment 摘要。**

### 4.6 全部借阅记录（管理员）

```
GET /borrows?page=1&page_size=10&user_id=0&equipment_id=0&status=
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页条数 |
| user_id | int | 0 | 按借用人筛选，0=全部 |
| equipment_id | int | 0 | 按设备筛选，0=全部 |
| status | string | "" | 按状态筛选 |

| 项目 | 内容 |
|------|------|
| 权限 | super_admin, lab_admin |

---

## 5 健康检查

```
GET /health
```

| 项目 | 内容 |
|------|------|
| 权限 | 无 |
| 成功 | 200 `{"code":0,"msg":"success","data":{"status":"ok"}}` |

**curl:**
```bash
curl http://localhost:8080/api/v1/health
```

---

## 6 接口速查表

| 方法 | 路径 | 权限 | 所属模块 |
|------|------|------|---------|
| GET | `/health` | 无 | 健康检查 |
| POST | `/auth/login` | 无 | IAM |
| POST | `/auth/logout` | 登录 | IAM |
| POST | `/auth/refresh` | 登录 | IAM |
| GET | `/roles` | 登录 | IAM |
| GET | `/users` | admin | IAM |
| GET | `/users/{id}` | admin/本人 | IAM |
| POST | `/users` | admin | IAM |
| PUT | `/users/{id}` | admin | IAM |
| DELETE | `/users/{id}` | super | IAM |
| PUT | `/users/{id}/password` | 登录/本人 | IAM |
| GET | `/equipments` | 登录 | Equipment |
| GET | `/equipments/{id}` | 登录 | Equipment |
| POST | `/equipments` | admin | Equipment |
| PUT | `/equipments/{id}` | admin | Equipment |
| DELETE | `/equipments/{id}` | super | Equipment |
| POST | `/borrows/apply` | 登录 | Borrow |
| POST | `/borrows/{id}/approve` | admin | Borrow |
| POST | `/borrows/{id}/return` | 登录 | Borrow |
| GET | `/borrows/my` | 登录 | Borrow |
| GET | `/borrows/pending` | admin | Borrow |
| GET | `/borrows` | admin | Borrow |

**总计：22 个接口（含 health）**
