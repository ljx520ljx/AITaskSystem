package main

import (
	"AITaskSystem/handler"
	"AITaskSystem/model"
	"AITaskSystem/repository"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CORS 中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	// 1. 初始化数据库
	db, err := gorm.Open(sqlite.Open("task.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	// 2. 自动迁移
	err = db.AutoMigrate(&model.Task{})
	if err != nil {
		log.Fatal("failed to migrate database", err)
	}

	// 3. 依赖注入
	taskRepo := repository.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepo)

	// 4. 设置 Gin 路由
	r := gin.Default()
	r.Use(Cors()) // 开启跨域

	// 静态文件服务 (前端入口)
	r.GET("/", func(c *gin.Context) {
		c.File("./index.html")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")
	{

		// 查
		v1.GET("/tasks", taskHandler.GetAllTasks)
		v1.GET("/tasks/:id", taskHandler.GetTaskByID)

		// 增 (统一入口：自动接入 AI 分析优先级、工时、依赖)
		v1.POST("/tasks", taskHandler.CreateTask)

		// 改
		v1.PUT("/tasks/:id", taskHandler.UpdateTask)
		v1.PUT("/tasks/:id/complete", taskHandler.MarkComplete)

		// 删
		v1.DELETE("/tasks/:id", taskHandler.DeleteTask)

		// 报表 (AI 自动生成)
		v1.GET("/report", taskHandler.GetWeeklyReport)
	}

	log.Println("Server starting on http://localhost:8080")
	r.Run(":8080")
}
