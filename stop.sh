#!/bin/bash

echo "⏹️  停止 talk-web 服务"
echo ""

# 停止前端
echo "停止前端..."
pkill -f "npm run dev" 2>/dev/null && echo "✓ 前端已停止" || echo "前端未运行"

# 停止后端
echo "停止后端..."
pkill -f "go.*main.go" 2>/dev/null && echo "✓ 后端已停止" || echo "后端未运行"

# 询问是否停止数据库
echo ""
read -p "是否停止数据库服务? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose down
    echo "✓ 数据库已停止"
fi

echo ""
echo "✅ 服务已停止"
