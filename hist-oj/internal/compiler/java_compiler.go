package compiler

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

// JavaCompiler Java 编译器
type JavaCompiler struct{}

// NewJavaCompiler 创建 Java 编译器
func NewJavaCompiler() *JavaCompiler {
	return &JavaCompiler{}
}

// Compile 编译 Java 代码
func (j *JavaCompiler) Compile(code, tempDir string) (string, error) {
	// 将 public class 替换为 Main
	re := regexp.MustCompile(`public\s+class\s+\w+`)
	code = re.ReplaceAllString(code, "public class Main")

	srcFile := filepath.Join(tempDir, "Main.java")

	// 写入源文件
	if err := writeFile(srcFile, code); err != nil {
		return "", fmt.Errorf("写入源文件失败: %w", err)
	}

	// 编译
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "javac", srcFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("编译失败:\n%s", string(output))
	}

	// 返回 "java" 作为标识，实际运行时需要指定 classpath
	return "java", nil
}

// Run 运行 Java 程序
func (j *JavaCompiler) Run(execPath, input, tempDir string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "java", "-cp", tempDir, "Main")

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
func (j *JavaCompiler) GetLanguageName() string {
	return "Java"
}
