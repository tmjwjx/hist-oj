package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/service"
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
			rating.GET("/contest/info/:contestId", handler.GetContestInfo)
			rating.POST("/batch", handler.GetBatchUserRating)
			rating.POST("/contest/batch", handler.GetBatchContestInfo)
			rating.GET("/rank", handler.GetRatingRank)

			// 管理接口：临时取消验证
			rating.POST("/calculate/:contestId", handler.CalculateRating)
			rating.POST("/initialize", handler.InitializeUserRating)
			rating.POST("/contest/set-rating-type", handler.SetContestRatingType)
			rating.POST("/trigger-scheduler", handler.TriggerScheduler)
		}

		// 判题终端相关接口
		judgeService := service.NewJudgeService(db)
		judgeHandler := NewJudgeHandler(judgeService)

		judge := api.Group("/judge")
		{
			judge.POST("/get-info", judgeHandler.GetInfo)
			judge.POST("/run-combined", judgeHandler.RunCombined)
		}
	}

	router.GET("/health", handler.HealthCheck)
}


