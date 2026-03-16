<template>
  <div class="registration-admin">
    <div class="container">
      <!-- 头部 -->
      <div class="header">
        <h1>📊 比赛报名管理系统</h1>
        <p>管理员控制面板</p>
      </div>

      <!-- 操作栏 -->
      <div class="action-bar">
        <router-link to="/admin/registration/create" class="btn btn-primary">
          + 创建新比赛
        </router-link>
      </div>

      <!-- 比赛列表 -->
      <div class="card">
        <h2>📋 比赛列表</h2>
        <div v-if="loading" style="text-align: center; padding: 40px;">
          <p>加载中...</p>
        </div>
        <div v-else-if="competitions.length === 0" style="text-align: center; padding: 40px; color: #718096;">
          <p>暂无比赛</p>
        </div>
        <table v-else class="competition-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>比赛名称</th>
              <th>报名时间</th>
              <th>报名人数</th>
              <th>可见性</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="comp in competitions" :key="comp.id">
              <td>{{ comp.id }}</td>
              <td>{{ comp.name }}</td>
              <td>
                <div>{{ formatTime(comp.startTime) }}</div>
                <div style="color: #718096; font-size: 12px;">至 {{ formatTime(comp.endTime) }}</div>
              </td>
              <td>
                <router-link
                  :to="{
                    path: `/admin/registration/list/${comp.id}`,
                    query: { competition: JSON.stringify(comp) }
                  }"
                  class="btn-link"
                >
                  {{ getRegistrationCount(comp.id) }} 人
                </router-link>
              </td>
              <td>
                <span class="badge" :class="{ 'badge-visible': comp.visible, 'badge-hidden': !comp.visible }">
                  {{ comp.visible ? '可见' : '隐藏' }}
                </span>
              </td>
              <td>
                <div class="action-buttons">
                  <router-link :to="`/admin/registration/edit/${comp.id}`" class="btn-small">编辑</router-link>
                  <router-link
                    :to="{
                      path: `/admin/registration/list/${comp.id}`,
                      query: { competition: JSON.stringify(comp) }
                    }"
                    class="btn-small"
                  >审核报名</router-link>
                  <button class="btn-small btn-danger" @click="confirmDelete(comp)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页组件 -->
        <div v-if="!loading && competitions.length > 0" class="pagination-container">
          <div class="pagination-info">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, total) }} 条，共 {{ total }} 条
          </div>
          <div class="pagination-controls">
            <button
              class="pagination-btn"
              :disabled="currentPage === 1"
              @click="handlePageChange(currentPage - 1)"
            >
              上一页
            </button>

            <template v-for="page in getPageNumbers()" :key="page">
              <span v-if="page === '...'" class="pagination-ellipsis">...</span>
              <button
                v-else
                class="pagination-btn"
                :class="{ 'pagination-btn-active': page === currentPage }"
                @click="handlePageChange(page)"
              >
                {{ page }}
              </button>
            </template>

            <button
              class="pagination-btn"
              :disabled="currentPage === getTotalPages()"
              @click="handlePageChange(currentPage + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </div>

      <!-- 聊天模态框 -->
      <div v-if="showChatModal" class="chat-modal" @click.self="closeChat">
        <div class="chat-modal-content">
          <div class="chat-modal-header">
            <h3>💬 与用户沟通 - {{ currentChatUser ? currentChatUser.name : '' }}</h3>
            <button class="close-btn" @click="closeChat">×</button>
          </div>
          <div class="chat-modal-body">
            <!-- 消息列表 -->
            <div class="chat-messages" ref="chatMessagesContainer">
              <div v-if="chatMessages.length === 0" class="no-messages">
                暂无消息，开始与用户沟通吧
              </div>
              <div
                v-for="(msg, index) in chatMessages"
                :key="index"
                class="chat-message"
                :class="{ 'message-admin': msg.sender === 'admin', 'message-user': msg.sender === 'user' }"
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
  </div>
</template>

<script>
import {
  adminGetCompetitions,
  adminCreateCompetition,
  adminUpdateCompetition,
  adminDeleteCompetition,
  adminGetRegistrations,
  adminUpdateRegistrationStatus
} from '@/api/registration'
import mMessage from '@/common/message'

export default {
  name: 'RegistrationAdmin',
  data() {
    return {
      loading: false,
      competitions: [],
      allCompetitions: [], // 存储所有比赛数据
      registrationCounts: {},

      // 分页相关
      currentPage: 1,
      pageSize: 20,
      total: 0,

      // 聊天相关 - 保留以便快速查看聊天
      showChatModal: false,
      currentChatUser: null,
      chatMessages: [],
      newMessage: '',
      sendingMessage: false,
      isComposing: false,
      justFinishedComposing: false,
      compositionEndTime: 0,
      pollingTimer: null
    }
  },
  methods: {
    async loadCompetitions() {
      try {
        this.loading = true
        const res = await adminGetCompetitions({ show_hidden: 'true' })
        if (res.code === 200) {
          // 保存所有数据
          this.allCompetitions = res.data
          this.total = res.data.length
          // 更新当前页显示的数据
          this.updatePageData()
          // 加载当前页比赛的报名数量
          for (const comp of this.competitions) {
            this.loadRegistrationCount(comp.id)
          }
        } else {
          mMessage.error(res.msg || '加载失败')
        }
      } catch (error) {
        console.error('加载比赛失败', error)
        mMessage.error('加载比赛失败')
      } finally {
        this.loading = false
      }
    },

    // 更新当前页数据
    updatePageData() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      this.competitions = this.allCompetitions.slice(start, end)
    },

    // 切换页码
    handlePageChange(page) {
      this.currentPage = page
      this.updatePageData()
      // 加载新页的报名数量
      for (const comp of this.competitions) {
        if (!this.registrationCounts[comp.id]) {
          this.loadRegistrationCount(comp.id)
        }
      }
      // 滚动到顶部
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },

    // 计算总页数
    getTotalPages() {
      return Math.ceil(this.total / this.pageSize)
    },

    // 生成分页按钮数组
    getPageNumbers() {
      const totalPages = this.getTotalPages()
      const current = this.currentPage
      const pages = []

      if (totalPages <= 7) {
        // 总页数小于等于7，显示所有页码
        for (let i = 1; i <= totalPages; i++) {
          pages.push(i)
        }
      } else {
        // 总页数大于7，显示部分页码
        if (current <= 4) {
          // 当前页在前面
          for (let i = 1; i <= 5; i++) {
            pages.push(i)
          }
          pages.push('...')
          pages.push(totalPages)
        } else if (current >= totalPages - 3) {
          // 当前页在后面
          pages.push(1)
          pages.push('...')
          for (let i = totalPages - 4; i <= totalPages; i++) {
            pages.push(i)
          }
        } else {
          // 当前页在中间
          pages.push(1)
          pages.push('...')
          for (let i = current - 1; i <= current + 1; i++) {
            pages.push(i)
          }
          pages.push('...')
          pages.push(totalPages)
        }
      }

      return pages
    },

    async loadRegistrationCount(competitionId) {
      try {
        const res = await adminGetRegistrations(competitionId)
        if (res.code === 200) {
          this.$set(this.registrationCounts, competitionId, res.data.length)
        }
      } catch (error) {
        // 忽略错误
      }
    },

    getRegistrationCount(competitionId) {
      return this.registrationCounts[competitionId] || 0
    },

    confirmDelete(comp) {
      if (confirm(`确定要删除比赛 "${comp.name}" 吗？此操作将同时删除所有报名记录，无法恢复。`)) {
        this.deleteCompetition(comp)
      }
    },

    async deleteCompetition(comp) {
      try {
        const res = await adminDeleteCompetition(comp.id)
        if (res.code === 200) {
          mMessage.success('删除成功')
          // 检查当前页是否还有数据，如果没有则跳转到上一页
          const totalPages = Math.ceil((this.total - 1) / this.pageSize)
          if (this.currentPage > totalPages && this.currentPage > 1) {
            this.currentPage = totalPages
          }
          await this.loadCompetitions()
        } else {
          mMessage.error(res.msg || '删除失败')
        }
      } catch (error) {
        console.error('删除失败', error)
        mMessage.error('删除失败')
      }
    },

    async updateStatus(reg, status) {
      try {
        const res = await adminUpdateRegistrationStatus(reg.id, { status })
        if (res.code === 200) {
          mMessage.success('更新成功')
        } else {
          mMessage.error(res.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败', error)
        mMessage.error('更新失败')
      }
    },

    viewDetail(reg) {
      // 可以打开详情对话框
      alert(`姓名: ${reg.name}\n班级: ${reg.class}\n学院: ${reg.college}\n学号: ${reg.studentId}`)
    },

    getStatusText(status) {
      if (status === 'pending') return '待审核'
      if (status === 'approved') return '已通过'
      if (status === 'rejected') return '已退回'
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

    formatDateTimeLocal(str) {
      if (!str) return ''
      const date = new Date(str)
      // 转换为 datetime-local 格式
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day}T${hours}:${minutes}`
    },

    // 聊天相关方法
    openChat(registration) {
      this.currentChatUser = registration
      this.showChatModal = true
      this.loadChatMessages()
      // 保存管理员查看时间
      this.saveAdminLastViewTime()
      // 开始轮询新消息
      this.startPolling()
      // 滚动到底部
      this.$nextTick(() => {
        this.scrollToBottom()
      })
    },

    closeChat() {
      // 关闭聊天窗口时，再次更新查看时间
      // 确保在聊天窗口期间看到的所有消息都标记为已读
      if (this.currentChatUser) {
        this.saveAdminLastViewTime()
      }
      this.showChatModal = false
      this.currentChatUser = null
      this.stopPolling()
    },

    async loadChatMessages() {
      if (!this.currentChatUser) return

      try {
        const remark = this.currentChatUser.remark || ''
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
      if (!this.newMessage.trim() || !this.currentChatUser) return

      try {
        this.sendingMessage = true
        const message = {
          sender: 'admin',
          message: this.newMessage.trim(),
          timestamp: new Date().toISOString()
        }

        this.chatMessages.push(message)

        // 更新报名信息，保存消息到 remark 字段
        const data = {
          remark: JSON.stringify(this.chatMessages)
        }

        const res = await adminUpdateRegistrationStatus(this.currentChatUser.id, data)
        if (res.code === 200) {
          this.newMessage = ''
          // 刷新报名信息
          await this.viewRegistrations(this.viewingCompetition)
          // 更新当前用户引用
          const updatedReg = this.registrations.find(r => r.id === this.currentChatUser.id)
          if (updatedReg) {
            this.currentChatUser = updatedReg
          }
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

    async saveAdminLastViewTime() {
      if (!this.currentChatUser || !this.currentChatUser.id) return

      try {
        const timestamp = new Date().toISOString()
        const data = {
          admin_last_view_time: timestamp
        }
        await adminUpdateRegistrationStatus(this.currentChatUser.id, data)
        // 更新本地数据 - 使用 Vue.set 确保响应式
        this.$set(this.currentChatUser, 'adminLastViewTime', timestamp)

        // 同时更新 registrations 数组中的对应项
        const index = this.registrations.findIndex(r => r.id === this.currentChatUser.id)
        if (index !== -1) {
          this.$set(this.registrations[index], 'adminLastViewTime', timestamp)
        }
      } catch (error) {
        console.error('保存管理员查看时间失败', error)
      }
    },

    getAdminUnreadCount(registration) {
      // 如果聊天窗口打开的是当前用户，未读计数为0
      if (this.showChatModal && this.currentChatUser && this.currentChatUser.id === registration.id) {
        return 0
      }

      if (!registration || !registration.remark) {
        return 0
      }

      try {
        const messages = JSON.parse(registration.remark)
        const lastViewTime = registration.adminLastViewTime || registration.created_at

        return messages.filter(msg => {
          return msg.sender === 'user' &&
                 new Date(msg.timestamp) > new Date(lastViewTime)
        }).length
      } catch (e) {
        // 如果解析失败，说明是旧格式的简单文本，没有未读消息
        return 0
      }
    },

    handleCompositionStart() {
      this.isComposing = true
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

    startPolling() {
      // 每5秒轮询一次新消息
      this.pollingTimer = setInterval(async () => {
        if (this.showChatModal && this.currentChatUser && this.viewingCompetition) {
          const oldLength = this.chatMessages.length
          // 只重新加载聊天消息，不重新加载整个列表（避免闪烁）
          await this.loadChatMessages()
          // 只有在有新消息时才更新查看时间和滚动
          if (this.chatMessages.length > oldLength) {
            // 更新查看时间，这样新消息不会被标记为未读
            this.saveAdminLastViewTime()
            this.$nextTick(() => {
              this.scrollToBottom()
            })
          }
        }
      }, 5000)
    },

    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }
    }
  },
  async mounted() {
    await this.loadCompetitions()
  },

  beforeDestroy() {
    this.stopPolling()
  }
}
</script>

<style scoped>
.registration-admin {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #ffffff;
  min-height: 100vh;
  padding: 20px;
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  color: #2d3748;
  padding: 30px;
  border-radius: 16px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.header h1 {
  font-size: 32px;
  font-weight: 700;
  color: #409EFF;
  
  
  margin-bottom: 8px;
}

.header p {
  color: #718096;
  font-size: 14px;
}

.action-bar {
  margin-bottom: 24px;
}

.card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 28px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.card h2 {
  font-size: 24px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 24px;
}

.competition-table,
.registration-table {
  width: 100%;
  border-collapse: collapse;
}

.competition-table th,
.competition-table td,
.registration-table th,
.registration-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.competition-table th,
.registration-table th {
  background: #f7fafc;
  font-weight: 600;
  color: #2d3748;
}

.competition-table tr:hover,
.registration-table tr:hover {
  background: #f7fafc;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
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

.btn-link {
  background: none;
  border: none;
  color: #4facfe;
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
}

.btn-small {
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 4px;
  background: #409EFF;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-danger {
  background: #fc8181;
  color: white;
}

.btn-success {
  background: #68d391;
  color: white;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.badge-visible {
  background: #d1fae5;
  color: #065f46;
}

.badge-hidden {
  background: #e2e8f0;
  color: #4a5568;
}

.badge-pending {
  background: #E6A23C;
  color: white;
}

.badge-approved {
  background: #409EFF;
  color: white;
}

.badge-rejected {
  background: #F56C6C;
  color: white;
}

/* 分页样式 */
.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
  flex-wrap: wrap;
  gap: 16px;
}

.pagination-info {
  color: #718096;
  font-size: 14px;
}

.pagination-controls {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.pagination-btn {
  min-width: 36px;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #4a5568;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pagination-btn:hover:not(:disabled) {
  background: #409EFF;
  color: white;
  border-color: #409EFF;
  transform: translateY(-1px);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #f7fafc;
}

.pagination-btn-active {
  background: #409EFF !important;
  color: white !important;
  border-color: #409EFF !important;
}

.pagination-ellipsis {
  color: #a0aec0;
  font-size: 14px;
  padding: 0 4px;
}

/* 响应式分页 */
@media (max-width: 768px) {
  .pagination-container {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .pagination-info {
    order: 2;
  }

  .pagination-controls {
    order: 1;
  }

  .pagination-btn {
    min-width: 32px;
    height: 32px;
    padding: 0 8px;
    font-size: 13px;
  }
}

/* Modal */
.modal {
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
}

.modal-content {
  background: white;
  border-radius: 16px;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow: auto;
}

.large-modal {
  max-width: 1200px;
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  font-size: 20px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
}

.checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #f7fafc;
  border-radius: 8px;
  cursor: pointer;
}

.checkbox-item input[type="checkbox"] {
  width: auto;
  cursor: pointer;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

/* 聊天按钮 */
.btn-chat {
  position: relative;
  padding: 6px 12px;
  background: #409EFF;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-chat:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.4);
}

.unread-badge {
  background: #e53e3e;
  color: white;
  border-radius: 8px;
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  min-width: 16px;
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
  background: #409EFF;
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

.message-admin {
  justify-content: flex-end;
}

.message-user {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 12px;
  position: relative;
}

.message-admin .message-bubble {
  background: #409EFF;
  color: white;
  border-bottom-right-radius: 4px;
}

.message-user .message-bubble {
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

.message-admin .message-time {
  color: white;
  text-align: right;
}

.message-user .message-time {
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
  background: #409EFF;
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
