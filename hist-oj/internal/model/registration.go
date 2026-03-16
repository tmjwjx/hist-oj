package model

import (
	"time"
)

// Competition 比赛
type Competition struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string          `gorm:"type:varchar(255);not null" json:"name"`
	StartTime   time.Time       `gorm:"type:datetime;not null;column:start_time" json:"startTime"`
	EndTime     time.Time       `gorm:"type:datetime;not null;column:end_time" json:"endTime"`
	Fields      string          `gorm:"type:text;not null" json:"fields"` // JSON字符串，存储FieldConfig
	LogoURL     string          `gorm:"type:text;column:logo_url" json:"logoUrl"`
	Description string          `gorm:"type:text" json:"description"` // HTML格式的比赛说明
	Visible     bool            `gorm:"type:tinyint(1);default:1" json:"visible"`
	CreatedAt   time.Time       `gorm:"autoCreateTime;column:created_at" json:"createdAt"`

	// 关联字段
	Registrations []Registration `gorm:"foreignKey:CompetitionID" json:"registrations,omitempty"`
}

// TableName 指定表名
func (Competition) TableName() string {
	return "histcontest_register_competitions"
}

// Registration 报名记录
type Registration struct {
	ID                uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	CompetitionID     uint64      `gorm:"type:int;not null;column:competition_id" json:"competitionId"`
	UserUUID          string      `gorm:"type:varchar(100);column:user_uuid" json:"userUuid"`
	Name              string      `gorm:"type:varchar(100);column:name" json:"name"`
	Class             string      `gorm:"type:varchar(100);column:class" json:"class"`
	College           string      `gorm:"type:varchar(100);column:college" json:"college"`
	StudentID         string      `gorm:"type:varchar(50);column:student_id" json:"studentId"`
	Gender            string      `gorm:"type:varchar(10);column:gender" json:"gender"`
	ShirtSize         string      `gorm:"type:varchar(10);column:shirt_size" json:"shirtSize"`
	TeamName          string      `gorm:"type:varchar(100);column:team_name" json:"teamName"`
	QQ                string      `gorm:"type:varchar(20);column:qq" json:"qq"`
	Status            string      `gorm:"type:varchar(20);default:'pending';column:status" json:"status"` // pending, approved, rejected
	Remark            string      `gorm:"type:text;column:remark" json:"remark"`
	LastViewTime      *time.Time  `gorm:"type:datetime;column:last_view_time" json:"lastViewTime"`
	AdminLastViewTime *time.Time  `gorm:"type:datetime;column:admin_last_view_time" json:"adminLastViewTime"`
	CreatedAt         time.Time   `gorm:"autoCreateTime;column:created_at" json:"createdAt"`

	// 关联字段
	Competition *Competition `gorm:"foreignKey:CompetitionID" json:"competition,omitempty"`
}

// TableName 指定表名
func (Registration) TableName() string {
	return "histcontest_register_registrations"
}

// FieldConfig 字段配置（用于存储在 Competition.Fields 中）
type FieldConfig struct {
	Name      bool `json:"name"`
	Class     bool `json:"class"`
	College   bool `json:"college"`
	StudentID bool `json:"studentId"`
	Gender    bool `json:"gender"`
	ShirtSize bool `json:"shirtSize"`
	TeamName  bool `json:"teamName"`
	QQ        bool `json:"qq"`
}
