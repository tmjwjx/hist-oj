package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// HOJClient HOJ 后端客户端
type HOJClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	token      string
}

// NewHOJClient 创建 HOJ 客户端
func NewHOJClient(baseURL, username, password string) *HOJClient {
	// 创建 Cookie Jar 用于保存会话
	jar, _ := cookiejar.New(nil)

	return &HOJClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Jar:     jar,
		},
		username: username,
		password: password,
	}
}

// Login 登录 HOJ
func (c *HOJClient) Login(username, password string) error {
	url := fmt.Sprintf("%s/api/login", c.baseURL)

	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("序列化登录数据失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取登录响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登录失败: HTTP %d", resp.StatusCode)
	}

	var hojResp HOJResponse
	if err := json.Unmarshal(body, &hojResp); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}

	// HOJ 登录成功时 msg 为 "success"，检查 status 或 code 字段
	if hojResp.Message != "success" && hojResp.Status != 200 && hojResp.Code != 200 {
		return fmt.Errorf("登录失败: %s", hojResp.Message)
	}

	// 保存 token（从响应头获取）
	if token := resp.Header.Get("Authorization"); token != "" {
		c.token = token
	}

	// 也保存用户名和密码用于后续可能的重新登录
	c.username = username
	c.password = password

	return nil
}

// TestJudgeReq 本地测试请求
type TestJudgeReq struct {
	Pid            int64  `json:"pid"`
	Type           string `json:"type"`
	Code           string `json:"code"`
	Language       string `json:"language"`
	UserInput      string `json:"userInput"`
	ExpectedOutput string `json:"expectedOutput"`
	IsRemoteJudge  bool   `json:"isRemoteJudge"`
	Mode           string `json:"mode,omitempty"` // 添加 mode 字段
}

// TestJudgeRes 本地测试结果
type TestJudgeRes struct {
	Status           int    `json:"status"`
	Time             int64  `json:"time"`
	Memory           int64  `json:"memory"`
	Input            string `json:"userInput"`
	Output           string `json:"userOutput"`
	ExpectedOutput   string `json:"expectedOutput"`
	Stderr           string `json:"stderr"`
	ProblemJudgeMode string `json:"problemJudgeMode"`
}

// SubmitJudgeReq 提交判题请求
type SubmitJudgeReq struct {
	Pid      string `json:"pid"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Cid      int64  `json:"cid"`
	IsRemote bool   `json:"isRemote"`
}

// JudgeResult 判题结果
type JudgeResult struct {
	SubmitId     int64  `json:"submitId"`
	Status       int    `json:"status"`
	Time         int    `json:"time"`
	Memory       int    `json:"memory"`
	ErrorMessage string `json:"errorMessage"`
	Score        int    `json:"score"`
}

// HOJResponse HOJ 统一响应格式
type HOJResponse struct {
	Status  int             `json:"status"` // HOJ 使用 status 字段
	Code    int             `json:"code"`   // 某些接口使用 code 字段
	Message string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

// SubmitTestJudge 提交本地测试
func (c *HOJClient) SubmitTestJudge(req *TestJudgeReq) (string, error) {
	url := fmt.Sprintf("%s/api/submit-problem-test-judge", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 调试：打印请求内容
	fmt.Printf("[HOJ Client] 提交测试请求: URL=%s, Body=%s\n", url, string(jsonData))
	fmt.Printf("[HOJ Client] 请求详情: pid=%d, type=%s, language=%s, inputLen=%d\n",
		req.Pid, req.Type, req.Language, len(req.UserInput))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// 添加 Authorization 头
	if c.token != "" {
		httpReq.Header.Set("Authorization", c.token)
		fmt.Printf("[HOJ Client] Authorization: %s\n", c.token[:20]+"...")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试：打印响应
	fmt.Printf("[HOJ Client] 响应: %s\n", string(body))

	var hojResp HOJResponse
	if err := json.Unmarshal(body, &hojResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// HOJ 使用 status 字段，如果不是 200 则为错误
	if hojResp.Status != 200 {
		return "", fmt.Errorf("HOJ 返回错误: %s", hojResp.Message)
	}

	// 返回 testJudgeKey
	var testJudgeKey string
	if err := json.Unmarshal(hojResp.Data, &testJudgeKey); err != nil {
		return "", fmt.Errorf("解析 testJudgeKey 失败: %w", err)
	}

	return testJudgeKey, nil
}

// GetTestJudgeResult 获取本地测试结果
func (c *HOJClient) GetTestJudgeResult(testJudgeKey string) (*TestJudgeRes, error) {
	url := fmt.Sprintf("%s/api/get-test-judge-result?testJudgeKey=%s", c.baseURL, testJudgeKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加 Authorization 头
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试：打印原始响应
	fmt.Printf("[HOJ Client] 获取测试结果响应: %s\n", string(body))

	var hojResp HOJResponse
	if err := json.Unmarshal(body, &hojResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查 status 或 code 字段
	if hojResp.Status != 200 && hojResp.Code != 200 {
		return nil, fmt.Errorf("HOJ 返回错误: %s", hojResp.Message)
	}

	var result TestJudgeRes
	if err := json.Unmarshal(hojResp.Data, &result); err != nil {
		return nil, fmt.Errorf("解析测试结果失败: %w", err)
	}

	return &result, nil
}

// SubmitProblemJudge 提交正式判题
func (c *HOJClient) SubmitProblemJudge(req *SubmitJudgeReq) (*JudgeResult, error) {
	url := fmt.Sprintf("%s/api/submit-problem-judge", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var hojResp HOJResponse
	if err := json.Unmarshal(body, &hojResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查 status 或 code 字段
	if hojResp.Status != 200 && hojResp.Code != 200 {
		return nil, fmt.Errorf("HOJ 返回错误: %s", hojResp.Message)
	}

	var result JudgeResult
	if err := json.Unmarshal(hojResp.Data, &result); err != nil {
		return nil, fmt.Errorf("解析判题结果失败: %w", err)
	}

	return &result, nil
}

// GetSubmissionDetail 获取提交详情
func (c *HOJClient) GetSubmissionDetail(submitId int64) (*JudgeResult, error) {
	url := fmt.Sprintf("%s/api/get-submission-detail?submitId=%d", c.baseURL, submitId)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var hojResp HOJResponse
	if err := json.Unmarshal(body, &hojResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查 status 或 code 字段
	if hojResp.Status != 200 && hojResp.Code != 200 {
		return nil, fmt.Errorf("HOJ 返回错误: %s", hojResp.Message)
	}

	var result JudgeResult
	if err := json.Unmarshal(hojResp.Data, &result); err != nil {
		return nil, fmt.Errorf("解析提交详情失败: %w", err)
	}

	return &result, nil
}
