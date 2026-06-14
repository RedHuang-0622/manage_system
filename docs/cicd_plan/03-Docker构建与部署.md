# 03 — Docker 构建与部署

## 1 Dockerfile（多阶段构建）

**文件：** `Dockerfile`

```dockerfile
# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 利用 Docker 缓存：先复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译（静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/manage_system ./cmd/

# ============================================================
# Stage 2: Runtime
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 从 builder 复制二进制
COPY --from=builder /app/manage_system .
COPY --from=builder /app/conf/ ./conf/

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./manage_system"]
```

## 2 Docker Compose（本地开发 + 测试）

**文件：** `docker-compose.yaml`

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: lab_mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-root123}
      MYSQL_DATABASE: lab_manage
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: lab_redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    build:
      context: .
      args:
        VERSION: ${VERSION:-dev}
    container_name: lab_app
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      LAB_SERVER_PORT: 8080
      LAB_SERVER_MODE: release
      LAB_MYSQL_HOST: mysql
      LAB_MYSQL_PORT: 3306
      LAB_MYSQL_USER: root
      LAB_MYSQL_PASSWORD: ${MYSQL_ROOT_PASSWORD:-root123}
      LAB_MYSQL_DATABASE: lab_manage
      LAB_REDIS_ADDR: redis:6379
      LAB_JWT_SECRET: ${JWT_SECRET:-change-me-in-production-min-32-chars}
      LAB_LOG_LEVEL: info

volumes:
  mysql_data:
```

### 2.1 测试用 Docker Compose

**文件：** `docker-compose.test.yaml`

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: test123
      MYSQL_DATABASE: lab_manage_test
    ports:
      - "3307:3306"
    tmpfs: /var/lib/mysql   # 内存存储，重启即清空
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

## 3 Docker 构建 CI 扩展

在 `.github/workflows/ci.yml` 中添加 Docker 构建 Job：

```yaml
  docker-build:
    name: Docker Build & Push
    runs-on: ubuntu-latest
    needs: integration-test
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ secrets.DOCKER_USERNAME }}/lab-manage-system
          tags: |
            type=sha,prefix=,format=short
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.sha }}
```

## 4 启动命令速查

```bash
# === 本地开发 ===
docker-compose up -d mysql redis     # 仅启动依赖
go run ./cmd/                         # 本地运行应用

# === 本地全栈启动 ===
docker-compose up -d                  # 启动全部（包括 app）

# === 本地测试 ===
docker-compose -f docker-compose.test.yaml up -d
go test ./... -tags=integration -v
docker-compose -f docker-compose.test.yaml down -v

# === 生产部署 ===
export JWT_SECRET="production-secret-at-least-32-chars"
export MYSQL_ROOT_PASSWORD="strong-password"
docker-compose up -d
```

## 5 镜像大小优化

多阶段构建后的预期镜像大小：

| 阶段 | 大小 | 说明 |
|------|------|------|
| Builder (golang:1.21-alpine) | ~400 MB | 仅构建时存在 |
| Runtime (alpine:3.19) | ~15 MB | 最终运行镜像 |
| 编译产物 (manage_system) | ~10 MB | `-ldflags="-s -w"` 后 |

进一步优化可选方案：
- 使用 `scratch` 基础镜像（~5 MB，但无 shell 和 ca-certificates）
- 使用 `distroless` 基础镜像（~8 MB，包含 CA 证书）
