# 修复 Markdown 和 LaTeX 数学公式渲染

## 问题原因

之前使用的 `marked` 库不支持 LaTeX 数学公式渲染，导致 `$\frac{3}{5}$` 这样的数学公式无法正确显示。

## 解决方案

将所有 Markdown 渲染从 `marked` 切换到 `markdown-it` + `@iktakahiro/markdown-it-katex`，完整支持 LaTeX 数学公式。

## 修改的文件

### 1. 学生作业详情页面
**文件**: `hoj-vue/src/views/classroom/student/HomeworkDetail.vue`

**修改内容**:
- 导入 `MarkdownIt` 和 `katex`
- 导入 KaTeX CSS 样式
- 配置 markdown-it 实例并启用 katex 插件
- 修改 `formatContent` 方法使用 `md.render()` 替代 `marked()`

```javascript
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex)

formatContent(content) {
  if (!content) return ''
  try {
    return md.render(content)
  } catch (e) {
    console.error('Markdown渲染失败:', e)
    return content
  }
}
```

### 2. 教师查看学生提交详情页面
**文件**: `hoj-vue/src/views/classroom/teacher/StudentSubmissionDetail.vue`

**修改内容**: 同上，替换为 `markdown-it` + `katex`

### 3. 题库管理页面
**文件**: `hoj-vue/src/views/classroom/teacher/QuestionBank.vue`

**修改内容**: 同上，替换为 `markdown-it` + `katex`

### 4. 编程题组件
**文件**: `hoj-vue/src/components/classroom/ProgrammingQuestion.vue`

**修改内容**:
- 添加 KaTeX CSS 导入
- 优化 markdown-it 配置选项

## LaTeX 数学公式支持

现在支持以下数学公式语法：

### 行内公式
使用 `$...$` 包裹：
```
这是一个行内公式：$a^2 + b^2 = c^2$
```

### 块级公式
使用 `$$...$$` 包裹：
```
$$
\frac{3}{5}
$$
```

### 支持的符号示例
- 分数: `\frac{a}{b}` → $\frac{a}{b}$
- 上标: `a^2` → $a^2$
- 下标: `a_n` → $a_n$
- 根号: `\sqrt{x}` → $\sqrt{x}$
- 求和: `\sum_{i=1}^{n}` → $\sum_{i=1}^{n}$
- 积分: `\int_{a}^{b}` → $\int_{a}^{b}$
- 希腊字母: `\alpha, \beta, \gamma` → $\alpha, \beta, \gamma$
- 矩阵: `\begin{matrix} ... \end{matrix}`

## 部署步骤

```bash
# 1. 重新构建前端
cd hoj-vue
npm run build

# 2. 部署新的 dist 到服务器
# 将 dist 目录的内容复制到 Nginx 或其他 Web 服务器

# 3. 清除浏览器缓存
# 使用 Ctrl+Shift+Delete 清除缓存
# 或使用无痕模式测试

# 4. 测试功能
# - 创建包含数学公式的题目
# - 查看学生作业页面
# - 验证公式正确渲染
```

## 测试用例

### 测试1：行内公式
**输入**: `计算 $a^2 + b^2 = c^2$ 的值`
**预期输出**: a² + b² = c² 正确显示

### 测试2：分数
**输入**: `$\frac{3}{5}$`
**预期输出**: 显示为 3/5 分数形式

### 测试3：复杂公式
**输入**: `$$\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}$$`
**预期输出**: 完整积分公式正确显示

### 测试4：矩阵
**输入**:
```
$$
\begin{pmatrix}
a & b \\
c & d
\end{pmatrix}
$$
```
**预期输出**: 2x2 矩阵正确显示

## 注意事项

1. **KaTeX CSS 必须导入**: `import 'katex/dist/katex.min.css'`
2. **配置选项**: 使用 `html: true` 允许 HTML 标签
3. **浏览器缓存**: 修改后必须清除浏览器缓存或强制刷新
4. **性能影响**: KaTeX 渲染比纯 Markdown 稍慢，但在可接受范围内

## 其他修复

### 后端 API 修复
**文件**: `hist-oj/internal/api/classroom_api_part2.go`

**修复内容**:
- CreateHomework API 添加 `ShowAnswer` 字段支持
- UpdateHomework API 添加 `ShowAnswer` 字段支持
- 确保教师设置的"允许查看答案"选项能正确保存

### 编程题错误提示修复
**文件**: `hoj-vue/src/components/classroom/ProgrammingQuestion.vue`

**修复内容**:
- 修改错误提示显示条件，只在真正加载失败时显示
- 避免题目加载成功后仍显示错误提示

## 验证清单

部署完成后，请验证：

- [ ] 数学公式 `$\frac{3}{5}$` 正确显示为分数
- [ ] 行内公式 `$a^2$` 正确显示
- [ ] 块级公式 `$$...$$` 正确显示
- [ ] Markdown 其他功能（标题、列表、代码块等）正常
- [ ] 学生作业页面答案正确显示
- [ ] 教师创建作业时"允许查看答案"选项能保存
- [ ] 编程题错误提示逻辑正确
- [ ] 所有页面样式正常

## 相关依赖

`package.json` 中已有的依赖：
```json
{
  "markdown-it": "^14.1.0",
  "@iktakahiro/markdown-it-katex": "^4.0.1",
  "katex": "^0.16.0"
}
```

无需安装新的依赖，所有必需的包都已存在。
