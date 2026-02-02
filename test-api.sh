#!/bin/bash

echo "🧪 测试 talk-web API"
echo "===================="
echo ""

# 测试健康检查
echo "1️⃣ 测试健康检查..."
curl -s http://localhost:8080/health | jq .
echo ""

# 测试登录
echo "2️⃣ 测试登录 (admin/admin123)..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r .token)
echo $LOGIN_RESPONSE | jq .
echo ""

# 测试获取当前用户信息
echo "3️⃣ 测试获取用户信息..."
curl -s http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 测试获取用户列表
echo "4️⃣ 测试获取用户列表..."
curl -s http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 测试创建用户
echo "5️⃣ 测试创建用户 (testuser)..."
curl -s -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","is_admin":false}' | jq .
echo ""

# 再次获取用户列表
echo "6️⃣ 验证用户已创建..."
curl -s http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "✅ API 测试完成！"
echo ""
echo "📊 服务状态："
echo "  - 后端: http://localhost:8080"
echo "  - 前端: http://localhost:5173"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6380"
