#!/bin/bash

echo "🌐 启动 Caddy - talk-web"
echo "========================"
echo ""

# 检查 Caddy 是否安装
if ! command -v caddy &> /dev/null; then
    echo "❌ 未安装 Caddy"
    echo ""
    echo "安装方法:"
    echo "Ubuntu/Debian:"
    echo "  sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl"
    echo "  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg"
    echo "  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list"
    echo "  sudo apt update && sudo apt install caddy"
    echo ""
    echo "macOS:"
    echo "  brew install caddy"
    exit 1
fi

# 检查后端是否运行
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  后端未运行，尝试启动..."
    cd server && nohup go run main.go > ../logs/server.log 2>&1 &
    sleep 2
fi

# 检查前端是否运行
if ! curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "⚠️  前端未运行，尝试启动..."
    cd web && nohup npm run dev > ../logs/web.log 2>&1 &
    sleep 3
fi

# 创建日志目录
sudo mkdir -p /var/log/caddy
sudo chown $USER:$USER /var/log/caddy

# 选择配置文件
echo "选择 Caddy 配置:"
echo "1) 开发模式 - HTTP + 代理到开发服务器 (推荐)"
echo "2) Homelab 模式 - 自签名 HTTPS + 静态文件"
echo "3) 本地模式 - HTTP + 静态文件"
echo ""
read -p "请选择 (1/2/3): " -n 1 -r choice
echo ""

case $choice in
    1)
        config_file="Caddyfile"
        echo "使用开发模式配置..."
        ;;
    2)
        config_file="Caddyfile.homelab"
        echo "使用 Homelab 模式配置..."
        echo "⚠️  需要先构建前端: cd web && npm run build"
        ;;
    3)
        config_file="Caddyfile.local"
        echo "使用本地模式配置..."
        ;;
    *)
        config_file="Caddyfile"
        echo "使用默认配置..."
        ;;
esac

# 验证配置
echo ""
echo "验证配置文件..."
if ! caddy validate --config "$config_file" 2>&1; then
    echo "❌ 配置文件验证失败"
    exit 1
fi

echo "✓ 配置文件有效"
echo ""

# 启动 Caddy
echo "启动 Caddy..."
echo ""
echo "访问地址:"
echo "  - http://localhost"
echo "  - http://talk.home.wbsays.com (如果配置了 DNS)"
echo "  - http://$(hostname -I | awk '{print $1}')"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

sudo caddy run --config "$config_file"
