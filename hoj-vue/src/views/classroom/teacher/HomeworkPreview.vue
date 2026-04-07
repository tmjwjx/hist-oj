<template>
  <div class="homework-preview">
    <div class="header">
      <h2>全卷预览（学生视角）</h2>
      <div class="actions">
        <el-button @click="goBack">返回</el-button>
      </div>
    </div>

    <div class="preview-container" v-loading="loading">
      <!-- 作业基本信息 -->
      <el-card class="preview-info" shadow="never">
        <h3>{{ homeworkInfo.title || '作业标题' }}</h3>
        <p><strong>描述：</strong>{{ homeworkInfo.description || '无' }}</p>
        <p><strong>开始时间：</strong>{{ formatTime(homeworkInfo.startTime) }}</p>
        <p><strong>结束时间：</strong>{{ formatTime(homeworkInfo.endTime) }}</p>
        <p><strong>总分：</strong>{{ getTotalScore() }} 分</p>
        <p><strong>题目数量：</strong>{{ selectedQuestions.length }} 题</p>
      </el-card>

      <!-- 题目列表（和学生看到的一样） -->
      <div class="questions-preview">
        <el-divider content-position="left">题目内容</el-divider>
        <div v-for="(item, index) in selectedQuestions" :key="item.id || item.problemId" class="question-item">
          <div class="question-header">
            <span class="question-number">{{ index + 1 }}.</span>
            <span class="question-type">({{ getQuestionTypeText(item.type) }})</span>
            <span class="question-score">{{ item.score }}分</span>
          </div>
          <div class="question-title markdown-body" v-html="renderMarkdown(item.title)"></div>
          <div class="question-content markdown-body" v-html="renderMarkdown(item.content)"></div>

          <!-- 单选题选项（预览模式，不可作答） -->
          <div v-if="item.type === 'single_choice'" class="question-options">
            <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
              <div class="option-preview">
                <span class="option-letter">{{ option.letter }}.</span>
                <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
              </div>
            </div>
          </div>

          <!-- 多选题选项（预览模式，不可作答） -->
          <div v-if="item.type === 'multiple_choice'" class="question-options">
            <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
              <div class="option-preview">
                <span class="option-letter">{{ option.letter }}.</span>
                <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
              </div>
            </div>
          </div>

          <!-- 判断题（预览模式，不可作答） -->
          <div v-if="item.type === 'judge'" class="question-options">
            <div class="option-preview">
              <span class="option-letter">✓</span>
              <span class="option-text">正确</span>
            </div>
            <div class="option-preview">
              <span class="option-letter">✗</span>
              <span class="option-text">错误</span>
            </div>
          </div>

          <!-- 主观题 -->
          <div v-if="item.type === 'subjective'" class="subjective-preview">
            <el-alert type="info" :closable="false">
              <i class="el-icon-edit"></i> 主观题，学生需要在此处输入文字答案
            </el-alert>
          </div>

          <!-- 编程题 -->
          <div v-if="item.type === 'programming'" class="programming-preview">
            <template v-if="getProgrammingProblemInfo(item.problemId)">
              <div class="problem-preview">
                <el-divider content-position="left">编程题详情</el-divider>
                <el-card>
                  <h3>{{ getProgrammingProblemInfo(item.problemId).problem.title }}</h3>
                  <div class="problem-meta">
                    <el-tag size="small">题目ID: {{ getProgrammingProblemInfo(item.problemId).problem.problemId }}</el-tag>
                    <el-tag size="small" type="info">时间限制: {{ getProgrammingProblemInfo(item.problemId).problem.timeLimit }}ms</el-tag>
                    <el-tag size="small" type="warning">内存限制: {{ getProgrammingProblemInfo(item.problemId).problem.memoryLimit }}MB</el-tag>
                    <el-tag size="small" type="success">判题模式: {{ getJudgeModeText(getProgrammingProblemInfo(item.problemId).problem.judgeMode) }}</el-tag>
                  </div>
                  <div class="problem-content">
                    <div class="content-section">
                      <h4>题目描述</h4>
                      <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.description)"></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.input">
                      <h4>输入格式</h4>
                      <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.input)"></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.output">
                      <h4>输出格式</h4>
                      <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.output)"></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingExamplesForProblem(item.problemId).length > 0">
                      <h4>样例</h4>
                      <div v-for="(example, idx) in getProgrammingExamplesForProblem(item.problemId)" :key="idx" class="example-item">
                        <el-alert :title="`样例 ${idx + 1}`" type="info" :closable="false">
                          <div slot="default">
                            <p><strong>输入：</strong></p>
                            <pre>{{ example.input }}</pre>
                            <p><strong>输出：</strong></p>
                            <pre>{{ example.output }}</pre>
                          </div>
                        </el-alert>
                      </div>
                    </div>
                  </div>
                </el-card>
              </div>
            </template>

            <el-alert v-else type="warning" :closable="false">
              <p>学生将看到完整的编程题目界面，包括题目描述、样例、代码编辑器和提交按钮</p>
            </el-alert>

            <el-alert v-if="!item.problemId" type="warning" :closable="false">
              此编程题未关联 BingOJ 题目 ID
            </el-alert>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import { getJudgeInfo } from '@/common/judgeTerminal'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: false
})
md.use(MarkdownItKatex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'HomeworkPreview',
  data() {
    return {
      loading: false,
      homeworkInfo: {
        title: '',
        description: '',
        startTime: null,
        endTime: null
      },
      selectedQuestions: [],
      programmingProblemsCache: {}
    }
  },
  computed: {
    classroomId() {
      return this.$route.params.classroomId
    }
  },
  async mounted() {
    await this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        // 从 sessionStorage 读取预览数据
        const previewData = sessionStorage.getItem('homework_preview_data')
        if (!previewData) {
          this.$message.error('未找到预览数据')
          this.goBack()
          return
        }

        const data = JSON.parse(previewData)
        this.homeworkInfo = data.homeworkInfo
        this.selectedQuestions = data.selectedQuestions

        // 加载编程题数据
        const programmingQuestions = this.selectedQuestions.filter(q => q.type === 'programming' && q.problemId)
        if (programmingQuestions.length > 0) {
          await this.loadProgrammingProblemsForPreview()
        }
      } catch (error) {
        this.$message.error('加载预览数据失败')
        console.error(error)
      } finally {
        this.loading = false
      }
    },
    async loadProgrammingProblemsForPreview() {
      const programmingQuestions = this.selectedQuestions.filter(q => q.type === 'programming' && q.problemId)
      const problemIds = programmingQuestions.map(q => q.problemId)
      await this.loadProgrammingProblemByIds(problemIds)
    },
    async loadProgrammingProblemByIds(problemIds) {
      if (!problemIds || problemIds.length === 0) return

      const userInfo = this.$store.getters.userInfo
      const token = localStorage.getItem('token')

      if (!userInfo || !token) {
        console.warn('未登录，无法加载编程题数据')
        return
      }

      const uncachedIds = problemIds.filter(id => !this.programmingProblemsCache[id])

      for (const problemId of uncachedIds) {
        try {
          const res = await getJudgeInfo({
            pid: problemId,
            cid: '0',
            mode: 'normal',
            username: userInfo.username,
            token: token,
            password: ''
          })

          if (res.code === 200 && res.data) {
            this.$set(this.programmingProblemsCache, problemId, res.data)
          }
        } catch (error) {
          console.error('加载编程题信息失败:', problemId, error)
        }
      }
    },
    getProgrammingProblemInfo(problemId) {
      return this.programmingProblemsCache[problemId] || null
    },
    getProgrammingExamplesForProblem(problemId) {
      const problemInfo = this.programmingProblemsCache[problemId]
      if (!problemInfo || !problemInfo.problem || !problemInfo.problem.examples) {
        return []
      }

      const examples = []
      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      let match

      while ((match = regex.exec(problemInfo.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      return examples
    },
    getTotalScore() {
      return this.selectedQuestions.reduce((sum, q) => sum + (q.score || 0), 0)
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        return options.map((opt, idx) => ({
          letter: String.fromCharCode(65 + idx),
          text: opt
        }))
      } catch (e) {
        return []
      }
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        composite: '组合题',
        subjective: this.$t('m.Subjective'),
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    getJudgeModeText(mode) {
      const modeMap = {
        'default': '默认模式',
        'spj': '特殊判题 (SPJ)',
        'interactive': '交互式',
        'subtask': '子任务'
      }
      return modeMap[mode] || mode || '默认模式'
    },
    renderMarkdown(text) {
      if (!text) return ''
      try {
        return md.render(text)
      } catch (e) {
        return text
      }
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    goBack() {
      this.$router.back()
    }
  }
}
</script>

<style scoped>
.homework-preview {
  padding: 20px;
  background: #f5f5f7;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
}

.header .el-button {
  background: #007aff;
  border: none;
  color: #ffffff;
  font-size: 14px;
  padding: 8px 16px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.header .el-button:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.preview-container {
  max-width: 900px;
  margin: 0 auto;
}

.preview-info {
  margin-bottom: 20px;
  background-color: #f5f7fa;
}

.preview-info h3 {
  margin: 0 0 15px 0;
  color: #409EFF;
  font-size: 20px;
}

.preview-info p {
  margin: 8px 0;
  color: #606266;
}

.questions-preview {
  margin-top: 20px;
}

.question-item {
  padding: 20px;
  margin-bottom: 20px;
  background-color: #fff;
  border: 1px solid #DCDFE6;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.question-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 2px solid #E4E7ED;
}

.question-number {
  font-size: 18px;
  font-weight: bold;
  color: #409EFF;
}

.question-type {
  color: #909399;
  font-size: 14px;
}

.question-score {
  margin-left: auto;
  color: #E6A23C;
  font-weight: bold;
}

.question-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 15px;
  color: #303133;
}

.question-content {
  margin: 15px 0;
  color: #606266;
  line-height: 1.8;
  font-size: 14px;
}

.question-options {
  margin-top: 15px;
  padding-left: 20px;
}

.option-item {
  margin-bottom: 10px;
}

.option-preview {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  background-color: #F5F7FA;
  border-radius: 8px;
  cursor: default;
}

.option-preview:hover {
  background-color: #ECF5FF;
}

.option-letter {
  display: inline-block;
  min-width: 30px;
  font-weight: bold;
  color: #409EFF;
  font-size: 15px;
}

.option-text {
  flex: 1;
  color: #606266;
  line-height: 1.6;
}

.subjective-preview,
.programming-preview {
  margin-top: 15px;
  padding: 15px;
  background-color: #FDF6EC;
  border-left: 3px solid #E6A23C;
  border-radius: 8px;
}

.problem-preview {
  margin: 20px 0;
}

.problem-preview h3 {
  margin-top: 0;
  color: #303133;
  font-size: 18px;
}

.problem-meta {
  margin: 15px 0;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.problem-content {
  margin-top: 20px;
}

.content-section {
  margin-bottom: 20px;
}

.content-section h4 {
  color: #409EFF;
  font-size: 16px;
  margin-bottom: 10px;
  border-left: 3px solid #409EFF;
  padding-left: 10px;
}

.content-section pre {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 10px 0;
}

.example-item {
  margin-bottom: 15px;
}

.example-item pre {
  margin: 8px 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 卡片样式优化 */
.homework-preview >>> .el-card {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: none;
}

.homework-preview >>> .el-card__body {
  padding: 24px;
}

/* 分割线样式 */
.homework-preview >>> .el-divider {
  margin: 20px 0;
  background-color: #e0e0e0;
}

/* Alert 样式优化 */
.homework-preview >>> .el-alert {
  border-radius: 8px;
  border: none;
}

/* Tag 样式优化 */
.homework-preview >>> .el-tag {
  border-radius: 6px;
  padding: 4px 12px;
  font-size: 13px;
  font-weight: 500;
}

.programming-content {
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.programming-content >>> * {
  max-width: 100%;
  overflow-x: auto;
}

.programming-content >>> pre,
.programming-content >>> code {
  white-space: pre-wrap !important;
  word-break: break-word !important;
  max-width: 100% !important;
  overflow-x: auto !important;
}

.programming-content >>> .katex,
.programming-content >>> .katex-display {
  max-width: 100% !important;
  overflow-x: auto !important;
}

.programming-content >>> table {
  max-width: 100% !important;
  overflow-x: auto !important;
  display: block !important;
}
</style>

<!-- 引入 KaTeX 样式 -->
<style>
@import '~katex/dist/katex.min.css';
</style>
