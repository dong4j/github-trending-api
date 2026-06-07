# ===========================================
# Stage 1: 构建阶段
# ===========================================
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
# 这样代码修改不会触发依赖下载
COPY go.mod go.sum* ./
RUN go mod download

# 复制源码并编译
COPY . .
# CGO_ENABLED=0 编译为静态二进制,适用于 alpine / scratch
# -ldflags="-w -s" 去除调试信息,减小二进制体积
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./cmd/server/

# ===========================================
# Stage 2: 运行阶段
# ===========================================
# 使用 alpine 基础镜像保持体积小巧
FROM alpine:3.19

# 安装 CA 证书 (HTTPS 请求 GitHub 需要)
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=UTC

# 创建非 root 用户运行服务 (安全最佳实践)
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 从 builder 阶段复制编译产物
COPY --from=builder /app/bin/server /app/server

# 切换到非 root 用户
USER app

# 暴露端口
EXPOSE 5002

# 健康检查 (可选)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:5002/ || exit 1

# 启动服务
ENTRYPOINT ["/app/server"]
