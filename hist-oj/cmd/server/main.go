package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/api"
	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/schedule"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := utils.InitLogger(nil); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	logger := utils.GetLogger()

	// 初始化数据库
	if err := client.InitDatabase(&cfg.Database); err != nil {
		logger.Fatal("Failed to init database", zap.Error(err))
	}

	// 初始化HOJ API客户端
	client.InitHojAPIClient(&cfg.HojAPI)

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建服务
	db := client.GetDB()
	ratingService := service.NewRatingService(db, &cfg.Rating)
	queryService := service.NewQueryService(db)

	// 创建处理器
	handler := api.NewHandler(ratingService, queryService)

	// 启动定时任务
	scheduler := schedule.NewScheduler(ratingService, &cfg.Rating)
	if err := scheduler.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}
	defer scheduler.Stop()

	// 将 scheduler 设置到 handler 中
	handler.SetScheduler(scheduler)

	// 设置路由
	router := gin.Default()

	// 添加 CORS 中间件（允许前端跨域访问）
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api.SetupRoutes(router, handler, cfg, db)

	// 静态文件服务（用于资料下载）
	// 创建上传目录
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir+"/classroom", 0755); err != nil {
		logger.Warn("Failed to create upload directory", zap.Error(err))
	}
	router.Static("/uploads", uploadDir)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器（在goroutine中）
	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

