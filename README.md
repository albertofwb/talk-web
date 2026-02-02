# talk-web

语音对讲 Web 应用 - 按住说话，松开上传，支持用户认证和管理。

## 功能特性

- 🎤 **语音录音** - 按住按钮录音，松开自动上传
- 🔐 **用户认证** - JWT 令牌认证机制
- 👥 **用户管理** - 管理员可创建、编辑、删除用户
- 🎨 **现代 UI** - 响应式设计，支持桌面和移动端
- 🔄 **自动转发** - 音频自动转发到 talk-server 进行 STT 处理

## 技术栈

### 后端
- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- JWT (认证)
- PostgreSQL (数据库)
- Redis (会话存储)

### 前端
- React 18
- TypeScript
- Vite
- TailwindCSS
- MediaRecorder API

## 快速开始

### 1. 初始化项目

```bash
# 安装依赖
make init
```

### 2. 启动数据库

```bash
# 启动 PostgreSQL 和 Redis
make up
```

### 3. 启动后端（新终端）

```bash
# 启动 Go 服务 (端口 8080)
make server
```

### 4. 启动前端（新终端）

```bash
# 启动 React 开发服务器 (端口 5173)
make web
```

### 5. 访问应用

打开浏览器访问: http://localhost:5173

默认管理员账号:
- 用户名: `admin`
- 密码: `admin123`

## 环境配置

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：
- `DB_*`: PostgreSQL 数据库配置
- `REDIS_ADDR`: Redis 地址
- `JWT_SECRET`: JWT 密钥（生产环境务必修改）
- `TALK_SERVER_URL`: 语音识别服务地址

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
│   └── middleware/     # 中间件
├── web/                # React 前端
│   └── src/
│       ├── pages/      # 页面组件
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

## 使用说明

1. **登录** - 使用默认管理员账号或创建的用户账号登录
2. **录音** - 在主页按住录音按钮开始录音，松开自动上传
3. **管理** - 管理员可访问后台创建和管理用户

## 注意事项

- 首次使用需要授权浏览器访问麦克风
- 确保 `TALK_SERVER_URL` 指向正确的 STT 服务
- 生产环境务必修改默认的 JWT 密钥
- 建议使用 HTTPS 部署（麦克风权限要求）

## 开发

### 后端开发

```bash
cd server
go run main.go
```

### 前端开发

```bash
cd web
npm run dev
```

### 数据库迁移

GORM 会自动迁移数据库表结构，无需手动操作。

## 许可证

MIT
