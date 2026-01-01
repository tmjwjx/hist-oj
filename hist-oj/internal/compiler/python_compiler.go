package compiler

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// PythonCompiler Python 编译器
type PythonCompiler struct{}

// NewPythonCompiler 创建 Python 编译器
func NewPythonCompiler() *PythonCompiler {
	return &PythonCompiler{}
}

// Compile Python 不需要编译，只需写入文件
func (p *PythonCompiler) Compile(code, tempDir string) (string, error) {
	srcFile := filepath.Join(tempDir, "main.py")

	// 写入源文件
	if err := writeFile(srcFile, code); err != nil {
		return "", fmt.Errorf("写入源文件失败: %w", err)
	}

	// 检查语法
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-m", "py_compile", srcFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("语法检查失败:\n%s", string(output))
	}

	return srcFile, nil
}

// Run 运行 Python 程序
func (p *PythonCompiler) Run(execPath, input, tempDir string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", execPath)

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
func (p *PythonCompiler) GetLanguageName() string {
	return "Python"
}
