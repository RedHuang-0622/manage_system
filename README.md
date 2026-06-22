# 实验室资产与人员管理系统

[![CI](https://github.com/RedHuang-0622/manage_system/actions/workflows/ci.yml/badge.svg)](https://github.com/RedHuang-0622/manage_system/actions/workflows/ci.yml)

基于 **Go + Gin** 的单体 REST API 后端系统，面向高校实验室的日常管理场景，涵盖**身份访问管理（IAM）**、**设备资产管理**、**借阅工单流转**三大核心模块。

## 技术栈

| 职责 | 技术 | 说明 |
|------|------|------|
| Web 框架 | Gin v1.9+ | 路由分发、中间件链 |
| ORM | GORM v2 | 连接池、事务、行锁 |
| 数据库 | MySQL 8.0+ | InnoDB 引擎 |
| 缓存 | Redis 7+ | 设备列表缓存、JWT 黑名单 |
| 认证 | JWT (golang-jwt v5) | Token 签发/解析/刷新/黑名单 |
| 鉴权 | Casbin v2 + GORM Adapter | RBAC，策略持久化到 MySQL |
| 配置 | Viper | config.yaml + 环境变量覆盖 |
| 日志 | Zap (uber-go) | 结构化 JSON，按天切割 |
| 校验 | go-playground/validator | 结构体 Tag 校验 |
| 密码 | bcrypt (cost=12) | 不可逆加密 |

## 系统架构

```
┌─────────────────────────────────────────────┐
│                   Router                     │
│   Recovery → RequestID → CORS → Logger      │
│              → Auth → Casbin                │
├──────────┬──────────┬───────────────────────┤
│ Auth     │ Equipment│ Borrow               │
│ Controller│Controller│Controller            │
├──────────┼──────────┼───────────────────────┤
│ Auth     │ Equipment│ Borrow               │
│ Service  │ Service  │ Service              │
├──────────┼──────────┼───────────────────────┤
│ User DAO │ Equip DAO│ Borrow DAO           │
│ Role DAO │          │                       │
├──────────┴──────────┴───────────────────────┤
│         MySQL · Redis · Casbin              │
└─────────────────────────────────────────────┘
```

### 角色体系

| 角色 | 权限 |
|------|------|
| `super_admin` | 全局管理（含用户创建、设备出入库） |
| `lab_admin` | 实验室管理（审批借阅、管理设备与成员） |
| `member` | 浏览设备、发起借阅、查看个人记录 |

## 快速开始

### 前置条件

- Go 1.23+
- MySQL 8.0+
- Redis 7+

### 1. 克隆项目

```bash
git clone git@github.com:RedHuang-0622/manage_system.git
cd manage_system
```

### 2. 配置数据库

```bash
# 创建 MySQL 数据库
mysql -u root -p -e "CREATE DATABASE lab_manage CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 3. 配置文件

```bash
# 复制模板并填入真实值
cp backend/conf/config.example.yaml backend/conf/config.yaml
```

修改 `backend/conf/config.yaml` 中的数据库连接信息和 JWT 密钥。

### 4. 启动服务

```bash
cd backend
go run cmd/main.go
```

服务启动于 `http://localhost:8080`，系统自动创建种子数据：

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `admin123` | super_admin |

### 5. 验证

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 查看设备列表
curl http://localhost:8080/api/v1/equipments \
  -H "Authorization: Bearer <token>"
```

## API 总览

> Base URL: `http://localhost:8080/api/v1`

### 认证接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/auth/login` | 用户登录 | 无 |
| POST | `/auth/logout` | 用户登出 | 登录用户 |
| POST | `/auth/refresh` | 刷新 Token | 登录用户 |

### 用户管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/roles` | 角色列表 | 登录用户 |
| POST | `/users` | 创建用户 | super_admin / lab_admin |
| GET | `/users` | 用户列表 | super_admin / lab_admin |
| GET | `/users/:id` | 用户详情 | super_admin / lab_admin |
| PUT | `/users/:id` | 更新用户 | super_admin / lab_admin |
| DELETE | `/users/:id` | 删除用户 | super_admin |
| PUT | `/users/:id/password` | 修改密码 | 本人或管理员 |

### 设备管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/equipments` | 设备入库 | super_admin / lab_admin |
| GET | `/equipments` | 设备列表（含缓存） | 登录用户 |
| GET | `/equipments/:id` | 设备详情 | 登录用户 |
| PUT | `/equipments/:id` | 更新设备 | super_admin / lab_admin |
| PUT | `/equipments/:id/status` | 设备状态变更 | super_admin / lab_admin |

### 借阅管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/borrows/apply` | 发起借阅申请 | member |
| GET | `/borrows/my` | 我的借阅记录 | member |
| GET | `/borrows/pending` | 待审批列表 | super_admin / lab_admin |
| POST | `/borrows/:id/approve` | 审批通过 | super_admin / lab_admin |
| POST | `/borrows/:id/reject` | 审批驳回 | super_admin / lab_admin |
| POST | `/borrows/:id/return` | 归还设备 | 借阅人 |
| POST | `/borrows/:id/cancel` | 取消申请 | 借阅人 |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 服务健康状态 |

## 项目结构

```
manage_system/
├── backend/
│   ├── cmd/main.go              # 入口：依赖注入、启动服务
│   ├── conf/
│   │   ├── config.example.yaml  # 配置模板
│   │   └── rbac_model.conf      # Casbin RBAC 模型
│   ├── router/
│   │   ├── router.go            # 路由注册
│   │   └── middleware/          # 6 层中间件链
│   ├── controller/              # 控制器层（参数校验、响应封装）
│   ├── service/                 # 业务逻辑层（事务、缓存、并发控制）
│   ├── dao/                     # 数据访问层（GORM 操作）
│   ├── models/                  # 数据实体
│   └── pkg/                     # 基础组件
│       ├── config/              # Viper 配置加载
│       ├── jwt/                 # JWT 服务
│       ├── redis/               # Redis 连接池
│       ├── zap/                 # 日志初始化
│       ├── errcode/             # 统一错误码
│       ├── response/            # 统一响应格式
│       └── safego/              # 并发安全工具
├── docs/
│   ├── api/                     # API 接口文档
│   ├── plan/                    # 模块设计文档
│   ├── design/                  # PRD 与视频脚本
│   ├── test_plan/               # 测试计划
│   └── cicd_plan/               # CI/CD 方案
├── test/                        # 测试套件
└── test_plan/                   # 测试计划文档
```

## 工程亮点

- **分层解耦**：Strict Router → Controller → Service → DAO 四层架构
- **双重鉴权**：JWT 身份认证 + Casbin RBAC 细粒度授权
- **并发安全**：借阅流程使用 GORM 事务 + MySQL 行锁，防止设备超卖
- **缓存策略**：设备列表 Redis 缓存，查询性能优化
- **优雅关闭**：信号监听 + Graceful Shutdown，确保请求处理完毕
- **统一错误码**：`code=0` 成功，`code>0` 区分业务异常与系统异常
- **防御式中间件**：Recovery + CORS + RateLimit + RequestID 全链路追踪

## 测试

```bash
cd test
go test ./... -v
```

覆盖 controller / service / dao / middleware / integration 各层测试，详见 `docs/test_plan/`。

## License

MIT
