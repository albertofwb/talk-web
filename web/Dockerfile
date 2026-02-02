# React 前端 Dockerfile

FROM node:18-alpine AS builder

WORKDIR /app

# 复制 package.json
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建生产版本
RUN npm run build

# 运行阶段 - 只保留构建后的文件
FROM scratch AS dist
COPY --from=builder /app/dist /app/dist

# 如果需要用 nginx 提供服务，使用下面的配置
# FROM nginx:alpine
# COPY --from=builder /app/dist /usr/share/nginx/html
# COPY nginx.conf /etc/nginx/conf.d/default.conf
# EXPOSE 80
# CMD ["nginx", "-g", "daemon off;"]
