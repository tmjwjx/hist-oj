package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 更新比赛 1002 的 is_rating 字段为 1
	contestID := 1002
	result := db.Exec("UPDATE contest SET is_rating = 1 WHERE id = ?", contestID)
	if result.Error != nil {
		log.Fatal("更新失败:", result.Error)
	}

	fmt.Printf("✅ 成功将比赛 %d 设置为 Rating 比赛（影响行数: %d）\n", contestID, result.RowsAffected)

	// 验证更新结果
	var contest struct {
		ID       int64 `gorm:"column:id"`
		IsRating bool  `gorm:"column:is_rating"`
		Title    string `gorm:"column:title"`
	}

	if err := db.Table("contest").Where("id = ?", contestID).First(&contest).Error; err != nil {
		log.Fatal("查询失败:", err)
	}

	fmt.Printf("\n比赛信息:\n")
	fmt.Printf("  ID: %d\n", contest.ID)
	fmt.Printf("  标题: %s\n", contest.Title)
	fmt.Printf("  是否为 Rating 比赛: %v\n", contest.IsRating)
}
