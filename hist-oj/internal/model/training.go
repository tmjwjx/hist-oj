package model

import "time"

// Training 训练实体（对应HOJ的training表）
type Training struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title      string    `gorm:"type:varchar(100);not null" json:"title"`
	Author     string    `gorm:"type:varchar(50);not null" json:"author"`
	Auth       string    `gorm:"type:varchar(20);not null;default:'Public'" json:"auth"` // Public, Private
	PrivatePwd string    `gorm:"type:varchar(100)" json:"privatePwd,omitempty"`           // 私有训练的密码
	Status     bool      `gorm:"type:tinyint(1);not null;default:1" json:"status"`        // 是否可用
	Rank       int       `gorm:"column:rank;not null" json:"rank"`                         // 编号，排序用
	IsGroup    bool      `gorm:"type:tinyint(1);not null;default:0" json:"isGroup"`       // 是否为团队训练
	GID        *uint64   `gorm:"type:bigint unsigned" json:"gid,omitempty"`               // 团队ID
	GmtCreate  time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
}

// TableName 指定表名
func (Training) TableName() string {
	return "training"
}

// TrainingRegister 私有训练注册记录（对应HOJ的training_register表）
type TrainingRegister struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TID        uint64    `gorm:"type:bigint unsigned;not null;index:idx_tid" json:"tid"`
	UID        string    `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	Status     bool      `gorm:"type:tinyint(1);not null;default:1" json:"status"`
	GmtCreate  time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
}

// TableName 指定表名
func (TrainingRegister) TableName() string {
	return "training_register"
}

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
