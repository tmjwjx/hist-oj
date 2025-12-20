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

	// 查询 user_record 表结构
	var columns []struct {
		Field   string
		Type    string
		Null    string
		Key     string
		Default *string
		Extra   string
	}

	if err := db.Raw("DESCRIBE user_record").Scan(&columns).Error; err != nil {
		log.Fatal("查询表结构失败:", err)
	}

	fmt.Println("user_record 表结构:")
	fmt.Println("----------------------------------------")
	for _, col := range columns {
		fmt.Printf("%-20s %-20s\n", col.Field, col.Type)
	}

	// 添加 hist_rating 列
	fmt.Println("\n添加 hist_rating 列...")
	if err := db.Exec("ALTER TABLE user_record ADD COLUMN hist_rating INT NULL COMMENT 'Hist Rating 分数'").Error; err != nil {
		fmt.Printf("添加列失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("✅ 成功添加 hist_rating 列")
	}

	// 再次查询表结构确认
	columns = nil
	if err := db.Raw("DESCRIBE user_record").Scan(&columns).Error; err != nil {
		log.Fatal("查询表结构失败:", err)
	}

	fmt.Println("\n更新后的 user_record 表结构:")
	fmt.Println("----------------------------------------")
	for _, col := range columns {
		fmt.Printf("%-20s %-20s\n", col.Field, col.Type)
	}
}
