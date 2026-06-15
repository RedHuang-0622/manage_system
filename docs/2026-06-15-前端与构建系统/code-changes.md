# 代码变更摘要

## 新增/修改文件

| 文件 | 类型 | 说明 | 设计模式 |
|------|------|------|---------|
| `backend/router/middleware/cors.go` | 修改 | CORS 中间件支持可配置 allowed_origins | Strategy |
| `backend/pkg/config/config.go` | 修改 | 新增 CORSConfig 结构体 | — |
| `backend/router/router.go` | 修改 | Dependencies 新增 CORSAllowedOrigins | — |
| `backend/cmd/main.go` | 修改 | 传递 cfg.CORS.AllowedOrigins | — |
| `backend/conf/config.example.yaml` | 修改 | 新增 cors.allowed_origins 配置 | — |
| `Makefile` | 新增 | 统一构建入口（build/dev/test/lint/clean/deploy） | Builder |
| `frontend/package.json` | 新增 | 依赖声明 | — |
| `frontend/vite.config.ts` | 新增 | Vite 构建配置 + API 代理 | — |
| `frontend/tsconfig.json` | 新增 | TypeScript 编译配置 | — |
| `frontend/tsconfig.node.json` | 新增 | Vite 配置文件的 TS 编译 | — |
| `frontend/index.html` | 新增 | SPA 入口 | — |
| `frontend/src/main.tsx` | 新增 | React 入口 + Ant Design ConfigProvider | — |
| `frontend/src/App.tsx` | 新增 | 路由配置 + 权限守卫 + 懒加载 | Strategy |
| `frontend/src/api/types.ts` | 新增 | 22 个 API 的 TS 类型定义（与 Go struct 1:1） | — |
| `frontend/src/api/client.ts` | 新增 | Axios 实例 + Token 注入 + 刷新锁 | Adapter + Decorator |
| `frontend/src/api/auth.ts` | 新增 | 认证 API（login/logout/refresh/roles） | Adapter |
| `frontend/src/api/users.ts` | 新增 | 用户管理 API（6 个接口） | Adapter |
| `frontend/src/api/equipment.ts` | 新增 | 设备管理 API（5 个接口） | Adapter |
| `frontend/src/api/borrows.ts` | 新增 | 借阅管理 API（7 个接口） | Adapter |
| `frontend/src/store/auth.ts` | 新增 | Zustand 认证状态 + localStorage 持久化 | Observer |
| `frontend/src/hooks/useAuth.ts` | 新增 | 登录/登出/Token 检查封装 | — |
| `frontend/src/hooks/usePermission.ts` | 新增 | 角色权限判断 Hook | — |
| `frontend/src/hooks/usePagination.ts` | 新增 | 分页逻辑 + Ant Table 适配 | — |
| `frontend/src/pages/Login.tsx` | 新增 | 登录页面 | — |
| `frontend/src/pages/Dashboard.tsx` | 新增 | 仪表盘首页 | — |
| `frontend/src/pages/NotFound.tsx` | 新增 | 404 页面 | — |
| `frontend/src/pages/equipment/List.tsx` | 新增 | 设备大厅（分页/搜索/筛选） | — |
| `frontend/src/pages/equipment/Detail.tsx` | 新增 | 设备详情 | — |
| `frontend/src/pages/equipment/Create.tsx` | 新增 | 设备入库 | — |
| `frontend/src/pages/equipment/Edit.tsx` | 新增 | 设备编辑 + 下架 | — |
| `frontend/src/pages/users/List.tsx` | 新增 | 用户列表 + 禁用 | — |
| `frontend/src/pages/users/Create.tsx` | 新增 | 创建用户 | — |
| `frontend/src/pages/users/Edit.tsx` | 新增 | 编辑用户 | — |
| `frontend/src/pages/users/ChangePassword.tsx` | 新增 | 修改密码 | — |
| `frontend/src/pages/borrows/MyRecords.tsx` | 新增 | 我的借阅 + 取消 | — |
| `frontend/src/pages/borrows/PendingList.tsx` | 新增 | 待审批列表 + 审批 | — |
| `frontend/src/pages/borrows/AllRecords.tsx` | 新增 | 全部记录 + 归还 | — |
| `frontend/src/pages/borrows/Apply.tsx` | 新增 | 发起借阅申请 | — |
| `frontend/src/components/Layout/index.tsx` | 新增 | 主布局（Sider + Content） | — |
| `frontend/src/components/Layout/Sidebar.tsx` | 新增 | 侧边栏（角色菜单） | Strategy |
| `frontend/src/components/Layout/TopBar.tsx` | 新增 | 顶部栏（用户信息/退出/改密） | — |
| `frontend/src/components/RoleGuard.tsx` | 新增 | 路由级角色守卫 | Strategy |
| `frontend/src/components/StatusBadge.tsx` | 新增 | 状态标签（设备/用户/借阅） | — |
| `frontend/src/components/ErrorBoundary.tsx` | 新增 | React 错误边界 | — |
| `frontend/src/components/Pagination/index.tsx` | 新增 | 分页组件封装 | — |
| `frontend/src/styles/variables.css` | 新增 | CSS 主题变量 | — |
| `frontend/src/styles/global.css` | 新增 | 全局样式 | — |
| `frontend/nginx.conf` | 新增 | 生产 Nginx 配置（SPA + API 反代） | — |
| `frontend/.env.development` | 新增 | 开发环境变量 | — |
| `frontend/.env.production` | 新增 | 生产环境变量 | — |
| `frontend/public/favicon.svg` | 新增 | 网站图标 | — |
| `.gitignore` | 修改 | 新增 frontend/ 产物忽略 | — |

**总计：49 个文件，+9864 行**

## API 变更
无 API 变更。后端所有 22 个接口行为不变。

## 设计模式使用

| 模式 | 文件 | 效果 |
|------|------|------|
| **Adapter** | `api/client.ts`, `api/*.ts` | 将 HTTP REST 适配为类型安全 TS 函数 |
| **Strategy** | `RoleGuard.tsx`, `Sidebar.tsx`, `cors.go` | 运行时按角色/配置选择行为 |
| **Decorator** | `api/client.ts` Axios 拦截器 | 非侵入式 Token 注入/刷新/错误处理 |
| **Observer** | `store/auth.ts` Zustand persist | 登录/登出自动同步 UI + localStorage |
| **Builder** | `Makefile` 变量覆盖 | `make build VERSION=v1.0` 灵活控制 |

## 接口抽象

| 接口 | 实现方 | 使用方 |
|------|-------|--------|
| `ApiResponse<T>` / `PageData<T>` | 后端 JSON 响应 | 所有 `api/*.ts` |
| `AuthState` (Zustand store) | `store/auth.ts` | 所有 pages + hooks + client.ts 拦截器 |
| `CORSConfig` (Go struct) | `middleware/cors.go` | `router/router.go` |

## 循环依赖检查
- [x] 确认无新增循环依赖 — 前端是完全独立项目，与后端仅 HTTP 协议交互

## Commit 记录

| Commit | Type | 子目标 | Message |
|--------|------|-------|---------|
| `b9878d1` | `refactor` | G1 | CORS middleware supports configurable allowed origins |
| `5b825e0` | `feat` | G2-G10 | Add React frontend and Makefile build system |
