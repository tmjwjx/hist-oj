package model

import (
	"time"
)

// PlagiarismCheckConfig 查重配置表
type PlagiarismCheckConfig struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CID       uint64    `gorm:"type:bigint unsigned;not null;index:idx_cid;uniqueIndex:unique_cid_pid;column:cid;comment:比赛ID" json:"cid"`
	PID       uint64    `gorm:"type:bigint unsigned;not null;index:idx_pid;uniqueIndex:unique_cid_pid;column:pid;comment:题目ID" json:"pid"`
	CPID      uint64    `gorm:"type:bigint unsigned;not null;index:idx_cpid;column:cpid;comment:比赛题目ID" json:"cpid"`
	Threshold int       `gorm:"type:int;not null;default:50;comment:查重率阈值(0-100)" json:"threshold"`
	CreatedBy string    `gorm:"type:varchar(32);column:created_by;comment:创建者用户ID" json:"createdBy"`
	CreatedAt time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	UpdatedAt time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
}

// TableName 指定表名
func (PlagiarismCheckConfig) TableName() string {
	return "plagiarism_check_config"
}

// PlagiarismCheck 查重任务表
type PlagiarismCheck struct {
	ID            uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	CID           uint64      `gorm:"type:bigint unsigned;not null;index:idx_cid;column:cid;comment:比赛ID" json:"cid"`
	Status        string      `gorm:"type:varchar(20);not null;default:'pending';index:idx_status;comment:状态" json:"status"` // pending, running, completed, failed
	TotalPairs    int         `gorm:"type:int;not null;default:0;column:total_pairs;comment:总对比对数" json:"totalPairs"`
	CheckedPairs  int         `gorm:"type:int;not null;default:0;column:checked_pairs;comment:已检查对数" json:"checkedPairs"`
	TotalSubmissions int       `gorm:"type:int;not null;default:0;column:total_submissions;comment:总提交数" json:"totalSubmissions"`
	Progress      float64     `gorm:"type:decimal(5,2);default:0.00;column:progress;comment:进度百分比" json:"progress"`
	ErrorMessage  string      `gorm:"type:varchar(500);column:error_message;comment:错误信息" json:"errorMessage"`
	StartedBy     string      `gorm:"type:varchar(32);column:started_by;comment:启动者用户ID" json:"startedBy"`
	CreatedAt     time.Time   `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	UpdatedAt     time.Time   `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
	StartedAt     *time.Time  `gorm:"column:started_at;index:idx_started_at;comment:开始时间" json:"startedAt"`
	CompletedAt   *time.Time  `gorm:"column:completed_at;comment:完成时间" json:"completedAt"`

	// 关联字段
	Contest *Contest `gorm:"foreignKey:CID;references:ID" json:"contest,omitempty"`
	Results []PlagiarismResult `gorm:"foreignKey:CheckID;references:ID" json:"results,omitempty"`
}

// TableName 指定表名
func (PlagiarismCheck) TableName() string {
	return "plagiarism_check"
}

// PlagiarismResult 查重结果表
type PlagiarismResult struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CheckID          uint64    `gorm:"type:bigint unsigned;not null;index:idx_check_id;column:check_id;comment:查重任务ID" json:"checkId"`
	CID              uint64    `gorm:"type:bigint unsigned;not null;index:idx_cid;column:cid;comment:比赛ID" json:"cid"`
	CPID             uint64    `gorm:"type:bigint unsigned;not null;index:idx_cpid;column:cpid;comment:比赛题目ID" json:"cpid"`
	PID              uint64    `gorm:"type:bigint unsigned;not null;column:pid;comment:题目ID" json:"pid"`
	DisplayID        string    `gorm:"type:varchar(255);column:display_id;comment:题目显示ID" json:"displayId"`
	ProblemTitle     string    `gorm:"type:varchar(255);column:problem_title;comment:题目标题" json:"problemTitle"`
	SubmitID1        uint64    `gorm:"type:bigint unsigned;not null;column:submit_id_1;comment:提交ID1(judge表)" json:"submitId1"`
	SubmitID2        uint64    `gorm:"type:bigint unsigned;not null;column:submit_id_2;comment:提交ID2(judge表)" json:"submitId2"`
	ContestRecordID1 uint64    `gorm:"type:bigint unsigned;column:contest_record_id_1;comment:比赛记录ID1(contest_record表)" json:"contestRecordId1"`
	ContestRecordID2 uint64    `gorm:"type:bigint unsigned;column:contest_record_id_2;comment:比赛记录ID2(contest_record表)" json:"contestRecordId2"`
	UID1             string    `gorm:"type:varchar(32);not null;index:idx_uid_1;column:uid_1;comment:用户1 ID" json:"uid1"`
	UID2             string    `gorm:"type:varchar(32);not null;index:idx_uid_2;column:uid_2;comment:用户2 ID" json:"uid2"`
	Username1        string    `gorm:"type:varchar(255);column:username_1;comment:用户1 用户名" json:"username1"`
	Username2        string    `gorm:"type:varchar(255);column:username_2;comment:用户2 用户名" json:"username2"`
	Language         string    `gorm:"type:varchar(50);column:language;comment:编程语言" json:"language"`
	Similarity1to2   int       `gorm:"type:int;column:similarity_1_to_2;comment:1对2的相似度" json:"similarity1to2"`
	Similarity2to1   int       `gorm:"type:int;column:similarity_2_to_1;comment:2对1的相似度" json:"similarity2to1"`
	MaxSimilarity    int       `gorm:"type:int;index:idx_max_similarity;column:max_similarity;comment:最大相似度" json:"maxSimilarity"`
	IsOverThreshold  bool      `gorm:"type:tinyint(1);default:0;index:idx_is_over_threshold;column:is_over_threshold;comment:是否超过阈值" json:"isOverThreshold"`
	CreatedAt        time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`

	// 关联字段
	Check      *PlagiarismCheck `gorm:"foreignKey:CheckID;references:ID" json:"check,omitempty"`
	Contest    *Contest         `gorm:"foreignKey:CID;references:ID" json:"contest,omitempty"`
}

// TableName 指定表名
func (PlagiarismResult) TableName() string {
	return "plagiarism_result"
}

// Judge 提交记录表（HOJ 原生表，用于查重）
type Judge struct {
	SubmitID   uint64    `gorm:"primaryKey;autoIncrement;column:submit_id" json:"submitId"`
	PID        uint64    `gorm:"primaryKey;type:bigint unsigned;not null;index:idx_pid;column:pid;comment:题目ID" json:"pid"`
	DisplayPID string    `gorm:"primaryKey;type:varchar(255);not null;column:display_pid;comment:题目展示ID" json:"displayPid"`
	UID        string    `gorm:"primaryKey;type:varchar(32);not null;index:idx_uid;column:uid;comment:用户ID" json:"uid"`
	CID        uint64    `gorm:"primaryKey;type:bigint unsigned;not null;default:0;index:idx_cid;column:cid;comment:比赛ID" json:"cid"`
	Username   string    `gorm:"type:varchar(255);index:idx_username;column:username;comment:用户名" json:"username"`
	SubmitTime time.Time `gorm:"type:datetime;not null;column:submit_time;comment:提交时间" json:"submitTime"`
	Status     int       `gorm:"type:int;column:status;comment:结果码" json:"status"`
	Code       string    `gorm:"type:longtext;not null;column:code;comment:代码" json:"code"`
	Language   string    `gorm:"type:varchar(255);column:language;comment:代码语言" json:"language"`
	CPID       uint64    `gorm:"type:bigint unsigned;default:0;index:idx_cpid;column:cpid;comment:比赛题目ID" json:"cpid"`
}

// TableName 指定表名
func (Judge) TableName() string {
	return "judge"
}

// ContestProblem 比赛题目关联表（HOJ 原生表）
type ContestProblem struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DisplayID    string    `gorm:"type:varchar(255);not null;column:display_id;comment:题目在比赛中的显示ID" json:"displayId"`
	CID          uint64    `gorm:"type:bigint unsigned;not null;index:idx_cid;column:cid;comment:比赛ID" json:"cid"`
	PID          uint64    `gorm:"type:bigint unsigned;not null;index:idx_pid;column:pid;comment:题目ID" json:"pid"`
	DisplayTitle string    `gorm:"type:varchar(255);not null;column:display_title;comment:题目在比赛中的标题" json:"displayTitle"`
	Color        string    `gorm:"type:varchar(255);column:color;comment:气球颜色" json:"color"`
	CreatedAt    time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	UpdatedAt    time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`

	// 关联字段
	Contest *Contest `gorm:"foreignKey:CID;references:ID" json:"contest,omitempty"`
}

// TableName 指定表名
func (ContestProblem) TableName() string {
	return "contest_problem"
}
