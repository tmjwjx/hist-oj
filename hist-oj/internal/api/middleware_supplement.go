package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/utils"
)

// AdminRoleAuthMiddleware 管理员权限验证中间件（root/admin）
func AdminRoleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 先确保用户已登录
		AdminAuthMiddleware()(c)
		if c.IsAborted() {
			return
		}

		roles, exists := c.Get("roles")
		if !exists {
			logger.Warn("未找到用户角色信息")
			c.JSON(200, errorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			logger.Warn("角色格式错误")
			c.JSON(200, errorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		isAdmin := false
		for _, role := range roleList {
			if role == "root" || role == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			logger.Warn("用户不是管理员", zap.Any("roles", roleList))
			c.JSON(200, errorResponse(403, "需要管理员权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminAuthMiddleware 超级管理员权限验证中间件
func SuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 先调用普通管理员认证，确保用户已登录
		AdminAuthMiddleware()(c)
		if c.IsAborted() {
			return
		}

		// 获取用户角色
		roles, exists := c.Get("roles")
		if !exists {
			logger.Warn("未找到用户角色信息")
			c.JSON(200, errorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		// 检查是否是超级管理员
		roleList, ok := roles.([]string)
		if !ok {
			logger.Warn("角色格式错误")
			c.JSON(200, errorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		// 检查是否包含超级管理员角色 (root 为超级管理员, admin 为普通管理员)
		isSuperAdmin := false
		for _, role := range roleList {
			if role == "root" || role == "admin" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			logger.Warn("用户不是超级管理员",
				zap.Any("roles", roleList))
			c.JSON(200, errorResponse(403, "需要超级管理员权限"))
			c.Abort()
			return
		}

		logger.Info("超级管理员权限验证通过")
		c.Next()
	}
}
