# 多阶段构建 - 构建阶段
FROM node:16-alpine AS builder

# 设置工作目录
WORKDIR /build

# 复制 package 文件
COPY package*.json ./

# 删除可能包含错误镜像源的 package-lock.json，重新安装依赖
RUN rm -f package-lock.json && \
    npm config set registry https://registry.npmjs.org && \
    npm config set strict-ssl false && \
    npm install --legacy-peer-deps --no-audit

# 复制源代码
COPY . .

# 构建生产版本
RUN npm run build

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
