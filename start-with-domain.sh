#!/bin/bash

echo "🌐 启动 talk-web (域名访问)"
echo "============================"
echo ""

# 检查服务
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "启动后端..."
    cd /home/albert/talk-web/server
    nohup go run main.go > ../logs/server.log 2>&1 &
    sleep 2
fi

if ! curl -s http://localhost:5174 > /dev/null 2>&1; then
    echo "启动前端..."
    cd /home/albert/talk-web/web
    nohup npm run dev > ../logs/web.log 2>&1 &
    sleep 3
fi

# 创建日志目录
sudo mkdir -p /var/log/caddy
sudo chown $USER:$USER /var/log/caddy

echo "✓ 服务已启动"
echo ""

# 检查 Caddy
if ! command -v caddy &> /dev/null; then
    echo "❌ Caddy 未安装"
    echo ""
    echo "安装方法:"
    echo "  sudo apt install caddy"
    exit 1
fi

echo "启动 Caddy..."
sudo caddy start --config /home/albert/talk-web/Caddyfile.domain

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 启动完成！"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📱 访问地址:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  http://talk.home.wbsays.com"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "💡 提示:"
    echo "  - 局域网内所有设备可访问"
    echo "  - 无需输入端口号"
    echo "  - 登录账号: admin / admin123"
    echo ""
    echo "🔧 管理命令:"
    echo "  查看状态: sudo caddy status"
    echo "  停止服务: sudo caddy stop"
    echo "  查看日志: tail -f /var/log/caddy/talk-domain.log"
    echo ""
else
    echo "❌ Caddy 启动失败"
fi
