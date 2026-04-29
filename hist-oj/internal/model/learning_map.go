package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	LearningMapStatusDraft     = "draft"
	LearningMapStatusPublished = "published"

	LearningMapNodeTypeKnowledge = "knowledge"
	LearningMapNodeTypeProblem   = "problem"

	LearningMapEdgeTypePrerequisite = "prerequisite"
	LearningMapEdgeTypeRelated      = "related"

	LearningMapProgressLocked     = "locked"
	LearningMapProgressAvailable  = "available"
	LearningMapProgressInProgress = "in_progress"
	LearningMapProgressCompleted  = "completed"
	LearningMapProgressMastered   = "mastered"
)

// LearningMap 算法航海图
// status: draft | published
type LearningMap struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"type:varchar(120);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"type:varchar(20);not null;default:'draft';index:idx_status" json:"status"`
	CreatedAt   time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`
}

func (LearningMap) TableName() string {
	return "learning_map"
}

// LearningMapNode 航海节点
type LearningMapNode struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MapID            uint64    `gorm:"type:bigint unsigned;not null;index:idx_map_id" json:"mapId"`
	Type             string    `gorm:"type:varchar(20);not null;index:idx_type" json:"type"` // knowledge | problem
	Title            string    `gorm:"type:varchar(160);not null" json:"title"`
	Description      string    `gorm:"type:text" json:"description"`
	Difficulty       string    `gorm:"type:varchar(20);default:'beginner'" json:"difficulty"`
	Tags             string    `gorm:"type:text" json:"tags"` // JSON string array
	X                float64   `gorm:"type:double;not null;default:0" json:"x"`
	Y                float64   `gorm:"type:double;not null;default:0" json:"y"`
	Level            int       `gorm:"type:int;default:0" json:"level"`
	Region           string    `gorm:"type:varchar(100)" json:"region"`
	Published        bool      `gorm:"type:tinyint(1);not null;default:1" json:"published"`
	KnowledgeContent string    `gorm:"type:longtext" json:"knowledgeContent"`
	ProblemID        *uint64   `gorm:"type:bigint unsigned;index:idx_problem_id" json:"problemId"`
	ProblemDisplayID string    `gorm:"type:varchar(80);index:idx_problem_display_id" json:"problemDisplayId"`
	Metadata         string    `gorm:"type:longtext" json:"metadata"` // JSON string object
	CreatedAt        time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`
}

func (LearningMapNode) TableName() string {
	return "learning_map_node"
}

// LearningMapEdge 连线
type LearningMapEdge struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MapID        uint64    `gorm:"type:bigint unsigned;not null;index:idx_map_id" json:"mapId"`
	SourceNodeID uint64    `gorm:"type:bigint unsigned;not null;index:idx_source_node_id" json:"sourceNodeId"`
	TargetNodeID uint64    `gorm:"type:bigint unsigned;not null;index:idx_target_node_id" json:"targetNodeId"`
	Type         string    `gorm:"type:varchar(20);not null;default:'prerequisite';index:idx_edge_type" json:"type"` // prerequisite | related
	CreatedAt    time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`
}

func (LearningMapEdge) TableName() string {
	return "learning_map_edge"
}

// UserLearningProgress 用户航海图进度
type UserLearningProgress struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      string     `gorm:"type:varchar(32);not null;index:idx_user_id;uniqueIndex:uk_user_map_node,priority:1" json:"userId"`
	MapID       uint64     `gorm:"type:bigint unsigned;not null;index:idx_map_id;uniqueIndex:uk_user_map_node,priority:2" json:"mapId"`
	NodeID      uint64     `gorm:"type:bigint unsigned;not null;index:idx_node_id;uniqueIndex:uk_user_map_node,priority:3" json:"nodeId"`
	Status      string     `gorm:"type:varchar(20);not null;default:'locked'" json:"status"`
	CompletedAt *time.Time `gorm:"type:datetime" json:"completedAt"`
	MasteredAt  *time.Time `gorm:"type:datetime" json:"masteredAt"`
	CreatedAt   time.Time  `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`
}

func (UserLearningProgress) TableName() string {
	return "user_learning_progress"
}

func InitLearningMapTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&LearningMap{},
		&LearningMapNode{},
		&LearningMapEdge{},
		&UserLearningProgress{},
	)
}
