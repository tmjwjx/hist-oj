package compiler

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// CppCompiler C++ 编译器
type CppCompiler struct {
	isC bool
}

// NewCppCompiler 创建 C++ 编译器
func NewCppCompiler() *CppCompiler {
	return &CppCompiler{isC: false}
}

// NewCCompiler 创建 C 编译器
func NewCCompiler() *CppCompiler {
	return &CppCompiler{isC: true}
}

// Compile 编译 C/C++ 代码
func (c *CppCompiler) Compile(code, tempDir string) (string, error) {
	srcFile := filepath.Join(tempDir, "main.cpp")
	exeFile := filepath.Join(tempDir, "main")

	if runtime.GOOS == "windows" {
		exeFile += ".exe"
	}

	// 写入源文件
	if err := writeFile(srcFile, code); err != nil {
		return "", fmt.Errorf("写入源文件失败: %w", err)
	}

	// 选择编译器
	compiler := "g++"
	args := []string{srcFile, "-o", exeFile, "-O2"}

	if c.isC {
		compiler = "gcc"
	} else {
		args = append(args, "-std=c++17")
	}

	// 编译
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, compiler, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("编译失败:\n%s", string(output))
	}

	if !fileExists(exeFile) {
		return "", fmt.Errorf("编译后未生成可执行文件")
	}

	return exeFile, nil
}

// Run 运行 C/C++ 程序
func (c *CppCompiler) Run(execPath, input, tempDir string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, execPath)

	// 设置输入
	if input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", "", fmt.Errorf("创建输入管道失败: %w", err)
		}

		go func() {
			defer stdin.Close()
			stdin.Write([]byte(input))
		}()
	}

	// 运行程序
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", "", fmt.Errorf("时间超限")
	}

	if err != nil {
		return "", string(output), fmt.Errorf("运行错误: %w", err)
	}

	return string(output), "", nil
}

// GetLanguageName 获取语言名称
func (c *CppCompiler) GetLanguageName() string {
	if c.isC {
		return "C"
	}
	return "C++"
}
