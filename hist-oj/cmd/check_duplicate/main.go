package main

import (
	"fmt"
	"log"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/model"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := client.InitDatabase(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	db := client.GetDB()

	// 查找热身赛
	var contest model.Contest
	if err := db.Where("title LIKE ?", "%热身赛%").First(&contest).Error; err != nil {
		log.Fatalf("查询热身赛失败: %v", err)
	}

	fmt.Printf("找到比赛: ID=%d, Title=%s\n", contest.ID, contest.Title)

	// 检查 rating_history 表中是否有重复记录
	var records []model.RatingHistory
	if err := db.Where("contest_id = ?", contest.ID).Order("uid, created_at").Find(&records).Error; err != nil {
		log.Fatalf("查询 rating_history 失败: %v", err)
	}

	fmt.Printf("\n共找到 %d 条 rating_history 记录\n", len(records))

	// 检查每个用户是否有重复记录
	uidMap := make(map[string][]model.RatingHistory)
	for _, record := range records {
		uidMap[record.UID] = append(uidMap[record.UID], record)
	}

	duplicateCount := 0
	for uid, userRecords := range uidMap {
		if len(userRecords) > 1 {
			duplicateCount++
			fmt.Printf("\n⚠️  用户 %s 有 %d 条记录:\n", uid, len(userRecords))
			for i, record := range userRecords {
				oldRating := 0
				if record.OldRating != nil {
					oldRating = *record.OldRating
				}
				fmt.Printf("  [%d] OldRating=%d, NewRating=%d, Change=%d, Rank=%d, CreatedAt=%v\n",
					i+1, oldRating, record.NewRating, record.RatingChange, record.Rank, record.CreatedAt)
			}
		}
	}

	if duplicateCount == 0 {
		fmt.Println("\n✅ 没有发现重复记录")
	} else {
		fmt.Printf("\n❌ 共有 %d 个用户存在重复记录\n", duplicateCount)
	}

	// 检查比赛状态
	var status model.ContestRatingStatus
	if err := db.Where("contest_id = ?", contest.ID).First(&status).Error; err != nil {
		log.Fatalf("查询比赛状态失败: %v", err)
	}

	fmt.Printf("\n比赛状态:\n")
	fmt.Printf("  IsRated: %v\n", status.IsRated)
	fmt.Printf("  RatingCalculated: %v\n", status.RatingCalculated)
	if status.CalculatedAt != nil {
		fmt.Printf("  CalculatedAt: %v\n", status.CalculatedAt)
	}
	fmt.Printf("  CreatedAt: %v\n", status.CreatedAt)
	fmt.Printf("  UpdatedAt: %v\n", status.UpdatedAt)

	// 检查是否有用户被更新了两次
	fmt.Printf("\n检查用户 rating 是否被重复更新...\n")
	updatedTwice := 0
	for uid, userRecords := range uidMap {
		if len(userRecords) > 1 {
			// 检查用户的最终 rating
			var user model.UserRecord
			if err := db.Where("uid = ?", uid).First(&user).Error; err == nil {
				finalRating := 1200 // 默认值
				if user.HistRating != nil {
					finalRating = *user.HistRating
				}
				// 最后一条记录的 NewRating 应该等于用户的当前 rating
				lastRecord := userRecords[len(userRecords)-1]
				if lastRecord.NewRating != finalRating {
					fmt.Printf("⚠️  用户 %s: 当前 rating=%d, 但历史记录最后一条=%d\n",
						uid, finalRating, lastRecord.NewRating)
					updatedTwice++
				}
			}
		}
	}
	if updatedTwice == 0 {
		fmt.Println("✅ 所有用户的 rating 都是正确的")
	} else {
		fmt.Printf("❌ 共有 %d 个用户的 rating 不匹配\n", updatedTwice)
	}
}
