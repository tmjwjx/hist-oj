package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/utils"
)

// JWTClaims JWT Claims结构（兼容HOJ的JWT格式）
type JWTClaims struct {
	jwt.RegisteredClaims
}

// JWTAuthMiddleware JWT认证中间件（必须认证）
func JWTAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()
		
		// 获取Authorization头（HOJ的token直接是JWT字符串，没有Bearer前缀）
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("请求缺少Authorization头", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权，请先登录",
			})
			c.Abort()
			return
		}

		// HOJ的token可能没有Bearer前缀，需要兼容处理
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		// 解析JWT Token（使用HS512算法，与HOJ一致）
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法（HOJ使用HS512）
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			logger.Warn("JWT Token解析失败", 
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 验证Claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// 检查是否过期
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				logger.Warn("JWT Token已过期", zap.String("userId", claims.Subject))
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Token已过期，请重新登录",
				})
				c.Abort()
				return
			}

			// HOJ的JWT token的subject是userId
			userId := claims.Subject
			if userId == "" {
				logger.Warn("JWT Token中缺少userId", zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Token格式错误",
				})
				c.Abort()
				return
			}

			// 将用户ID存储到上下文，供后续使用
			c.Set("userId", userId)
			c.Set("claims", claims)
			
			logger.Debug("JWT认证成功", 
				zap.String("userId", userId),
				zap.String("path", c.Request.URL.Path))
			
			c.Next()
		} else {
			logger.Warn("JWT Token验证失败", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token验证失败",
			})
			c.Abort()
			return
		}
	}
}

// OptionalAuthMiddleware 可选认证中间件（有token就验证，没有就跳过）
func OptionalAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			// 没有token，继续执行（允许未登录访问）
			c.Next()
			return
		}

		// 有token，进行验证
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				// 检查是否过期
				if claims.ExpiresAt == nil || claims.ExpiresAt.Time.After(time.Now()) {
					userId := claims.Subject
					if userId != "" {
						c.Set("userId", userId)
						c.Set("claims", claims)
						logger.Debug("可选认证：JWT验证成功", zap.String("userId", userId))
					}
				}
			}
		} else {
			logger.Debug("可选认证：JWT验证失败，但继续执行", zap.Error(err))
		}

		c.Next()
	}
}

