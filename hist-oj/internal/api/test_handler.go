package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestInsertOperationLog 测试插入操作日志
func (h *Handler) TestInsertOperationLog(c *gin.Context) {
	// 直接调用 LogOperation 方法
	h.ratingService.LogOperation("admin", "admin", "personal_adjust", "user", "testuser", `{"username":"testuser","oldRating":0,"newRating":10,"ratingChange":10,"reason":"测试：手动调整Rating"}`, "127.0.0.1")
	h.ratingService.LogOperation("admin", "admin", "skip_user", "contest", "1000", `{"contestId":1000,"usernames":["user1","user2"],"reason":"代码抄袭","autoRecalc":true}`, "127.0.0.1")
	h.ratingService.LogOperation("admin", "admin", "cancel_skip", "contest", "1000", `{"contestId":1000,"uids":["uuid1","uuid2"],"count":2}`, "127.0.0.1")
	h.ratingService.LogOperation("admin", "admin", "recalculate", "contest", "1000", `{"taskId":1,"contestId":1000,"contestCount":5}`, "127.0.0.1")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "测试数据插入成功",
		"data": gin.H{
			"inserted": 4,
		},
	})
}
