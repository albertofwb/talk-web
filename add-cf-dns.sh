#!/bin/bash

echo "📡 添加 Cloudflare DNS 记录 - talk.home.wbsays.com"
echo "================================================="
echo ""

# 配置
CF_API_TOKEN="BUk-kU7WhADREbhaz6RBkGLVjz9CqdElSwZ6Dfnb"
ZONE_ID="bb014402ba003151b6a9d3f3a95d2016"
RECORD_NAME="talk.home.wbsays.com"
LAN_IP="192.168.1.101"

echo "配置信息:"
echo "  域名: $RECORD_NAME"
echo "  IP: $LAN_IP (局域网)"
echo ""

# 检查记录是否存在
echo "检查现有记录..."
RESPONSE=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records?name=$RECORD_NAME&type=A" \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -H "Content-Type: application/json")

RECORD_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$RECORD_ID" ]; then
    echo "✓ 找到现有记录: $RECORD_ID"
    echo ""
    echo "更新记录..."

    # 更新记录
    RESULT=$(curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$RECORD_ID" \
      -H "Authorization: Bearer $CF_API_TOKEN" \
      -H "Content-Type: application/json" \
      --data "{
        \"type\": \"A\",
        \"name\": \"$RECORD_NAME\",
        \"content\": \"$LAN_IP\",
        \"ttl\": 300,
        \"proxied\": false
      }")
else
    echo "记录不存在，创建新记录..."
    echo ""

    # 创建记录
    RESULT=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records" \
      -H "Authorization: Bearer $CF_API_TOKEN" \
      -H "Content-Type: application/json" \
      --data "{
        \"type\": \"A\",
        \"name\": \"$RECORD_NAME\",
        \"content\": \"$LAN_IP\",
        \"ttl\": 300,
        \"proxied\": false
      }")
fi

# 检查结果
if echo "$RESULT" | grep -q '"success":true'; then
    echo "✅ DNS 记录配置成功！"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📋 配置详情:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  类型: A 记录"
    echo "  域名: $RECORD_NAME"
    echo "  IP: $LAN_IP"
    echo "  TTL: 300 秒"
    echo "  代理: 关闭 (DNS Only)"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "⏱️  等待 DNS 生效（约 1-5 分钟）..."
    echo ""
    echo "测试命令:"
    echo "  dig talk.home.wbsays.com"
    echo "  nslookup talk.home.wbsays.com"
    echo ""
    echo "访问地址:"
    echo "  http://talk.home.wbsays.com:5174"
    echo ""
else
    echo "❌ 配置失败"
    echo ""
    echo "错误信息:"
    echo "$RESULT" | grep -o '"message":"[^"]*"' || echo "$RESULT"
    echo ""
fi
