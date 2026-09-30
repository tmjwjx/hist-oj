<template>
  <div class="contest-question-qa">
    <el-card shadow>
      <div slot="header">
        <span>{{ $t('m.CQ_My_Questions') }}</span>
        <el-button
          style="float: right"
          type="primary"
          size="small"
          icon="el-icon-plus"
          @click="showCreateDialog = true"
        >
          {{ $t('m.CQ_Ask') }}
        </el-button>
      </div>

      <div v-loading="loading">
        <el-empty v-if="questions.length === 0 && !loading" :description="$t('m.CQ_No_Questions')"></el-empty>

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
              <span class="question-title">{{ question.title }}</span>
              <span class="question-time">{{ formatTime(question.createdAt) }}</span>
            </div>
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

    <!-- 创建问题对话框 -->
    <el-dialog
      :title="$t('m.CQ_Ask')"
      :visible.sync="showCreateDialog"
      width="600px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="form" label-width="80px">
        <el-form-item :label="$t('m.Title')" prop="title">
          <el-input
            v-model="form.title"
            :placeholder="$t('m.CQ_Title_Placeholder')"
            maxlength="200"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.Content')" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="8"
            :placeholder="$t('m.CQ_Content_Placeholder')"
            maxlength="1000"
            show-word-limit
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ $t('m.CQ_Submit') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'

export default {
  name: 'ContestQuestionQA',
  data() {
    return {
      loading: false,
      questions: [],
      page: 1,
      limit: 20,
      total: 0,
      showCreateDialog: false,
      submitting: false,
      form: {
        title: '',
        content: ''
      },
      rules: {
        title: [{ required: true, message: this.$t('m.CQ_Title_Required'), trigger: 'blur' }],
        content: [{ required: true, message: this.$t('m.CQ_Content_Required'), trigger: 'blur' }]
      }
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
        const res = await this.$store.dispatch('contestQuestion/getQuestions', {
          contestId: this.contestID,
          params: { page: this.page, limit: this.limit }
        })
        if (res.data.code === 200) {
          this.questions = res.data.data.list || []
          this.total = res.data.data.total || 0
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
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
      // 暂时使用对话框显示详情
      this.$router.push({
        name: 'ContestQuestionDetail',
        params: { contestID: this.contestID, questionId }
      })
    },
    async handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return

        this.submitting = true
        try {
          const res = await this.$store.dispatch('contestQuestion/createQuestion', {
            contestId: Number(this.contestID),
            title: this.form.title,
            content: this.form.content
          })
          if (res.data.code === 200) {
            this.$message.success(this.$t('m.CQ_Ask_Success'))
            this.showCreateDialog = false
            this.resetForm()
            this.loadQuestions()
          } else {
            this.$message.error(res.data.msg || this.$t('m.CQ_Ask_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.CQ_Ask_Failed'))
        } finally {
          this.submitting = false
        }
      })
    },
    resetForm() {
      this.form = {
        title: '',
        content: ''
      }
      this.$refs.form && this.$refs.form.clearValidate()
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
        pending: this.$t('m.CQ_Status_Pending'),
        answered: this.$t('m.CQ_Status_Answered'),
        closed: this.$t('m.CQ_Status_Closed')
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
.contest-question-qa {
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

.question-title {
  flex: 1;
  margin-left: 10px;
  font-weight: 500;
  color: #303133;
}

.question-time {
  font-size: 12px;
  color: #909399;
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
