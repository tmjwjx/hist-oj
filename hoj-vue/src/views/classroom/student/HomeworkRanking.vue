<template>
  <div class="student-homework-ranking">
    <div class="header">
      <h3>{{ homework.title || '作业排行榜' }}</h3>
      <div class="header-actions">
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <el-card v-loading="loading">
      <div v-if="homework.id" class="homework-meta">
        <p><strong>{{ $t('m.Homework_Title') }}:</strong> {{ homework.title }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(homework.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(homework.endTime) }}</p>
      </div>

      <el-divider></el-divider>

      <el-alert
        v-if="message"
        :title="message"
        type="info"
        :closable="false"
        show-icon
      />

      <el-table
        v-else
        :data="rankings"
        stripe
        border
        max-height="640"
        empty-text="暂无排行榜数据"
      >
        <el-table-column prop="rank" label="总分排名" width="110" align="center">
          <template slot-scope="{ row }">
            <el-tag type="warning" size="small">#{{ row.rank || '-' }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="realName" label="姓名" min-width="130" />

        <el-table-column label="用户名" min-width="160">
          <template slot-scope="{ row }">
            <UserName
              v-if="row.username"
              :username="row.username"
              :show-tooltip="true"
              :bold="true"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>

        <el-table-column label="总分" width="120" align="center">
          <template slot-scope="{ row }">
            <span v-if="row.hasScore">{{ formatScore(row.totalScore) }}</span>
            <span class="unanswered" v-else>未作答</span>
          </template>
        </el-table-column>

        <el-table-column
          v-for="column in questionColumns"
          :key="column.key"
          :label="column.label"
          width="100"
          align="center"
        >
          <template slot-scope="{ row }">
            <span v-if="hasQuestionScore(row, column.key)">{{ formatScore(row.questionScores[column.key]) }}</span>
            <span class="unanswered" v-else>未作答</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'
import UserName from '@/components/oj/common/UserName.vue'

export default {
  name: 'HomeworkRanking',
  components: {
    UserName
  },
  data() {
    return {
      loading: false,
      homework: {},
      questionColumns: [],
      rankings: [],
      message: ''
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      if (this.loading) return
      this.loading = true
      this.message = ''

      try {
        const homeworkId = this.$route.params.homeworkId
        const [detailRes, rankingRes] = await Promise.all([
          this.$store.dispatch('classroom/getHomeworkDetail', homeworkId),
          this.$store.dispatch('classroom/getHomeworkRanking', homeworkId)
        ])

        if (detailRes && detailRes.code === 200 && detailRes.data) {
          this.homework = detailRes.data
        }

        if (rankingRes && rankingRes.code === 200 && rankingRes.data) {
          this.questionColumns = Array.isArray(rankingRes.data.questionColumns) ? rankingRes.data.questionColumns : []
          this.rankings = Array.isArray(rankingRes.data.rankings) ? rankingRes.data.rankings : []
          this.message = ''
        } else {
          this.questionColumns = []
          this.rankings = []
          this.message = rankingRes.message || '教师未开放排行榜'
        }
      } catch (error) {
        this.questionColumns = []
        this.rankings = []
        this.message = '排行榜加载失败'
      } finally {
        this.loading = false
      }
    },
    hasQuestionScore(row, columnKey) {
      if (!row || !row.questionScores || !columnKey) return false
      return Object.prototype.hasOwnProperty.call(row.questionScores, columnKey)
    },
    formatScore(score) {
      const numericScore = Number(score)
      if (!Number.isFinite(numericScore)) return '-'
      return Number.isInteger(numericScore) ? numericScore : numericScore.toFixed(2)
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    goBack() {
      this.$router.back()
    }
  }
}
</script>

<style scoped>
.student-homework-ranking {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.homework-meta {
  font-size: 15px;
  line-height: 1.8;
  color: #303133;
}

.homework-meta p {
  margin: 6px 0;
}

.unanswered {
  color: #909399;
}
</style>
