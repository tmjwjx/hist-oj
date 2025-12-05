# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go mod 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hist-oj-service ./cmd/server

# 运行阶段
FROM alpine:3.18

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/hist-oj-service .

# 复制配置文件目录（运行时通过环境变量覆盖）
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./hist-oj-service"]
