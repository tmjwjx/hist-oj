<template>
  <div class="contest-question-detail">
    <el-card shadow>
      <div slot="header">
        <el-button icon="el-icon-arrow-left" size="small" @click="goBack">返回</el-button>
        <el-dropdown
          v-if="isContestAdmin"
          style="float: right; margin-left: 10px"
          @command="handleStatusChange"
        >
          <el-button size="small">
            状态：{{ getStatusText(question.status) }}<i class="el-icon-arrow-down el-icon--right"></i>
          </el-button>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item command="pending">待回复</el-dropdown-item>
            <el-dropdown-item command="answered">已回复</el-dropdown-item>
            <el-dropdown-item command="closed">已关闭</el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
        <el-button
          v-if="isContestAdmin || question.questionerId === currentUserId"
          type="danger"
          size="small"
          style="float: right"
          @click="handleDelete"
        >
          删除
        </el-button>
      </div>

      <div v-loading="loading">
        <el-empty v-if="!question.id && !loading" description="问题不存在或加载失败"></el-empty>

        <div v-if="question.id" class="question-info">
          <h2 class="question-title">{{ question.title }}</h2>
          <div class="question-meta">
            <UserName
              v-if="question.questioner"
              :username="question.questioner.username"
              :rating="question.questioner.histRating"
              :bold="true"
            >
              {{ question.questioner.username }}
            </UserName>
            <span class="meta-item">状态：{{ getStatusText(question.status) }}</span>
            <span class="meta-item">{{ formatTime(question.createdAt) }}</span>
          </div>
          <div class="question-content">{{ question.content }}</div>
        </div>

        <el-divider></el-divider>

        <div class="reply-list" ref="replyList">
          <div
            v-for="reply in replies"
            :key="reply.id"
            class="reply-item"
            :class="isMyReply(reply) ? 'my-reply' : 'other-reply'"
          >
            <div v-if="!isMyReply(reply)" class="reply-avatar">
              <el-avatar :size="40" :src="reply.sender?.avatar">
                <i class="el-icon-user-solid"></i>
              </el-avatar>
            </div>
            <div class="reply-body">
              <div class="reply-meta">
                <UserName
                  v-if="!isMyReply(reply)"
                  :username="reply.sender?.username"
                  :rating="reply.sender?.histRating"
                  :bold="true"
                  class="sender-name"
                >
                  {{ reply.sender?.username }}
                </UserName>
                <span class="reply-time">{{ formatTime(reply.createdAt) }}</span>
              </div>
              <div class="reply-bubble">{{ reply.content }}</div>
            </div>
            <div v-if="isMyReply(reply)" class="reply-avatar">
              <el-avatar :size="40" :src="reply.sender?.avatar">
                <i class="el-icon-user-solid"></i>
              </el-avatar>
            </div>
          </div>
        </div>

        <div class="reply-input" v-if="question.status !== 'closed' || isContestAdmin">
          <div v-if="question.status === 'closed' && isContestAdmin" class="closed-notice">
            <i class="el-icon-warning"></i> 问题已关闭，管理员仍可回复
          </div>
          <el-input
            v-model="newReply"
            type="textarea"
            :rows="3"
            placeholder="输入回复内容..."
            @keydown.enter.native="handleEnterKey"
          ></el-input>
          <div class="input-actions">
            <el-button
              type="primary"
              @click="sendReply"
              :loading="sending"
              :disabled="!newReply.trim()"
            >
              发送
            </el-button>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'
import { mapGetters } from 'vuex'

export default {
  name: 'ContestQuestionDetail',
  components: {
    UserName
  },
  mixins: [realtimeSync],
  data() {
    return {
      loading: false,
      question: {},
      replies: [],
      newReply: '',
      sending: false,
      currentUserId: null,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadQuestionDetail',
        immediate: true
      }
    }
  },
  computed: {
    ...mapGetters(['userInfo', 'isContestAdmin']),
    questionId() {
      return this.$route.params.questionId
    },
    contestID() {
      return this.$route.params.contestID
    }
  },
  mounted() {
    this.currentUserId = this.userInfo?.uuid
  },
  methods: {
    async loadQuestionDetail() {
      if (this.loading) return

      const isFirstLoad = !this.question.id
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('contestQuestion/getQuestionDetail', this.questionId)

        if (res.data.code === 200) {
          const data = res.data.data
          const newReplies = data.replies || []

          // 检查回复是否有变化
          if (isFirstLoad) {
            this.question = data
            this.replies = newReplies
            this.$nextTick(() => {
              this.scrollToBottom()
            })
          } else {
            const currentIds = this.replies.map(r => r.id).join(',')
            const newIds = newReplies.map(r => r.id).join(',')

            if (currentIds !== newIds) {
              this.replies = newReplies
              this.question.status = data.status
              this.$nextTick(() => {
                this.scrollToBottom()
              })
            }
          }
        } else {
          this.$message.error(res.data.msg || '加载失败')
        }
      } catch (error) {
        this.$message.error('加载失败: ' + (error.response?.data?.msg || error.message))
      } finally {
        this.loading = false
      }
    },
    async sendReply() {
      if (!this.newReply.trim()) return

      this.sending = true
      try {
        const res = await this.$store.dispatch('contestQuestion/sendReply', {
          questionId: this.questionId,
          content: this.newReply
        })
        if (res.data.code === 200) {
          this.newReply = ''
          this.loadQuestionDetail()
        } else {
          this.$message.error(res.data.msg || '发送失败')
        }
      } catch (error) {
        this.$message.error('发送失败')
      } finally {
        this.sending = false
      }
    },
    handleEnterKey(e) {
      if (!e.shiftKey) {
        e.preventDefault()
        this.sendReply()
      }
    },
    async handleStatusChange(status) {
      try {
        const res = await this.$store.dispatch('contestQuestion/updateQuestionStatus', {
          questionId: this.questionId,
          status
        })
        if (res.data.code === 200) {
          this.$message.success('状态更新成功')
          this.loadQuestionDetail()
        } else {
          this.$message.error(res.data.msg || '更新失败')
        }
      } catch (error) {
        this.$message.error('更新失败')
      }
    },
    async handleDelete() {
      this.$confirm('确定要删除这个问题吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('contestQuestion/deleteQuestion', this.questionId)
          if (res.data.code === 200) {
            this.$message.success('删除成功')
            this.goBack()
          } else {
            this.$message.error(res.data.msg || '删除失败')
          }
        } catch (error) {
          this.$message.error('删除失败')
        }
      })
    },
    isMyReply(reply) {
      return reply.senderId === this.currentUserId
    },
    goBack() {
      this.$router.back()
    },
    scrollToBottom() {
      const container = this.$refs.replyList
      if (container) {
        container.scrollTop = container.scrollHeight
      }
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
.contest-question-detail {
  padding: 10px;
}

.question-info {
  margin-bottom: 20px;
}

.question-title {
  font-size: 20px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 15px;
}

.question-meta {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  color: #909399;
  font-size: 14px;
}

.meta-item {
  margin-left: 20px;
}

.question-content {
  color: #606266;
  font-size: 15px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.reply-list {
  min-height: 300px;
  max-height: 500px;
  overflow-y: auto;
  margin-bottom: 20px;
  padding: 10px;
}

.reply-item {
  display: flex;
  margin-bottom: 20px;
}

.other-reply {
  justify-content: flex-start;
}

.my-reply {
  justify-content: flex-end;
}

.reply-avatar {
  flex-shrink: 0;
}

.other-reply .reply-avatar {
  margin-right: 10px;
}

.my-reply .reply-avatar {
  margin-left: 10px;
}

.reply-body {
  max-width: 70%;
}

.reply-meta {
  display: flex;
  align-items: center;
  margin-bottom: 5px;
  font-size: 12px;
  color: #909399;
}

.other-reply .reply-meta {
  justify-content: flex-start;
}

.my-reply .reply-meta {
  justify-content: flex-end;
}

.sender-name {
  margin-right: 10px;
}

.reply-bubble {
  padding: 10px 15px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.other-reply .reply-bubble {
  background-color: #f5f7fa;
  color: #303133;
  border-top-left-radius: 0;
}

.my-reply .reply-bubble {
  background-color: #4a90e2;
  color: #ffffff;
  border-top-right-radius: 0;
}

.reply-input {
  margin-top: 20px;
}

.input-actions {
  margin-top: 10px;
  text-align: right;
}
</style>
