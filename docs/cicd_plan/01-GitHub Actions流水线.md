# 01 — GitHub Actions 流水线

## 1 流水线总文件

**文件：** `.github/workflows/ci.yml`

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop, 'feature/**']
  pull_request:
    branches: [main]
  workflow_dispatch:

env:
  GO_VERSION: '1.21'
  MYSQL_ROOT_PASSWORD: test123
  MYSQL_DATABASE: lab_manage_test

jobs:
  # ============================================================
  # Stage 1: Lint & Static Check
  # ============================================================
  lint:
    name: Lint & Static Analysis
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: go vet
        run: go vet ./...

      - name: go mod tidy check
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum || \
            (echo "go.mod/go.sum not tidy. Run 'go mod tidy' locally." && exit 1)

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m --out-format=colored-line-number
          only-new-issues: true

      - name: gofmt check
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            echo "Files not gofmt'd:"
            echo "$unformatted"
            exit 1
          fi

  # ============================================================
  # Stage 2: Unit Test
  # ============================================================
  unit-test:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run unit tests (with race detector)
        run: |
          go test ./... -short -race -coverprofile=coverage.out \
            -covermode=atomic -v 2>&1 | tee unit-test.log

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
          echo "Overall coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "::warning::Coverage ${COVERAGE}% is below 70% threshold"
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: unit-coverage
          fail_ci_if_error: false

      - name: Upload coverage artifact
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out

      - name: Run Service layer coverage check
        run: |
          go test ./service/... -coverprofile=service_coverage.out -covermode=atomic
          SVC_COV=$(go tool cover -func=service_coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
          echo "Service layer coverage: ${SVC_COV}%"
          if (( $(echo "$SVC_COV < 85" | bc -l) )); then
            echo "::warning::Service coverage ${SVC_COV}% below 85% threshold"
          fi

  # ============================================================
  # Stage 3: Integration Test
  # ============================================================
  integration-test:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: unit-test
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: ${{ env.MYSQL_ROOT_PASSWORD }}
          MYSQL_DATABASE: ${{ env.MYSQL_DATABASE }}
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping -h localhost"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd="redis-cli ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Wait for MySQL
        run: |
          for i in {1..30}; do
            if mysqladmin ping -h 127.0.0.1 -u root -p${{ env.MYSQL_ROOT_PASSWORD }} --silent; then
              echo "MySQL is ready"
              break
            fi
            echo "Waiting for MySQL... ($i)"
            sleep 2
          done

      - name: Wait for Redis
        run: |
          for i in {1..10}; do
            if redis-cli -h 127.0.0.1 ping | grep -q PONG; then
              echo "Redis is ready"
              break
            fi
            echo "Waiting for Redis... ($i)"
            sleep 1
          done

      - name: Generate test config
        run: |
          cat > conf/config.test.yaml << EOF
          server:
            port: 8080
            mode: test
          mysql:
            host: 127.0.0.1
            port: 3306
            user: root
            password: ${{ env.MYSQL_ROOT_PASSWORD }}
            database: ${{ env.MYSQL_DATABASE }}
            charset: utf8mb4
            max_idle_conns: 5
            max_open_conns: 20
            conn_max_lifetime: 3600
          redis:
            addr: 127.0.0.1:6379
            password: ""
            db: 0
            pool_size: 10
            min_idle_conns: 5
          jwt:
            secret: "test-secret-key-for-ci-at-least-32-chars"
            expire: 7200
            issuer: "lab-system-test"
          casbin:
            model_path: "conf/rbac_model.conf"
          log:
            path: "logs/test.log"
            level: debug
            max_size: 10
            max_backups: 3
            max_age: 1
          EOF

      - name: Run integration tests
        run: |
          go test ./... -tags=integration -v -count=1 -timeout 120s 2>&1 | \
          tee integration-test.log

      - name: Upload integration test log
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: integration-test-log
          path: integration-test.log

  # ============================================================
  # Stage 4: Build
  # ============================================================
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: integration-test
    if: github.ref == 'refs/heads/main' || github.event_name == 'workflow_dispatch'
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Go build (Linux)
        run: |
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
            go build -ldflags="-s -w -X main.Version=${GITHUB_SHA::7} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            -o bin/manage_system_linux ./cmd/

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: manage_system_linux
          path: bin/manage_system_linux
```

## 2 Stage 拆分说明

### 2.1 Lint & Static Check

| 检查项 | 工具 | 阻断 |
|--------|------|------|
| 静态分析 | `go vet ./...` | 是 |
| 依赖完整性 | `go mod tidy` + diff 检查 | 是 |
| 代码规范 | `golangci-lint` | 是 |
| 格式化 | `gofmt -l .` | 是 |

### 2.2 Unit Test

| 步骤 | 说明 |
|------|------|
| `-short` flag | 跳过需要外部依赖的测试 |
| `-race` flag | 启用竞态检测（go test 内置） |
| `-covermode=atomic` | 并发安全的覆盖率统计 |
| 覆盖率阈值 | 整体 ≥ 70%，Service 层 ≥ 85%（低于阈值发 Warning） |
| 报告上传 | Codecov + Artifact 双通道 |

### 2.3 Integration Test

| 要素 | 实现 |
|------|------|
| MySQL | GitHub Actions `services.mysql`，健康检查等待 |
| Redis | GitHub Actions `services.redis`，健康检查等待 |
| 配置注入 | 动态生成 `conf/config.test.yaml` |
| `-tags=integration` | 隔离集成测试，避免 `go test ./...` 默认运行 |

### 2.4 Build

| 要素 | 实现 |
|------|------|
| 触发条件 | 仅 main 分支 Push 或手动触发 |
| 编译优化 | `-ldflags="-s -w"` 去除调试信息，减小二进制 |
| 版本注入 | `-X main.Version` / `-X main.BuildTime` 编译时注入 |
| 产物 | 上传 Linux amd64 二进制为 Artifact |

## 3 golangci-lint 配置

**文件：** `.golangci.yml`

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck      # 未处理的 error
    - gosimple      # 简化建议
    - govet         # go vet
    - ineffassign   # 无效赋值
    - staticcheck   # 静态分析
    - unused        # 未使用变量
    - gofmt         # 格式化
    - goimports     # import 排序
    - misspell      # 拼写检查
    - bodyclose     # HTTP body 未关闭
    - nilerr        # 返回 nil error
    - prealloc      # slice 预分配

linters-settings:
  errcheck:
    check-blank: true
  govet:
    check-shadowing: true

issues:
  exclude-dirs:
    - vendor
    - bin
  max-issues-per-linter: 50
  max-same-issues: 10
```

## 4 本地运行验证

开发者在提交前可本地运行完整的 CI 流程：

```bash
# === Stage 1: Lint ===
go vet ./...
go mod tidy && git diff --exit-code go.mod go.sum
golangci-lint run --timeout=5m
test -z "$(gofmt -l .)"

# === Stage 2: Unit Test ===
go test ./... -short -race -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1   # 查看整体覆盖率

# === Stage 3: Integration Test ===
docker-compose -f docker-compose.test.yaml up -d
# 等待 MySQL + Redis ready...
go test ./... -tags=integration -v -count=1
docker-compose -f docker-compose.test.yaml down -v

# === Stage 4: Build ===
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/manage_system ./cmd/
```
