<template>
  <div class="submission-detail">
    <div class="header">
      <h3>{{ $t('m.Student_Submission_Detail') }}</h3>
      <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-card>
      <div v-if="studentSubmission">
        <!-- 学生信息 -->
        <div class="student-info">
          <p><strong>{{ $t('m.Student') }}:</strong> <UserName :username="studentSubmission.studentName" /></p>
          <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(studentSubmission.submitTime) }}</p>
          <p><strong>{{ $t('m.Total_Score') }}:</strong> {{ studentSubmission.totalScore }}</p>
        </div>

        <el-divider></el-divider>

        <!-- 题目列表 -->
        <div v-for="(submit, index) in studentSubmission.questions" :key="submit.id" class="question-item">
          <!-- 编程题 -->
          <div v-if="submit.problemId" class="programming-question-item">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <el-tag type="warning" size="small">
                {{ $t('m.Programming') }}
              </el-tag>
              <span class="question-score">{{ submit.score || 0 }}分</span>
            </div>
            <div class="question-title">BingOJ 编程题 - {{ submit.problemId }}</div>

            <!-- 解析学生答案（JSON格式，包含code和language） -->
            <div v-if="submit.answer" class="programming-answer">
              <div v-if="parseProgrammingAnswer(submit.answer)" class="code-info">
                <p><strong>编程语言:</strong> {{ parseProgrammingAnswer(submit.answer).language }}</p>
                <el-divider></el-divider>
                <p><strong>学生代码:</strong></p>
                <div class="code-preview-wrapper">
                  <pre class="code-preview"><code :ref="`codeBlock_${index}`" :class="`language-${mapLanguage(parseProgrammingAnswer(submit.answer).language)}`">{{ parseProgrammingAnswer(submit.answer).code }}</code></pre>
                </div>
              </div>
            </div>

            <!-- 评测结果 -->
            <div v-if="submit.judgeResult" class="judge-result">
              <p><strong>评测结果:</strong></p>
              <el-tag :type="getJudgeResultType(submit.judgeResult)" size="small">
                {{ submit.judgeResult }}
              </el-tag>
            </div>

            <!-- 评分区域 - 编程题由系统自动评分，不显示手动评分按钮 -->
            <div class="grading-area">
              <el-tag :type="submit.isScored === 1 ? 'success' : 'warning'" size="small">
                {{ submit.isScored === 1 ? $t('m.Graded') : $t('m.Not_Graded') }}
              </el-tag>
              <span class="current-score">
                {{ $t('m.Current_Score') }}: {{ submit.score }}
              </span>
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                <i class="el-icon-info"></i> 编程题由系统自动评分
              </span>
            </div>
          </div>

          <!-- 普通题目 -->
          <div v-else-if="submit.question">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <el-tag :type="getQuestionTypeTag(submit.question.type)" size="small">
                {{ getQuestionTypeText(submit.question.type) }}
              </el-tag>
              <span class="question-score">{{ submit.question.score || 0 }}分</span>
            </div>

            <div class="question-title markdown-body" v-html="renderMarkdown(submit.question.title)"></div>
            <div v-if="submit.question.content" class="question-content markdown-body" v-html="renderMarkdown(submit.question.content)"></div>

            <!-- 显示题目选项（单选、多选） -->
            <div v-if="submit.question.type === 'single_choice' || submit.question.type === 'multiple_choice'" class="question-options">
              <div v-for="option in parseOptions(submit.question.options)" :key="option.letter" class="option-item">
                <div class="option-display">
                  <el-tag
                    :type="isOptionSelected(submit.answer, option.letter, submit.question.type) ? 'primary' : 'info'"
                    size="small"
                    effect="plain"
                  >
                    <span class="option-letter">{{ option.letter }}</span>
                  </el-tag>
                  <span v-html="renderMarkdown(option.text)" class="markdown-body option-text"></span>
                  <!-- 正确答案标识 -->
                  <el-tag v-if="isCorrectAnswer(submit.question.answer, option.letter, submit.question.type)"
                    type="success"
                    size="mini"
                    style="margin-left: 8px;">
                    ✓ {{ $t('m.Correct_Answer') }}
                  </el-tag>
                  <!-- 学生选择标识 -->
                  <el-tag v-if="isOptionSelected(submit.answer, option.letter, submit.question.type)"
                    type="primary"
                    size="mini"
                    style="margin-left: 4px;">
                    {{ $t('m.Selected') }}
                  </el-tag>
                </div>
              </div>
            </div>

            <!-- 判断题选项 -->
            <div v-if="submit.question.type === 'judge'" class="question-options">
              <div class="option-display">
                <el-tag
                  :type="parseJudgeAnswer(submit.answer) === true ? 'primary' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ $t('m.True') }}
                </el-tag>
                <!-- 正确答案标识 -->
                <el-tag v-if="parseJudgeAnswer(submit.question.answer) === true"
                  type="success"
                  size="mini"
                  style="margin-left: 8px;">
                  ✓ {{ $t('m.Correct_Answer') }}
                </el-tag>
                <!-- 学生选择标识 -->
                <el-tag v-if="parseJudgeAnswer(submit.answer) === true"
                  type="primary"
                  size="mini"
                  style="margin-left: 4px;">
                  {{ $t('m.Selected') }}
                </el-tag>
              </div>
              <div class="option-display">
                <el-tag
                  :type="parseJudgeAnswer(submit.answer) === false ? 'primary' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ $t('m.False') }}
                </el-tag>
                <!-- 正确答案标识 -->
                <el-tag v-if="parseJudgeAnswer(submit.question.answer) === false"
                  type="success"
                  size="mini"
                  style="margin-left: 8px;">
                  ✓ {{ $t('m.Correct_Answer') }}
                </el-tag>
                <!-- 学生选择标识 -->
                <el-tag v-if="parseJudgeAnswer(submit.answer) === false"
                  type="primary"
                  size="mini"
                  style="margin-left: 4px;">
                  {{ $t('m.Selected') }}
                </el-tag>
              </div>
            </div>

            <!-- 主观题答案 -->
            <div v-if="submit.question.type === 'subjective'" class="subjective-answer">
              <!-- 显示学生上传的图片 -->
              <div v-if="getAttachmentImages(submit.attachment).length > 0" class="attachment-images">
                <div class="attachment-title">
                  <i class="el-icon-picture"></i> 学生上传的图片：
                </div>
                <div class="attachment-list">
                  <el-image
                    v-for="(img, idx) in getAttachmentImages(submit.attachment)"
                    :key="idx"
                    :src="img"
                    :preview-src-list="getAttachmentImages(submit.attachment)"
                    fit="cover"
                    style="width: 120px; height: 120px; margin: 5px; border-radius: 4px;"
                  >
                  </el-image>
                </div>
              </div>

              <!-- 显示文字答案（如果有） -->
              <div v-if="submit.answer" class="answer-content">{{ submit.answer }}</div>
              <div v-else class="answer-content">{{ $t('m.No_Answer') }}</div>
            </div>

            <!-- 答案对比 -->
            <div class="answer-comparison">
              <p><strong>{{ $t('m.Correct_Answer') }}:</strong>
                <span v-if="submit.question.type === 'single_choice'">
                  {{ submit.question.answer }}
                </span>
                <span v-else-if="submit.question.type === 'judge'">
                  {{ parseJudgeAnswer(submit.question.answer) ? $t('m.True') : $t('m.False') }}
                </span>
                <span v-else-if="submit.question.type === 'multiple_choice'">
                  {{ parseMultipleChoiceAnswer(submit.question.answer) }}
                </span>
                <span v-else-if="submit.question.type === 'subjective'">
                  {{ submit.question.answer || '-' }}
                </span>
                <span v-else>
                  {{ submit.question.answer || '-' }}
                </span>
              </p>
              <!-- 主观题不显示学生答案（因为上面已经显示了） -->
              <p v-if="submit.question.type !== 'subjective'"><strong>{{ $t('m.Student_Answer') }}:</strong>
                <span v-if="submit.question.type === 'judge'">
                  <span v-if="!submit.answer || submit.answer === ''">{{ $t('m.No_Answer') }}</span>
                  <span v-else-if="parseJudgeAnswer(submit.answer) === true">{{ $t('m.True') }}</span>
                  <span v-else-if="parseJudgeAnswer(submit.answer) === false">{{ $t('m.False') }}</span>
                  <span v-else>{{ $t('m.No_Answer') }}</span>
                </span>
                <span v-else-if="submit.question.type === 'multiple_choice'">
                  <span v-if="!submit.answer || submit.answer === ''">{{ $t('m.No_Answer') }}</span>
                  <span v-else>{{ parseMultipleChoiceAnswer(submit.answer) }}</span>
                </span>
                <span v-else>
                  <span v-if="!submit.answer || submit.answer === ''">{{ $t('m.No_Answer') }}</span>
                  <span v-else>{{ submit.answer }}</span>
                </span>
              </p>
            </div>

            <!-- 评分区域 -->
            <div class="grading-area">
              <el-tag :type="submit.isScored === 1 ? 'success' : 'warning'" size="small">
                {{ submit.isScored === 1 ? $t('m.Graded') : $t('m.Not_Graded') }}
              </el-tag>
              <span class="current-score">
                {{ $t('m.Current_Score') }}: {{ submit.score }}
              </span>

              <!-- 操作按钮 -->
              <div class="action-buttons">
                <el-button size="small" type="primary" @click="openGradeDialog(submit)">
                  <i class="el-icon-edit"></i> {{ $t('m.Grade') }}
                </el-button>
                <el-button size="small" @click="recalculateScore(submit)">
                  <i class="el-icon-refresh"></i> {{ $t('m.Recalculate') }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 评分对话框 -->
    <el-dialog
      :title="$t('m.Manual_Grading')"
      :visible.sync="gradeDialogVisible"
      width="500px"
    >
      <el-form :model="gradeForm" label-width="100px">
        <el-form-item :label="$t('m.Question_Type')">
          <el-tag v-if="currentQuestion?.problemId" type="warning">{{ $t('m.Programming') }}</el-tag>
          <el-tag v-else :type="getQuestionTypeTag(currentQuestion?.question?.type)">
            {{ getQuestionTypeText(currentQuestion?.question?.type) }}
          </el-tag>
        </el-form-item>
        <el-form-item :label="$t('m.Question_Score')">
          <span>{{ getCurrentQuestionScore() }} {{ $t('m.Points') }}</span>
        </el-form-item>
        <el-form-item :label="$t('m.Enter_Score')">
          <el-input-number
            v-model="gradeForm.score"
            :min="0"
            :max="getCurrentQuestionScore()"
            :precision="1"
          ></el-input-number>
          <span style="margin-left: 10px; color: #909399;">
            (0 - {{ getCurrentQuestionScore() }})
          </span>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="gradeDialogVisible = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="submitGrade" :loading="grading">
          {{ $t('m.Confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// 配置 markdown-it 支持 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex)

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'StudentSubmissionDetail',
  components: {
    UserName
  },
  mixins: [realtimeSync, teacherAuth],
  data() {
    return {
      loading: false,
      homework: null,
      submissions: [],
      studentSubmission: null,
      gradeDialogVisible: false,
      currentQuestion: null,
      gradeForm: {
        score: 0
      },
      grading: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadData',
        immediate: true
      }
    }
  },
  mounted() {
    // 由 realtimeSync mixin 自动启动同步
  },
  methods: {
    async loadData() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = !this.homework || Object.keys(this.homework).length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const homeworkId = this.$route.params.homeworkId
        const studentUid = this.$route.query.uid

        // 并行加载作业详情和提交记录
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
            this.buildStudentSubmission(studentUid)
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    buildStudentSubmission(studentUid) {
      // 筛选该学生的所有提交记录
      const studentSubmits = this.submissions.filter(s => s.uid === studentUid)


      if (studentSubmits.length === 0) {
        this.$message.warning(this.$t('m.No_Submission_Found'))
        return
      }

      // 构建学生提交数据
      let totalScore = 0
      const questions = studentSubmits.map(submit => {
        totalScore += submit.score || 0
        return submit
      })

      this.studentSubmission = {
        uid: studentUid,
        studentName: studentSubmits[0].student?.username || studentSubmits[0].student?.realName || studentUid,
        submitTime: Math.max(...studentSubmits.map(s => new Date(s.createdAt).getTime())),
        totalScore: totalScore,
        questions: questions
      }

      // 应用代码高亮
      this.highlightAllCodeBlocks()
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        // 如果已经是数组对象格式 [{letter: 'A', text: '...'}]
        if (Array.isArray(options) && options.length > 0 && options[0].letter) {
          return options
        }
        // 如果是字符串数组格式 ["A. 选项1", "B. 选项2"]
        if (Array.isArray(options) && typeof options[0] === 'string') {
          return options.map((opt, idx) => {
            // 尝试匹配 "A. xxx" 或 "A、xxx" 等格式
            const match = opt.match(/^([A-Z])[\.\、]\s*(.*)$/)
            if (match) {
              return { letter: match[1], text: match[2] }
            }
            // 如果没有匹配到，生成字母（A, B, C...）
            return { letter: String.fromCharCode(65 + idx), text: opt }
          })
        }
        return []
      } catch (e) {
        console.error('解析选项失败:', e, optionsStr)
        return []
      }
    },
    parseMultipleChoiceAnswer(answer) {
      if (!answer) return ''
      try {
        const arr = JSON.parse(answer)
        return Array.isArray(arr) ? arr.sort().join(', ') : answer
      } catch (e) {
        return answer
      }
    },
    parseJudgeAnswer(answer) {
      // 处理判断题答案，返回布尔值或null
      // null 表示未作答，true/false 表示已作答
      // 支持多种格式：true/false, 1/0, 对/错, 正确/错误, True/False
      if (answer === '' || answer === null || answer === undefined) {
        return null // 未作答
      }
      if (answer === true || answer === 'true' || answer === 1 || answer === '1' || answer === '对' || answer === '正确' || answer === 'True' || answer === 'TRUE') {
        return true
      }
      if (answer === false || answer === 'false' || answer === 0 || answer === '0' || answer === '错' || answer === '错误' || answer === 'False' || answer === 'FALSE') {
        return false
      }
      // 未知格式，返回null表示未作答
      return null
    },
    // 渲染 Markdown
    renderMarkdown(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return content
      }
    },
    // 解析编程题答案（JSON格式，包含code和language）
    parseProgrammingAnswer(answer) {
      if (!answer) return null
      try {
        const parsed = JSON.parse(answer)
        return parsed
      } catch (e) {
        return null
      }
    },
    // 获取评测结果的标签类型
    getJudgeResultType(result) {
      if (!result) return 'info'
      const resultMap = {
        'Accepted': 'success',
        'AC': 'success',
        'Presentation Error': 'warning',
        'PE': 'warning',
        'Wrong Answer': 'danger',
        'WA': 'danger',
        'Time Limit Exceeded': 'warning',
        'TLE': 'warning',
        'Memory Limit Exceeded': 'warning',
        'MLE': 'warning',
        'Runtime Error': 'danger',
        'RE': 'danger',
        'Compilation Error': 'danger',
        'CE': 'danger'
      }
      return resultMap[result] || 'info'
    },
    isOptionSelected(answer, optionLetter, questionType) {
      if (questionType === 'single_choice' || questionType === 'judge') {
        return answer === optionLetter
      } else if (questionType === 'multiple_choice') {
        try {
          const selected = JSON.parse(answer || '[]')
          return selected.includes(optionLetter)
        } catch (e) {
          return false
        }
      }
      return false
    },
    isCorrectAnswer(correctAnswer, optionLetter, questionType) {
      if (questionType === 'single_choice' || questionType === 'judge') {
        return correctAnswer === optionLetter
      } else if (questionType === 'multiple_choice') {
        try {
          const correct = JSON.parse(correctAnswer || '[]')
          return correct.includes(optionLetter)
        } catch (e) {
          return false
        }
      }
      return false
    },
    openGradeDialog(submit) {
      this.currentQuestion = submit
      this.gradeForm.score = submit.score || 0
      this.gradeDialogVisible = true
    },
    async submitGrade() {
      // 验证分数范围
      const maxScore = this.currentQuestion?.question?.score || 0
      const scoreValue = Number(this.gradeForm.score) || 0 // 确保转换为数字，null/undefined 时默认为 0
      if (scoreValue < 0 || scoreValue > maxScore) {
        this.$message.error(`${this.$t('m.Score_Range_Error')}: 0 - ${maxScore}`)
        return
      }

      // 使用 homeworkQuestionId（后端返回的新字段）
      const questionId = this.currentQuestion.homeworkQuestionId || this.currentQuestion.questionId
      if (!questionId) {
        this.$message.error('无法获取题目ID')
        return
      }

      this.grading = true
      try {
        const res = await this.$store.dispatch('classroom/gradeHomework', {
          homeworkId: parseInt(this.$route.params.homeworkId),
          questionId: questionId,
          uid: this.currentQuestion.uid,
          score: scoreValue
        })

        if (res.code === 200) {
          this.$message.success(this.$t('m.Grade_Success'))
          this.gradeDialogVisible = false
          // 重新加载数据
          await this.loadData()
        } else {
          this.$message.error(res.message || this.$t('m.Grade_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Grade_Failed'))
      } finally {
        this.grading = false
      }
    },
    async recalculateScore(submit) {
      const questionType = submit.question?.type

      // 只有客观题才能重新计算
      if (questionType !== 'single_choice' && questionType !== 'multiple_choice' && questionType !== 'judge') {
        this.$message.warning(this.$t('m.Cannot_Recalculate_Subjective'))
        return
      }

      try {
        await this.$confirm(
          this.$t('m.Recalculate_Confirm'),
          this.$t('m.Tips'),
          {
            confirmButtonText: this.$t('m.Confirm'),
            cancelButtonText: this.$t('m.Cancel'),
            type: 'warning'
          }
        )

        // 调用后端重新评分接口
        const res = await this.$store.dispatch('classroom/recalculateScore', {
          homeworkId: parseInt(this.$route.params.homeworkId),
          questionId: submit.questionId,
          uid: submit.uid
        })

        if (res.code === 200) {
          this.$message.success(this.$t('m.Recalculate_Success'))
          await this.loadData()
        } else {
          this.$message.error(res.message || this.$t('m.Recalculate_Failed'))
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error(this.$t('m.Recalculate_Failed'))
        }
      }
    },
    goBack() {
      this.$router.back()
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
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
        subjective: 'info',
        programming: 'danger'
      }
      return map[type] || ''
    },
    getCurrentQuestionScore() {
      // 编程题：从 homeworkQuestionId 关联查询分数
      if (this.currentQuestion?.problemId) {
        // 需要从 homework.questions 中查找对应的分数
        const homeworkQuestion = this.homework?.questions?.find(q => {
          // 编程题匹配 problemId
          if (this.currentQuestion.problemId) {
            return q.problemId === this.currentQuestion.problemId
          }
          return false
        })
        return homeworkQuestion?.score || 0
      }
      // 普通题目：直接从 question 获取
      return this.currentQuestion?.question?.score || 0
    },
    // 解析附件字符串为图片URL数组
    getAttachmentImages(attachment) {
      if (!attachment) {
        return []
      }
      // 确保是字符串类型
      if (typeof attachment !== 'string') {
        return []
      }
      // attachment是用逗号分隔的URL字符串
      try {
        const urls = attachment.split(',').filter(url => url && url.trim())
        return urls
      } catch (e) {
        console.error('解析attachment失败:', e, attachment)
        return []
      }
    },
    // 高亮所有代码块
    highlightAllCodeBlocks() {
      this.$nextTick(() => {
        if (this.studentSubmission && this.studentSubmission.questions) {
          this.studentSubmission.questions.forEach((submit, index) => {
            if (submit.problemId && submit.answer && this.parseProgrammingAnswer(submit.answer)) {
              const refName = `codeBlock_${index}`
              const codeBlock = this.$refs[refName]
              if (codeBlock) {
                // 如果是数组（v-for产生的ref），取第一个元素
                const element = Array.isArray(codeBlock) ? codeBlock[0] : codeBlock
                if (element) {
                  hljs.highlightElement(element)
                }
              }
            }
          })
        }
      })
    },
    // 映射编程语言到 highlight.js 支持的语言标识
    mapLanguage(lang) {
      const languageMap = {
        // C语言变体
        'c': 'c',
        'C': 'c',
        'C With O2': 'c',

        // C++变体
        'cpp': 'cpp',
        'C++': 'cpp',
        'c++': 'cpp',
        'C++ 17 With O2': 'cpp',
        'C++ 17': 'cpp',
        'C++ 20 With O2': 'cpp',
        'C++ 20': 'cpp',

        // Java
        'java': 'java',
        'Java': 'java',

        // Python变体
        'python': 'python',
        'Python': 'python',
        'py': 'py',
        'python3': 'python',
        'Python3': 'python',
        'python2': 'python',
        'Python2': 'python',
        'pypy3': 'python',
        'PyPy3': 'python',
        'pypy2': 'python',
        'PyPy2': 'python',

        // Go
        'go': 'go',
        'golang': 'go',
        'Go': 'go',

        // Rust
        'rust': 'rust',
        'Rust': 'rust',

        // JavaScript变体
        'javascript': 'javascript',
        'js': 'javascript',
        'JavaScript': 'javascript',
        'javascript node': 'javascript',
        'JavaScript Node': 'javascript',
        'javascript v8': 'javascript',
        'JavaScript V8': 'javascript',

        // TypeScript
        'typescript': 'typescript',
        'ts': 'typescript',
        'TypeScript': 'typescript',

        // PHP
        'php': 'php',
        'PHP': 'php',

        // Ruby
        'ruby': 'ruby',
        'Ruby': 'ruby',

        // Kotlin
        'kotlin': 'kotlin',
        'Kotlin': 'kotlin',

        // Scala
        'scala': 'scala',
        'Scala': 'scala',

        // C#
        'csharp': 'csharp',
        'c#': 'csharp',
        'C#': 'csharp',
        'CSharp': 'csharp'
      }
      return languageMap[lang] || 'plaintext'
    }
  }
}
</script>

<style scoped>
.submission-detail {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.student-info {
  margin-bottom: 20px;
}

.student-info p {
  margin: 5px 0;
}

.question-item {
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.question-number {
  font-weight: bold;
  font-size: 16px;
}

.question-score {
  margin-left: auto;
  color: #409EFF;
  font-weight: bold;
}

.question-title {
  margin-bottom: 15px;
  font-size: 15px;
}

.question-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 15px;
}

.option-item {
  margin-bottom: 5px;
}

.subjective-answer {
  margin-bottom: 15px;
}

.answer-content {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  white-space: pre-wrap;
}

.answer-comparison {
  padding: 15px;
  background: #f0f9ff;
  border-radius: 4px;
  margin-bottom: 15px;
}

.answer-comparison p {
  margin: 5px 0;
}

.grading-area {
  display: flex;
  align-items: center;
  gap: 15px;
  padding-top: 15px;
  border-top: 1px solid #ebeef5;
}

.current-score {
  font-weight: bold;
  color: #409EFF;
}

.action-buttons {
  margin-left: auto;
}

/* 编程题样式 */
.programming-question-item {
  margin-bottom: 30px;
  padding: 20px;
  background: #fafafa;
  border-radius: 4px;
}

.programming-answer {
  margin-top: 15px;
}

.code-info {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
}

.code-preview-wrapper {
  max-height: 500px;
  overflow: auto;
  background: #282c34;
  border-radius: 4px;
}

.code-preview {
  margin: 0;
  padding: 20px;
  background: #282c34;
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #abb2bf;
}

.code-preview code {
  background: transparent !important;
  padding: 0 !important;
  display: block;
  white-space: pre;
}

.judge-result {
  margin-top: 15px;
  padding: 10px;
  background: #f0f9ff;
  border-radius: 4px;
}

/* 选项样式优化 */
.option-item {
  margin-bottom: 12px;
}

.option-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #F5F7FA;
  border-radius: 4px;
  border: 1px solid #E4E7ED;
  /* 移除 transition 避免轮询时闪烁 */
}

.option-display:hover {
  background: #ECF5FF;
  border-color: #409EFF;
}

.option-text {
  flex: 1;
  line-height: 1.6;
  color: #606266;
}

.question-options .option-letter {
  font-weight: bold;
  font-size: 14px;
  min-width: 24px;
}
</style>

<!-- 非scoped样式，确保markdown-body样式生效 -->
<style>
.submission-detail .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.submission-detail .markdown-body h1,
.submission-detail .markdown-body h2,
.submission-detail .markdown-body h3,
.submission-detail .markdown-body h4,
.submission-detail .markdown-body h5,
.submission-detail .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.submission-detail .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.submission-detail .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.submission-detail .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.submission-detail .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.submission-detail .markdown-body p {
  margin-bottom: 16px !important;
}

.submission-detail .markdown-body strong {
  font-weight: bold !important;
}

.submission-detail .markdown-body em {
  font-style: italic !important;
}

.submission-detail .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.submission-detail .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 附件图片样式 */
.attachment-images {
  margin-top: 15px;
  padding: 15px;
  background: #f0f9ff;
  border-radius: 4px;
  border: 1px solid #b3d8ff;
}

.attachment-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
  font-size: 14px;
}

.attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

/* Highlight.js 代码高亮全局样式 - 与查重代码一致 */
.code-preview .hljs {
  display: block;
  overflow-x: auto;
  padding: 0;
  background: #282c34;
  color: #abb2bf;
}

.code-preview .hljs-comment,
.code-preview .hljs-quote {
  color: #5c6370;
  font-style: italic;
}

.code-preview .hljs-keyword,
.code-preview .hljs-selector-tag,
.code-preview .hljs-subst {
  color: #c678dd;
}

.code-preview .hljs-number,
.code-preview .hljs-literal,
.code-preview .hljs-variable,
.code-preview .hljs-template-variable,
.code-preview .hljs-tag .hljs-attr {
  color: #d19a66;
}

.code-preview .hljs-string,
.code-preview .hljs-doctag {
  color: #98c379;
}

.code-preview .hljs-title,
.code-preview .hljs-section,
.code-preview .hljs-selector-id {
  color: #61afef;
}

.code-preview .hljs-type,
.code-preview .hljs-class .hljs-title {
  color: #e5c07b;
}

.code-preview .hljs-tag,
.code-preview .hljs-name,
.code-preview .hljs-attribute {
  color: #e06c75;
  font-weight: normal;
}

.code-preview .hljs-regexp,
.code-preview .hljs-link {
  color: #56b6c2;
}

.code-preview .hljs-symbol,
.code-preview .hljs-bullet {
  color: #61afef;
}

.code-preview .hljs-built_in,
.code-preview .hljs-builtin-name {
  color: #e6c07b;
}

.code-preview .hljs-meta {
  color: #61afef;
}

.code-preview .hljs-deletion {
  background: #f8756f;
}

.code-preview .hljs-addition {
  background: #98c379;
}

.code-preview .hljs-emphasis {
  font-style: italic;
}

.code-preview .hljs-strong {
  font-weight: bold;
}
</style>
