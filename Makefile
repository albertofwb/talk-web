.PHONY: help init up down server run logs clean build deploy

help:
	@echo "talk-web 项目管理"
	@echo ""
	@echo "可用命令:"
	@echo "  make init    - 初始化项目（安装依赖）"
	@echo "  make up      - 启动数据库服务"
	@echo "  make down    - 停止数据库服务"
	@echo "  make server  - 启动 Go 后端"
	@echo "  make run     - 启动 React 前端"
	@echo "  make build   - 构建前端生产版本"
	@echo "  make deploy  - 部署到生产环境（构建前端 + 重启服务）"
	@echo "  make logs    - 查看数据库日志"
	@echo "  make clean   - 清理数据"

init:
	@echo "📦 安装 Go 依赖..."
	cd server && go mod download
	@echo "📦 安装 Node 依赖..."
	cd web && npm install
	@echo "✓ 依赖安装完成"

up:
	@echo "🚀 启动数据库服务..."
	docker-compose up -d
	@echo "✓ 数据库已启动"
	@echo ""
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

down:
	@echo "⏹️  停止数据库服务..."
	docker-compose down
	@echo "✓ 数据库已停止"

server:
	@echo "🚀 启动 Go 后端 (端口 8080)..."
	cd server && go run main.go

run:
	@echo "🚀 启动 React 前端 (端口 5173)..."
	cd web && npm run dev

build:
	@echo "📦 构建前端生产版本..."
	cd web && npm run build
	@echo "✓ 前端构建完成: web/dist/"

deploy: build
	@echo "🚀 部署到生产环境..."
	sudo systemctl restart talk-web
	@echo "✓ 服务已重启"
	@echo ""
	@echo "访问地址: https://talk.home.wbsays.com"

logs:
	docker-compose logs -f

clean:
	@echo "🗑️  清理数据..."
	docker-compose down -v
	@echo "✓ 数据已清理"
