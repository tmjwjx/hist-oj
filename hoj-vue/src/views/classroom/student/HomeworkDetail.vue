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

    <el-card v-loading="loading">
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
            <div class="question-title">{{ item.question.title }}</div>
            <div class="question-content" v-html="formatContent(item.question.content)"></div>

            <!-- 单选题选项 -->
            <div v-if="item.question.type === 'single_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.question.options)" :key="idx" class="option-item">
                <el-radio
                  v-model="answers[item.question.id]"
                  :label="option.letter"
                  @change="handleAnswerChange"
                >
                  {{ option.letter }}. {{ option.text }}
                </el-radio>
              </div>
              <!-- 显示学生已选择的选项 -->
              <div v-if="answers[item.question.id]" class="student-answer">
                <el-tag type="info">已选: {{ answers[item.question.id] }}</el-tag>
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
                  {{ option.letter }}. {{ option.text }}
                </el-checkbox>
              </div>
              <!-- 显示学生已选择的选项（按字典序排列） -->
              <div v-if="multipleAnswers[item.question.id] && multipleAnswers[item.question.id].length > 0" class="student-answer">
                <el-tag type="info">已选: {{ [...multipleAnswers[item.question.id]].sort().join(', ') }}</el-tag>
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
            </div>

            <!-- 编程题 -->
            <div v-if="item.question.type === 'programming'" class="programming-answer">
              <p>此题为编程题，请在 HOJ 平台提交</p>
              <el-button v-if="item.question.problemId" type="primary" @click="goToProblem(item.question.problemId)">
                前往 HOJ 提交
              </el-button>
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

export default {
  name: 'HomeworkDetail',
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
      questionIsScored: {} // 每题是否已评分 { questionId: boolean }
    }
  },
  computed: {
    canSubmit() {
      return this.homework.status === 2 && !this.isSubmitted
    },
    // 是否可以查看作业（题目和答案）
    canViewHomework() {
      // 如果已提交，根据教师设置判断
      if (this.isSubmitted) {
        return this.homework.showHomework === 1
      }
      // 未提交时，作业进行中可以查看
      return this.homework.status === 2
    },
    // 是否可以查看分数
    canViewScore() {
      // 必须已提交且教师允许查看分数
      return this.isSubmitted && this.homework.showScore === 1
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
    // 页面退出前总是保存所有答案（确保不丢失数据）
    this.saveDraftSync()
  },
  methods: {
    async loadHomeworkDetail() {
      this.loading = true
      try {
        const homeworkId = this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', homeworkId)
        if (res.code === 200) {
          this.homework = res.data
          // 初始化所有题目的答案对象
          this.homework.questions.forEach(item => {
            const qid = item.question.id
            if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
              // 使用 $set 确保响应式
              this.$set(this.answers, qid, '')
            } else if (item.question.type === 'multiple_choice') {
              this.$set(this.multipleAnswers, qid, [])
            }
          })

          // 加载已提交的答案
          await this.loadSubmission()
          // 初始化答案（确保所有题目都有记录）
          await this.initializeAnswers()
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    async loadSubmission() {
      try {
        const homeworkId = this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getStudentHomeworkDetail', homeworkId)

        if (res.code === 200 && res.data) {
          this.submission = res.data

          // 恢复已提交的答案
          if (res.data.answers) {
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

            // 根据后端返回的 isOfficiallySubmitted 字段设置提交状态
            // 只有正式提交后才显示分数和提交时间
            if (res.data.isOfficiallySubmitted) {
              this.isSubmitted = true
            }
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
            answers: answersData
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
        if (item.question.type === 'programming') return // 编程题不需要在这里提交

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

        // 检查是否有答案
        if (Object.keys(answersData).length === 0) {
          this.$message.warning('请至少作答一道题目')
          return
        }

        // 批量提交
        const res = await this.$store.dispatch('classroom/submitHomework', {
          homeworkId: homeworkId,
          answers: answersData
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
      // 简单的内容格式化，可以根据需要扩展
      return content
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
  transition: all 0.2s;
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
