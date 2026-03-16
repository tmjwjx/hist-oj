package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	"github.com/hoj/hist-oj/internal/websocket"

	"github.com/gin-contrib/gzip"
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

	// 初始化数据库表（查重表需要使用 SQL 脚本手动创建，不使用 AutoMigrate）
	db := client.GetDB()
	// 查重表通过 SQL 脚本手动创建，不使用 AutoMigrate（避免字段名问题）
	// if err := model.InitPlagiarismTables(db); err != nil {
	// 	logger.Warn("Failed to init plagiarism tables (service will continue)", zap.Error(err))
	// }

	// 初始化HOJ API客户端
	client.InitHojAPIClient(&cfg.HojAPI)

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建服务
	ratingService := service.NewRatingService(db, &cfg.Rating)
	queryService := service.NewQueryService(db)

	// 创建WebSocket Hub
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	// 创建处理器
	handler := api.NewHandler(ratingService, queryService, wsHub)

	// 启动定时任务
	scheduler := schedule.NewScheduler(ratingService, &cfg.Rating)
	if err := scheduler.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}
	defer scheduler.Stop()

	// 将 scheduler 设置到 handler 中
	handler.SetScheduler(scheduler)

	// 创建路由并添加 gzip 中间件
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 添加 CORS 中间件（允许前端跨域访问）
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api.SetupRoutes(router, handler, cfg, db)

	// 静态文件服务（用于资料预览和下载）
	// 创建上传目录
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir+"/classroom", 0755); err != nil {
		logger.Warn("Failed to create upload directory", zap.Error(err))
	}

	// 自定义静态文件服务 - 带认证和权限检查
	router.GET("/uploads/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")

		// 安全检查：防止路径遍历攻击
		filepath = strings.TrimPrefix(filepath, "/")
		if strings.Contains(filepath, "..") {
			logger.Warn("检测到路径遍历攻击尝试", zap.String("path", filepath))
			c.JSON(http.StatusForbidden, gin.H{"error": "禁止访问"})
			return
		}

		// 检查Referer头和Origin头 - 防止直接URL访问和外部下载
		referer := c.GetHeader("Referer")
		originHeader := c.GetHeader("Origin") // 对于 CORS 请求
		host := c.Request.Host
		userAgent := c.GetHeader("User-Agent")

		// 允许的来源：必须来自本站
		// 检查 Referer 或 Origin 头
		allowedFromReferer := referer != "" && (strings.Contains(referer, "bingoj.cn") || strings.Contains(referer, host))
		allowedFromOrigin := originHeader != "" && (strings.Contains(originHeader, "bingoj.cn") || strings.Contains(originHeader, host))
		// 允许浏览器的 fetch 请求（User-Agent 包含 "Mozilla"）
		allowedFromBrowser := userAgent != "" && strings.Contains(userAgent, "Mozilla")

		allowed := allowedFromReferer || allowedFromOrigin || allowedFromBrowser

		// 如果不是从本站访问，拒绝请求
		if !allowed {
			logger.Warn("拒绝非本站请求",
				zap.String("referer", referer),
				zap.String("origin", originHeader),
				zap.String("host", host),
				zap.String("path", filepath))
			c.JSON(http.StatusForbidden, gin.H{"error": "禁止直接访问"})
			return
		}

		// 构建完整文件路径
		fullPath := uploadDir + "/" + filepath

		// 检查文件是否存在
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}

		// 设置响应头
		c.Header("Content-Disposition", "inline")
		// 禁用缓存，确保权限检查每次都执行
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("X-Content-Type-Options", "nosniff")
		// 确保允许跨域读取（用于 Canvas 渲染）
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET")
		c.Header("Access-Control-Allow-Headers", "*")

		c.File(fullPath)
	})

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

