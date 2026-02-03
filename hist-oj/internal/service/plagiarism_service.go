package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/model"
)

// PlagiarismService 查重服务
type PlagiarismService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPlagiarismService 创建查重服务
func NewPlagiarismService(db *gorm.DB) *PlagiarismService {
	return &PlagiarismService{
		db:     db,
		logger: getLogger(),
	}
}

// SaveConfig 保存查重配置
func (s *PlagiarismService) SaveConfig(cid uint64, configs []model.PlagiarismCheckConfig, createdBy string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除该比赛原有的配置
		if err := tx.Where("cid = ?", cid).Delete(&model.PlagiarismCheckConfig{}).Error; err != nil {
			s.logger.Error("删除原查重配置失败", zap.Uint64("cid", cid), zap.Error(err))
			return fmt.Errorf("删除原查重配置失败: %w", err)
		}

		// 批量创建新配置
		for i := range configs {
			configs[i].CID = cid
			configs[i].CreatedBy = createdBy

			// 根据 CPID 查询 PID（题目ID）
			var contestProblem model.ContestProblem
			if err := tx.Where("id = ?", configs[i].CPID).First(&contestProblem).Error; err != nil {
				s.logger.Error("查询比赛题目失败", zap.Uint64("cpid", configs[i].CPID), zap.Error(err))
				return fmt.Errorf("查询比赛题目失败: cpid=%d, error=%w", configs[i].CPID, err)
			}
			configs[i].PID = contestProblem.PID
		}

		if len(configs) > 0 {
			if err := tx.Create(&configs).Error; err != nil {
				s.logger.Error("创建查重配置失败", zap.Uint64("cid", cid), zap.Error(err))
				return fmt.Errorf("创建查重配置失败: %w", err)
			}
		}

		return nil
	})
}

// GetConfig 获取查重配置
func (s *PlagiarismService) GetConfig(cid uint64) ([]model.PlagiarismCheckConfig, error) {
	var configs []model.PlagiarismCheckConfig
	err := s.db.Where("cid = ?", cid).Find(&configs).Error
	if err != nil {
		s.logger.Error("查询查重配置失败", zap.Uint64("cid", cid), zap.Error(err))
		return nil, fmt.Errorf("查询查重配置失败: %w", err)
	}
	return configs, nil
}

// GetContestProblems 获取比赛题目列表
func (s *PlagiarismService) GetContestProblems(cid uint64) ([]model.ContestProblem, error) {
	var problems []model.ContestProblem
	err := s.db.Where("cid = ?", cid).Find(&problems).Error
	if err != nil {
		s.logger.Error("查询比赛题目失败", zap.Uint64("cid", cid), zap.Error(err))
		return nil, fmt.Errorf("查询比赛题目失败: %w", err)
	}
	return problems, nil
}

// CreateCheck 创建查重任务
func (s *PlagiarismService) CreateCheck(cid uint64, startedBy string) (*model.PlagiarismCheck, error) {
	// 检查比赛状态
	var contest model.Contest
	if err := s.db.Where("id = ?", cid).First(&contest).Error; err != nil {
		return nil, fmt.Errorf("查询比赛失败: %w", err)
	}

	if contest.Status != 1 {
		return nil, fmt.Errorf("比赛未结束，无法进行查重")
	}

	// 查询配置
	configs, err := s.GetConfig(cid)
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("请先设置查重配置")
	}

	// 删除该比赛的旧查重任务和结果（支持重新查重）
	// 注意：先删除结果，再删除任务，避免外键约束问题
	deleteErr := s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除查重结果
		if err := tx.Where("cid = ?", cid).Delete(&model.PlagiarismResult{}).Error; err != nil {
			s.logger.Error("删除查重结果失败", zap.Uint64("cid", cid), zap.Error(err))
			return fmt.Errorf("删除查重结果失败: %w", err)
		}
		// 再删除非 completed 状态的查重任务
		if err := tx.Where("cid = ? AND status IN ('pending', 'running', 'failed')", cid).Delete(&model.PlagiarismCheck{}).Error; err != nil {
			s.logger.Error("删除旧查重任务失败", zap.Uint64("cid", cid), zap.Error(err))
			return fmt.Errorf("删除旧查重任务失败: %w", err)
		}
		return nil
	})
	if deleteErr != nil {
		s.logger.Error("删除旧查重数据失败", zap.Uint64("cid", cid), zap.Error(deleteErr))
		return nil, fmt.Errorf("删除旧查重数据失败: %w", deleteErr)
	}
	s.logger.Info("已删除旧的查重任务", zap.Uint64("cid", cid))

	// 创建查重任务
	check := &model.PlagiarismCheck{
		CID:       cid,
		Status:    "pending",
		StartedBy: startedBy,
	}

	if err := s.db.Create(check).Error; err != nil {
		s.logger.Error("创建查重任务失败", zap.Uint64("cid", cid), zap.Error(err))
		return nil, fmt.Errorf("创建查重任务失败: %w", err)
	}

	s.logger.Info("创建查重任务", zap.Uint64("cid", cid), zap.Uint64("checkId", check.ID))

	return check, nil
}

// RunCheck 执行查重任务
func (s *PlagiarismService) RunCheck(checkID uint64) error {
	var check model.PlagiarismCheck
	if err := s.db.Where("id = ?", checkID).First(&check).Error; err != nil {
		return fmt.Errorf("查询查重任务失败: %w", err)
	}

	s.logger.Info("RunCheck: 开始执行查重任务", zap.Uint64("checkId", checkID), zap.Uint64("cid", check.CID))

	// 更新状态为运行中
	now := time.Now()
	check.Status = "running"
	check.StartedAt = &now
	s.db.Save(&check)

	s.logger.Info("RunCheck: 已更新状态为running，准备启动executeCheck", zap.Uint64("checkId", checkID))

	// 异步执行查重
	go s.executeCheck(&check)

	s.logger.Info("RunCheck: 已启动executeCheck goroutine", zap.Uint64("checkId", checkID))

	return nil
}

// executeCheck 执行查重逻辑
func (s *PlagiarismService) executeCheck(check *model.PlagiarismCheck) {
	s.logger.Info("executeCheck: 函数开始执行", zap.Uint64("checkId", check.ID), zap.Uint64("cid", check.CID))

	// 查询比赛信息（获取开始和结束时间）
	var contest model.Contest
	if err := s.db.Where("id = ?", check.CID).First(&contest).Error; err != nil {
		s.logger.Error("查询比赛失败", zap.Uint64("cid", check.CID), zap.Error(err))
		s.updateCheckError(check.ID, fmt.Sprintf("查询比赛失败: %v", err))
		return
	}

	s.logger.Info("executeCheck: 查询比赛信息成功", zap.Uint64("cid", check.CID), zap.Int("status", contest.Status))

	// 查询配置
	configs, err := s.GetConfig(check.CID)
	if err != nil {
		s.logger.Error("获取配置失败", zap.Uint64("cid", check.CID), zap.Error(err))
		s.updateCheckError(check.ID, err.Error())
		return
	}

	s.logger.Info("executeCheck: 查询配置成功", zap.Uint64("cid", check.CID), zap.Int("configCount", len(configs)))

	// 创建阈值映射（支持 CPID 和 PID）
	thresholdMapByCPID := make(map[uint64]int)  // config.CPID -> threshold
	thresholdMapByPID := make(map[uint64]int)   // config.PID -> threshold
	for _, config := range configs {
		thresholdMapByCPID[config.CPID] = config.Threshold
		thresholdMapByPID[config.PID] = config.Threshold
	}

	// 预加载比赛的所有题目，建立 cpid -> contest_problem 映射
	var contestProblems []model.ContestProblem
	if err := s.db.Where("cid = ?", check.CID).Find(&contestProblems).Error; err != nil {
		s.logger.Error("查询比赛题目失败", zap.Uint64("cid", check.CID), zap.Error(err))
		s.updateCheckError(check.ID, fmt.Sprintf("查询比赛题目失败: %v", err))
		return
	}

	// 建立 cpid -> contest_problem 映射
	problemMap := make(map[uint64]model.ContestProblem)
	for _, problem := range contestProblems {
		problemMap[problem.ID] = problem
	}
	s.logger.Info("预加载比赛题目成功",
		zap.Uint64("cid", check.CID),
		zap.Int("problemCount", len(contestProblems)))

	// 查询比赛期间的提交记录（只查比赛开始到结束之间且AC的提交）
	s.logger.Info("准备查询比赛期间AC提交记录",
		zap.Uint64("cid", check.CID),
		zap.Time("startTime", contest.StartTime),
		zap.Time("endTime", contest.EndTime))

	var submissions []model.Judge
	if err := s.db.Where("cid = ? AND submit_time >= ? AND submit_time <= ? AND status = 0",
		check.CID, contest.StartTime, contest.EndTime).Find(&submissions).Error; err != nil {
		s.logger.Error("查询提交记录失败", zap.Uint64("cid", check.CID), zap.Error(err))
		s.updateCheckError(check.ID, err.Error())
		return
	}

	s.logger.Info("查询比赛期间AC提交记录完成",
		zap.Uint64("cid", check.CID),
		zap.Time("startTime", contest.StartTime),
		zap.Time("endTime", contest.EndTime),
		zap.Int("count", len(submissions)))

	s.logger.Info("查询到提交记录", zap.Int("count", len(submissions)), zap.Uint64("cid", check.CID))

	// 如果没有提交记录，直接完成
	if len(submissions) == 0 {
		s.logger.Info("没有AC提交记录，查重任务完成", zap.Uint64("cid", check.CID))
		completedAt := time.Now()
		check.Status = "completed"
		check.Progress = 100.0
		check.TotalSubmissions = 0
		check.TotalPairs = 0
		check.CheckedPairs = 0
		check.CompletedAt = &completedAt
		s.db.Save(check)
		return
	}

	// 按题目和语言分组
	grouped := make(map[string][]model.Judge) // key: "cpid:language"
	for _, sub := range submissions {
		key := fmt.Sprintf("%d:%s", sub.CPID, sub.Language)
		grouped[key] = append(grouped[key], sub)
	}

	s.logger.Info("分组完成", zap.Int("groups", len(grouped)))

	// 计算总对比对数
	totalPairs := 0
	for _, subs := range grouped {
		n := len(subs)
		totalPairs += n * (n - 1) / 2
	}

	s.logger.Info("计算总对比对数", zap.Int("totalPairs", totalPairs))

	// 如果没有对比对数（所有题目都只有1人AC），直接完成
	if totalPairs == 0 {
		s.logger.Info("没有需要对比的代码对，查重任务完成", zap.Uint64("cid", check.CID))
		completedAt := time.Now()
		check.Status = "completed"
		check.Progress = 100.0
		check.TotalSubmissions = len(submissions)
		check.TotalPairs = 0
		check.CheckedPairs = 0
		check.CompletedAt = &completedAt
		s.db.Save(check)
		return
	}

	check.TotalPairs = totalPairs
	check.TotalSubmissions = len(submissions)
	s.db.Save(check)

	s.logger.Info("executeCheck: 准备预加载数据", zap.Uint64("checkId", check.ID), zap.Int("totalPairs", totalPairs))

	// 创建临时目录
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("plagiarism_%d", check.ID))
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// 预加载 contest_record 数据，避免在对比时频繁查询数据库
	s.logger.Info("executeCheck: 开始预加载contest_record", zap.Uint64("cid", check.CID))
	var contestRecords []model.ContestRecord
	if err := s.db.Where("cid = ?", check.CID).Find(&contestRecords).Error; err != nil {
		s.logger.Error("查询比赛记录失败", zap.Uint64("cid", check.CID), zap.Error(err))
		s.updateCheckError(check.ID, fmt.Sprintf("查询比赛记录失败: %v", err))
		return
	}

	// 建立 (uid, cpid) -> record 的映射
	recordMap := make(map[string]model.ContestRecord)
	for _, record := range contestRecords {
		key := fmt.Sprintf("%s:%d", record.UID, record.CPID)
		recordMap[key] = record
	}
	s.logger.Info("预加载比赛记录成功", zap.Uint64("cid", check.CID), zap.Int("recordCount", len(contestRecords)))

	// 执行查重
	var wg sync.WaitGroup
	resultsChan := make(chan model.PlagiarismResult, 100)
	progressChan := make(chan bool, 1000) // 进度更新通道（批量处理）
	progressDoneChan := make(chan bool)    // 进度更新完成信号

	// 启动结果收集协程
	go s.collectResults(check.ID, resultsChan, &wg)
	// 启动进度更新协程（批量更新数据库）
	go s.batchUpdateProgress(check.ID, progressChan, totalPairs, progressDoneChan)

	// 限制并发数，避免同时启动过多 goroutine
	// 注意：此值需要与数据库连接池大小匹配（database.go中SetMaxOpenConns=300）
	maxConcurrency := 100 // 提高到100个并发对比，平衡性能和数据库连接占用
	semaphore := make(chan struct{}, maxConcurrency)

	s.logger.Info("executeCheck: 准备启动对比任务", zap.Uint64("checkId", check.ID), zap.Int("groupCount", len(grouped)), zap.Int("maxConcurrency", maxConcurrency))

	// 添加 defer recover，捕获 panic
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("executeCheck 发生 panic", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
			s.updateCheckError(check.ID, fmt.Sprintf("查重任务异常: %v", r))
		}
	}()

	// 对每组进行两两对比
	for key, subs := range grouped {
		if len(subs) < 2 {
			continue
		}

		parts := strings.Split(key, ":")
		language := parts[1]

		// 使用 PID 查找阈值（因为 judge.CPID 可能为 0）
		threshold := 50 // 默认阈值
		if len(subs) > 0 {
			pid := subs[0].PID
			if t, ok := thresholdMapByPID[pid]; ok {
				threshold = t
			}
		}

		// 从预加载的映射中获取题目信息
		displayID := ""
		displayTitle := ""
		contestedProblemID := uint64(0) // contest_problem 表的主键 ID

		// 使用提交的 PID 从 contest_problem 中查找题目
		if len(subs) > 0 {
			pid := subs[0].PID

			// 从预加载的题目中查找匹配 PID 的题目
			for _, problem := range contestProblems {
				if problem.PID == pid {
					displayID = problem.DisplayID
					displayTitle = problem.DisplayTitle
					contestedProblemID = problem.ID // 保存 contest_problem.ID
					break
				}
			}

			// 如果还是找不到，使用 judge.DisplayPID 作为备用
			if displayID == "" {
				displayID = subs[0].DisplayPID
				displayTitle = displayID
			}
		}

		// 两两对比
		comparisonCount := 0
		skippedCount := 0 // 跳过的同一用户对比数量
		for i := 0; i < len(subs); i++ {
			for j := i + 1; j < len(subs); j++ {
				// 跳过同一用户的提交（不自比较）
				if subs[i].UID == subs[j].UID {
					skippedCount++
					continue
				}
				comparisonCount++
				wg.Add(1)

				// 启动 goroutine（在 goroutine 内部获取信号量，避免主线程阻塞）
				go s.compareSubmissions(subs[i], subs[j], check.ID, check.CID, contestedProblemID,
					displayID, displayTitle, language, threshold, resultsChan, progressChan, tmpDir, &wg, semaphore, recordMap)
			}
		}
		s.logger.Info("启动对比任务", zap.String("key", key), zap.Int("subs", len(subs)),
			zap.Int("comparisons", comparisonCount), zap.Int("skippedSameUser", skippedCount))
	}

	// 等待所有对比完成
	s.logger.Info("executeCheck: 等待所有对比任务完成", zap.Uint64("checkId", check.ID))
	wg.Wait()
	s.logger.Info("executeCheck: 所有对比任务已完成", zap.Uint64("checkId", check.ID))

	// 关闭进度通道，触发批量更新协程进行最后的更新
	close(progressChan)

	// 等待进度更新协程完成最后的更新
	<-progressDoneChan
	s.logger.Info("executeCheck: 进度更新协程已完成", zap.Uint64("checkId", check.ID))

	// 关闭结果通道
	close(resultsChan)
	s.logger.Info("executeCheck: 已关闭resultsChan", zap.Uint64("checkId", check.ID))

	// 更新任务状态
	completedAt := time.Now()
	check.Status = "completed"
	check.Progress = 100.0
	check.CompletedAt = &completedAt

	// 从数据库重新读取最新的 checked_pairs
	var updatedCheck model.PlagiarismCheck
	if err := s.db.Where("id = ?", check.ID).First(&updatedCheck).Error; err == nil {
		check.CheckedPairs = updatedCheck.CheckedPairs
		s.logger.Info("executeCheck: 最终状态", zap.Uint64("checkId", check.ID),
			zap.Int("checkedPairs", check.CheckedPairs),
			zap.Int("totalPairs", check.TotalPairs),
			zap.Int("totalSubmissions", check.TotalSubmissions))
	}

	s.db.Save(check)

	s.logger.Info("查重任务完成", zap.Uint64("checkId", check.ID), zap.Float64("progress", check.Progress))
}

// compareSubmissions 对比两份代码
func (s *PlagiarismService) compareSubmissions(sub1, sub2 model.Judge, checkID, cid, cpid uint64,
	displayID, displayTitle, language string, threshold int, resultsChan chan<- model.PlagiarismResult, progressChan chan bool, tmpDir string, wg *sync.WaitGroup, semaphore chan struct{}, recordMap map[string]model.ContestRecord) {

	// 获取信号量（在 goroutine 启动时立即获取，避免主线程阻塞）
	semaphore <- struct{}{}

	defer func() {
		// 释放信号量
		<-semaphore
		// 通过通道通知进度更新（批量处理，减少数据库压力）
		progressChan <- true
		// 标记任务完成
		wg.Done()
	}()

	// 从预加载的映射中获取 contest_record.submit_id
	key1 := fmt.Sprintf("%s:%d", sub1.UID, cpid)
	key2 := fmt.Sprintf("%s:%d", sub2.UID, cpid)
	recordID1 := uint64(0)
	recordID2 := uint64(0)
	if record, ok := recordMap[key1]; ok {
		recordID1 = record.SubmitID
	}
	if record, ok := recordMap[key2]; ok {
		recordID2 = record.SubmitID
	}

	sim1to2, sim2to1, err := s.runSimCheck(sub1.Code, sub2.Code, language, tmpDir)
	if err != nil {
		s.logger.Error("查重失败",
			zap.Uint64("submit1", sub1.SubmitID),
			zap.Uint64("submit2", sub2.SubmitID),
			zap.String("language", language),
			zap.Int("code1Length", len(sub1.Code)),
			zap.Int("code2Length", len(sub2.Code)),
			zap.Error(err))
		return
	}

	maxSim := sim1to2
	if sim2to1 > maxSim {
		maxSim = sim2to1
	}

	result := model.PlagiarismResult{
		CheckID:           checkID,
		CID:               cid,
		CPID:              cpid,
		PID:               sub1.PID,
		DisplayID:         displayID,
		ProblemTitle:      displayTitle,
		SubmitID1:         sub1.SubmitID,
		SubmitID2:         sub2.SubmitID,
		ContestRecordID1:  recordID1,
		ContestRecordID2:  recordID2,
		UID1:              sub1.UID,
		UID2:              sub2.UID,
		Username1:         sub1.Username,
		Username2:         sub2.Username,
		Language:          language,
		Similarity1to2:    sim1to2,
		Similarity2to1:    sim2to1,
		MaxSimilarity:     maxSim,
		IsOverThreshold:   maxSim >= threshold,
	}

	resultsChan <- result
}

// runSimCheck 调用 sim 工具进行查重（带超时控制）
func (s *PlagiarismService) runSimCheck(code1, code2, language, tmpDir string) (sim1to2, sim2to1 int, err error) {
	// 写入临时文件
	file1 := filepath.Join(tmpDir, fmt.Sprintf("%d_1.tmp", time.Now().UnixNano()))
	file2 := filepath.Join(tmpDir, fmt.Sprintf("%d_2.tmp", time.Now().UnixNano()))

	if err := os.WriteFile(file1, []byte(code1), 0644); err != nil {
		return 0, 0, err
	}
	defer os.Remove(file1)

	if err := os.WriteFile(file2, []byte(code2), 0644); err != nil {
		return 0, 0, err
	}
	defer os.Remove(file2)

	// 选择 sim 工具
	languageLower := strings.ToLower(language)
	var simCmd string
	// 使用模糊匹配来支持各种语言名称格式
	// C/C++ 包括：c, c++, cpp, gcc, g++ 等
	if strings.Contains(languageLower, "c++") || strings.Contains(languageLower, "cpp") ||
	   strings.Contains(languageLower, "gcc") || strings.Contains(languageLower, "g++") ||
	   languageLower == "c" || strings.HasPrefix(languageLower, "c ") {
		simCmd = "sim_c"
	} else if strings.Contains(languageLower, "java") {
		simCmd = "sim_java"
	} else if strings.Contains(languageLower, "pascal") || strings.Contains(languageLower, "pas") {
		simCmd = "sim_pasc"
	} else {
		// 不支持的语言（如 Python, Go, Rust 等），返回相似度 0
		// 这些语言没有对应的 sim 工具，但不应该报错
		s.logger.Debug("跳过不支持的语言（无sim工具）", zap.String("language", language))
		return 0, 0, nil
	}

	// 执行 sim 命令（使用完整路径）
	// -p: 输出相似度百分比
	// -a: 比较所有文件对（获得双向百分比）
	// -t 1: 阈值设置为1（显示所有相似度）
	simPath := fmt.Sprintf("/usr/local/bin/%s", simCmd)
	cmd := exec.Command(simPath, "-p", "-a", "-t", "1", file1, file2)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 添加超时控制：每个 sim 比较最多5秒
	timeout := 5 * time.Second
	timer := time.AfterFunc(timeout, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("sim执行失败: %w, stderr: %s, path: %s", err, stderr.String(), simPath)
	}

	// 解析输出
	output := stdout.String()
	sim1to2, sim2to1, err = parseSimOutput(output)
	if err != nil {
		return 0, 0, fmt.Errorf("解析sim输出失败: %w, output: %s", err, output)
	}

	return sim1to2, sim2to1, nil
}

// parseSimOutput 解析 sim 输出
func parseSimOutput(output string) (sim1to2, sim2to1 int, err error) {
	// 示例输出（使用 -p -a 参数）:
	// File a.cpp: 374 tokens, 82 lines
	// File b.cpp: 388 tokens, 81 lines
	// Total input: 2 files (2 new, 0 old), 762 tokens
	// a.cpp consists for 89 % of b.cpp material
	// b.cpp consists for 86 % of a.cpp material
	//
	// 注意：当两个代码差异较大时，sim 不会输出百分比信息
	// 此时应该将相似度设置为 0

	// 正则匹配相似度
	re := regexp.MustCompile(`(\d+)\s*%`)

	matches := re.FindAllStringSubmatch(output, -1)
	if len(matches) >= 2 {
		// 有两个百分比，双向相似度
		sim1to2, _ = strconv.Atoi(matches[0][1])
		sim2to1, _ = strconv.Atoi(matches[1][1])
		return sim1to2, sim2to1, nil
	} else if len(matches) == 1 {
		// 只有一个百分比，可能是单向相似度（两个文件差异较大）
		sim1to2, _ = strconv.Atoi(matches[0][1])
		sim2to1 = sim1to2 // 假设双向相同
		return sim1to2, sim2to1, nil
	}

	// 没有百分比输出，说明两个代码差异很大，相似度为 0
	// 这不是错误，而是正常情况
	return 0, 0, nil
}

// collectResults 收集查重结果
func (s *PlagiarismService) collectResults(checkID uint64, resultsChan <-chan model.PlagiarismResult, wg *sync.WaitGroup) {
	resultCount := 0
	// 批量插入，减少数据库操作
	batchSize := 50 // 从100降低到50，更频繁保存
	batch := make([]model.PlagiarismResult, 0, batchSize)

	s.logger.Info("collectResults协程启动", zap.Uint64("checkId", checkID))

	for result := range resultsChan {
		resultCount++
		batch = append(batch, result)

		// 达到批次大小，批量插入
		if len(batch) >= batchSize {
			if err := s.db.Create(&batch).Error; err != nil {
				s.logger.Error("批量保存查重结果失败", zap.Uint64("checkId", checkID), zap.Error(err))
			} else {
				// 每50条记录打印一次日志（从1000降低到50）
				if resultCount % 50 == 0 {
					s.logger.Info("已保存查重结果", zap.Uint64("checkId", checkID), zap.Int("resultCount", resultCount))
				}
			}
			batch = batch[:0] // 清空batch
		}
	}

	// 保存剩余的结果
	if len(batch) > 0 {
		if err := s.db.Create(&batch).Error; err != nil {
			s.logger.Error("保存剩余查重结果失败", zap.Uint64("checkId", checkID), zap.Error(err))
		}
	}

	s.logger.Info("所有查重结果已保存", zap.Uint64("checkId", checkID), zap.Int("totalResults", resultCount))
}

// batchUpdateProgress 批量更新查重进度
func (s *PlagiarismService) batchUpdateProgress(checkID uint64, progressChan <-chan bool, totalPairs int, doneChan chan<- bool) {
	batchSize := 0
	totalProcessed := 0
	ticker := time.NewTicker(200 * time.Millisecond) // 每200ms更新一次数据库（从500ms降低）
	defer ticker.Stop()

	s.logger.Info("启动进度更新协程", zap.Uint64("checkId", checkID), zap.Int("totalPairs", totalPairs))

	for {
		select {
		case _, ok := <-progressChan:
			if !ok {
				// 通道关闭，更新剩余的进度
				if batchSize > 0 {
					newCheckedPairs := totalProcessed + batchSize
					progress := float64(newCheckedPairs) * 100.0 / float64(totalPairs)
					if progress > 100 {
						progress = 100
					}
					result := s.db.Exec("UPDATE plagiarism_check SET checked_pairs = ?, progress = ? WHERE id = ?", newCheckedPairs, progress, checkID)
					if result.Error != nil {
						s.logger.Error("更新进度失败", zap.Uint64("checkId", checkID), zap.Error(result.Error))
					} else {
						s.logger.Info("最终进度更新", zap.Uint64("checkId", checkID), zap.Int("totalProcessed", newCheckedPairs), zap.Float64("progress", progress))
					}
				}
				// 通知主线程进度更新已完成
				doneChan <- true
				return
			}
			batchSize++

			// 达到批次大小，更新数据库（从100降低到20，更频繁更新）
			if batchSize >= 20 {
				newCheckedPairs := totalProcessed + batchSize
				progress := float64(newCheckedPairs) * 100.0 / float64(totalPairs)
				if progress > 100 {
					progress = 100
				}
				result := s.db.Exec("UPDATE plagiarism_check SET checked_pairs = ?, progress = ? WHERE id = ?", newCheckedPairs, progress, checkID)
				if result.Error != nil {
					s.logger.Error("批量更新进度失败", zap.Uint64("checkId", checkID), zap.Error(result.Error))
				}
				totalProcessed = newCheckedPairs
				batchSize = 0
			}
		case <-ticker.C:
			// 定时更新（即使未达到批次大小）
			if batchSize > 0 {
				newCheckedPairs := totalProcessed + batchSize
				progress := float64(newCheckedPairs) * 100.0 / float64(totalPairs)
				if progress > 100 {
					progress = 100
				}
				result := s.db.Exec("UPDATE plagiarism_check SET checked_pairs = ?, progress = ? WHERE id = ?", newCheckedPairs, progress, checkID)
				if result.Error != nil {
					s.logger.Error("定时更新进度失败", zap.Uint64("checkId", checkID), zap.Error(result.Error))
				}
				totalProcessed = newCheckedPairs
				batchSize = 0
			}
		}
	}
}

// updateCheckError 更新查重任务错误状态
func (s *PlagiarismService) updateCheckError(checkID uint64, errMsg string) {
	s.db.Model(&model.PlagiarismCheck{}).Where("id = ?", checkID).Updates(map[string]interface{}{
		"status":       "failed",
		"error_message": errMsg,
		"completed_at": time.Now(),
	})
}

// GetProgress 获取查重进度
func (s *PlagiarismService) GetProgress(checkID uint64) (*model.PlagiarismCheck, error) {
	var check model.PlagiarismCheck
	err := s.db.Where("id = ?", checkID).First(&check).Error
	if err != nil {
		return nil, fmt.Errorf("查询查重任务失败: %w", err)
	}
	return &check, nil
}

// GetResults 获取查重结果（只返回超过阈值的结果，可按题号筛选）
func (s *PlagiarismService) GetResults(checkID uint64, displayId string) ([]model.PlagiarismResult, int64, error) {
	startTime := time.Now()

	// 构建查询条件：只返回超过阈值的结果
	query := s.db.Where("check_id = ? AND is_over_threshold = ?", checkID, true)

	// 如果指定了题号，添加筛选条件
	if displayId != "" {
		query = query.Where("display_id = ?", displayId)
		s.logger.Info("GetResults: 按题号筛选", zap.Uint64("checkId", checkID), zap.String("displayId", displayId))
	}

	// 先统计总数
	var totalCount int64
	query.Model(&model.PlagiarismResult{}).Count(&totalCount)

	// 查询结果
	var results []model.PlagiarismResult
	if err := query.Order("max_similarity DESC").Find(&results).Error; err != nil {
		s.logger.Error("查询查重结果失败", zap.Uint64("checkId", checkID), zap.Error(err))
		return nil, 0, fmt.Errorf("查询查重结果失败: %w", err)
	}

	elapsed := time.Since(startTime).Milliseconds()
	s.logger.Info("GetResults: 查询完成", zap.Uint64("checkId", checkID),
		zap.String("displayId", displayId), zap.Int("count", len(results)),
		zap.Int64("totalCount", totalCount), zap.Int64("elapsedMs", elapsed))

	return results, totalCount, nil
}

// GetLatestCheck 获取最新的查重任务
func (s *PlagiarismService) GetLatestCheck(cid uint64) (*model.PlagiarismCheck, error) {
	var check model.PlagiarismCheck
	err := s.db.Where("cid = ?", cid).Order("id DESC").First(&check).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询查重任务失败: %w", err)
	}
	return &check, nil
}

// GetSubmission 获取提交详情
func (s *PlagiarismService) GetSubmission(submitID uint64) (*model.Judge, error) {
	var submission model.Judge
	err := s.db.Where("submit_id = ?", submitID).First(&submission).Error
	if err != nil {
		return nil, fmt.Errorf("查询提交详情失败: %w", err)
	}
	return &submission, nil
}

// getContestRecordSubmitID 查询 contest_record 表的 submit_id

// getLogger 获取日志记录器
func getLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// csvFieldEscape CSV 字段转义，处理逗号、引号、换行符
func csvFieldEscape(field string) string {
	// 如果字段包含逗号、引号或换行符，需要用引号包裹，并将内部引号转义
	if strings.ContainsAny(field, "\",\"\n") {
		// 将引号转义为两个引号
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return field
}

// ExportResults 导出查重结果为 Excel
func (s *PlagiarismService) ExportResults(checkID uint64) ([]byte, error) {
	// 查询所有结果（不限制数量）
	var results []model.PlagiarismResult
	if err := s.db.Where("check_id = ?", checkID).
		Order("max_similarity DESC").
		Find(&results).Error; err != nil {
		s.logger.Error("查询查重结果失败", zap.Uint64("checkId", checkID), zap.Error(err))
		return nil, fmt.Errorf("查询查重结果失败: %w", err)
	}

	s.logger.Info("ExportResults: 开始生成Excel", zap.Uint64("checkId", checkID), zap.Int("totalRows", len(results)))

	// 使用 excelize 生成 Excel（在内存中）
	// 注意：这里我们手动实现简单的CSV格式，避免 excelize 在某些环境下的问题
	var buf bytes.Buffer

	// 写入 CSV 头（带 BOM 以支持 Excel 打开中文）
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	buf.WriteString("题目编号,题目标题,用户1,用户2,语言,相似度1to2(%),相似度2to1(%),最大相似度(%),超过阈值\n")

	// 写入数据
	for _, r := range results {
		overThreshold := "否"
		if r.IsOverThreshold {
			overThreshold = "是"
		}
		// 对所有文本字段进行 CSV 转义，防止逗号等特殊字符破坏格式
		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%d,%s\n",
			csvFieldEscape(r.DisplayID),
			csvFieldEscape(r.ProblemTitle),
			csvFieldEscape(r.Username1),
			csvFieldEscape(r.Username2),
			csvFieldEscape(r.Language),
			r.Similarity1to2,
			r.Similarity2to1,
			r.MaxSimilarity,
			csvFieldEscape(overThreshold),
		))
	}

	s.logger.Info("ExportResults: Excel生成完成", zap.Int("fileSize", buf.Len()))
	return buf.Bytes(), nil
}
