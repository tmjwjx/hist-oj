package model

import (
	"time"
)

// BattleRoom 对战房间模型
type BattleRoom struct {
	ID                  int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RoomID              string     `json:"roomId" gorm:"column:room_id;uniqueIndex:uk_room_id"`
	HostID              string     `json:"hostId" gorm:"column:host_id;index:idx_host_id"`
	HostUsername        string     `json:"hostUsername" gorm:"column:host_username"`
	HostRating          *int       `json:"hostRating" gorm:"-"` // 从 user_record 表关联查询
	ChallengerID        *string    `json:"challengerId" gorm:"column:challenger_id"`
	ChallengerUsername  *string    `json:"challengerUsername" gorm:"column:challenger_username"`
	ChallengerRating    *int       `json:"challengerRating" gorm:"-"` // 从 user_record 表关联查询
	ChallengerReady     bool       `json:"challengerReady" gorm:"column:challenger_ready;default:false"` // 挑战者是否已准备
	ProblemID           *string    `json:"problemId" gorm:"column:problem_id"` // 显示ID（如 0051, Z001）
	Status              BattleStatus `json:"status" gorm:"column:status;index:idx_status;default:0"`
	WinnerID            *string    `json:"winnerId" gorm:"column:winner_id"`
	EndReason           *string    `json:"endReason" gorm:"column:end_reason"`
	StartTime           *time.Time `json:"startTime" gorm:"column:start_time"`
	EndTime             *time.Time `json:"endTime" gorm:"column:end_time"`
	GmtCreate           time.Time  `json:"gmtCreate" gorm:"column:gmt_create;autoCreateTime"`
	GmtModified         time.Time  `json:"gmtModified" gorm:"column:gmt_modified;autoUpdateTime"`
}

// TableName 指定表名
func (BattleRoom) TableName() string {
	return "battle_room"
}

// BattleRoomStatus 房间状态枚举
type BattleStatus int

const (
	BattleStatusWaiting BattleStatus = iota // 0-等待中
	BattleStatusInProgress                 // 1-对战中
	BattleStatusEnded                      // 2-已结束
)

// BattleEndReason 对战结束原因
const (
	EndReasonAC     = "ac"     // 一方AC
	EndReasonGiveUp = "giveup" // 一方放弃
	EndReasonTimeout = "timeout" // 超时
)

// BattleRecord 对战记录模型
type BattleRecord struct {
	ID                int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RoomID            string    `json:"roomId" gorm:"column:room_id"`
	UserID            string    `json:"userId" gorm:"column:user_id;index:idx_user_id"`
	Username          string    `json:"username" gorm:"column:username"`
	UserRating        *int      `json:"userRating" gorm:"-"` // 用户自己的rating（不存储在数据库，仅用于显示）
	OpponentID        string    `json:"opponentId" gorm:"column:opponent_id"`
	OpponentUsername  string    `json:"opponentUsername" gorm:"column:opponent_username"`
	OpponentRating    *int      `json:"opponentRating" gorm:"column:opponent_rating"` // 对手的rating
	ProblemID         string    `json:"problemId" gorm:"column:problem_id"` // 改为 string，支持显示ID
	ProblemTitle      string    `json:"problemTitle" gorm:"column:problem_title"`
	IsWinner          bool      `json:"isWinner" gorm:"column:is_winner;index:idx_is_winner"`
	EndReason         string    `json:"endReason" gorm:"column:end_reason"`
	SubmitCount       int       `json:"SubmitCount" gorm:"column:submit_count;default:0"`
	BattleTime        *int      `json:"battleTime" gorm:"column:battle_time;index"`
	GmtCreate         time.Time `json:"gmtCreate" gorm:"column:gmt_create;autoCreateTime;index"`
}

// TableName 指定表名
func (BattleRecord) TableName() string {
	return "battle_record"
}

// UserBattleStats 用户对战统计模型
type UserBattleStats struct {
	ID                int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID            string    `json:"userId" gorm:"column:user_id;uniqueIndex:uk_user_id"`
	Username          string    `json:"username" gorm:"column:username"`
	TotalBattles      int       `json:"totalBattles" gorm:"column:total_battles;default:0"`
	WinCount          int       `json:"winCount" gorm:"column:win_count;index:idx_win_count;default:0"`
	LoseCount         int       `json:"loseCount" gorm:"column:lose_count;default:0"`
	WinRate           float64   `json:"winRate" gorm:"column:win_rate;index:idx_win_rate;default:0.00"`
	TotalSubmitCount  int       `json:"totalSubmitCount" gorm:"column:total_submit_count;default:0"`
	AvgBattleTime     *int      `json:"avgBattleTime" gorm:"column:avg_battle_time"`
	Rating            *int      `json:"rating" gorm:"-"` // 从 rating 表关联查询
	GmtCreate         time.Time `json:"gmtCreate" gorm:"column:gmt_create;autoCreateTime"`
	GmtModified       time.Time `json:"gmtModified" gorm:"column:gmt_modified;autoUpdateTime"`
}

// TableName 指定表名
func (UserBattleStats) TableName() string {
	return "user_battle_stats"
}

// UserACProblem 用户AC题目（从HOJ获取）
type UserACProblem struct {
	Pid      int64  `json:"pid"`
	SubmitID int64  `json:"submitId"`
}

// Problem 题目信息（从HOJ获取）
type Problem struct {
	Pid        int64  `json:"pid"`
	ProblemID  string `json:"problemId"`
	Title      string `json:"title"`
	Difficulty int    `json:"difficulty"`
	Auth       int    `json:"auth"` // 1-公开，2-私有
}

// JudgeStatus 判题结果（从HOJ获取）
type JudgeStatus struct {
	SubmitID   int64  `json:"submitId"`
	Pid        int64  `json:"pid"`
	UID        string `json:"uid"`
	Username   string `json:"username"`
	Status     int    `json:"status"` // 0-Pending, 1-Accepted, ...
	SubmitTime string `json:"submitTime"`
}
