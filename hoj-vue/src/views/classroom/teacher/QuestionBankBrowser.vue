<template>
  <div class="question-bank-browser">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
        <h3>客观题题库</h3>
      </div>
      <div class="header-right">
        <el-tag type="success" effect="dark">已添加 {{ selectedQuestions.length }} 题（自动同步）</el-tag>
      </div>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card" shadow="hover">
      <el-row :gutter="15">
        <el-col :span="4">
          <el-select v-model="filters.type" placeholder="题型筛选" clearable size="small" @change="handleFilterChange">
            <el-option label="全部题型" value=""></el-option>
            <el-option label="单选题" value="single_choice"></el-option>
            <el-option label="多选题" value="multiple_choice"></el-option>
            <el-option label="判断题" value="judge"></el-option>
            <el-option label="填空题" value="fill_blank"></el-option>
            <el-option label="主观题" value="subjective"></el-option>
            <el-option label="组合题" value="composite"></el-option>
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.course" placeholder="课程筛选" clearable size="small" filterable @change="handleFilterChange">
            <el-option label="全部课程" value=""></el-option>
            <el-option
              v-for="course in commonCourses"
              :key="course"
              :label="course"
              :value="course"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.difficulty" placeholder="难度筛选" clearable size="small" @change="handleFilterChange">
            <el-option label="全部难度" value=""></el-option>
            <el-option label="简单" value="1"></el-option>
            <el-option label="中等" value="2"></el-option>
            <el-option label="困难" value="3"></el-option>
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.tag"
            placeholder="标签筛选"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.questionId"
            placeholder="搜索题目ID"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索标题"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
      </el-row>
    </el-card>

    <!-- 题目列表 -->
    <el-card class="questions-table-card" shadow="hover" v-loading="loading">
      <el-table
        :data="questions"
        class="question-bank-table"
        :row-class-name="getQuestionRowClass"
        stripe
      >
        <el-table-column prop="title" label="题目标题" min-width="300">
          <template slot-scope="{ row }">
            <div class="inline-question-cell">
              <div class="inline-meta-row">
                <el-tag type="info" size="mini">ID:{{ row.id }}</el-tag>
                <el-tag :type="getQuestionTypeColor(row.type)" size="mini">{{ getQuestionTypeName(row.type) }}</el-tag>
                <el-tag type="danger" size="mini">难度：{{ getDifficultyText(row.difficulty) }}</el-tag>
                <el-tag type="primary" size="mini">分值：{{ Number(row.score || 0) }}分</el-tag>
                <el-tag type="info" size="mini">创建者：{{ row.creator ? row.creator.username : (row.creatorId || '-') }}</el-tag>
                <el-tag type="info" size="mini">创建时间：{{ formatMetaTime(row.createTime || row.createdAt) }}</el-tag>
                <el-tag v-if="row.course" type="warning" size="mini">所属课程：{{ row.course }}</el-tag>
                <el-tag
                  v-for="(tag, idx) in parseQuestionTags(row.tags)"
                  :key="`meta-tag-${row.id}-${idx}`"
                  size="mini"
                  type="info"
                >
                  标签：{{ tag }}
                </el-tag>
                <el-tag :type="row.isShared ? 'success' : 'info'" size="mini">
                  开放权限：{{ row.isShared ? '共享' : '个人' }}
                </el-tag>
              </div>
              <div v-html="renderMarkdown(row.title)" class="inline-question-title markdown-body" v-highlight></div>
              <div
                v-if="row.content"
                v-html="renderMarkdown(row.content)"
                class="inline-question-content markdown-body"
                v-highlight
              ></div>

              <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="inline-options-list">
                <div
                  v-for="(option, index) in parseOptionEntries(row.options)"
                  :key="`opt-${row.id}-${index}`"
                  class="inline-option-item"
                >
                  <span class="inline-option-label">{{ option.letter }}.</span>
                  <span class="inline-option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                </div>
              </div>

              <div v-else-if="row.type === 'judge'" class="inline-options-list">
                <div class="inline-option-item">
                  <span class="inline-option-label">A.</span>
                  <span class="inline-option-text">正确</span>
                </div>
                <div class="inline-option-item">
                  <span class="inline-option-label">B.</span>
                  <span class="inline-option-text">错误</span>
                </div>
              </div>

              <div v-else-if="row.type === 'composite'" class="inline-composite-list">
                <div
                  v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(row.options)"
                  :key="subQuestion.id || subIndex"
                  class="inline-composite-item"
                >
                  <div class="inline-composite-head">子题 {{ subIndex + 1 }}（{{ Number(subQuestion.score || 0) }}分）</div>
                  <div
                    v-html="renderMarkdown(subQuestion.content || '')"
                    class="markdown-body inline-composite-content"
                    v-highlight
                  ></div>
                  <div class="inline-options-list">
                    <div
                      v-for="(option, optionIndex) in subQuestion.options"
                      :key="`subopt-${row.id}-${subIndex}-${optionIndex}`"
                      class="inline-option-item"
                    >
                      <span class="inline-option-label">{{ String.fromCharCode(65 + optionIndex) }}.</span>
                      <span class="inline-option-text markdown-body" v-html="renderMarkdown(option)" v-highlight></span>
                    </div>
                  </div>
                  <div class="answer-info compact-answer-info">
                    <strong>正确答案：</strong>
                    <el-tag type="success">{{ getCompositeCorrectAnswer(row.answer, subQuestion.id, subIndex) }}</el-tag>
                  </div>
                </div>
              </div>

              <div v-if="row.type !== 'composite'" class="inline-answer-row answer-info compact-answer-info">
                <span class="inline-answer-label">{{ row.type === 'subjective' ? '参考答案：' : '正确答案：' }}</span>
                <span v-if="row.type !== 'subjective'" class="inline-answer-text">{{ formatInlineAnswer(row) }}</span>
                <span
                  v-else-if="row.answer"
                  class="inline-answer-text markdown-body"
                  v-html="renderMarkdown(row.answer)"
                  v-highlight
                ></span>
                <span v-else class="inline-answer-text">暂无答案</span>
              </div>

              <div v-if="row.analysis" class="inline-analysis">
                <span class="inline-analysis-label">题目解析：</span>
                <div class="markdown-body" v-html="renderMarkdown(row.analysis)" v-highlight></div>
              </div>
              <div class="inline-actions-row">
                <el-button
                  v-if="isQuestionAdded(row)"
                  type="danger"
                  size="mini"
                  icon="el-icon-minus"
                  @click="removeQuestion(row)"
                >
                  移除题目
                </el-button>
                <el-button
                  v-else
                  type="primary"
                  size="mini"
                  icon="el-icon-plus"
                  @click="addQuestion(row)"
                >
                  添加题目
                </el-button>
              </div>
            </div>
          </template>
        </el-table-column>

      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
          :current-page="pagination.page"
          :page-sizes="[10, 20, 50, 100]"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
        >
        </el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(MarkdownItKatex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'QuestionBankBrowser',
  data() {
    return {
      loading: false,
      questions: [],
      selectedQuestions: [],
      filters: {
        type: '',
        course: '',
        difficulty: '',
        tag: '',
        questionId: '',
        keyword: ''
      },
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      commonCourses: [
        '数据结构',
        '算法设计与分析',
        '计算机网络',
        '操作系统',
        '计算机组成原理',
        '高等数学',
        '线性代数',
        '政治',
        '英语'
      ]
    }
  },
  mounted() {
    this.initSelectedQuestionsFromStore()
    this.loadQuestions()
  },
  methods: {
    initSelectedQuestionsFromStore() {
      const storedQuestions = this.$store.state.classroom.selectedQuestions || []
      this.selectedQuestions = Array.isArray(storedQuestions)
        ? storedQuestions.filter(question => question && question.id)
        : []
    },
    handleFilterChange() {
      this.pagination.page = 1
      this.loadQuestions()
    },
    async loadQuestions() {
      this.loading = true
      try {
        const questionId = String(this.filters.questionId || '').trim()
        const keyword = String(this.filters.keyword || '').trim()

        const params = {
          classroomId: this.$route.params.classroomId,
          page: this.pagination.page,
          limit: this.pagination.pageSize
        }

        if (this.filters.type) params.type = this.filters.type;
        if (this.filters.course) params.course = this.filters.course;
        if (this.filters.difficulty) params.difficulty = this.filters.difficulty;
        if (this.filters.tag) params.tag = this.filters.tag;
        if (questionId) {
          params.questionId = questionId
        } else if (keyword) {
          params.keyword = keyword
        }

        const res = await this.$store.dispatch('classroom/getQuestionBank', params)
        if (res.code === 200) {
          const questions = res.data.questions || res.data || []
          this.questions = questions.map(q => ({
            ...q,
            difficulty: parseInt(q.difficulty) || 2
          }))
          this.pagination.total = res.data.total || questions.length || 0
        }
      } catch (error) {
        console.error('加载题库失败:', error)
        this.$message.error('加载题库失败')
      } finally {
        this.loading = false
      }
    },
    getQuestionIdentity(question) {
      if (!question) return null
      const rawId = question.questionId !== undefined && question.questionId !== null ? question.questionId : question.id
      if (rawId === undefined || rawId === null || rawId === '') return null
      return String(rawId)
    },
    isQuestionAdded(question) {
      const identity = this.getQuestionIdentity(question)
      if (!identity) return false
      return this.selectedQuestions.some(item => this.getQuestionIdentity(item) === identity)
    },
    syncSelectedQuestionsToStore() {
      this.$store.commit('classroom/SET_SELECTED_QUESTIONS', this.selectedQuestions)
    },
    addQuestion(question) {
      if (this.isQuestionAdded(question)) {
        this.$message.warning('该题目已添加')
        return
      }

      const nextQuestion = {
        ...question,
        difficulty: parseInt(question.difficulty) || 2
      }
      this.selectedQuestions.push(nextQuestion)
      this.syncSelectedQuestionsToStore()
      this.$message.success('添加成功')
    },
    removeQuestion(question) {
      const identity = this.getQuestionIdentity(question)
      if (!identity) return
      const index = this.selectedQuestions.findIndex(item => this.getQuestionIdentity(item) === identity)
      if (index === -1) return
      this.selectedQuestions.splice(index, 1)
      this.syncSelectedQuestionsToStore()
      this.$message.success('已移除题目')
    },
    goBack() {
      this.$router.back()
    },
    handlePageChange(page) {
      this.pagination.page = page
      this.loadQuestions()
    },
    handleSizeChange(size) {
      this.pagination.pageSize = size
      this.pagination.page = 1
      this.loadQuestions()
    },
    getQuestionRowClass({ row }) {
      const normalizedType = String(row && row.type ? row.type : '').replace(/_/g, '-')
      const selectedClass = this.isQuestionAdded(row) ? 'is-added' : ''
      return `question-bank-row type-${normalizedType} ${selectedClass}`.trim()
    },
    parseMaybeSerializedJson(rawValue, maxDepth = 2) {
      if (rawValue === null || rawValue === undefined) return rawValue
      let current = rawValue
      for (let i = 0; i < maxDepth; i++) {
        if (typeof current !== 'string') break
        const trimmed = current.trim()
        if (!trimmed) return ''
        const looksLikeJson = (
          (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
          (trimmed.startsWith('[') && trimmed.endsWith(']')) ||
          (trimmed.startsWith('"') && trimmed.endsWith('"'))
        )
        if (!looksLikeJson) return trimmed
        try {
          current = JSON.parse(trimmed)
        } catch (e) {
          return trimmed
        }
      }
      return current
    },
    normalizeTextValue(rawValue) {
      const parsed = this.parseMaybeSerializedJson(rawValue)
      if (parsed === null || parsed === undefined) return ''
      if (typeof parsed === 'string') return parsed
      if (typeof parsed === 'number' || typeof parsed === 'boolean') return String(parsed)
      return ''
    },
    stripOptionPrefix(optionText, letterHint = '') {
      const normalizedText = String(optionText || '').trim()
      if (!normalizedText) return ''
      const escapedHint = letterHint ? letterHint.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') : '[A-Za-z]'
      const prefixRegex = new RegExp(`^\\s*(?:${escapedHint}|[A-Za-z])\\s*[\\.\\)、:：]\\s*`)
      return normalizedText.replace(prefixRegex, '').trim()
    },
    renderMarkdown(text) {
      const normalizedText = this.normalizeTextValue(text)
      if (!normalizedText) return ''
      return md.render(normalizedText)
    },
    getQuestionTypeName(type) {
      const typeMap = {
        'single_choice': '单选题',
        'multiple_choice': '多选题',
        'judge': '判断题',
        'fill_blank': '填空题',
        'subjective': '主观题',
        'composite': '组合题'
      }
      return typeMap[type] || type
    },
    getQuestionTypeColor(type) {
      const colorMap = {
        'single_choice': 'primary',
        'multiple_choice': 'success',
        'judge': 'warning',
        'fill_blank': 'success',
        'subjective': 'info',
        'composite': 'danger'
      }
      return colorMap[type] || ''
    },
    getDifficultyStars(difficulty) {
      return parseInt(difficulty) || 1
    },
    getDifficultyText(difficulty) {
      const level = Number(difficulty)
      if (level === 1) return '简单'
      if (level === 2) return '中等'
      if (level === 3) return '困难'
      return `等级${difficulty}`
    },
    formatMetaTime(value) {
      if (!value) return '--'
      try {
        return new Date(value).toLocaleString('zh-CN')
      } catch (e) {
        return '--'
      }
    },
    parseQuestionTags(tags) {
      const parsed = this.parseMaybeSerializedJson(tags)
      if (!Array.isArray(parsed)) return []
      return parsed.map(tag => String(tag || '').trim()).filter(Boolean)
    },
    parseQuestionOptions(options) {
      if (!options) return []
      const optionsArray = this.parseMaybeSerializedJson(options)
      if (!Array.isArray(optionsArray)) return []
      return optionsArray.map((opt, idx) => {
        if (opt && typeof opt === 'object' && !Array.isArray(opt)) {
          const normalizedLetter = String(opt.letter || opt.label || String.fromCharCode(65 + idx))
            .trim()
            .replace(/[^A-Za-z0-9]/g, '')
            .toUpperCase()
          const letter = normalizedLetter || String.fromCharCode(65 + idx)
          const rawText = opt.text !== undefined
            ? opt.text
            : (opt.content !== undefined ? opt.content : '')
          return this.stripOptionPrefix(rawText, letter)
        }
        return this.stripOptionPrefix(opt, String.fromCharCode(65 + idx))
      })
    },
    parseOptionEntries(optionsInput) {
      const parsedOptions = this.parseQuestionOptions(optionsInput)
      return parsedOptions.map((text, idx) => ({
        letter: String.fromCharCode(65 + idx),
        text
      }))
    },
    parseAnswerArray(answerInput, { allowCommaSplit = true } = {}) {
      const parsed = this.parseMaybeSerializedJson(answerInput)
      if (Array.isArray(parsed)) {
        return parsed.map(item => String(item || '').trim()).filter(Boolean)
      }
      const raw = String(parsed || '').trim()
      if (!raw) return []
      if (allowCommaSplit && (raw.includes(',') || raw.includes('，'))) {
        return raw.split(/[，,]/).map(item => item.trim()).filter(Boolean)
      }
      return [raw]
    },
    parseCompositeSubQuestions(optionsInput) {
      const parsed = this.parseMaybeSerializedJson(optionsInput)
      if (!Array.isArray(parsed)) return []
      return parsed.map((subQuestion, index) => {
        const optionSource = subQuestion && subQuestion.options !== undefined
          ? subQuestion.options
          : (subQuestion && subQuestion.choiceOptions !== undefined ? subQuestion.choiceOptions : [])
        return {
          id: String((subQuestion && subQuestion.id) || `sq_${index + 1}`),
          content: (subQuestion && subQuestion.content) || '',
          options: this.parseQuestionOptions(optionSource),
          score: Number((subQuestion && (subQuestion.score || subQuestion.subScore)) || 0)
        }
      })
    },
    parseCompositeAnswerMap(answerInput) {
      const parsed = this.parseMaybeSerializedJson(answerInput)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed
      }
      return {}
    },
    getCompositeCorrectAnswer(answerInput, subQuestionId, subIndex) {
      const answerMap = this.parseCompositeAnswerMap(answerInput)
      const candidates = [
        String(subQuestionId || ''),
        String(subIndex + 1),
        String(subIndex),
        `sub_${subIndex + 1}`
      ]
      for (const key of candidates) {
        if (key && Object.prototype.hasOwnProperty.call(answerMap, key)) {
          const value = String(answerMap[key] || '').trim()
          if (value) return value
        }
      }
      return '-'
    },
    isJudgeTrue(answer) {
      const raw = String(this.parseMaybeSerializedJson(answer) || '').trim().toLowerCase()
      return ['true', '1', 'yes', 'y', '正确'].includes(raw)
    },
    formatInlineAnswer(row) {
      if (row.type === 'single_choice') {
        const value = String(this.parseMaybeSerializedJson(row.answer) || '').trim()
        return value || '-'
      }
      if (row.type === 'multiple_choice') {
        const answers = this.parseAnswerArray(row.answer)
        return answers.length > 0 ? answers.join('、') : '-'
      }
      if (row.type === 'judge') {
        return this.isJudgeTrue(row.answer) ? '正确' : '错误'
      }
      if (row.type === 'fill_blank') {
        const values = this.parseAnswerArray(row.answer, { allowCommaSplit: false })
        return values.length > 0 ? values.join(' / ') : '-'
      }
      return '-'
    }
  }
}
</script>

<style scoped>
.question-bank-browser {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 15px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.header-left h3 {
  margin: 0;
  color: #409EFF;
}

.filter-card {
  margin-bottom: 20px;
}

.questions-table-card {
  min-height: 500px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.question-bank-browser >>> .question-bank-table.el-table::before {
  height: 0;
}

.question-bank-browser >>> .question-bank-table .el-table__body-wrapper table {
  border-collapse: separate;
  border-spacing: 0 10px;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row > td {
  background: #fff;
  border-top: 1px solid #ebeef5;
  border-bottom: 1px solid #ebeef5;
  vertical-align: top;
  transition: background-color 0.2s ease;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row > td:first-child {
  border-left: 4px solid #dcdfe6;
  border-radius: 10px 0 0 10px;
  padding-left: 10px;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row > td:last-child {
  border-right: 1px solid #ebeef5;
  border-radius: 0 10px 10px 0;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row:hover > td {
  background: #f8fbff;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.is-added > td {
  background: #f0f9eb;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-single-choice > td:first-child {
  border-left-color: #409eff;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-multiple-choice > td:first-child {
  border-left-color: #67c23a;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-judge > td:first-child {
  border-left-color: #e6a23c;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-fill-blank > td:first-child {
  border-left-color: #67c23a;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-subjective > td:first-child {
  border-left-color: #909399;
}

.question-bank-browser >>> .question-bank-table .el-table__body tr.question-bank-row.type-composite > td:first-child {
  border-left-color: #f56c6c;
}

.inline-question-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.inline-actions-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.inline-meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: nowrap;
  overflow-x: auto;
  white-space: nowrap;
  padding-bottom: 2px;
}

.inline-meta-row::-webkit-scrollbar {
  height: 4px;
}

.inline-meta-row::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 2px;
}

.inline-question-title {
  font-weight: 600;
  color: #303133;
}

.inline-question-content {
  color: #606266;
}

.inline-options-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px 12px;
}

.inline-option-item {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  min-width: 0;
  padding: 4px 6px;
  border-radius: 4px;
  background: #f8fafc;
}

.inline-option-label {
  flex-shrink: 0;
  width: 20px;
  color: #409eff;
  font-weight: 600;
  line-height: 1.6;
}

.inline-option-text {
  flex: 1;
  min-width: 0;
  line-height: 1.6;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.inline-composite-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.inline-composite-item {
  padding: 8px;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
}

.inline-composite-head {
  font-weight: 600;
  color: #606266;
  margin-bottom: 6px;
}

.inline-composite-content {
  margin-bottom: 6px;
}

.inline-answer-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 4px;
  background: #f0f9ff;
  border: 1px solid #d9ecff;
}

.inline-answer-label {
  flex-shrink: 0;
  color: #409eff;
  font-weight: 600;
}

.inline-answer-text {
  flex: 1;
  min-width: 0;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.inline-analysis {
  padding: 6px 8px;
  border-radius: 4px;
  background: #fff9e6;
  border: 1px solid #faecd8;
}

.inline-analysis-label {
  display: inline-block;
  margin-bottom: 4px;
  color: #e6a23c;
  font-weight: 600;
}

.inline-question-cell >>> .inline-option-text p,
.inline-question-cell >>> .inline-question-content p,
.inline-question-cell >>> .inline-answer-text p {
  margin: 0;
  text-indent: 0 !important;
}

.inline-question-cell >>> .inline-option-text pre,
.inline-question-cell >>> .inline-question-content pre,
.inline-question-cell >>> .inline-answer-text pre {
  margin: 4px 0 0;
  max-width: 100%;
  overflow-x: auto;
}

.inline-question-cell >>> .markdown-body pre {
  margin-left: 0 !important;
  text-indent: 0 !important;
  padding: 10px 12px !important;
  overflow-x: auto !important;
}

.inline-question-cell >>> .markdown-body pre code,
.inline-question-cell >>> .markdown-body code.hljs {
  margin-left: 0 !important;
  text-indent: 0 !important;
  padding-left: 0 !important;
  display: block;
  white-space: pre !important;
}
</style>
