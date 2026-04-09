<template>
  <div class="practice-detail-page" v-loading="loading">
    <el-card shadow="never" v-if="paper.id">
      <div class="header">
        <div>
          <h2 class="title">{{ paper.title }}</h2>
          <p class="desc">{{ paper.description || '暂无简介' }}</p>
          <div class="meta">
            <span>题目数量: {{ paper.questionCount || 0 }}</span>
            <span>总分: {{ paper.totalScore || 0 }}</span>
            <span>作者: {{ paper.creator ? paper.creator.username : '-' }}</span>
          </div>
        </div>
        <div class="actions">
          <el-button @click="$router.push({ name: 'PracticeList' })">返回列表</el-button>
          <el-button v-if="!started" type="primary" @click="startAnswer">开始作答</el-button>
        </div>
      </div>
    </el-card>

    <el-card v-if="paper.id && !started" class="start-tip" shadow="never">
      <el-alert
        type="info"
        :closable="false"
        title="点击“开始作答”后进入练习，每道题都可以单独点击“查看答案”查看标准答案。"
      ></el-alert>
    </el-card>

    <div v-if="paper.id && started" class="questions-wrap">
      <el-card shadow="never" class="global-status-card">
        <div class="global-status-top">
          <div class="status-item">
            <span class="status-label">当前题目</span>
            <span class="status-value">{{ currentQuestionIndex + 1 }} / {{ questions.length }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">已作答</span>
            <span class="status-value">{{ answeredCount }} / {{ questions.length }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">已查看答案</span>
            <span class="status-value">{{ viewedAnswerCount }}</span>
          </div>
        </div>
        <el-progress :percentage="progressPercent" :stroke-width="16"></el-progress>
        <div class="global-nav-actions">
          <el-button-group>
            <el-button size="small" icon="el-icon-arrow-left" :disabled="currentQuestionIndex === 0" @click="goPrevQuestion">
              上一题
            </el-button>
            <el-button size="small" :disabled="currentQuestionIndex >= questions.length - 1" @click="goNextQuestion">
              下一题
              <i class="el-icon-arrow-right el-icon--right"></i>
            </el-button>
          </el-button-group>
          <span class="current-text">第 {{ currentQuestionIndex + 1 }} 题</span>
        </div>
        <div class="question-index-list">
          <el-button
            v-for="(q, idx) in questions"
            :key="getQuestionKey(q, idx)"
            size="mini"
            :type="getQuestionButtonType(q, idx)"
            @click="jumpToQuestion(idx)"
          >
            {{ idx + 1 }}
          </el-button>
        </div>
      </el-card>

      <el-card v-if="currentQuestion" :key="getQuestionKey(currentQuestion, currentQuestionIndex)" class="question-card" shadow="never">
        <template v-if="currentQuestion">
        <div class="q-header">
          <div>
            <span class="q-index">{{ currentQuestionIndex + 1 }}.</span>
            <el-tag size="mini" :type="getTypeTag(currentQuestion.questionType)">{{ getTypeText(currentQuestion.questionType) }}</el-tag>
            <span class="q-score">{{ currentQuestion.score || 0 }} 分</span>
          </div>
          <div class="header-right-actions">
            <el-button size="mini" @click="toggleDone(currentQuestion, currentQuestionIndex)">
              {{ isQuestionAnswered(currentQuestion, currentQuestionIndex) ? '标记未作答' : '标记已作答' }}
            </el-button>
            <el-button size="mini" type="success" @click="toggleAnswer(currentQuestion, currentQuestionIndex)">
              {{ showAnswerMap[getQuestionKey(currentQuestion, currentQuestionIndex)] ? '收起答案' : '查看答案' }}
            </el-button>
          </div>
        </div>

        <template v-if="currentQuestion.questionType === 'programming'">
          <div class="q-content">
            <p>编程题编号: <strong>{{ currentQuestion.problemId || '-' }}</strong></p>
            <el-link v-if="currentQuestion.problemId" type="primary" :underline="false" @click="openProblem(currentQuestion.problemId)">
              前往题目页面
            </el-link>
          </div>
          <el-alert
            v-if="showAnswerMap[getQuestionKey(currentQuestion, currentQuestionIndex)]"
            type="warning"
            :closable="false"
            title="编程题标准答案请前往题目页面按测试数据验证。"
          ></el-alert>
        </template>

        <template v-else-if="currentQuestion.question">
          <div class="q-title markdown-body" v-html="renderMarkdown(currentQuestion.question.title || '')" v-highlight></div>
          <div class="q-content markdown-body" v-html="renderMarkdown(currentQuestion.question.content || '')" v-highlight></div>

          <div v-if="currentQuestion.question.type === 'single_choice'" class="options">
            <el-radio-group v-model="singleAnswers[currentQuestion.question.id]">
              <el-radio v-for="(opt, idx) in parseOptions(currentQuestion.question.options)" :key="idx" :label="opt.letter">
                <span class="option-rich-text">
                  <span class="option-letter option-head">{{ opt.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(opt.text)" v-highlight></span>
                </span>
              </el-radio>
            </el-radio-group>
          </div>

          <div v-if="currentQuestion.question.type === 'multiple_choice'" class="options">
            <el-checkbox-group v-model="multipleAnswers[currentQuestion.question.id]">
              <el-checkbox v-for="(opt, idx) in parseOptions(currentQuestion.question.options)" :key="idx" :label="opt.letter">
                <span class="option-rich-text">
                  <span class="option-letter option-head">{{ opt.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(opt.text)" v-highlight></span>
                </span>
              </el-checkbox>
            </el-checkbox-group>
          </div>

          <div v-if="currentQuestion.question.type === 'judge'" class="options">
            <el-radio-group v-model="singleAnswers[currentQuestion.question.id]">
              <el-radio label="true">正确</el-radio>
              <el-radio label="false">错误</el-radio>
            </el-radio-group>
          </div>

          <div v-if="currentQuestion.question.type === 'fill_blank'" class="options">
            <el-input
              v-model="singleAnswers[currentQuestion.question.id]"
              type="textarea"
              :rows="3"
              placeholder="请输入你的填空答案"
            ></el-input>
          </div>

          <div v-if="currentQuestion.question.type === 'subjective'" class="options">
            <el-input
              v-model="singleAnswers[currentQuestion.question.id]"
              type="textarea"
              :rows="4"
              placeholder="请输入你的答案"
            ></el-input>
          </div>

          <div v-if="currentQuestion.question.type === 'composite'" class="options composite-options">
            <div
              v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(currentQuestion.question.options)"
              :key="subQuestion.id || subIndex"
              class="composite-sub-question"
            >
              <div class="composite-sub-header">
                <span>子题 {{ subIndex + 1 }}</span>
                <span v-if="subQuestion.score !== null && subQuestion.score !== undefined" class="composite-sub-score">
                  {{ subQuestion.score }} 分
                </span>
              </div>

              <div
                v-if="subQuestion.content"
                class="composite-sub-content markdown-body"
                v-html="renderMarkdown(subQuestion.content)"
                v-highlight
              ></div>
              <div v-else class="composite-sub-placeholder">暂无子题题干</div>

              <el-radio-group
                :value="getCompositeSelectedAnswer(currentQuestion.question.id, subQuestion.id)"
                @input="handleCompositeChoiceSelect(currentQuestion.question.id, subQuestion.id, $event)"
              >
                <el-radio
                  v-for="(opt, idx) in normalizeCompositeOptions(subQuestion.options)"
                  :key="`${subQuestion.id || subIndex}_${idx}`"
                  :label="opt.letter"
                >
                  <span class="option-rich-text">
                    <span class="option-letter option-head">{{ opt.letter }}.</span>
                    <span class="option-text markdown-body" v-html="renderMarkdown(opt.text)" v-highlight></span>
                  </span>
                </el-radio>
              </el-radio-group>
            </div>
          </div>

          <div v-if="showAnswerMap[getQuestionKey(currentQuestion, currentQuestionIndex)]" class="answer-box answer-info compact-answer-info">
            <div class="answer-title">标准答案</div>
            <div
              class="answer-content markdown-body"
              v-highlight
              v-html="
                renderMarkdown(
                  formatLoadedAnswer(
                    answerDataMap[getQuestionKey(currentQuestion, currentQuestionIndex)] &&
                      answerDataMap[getQuestionKey(currentQuestion, currentQuestionIndex)].answer,
                    currentQuestion.question.type,
                    currentQuestion.question
                  )
                )
              "
            ></div>
            <div
              v-if="
                answerDataMap[getQuestionKey(currentQuestion, currentQuestionIndex)] &&
                  answerDataMap[getQuestionKey(currentQuestion, currentQuestionIndex)].analysis
              "
              class="analysis markdown-body"
              v-highlight
              v-html="renderMarkdown(answerDataMap[getQuestionKey(currentQuestion, currentQuestionIndex)].analysis)"
            ></div>
          </div>
        </template>
        </template>
      </el-card>
    </div>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import classroomApi from '@/api/classroom'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false
})

export default {
  name: 'PracticeDetail',
  data() {
    return {
      loading: false,
      started: false,
      paper: {},
      currentQuestionIndex: 0,
      showAnswerMap: {},
      answerDataMap: {},
      manualDoneMap: {},
      singleAnswers: {},
      multipleAnswers: {},
      compositeAnswers: {}
    }
  },
  created() {
    this.loadPaperDetail()
  },
  computed: {
    questions() {
      return this.paper && Array.isArray(this.paper.questions) ? this.paper.questions : []
    },
    currentQuestion() {
      if (!this.questions.length) return null
      if (this.currentQuestionIndex < 0) return this.questions[0]
      if (this.currentQuestionIndex >= this.questions.length) return this.questions[this.questions.length - 1]
      return this.questions[this.currentQuestionIndex]
    },
    answeredCount() {
      return this.questions.filter((q, idx) => this.isQuestionAnswered(q, idx)).length
    },
    viewedAnswerCount() {
      return this.questions.filter((q, idx) => !!this.showAnswerMap[this.getQuestionKey(q, idx)]).length
    },
    progressPercent() {
      if (!this.questions.length) return 0
      return Number(((this.answeredCount / this.questions.length) * 100).toFixed(2))
    }
  },
  methods: {
    async loadPaperDetail() {
      this.loading = true
      try {
        const paperId = this.$route.params.paperId
        const res = await classroomApi.getPublicExamPaperDetail(paperId)
        if (res.data && res.data.code === 200) {
          this.paper = res.data.data || {}
          ;(this.paper.questions || []).forEach(item => {
            if (!item.question) return
            const qid = item.question.id
            if (item.question.type === 'multiple_choice' && !Array.isArray(this.multipleAnswers[qid])) {
              this.$set(this.multipleAnswers, qid, [])
            }
            if (
              item.question.type === 'composite' &&
              (!this.compositeAnswers[qid] ||
                typeof this.compositeAnswers[qid] !== 'object' ||
                Array.isArray(this.compositeAnswers[qid]))
            ) {
              this.$set(this.compositeAnswers, qid, {})
            }
            if (
              (item.question.type === 'single_choice' ||
                item.question.type === 'judge' ||
                item.question.type === 'fill_blank' ||
                item.question.type === 'subjective') &&
              this.singleAnswers[qid] === undefined
            ) {
              this.$set(this.singleAnswers, qid, '')
            }
          })
        } else {
          this.$message.error((res.data && res.data.message) || '加载练习详情失败')
          this.$router.push({ name: 'PracticeList' })
        }
      } catch (error) {
        this.$message.error('加载练习详情失败')
        this.$router.push({ name: 'PracticeList' })
      } finally {
        this.loading = false
      }
    },
    startAnswer() {
      this.started = true
      this.currentQuestionIndex = 0
      this.$message.success('已开始作答')
    },
    goPrevQuestion() {
      if (this.currentQuestionIndex <= 0) return
      this.currentQuestionIndex -= 1
    },
    goNextQuestion() {
      if (this.currentQuestionIndex >= this.questions.length - 1) return
      this.currentQuestionIndex += 1
    },
    jumpToQuestion(index) {
      if (index < 0 || index >= this.questions.length) return
      this.currentQuestionIndex = index
    },
    getQuestionKey(item, index) {
      if (item.questionId) return `q_${item.questionId}`
      if (item.problemId) return `p_${item.problemId}`
      return `idx_${index}`
    },
    isQuestionAnswered(item, index) {
      const key = this.getQuestionKey(item, index)
      if (this.manualDoneMap[key]) return true

      if (item.questionType === 'programming') return false
      if (!item.question || !item.question.id) return false

      const qid = item.question.id
      if (item.question.type === 'multiple_choice') {
        return Array.isArray(this.multipleAnswers[qid]) && this.multipleAnswers[qid].length > 0
      }
      if (item.question.type === 'composite') {
        const subQuestions = this.parseCompositeSubQuestions(item.question.options)
        if (!subQuestions.length) return false
        const subAnswers = this.compositeAnswers[qid] || {}
        return subQuestions.every(subQuestion => {
          const answer = subAnswers[String(subQuestion.id || '')]
          return answer !== undefined && answer !== null && String(answer).trim() !== ''
        })
      }
      const value = this.singleAnswers[qid]
      return value !== undefined && value !== null && String(value).trim() !== ''
    },
    toggleDone(item, index) {
      const key = this.getQuestionKey(item, index)
      this.$set(this.manualDoneMap, key, !this.manualDoneMap[key])
    },
    getQuestionButtonType(item, index) {
      if (index === this.currentQuestionIndex) return 'primary'
      return this.isQuestionAnswered(item, index) ? 'success' : 'info'
    },
    async toggleAnswer(item, index) {
      const key = this.getQuestionKey(item, index)
      const next = !this.showAnswerMap[key]
      this.$set(this.showAnswerMap, key, next)

      if (!next) return

      // 编程题无题库答案，维持前端提示
      if (item.questionType === 'programming') return
      if (!item.question || !item.question.id) return

      // 已加载过不重复请求
      if (this.answerDataMap[key]) return

      try {
        const res = await classroomApi.getPublicExamPaperQuestionAnswer(this.paper.id, item.question.id)
        if (res.data && res.data.code === 200) {
          this.$set(this.answerDataMap, key, res.data.data || {})
        } else {
          this.$message.error((res.data && res.data.message) || '加载答案失败')
        }
      } catch (error) {
        this.$message.error('加载答案失败')
      }
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        if (!Array.isArray(options)) return []
        return options.map((opt, index) => {
          const letter = String.fromCharCode(65 + index)
          const rawText = typeof opt === 'string' ? opt : String(opt || '')
          // 选项内容若已包含 "A."/"B、" 等前缀，去掉后再由前端统一渲染字母
          const normalizedText = rawText.replace(/^\s*[A-Za-z]\s*[\.\)、:：]\s*/, '')
          return {
            letter,
            text: normalizedText
          }
        })
      } catch (e) {
        return []
      }
    },
    parseCompositeSubQuestions(optionsStr) {
      if (!optionsStr) return []
      try {
        const parsed = typeof optionsStr === 'string' ? JSON.parse(optionsStr) : optionsStr
        if (!Array.isArray(parsed)) return []
        return parsed.map((item, index) => ({
          id: String((item && item.id) || `sub_${index + 1}`),
          content: item && item.content ? String(item.content) : '',
          options: Array.isArray(item && item.options) ? item.options : [],
          score:
            item && item.score !== undefined
              ? item.score
              : item && item.subScore !== undefined
                ? item.subScore
                : null
        }))
      } catch (e) {
        return []
      }
    },
    normalizeCompositeOptions(options) {
      if (!Array.isArray(options)) return []
      return options.map((opt, index) => {
        const letter = String.fromCharCode(65 + index)
        const rawText = typeof opt === 'string' ? opt : String(opt || '')
        const normalizedText = rawText.replace(/^\s*[A-Za-z]\s*[\.\)、:：]\s*/, '')
        return {
          letter,
          text: normalizedText
        }
      })
    },
    getCompositeSelectedAnswer(questionId, subQuestionId) {
      const subAnswers = this.compositeAnswers[questionId]
      if (!subAnswers || typeof subAnswers !== 'object') return ''
      return subAnswers[String(subQuestionId)] || ''
    },
    handleCompositeChoiceSelect(questionId, subQuestionId, answerLetter) {
      const current = this.compositeAnswers[questionId] && typeof this.compositeAnswers[questionId] === 'object'
        ? { ...this.compositeAnswers[questionId] }
        : {}
      current[String(subQuestionId)] = answerLetter
      this.$set(this.compositeAnswers, questionId, current)
    },
    getTypeTag(type) {
      const map = {
        single_choice: 'success',
        multiple_choice: 'warning',
        judge: 'info',
        fill_blank: 'success',
        composite: 'danger',
        subjective: 'primary',
        programming: 'danger'
      }
      return map[type] || 'info'
    },
    getTypeText(type) {
      const map = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        fill_blank: '填空题',
        composite: '组合题',
        subjective: '主观题',
        programming: '编程题'
      }
      return map[type] || type
    },
    formatLoadedAnswer(answer, questionType, question = null) {
      if (!answer) return '暂无标准答案'
      if (questionType === 'multiple_choice') {
        try {
          const arr = JSON.parse(answer)
          if (Array.isArray(arr)) return arr.join('、')
        } catch (e) {}
      }
      if (questionType === 'judge') {
        const normalized = String(answer).toLowerCase()
        if (['true', '正确', '对', '1'].includes(normalized)) return '正确'
        if (['false', '错误', '错', '0'].includes(normalized)) return '错误'
      }
      if (questionType === 'fill_blank') {
        try {
          const parsed = typeof answer === 'string' ? JSON.parse(answer) : answer
          if (Array.isArray(parsed)) {
            const lines = parsed.map(item => String(item || '').trim()).filter(Boolean)
            return lines.length ? lines.map(item => `- ${item}`).join('\n') : '暂无标准答案'
          }
        } catch (e) {
          // fall through
        }
      }
      if (questionType === 'composite') {
        try {
          const answerMap = typeof answer === 'string' ? JSON.parse(answer) : answer
          if (!answerMap || typeof answerMap !== 'object' || Array.isArray(answerMap)) {
            return String(answer)
          }

          const subQuestions = question ? this.parseCompositeSubQuestions(question.options) : []
          if (subQuestions.length > 0) {
            const lines = subQuestions.map((subQuestion, index) => {
              const subAnswer =
                answerMap[subQuestion.id] ||
                answerMap[String(index + 1)] ||
                answerMap[index] ||
                answerMap[`sub_${index + 1}`]
              return `- 子题 ${index + 1}: ${subAnswer || '-'}`
            })
            return lines.join('\n')
          }

          const lines = Object.keys(answerMap).map(key => `- ${key}: ${answerMap[key] || '-'}`)
          return lines.length ? lines.join('\n') : '暂无标准答案'
        } catch (e) {
          return String(answer)
        }
      }
      return answer
    },
    renderMarkdown(content) {
      if (!content) return ''
      return md.render(content)
    },
    openProblem(problemId) {
      window.open(`/problem/${problemId}`, '_blank')
    }
  }
}
</script>

<style scoped>
.practice-detail-page {
  max-width: 1100px;
  margin: 20px auto;
}

.header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
}

.title {
  margin: 0 0 8px;
  color: #303133;
}

.desc {
  color: #606266;
  margin: 0 0 10px;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: #909399;
  font-size: 13px;
}

.actions {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.start-tip {
  margin-top: 12px;
}

.questions-wrap {
  margin-top: 12px;
}

.global-status-card {
  margin-bottom: 12px;
}

.global-status-top {
  display: grid;
  grid-template-columns: repeat(3, minmax(120px, 1fr));
  gap: 10px;
  margin-bottom: 10px;
}

.status-item {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 10px;
}

.status-label {
  display: block;
  font-size: 12px;
  color: #909399;
}

.status-value {
  display: block;
  margin-top: 4px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.global-nav-actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.current-text {
  color: #606266;
  font-weight: 500;
}

.question-index-list {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.question-card {
  margin-bottom: 12px;
}

.q-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.header-right-actions {
  display: flex;
  gap: 8px;
}

.q-index {
  font-weight: 600;
  margin-right: 8px;
}

.q-score {
  margin-left: 10px;
  color: #e6a23c;
}

.q-title {
  font-size: 16px;
  margin-bottom: 8px;
}

.q-content {
  margin-bottom: 12px;
  color: #606266;
  line-height: 1.7;
}

.options {
  margin-bottom: 12px;
}

.options /deep/ .el-radio-group,
.options /deep/ .el-checkbox-group {
  display: block;
}

.options /deep/ .el-radio,
.options /deep/ .el-checkbox {
  display: flex;
  align-items: flex-start;
  margin: 0 0 10px 0;
  height: auto;
  white-space: normal;
  line-height: 1.6;
}

.options /deep/ .el-radio__input,
.options /deep/ .el-checkbox__input {
  margin-top: 3px;
}

.options /deep/ .el-radio__label,
.options /deep/ .el-checkbox__label {
  flex: 1;
  display: block;
  width: 100%;
  padding-left: 0;
  white-space: normal;
  line-height: 1.7;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.option-rich-text {
  display: block;
  width: 100%;
}

.option-letter {
  color: #606266;
  font-weight: 600;
}

.option-head {
  display: block;
  margin-bottom: 6px;
}

.option-text {
  display: block;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.options /deep/ .el-radio__label code,
.options /deep/ .el-checkbox__label code {
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.options /deep/ .option-text p {
  margin: 0;
}

.options /deep/ .option-text pre {
  margin: 0;
  max-width: 100%;
  overflow-x: auto;
}

.composite-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.composite-sub-question {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 10px;
  background: #fafbfd;
}

.composite-sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}

.composite-sub-score {
  color: #e6a23c;
  font-size: 12px;
  font-weight: 500;
}

.composite-sub-content {
  margin-bottom: 10px;
}

.composite-sub-placeholder {
  color: #909399;
  margin-bottom: 10px;
}

.practice-detail-page /deep/ .markdown-body pre {
  padding: 0 !important;
}

.practice-detail-page /deep/ .markdown-body pre code {
  margin: 0 !important;
  text-indent: 0 !important;
}

.practice-detail-page /deep/ .markdown-body pre ol.pre-numbering {
  display: none !important;
}

.answer-box {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
}

.answer-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.analysis {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #dcdfe6;
}

@media (max-width: 768px) {
  .global-status-top {
    grid-template-columns: 1fr;
  }

  .global-nav-actions {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
