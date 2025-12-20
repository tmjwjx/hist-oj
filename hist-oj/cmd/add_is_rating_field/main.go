package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 连接数据库
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 检查 contest 表是否已有 is_rating 字段
	var count int64
	err = db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = 'hoj'
		AND TABLE_NAME = 'contest'
		AND COLUMN_NAME = 'is_rating'
	`).Scan(&count).Error

	if err != nil {
		fmt.Printf("❌ 查询字段失败: %v\n", err)
		return
	}

	if count > 0 {
		fmt.Println("✅ is_rating 字段已存在，无需添加")
		return
	}

	fmt.Println("📝 开始添加 is_rating 字段...")

	// 添加 is_rating 字段，默认值为 0 (unrated)
	err = db.Exec(`
		ALTER TABLE contest
		ADD COLUMN is_rating TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为 Rating 比赛：0-否，1-是'
		AFTER type
	`).Error

	if err != nil {
		fmt.Printf("❌ 添加字段失败: %v\n", err)
		return
	}

	fmt.Println("✅ is_rating 字段添加成功")

	// 查询当前所有比赛
	var contests []struct {
		ID    uint64 `gorm:"column:id"`
		Title string `gorm:"column:title"`
	}

	err = db.Table("contest").Select("id, title").Find(&contests).Error
	if err != nil {
		fmt.Printf("❌ 查询比赛失败: %v\n", err)
		return
	}

	fmt.Printf("\n📊 当前有 %d 场比赛，默认都设置为 unrated\n", len(contests))
	fmt.Println("   如需设置为 rated，请使用管理员接口或手动修改数据库")
}
