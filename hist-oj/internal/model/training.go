package model

import "time"

// TrainingProgress 训练进度信息
type TrainingProgress struct {
	TrainingID  uint64 `json:"trainingId"`
	SolvedCount int64  `json:"solvedCount"`
	TotalCount  int64  `json:"totalCount"`
	Status      string `json:"status"`
}

// TrainingParticipant 训练参与记录表(避免与HOJ的training_record冲突)
type TrainingParticipant struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainingID uint64     `gorm:"type:bigint unsigned;not null;index:idx_training_id" json:"trainingId"`
	UID        string     `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	Status     string     `gorm:"type:varchar(20);default:not_started" json:"status"` // not_started, in_progress, completed
	JoinTime   *time.Time `gorm:"type:datetime" json:"joinTime"`
	GmtCreate  time.Time  `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`

	// 关联字段
	User    *UserInfo         `gorm:"foreignKey:UID;references:UUID" json:"user,omitempty"`
	Progress *TrainingProgress `gorm:"-" json:"progress,omitempty"` // 不存储在数据库，仅用于API返回
}

// TableName 指定表名
func (TrainingParticipant) TableName() string {
	return "training_participant"
}
