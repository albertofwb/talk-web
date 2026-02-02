# talk-web 项目计划

语音对讲 Web 应用 - 按住说话，松开上传，支持用户认证和管理。

## 架构

```
┌─────────────────────────────────────────────┐
│                  talk-web                    │
├──────────────────┬──────────────────────────┤
│     /web         │        /server           │
│     React        │        Go (Gin+GORM)     │
│                  │                          │
│  - 登录页        │  - Auth API              │
│  - 录音界面      │  - 用户管理 (Admin)      │
│  - 管理后台      │  - 音频上传接收          │
└──────────────────┴──────────────────────────┘
                      │
         ┌────────────┼────────────┐
         ▼            ▼            ▼
     PostgreSQL    Redis      talk-server
     (用户/配置)   (Session)   (STT+路由)
```

## 目录结构

```
talk-web/
├── server/                 # Go 后端
│   ├── main.go
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── auth.go        # 登录/登出
│   │   ├── admin.go       # 用户管理
│   │   └── upload.go      # 音频上传
│   ├── model/
│   │   └── user.go
│   ├── middleware/
│   │   └── jwt.go
│   └── go.mod
├── web/                    # React 前端
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Login.tsx
│   │   │   ├── Talk.tsx   # 主录音页
│   │   │   └── Admin.tsx  # 用户管理
│   │   ├── components/
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml      # PG + Redis
├── Dockerfile
├── Makefile
└── README.md
```

## API 设计

### 认证
```
POST   /api/auth/login      # 登录 → JWT
POST   /api/auth/logout     # 登出
GET    /api/auth/me         # 当前用户信息
```

### 核心功能
```
POST   /api/upload          # 上传音频 (需认证)
```

### 管理后台
```
GET    /api/admin/users     # 用户列表
POST   /api/admin/users     # 创建用户
PUT    /api/admin/users/:id # 修改用户
DELETE /api/admin/users/:id # 删除用户
```

## 数据模型

### User
```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"` // bcrypt hash
    IsAdmin   bool      `gorm:"default:false"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## 技术栈

### 后端
- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- JWT (认证)
- bcrypt (密码哈希)

### 前端
- React 18
- Vite
- TypeScript
- TailwindCSS
- MediaRecorder API (录音)

### 基础设施
- PostgreSQL 15
- Redis 7
- Docker Compose

## 开发阶段

| 阶段 | 内容 | 预计时间 | 状态 |
|------|------|----------|------|
| 1 | 项目骨架 + Docker + 数据库 | 30min | ⬜ |
| 2 | Auth API (登录/JWT/用户管理) | 1h | ⬜ |
| 3 | React 登录页 + 录音页 | 1.5h | ⬜ |
| 4 | Admin 用户管理页面 | 1h | ⬜ |
| 5 | 对接 talk-server (转发音频) | 30min | ⬜ |

## 运行方式

### 开发环境
```bash
# 启动数据库
docker-compose up -d

# 后端
cd server && go run main.go

# 前端
cd web && npm run dev
```

### 生产环境
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 环境变量

```env
# Server
DB_HOST=localhost
DB_PORT=5432
DB_USER=talk
DB_PASSWORD=talk
DB_NAME=talk
REDIS_ADDR=localhost:6379
JWT_SECRET=your-secret-key
TALK_SERVER_URL=http://localhost:5000

# Web (build time)
VITE_API_URL=/api
```
