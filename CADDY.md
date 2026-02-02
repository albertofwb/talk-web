# Caddy 部署指南

Caddy 是一个现代化的 Web 服务器，自带自动 HTTPS 功能。

## 快速开始

### 本地开发（不使用 Docker）

```bash
# 1. 安装 Caddy
# Ubuntu/Debian
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# macOS
brew install caddy

# 2. 启动后端和前端
make up      # 启动数据库
make server  # 启动后端（新终端）
make web     # 启动前端（新终端）

# 3. 启动 Caddy（新终端）
sudo caddy run --config Caddyfile

# 4. 访问
# http://localhost
```

### Docker Compose 部署（推荐）

```bash
# 1. 配置环境变量
cp .env.example .env
nano .env  # 修改配置

# 2. 修改域名（如果有）
nano Caddyfile.prod
# 替换 talk.yourdomain.com 为你的域名

# 3. 启动所有服务
docker-compose -f docker-compose.caddy.yml up -d

# 4. 查看日志
docker-compose -f docker-compose.caddy.yml logs -f

# 5. 访问
# http://your-domain.com
# https://your-domain.com (自动配置 HTTPS)
```

## 配置文件说明

### Caddyfile（开发环境）

```caddy
:80 {
    # 前端代理到 React dev server
    handle /* {
        reverse_proxy localhost:5173
    }

    # API 代理到 Go 后端
    handle /api/* {
        reverse_proxy localhost:8080
    }
}
```

**特点：**
- 监听 80 端口
- 前端开发服务器热重载
- 后端 API 代理

### Caddyfile.prod（生产环境）

```caddy
talk.yourdomain.com {
    # 自动 HTTPS
    tls your-email@example.com

    # 静态文件服务
    root * /var/www/talk-web

    # API 代理
    handle /api/* {
        reverse_proxy localhost:8080
    }

    # SPA 路由支持
    try_files {path} /index.html
}
```

**特点：**
- 自动申请 Let's Encrypt 证书
- HTTPS/HTTP2/HTTP3 支持
- 安全头配置
- 静态资源缓存
- 健康检查

## 配置步骤

### 1. 修改域名

编辑 `Caddyfile.prod`：

```caddy
# 替换为你的域名
talk.yourdomain.com {
    tls your-email@example.com {
        # ...
    }
}
```

### 2. 配置环境变量

编辑 `.env`：

```bash
# 数据库密码
DB_PASSWORD=your-secure-password

# JWT 密钥（必须修改！）
JWT_SECRET=your-very-secure-random-secret-key

# Talk Server URL
TALK_SERVER_URL=http://your-talk-server:5000
```

### 3. 构建前端

```bash
cd web
npm run build
# 构建产物在 dist/ 目录
```

### 4. 启动服务

**方式一：Docker Compose**

```bash
docker-compose -f docker-compose.caddy.yml up -d
```

**方式二：手动启动**

```bash
# 启动数据库
docker-compose up -d postgres redis

# 启动后端
cd server
go run main.go

# 启动 Caddy
sudo caddy run --config Caddyfile.prod
```

## 常用命令

### Caddy 管理

```bash
# 启动 Caddy
caddy run --config Caddyfile

# 后台运行
caddy start --config Caddyfile

# 停止
caddy stop

# 重载配置（无需停机）
caddy reload --config Caddyfile

# 验证配置
caddy validate --config Caddyfile

# 格式化配置文件
caddy fmt --overwrite Caddyfile
```

### Docker Compose 管理

```bash
# 启动所有服务
docker-compose -f docker-compose.caddy.yml up -d

# 停止所有服务
docker-compose -f docker-compose.caddy.yml down

# 查看日志
docker-compose -f docker-compose.caddy.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.caddy.yml logs -f caddy
docker-compose -f docker-compose.caddy.yml logs -f backend

# 重启服务
docker-compose -f docker-compose.caddy.yml restart

# 查看状态
docker-compose -f docker-compose.caddy.yml ps
```

## HTTPS 证书

Caddy 自动管理 HTTPS 证书：

1. **自动申请** - 首次访问时自动向 Let's Encrypt 申请证书
2. **自动续期** - 证书到期前自动续期
3. **无需配置** - 只需提供域名和邮箱

### 证书存储位置

```bash
# Docker
docker volume ls | grep caddy
docker volume inspect talk-web_caddy_data

# 本地安装
~/.local/share/caddy/  # Linux
~/Library/Application Support/Caddy/  # macOS
```

### 使用自己的证书

```caddy
talk.yourdomain.com {
    tls /path/to/cert.pem /path/to/key.pem
}
```

## 性能优化

### 1. 启用 Gzip/Zstd 压缩

```caddy
encode gzip zstd
```

### 2. 静态资源缓存

```caddy
@static {
    path *.js *.css *.png *.jpg
}
handle @static {
    header Cache-Control "public, max-age=31536000"
}
```

### 3. HTTP/3 支持

Caddy 默认启用 HTTP/3（QUIC）

```bash
# 需要开放 UDP 443 端口
sudo ufw allow 443/udp
```

## 安全配置

### 1. 安全头

已在 `Caddyfile.prod` 中配置：
- HSTS（强制 HTTPS）
- CSP（内容安全策略）
- X-Frame-Options（防点击劫持）
- X-Content-Type-Options（防 MIME 嗅探）

### 2. 限流

```caddy
rate_limit {
    zone dynamic {
        key {remote_host}
        events 100
        window 1m
    }
}
```

### 3. IP 白名单

```caddy
@allowed {
    remote_ip 192.168.1.0/24 10.0.0.0/8
}
handle @allowed {
    # 只允许特定 IP 访问
}
```

## 监控和日志

### 查看访问日志

```bash
# Docker
docker exec talk-web-caddy tail -f /var/log/caddy/access.log

# 本地
tail -f /var/log/caddy/access.log
```

### 日志格式

JSON 格式，便于分析：

```json
{
  "ts": 1709366400,
  "request": {
    "remote_ip": "192.168.1.1",
    "method": "GET",
    "uri": "/api/auth/login"
  },
  "duration": 0.123,
  "status": 200
}
```

## 故障排查

### 1. 证书申请失败

```bash
# 检查域名 DNS 解析
dig talk.yourdomain.com

# 检查端口开放
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# 查看 Caddy 日志
docker-compose -f docker-compose.caddy.yml logs caddy
```

### 2. 代理失败

```bash
# 检查后端是否运行
curl http://localhost:8080/health

# 检查 Caddy 配置
caddy validate --config Caddyfile.prod
```

### 3. 权限问题

```bash
# Caddy 需要 root 权限监听 80/443 端口
sudo caddy run --config Caddyfile

# 或使用 systemd
sudo systemctl start caddy
```

## 迁移到 Caddy

### 从 Nginx 迁移

Nginx 配置：
```nginx
server {
    listen 80;
    server_name talk.yourdomain.com;

    location /api {
        proxy_pass http://localhost:8080;
    }

    location / {
        root /var/www/talk-web;
        try_files $uri /index.html;
    }
}
```

等价的 Caddy 配置：
```caddy
talk.yourdomain.com {
    root * /var/www/talk-web

    handle /api/* {
        reverse_proxy localhost:8080
    }

    try_files {path} /index.html
    file_server
}
```

## 参考资源

- [Caddy 官方文档](https://caddyserver.com/docs/)
- [Caddyfile 语法](https://caddyserver.com/docs/caddyfile)
- [Caddy 社区](https://caddy.community/)

---

需要帮助？查看日志或提交 Issue！
