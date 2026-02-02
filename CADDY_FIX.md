# Caddy WebSocket 配置修复

## 问题
之前的配置中添加了不必要的 header_up 配置，导致 WebSocket 握手失败（HTTP 400）

## 原因
Caddy 2 默认支持 WebSocket，会自动处理 Upgrade 和 Connection headers。
手动配置 header_up 反而干扰了正常的 WebSocket 握手过程。

## 修复
移除了以下配置：
```
header_up Upgrade {>Upgrade}
header_up Connection {>Connection}
```

只需要简单的 reverse_proxy 配置即可：
```
handle /api/ws {
    reverse_proxy localhost:8080
}
```

## 测试
使用 Python websockets 库测试通过：
- 直接连接: ws://localhost:8080/api/ws ✓
- Caddy 代理: wss://talk.home.wbsays.com/api/ws ✓
