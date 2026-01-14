# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 使用阿里云镜像源（加速国内访问）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的构建工具（重试机制）
RUN apk add --no-cache git || \
    apk add --no-cache git || \
    apk add --no-cache git

# 设置 Go 代理（加速依赖下载）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 复制 go mod 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hist-oj-service ./cmd/server

# 运行阶段
FROM alpine:3.18

# 使用阿里云镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖（重试机制）
RUN apk add --no-cache ca-certificates tzdata || \
    apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/hist-oj-service .

# 复制配置文件目录（运行时通过环境变量覆盖）
COPY --from=builder /app/configs ./configs

# 创建 uploads 目录（用于持久化上传文件）
RUN mkdir -p /app/uploads

# 暴露端口
EXPOSE 9527

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9527/health || exit 1

# 启动命令
CMD ["./hist-oj-service"]
