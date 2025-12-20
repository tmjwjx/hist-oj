package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type RatingHistory struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UID          string    `gorm:"column:uid;not null"`
	ContestID    uint64    `gorm:"column:contest_id;not null"`
	Rank         int       `gorm:"column:rank"`
	OldRating    *int      `gorm:"column:old_rating"`
	NewRating    int       `gorm:"column:new_rating"`
	RatingChange int       `gorm:"column:rating_change"`
	Participants int       `gorm:"column:participants"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (RatingHistory) TableName() string {
	return "rating_history"
}

type ContestRatingStatus struct {
	ContestID        uint64     `gorm:"column:contest_id;primaryKey"`
	IsRated          bool       `gorm:"column:is_rated;default:false"`
	RatingCalculated bool       `gorm:"column:rating_calculated;default:false"`
	CalculatedAt     *time.Time `gorm:"column:calculated_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (ContestRatingStatus) TableName() string {
	return "contest_rating_status"
}

func main() {
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 从数据库查询比赛 1002 的实际参赛者
	type ContestRecord struct {
		UID      string `gorm:"column:uid"`
		Username string `gorm:"column:username"`
	}

	var records []ContestRecord
	if err := db.Table("contest_record").
		Select("DISTINCT contest_record.uid, user_info.username").
		Joins("LEFT JOIN user_info ON contest_record.uid = user_info.uuid").
		Where("contest_record.cid = ?", 1002).
		Limit(10).
		Scan(&records).Error; err != nil {
		fmt.Printf("❌ 查询参赛者失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("❌ 没有找到比赛 1002 的参赛者数据")
		return
	}

	fmt.Printf("✅ 找到 %d 名参赛者\n", len(records))

	// 更新比赛状态
	now := time.Now()
	var existingStatus ContestRatingStatus
	result := db.Where("contest_id = ?", 1002).First(&existingStatus)

	if result.Error == nil {
		db.Model(&ContestRatingStatus{}).
			Where("contest_id = ?", 1002).
			Updates(map[string]interface{}{
				"is_rated":          true,
				"rating_calculated": true,
				"calculated_at":     now,
			})
	} else {
		status := ContestRatingStatus{
			ContestID:        1002,
			IsRated:          true,
			RatingCalculated: true,
			CalculatedAt:     &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		db.Create(&status)
	}
	fmt.Println("✅ 比赛状态已更新")

	// 插入 Rating 数据
	successCount := 0
	participants := len(records)

	fmt.Println("\n开始插入 Rating 数据...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, record := range records {
		rank := i + 1

		// 模拟 Rating 变化：前 5 名上升，后 5 名下降
		var oldRating, newRating, ratingChange int
		if rank <= 5 {
			oldRating = 1500 - (rank-1)*10
			ratingChange = 50 - (rank-1)*5
		} else {
			oldRating = 1500 + (rank-6)*10
			ratingChange = -10 - (rank-6)*5
		}
		newRating = oldRating + ratingChange

		// 删除旧记录
		db.Where("uid = ? AND contest_id = ?", record.UID, 1002).Delete(&RatingHistory{})

		// 插入新记录
		history := RatingHistory{
			UID:          record.UID,
			ContestID:    1002,
			Rank:         rank,
			OldRating:    &oldRating,
			NewRating:    newRating,
			RatingChange: ratingChange,
			Participants: participants,
		}

		if err := db.Create(&history).Error; err != nil {
			fmt.Printf("❌ 第 %d 名 %s: 插入失败 - %v\n", rank, record.Username, err)
			continue
		}

		changeSymbol := "+"
		if ratingChange < 0 {
			changeSymbol = ""
		}
		fmt.Printf("✅ 第 %d 名 %s: %d → %d (%s%d)\n",
			rank, record.Username, oldRating, newRating, changeSymbol, ratingChange)
		successCount++
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n✅ 插入完成！成功: %d 条\n", successCount)
	fmt.Println("\n🎉 现在可以访问 http://localhost:8066/contest/1002/rank 查看效果！")
	fmt.Println("   刷新页面后，你应该能看到用户名旁边显示 Rating 变化（绿色 +X 或红色 -X）")
}
