<template>
  <div class="homework-detail">
    <div class="header">
      <h3>{{ homework.title || '作业详情' }}</h3>
      <div>
        <el-button @click="editHomework">{{ $t('m.Edit') }}</el-button>
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <el-card v-loading="loading">
      <div v-if="homework.id">
        <p><strong>{{ $t('m.Homework_Title') }}:</strong> {{ homework.title }}</p>
        <p><strong>{{ $t('m.Description') }}:</strong> {{ homework.description || '-' }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(homework.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(homework.endTime) }}</p>

        <el-divider></el-divider>

        <h4>{{ $t('m.Questions') }}</h4>
        <el-table :data="homework.questions || []" stripe>
          <el-table-column prop="question.title" :label="$t('m.Question_Title')" />
          <el-table-column prop="question.type" :label="$t('m.Question_Type')" width="120">
            <template slot-scope="{ row }">
              {{ getQuestionTypeText(row.question.type) }}
            </template>
          </el-table-column>
          <el-table-column prop="score" :label="$t('m.Score')" width="80" />
        </el-table>

        <el-divider></el-divider>

        <h4>{{ $t('m.Student_Submissions') }}</h4>
        <div style="margin-bottom: 15px;">
          <el-button type="success" icon="el-icon-download" @click="exportHomeworkScores">
            导出成绩
          </el-button>
        </div>
        <el-table :data="studentSubmissions" stripe>
          <el-table-column prop="studentName" :label="$t('m.Student')" />
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
      submissions: [],
      studentSubmissions: [] // 聚合后的学生提交数据
    }
  },
  mounted() {
    this.loadHomeworkDetail()
  },
  methods: {
    async loadHomeworkDetail() {
      this.loading = true
      try {
        const homeworkId = this.$route.params.homeworkId
        const [homeworkRes, submissionsRes] = await Promise.all([
          this.$store.dispatch('classroom/getHomeworkDetail', homeworkId),
          this.$store.dispatch('classroom/getHomeworkSubmissions', homeworkId)
        ])

        if (homeworkRes.code === 200) {
          this.homework = homeworkRes.data
        }

        if (submissionsRes.code === 200) {
          this.submissions = submissionsRes.data || []
          this.aggregateStudentSubmissions()
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    aggregateStudentSubmissions() {
      // 将按题目分组的提交数据聚合为学生维度
      const studentMap = new Map()

      this.submissions.forEach(submit => {
        const uid = submit.uid
        if (!studentMap.has(uid)) {
          studentMap.set(uid, {
            uid: uid,
            studentName: submit.student?.username || submit.student?.realName || uid,
            submitTime: submit.createdAt,
            totalScore: 0,
            completedCount: 0,
            totalCount: 0,
            questions: []
          })
        }

        const student = studentMap.get(uid)
        student.totalCount++
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
      this.$router.push({
        name: 'StudentSubmissionDetail',
        params: { homeworkId: this.$route.params.homeworkId },
        query: { uid: submission.uid }
      })
    },
    editHomework() {
      // 跳转到编辑页面，复用 CreateHomework 组件
      this.$router.push({
        name: 'CreateHomework',
        params: { classroomId: this.$route.params.classroomId },
        query: { editId: this.homework.id }
      })
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

      // 表头：学生姓名 + 每题得分
      const headers = ['学生姓名']
      this.homework.questions.forEach((item, index) => {
        headers.push(`题目${index + 1}_${item.question?.title || ''}(${item.score}分)`)
      })
      headers.push('总分', '提交时间', '完成进度', '评分状态')
      csvContent += headers.map(h => `"${h}"`).join(',') + '\n'

      // 数据行
      this.studentSubmissions.forEach(student => {
        const row = [student.studentName]

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
</style>
