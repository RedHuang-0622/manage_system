# 01 — IAM 模块测试（身份与访问管理）

## 1 测试范围

覆盖三个子域：

| 子域 | DAO 层 | Service 层 | Middleware | Controller |
|------|--------|-----------|------------|------------|
| 认证 (auth) | — | 登录逻辑、Token 签发 | Auth 中间件 | 请求绑定、响应格式 |
| 用户管理 (user) | CRUD、分页、唯一性 | 密码加密、字段校验 | — | 参数校验、权限检查 |
| 权限控制 (casbin) | 策略 CRUD | 策略初始化 | Casbin 中间件 | — |

## 2 DAO 层测试

### 2.1 UserDAO

| 编号 | 测试用例 | 输入 | 预期 | 级别 |
|------|---------|------|------|------|
| U-DAO-01 | Create — 正常创建用户 | 完整 SysUser 对象 | 无 error，ID 自增回填 | P0 |
| U-DAO-02 | Create — 用户名重复 | 相同 username 两次 | 第二次返回唯一约束错误 | P0 |
| U-DAO-03 | FindByUsername — 存在 | "zhangsan" | 返回完整用户对象 | P0 |
| U-DAO-04 | FindByUsername — 不存在 | "nonexist" | 返回 ErrRecordNotFound | P1 |
| U-DAO-05 | FindByID — 存在 | id=1 | 返回用户 + Preload Role | P0 |
| U-DAO-06 | FindByID — 不存在 | id=9999 | 返回 ErrRecordNotFound | P1 |
| U-DAO-07 | FindPage — 无筛选 | page=1, size=10 | 返回全部用户 + 总数 | P0 |
| U-DAO-08 | FindPage — keyword 搜索 | keyword="张" | 返回 username/real_name 含"张"的用户 | P1 |
| U-DAO-09 | FindPage — status 筛选 | status=1 | 只返回启用的用户 | P1 |
| U-DAO-10 | FindPage — role_id 筛选 | role_id=3 | 只返回 member 角色用户 | P1 |
| U-DAO-11 | FindPage — 空结果 | keyword="zzzznotfound" | total=0, list=[] | P2 |
| U-DAO-12 | FindPage — 分页边界 | page=1, size=1, 共 3 条 | total=3, len(list)=1 | P2 |
| U-DAO-13 | UpdateFields — 更新 email | id=1, {email: "new@test.com"} | 无 error，数据库值已改变 | P0 |
| U-DAO-14 | UpdateFields — 更新 status 禁用 | id=1, {status: 0} | 无 error | P1 |
| U-DAO-15 | UpdateFields — 用户不存在 | id=9999, {email: "x"} | 返回 ErrRecordNotFound | P2 |

### 2.2 RoleDAO

| 编号 | 测试用例 | 输入 | 预期 | 级别 |
|------|---------|------|------|------|
| R-DAO-01 | FindAll — 返回全部角色 | — | 返回 3 个角色 | P0 |
| R-DAO-02 | FindByID — 存在 | id=1 | 返回 super_admin 角色 | P1 |
| R-DAO-03 | FindByID — 不存在 | id=999 | 返回 ErrRecordNotFound | P2 |
| R-DAO-04 | FindByName — 存在 | "member" | 返回 member 角色 | P1 |

## 3 Service 层测试

### 3.1 认证 (AuthService)

| 编号 | 测试用例 | 输入 | Mock 行为 | 预期 |
|------|---------|------|----------|------|
| A-SVC-01 | Login — 正常登录 | {username:"admin", password:"admin123"} | DAO 返回有效用户，bcrypt 匹配成功，JWT 签发成功 | 返回 Token + expires_in |
| A-SVC-02 | Login — 用户不存在 | {username:"nobody", password:"x"} | DAO 返回 ErrRecordNotFound | 返回 ErrAuthFailed |
| A-SVC-03 | Login — 密码错误 | {username:"admin", password:"wrong"} | DAO 返回用户，bcrypt 比较失败 | 返回 ErrAuthFailed |
| A-SVC-04 | Login — 账号已禁用 | {username:"disabled", password:"x"} | DAO 返回 status=0 的用户 | 返回 ErrAccountDisabled |
| A-SVC-05 | Logout — 正常登出 | 有效 Token | Redis mock 写入成功 | 无 error |
| A-SVC-06 | Logout — Token 已在黑名单 | 已登出过的 Token | Redis mock 已存在 | 幂等，无 error |
| A-SVC-07 | Refresh — 正常刷新 | 有效旧 Token | Parse 成功，黑名单无记录 | 返回新 Token，旧 Token 入黑名单 |
| A-SVC-08 | Refresh — 旧 Token 过期 | 过期 Token | Parse 返回过期错误 | 返回 ErrTokenInvalid |
| A-SVC-09 | Refresh — 旧 Token 在黑名单 | 已登出的 Token | 黑名单检查返回 true | 返回 ErrTokenInvalid |
| A-SVC-10 | GenerateToken — 签发成功 | userID=1, role=super_admin | — | Token 可解析，Claims 正确 |
| A-SVC-11 | GenerateToken — 过期时间计算 | expire=7200 | — | exp - iat ≈ 7200s |

### 3.2 用户管理 (UserService)

| 编号 | 测试用例 | 输入 | Mock 行为 | 预期 |
|------|---------|------|----------|------|
| U-SVC-01 | Create — 正常创建 | 完整 CreateUserReq | DAO 查 username 不存在，Insert 成功 | 返回 SysUser，password_hash 不为空 |
| U-SVC-02 | Create — 用户名冲突 | username="admin" | DAO FindByUsername 返回已有用户 | 返回 ErrUserExists |
| U-SVC-03 | Create — bcrypt 加密验证 | password="test123" | — | password_hash != "test123"，bcrypt 可以验证 |
| U-SVC-04 | ListPage — 正常分页 | page=1, size=10 | DAO 返回 50 条 | 返回 PageResult{total:50} |
| U-SVC-05 | GetByID — 存在 | id=1 | DAO 返回用户 | 返回用户，含 Role 关联 |
| U-SVC-06 | GetByID — 不存在 | id=9999 | DAO 返回 ErrRecordNotFound | 返回 ErrUserNotFound |
| U-SVC-07 | Update — 部分字段更新 | {email:"new@test.com"} | 用户存在，UpdateFields 成功 | 只更新 email 字段 |
| U-SVC-08 | Update — 用户不存在 | id=9999 | DAO FindByID 失败 | 返回 ErrUserNotFound |
| U-SVC-09 | Disable — 正常 | id=3, operatorID=1 | 用户存在，无未归还借阅 | status 改为 0 |
| U-SVC-10 | Disable — 禁止禁用自己 | id=1, operatorID=1 | — | 返回错误 |
| U-SVC-11 | Disable — 有未归还借阅 | id=3 | borrowDAO.CountActive 返回 >0 | 返回错误 |
| U-SVC-12 | ChangePassword — 正常 | old:"old", new:"new" | bcrypt 验证旧密码成功 | password_hash 已更新 |
| U-SVC-13 | ChangePassword — 旧密码错误 | old:"wrong", new:"new" | bcrypt 比较失败 | 返回 ErrAuthFailed |
| U-SVC-14 | ChangePassword — 管理员免旧密码 | isAdmin=true | — | 跳过旧密码验证，直接更新 |

## 4 Middleware 测试

### 4.1 Auth 中间件

| 编号 | 测试用例 | 请求 Header | 预期 |
|------|---------|------------|------|
| M-AUTH-01 | 正常 Token | Authorization: Bearer <valid_token> | c.Next() 被调用，Context 含 user_id/role_name |
| M-AUTH-02 | 缺少 Header | 无 Authorization | 401 {code:2003} |
| M-AUTH-03 | 格式错误 | Authorization: Basic xxx | 401 {code:2003} |
| M-AUTH-04 | Token 在黑名单 | Bearer <blacklisted_token> | 401 {code:2004} |
| M-AUTH-05 | Token 过期 | Bearer <expired_token> | 401 {code:2004} |
| M-AUTH-06 | Token 签名错误 | Bearer <tampered_token> | 401 {code:2004} |
| M-AUTH-07 | 白名单路径跳过 | GET /api/v1/health | 不经过 Auth，直接 200 |
| M-AUTH-08 | 白名单路径跳过 | POST /api/v1/auth/login | 不经过 Auth |

### 4.2 Casbin 中间件

| 编号 | 测试用例 | 角色 | 路径 | 方法 | 预期 |
|------|---------|------|------|------|------|
| M-CAS-01 | member 浏览设备 | member | /api/v1/equipments | GET | 通过 |
| M-CAS-02 | member 发起借阅 | member | /api/v1/borrows/apply | POST | 通过 |
| M-CAS-03 | member 越权审批 | member | /api/v1/borrows/1/approve | POST | 403 {code:2005} + Warn 日志 |
| M-CAS-04 | member 越权创建用户 | member | /api/v1/users | POST | 403 {code:2005} |
| M-CAS-05 | lab_admin 管理用户 | lab_admin | /api/v1/users | POST | 通过 |
| M-CAS-06 | lab_admin 审批借阅 | lab_admin | /api/v1/borrows/1/approve | POST | 通过 |
| M-CAS-07 | super_admin 全权限 | super_admin | /api/v1/* | ANY | 通过 |
| M-CAS-08 | 未登录（无 role_name） | — | /api/v1/equipments | GET | 403 {code:2005} |

## 5 Controller 层测试

### 5.1 Auth Controller

| 编号 | 测试用例 | 请求 | 预期 |
|------|---------|------|------|
| C-AUTH-01 | POST /auth/login — 正常 | {username, password} 有效 | 200 {code:0, data:{token, expires_in}} |
| C-AUTH-02 | POST /auth/login — 参数缺失 | {} 空 Body | 400 {code:1001} 参数校验失败 |
| C-AUTH-03 | POST /auth/login — username 过短 | {username:"ab"} | 400 {code:1001} |
| C-AUTH-04 | POST /auth/logout — 正常 | 有效 Token | 200 {code:0} |
| C-AUTH-05 | POST /auth/refresh — 正常 | {token: valid_old} | 200 {code:0, data:{token: new}} |

### 5.2 User Controller

| 编号 | 测试用例 | 请求 | 预期 |
|------|---------|------|------|
| C-USER-01 | POST /users — 正常 | 完整用户 JSON | 201 {code:0, data:{id, username, ...}} |
| C-USER-02 | POST /users — 缺少必填项 | {username: "test"} 无 password | 400 {code:1001} |
| C-USER-03 | POST /users — email 格式错误 | {email: "not-email"} | 400 {code:1001} |
| C-USER-04 | GET /users — 默认分页 | GET /users | 200 {code:0, data:{total, list}} |
| C-USER-05 | GET /users — 自定义分页 | GET /users?page=2&page_size=5 | 200，len(list)≤5 |
| C-USER-06 | GET /users/:id — 存在 | GET /users/1 | 200 {code:0, data:{...}} |
| C-USER-07 | GET /users/:id — 不存在 | GET /users/9999 | 404 {code:3002} |
| C-USER-08 | PUT /users/:id — 部分更新 | {email: "new@test.com"} | 200 {code:0} |
| C-USER-09 | DELETE /users/:id — 正常 | member 角色调用 | 403 {code:2005} |
| C-USER-10 | PUT /users/:id/password — 正常 | {old, new} | 200 {code:0} |

## 6 pkg/jwt 包测试

| 编号 | 测试用例 | 输入 | 预期 |
|------|---------|------|------|
| J-PKG-01 | GenerateToken + ParseToken | userID=1, role=super_admin | 解析出的 Claims 与输入一致 |
| J-PKG-02 | ParseToken — 过期 Token | expire=-1s 的 Token | 返回解析错误 |
| J-PKG-03 | ParseToken — 错误密钥 | 用不同 secret 签名的 Token | 返回签名验证错误 |
| J-PKG-04 | AddToBlacklist + IsInBlacklist | 正常 Token | 加入后 IsInBlacklist 返回 true |
| J-PKG-05 | IsInBlacklist — 未加入 | 新 Token | 返回 false |
| J-PKG-06 | 黑名单 Key — 多次登出 | 同一 Token 两次 AddToBlacklist | 幂等，不报错 |
| J-PKG-07 | 黑名单 TTL — 过期自动清除 | 设置 TTL=1s，等待 2s | IsInBlacklist 返回 false |
