package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
)

var errNativeJudgeUnavailable = errors.New("native judge unavailable")

//go:embed assets/testlib.h
var embeddedTestlibHeader string

const (
	sandboxDefaultURL     = "http://127.0.0.1:5050"
	sandboxStdIOMaxBytes  = 16 * 1024 * 1024
	sandboxOutputMaxBytes = 32 * 1024 * 1024
	sandboxProcLimit      = 64
	sandboxStackLimitMB   = 128

	terminalSPJCodePC    = 99
	terminalSPJCodeAC    = 100
	terminalSPJCodePE    = 101
	terminalSPJCodeWA    = 102
	terminalSPJCodeError = 103
)

var sandboxDefaultEnv = []string{
	"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	"LANG=en_US.UTF-8",
	"LANGUAGE=en_US:en",
	"LC_ALL=en_US.UTF-8",
}

var sandboxPythonEnv = []string{
	"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	"LANG=en_US.UTF-8",
	"LANGUAGE=en_US:en",
	"LC_ALL=en_US.UTF-8",
	"PYTHONIOENCODING=utf-8",
}

type terminalJudgeCase struct {
	ProblemID   int64
	CaseID      int64
	Seq         int
	InputRef    string
	ExpectedRef string
	Score       *int
	GroupNum    *int
	Mode        string
}

type terminalJudgeRunResult struct {
	FinalStatus       int
	FirstFailedSeq    int
	FirstFailedStatus int
	ErrorMessage      string
	CaseDetails       []*JudgeCaseDetail
	CaseRecords       []*model.SubmissionHistoryCase
	MaxTime           int
	MaxMemory         int
}

type problemCaseRow struct {
	ID       int64  `gorm:"column:id"`
	Input    string `gorm:"column:input"`
	Output   string `gorm:"column:output"`
	Score    *int   `gorm:"column:score"`
	GroupNum *int   `gorm:"column:group_num"`
	Mode     string `gorm:"column:mode"`
}

type terminalLanguageProfile struct {
	SourceName      string
	ExecName        string
	CompileArgs     []string
	RunArgs         []string
	CompileEnv      []string
	RunEnv          []string
	NeedCompile     bool
	NeedCachedExec  bool
	NeedSourceAtRun bool
	InjectTestlib   bool
}

type terminalSPJProgram struct {
	Code      string
	Language  string
	ExtraFile map[string]string
}

type terminalSPJProfile struct {
	SourceName  string
	ExecName    string
	CompileArgs []string
	RunArgs     []string
	CompileEnv  []string
	RunEnv      []string
}

type terminalSPJRunResult struct {
	Code       int
	ErrMessage string
	Percentage float64
}

type terminalSandboxJudgeEngine struct {
	runner      *sandboxRunner
	profile     *terminalLanguageProfile
	sourceCode  string
	execFileID  string
	cachedFiles []string
}

type terminalSPJEngine struct {
	runner      *sandboxRunner
	profile     *terminalSPJProfile
	execFileID  string
	cachedFiles []string
}

type sandboxRunner struct {
	baseURL string
	client  *http.Client
}

type sandboxRunRequest struct {
	Cmd []map[string]interface{} `json:"cmd"`
}

type sandboxCmdResult struct {
	Status     string            `json:"status"`
	ExitStatus int               `json:"exitStatus"`
	Time       int64             `json:"time"`
	Memory     int64             `json:"memory"`
	RunTime    int64             `json:"runTime"`
	Files      map[string]string `json:"files"`
	FileIDs    map[string]string `json:"fileIds"`
}

// IsNativeJudgeUnavailable 判断是否是本地判题环境缺失错误（例如沙箱服务不可达）
func IsNativeJudgeUnavailable(err error) bool {
	return errors.Is(err, errNativeJudgeUnavailable)
}

// GenerateLocalSubmitID 生成判题终端本地提交ID
func GenerateLocalSubmitID() string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("LOCAL-%d", time.Now().UnixMilli())
	}
	return fmt.Sprintf("LOCAL-%d-%s", time.Now().UnixMilli(), hex.EncodeToString(buf))
}

// BuildTerminalJudgeCases 构建判题终端独立判题用的测试点集合。
// 优先从 problem_case + 测试文件目录读取，失败时回落到题目 examples。
func (s *JudgeService) BuildTerminalJudgeCases(problemID int64, fallbackSamples []SampleResult) ([]*terminalJudgeCase, string, error) {
	var rows []*problemCaseRow
	baseQuery := func() *gorm.DB {
		return s.db.Table("problem_case").
			Where("pid = ? AND status = 0", problemID).
			Order("id ASC")
	}

	err := baseQuery().
		Select("id, input, output, score, group_num, mode").
		Scan(&rows).Error
	if isUnknownColumnError(err, "mode") {
		// 兼容旧版数据库：problem_case 还没有 mode 字段
		err = baseQuery().
			Select("id, input, output, score, group_num").
			Scan(&rows).Error
	}
	if err != nil {
		fallback := buildCasesFromSamples(problemID, fallbackSamples)
		if len(fallback) > 0 {
			return fallback, "examples", fmt.Errorf("读取 problem_case 失败，已回退到样例: %w", err)
		}
		return nil, "", fmt.Errorf("读取 problem_case 失败: %w", err)
	}

	if len(rows) == 0 {
		fallback := buildCasesFromSamples(problemID, fallbackSamples)
		if len(fallback) > 0 {
			return fallback, "examples", fmt.Errorf("problem_case 为空，已回退到样例")
		}
		return nil, "", fmt.Errorf("题目无可用测试点")
	}

	cases := make([]*terminalJudgeCase, 0, len(rows))
	for idx, row := range rows {
		inErr := validateProblemCaseRef(problemID, row.Input)
		if inErr != nil {
			fallback := buildCasesFromSamples(problemID, fallbackSamples)
			if len(fallback) > 0 {
				return fallback, "examples", fmt.Errorf("读取测试点输入失败，已回退到样例: %w", inErr)
			}
			return nil, "", inErr
		}

		outErr := validateProblemCaseRef(problemID, row.Output)
		if outErr != nil {
			fallback := buildCasesFromSamples(problemID, fallbackSamples)
			if len(fallback) > 0 {
				return fallback, "examples", fmt.Errorf("读取测试点输出失败，已回退到样例: %w", outErr)
			}
			return nil, "", outErr
		}

		seq := idx + 1
		mode := strings.TrimSpace(row.Mode)
		if mode == "" {
			mode = "default"
		}
		cases = append(cases, &terminalJudgeCase{
			ProblemID:   problemID,
			CaseID:      row.ID,
			Seq:         seq,
			InputRef:    row.Input,
			ExpectedRef: row.Output,
			Score:       row.Score,
			GroupNum:    row.GroupNum,
			Mode:        mode,
		})
	}

	return cases, "problem_case", nil
}

// TestLocalSamplesNative 判题终端本地样例自测（走 go-judge 沙箱，不依赖 Java test-judge）。
func (s *JudgeService) TestLocalSamplesNative(
	language, code string,
	samples []SampleResult,
	timeLimitMS, memoryLimitMB int64,
	problemID int64,
	problemJudgeMode string,
) ([]SampleResult, error) {
	engine, err := newTerminalSandboxJudgeEngine(language, code)
	if err != nil {
		return nil, err
	}
	defer engine.Close(s.logger)

	spjEngine, err := s.newTerminalSPJEngineIfNeeded(problemID, problemJudgeMode)
	if err != nil {
		return nil, err
	}
	if spjEngine != nil {
		defer spjEngine.Close(s.logger)
	}

	results := make([]SampleResult, 0, len(samples))
	timeout := terminalCaseTimeout(timeLimitMS)
	effectiveMemoryMB := sandboxRunMemoryLimitMB(memoryLimitMB)

	for i, sample := range samples {
		result := sample
		runRes, runErr := engine.Run(sample.Input, timeout, effectiveMemoryMB)
		if runErr != nil {
			if IsNativeJudgeUnavailable(runErr) {
				return nil, runErr
			}
			result.Status = 4
			result.IsOK = false
			result.Stderr = trimToMax(runErr.Error(), 1024*1024)
			result.Output = "(无输出)"
			results = append(results, result)
			continue
		}

		rawStdout := runRes.Stdout
		result.Output = trimToMax(rawStdout, 20000)
		result.Stderr = trimToMax(runRes.Stderr, 1024*1024)
		result.DetailedStderr = ""

		sandboxStatus := mapSandboxStatus(runRes.Status, runRes.ExitStatus)
		switch sandboxStatus {
		case 0:
			judgeStatus, judgeStderr, judgeErr := evaluateTerminalOutput(sample.Input, sample.Expected, rawStdout, result.Stderr, spjEngine)
			if judgeErr != nil {
				result.Status = 4
				result.IsOK = false
				result.Stderr = trimToMax(judgeErr.Error(), 1024*1024)
			} else {
				result.Status = judgeStatus
				result.IsOK = (judgeStatus == 0)
				result.Stderr = trimToMax(judgeStderr, 1024*1024)
			}
		default:
			result.Status = sandboxStatus
			result.IsOK = false
			runtimeMsg := buildSandboxRuntimeMessage(runRes)
			if strings.TrimSpace(result.Stderr) == "" {
				result.Stderr = runtimeMsg
			} else if runtimeMsg != "" && !strings.Contains(result.Stderr, runtimeMsg) {
				result.DetailedStderr = runtimeMsg
			}
		}

		if result.Output == "" && result.Stderr != "" {
			result.Output = result.Stderr
		}
		if result.Output == "" {
			result.Output = "(无输出)"
		}

		s.logger.Info("本地样例测试完成",
			zap.Int("样例编号", i+1),
			zap.Int("status", result.Status),
			zap.Bool("is_ok", result.IsOK),
			zap.String("stderr", trimToMax(result.Stderr, 500)))

		results = append(results, result)
	}

	return results, nil
}

// RunTerminalJudge 执行判题终端独立判题（全测试点）。
func (s *JudgeService) RunTerminalJudge(
	language, code string,
	cases []*terminalJudgeCase,
	timeLimitMS, memoryLimitMB int64,
	problemJudgeMode string,
	problemJudgeCaseMode string,
	onProgress func(done, total, seq, status int),
) (*terminalJudgeRunResult, error) {
	if len(cases) == 0 {
		return nil, fmt.Errorf("无可评测测试点")
	}

	engine, err := newTerminalSandboxJudgeEngine(language, code)
	if err != nil {
		return nil, err
	}
	defer engine.Close(s.logger)

	problemID := cases[0].ProblemID
	spjEngine, err := s.newTerminalSPJEngineIfNeeded(problemID, problemJudgeMode)
	if err != nil {
		return nil, err
	}
	if spjEngine != nil {
		defer spjEngine.Close(s.logger)
	}

	result := &terminalJudgeRunResult{
		FinalStatus: 0,
		CaseDetails: make([]*JudgeCaseDetail, 0, len(cases)),
		CaseRecords: make([]*model.SubmissionHistoryCase, 0, len(cases)),
	}
	timeout := terminalCaseTimeout(timeLimitMS)
	effectiveMemoryMB := sandboxRunMemoryLimitMB(memoryLimitMB)

	for _, tc := range cases {
		if onProgress != nil {
			onProgress(len(result.CaseDetails), len(cases), tc.Seq, 7)
		}

		inputContent, inErr := resolveProblemCaseContent(tc.ProblemID, tc.InputRef)
		if inErr != nil {
			status := 4
			errMsg := trimToMax(inErr.Error(), 1024*1024)
			timeCopy := 0
			memoryCopy := 0
			statusCopy := status
			seqCopy := tc.Seq

			if result.FinalStatus == 0 {
				result.FinalStatus = status
				result.FirstFailedSeq = tc.Seq
				result.FirstFailedStatus = status
				result.ErrorMessage = terminalFinalErrorMessage(tc.Seq, status, errMsg)
			}

			result.CaseDetails = append(result.CaseDetails, &JudgeCaseDetail{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			result.CaseRecords = append(result.CaseRecords, &model.SubmissionHistoryCase{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			if onProgress != nil {
				onProgress(len(result.CaseDetails), len(cases), tc.Seq, status)
			}
			if shouldStopAfterCase(tc.Mode, problemJudgeCaseMode, status) {
				break
			}
			continue
		}

		expectedContent, outErr := resolveProblemCaseContent(tc.ProblemID, tc.ExpectedRef)
		if outErr != nil {
			status := 4
			errMsg := trimToMax(outErr.Error(), 1024*1024)
			timeCopy := 0
			memoryCopy := 0
			statusCopy := status
			seqCopy := tc.Seq

			if result.FinalStatus == 0 {
				result.FinalStatus = status
				result.FirstFailedSeq = tc.Seq
				result.FirstFailedStatus = status
				result.ErrorMessage = terminalFinalErrorMessage(tc.Seq, status, errMsg)
			}

			result.CaseDetails = append(result.CaseDetails, &JudgeCaseDetail{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			result.CaseRecords = append(result.CaseRecords, &model.SubmissionHistoryCase{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			if onProgress != nil {
				onProgress(len(result.CaseDetails), len(cases), tc.Seq, status)
			}
			if shouldStopAfterCase(tc.Mode, problemJudgeCaseMode, status) {
				break
			}
			continue
		}

		runRes, runErr := engine.Run(inputContent, timeout, effectiveMemoryMB)
		if runErr != nil {
			if IsNativeJudgeUnavailable(runErr) {
				return nil, runErr
			}

			status := 4
			errMsg := trimToMax(runErr.Error(), 1024*1024)
			timeCopy := 0
			memoryCopy := 0
			statusCopy := status
			seqCopy := tc.Seq

			if result.FinalStatus == 0 {
				result.FinalStatus = status
				result.FirstFailedSeq = tc.Seq
				result.FirstFailedStatus = status
				result.ErrorMessage = terminalFinalErrorMessage(tc.Seq, status, errMsg)
			}

			result.CaseDetails = append(result.CaseDetails, &JudgeCaseDetail{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			result.CaseRecords = append(result.CaseRecords, &model.SubmissionHistoryCase{
				CaseID:   tc.CaseID,
				Status:   &statusCopy,
				Time:     &timeCopy,
				Memory:   &memoryCopy,
				Score:    tc.Score,
				GroupNum: tc.GroupNum,
				Seq:      &seqCopy,
				Mode:     tc.Mode,
				Stderr:   errMsg,
			})
			if onProgress != nil {
				onProgress(len(result.CaseDetails), len(cases), tc.Seq, status)
			}
			if shouldStopAfterCase(tc.Mode, problemJudgeCaseMode, status) {
				break
			}
			continue
		}

		status := mapSandboxStatus(runRes.Status, runRes.ExitStatus)
		stderr := trimToMax(runRes.Stderr, 1024*1024)
		rawStdout := runRes.Stdout

		if status == 0 {
			judgeStatus, judgeStderr, judgeErr := evaluateTerminalOutput(inputContent, expectedContent, rawStdout, stderr, spjEngine)
			if judgeErr != nil {
				status = 4
				stderr = trimToMax(judgeErr.Error(), 1024*1024)
			} else {
				status = judgeStatus
				stderr = trimToMax(judgeStderr, 1024*1024)
			}
		} else if strings.TrimSpace(stderr) == "" {
			stderr = buildSandboxRuntimeMessage(runRes)
		}

		timeCopy := runRes.TimeMS
		memoryCopy := runRes.MemoryKB
		statusCopy := status
		seqCopy := tc.Seq

		if timeCopy > result.MaxTime {
			result.MaxTime = timeCopy
		}
		if memoryCopy > result.MaxMemory {
			result.MaxMemory = memoryCopy
		}

		if result.FinalStatus == 0 && status != 0 {
			result.FinalStatus = status
			result.FirstFailedSeq = tc.Seq
			result.FirstFailedStatus = status
			result.ErrorMessage = terminalFinalErrorMessage(tc.Seq, status, stderr)
		}

		result.CaseDetails = append(result.CaseDetails, &JudgeCaseDetail{
			CaseID:   tc.CaseID,
			Status:   &statusCopy,
			Time:     &timeCopy,
			Memory:   &memoryCopy,
			Score:    tc.Score,
			GroupNum: tc.GroupNum,
			Seq:      &seqCopy,
			Mode:     tc.Mode,
			Stderr:   stderr,
		})
		result.CaseRecords = append(result.CaseRecords, &model.SubmissionHistoryCase{
			CaseID:   tc.CaseID,
			Status:   &statusCopy,
			Time:     &timeCopy,
			Memory:   &memoryCopy,
			Score:    tc.Score,
			GroupNum: tc.GroupNum,
			Seq:      &seqCopy,
			Mode:     tc.Mode,
			Stderr:   stderr,
		})
		if onProgress != nil {
			onProgress(len(result.CaseDetails), len(cases), tc.Seq, status)
		}
		if shouldStopAfterCase(tc.Mode, problemJudgeCaseMode, status) {
			break
		}
	}

	return result, nil
}

// SaveLocalJudgeCaseDetails 保存判题终端本地判题测试点详情
func (s *JudgeService) SaveLocalJudgeCaseDetails(submitID string, records []*model.SubmissionHistoryCase) error {
	if len(records) == 0 {
		return nil
	}
	for _, item := range records {
		item.SubmitID = submitID
	}
	return s.db.Create(records).Error
}

type terminalRunResult struct {
	Status     string
	ExitStatus int
	Stdout     string
	Stderr     string
	TimeMS     int
	MemoryKB   int
}

func newTerminalSandboxJudgeEngine(language, code string) (*terminalSandboxJudgeEngine, error) {
	profile, err := getTerminalLanguageProfile(language)
	if err != nil {
		return nil, err
	}

	engine := &terminalSandboxJudgeEngine{
		runner:     newSandboxRunner(),
		profile:    profile,
		sourceCode: code,
	}

	if err := engine.prepare(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *terminalSandboxJudgeEngine) prepare() error {
	if !e.profile.NeedCompile {
		return nil
	}

	copyIn := map[string]interface{}{
		e.profile.SourceName: map[string]interface{}{"content": e.sourceCode},
	}
	if e.profile.InjectTestlib {
		copyIn["testlib.h"] = map[string]interface{}{"content": embeddedTestlibHeader}
	}

	cmd := map[string]interface{}{
		"args":        e.profile.CompileArgs,
		"env":         e.profile.CompileEnv,
		"files":       sandboxCompileFiles(),
		"cpuLimit":    int64(20) * 1000 * 1000 * 1000,
		"clockLimit":  int64(60) * 1000 * 1000 * 1000,
		"memoryLimit": int64(512) * 1024 * 1024,
		"procLimit":   sandboxProcLimit,
		"stackLimit":  int64(256) * 1024 * 1024,
		"copyIn":      copyIn,
		"copyOut":     []string{"stdout", "stderr"},
	}

	if e.profile.NeedCachedExec {
		cmd["copyOutCached"] = []string{e.profile.ExecName}
	}

	res, err := e.runner.runOne(context.Background(), cmd)
	if err != nil {
		return err
	}

	if res.Status != "Accepted" {
		return fmt.Errorf("编译失败: %s", buildCompileErrorMessage(res))
	}

	if e.profile.NeedCachedExec {
		fileID := ""
		if res.FileIDs != nil {
			fileID = res.FileIDs[e.profile.ExecName]
		}
		if fileID == "" {
			return fmt.Errorf("编译失败: 沙箱未返回可执行文件ID")
		}
		e.execFileID = fileID
		e.cachedFiles = append(e.cachedFiles, fileID)
	}

	return nil
}

func (e *terminalSandboxJudgeEngine) Run(input string, timeout time.Duration, memoryLimitMB int64) (*terminalRunResult, error) {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	copyIn := map[string]interface{}{}
	if e.profile.NeedCachedExec {
		copyIn[e.profile.ExecName] = map[string]interface{}{"fileId": e.execFileID}
	}
	if e.profile.NeedSourceAtRun {
		copyIn[e.profile.SourceName] = map[string]interface{}{"content": e.sourceCode}
	}

	normalizedInput := input
	if normalizedInput != "" && !strings.HasSuffix(normalizedInput, "\n") {
		normalizedInput += "\n"
	}

	cpuLimitNs := timeout.Nanoseconds()
	// Clock limit 适度高于 CPU limit，避免调度抖动导致误杀，
	// 同时避免出现 wall-clock 远高于题面时限的“慢判题”体验。
	clockLimitNs := cpuLimitNs + int64(200)*1000*1000
	minClockLimitNs := cpuLimitNs * 12 / 10 // 至少 1.2x
	maxClockLimitNs := cpuLimitNs * 2       // 最多 2x
	if clockLimitNs < minClockLimitNs {
		clockLimitNs = minClockLimitNs
	}
	if clockLimitNs > maxClockLimitNs {
		clockLimitNs = maxClockLimitNs
	}

	cmd := map[string]interface{}{
		"args":        e.profile.RunArgs,
		"env":         e.profile.RunEnv,
		"files":       sandboxRunFiles(normalizedInput),
		"cpuLimit":    cpuLimitNs,
		"clockLimit":  clockLimitNs,
		"memoryLimit": memoryLimitMB * 1024 * 1024,
		"procLimit":   sandboxProcLimit,
		"stackLimit":  int64(sandboxStackLimitMB) * 1024 * 1024,
		"copyIn":      copyIn,
		"copyOut":     []string{"stdout", "stderr"},
	}

	res, err := e.runner.runOne(context.Background(), cmd)
	if err != nil {
		return nil, err
	}

	r := &terminalRunResult{
		Status:     res.Status,
		ExitStatus: res.ExitStatus,
	}
	if res.Files != nil {
		r.Stdout = res.Files["stdout"]
		r.Stderr = res.Files["stderr"]
	}
	if res.Time > 0 {
		r.TimeMS = int(res.Time / (1000 * 1000))
	}
	if res.Memory > 0 {
		r.MemoryKB = int((res.Memory + 1023) / 1024)
	}
	return r, nil
}

func (e *terminalSandboxJudgeEngine) Close(logger *zap.Logger) {
	for _, fileID := range e.cachedFiles {
		if err := e.runner.deleteFile(fileID); err != nil && logger != nil {
			logger.Debug("删除沙箱缓存文件失败", zap.String("file_id", fileID), zap.Error(err))
		}
	}
}

func (s *JudgeService) newTerminalSPJEngineIfNeeded(problemID int64, judgeMode string) (*terminalSPJEngine, error) {
	if !isSPJMode(judgeMode) {
		return nil, nil
	}
	if problemID <= 0 {
		return nil, fmt.Errorf("SPJ 题目缺少题目ID，无法执行 SPJ 检查")
	}

	program, err := s.loadTerminalSPJProgram(problemID)
	if err != nil {
		return nil, err
	}
	return newTerminalSPJEngine(program)
}

func (s *JudgeService) loadTerminalSPJProgram(problemID int64) (*terminalSPJProgram, error) {
	type row struct {
		SPJCode        string `gorm:"column:spj_code"`
		SPJLanguage    string `gorm:"column:spj_language"`
		JudgeExtraFile string `gorm:"column:judge_extra_file"`
	}

	var data row
	query := s.db.Table("problem").Where("id = ?", problemID)

	err := query.Select("spj_code, spj_language, judge_extra_file").Scan(&data).Error
	if isUnknownColumnError(err, "judge_extra_file") {
		err = query.Select("spj_code, spj_language").Scan(&data).Error
	}
	if isUnknownColumnError(err, "spj_code") || isUnknownColumnError(err, "spj_language") {
		return nil, fmt.Errorf("当前数据库缺少 SPJ 字段，无法执行 SPJ 判题")
	}
	if err != nil {
		return nil, fmt.Errorf("读取 SPJ 配置失败: %w", err)
	}

	code := strings.TrimSpace(data.SPJCode)
	if code == "" {
		return nil, fmt.Errorf("SPJ 代码为空，无法执行 SPJ 判题")
	}
	lang := strings.TrimSpace(data.SPJLanguage)
	if lang == "" {
		lang = "C++"
	}

	return &terminalSPJProgram{
		Code:      data.SPJCode,
		Language:  lang,
		ExtraFile: parseTerminalExtraFiles(data.JudgeExtraFile),
	}, nil
}

func parseTerminalExtraFiles(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var tmp map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return nil
	}

	result := make(map[string]string, len(tmp))
	for name, content := range tmp {
		if strings.TrimSpace(name) == "" {
			continue
		}
		switch v := content.(type) {
		case string:
			result[name] = v
		default:
			bytes, _ := json.Marshal(v)
			result[name] = string(bytes)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func newTerminalSPJEngine(program *terminalSPJProgram) (*terminalSPJEngine, error) {
	if program == nil {
		return nil, fmt.Errorf("SPJ 配置为空")
	}
	profile, err := getTerminalSPJProfile(program.Language)
	if err != nil {
		return nil, err
	}

	engine := &terminalSPJEngine{
		runner:  newSandboxRunner(),
		profile: profile,
	}
	if err := engine.prepare(program.Code, program.ExtraFile); err != nil {
		return nil, err
	}
	return engine, nil
}

func (e *terminalSPJEngine) prepare(code string, extraFiles map[string]string) error {
	copyIn := map[string]interface{}{
		e.profile.SourceName: map[string]interface{}{"content": code},
	}
	if _, ok := extraFiles["testlib.h"]; !ok {
		copyIn["testlib.h"] = map[string]interface{}{"content": embeddedTestlibHeader}
	}
	for name, content := range extraFiles {
		if strings.TrimSpace(name) == "" {
			continue
		}
		copyIn[name] = map[string]interface{}{"content": content}
	}

	cmd := map[string]interface{}{
		"args":          e.profile.CompileArgs,
		"env":           e.profile.CompileEnv,
		"files":         sandboxCompileFiles(),
		"cpuLimit":      int64(20) * 1000 * 1000 * 1000,
		"clockLimit":    int64(60) * 1000 * 1000 * 1000,
		"memoryLimit":   int64(512) * 1024 * 1024,
		"procLimit":     sandboxProcLimit,
		"stackLimit":    int64(256) * 1024 * 1024,
		"copyIn":        copyIn,
		"copyOut":       []string{"stdout", "stderr"},
		"copyOutCached": []string{e.profile.ExecName},
	}

	res, err := e.runner.runOne(context.Background(), cmd)
	if err != nil {
		return err
	}
	if res.Status != "Accepted" {
		return fmt.Errorf("SPJ 编译失败: %s", buildCompileErrorMessage(res))
	}

	fileID := ""
	if res.FileIDs != nil {
		fileID = res.FileIDs[e.profile.ExecName]
	}
	if strings.TrimSpace(fileID) == "" {
		return fmt.Errorf("SPJ 编译失败: 沙箱未返回可执行文件ID")
	}
	e.execFileID = fileID
	e.cachedFiles = append(e.cachedFiles, fileID)
	return nil
}

func (e *terminalSPJEngine) Check(input, expected, output string) (*terminalSPJRunResult, error) {
	copyIn := map[string]interface{}{
		e.profile.ExecName: map[string]interface{}{"fileId": e.execFileID},
		"std_input.txt":    map[string]interface{}{"content": normalizeLineBreak(input)},
		"user_output.txt":  map[string]interface{}{"content": normalizeLineBreak(output)},
		"std_output.txt":   map[string]interface{}{"content": normalizeLineBreak(expected)},
	}

	cmd := map[string]interface{}{
		"args":        e.profile.RunArgs,
		"env":         e.profile.RunEnv,
		"files":       sandboxSPJRunFiles(),
		"cpuLimit":    int64(10) * 1000 * 1000 * 1000,
		"clockLimit":  int64(20) * 1000 * 1000 * 1000,
		"memoryLimit": int64(512) * 1024 * 1024,
		"procLimit":   sandboxProcLimit,
		"stackLimit":  int64(256) * 1024 * 1024,
		"copyIn":      copyIn,
		"copyOut":     []string{"stdout", "stderr"},
	}

	res, err := e.runner.runOne(context.Background(), cmd)
	if err != nil {
		return nil, err
	}
	return parseTerminalSPJRunResult(res), nil
}

func (e *terminalSPJEngine) Close(logger *zap.Logger) {
	for _, fileID := range e.cachedFiles {
		if err := e.runner.deleteFile(fileID); err != nil && logger != nil {
			logger.Debug("删除 SPJ 缓存文件失败", zap.String("file_id", fileID), zap.Error(err))
		}
	}
}

func newSandboxRunner() *sandboxRunner {
	baseURL := strings.TrimSpace(os.Getenv("HISTOJ_SANDBOX_URL"))
	if baseURL == "" {
		baseURL = sandboxDefaultURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &sandboxRunner{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 80 * time.Second,
		},
	}
}

func (r *sandboxRunner) runOne(ctx context.Context, cmd map[string]interface{}) (*sandboxCmdResult, error) {
	reqBody := sandboxRunRequest{Cmd: []map[string]interface{}{cmd}}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化沙箱请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/run", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建沙箱请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		if isSandboxUnavailableError(err) {
			return nil, fmt.Errorf("%w: 无法连接沙箱服务 %s: %v", errNativeJudgeUnavailable, r.baseURL, err)
		}
		return nil, fmt.Errorf("调用沙箱失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("%w: 沙箱服务返回状态 %d: %s", errNativeJudgeUnavailable, resp.StatusCode, trimToMax(string(body), 500))
		}
		return nil, fmt.Errorf("沙箱执行失败: HTTP %d: %s", resp.StatusCode, trimToMax(string(body), 1000))
	}

	var results []sandboxCmdResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("解析沙箱响应失败: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("沙箱响应为空")
	}

	return &results[0], nil
}

func (r *sandboxRunner) deleteFile(fileID string) error {
	if strings.TrimSpace(fileID) == "" {
		return nil
	}
	req, err := http.NewRequest(http.MethodDelete, r.baseURL+"/file/"+url.PathEscape(fileID), nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete file failed: status=%d body=%s", resp.StatusCode, trimToMax(string(body), 200))
	}
	return nil
}

func isSandboxUnavailableError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	return false
}

func sandboxCompileFiles() []interface{} {
	return []interface{}{
		map[string]interface{}{"content": ""},
		map[string]interface{}{"name": "stdout", "max": sandboxStdIOMaxBytes},
		map[string]interface{}{"name": "stderr", "max": sandboxStdIOMaxBytes},
	}
}

func sandboxRunFiles(input string) []interface{} {
	return []interface{}{
		map[string]interface{}{"content": input},
		map[string]interface{}{"name": "stdout", "max": sandboxOutputMaxBytes},
		map[string]interface{}{"name": "stderr", "max": sandboxStdIOMaxBytes},
	}
}

func sandboxSPJRunFiles() []interface{} {
	return []interface{}{
		map[string]interface{}{"content": ""},
		map[string]interface{}{"name": "stdout", "max": sandboxStdIOMaxBytes},
		map[string]interface{}{"name": "stderr", "max": sandboxStdIOMaxBytes},
	}
}

func buildCompileErrorMessage(res *sandboxCmdResult) string {
	if res == nil {
		return "未知编译错误"
	}
	stderr := ""
	stdout := ""
	if res.Files != nil {
		stderr = strings.TrimSpace(res.Files["stderr"])
		stdout = strings.TrimSpace(res.Files["stdout"])
	}
	statusLine := ""
	if strings.TrimSpace(res.Status) != "" {
		statusLine = "Sandbox Status: " + res.Status
	}
	return pickFirstNonEmpty(stderr, stdout, statusLine, "编译失败")
}

func getTerminalLanguageProfile(language string) (*terminalLanguageProfile, error) {
	profile := &terminalLanguageProfile{
		CompileEnv:  sandboxDefaultEnv,
		RunEnv:      sandboxDefaultEnv,
		NeedCompile: true,
	}

	switch {
	case strings.Contains(language, "C++"):
		profile.SourceName = "main.cpp"
		profile.ExecName = "main"
		profile.CompileArgs = []string{"/usr/bin/g++", "-DONLINE_JUDGE", "-O2", "-w", "-fmax-errors=1", "-std=gnu++17", "-I", "/w", profile.SourceName, "-lm", "-o", profile.ExecName}
		profile.RunArgs = []string{"/w/" + profile.ExecName}
		profile.NeedCachedExec = true
		profile.InjectTestlib = true
	case strings.Contains(language, "C With"):
		profile.SourceName = "main.c"
		profile.ExecName = "main"
		profile.CompileArgs = []string{"/usr/bin/gcc", "-DONLINE_JUDGE", "-O2", "-w", "-fmax-errors=1", "-std=c11", "-I", "/w", profile.SourceName, "-lm", "-o", profile.ExecName}
		profile.RunArgs = []string{"/w/" + profile.ExecName}
		profile.NeedCachedExec = true
	case strings.Contains(language, "Java"):
		profile.SourceName = "Main.java"
		profile.ExecName = "Main.jar"
		profile.CompileArgs = []string{"/bin/bash", "-c", "javac -encoding utf-8 Main.java && jar -cf Main.jar *.class"}
		profile.RunArgs = []string{"/usr/bin/java", "-Dfile.encoding=UTF-8", "-cp", "/w/Main.jar", "Main"}
		profile.NeedCachedExec = true
	case strings.Contains(language, "Python"):
		profile.SourceName = "main.py"
		profile.ExecName = "main.py"
		profile.CompileArgs = []string{"/usr/bin/python3", "-m", "py_compile", profile.SourceName}
		profile.RunArgs = []string{"/usr/bin/python3", "/w/" + profile.SourceName}
		profile.CompileEnv = sandboxPythonEnv
		profile.RunEnv = sandboxPythonEnv
		profile.NeedCachedExec = false
		profile.NeedSourceAtRun = true
	default:
		return nil, fmt.Errorf("不支持的语言: %s", language)
	}

	return profile, nil
}

func getTerminalSPJProfile(language string) (*terminalSPJProfile, error) {
	profile := &terminalSPJProfile{
		ExecName:   "spj",
		RunArgs:    []string{"/w/spj", "/w/std_input.txt", "/w/user_output.txt", "/w/std_output.txt"},
		CompileEnv: sandboxDefaultEnv,
		RunEnv:     sandboxDefaultEnv,
	}

	lang := strings.TrimSpace(strings.ToLower(language))
	switch {
	case strings.Contains(lang, "c++"):
		profile.SourceName = "spj.cpp"
		profile.CompileArgs = []string{
			"/usr/bin/g++", "-DONLINE_JUDGE", "-O2", "-w", "-fmax-errors=3", "-std=c++14",
			"-I", "/w", profile.SourceName, "-lm", "-o", profile.ExecName,
		}
	case strings.HasPrefix(lang, "c"):
		profile.SourceName = "spj.c"
		profile.CompileArgs = []string{
			"/usr/bin/gcc", "-DONLINE_JUDGE", "-O2", "-w", "-fmax-errors=3", "-std=c11",
			"-I", "/w", profile.SourceName, "-lm", "-o", profile.ExecName,
		}
	default:
		return nil, fmt.Errorf("不支持的 SPJ 语言: %s", language)
	}
	return profile, nil
}

func sandboxRunMemoryLimitMB(problemLimitMB int64) int64 {
	if problemLimitMB <= 0 {
		return 512
	}
	mb := problemLimitMB + 128
	if mb < 256 {
		mb = 256
	}
	if mb > 2048 {
		mb = 2048
	}
	return mb
}

func mapSandboxStatus(status string, exitStatus int) int {
	switch status {
	case "Accepted":
		return 0
	case "Time Limit Exceeded":
		return 1
	case "Memory Limit Exceeded":
		return 2
	case "Output Limit Exceeded", "Nonzero Exit Status", "Signalled":
		return 3
	case "Internal Error", "File Error":
		return 4
	default:
		if exitStatus != 0 {
			return 3
		}
		return 4
	}
}

func parseTerminalSPJRunResult(res *sandboxCmdResult) *terminalSPJRunResult {
	result := &terminalSPJRunResult{
		Code: terminalSPJCodeError,
	}
	if res == nil {
		result.ErrMessage = "SPJ 执行失败"
		return result
	}

	spjStdout := ""
	spjStderr := ""
	if res.Files != nil {
		spjStdout = strings.TrimSpace(res.Files["stdout"])
		spjStderr = strings.TrimSpace(res.Files["stderr"])
	}
	if spjStderr != "" {
		result.ErrMessage = trimToMax(spjStderr, 1024*1024)
	}

	sandboxStatus := mapSandboxStatus(res.Status, res.ExitStatus)
	exitCode := res.ExitStatus

	switch sandboxStatus {
	case 0:
		if exitCode == 0 {
			result.Code = terminalSPJCodeAC
		} else {
			result.Code = exitCode
		}
	case 3:
		switch exitCode {
		case terminalSPJCodeWA, terminalSPJCodeError, terminalSPJCodeAC, terminalSPJCodePE:
			result.Code = exitCode
		case terminalSPJCodePC:
			result.Code = terminalSPJCodePC
			if pct, ok := parseTerminalSPJPercentage(spjStdout); ok {
				result.Percentage = pct
				if pct >= 1 {
					result.Code = terminalSPJCodeAC
				}
			}
		default:
			if spjStderr != "" {
				return parseTerminalTestlibErr(spjStderr)
			}
			result.Code = terminalSPJCodeError
			result.ErrMessage = pickFirstNonEmpty(result.ErrMessage, buildSandboxSPJRuntimeMessage(res))
		}
	default:
		result.Code = terminalSPJCodeError
		result.ErrMessage = pickFirstNonEmpty(result.ErrMessage, buildSandboxSPJRuntimeMessage(res))
	}

	return result
}

func parseTerminalSPJPercentage(raw string) (float64, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0, false
	}
	number, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return number / 100.0, true
}

func parseTerminalTestlibErr(msg string) *terminalSPJRunResult {
	output := strings.TrimSpace(msg)
	if output == "" {
		return &terminalSPJRunResult{Code: terminalSPJCodeError}
	}
	output = trimToMax(output, 1024)

	lower := strings.ToLower(output)
	switch {
	case strings.HasPrefix(lower, "ok "):
		return &terminalSPJRunResult{
			Code:       terminalSPJCodeAC,
			ErrMessage: strings.TrimSpace(output[3:]),
		}
	case strings.HasPrefix(lower, "wrong answer "):
		return &terminalSPJRunResult{
			Code:       terminalSPJCodeWA,
			ErrMessage: strings.TrimSpace(output[len("wrong answer "):]),
		}
	case strings.HasPrefix(lower, "wrong output format "):
		return &terminalSPJRunResult{
			Code:       terminalSPJCodeWA,
			ErrMessage: "May be output presentation error. " + strings.TrimSpace(output[len("wrong output format "):]),
		}
	case strings.HasPrefix(lower, "partially correct "):
		result := &terminalSPJRunResult{
			Code:       terminalSPJCodePC,
			ErrMessage: strings.TrimSpace(output[len("partially correct "):]),
		}
		left := strings.Index(output, "(")
		right := strings.Index(output, ")")
		if left >= 0 && right > left+1 {
			value := strings.TrimSpace(output[left+1 : right])
			if i, err := strconv.Atoi(value); err == nil {
				result.Percentage = float64(i) / 100.0
			}
		}
		return result
	case strings.HasPrefix(lower, "points "):
		result := &terminalSPJRunResult{
			Code:       terminalSPJCodePC,
			ErrMessage: strings.TrimSpace(output[len("points "):]),
		}
		fields := strings.Fields(strings.TrimSpace(output[len("points "):]))
		if len(fields) > 0 {
			if f, err := strconv.ParseFloat(fields[0], 64); err == nil {
				result.Percentage = f / 100.0
				if math.Abs(result.Percentage-1.0) < 1e-9 {
					result.Code = terminalSPJCodeAC
				}
			}
		}
		return result
	case strings.HasPrefix(lower, "fail "):
		return &terminalSPJRunResult{
			Code:       terminalSPJCodeError,
			ErrMessage: strings.TrimSpace(output[len("FAIL "):]),
		}
	default:
		return &terminalSPJRunResult{
			Code:       terminalSPJCodeError,
			ErrMessage: output,
		}
	}
}

func buildSandboxSPJRuntimeMessage(res *sandboxCmdResult) string {
	if res == nil {
		return "SPJ 运行失败"
	}
	parts := make([]string, 0, 3)
	if strings.TrimSpace(res.Status) != "" {
		parts = append(parts, "Sandbox Status: "+res.Status)
	}
	if res.ExitStatus != 0 {
		parts = append(parts, fmt.Sprintf("Exit Status: %d", res.ExitStatus))
	}
	if res.Files != nil && strings.TrimSpace(res.Files["stderr"]) != "" {
		parts = append(parts, strings.TrimSpace(res.Files["stderr"]))
	}
	if len(parts) == 0 {
		return "SPJ 运行失败"
	}
	return strings.Join(parts, "\n")
}

func buildSandboxRuntimeMessage(res *terminalRunResult) string {
	if res == nil {
		return "运行失败"
	}
	parts := make([]string, 0, 3)
	if strings.TrimSpace(res.Status) != "" {
		parts = append(parts, "Sandbox Status: "+res.Status)
	}
	if res.ExitStatus != 0 {
		parts = append(parts, fmt.Sprintf("Exit Status: %d", res.ExitStatus))
	}
	if strings.TrimSpace(res.Stderr) != "" {
		parts = append(parts, strings.TrimSpace(res.Stderr))
	}
	if len(parts) == 0 {
		return "运行失败"
	}
	return strings.Join(parts, "\n")
}

func buildCasesFromSamples(problemID int64, samples []SampleResult) []*terminalJudgeCase {
	if len(samples) == 0 {
		return nil
	}
	result := make([]*terminalJudgeCase, 0, len(samples))
	for i, sample := range samples {
		seq := i + 1
		caseID := int64(seq)
		result = append(result, &terminalJudgeCase{
			ProblemID:   problemID,
			CaseID:      caseID,
			Seq:         seq,
			InputRef:    normalizeLineBreak(sample.Input),
			ExpectedRef: normalizeLineBreak(sample.Expected),
			Mode:        "default",
		})
	}
	return result
}

func isSPJMode(mode string) bool {
	return strings.EqualFold(strings.TrimSpace(mode), "spj")
}

func evaluateTerminalOutput(input, expected, output, currentStderr string, spjEngine *terminalSPJEngine) (int, string, error) {
	stderr := currentStderr
	if spjEngine == nil {
		if judgeOutputEquals(output, expected) {
			return 0, stderr, nil
		}
		if strings.TrimSpace(stderr) == "" {
			stderr = buildWrongAnswerMessage(expected, output)
		}
		return -1, stderr, nil
	}

	spjResult, err := spjEngine.Check(input, expected, output)
	if err != nil {
		return 4, err.Error(), err
	}

	if strings.TrimSpace(spjResult.ErrMessage) != "" {
		stderr = spjResult.ErrMessage
	}

	status := mapTerminalSPJCodeToJudgeStatus(spjResult.Code)
	switch status {
	case 0:
		return 0, stderr, nil
	case -1:
		if strings.TrimSpace(stderr) == "" {
			stderr = "Wrong Answer"
		}
	case -3:
		if strings.TrimSpace(stderr) == "" {
			stderr = "Presentation Error"
		}
	case 8:
		if strings.TrimSpace(stderr) == "" {
			stderr = fmt.Sprintf("Partially Correct (%.2f%%)", spjResult.Percentage*100)
		}
	default:
		if strings.TrimSpace(stderr) == "" {
			stderr = "SPJ 评测失败"
		}
	}
	return status, stderr, nil
}

func mapTerminalSPJCodeToJudgeStatus(code int) int {
	switch code {
	case terminalSPJCodeAC:
		return 0
	case terminalSPJCodeWA:
		return -1
	case terminalSPJCodePE:
		return -3
	case terminalSPJCodePC:
		return 8
	default:
		return 4
	}
}

func validateProblemCaseRef(problemID int64, raw string) error {
	if !looksLikeCaseFileName(raw) {
		return nil
	}
	caseDir := fmt.Sprintf("problem_%d", problemID)
	for _, base := range testcaseBaseDirs() {
		target := filepath.Join(base, caseDir, raw)
		info, err := os.Stat(target)
		if err == nil && !info.IsDir() {
			return nil
		}
	}
	return fmt.Errorf("测试文件不存在: %s (problem_%d)", raw, problemID)
}

func shouldStopAfterCase(caseMode, problemMode string, status int) bool {
	if status == 0 {
		return false
	}
	mode := normalizeJudgeCaseMode(caseMode)
	if mode == "" || mode == "default" {
		mode = normalizeJudgeCaseMode(problemMode)
	}
	return mode == "ergodic_without_error"
}

func normalizeJudgeCaseMode(mode string) string {
	return strings.TrimSpace(strings.ToLower(mode))
}

func isUnknownColumnError(err error, column string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	col := strings.ToLower(strings.TrimSpace(column))
	if col == "" {
		return false
	}
	return strings.Contains(msg, "unknown column") && strings.Contains(msg, "'"+col+"'")
}

func resolveProblemCaseContent(problemID int64, raw string) (string, error) {
	if !looksLikeCaseFileName(raw) {
		return normalizeLineBreak(raw), nil
	}

	caseDir := fmt.Sprintf("problem_%d", problemID)
	for _, base := range testcaseBaseDirs() {
		target := filepath.Join(base, caseDir, raw)
		content, err := os.ReadFile(target)
		if err == nil {
			return normalizeLineBreak(string(content)), nil
		}
	}

	return "", fmt.Errorf("测试文件不存在: %s (problem_%d)", raw, problemID)
}

func testcaseBaseDirs() []string {
	defaults := []string{"/hoj/testcase", "/judge/test_case"}
	env := strings.TrimSpace(os.Getenv("HISTOJ_TESTCASE_DIR"))
	if env == "" {
		return defaults
	}

	parts := strings.FieldsFunc(env, func(r rune) bool {
		return r == ',' || r == ';'
	})

	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts)+len(defaults))
	for _, p := range parts {
		dir := strings.TrimSpace(p)
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		result = append(result, dir)
	}
	for _, d := range defaults {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		result = append(result, d)
	}
	return result
}

func looksLikeCaseFileName(value string) bool {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" {
		return false
	}
	return strings.HasSuffix(v, ".in") ||
		strings.HasSuffix(v, ".out") ||
		strings.HasSuffix(v, ".ans") ||
		strings.HasSuffix(v, ".txt")
}

func terminalCaseTimeout(timeLimitMS int64) time.Duration {
	if timeLimitMS <= 0 {
		return 2 * time.Second
	}
	// 与题面时间限制对齐，不再统一额外 +500ms。
	timeout := timeLimitMS
	if timeout < 50 {
		timeout = 50
	}
	if timeout > 120000 {
		timeout = 120000
	}
	return time.Duration(timeout) * time.Millisecond
}

func terminalFinalErrorMessage(seq int, status int, stderr string) string {
	statusText := client.GetStatusText(status)
	if strings.TrimSpace(stderr) == "" {
		return fmt.Sprintf("Case #%d: %s", seq, statusText)
	}
	return fmt.Sprintf("Case #%d: %s\n%s", seq, statusText, trimToMax(stderr, 2000))
}

func buildWrongAnswerMessage(expected, output string) string {
	expectedPreview := trimToMax(strings.TrimSpace(normalizeLineBreak(expected)), 300)
	outputPreview := trimToMax(strings.TrimSpace(normalizeLineBreak(output)), 300)
	if outputPreview == "" {
		outputPreview = "(空输出)"
	}
	return fmt.Sprintf("Wrong Answer\nExpected: %s\nOutput: %s", expectedPreview, outputPreview)
}

func judgeOutputEquals(output, expected string) bool {
	return normalizeJudgeOutput(output) == normalizeJudgeOutput(expected)
}

func normalizeJudgeOutput(value string) string {
	value = normalizeLineBreak(value)
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func normalizeLineBreak(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}

func pickFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func trimToMax(value string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
