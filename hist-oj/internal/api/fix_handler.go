package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FixOperationLogsUsername 修复操作日志中缺失的操作人用户名
func (h *Handler) FixOperationLogsUsername(c *gin.Context) {
	db := h.ratingService.GetDB()

	// 1. 修复 rating_operation_logs 表
	sql1 := `
		UPDATE rating_operation_logs rol
		LEFT JOIN user u ON rol.operator_uid = u.uuid
		SET rol.operator_username = u.username
		WHERE rol.operator_username IS NULL OR rol.operator_username = ''
	`
	if err := db.Exec(sql1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "修复操作日志失败",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	// 2. 修复 contest_skip_users 表
	sql2 := `
		UPDATE contest_skip_users csu
		LEFT JOIN user u ON csu.operator_uid = u.uuid
		SET csu.operator_username = u.username
		WHERE csu.operator_username IS NULL OR csu.operator_username = ''
	`
	if err := db.Exec(sql2).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "修复skip用户记录失败",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	// 3. 查询修复结果
	var result []struct {
		TableName       string `json:"tableName" gorm:"column:table_name"`
		TotalRecords    int64  `json:"totalRecords" gorm:"column:total_records"`
		WithUsername    int64  `json:"withUsername" gorm:"column:with_username"`
		WithoutUsername int64  `json:"withoutUsername" gorm:"column:without_username"`
	}

	// 查询 rating_operation_logs 统计
	db.Raw(`
		SELECT
			'rating_operation_logs' as table_name,
			COUNT(*) as total_records,
			SUM(CASE WHEN operator_username IS NOT NULL AND operator_username != '' THEN 1 ELSE 0 END) as with_username,
			SUM(CASE WHEN operator_username IS NULL OR operator_username = '' THEN 1 ELSE 0 END) as without_username
		FROM rating_operation_logs
	`).Scan(&result)

	// 查询 contest_skip_users 统计
	db.Raw(`
		SELECT
			'contest_skip_users' as table_name,
			COUNT(*) as total_records,
			SUM(CASE WHEN operator_username IS NOT NULL AND operator_username != '' THEN 1 ELSE 0 END) as with_username,
			SUM(CASE WHEN operator_username IS NULL OR operator_username = '' THEN 1 ELSE 0 END) as without_username
		FROM contest_skip_users
	`).Scan(&result)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "操作日志用户名修复完成",
		"data": gin.H{
			"statistics": result,
		},
	})
}
