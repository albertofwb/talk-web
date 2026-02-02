#!/bin/bash

# Talk-Web 服务管理脚本

case "$1" in
    start)
        echo "🚀 启动 talk-web 服务..."
        sudo systemctl start talk-web
        sudo systemctl status talk-web --no-pager
        ;;
    stop)
        echo "🛑 停止 talk-web 服务..."
        sudo systemctl stop talk-web
        ;;
    restart)
        echo "🔄 重启 talk-web 服务..."
        sudo systemctl restart talk-web
        sleep 2
        sudo systemctl status talk-web --no-pager
        ;;
    status)
        echo "📊 talk-web 服务状态:"
        sudo systemctl status talk-web --no-pager
        ;;
    logs)
        echo "📋 talk-web 服务日志:"
        if [ -n "$2" ]; then
            sudo journalctl -u talk-web -n "$2" --no-pager
        else
            sudo journalctl -u talk-web -n 50 --no-pager
        fi
        ;;
    follow)
        echo "📋 实时跟踪日志 (Ctrl+C 退出):"
        sudo journalctl -u talk-web -f
        ;;
    enable)
        echo "✅ 启用开机自启动..."
        sudo systemctl enable talk-web
        ;;
    disable)
        echo "❌ 禁用开机自启动..."
        sudo systemctl disable talk-web
        ;;
    *)
        echo "Talk-Web 服务管理"
        echo ""
        echo "用法: $0 {start|stop|restart|status|logs|follow|enable|disable}"
        echo ""
        echo "命令说明:"
        echo "  start    - 启动服务"
        echo "  stop     - 停止服务"
        echo "  restart  - 重启服务"
        echo "  status   - 查看状态"
        echo "  logs     - 查看日志 (最近50条)"
        echo "  logs N   - 查看最近N条日志"
        echo "  follow   - 实时跟踪日志"
        echo "  enable   - 启用开机自启动"
        echo "  disable  - 禁用开机自启动"
        exit 1
        ;;
esac
