package model

import (
	"time"

	"gorm.io/gorm"
)

// ClassroomUserRole 班级用户角色表
type ClassroomUserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UID       string    `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	Role      string    `gorm:"type:varchar(20);not null" json:"role"` // teacher, student, admin
	CreatedAt time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (ClassroomUserRole) TableName() string {
	return "classroom_user_role"
}

// ClassroomTeacher 班级教师关联表（支持一个班级多个教师）
type ClassroomTeacher struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID uint64    `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	TeacherID   string    `gorm:"type:varchar(32);not null;index:idx_teacher_id" json:"teacherId"`
	Status      int       `gorm:"type:int;default:1;index:idx_status" json:"status"` // 1: 正常, 0: 已移除
	CreatedAt   time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Classroom *Classroom `gorm:"foreignKey:ClassroomID" json:"classroom,omitempty"`
	Teacher   *UserInfo  `gorm:"foreignKey:TeacherID;references:UUID" json:"teacher,omitempty"`
}

// TableName 指定表名
func (ClassroomTeacher) TableName() string {
	return "classroom_teacher"
}

// Classroom 班级表
type Classroom struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassName   string    `gorm:"type:varchar(100);not null" json:"className"`
	ClassBelong string    `gorm:"type:varchar(100);not null" json:"classBelong"`
	ClassCode   string    `gorm:"type:varchar(8);not null;uniqueIndex" json:"classCode"`
	TeacherID   string    `gorm:"type:varchar(32);not null;index:idx_teacher_id" json:"teacherId"` // 主教师ID，保留用于向后兼容
	Status      int       `gorm:"type:int;default:1" json:"status"` // 1: 正常, 0: 已删除
	CreatedAt   time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Teacher  *UserInfo            `gorm:"foreignKey:TeacherID;references:UUID" json:"teacher,omitempty"`
	Teachers []ClassroomTeacher   `gorm:"foreignKey:ClassroomID" json:"teachers,omitempty"` // 班级所有教师
	Students []ClassroomStudent   `gorm:"foreignKey:ClassroomID" json:"students,omitempty"`
}

// TableName 指定表名
func (Classroom) TableName() string {
	return "classroom"
}

// ClassroomStudent 班级学生表
type ClassroomStudent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID  uint64    `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	UID          string    `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	RealName     string    `gorm:"type:varchar(50);not null" json:"realName"`
	Gender       string    `gorm:"type:varchar(10)" json:"gender"`
	StudentClass string    `gorm:"type:varchar(100)" json:"studentClass"`
	StudentNo    string    `gorm:"type:varchar(50)" json:"studentNo"`
	Status       int       `gorm:"type:int;default:1" json:"status"` // 1: 正常, 0: 已移除
	CreatedAt    time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	User *UserInfo `gorm:"foreignKey:UID;references:UUID" json:"user,omitempty"`
}

// TableName 指定表名
func (ClassroomStudent) TableName() string {
	return "classroom_student"
}

// ClassroomCheckin 签到表
type ClassroomCheckin struct {
	ID                     uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID            uint64     `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	CheckinCode            string     `gorm:"type:varchar(20);not null;index:idx_checkin_code" json:"checkinCode"`
	CheckinName            string     `gorm:"type:varchar(100)" json:"checkinName"`
	CheckinType            string     `gorm:"type:varchar(20);default:'code'" json:"checkinType"` // code 或 qrcode
	QrcodeToken            string     `gorm:"type:varchar(255)" json:"qrcodeToken"`
	QrcodeExpiresAt        *time.Time `gorm:"type:datetime" json:"qrcodeExpiresAt"`
	QrcodeRefreshInterval   int        `gorm:"type:int;default:15" json:"qrcodeRefreshInterval"` // 二维码刷新间隔(秒)
	StartTime              time.Time  `gorm:"type:datetime;not null" json:"startTime"`
	EndTime                *time.Time `gorm:"type:datetime" json:"endTime"`
	Status                 int        `gorm:"type:int;default:1" json:"status"` // 1: 进行中, 2: 已结束
	CreatedAt              time.Time  `gorm:"column:create_time;autoCreateTime" json:"createdAt"`

	// 关联字段
	Records []ClassroomCheckinRecord `gorm:"foreignKey:CheckinID" json:"records,omitempty"`
}

// TableName 指定表名
func (ClassroomCheckin) TableName() string {
	return "classroom_checkin"
}

// ClassroomCheckinRecord 签到记录表
type ClassroomCheckinRecord struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CheckinID   uint64     `gorm:"type:bigint unsigned;not null;index:idx_checkin_id" json:"checkinId"`
	UID         string     `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	Status      string     `gorm:"type:varchar(20);default:present" json:"status"` // present, absent, sick_leave, personal_leave
	CheckinTime *time.Time `gorm:"type:datetime" json:"checkinTime"`
	Remark      string     `gorm:"type:varchar(200)" json:"remark"`
	CreatedAt   time.Time  `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Student      *UserInfo           `gorm:"foreignKey:UID;references:UUID" json:"student,omitempty"`
	ClassStudent *ClassroomStudent   `gorm:"-" json:"classStudent,omitempty"` // 班级学生信息（包含真实姓名）
}

// TableName 指定表名
func (ClassroomCheckinRecord) TableName() string {
	return "classroom_checkin_record"
}

// QuestionBank 题库表
type QuestionBank struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title      string    `gorm:"type:varchar(200);not null" json:"title"`
	Type       string    `gorm:"type:varchar(20);not null;index:idx_type" json:"type"` // choice, judge, subjective, programming
	Content    string    `gorm:"type:text;not null" json:"content"`
	Options    *string   `gorm:"type:json" json:"options"` // JSON格式的选项（可为NULL）
	Answer     string    `gorm:"type:text" json:"answer"`
	Difficulty int       `gorm:"type:int;default:1" json:"difficulty"` // 1: 简单, 2: 中等, 3: 困难
	Score      int       `gorm:"type:int;default:2" json:"score"`
	CreatorID  string    `gorm:"type:varchar(32);not null;index:idx_creator_id" json:"creatorId"`
	IsShared   int       `gorm:"type:int;default:0;index:idx_is_shared" json:"isShared"` // 0: 个人, 1: 共享
	ProblemID  *string   `gorm:"type:varchar(50)" json:"problemId"` // 关联的OJ题目ID（编程题,字符串类型支持"0001"等格式）
	Status     int       `gorm:"type:int;default:1;index:idx_status" json:"status"` // 1: 正常, 0: 已删除
	CreatedAt  time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Creator *UserInfo `gorm:"foreignKey:CreatorID;references:UUID" json:"creator,omitempty"`
}

// TableName 指定表名
func (QuestionBank) TableName() string {
	return "question_bank"
}

// ClassroomHomework 作业/考试表
type ClassroomHomework struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID  uint64    `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	Title        string    `gorm:"type:varchar(200);not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	StartTime    time.Time `gorm:"type:datetime;not null" json:"startTime"`
	EndTime      time.Time `gorm:"type:datetime;not null" json:"endTime"`
	ShowScore    int       `gorm:"type:int;default:1" json:"showScore"` // 完成后是否显示成绩
	ShowHomework int       `gorm:"type:int;default:0" json:"showHomework"` // 完成后是否查看作业题目
	ShowAnswer   int       `gorm:"type:int;default:0" json:"showAnswer"` // 学生提交后是否可以查看答案
	Status       int       `gorm:"type:int;default:1;index:idx_status" json:"status"` // 1: 未开始, 2: 进行中, 3: 已结束
	CreatedAt    time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 考试模式字段
	IsExamMode              int `gorm:"type:int;default:0;comment:是否考试模式(0否1是)" json:"isExamMode"`
	ExamDuration            int `gorm:"type:int;not null;default:60;comment:考试时长(分钟)" json:"examDuration"`
	AllowSubmitAfterMinutes int `gorm:"type:int;default:0;comment:开考后多少分钟允许交卷(0表示立即允许)" json:"allowSubmitAfterMinutes"`
	DisableCopyPaste        int `gorm:"type:int;default:1;comment:是否禁止复制粘贴(0否1是)" json:"disableCopyPaste"`
	RequireFullscreen       int `gorm:"type:int;default:1;comment:是否要求全屏(0否1是)" json:"requireFullscreen"`
	DisallowTabSwitch       int `gorm:"type:int;default:1;comment:是否禁止切换标签页(0否1是)" json:"disallowTabSwitch"`

	// 关联字段
	Questions   []HomeworkQuestion `gorm:"foreignKey:HomeworkID" json:"questions,omitempty"`
	IsCompleted bool               `gorm:"-" json:"isCompleted"` // 学生是否已完成（非数据库字段，仅用于API返回）
}

// TableName 指定表名
func (ClassroomHomework) TableName() string {
	return "classroom_homework"
}

// HomeworkQuestion 作业题目关联表
type HomeworkQuestion struct {
	ID            uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeworkID    uint64        `gorm:"type:bigint unsigned;not null;index:idx_homework_id" json:"homeworkId"`
	QuestionID    *uint64       `gorm:"type:bigint unsigned;index:idx_question_id" json:"questionId,omitempty"`    // 题库题目ID（可为空，编程题为空）
	ProblemID     *string       `gorm:"type:varchar(50);index:idx_problem_id" json:"problemId,omitempty"`       // HOJ 题目ID（编程题使用，字符串类型）
	QuestionOrder int           `gorm:"type:int;not null" json:"questionOrder"`
	Score         int           `gorm:"type:int;default:2" json:"score"`
	CreatedAt     time.Time     `gorm:"column:create_time;autoCreateTime" json:"createdAt"`

	// 关联字段
	Question *QuestionBank `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

// TableName 指定表名
func (HomeworkQuestion) TableName() string {
	return "homework_question"
}

// HomeworkSubmit 作业提交记录表
type HomeworkSubmit struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeworkID          uint64   `gorm:"type:bigint unsigned;not null;index:idx_homework_id" json:"homeworkId"`
	QuestionID          *uint64  `gorm:"type:bigint unsigned;index:idx_question_id" json:"questionId,omitempty"`  // 题库题目ID（可为空）
	ProblemID           *string  `gorm:"type:varchar(50);index:idx_problem_id" json:"problemId,omitempty"`       // HOJ 题目ID（编程题使用）
	UID                 string   `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	Answer              string   `gorm:"type:text" json:"answer"`
	Attachment          string   `gorm:"type:varchar(1000)" json:"attachment"` // 图片附件URL（主观题使用，多个图片用逗号分隔）
	SubmitID            *uint64  `gorm:"type:bigint unsigned" json:"submitId"` // 提交记录ID（编程题）
	Score               float64  `gorm:"type:decimal(5,2);default:0" json:"score"`
	IsScored            int      `gorm:"type:int;default:0" json:"isScored"` // 是否已批改
	IsOfficiallySubmitted int    `gorm:"type:int;default:0" json:"isOfficiallySubmitted"` // 是否已正式提交（0=草稿自动保存，1=用户点击提交）
	JudgeResult         string   `gorm:"type:varchar(50)" json:"judgeResult"` // 评测结果（编程题）
	CreatedAt           time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 考试模式字段
	ExamStartTime           *time.Time `gorm:"type:datetime;comment:考试开始时间" json:"examStartTime,omitempty"`
	ExamEndTime             *time.Time `gorm:"type:datetime;comment:考试结束时间" json:"examEndTime,omitempty"`
	IsForcedSubmit          int        `gorm:"type:int;default:0;comment:是否强制收卷(0否1是)" json:"isForcedSubmit"`
	TabSwitchCount          int        `gorm:"type:int;default:0;comment:切换标签页次数" json:"tabSwitchCount"`
	FullscreenExitCount     int        `gorm:"type:int;default:0;comment:退出全屏次数" json:"fullscreenExitCount"`
	CopyPasteAttemptCount   int        `gorm:"type:int;default:0;comment:尝试复制粘贴次数" json:"copyPasteAttemptCount"`
	DeviceInfo              string     `gorm:"type:varchar(500);comment:设备信息" json:"deviceInfo"`
	BrowserInfo             string     `gorm:"type:varchar(500);comment:浏览器信息" json:"browserInfo"`

	// 关联字段
	Student  *UserInfo     `gorm:"foreignKey:UID;references:UUID" json:"student,omitempty"`
	Question *QuestionBank `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

// TableName 指定表名
func (HomeworkSubmit) TableName() string {
	return "homework_submit"
}

// ClassroomFolder 资料库文件夹表
type ClassroomFolder struct {
	ID          uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID uint64            `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	FolderName  string            `gorm:"type:varchar(100);not null" json:"folderName"`
	ParentID    uint64            `gorm:"type:bigint unsigned;default:0;index:idx_parent_id" json:"parentId"` // 0表示根目录
	CreatorID   string            `gorm:"type:varchar(32);not null" json:"creatorId"`
	SortOrder   int               `gorm:"type:int;default:0" json:"sortOrder"`
	Status      int               `gorm:"type:int;default:1;index:idx_status" json:"status"` // 1: 正常, 0: 已删除
	CreatedAt   time.Time         `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time         `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Children  []ClassroomFolder  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Materials []ClassroomMaterial `gorm:"foreignKey:FolderID" json:"materials,omitempty"`
}

// TableName 指定表名
func (ClassroomFolder) TableName() string {
	return "classroom_folder"
}

// ClassroomMaterial 资料库文件表
type ClassroomMaterial struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	FolderID      uint64    `gorm:"type:bigint unsigned;not null;index:idx_folder_id" json:"folderId"`
	FileName      string    `gorm:"type:varchar(255);not null" json:"fileName"`
	FileType      string    `gorm:"type:varchar(20);not null" json:"fileType"` // pdf, word, ppt, txt, mp4
	FilePath      string    `gorm:"type:varchar(500);not null" json:"filePath"`
	FileSize      uint64    `gorm:"type:bigint unsigned" json:"fileSize"` // 字节
	CreatorID     string    `gorm:"type:varchar(32);not null;index:idx_creator_id" json:"creatorId"`
	IsShared      int       `gorm:"type:int;default:0;index:idx_is_shared" json:"isShared"` // 0: 个人, 1: 共享
	DownloadCount int       `gorm:"type:int;default:0" json:"downloadCount"`
	Status        int       `gorm:"type:int;default:1;index:idx_status" json:"status"` // 1: 正常, 0: 已删除
	CreatedAt     time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Creator    *UserInfo                      `gorm:"foreignKey:CreatorID;references:UUID" json:"creator,omitempty"`
	Permission *ClassroomMaterialPermission   `gorm:"-" json:"permission,omitempty"` // 当前用户的权限（不存储在数据库）
}

// TableName 指定表名
func (ClassroomMaterial) TableName() string {
	return "classroom_material"
}

// ClassroomMaterialPermission 资料库文件权限表
type ClassroomMaterialPermission struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID   uint64    `gorm:"type:bigint unsigned;not null;index:idx_material_id" json:"materialId"`
	StudentUID   string    `gorm:"type:varchar(32);not null;index:idx_student_uid" json:"studentUid"`
	CanPreview   int       `gorm:"type:int;default:0;comment:是否可预览(0否1是)" json:"canPreview"`
	CanDownload  int       `gorm:"type:int;default:0;comment:是否可下载(0否1是)" json:"canDownload"`
	CreatedAt    time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updatedAt"`

	// 关联字段
	Material *ClassroomMaterial `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	Student  *UserInfo          `gorm:"foreignKey:StudentUID;references:UUID" json:"student,omitempty"`
}

// TableName 指定表名
func (ClassroomMaterialPermission) TableName() string {
	return "classroom_material_permission"
}

// ClassroomRandomPick 随机选人记录表
type ClassroomRandomPick struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID uint64    `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	PickedUID   string    `gorm:"type:varchar(32);not null" json:"pickedUid"`
	PickTime    time.Time `gorm:"type:datetime;autoCreateTime;index:idx_pick_time" json:"pickTime"`

	// 关联字段
	PickedUser   *UserInfo         `gorm:"foreignKey:PickedUID;references:UUID" json:"pickedUser,omitempty"`
	StudentInfo *ClassroomStudent `gorm:"-" json:"studentInfo,omitempty"` // 不在数据库中，仅用于API返回
}

// TableName 指定表名
func (ClassroomRandomPick) TableName() string {
	return "classroom_random_pick"
}

// ClassroomMessage 班级消息表
type ClassroomMessage struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassroomID uint64    `gorm:"type:bigint unsigned;not null;index:idx_classroom_id" json:"classroomId"`
	SenderID    string    `gorm:"type:varchar(32);not null" json:"senderId"`
	Content     string    `gorm:"type:text" json:"content"`
	ImageURL    string    `gorm:"type:varchar(500)" json:"imageUrl"`
	MsgType     string    `gorm:"type:varchar(20);default:text" json:"msgType"` // text, image, system
	CreatedAt   time.Time `gorm:"column:create_time;autoCreateTime;index:idx_create_time" json:"createdAt"`

	// 关联字段
	Sender      *UserInfo         `gorm:"foreignKey:SenderID;references:UUID" json:"sender,omitempty"`
	StudentInfo *ClassroomStudent `gorm:"-" json:"studentInfo,omitempty"` // 不存储到数据库，仅用于查询时返回
}

// TableName 指定表名
func (ClassroomMessage) TableName() string {
	return "classroom_message"
}

// StudentQuestionOrder 学生题目顺序映射表（考试模式用）
type StudentQuestionOrder struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeworkID   uint64 `gorm:"type:bigint unsigned;not null;index:idx_homework_id" json:"homeworkId"`
	UID          string `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	OrderMapping string `gorm:"type:text;not null;comment:题目顺序映射JSON" json:"orderMapping"`
	CreatedAt    time.Time `gorm:"column:create_time;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (StudentQuestionOrder) TableName() string {
	return "student_question_order"
}

// ExamViolationLog 考试违规日志表
type ExamViolationLog struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeworkID    uint64 `gorm:"type:bigint unsigned;not null;index:idx_homework_id" json:"homeworkId"`
	UID           string `gorm:"type:varchar(32);not null;index:idx_uid" json:"uid"`
	ViolationType string `gorm:"type:varchar(50);not null;index:idx_violation_type;comment:违规类型(tab_switch, fullscreen_exit, copy_attempt, paste_attempt, context_menu, devtools_attempt)" json:"violationType"`
	Description   string `gorm:"type:text;comment:违规详情" json:"description"`
	IP            string `gorm:"type:varchar(50);comment:IP地址" json:"ip"`
	CreatedAt     time.Time `gorm:"column:create_time;autoCreateTime;index:idx_create_time" json:"createdAt"`
}

// TableName 指定表名
func (ExamViolationLog) TableName() string {
	return "exam_violation_log"
}

// InitClassroomTables 初始化班级相关数据库表
func InitClassroomTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&ClassroomUserRole{},
		&Classroom{},
		&ClassroomTeacher{},
		&ClassroomStudent{},
		&ClassroomCheckin{},
		&ClassroomCheckinRecord{},
		&QuestionBank{},
		&ClassroomHomework{},
		&HomeworkQuestion{},
		&HomeworkSubmit{},
		&ClassroomFolder{},
		&ClassroomMaterial{},
		&ClassroomMaterialPermission{},
		&ClassroomRandomPick{},
		&ClassroomMessage{},
		&StudentQuestionOrder{},
		&ExamViolationLog{},
	)
}
