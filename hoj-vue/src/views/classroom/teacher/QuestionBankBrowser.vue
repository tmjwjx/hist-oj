<template>
  <div class="question-bank-browser">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
        <h3>客观题题库</h3>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="addSelectedQuestions" :disabled="selectedQuestions.length === 0">
          <i class="el-icon-check"></i>
          已选 {{ selectedQuestions.length }} 题 - 添加到作业
        </el-button>
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

    <!-- 题目列表 - 使用 expandable table -->
    <el-card class="questions-table-card" shadow="hover" v-loading="loading">
      <el-table
        :data="questions"
        stripe
        @selection-change="handleSelectionChange"
        ref="questionTable"
      >
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="id" label="ID" width="90" align="center"></el-table-column>

        <el-table-column type="expand">
          <template slot-scope="{ row }">
            <div class="question-detail-expand">
              <!-- 基本信息 -->
              <el-descriptions :column="2" border size="small" class="detail-descriptions">
                <el-descriptions-item label="题型">
                  <el-tag :type="getQuestionTypeColor(row.type)" size="small">
                    {{ getQuestionTypeName(row.type) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="难度">
                  <el-rate :value="getDifficultyStars(row.difficulty)" :max="3" disabled />
                </el-descriptions-item>
                <el-descriptions-item label="分值">{{ row.score }} 分</el-descriptions-item>
                <el-descriptions-item label="共享">
                  <el-tag :type="row.isShared ? 'success' : 'info'" size="small">
                    {{ row.isShared ? '是' : '否' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="课程" :span="2">
                  <el-tag v-if="row.course" type="warning" size="small">{{ row.course }}</el-tag>
                  <span v-else style="color: #909399;">-</span>
                </el-descriptions-item>
                <el-descriptions-item label="标签" :span="2">
                  <el-tag
                    v-for="(tag, idx) in parseQuestionTags(row.tags)"
                    :key="idx"
                    size="mini"
                    type="info"
                    style="margin-right: 5px;"
                  >
                    {{ tag }}
                  </el-tag>
                  <span v-if="!row.tags || row.tags === '[]'" style="color: #909399;">-</span>
                </el-descriptions-item>
              </el-descriptions>

              <!-- 题目内容 -->
              <div class="detail-section">
                <h4 class="detail-label">题目标题</h4>
                <div
                  v-html="renderMarkdown(row.title)"
                  class="markdown-body detail-content detail-markdown"
                  v-highlight
                ></div>
              </div>

              <div v-if="row.content" class="detail-section">
                <h4 class="detail-label">题目描述</h4>
                <div
                  v-html="renderMarkdown(row.content)"
                  class="markdown-body detail-content detail-markdown"
                  v-highlight
                ></div>
              </div>

              <!-- 选择题选项 -->
              <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="detail-section">
                <h4 class="detail-label">选项</h4>
                <div class="options-display">
                  <div v-if="parseQuestionOptions(row.options).length > 0">
                    <div v-for="(option, index) in parseQuestionOptions(row.options)" :key="index" class="option-item">
                      <span class="option-label">{{ String.fromCharCode(65 + index) }}.</span>
                      <span
                        v-html="renderMarkdown(option)"
                        class="option-text markdown-body detail-markdown"
                        v-highlight
                      ></span>
                    </div>
                  </div>
                  <el-alert v-else type="info" :closable="false">暂无选项数据</el-alert>
                </div>
              </div>

              <!-- 组合题子题 -->
              <div v-if="row.type === 'composite'" class="detail-section">
                <h4 class="detail-label">组合题子题</h4>
                <div class="composite-list">
                  <div
                    v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(row.options)"
                    :key="subQuestion.id || subIndex"
                    class="composite-item"
                  >
                    <div class="composite-header">
                      <span>子题 {{ subIndex + 1 }}</span>
                      <span>{{ Number(subQuestion.score || 0) }}分</span>
                    </div>
                    <div
                      v-html="renderMarkdown(subQuestion.content || '')"
                      class="markdown-body detail-content detail-markdown"
                      v-highlight
                    ></div>
                    <div class="options-display composite-options">
                      <div v-if="parseQuestionOptions(subQuestion.options).length > 0">
                        <div
                          v-for="(option, optionIndex) in parseQuestionOptions(subQuestion.options)"
                          :key="optionIndex"
                          class="option-item"
                        >
                          <span class="option-label">{{ String.fromCharCode(65 + optionIndex) }}.</span>
                          <span
                            v-html="renderMarkdown(option)"
                            class="option-text markdown-body detail-markdown"
                            v-highlight
                          ></span>
                        </div>
                      </div>
                    </div>
                    <div class="composite-answer answer-info compact-answer-info">
                      <el-tag type="success" size="small">
                        正确答案：{{ getCompositeCorrectAnswer(row.answer, subQuestion.id, subIndex) }}
                      </el-tag>
                    </div>
                  </div>
                  <el-empty
                    v-if="parseCompositeSubQuestions(row.options).length === 0"
                    description="暂无组合题子题数据"
                    :image-size="90"
                  />
                </div>
              </div>

              <!-- 正确答案 -->
              <div class="detail-section">
                <h4 class="detail-label">正确答案</h4>
                <div class="answer-display answer-info compact-answer-info">
                  <el-tag v-if="row.type === 'single_choice'" type="success" size="small">
                    {{ parseSingleChoiceAnswer(row) }}
                  </el-tag>
                  <el-tag v-else-if="row.type === 'multiple_choice'" type="success" size="small">
                    {{ parseMultipleChoiceAnswer(row) }}
                  </el-tag>
                  <el-tag v-else-if="row.type === 'judge'" :type="isJudgeTrue(row.answer) ? 'success' : 'danger'" size="small">
                    {{ isJudgeTrue(row.answer) ? '正确' : '错误' }}
                  </el-tag>
                  <el-tag v-else-if="row.type === 'fill_blank'" type="success" size="small">
                    {{ parseFillBlankAnswer(row.answer) }}
                  </el-tag>
                  <el-tag v-else-if="row.type === 'composite'" type="success" size="small">
                    组合题答案见上方子题详情
                  </el-tag>
                  <div v-else-if="row.type === 'subjective'" class="subjective-answer">
                    <div
                      v-if="row.answer"
                      v-html="renderMarkdown(row.answer)"
                      class="markdown-body detail-markdown"
                      v-highlight
                    ></div>
                    <span v-else style="color: #909399;">暂无参考答案</span>
                  </div>
                </div>
              </div>

              <!-- 题目解析 -->
              <div v-if="row.analysis" class="detail-section">
                <el-divider content-position="left">
                  <i class="el-icon-document" style="color: #E6A23C;"></i>
                  <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
                </el-divider>
                <div
                  v-html="renderMarkdown(row.analysis)"
                  class="markdown-body detail-content detail-markdown"
                  v-highlight
                ></div>
              </div>

              <!-- 创建时间 -->
              <div class="detail-meta">
                <el-tag size="mini" type="info">创建：{{ formatTime(row.createdAt) }}</el-tag>
                <el-tag size="mini" type="info" style="margin-left: 10px;">更新：{{ formatTime(row.updatedAt) }}</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="题目标题" min-width="300">
          <template slot-scope="{ row }">
            <div v-html="renderMarkdown(row.title)" class="title-preview"></div>
          </template>
        </el-table-column>

        <el-table-column prop="type" label="题型" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="getQuestionTypeColor(row.type)" size="small">
              {{ getQuestionTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="difficulty" label="难度" width="120">
          <template slot-scope="{ row }">
            <el-rate :value="getDifficultyStars(row.difficulty)" :max="3" disabled />
          </template>
        </el-table-column>

        <el-table-column prop="course" label="课程" width="120">
          <template slot-scope="{ row }">
            <el-tag v-if="row.course" type="warning" size="small">{{ row.course }}</el-tag>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>

        <el-table-column prop="tags" label="标签" width="180">
          <template slot-scope="{ row }">
            <el-tag
              v-for="(tag, idx) in parseQuestionTags(row.tags)"
              :key="idx"
              size="mini"
              type="info"
              style="margin-right: 3px;"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!row.tags || row.tags === '[]'" style="color: #909399;">-</span>
          </template>
        </el-table-column>

        <el-table-column prop="score" label="分值" width="80">
          <template slot-scope="{ row }">
            <span style="font-weight: bold; color: #409EFF;">{{ row.score }}分</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="small" @click="toggleExpand(row)">
              {{ isExpanded(row) ? '收起' : '查看详情' }}
            </el-button>
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
      ],
      expandedRows: []
    }
  },
  mounted() {
    this.loadQuestions()
  },
  methods: {
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
    handleSelectionChange(selection) {
      this.selectedQuestions = selection
    },
    toggleExpand(row) {
      const table = this.$refs.questionTable
      table.toggleRowExpansion(row)

      // 更新展开状态
      const index = this.expandedRows.indexOf(row.id)
      if (index > -1) {
        this.expandedRows.splice(index, 1)
      } else {
        this.expandedRows.push(row.id)
      }
    },
    isExpanded(row) {
      return this.expandedRows.includes(row.id)
    },
    addSelectedQuestions() {
      if (this.selectedQuestions.length === 0) {
        this.$message.warning('请先选择题目')
        return
      }

      // 通过 store 传递选中的题目
      this.$store.commit('classroom/SET_SELECTED_QUESTIONS', this.selectedQuestions)
      this.$message.success(`已添加 ${this.selectedQuestions.length} 道题目`)

      // 返回上一页
      this.goBack()
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
    parseSingleChoiceAnswer(question) {
      const value = String(this.parseMaybeSerializedJson(question.answer) || '').trim()
      return value ? `答案：${value}` : '暂无答案'
    },
    parseMultipleChoiceAnswer(question) {
      const answers = this.parseAnswerArray(question.answer)
      if (answers.length > 0) {
        return `答案：${answers.join('、')}`
      }
      return '暂无答案'
    },
    isJudgeTrue(answer) {
      const raw = String(this.parseMaybeSerializedJson(answer) || '').trim().toLowerCase()
      return ['true', '1', 'yes', 'y', '正确'].includes(raw)
    },
    parseFillBlankAnswer(answer) {
      const values = this.parseAnswerArray(answer, { allowCommaSplit: false })
      if (values.length > 0) {
        return `答案：${values.join(' / ')}`
      }
      return '暂无答案'
    },
    formatTime(time) {
      if (!time) return '--'
      try {
        const date = new Date(time)
        return date.toLocaleString('zh-CN')
      } catch (e) {
        return []
      }
      return '--'
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

/* 展开的题目详情样式 */
.question-detail-expand {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.detail-descriptions {
  margin-bottom: 20px;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-label {
  color: #409EFF;
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 10px;
}

.detail-content {
  padding: 12px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
  border: 1px solid #e0e6ed;
}

.options-display {
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e0e6ed;
}

.composite-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.composite-item {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 12px;
  background: #fff;
}

.composite-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
  color: #303133;
}

.composite-options {
  margin-top: 8px;
}

.composite-answer {
  margin-top: 8px;
}

.option-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.option-item:last-child {
  border-bottom: none;
}

.option-label {
  font-weight: bold;
  color: #409EFF;
  margin-right: 10px;
  min-width: 25px;
}

.option-text {
  flex: 1;
  word-wrap: break-word;
}

.answer-display {
  padding: 12px;
  background: #f0f9ff;
  border-radius: 4px;
  border: 1px solid #b3d8ff;
}

.subjective-answer {
  padding: 12px;
  background: white;
  border-radius: 4px;
}

.detail-meta {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
  display: flex;
  align-items: center;
}

.title-preview {
  max-height: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>

<style>
/* Markdown 样式 */
.question-detail-expand .markdown-body h1,
.question-detail-expand .markdown-body h2,
.question-detail-expand .markdown-body h3 {
  margin-top: 1em;
  margin-bottom: 0.5em;
  font-weight: 600;
}

.question-detail-expand .markdown-body p {
  margin-bottom: 0.8em;
  line-height: 1.8;
}

.question-detail-expand .markdown-body code {
  background: #f8f8f9;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
}

.question-detail-expand .markdown-body pre {
  background: #f8f8f9;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
}

.question-detail-expand .detail-markdown pre {
  margin-left: 0 !important;
  text-indent: 0 !important;
  padding: 10px 12px !important;
  overflow-x: auto !important;
}

.question-detail-expand .detail-markdown pre code,
.question-detail-expand .detail-markdown code.hljs {
  margin-left: 0 !important;
  padding-left: 0 !important;
  text-indent: 0 !important;
  display: block;
  white-space: pre !important;
}

.question-detail-expand .detail-markdown p {
  text-indent: 0 !important;
}
</style>
