<template>
  <div class="contest-question-list">
    <el-card shadow>
      <div slot="header">
        <span>答疑列表</span>
        <el-select
          v-model="filterStatus"
          size="small"
          style="float: right; width: 120px"
          @change="loadQuestions"
        >
          <el-option label="全部" value="all"></el-option>
          <el-option label="待回复" value="pending"></el-option>
          <el-option label="已回复" value="answered"></el-option>
          <el-option label="已关闭" value="closed"></el-option>
        </el-select>
      </div>

      <div v-loading="loading">
        <el-empty v-if="questions.length === 0 && !loading" description="暂无提问"></el-empty>

        <div v-else class="question-list">
          <div
            v-for="question in questions"
            :key="question.id"
            class="question-item"
            @click="viewQuestion(question.id)"
          >
            <div class="question-header">
              <el-tag :type="getStatusType(question.status)" size="small">
                {{ getStatusText(question.status) }}
              </el-tag>
              <UserName
                :username="question.questioner?.username || question.questionerId"
                :rating="question.questioner?.histRating"
                :bold="true"
                class="questioner-name"
              >
                {{ question.questioner?.username || question.questionerId }}
              </UserName>
              <span class="question-time">{{ formatTime(question.createdAt) }}</span>
            </div>
            <div class="question-title">{{ question.title }}</div>
            <div class="question-content">{{ question.content }}</div>
          </div>
        </div>

        <el-pagination
          v-if="total > limit"
          @current-change="handlePageChange"
          :current-page="page"
          :page-size="limit"
          :total="total"
          layout="prev, pager, next"
        >
        </el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'
import UserName from '@/components/oj/common/UserName.vue'

export default {
  name: 'ContestQuestionList',
  components: {
    UserName
  },
  data() {
    return {
      loading: false,
      questions: [],
      page: 1,
      limit: 20,
      total: 0,
      filterStatus: 'all'
    }
  },
  computed: {
    contestID() {
      return this.$route.params.contestID
    }
  },
  mounted() {
    this.loadQuestions()
  },
  methods: {
    async loadQuestions() {
      this.loading = true
      try {
        const params = { page: this.page, limit: this.limit, all: 'true' }
        if (this.filterStatus !== 'all') {
          params.status = this.filterStatus
        }
        const res = await this.$store.dispatch('contestQuestion/getQuestions', {
          contestId: this.contestID,
          params
        })
        if (res.data.code === 200) {
          this.questions = res.data.data.list || []
          this.total = res.data.data.total || 0
        }
      } catch (error) {
        this.$message.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    handlePageChange(page) {
      this.page = page
      this.loadQuestions()
    },
    viewQuestion(questionId) {
      // 跳转到问题详情页（需要实现 QuestionDetail 组件）
      this.$router.push({
        name: 'ContestQuestionDetail',
        params: { contestID: this.contestID, questionId }
      })
    },
    getStatusType(status) {
      const map = {
        pending: 'warning',
        answered: 'success',
        closed: 'info'
      }
      return map[status] || 'info'
    },
    getStatusText(status) {
      const map = {
        pending: '待回复',
        answered: '已回复',
        closed: '已关闭'
      }
      return map[status] || status
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    }
  }
}
</script>

<style scoped>
.contest-question-list {
  padding: 10px;
}

.question-list {
  margin-top: 10px;
}

.question-item {
  padding: 15px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.3s;
}

.question-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.question-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.questioner-name {
  flex: 1;
  margin-left: 10px;
}

.question-time {
  font-size: 12px;
  color: #909399;
}

.question-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.question-content {
  color: #606266;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.el-pagination {
  margin-top: 20px;
  text-align: center;
}
</style>
