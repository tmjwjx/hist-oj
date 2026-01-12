# 报名系统后端 Dockerfile
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 Go 代理（加速依赖下载）
ENV GOPROXY=https://goproxy.io,https://goproxy.cn,https://proxy.golang.org,direct
ENV GO111MODULE=on

# 安装 ca-certificates（避免 TLS 证书问题）
RUN apk --no-cache add ca-certificates git

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖（添加重试机制）
RUN go mod download || \
    (echo "Retrying with different proxy..." && \
     GOPROXY=https://proxy.golang.org,https://goproxy.cn,direct go mod download) || \
    (echo "Retrying one more time..." && \
     GOPROXY=https://goproxy.cn,direct go mod download)

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 最终镜像
FROM alpine:3.18

# 安装 ca-certificates、curl 和时区数据
RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 创建 uploads 目录
RUN mkdir -p /app/uploads

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/api/competitions || exit 1

# 运行应用
CMD ["./main"]
