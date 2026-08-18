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
          <p><strong>{{ $t('m.Student_Name') }}:</strong> {{ studentSubmission.realName || '-' }}</p>
          <p><strong>{{ $t('m.System_Username') }}:</strong> <UserName :username="studentSubmission.studentName" /></p>
          <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(studentSubmission.submitTime) }}</p>
          <p><strong>{{ $t('m.Total_Score') }}:</strong> {{ studentSubmission.totalScore }}</p>
        </div>

        <el-divider></el-divider>

        <!-- 题目列表 -->
        <div v-for="(submit, index) in studentSubmission.questions" :key="submit.id" class="question-item">
          <!-- 未提交的题目 -->
          <div v-if="submit._unsubmitted" class="unsubmitted-item">
            <!-- 编程题未提交 -->
            <div v-if="submit.problemId">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <el-tag type="warning" size="small">
                  {{ $t('m.Programming') }}
                </el-tag>
                <el-tag type="danger" size="small" effect="plain" style="margin-left: 8px;">
                  <i class="el-icon-warning-outline"></i> 未提交
                </el-tag>
                <span class="question-score">{{ submit.maxScore || 0 }}分</span>
              </div>
              <div class="question-title">BingOJ 编程题 - {{ submit.problemId }}</div>

              <!-- 未提交提示 -->
              <div class="programming-unsubmitted">
                <el-alert
                  type="warning"
                  :closable="false"
                  show-icon>
                  <template slot="title">
                    <span style="font-size: 14px;">该学生在作业发布后未提交此题目</span>
                  </template>
                </el-alert>
              </div>

              <!-- 评分区域 - 显示0分状态 -->
              <div class="grading-area">
                <el-tag type="warning" size="small">
                  {{ $t('m.Not_Graded') }}
                </el-tag>
                <span class="current-score">
                  {{ $t('m.Current_Score') }}: 0
                </span>
                <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                  <i class="el-icon-warning-outline"></i> 学生未提交，无法评分
                </span>
              </div>
            </div>

            <!-- 普通题目未提交 -->
            <div v-else-if="submit.question">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <el-tag :type="getQuestionTypeTag(submit.question.type)" size="small">
                  {{ getQuestionTypeText(submit.question.type) }}
                </el-tag>
                <el-tag type="danger" size="small" effect="plain" style="margin-left: 8px;">
                  <i class="el-icon-warning-outline"></i> 未提交
                </el-tag>
                <span class="question-score">{{ submit.maxScore || submit.question.score || 0 }}分</span>
              </div>

              <div class="question-title markdown-body" v-html="renderMarkdown(submit.question.title)" v-highlight></div>
              <div v-if="submit.question.content" class="question-content markdown-body" v-html="renderMarkdown(submit.question.content)" v-highlight></div>

              <!-- 显示题目选项（单选、多选） -->
              <div v-if="submit.question.type === 'single_choice' || submit.question.type === 'multiple_choice'" class="question-options">
                <div v-for="option in parseOptions(submit.question.options)" :key="option.letter" class="option-item">
                  <div class="option-display">
                    <el-tag type="info" size="small" effect="plain">
                      <span class="option-letter">{{ option.letter }}</span>
                    </el-tag>
                    <span v-html="renderMarkdown(option.text)" class="markdown-body option-text" v-highlight></span>
                    <!-- 正确答案标识 -->
                    <el-tag v-if="isCorrectAnswer(submit.question.answer, option.letter, submit.question.type)"
                      type="success"
                      size="mini"
                      style="margin-left: 8px;">
                      ✓ {{ $t('m.Correct_Answer') }}
                    </el-tag>
                  </div>
                </div>
              </div>

              <!-- 组合题子题（未提交） -->
              <div v-else-if="submit.question.type === 'composite'" class="question-options composite-answer-area">
                <div
                  v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(submit.question.options)"
                  :key="subQuestion.id || subIndex"
                  class="composite-sub-question-card"
                >
                  <div class="composite-sub-header">
                    <span>子题 {{ subIndex + 1 }}</span>
                    <span class="question-score">{{ Number(subQuestion.score || 0) }}分</span>
                  </div>
                  <div class="markdown-body" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>

                  <div
                    v-for="(option, optionIndex) in parseOptions(JSON.stringify(subQuestion.options || []))"
                    :key="`${subQuestion.id || subIndex}_${optionIndex}`"
                    class="option-item"
                  >
                    <div class="option-display">
                      <el-tag type="info" size="small" effect="plain">
                        <span class="option-letter">{{ option.letter }}</span>
                      </el-tag>
                      <span v-html="renderMarkdown(option.text)" class="markdown-body option-text" v-highlight></span>
                      <el-tag
                        v-if="getCompositeCorrectAnswer(submit.question.answer, subQuestion.id, subIndex) === option.letter"
                        type="success"
                        size="mini"
                        style="margin-left: 8px;"
                      >
                        ✓ {{ $t('m.Correct_Answer') }}
                      </el-tag>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 填空题显示答案 -->
              <div v-else-if="submit.question.type === 'fill_blank'" class="answer-display answer-info compact-answer-info">
                <p><strong>{{ $t('m.Correct_Answer') }}:</strong></p>
                <div v-html="renderMarkdown(formatFillBlankAnswer(submit.question.answer))" class="markdown-body" v-highlight></div>
              </div>

              <!-- 简答题显示答案 -->
              <div v-else-if="submit.question.type === 'essay' || submit.question.type === 'subjective'" class="answer-display answer-info compact-answer-info">
                <p><strong>{{ $t('m.Correct_Answer') }}:</strong></p>
                <div v-html="renderMarkdown(submit.question.answer || '')" class="markdown-body" v-highlight></div>
              </div>

              <!-- 未提交提示 -->
              <div class="unsubmitted-notice">
                <el-alert
                  type="warning"
                  :closable="false"
                  show-icon>
                  <template slot="title">
                    <span style="font-size: 14px;">该学生在作业发布后未提交此题目</span>
                  </template>
                </el-alert>
              </div>

              <!-- 评分区域 - 显示0分状态 -->
              <div class="grading-area">
                <el-tag type="warning" size="small">
                  {{ $t('m.Not_Graded') }}
                </el-tag>
                <span class="current-score">
                  {{ $t('m.Current_Score') }}: 0
                </span>
                <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                  <i class="el-icon-warning-outline"></i> 学生未提交，无法评分
                </span>
              </div>
            </div>
          </div>

          <!-- 编程题 -->
          <div v-else-if="submit.problemId" class="programming-question-item">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <el-tag type="warning" size="small">
                {{ $t('m.Programming') }}
              </el-tag>
              <span class="question-score">{{ submit.maxScore || 0 }}分</span>
            </div>
            <div class="question-title">BingOJ 编程题 - {{ submit.problemId }}</div>

            <!-- 解析学生答案（JSON格式，包含code和language） -->
            <div v-if="submit.answer" class="programming-answer">
              <div v-if="parseProgrammingAnswer(submit.answer)" class="code-info">
                <p><strong>编程语言:</strong> {{ parseProgrammingAnswer(submit.answer).language }}</p>
                <el-divider></el-divider>
                <p><strong>学生代码:</strong></p>
                <Highlight
                  :code="parseProgrammingAnswer(submit.answer).code"
                  :language="mapLanguage(parseProgrammingAnswer(submit.answer).language)"
                ></Highlight>
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
              <span class="question-score">{{ submit.maxScore || submit.question.score || 0 }}分</span>
            </div>

            <div class="question-title markdown-body" v-html="renderMarkdown(submit.question.title)" v-highlight></div>
            <div v-if="submit.question.content" class="question-content markdown-body" v-html="renderMarkdown(submit.question.content)" v-highlight></div>

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
                  <span v-html="renderMarkdown(option.text)" class="markdown-body option-text" v-highlight></span>
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

            <!-- 组合题子题（已提交） -->
            <div v-if="submit.question.type === 'composite'" class="question-options composite-answer-area">
              <div
                v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(submit.question.options)"
                :key="subQuestion.id || subIndex"
                class="composite-sub-question-card"
              >
                <div class="composite-sub-header">
                  <span>子题 {{ subIndex + 1 }}</span>
                  <span class="question-score">{{ Number(subQuestion.score || 0) }}分</span>
                </div>
                <div class="markdown-body" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>

                <div
                  v-for="(option, optionIndex) in parseOptions(JSON.stringify(subQuestion.options || []))"
                  :key="`${subQuestion.id || subIndex}_${optionIndex}`"
                  class="option-item"
                >
                  <div class="option-display">
                    <el-tag
                      :type="getCompositeStudentAnswer(submit.answer, subQuestion.id, subIndex) === option.letter ? 'primary' : 'info'"
                      size="small"
                      effect="plain"
                    >
                      <span class="option-letter">{{ option.letter }}</span>
                    </el-tag>
                    <span v-html="renderMarkdown(option.text)" class="markdown-body option-text" v-highlight></span>
                    <el-tag
                      v-if="getCompositeCorrectAnswer(submit.question.answer, subQuestion.id, subIndex) === option.letter"
                      type="success"
                      size="mini"
                      style="margin-left: 8px;"
                    >
                      ✓ {{ $t('m.Correct_Answer') }}
                    </el-tag>
                    <el-tag
                      v-if="getCompositeStudentAnswer(submit.answer, subQuestion.id, subIndex) === option.letter"
                      type="primary"
                      size="mini"
                      style="margin-left: 4px;"
                    >
                      {{ $t('m.Selected') }}
                    </el-tag>
                  </div>
                </div>

                <div class="answer-comparison answer-info compact-answer-info" style="margin-top: 10px;">
                  <p><strong>{{ $t('m.Correct_Answer') }}:</strong> {{ getCompositeCorrectAnswer(submit.question.answer, subQuestion.id, subIndex) || '-' }}</p>
                  <p><strong>{{ $t('m.Student_Answer') }}:</strong>
                    <span v-if="getCompositeStudentAnswer(submit.answer, subQuestion.id, subIndex)">
                      {{ getCompositeStudentAnswer(submit.answer, subQuestion.id, subIndex) }}
                    </span>
                    <span v-else>{{ $t('m.No_Answer') }}</span>
                  </p>
                </div>
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
            <div class="answer-comparison answer-info compact-answer-info">
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
                <span v-else-if="submit.question.type === 'fill_blank'">
                  {{ formatFillBlankAnswer(submit.question.answer) }}
                </span>
                <span v-else-if="submit.question.type === 'composite'">
                  {{ formatCompositeAnswerSummary(submit.question.answer, submit.question.options, '-') }}
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
                <span v-else-if="submit.question.type === 'fill_blank'">
                  <span v-if="!submit.answer || submit.answer === ''">{{ $t('m.No_Answer') }}</span>
                  <span v-else>{{ formatFillBlankAnswer(submit.answer) }}</span>
                </span>
                <span v-else-if="submit.question.type === 'composite'">
                  <span v-if="!submit.answer || submit.answer === ''">{{ $t('m.No_Answer') }}</span>
                  <span v-else>{{ formatCompositeAnswerSummary(submit.answer, submit.question.options, $t('m.No_Answer')) }}</span>
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
const Highlight = () => import('@/components/oj/common/Highlight')
import { addCodeBtn } from '@/common/codeblock'

// 配置 markdown-it 支持 KaTeX
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

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'StudentSubmissionDetail',
  components: {
    UserName,
    Highlight
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
    setTimeout(() => {
      this.$nextTick(() => {
        addCodeBtn()
      })
    }, 200)
  },
  watch: {
    homework(newVal, oldVal) {
      if (newVal !== oldVal) {
        this.$nextTick(() => {
          setTimeout(() => {
            addCodeBtn()
          }, 100)
        })
      }
    },
    questions(newVal, oldVal) {
      if (newVal !== oldVal) {
        this.$nextTick(() => {
          setTimeout(() => {
            addCodeBtn()
          }, 100)
        })
      }
    }
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
        // 兼容教师端（query）和管理员端（params）的路由参数
        const studentUid = this.$route.params.uid || this.$route.query.uid

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

            // 重要：作业详情更新后（如添加了新题目），需要重新构建学生提交数据
            // 否则新题目不会显示在学生提交详情中
            if (Array.isArray(this.submissions)) {
              const studentUid = this.$route.params.uid || this.$route.query.uid
              this.buildStudentSubmission(studentUid)
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

      if (!this.homework || !this.homework.questions || this.homework.questions.length === 0) {
        this.$message.warning('作业信息不完整')
        return
      }

      // 遍历作业的所有题目，确保显示所有题目（包括后来新增的）
      let totalScore = 0
      let latestSubmitTime = 0
      const questions = this.homework.questions.map(homeworkQuestion => {
        let submit = null

        // 根据题目类型查找提交记录
        if (homeworkQuestion.problemId) {
          // 编程题：通过 problemId 匹配
          submit = studentSubmits.find(s => s.problemId && s.problemId === homeworkQuestion.problemId)
        } else if (homeworkQuestion.question) {
          // 普通题目：通过 questionId 匹配
          submit = studentSubmits.find(s => s.questionId && s.questionId === homeworkQuestion.question.id)
        }

        if (submit) {
          // 有提交记录
          totalScore += submit.score || 0
          const submitTime = new Date(submit.createdAt).getTime()
          if (submitTime > latestSubmitTime) {
            latestSubmitTime = submitTime
          }
          // 返回提交记录，包含题目信息
          return {
            ...submit,
            // 确保题目信息存在（用于显示题目内容）
            question: submit.question || homeworkQuestion.question,
            problemId: submit.problemId || homeworkQuestion.problemId,
            // 添加题目满分（用于编程题显示）
            maxScore: homeworkQuestion.score || (submit.question?.score) || 0
          }
        } else {
          // 没有提交记录，创建一个空记录
          return {
            id: `unsubmitted-${homeworkQuestion.id}`,
            questionId: homeworkQuestion.question ? homeworkQuestion.question.id : null,
            homeworkId: this.homework.id,
            uid: studentUid,
            answer: null,
            score: 0,
            judgeResult: '未提交',
            isScored: 0,
            createdAt: null,
            updatedAt: null,
            // 包含题目信息以便显示
            question: homeworkQuestion.question,
            problemId: homeworkQuestion.problemId,
            // 添加题目满分
            maxScore: homeworkQuestion.score || (homeworkQuestion.question?.score) || 0,
            _unsubmitted: true // 标记为未提交
          }
        }
      })

      // 班级姓名只使用 realName；系统用户名只使用 student.username
      let studentName = studentUid
      let studentRealName = this.$route.query.realName || ''
      if (studentSubmits.length > 0) {
        const firstSubmit = studentSubmits[0]
        studentName = firstSubmit.student?.username || studentUid
        studentRealName = firstSubmit.realName || studentRealName
      }

      this.studentSubmission = {
        uid: studentUid,
        studentName: studentName,
        realName: studentRealName,
        submitTime: latestSubmitTime,
        totalScore: totalScore,
        questions: questions
      }

      // 应用代码高亮
      this.$nextTick(() => {
        this.highlightAllCodeBlocks()
      })
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
    parseCompositeSubQuestions(optionsInput) {
      if (!optionsInput) return []
      try {
        const parsed = Array.isArray(optionsInput)
          ? optionsInput
          : JSON.parse(optionsInput)
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
    parseCompositeAnswerMap(answerInput) {
      if (!answerInput) return {}
      try {
        const parsed = typeof answerInput === 'string' ? JSON.parse(answerInput) : answerInput
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          return parsed
        }
      } catch (e) {
        // ignore parse failure
      }
      return {}
    },
    getCompositeAnswerBySubQuestion(answerInput, subQuestionId, subIndex) {
      const answerMap = this.parseCompositeAnswerMap(answerInput)
      const candidates = [
        String(subQuestionId || ''),
        String(subIndex + 1),
        String(subIndex),
        `sub_${subIndex + 1}`,
        `sq_${subIndex + 1}`
      ]
      for (const key of candidates) {
        if (!key) continue
        if (Object.prototype.hasOwnProperty.call(answerMap, key)) {
          const value = String(answerMap[key] || '').trim()
          if (value) return value
        }
      }
      return ''
    },
    getCompositeCorrectAnswer(correctAnswerInput, subQuestionId, subIndex) {
      return this.getCompositeAnswerBySubQuestion(correctAnswerInput, subQuestionId, subIndex)
    },
    getCompositeStudentAnswer(studentAnswerInput, subQuestionId, subIndex) {
      return this.getCompositeAnswerBySubQuestion(studentAnswerInput, subQuestionId, subIndex)
    },
    formatCompositeAnswerSummary(answerInput, optionsInput, emptyText = '') {
      const subQuestions = this.parseCompositeSubQuestions(optionsInput)
      if (subQuestions.length === 0) {
        const fallbackMap = this.parseCompositeAnswerMap(answerInput)
        const fallbackEntries = Object.keys(fallbackMap).map(key => `${key}: ${String(fallbackMap[key] || '').trim()}`).filter(Boolean)
        return fallbackEntries.length > 0 ? fallbackEntries.join('； ') : emptyText
      }

      const segments = []
      subQuestions.forEach((subQuestion, index) => {
        const value = this.getCompositeAnswerBySubQuestion(answerInput, subQuestion.id, index)
        if (value) {
          segments.push(`子题${index + 1}: ${value}`)
        }
      })
      return segments.length > 0 ? segments.join('； ') : emptyText
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
      if (answer === '' || answer === null || answer === undefined) {
        return null // 未作答
      }
      const normalized = String(answer).toLowerCase().trim()
      if (normalized === 'true') {
        return true
      }
      if (normalized === 'false') {
        return false
      }
      // 未知格式，返回null表示未作答
      return null
    },
    formatFillBlankAnswer(answer) {
      if (!answer) return ''
      try {
        const parsed = typeof answer === 'string' ? JSON.parse(answer) : answer
        if (Array.isArray(parsed)) {
          return parsed.map(item => String(item || '').trim()).filter(Boolean).join(' / ')
        }
      } catch (e) {
        // fall through
      }
      return String(answer || '').trim()
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
    // 获取评测结果的标签类型（与 HOJ 官方状态码保持一致）
    getJudgeResultType(result) {
      if (!result) return 'info'

      // HOJ 官方状态码映射（参考 constants.js 中的 JUDGE_STATUS）
      const resultMap = {
        // 通过状态
        'AC': 'success',
        'Accepted': 'success',

        // 错误状态（红色）
        'WA': 'danger',
        'Wrong Answer': 'danger',
        'RE': 'danger',
        'Runtime Error': 'danger',
        'CE': 'danger',
        'Compilation Error': 'danger',

        // 警告状态（黄色）
        'TLE': 'warning',
        'Time Limit Exceeded': 'warning',
        'MLE': 'warning',
        'Memory Limit Exceeded': 'warning',
        'PE': 'warning',
        'Presentation Error': 'warning',

        // 系统状态（灰色/蓝色）
        'SE': 'info',
        'System Error': 'info',
        'SF': 'info',
        'Submitted Failed': 'info',

        // 蓝色状态
        'PAC': 'primary',
        'Partial Accepted': 'primary',

        // 其他状态
        'CA': 'info',
        'Cancelled': 'info',
        'SNR': 'info',
        'Submitted Unknown Result': 'info',

        // 兼容旧格式（如果存在）
        'NO': 'danger',  // BingOJ 的旧状态码 4
        'PC': 'primary'  // BingOJ 的旧状态码 8
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
      const maxScore = this.getCurrentQuestionScore()
      const scoreValue = Number(this.gradeForm.score) || 0 // 确保转换为数字，null/undefined 时默认为 0
      if (scoreValue < 0 || scoreValue > maxScore) {
        this.$message.error(`${this.$t('m.Score_Range_Error')}: 0 - ${maxScore}`)
        return
      }

      // Java 合并后的评分接口以具体提交记录为准，避免同一题多次提交时评错记录。
      const submissionId = this.currentQuestion.id
      if (!submissionId) {
        this.$message.error('无法获取提交记录ID')
        return
      }

      this.grading = true
      try {
        const res = await this.$store.dispatch('classroom/gradeHomework', {
          homeworkId: parseInt(this.$route.params.homeworkId),
          submissionId,
          questionId: this.currentQuestion.homeworkQuestionId || this.currentQuestion.questionId,
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
      if (questionType !== 'single_choice' && questionType !== 'multiple_choice' && questionType !== 'judge' && questionType !== 'fill_blank' && questionType !== 'composite') {
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
      // 题目满分属于作业-题目关系，新 Java 接口放在提交视图的 maxScore 外层。
      return this.currentQuestion?.maxScore || this.currentQuestion?.question?.score || 0
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

/* 未提交题目的样式 */
.unsubmitted-item {
  background-color: #fafafa;
  border-left: 4px solid #E6A23C;
  opacity: 0.9;
}

.unsubmitted-item:hover {
  opacity: 1;
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

.question-options.composite-answer-area {
  display: block;
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

/* 编程题未提交样式 */
.programming-unsubmitted {
  margin-top: 15px;
  margin-bottom: 15px;
}

.programming-unsubmitted .el-alert {
  border-radius: 4px;
}

.programming-unsubmitted .el-alert__title {
  line-height: 1.6;
}

/* 普通题目未提交提示 */
.unsubmitted-notice {
  margin-top: 20px;
  margin-bottom: 15px;
}

.unsubmitted-notice .el-alert {
  border-radius: 4px;
}

.unsubmitted-notice .el-alert__title {
  line-height: 1.6;
}


/* 选项样式优化 */
.option-item {
  margin-bottom: 12px;
}

.composite-sub-question-card {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #fff;
  margin-bottom: 14px;
}

.composite-sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
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

/* 答案显示区域 */
.answer-display {
  margin-top: 15px;
  padding: 15px;
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
}

.answer-display p {
  margin: 0 0 10px 0;
  font-weight: 600;
  color: #409EFF;
}

.answer-display div {
  margin-top: 8px;
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
  padding: 0 10px 0 40px !important;  /* 上下0，左侧40px给行号留空间 */
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

.submission-detail .markdown-body pre code {
  padding-left: 0 !important;
}

.submission-detail .markdown-body pre ol.pre-numbering li:before {
  width: 40px !important;
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
</style>
