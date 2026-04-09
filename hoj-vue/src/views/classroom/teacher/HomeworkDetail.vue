<template>
  <div class="homework-detail">
    <div class="header">
      <h3>{{ homework.title || '作业详情' }}</h3>
      <div>
        <el-button v-if="isExamMode" type="warning" @click="viewExamMonitoring">
          <i class="el-icon-view"></i>
          <span>考试监控</span>
        </el-button>
        <el-button type="primary" @click="viewAnalysis">学情分析</el-button>
        <el-button @click="editHomework">{{ $t('m.Edit') }}</el-button>
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-card>
      <div v-if="homework.id">
        <p><strong>{{ $t('m.Homework_Title') }}:</strong> {{ homework.title }}</p>
        <p><strong>{{ $t('m.Description') }}:</strong> {{ homework.description || '-' }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(homework.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(homework.endTime) }}</p>

        <el-divider></el-divider>

        <h4>{{ $t('m.Student_Submissions') }}</h4>
        <div style="margin-bottom: 15px;">
          <el-button type="primary" icon="el-icon-document" @click="showPaperContentDialog = true">
            查看试卷内容
          </el-button>
          <el-button type="success" icon="el-icon-download" @click="exportHomeworkScores">
            导出成绩
          </el-button>
        </div>
        <el-table :data="studentSubmissions" stripe>
          <el-table-column :label="$t('m.Student_Name')" width="150">
            <template slot-scope="{ row }">
              {{ row.realName || '-' }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.System_Username')">
            <template slot-scope="{ row }">
              <UserName :username="row.studentName" />
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Submit_Time')">
            <template slot-scope="{ row }">{{ formatTime(row.submitTime) }}</template>
          </el-table-column>
          <el-table-column :label="$t('m.Progress')" width="120">
            <template slot-scope="{ row }">
              {{ row.completedCount }}/{{ row.totalCount }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Score')" width="100">
            <template slot-scope="{ row }">
              <span v-if="row.completedCount > 0">{{ row.totalScore }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Grading_Status')" width="120">
            <template slot-scope="{ row }">
              <el-tag v-if="isFullyGraded(row)" type="success" size="small">
                {{ $t('m.Graded') }}
              </el-tag>
              <el-tag v-else type="warning" size="small">
                {{ $t('m.Partial_Graded') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="违规" width="100">
            <template slot-scope="{ row }">
              <el-button
                v-if="hasViolations(row)"
                type="danger"
                size="mini"
                @click="viewViolations(row)"
                circle
                icon="el-icon-warning"
                :title="`有 ${getViolationCount(row)} 条违规记录`"
              >
              </el-button>
              <span v-else style="color: #909399; font-size: 12px;">无违规</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Operation')" width="150">
            <template slot-scope="{ row }">
              <el-button size="small" type="primary" @click="viewSubmission(row)">
                {{ $t('m.View_Detail') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog
      title="试卷内容"
      :visible.sync="showPaperContentDialog"
      width="900px"
      top="5vh"
    >
      <div v-if="homework.questions && homework.questions.length > 0" class="paper-content-preview">
        <div
          v-for="(item, index) in homework.questions"
          :key="item.id || item.problemId || index"
          class="paper-question-item"
        >
          <div class="paper-question-header">
            <span class="paper-question-index">{{ index + 1 }}.</span>
            <el-tag
              v-if="item.problemId"
              type="warning"
              size="small"
            >
              编程题
            </el-tag>
            <el-tag
              v-else-if="item.question"
              :type="getQuestionTypeTag(item.question.type)"
              size="small"
            >
              {{ getQuestionTypeText(item.question.type) }}
            </el-tag>
            <span class="paper-question-score">{{ item.score || 0 }}分</span>
          </div>

          <template v-if="item.question">
            <div class="markdown-body paper-question-title" v-html="renderMarkdown(item.question.title)" v-highlight></div>
            <div
              v-if="item.question.content"
              class="markdown-body paper-question-content"
              v-html="renderMarkdown(item.question.content)"
              v-highlight
            ></div>

            <div v-if="item.question.type === 'single_choice' || item.question.type === 'multiple_choice'" class="paper-question-options">
              <div v-for="(option, optionIndex) in parseOptions(item.question.options)" :key="optionIndex" class="paper-option-item">
                <span class="paper-option-label">{{ option.letter }}.</span>
                <span class="markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>

            <div v-else-if="item.question.type === 'judge'" class="paper-question-options">
              <div class="paper-option-item"><span class="paper-option-label">✓</span><span>正确</span></div>
              <div class="paper-option-item"><span class="paper-option-label">✗</span><span>错误</span></div>
            </div>

            <div v-else-if="item.question.type === 'composite'" class="paper-composite-list">
              <div
                v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(item.question.options)"
                :key="subQuestion.id || subIndex"
                class="paper-composite-item"
              >
                <div class="paper-composite-header">
                  <span>子题 {{ subIndex + 1 }}</span>
                  <span>{{ Number(subQuestion.score || 0) }}分</span>
                </div>
                <div class="markdown-body" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                <div class="paper-question-options">
                  <div
                    v-for="(option, optionIndex) in parseOptions(subQuestion.options)"
                    :key="optionIndex"
                    class="paper-option-item"
                  >
                    <span class="paper-option-label">{{ option.letter }}.</span>
                    <span class="markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
                <div class="paper-answer-box answer-info compact-answer-info">
                  <span class="paper-answer-label">正确答案：</span>
                  <span class="paper-answer-value">{{ getCompositeCorrectAnswer(item.question.answer, subQuestion.id, subIndex) }}</span>
                </div>
              </div>
            </div>

            <div v-else-if="item.question.type === 'fill_blank'" class="paper-hint-box">
              填空题，学生作答时填写文本答案
            </div>
            <div v-else-if="item.question.type === 'subjective'" class="paper-hint-box">
              主观题，学生作答时输入文字答案
            </div>

            <div
              v-if="['single_choice', 'multiple_choice', 'judge', 'fill_blank'].includes(item.question.type)"
              class="paper-answer-box answer-info compact-answer-info"
            >
              <span class="paper-answer-label">正确答案：</span>
              <span class="paper-answer-value">{{ formatAnswerForDisplay(item.question) }}</span>
            </div>
            <div v-else-if="item.question.type === 'subjective'" class="paper-answer-box answer-info compact-answer-info">
              <span class="paper-answer-label">参考答案：</span>
              <div
                v-if="normalizeTextValue(item.question.answer)"
                class="markdown-body paper-answer-markdown"
                v-html="renderMarkdown(item.question.answer)"
                v-highlight
              ></div>
              <span v-else class="paper-answer-value">暂无参考答案</span>
            </div>
          </template>

          <template v-else>
            <div class="paper-question-content">BingOJ 编程题 - {{ item.problemId }}</div>
            <div class="paper-answer-box answer-info compact-answer-info">
              <span class="paper-answer-label">参考答案：</span>
              <span class="paper-answer-value">请在题库中查看标准答案与解析</span>
            </div>
          </template>
        </div>
      </div>
      <el-empty v-else description="暂无试卷内容"></el-empty>
    </el-dialog>

    <!-- 违规记录对话框 -->
    <el-dialog
      title="学生违规记录"
      :visible.sync="violationsDialogVisible"
      width="600px"
    >
      <div v-if="currentStudent">
        <p><strong>学生：</strong>{{ currentStudent.realName || currentStudent.studentName }}</p>
        <el-divider></el-divider>
        <div v-if="currentStudentViolations && currentStudentViolations.length > 0">
          <el-timeline>
            <el-timeline-item
              v-for="(violation, index) in currentStudentViolations"
              :key="index"
              :timestamp="formatTime(violation.createdAt)"
              placement="top"
              :type="getViolationType(violation.violationType)"
            >
              <el-card>
                <h4>{{ getViolationTypeText(violation.violationType) }}</h4>
                <p v-if="violation.description">{{ violation.description }}</p>
                <p style="color: #909399; font-size: 12px;">
                  <i class="el-icon-time"></i> {{ formatTime(violation.createdAt) }}
                </p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </div>
        <div v-else>
          <el-empty description="暂无违规记录"></el-empty>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import teacherAuth from '@/mixins/teacherAuth'
import moment from 'moment'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'HomeworkDetail',
  components: {
    UserName
  },
  mixins: [teacherAuth, realtimeSync],
  data() {
    return {
      loading: false,
      homework: {},
      submissions: [],
      studentSubmissions: [], // 聚合后的学生提交数据
      // 违规记录相关
      violationsDialogVisible: false,
      currentStudent: null,
      currentStudentViolations: [],
      allViolationsMap: {}, // uid -> violations[]
      showPaperContentDialog: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadHomeworkDetail',
        immediate: true
      }
    }
  },
  mounted() {
    // 由 realtimeSync mixin 自动启动同步
  },
  computed: {
    isExamMode() {
      return this.homework && this.homework.isExamMode === 1
    },
    classroomId() {
      // 优先从 query 参数获取（适配 /admin/classroom 路由），没有才从 params 获取
      return this.$route.query.classroomId || this.$route.params.classroomId
    },
    homeworkId() {
      // 优先从 query 参数获取（适配 /admin/classroom 路由），没有才从 params 获取
      return this.$route.query.homeworkId || this.$route.params.homeworkId
    }
  },
  methods: {
    async loadHomeworkDetail() {
      // 避免重复请求
      if (this.loading) return

      // 使用 computed 属性获取 homeworkId（优先从 query 获取，适配 ClassroomAdmin 路由）
      const homeworkId = this.homeworkId
      if (!homeworkId) {
        console.error('homeworkId is undefined, route:', this.$route)
        return
      }

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = !this.homework || Object.keys(this.homework).length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const [homeworkRes, submissionsRes] = await Promise.all([
          this.$store.dispatch('classroom/getHomeworkDetail', homeworkId),
          this.$store.dispatch('classroom/getHomeworkSubmissions', homeworkId)
        ])

        if (homeworkRes.code === 200) {
          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentHomeworkString = JSON.stringify(this.homework)
          const newHomeworkString = JSON.stringify(homeworkRes.data)

          if (currentHomeworkString !== newHomeworkString) {
            // 数据真的变化了，才更新
            this.homework = homeworkRes.data

            // 重要：作业详情更新后（如添加了新题目），需要重新聚合学生提交数据
            // 这样 totalCount 会基于新的题目数量更新
            if (this.submissions && this.submissions.length > 0) {
              this.aggregateStudentSubmissions()
            }
          }
        }

        if (submissionsRes.code === 200) {
          const newSubmissions = submissionsRes.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentSubmissionsString = JSON.stringify(this.submissions)
          const newSubmissionsString = JSON.stringify(newSubmissions)

          if (currentSubmissionsString !== newSubmissionsString) {
            // 数据真的变化了，才更新
            this.submissions = newSubmissions
            this.aggregateStudentSubmissions()
          }
        }
      } catch (error) {
        console.error('加载作业详情失败:', error)
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    aggregateStudentSubmissions() {
      // 将按题目分组的提交数据聚合为学生维度
      const studentMap = new Map()

      // 获取作业的总题目数（用于计算完成进度）
      const totalQuestionsCount = this.homework.questions?.length || 0

      this.submissions.forEach(submit => {
        const uid = submit.uid
        if (!studentMap.has(uid)) {
          studentMap.set(uid, {
            uid: uid,
            studentName: submit.student?.username || submit.student?.realName || uid,
            realName: submit.realName || '', // 班级学生真实姓名
            submitTime: submit.createdAt,
            totalScore: 0,
            completedCount: 0,
            totalCount: totalQuestionsCount,  // 使用作业的题目总数
            questions: []
          })
        }

        const student = studentMap.get(uid)
        // 不再 totalCount++，因为已经在初始化时设置了正确的值
        student.questions.push(submit)

        // 只统计已正式提交的题目
        if (submit.isOfficiallySubmitted === 1) {
          student.totalScore += submit.score || 0
          student.completedCount++
        }

        // 更新提交时间（取最新）
        if (submit.createdAt > student.submitTime) {
          student.submitTime = submit.createdAt
        }
      })

      this.studentSubmissions = Array.from(studentMap.values())
    },
    viewSubmission(submission) {
      // 检查当前是否在管理员路由下
      const isAdminRoute = this.$route.path.startsWith('/admin/classroom')

      if (isAdminRoute) {
        // 在管理员路由下，使用命名路由
        this.$router.push({
          name: 'admin-student-submission-detail',
          params: {
            homeworkId: this.homeworkId,
            uid: submission.uid
          },
          query: {
            realName: submission.realName || ''
          }
        })
      } else {
        // 在教师端路由下，使用命名路由
        this.$router.push({
          name: 'StudentSubmissionDetail',
          params: { homeworkId: this.homeworkId },
          query: {
            uid: submission.uid,
            realName: submission.realName || ''
          }
        })
      }
    },
    editHomework() {
      // 检查当前是否在管理员路由下
      const isAdminRoute = this.$route.path.startsWith('/admin/classroom')

      if (isAdminRoute) {
        // 管理员路由：跳转到管理员端的编辑页面
        this.$router.push({
          name: 'admin-edit-homework',
          params: {
            classroomId: this.classroomId,
            homeworkId: this.homework.id
          },
          query: {
            isExamMode: this.homework.isExamMode || 0
          }
        })
      } else {
        // 教师路由：跳转到教师端的编辑页面
        this.$router.push({
          name: 'CreateHomework',
          params: {
            classroomId: this.classroomId
          },
          query: {
            editId: this.homework.id,
            isExamMode: this.homework.isExamMode || 0
          }
        })
      }
    },
    goBack() {
      // 检查当前是否在管理员路由下
      const isAdminRoute = this.$route.path.startsWith('/admin/classroom')

      if (isAdminRoute) {
        // 在管理员路由下，使用 query 参数返回，清除 homeworkId
        this.$router.push({
          query: {
            classroomId: this.classroomId,
            activeTab: 'homework'
          }
        })
      } else {
        // 在教师端路由下，使用命名路由
        // 注意：必须添加 tab=homework 参数，否则 ClassroomDetail 会使用默认的 students tab
        this.$router.push({
          name: 'TeacherHomework',
          params: { classroomId: this.classroomId },
          query: { tab: 'homework' }
        })
      }
    },
    viewAnalysis() {
      // 检查当前是否在管理员路由下
      const isAdminRoute = this.$route.path.startsWith('/admin/classroom')

      if (isAdminRoute) {
        // 在管理员路由下，跳转到管理员专用的分析页面
        this.$router.push({
          name: 'admin-classroom-homework-analysis',
          query: {
            classroomId: this.classroomId,
            homeworkId: this.homeworkId
          }
        })
      } else {
        // 在教师端路由下，使用命名路由
        this.$router.push({
          name: 'HomeworkAnalysis',
          params: {
            classroomId: this.classroomId,
            homeworkId: this.homeworkId
          }
        })
      }
    },
    viewExamMonitoring() {
      // 检查当前是否在管理员路由下
      const isAdminRoute = this.$route.path.startsWith('/admin/classroom')

      if (isAdminRoute) {
        // 在管理员路由下，跳转到管理员专用的监控页面
        this.$router.push({
          name: 'admin-classroom-exam-monitoring',
          query: {
            classroomId: this.classroomId,
            homeworkId: this.homeworkId
          }
        })
      } else {
        // 在教师端路由下，使用命名路由
        this.$router.push({
          name: 'ExamMonitoring',
          params: {
            classroomId: this.classroomId,
            homeworkId: this.homeworkId
          }
        })
      }
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        fill_blank: '填空题',
        composite: '组合题',
        subjective: this.$t('m.Subjective'),
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    getQuestionTypeTag(type) {
      const map = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        fill_blank: 'success',
        composite: 'danger',
        subjective: 'info'
      }
      return map[type] || ''
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
    parseCompositeAnswerMap(answerInput) {
      const parsed = this.parseMaybeSerializedJson(answerInput)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed
      }
      return {}
    },
    stripOptionPrefix(optionText, letterHint = '') {
      const normalizedText = String(optionText || '').trim()
      if (!normalizedText) return ''
      const escapedHint = letterHint ? letterHint.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') : '[A-Za-z]'
      const prefixRegex = new RegExp(`^\\s*(?:${escapedHint}|[A-Za-z])\\s*[\\.\\)、:：]\\s*`)
      return normalizedText.replace(prefixRegex, '').trim()
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
    formatAnswerForDisplay(question) {
      if (!question) return '-'
      switch (question.type) {
        case 'single_choice': {
          const value = String(this.parseMaybeSerializedJson(question.answer) || '').trim()
          return value || '-'
        }
        case 'multiple_choice': {
          const values = this.parseAnswerArray(question.answer)
          return values.length > 0 ? values.join('、') : '-'
        }
        case 'judge': {
          const raw = String(this.parseMaybeSerializedJson(question.answer) || '').trim().toLowerCase()
          if (['true', '1', 'yes', 'y', '正确'].includes(raw)) return '正确'
          if (['false', '0', 'no', 'n', '错误'].includes(raw)) return '错误'
          return '-'
        }
        case 'fill_blank': {
          const values = this.parseAnswerArray(question.answer, { allowCommaSplit: false })
          return values.length > 0 ? values.join(' / ') : '-'
        }
        default:
          return '-'
      }
    },
    renderMarkdown(content) {
      const normalized = this.normalizeTextValue(content)
      if (!normalized) return ''
      try {
        return md.render(normalized)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return normalized
      }
    },
    parseOptions(optionsInput) {
      if (!optionsInput) return []
      const parsed = this.parseMaybeSerializedJson(optionsInput)
      if (!Array.isArray(parsed)) return []
      return parsed.map((option, index) => {
        if (option && typeof option === 'object' && !Array.isArray(option)) {
          const normalizedLetter = String(option.letter || option.label || String.fromCharCode(65 + index))
            .trim()
            .replace(/[^A-Za-z0-9]/g, '')
            .toUpperCase()
          const letter = normalizedLetter || String.fromCharCode(65 + index)
          const rawText = option.text !== undefined
            ? option.text
            : (option.content !== undefined ? option.content : '')
          const normalizedText = this.normalizeTextValue(rawText)
          return {
            letter,
            text: this.stripOptionPrefix(normalizedText || String(rawText || ''), letter)
          }
        }
        const letter = String.fromCharCode(65 + index)
        const normalizedText = this.normalizeTextValue(option)
        return {
          letter,
          text: this.stripOptionPrefix(normalizedText || String(option || ''), letter)
        }
      })
    },
    parseCompositeSubQuestions(optionsInput) {
      if (!optionsInput) return []
      const parsed = this.parseMaybeSerializedJson(optionsInput)
      if (!Array.isArray(parsed)) return []
      return parsed.map((subQuestion, index) => {
        const optionSource = subQuestion && subQuestion.options !== undefined
          ? subQuestion.options
          : (subQuestion && subQuestion.choiceOptions !== undefined ? subQuestion.choiceOptions : [])
        const options = this.parseMaybeSerializedJson(optionSource)
        return {
          id: String((subQuestion && subQuestion.id) || `sq_${index + 1}`),
          content: (subQuestion && subQuestion.content) || '',
          options: Array.isArray(options) ? options : [],
          score: Number((subQuestion && (subQuestion.score || subQuestion.subScore)) || 0)
        }
      })
    },
    isFullyGraded(studentSubmission) {
      // 检查是否所有已提交的题目都已评分
      const submittedQuestions = studentSubmission.questions.filter(q => q.isOfficiallySubmitted === 1)
      if (submittedQuestions.length === 0) return false

      // 如果有主观题，需要等待教师手动评分
      const hasSubjective = submittedQuestions.some(q => {
        const questionType = q.question?.type
        return questionType === 'subjective' || questionType === 'programming'
      })

      if (hasSubjective) {
        // 检查主观题是否都已评分
        return submittedQuestions.every(q => {
          const questionType = q.question?.type
          if (questionType === 'subjective' || questionType === 'programming') {
            return q.isScored === 1
          }
          return true
        })
      }

      // 全是客观题，都应该是已评分状态
      return submittedQuestions.every(q => q.isScored === 1)
    },
    exportHomeworkScores() {
      if (!this.homework.questions || this.homework.questions.length === 0) {
        this.$message.warning('没有题目数据')
        return
      }

      // 创建CSV内容
      let csvContent = '\uFEFF' // UTF-8 BOM
      csvContent += `作业成绩_${this.homework.title}_${this.formatTime(this.homework.endTime)}\n\n`

      // 表头：真实姓名 + 系统用户名 + 每题得分
      const headers = ['真实姓名', '系统用户名']
      this.homework.questions.forEach((item, index) => {
        headers.push(`题目${index + 1}_${item.question?.title || ''}(${item.score}分)`)
      })
      headers.push('总分', '提交时间', '完成进度', '评分状态')
      csvContent += headers.map(h => `"${h}"`).join(',') + '\n'

      // 数据行
      this.studentSubmissions.forEach(student => {
        const row = [student.realName || '-', student.studentName]

        // 每题得分
        this.homework.questions.forEach(questionItem => {
          const submission = student.questions.find(q => q.questionId === questionItem.questionId)
          if (submission && submission.isOfficiallySubmitted === 1) {
            row.push(submission.score || 0)
          } else {
            row.push('-')
          }
        })

        // 总分
        const totalScore = student.completedCount > 0 ? student.totalScore : '-'
        row.push(totalScore)

        // 提交时间
        row.push(student.submitTime ? this.formatTime(student.submitTime) : '-')

        // 完成进度
        row.push(`${student.completedCount}/${student.totalCount}`)

        // 评分状态
        const gradingStatus = this.isFullyGraded(student) ? '已评分' : '部分评分'
        row.push(gradingStatus)

        csvContent += row.map(field => `"${field}"`).join(',') + '\n'
      })

      // 创建Blob并下载
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', `作业成绩_${this.homework.title}_${new Date().getTime()}.csv`)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    // ==================== 违规记录相关方法 ====================
    // 检查学生是否有违规记录
    hasViolations(studentSubmission) {
      if (!studentSubmission.questions || studentSubmission.questions.length === 0) {
        return false
      }
      // 检查是否有任何违规计数大于0
      return studentSubmission.questions.some(q => {
        const tabSwitchCount = q.tabSwitchCount || 0
        const fullscreenExitCount = q.fullscreenExitCount || 0
        const copyPasteAttemptCount = q.copyPasteAttemptCount || 0
        return tabSwitchCount > 0 || fullscreenExitCount > 0 || copyPasteAttemptCount > 0
      })
    },
    // 获取学生违规总数
    getViolationCount(studentSubmission) {
      if (!studentSubmission.questions || studentSubmission.questions.length === 0) {
        return 0
      }
      return studentSubmission.questions.reduce((total, q) => {
        const tabSwitchCount = q.tabSwitchCount || 0
        const fullscreenExitCount = q.fullscreenExitCount || 0
        const copyPasteAttemptCount = q.copyPasteAttemptCount || 0
        return total + tabSwitchCount + fullscreenExitCount + copyPasteAttemptCount
      }, 0)
    },
    // 查看违规记录
    async viewViolations(studentSubmission) {
      this.currentStudent = studentSubmission
      this.currentStudentViolations = []

      // 构建违规记录列表
      const violations = []
      if (studentSubmission.questions && studentSubmission.questions.length > 0) {
        studentSubmission.questions.forEach(q => {
          const tabSwitchCount = q.tabSwitchCount || 0
          const fullscreenExitCount = q.fullscreenExitCount || 0
          const copyPasteAttemptCount = q.copyPasteAttemptCount || 0

          // 添加违规记录
          for (let i = 0; i < tabSwitchCount; i++) {
            violations.push({
              violationType: 'tab_switch',
              description: '切换浏览器标签页',
              createdAt: q.createdAt
            })
          }
          for (let i = 0; i < fullscreenExitCount; i++) {
            violations.push({
              violationType: 'fullscreen_exit',
              description: '退出全屏模式',
              createdAt: q.createdAt
            })
          }
          for (let i = 0; i < copyPasteAttemptCount; i++) {
            violations.push({
              violationType: 'copy_paste_attempt',
              description: '尝试复制或粘贴',
              createdAt: q.createdAt
            })
          }
        })
      }

      this.currentStudentViolations = violations
      this.violationsDialogVisible = true
    },
    // 获取违规类型（用于 timeline 颜色）
    getViolationType(type) {
      const typeMap = {
        'tab_switch': 'warning',
        'fullscreen_exit': 'danger',
        'copy_paste_attempt': 'danger',
        'context_menu': 'warning',
        'devtools_attempt': 'danger'
      }
      return typeMap[type] || 'primary'
    },
    // 获取违规类型文本
    getViolationTypeText(type) {
      const typeMap = {
        'tab_switch': '切换标签页',
        'fullscreen_exit': '退出全屏',
        'copy_paste_attempt': '尝试复制/粘贴',
        'context_menu': '右键菜单',
        'devtools_attempt': '开发者工具'
      }
      return typeMap[type] || type
    }
  }
}
</script>

<style scoped>
.homework-detail {
  padding: 20px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.paper-content-preview {
  max-height: 70vh;
  overflow-y: auto;
}

.paper-question-item {
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 14px;
}

.paper-question-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.paper-question-index {
  font-weight: 700;
  color: #409EFF;
}

.paper-question-score {
  margin-left: auto;
  color: #E6A23C;
  font-weight: 600;
}

.paper-question-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
}

.paper-question-content {
  margin-bottom: 10px;
}

.paper-question-options {
  margin-top: 8px;
}

.paper-option-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 8px;
}

.paper-option-label {
  min-width: 20px;
  color: #409EFF;
  font-weight: 700;
}

.paper-composite-list {
  margin-top: 10px;
}

.paper-composite-item {
  padding: 10px;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  margin-bottom: 10px;
}

.paper-composite-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-weight: 600;
}

.paper-hint-box {
  margin-top: 8px;
  padding: 10px;
  background: #fdf6ec;
  border: 1px solid #faecd8;
  color: #8a6d3b;
  border-radius: 4px;
}

.paper-answer-box {
  margin-top: 10px;
  padding: 10px;
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
}

.paper-answer-label {
  color: #409EFF;
  font-weight: 600;
  margin-right: 6px;
}

.paper-answer-value {
  color: #303133;
}

.paper-answer-markdown {
  margin-top: 6px;
}

.paper-content-preview >>> .markdown-body pre {
  margin-left: 0 !important;
  text-indent: 0 !important;
  padding-left: 0 !important;
  padding: 10px 12px !important;
  overflow-x: auto !important;
}

.paper-content-preview >>> .markdown-body pre code,
.paper-content-preview >>> .markdown-body code.hljs {
  margin-left: 0 !important;
  padding-left: 0 !important;
  text-indent: 0 !important;
  display: block;
  white-space: pre !important;
}

.paper-content-preview >>> .markdown-body p {
  text-indent: 0 !important;
}
</style>
