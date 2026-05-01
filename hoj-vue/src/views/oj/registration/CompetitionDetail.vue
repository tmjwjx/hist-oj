<template>
  <div class="competition-detail-container">
    <div class="container">
      <!-- 头部 -->
      <div class="header">
        <div>
          <h1 v-if="competition">{{ competition.name }}</h1>
          <h1 v-else>加载中...</h1>
        </div>
        <button class="back-btn btn-primary" @click="goBack">← 返回列表</button>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="card" style="text-align: center; padding: 60px;">
        <p>正在加载比赛信息...</p>
      </div>

      <!-- 比赛详情 -->
      <div v-else-if="competition">
        <!-- 比赛信息 -->
        <div class="card">
          <h2>📋 比赛信息</h2>
          <div class="info-row">
            <span class="label">📅 报名时间：</span>
            <span>{{ formatTime(competition.startTime) }} - {{ formatTime(competition.endTime) }}</span>
          </div>
          <div class="info-row">
            <span class="label">📊 报名状态：</span>
            <span class="badge" :class="getRegistrationStatusClass()">
              {{ getRegistrationStatusText() }}
            </span>
          </div>
          <div v-if="competition.logoUrl" class="logo-container">
            <img :src="competition.logoUrl" :alt="competition.name" class="logo">
          </div>

          <!-- 比赛说明 -->
          <div v-if="competition.description" class="description-section">
            <h3>📖 比赛说明</h3>
            <div class="description-content" v-html="renderedDescription"></div>
          </div>
        </div>

        <!-- 我的报名状态 -->
        <div v-if="myRegistration" class="card">
          <h2>📝 我的报名</h2>
          <div class="registration-info">
            <div class="info-row">
              <span class="label">审核状态：</span>
              <span class="badge" :class="myRegistration.status">
                {{ getStatusText(myRegistration.status) }}
              </span>
            </div>
            <div class="info-row">
              <span class="label">与管理员沟通：</span>
              <button class="btn btn-chat" @click="openChat">
                💬 消息
                <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }}</span>
              </button>
            </div>
            <div v-if="myRegistration.status === 'rejected' && canEdit" class="edit-section">
              <p style="color: #e53e3e; margin-bottom: 16px;">您的报名被退回，请修改后重新提交</p>
              <button class="btn btn-primary" @click="showEditForm = true">修改报名信息</button>
            </div>
          </div>
        </div>

        <!-- 报名表单 -->
        <div v-if="!myRegistration && canRegister" class="card">
          <h2>✏️ 填写报名信息</h2>
          <form @submit.prevent="submitRegistration">
            <div v-for="(enabled, field) in fieldConfig" :key="field" class="form-group" v-show="enabled">
              <label>
                {{ getFieldLabel(field) }}
                <span v-if="isRequired(field)" class="required-mark">*</span>
              </label>
              <input
                v-if="field !== 'gender'"
                v-model="formData[field]"
                :type="getInputType(field)"
                :placeholder="getFieldPlaceholder(field)">
              <select v-else v-model="formData.gender">
                <option value="">请选择性别</option>
                <option value="男">男</option>
                <option value="女">女</option>
              </select>
            </div>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '提交中...' : '提交报名' }}
            </button>
          </form>
        </div>

        <!-- 编辑表单 -->
        <div v-if="showEditForm && myRegistration" class="card">
          <h2>✏️ 修改报名信息</h2>
          <form @submit.prevent="updateRegistration">
            <div v-for="(enabled, field) in fieldConfig" :key="field" class="form-group" v-show="enabled">
              <label>
                {{ getFieldLabel(field) }}
                <span v-if="isRequired(field)" class="required-mark">*</span>
              </label>
              <input
                v-if="field !== 'gender'"
                v-model="editFormData[field]"
                :type="getInputType(field)"
                :placeholder="getFieldPlaceholder(field)">
              <select v-else v-model="editFormData.gender">
                <option value="">请选择性别</option>
                <option value="男">男</option>
                <option value="女">女</option>
              </select>
            </div>
            <div style="display: flex; gap: 12px;">
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '更新中...' : '更新报名' }}
              </button>
              <button type="button" class="btn" @click="showEditForm = false">取消</button>
            </div>
          </form>
        </div>

        <!-- 不可报名提示 -->
        <div v-if="!canRegister && !myRegistration" class="card" style="text-align: center; padding: 40px;">
          <p style="font-size: 18px; color: #718096;">{{ cannotRegisterReason }}</p>
        </div>
      </div>

      <!-- 比赛不存在 -->
      <div v-else class="card" style="text-align: center; padding: 60px;">
        <p style="font-size: 18px; color: #e53e3e;">比赛不存在或已被删除</p>
      </div>
    </div>

    <!-- 聊天模态框 -->
    <div v-if="showChatModal" class="chat-modal" @click.self="closeChat">
      <div class="chat-modal-content">
        <div class="chat-modal-header">
          <h3>💬 与管理员沟通</h3>
          <button class="close-btn" @click="closeChat">×</button>
        </div>
        <div class="chat-modal-body">
          <!-- 消息列表 -->
          <div class="chat-messages" ref="chatMessagesContainer">
            <div v-if="chatMessages.length === 0" class="no-messages">
              暂无消息，开始与管理员沟通吧
            </div>
            <div
              v-for="(msg, index) in chatMessages"
              :key="index"
              class="chat-message"
              :class="{ 'message-user': msg.sender === 'user', 'message-admin': msg.sender === 'admin' }"
            >
              <div class="message-bubble">
                <div class="message-text">{{ msg.message }}</div>
                <div class="message-time">{{ formatMessageTime(msg.timestamp) }}</div>
              </div>
            </div>
          </div>

          <!-- 输入区域 -->
          <div class="chat-input-area">
            <textarea
              v-model="newMessage"
              ref="messageInput"
              class="chat-input"
              placeholder="输入消息... (Enter发送，Shift+Enter换行)"
              rows="3"
              @compositionstart="handleCompositionStart"
              @compositionend="handleCompositionEnd"
              @keydown.enter.exact.prevent="handleEnterKey"
            ></textarea>
            <button
              class="btn-send"
              @click="sendMessage"
              :disabled="!newMessage.trim() || sendingMessage"
            >
              {{ sendingMessage ? '发送中...' : '发送' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getCompetition, createRegistration, updateRegistration, getMyRegistration } from '@/api/registration'
import mMessage from '@/common/message'
import { wsManager } from '@/utils/websocket'
import MarkdownIt from 'markdown-it'
import katex from 'katex'
import '@iktakahiro/markdown-it-katex'

// 配置 markdown-it，支持 LaTeX 数学公式
const md = new MarkdownIt({
  html: true,          // 启用 HTML 标签（用于调整图片大小）
  linkify: true,       // 自动转换 URL
  typographer: true,   // 启用一些语言中立的替换和引号美化
})

// 使用 katex 插件
md.use(require('@iktakahiro/markdown-it-katex'), {
  throwOnError: false,
  errorColor: '#cc0000'
})

// 渲染 Markdown（支持 LaTeX 数学公式）
function renderMarkdown(text) {
  if (!text) return ''

  try {
    return md.render(text)
  } catch (error) {
    console.error('Markdown 渲染错误:', error)
    return `<p style="color: #f56c6c;">渲染错误: ${error.message}</p>`
  }
}

export default {
  name: 'CompetitionDetail',
  data() {
    return {
      loading: true,
      competition: null,
      myRegistration: null,
      fieldConfig: {},
      formData: {},
      editFormData: {},
      submitting: false,
      showEditForm: false,
      currentUser: null,

      // 聊天相关
      showChatModal: false,
      chatMessages: [],
      newMessage: '',
      sendingMessage: false,
      isComposing: false,
      justFinishedComposing: false,
      compositionEndTime: 0,
      unreadCount: 0
    }
  },
  computed: {
    renderedDescription() {
      if (!this.competition || !this.competition.description) return ''
      return renderMarkdown(this.competition.description)
    },
    canRegister() {
      if (!this.competition) return false
      const now = new Date()
      const start = new Date(this.competition.startTime)
      const end = new Date(this.competition.endTime)
      return now >= start && now <= end && !this.myRegistration
    },
    canEdit() {
      if (!this.myRegistration) return false
      if (this.myRegistration.status !== 'rejected') return false
      const now = new Date()
      const end = new Date(this.competition.endTime)
      return now <= end
    },
    cannotRegisterReason() {
      if (!this.competition) return ''
      const now = new Date()
      const start = new Date(this.competition.startTime)
      const end = new Date(this.competition.endTime)

      if (now < start) return '报名尚未开始'
      if (now > end) return '报名时间已截止'
      return ''
    }
  },
  methods: {
    async loadCompetition() {
      try {
        this.loading = true
        const res = await getCompetition(this.$route.params.id)
        if (res.code === 200) {
          this.competition = res.data
          this.fieldConfig = JSON.parse(res.data.fields || '{}')
          await this.loadMyRegistration()
        } else {
          mMessage.error(res.msg || '加载比赛失败')
        }
      } catch (error) {
        console.error('加载比赛失败', error)
        mMessage.error('加载比赛失败')
      } finally {
        this.loading = false
      }
    },

    async loadMyRegistration() {
      if (!this.currentUser) {
        return
      }

      if (!this.competition || !this.competition.id) {
        return
      }

      try {
        const res = await getMyRegistration(this.competition.id, this.currentUser.uuid)

        // 处理后端的标准响应格式: { code: 200, message: "success", data: {...} }
        if (res && res.code === 200 && res.data) {
          const registrationData = res.data

          // 检查是否是有效的报名数据（有id或其他关键字段）
          if (registrationData && registrationData.id) {
            this.myRegistration = registrationData
            this.calculateUnreadCount()
          } else {
            this.myRegistration = null
          }
        } else if (res && res.code === 404) {
          this.myRegistration = null
        } else {
          this.myRegistration = null
        }
      } catch (error) {
        // 404 或其他错误表示未报名
        this.myRegistration = null
      }
    },

    async submitRegistration() {
      if (!this.validateForm()) return

      try {
        this.submitting = true
        // 转换为snake_case以匹配后端API
        const data = {
          competition_id: this.competition.id,
          user_uuid: this.currentUser.uuid
        }
        // 转换formData字段：camelCase -> snake_case
        const fieldMapping = {
          name: 'name',
          class: 'class',
          college: 'college',
          studentId: 'student_id',
          gender: 'gender',
          shirtSize: 'shirt_size',
          teamName: 'team_name',
          qq: 'qq'
        }
        for (const [camelKey, snakeKey] of Object.entries(fieldMapping)) {
          if (this.formData[camelKey] !== undefined && this.formData[camelKey] !== '') {
            data[snakeKey] = this.formData[camelKey]
          }
        }
        const res = await createRegistration(data)

        if (res && res.code === 200) {
          mMessage.success('报名成功！')
          await this.loadMyRegistration()
        } else {
          mMessage.error(res?.message || res?.msg || '报名失败')
        }
      } catch (error) {
        console.error('报名失败', error)
        mMessage.error('报名失败')
      } finally {
        this.submitting = false
      }
    },

    async updateRegistration() {
      if (!this.validateEditForm()) return

      try {
        this.submitting = true
        // 转换为snake_case以匹配后端API
        const data = {
          status: 'pending'  // 重置为待审核
        }
        // 转换editFormData字段：camelCase -> snake_case
        const fieldMapping = {
          name: 'name',
          class: 'class',
          college: 'college',
          studentId: 'student_id',
          gender: 'gender',
          shirtSize: 'shirt_size',
          teamName: 'team_name',
          qq: 'qq'
        }
        for (const [camelKey, snakeKey] of Object.entries(fieldMapping)) {
          if (this.editFormData[camelKey] !== undefined && this.editFormData[camelKey] !== '') {
            data[snakeKey] = this.editFormData[camelKey]
          }
        }
        const res = await updateRegistration(this.myRegistration.id, data)

        if (res && res.code === 200) {
          mMessage.success('更新成功！')
          this.showEditForm = false
          await this.loadMyRegistration()
        } else {
          mMessage.error(res?.message || res?.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败', error)
        mMessage.error('更新失败')
      } finally {
        this.submitting = false
      }
    },

    validateForm() {
      for (const [field, enabled] of Object.entries(this.fieldConfig)) {
        if (enabled && this.isRequired(field) && !this.formData[field]) {
          mMessage.error(`请填写${this.getFieldLabel(field)}`)
          return false
        }
      }
      return true
    },

    validateEditForm() {
      for (const [field, enabled] of Object.entries(this.fieldConfig)) {
        if (enabled && this.isRequired(field) && !this.editFormData[field]) {
          mMessage.error(`请填写${this.getFieldLabel(field)}`)
          return false
        }
      }
      return true
    },

    getFieldLabel(field) {
      const labels = {
        name: '姓名',
        class: '班级',
        college: '学院',
        studentId: '学号',
        gender: '性别',
        shirtSize: 'T恤尺码',
        teamName: '队伍名称',
        qq: 'QQ号'
      }
      return labels[field] || field
    },

    getFieldPlaceholder(field) {
      return `请输入${this.getFieldLabel(field)}`
    },

    getInputType(field) {
      if (field === 'qq' || field === 'studentId') return 'text'
      return 'text'
    },

    isRequired(field) {
      // 所有启用的字段都是必填的
      return this.fieldConfig[field]
    },

    getRegistrationStatusClass() {
      if (!this.myRegistration) return ''
      return this.myRegistration.status
    },

    getRegistrationStatusText() {
      if (!this.myRegistration) return '未报名'
      if (this.myRegistration.status === 'pending') return '待审核'
      if (this.myRegistration.status === 'approved') return '报名成功'
      if (this.myRegistration.status === 'rejected') return '报名失败'
    },

    getStatusText(status) {
      if (status === 'pending') return '待审核'
      if (status === 'approved') return '报名成功'
      if (status === 'rejected') return '报名失败'
    },

    formatTime(str) {
      if (!str) return ''
      const date = new Date(str)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
        timeZone: 'Asia/Shanghai'
      })
    },

    goBack() {
      this.$router.push({ name: 'RegistrationList' })
    },

    getCurrentUser() {
      const userInfoStr = localStorage.getItem('userInfo')
      if (userInfoStr) {
        try {
          const hojUserInfo = JSON.parse(userInfoStr)
          return {
            uuid: hojUserInfo.uid || hojUserInfo.uuid,
            username: hojUserInfo.username
          }
        } catch (e) {
          return null
        }
      }
      return null
    },

    // 聊天相关方法
    openChat() {
      this.showChatModal = true
      this.loadChatMessages()
      // 保存当前查看时间
      this.saveLastViewTime()
      // 滚动到底部
      this.$nextTick(() => {
        this.scrollToBottom()
      })
    },

    closeChat() {
      // 关闭聊天窗口时，再次更新查看时间
      // 确保在聊天窗口期间看到的所有消息都标记为已读
      this.saveLastViewTime()
      this.showChatModal = false
    },

    async loadChatMessages() {
      if (!this.myRegistration) return

      try {
        const remark = this.myRegistration.remark || ''
        if (remark) {
          try {
            // 尝试解析为JSON数组（新格式）
            this.chatMessages = JSON.parse(remark)
          } catch (e) {
            // 如果解析失败，当作简单文本处理（旧格式）
            this.chatMessages = [{
              sender: 'admin',
              message: remark,
              timestamp: new Date().toISOString()
            }]
          }
        } else {
          this.chatMessages = []
        }
      } catch (error) {
        console.error('加载聊天消息失败', error)
        this.chatMessages = []
      }
    },

    async sendMessage() {
      if (!this.newMessage.trim() || !this.myRegistration) return

      try {
        this.sendingMessage = true
        const message = {
          sender: 'user',
          message: this.newMessage.trim(),
          timestamp: new Date().toISOString()
        }

        this.chatMessages.push(message)

        // 更新报名信息，保存消息到 remark 字段
        const data = {
          remark: JSON.stringify(this.chatMessages)
        }

        const res = await updateRegistration(this.myRegistration.id, data)
        if (res.code === 200) {
          this.newMessage = ''
          // 刷新报名信息
          await this.loadMyRegistration()
          // 滚动到底部
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        } else {
          mMessage.error(res.msg || '发送失败')
          // 移除刚才添加的消息
          this.chatMessages.pop()
        }
      } catch (error) {
        console.error('发送消息失败', error)
        mMessage.error('发送消息失败')
        // 移除刚才添加的消息
        this.chatMessages.pop()
      } finally {
        this.sendingMessage = false
      }
    },

    async saveLastViewTime() {
      if (!this.myRegistration) return

      try {
        const timestamp = new Date().toISOString()
        await updateRegistration(this.myRegistration.id, {
          last_view_time: timestamp
        })
        // 更新本地数据
        this.myRegistration.lastViewTime = timestamp
        this.calculateUnreadCount()
      } catch (error) {
        console.error('保存查看时间失败', error)
      }
    },

    calculateUnreadCount() {
      // 如果聊天窗口是打开的，未读计数为0
      if (this.showChatModal) {
        this.unreadCount = 0
        return
      }

      if (!this.myRegistration || !this.myRegistration.remark) {
        this.unreadCount = 0
        return
      }

      try {
        const messages = JSON.parse(this.myRegistration.remark)
        const lastViewTime = this.myRegistration.lastViewTime || this.myRegistration.created_at

        this.unreadCount = messages.filter(msg => {
          return msg.sender === 'admin' &&
                 new Date(msg.timestamp) > new Date(lastViewTime)
        }).length
      } catch (e) {
        // 如果解析失败，说明是旧格式的简单文本，没有未读消息
        this.unreadCount = 0
      }
    },

    handleCompositionStart() {
      this.isComposing = true
      this.justFinishedComposing = false
    },

    handleCompositionEnd() {
      this.isComposing = false
      this.compositionEndTime = Date.now()
      this.justFinishedComposing = true
      // 150ms后恢复常态
      setTimeout(() => {
        this.justFinishedComposing = false
      }, 150)
    },

    handleEnterKey(e) {
      const now = Date.now()
      // 如果正在输入法组合中，或者刚完成输入法组合（150ms内），不发送
      if (this.isComposing || (this.justFinishedComposing && now - this.compositionEndTime < 150)) {
        e.preventDefault()
        return
      }
      this.sendMessage()
    },

    scrollToBottom() {
      const container = this.$refs.chatMessagesContainer
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    },

    formatMessageTime(timestamp) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      })
    },

    // WebSocket 相关方法
    connectWebSocket() {
      if (!this.competition || !this.competition.id) {
        return
      }

      // 监听报名更新事件
      wsManager.on('registrationUpdate', async (registration) => {
        // 更新本地报名信息
        if (this.myRegistration && this.myRegistration.id === registration.id) {
          this.myRegistration = registration

          // 如果聊天窗口打开，重新加载聊天消息
          if (this.showChatModal) {
            const oldLength = this.chatMessages.length
            await this.loadChatMessages()

            // 如果有新消息，滚动到底部
            if (this.chatMessages.length > oldLength) {
              this.$nextTick(() => {
                this.scrollToBottom()
              })
            }
          } else {
            // 如果聊天窗口关闭，更新未读计数
            this.calculateUnreadCount()
          }
        }
      })

      // 监听比赛信息更新（当管理员修改比赛描述、图片等）
      wsManager.on('competitionUpdate', (competition) => {
        // 直接使用 WebSocket 广播的数据，无需重新加载
        if (this.competition && this.competition.id === competition.id) {
          // 使用 $set 确保 Vue 响应式更新
          this.$set(this, 'competition', competition)

          // 解析 fields 字段
          if (competition.fields) {
            try {
              const parsed = JSON.parse(competition.fields)
              this.$set(this, 'fieldConfig', parsed)
            } catch (e) {
              console.error('解析 fields 失败:', e)
            }
          }

          // 强制更新 Markdown 渲染
          this.$forceUpdate()

          mMessage.success('比赛信息已更新')
        }
      })

      // 连接 WebSocket
      wsManager.connect(this.competition.id)
    }
  },
  async mounted() {
    this.currentUser = this.getCurrentUser()
    if (!this.currentUser) {
      this.$router.replace({ path: '/home' })
      return
    }
    await this.loadCompetition()
    // 连接 WebSocket（实时接收管理员消息）
    this.connectWebSocket()
  },

  beforeDestroy() {
    // 断开 WebSocket 连接
    wsManager.disconnect()
  }
}
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.competition-detail-container {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #ffffff;
  min-height: 100vh;
  padding: 20px;
}

.container {
  max-width: 800px;
  margin: 0 auto;
}

.header {
  background: white;
  border-radius: 16px;
  padding: 30px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #409EFF;
  
  
}

.back-btn {
  padding: 10px 20px;
  color: #409EFF;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s;
}

.back-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(79, 172, 254, 0.4);
}

.card {
  background: white;
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  margin-bottom: 24px;
}

.card h2 {
  font-size: 24px;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 20px;
}

.card h3 {
  font-size: 20px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 16px;
  margin-top: 24px;
}

.info-row {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 15px;
}

.label {
  font-weight: 600;
  color: #4a5568;
  margin-right: 8px;
  min-width: 120px;
}

.logo-container {
  margin-top: 20px;
  text-align: center;
}

.logo {
  max-width: 200px;
  max-height: 200px;
  object-fit: contain;
  border-radius: 12px;
}

.description-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e2e8f0;
}

.description-content {
  line-height: 1.6;
  color: #4a5568;
}

.description-content >>> h1,
.description-content >>> h2,
.description-content >>> h3 {
  margin-top: 16px;
  margin-bottom: 8px;
  color: #2d3748;
}

.description-content >>> p {
  margin-bottom: 12px;
}

.description-content >>> ul,
.description-content >>> ol {
  margin-left: 20px;
  margin-bottom: 12px;
}

.description-content >>> img {
  /* 保持响应式，但不覆盖指定的宽度 */
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 12px 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 对于有指定宽度的图片，使用 HTML width 属性 */
.description-content >>> img[width] {
  /* 让 HTML 的 width 属性生效，不要覆盖它 */
  max-width: 100%;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 8px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 12px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.3s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #4facfe;
  box-shadow: 0 0 0 3px rgba(79, 172, 254, 0.1);
}

.required-mark {
  color: #e53e3e;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 16px;
  background: #409EFF;
  color: white;
}

.btn-primary {
  background: #409EFF;
  color: white;
}

.btn:hover:not(:disabled) {
  background: #66b1ff;
  transform: translateY(-2px);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.badge {
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
}

.badge.pending {
  background: #E6A23C;
  color: white;
}

.badge.approved {
  background: #67C23A;
  color: white;
}

.badge.rejected {
  background: #F56C6C;
  color: white;
}

.edit-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

/* 聊天按钮 */
.btn-chat {
  position: relative;
  padding: 8px 16px;
  background: #67C23A;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-chat:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.unread-badge {
  background: #e53e3e;
  color: white;
  border-radius: 10px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 700;
  min-width: 18px;
  text-align: center;
}

/* 聊天模态框 */
.chat-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.chat-modal-content {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 600px;
  height: 80vh;
  max-height: 700px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.chat-modal-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #67C23A;
  border-radius: 16px 16px 0 0;
  color: white;
}

.chat-modal-header h3 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.close-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 28px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: rotate(90deg);
}

.chat-modal-body {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f7fafc;
}

.no-messages {
  text-align: center;
  color: #a0aec0;
  padding: 40px 20px;
  font-size: 14px;
}

.chat-message {
  display: flex;
  margin-bottom: 16px;
  animation: messageSlide 0.3s ease-out;
}

@keyframes messageSlide {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-user {
  justify-content: flex-end;
}

.message-admin {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 12px;
  position: relative;
}

.message-user .message-bubble {
  background: #67C23A;
  color: white;
  border-bottom-right-radius: 4px;
}

.message-admin .message-bubble {
  background: white;
  color: #2d3748;
  border: 1px solid #e2e8f0;
  border-bottom-left-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.message-text {
  word-wrap: break-word;
  white-space: pre-wrap;
  line-height: 1.5;
  font-size: 14px;
}

.message-time {
  margin-top: 4px;
  font-size: 11px;
  opacity: 0.7;
}

.message-user .message-time {
  color: white;
  text-align: right;
}

.message-admin .message-time {
  color: #a0aec0;
}

.chat-input-area {
  padding: 16px;
  background: white;
  border-top: 1px solid #e2e8f0;
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.chat-input {
  flex: 1;
  padding: 12px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  transition: all 0.3s;
  max-height: 120px;
}

.chat-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.btn-send {
  padding: 12px 24px;
  background: #67C23A;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  height: fit-content;
}

.btn-send:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-send:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
