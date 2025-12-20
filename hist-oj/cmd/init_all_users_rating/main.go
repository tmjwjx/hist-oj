package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserInfo struct {
	UUID     string `gorm:"column:uuid;primaryKey"`
	Username string `gorm:"column:username"`
}

func (UserInfo) TableName() string {
	return "user_info"
}

type UserRecord struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UID        string    `gorm:"column:uid;not null;uniqueIndex"`
	HistRating *int      `gorm:"column:hist_rating"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserRecord) TableName() string {
	return "user_record"
}

func main() {
	// 连接数据库
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 查询所有用户
	var users []UserInfo
	if err := db.Find(&users).Error; err != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 个用户\n", len(users))

	// 查询已有 rating 记录的用户
	var existingRecords []UserRecord
	if err := db.Find(&existingRecords).Error; err != nil {
		fmt.Printf("❌ 查询已有 rating 记录失败: %v\n", err)
		return
	}

	// 构建已有记录的 UID 集合
	existingUIDs := make(map[string]bool)
	for _, record := range existingRecords {
		existingUIDs[record.UID] = true
	}

	fmt.Printf("📊 已有 %d 个用户有 rating 记录\n", len(existingRecords))

	// 初始化没有 rating 记录的用户
	defaultRating := 1200
	successCount := 0
	skipCount := 0

	fmt.Println("\n开始初始化用户 Rating...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, user := range users {
		// 跳过已有记录的用户
		if existingUIDs[user.UUID] {
			skipCount++
			continue
		}

		// 创建新的 rating 记录
		record := UserRecord{
			UID:        user.UUID,
			HistRating: &defaultRating,
		}

		if err := db.Create(&record).Error; err != nil {
			fmt.Printf("❌ 用户 %s (%s): 初始化失败 - %v\n", user.Username, user.UUID, err)
			continue
		}

		fmt.Printf("✅ 用户 %s (%s): 初始化 Rating = %d\n", user.Username, user.UUID, defaultRating)
		successCount++
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n✅ 初始化完成！\n")
	fmt.Printf("   - 总用户数: %d\n", len(users))
	fmt.Printf("   - 已有记录: %d\n", skipCount)
	fmt.Printf("   - 新增记录: %d\n", successCount)
	fmt.Printf("   - 失败: %d\n", len(users)-skipCount-successCount)
}
