package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type PlagiarismResult struct {
	CheckID         uint64 `gorm:"column:check_id"`
	IsOverThreshold bool   `gorm:"column:is_over_threshold"`
}

func (PlagiarismResult) TableName() string {
	return "plagiarism_result"
}

func main() {
	// Connect to database
	dsn := "root:hist2025@tcp(43.143.133.62:3306)/hoj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get all check IDs
	var checkIDs []uint64
	db.Model(&PlagiarismResult{}).Distinct("check_id").Pluck("check_id", &checkIDs)

	fmt.Printf("Found %d check IDs\n", len(checkIDs))
	for _, checkID := range checkIDs {
		var totalCount int64
		var overThresholdCount int64

		db.Model(&PlagiarismResult{}).Where("check_id = ?", checkID).Count(&totalCount)
		db.Model(&PlagiarismResult{}).Where("check_id = ? AND is_over_threshold = ?", checkID, true).Count(&overThresholdCount)

		fmt.Printf("CheckID %d: Total=%d, OverThreshold=%d\n", checkID, totalCount, overThresholdCount)
	}
}
