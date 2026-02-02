#!/bin/bash

echo "📱 Android 访问配置向导"
echo "======================"
echo ""

# 获取服务器 IP
SERVER_IP=$(hostname -I | awk '{print $1}')
echo "✓ 服务器 IP: $SERVER_IP"
echo ""

# 检查服务状态
echo "检查服务状态..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  后端未运行"
    read -p "是否启动后端? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        cd server && nohup go run main.go > ../logs/server.log 2>&1 &
        sleep 2
        echo "✓ 后端已启动"
    fi
fi

if ! curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "⚠️  前端未运行"
    read -p "是否启动前端? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        cd web && nohup npm run dev > ../logs/web.log 2>&1 &
        sleep 3
        echo "✓ 前端已启动"
    fi
fi

echo ""
echo "配置方式选择:"
echo "1) 自签名 HTTPS - 简单快速（推荐）"
echo "2) mkcert 证书 - 最佳体验（需安装 CA）"
echo "3) HTTP only - 不支持麦克风"
echo ""
read -p "请选择 (1/2/3): " -n 1 -r choice
echo ""

case $choice in
    1)
        echo ""
        echo "=== 配置自签名 HTTPS ==="
        echo ""

        # 创建配置
        cat > Caddyfile.android <<EOF
# Android 访问配置 - 自签名 HTTPS
https://$SERVER_IP, https://talk.home.wbsays.com {
    tls internal {
        on_demand
    }

    # API 代理
    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /health {
        reverse_proxy localhost:8080
    }

    # 前端代理
    handle /* {
        reverse_proxy localhost:5173
    }

    # 日志
    log {
        output file /var/log/caddy/android.log
    }
}

# HTTP 跳转
http://$SERVER_IP {
    redir https://$SERVER_IP{uri}
}
EOF

        echo "✓ 配置文件已创建: Caddyfile.android"
        echo ""

        # 创建日志目录
        sudo mkdir -p /var/log/caddy
        sudo chown $USER:$USER /var/log/caddy

        echo "启动 Caddy..."
        sudo caddy start --config Caddyfile.android

        if [ $? -eq 0 ]; then
            echo ""
            echo "✅ Caddy 已启动！"
            echo ""
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "📱 Android 访问步骤:"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""
            echo "1️⃣  在 Android Chrome 打开:"
            echo "   https://$SERVER_IP"
            echo ""
            echo "2️⃣  会看到安全警告，点击:"
            echo "   '高级' → '继续前往 $SERVER_IP (不安全)'"
            echo ""
            echo "3️⃣  完成！可以正常使用麦克风了"
            echo ""
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""
            echo "💡 提示:"
            echo "  - 首次访问需要接受证书警告"
            echo "  - 之后访问不会再提示"
            echo "  - 麦克风权限正常工作"
            echo ""
        else
            echo "❌ Caddy 启动失败"
            exit 1
        fi
        ;;

    2)
        echo ""
        echo "=== 配置 mkcert 证书 ==="
        echo ""

        # 检查 mkcert
        if ! command -v mkcert &> /dev/null; then
            echo "安装 mkcert..."
            wget -q https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
            chmod +x mkcert-v1.4.4-linux-amd64
            sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert
            echo "✓ mkcert 已安装"
        fi

        echo "初始化 mkcert..."
        mkcert -install

        echo "生成证书..."
        mkcert $SERVER_IP talk.home.wbsays.com localhost 127.0.0.1

        # 查找证书文件
        CERT_FILE=$(ls ${SERVER_IP}+*.pem 2>/dev/null | head -1)
        KEY_FILE=$(ls ${SERVER_IP}+*-key.pem 2>/dev/null | head -1)

        if [ -z "$CERT_FILE" ] || [ -z "$KEY_FILE" ]; then
            echo "❌ 证书生成失败"
            exit 1
        fi

        echo "✓ 证书已生成: $CERT_FILE"

        # 创建配置
        cat > Caddyfile.android <<EOF
# Android 访问配置 - mkcert 证书
https://$SERVER_IP {
    tls ./$CERT_FILE ./$KEY_FILE

    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /health {
        reverse_proxy localhost:8080
    }

    handle /* {
        reverse_proxy localhost:5173
    }

    log {
        output file /var/log/caddy/android.log
    }
}
EOF

        # 创建日志目录
        sudo mkdir -p /var/log/caddy
        sudo chown $USER:$USER /var/log/caddy

        echo "启动 Caddy..."
        sudo caddy start --config Caddyfile.android

        # 准备 CA 证书供下载
        CA_ROOT=$(mkcert -CAROOT)
        cp "$CA_ROOT/rootCA.pem" ~/rootCA.crt

        echo ""
        echo "启动 HTTP 服务器（用于下载证书）..."
        cd ~ && python3 -m http.server 8000 > /dev/null 2>&1 &
        HTTP_PID=$!

        echo ""
        echo "✅ 配置完成！"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "📱 Android 安装证书步骤:"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "1️⃣  在 Android Chrome 下载证书:"
        echo "   http://$SERVER_IP:8000/rootCA.crt"
        echo ""
        echo "2️⃣  安装证书:"
        echo "   设置 → 安全 → 加密与凭据"
        echo "   → 安装证书 → CA 证书"
        echo "   → 选择 rootCA.crt"
        echo ""
        echo "3️⃣  访问应用:"
        echo "   https://$SERVER_IP"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "💡 提示:"
        echo "  - 证书只需安装一次"
        echo "  - 无任何安全警告"
        echo "  - 完美的 HTTPS 体验"
        echo ""
        echo "🗑️  证书下载后按任意键停止 HTTP 服务器..."
        read -n 1
        kill $HTTP_PID 2>/dev/null
        echo "✓ HTTP 服务器已停止"
        ;;

    3)
        echo ""
        echo "=== 配置 HTTP 模式 ==="
        echo ""
        echo "⚠️  注意: HTTP 模式下麦克风功能不可用！"
        echo ""

        sudo caddy start --config Caddyfile.local

        echo ""
        echo "✅ Caddy 已启动（HTTP 模式）"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "📱 Android 访问:"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "   http://$SERVER_IP"
        echo ""
        echo "⚠️  限制: 无法使用麦克风功能"
        echo ""
        ;;

    *)
        echo "无效选择"
        exit 1
        ;;
esac

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 管理命令:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "查看状态:  ./status.sh"
echo "查看日志:  sudo caddy logs"
echo "停止服务:  sudo caddy stop"
echo "重启服务:  sudo caddy reload --config Caddyfile.android"
echo ""
