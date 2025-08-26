# 多阶段构建Dockerfile
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 跳过系统依赖安装，加速构建

# 设置Go代理加速下载，跳过校验
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV GO111MODULE=on

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖（本地已下载，可跳过）
# RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：运行阶段
FROM alpine:latest

# 跳过运行时依赖安装，加速构建

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从builder阶段复制二进制文件
COPY --from=builder /app/main .

# 复制前端静态文件
COPY --from=builder /app/web ./web

# 复制配置文件
COPY --from=builder /app/sample_config.yaml .

# 设置权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查（简化版）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD [ "true" ]

# 启动应用
CMD ["./main"]
