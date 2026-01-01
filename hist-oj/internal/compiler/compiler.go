package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Compiler 编译器接口
type Compiler interface {
	// Compile 编译代码，返回可执行文件路径或错误
	Compile(code, tempDir string) (string, error)
	// Run 运行代码，返回输出、错误信息和运行错误
	Run(execPath, input, tempDir string, timeout time.Duration) (string, string, error)
	// GetLanguageName 获取语言名称
	GetLanguageName() string
}

// CompileResult 编译结果
type CompileResult struct {
	Success    bool
	ExecPath   string
	ErrorMsg   string
}

// RunResult 运行结果
type RunResult struct {
	Success    bool
	Output     string
	ErrorMsg   string
	IsTimeout  bool
}

// GetCompiler 根据语言获取编译器
func GetCompiler(language string) (Compiler, error) {
	switch {
	case strings.Contains(language, "C++"):
		return NewCppCompiler(), nil
	case strings.Contains(language, "C With"):
		return NewCCompiler(), nil
	case strings.Contains(language, "Python"):
		return NewPythonCompiler(), nil
	case strings.Contains(language, "Java"):
		return NewJavaCompiler(), nil
	default:
		return nil, fmt.Errorf("不支持的语言: %s", language)
	}
}

// runCommand 运行命令并返回输出
func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// writeFile 写入文件
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
