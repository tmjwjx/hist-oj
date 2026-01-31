<template>
  <div class="contest-question-qa">
    <el-card shadow>
      <div slot="header">
        <span>我的问题</span>
        <el-button
          style="float: right"
          type="primary"
          size="small"
          icon="el-icon-plus"
          @click="showCreateDialog = true"
        >
          提问
        </el-button>
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
      title="提问"
      :visible.sync="showCreateDialog"
      width="600px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="form" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input
            v-model="form.title"
            placeholder="请输入问题标题"
            maxlength="200"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="8"
            placeholder="请详细描述您的问题"
            maxlength="1000"
            show-word-limit
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          提交
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
        title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
        content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
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
            this.$message.success('提问成功')
            this.showCreateDialog = false
            this.resetForm()
            this.loadQuestions()
          } else {
            this.$message.error(res.data.msg || '提问失败')
          }
        } catch (error) {
          this.$message.error('提问失败')
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
