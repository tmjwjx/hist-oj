package model

import (
	"time"
)

// SubmissionHistory 提交历史记录
type SubmissionHistory struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SubmitID    string    `gorm:"type:varchar(50);index;column:submit_id" json:"submit_id"`
	PID         string    `gorm:"type:varchar(50);index:idx_pid_cid;column:pid" json:"pid"`
	CID         string    `gorm:"type:varchar(50);default:'0';index:idx_pid_cid;column:cid" json:"cid"`
	Username    string    `gorm:"type:varchar(100);index;column:username" json:"username"`
	Result      string    `gorm:"type:varchar(50);column:result" json:"result"`
	TimeUsed    string    `gorm:"type:varchar(20);column:time_used" json:"time_used"`
	MemoryUsed  string    `gorm:"type:varchar(20);column:memory_used" json:"memory_used"`
	Language    string    `gorm:"type:varchar(50);column:language" json:"language"`
	Code        string    `gorm:"type:mediumtext;column:code" json:"code"`
	LocalInfo   string    `gorm:"type:varchar(100);default:'';column:local_info" json:"local_info"`
	SubmitTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;column:submit_time" json:"submit_time"`
}

// TableName 指定表名
func (SubmissionHistory) TableName() string {
	return "submission_history"
}
