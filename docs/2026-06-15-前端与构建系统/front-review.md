# 前置审查报告 — 前端与构建系统

## 需求摘要
为现有 Go 后端系统新增前后端分离的 Web 前端应用，覆盖全部 22 个 REST API 接口，同时提供 Makefile 统一构建前后端、测试、部署等流程。

---

## 影响文件清单

### 前端项目（全部新建）

| 文件路径 | 修改类型 | 说明 |
|---------|---------|------|
| `frontend/` | 新增目录 | 前端项目根目录（完全独立于 backend/） |
| `frontend/package.json` | 新增 | React + Vite + TypeScript 依赖声明 |
| `frontend/vite.config.ts` | 新增 | Vite 构建配置（含 API 代理） |
| `frontend/tsconfig.json` | 新增 | TypeScript 编译配置 |
| `frontend/index.html` | 新增 | SPA 入口 HTML |
| `frontend/src/` | 新增目录 | 前端源码目录 |
| `frontend/src/main.tsx` | 新增 | React 应用入口 |
| `frontend/src/App.tsx` | 新增 | 根组件（路由 + 布局 + 权限守卫） |
| `frontend/src/api/` | 新增目录 | API 调用层 |
| `frontend/src/api/client.ts` | 新增 | Axios 实例（拦截器：Token注入/刷新/错误统一处理） |
| `frontend/src/api/auth.ts` | 新增 | 认证 API（login/logout/refresh） |
| `frontend/src/api/users.ts` | 新增 | 用户管理 API |
| `frontend/src/api/equipment.ts` | 新增 | 设备管理 API |
| `frontend/src/api/borrows.ts` | 新增 | 借阅管理 API |
| `frontend/src/api/types.ts` | 新增 | API 请求/响应 TypeScript 类型定义（与后端模型 1:1 映射） |
| `frontend/src/pages/` | 新增目录 | 页面组件 |
| `frontend/src/pages/Login.tsx` | 新增 | 登录页 |
| `frontend/src/pages/dashboard/` | 新增 | 仪表盘（首页） |
| `frontend/src/pages/equipment/` | 新增 | 设备大厅 / 设备详情 / 设备入库 / 设备编辑 |
| `frontend/src/pages/users/` | 新增 | 用户列表 / 用户创建 / 用户编辑 / 修改密码 |
| `frontend/src/pages/borrows/` | 新增 | 我的借阅 / 待审批 / 全部记录 / 申请借阅 |
| `frontend/src/components/` | 新增目录 | 通用组件 |
| `frontend/src/components/Layout/` | 新增 | 主布局（侧边栏 + 顶部导航 + 内容区） |
| `frontend/src/components/Pagination/` | 新增 | 分页组件 |
| `frontend/src/components/StatusBadge/` | 新增 | 状态标签（设备上下架/借阅状态） |
| `frontend/src/components/RoleGuard/` | 新增 | 角色权限守卫（按角色显示/隐藏） |
| `frontend/src/components/ErrorBoundary/` | 新增 | 错误边界 |
| `frontend/src/hooks/` | 新增目录 | 自定义 Hooks |
| `frontend/src/hooks/useAuth.ts` | 新增 | 认证状态管理（含 Token 持久化） |
| `frontend/src/hooks/usePagination.ts` | 新增 | 分页逻辑封装 |
| `frontend/src/hooks/usePermission.ts` | 新增 | 权限判断 Hook |
| `frontend/src/store/` | 新增目录 | 状态管理 |
| `frontend/src/store/auth.ts` | 新增 | 认证状态（Zustand） |
| `frontend/src/styles/` | 新增目录 | 样式 |
| `frontend/src/styles/global.css` | 新增 | 全局样式 + CSS 变量 |
| `frontend/src/styles/variables.css` | 新增 | 主题变量（色板/间距/圆角/阴影） |
| `frontend/public/` | 新增目录 | 静态资源 |

### Makefile（新建）

| 文件路径 | 修改类型 | 说明 |
|---------|---------|------|
| `Makefile` | 新增 | 项目根目录统一构建入口 |

### 后端（少量适配）

| 文件路径 | 修改类型 | 具体位置 | 说明 |
|---------|---------|---------|------|
| `backend/router/middleware/cors.go` | 修改 | L13 | CORS 改为可配置（生产环境禁止 `*`） |
| `backend/conf/config.example.yaml` | 修改 | — | 新增 `cors.allowed_origins` 字段 |

### 其他（新建）

| 文件路径 | 修改类型 | 说明 |
|---------|---------|------|
| `.env.example` | 新增 | 前端环境变量模板（VITE_API_BASE_URL） |
| `.env.development` | 新增 | 开发环境（代理到 localhost:8080） |
| `.env.production` | 新增 | 生产环境（直连后端或 Nginx） |
| `frontend/nginx.conf` | 新增 | 生产部署 Nginx 配置（SPA fallback + API 反代） |
| `docker-compose.yml` | 可选新增 | 如果后续需要容器化 |

---

## 技术选型

### 前端核心技术栈

| 维度 | 选择 | 理由 |
|------|------|------|
| 框架 | **React 18 + TypeScript** | 生态最成熟，适合管理后台；TypeScript 确保类型安全，与后端 Go 的静态类型文化匹配 |
| 构建工具 | **Vite 5** | 极速 HMR（<1s），原生 ESBuild 打包，开发体验碾压 Webpack |
| 路由 | **React Router v6** | 官方推荐，支持嵌套路由、懒加载、loader/action 模式 |
| HTTP 客户端 | **Axios** | 拦截器架构天然适配 Token 注入/刷新逻辑 |
| 状态管理 | **Zustand** | 轻量（<1KB），无 boilerplate，比 Redux 更适合中小型管理后台 |
| UI 组件库 | **Ant Design 5** | 企业级管理后台标配，Table/Form/Modal 开箱即用，中文生态好 |
| 图标 | **@ant-design/icons** | 与 Ant Design 风格统一 |
| CSS 方案 | **CSS Modules + Ant Design Token** | CSS Modules 无运行时开销；Ant Design Token 支持主题定制 |

### 为什么不选 Vue3？
- 该项目是高校实验室管理系统，后续可能对接校内其他系统。React 生态的 Ant Design Pro、UmiJS 等管理后台方案更丰富。
- Go + React 是常见全栈组合，社区资料丰富。

---

## 前端项目结构

```
frontend/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API 调用层
│   │   ├── client.ts           # Axios 实例 + 拦截器
│   │   ├── types.ts            # 请求/响应 TS 类型
│   │   ├── auth.ts             # 认证接口
│   │   ├── users.ts            # 用户管理接口
│   │   ├── equipment.ts        # 设备管理接口
│   │   └── borrows.ts          # 借阅管理接口
│   ├── components/             # 通用组件
│   │   ├── Layout/
│   │   │   ├── index.tsx       # 主布局（侧边栏 + TopBar + Outlet）
│   │   │   ├── Sidebar.tsx     # 侧边栏导航（根据角色显示菜单项）
│   │   │   └── TopBar.tsx      # 顶部栏（用户信息/退出）
│   │   ├── Pagination/
│   │   ├── StatusBadge/
│   │   ├── RoleGuard/          # 角色守卫（包裹需要特定角色的区域）
│   │   └── ErrorBoundary/
│   ├── pages/                  # 页面组件
│   │   ├── Login.tsx           # 登录页
│   │   ├── Dashboard.tsx       # 首页
│   │   ├── equipment/
│   │   │   ├── List.tsx        # 设备大厅（含搜索/分类筛选/库存过滤）
│   │   │   ├── Detail.tsx      # 设备详情
│   │   │   ├── Create.tsx      # 入库（管理员）
│   │   │   └── Edit.tsx        # 编辑（管理员）
│   │   ├── users/
│   │   │   ├── List.tsx        # 用户列表
│   │   │   ├── Create.tsx      # 创建用户
│   │   │   ├── Edit.tsx        # 编辑用户
│   │   │   └── ChangePassword.tsx
│   │   ├── borrows/
│   │   │   ├── MyRecords.tsx   # 我的借阅
│   │   │   ├── PendingList.tsx # 待审批
│   │   │   ├── AllRecords.tsx  # 全部记录（管理员）
│   │   │   └── Apply.tsx       # 发起申请
│   │   └── NotFound.tsx        # 404
│   ├── hooks/                  # 自定义 Hooks
│   │   ├── useAuth.ts          # 认证（login/logout/token检查）
│   │   ├── usePagination.ts    # 分页参数管理
│   │   └── usePermission.ts    # 角色权限判断
│   ├── store/                  # 状态
│   │   └── auth.ts             # 用户信息 + Token（Zustand + persist）
│   ├── styles/
│   │   ├── variables.css       # CSS 自定义属性（主题）
│   │   └── global.css          # 全局样式
│   ├── App.tsx                 # 路由配置 + 权限守卫
│   ├── main.tsx                # 入口
│   └── vite-env.d.ts           # Vite 类型声明
├── .env.development            # 开发环境变量
├── .env.production             # 生产环境变量
├── index.html                  # SPA 入口
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── nginx.conf                  # 生产部署配置
```

---

## Makefile 设计

### 构建流程编排

```
Makefile (根目录)
├── make build          # 构建前端（dist/）+ 编译后端（bin/server）
├── make build-frontend # 仅构建前端
├── make build-backend  # 仅编译后端
├── make dev            # 开发模式：同时启动后端 + 前端 dev server
├── make dev-backend    # 仅启动后端
├── make dev-frontend   # 仅启动前端 dev server
├── make test           # 运行全部测试（前端 + 后端）
├── make test-backend   # 后端测试（go test -race -cover）
├── make test-frontend  # 前端测试（vitest）
├── make lint           # 代码检查（go vet + eslint）
├── make clean          # 清理构建产物
├── make install        # 安装全部依赖
├── make deploy         # 生产构建 + 部署
├── make docker-build   # Docker 镜像构建
└── make help           # 显示帮助
```

### Makefile 变量设计

```makefile
# 可覆盖的变量
GO          ?= go
NPM         ?= npm
BACKEND_DIR  = backend
FRONTEND_DIR = frontend
BUILD_DIR    = dist
BINARY       = bin/server

# 版本注入
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS     := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'
```

---

## 依赖分析

### 前端 → 后端

```
frontend/src/api/client.ts ──HTTP──► backend (Gin Router)
       │                                  │
       ├─ /api/v1/auth/*                  ├─ Auth Controller
       ├─ /api/v1/users/*                 ├─ IAM Service
       ├─ /api/v1/roles                   ├─ User DAO / Role DAO
       ├─ /api/v1/equipments/*            ├─ Equipment Controller
       └─ /api/v1/borrows/*               └─ Borrow Controller
```

**协议约定**：
- 请求：JSON，Authorization: Bearer \<token\>
- 响应：统一格式 `{code: number, msg: string, data: T | null}`
- 分页：`{total, page, page_size, list}`
- 错误码：见 `backend/pkg/errcode/errcode.go`（0=成功，1xxx/2xxx/3xxx/5xxx）

### 开发代理链路

```
Browser (localhost:5173)
  │
  ├─ /api/* ──proxy──► Backend (localhost:8080)
  │                         │
  └─ 其他 ───Vite Dev Server──► React HMR
```

### 生产部署链路

```
Browser (HTTPS)
  │
  ▼
Nginx (:443)
  ├─ /api/* ──proxy──► Backend (:8080)
  └─ /* ──static──► frontend/dist/
```

### Makefile 协调

```
make dev
  ├─ 后台: (cd backend && go run cmd/main.go)
  └─ 前台: (cd frontend && npm run dev)

make build
  ├─ (cd backend && go build -ldflags="..." -o bin/server cmd/main.go)
  └─ (cd frontend && npm run build) → frontend/dist/

make deploy
  ├─ make build
  └─ (可选) 复制 dist/ + bin/server 到部署目标
```

---

## 循环依赖检查

- [x] **确认无新增循环依赖** — 前端是完全独立的项目，与后端零编译期依赖
- [x] Makefile 仅做流程编排，不引入编译期耦合
- [x] 前端 `api/types.ts` 独立定义类型，不依赖后端代码生成，避免工具链循环

---

## 风险预估

| 风险 | 概率 | 严重程度 | 说明 |
|------|------|---------|------|
| **CORS 策略在开发/生产不一致** | 中 | 高 | 当前 CORS 用 `*`，生产环境必须限制为具体 origin。修改时需确保开发代理也不受影响 |
| **Token 刷新竞态** | 中 | 中 | 多个请求同时 401 时，Axios 拦截器可能出现竞速刷新 Token。需在 `client.ts` 中加刷新锁（Promise 单例模式） |
| **Casbin 路由匹配变更** | 低 | 高 | 后端路由定义（`POST /users/:id/disable`）与 API 文档（`DELETE /users/:id`）不一致。前端必须以后端实际路由为准 |
| **Vite 开发代理路径冲突** | 低 | 中 | 如果前端路由包含 `/api/` 前缀，Vite 代理会误拦截。解决方案：前端路由避免使用 `/api/` 前缀 |
| **构建产物不一致** | 低 | 中 | Windows/Linux/macOS 三平台 Makefile 兼容性。Go 交叉编译天然支持；前端 `npm install` 依赖平台无关。需注意 Makefile 中路径分隔符 |
| **Ant Design 5 Tree Shaking** | 低 | 低 | Ant Design 5 默认支持 CSS-in-JS，无需额外配置 babel-plugin-import |

---

## 建议方案

### 前端实现路径（推荐 React + TypeScript + Vite + Ant Design）

**Phase 1 — 基础设施**
1. `npm create vite@latest frontend -- --template react-ts` 初始化
2. 安装依赖：`antd @ant-design/icons axios react-router-dom zustand`
3. 配置 `vite.config.ts`（API 代理、路径别名 @/）
4. 实现 `api/client.ts`（Axios 实例、请求拦截器注入 Token、响应拦截器处理 401 刷新）

**Phase 2 — 认证与布局**
5. 实现 `store/auth.ts`（Zustand + localStorage 持久化）
6. 实现 `pages/Login.tsx` + `hooks/useAuth.ts`
7. 实现 `components/Layout/`（侧边栏根据角色动态显示菜单）
8. 实现 `components/RoleGuard/` + 路由权限配置

**Phase 3 — 业务页面**
9. 设备管理：List / Detail / Create / Edit（含库存可视化）
10. 用户管理：List / Create / Edit / ChangePassword
11. 借阅管理：MyRecords / PendingList / AllRecords / Apply
12. 仪表盘首页

**Phase 4 — 构建与部署**
13. Makefile（build/dev/test/lint/clean/deploy）
14. `frontend/nginx.conf` 生产部署配置
15. CORS 中间件改造（可配置 allowed_origins）

### 关键设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| Token 存储 | localStorage | 简单可靠；管理后台非金融级安全，可接受 XSS 风险 |
| 路由懒加载 | React.lazy + Suspense | 减少首屏包体积 |
| 表单方案 | Ant Design Form | 与 UI 库统一，减少依赖 |
| 表格方案 | Ant Design Table | 内置分页/排序/筛选，完美适配后端分页接口 |
| 错误处理 | Axios 响应拦截器统一处理 | 根据 `code` 区分：2xxx → 跳登录，3xxx → Toast，5xxx → 错误页 |
| 权限控制 | 双重：路由级（beforeEnter）+ 组件级（RoleGuard） | 路由级防未授权访问，组件级控制按钮/操作显隐 |

### 后端适配改动（最小化）

只改一处：

**`backend/router/middleware/cors.go`**：
```go
// 改造前
c.Header("Access-Control-Allow-Origin", "*")

// 改造后
origin := c.GetHeader("Origin")
if isAllowedOrigin(origin) {
    c.Header("Access-Control-Allow-Origin", origin)
}
c.Header("Access-Control-Allow-Credentials", "true")
```

**`backend/conf/config.example.yaml`** 新增：
```yaml
cors:
  allowed_origins:
    - "http://localhost:5173"
    - "https://your-domain.com"
```
