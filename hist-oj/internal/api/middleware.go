package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/utils"
)

// AdminAuthMiddleware 管理员权限验证中间件
// TODO: 当前版本简化实现，不进行实际的权限验证
// 后续需要接入 HOJ 后端的认证系统，从 JWT Token 中解析用户信息
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// TODO: 实际实现应该：
		// 1. 从 Header 或 Query 中获取 token
		// 2. 调用 HOJ API 验证 token
		// 3. 检查用户是否为管理员
		// 4. 将操作人信息存入 context

		// 临时方案：从请求头或参数中获取操作人UID
		operatorUID := c.GetHeader("X-Operator-UID")
		if operatorUID == "" {
			operatorUID = c.Query("operatorUID")
		}

		// 如果没有提供操作人UID，使用默认值
		if operatorUID == "" {
			operatorUID = "admin"
		}

		logger.Info("管理员权限验证（临时简化版）",
			zap.String("operator_uid", operatorUID),
			zap.String("path", c.Request.URL.Path))

		// 将操作人UID存入context，供后续handler使用
		c.Set("operatorUID", operatorUID)

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
