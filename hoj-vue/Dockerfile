# 多阶段构建 - 构建阶段
FROM node:16-alpine AS builder

# 设置工作目录
WORKDIR /build

# 复制 package 文件
COPY package*.json ./

# 设置 npm 镜像源（支持多源回退 + 重试）
ARG NPM_REGISTRIES="https://registry.npmmirror.com https://registry.npmjs.org"
ARG NPM_INSTALL_RETRIES=3
RUN set -eux; \
    npm config set strict-ssl false; \
    npm config set fetch-timeout 600000; \
    npm config set fetch-retries 5; \
    npm config set fetch-retry-mintimeout 20000; \
    npm config set fetch-retry-maxtimeout 120000; \
    success=0; \
    for registry in ${NPM_REGISTRIES}; do \
      npm config set registry "${registry}"; \
      echo "Using npm registry: ${registry}"; \
      attempt=1; \
      while [ "${attempt}" -le "${NPM_INSTALL_RETRIES}" ]; do \
        echo "npm install attempt ${attempt}/${NPM_INSTALL_RETRIES} on ${registry}"; \
        if npm install --legacy-peer-deps --no-audit --prefer-offline; then \
          success=1; \
          break; \
        fi; \
        npm cache clean --force || true; \
        attempt=$((attempt + 1)); \
        sleep 5; \
      done; \
      [ "${success}" -eq 1 ] && break; \
    done; \
    [ "${success}" -eq 1 ]

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
