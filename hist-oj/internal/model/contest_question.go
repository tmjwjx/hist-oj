package model

import (
	"time"
)

// ContestQuestion 比赛问题表
type ContestQuestion struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestID    uint64    `gorm:"type:bigint unsigned;not null;index:idx_contest_id" json:"contestId"`
	QuestionerID string    `gorm:"type:varchar(32);not null;index:idx_questioner_id" json:"questionerId"`
	Title        string    `gorm:"type:varchar(200);not null" json:"title"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	Status       string    `gorm:"type:varchar(20);default:'pending';index:idx_status" json:"status"` // pending, answered, closed
	Priority     int       `gorm:"type:int;default:0" json:"priority"` // 0: 普通, 1: 高
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Questioner *UserInfo               `gorm:"foreignKey:QuestionerID;references:UUID" json:"questioner,omitempty"`
	Replies    []ContestQuestionReply  `gorm:"foreignKey:QuestionID" json:"replies"`
}

// TableName 指定表名
func (ContestQuestion) TableName() string {
	return "contest_question"
}

// ContestQuestionReply 比赛问题回复表
type ContestQuestionReply struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	QuestionID uint64   `gorm:"type:bigint unsigned;not null;index:idx_question_id" json:"questionId"`
	SenderID  string    `gorm:"type:varchar(32);not null" json:"senderId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_created_at" json:"createdAt"`

	// 关联字段
	Sender *UserInfo `gorm:"foreignKey:SenderID;references:UUID" json:"sender,omitempty"`
}

// TableName 指定表名
func (ContestQuestionReply) TableName() string {
	return "contest_question_reply"
}
