# talk-web 部署指南

## 开发环境部署

### 前置要求
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Git

### 快速启动

```bash
# 1. 克隆仓库
git clone <repo-url>
cd talk-web

# 2. 启动数据库
docker-compose up -d

# 3. 安装依赖
make init
# 或手动：
cd server && go mod download
cd ../web && npm install

# 4. 配置环境变量
cp .env.example .env
# 编辑 .env 配置 TALK_SERVER_URL

# 5. 启动服务
# 方式 1: 使用脚本
./start.sh

# 方式 2: 分别启动
# 终端1 - 后端
cd server && go run main.go

# 终端2 - 前端
cd web && npm run dev
```

### 访问应用
- 前端: http://localhost:5173
- 后端 API: http://localhost:8080
- 默认账号: admin / admin123

## 生产环境部署

### 使用 Docker Compose

1. 创建 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: talk
      POSTGRES_PASSWORD: <strong-password>
      POSTGRES_DB: talk
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

  redis:
    image: redis:7-alpine
    restart: always

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=<change-me>
      - TALK_SERVER_URL=http://talk-server:5000
      - GIN_MODE=release
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    restart: always

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: always

volumes:
  postgres_data:
```

2. 创建 `Dockerfile.backend`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

3. 创建 `web/Dockerfile`:

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

4. 创建 `web/nginx.conf`:

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

5. 启动生产环境：

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 环境变量配置

### 必须配置
- `JWT_SECRET`: JWT 密钥，生产环境必须修改
- `TALK_SERVER_URL`: 语音识别服务地址

### 数据库配置
- `DB_HOST`: 数据库主机 (默认 localhost)
- `DB_PORT`: 数据库端口 (默认 5432)
- `DB_USER`: 数据库用户 (默认 talk)
- `DB_PASSWORD`: 数据库密码 (默认 talk)
- `DB_NAME`: 数据库名 (默认 talk)

### Redis 配置
- `REDIS_ADDR`: Redis 地址 (默认 localhost:6380)

### 服务配置
- `PORT`: 后端服务端口 (默认 8080)
- `GIN_MODE`: Gin 模式 (release/debug)

## 数据备份

### PostgreSQL 备份
```bash
# 备份
docker exec talk-web-postgres pg_dump -U talk talk > backup.sql

# 恢复
docker exec -i talk-web-postgres psql -U talk talk < backup.sql
```

## 监控和日志

### 查看日志
```bash
# 后端日志
tail -f logs/server.log

# 前端日志
tail -f logs/web.log

# 数据库日志
docker-compose logs -f postgres
```

## 安全建议

1. **修改默认密码**
   - 登录后立即修改 admin 账号密码
   - 生产环境使用强密码

2. **JWT 密钥**
   - 使用强随机密钥
   - 定期轮换密钥

3. **HTTPS**
   - 生产环境必须使用 HTTPS
   - 浏览器麦克风权限要求 HTTPS

4. **数据库安全**
   - 不要暴露数据库端口到公网
   - 使用强密码
   - 定期备份

5. **CORS 配置**
   - 生产环境限制允许的域名
   - 不要使用 `*` 通配符

## 性能优化

1. **前端优化**
   - 使用 CDN 加速
   - 启用 gzip 压缩
   - 配置浏览器缓存

2. **后端优化**
   - 启用 Redis 缓存
   - 数据库连接池
   - 限流和防护

3. **数据库优化**
   - 添加索引
   - 定期清理日志
   - 配置连接池

## 故障排查

### 后端无法启动
```bash
# 检查数据库连接
docker-compose ps
psql -h localhost -U talk -d talk

# 查看日志
tail -f logs/server.log
```

### 前端无法访问
```bash
# 检查前端服务
curl http://localhost:5173

# 查看日志
tail -f logs/web.log
```

### 音频上传失败
- 检查 `TALK_SERVER_URL` 配置
- 确认 talk-server 服务运行中
- 查看网络连接
