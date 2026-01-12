<template>
  <div class="submission-detail">
    <div class="header">
      <h3>{{ $t('m.Student_Submission_Detail') }}</h3>
      <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
    </div>

    <el-card v-loading="loading">
      <div v-if="studentSubmission">
        <!-- 学生信息 -->
        <div class="student-info">
          <p><strong>{{ $t('m.Student') }}:</strong> {{ studentSubmission.studentName }}</p>
          <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(studentSubmission.submitTime) }}</p>
          <p><strong>{{ $t('m.Total_Score') }}:</strong> {{ studentSubmission.totalScore }}</p>
        </div>

        <el-divider></el-divider>

        <!-- 题目列表 -->
        <div v-for="(submit, index) in studentSubmission.questions" :key="submit.id" class="question-item">
          <div class="question-header">
            <span class="question-number">{{ index + 1 }}.</span>
            <el-tag :type="getQuestionTypeTag(submit.question?.type)" size="small">
              {{ getQuestionTypeText(submit.question?.type) }}
            </el-tag>
            <span class="question-score">{{ submit.question?.score || 0 }}分</span>
          </div>

          <div class="question-title">{{ submit.question?.title }}</div>

          <!-- 显示题目选项（单选、多选、判断题） -->
          <div v-if="submit.question?.type === 'single_choice' || submit.question?.type === 'multiple_choice'" class="question-options">
            <div v-for="option in parseOptions(submit.question.options)" :key="option.letter" class="option-item">
              <el-tag :type="isOptionSelected(submit.answer, option.letter, submit.question.type) ? 'primary' : 'info'" size="small">
                {{ option.letter }}. {{ option.text }}
              </el-tag>
            </div>
          </div>

          <!-- 判断题选项 -->
          <div v-if="submit.question?.type === 'judge'" class="question-options">
            <el-tag :type="submit.answer === 'true' ? 'primary' : 'info'" size="small">
              {{ submit.answer === 'true' ? $t('m.True') : $t('m.False') }}
            </el-tag>
          </div>

          <!-- 主观题答案 -->
          <div v-if="submit.question?.type === 'subjective'" class="subjective-answer">
            <p><strong>{{ $t('m.Student_Answer') }}:</strong></p>
            <div class="answer-content">{{ submit.answer || $t('m.No_Answer') }}</div>
          </div>

          <!-- 答案对比 -->
          <div class="answer-comparison">
            <p><strong>{{ $t('m.Correct_Answer') }}:</strong>
              <span v-if="submit.question?.type === 'single_choice' || submit.question?.type === 'judge'">
                {{ submit.question?.answer }}
              </span>
              <span v-else-if="submit.question?.type === 'multiple_choice'">
                {{ parseMultipleChoiceAnswer(submit.question?.answer) }}
              </span>
              <span v-else>
                {{ submit.question?.answer || '-' }}
              </span>
            </p>
            <p><strong>{{ $t('m.Student_Answer') }}:</strong>
              <span v-if="submit.question?.type === 'multiple_choice'">
                {{ parseMultipleChoiceAnswer(submit.answer) }}
              </span>
              <span v-else>
                {{ submit.answer }}
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
    </el-card>

    <!-- 评分对话框 -->
    <el-dialog
      :title="$t('m.Manual_Grading')"
      :visible.sync="gradeDialogVisible"
      width="500px"
    >
      <el-form :model="gradeForm" label-width="100px">
        <el-form-item :label="$t('m.Question_Type')">
          <el-tag>{{ getQuestionTypeText(currentQuestion?.question?.type) }}</el-tag>
        </el-form-item>
        <el-form-item :label="$t('m.Question_Score')">
          <span>{{ currentQuestion?.question?.score || 0 }} {{ $t('m.Points') }}</span>
        </el-form-item>
        <el-form-item :label="$t('m.Enter_Score')">
          <el-input-number
            v-model="gradeForm.score"
            :min="0"
            :max="currentQuestion?.question?.score || 0"
            :precision="1"
          ></el-input-number>
          <span style="margin-left: 10px; color: #909399;">
            (0 - {{ currentQuestion?.question?.score || 0 }})
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

export default {
  name: 'StudentSubmissionDetail',
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
      grading: false
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const homeworkId = this.$route.params.homeworkId
        const studentUid = this.$route.query.uid

        // 并行加载作业详情和提交记录
        const [homeworkRes, submissionsRes] = await Promise.all([
          this.$store.dispatch('classroom/getHomeworkDetail', homeworkId),
          this.$store.dispatch('classroom/getHomeworkSubmissions', homeworkId)
        ])

        if (homeworkRes.code === 200) {
          this.homework = homeworkRes.data
        }

        if (submissionsRes.code === 200) {
          this.submissions = submissionsRes.data || []
          this.buildStudentSubmission(studentUid)
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
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
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        return JSON.parse(optionsStr)
      } catch (e) {
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

      this.grading = true
      try {
        const res = await this.$store.dispatch('classroom/gradeHomework', {
          homeworkId: parseInt(this.$route.params.homeworkId),
          questionId: this.currentQuestion.questionId,
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
</style>
