package model

import (
	"time"

	"gorm.io/gorm"
)

// ProblemSet 题目集
type ProblemSet struct {
	ID          uint64            `json:"id" gorm:"primaryKey"`
	UserID      uint64            `json:"user_id" gorm:"not null;index:idx_user_id;index:idx_created_at"`
	Title       string            `json:"title" gorm:"size:255;not null"`
	Author      string            `json:"author" gorm:"size:100"`
	ContestDate *time.Time        `json:"contest_date" gorm:"type:date"`
	CreatedAt   time.Time         `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Problems    []ProblemSetProblem `json:"problems,omitempty" gorm:"foreignKey:SetID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ProblemSet) TableName() string {
	return "xcpc_problem_set"
}

// ProblemSetProblem 题目集中的题目
type ProblemSetProblem struct {
	ID            uint64          `json:"id" gorm:"primaryKey"`
	SetID         uint64          `json:"set_id" gorm:"not null;index:idx_set_id"`
	ProblemLetter string          `json:"problem_letter" gorm:"size:1"`
	Title         string          `json:"title" gorm:"size:255;not null"`
	TimeLimit     int             `json:"time_limit" gorm:"default:1000"`
	Description   string          `json:"description" gorm:"type:text"`
	InputFormat   string          `json:"input_format" gorm:"type:text"`
	OutputFormat  string          `json:"output_format" gorm:"type:text"`
	Note          string          `json:"note" gorm:"type:text"`
	SortOrder     int             `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Examples      []ProblemExample `json:"examples,omitempty" gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ProblemSetProblem) TableName() string {
	return "xcpc_problem_set_problem"
}

// ProblemExample 题目样例
type ProblemExample struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	ProblemID   uint64    `json:"problem_id" gorm:"not null;index:idx_problem_id;uniqueIndex:uk_problem_example"`
	ExampleNo   int       `json:"example_no" gorm:"not null;uniqueIndex:uk_problem_example"`
	Description string    `json:"description" gorm:"size:500"`
	Input       string    `json:"input" gorm:"type:text;not null"`
	Output      string    `json:"output" gorm:"type:text;not null"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProblemExample) TableName() string {
	return "xcpc_problem_example"
}

// BeforeCreate GORM hook
func (p *ProblemSet) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeCreate GORM hook
func (p *ProblemSetProblem) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeCreate GORM hook
func (p *ProblemExample) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// ProblemSetImage 题目集图片
type ProblemSetImage struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	SetID      uint64    `json:"set_id" gorm:"not null;index:idx_set_id;uniqueIndex:uk_set_filename"`
	Filename   string    `json:"filename" gorm:"size:255;not null;uniqueIndex:uk_set_filename"`
	FilePath   string    `json:"file_path" gorm:"size:500;not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	MimeType   string    `json:"mime_type" gorm:"size:100"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProblemSetImage) TableName() string {
	return "xcpc_problem_set_image"
}

// BeforeCreate GORM hook
func (p *ProblemSetImage) BeforeCreate(tx *gorm.DB) error {
	return nil
}
