package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/service"
)

func SetupRoutes(router *gin.Engine, handler *Handler, cfg *config.Config, db *gorm.DB) {
	// 应用CORS中间件到所有路由
	router.Use(CORSMiddleware())

	api := router.Group("/api")
	{
		// Rating 相关接口
		rating := api.Group("/rating")
		{
			// 查询接口：公开访问
			rating.GET("/user/:uid", handler.GetUserRating)
			rating.GET("/history/:uid", handler.GetRatingHistory)
			rating.GET("/color/:rating", handler.GetRatingColor)
			rating.GET("/contest/:contestId", handler.GetContestParticipantsRating)
			rating.GET("/contest/info/:contestId", handler.GetContestInfo)
			rating.POST("/batch", handler.GetBatchUserRating)
			rating.POST("/contest/batch", handler.GetBatchContestInfo)
			rating.GET("/rank", handler.GetRatingRank)

			// 管理接口：临时取消验证
			rating.POST("/calculate/:contestId", handler.CalculateRating)
			rating.POST("/initialize", handler.InitializeUserRating)
			rating.POST("/contest/set-rating-type", handler.SetContestRatingType)
			rating.POST("/trigger-scheduler", handler.TriggerScheduler)

			// 手动调整 rating（管理员，需要权限验证）
			admin := rating.Group("/admin")
			admin.Use(AdminAuthMiddleware())
			{
				admin.POST("/adjust", handler.AdjustUserRating)
				admin.GET("/history", handler.GetManualAdjustmentHistory)

				// Skip用户管理
				admin.POST("/contest/skip-users", handler.BatchSkipContestUsers)
				admin.GET("/contest/:contestId/skip-users", handler.GetContestSkipUsers)
				admin.DELETE("/contest/skip-users", handler.CancelSkip)
				admin.POST("/contest/:contestId/recalculate", handler.RecalculateFromContest)
				admin.GET("/recalculate-progress/:taskId", handler.GetRecalculateProgress)
				admin.POST("/contest/:contestId/reset-recalculate-lock", handler.ResetRecalculateLock)
				admin.POST("/contest/:contestId/sync-skip", handler.SyncContestSkipFlag)

				// 操作日志
				admin.GET("/operation-logs", handler.GetOperationLogs)
				admin.POST("/migrate-logs", handler.MigrateOperationLogs)
				admin.POST("/fix-logs-username", handler.FixOperationLogsUsername)

				// 测试接口（插入测试数据）
				admin.POST("/test-insert-logs", handler.TestInsertOperationLog)
			}
		}

		// 判题终端相关接口
		judgeService := service.NewJudgeService(db, cfg.HojAPI.BaseURL)
		handler.SetJudgeService(judgeService) // 设置 judge service 到 handler
		judgeHandler := NewJudgeHandler(judgeService)

		judge := api.Group("/judge")
		{
			judge.POST("/get-info", judgeHandler.GetInfo)
			judge.POST("/get-history", judgeHandler.GetHistory)
			judge.POST("/run-combined", judgeHandler.RunCombined)
			judge.POST("/submit", judgeHandler.Submit)
		}

		// 注册对战相关路由
		RegisterBattleRoutes(api)

		// 注册训练相关路由
		RegisterTrainingRoutes(api)

		// 班级功能相关接口
		classroom := api.Group("/classroom")
		{
			// 用户角色查询（需要认证）
			classroom.GET("/user/roles", AuthMiddleware(), handler.GetCurrentUserRoles)

			// 权限管理（管理员）
			admin := classroom.Group("/admin")
			{
				admin.POST("/role/grant", handler.GrantUserRole)
				admin.DELETE("/role/revoke", handler.RevokeUserRole)
				admin.GET("/roles/:userId", handler.GetUserRoles)
				admin.PUT("/user/:uid/roles", handler.UpdateUserRoles) // 更新 HOJ 用户角色
				admin.GET("/classrooms", handler.GetAllClassrooms) // 获取所有班级（管理员）
			}

			// 班级管理（教师）- 需要认证
			classroom.POST("/create", AuthMiddleware(), handler.CreateClassroom)
			classroom.DELETE("/:classroomId", AuthMiddleware(), handler.DeleteClassroom)
			classroom.GET("/list", AuthMiddleware(), handler.GetClassroomList)
			classroom.GET("/:classroomId", handler.GetClassroomDetail)

			// 学生加入班级 - 需要认证
			classroom.POST("/join", AuthMiddleware(), handler.JoinClassroom)
			classroom.GET("/my-classrooms", AuthMiddleware(), handler.GetStudentClassrooms)
			classroom.GET("/teacher-classrooms", AuthMiddleware(), handler.GetTeacherClassrooms)

			// 学生管理（教师）- 需要认证
			classroom.GET("/:classroomId/students", AuthMiddleware(), handler.GetClassroomStudents)
			classroom.DELETE("/student", AuthMiddleware(), handler.RemoveStudent)
			classroom.PUT("/student", AuthMiddleware(), handler.UpdateStudentInfo)

			// 学生班级个人信息管理（学生）- 需要认证
			classroom.GET("/:classroomId/my-info", AuthMiddleware(), handler.GetClassroomStudentInfo)
			classroom.PUT("/:classroomId/my-info", AuthMiddleware(), handler.UpdateClassroomStudentInfo)

			// 签到功能 - 需要认证
			classroom.POST("/checkin", AuthMiddleware(), handler.CreateCheckin)
			classroom.POST("/checkin/submit", AuthMiddleware(), handler.StudentCheckin)
			classroom.GET("/:classroomId/checkins", AuthMiddleware(), handler.GetCheckinList)
			classroom.GET("/:classroomId/checkins/student", AuthMiddleware(), handler.GetCheckinListForStudent) // 学生端安全API
			classroom.GET("/checkin/:checkinId/records", AuthMiddleware(), handler.GetCheckinRecords)
			classroom.POST("/checkin/:checkinId/record", AuthMiddleware(), handler.CreateCheckinRecord)
			classroom.PUT("/checkin/record", AuthMiddleware(), handler.UpdateCheckinRecord)
			classroom.POST("/checkin/:checkinId/end", AuthMiddleware(), handler.EndCheckin)
			classroom.PUT("/checkin/:checkinId", AuthMiddleware(), handler.UpdateCheckin)
			classroom.DELETE("/checkin/:checkinId", AuthMiddleware(), handler.DeleteCheckin)

			// 二维码签到功能 - 需要认证
			classroom.GET("/checkin/:checkinId/qrcode", AuthMiddleware(), handler.GetQrcodeInfo)
			classroom.POST("/checkin/:checkinId/qrcode/refresh", AuthMiddleware(), handler.RefreshQrcode)
			classroom.POST("/checkin/qrcode/submit", AuthMiddleware(), handler.SubmitQrcodeCheckin)

			// 题库功能（教师）- 需要认证
			classroom.POST("/question", AuthMiddleware(), handler.CreateQuestion)
			classroom.GET("/questions", AuthMiddleware(), handler.GetQuestionBank)
			classroom.PUT("/question/:questionId", AuthMiddleware(), handler.UpdateQuestion)
			classroom.DELETE("/question/:questionId", AuthMiddleware(), handler.DeleteQuestion)
			classroom.GET("/question/:questionId", handler.GetQuestionDetail)

			// 作业功能 - 需要认证
			classroom.POST("/homework", AuthMiddleware(), handler.CreateHomework)
			classroom.GET("/:classroomId/homeworks", AuthMiddleware(), handler.GetHomeworkList)
			// 具体路径要放在参数化路径之前
			classroom.POST("/homework/upload-attachment", AuthMiddleware(), handler.UploadHomeworkAttachment) // 上传作业附件
			classroom.POST("/homework/draft", AuthMiddleware(), handler.SaveHomeworkDraft) // 保存草稿
			classroom.POST("/homework/submit", AuthMiddleware(), handler.SubmitHomework)    // 正式提交
			classroom.POST("/homework/grade", AuthMiddleware(), handler.GradeHomework)
			classroom.POST("/homework/programming/grade", AuthMiddleware(), handler.GradeProgrammingHomework)
			classroom.POST("/homework/recalculate", AuthMiddleware(), handler.RecalculateScore)
			// 考试模式相关接口 - 需要认证
			classroom.POST("/homework/:homeworkId/start-exam", AuthMiddleware(), handler.StartExam)
			classroom.GET("/homework/:homeworkId/exam-status", AuthMiddleware(), handler.GetExamStatus)
			classroom.POST("/homework/violation", AuthMiddleware(), handler.LogViolation)
			classroom.GET("/homework/:homeworkId/exam-monitoring", AuthMiddleware(), handler.GetExamMonitoring)
			classroom.POST("/homework/force-submit", AuthMiddleware(), handler.ForceSubmit)
			classroom.POST("/homework/:homeworkId/force-submit-all", AuthMiddleware(), handler.ForceSubmitAll)
			// 参数化路径放在最后
			classroom.GET("/homework/:homeworkId", AuthMiddleware(), handler.GetHomeworkDetail)
			classroom.PUT("/homework/:homeworkId", AuthMiddleware(), handler.UpdateHomework)
			classroom.DELETE("/homework/:homeworkId", AuthMiddleware(), handler.DeleteHomework)
			classroom.GET("/homework/:homeworkId/submissions", AuthMiddleware(), handler.GetHomeworkSubmissions)
			classroom.GET("/homework/:homeworkId/status", AuthMiddleware(), handler.GetStudentHomeworkStatus)
			classroom.GET("/homework/:homeworkId/my-detail", AuthMiddleware(), handler.GetStudentHomeworkDetail)
			classroom.GET("/homework/:homeworkId/analysis", AuthMiddleware(), handler.GetHomeworkAnalysis)

			// 编程题提交记录 - 需要认证
			classroom.POST("/programming/submission", AuthMiddleware(), handler.SaveProgrammingSubmission)
			classroom.GET("/programming/submissions", AuthMiddleware(), handler.GetProgrammingSubmissions)

			// 资料库功能 - 需要认证
			classroom.POST("/folder", AuthMiddleware(), handler.CreateFolder)
			classroom.GET("/:classroomId/folders", AuthMiddleware(), handler.GetFolders)
			classroom.DELETE("/folder/:folderId", AuthMiddleware(), handler.DeleteFolder)
			classroom.PUT("/folder", AuthMiddleware(), handler.UpdateFolder)
			classroom.POST("/material/upload", AuthMiddleware(), handler.UploadMaterial)
			classroom.GET("/folder/:folderId/materials", AuthMiddleware(), handler.GetMaterials)
			classroom.DELETE("/material/:materialId", AuthMiddleware(), handler.DeleteMaterial)
			classroom.POST("/material/copy", AuthMiddleware(), handler.CopyMaterialToClassroom)

			// 资料库权限管理 - 需要认证
			classroom.GET("/material/:materialId/permissions", AuthMiddleware(), handler.GetMaterialPermissions)      // 获取权限设置
			classroom.POST("/material/permissions", AuthMiddleware(), handler.SetMaterialPermissions)              // 批量设置权限
			classroom.POST("/material/:materialId/permissions/batch", AuthMiddleware(), handler.BatchSetAllMaterialPermissions) // 全部开启/关闭权限
			classroom.GET("/material/:materialId/download", AuthMiddleware(), handler.DownloadMaterial)            // 下载资料（带权限验证）
			classroom.GET("/material/:materialId/pdf", AuthMiddleware(), handler.GetMaterialPDFBase64)             // 获取PDF base64（旧版，兼容）
			classroom.GET("/material/:materialId/pdf/binary", AuthMiddleware(), handler.GetMaterialPDFBinary)       // 获取PDF二进制（新版，性能更好）

			// 随机选人（教师）- 需要认证
			classroom.POST("/:classroomId/random-pick", AuthMiddleware(), handler.RandomPick)
			classroom.GET("/:classroomId/pick-history", AuthMiddleware(), handler.GetPickHistory)

			// 班级消息 - 需要认证
			classroom.POST("/message", AuthMiddleware(), handler.SendMessage)
			classroom.GET("/:classroomId/messages", AuthMiddleware(), handler.GetMessages)
			classroom.POST("/message/upload-image", AuthMiddleware(), handler.UploadMessageImage)
			classroom.DELETE("/message/:messageId", AuthMiddleware(), handler.RecallMessage)
		}

		// 比赛问题答疑 - 需要认证
		contestQuestion := api.Group("/contest-question")
		{
			contestQuestion.POST("", AuthMiddleware(), handler.CreateContestQuestion)
			contestQuestion.GET(":contestId/questions", AuthMiddleware(), handler.GetContestQuestions)
			contestQuestion.GET("/question/:questionId", AuthMiddleware(), handler.GetContestQuestionDetail)
			contestQuestion.PUT("/question/:questionId/status", AuthMiddleware(), handler.UpdateQuestionStatus)
			contestQuestion.POST("/question/:questionId/reply", AuthMiddleware(), handler.SendQuestionReply)
			contestQuestion.DELETE("/question/:questionId", AuthMiddleware(), handler.DeleteContestQuestion)
		}

		// 代码查重功能 - 管理员
		RegisterPlagiarismRoutes(api, cfg, db)
	}

	router.GET("/health", handler.HealthCheck)
}


