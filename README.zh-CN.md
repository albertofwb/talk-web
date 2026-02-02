# talk-web - 语音对讲 Web 应用

<div align="center">

🎤 **按住说话 · 松开发送 · 智能识别**

一个现代化的语音对讲 Web 应用，支持实时录音、自动上传和语音识别

[功能特性](#功能特性) · [快速开始](#快速开始) · [在线演示](#在线演示) · [文档](#文档)

</div>

---

## 📸 预览

```
┌─────────────────────────────────────┐
│          语音对讲系统                │
│                                     │
│     [     按住说话     ]             │
│                                     │
│  🎤 鼠标按住录音，松开发送            │
│  📱 触摸屏按住录音，松开发送          │
└─────────────────────────────────────┘
```

## ✨ 功能特性

### 核心功能
- 🎙️ **实时录音** - 按住按钮即可录音，松开自动上传
- 🔄 **自动处理** - 音频自动转发到语音识别服务（STT）
- 🔐 **用户认证** - 安全的 JWT 认证机制
- 👥 **用户管理** - 管理员可管理所有用户
- 📱 **响应式设计** - 完美支持桌面和移动端

### 技术亮点
- ⚡ **快速部署** - 5 分钟快速启动
- 🐳 **Docker 支持** - 一键启动数据库服务
- 🔧 **开发友好** - 热重载、完整日志
- 📊 **生产就绪** - 完善的错误处理和日志

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

### 一键启动（推荐）

```bash
# 1. 克隆仓库
git clone <your-repo-url>
cd talk-web

# 2. 安装依赖
make init

# 3. 启动数据库
make up

# 4. 启动服务（自动启动前后端）
./start.sh
```

### 分步启动

```bash
# 1. 启动数据库
docker-compose up -d

# 2. 启动后端（终端 1）
cd server
go run main.go

# 3. 启动前端（终端 2）
cd web
npm install
npm run dev
```

### 访问应用

- 🌐 **前端**: http://localhost:5173
- 🔧 **API**: http://localhost:8080
- 👤 **默认账号**: `admin` / `admin123`

## 📖 使用指南

### 1. 登录系统
1. 打开浏览器访问 http://localhost:5173
2. 使用默认账号 `admin / admin123` 登录

### 2. 开始录音
1. 按住页面中央的录音按钮
2. 对着麦克风说话
3. 松开按钮，音频自动上传并处理

### 3. 管理用户（管理员）
1. 点击右上角"管理后台"
2. 可以创建、编辑、删除用户
3. 可以设置用户管理员权限

## 🏗️ 技术架构

```
┌─────────────────────────────────────────┐
│              talk-web                   │
├──────────────┬──────────────────────────┤
│   Frontend   │        Backend           │
│   React 18   │   Go + Gin + GORM        │
│              │                          │
│ - 登录页     │ - JWT 认证               │
│ - 录音界面   │ - 用户管理               │
│ - 管理后台   │ - 音频转发               │
└──────────────┴──────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
   PostgreSQL     Redis      Talk-Server
   (数据存储)   (会话管理)    (语音识别)
```

### 技术栈

**后端**
- Go 1.21
- Gin Web Framework
- GORM (PostgreSQL)
- JWT 认证
- bcrypt 密码加密

**前端**
- React 18
- TypeScript
- Vite
- TailwindCSS
- Axios
- MediaRecorder API

**基础设施**
- PostgreSQL 15
- Redis 7
- Docker Compose

## 📁 项目结构

```
talk-web/
├── server/              # Go 后端
│   ├── main.go         # 入口文件
│   ├── config/         # 配置管理
│   ├── handler/        # API 处理器
│   ├── model/          # 数据模型
│   └── middleware/     # 中间件
│
├── web/                # React 前端
│   ├── src/
│   │   ├── pages/      # 页面组件
│   │   ├── utils/      # 工具函数
│   │   └── App.tsx     # 主应用
│   └── package.json
│
├── docker-compose.yml  # 数据库服务
├── Makefile           # 快捷命令
└── README.md          # 项目文档
```

## 🔧 配置说明

### 环境变量

复制 `.env.example` 为 `.env`：

```bash
cp .env.example .env
```

主要配置项：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=talk
DB_PASSWORD=talk
DB_NAME=talk

# Redis 配置
REDIS_ADDR=localhost:6380

# JWT 密钥（生产环境必须修改！）
JWT_SECRET=your-secret-key-change-in-production

# Talk Server 地址（语音识别服务）
TALK_SERVER_URL=http://localhost:5000

# 服务端口
PORT=8080
```

## 📋 API 文档

### 认证接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/auth/login` | 用户登录 | ❌ |
| POST | `/api/auth/logout` | 用户登出 | ✅ |
| GET | `/api/auth/me` | 获取当前用户信息 | ✅ |

### 音频接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/upload` | 上传音频文件 | ✅ |

### 管理接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/admin/users` | 获取用户列表 | 🔐 管理员 |
| POST | `/api/admin/users` | 创建新用户 | 🔐 管理员 |
| PUT | `/api/admin/users/:id` | 更新用户信息 | 🔐 管理员 |
| DELETE | `/api/admin/users/:id` | 删除用户 | 🔐 管理员 |

## 🛠️ 开发工具

### 快捷脚本

```bash
./status.sh     # 查看服务状态
./start.sh      # 一键启动所有服务
./stop.sh       # 停止所有服务
./test-api.sh   # 测试 API 接口
```

### Makefile 命令

```bash
make init       # 安装所有依赖
make up         # 启动数据库
make down       # 停止数据库
make server     # 启动后端
make web        # 启动前端
make logs       # 查看数据库日志
make clean      # 清理所有数据
```

## 📊 数据模型

### User (用户)

```go
type User struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`     // 用户名（唯一）
    Password  string    `json:"-"`            // 密码（bcrypt 哈希）
    IsAdmin   bool      `json:"is_admin"`     // 是否管理员
    CreatedAt time.Time `json:"created_at"`   // 创建时间
    UpdatedAt time.Time `json:"updated_at"`   // 更新时间
}
```

## 🚢 部署指南

详细部署文档请查看 [DEPLOY.md](./DEPLOY.md)

### 快速部署（Docker Compose）

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d
```

### 安全建议

⚠️ **生产环境必须做的事情：**

1. ✅ 修改 `JWT_SECRET` 为强随机密钥
2. ✅ 修改数据库密码
3. ✅ 修改默认管理员密码
4. ✅ 启用 HTTPS（麦克风权限要求）
5. ✅ 配置防火墙规则
6. ✅ 定期备份数据库

## 📝 更新日志

### v0.1.0 (2026-02-02)

**初始版本**
- ✅ 用户认证系统
- ✅ 语音录音功能
- ✅ 音频上传转发
- ✅ 用户管理后台
- ✅ Docker 支持
- ✅ 完整文档

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 💬 支持

- 📧 Email: your-email@example.com
- 🐛 Issues: [GitHub Issues](your-repo-url/issues)
- 📚 文档: [完整文档](./DEPLOY.md)

## 🙏 致谢

感谢所有贡献者和使用者！

---

<div align="center">
Made with ❤️ by Your Team
</div>
