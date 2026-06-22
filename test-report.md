# 测试报告

## 概览

| 维度 | 总数 | 通过 | 失败 | 跳过 | 覆盖 |
|------|------|------|------|------|------|
| Backend Unit | 5 packages | ✅ 5 | 0 | 0 | 0.1% ~ 92.4% |
| Frontend Unit | 3 files, 27 tests | ✅ 27 | 0 | 0 | — |
| Integration | 9 scenarios | ⏸️ 未跑 (需 -tags=integration) | — | — | — |
| E2E (Playwright) | 5 specs | ⏸️ 未跑 (需启动服务) | — | — | — |

## 各维度详情

### 1. 后端单元测试 (`go test -race -coverpkg=manage_system/...`)

| Package | 状态 | 覆盖率 |
|---------|:----:|:------:|
| `test/dao/` | ✅ | 92.4% |
| `test/service/` | ✅ | 51.2% |
| `test/middleware/` | ✅ | 36.2% |
| `test/pkg/` | ✅ | 40.2% |
| `test/controller/` | ✅ | 0.1% |

- **框架**: testify mock + SQLite 内存库 + miniredis
- **竞态检测**: `-race` 全部通过
- **综合评价**: DAO 层覆盖优秀 (92.4%)，Service 层覆盖中等 (51.2%)，Controller 层仅测参数校验

### 2. 前端单元测试 (`vitest`)

| 文件 | 测试数 | 状态 |
|------|:------:|:----:|
| `src/api/__tests__/auth.test.ts` | 6 | ✅ |
| `src/store/__tests__/auth.test.ts` | 13 | ✅ |
| `src/hooks/__tests__/usePermission.test.ts` | 8 | ✅ |

- **框架**: vitest + jsdom + @testing-library/react
- **测试覆盖**: JWT decode、Auth Store 状态管理、5 种角色权限矩阵

### 3. 静态分析

| 检查 | 状态 |
|------|:----:|
| `go vet ./...` (backend) | ✅ |
| `tsc --noEmit` (frontend) | ✅ |
| `eslint` | ⚠️ 配置已创建 (.eslintrc.cjs)，待验证 |

### 4. CI 流水线

| Job | 触发 | 阻塞 |
|-----|:----:|:----:|
| backend-vet (vet + build) | push/PR | ✅ |
| frontend-vet (tsc + lint + vite build) | push/PR | ✅ |
| backend-test (race + coverage, Go 1.23/1.24) | push/PR | ✅ |
| frontend-test (vitest + coverage) | push/PR | ✅ |
| backend-integration (MySQL + Redis) | push/PR | ✅ |
| e2e (Playwright) | push/PR | ✅ |
| static-analysis (gofmt, govulncheck, go mod tidy) | push/PR | ✅ |
| summary (test report) | push/PR | — |

## 修复记录

| # | 文件 | 问题 | 状态 |
|---|------|------|:----:|
| 1 | `backend/cmd/main.go` | `db.DB()` 错误被忽略 | ✅ |
| 2 | `backend/cmd/main.go` | `initCasbin` 用 `log.Fatalf` 而非 `logger.Fatal` | ✅ |
| 3 | `backend/cmd/main.go` | 死代码 `hasPolicy` 函数 | ✅ |
| 4 | `backend/cmd/main.go` | 明文密码打在日志 | ✅ |
| 5 | `backend/cmd/main.go` | 变量遮蔽 `admin` | ✅ |
| 6 | `frontend/src/components/RoleGuard.tsx` | `equipment_manager` 不在类型联合 | ✅ |
| 7 | `frontend/src/components/Layout/Sidebar.tsx` | `EyeOutlined` 未使用 | ✅ |
| 8 | `frontend/src/pages/equipment/List.tsx` | `setSearchParams` 未使用 | ✅ |
| 9 | `frontend/src/pages/users/List.tsx` | `fetchData()` 缺参数 | ✅ |
| 10 | `frontend/playwright.config.ts` | 硬编码 Chrome 路径 | ✅ |
| 11 | `frontend/e2e/global.setup.ts` | 硬编码 Chrome 路径 + BASE URL | ✅ |

## 新增文件

| 文件 | 说明 |
|------|------|
| `.github/workflows/ci.yml` | CI 流水线 (8 jobs) |
| `frontend/vitest.config.ts` | Vitest 配置 (jsdom + coverage) |
| `frontend/.eslintrc.cjs` | ESLint 配置 |
| `frontend/src/test-setup.ts` | Jest-dom matchers |
| `frontend/src/api/__tests__/auth.test.ts` | JWT 解码测试 (6) |
| `frontend/src/store/__tests__/auth.test.ts` | Auth Store 测试 (13) |
| `frontend/src/hooks/__tests__/usePermission.test.ts` | 权限 Hook 测试 (8) |

## 综合判断

- [x] ✅ **通过** — 所有可本地运行的测试全部通过
- [ ] ⚠️ 有条件通过
- [ ] 🚨 不通过

> **注**: 集成测试 (`-tags=integration`) 和 E2E 测试需要 MySQL + Redis + 启动服务，仅在 CI 环境中运行。CI 流水线已完整配置。
