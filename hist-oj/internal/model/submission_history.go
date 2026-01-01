package model

import (
	"time"
)

// SubmissionHistory 提交历史记录
type SubmissionHistory struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SubmitID    string    `gorm:"type:varchar(50);index" json:"submit_id"`
	PID         string    `gorm:"type:varchar(50);index:idx_pid_cid" json:"pid"`
	CID         string    `gorm:"type:varchar(50);default:'0';index:idx_pid_cid" json:"cid"`
	Username    string    `gorm:"type:varchar(100);index" json:"username"`
	Result      string    `gorm:"type:varchar(50)" json:"result"`
	TimeUsed    string    `gorm:"type:varchar(20)" json:"time_used"`
	MemoryUsed  string    `gorm:"type:varchar(20)" json:"memory_used"`
	Language    string    `gorm:"type:varchar(50)" json:"language"`
	Code        string    `gorm:"type:mediumtext" json:"code"`
	LocalInfo   string    `gorm:"type:varchar(100);default:''" json:"local_info"`
	SubmitTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index" json:"submit_time"`
}

// TableName 指定表名
func (SubmissionHistory) TableName() string {
	return "submission_history"
}
