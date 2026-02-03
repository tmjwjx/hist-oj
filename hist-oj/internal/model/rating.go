package model

import (
	"time"

	"gorm.io/gorm"
)

// RatingHistory Rating历史记录
type RatingHistory struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UID          string     `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	ContestID    *uint64    `gorm:"type:bigint unsigned;default:0;index:idx_contest_id" json:"contest_id"` // 改为指针，0 表示不关联比赛（手动调整）
	ContestTitle string     `gorm:"-" json:"contest_title"` // 不映射到数据库，仅用于返回
	ContestTime  *time.Time `gorm:"-" json:"contest_time"` // 比赛时间，从Contest表关联查询
	OldRating    *int       `gorm:"type:int" json:"old_rating"`
	NewRating    int        `gorm:"type:int;not null" json:"new_rating"`
	RatingChange int        `gorm:"type:int;not null" json:"rating_change"`
	Rank         int        `gorm:"type:int;not null" json:"rank"`
	Participants int        `gorm:"type:int;not null" json:"participants"`
	Reason       string     `gorm:"type:varchar(255);default:''" json:"reason"` // 操作原因（手动调整时的备注）
	IsManual     bool       `gorm:"type:tinyint(1);default:0;index:idx_is_manual" json:"is_manual"` // 是否为手动调整
	OperatorUID  string     `gorm:"type:varchar(32);default:''" json:"operator_uid"` // 操作人UID
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"` // 记录创建时间
}

// TableName 指定表名
func (RatingHistory) TableName() string {
	return "rating_history"
}

// ContestRatingStatus 比赛Rating状态
type ContestRatingStatus struct {
	ContestID        uint64     `gorm:"primaryKey;type:bigint unsigned;column:contest_id" json:"contestId"`
	IsRated          bool       `gorm:"type:tinyint(1);default:0;column:is_rated" json:"isRated"`
	RatingCalculated bool       `gorm:"type:tinyint(1);default:0;index:idx_rating_calculated;column:rating_calculated" json:"ratingCalculated"`
	CalculatedAt     *time.Time `gorm:"type:datetime;column:calculated_at" json:"calculatedAt"`
	CreatedAt        time.Time  `gorm:"type:datetime(3);column:created_at;autoCreateTime:milli" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"type:datetime(3);column:updated_at;autoUpdateTime:milli" json:"updatedAt"`
}

// TableName 指定表名
func (ContestRatingStatus) TableName() string {
	return "contest_rating_status"
}

// UserRecord 用户记录（对应HOJ的user_record表）
type UserRecord struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UID         string    `gorm:"type:varchar(32);not null;uniqueIndex" json:"uid"`
	Rating      *int      `gorm:"type:int" json:"rating"`           // HOJ 原有的 rating 字段
	HistRating  *int      `gorm:"type:int;column:hist_rating" json:"histRating"` // 新增的 hist_rating 字段
	GmtCreate   time.Time `gorm:"type:datetime;column:gmt_create" json:"gmtCreate"`
	GmtModified time.Time `gorm:"type:datetime;column:gmt_modified;autoUpdateTime" json:"gmtModified"`
}

// TableName 指定表名
func (UserRecord) TableName() string {
	return "user_record"
}

// UserInfo 用户信息（对应HOJ的user_info表）
type UserInfo struct {
	UUID      string    `gorm:"primaryKey;type:varchar(32)" json:"uuid"`
	Username  string    `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	Password  string    `gorm:"type:varchar(255)" json:"password"`
	Nickname  string    `gorm:"type:varchar(255)" json:"nickname"`
	Realname  string    `gorm:"type:varchar(255)" json:"realname"` // 真实姓名
	Rating    int       `gorm:"type:int;default:1200" json:"rating"` // 当前 Rating（已废弃，使用HistRating）
	HistRating int      `gorm:"-" json:"histRating"` // Hist Rating（从user_record表获取）
	Status    int       `gorm:"type:int;default:0" json:"status"` // 0: 正常, 1: 禁用
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (UserInfo) TableName() string {
	return "user_info"
}

// Contest 比赛信息（对应HOJ的contest表）
type Contest struct {
	ID        uint64    `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	UID       string    `gorm:"type:varchar(32);not null;column:uid" json:"uid"` // 创建者ID
	Author    string    `gorm:"type:varchar(255);column:author" json:"author"` // 创建者用户名
	Type      int       `gorm:"type:int;not null;default:0" json:"type"` // 0: ACM, 1: OI
	IsRating  bool      `gorm:"type:tinyint(1);not null;default:0;column:is_rating" json:"isRating"` // 0: unrated, 1: rated
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	StartTime time.Time `gorm:"type:datetime;column:start_time" json:"startTime"`
	EndTime   time.Time `gorm:"type:datetime;column:end_time" json:"endTime"`
	Duration  int64     `gorm:"type:bigint" json:"duration"`
	Status    int       `gorm:"type:int" json:"status"` // -1: 未开始, 0: 进行中, 1: 已结束
	CreatedAt time.Time `gorm:"column:gmt_create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:gmt_modified" json:"updatedAt"`
}

// TableName 指定表名
func (Contest) TableName() string {
	return "contest"
}

// ContestRecord 比赛记录（对应HOJ的contest_record表）
// 这个表存储的是每个用户的每次提交记录
type ContestRecord struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CID         uint64    `gorm:"type:bigint unsigned;not null;index:idx_cid" json:"cid"`
	UID         string    `gorm:"type:varchar(255);not null;index:idx_uid" json:"uid"`
	PID         uint64    `gorm:"type:bigint unsigned" json:"pid"`
	CPID        uint64    `gorm:"type:bigint unsigned" json:"cpid"`
	Username    string    `gorm:"type:varchar(255)" json:"username"`
	Realname    string    `gorm:"type:varchar(255)" json:"realname"`
	DisplayID   string    `gorm:"type:varchar(255);column:display_id" json:"displayId"`
	SubmitID    uint64    `gorm:"type:bigint unsigned;column:submit_id" json:"submitId"`
	Status      int       `gorm:"type:int" json:"status"` // 0: AC, 其他: 非AC
	SubmitTime  time.Time `gorm:"type:datetime;column:submit_time" json:"submitTime"`
	Time        uint64    `gorm:"type:bigint unsigned" json:"time"`
	Score       *int      `gorm:"type:int" json:"score"`
	UseTime     int       `gorm:"type:int;column:use_time" json:"useTime"`
	FirstBlood  bool      `gorm:"type:tinyint(1);column:first_blood" json:"firstBlood"`
	Checked     bool      `gorm:"type:tinyint(1)" json:"checked"`
	GmtCreate   time.Time `gorm:"type:datetime;column:gmt_create" json:"gmtCreate"`
	GmtModified time.Time `gorm:"type:datetime;column:gmt_modified" json:"gmtModified"`
}

// TableName 指定表名
func (ContestRecord) TableName() string {
	return "contest_record"
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
	// 创建提交历史表
	if err := db.AutoMigrate(&SubmissionHistory{}); err != nil {
		return err
	}
	// 创建班级相关表
	if err := InitClassroomTables(db); err != nil {
		return err
	}
	// 创建比赛问题答疑相关表
	if err := db.AutoMigrate(&ContestQuestion{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&ContestQuestionReply{}); err != nil {
		return err
	}
	// 创建查重相关表
	if err := InitPlagiarismTables(db); err != nil {
		return err
	}
	return nil
}

// InitPlagiarismTables 初始化查重相关表
func InitPlagiarismTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&PlagiarismCheckConfig{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&PlagiarismCheck{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&PlagiarismResult{}); err != nil {
		return err
	}
	return nil
}

