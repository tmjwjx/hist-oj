package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MigrateOperationLogs 从历史数据迁移操作日志
func (h *Handler) MigrateOperationLogs(c *gin.Context) {
	migratedCount, errors, err := h.ratingService.MigrateOperationLogs()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "操作日志迁移失败",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "操作日志迁移完成",
		"data": gin.H{
			"migrated": migratedCount,
			"errors":   errors,
		},
	})
}
