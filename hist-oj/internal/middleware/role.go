package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/utils"
)

// 角色常量定义（与HOJ保持一致）
const (
	RoleRoot         = "root"          // 超级管理员
	RoleAdmin        = "admin"         // 管理员
	RoleProblemAdmin = "problem_admin" // 题目管理员
	RoleDefaultUser  = "default_user"  // 默认用户
)

// UserRole 用户角色关系（对应HOJ的user_role表）
type UserRole struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UID    string `gorm:"type:varchar(32);not null" json:"uid"`
	RoleID uint64 `gorm:"type:bigint unsigned;not null" json:"roleId"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_role"
}

// Role 角色信息（对应HOJ的role表）
type Role struct {
	ID          uint64 `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Role        string `gorm:"type:varchar(50);not null" json:"role"`
	Description string `gorm:"type:varchar(100)" json:"description"`
	Status      int    `gorm:"type:int;default:0" json:"status"` // 0可用，1不可用
}

// TableName 指定表名
func (Role) TableName() string {
	return "role"
}

// GetUserRoles 获取用户的所有角色
func GetUserRoles(db *gorm.DB, uid string) ([]string, error) {
	var userRoles []UserRole
	if err := db.Where("uid = ?", uid).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []string{}, nil
	}

	// 获取角色ID列表
	roleIDs := make([]uint64, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 查询角色信息
	var roles []Role
	if err := db.Where("id IN ? AND status = 0", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}

	// 提取角色名称
	roleNames := make([]string, 0, len(roles))
	for _, r := range roles {
		roleNames = append(roleNames, r.Role)
	}

	return roleNames, nil
}

// HasRole 检查用户是否拥有指定角色
func HasRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

// HasAnyRole 检查用户是否拥有任意一个指定角色
func HasAnyRole(roles []string, targetRoles []string) bool {
	for _, role := range roles {
		for _, target := range targetRoles {
			if role == target {
				return true
			}
		}
	}
	return false
}

// AdminAuthMiddleware 管理员权限中间件（root 或 admin）
func AdminAuthMiddleware(cfg *config.JWTConfig, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 先进行JWT认证
		JWTAuthMiddleware(cfg)(c)
		if c.IsAborted() {
			return
		}

		// 获取用户ID
		userId, exists := c.Get("userId")
		if !exists {
			logger.Warn("管理员权限检查：用户ID不存在", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		uid := userId.(string)

		// 获取用户角色
		roles, err := GetUserRoles(db, uid)
		if err != nil {
			logger.Error("管理员权限检查：查询用户角色失败",
				zap.String("uid", uid),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员
		if !HasAnyRole(roles, []string{RoleRoot, RoleAdmin}) {
			logger.Warn("管理员权限检查：用户权限不足",
				zap.String("uid", uid),
				zap.Strings("roles", roles),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		// 将角色信息存储到上下文
		c.Set("roles", roles)
		logger.Debug("管理员权限验证通过",
			zap.String("uid", uid),
			zap.Strings("roles", roles))

		c.Next()
	}
}

// RootAuthMiddleware 超级管理员权限中间件（仅 root）
func RootAuthMiddleware(cfg *config.JWTConfig, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 先进行JWT认证
		JWTAuthMiddleware(cfg)(c)
		if c.IsAborted() {
			return
		}

		// 获取用户ID
		userId, exists := c.Get("userId")
		if !exists {
			logger.Warn("超级管理员权限检查：用户ID不存在", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		uid := userId.(string)

		// 获取用户角色
		roles, err := GetUserRoles(db, uid)
		if err != nil {
			logger.Error("超级管理员权限检查：查询用户角色失败",
				zap.String("uid", uid),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查是否为超级管理员
		if !HasRole(roles, RoleRoot) {
			logger.Warn("超级管理员权限检查：用户权限不足",
				zap.String("uid", uid),
				zap.Strings("roles", roles),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要超级管理员权限",
			})
			c.Abort()
			return
		}

		// 将角色信息存储到上下文
		c.Set("roles", roles)
		logger.Debug("超级管理员权限验证通过",
			zap.String("uid", uid),
			zap.Strings("roles", roles))

		c.Next()
	}
}

// RoleAuthMiddleware 自定义角色权限中间件
func RoleAuthMiddleware(cfg *config.JWTConfig, db *gorm.DB, allowedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// 先进行JWT认证
		JWTAuthMiddleware(cfg)(c)
		if c.IsAborted() {
			return
		}

		// 获取用户ID
		userId, exists := c.Get("userId")
		if !exists {
			logger.Warn("角色权限检查：用户ID不存在", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		uid := userId.(string)

		// 获取用户角色
		roles, err := GetUserRoles(db, uid)
		if err != nil {
			logger.Error("角色权限检查：查询用户角色失败",
				zap.String("uid", uid),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查是否拥有允许的角色
		if !HasAnyRole(roles, allowedRoles) {
			logger.Warn("角色权限检查：用户权限不足",
				zap.String("uid", uid),
				zap.Strings("user_roles", roles),
				zap.Strings("required_roles", allowedRoles),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 将角色信息存储到上下文
		c.Set("roles", roles)
		logger.Debug("角色权限验证通过",
			zap.String("uid", uid),
			zap.Strings("roles", roles))

		c.Next()
	}
}
