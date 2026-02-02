#!/bin/bash

echo "📊 talk-web 项目状态"
echo "===================="
echo ""

# 检查数据库
echo "🗄️  数据库服务："
if docker ps | grep -q talk-web-postgres; then
    echo "  ✓ PostgreSQL 运行中 (localhost:5432)"
else
    echo "  ✗ PostgreSQL 未运行"
fi

if docker ps | grep -q talk-web-redis; then
    echo "  ✓ Redis 运行中 (localhost:6380)"
else
    echo "  ✗ Redis 未运行"
fi
echo ""

# 检查后端
echo "🔧 后端服务："
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "  ✓ Go 后端运行中 (http://localhost:8080)"
    echo "    - 健康检查: OK"
else
    echo "  ✗ Go 后端未运行"
fi
echo ""

# 检查前端
echo "🎨 前端服务："
if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "  ✓ React 前端运行中 (http://localhost:5173)"
else
    echo "  ✗ React 前端未运行"
fi
echo ""

# 进程信息
echo "📝 运行进程："
ps aux | grep -E "(go.*main.go|npm run dev)" | grep -v grep | awk '{print "  - PID", $2, ":", $11, $12, $13}'
echo ""

# 端口占用
echo "🔌 端口占用："
netstat -tlnp 2>/dev/null | grep -E "(5432|6380|8080|5173)" | awk '{print "  -", $4, $7}' || ss -tlnp | grep -E "(5432|6380|8080|5173)" | awk '{print "  -", $4, $7}'
echo ""

echo "📚 快速命令："
echo "  - 启动所有服务: ./start.sh"
echo "  - 测试 API: ./test-api.sh"
echo "  - 查看日志: tail -f logs/*.log"
echo "  - 停止数据库: make down"
echo ""

echo "🌐 访问地址："
echo "  - 前端页面: http://localhost:5173"
echo "  - 后端 API: http://localhost:8080/api"
echo "  - 健康检查: http://localhost:8080/health"
echo ""

echo "👤 默认账号："
echo "  - 用户名: admin"
echo "  - 密码: admin123"
