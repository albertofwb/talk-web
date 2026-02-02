import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',  // 监听所有网络接口（支持局域网和 Tailscale 访问）
    port: 5173,
    strictPort: false,  // 如果端口被占用，自动换端口
    allowedHosts: [
      'talk.home.wbsays.com',
      '.home.wbsays.com',  // 允许所有 *.home.wbsays.com 子域名
      '.tail96df5.ts.net',  // 允许 Tailscale 域名
    ],
    hmr: {
      clientPort: 5174,  // HMR 使用的端口
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,  // 启用 WebSocket 代理
      },
    },
  },
})
