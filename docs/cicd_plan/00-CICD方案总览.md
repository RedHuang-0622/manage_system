# 00 — CI/CD 自动化测试方案总览

## 1 目标

为"实验室资产与人员管理系统"建立完整的 CI/CD 流水线，实现代码提交 → 自动检查 → 自动测试 → 自动构建的闭环，确保每次提交的质量可追溯。

## 2 流水线架构

```
Git Push / PR
      │
      ▼
┌─────────────────────────────────────────────────────┐
│                    CI Pipeline                       │
│                                                     │
│  Stage 1: Lint & Static Check                       │
│  ├── go vet        (静态分析)                        │
│  ├── golangci-lint (代码规范)                        │
│  └── go mod tidy   (依赖完整性检查)                   │
│                                                     │
│  Stage 2: Unit Test                                 │
│  ├── go test ./... -race -cover (单元+竞态)         │
│  └── 覆盖率报告上传                                  │
│                                                     │
│  Stage 3: Integration Test                          │
│  ├── docker-compose up (MySQL + Redis)              │
│  ├── go test -tags=integration ./...                │
│  └── docker-compose down                            │
│                                                     │
│  Stage 4: Build                                     │
│  ├── go build -ldflags="-s -w" (编译)               │
│  ├── Docker image build (镜像)                       │
│  └── Artifact upload (产物归档)                      │
│                                                     │
└─────────────────────────────────────────────────────┘
      │
      ▼ (main 分支合并后)
┌─────────────────────────────────────────────────────┐
│                    CD Pipeline (可选)                 │
│  ├── Push Docker Image to Registry                  │
│  ├── Deploy to Test Environment                     │
│  └── Smoke Test                                     │
└─────────────────────────────────────────────────────┘
```

## 3 触发策略

| 触发条件 | 执行 Stage | 说明 |
|---------|-----------|------|
| 所有分支 Push | Lint + Unit Test | 快速反馈 |
| PR → main | Lint + Unit Test + Integration Test | 合并前全量验证 |
| main 分支 Push | Lint + Unit + Integration + Build + Deploy | 自动部署测试环境 |
| 手动触发 (workflow_dispatch) | 全部 Stage | 按需执行 |

## 4 质量门禁

| 门禁 | 阈值 | 阻塞级别 |
|------|------|---------|
| golangci-lint 错误 | 0 errors | 阻塞合并 |
| 单元测试通过率 | 100% | 阻塞合并 |
| 代码覆盖率 | ≥ 70% (整体), ≥ 85% (Service 层) | 告警，不阻塞 |
| 竞态检测 | 0 race conditions | 阻塞合并 |
| 集成测试通过率 | 100% | 阻塞合并 |
| 编译成功 | Go build 无 error | 阻塞合并 |

## 5 文档索引

| 编号 | 文档 | 内容 |
|------|------|------|
| 01 | [GitHub Actions流水线.md](01-GitHub Actions流水线.md) | 完整 YAML 配置、Stage 拆解 |
| 02 | [测试报告与监控.md](02-测试报告与监控.md) | 覆盖率报告、JUnit 输出、通知 |
| 03 | [Docker构建与部署.md](03-Docker构建与部署.md) | 多阶段构建、镜像优化、部署脚本 |
