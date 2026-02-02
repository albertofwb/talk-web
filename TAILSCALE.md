# Tailscale 访问指南

既然你们都有 Tailscale，那就太简单了！🎉

## 🚀 超级简单的方案

### 优势
- ✅ **自动 HTTPS** - Tailscale 自带证书
- ✅ **麦克风权限** - 完美支持
- ✅ **随时随地访问** - 不限局域网
- ✅ **无需配置** - 零配置

---

## 方式一：直接访问（最简单）⭐⭐⭐

### 1. 获取 Tailscale IP

在服务器上：

```bash
# 查看 Tailscale IP
tailscale ip -4

# 例如: 100.64.1.10
```

你的 Tailscale IP: **100.x.x.x**

### 2. 启动服务

**不需要 Caddy！** 直接用开发服务器：

```bash
# 启动数据库
make up

# 启动后端
cd server && go run main.go

# 启动前端（新终端）
cd web && npm run dev
```

### 3. 修改前端监听地址

编辑 `web/vite.config.ts`：

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',  // 添加这行 - 监听所有网络接口
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

重启前端：

```bash
cd web && npm run dev
```

### 4. Android 访问

在 Android Chrome 打开：

```
http://100.64.1.10:5173
```

**完成！** ✅

---

## 方式二：使用 Caddy + Tailscale（推荐）⭐⭐⭐⭐

### 优势
- 不需要端口号
- 统一入口
- 更专业

### 1. 获取 Tailscale IP 和主机名

```bash
# 查看 IP
tailscale ip -4
# 例如: 100.64.1.10

# 查看主机名
hostname
# 例如: albert-pc
```

### 2. 创建 Tailscale 专用配置

```bash
cat > Caddyfile.tailscale <<'EOF'
# Tailscale 访问配置

# 使用 Tailscale IP
http://100.64.1.10 {
    # 前端代理
    handle /* {
        reverse_proxy localhost:5173
    }

    # API 代理
    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /health {
        reverse_proxy localhost:8080
    }
}

# 或者使用 MagicDNS 域名
# http://albert-pc.tail-scale.ts.net {
#     handle /* {
#         reverse_proxy localhost:5173
#     }
#     handle /api/* {
#         reverse_proxy localhost:8080
#     }
# }
EOF
```

### 3. 启动 Caddy

```bash
sudo caddy run --config Caddyfile.tailscale
```

### 4. Android 访问

```
http://100.64.1.10
```

或者使用 MagicDNS（如果启用）：

```
http://albert-pc.tail-scale.ts.net
```

---

## 方式三：Tailscale HTTPS（最完美）⭐⭐⭐⭐⭐

Tailscale 提供免费的 HTTPS 证书！

### 1. 启用 Tailscale HTTPS

在服务器上：

```bash
# 启用 HTTPS
tailscale cert
```

### 2. 获取 MagicDNS 域名

```bash
# 查看你的 Tailscale 域名
tailscale status

# 格式: <hostname>.<tailnet-name>.ts.net
# 例如: albert-pc.tail12345.ts.net
```

### 3. 创建 HTTPS 配置

```bash
TAILSCALE_HOSTNAME=$(tailscale status --json | jq -r '.Self.DNSName' | tr -d '.')

cat > Caddyfile.tailscale-https <<EOF
https://$TAILSCALE_HOSTNAME {
    tls {
        get_certificate tailscale
    }

    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /* {
        reverse_proxy localhost:5173
    }
}
EOF
```

### 4. 启动 Caddy

```bash
sudo caddy run --config Caddyfile.tailscale-https
```

### 5. Android 访问

```
https://albert-pc.tail12345.ts.net
```

**完美的 HTTPS！** ✅
- 绿色小锁
- 受信任的证书
- 麦克风权限正常

---

## 🎯 推荐方案对比

| 方案 | 配置难度 | 麦克风 | HTTPS | 推荐指数 |
|------|---------|--------|-------|---------|
| 方式一：直接访问 | ⭐ | ✅ | ❌ | ⭐⭐⭐ |
| 方式二：Caddy HTTP | ⭐⭐ | ✅ | ❌ | ⭐⭐⭐⭐ |
| 方式三：Tailscale HTTPS | ⭐⭐⭐ | ✅ | ✅ | ⭐⭐⭐⭐⭐ |

---

## 快速开始（懒人版）

```bash
# 1. 启动后端和前端
./start.sh

# 2. 修改前端监听
echo "修改 web/vite.config.ts 添加 host: '0.0.0.0'"

# 3. 重启前端
cd web && npm run dev

# 4. 在 Android 访问
# http://$(tailscale ip -4):5173
```

完成！

---

## 常见问题

### Q: 为什么推荐用 Tailscale？

**A:** 因为：
1. ✅ 你们都已经有了
2. ✅ 零配置的安全连接
3. ✅ 不需要处理证书
4. ✅ 随时随地访问（不限局域网）
5. ✅ 麦克风权限自动工作

### Q: 需要开放防火墙吗？

**A:** 不需要！Tailscale 自动处理所有网络连接。

### Q: 可以在外网访问吗？

**A:** 可以！只要手机和服务器都连接到 Tailscale，无论在哪里都能访问。

### Q: 性能怎么样？

**A:** 局域网内 Tailscale 会自动使用直连，性能和普通局域网一样。

---

## 一键配置脚本

创建 `setup-tailscale.sh`：

```bash
#!/bin/bash

echo "🔐 Tailscale 访问配置"
echo "===================="
echo ""

# 检查 Tailscale
if ! command -v tailscale &> /dev/null; then
    echo "❌ 未安装 Tailscale"
    echo "请访问: https://tailscale.com/download"
    exit 1
fi

# 获取 Tailscale 信息
TS_IP=$(tailscale ip -4)
TS_HOSTNAME=$(hostname)

echo "✓ Tailscale IP: $TS_IP"
echo "✓ 主机名: $TS_HOSTNAME"
echo ""

# 修改前端配置
echo "配置前端监听所有接口..."
if grep -q "host:" web/vite.config.ts; then
    echo "已配置"
else
    sed -i "s/server: {/server: {\n    host: '0.0.0.0',/" web/vite.config.ts
    echo "✓ 已修改 vite.config.ts"
fi

# 启动服务
echo ""
echo "启动服务..."
./start.sh &

sleep 3

echo ""
echo "✅ 配置完成！"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 Android 访问地址:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  http://$TS_IP:5173"
echo ""
echo "💡 使用 Tailscale 连接即可访问"
echo ""
```

使用：

```bash
chmod +x setup-tailscale.sh
./setup-tailscale.sh
```

---

## 总结

有 Tailscale 就不要折腾证书了！

**最简单的方式：**
1. `tailscale ip -4` 获取 IP
2. 修改前端监听 `0.0.0.0`
3. Android 访问 `http://100.x.x.x:5173`

**搞定！** 🎉
