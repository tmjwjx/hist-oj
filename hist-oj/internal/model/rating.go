package model

import (
	"time"

	"gorm.io/gorm"
)

// RatingHistory Rating历史记录
type RatingHistory struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UID          string     `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	ContestID    uint64     `gorm:"type:bigint unsigned;not null;index:idx_contest_id" json:"contestId"`
	OldRating    *int       `gorm:"type:int" json:"oldRating"`
	NewRating    int        `gorm:"type:int;not null" json:"newRating"`
	RatingChange int        `gorm:"type:int;not null" json:"ratingChange"`
	Rank         int        `gorm:"type:int;not null" json:"rank"`
	Participants int       `gorm:"type:int;not null" json:"participants"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (RatingHistory) TableName() string {
	return "rating_history"
}

// ContestRatingStatus 比赛Rating状态
type ContestRatingStatus struct {
	ContestID       uint64     `gorm:"primaryKey;type:bigint unsigned" json:"contestId"`
	IsRated         bool       `gorm:"type:tinyint(1);default:0" json:"isRated"`
	RatingCalculated bool       `gorm:"type:tinyint(1);default:0;index:idx_rating_calculated" json:"ratingCalculated"`
	CalculatedAt    *time.Time `gorm:"type:datetime" json:"calculatedAt"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (ContestRatingStatus) TableName() string {
	return "contest_rating_status"
}

// UserRecord 用户记录（对应HOJ的user_record表）
type UserRecord struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UID        string    `gorm:"type:varchar(32);not null;uniqueIndex" json:"uid"`
	HistRating *int      `gorm:"type:int;column:hist_rating" json:"histRating"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (UserRecord) TableName() string {
	return "user_record"
}

// Contest 比赛信息（对应HOJ的contest表）
type Contest struct {
	ID        uint64    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Type      int       `gorm:"type:int;not null;default:0" json:"type"` // 0: ACM, 1: OI
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	StartTime time.Time `gorm:"type:datetime" json:"startTime"`
	EndTime   time.Time `gorm:"type:datetime" json:"endTime"`
	Duration  int64     `gorm:"type:bigint" json:"duration"`
	Status    int       `gorm:"type:int" json:"status"` // -1: 未开始, 0: 进行中, 1: 已结束
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Contest) TableName() string {
	return "contest"
}

// InitTables 初始化数据库表
func InitTables(db *gorm.DB) error {
	// 创建rating相关表
	if err := db.AutoMigrate(&RatingHistory{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&ContestRatingStatus{}); err != nil {
		return err
	}
	return nil
}

