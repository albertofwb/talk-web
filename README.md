# talk-web

语音对讲 Web 应用 - 按住说话，自动识别，智能回复。

## 功能特性

- 🎤 **语音录音** - 按住按钮录音（最少1秒），松开自动上传
- 🗣️ **语音识别 (STT)** - 本地 Whisper 模型，隐私保护
- 🤖 **智能对话** - 集成 Telegram Bot，AI 回复
- 🔊 **语音合成 (TTS)** - 自动播放回复语音
- 🔐 **用户认证** - JWT 令牌认证机制
- 👥 **用户管理** - 管理员可创建、编辑、删除用户
- 🎨 **现代 UI** - 响应式设计，支持桌面和移动端
- 🔒 **HTTPS 支持** - Caddy + Cloudflare DNS 自动证书

## 技术栈

### 后端
- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- JWT (认证)
- PostgreSQL (数据库)
- Redis (消息队列 + 会话存储)

### 前端
- React 18
- TypeScript
- Vite
- TailwindCSS
- MediaRecorder API

### AI/语音
- **STT**: faster-whisper (OpenAI Whisper 本地模型)
- **TTS**: edge-tts (Microsoft Edge 语音合成)
- **消息队列**: Redis (tg/th 命令集成)

## 工作流程

```
用户说话 → 前端录音 → 后端 STT 识别
    ↓
发送到 Redis 队列 (tg)
    ↓
Telegram Bot 处理 → AI 回复 → Redis 收件箱
    ↓
后端轮询获取回复 (th) → TTS 生成语音 → 前端播放
```

## 快速开始

### 前置要求

1. **安装依赖**
   ```bash
   # Go 1.21+
   # Node.js 18+
   # PostgreSQL 15+
   # Redis 7+
   # Python 3.10+ (uv)

   # STT/TTS 工具
   pip install faster-whisper edge-tts
   ```

2. **配置 tg/th 命令**（继承现有配置）
   - `tg` - 发送消息到 Telegram Bot
   - `th` - 从 Telegram 获取回复
   - 位置: `~/.local/bin/`

### 启动步骤

#### 1. 初始化项目

```bash
# 安装依赖
make init
```

#### 2. 启动数据库

```bash
# 启动 PostgreSQL 和 Redis
make up
```

#### 3. 启动后端（新终端）

```bash
# 启动 Go 服务 (端口 8080)
cd server
go run main.go
```

#### 4. 启动前端（新终端）

```bash
# 启动 React 开发服务器 (端口 5173)
cd web
npm run dev
```

#### 5. 访问应用

- **开发环境**: http://localhost:5173
- **HTTPS (推荐)**: https://100.118.236.127 或 https://talk.home.wbsays.com

默认管理员账号:
- 用户名: `admin`
- 密码: `c2h5oh.home`

## 环境配置

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：
```env
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=talk
DB_PASSWORD=talk
DB_NAME=talk

# Redis
REDIS_ADDR=localhost:6379

# 认证
JWT_SECRET=your-secret-key

# Telegram Bot（通过 Redis 队列）
TG_RECIPIENT=AlbertClaudeBot    # 默认接收者
TG_USERNAME=WbsaysVoiceBot      # 当前用户收件箱
```

## API 文档

### 认证接口

```
POST   /api/auth/login      # 登录
POST   /api/auth/logout     # 登出
GET    /api/auth/me         # 获取当前用户信息
```

### 核心功能

```
POST   /api/upload          # 上传音频（需认证）
                            # 流程: STT → Telegram → 等待回复 → TTS
GET    /api/audio/:filename # 下载 TTS 生成的音频
```

### 管理后台

```
GET    /api/admin/users     # 用户列表（需管理员）
POST   /api/admin/users     # 创建用户（需管理员）
PUT    /api/admin/users/:id # 修改用户（需管理员）
DELETE /api/admin/users/:id # 删除用户（需管理员）
```

## 项目结构

```
talk-web/
├── server/              # Go 后端
│   ├── main.go
│   ├── config/         # 配置
│   ├── handler/        # API 处理器
│   ├── model/          # 数据模型
│   ├── middleware/     # 中间件
│   └── pkg/           # 核心包
│       ├── stt/       # 语音识别
│       ├── tts/       # 语音合成
│       └── telegram/  # Telegram 集成 (tg/th)
├── web/                # React 前端
│   └── src/
│       ├── pages/      # 页面组件
│       │   ├── Login.tsx
│       │   ├── Talk.tsx    # 主录音页面
│       │   └── Admin.tsx
│       ├── utils/      # 工具函数
│       └── App.tsx
├── docker-compose.yml  # 数据库服务
├── Makefile           # 项目管理
└── README.md
```

## 常用命令

### 快捷脚本
```bash
./status.sh    # 查看服务状态
./start.sh     # 一键启动所有服务
./stop.sh      # 停止所有服务
./test-api.sh  # 测试 API 接口
```

### Makefile 命令
```bash
make help    # 查看所有命令
make init    # 安装依赖
make up      # 启动数据库
make down    # 停止数据库
make server  # 启动后端
make web     # 启动前端
make logs    # 查看数据库日志
make clean   # 清理数据（会删除所有数据）
```

### tg/th 命令（继承现有工具）
```bash
# 发送消息到 Telegram Bot
tg "你好世界"
tg @AlbertClaudeBot "测试消息"

# 查看收件箱
th -l                      # 列出所有对话
th @WbsaysVoiceBot         # 查看消息
th @WbsaysVoiceBot 50      # 查看最近50条
th --clear @WbsaysVoiceBot # 清空收件箱
```

## 使用说明

### 基本使用

1. **登录** - 使用管理员账号登录
2. **授权麦克风** - 首次使用需要授权浏览器访问麦克风（仅一次）
3. **按住录音** - 按住按钮至少 1 秒，说话清晰
4. **松开发送** - 松开按钮自动上传、识别、获取回复
5. **听回复** - 自动播放 TTS 生成的语音回复

### 录音技巧

- ✅ 按住至少 1 秒
- ✅ 说话清晰，环境安静
- ✅ 确保 HTTPS 环境（麦克风权限要求）
- ❌ 不要松手太快
- ❌ 避免嘈杂环境

### 管理用户

管理员可访问 `/admin` 页面：
- 创建新用户
- 修改用户信息
- 删除用户
- 设置管理员权限

## HTTPS 部署

### Caddy 配置（推荐）

已配置 Caddy 反向代理和自动 HTTPS：

1. **Cloudflare DNS** - `talk.home.wbsays.com`
2. **Tailscale HTTPS** - `https://100.118.236.127`
3. **自动证书续期**

配置文件位置：
- `/etc/caddy/sites/talk.home.wbsays.com`
- `/etc/caddy/sites/tailscale-https`

## Telegram 集成原理

### 消息流

1. **发送消息** (tg → Redis)
   ```
   用户文本 → Redis队列 (message_queue)
   格式: {"text": "...", "recipient": "AlbertClaudeBot"}
   ```

2. **后台处理** (Telegram Bot)
   ```
   Bot 从队列读取 → 处理 → 回复到收件箱
   ```

3. **获取回复** (Redis → th)
   ```
   轮询收件箱 (inbox:WbsaysVoiceBot)
   格式: {"text": "...", "sender": "...", "timestamp": "..."}
   ```

### Go 集成

后端直接使用 Redis 客户端，不调用命令行：

```go
import "talk-web/server/pkg/telegram"

// 发送消息
tg := telegram.NewTelegramClient()
tg.SendToTelegram("你好", "AlbertClaudeBot")

// 等待回复（30秒超时）
reply, err := tg.WaitForReply("WbsaysVoiceBot", 30*time.Second)
```

## 开发

### 后端开发

```bash
cd server
go run main.go

# 热重载（可选）
air
```

### 前端开发

```bash
cd web
npm run dev

# 自动热更新
```

### 数据库

GORM 会自动迁移数据库表结构，无需手动操作。

如需重置数据库：
```bash
make clean
make up
```

## 故障排查

### 麦克风权限问题
- **症状**: 浏览器一直弹出权限请求
- **解决**: 使用 HTTPS 访问（推荐 Tailscale HTTPS）

### 录音识别失败
- **症状**: 500 错误，"InvalidDataError"
- **原因**: 录音时间太短或音频损坏
- **解决**: 按住至少 1 秒，等待录音状态变红

### Telegram 回复超时
- **症状**: "timeout waiting for reply"
- **原因**: Bot 没有运行或队列阻塞
- **解决**: 检查 Redis 队列和 Bot 状态

### 音频无法播放
- **症状**: 识别成功但听不到回复
- **原因**: TTS 生成失败或音频路径错误
- **解决**: 检查 `/tmp/xiaoxiao-*.opus` 文件和日志

## 性能优化

- **STT 模型**: 默认 `base`，可升级到 `large-v3` 提高准确度
- **并发控制**: STT 模型加载占内存，建议限制并发数
- **Redis 连接池**: 使用连接池提高性能
- **音频格式**: webm/opus 格式，兼容性好且体积小

## 注意事项

- ⚠️ 首次使用需要授权浏览器访问麦克风
- ⚠️ 必须使用 HTTPS 才能访问麦克风（浏览器安全策略）
- ⚠️ 生产环境务必修改默认的 JWT 密钥和管理员密码
- ⚠️ STT 第一次使用会下载模型（~140MB），需要网络
- ⚠️ TTS 生成的音频文件在 `/tmp` 目录，会定期清理

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- **GitHub**: https://github.com/albertofwb/talk-web
- **OpenAI Whisper**: https://github.com/openai/whisper
- **faster-whisper**: https://github.com/guillaumekln/faster-whisper
- **edge-tts**: https://github.com/rany2/edge-tts
