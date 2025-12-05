package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/middleware"
)

func SetupRoutes(router *gin.Engine, handler *Handler, cfg *config.Config, db *gorm.DB) {
	api := router.Group("/api")
	{
		// Rating 相关接口
		rating := api.Group("/rating")
		{
			// 查询接口：公开访问
			rating.GET("/user/:uid", handler.GetUserRating)
			rating.GET("/history/:uid", handler.GetRatingHistory)
			rating.GET("/color/:rating", handler.GetRatingColor)
			rating.GET("/contest/:contestId", handler.GetContestParticipantsRating)

			// 计算接口：仅管理员可触发
			rating.POST("/calculate/:contestId",
				middleware.AdminAuthMiddleware(&cfg.JWT, db),
				handler.CalculateRating)
		}
	}

	router.GET("/health", handler.HealthCheck)
}

