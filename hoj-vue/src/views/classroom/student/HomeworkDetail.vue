<template>
  <div class="homework-detail">
    <div class="header">
      <h3>{{ homework.title || '作业详情' }}</h3>
      <div class="header-actions">
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <!-- 自动保存提示 -->
    <div v-if="autoSaving" class="auto-save-tip">
      <i class="el-icon-loading"></i> 正在自动保存...
    </div>
    <div v-else-if="lastSaveTime" class="auto-save-tip success">
      <i class="el-icon-check"></i> 已于 {{ lastSaveTime }} 自动保存
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-card>
      <div v-if="homework.id">
        <p><strong>{{ $t('m.Homework_Title') }}:</strong> {{ homework.title }}</p>
        <p><strong>{{ $t('m.Description') }}:</strong> {{ homework.description || '-' }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(homework.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(homework.endTime) }}</p>

        <el-divider></el-divider>

        <!-- 试卷列表 - 根据教师设置控制是否显示 -->
        <div v-if="canViewHomework" class="questions-container">
          <h4>{{ $t('m.Questions') }}</h4>
          <div v-for="(item, index) in homework.questions" :key="item.id" class="question-item">
            <!-- 编程题：使用 problemId 判断 -->
            <div v-if="item.problemId" class="programming-question-wrapper">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <span class="question-type">(编程题)</span>
                <span class="question-score">{{ item.score }}分</span>
                <!-- 编程题作答状态 -->
                <el-tag
                  :type="programmingStatus[item.problemId] === 'submitted' ? 'success' : 'info'"
                  size="small"
                  style="margin-left: 10px"
                >
                  <i v-if="programmingStatus[item.problemId] === 'checking'" class="el-icon-loading"></i>
                  {{ programmingStatusText[item.problemId] || '未作答' }}
                </el-tag>
                <!-- 已提交后显示得分 -->
                <span v-if="canViewScore && questionScores[item.problemId] !== undefined" class="question-score-earned">
                  得分: <span :style="{ color: questionScores[item.problemId] === 0 ? '#F56C6C' : '#67C23A', fontWeight: 'bold' }">{{ questionScores[item.problemId] }}</span>
                  <span v-if="!questionIsScored[item.problemId]" style="color: #909399; font-size: 12px; margin-left: 5px;">(未评分)</span>
                </span>
              </div>
              <ProgrammingQuestion
                :problem-id="item.problemId"
                :question-id="item.id"
                :homework-id="homework.id"
                :can-view-homework="canViewHomework"
                :can-view-score="canViewScore"
                :can-view-answer="canViewAnswer"
                :is-submitted="isSubmitted"
              />
            </div>

            <!-- 普通题目：使用 question 判断 -->
            <div v-else-if="item.question">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <span class="question-type">({{ getQuestionTypeText(item.question.type) }})</span>
                <span class="question-score">{{ item.score }}分</span>
                <!-- 已提交后显示得分 -->
                <span v-if="canViewScore && questionScores[item.question.id] !== undefined" class="question-score-earned">
                  得分: <span :style="{ color: questionScores[item.question.id] === 0 ? '#F56C6C' : '#67C23A', fontWeight: 'bold' }">{{ questionScores[item.question.id] }}</span>
                  <span v-if="!questionIsScored[item.question.id]" style="color: #909399; font-size: 12px; margin-left: 5px;">(未评分)</span>
                </span>
              </div>
              <div class="question-title markdown-body" v-html="formatContent(item.question.title)"></div>
              <div class="question-content markdown-body" v-html="formatContent(item.question.content)"></div>

              <!-- 单选题选项 -->
              <div v-if="item.question.type === 'single_choice'" class="question-options">
                <div v-for="(option, idx) in parseOptions(item.question.options)" :key="idx" class="option-item">
                  <el-radio
                    v-model="answers[item.question.id]"
                    :label="option.letter"
                    @change="handleAnswerChange"
                  >
                    <span v-html="`${option.letter}. ${formatContent(option.text)}`" class="markdown-body"></span>
                  </el-radio>
                </div>
                <!-- 显示学生已选择的选项 -->
                <div v-if="answers[item.question.id]" class="student-answer">
                  <el-tag type="info">已选: {{ answers[item.question.id] }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ item.question.answer }}</el-tag>
                </div>
              </div>

              <!-- 多选题选项 -->
              <div v-if="item.question.type === 'multiple_choice'" class="question-options">
                <div v-for="(option, idx) in parseOptions(item.question.options)" :key="idx" class="option-item">
                  <el-checkbox
                    v-model="multipleAnswers[item.question.id]"
                    :label="option.letter"
                    @change="handleMultipleChoiceChange(item.question.id)"
                  >
                    <span v-html="`${option.letter}. ${formatContent(option.text)}`" class="markdown-body"></span>
                  </el-checkbox>
                </div>
                <!-- 显示学生已选择的选项（按字典序排列） -->
                <div v-if="multipleAnswers[item.question.id] && multipleAnswers[item.question.id].length > 0" class="student-answer">
                  <el-tag type="info">已选: {{ [...multipleAnswers[item.question.id]].sort().join(', ') }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ item.question.answer }}</el-tag>
                </div>
              </div>

              <!-- 判断题 -->
              <div v-if="item.question.type === 'judge'" class="question-options">
                <el-radio
                  v-model="answers[item.question.id]"
                  label="true"
                  @change="handleAnswerChange"
                >对</el-radio>
                <el-radio
                  v-model="answers[item.question.id]"
                  label="false"
                  @change="handleAnswerChange"
                >错</el-radio>
                <!-- 显示学生已选择的选项 -->
                <div v-if="answers[item.question.id]" class="student-answer">
                  <el-tag type="info">已选: {{ answers[item.question.id] === 'true' ? '对' : '错' }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ item.question.answer === 'true' || item.question.answer === '对' ? '对' : '错' }}</el-tag>
                </div>
              </div>

              <!-- 主观题 -->
              <div v-if="item.question.type === 'subjective'" class="subjective-answer">
                <el-input
                  v-model="answers[item.question.id]"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入你的答案"
                  @blur="handleAnswerChange"
                />
                <!-- 图片上传区域 -->
                <div v-if="!isSubmitted" class="image-upload-area">
                  <div class="upload-tip">
                    <i class="el-icon-picture-outline"></i> 支持上传图片作为答案（可选）
                  </div>
                  <el-upload
                    :action="uploadUrl"
                    :headers="uploadHeaders"
                    :on-success="(response, file, fileList) => handleUploadSuccess(response, file, fileList, item.question.id)"
                    :on-error="handleUploadError"
                    :on-remove="(file, fileList) => handleRemoveFile(file, fileList, item.question.id)"
                    :file-list="getImageList(item.question.id)"
                    :limit="5"
                    :on-exceed="handleExceed"
                    accept="image/*"
                    list-type="picture-card"
                    :class="{ 'hide-upload-btn': getImageList(item.question.id).length >= 5 }"
                  >
                    <i class="el-icon-plus"></i>
                  </el-upload>
                  <div class="upload-limit-tip">
                    <i class="el-icon-info"></i> 最多上传5张图片，每张图片不超过10MB
                  </div>
                </div>
                <!-- 已提交的图片显示 -->
                <div v-else-if="getSubmittedImages(item.question.id).length > 0" class="submitted-images">
                  <div class="submitted-images-title">
                    <i class="el-icon-picture"></i> 已提交的图片：
                  </div>
                  <div class="submitted-images-list">
                    <el-image
                      v-for="(img, idx) in getSubmittedImages(item.question.id)"
                      :key="idx"
                      :src="img"
                      :preview-src-list="getSubmittedImages(item.question.id)"
                      fit="cover"
                      style="width: 100px; height: 100px; margin: 5px;"
                    >
                    </el-image>
                  </div>
                </div>
                <!-- 显示参考答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="reference-answer">
                  <div class="reference-answer-title">
                    <i class="el-icon-document"></i> 参考答案：
                  </div>
                  <div v-if="item.question.answer" class="reference-answer-content markdown-body" v-html="formatContent(item.question.answer)"></div>
                  <div v-else class="reference-answer-empty">教师未设置答案</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <el-divider></el-divider>

        <!-- 操作区域 -->
        <div v-if="canSubmit" class="action-area">
          <el-button type="primary" @click="showSubmitConfirm" :loading="submitting">
            <i class="el-icon-upload"></i> 提交作业
          </el-button>
        </div>

        <!-- 已提交信息（只有正式提交后才显示） -->
        <div v-if="isSubmitted && submission">
          <!-- 当可以查看作业时显示详细信息 -->
          <el-alert v-if="canViewHomework" type="success" :closable="false">
            <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(submission.submitTime) }}</p>
            <!-- 根据教师设置决定是否显示分数 -->
            <div v-if="canViewScore">
              <!-- 有未评分题目时显示详细分数信息 -->
              <div v-if="submission.hasUngraded">
                <p><strong>{{ $t('m.Graded_Score') }}:</strong> {{ submission.gradedScore !== undefined ? submission.gradedScore : 0 }}</p>
                <p style="color: #909399; font-size: 12px;">
                  <i class="el-icon-info"></i>
                  {{ $t('m.Has_Ungraded_Questions_Tip') }}
                </p>
              </div>
              <!-- 所有题目都已评分 -->
              <p v-else><strong>{{ $t('m.Score') }}:</strong> <span :style="{ color: submission.score === 0 ? '#F56C6C' : '' }">{{ submission.score !== undefined ? submission.score : $t('m.Not_Graded') }}</span></p>
            </div>
            <p v-else><strong>{{ $t('m.Score') }}:</strong> {{ $t('m.Score_Hidden_Tip') }}</p>
          </el-alert>

          <!-- 当不能查看作业时，显示简化提示 -->
          <el-alert v-else type="info" :closable="false">
            <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(submission.submitTime) }}</p>
            <div v-if="canViewScore">
              <!-- 有未评分题目时显示详细分数信息 -->
              <div v-if="submission.hasUngraded">
                <p><strong>{{ $t('m.Graded_Score') }}:</strong> {{ submission.gradedScore !== undefined ? submission.gradedScore : 0 }}</p>
                <p style="color: #909399; font-size: 12px;">
                  <i class="el-icon-info"></i>
                  {{ $t('m.Has_Ungraded_Questions_Tip') }}
                </p>
              </div>
              <!-- 所有题目都已评分 -->
              <p v-else><strong>{{ $t('m.Score') }}:</strong> <span :style="{ color: submission.score === 0 ? '#F56C6C' : '' }">{{ submission.score !== undefined ? submission.score : $t('m.Not_Graded') }}</span></p>
            </div>
            <p v-else><strong>{{ $t('m.Score') }}:</strong> {{ $t('m.Score_Hidden_Tip') }}</p>
            <p style="margin-top: 10px; color: #909399;">
              <i class="el-icon-info"></i>
              {{ $t('m.Homework_Content_Hidden_Tip') }}
            </p>
          </el-alert>
        </div>
      </div>
    </el-card>

    <!-- 提交确认对话框 -->
    <el-dialog
      title="确认提交"
      :visible.sync="showConfirmDialog"
      width="400px"
    >
      <p>确定要提交作业吗？提交后将无法修改答案。</p>
      <span slot="footer">
        <el-button @click="showConfirmDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmSubmit" :loading="submitting">确认提交</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import ProgrammingQuestion from '@/components/classroom/ProgrammingQuestion.vue'
import realtimeSync from '@/mixins/realtimeSync'

// 配置 markdown-it 支持 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex)

export default {
  name: 'HomeworkDetail',
  mixins: [realtimeSync],
  components: {
    ProgrammingQuestion
  },
  data() {
    return {
      loading: false,
      homework: {},
      submission: null,
      submitting: false,
      saving: false,
      showConfirmDialog: false,
      autoSaving: false,
      lastSaveTime: '',
      hasUnsavedChanges: false,
      autoSaveTimer: null,
      isSubmitted: false, // 是否已正式提交（区别于草稿）
      answers: {}, // 单选题、判断题、主观题答案
      multipleAnswers: {}, // 多选题答案数组 { questionId: ['A', 'B'] }
      questionScores: {}, // 每题得分 { questionId: score }
      questionIsScored: {}, // 每题是否已评分 { questionId: boolean }
      programmingStatus: {}, // 编程题状态 { problemId: 'not_started' | 'checking' | 'submitted' }
      programmingStatusText: {}, // 编程题状态文本 { problemId: '未作答' | '检测中...' | '已作答' }
      pollingTimer: null, // 轮询定时器
      attachments: {}, // 每题的图片URL列表 { questionId: [url1, url2, ...] }
      uploadedImages: {}, // 已上传的图片文件列表 { questionId: [{name, url, uid}, ...] }
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadHomeworkDetail',
        immediate: true
      }
    }
  },
  computed: {
    canSubmit() {
      return this.homework.status === 2 && !this.isSubmitted
    },
    // 是否可以查看作业（题目和答案）
    canViewHomework() {
      // 已提交后，根据教师设置判断
      if (this.isSubmitted) {
        // 如果教师允许查看作业内容（showHomework === 1），则可以查看
        return this.homework.showHomework === 1
      }
      // 未提交时，作业进行中可以查看
      return this.homework.status === 2
    },
    // 是否可以查看分数
    canViewScore() {
      // 必须已提交且教师允许查看分数
      return this.isSubmitted && this.homework.showScore === 1
    },
    // 是否可以查看答案
    canViewAnswer() {
      // 必须已提交且教师允许查看答案
      return this.isSubmitted && this.homework.showAnswer === 1
    },
    // 图片上传URL
    uploadUrl() {
      return '/rating-api/api/classroom/homework/upload-attachment'
    },
    // 上传请求头
    uploadHeaders() {
      const token = this.$store.getters.token
      return {
        'Authorization': 'Bearer ' + token
      }
    }
  },
  mounted() {
    this.loadHomeworkDetail()
  },
  beforeDestroy() {
    // 清除自动保存定时器
    if (this.autoSaveTimer) {
      clearTimeout(this.autoSaveTimer)
    }
    // 清除轮询定时器
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer)
    }
    // 页面退出前总是保存所有答案（确保不丢失数据）
    this.saveDraftSync()
  },
  methods: {
    async loadHomeworkDetail() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = !this.homework || !this.homework.id
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const homeworkId = this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', homeworkId)
        if (res.code === 200) {
          // 安全措施：清空所有题目的答案和难度字段（防止前端泄露）
          // 注意：学生始终看不到难度，只有在已提交且教师允许时才能看到答案
          if (res.data.questions && Array.isArray(res.data.questions)) {
            res.data.questions.forEach(item => {
              if (item.question) {
                // 学生始终看不到难度
                if (item.question.difficulty) {
                  item.question.difficulty = 0
                }
                // 只有在未提交或教师不允许查看答案时，才清空答案字段
                if (!this.canViewAnswer && item.question.answer) {
                  item.question.answer = ''
                }
              }
            })
          }

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentHomeworkString = JSON.stringify(this.homework)
          const newHomeworkString = JSON.stringify(res.data)

          if (currentHomeworkString !== newHomeworkString) {
            // 数据真的变化了，才更新
            this.homework = res.data

            // 只在首次加载时初始化答案对象，避免覆盖用户正在输入的答案
            if (isFirstLoad) {
              // 初始化所有题目的答案对象
              this.homework.questions.forEach(item => {
                // 编程题跳过
                if (item.problemId) return

                // 普通题目处理
                if (!item.question) return

                const qid = item.question.id
                if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
                  // 使用 $set 确保响应式
                  this.$set(this.answers, qid, '')
                } else if (item.question.type === 'multiple_choice') {
                  this.$set(this.multipleAnswers, qid, [])
                }
              })
            }
          }

          // 加载已提交的答案
          await this.loadSubmission(isFirstLoad)

          // 只在首次加载时初始化答案
          if (isFirstLoad) {
            // 初始化答案（确保所有题目都有记录）
            await this.initializeAnswers()
          }

          // 初始化编程题状态检测
          this.initializeProgrammingStatus()
          // 启动轮询检测编程题提交状态
          this.startPollingProgrammingStatus()
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
    async loadSubmission(shouldRestoreAnswers = true) {
      try {
        const homeworkId = this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getStudentHomeworkDetail', homeworkId)

        if (res.code === 200 && res.data) {
          this.submission = res.data

          // 根据后端返回的 isOfficiallySubmitted 字段设置提交状态
          // 只要有正式提交（包括编程题提交），就设置为已提交
          if (res.data.isOfficiallySubmitted) {
            this.isSubmitted = true
          }

          // 只在首次加载或明确要求时恢复答案
          // 避免在用户正在输入时覆盖其输入
          if (shouldRestoreAnswers && res.data.answers) {
            const answersData = JSON.parse(res.data.answers || '{}')

            // 统计已作答的题目数
            let answeredCount = 0

            Object.keys(answersData).forEach(qid => {
              answeredCount++

              // 尝试解析为数组（多选题）
              try {
                const parsed = JSON.parse(answersData[qid])
                if (Array.isArray(parsed)) {
                  // 多选题
                  this.$set(this.multipleAnswers, qid, parsed)
                } else {
                  // 单选题、判断题、主观题
                  this.$set(this.answers, qid, answersData[qid])
                }
              } catch (e) {
                // 解析失败，说明是普通字符串
                this.$set(this.answers, qid, answersData[qid])
              }
            })
          }

          // 恢复每题得分
          if (res.data.scores) {
            const scoresData = JSON.parse(res.data.scores || '{}')
            Object.keys(scoresData).forEach(qid => {
              this.$set(this.questionScores, qid, scoresData[qid])
            })
          }

          // 恢复每题评分状态
          if (res.data.isScoredMap) {
            const isScoredData = JSON.parse(res.data.isScoredMap || '{}')
            Object.keys(isScoredData).forEach(qid => {
              this.$set(this.questionIsScored, qid, isScoredData[qid])
            })
          }

          // 恢复每题附件
          if (res.data.attachments && shouldRestoreAnswers) {
            const attachmentsData = JSON.parse(res.data.attachments || '{}')
            Object.keys(attachmentsData).forEach(qid => {
              const attachmentStr = attachmentsData[qid]
              if (attachmentStr) {
                // 将逗号分隔的URL字符串转换为数组
                const urls = attachmentStr.split(',').filter(url => url.trim())
                this.$set(this.attachments, qid, urls)
                // 构建上传文件列表（用于el-upload显示）
                const fileList = urls.map((url, idx) => ({
                  name: `image_${idx + 1}`,
                  url: url,
                  uid: Date.now() + idx,
                  response: { data: { url: url } }
                }))
                this.$set(this.uploadedImages, qid, fileList)
              }
            })
          }
        }
      } catch (error) {
        // 忽略错误，可能还没有提交
      }
    },
    // 处理答案变化 - 触发自动保存
    handleAnswerChange() {
      this.hasUnsavedChanges = true
      this.scheduleAutoSave()
    },
    // 页面初始化后立即保存一次（确保所有题目都有记录）
    async initializeAnswers() {
      try {
        const homeworkId = parseInt(this.$route.params.homeworkId)

        // 初始化所有题目的答案（即使是空答案）
        const answersData = {}
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            // 如果已有答案就用已有的，否则用空字符串
            answersData[qid] = this.answers[qid] || ''
          }
        })

        // 初始化多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 调用草稿保存 API（不判分，不标记为已提交）
        const res = await this.$store.dispatch('classroom/saveHomeworkDraft', {
          homeworkId: homeworkId,
          answers: answersData
        })
      } catch (error) {
        // Silently handle error
      }
    },
    // 处理多选题选项变化
    handleMultipleChoiceChange(questionId) {
      // 确保 multipleAnswers[questionId] 是数组
      if (!Array.isArray(this.multipleAnswers[questionId])) {
        this.$set(this.multipleAnswers, questionId, [])
      }
      // 多选题不立即触发自动保存，只在失去焦点或页面退出时保存
      // this.handleAnswerChange()
    },
    // 安排自动保存（延迟执行，避免频繁保存）
    scheduleAutoSave() {
      if (this.autoSaveTimer) {
        clearTimeout(this.autoSaveTimer)
      }
      // 3秒后自动保存
      this.autoSaveTimer = setTimeout(() => {
        this.saveDraft()
      }, 3000)
    },
    // 保存草稿（批量提交）
    async saveDraft() {
      this.autoSaving = true
      try {
        const homeworkId = parseInt(this.$route.params.homeworkId)

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案（包括空值，确保保存所有题目的状态）
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 调用草稿保存 API（不判分，不标记为已提交）
        const res = await this.$store.dispatch('classroom/saveHomeworkDraft', {
          homeworkId: homeworkId,
          answers: answersData
        })

        if (res.code === 200) {
          this.hasUnsavedChanges = false
          this.lastSaveTime = moment().format('HH:mm:ss')
          this.$message.success('草稿保存成功')
          // 草稿保存后不需要重新加载提交状态（因为不会改变 isSubmitted）
        } else {
          this.$message.error(res.message || '保存失败')
        }
      } catch (error) {
        this.$message.error('保存失败')
      } finally {
        this.autoSaving = false
      }
    },
    // 同步保存草稿（用于页面退出时，不显示提示）
    saveDraftSync() {
      try {
        const homeworkId = parseInt(this.$route.params.homeworkId)

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案（包括空值）
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 准备附件数据（将URL数组用逗号连接）
        const attachmentsData = {}
        Object.keys(this.attachments).forEach(qid => {
          if (Array.isArray(this.attachments[qid]) && this.attachments[qid].length > 0) {
            attachmentsData[qid] = this.attachments[qid].join(',')
          }
        })

        // 调试日志：打印要发送的附件数据

        // 使用 fetch 的 keepalive 选项来确保页面关闭时也能发送请求
        // 调用草稿保存 API（不判分，不标记为已提交）
        fetch('/api/classroom/homework/draft', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': localStorage.getItem('token') || ''
          },
          body: JSON.stringify({
            homeworkId: homeworkId,
            answers: answersData,
            attachments: attachmentsData
          }),
          keepalive: true
        }).then(response => {
          // Silently handle response
        }).catch(error => {
          // Silently handle error
        })
      } catch (error) {
        // Silently handle error
      }
    },
    // 显示提交确认对话框
    showSubmitConfirm() {
      // 检查是否有未完成的题目
      const unanswered = this.getUnansweredQuestions()
      if (unanswered.length > 0) {
        this.$message.warning(`请完成所有题目后再提交`)
        return
      }
      this.showConfirmDialog = true
    },
    // 获取未完成的题目列表
    getUnansweredQuestions() {
      const unanswered = []
      this.homework.questions.forEach(item => {
        // 编程题处理
        if (item.problemId) {
          // 检查编程题是否已提交
          const status = this.programmingStatus[item.problemId]
          if (status !== 'submitted') {
            unanswered.push(`第${this.getQuestionIndex(item)}题 (编程题)`)
          }
          return
        }

        // 普通题目处理
        if (!item.question) return

        const qid = item.question.id
        if (item.question.type === 'multiple_choice') {
          // 多选题：检查是否至少选择了一个选项
          if (!this.multipleAnswers[qid] || this.multipleAnswers[qid].length === 0) {
            unanswered.push(`第${this.getQuestionIndex(item)}题`)
          }
        } else {
          // 其他题型：检查是否有答案
          if (!this.answers[qid]) {
            unanswered.push(`第${this.getQuestionIndex(item)}题`)
          }
        }
      })
      return unanswered
    },
    // 获取题目序号
    getQuestionIndex(item) {
      const index = this.homework.questions.findIndex(q => q.id === item.id)
      return index + 1
    },
    // 确认提交
    async confirmSubmit() {
      this.submitting = true
      try {
        const homeworkId = parseInt(this.$route.params.homeworkId)

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 准备附件数据（将URL数组用逗号连接）

        const attachmentsData = {}
        Object.keys(this.attachments).forEach(qid => {
          if (Array.isArray(this.attachments[qid]) && this.attachments[qid].length > 0) {
            attachmentsData[qid] = this.attachments[qid].join(',')
          } else {
          }
        })

        // 调试日志：打印要发送的附件数据

        // 检查是否有答案
        if (Object.keys(answersData).length === 0) {
          this.$message.warning('请至少作答一道题目')
          return
        }

        // 批量提交
        const res = await this.$store.dispatch('classroom/submitHomework', {
          homeworkId: homeworkId,
          answers: answersData,
          attachments: attachmentsData
        })

        if (res.code === 200) {
          this.$message.success(this.$t('m.Submit_Success'))
          this.showConfirmDialog = false
          this.hasUnsavedChanges = false
          this.isSubmitted = true // 标记为已正式提交
          await this.loadSubmission()
        } else {
          this.$message.error(res.message || this.$t('m.Submit_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Submit_Failed'))
      } finally {
        this.submitting = false
      }
    },
    parseOptions(optionsStr) {
      try {
        const options = JSON.parse(optionsStr || '[]')
        return options.map((opt, idx) => ({
          letter: ['A', 'B', 'C', 'D'][idx],
          text: opt.replace(/^[A-D]\.\s*/, '')
        }))
      } catch {
        return []
      }
    },
    formatContent(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return content
      }
    },
    goToProblem(problemId) {
      window.open(`/problem/${problemId}`, '_blank')
    },
    goBack() {
      this.$router.back()
    },
    formatTime(time) {
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
    // 初始化编程题状态
    async initializeProgrammingStatus() {
      // 为所有编程题初始化状态
      this.homework.questions.forEach(item => {
        if (item.problemId) {
          this.$set(this.programmingStatus, item.problemId, 'not_started')
          this.$set(this.programmingStatusText, item.problemId, '未作答')
        }
      })

      // 2秒后开始第一次检测
      setTimeout(() => {
        this.checkProgrammingStatus()
      }, 2000)
    },
    // 启动轮询检测编程题提交状态
    startPollingProgrammingStatus() {
      this.pollingTimer = setInterval(() => {
        this.checkProgrammingStatus()
      }, 500)
    },
    // 检测编程题提交状态
    async checkProgrammingStatus() {
      try {
        const homeworkId = this.$route.params.homeworkId

        // 获取所有编程题
        const programmingQuestions = this.homework.questions.filter(q => q.problemId)

        for (const question of programmingQuestions) {
          try {
            // 调用后端API查询该题目的提交记录
            const res = await this.$store.dispatch('classroom/getProgrammingSubmissions', {
              homeworkId: homeworkId,
              questionId: question.id
            })

            const hasSubmission = res.code === 200 && res.data && res.data.length > 0
            const currentStatus = this.programmingStatus[question.problemId]

            if (hasSubmission && currentStatus !== 'submitted') {
              // 有提交记录且状态改变
              this.$set(this.programmingStatus, question.problemId, 'submitted')
              this.$set(this.programmingStatusText, question.problemId, '已作答')
            } else if (!hasSubmission && currentStatus !== 'not_started') {
              // 没有提交记录且状态改变
              this.$set(this.programmingStatus, question.problemId, 'not_started')
              this.$set(this.programmingStatusText, question.problemId, '未作答')
            }
            // 如果状态没有改变，不更新，避免闪烁
          } catch (error) {
            console.error('检测编程题状态失败:', error)
          }
        }
      } catch (error) {
        console.error('检测编程题状态失败:', error)
      }
    },
    // 图片上传成功回调
    handleUploadSuccess(response, file, fileList, questionId) {
      if (response.code === 200 && response.data && response.data.url) {
        const url = response.data.url
        // 统一转换为字符串类型作为键
        const qid = String(questionId)
        // 初始化该题目的图片数组
        if (!this.attachments[qid]) {
          this.$set(this.attachments, qid, [])
        }
        // 添加URL到数组
        this.attachments[qid].push(url)
        // 保存文件列表
        this.$set(this.uploadedImages, qid, fileList)
        // 触发自动保存
        this.handleAnswerChange()
      } else {
        this.$message.error(response.message || '图片上传失败')
      }
    },
    // 图片上传失败回调
    handleUploadError(err, file, fileList) {
      this.$message.error('图片上传失败，请重试')
    },
    // 删除图片回调
    handleRemoveFile(file, fileList, questionId) {
      // 统一转换为字符串类型作为键
      const qid = String(questionId)
      // 从attachments中移除该图片URL
      if (file.response && file.response.data && file.response.data.url) {
        const url = file.response.data.url
        if (this.attachments[qid]) {
          const index = this.attachments[qid].indexOf(url)
          if (index > -1) {
            this.attachments[qid].splice(index, 1)
          }
        }
      }
      // 更新文件列表
      this.$set(this.uploadedImages, qid, fileList)
      // 触发自动保存
      this.handleAnswerChange()
    },
    // 超出上传数量限制
    handleExceed(files, fileList) {
      this.$message.warning('最多只能上传5张图片')
    },
    // 获取某题目的图片列表（用于el-upload）
    getImageList(questionId) {
      return this.uploadedImages[questionId] || []
    },
    // 获取已提交的图片列表（用于显示）
    getSubmittedImages(questionId) {
      try {
        const qid = String(questionId)

        // 优先从 this.attachments 获取（loadSubmission 时已恢复）
        if (this.attachments && this.attachments[qid]) {
          return this.attachments[qid]
        }

        // 如果 this.attachments 没有，尝试从 this.submission.attachments 获取
        if (this.submission && this.submission.attachments) {
          const attachmentsData = JSON.parse(this.submission.attachments || '{}')
          const attachmentStr = attachmentsData[qid]
          if (attachmentStr) {
            const urls = attachmentStr.split(',').filter(url => url.trim())
            return urls
          }
        }

        return []
      } catch (e) {
        console.error('解析attachments失败:', e)
        return []
      }
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
.questions-container {
  margin: 20px 0;
}
.question-item {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #eee;
  border-radius: 8px;
  background: #fafafa;
}
.question-header {
  margin-bottom: 15px;
  font-weight: bold;
}
.question-number {
  font-size: 18px;
  color: #409EFF;
}
.question-type {
  color: #909399;
  margin: 0 10px;
  font-size: 14px;
}
.question-score {
  color: #67C23A;
  font-size: 14px;
}
.question-score-earned {
  color: #409EFF;
  font-size: 14px;
  margin-left: auto;
}
.question-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 10px;
}
.question-content {
  margin-bottom: 15px;
  line-height: 1.6;
}
.question-options {
  margin-top: 15px;
}
.option-item {
  margin: 10px 0;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  /* 移除 transition 避免轮询时闪烁 */
}
.option-item:hover {
  background: #f5f7fa;
}
.student-answer {
  margin-top: 15px;
  padding: 10px;
  background: #e6f7ff;
  border-left: 3px solid #409EFF;
  border-radius: 4px;
}
.action-area {
  text-align: center;
  padding: 20px 0;
}
.submit-area {
  text-align: center;
  padding: 20px 0;
}
.subjective-answer {
  margin-top: 15px;
}
.programming-answer {
  margin-top: 15px;
  text-align: center;
  padding: 20px;
  background: #f0f9ff;
  border-radius: 4px;
}
</style>

<!-- 非scoped样式，确保markdown-body样式生效 -->
<style>
.homework-detail .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.homework-detail .markdown-body h1,
.homework-detail .markdown-body h2,
.homework-detail .markdown-body h3,
.homework-detail .markdown-body h4,
.homework-detail .markdown-body h5,
.homework-detail .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.homework-detail .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.homework-detail .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.homework-detail .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.homework-detail .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.homework-detail .markdown-body p {
  margin-bottom: 16px !important;
}

.homework-detail .markdown-body strong {
  font-weight: bold !important;
}

.homework-detail .markdown-body em {
  font-style: italic !important;
}

.homework-detail .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.homework-detail .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 答案显示样式 */
.correct-answer {
  margin-top: 15px;
  padding: 10px;
  background: #f0f9ff;
  border-left: 3px solid #67C23A;
  border-radius: 4px;
}

.reference-answer {
  margin-top: 15px;
  padding: 15px;
  background: #f5f7fa;
  border-left: 3px solid #409EFF;
  border-radius: 4px;
}

.reference-answer-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
  font-size: 14px;
}

.reference-answer-content {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
}

.reference-answer-empty {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  color: #909399;
  font-style: italic;
}

/* 图片上传样式 */
.image-upload-area {
  margin-top: 15px;
  padding: 15px;
  background: #f9f9f9;
  border-radius: 4px;
  border: 1px dashed #d9d9d9;
}

.upload-tip {
  margin-bottom: 10px;
  color: #606266;
  font-size: 14px;
}

.upload-limit-tip {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

.hide-upload-btn .el-upload--picture-card {
  display: none;
}

.submitted-images {
  margin-top: 15px;
  padding: 15px;
  background: #f0f9ff;
  border-radius: 4px;
  border: 1px solid #b3d8ff;
}

.submitted-images-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
  font-size: 14px;
}

.submitted-images-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
