package model

import "time"

// SubmissionHistoryCase 判题终端本地判题测试点详情
type SubmissionHistoryCase struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SubmitID  string    `gorm:"type:varchar(50);index:idx_submit_seq,priority:1;column:submit_id" json:"submit_id"`
	CaseID    int64     `gorm:"column:case_id" json:"case_id"`
	Status    *int      `gorm:"column:status" json:"status"`
	Time      *int      `gorm:"column:time" json:"time"`
	Memory    *int      `gorm:"column:memory" json:"memory"`
	Score     *int      `gorm:"column:score" json:"score"`
	GroupNum  *int      `gorm:"column:group_num" json:"group_num"`
	Seq       *int      `gorm:"index:idx_submit_seq,priority:2;column:seq" json:"seq"`
	Mode      string    `gorm:"type:varchar(32);default:'default';column:mode" json:"mode"`
	Stderr    string    `gorm:"type:text;column:stderr" json:"stderr"`
	CreatedAt time.Time `gorm:"column:create_time;autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (SubmissionHistoryCase) TableName() string {
	return "submission_history_case"
}
