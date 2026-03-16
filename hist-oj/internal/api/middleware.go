package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// AuthMiddleware 用户认证中间件
// 从请求头中获取 Authorization token，调用 HOJ API 验证并设置用户信息到 context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 从 Header 中获取 token
		token := c.GetHeader("Authorization")
		urlType := c.GetHeader("Url-Type")

		logger.Info("AuthMiddleware收到请求",
			zap.String("path", c.Request.URL.Path),
			zap.String("urlType", urlType),
			zap.Bool("token_exists", token != ""),
			zap.Int("token_length", len(token)))

		if token == "" {
			logger.Warn("未提供认证token", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
			c.Abort()
			return
		}

		// 调用 HOJ API 验证 token
		userAuth, err := client.ValidateToken(token)
		if err != nil {
			logger.Warn("token验证失败",
				zap.Error(err),
				zap.String("token", token[:min(50, len(token))]),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusOK, errorResponse(401, "用户认证失败"))
			c.Abort()
			return
		}

		// 创建 UserInfo 对象供需要完整用户信息的 handler 使用
		userInfo := &model.UserInfo{
			UUID:     userAuth.UID,
			Username: userAuth.Username,
		}

		// 将用户信息存入context，供后续handler使用
		c.Set("user", userInfo)        // 完整用户对象（GetMyRegistration 等需要）
		c.Set("userId", userAuth.UID)  // 兼容旧代码
		c.Set("uid", userAuth.UID)     // 新代码使用 uid
		c.Set("username", userAuth.Username)
		c.Set("roles", userAuth.Roles)

		logger.Info("用户认证成功",
			zap.String("uid", userAuth.UID),
			zap.String("username", userAuth.Username),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AdminAuthMiddleware 管理员权限验证中间件
// 从请求头中获取 Authorization token，验证并设置用户信息到 context
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 从 Header 中获取 token
		token := c.GetHeader("Authorization")

		if token == "" {
			logger.Warn("未提供认证token")
			c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
			c.Abort()
			return
		}

		// 调用 ValidateToken 验证 token 并获取用户信息
		userAuth, err := client.ValidateToken(token)
		if err != nil {
			logger.Warn("token验证失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(401, "用户认证失败"))
			c.Abort()
			return
		}

		// 将用户信息存入context，供后续handler使用
		c.Set("userId", userAuth.UID)    // 兼容旧代码
		c.Set("uid", userAuth.UID)       // 新代码使用 uid
		c.Set("username", userAuth.Username)
		c.Set("roles", userAuth.Roles)

		logger.Info("管理员认证成功",
			zap.String("uid", userAuth.UID),
			zap.String("username", userAuth.Username),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的路由组
func RequireAdmin(handler *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 调用管理员权限验证中间件
		AdminAuthMiddleware()(c)
		// 如果验证通过，继续执行
		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
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
