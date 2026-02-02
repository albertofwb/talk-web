# Android 访问指南

## 📱 从 Android Chrome 访问 talk-web

### 前提条件

1. ✅ Android 设备和服务器在同一 WiFi 网络
2. ✅ 服务器已启动（后端 + 前端 或 Caddy）
3. ✅ 防火墙允许局域网访问

---

## 🚀 快速开始

### 步骤 1：获取服务器 IP

在服务器上运行：

```bash
# 查看 IP 地址
hostname -I
# 或
ip addr show | grep "inet " | grep -v 127.0.0.1
```

假设服务器 IP 是：`192.168.1.100`

### 步骤 2：启动 Caddy（推荐）

```bash
# 使用一键脚本
./start-caddy.sh

# 或手动启动
sudo caddy run --config Caddyfile.local
```

### 步骤 3：在 Android Chrome 访问

打开 Chrome 浏览器，输入：

```
http://192.168.1.100
```

✅ **完成！**

---

## ⚠️ 麦克风权限问题

### 问题说明

Android Chrome **要求 HTTPS** 才能使用麦克风（除了 localhost）。

HTTP 访问会提示：
```
"此网站需要 HTTPS 才能访问麦克风"
```

### 解决方案

有 3 种方法解决：

---

## 方式一：使用自签名 HTTPS（推荐）⭐

### 1. 启动 HTTPS 模式

```bash
# 使用 Homelab 配置
sudo caddy run --config Caddyfile.homelab
```

### 2. 修改 Caddyfile.homelab

```bash
nano Caddyfile.homelab
```

确保包含你的服务器 IP：

```caddy
https://192.168.1.100, talk.home.wbsays.com {
    tls internal  # 自动生成自签名证书

    root * /var/www/talk-web

    handle /api/* {
        reverse_proxy localhost:8080
    }

    try_files {path} /index.html
    file_server
}
```

### 3. Android 访问并信任证书

1. 在 Chrome 打开：`https://192.168.1.100`
2. 会看到"您的连接不是私密连接"警告
3. 点击 **"高级"**
4. 点击 **"继续前往 192.168.1.100（不安全）"**
5. ✅ 现在可以使用麦克风了！

**优点：**
- ✅ 支持麦克风
- ✅ 配置简单
- ✅ 局域网内所有设备可用

**缺点：**
- ⚠️ 每次访问都有安全警告（点击"高级"跳过）

---

## 方式二：使用 mkcert 受信任证书（最佳）⭐⭐⭐

### 在服务器上操作：

#### 1. 安装 mkcert

```bash
# 下载 mkcert
wget https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
chmod +x mkcert-v1.4.4-linux-amd64
sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert

# 安装本地 CA
mkcert -install
```

#### 2. 生成证书

```bash
# 获取服务器 IP
SERVER_IP=$(hostname -I | awk '{print $1}')

# 生成证书（包含 IP 和域名）
mkcert $SERVER_IP talk.home.wbsays.com localhost 127.0.0.1

# 会生成两个文件，例如：
# 192.168.1.100+3.pem
# 192.168.1.100+3-key.pem
```

#### 3. 配置 Caddy 使用证书

创建或修改 `Caddyfile.android`：

```caddy
https://192.168.1.100 {
    tls ./192.168.1.100+3.pem ./192.168.1.100+3-key.pem

    root * /var/www/talk-web

    handle /api/* {
        reverse_proxy localhost:8080
    }

    try_files {path} /index.html
    file_server
}
```

#### 4. 启动 Caddy

```bash
sudo caddy run --config Caddyfile.android
```

### 在 Android 上操作：

#### 5. 导出 CA 证书

在服务器上：

```bash
# 找到 CA 证书位置
mkcert -CAROOT

# 复制证书到可访问位置
cp "$(mkcert -CAROOT)/rootCA.pem" ~/rootCA.crt

# 启动简单 HTTP 服务器供下载
cd ~
python3 -m http.server 8000
```

#### 6. 在 Android 安装证书

1. 在 Android Chrome 访问：`http://192.168.1.100:8000/rootCA.crt`
2. 下载证书
3. 打开 Android **设置** → **安全** → **加密与凭据** → **安装证书**
4. 选择 **CA 证书**
5. 找到下载的 `rootCA.crt` 并安装
6. 输入锁屏密码确认

#### 7. 访问应用

现在访问 `https://192.168.1.100`：
- ✅ 无警告
- ✅ 绿色小锁
- ✅ 麦克风权限正常

**优点：**
- ✅ 完全受信任的 HTTPS
- ✅ 无安全警告
- ✅ 最佳用户体验

**缺点：**
- ⚠️ 需要在 Android 安装证书（一次性操作）

---

## 方式三：使用域名 + DNS（专业）

### 1. 配置路由器 DNS

在路由器管理界面添加：

```
主机名: talk
域名: home.wbsays.com
IP: 192.168.1.100
```

### 2. 修改 Caddyfile

```caddy
https://talk.home.wbsays.com {
    tls internal
    # 配置...
}
```

### 3. Android 访问

访问：`https://talk.home.wbsays.com`

---

## 🔥 一键配置脚本（推荐使用）

创建 `setup-android.sh`：

```bash
#!/bin/bash

echo "📱 Android 访问配置向导"
echo "======================"
echo ""

# 获取服务器 IP
SERVER_IP=$(hostname -I | awk '{print $1}')
echo "服务器 IP: $SERVER_IP"
echo ""

echo "选择配置方式:"
echo "1) 自签名 HTTPS（简单，有警告）"
echo "2) mkcert 证书（最佳，需安装 CA）"
echo "3) HTTP only（不支持麦克风）"
echo ""
read -p "请选择 (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo "使用自签名 HTTPS..."

        # 创建配置
        cat > Caddyfile.android <<EOF
https://$SERVER_IP, https://talk.home.wbsays.com {
    tls internal

    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /* {
        reverse_proxy localhost:5173
    }
}
EOF

        sudo caddy run --config Caddyfile.android &

        echo ""
        echo "✓ Caddy 已启动"
        echo ""
        echo "📱 Android 访问步骤："
        echo "1. 打开 Chrome"
        echo "2. 访问: https://$SERVER_IP"
        echo "3. 点击 '高级' → '继续访问'"
        echo ""
        ;;

    2)
        echo ""
        echo "安装 mkcert..."

        if ! command -v mkcert &> /dev/null; then
            wget -q https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
            chmod +x mkcert-v1.4.4-linux-amd64
            sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert
        fi

        mkcert -install

        echo "生成证书..."
        mkcert $SERVER_IP talk.home.wbsays.com localhost

        CERT_FILE="${SERVER_IP}+3.pem"
        KEY_FILE="${SERVER_IP}+3-key.pem"

        # 创建配置
        cat > Caddyfile.android <<EOF
https://$SERVER_IP {
    tls ./$CERT_FILE ./$KEY_FILE

    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /* {
        reverse_proxy localhost:5173
    }
}
EOF

        sudo caddy run --config Caddyfile.android &

        # 准备 CA 证书供下载
        cp "$(mkcert -CAROOT)/rootCA.pem" ~/rootCA.crt
        cd ~ && python3 -m http.server 8000 &

        echo ""
        echo "✓ Caddy 已启动"
        echo "✓ HTTP 服务器已启动（端口 8000）"
        echo ""
        echo "📱 Android 配置步骤："
        echo "1. 下载证书: http://$SERVER_IP:8000/rootCA.crt"
        echo "2. 设置 → 安全 → 加密与凭据 → 安装证书 → CA 证书"
        echo "3. 选择下载的 rootCA.crt"
        echo "4. 访问: https://$SERVER_IP"
        echo ""
        ;;

    3)
        echo ""
        echo "使用 HTTP 模式（不支持麦克风）"
        sudo caddy run --config Caddyfile.local &

        echo ""
        echo "✓ Caddy 已启动"
        echo ""
        echo "📱 Android 访问:"
        echo "http://$SERVER_IP"
        echo ""
        echo "⚠️  注意: HTTP 模式下麦克风功能不可用"
        ;;
esac

echo ""
echo "按 Ctrl+C 停止服务"
wait
```

使用：

```bash
chmod +x setup-android.sh
./setup-android.sh
```

---

## 🔧 故障排查

### 问题 1: 无法访问

```bash
# 检查防火墙
sudo ufw status
sudo ufw allow from 192.168.1.0/24

# 检查服务
./status.sh

# 检查端口
sudo netstat -tlnp | grep -E "(80|443|8080|5173)"
```

### 问题 2: 麦克风不工作

1. 确认使用 **HTTPS** 访问
2. 检查 Chrome 权限：设置 → 网站设置 → 麦克风
3. 清除网站数据重试

### 问题 3: 证书警告

- 自签名证书正常会有警告
- 点击"高级" → "继续访问"即可
- 或使用 mkcert 方式解决

---

## 📋 快速参考

| 访问方式 | 地址 | 麦克风 | 警告 |
|---------|------|--------|------|
| HTTP | `http://192.168.1.100` | ❌ | ❌ |
| HTTPS (自签名) | `https://192.168.1.100` | ✅ | ⚠️ |
| HTTPS (mkcert) | `https://192.168.1.100` | ✅ | ✅ |

---

## 🎯 推荐配置

**日常使用：** 方式一（自签名 HTTPS）
- 配置简单
- 一次警告后正常使用

**最佳体验：** 方式二（mkcert）
- 一次性配置
- 完美的 HTTPS 体验

---

需要帮助？查看服务器日志：
```bash
tail -f /var/log/caddy/talk-web.log
```
