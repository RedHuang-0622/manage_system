# 实现方案

## 设计目标

1. **前端零侵入后端**：前端完全独立项目，通过 HTTP REST 与后端通信，无需修改后端代码（除 CORS 配置增强）
2. **构建统一入口**：Makefile 作为唯一真相源，`make build` 一键产出前后端可部署产物
3. **类型安全**：TypeScript 类型定义与后端 Go 结构体 1:1 映射，编译期拦截接口不匹配
4. **权限可视化**：前端根据 JWT Claims 中的 role_name 动态控制 UI 显隐 + 路由守卫
5. **可维护性**：清晰分层（api → hooks → store → pages），新页面增量添加成本低

## 设计模式选择

| 模式 | 语言实现 | 应用位置 | 理由 |
|------|---------|---------|------|
| **Adapter** | Axios 封装 + `api/client.ts` | API 调用层 | 将 HTTP 请求适配为类型安全的 TS 函数，屏蔽底层传输 |
| **Strategy** | 角色路由守卫 `RoleGuard` | 权限控制 | 不同角色对应不同 UI 策略，运行时按 `role_name` 选择 |
| **Decorator** | Axios 拦截器链 | Token 注入 / 刷新 / 错误处理 | 非侵入式增强请求/响应，符合 Go 中间件链思维 |
| **Observer** | Zustand `subscribe` + `persist` | 认证状态同步 | 登录/登出时自动更新 UI 和 localStorage |
| **Module Pattern** | ES Module + barrel exports | 前端整体 | 每个 feature（auth/equipment/borrows）独立模块边界 |
| **Builder** | Makefile 变量覆盖 + `.env` 合并 | 构建系统 | `make build VERSION=v1.0` 灵活控制构建参数 |

---

## 方案 A：React + Vite + Ant Design 5 + Zustand（推荐）

### 核心思路

使用 React 18 生态构建 SPA 管理后台。Ant Design 5 提供企业级 UI 组件（Table/Form/Modal/Layout），Zustand 做轻量状态管理，Vite 做构建工具。前端完全独立于后端目录，通过 `vite.config.ts` 的 proxy 功能在开发时转发 `/api/*` 到后端。

### 架构设计

```
frontend/
├── src/
│   ├── api/              ← Adapter 层：每个后端模块一个 API 文件
│   │   ├── client.ts     ← Axios 实例 + 拦截器（Token注入/刷新/错误映射）
│   │   ├── types.ts      ← DTO：与后端 models/*.go 1:1 的 TS interface
│   │   ├── auth.ts
│   │   ├── users.ts
│   │   ├── equipment.ts
│   │   └── borrows.ts
│   ├── store/            ← Observer：全局状态（仅认证信息）
│   │   └── auth.ts       ← Zustand store + localStorage persist
│   ├── hooks/            ← 可复用逻辑
│   │   ├── useAuth.ts    ← 封装 login/logout/checkAuth
│   │   ├── usePagination.ts
│   │   └── usePermission.ts
│   ├── pages/            ← 每个页面 = 一个路由
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── equipment/{List,Detail,Create,Edit}.tsx
│   │   ├── users/{List,Create,Edit,ChangePassword}.tsx
│   │   └── borrows/{MyRecords,PendingList,AllRecords,Apply}.tsx
│   ├── components/       ← 可复用 UI 块
│   │   ├── Layout/       ← Ant Design Layout + 侧边栏角色菜单
│   │   ├── RoleGuard.tsx ← Strategy：按 role_name 条件渲染
│   │   └── ErrorBoundary.tsx
│   └── App.tsx           ← React Router v6 路由配置
```

### 关键接口契约（TypeScript ← → Go）

```typescript
// —— API 统一响应 (对应 backend/pkg/response/response.go) ——
interface ApiResponse<T> {
  code: number;     // 0=成功 | 1xxx=通用 | 2xxx=认证 | 3xxx=业务 | 5xxx=系统
  msg: string;
  data: T;
}

interface PageData<T> {
  total: number;
  page: number;
  page_size: number;
  list: T[];
}

// —— JWT Claims (对应 backend/pkg/jwt/jwt.go Claims) ——
interface UserInfo {
  user_id: number;
  username: string;
  role_id: number;
  role_name: "super_admin" | "lab_admin" | "member";
}

// —— 设备 (对应 backend/models/equipment.go LabEquipment) ——
interface Equipment {
  id: number;
  name: string;
  model: string;
  category: string;
  total_stock: number;
  available_stock: number;
  location: string;
  description: string;
  status: 0 | 1;  // 0=下架 1=上架
  created_at: string;
  updated_at: string;
}

// —— 借阅记录 (对应 backend/models/borrow.go BorrowRecord) ——
type BorrowStatus = "申请中" | "已借出" | "已归还" | "被拒绝";

interface BorrowRecord {
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
```

### 数据流

```
User Action → Page Component
    │ dispatch async thunk
    ▼
API Layer (api/*.ts)
    │ call client.get/post
    ▼
Axios Client (client.ts)
    │ intercept request → inject Authorization header
    │ intercept response → check code → refreshToken if 401 → retry
    ▼
Backend (Gin Router)
    │ Auth MW → Casbin MW → Controller → Service → DAO
    ▼
Response {code, msg, data}
```

### 路由设计 & 权限矩阵

```typescript
const routes = [
  { path: "/login",         element: <Login />,       auth: false },
  { path: "/",              element: <Dashboard />,    auth: true,  roles: ["*"] },
  { path: "/equipments",    element: <EquipList />,    auth: true,  roles: ["*"] },
  { path: "/equipments/:id",element: <EquipDetail />,  auth: true,  roles: ["*"] },
  { path: "/equipments/new",element: <EquipCreate />,  auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/equipments/:id/edit",element:<EquipEdit />,auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/users",         element: <UserList />,     auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/users/new",     element: <UserCreate />,   auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/users/:id/edit",element: <UserEdit />,     auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/borrows/my",    element: <MyRecords />,    auth: true,  roles: ["*"] },
  { path: "/borrows/pending",element:<PendingList />,  auth: true,  roles: ["super_admin","lab_admin"] },
  { path: "/borrows/all",   element: <AllRecords />,   auth: true,  roles: ["super_admin","lab_admin"] },
];
```

### Token 刷新并发安全

```typescript
// client.ts — 响应拦截器中的刷新锁
let isRefreshing = false;
let failedQueue: Array<{resolve: Function; reject: Function}> = [];

const processQueue = (error: unknown, token: string | null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) reject(error);
    else resolve(token);
  });
  failedQueue = [];
};

client.interceptors.response.use(
  (response) => {
    const { code } = response.data;
    if (code !== 0) {
      // 非 2004 错误直接抛
      if (code !== 2004) return Promise.reject(response.data);
      // Token 过期 → 尝试刷新
      if (!isRefreshing) {
        isRefreshing = true;
        return refreshToken().then((newToken) => {
          processQueue(null, newToken);
          isRefreshing = false;
          // 用新 Token 重试原请求
          response.config.headers.Authorization = `Bearer ${newToken}`;
          return client(response.config);
        }).catch((err) => {
          processQueue(err, null);
          isRefreshing = false;
          authStore.getState().logout();
          return Promise.reject(err);
        });
      }
      // 已在刷新中 → 排队等待
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then((token) => {
        response.config.headers.Authorization = `Bearer ${token}`;
        return client(response.config);
      });
    }
    return response.data; // 解包：直接返回 {code, msg, data}
  },
  (error) => Promise.reject(error)
);
```

### Makefile 设计

```makefile
.PHONY: build dev test lint clean deploy help

GO           := go
NPM          := npm
BACKEND_DIR  := backend
FRONTEND_DIR := frontend
BUILD_DIR    := dist
BIN_DIR      := bin
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME   := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS      := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

##@ Build
build: build-backend build-frontend

build-backend:
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/server ./cmd/main.go

build-frontend:
	cd $(FRONTEND_DIR) && $(NPM) ci --silent && $(NPM) run build

##@ Development
dev: dev-backend dev-frontend

dev-backend:
	cd $(BACKEND_DIR) && $(GO) run ./cmd/main.go &

dev-frontend:
	cd $(FRONTEND_DIR) && $(NPM) run dev

##@ Test
test: test-backend test-frontend

test-backend:
	cd $(BACKEND_DIR) && $(GO) vet ./...
	cd test && $(GO) test -race -count=3 -cover ./...

test-frontend:
	cd $(FRONTEND_DIR) && $(NPM) run lint && $(NPM) run test

##@ Utility
lint:
	cd $(BACKEND_DIR) && $(GO) vet ./...
	cd $(FRONTEND_DIR) && $(NPM) run lint

clean:
	rm -rf $(BACKEND_DIR)/$(BIN_DIR)
	rm -rf $(FRONTEND_DIR)/$(BUILD_DIR)

install:
	cd $(FRONTEND_DIR) && $(NPM) ci

deploy: build
	@echo "部署产物: $(BACKEND_DIR)/$(BIN_DIR)/server + $(FRONTEND_DIR)/$(BUILD_DIR)/"

##@ Docker
docker-build:
	docker build -t lab-system:$(VERSION) -f Dockerfile .

##@ Help
help:
	@echo "可用目标: build dev test lint clean install deploy docker-build"
```

### 变更范围

| 类别 | 文件数 | 说明 |
|------|-------|------|
| 新建 `frontend/` | ~40 文件 | 完整 React SPA |
| 新建 `Makefile` | 1 文件 | 根目录 |
| 修改后端 | 1 文件 | `backend/router/middleware/cors.go` — CORS 可配置化 |
| 新增配置字段 | 1 处 | `backend/conf/config.example.yaml` — `cors.allowed_origins` |

---

## 方案 B：Vue 3 + Vite + Element Plus + Pinia

### 核心思路

使用 Vue 3 Composition API + Element Plus（与 Ant Design 同级别的 Vue 生态 UI 库）构建。Pinia 做状态管理（Vue 官方推荐替代 Vuex）。整体架构与方案 A 对称，但采用 Vue 生态。

### 关键差异

| 维度 | 方案 A (React) | 方案 B (Vue 3) |
|------|---------------|----------------|
| UI 库 | Ant Design 5 | Element Plus |
| 状态管理 | Zustand | Pinia |
| 路由 | React Router v6 | Vue Router 4 |
| HTTP | Axios（相同） | Axios（相同） |
| 模板语法 | JSX | SFC (.vue) |
| 权限守卫 | `<RoleGuard>` 组件 | `v-permission` 自定义指令 |
| 响应式 | Hooks | `ref()` / `reactive()` |

### 架构设计

```
frontend/
├── src/
│   ├── api/              ← Axios 封装（与方案 A 完全复用）
│   │   ├── client.ts     ← 相同拦截器逻辑
│   │   └── ...
│   ├── stores/           ← Pinia stores
│   │   └── auth.ts       ← defineStore + persist plugin
│   ├── composables/      ← Vue Composables（等价 React Hooks）
│   │   ├── useAuth.ts
│   │   ├── usePagination.ts
│   │   └── usePermission.ts
│   ├── views/            ← 页面组件（SFC）
│   ├── components/       ← 通用组件
│   ├── directives/       ← 自定义指令
│   │   └── permission.ts ← v-permission="['super_admin']"
│   └── router/index.ts   ← Vue Router 配置 + 导航守卫
```

### 优势 vs 方案 A

- SFC 模板语法对后端转前端的开发者更直观
- Element Plus 中文文档质量高
- `v-permission` 指令比 React 的 `<RoleGuard>` 包裹更简洁

### 劣势 vs 方案 A

- TypeScript 支持弱于 React（Vue SFC 中 TS 推导有边界）
- 生态规模小于 React（特别是管理后台模板/Pro 版）
- 导师/学生群体更大概率有 React 基础（前端课程多用 React）

---

## 方案 C：Go 模板渲染 + HTMX（不同范式） 

### 核心思路

**不走 SPA 路线**。利用 Go 的 `html/template` 做服务端渲染，配合 HTMX（~14KB 无依赖 JS 库）实现局部刷新。前端是 Go 模板 + 少量 HTMX 属性，没有 npm/node 依赖。

### 架构设计

```
backend/
├── cmd/main.go
├── ...
├── templates/              ← Go html/template 文件
│   ├── base.html           ← 母版（Layout）
│   ├── login.html
│   ├── dashboard.html
│   ├── equipment/
│   │   ├── list.html
│   │   ├── detail.html
│   │   └── form.html       ← 创建/编辑共用
│   ├── users/
│   │   ├── list.html
│   │   └── form.html
│   └── borrows/
│       ├── my.html
│       ├── pending.html
│       └── all.html
├── static/                 ← 静态资源
│   ├── css/
│   │   └── app.css         ← 手写 CSS（或 Tailwind CDN）
│   ├── js/
│   │   └── htmx.min.js     ← HTMX 库（单文件，~14KB）
│   └── favicon.ico
```

### 关键差异

| 维度 | 方案 A (React SPA) | 方案 C (Go + HTMX) |
|------|-------------------|-------------------|
| 前端运行时 | React (JavaScript) | Go template + HTMX |
| 构建工具 | Vite + npm | 无（Go 自带构建） |
| 状态管理 | Zustand (client-side) | 服务端 Session |
| 路由 | React Router (client-side) | Gin Router (server-side) |
| UI 组件 | Ant Design | 手写 HTML + CSS（或 Tailwind CDN） |
| 前后端耦合 | 完全解耦（HTTP API） | 高耦合（模板嵌入后端） |
| npm 依赖 | 有（~200MB） | 无 |
| 部署复杂度 | Nginx + 静态文件 + API 反代 | 单一 Go binary（内嵌模板） |

### 优势 vs 方案 A

- **零前端工具链**：不需要 Node.js / npm / Vite，Go 开发者即可维护
- **单一部署产物**：一个 Go binary 包含后端 + 前端模板
- **SEO 友好**：服务端渲染原生支持
- **包体积小**：HTMX 14KB vs React + Ant Design ~200KB+

### 劣势 vs 方案 A

- **交互体验降级**：HTMX 局部刷新不如 React SPA 即时
- **UI 开发效率低**：无组件库，手写 HTML/CSS 工作量大
- **无法热更新**：改模板需重启后端（或配置热重载）
- **生态孤岛**：HTMX 社区小，复杂交互（拖拽、图表）需额外处理
- **与现有架构冲突**：当前后端已是纯 API 架构，引入模板违反分层

---

## 方案定性对比

| 维度 | 方案 A: React SPA | 方案 B: Vue 3 SPA | 方案 C: Go + HTMX |
|------|-------------------|-------------------|-------------------|
| **耦合度** | 低 — 前后端仅 HTTP 协议耦合 | 低 — 同方案 A | 高 — 模板嵌入后端，同进程部署 |
| **内聚性** | 高 — 前端独立模块边界清晰 | 高 — 同方案 A | 中 — 模板与 Go 代码混合 |
| **可测试性** | 高 — 前端可独立 vitest 测试 | 高 — 同方案 A | 低 — 需 Go 集成测试覆盖模板 |
| **实现成本** | 中 — Ant Design 开箱即用，~40 文件 | 中 — Element Plus 开箱即用，~40 文件 | 高 — 手写 UI，需转换后端路由结构 |
| **改动面** | 极小 — 仅 CORS 1 处后端改动 | 极小 — 同方案 A | 大 — 后端新增 templates/ 目录 + 路由 + static 服务 |
| **可回滚性** | 极好 — 删 frontend/ + Makefile 即回滚 | 极好 — 同方案 A | 差 — 后端改动多，回滚需 git revert |
| **团队适配** | ⭐⭐⭐ Go + React 是全栈常见组合 | ⭐⭐ Vue 入门曲线平但生态较小 | ⭐ Go 开发者友好但不符合 SPA 趋势 |
| **可扩展性** | ⭐⭐⭐ 易扩展新页面/WebSocket/移动端 | ⭐⭐⭐ 同方案 A | ⭐ 扩展需修改 Go 路由 + 模板 |
| **生产成熟度** | ⭐⭐⭐ Ant Design Pro 企业级验证 | ⭐⭐ Element Plus Admin 可用 | ⭐ HTMX 生产案例少 |
| **Token 刷新** | Axios 拦截器 + 刷新锁 | 同方案 A | 无需（Session 机制） |

---

## 推荐：方案 A — React + Vite + Ant Design 5 + Zustand

### 推荐理由

1. **最小改动面**：前后端完全解耦，后端仅改 CORS 一处（从 `*` 到可配置），风险极低
2. **最佳可回滚性**：`rm -rf frontend/ Makefile` + `git checkout backend/router/middleware/cors.go` 即回滚
3. **最优生态匹配**：React + Ant Design 在管理后台场景有 Ant Design Pro 做参考
4. **Go 文化契合**：TypeScript 静态类型 + ES Module 显式导入，与 Go 的显式依赖文化一致
5. **可测试性**：前端 vitest + 后端 go test 独立运行，互不干扰
6. **未来扩展**：保留 WebSocket / 移动端 / 微前端扩展可能

### 最大风险

| 风险 | 缓解措施 |
|------|---------|
| **Token 刷新并发竞态** | client.ts 中实现 Promise 单例刷新锁 + 失败队列（已在方案中设计） |
| **Ant Design 5 bundle 体积** | Vite tree-shaking + 路由懒加载 + `antd/es` 按需导入 |
| **Windows 下 Makefile 兼容** | 使用纯 shell 命令（cd / rm / mkdir），依赖 Git Bash 环境（项目已配置） |

---

## 循环依赖检查

- [x] `api/client.ts` → `store/auth.ts`（读取 Token 注入请求） — ✅ 单向，auth store 不依赖 client
- [x] `api/*.ts` → `api/client.ts`（使用 Axios 实例） — ✅ 单向
- [x] `pages/*.tsx` → `api/*.ts` + `hooks/*.ts` — ✅ 单向
- [x] `components/Layout` → `store/auth.ts`（读取角色） — ✅ 单向
- [x] `Makefile` → `backend/` + `frontend/` — ✅ Makefile 仅调用子目录命令，不编译期依赖
- [x] `frontend/` ↔ `backend/` — ✅ 零编译期依赖，仅运行时 HTTP

**结论：零循环依赖。** 前端完全是独立项目。

---

## CORS 中间件改造（唯一后端改动）

### 当前代码 `backend/router/middleware/cors.go`

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        // ...
    }
}
```

### 改造后

```go
type CORSConfig struct {
    AllowedOrigins []string
}

func CORS(cfg *CORSConfig) gin.HandlerFunc {
    allowed := make(map[string]bool)
    for _, o := range cfg.AllowedOrigins {
        allowed[o] = true
    }

    // 如果未配置，开发模式允许 localhost
    if len(allowed) == 0 {
        allowed["http://localhost:5173"] = true
        allowed["http://localhost:3000"] = true
    }

    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        if allowed[origin] || allowed["*"] {
            c.Header("Access-Control-Allow-Origin", origin)
        }
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
        c.Header("Access-Control-Expose-Headers", "X-Request-ID")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    }
}
```

**调用方变更** `backend/router/router.go`：
```go
// 改造前
r.Use(middleware.CORS())

// 改造后
r.Use(middleware.CORS(&middleware.CORSConfig{
    AllowedOrigins: cfg.CORS.AllowedOrigins, // 来自 config.yaml
}))
```

---

## 实现步骤

| # | 步骤 | 涉及文件 | 设计模式 | 预计变更 |
|---|------|---------|---------|---------|
| 1 | 后端 CORS 可配置化 | `middleware/cors.go`, `router/router.go`, `config/config.go` | Strategy | +30 -5 行 |
| 2 | Vite 项目初始化 | `frontend/` 骨架文件 | — | 脚手架生成 |
| 3 | API 类型 + Axios 客户端 | `api/types.ts`, `api/client.ts` | Adapter + Decorator | ~200 行 |
| 4 | 认证模块 | `store/auth.ts`, `hooks/useAuth.ts`, `pages/Login.tsx` | Observer | ~150 行 |
| 5 | 布局 + 路由守卫 | `components/Layout/`, `App.tsx` | Strategy | ~200 行 |
| 6 | API 函数（22 个接口） | `api/auth.ts`, `api/users.ts`, `api/equipment.ts`, `api/borrows.ts` | Adapter | ~300 行 |
| 7 | 设备管理页面 | `pages/equipment/*.tsx` | — | ~400 行 |
| 8 | 用户管理页面 | `pages/users/*.tsx` | — | ~350 行 |
| 9 | 借阅管理页面 | `pages/borrows/*.tsx` | — | ~450 行 |
| 10 | 仪表盘首页 | `pages/Dashboard.tsx` | — | ~100 行 |
| 11 | Makefile | `Makefile` | Builder | ~80 行 |
| 12 | Nginx 配置 + 环境变量 | `nginx.conf`, `.env.*` | — | ~50 行 |

**总预计变更**：~2300 行新增代码 + ~35 行后端修改

## 测试策略

| 层 | 方案 | 覆盖率目标 |
|----|------|----------|
| API 类型 | TypeScript 编译（零错误） | — |
| API 函数 | vitest + MSW mock（模拟 22 个接口响应） | ≥ 80% |
| 认证流程 | E2E: 登录 → 获取 Token → 自动刷新 → 登出 | 关键路径 |
| Hooks | vitest + @testing-library/react-hooks | ≥ 80% |
| 页面组件 | @testing-library/react 快照测试 | 关键页面 |
| Makefile | `make build` / `make clean` / `make help` 手动验证 | 全目标 |
| 后端 CORS | go test — 验证 OPTIONS 预检 + Origin 限制 | 新增用例 |

## 回滚方案

- **紧急回滚**：`make build` 产物替换为旧版 binary + 无前端（回退到纯 API 模式）
- **代码回滚**：`git revert` 最后一个 commit
- **部分回滚**：删 `frontend/` 目录即回退到纯后端模式，后端 API 照常工作
