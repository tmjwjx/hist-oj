# 多阶段构建 - 构建阶段
FROM node:16-alpine AS builder

# 设置工作目录
WORKDIR /build

# 复制 package 文件
COPY package*.json ./

# 设置 npm 镜像源（支持构建参数）
ARG NPM_REGISTRY=https://registry.npmmirror.com
RUN rm -f package-lock.json && \
    npm config set registry ${NPM_REGISTRY} && \
    npm config set strict-ssl false && \
    npm config set fetch-timeout 300000 && \
    npm config set fetch-retries 3 && \
    npm config set fetch-retry-mintimeout 10000 && \
    npm config set fetch-retry-maxtimeout 60000 && \
    npm install --legacy-peer-deps --no-audit

# 复制源代码
COPY . .

# 构建生产版本
RUN rm -rf dist node_modules/.cache && npm run build

# 运行阶段 - 使用 Nginx
FROM nginx:alpine

# 复制自定义 Nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 从构建阶段复制构建好的文件
COPY --from=builder /build/dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]
