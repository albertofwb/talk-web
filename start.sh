#!/bin/bash

echo "🚀 启动 talk-web 项目"
echo ""

# 检查数据库
echo "📊 检查数据库状态..."
if ! docker ps | grep -q talk-web-postgres; then
    echo "启动数据库..."
    docker-compose up -d
    sleep 3
fi

# 启动后端
echo "🔧 启动 Go 后端 (端口 8080)..."
cd server
nohup go run main.go > ../logs/server.log 2>&1 &
SERVER_PID=$!
echo "后端 PID: $SERVER_PID"
cd ..

# 等待后端启动
echo "等待后端启动..."
for i in {1..20}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✓ 后端启动成功"
        break
    fi
    sleep 0.5
done

# 启动前端
echo "🎨 启动 React 前端 (端口 5173)..."
cd web
if [ ! -d "node_modules" ]; then
    echo "安装前端依赖..."
    npm install
fi
npm run dev

# 清理
trap 'kill $SERVER_PID 2>/dev/null' EXIT
