package main

import (
	"fmt"
	"log"
	"time"

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

	// 查询所有 rating_history 记录，按创建时间排序
	var records []model.RatingHistory
	if err := db.Where("contest_id = ?", contest.ID).Order("uid, created_at").Find(&records).Error; err != nil {
		log.Fatalf("查询 rating_history 失败: %v", err)
	}

	fmt.Printf("\n共找到 %d 条 rating_history 记录\n", len(records))

	// 找出重复的记录（每个用户保留最新的）
	uidMap := make(map[string][]model.RatingHistory)
	for _, record := range records {
		uidMap[record.UID] = append(uidMap[record.UID], record)
	}

	duplicateCount := 0
	var idsToDelete []uint64

	for uid, userRecords := range uidMap {
		if len(userRecords) > 1 {
			duplicateCount++
			fmt.Printf("\n用户 %s 有 %d 条记录，将删除前 %d 条旧记录:\n",
				uid, len(userRecords), len(userRecords)-1)

			// 保留最后一条（最新的），删除其余的
			for i := 0; i < len(userRecords)-1; i++ {
				record := userRecords[i]
				idsToDelete = append(idsToDelete, record.ID)
				oldRating := 0
				if record.OldRating != nil {
					oldRating = *record.OldRating
				}
				fmt.Printf("  删除: ID=%d, OldRating=%d, NewRating=%d, Change=%d, CreatedAt=%v\n",
					record.ID, oldRating, record.NewRating, record.RatingChange, record.CreatedAt)
			}

			// 显示保留的记录
			lastRecord := userRecords[len(userRecords)-1]
			oldRating := 0
			if lastRecord.OldRating != nil {
				oldRating = *lastRecord.OldRating
			}
			fmt.Printf("  保留: ID=%d, OldRating=%d, NewRating=%d, Change=%d, CreatedAt=%v\n",
				lastRecord.ID, oldRating, lastRecord.NewRating, lastRecord.RatingChange, lastRecord.CreatedAt)
		}
	}

	if duplicateCount == 0 {
		fmt.Println("\n✅ 没有发现重复记录，无需清理")
		return
	}

	fmt.Printf("\n共需要删除 %d 条重复记录\n", len(idsToDelete))

	// 确认删除
	fmt.Print("\n是否确认删除？(yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" && confirm != "y" {
		fmt.Println("已取消删除")
		return
	}

	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("发生panic: %v", r)
		}
	}()

	// 删除重复记录
	if len(idsToDelete) > 0 {
		batchSize := 100
		for i := 0; i < len(idsToDelete); i += batchSize {
			end := i + batchSize
			if end > len(idsToDelete) {
				end = len(idsToDelete)
			}
			batch := idsToDelete[i:end]

			if err := tx.Delete(&model.RatingHistory{}, batch).Error; err != nil {
				tx.Rollback()
				log.Fatalf("删除失败: %v", err)
			}
			fmt.Printf("已删除 %d/%d 条记录\n", end, len(idsToDelete))
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}

	fmt.Println("\n✅ 清理完成！")

	// 重置比赛的 RatingCalculated 状态，允许重新计算
	fmt.Println("\n是否要重置比赛状态以允许重新计算？(yes/no): ")
	confirm = ""
	fmt.Scanln(&confirm)

	if confirm == "yes" || confirm == "y" {
		now := time.Now()
		if err := db.Model(&model.ContestRatingStatus{}).
			Where("contest_id = ?", contest.ID).
			Updates(map[string]interface{}{
				"rating_calculated": false,
				"calculated_at":     nil,
				"updated_at":        now,
			}).Error; err != nil {
			log.Fatalf("重置比赛状态失败: %v", err)
		}
		fmt.Println("✅ 比赛状态已重置，可以重新计算 Rating")
	}
}
