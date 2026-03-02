# 构建阶段
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 使用阿里云镜像源（加速国内访问）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的构建工具（包括 sim 编译依赖）
RUN apk add --no-cache git gcc make flex musl-dev || \
    apk add --no-cache git gcc make flex musl-dev

# 设置 Go 代理（加速依赖下载）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 复制 go mod 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 复制并编译 sim 工具（代码查重功能）
RUN if [ -d third_party/sim/sim_src ]; then \
    echo "正在编译 sim 工具..." && \
    cd third_party/sim/sim_src && \
    # 清理预编译的二进制文件（可能是其他架构的）\
    make clean || true && \
    rm -f sim_c sim_java sim_text sim_cc *.o 2>/dev/null || true && \
    # 修复 Makefile: 替换路径和添加编译标志 \
    sed -i 's|/home/dick|/usr/local|g' Makefile && \
    sed -i 's|^CC[[:space:]]*=[[:space:]]*gcc|CC = gcc -std=gnu89|' Makefile && \
    echo "开始编译..." && \
    # 编译 sim_c 和 sim_java \
    make sim_c sim_java && \
    # 验证编译结果 \
    echo "编译产物:" && \
    ls -la sim_c sim_java && \
    mkdir -p /tmp/sim-bin && \
    cp sim_c sim_java /tmp/sim-bin/ && \
    echo "✓ sim 工具编译成功（sim_c, sim_java）"; \
    else \
    echo "警告: third_party/sim/sim_src 不存在，跳过 sim 工具编译"; \
    fi

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hist-oj-service ./cmd/server

# 运行阶段
FROM alpine:3.18

# 使用多个镜像源（阿里云 + 官方源备份）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.18/main/" >> /etc/apk/repositories && \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.18/main/" >> /etc/apk/repositories

# 安装必要的运行时依赖（重试机制）
RUN apk add --no-cache ca-certificates tzdata wget python3 py3-pip || \
    apk add --no-cache ca-certificates tzdata wget python3 py3-pip

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/hist-oj-service .

# 复制 sim 工具二进制文件（如果存在）
RUN mkdir -p /usr/local/bin
COPY --from=builder /tmp/sim-bin /usr/local/bin/sim-tools
RUN echo "检查 sim-tools 目录..." && \
    ls -la /usr/local/bin/ && \
    if [ -d /usr/local/bin/sim-tools ]; then \
    echo "sim-tools 存在，内容:" && \
    ls -la /usr/local/bin/sim-tools/ && \
    cp /usr/local/bin/sim-tools/* /usr/local/bin/ && \
    chmod +x /usr/local/bin/sim_* && \
    rm -rf /usr/local/bin/sim-tools && \
    echo "✓ sim工具已安装：" && \
    ls -la /usr/local/bin/sim_*; \
    else \
    echo "⚠️  警告: sim工具不存在，查重功能将无法使用"; \
    echo "可能原因:" && \
    echo "  1. third_party/sim/sim_src 目录不存在"; \
    echo "  2. sim 编译失败"; \
    echo "  3. 构建架构不匹配"; \
    fi

# 复制配置文件目录（运行时通过环境变量覆盖）
COPY --from=builder /app/configs ./configs

# 创建持久化存储目录
RUN mkdir -p /app/uploads

# 声明数据卷
VOLUME ["/app/uploads"]

# 暴露端口
EXPOSE 9527

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9527/health || exit 1

# 启动命令
CMD ["./hist-oj-service"]
