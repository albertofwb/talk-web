# Homelab / 内网部署指南

talk-web 在家庭网络或内网环境中的部署配置。

## 网络拓扑

```
家庭网络/内网
├── 路由器 (192.168.1.1)
├── 服务器 (192.168.1.100) - talk-web
└── 其他设备 (192.168.1.x)
```

## 快速开始

### 1. 确定服务器 IP

```bash
# 查看本机 IP
hostname -I
ip addr show
```

假设服务器 IP 是 `192.168.1.100`

### 2. 配置本地 DNS（可选）

#### 方式一：修改 /etc/hosts

在**所有设备**上添加：

```bash
# Windows: C:\Windows\System32\drivers\etc\hosts
# Linux/Mac: /etc/hosts

192.168.1.100  talk.home.local
192.168.1.100  talk.home.wbsays.com
```

#### 方式二：配置路由器 DNS

在路由器管理界面添加本地 DNS 记录：
- 主机名: `talk`
- IP: `192.168.1.100`
- 域名: `home.local` 或 `home.wbsays.com`

### 3. 启动服务

```bash
# 方式一：使用脚本（推荐）
./start-caddy.sh

# 方式二：手动启动
sudo caddy run --config Caddyfile.local
```

### 4. 访问应用

从任意设备访问：
- `http://talk.home.local`
- `http://talk.home.wbsays.com`
- `http://192.168.1.100`

## 配置选项

### 选项 1: HTTP 模式（推荐内网使用）

**优点：**
- 配置简单
- 无需证书
- 性能好

**缺点：**
- 无加密
- 浏览器可能限制某些功能（麦克风需要 HTTPS）

**配置：** `Caddyfile.local`

```caddy
http://talk.home.local:80 {
    handle /* {
        reverse_proxy localhost:5173
    }
    handle /api/* {
        reverse_proxy localhost:8080
    }
}
```

### 选项 2: 自签名 HTTPS（推荐 Homelab）

**优点：**
- 加密传输
- 支持所有浏览器功能
- 类生产环境

**缺点：**
- 需要信任证书
- 首次访问有警告

**配置：** `Caddyfile.homelab`

```caddy
talk.home.wbsays.com {
    tls internal  # Caddy 自动生成自签名证书
    # 其他配置...
}
```

### 选项 3: 本地 CA 证书（最佳体验）

使用 mkcert 创建本地受信任的证书：

```bash
# 安装 mkcert
brew install mkcert  # macOS
# 或
wget -O mkcert https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
chmod +x mkcert
sudo mv mkcert /usr/local/bin/

# 安装本地 CA
mkcert -install

# 生成证书
mkcert talk.home.local talk.home.wbsays.com localhost 192.168.1.100

# 会生成两个文件：
# - talk.home.local+3.pem (证书)
# - talk.home.local+3-key.pem (私钥)

# 在 Caddyfile 中使用
talk.home.local {
    tls talk.home.local+3.pem talk.home.local+3-key.pem
    # 其他配置...
}
```

## 麦克风权限问题

浏览器要求麦克风权限需要 **HTTPS** 或 **localhost**。

### 解决方案

#### 1. 使用 localhost（开发）

```bash
# 在服务器上访问
http://localhost:5173
```

#### 2. 使用自签名证书（推荐）

```bash
# 启用 HTTPS
sudo caddy run --config Caddyfile.homelab

# 访问 https://talk.home.local
# 首次访问点击"高级" -> "继续访问"
```

#### 3. Chrome 强制允许（仅开发）

```bash
# 启动 Chrome 时添加参数
google-chrome --unsafely-treat-insecure-origin-as-secure="http://192.168.1.100" --user-data-dir=/tmp/chrome-dev
```

#### 4. 使用 mkcert（最佳）

按上面"选项 3"步骤操作，所有设备都信任证书。

## Docker 部署（内网）

### docker-compose.homelab.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: talk
      POSTGRES_PASSWORD: talk
      POSTGRES_DB: talk
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - talk-network
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    networks:
      - talk-network
    restart: unless-stopped

  backend:
    build: ./server
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=homelab-secret
    depends_on:
      - postgres
      - redis
    networks:
      - talk-network
    restart: unless-stopped

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile.homelab:/etc/caddy/Caddyfile
      - ./web/dist:/var/www/talk-web
      - caddy_data:/data
    depends_on:
      - backend
    networks:
      - talk-network
    restart: unless-stopped

networks:
  talk-network:

volumes:
  postgres_data:
  caddy_data:
```

### 启动

```bash
# 构建前端
cd web && npm run build

# 启动所有服务
docker-compose -f docker-compose.homelab.yml up -d

# 查看日志
docker-compose -f docker-compose.homelab.yml logs -f
```

## 端口转发（外网访问）

如果需要从外网访问：

### 1. 路由器端口转发

在路由器设置：
```
外部端口 80 -> 192.168.1.100:80
外部端口 443 -> 192.168.1.100:443
```

### 2. 使用 Tailscale/ZeroTier（推荐）

更安全的方案：

```bash
# 安装 Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# 启动
sudo tailscale up

# 获取 Tailscale IP
tailscale ip -4

# 通过 Tailscale IP 访问
http://100.x.x.x
```

### 3. 使用 Cloudflare Tunnel

```bash
# 安装 cloudflared
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb

# 登录并创建隧道
cloudflared tunnel login
cloudflared tunnel create talk-web
cloudflared tunnel route dns talk-web talk.yourdomain.com

# 配置
cat > ~/.cloudflared/config.yml <<EOF
tunnel: <tunnel-id>
credentials-file: /home/$USER/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: talk.yourdomain.com
    service: http://localhost:80
  - service: http_status:404
EOF

# 启动
cloudflared tunnel run talk-web
```

## 性能优化

### 1. 静态资源缓存

```caddy
@static {
    path *.js *.css *.png *.jpg
}
handle @static {
    header Cache-Control "public, max-age=86400"
}
```

### 2. 压缩

```caddy
encode gzip zstd
```

### 3. HTTP/2

Caddy 默认启用 HTTP/2

## 监控

### 查看访问日志

```bash
tail -f /var/log/caddy/talk-web.log
```

### 系统资源监控

```bash
# 安装 htop
sudo apt install htop

# 查看资源使用
htop

# 查看端口
sudo netstat -tlnp | grep -E "(80|443|8080|5173)"
```

## 备份

```bash
# 备份数据库
docker exec talk-web-postgres pg_dump -U talk talk > backup-$(date +%Y%m%d).sql

# 备份配置
tar -czf config-backup-$(date +%Y%m%d).tar.gz \
    .env Caddyfile* docker-compose*.yml

# 自动备份脚本
cat > backup.sh <<'EOF'
#!/bin/bash
BACKUP_DIR=~/backups/talk-web
mkdir -p $BACKUP_DIR
docker exec talk-web-postgres pg_dump -U talk talk > $BACKUP_DIR/db-$(date +%Y%m%d).sql
find $BACKUP_DIR -name "db-*.sql" -mtime +7 -delete
EOF

chmod +x backup.sh

# 添加到 crontab
crontab -e
# 每天凌晨 2 点备份
0 2 * * * /path/to/backup.sh
```

## 故障排查

### 无法访问

```bash
# 检查服务状态
./status.sh

# 检查防火墙
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 检查 Caddy
sudo caddy validate --config Caddyfile.local
```

### 证书问题

```bash
# 重新生成自签名证书
mkcert -uninstall
mkcert -install
mkcert talk.home.local
```

---

现在你的 talk-web 已经可以在家庭网络中完美运行了！ 🏠
