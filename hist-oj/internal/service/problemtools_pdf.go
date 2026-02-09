package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoj/hist-oj/internal/model"
)

// ProblemToolsPDFGenerator 使用 Kattis problemtools 生成 PDF
type ProblemToolsPDFGenerator struct {
	problemtoolsPath string
	tempDir          string
}

// NewProblemToolsPDFGenerator 创建 ProblemTools PDF 生成器
func NewProblemToolsPDFGenerator(problemtoolsPath, tempDir string) *ProblemToolsPDFGenerator {
	return &ProblemToolsPDFGenerator{
		problemtoolsPath: problemtoolsPath,
		tempDir:          tempDir,
	}
}

// GenerateProblemSetPDF 生成题目集 PDF
func (g *ProblemToolsPDFGenerator) GenerateProblemSetPDF(problemSet *model.ProblemSet, images []model.ProblemSetImage) ([]byte, error) {
	// 创建临时工作目录
	workDir := filepath.Join(g.tempDir, fmt.Sprintf("problemset_%d_%d", problemSet.ID, time.Now().Unix()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)

	// 创建 problemset 主目录和各个题目子目录
	// 结构: workDir/problemset/{problem_1, problem_2, ...}
	problemsetDir := filepath.Join(workDir, "problemset")
	if err := os.MkdirAll(problemsetDir, 0755); err != nil {
		return nil, fmt.Errorf("创建 problemset 目录失败: %w", err)
	}

	// 复制图片到工作目录（放在 workDir 下，与 problemset.tex 同级，方便 \includegraphics 引用）
	for _, image := range images {
		dstPath := filepath.Join(workDir, image.Filename)
		if err := copyFile(image.FilePath, dstPath); err != nil {
			return nil, fmt.Errorf("复制图片失败: %w", err)
		}
	}

	// 为每个题目创建 problemtools 格式的包
	for _, problem := range problemSet.Problems {
		problemDir := filepath.Join(problemsetDir, fmt.Sprintf("problem_%d", problem.ID))
		if err := g.createProblemPackage(problemDir, problem); err != nil {
			return nil, fmt.Errorf("创建题目包失败: %w", err)
		}
	}

	// 创建 problemtools 格式的主 LaTeX 文件（放在 workDir 下，与 problemset 目录同级）
	texPath := filepath.Join(workDir, "problemset.tex")
	if err := g.createProblemSetTex(texPath, problemSet, problemsetDir); err != nil {
		return nil, fmt.Errorf("创建题目集 LaTeX 失败: %w", err)
	}

	// 复制 problemset.cls 到 workDir
	problemsetClsSrc := filepath.Join(g.problemtoolsPath, "problemtools/templates/latex/problemset.cls")
	problemsetClsDst := filepath.Join(workDir, "problemset.cls")
	if err := copyFile(problemsetClsSrc, problemsetClsDst); err != nil {
		return nil, fmt.Errorf("复制 problemset.cls 失败: %w", err)
	}

	// 调用 lualatex 直接编译（这是 problemtools 内部使用的方式）
	pdfBytes, err := g.compileProblemsetPDF(workDir, texPath)
	if err != nil {
		return nil, fmt.Errorf("编译 PDF 失败: %w", err)
	}

	return pdfBytes, nil
}

// createProblemPackage 创建单个题目的 problemtools 格式包
func (g *ProblemToolsPDFGenerator) createProblemPackage(problemDir string, problem model.ProblemSetProblem) error {
	// 创建必要的目录
	statementDir := filepath.Join(problemDir, "problem_statement")
	sampleDir := filepath.Join(problemDir, "data", "sample")
	secretDir := filepath.Join(problemDir, "data", "secret")

	if err := os.MkdirAll(statementDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(sampleDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(secretDir, 0755); err != nil {
		return err
	}

	// 写入 problem.yaml
	if err := g.writeProblemYaml(problemDir, problem); err != nil {
		return err
	}

	// 写入题目描述 LaTeX（使用 problemtools 格式）
	if err := g.writeProblemStatement(statementDir, problem); err != nil {
		return err
	}

	// 写入样例数据
	if err := g.writeSampleData(sampleDir, problem); err != nil {
		return err
	}

	// 创建一个空的测试文件以满足 problemtools 要求
	emptyTestFile := filepath.Join(secretDir, "testdata.yaml")
	testDataContent := "input: test.in\noutput: test.ans\n"
	if err := os.WriteFile(emptyTestFile, []byte(testDataContent), 0644); err != nil {
		return err
	}

	return nil
}

// writeProblemYaml 写入 problem.yaml
func (g *ProblemToolsPDFGenerator) writeProblemYaml(problemDir string, problem model.ProblemSetProblem) error {
	yamlPath := filepath.Join(problemDir, "problem.yaml")

	// 时间限制：从毫秒转换为秒
	timeLimitSeconds := float64(problem.TimeLimit) / 1000.0

	content := fmt.Sprintf(`name: %s
license: unknown
limits:
  time: %.1f
`,
		problem.Title,
		timeLimitSeconds,
	)

	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 problem.yaml 失败: %w", err)
	}

	return nil
}

// writeProblemStatement 写入题目描述 LaTeX 文件（使用 problemtools 格式）
func (g *ProblemToolsPDFGenerator) writeProblemStatement(statementDir string, problem model.ProblemSetProblem) error {
	texPath := filepath.Join(statementDir, "problem.en.tex")

	// 使用 problemtools 的命令格式
	var builder strings.Builder

	// 题目名称
	builder.WriteString(fmt.Sprintf(`\problemname{%s}
`, problem.Title))

	// 题目描述
	builder.WriteString(convertToProblemtoolsFormat(problem.Description))

	// Input
	builder.WriteString("\n\\section*{Input}\n")
	builder.WriteString(convertToProblemtoolsFormat(problem.InputFormat))

	// Output
	builder.WriteString("\n\\section*{Output}\n")
	builder.WriteString(convertToProblemtoolsFormat(problem.OutputFormat))

	// 样例说明
	for _, example := range problem.Examples {
		if example.Description != "" {
			builder.WriteString(fmt.Sprintf("\n\\section*{Explanation}\n%s\n", convertToProblemtoolsFormat(example.Description)))
		}
	}

	// Note
	if problem.Note != "" {
		builder.WriteString("\n\\section*{Note}\n")
		builder.WriteString(convertToProblemtoolsFormat(problem.Note))
	}

	if err := os.WriteFile(texPath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("写入题目描述失败: %w", err)
	}

	return nil
}

// writeSampleData 写入样例数据
func (g *ProblemToolsPDFGenerator) writeSampleData(sampleDir string, problem model.ProblemSetProblem) error {
	for _, example := range problem.Examples {
		inputFile := filepath.Join(sampleDir, fmt.Sprintf("%d.in", example.ExampleNo))
		outputFile := filepath.Join(sampleDir, fmt.Sprintf("%d.ans", example.ExampleNo))

		if err := os.WriteFile(inputFile, []byte(example.Input), 0644); err != nil {
			return fmt.Errorf("写入样例输入失败: %w", err)
		}

		if err := os.WriteFile(outputFile, []byte(example.Output), 0644); err != nil {
			return fmt.Errorf("写入样例输出失败: %w", err)
		}

		// 创建 testdata.yaml
		testDataFile := filepath.Join(sampleDir, fmt.Sprintf("%d.yaml", example.ExampleNo))
		testDataContent := fmt.Sprintf("input: %d.in\noutput: %d.ans\n", example.ExampleNo, example.ExampleNo)
		if err := os.WriteFile(testDataFile, []byte(testDataContent), 0644); err != nil {
			return fmt.Errorf("写入 testdata.yaml 失败: %w", err)
		}
	}
	return nil
}

// createProblemSetTex 创建 problemtools 格式的主 LaTeX 文件
func (g *ProblemToolsPDFGenerator) createProblemSetTex(texPath string, problemSet *model.ProblemSet, problemsetDir string) error {
	// 格式化日期
	dateStr := time.Now().Format("2006-01-02")
	if problemSet.ContestDate != nil {
		dateStr = problemSet.ContestDate.Format("2006-01-02")
	}

	// 使用 problemtools 的 problemset documentclass
	// 添加中文支持（lualatex 使用 luatexja 而不是 xeCJK）
	content := fmt.Sprintf(`\documentclass[plainproblems]{problemset}

\usepackage{luatexja}
\usepackage{luatexja-fontspec}

\problemparentpath{problemset}

\begin{document}
\thispagestyle{empty}
\begin{center}
  \vspace*{2cm}
  \Huge \textbf{%s} \\[0.5em]
  \Large \textbf{Problem Set} \\[2cm]
  \Large %s \\[1cm]
  \large %s
\end{center}
\clearpage

`, problemSet.Title, problemSet.Author, dateStr)

	// 添加每个题目（使用 problemtools 的 \includeproblem 命令）
	// 题目在 problemset/problem_1, problemset/problem_2, 等目录中
	// 需要在每个 \includeproblem 前设置 statementdirectory 和 statementfilename
	for _, problem := range problemSet.Problems {
		problemName := fmt.Sprintf("problem_%d", problem.ID)
		// 设置 statement 目录和文件名
		content += fmt.Sprintf(`%% %s
\statementdirectory{problem_statement}
\statementfilename{problem.en.tex}
\includeproblem{%s}
`, problem.Title, problemName)
	}

	content += "\n\\end{document}\n"

	if err := os.WriteFile(texPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入题目集 LaTeX 失败: %w", err)
	}

	return nil
}

// compileProblemsetPDF 使用 lualatex 编译 problemset.tex（这是 problemtools 内部使用的方式）
func (g *ProblemToolsPDFGenerator) compileProblemsetPDF(workDir, texPath string) ([]byte, error) {
	pdfPath := filepath.Join(workDir, "problemset.pdf")

	// 编译两次（解决引用问题）
	for i := 0; i < 2; i++ {
		cmd := exec.Command("lualatex", "-interaction=nonstopmode", filepath.Base(texPath))
		cmd.Dir = workDir

		// 捕获输出用于错误调试
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 返回详细的错误信息
			return nil, fmt.Errorf("lualatex 编译失败 (第%d次): %w\n输出:\n%s", i+1, err, string(output))
		}
	}

	// 读取生成的 PDF
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("读取 PDF 失败: %w", err)
	}

	return pdfBytes, nil
}

// convertToProblemtoolsFormat 转换为 problemtools LaTeX 格式
func convertToProblemtoolsFormat(text string) string {
	// 简单的文本转换
	// TODO: 可以添加更完整的 Markdown 到 LaTeX 转换
	return text
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
