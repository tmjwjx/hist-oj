<template>
  <div class="homework-preview">
    <div class="header">
      <h2>{{ $t('m.Full_Paper_Preview') }}</h2>
      <div class="actions">
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <div class="preview-container" v-loading="loading">
      <!-- 作业基本信息 -->
      <el-card class="preview-info" shadow="never">
        <h3>{{ homeworkInfo.title || $t('m.Homework_Title') }}</h3>
        <p><strong>{{ $t('m.Description') }}:</strong> {{ homeworkInfo.description || $t('m.None') }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong>{{ formatTime(homeworkInfo.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong>{{ formatTime(homeworkInfo.endTime) }}</p>
        <p><strong>{{ $t('m.Total_Score') }}:</strong> {{ getTotalScore() }}</p>
        <p><strong>{{ $t('m.Question_Count') }}:</strong> {{ selectedQuestions.length }}</p>
      </el-card>

      <!-- 题目列表（和学生看到的一样） -->
      <div class="questions-preview">
        <el-divider content-position="left">{{ $t('m.Question_Content') }}</el-divider>
        <div v-for="(item, index) in selectedQuestions" :key="item.id || item.problemId" class="question-item">
          <div class="question-header">
            <span class="question-number">{{ index + 1 }}.</span>
            <span class="question-type">({{ getQuestionTypeText(item.type) }})</span>
            <span class="question-score">{{ item.score }}{{ $t('m.Score_Unit') }}</span>
          </div>
          <div class="question-title markdown-body preview-markdown" v-html="renderMarkdown(item.title)" v-highlight></div>
          <div
            v-if="item.content"
            class="question-content markdown-body preview-markdown"
            v-html="renderMarkdown(item.content)"
            v-highlight
          ></div>

          <!-- 单选题选项（预览模式，不可作答） -->
          <div v-if="item.type === 'single_choice'" class="question-options">
            <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
              <div class="option-preview">
                <span class="option-letter">{{ option.letter }}.</span>
                <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>
          </div>

          <!-- 多选题选项（预览模式，不可作答） -->
          <div v-if="item.type === 'multiple_choice'" class="question-options">
            <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
              <div class="option-preview">
                <span class="option-letter">{{ option.letter }}.</span>
                <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>
          </div>

          <!-- 判断题（预览模式，不可作答） -->
          <div v-if="item.type === 'judge'" class="question-options">
            <div class="option-preview">
              <span class="option-letter">✓</span>
              <span class="option-text">{{ $t('m.True') }}</span>
            </div>
            <div class="option-preview">
              <span class="option-letter">✗</span>
              <span class="option-text">{{ $t('m.False') }}</span>
            </div>
          </div>

          <!-- 组合题 -->
          <div v-if="item.type === 'composite'" class="question-options composite-preview-list">
            <div
              v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(item.options)"
              :key="subQuestion.id || subIndex"
              class="composite-preview-item"
            >
              <div class="composite-preview-header">
                <span>{{ $t('m.Sub_Question') }} {{ subIndex + 1 }}</span>
                <span>{{ Number(subQuestion.score || 0) }}{{ $t('m.Score_Unit') }}</span>
              </div>
              <div class="markdown-body preview-markdown" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
              <div class="question-options" style="padding-left: 0;">
                <div v-for="(option, idx) in parseOptions(subQuestion.options)" :key="idx" class="option-item">
                  <div class="option-preview">
                    <span class="option-letter">{{ option.letter }}.</span>
                    <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 填空题 -->
          <div v-if="item.type === 'fill_blank'" class="subjective-preview">
            <el-alert type="success" :closable="false">
              <i class="el-icon-edit"></i> {{ $t('m.Fill_Blank_Student_Hint') }}
            </el-alert>
          </div>

          <!-- 主观题 -->
          <div v-if="item.type === 'subjective'" class="subjective-preview">
            <el-alert type="info" :closable="false">
              <i class="el-icon-edit"></i> {{ $t('m.Subjective_Student_Hint') }}
            </el-alert>
          </div>

          <!-- 编程题 -->
          <div v-if="item.type === 'programming'" class="programming-preview">
            <template v-if="getProgrammingProblemInfo(item.problemId)">
              <div class="problem-preview">
                <el-divider content-position="left">{{ $t('m.Programming_Detail') }}</el-divider>
                <el-card>
                  <h3>{{ getProgrammingProblemInfo(item.problemId).problem.title }}</h3>
                  <div class="problem-meta">
                    <el-tag size="small">{{ $t('m.Question_Id') }}: {{ getProgrammingProblemInfo(item.problemId).problem.problemId }}</el-tag>
                    <el-tag size="small" type="info">{{ $t('m.Time_Limit') }}: {{ getProgrammingProblemInfo(item.problemId).problem.timeLimit }}ms</el-tag>
                    <el-tag size="small" type="warning">{{ $t('m.Memory_Limit') }}: {{ getProgrammingProblemInfo(item.problemId).problem.memoryLimit }}MB</el-tag>
                    <el-tag size="small" type="success">{{ $t('m.Judge_Mode') }}: {{ getJudgeModeText(getProgrammingProblemInfo(item.problemId).problem.judgeMode) }}</el-tag>
                  </div>
                  <div class="problem-content">
                    <div class="content-section">
                      <h4>{{ $t('m.Problem_Description') }}</h4>
                      <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.description)" v-highlight></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.input">
                      <h4>{{ $t('m.Input_Format') }}</h4>
                      <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.input)" v-highlight></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.output">
                      <h4>{{ $t('m.Output_Format') }}</h4>
                      <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.output)" v-highlight></div>
                    </div>
                    <div class="content-section" v-if="getProgrammingExamplesForProblem(item.problemId).length > 0">
                      <h4>{{ $t('m.Samples') }}</h4>
                      <div v-for="(example, idx) in getProgrammingExamplesForProblem(item.problemId)" :key="idx" class="example-item">
                        <el-alert :title="`${$t('m.Sample')} ${idx + 1}`" type="info" :closable="false">
                          <div slot="default">
                            <p><strong>{{ $t('m.Input') }}:</strong></p>
                            <pre>{{ example.input }}</pre>
                            <p><strong>{{ $t('m.Output') }}:</strong></p>
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
              <p>{{ $t('m.Programming_Student_View_Hint') }}</p>
            </el-alert>

            <el-alert v-if="!item.problemId" type="warning" :closable="false">
              {{ $t('m.Programming_Not_Linked') }}
            </el-alert>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import { getClassroomProblem } from '@/common/classroomProblem'
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
          this.$message.error(this.$t('m.Preview_Data_Not_Found'))
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
        this.$message.error(this.$t('m.Load_Preview_Failed'))
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
          const res = await getClassroomProblem({
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
    parseOptions(optionsInput) {
      if (!optionsInput) return []
      try {
        const options = Array.isArray(optionsInput) ? optionsInput : JSON.parse(optionsInput)
        if (!Array.isArray(options)) return []
        return options.map((opt, idx) => ({
          letter: String.fromCharCode(65 + idx),
          text: String(opt || '').replace(/^\s*[A-Za-z][\.\)、:：]\s*/, '').trim()
        }))
      } catch (e) {
        return []
      }
    },
    parseCompositeSubQuestions(optionsInput) {
      if (!optionsInput) return []
      try {
        const parsed = Array.isArray(optionsInput) ? optionsInput : JSON.parse(optionsInput)
        if (!Array.isArray(parsed)) return []
        return parsed.map((subQuestion, index) => ({
          id: String(subQuestion.id || `sq_${index + 1}`),
          content: subQuestion.content || '',
          options: Array.isArray(subQuestion.options)
            ? subQuestion.options
            : (Array.isArray(subQuestion.choiceOptions) ? subQuestion.choiceOptions : []),
          score: Number(subQuestion.score || subQuestion.subScore || 0)
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
        fill_blank: this.$t('m.Fill_Blank'),
        composite: this.$t('m.Composite_Question'),
        subjective: this.$t('m.Subjective'),
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    getJudgeModeText(mode) {
      const modeMap = {
        'default': this.$t('m.Judge_Mode_Default'),
        'spj': this.$t('m.Judge_Mode_Spj'),
        'interactive': this.$t('m.Judge_Mode_Interactive'),
        'subtask': this.$t('m.Judge_Mode_Subtask')
      }
      return modeMap[mode] || mode || this.$t('m.Judge_Mode_Default')
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

.preview-container >>> .preview-markdown pre {
  margin-left: 0 !important;
  padding-left: 0 !important;
  text-indent: 0 !important;
}

.preview-container >>> .preview-markdown pre code,
.preview-container >>> .preview-markdown code.hljs {
  margin-left: 0 !important;
  padding-left: 0 !important;
  text-indent: 0 !important;
}

.preview-container >>> .preview-markdown p {
  text-indent: 0 !important;
}

.option-item {
  margin-bottom: 10px;
}

.composite-preview-list {
  margin-top: 12px;
  padding-left: 0;
}

.composite-preview-item {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
  background: #fff;
}

.composite-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
  color: #303133;
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
