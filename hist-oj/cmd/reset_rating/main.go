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
	CreatedAt  time.Time `gorm:"column:gmt_create;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:gmt_modified;autoUpdateTime"`
}

func (UserRecord) TableName() string {
	return "user_record"
}

type RatingHistory struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UID          string    `gorm:"column:uid;not null"`
	ContestID    uint64    `gorm:"column:contest_id;not null"`
	OldRating    int64     `gorm:"column:old_rating"`
	NewRating    int64     `gorm:"column:new_rating;not null"`
	RatingChange int64     `gorm:"column:rating_change;not null"`
	Rank         int64     `gorm:"column:rank;not null"`
	Participants int64     `gorm:"column:participants;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (RatingHistory) TableName() string {
	return "rating_history"
}

type ContestRatingStatus struct {
	ContestID        uint64     `gorm:"column:contest_id;primaryKey"`
	IsRated          bool       `gorm:"column:is_rated;default:0"`
	RatingCalculated bool       `gorm:"column:rating_calculated;default:0"`
	CalculatedAt     *time.Time `gorm:"column:calculated_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (ContestRatingStatus) TableName() string {
	return "contest_rating_status"
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
	fmt.Println()
	fmt.Println("⚠️  警告：此操作将清理所有 Rating 数据并重置为 0")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// 1. 清理 rating_history 表
	fmt.Println("🗑️  步骤 1: 清理 rating_history 表...")
	var historyCount int64
	if err := db.Model(&RatingHistory{}).Count(&historyCount).Error; err != nil {
		fmt.Printf("❌ 查询 rating_history 记录数失败: %v\n", err)
		return
	}
	fmt.Printf("   找到 %d 条历史记录\n", historyCount)

	if historyCount > 0 {
		if err := db.Exec("TRUNCATE TABLE rating_history").Error; err != nil {
			fmt.Printf("❌ 清理 rating_history 表失败: %v\n", err)
			return
		}
		fmt.Printf("   ✅ 已清理 %d 条历史记录\n", historyCount)
	} else {
		fmt.Println("   ℹ️  表中无数据，跳过")
	}
	fmt.Println()

	// 2. 重置 contest_rating_status 表的计算状态
	fmt.Println("🔄 步骤 2: 重置 contest_rating_status 表的计算状态...")
	var statusCount int64
	if err := db.Model(&ContestRatingStatus{}).Where("rating_calculated = ?", true).Count(&statusCount).Error; err != nil {
		fmt.Printf("❌ 查询 contest_rating_status 记录数失败: %v\n", err)
		return
	}
	fmt.Printf("   找到 %d 条已计算的比赛\n", statusCount)

	if statusCount > 0 {
		// 只重置 rating_calculated 和 calculated_at 字段，保留 is_rated 字段
		if err := db.Model(&ContestRatingStatus{}).
			Where("rating_calculated = ?", true).
			Updates(map[string]interface{}{
				"rating_calculated": false,
				"calculated_at":     nil,
			}).Error; err != nil {
			fmt.Printf("❌ 重置 contest_rating_status 失败: %v\n", err)
			return
		}
		fmt.Printf("   ✅ 已重置 %d 条比赛的计算状态\n", statusCount)
	} else {
		fmt.Println("   ℹ️  无需重置")
	}
	fmt.Println()

	// 3. 清理 user_record 表
	fmt.Println("🗑️  步骤 3: 清理 user_record 表...")
	var recordCount int64
	if err := db.Model(&UserRecord{}).Count(&recordCount).Error; err != nil {
		fmt.Printf("❌ 查询 user_record 记录数失败: %v\n", err)
		return
	}
	fmt.Printf("   找到 %d 条用户记录\n", recordCount)

	if recordCount > 0 {
		if err := db.Exec("TRUNCATE TABLE user_record").Error; err != nil {
			fmt.Printf("❌ 清理 user_record 表失败: %v\n", err)
			return
		}
		fmt.Printf("   ✅ 已清理 %d 条用户记录\n", recordCount)
	} else {
		fmt.Println("   ℹ️  表中无数据，跳过")
	}
	fmt.Println()

	// 4. 初始化所有用户的 Rating 为 0
	fmt.Println("🔄 步骤 4: 初始化所有用户 Rating 为 0...")
	var users []UserInfo
	if err := db.Find(&users).Error; err != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", err)
		return
	}

	fmt.Printf("   找到 %d 个用户\n", len(users))

	defaultRating := 0
	successCount := 0
	failCount := 0

	for i, user := range users {
		record := UserRecord{
			UID:        user.UUID,
			HistRating: &defaultRating,
		}

		if err := db.Create(&record).Error; err != nil {
			fmt.Printf("   ❌ 用户 %s (%s): 初始化失败 - %v\n", user.Username, user.UUID, err)
			failCount++
			continue
		}

		successCount++
		if (i+1)%10 == 0 || i+1 == len(users) {
			fmt.Printf("   进度: %d/%d\n", i+1, len(users))
		}
	}

	fmt.Printf("   ✅ 成功初始化 %d 个用户\n", successCount)
	if failCount > 0 {
		fmt.Printf("   ⚠️  失败: %d 个用户\n", failCount)
	}
	fmt.Println()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Rating 数据重置完成！")
	fmt.Println()
	fmt.Printf("📊 统计信息:\n")
	fmt.Printf("   - 清理历史记录: %d 条\n", historyCount)
	fmt.Printf("   - 重置比赛状态: %d 条\n", statusCount)
	fmt.Printf("   - 清理用户记录: %d 条\n", recordCount)
	fmt.Printf("   - 初始化用户: %d 个 (Rating = 0)\n", successCount)
	fmt.Printf("   - 失败: %d 个\n", failCount)
	fmt.Println()
	fmt.Println("💡 提示: 所有用户的 Rating 已重置为 0，历史记录已清空，比赛状态已重置")
}
