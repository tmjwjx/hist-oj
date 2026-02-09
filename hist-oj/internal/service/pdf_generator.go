package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hoj/hist-oj/internal/model"
)

// PDFGenerator PDF 生成服务
type PDFGenerator struct {
	templateDir string
	tempDir     string
}

// NewPDFGenerator 创建 PDF 生成器
func NewPDFGenerator(templateDir, tempDir string) *PDFGenerator {
	return &PDFGenerator{
		templateDir: templateDir,
		tempDir:     tempDir,
	}
}

// ProblemSetData 题目集数据
type ProblemSetData struct {
	Title               string
	Author              string
	ContestDate         *time.Time
	ContestDateFormatted string
	Problems            []ProblemData
}

// ProblemData 题目数据
type ProblemData struct {
	ID            uint64
	ProblemLetter string
	Title         string
	TimeLimit     int
	Description   string
	InputFormat   string
	OutputFormat  string
	Note          string
	Examples      []ExampleData
}

// ExampleData 样例数据
type ExampleData struct {
	ExampleNo   int
	Description string
	Input       string
	Output      string
	InputFile   string // 文件路径（用于 lstinputlisting）
	OutputFile  string // 文件路径（用于 lstinputlisting）
}

// escapeLaTeX 转义 LaTeX 特殊字符，但保护 $...$ 数学公式
func escapeLaTeX(text string) string {
	if text == "" {
		return ""
	}

	// 先找到所有 $...$ 数学公式块，用占位符替换
	var mathBlocks []string
	placeholder := func(index int) string {
		return fmt.Sprintf("XXXMATHBLOCK%dXXX", index)
	}

	result := text
	blockIndex := 0

	// 查找 $...$ 块（简单实现，不支持嵌套）
	for {
		start := strings.Index(result, "$")
		if start == -1 {
			break
		}
		end := strings.Index(result[start+1:], "$")
		if end == -1 {
			break
		}
		end = start + 1 + end

		// 提取数学公式块
		mathBlock := result[start : end+1]
		mathBlocks = append(mathBlocks, mathBlock)

		// 用占位符替换
		result = result[:start] + placeholder(blockIndex) + result[end+1:]
		blockIndex++
	}

	// 对剩余文本进行转义（不转义 $）
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"&", "\\&",
		"%", "\\%",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\^{}",
		"+", "$+$",           // 转义 + 号（在数学模式中）
		"<", "$<$",           // 转义 < 号
		">", "$>$",           // 转义 > 号
		"|", "\\textbar{}",   // 转义 | 号
	)
	result = replacer.Replace(result)

	// 恢复数学公式块
	for i, block := range mathBlocks {
		result = strings.Replace(result, placeholder(i), block, 1)
	}

	return result
}

// GeneratePDF 生成 PDF
func (g *PDFGenerator) GeneratePDF(problemSet *model.ProblemSet) ([]byte, error) {
	// 诊断：输出 PDF 生成器配置
	fmt.Printf("[PDF-GEN] TemplateDir: %s\n", g.templateDir)
	fmt.Printf("[PDF-GEN] TempDir: %s\n", g.tempDir)
	fmt.Printf("[PDF-GEN] ProblemSet ID: %d, Title: %s\n", problemSet.ID, problemSet.Title)
	fmt.Printf("[PDF-GEN] Problems count: %d\n", len(problemSet.Problems))

	// 创建临时目录
	workDir := filepath.Join(g.tempDir, fmt.Sprintf("problem_set_%d_%d", problemSet.ID, time.Now().Unix()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)

	// 准备数据（需要workDir来写入样例文件）
	data := g.prepareData(problemSet, workDir)

	// 生成主 LaTeX 文件
	mainTexPath := filepath.Join(workDir, "main.tex")
	if err := g.generateMainTex(mainTexPath, data); err != nil {
		return nil, fmt.Errorf("生成主 LaTeX 文件失败: %w", err)
	}

	// 验证生成的 main.tex 文件
	if info, err := os.Stat(mainTexPath); err != nil {
		return nil, fmt.Errorf("无法访问生成的 main.tex: %w", err)
	} else if info.Size() == 0 {
		return nil, fmt.Errorf("生成的 main.tex 文件大小为 0")
	}

	// 生成各个题目的 LaTeX 文件
	for _, problem := range data.Problems {
		problemTexPath := filepath.Join(workDir, fmt.Sprintf("problem_%d.tex", problem.ID))
		if err := g.generateProblemTex(problemTexPath, problem); err != nil {
			return nil, fmt.Errorf("生成题目 %d LaTeX 文件失败: %w", problem.ID, err)
		}
	}

	// 编译 LaTeX 生成 PDF
	pdfPath := filepath.Join(workDir, "main.pdf")
	if err := g.compileLatex(workDir, "main.tex"); err != nil {
		return nil, fmt.Errorf("编译 LaTeX 失败: %w", err)
	}

	// 读取 PDF 文件
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("读取 PDF 文件失败: %w", err)
	}

	// 验证 PDF 文件内容（PDF 文件应该以 %PDF- 开头）
	if len(pdfBytes) < 5 {
		return nil, fmt.Errorf("生成的 PDF 文件太小，可能已损坏")
	}
	if string(pdfBytes[:5]) != "%PDF-" {
		return nil, fmt.Errorf("生成的文件不是有效的 PDF 格式")
	}

	return pdfBytes, nil
}

// prepareData 准备模板数据
func (g *PDFGenerator) prepareData(problemSet *model.ProblemSet, workDir string) *ProblemSetData {
	data := &ProblemSetData{
		Title:       escapeLaTeX(problemSet.Title),
		Author:      escapeLaTeX(problemSet.Author),
		ContestDate: problemSet.ContestDate,
	}

	// 格式化日期
	if problemSet.ContestDate != nil {
		data.ContestDateFormatted = problemSet.ContestDate.Format("January 2, 2006")
	} else {
		data.ContestDateFormatted = time.Now().Format("January 2, 2006")
	}

	// 创建样例文件目录
	samplesDir := filepath.Join(workDir, "samples")
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		fmt.Printf("[WARNING] Failed to create samples directory: %v\n", err)
	}

	// 处理题目并按 sort_order 排序
	problems := make([]model.ProblemSetProblem, len(problemSet.Problems))
	copy(problems, problemSet.Problems)

	// 按排序字段排序
	for i := 0; i < len(problems); i++ {
		for j := i + 1; j < len(problems); j++ {
			if problems[i].SortOrder > problems[j].SortOrder {
				problems[i], problems[j] = problems[j], problems[i]
			}
		}
	}

	for _, problem := range problems {
		// Debug: 输出题目原始数据
		fmt.Printf("[DEBUG] Problem raw data: Letter='%s', Title='%s', TimeLimit=%d\n",
			problem.ProblemLetter, problem.Title, problem.TimeLimit)

		problemData := ProblemData{
			ID:            problem.ID,
			ProblemLetter: problem.ProblemLetter, // 不转义，应该是简单的 A-Z
			Title:         escapeLaTeX(problem.Title),
			TimeLimit:     problem.TimeLimit, // 保持毫秒单位，Go模板会自动转为字符串
			Description:   escapeLaTeX(problem.Description),
			InputFormat:   escapeLaTeX(problem.InputFormat),
			OutputFormat:  escapeLaTeX(problem.OutputFormat),
			Note:          escapeLaTeX(problem.Note),
		}

		// Debug: 输出转义后的数据
		fmt.Printf("[DEBUG] Problem escaped data: Letter='%s', Title='%s', TimeLimit=%d\n",
			problemData.ProblemLetter, problemData.Title, problemData.TimeLimit)

		// 处理样例 - 写入到文件，避免转义问题
		for _, example := range problem.Examples {
			// 生成文件路径
			inputFile := filepath.Join("samples", fmt.Sprintf("problem_%d_sample_%d.in", problem.ID, example.ExampleNo))
			outputFile := filepath.Join("samples", fmt.Sprintf("problem_%d_sample_%d.ans", problem.ID, example.ExampleNo))

			// 写入样例输入文件（不进行任何转义，保持原始内容）
			inputPath := filepath.Join(workDir, inputFile)
			if err := os.WriteFile(inputPath, []byte(example.Input), 0644); err != nil {
				fmt.Printf("[WARNING] Failed to write sample input file: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Wrote sample input file: %s\n", inputPath)
			}

			// 写入样例输出文件（不进行任何转义，保持原始内容）
			outputPath := filepath.Join(workDir, outputFile)
			if err := os.WriteFile(outputPath, []byte(example.Output), 0644); err != nil {
				fmt.Printf("[WARNING] Failed to write sample output file: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Wrote sample output file: %s\n", outputPath)
			}

			// Debug: 输出样例原始数据
			fmt.Printf("[DEBUG] Example %d raw input: '%s'\n", example.ExampleNo, example.Input)
			fmt.Printf("[DEBUG] Example %d raw output: '%s'\n", example.ExampleNo, example.Output)

			problemData.Examples = append(problemData.Examples, ExampleData{
				ExampleNo:   example.ExampleNo,
				Description: escapeLaTeX(example.Description),
				Input:       example.Input, // 保留原始数据（虽然现在不再使用）
				Output:      example.Output, // 保留原始数据（虽然现在不再使用）
				InputFile:   inputFile,
				OutputFile:  outputFile,
			})
		}

		data.Problems = append(data.Problems, problemData)
	}

	return data
}

// generateMainTex 生成主 LaTeX 文件
func (g *PDFGenerator) generateMainTex(outputPath string, data *ProblemSetData) error {
	tmplPath := filepath.Join(g.templateDir, "main.tex")

	tmpl, err := template.New("main.tex").Delims("[[", "]]").Funcs(template.FuncMap{
		"escapeLaTeX": escapeLaTeX,
	}).ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 先渲染到 buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行模板失败: %w", err)
	}

	// 构建问题列表 LaTeX 代码
	// 使用更安全的格式：字母和标题分开，避免花括号冲突
	var problemsListBuilder strings.Builder
	for i, p := range data.Problems {
		// 生成格式：\Large \textbf{A} \hspace{10pt} Title
		// 不使用 \makebox，避免转义后的花括号破坏参数结构
		line := fmt.Sprintf("\\Large \\textbf{%s} \\hspace{10pt} %s", p.ProblemLetter, p.Title)
		if i < len(data.Problems)-1 {
			// 非最后一行：添加 LaTeX 换行符 \\
			line += `\\` + "\n"
		} else {
			// 最后一行：不添加 LaTeX 换行符
			line += "\n"
		}
		problemsListBuilder.WriteString(line)
	}
	problemsList := problemsListBuilder.String()

	// 构建 \input 命令列表
	var problemsInputsBuilder strings.Builder
	for _, p := range data.Problems {
		problemsInputsBuilder.WriteString(fmt.Sprintf("\\input{problem_%d.tex}\n\\clearpage\n", p.ID))
	}
	problemsInputs := problemsInputsBuilder.String()

	// 替换占位符（避免花括号问题）
	content := buf.String()
	content = strings.Replace(content, "TITLEPLACEHOLDER", data.Title, -1)
	content = strings.Replace(content, "AUTHORPLACEHOLDER", data.Author, -1)
	content = strings.Replace(content, "DATEPLACEHOLDER", data.ContestDateFormatted, -1)
	content = strings.Replace(content, "PROBLEMSLISTPLACEHOLDER", problemsList, -1)
	content = strings.Replace(content, "PROBLEMSINPUTS", problemsInputs, -1)

	// 调试：记录生成的 LaTeX 内容（前500字符）
	if len(content) > 500 {
		fmt.Printf("[DEBUG] Generated main.tex (first 500 chars):\n%s\n...\n", content[:500])
	} else {
		fmt.Printf("[DEBUG] Generated main.tex:\n%s\n", content)
	}
	fmt.Printf("[DEBUG] Problems list:\n%s\n", problemsList)
	fmt.Printf("[DEBUG] Problems inputs:\n%s\n", problemsInputs)

	// 写入文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.WriteString(content); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// generateProblemTex 生成题目 LaTeX 文件
func (g *PDFGenerator) generateProblemTex(outputPath string, data ProblemData) error {
	tmplPath := filepath.Join(g.templateDir, "problem.tex")

	// 使用自定义分隔符，避免与LaTeX花括号冲突
	tmpl, err := template.New("problem.tex").Delims("[[", "]]").ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 先渲染到 buffer 以便调试
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行模板失败: %w", err)
	}

	// Debug: 输出生成的 LaTeX 前 200 字符
	generatedContent := buf.String()
	if len(generatedContent) > 200 {
		fmt.Printf("[DEBUG] Generated problem_%d.tex (first 200 chars):\n%s\n...\n",
			data.ID, generatedContent[:200])
	} else {
		fmt.Printf("[DEBUG] Generated problem_%d.tex:\n%s\n", data.ID, generatedContent)
	}

	// 写入文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.WriteString(generatedContent); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// compileLatex 编译 LaTeX 文件（使用 lualatex 以更好地支持中文和Unicode）
func (g *PDFGenerator) compileLatex(workDir, texFile string) error {
	pdfPath := filepath.Join(workDir, "main.pdf")

	// 第一次编译（使用 lualatex，类似 Kattis/problemtools）
	cmd := exec.Command("lualatex", "-interaction=nonstopmode", texFile)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行编译，但不立即报错（因为PDF可能已经生成）
	compileErr := cmd.Run()

	// 检查 PDF 是否已生成（即使编译返回错误，只要PDF生成就算成功）
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		// PDF未生成，这才是真正的错误
		output := stdout.String()
		if compileErr != nil {
			if output != "" {
				return fmt.Errorf("lualatex 第一次编译失败: %w\n输出:\n%s", compileErr, output)
			}
			return fmt.Errorf("lualatex 第一次编译失败: %w", compileErr)
		}
		return fmt.Errorf("第一次编译后 PDF 文件未生成，lualatex 输出:\n%s", output)
	}

	// PDF已生成，记录警告但继续（可能是字体警告等非致命错误）
	if compileErr != nil {
		fmt.Printf("[WARNING] lualatex 第一次编译有警告但PDF已生成: %v\n", compileErr)
	}

	// 第二次编译（用于生成正确的目录和引用）
	cmd = exec.Command("lualatex", "-interaction=nonstopmode", texFile)
	cmd.Dir = workDir

	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	compileErr = cmd.Run()

	// 检查PDF是否仍然存在且有效
	if info, err := os.Stat(pdfPath); err != nil {
		return fmt.Errorf("编译完成后无法访问 PDF 文件: %w", err)
	} else if info.Size() == 0 {
		return fmt.Errorf("生成的 PDF 文件大小为 0")
	}

	// 记录第二次编译的警告（如果有）
	if compileErr != nil {
		fmt.Printf("[WARNING] lualatex 第二次编译有警告但PDF有效: %v\n", compileErr)
	}

	return nil
}
