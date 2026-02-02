package main

import (
	"fmt"
	"log"
	"talk-web/server/config"
	"talk-web/server/handler"
	"talk-web/server/middleware"
	"talk-web/server/model"
	"talk-web/server/pkg/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.User{}, &model.Message{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建默认管理员账号（如果不存在）
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		admin := model.User{
			Username: "admin",
			IsAdmin:  true,
		}
		admin.SetPassword("admin123")
		if err := db.Create(&admin).Error; err != nil {
			log.Fatal("创建默认管理员失败:", err)
		}
		log.Println("✓ 已创建默认管理员: admin/admin123")
	}

	// 初始化JWT
	middleware.InitJWT(cfg.JWTSecret)

	// 初始化 WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 创建路由
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"https://100.118.236.127",
			"https://talk.home.wbsays.com",
			"http://talk.home.wbsays.com",
			"https://home.tail96df5.ts.net",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 初始化handlers
	authHandler := handler.NewAuthHandler(db)
	adminHandler := handler.NewAdminHandler(db)
	uploadHandler := handler.NewUploadHandler(cfg.TalkServerURL, db, hub)
	wsHandler := handler.NewWebSocketHandler(hub)

	// 路由
	api := r.Group("/api")
	{
		// 认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(), authHandler.Me)
		}

		// 上传音频
		api.POST("/upload", middleware.AuthRequired(), uploadHandler.Upload)

		// WebSocket 连接（实时推送，认证在 handler 内部处理）
		api.GET("/ws", wsHandler.ServeWS)

		// 轮询获取回复（兼容旧版本）
		api.GET("/reply", middleware.AuthRequired(), uploadHandler.GetReply)

		// 获取历史记录
		api.GET("/history", middleware.AuthRequired(), uploadHandler.GetHistory)

		// 下载音频文件
		api.GET("/audio/:filename", middleware.AuthRequired(), func(c *gin.Context) {
			filename := c.Param("filename")
			filePath := fmt.Sprintf("/tmp/%s", filename)
			c.File(filePath)
		})

		// 管理后台
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.POST("/users", adminHandler.CreateUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
