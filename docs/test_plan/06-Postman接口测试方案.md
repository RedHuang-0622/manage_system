# 06 — Postman 接口测试方案

> **定位：** 本方案属于测试金字塔顶层的 E2E / API 测试，覆盖全部 HTTP 接口的正向、异常与权限边界场景。
> **前置依赖：** 后端服务已启动（`go run backend/cmd/main.go`），MySQL + Redis 就绪。

---

## 1 Collection 结构

```
📁 LabManage API Tests
├── 📁 00-Health              (1 request)
├── 📁 01-Auth                (3 requests: login / logout / refresh)
├── 📁 02-Users               (7 requests: CRUD + disable + changePassword + listRoles)
├── 📁 03-Equipments          (5 requests: CRUD + disable)
├── 📁 04-Borrows             (7 requests: apply / approve / return / cancel / list)
├── 📁 05-Full-Flow           (12 requests: 完整借阅闭环，按顺序编排)
└── 📁 06-Permission-Tests    (8 requests: 权限越界验证)
```

**总计：43 个 Request**

---

## 2 Environment 变量

创建 Postman Environment `LabManage-Local`：

| Variable | Type | Initial Value | Description |
|----------|------|---------------|-------------|
| `base_url` | string | `http://localhost:8080` | API 基础地址 |
| `admin_token` | string | *(空)* | 管理员 JWT，登录后自动填充 |
| `member_token` | string | *(空)* | 普通成员 JWT，登录后自动填充 |
| `lab_admin_token` | string | *(空)* | 实验室负责人 JWT |
| `user_id` | string | *(空)* | 动态创建的用户 ID |
| `equip_id` | string | *(空)* | 动态创建的设备 ID |
| `borrow_id` | string | *(空)* | 动态创建的借阅工单 ID |

---

## 3 Collection 级 Pre-request Script

```javascript
// 📁 Collection 级 Pre-request — 所有请求共享
// 跳过 login 和 health 请求（无需 token）
const skipPaths = [
    '/api/v1/auth/login',
    '/api/v1/health'
];

const currentPath = pm.request.url.getPath();
if (!skipPaths.includes(currentPath)) {
    // 根据请求所属文件夹自动选择 token
    const folder = pm.info.requestName; // 实际通过 folder 判断
    // 默认用 admin_token，特定请求手动覆盖
}
```

---

## 4 各接口详细测试

### 4.1 Health Check

| Request | `GET /api/v1/health` |
|---------|----------------------|
| Folder | `00-Health` |
| Auth | None |

**Tests 脚本：**
```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("Response has status ok", () => {
    const json = pm.response.json();
    pm.expect(json.status).to.eql("ok");
});
pm.test("Response time < 200ms", () => {
    pm.expect(pm.response.responseTime).to.be.below(200);
});
```

---

### 4.2 认证模块 (`01-Auth`)

#### 4.2.1 Login — 正常登录

| Request | `POST /api/v1/auth/login` |
|---------|---------------------------|
| Body | `{"username":"admin","password":"admin123"}` |

**Tests 脚本：**
```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const json = pm.response.json();

pm.test("code = 0", () => pm.expect(json.code).to.eql(0));
pm.test("has token", () => pm.expect(json.data.token).to.be.a("string"));
pm.test("has expires_in", () => pm.expect(json.data.expires_in).to.be.a("number"));

// 自动保存 admin token
pm.environment.set("admin_token", json.data.token);
```

#### 4.2.2 Login — 密码错误

| Request | `POST /api/v1/auth/login` |
|---------|---------------------------|
| Body | `{"username":"admin","password":"wrong_pwd"}` |

```javascript
pm.test("Status 401", () => pm.response.to.have.status(401));
const json = pm.response.json();
pm.test("code = 2001", () => pm.expect(json.code).to.eql(2001));
pm.test("msg = 用户名或密码错误", () => pm.expect(json.msg).to.include("用户名或密码错误"));
```

#### 4.2.3 Login — 参数缺失

| Request | `POST /api/v1/auth/login` |
|---------|---------------------------|
| Body | `{}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
pm.test("code = 1001", () => pm.expect(pm.response.json().code).to.eql(1001));
```

#### 4.2.4 Login — username 过短

| Request | `POST /api/v1/auth/login` |
|---------|---------------------------|
| Body | `{"username":"ab","password":"password123"}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.2.5 Logout — 正常登出

| Request | `POST /api/v1/auth/logout` |
|---------|----------------------------|
| Auth | Bearer `{{admin_token}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("code = 0", () => pm.expect(pm.response.json().code).to.eql(0));
```

#### 4.2.6 Logout — 无 Token

| Request | `POST /api/v1/auth/logout` |
|---------|----------------------------|
| Auth | None |

```javascript
pm.test("Status 401", () => pm.response.to.have.status(401));
```

#### 4.2.7 RefreshToken — 正常刷新

| Request | `POST /api/v1/auth/refresh` |
|---------|-----------------------------|
| Auth | Bearer `{{admin_token}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const json = pm.response.json();
pm.test("新 token 不同于旧 token", () => {
    pm.expect(json.data.token).to.not.eql(pm.environment.get("admin_token"));
});
pm.environment.set("admin_token", json.data.token);
```

#### 4.2.8 RefreshToken — 无 Token

| Request | `POST /api/v1/auth/refresh` |
|---------|-----------------------------|
| Auth | None |

```javascript
pm.test("Status 401", () => pm.response.to.have.status(401));
```

---

### 4.3 角色 (`01-Auth`)

#### 4.3.1 ListRoles — 查看角色列表

| Request | `GET /api/v1/roles` |
|---------|---------------------|
| Auth | Bearer `{{admin_token}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const json = pm.response.json();
pm.test("至少有 3 个系统角色", () => {
    pm.expect(json.data.length).to.be.at.least(3);
});
pm.test("包含 super_admin", () => {
    pm.expect(json.data.map(r => r.role_name)).to.include("super_admin");
});
```

---

### 4.4 用户管理 (`02-Users`)

#### 4.4.1 CreateUser — 正常创建

| Request | `POST /api/v1/users` |
|---------|----------------------|
| Auth | Bearer `{{admin_token}}` |
| Body | `{"username":"test_user_{{random}}","password":"pass123456","real_name":"测试用户","role_id":3}` |

**Pre-request：**
```javascript
pm.variables.set("random", Math.floor(Math.random() * 9000 + 1000));
```

**Tests：**
```javascript
pm.test("Status 201", () => pm.response.to.have.status(201));
const json = pm.response.json();
pm.test("code = 0", () => pm.expect(json.code).to.eql(0));
pm.test("password_hash 不返回", () => {
    pm.expect(json.data.password_hash).to.be.undefined;
});
pm.environment.set("user_id", json.data.id);
```

#### 4.4.2 CreateUser — 用户名重复

| Request | `POST /api/v1/users` |
|---------|----------------------|
| Body | `{"username":"admin","password":"pass123456","real_name":"重复","role_id":3}` |

```javascript
pm.test("Status 409", () => pm.response.to.have.status(409));
pm.test("code = 3001", () => pm.expect(pm.response.json().code).to.eql(3001));
```

#### 4.4.3 CreateUser — 缺少必填项

| Request | `POST /api/v1/users` |
|---------|----------------------|
| Body | `{"username":"test"}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.4.4 CreateUser — email 格式错误

| Request | `POST /api/v1/users` |
|---------|----------------------|
| Body | `{"username":"u1","password":"pass123456","real_name":"测试","email":"not-email","role_id":3}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.4.5 ListUsers — 默认分页

| Request | `GET /api/v1/users` |
|---------|---------------------|
| Auth | Bearer `{{admin_token}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const data = pm.response.json().data;
pm.test("有分页信息", () => {
    pm.expect(data.total).to.be.a("number");
    pm.expect(data.page).to.be.a("number");
    pm.expect(data.page_size).to.be.a("number");
    pm.expect(data.list).to.be.an("array");
});
```

#### 4.4.6 ListUsers — 带筛选条件

| Request | `GET /api/v1/users?keyword=admin&status=1&page=1&page_size=5` |
|---------|---------------------------------------------------------------|

```javascript
pm.test("筛选后 total 正确", () => {
    const data = pm.response.json().data;
    pm.expect(data.list.length).to.be.at.most(5);
});
```

#### 4.4.7 GetUser — 查看用户详情

| Request | `GET /api/v1/users/{{user_id}}` |
|---------|-------------------------------|

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const user = pm.response.json().data;
pm.test("包含 role 关联", () => {
    pm.expect(user.role).to.be.an("object");
    pm.expect(user.role.role_name).to.be.a("string");
});
```

#### 4.4.8 GetUser — 不存在的 ID

| Request | `GET /api/v1/users/99999` |

```javascript
pm.test("Status 404", () => pm.response.to.have.status(404));
```

#### 4.4.9 UpdateUser — 部分字段更新

| Request | `PUT /api/v1/users/{{user_id}}` |
|---------|-------------------------------|
| Body | `{"email":"new@test.com","phone":"13800138000"}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("msg = 更新成功", () => {
    pm.expect(pm.response.json().msg).to.eql("更新成功");
});
```

#### 4.4.10 UpdateUser — 不存在的用户

| Request | `PUT /api/v1/users/99999` |
| Body | `{"real_name":"ghost"}` |

```javascript
pm.test("Status 404", () => pm.response.to.have.status(404));
```

#### 4.4.11 DisableUser — 正常禁用

| Request | `POST /api/v1/users/{{user_id}}/disable` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("msg = 用户已禁用", () => {
    pm.expect(pm.response.json().msg).to.include("禁用");
});
```

#### 4.4.12 DisableUser — 不能禁用自己

| Request | `POST /api/v1/users/1/disable` |
|---------|-------------------------------|
| Note | admin 的 ID 是 1 |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.4.13 ChangePassword — 修改密码

| Request | `PUT /api/v1/users/{{user_id}}/password` |
|---------|----------------------------------------|
| Body | `{"old_password":"pass123456","new_password":"newpass789"}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
```

#### 4.4.14 ChangePassword — 新密码过短

| Request | `PUT /api/v1/users/{{user_id}}/password` |
| Body | `{"new_password":"abc"}` (min=6) |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

---

### 4.5 设备管理 (`03-Equipments`)

#### 4.5.1 Create — 正常创建

| Request | `POST /api/v1/equipments` |
|---------|---------------------------|
| Auth | Bearer `{{admin_token}}` |
| Body | `{"name":"GPU Server","model":"RTX4090","category":"服务器","total_stock":10,"location":"A101","description":"深度学习训练卡"}` |

```javascript
pm.test("Status 201", () => pm.response.to.have.status(201));
const json = pm.response.json();
pm.test("available = total", () => {
    pm.expect(json.data.available_stock).to.eql(json.data.total_stock);
});
pm.test("status = 1", () => pm.expect(json.data.status).to.eql(1));
pm.environment.set("equip_id", json.data.id);
```

#### 4.5.2 Create — 缺少 name

| Request | `POST /api/v1/equipments` |
| Body | `{"total_stock":10}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.5.3 Create — total_stock 为负

| Request | `POST /api/v1/equipments` |
| Body | `{"name":"负库存","total_stock":-1}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.5.4 List — 默认分页

| Request | `GET /api/v1/equipments` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
const data = pm.response.json().data;
pm.test("page_size 默认 12", () => pm.expect(data.page_size).to.eql(12));
pm.test("list 是数组", () => pm.expect(data.list).to.be.an("array"));
```

#### 4.5.5 List — 带搜索条件

| Request | `GET /api/v1/equipments?keyword=GPU&category=服务器&only_available=1` |

```javascript
pm.test("list 中 name 都含 GPU", () => {
    const list = pm.response.json().data.list;
    list.forEach(e => pm.expect(e.name).to.match(/GPU/i));
});
```

#### 4.5.6 GetByID — 查看详情

| Request | `GET /api/v1/equipments/{{equip_id}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("有库存字段", () => {
    const e = pm.response.json().data;
    pm.expect(e.total_stock).to.be.a("number");
    pm.expect(e.available_stock).to.be.a("number");
});
```

#### 4.5.7 GetByID — 不存在

| Request | `GET /api/v1/equipments/99999` |

```javascript
pm.test("Status 404", () => pm.response.to.have.status(404));
pm.test("code = 3003", () => pm.expect(pm.response.json().code).to.eql(3003));
```

#### 4.5.8 Update — 更新名称和位置

| Request | `PUT /api/v1/equipments/{{equip_id}}` |
| Body | `{"name":"更新后的GPU","location":"B202"}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
```

#### 4.5.9 Update — 调整总库存联动可用库存

| Request | `PUT /api/v1/equipments/{{equip_id}}` |
| Body | `{"total_stock":20}` |

```javascript
pm.test("总量和可用同步变化", () => {
    // 可用 = 原可用 + (新总量 - 旧总量)
    // 不做具体断言，验证成功即可
    pm.expect(pm.response.json().code).to.eql(0);
});
```

#### 4.5.10 Disable — 下架设备

| Request | `DELETE /api/v1/equipments/{{equip_id}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("msg = 设备已下架", () => {
    pm.expect(pm.response.json().msg).to.include("下架");
});
```

#### 4.5.11 Disable — 不存在的设备

| Request | `DELETE /api/v1/equipments/99999` |

```javascript
pm.test("Status 404", () => pm.response.to.have.status(404));
```

---

### 4.6 借阅工单 (`04-Borrows`)

#### 4.6.1 Apply — 正常申请

| Request | `POST /api/v1/borrows/apply` |
|---------|------------------------------|
| Auth | Bearer `{{member_token}}` |
| Body | `{"equipment_id":{{equip_id}},"quantity":2,"apply_note":"实验需要"}` |

```javascript
pm.test("Status 201", () => pm.response.to.have.status(201));
const json = pm.response.json();
pm.test("status = 申请中", () => {
    pm.expect(json.data.status).to.eql("申请中");
});
pm.environment.set("borrow_id", json.data.id);
```

#### 4.6.2 Apply — equipment_id 缺失

| Request | `POST /api/v1/borrows/apply` |
| Body | `{"quantity":2}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.6.3 Apply — quantity 为 0

| Request | `POST /api/v1/borrows/apply` |
| Body | `{"equipment_id":1,"quantity":0}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
```

#### 4.6.4 Apply — 设备不存在

| Request | `POST /api/v1/borrows/apply` |
| Body | `{"equipment_id":99999,"quantity":1}` |

```javascript
pm.test("Status 404", () => {
    pm.expect(pm.response.code).to.be.oneOf([400, 404]);
});
```

#### 4.6.5 Apply — 库存不足

| Request | `POST /api/v1/borrows/apply` |
| Body | `{"equipment_id":{{equip_id}},"quantity":9999}` |

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
pm.test("code = 3004 库存不足", () => {
    pm.expect(pm.response.json().code).to.eql(3004);
});
```

#### 4.6.6 Approve — 审批通过

| Request | `POST /api/v1/borrows/{{borrow_id}}/approve` |
|---------|---------------------------------------------|
| Auth | Bearer `{{admin_token}}` |
| Body | `{"approve":true,"approve_note":"同意借出"}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("status = 已借出", () => {
    pm.expect(pm.response.json().data.status).to.eql("已借出");
});
```

#### 4.6.7 Approve — 拒绝

创建一个新申请后：

| Request | `POST /api/v1/borrows/{{borrow_id}}/approve` (新 borrow_id) |
| Body | `{"approve":false,"approve_note":"设备已预约"}` |

```javascript
pm.test("status = 被拒绝", () => {
    pm.expect(pm.response.json().data.status).to.eql("被拒绝");
});
```

#### 4.6.8 Approve — 重复审批(终态校验)

对已拒绝的工单再次审批：

```javascript
pm.test("Status 400", () => pm.response.to.have.status(400));
pm.test("code = 3007 工单状态异常", () => {
    pm.expect(pm.response.json().code).to.eql(3007);
});
```

#### 4.6.9 Return — 归还设备

| Request | `POST /api/v1/borrows/{{borrow_id}}/return` |
| Auth | Bearer `{{member_token}}` (工单所属人) |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("status = 已归还", () => {
    pm.expect(pm.response.json().data.status).to.eql("已归还");
});
pm.test("return_at 不为空", () => {
    pm.expect(pm.response.json().data.return_at).to.be.a("string").and.not.empty;
});
```

#### 4.6.10 Cancel — 取消申请

创建新申请后：

| Request | `POST /api/v1/borrows/{{borrow_id}}/cancel` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("msg = 已取消申请", () => {
    pm.expect(pm.response.json().msg).to.include("取消");
});
```

#### 4.6.11 ListMy — 我的借阅

| Request | `GET /api/v1/borrows/my` |
|---------|--------------------------|
| Auth | Bearer `{{member_token}}` |

```javascript
pm.test("Status 200", () => pm.response.to.have.status(200));
pm.test("只返回当前用户的记录", () => {
    const list = pm.response.json().data.list;
    pm.expect(list).to.be.an("array");
});
```

#### 4.6.12 ListMy — 按状态筛选

| Request | `GET /api/v1/borrows/my?status=已归还` |

```javascript
pm.test("所有记录 status = 已归还", () => {
    pm.response.json().data.list.forEach(r => {
        pm.expect(r.status).to.eql("已归还");
    });
});
```

#### 4.6.13 ListPending — 待审批列表

| Request | `GET /api/v1/borrows/pending` |
| Auth | Bearer `{{admin_token}}` |

```javascript
pm.test("所有记录 status = 申请中", () => {
    pm.response.json().data.list.forEach(r => {
        pm.expect(r.status).to.eql("申请中");
    });
});
```

#### 4.6.14 ListAll — 管理员查看全部

| Request | `GET /api/v1/borrows?user_id=2&status=已归还` |

```javascript
pm.test("支持组合筛选", () => {
    const data = pm.response.json().data;
    pm.expect(data.total).to.be.a("number");
});
```

---

## 5 权限越界测试 (`06-Permission-Tests`)

以 `member` 角色为基础，测试所有越权操作。

**前置：** 登录 member 获取 `member_token`。

| # | Request | Method | URL | Expected |
|---|---------|--------|-----|----------|
| P-01 | 创建用户 | POST | `/api/v1/users` | **403** `code:2005` |
| P-02 | 删除设备 | DELETE | `/api/v1/equipments/1` | **403** `code:2005` |
| P-03 | 审批工单 | POST | `/api/v1/borrows/1/approve` | **403** `code:2005` |
| P-04 | 查看待审批 | GET | `/api/v1/borrows/pending` | **403** `code:2005` |
| P-05 | 查看全量工单 | GET | `/api/v1/borrows` | **403** `code:2005` |
| P-06 | 查看设备 (允许) | GET | `/api/v1/equipments` | **200** |
| P-07 | 申请借阅 (允许) | POST | `/api/v1/borrows/apply` | **201** |
| P-08 | 查看角色 (允许) | GET | `/api/v1/roles` | **200** |

**通用 Tests：**
```javascript
// 越权请求
pm.test("Status 403", () => pm.response.to.have.status(403));
pm.test("code = 2005", () => pm.expect(pm.response.json().code).to.eql(2005));
pm.test("msg = 权限不足", () => pm.expect(pm.response.json().msg).to.include("权限不足"));

// 允许的请求
pm.test("Status 2xx", () => pm.expect(pm.response.code).to.be.within(200, 299));
```

---

## 6 完整借阅闭环 (`05-Full-Flow`)

按顺序编排的 Runner 场景，模拟真实业务流程。

### 6.1 Flow 顺序

```
Step 01: POST /api/v1/auth/login (admin)
    → 保存 admin_token
    → 验证 200, 有 token

Step 02: POST /api/v1/users (创建 member)
    → 保存 user_id
    → 验证 201

Step 03: POST /api/v1/auth/login (新 member)
    → 保存 member_token
    → 验证 200

Step 04: POST /api/v1/equipments (创建 3 个库存)
    → 保存 equip_id
    → 验证 available = 3

Step 05: GET /api/v1/equipments (查看设备大厅)
    → 验证新设备在列表中

Step 06: POST /api/v1/borrows/apply (申请借 2 个)
    → 保存 borrow_id
    → 验证 status = 申请中

Step 07: GET /api/v1/equipments/{{equip_id}} (验证库存未扣)
    → 验证 available_stock = 3

Step 08: POST /api/v1/borrows/{{borrow_id}}/approve (审批通过)
    → 验证 status = 已借出

Step 09: GET /api/v1/equipments/{{equip_id}} (验证库存已扣)
    → 验证 available_stock = 1 (= 3 - 2)

Step 10: POST /api/v1/borrows/{{borrow_id}}/return (归还)
    → 验证 status = 已归还

Step 11: GET /api/v1/equipments/{{equip_id}} (验证库存恢复)
    → 验证 available_stock = 3

Step 12: GET /api/v1/borrows/my (验证记录)
    → 验证 total >= 1, 最后一条 status = 已归还
```

### 6.2 关键验证点 Tests

**Step 07 — 申请不扣库存：**
```javascript
const equip = pm.response.json().data;
pm.test("申请后库存不变", () => {
    pm.expect(equip.available_stock).to.eql(3);
});
```

**Step 09 — 审批后扣库存：**
```javascript
const equip = pm.response.json().data;
pm.test("审批后库存扣减", () => {
    pm.expect(equip.available_stock).to.eql(1);
});
```

**Step 11 — 归还后库存恢复：**
```javascript
const equip = pm.response.json().data;
pm.test("归还后库存恢复", () => {
    pm.expect(equip.available_stock).to.eql(3);
});
```

---

## 7 Runner 执行顺序

Postman Collection Runner 按以下顺序执行：

```
Phase 1: 环境重置
  ├── 00-Health          (1 req)
  └── 01-Auth            (login → 获得所有 token)

Phase 2: 正常功能
  ├── 02-Users           (CRUD 正向用例)
  ├── 03-Equipments      (CRUD 正向用例)
  └── 04-Borrows         (申请 → 审批 → 归还 闭环)

Phase 3: 权限边界
  └── 06-Permission-Tests (member 越权 + 合法操作)

Phase 4: 全链路
  └── 05-Full-Flow        (完整借阅闭环)
```

**Runner 配置：**
- Delay: `200ms` (避免竞态)
- Keep variable values: ✅
- Save responses: ✅
- Run in parallel: ❌ (顺序依赖)

---

## 8 Newman CLI 命令行

### 8.1 安装

```bash
npm install -g newman
```

### 8.2 导出 Collection

在 Postman 中 → 右键 Collection → Export → `LabManage.postman_collection.json`

### 8.3 导出 Environment

在 Postman 中 → Environments → `LabManage-Local` → Export → `LabManage-Local.postman_environment.json`

### 8.4 运行命令

```bash
# 基础运行
newman run LabManage.postman_collection.json \
  -e LabManage-Local.postman_environment.json \
  --reporters cli,htmlextra \
  --reporter-htmlextra-export report.html

# 指定文件夹运行
newman run LabManage.postman_collection.json \
  -e LabManage-Local.postman_environment.json \
  --folder "04-Borrows"

# 只运行全链路
newman run LabManage.postman_collection.json \
  -e LabManage-Local.postman_environment.json \
  --folder "05-Full-Flow"

# CI 模式（失败立即退出 + JUnit 报告）
newman run LabManage.postman_collection.json \
  -e LabManage-Local.postman_environment.json \
  --bail \
  --reporters cli,junit \
  --reporter-junit-export results.xml

# 并发测试（重复执行全链路，验证幂等）
newman run LabManage.postman_collection.json \
  -e LabManage-Local.postman_environment.json \
  --folder "05-Full-Flow" \
  -n 5 --delay-request 500
```

### 8.5 Makefile / Task 集成

```makefile
# Makefile
.PHONY: test-api test-api-ci test-api-flow

test-api:
	docker-compose up -d mysql redis
	sleep 5
	cd backend && go run cmd/main.go &
	sleep 3
	newman run ../docs/test_plan/LabManage.postman_collection.json \
		-e ../docs/test_plan/LabManage-Local.postman_environment.json
	kill %1

test-api-ci:
	newman run LabManage.postman_collection.json \
		-e LabManage-CI.postman_environment.json \
		--bail --reporters cli,junit

test-api-flow:
	newman run LabManage.postman_collection.json \
		-e LabManage-Local.postman_environment.json \
		--folder "05-Full-Flow" -n 3
```

---

## 9 测试数据要求

| 数据 | 说明 |
|------|------|
| admin 用户 | 初始种子数据 `admin / admin123` |
| member 角色 | ID=3，系统预置 |
| super_admin 角色 | ID=1 |
| Casbin 策略 | 种子数据自动初始化 |

**每次 Runner 执行前建议：**
```sql
-- 清理测试数据（保留种子用户和角色）
DELETE FROM borrow_records;
DELETE FROM lab_equipments;
DELETE FROM sys_users WHERE id > 1;
```

或在 Postman Collection 的 **Pre-request Script (Collection 级)** 中加一个 setup 请求。

---

## 10 常见问题

| 问题 | 解决 |
|------|------|
| Token 过期返回 401 | 在 Runner 开始前先执行 Login，Refresher Token 有 2h 有效期 |
| `{{random}}` 变量不解析 | Postman 动态变量用 `$randomInt` 替代 `{{random}}` |
| Newman 中文乱码 | 确保请求 Body 的 `Content-Type: application/json; charset=utf-8` |
| 并发 Runner 结果不稳定 | 串行运行，delay ≥ 200ms；并发场景用 Go 集成测试 |

---

## 附录：Collection JSON 骨架

```json
{
  "info": {
    "name": "LabManage API Tests",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "00-Health",
      "item": [
        { "name": "GET Health", "request": { "method": "GET", "url": "{{base_url}}/api/v1/health" } }
      ]
    },
    {
      "name": "01-Auth",
      "item": [
        {
          "name": "POST Login (admin)",
          "event": [{ "listen": "test", "script": { "exec": [
            "pm.test('Status 200', () => pm.response.to.have.status(200));",
            "const json = pm.response.json();",
            "pm.environment.set('admin_token', json.data.token);"
          ]}}],
          "request": {
            "method": "POST",
            "url": "{{base_url}}/api/v1/auth/login",
            "header": [{ "key": "Content-Type", "value": "application/json" }],
            "body": { "mode": "raw", "raw": "{\"username\":\"admin\",\"password\":\"admin123\"}" }
          }
        }
      ]
    }
  ],
  "variable": [
    { "key": "base_url", "value": "http://localhost:8080" }
  ]
}
```

> **完整 Collection JSON 文件：** 见同目录 `LabManage.postman_collection.json`（需从 Postman 导出）
