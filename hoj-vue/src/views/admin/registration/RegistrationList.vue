<template>
  <div class="registration-list-page">
    <div class="container">
      <!-- 头部 -->
      <div class="header">
        <div>
          <h1>{{ competition ? competition.name : '加载中...' }} - 报名列表</h1>
          <p>管理比赛报名信息</p>
        </div>
        <div class="header-actions">
          <!-- 导出Excel按钮 -->
          <button type="button" class="btn btn-success" @click="exportToExcel" :disabled="exporting || registrations.length === 0">
            {{ exporting ? '导出中...' : '📊 导出Excel' }}
          </button>
          <router-link to="/admin/registration" class="btn">返回</router-link>
        </div>
      </div>

      <!-- 统计信息 -->
      <div class="stats-section">
        <div class="stat-card">
          <div class="stat-number">{{ registrations.length }}</div>
          <div class="stat-label">总报名数</div>
        </div>
        <div class="stat-card">
          <div class="stat-number pending">{{ getStatusCount('pending') }}</div>
          <div class="stat-label">待审核</div>
        </div>
        <div class="stat-card">
          <div class="stat-number approved">{{ getStatusCount('approved') }}</div>
          <div class="stat-label">已通过</div>
        </div>
        <div class="stat-card">
          <div class="stat-number rejected">{{ getStatusCount('rejected') }}</div>
          <div class="stat-label">已退回</div>
        </div>
      </div>

      <!-- 报名列表 -->
      <div class="card">
        <div class="card-header">
          <h2>
            📋 报名成员
            <span v-if="totalUnreadCount > 0" class="total-unread-badge">
              {{ totalUnreadCount > 99 ? '99+' : totalUnreadCount }}
            </span>
          </h2>
        </div>

        <!-- 筛选区域 -->
        <div class="filter-section">
          <div class="filter-row">
            <div class="filter-item">
              <label>姓名：</label>
              <input v-model="filterName" placeholder="输入姓名筛选" class="filter-input">
            </div>
            <div class="filter-item">
              <label>学号：</label>
              <input v-model="filterStudentId" placeholder="输入学号筛选" class="filter-input">
            </div>
            <div class="filter-item">
              <label>审核状态：</label>
              <select v-model="filterStatus" class="filter-select">
                <option value="">全部</option>
                <option value="pending">待审核</option>
                <option value="approved">已通过</option>
                <option value="rejected">已退回</option>
              </select>
            </div>
            <button class="btn-small" @click="clearFilters">清空筛选</button>
          </div>
          <div class="filter-info">
            共找到 <strong>{{ filteredRegistrations.length }}</strong> 条记录
          </div>
        </div>

        <div v-if="loading" style="text-align: center; padding: 40px;">
          <p>加载中...</p>
        </div>

        <div v-else-if="filteredRegistrations.length === 0" style="text-align: center; padding: 40px; color: #718096;">
          <p>{{ registrations.length === 0 ? '暂无报名' : '没有找到符合条件的记录' }}</p>
        </div>

        <table v-else class="registration-table">
          <thead>
            <tr>
              <th>姓名</th>
              <th>班级</th>
              <th>学院</th>
              <th>学号</th>
              <th>性别</th>
              <th>队伍</th>
              <th>QQ</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="reg in paginatedRegistrations" :key="reg.id">
              <td>{{ reg.name }}</td>
              <td>{{ reg.class }}</td>
              <td>{{ reg.college }}</td>
              <td>{{ reg.studentId }}</td>
              <td>{{ reg.gender }}</td>
              <td>{{ reg.teamName }}</td>
              <td>{{ reg.qq }}</td>
              <td>
                <span class="badge" :class="'badge-' + reg.status">
                  {{ getStatusText(reg.status) }}
                </span>
              </td>
              <td>
                <div class="action-buttons">
                  <button v-if="reg.status === 'pending'" class="btn-small btn-success" @click="updateStatus(reg, 'approved')">通过</button>
                  <button v-if="reg.status === 'pending'" class="btn-small btn-danger" @click="updateStatus(reg, 'rejected')">退回</button>
                  <button class="btn-small" @click="viewDetail(reg)">详情</button>
                  <button class="btn-small btn-chat" @click="openChat(reg)">
                    💬 沟通
                    <span v-if="getAdminUnreadCount(reg) > 0" class="unread-badge">{{ getAdminUnreadCount(reg) }}</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div v-if="filteredRegistrations.length > 0" class="pagination-container">
          <div class="pagination-info">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, filteredRegistrations.length) }} 条，共 {{ filteredRegistrations.length }} 条
          </div>
          <div class="pagination-controls">
            <button class="page-btn" @click="currentPage = 1" :disabled="currentPage === 1">首页</button>
            <button class="page-btn" @click="currentPage--" :disabled="currentPage === 1">上一页</button>
            <template v-for="page in pageNumbers">
              <button v-if="page === '...'" class="page-btn page-ellipsis" disabled>...</button>
              <button v-else class="page-btn" :class="{ active: page === currentPage }" @click="currentPage = page">{{ page }}</button>
            </template>
            <button class="page-btn" @click="currentPage++" :disabled="currentPage === totalPages">下一页</button>
            <button class="page-btn" @click="currentPage = totalPages" :disabled="currentPage === totalPages">末页</button>
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
import { adminGetRegistrations, adminUpdateRegistrationStatus } from '@/api/registration'
import mMessage from '@/common/message'
import { wsManager } from '@/utils/websocket'

export default {
  name: 'RegistrationList',
  data() {
    return {
      competition: null,
      competitionId: null,
      loading: false,
      registrations: [],
      exporting: false,

      // 筛选相关
      filterName: '',
      filterStudentId: '',
      filterStatus: '',

      // 分页相关
      currentPage: 1,
      pageSize: 20,

      // 聊天相关
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
  computed: {
    filteredRegistrations() {
      return this.registrations.filter(reg => {
        const matchName = !this.filterName || reg.name.toLowerCase().includes(this.filterName.toLowerCase())
        const matchStudentId = !this.filterStudentId || reg.studentId.toLowerCase().includes(this.filterStudentId.toLowerCase())
        const matchStatus = !this.filterStatus || reg.status === this.filterStatus
        return matchName && matchStudentId && matchStatus
      })
    },

    // 计算总未读消息数
    totalUnreadCount() {
      return this.registrations.reduce((total, reg) => {
        return total + this.getAdminUnreadCount(reg)
      }, 0)
    },

    // 总页数
    totalPages() {
      return Math.ceil(this.filteredRegistrations.length / this.pageSize)
    },

    // 当前页显示的数据
    paginatedRegistrations() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return this.filteredRegistrations.slice(start, end)
    },

    // 页码列表
    pageNumbers() {
      const pages = []
      const total = this.totalPages

      if (total <= 7) {
        // 总页数小于等于7，显示所有页码
        for (let i = 1; i <= total; i++) {
          pages.push(i)
        }
      } else {
        // 总页数大于7，使用省略号
        if (this.currentPage <= 4) {
          // 当前页靠近开头
          for (let i = 1; i <= 5; i++) {
            pages.push(i)
          }
          pages.push('...')
          pages.push(total)
        } else if (this.currentPage >= total - 3) {
          // 当前页靠近末尾
          pages.push(1)
          pages.push('...')
          for (let i = total - 4; i <= total; i++) {
            pages.push(i)
          }
        } else {
          // 当前页在中间
          pages.push(1)
          pages.push('...')
          for (let i = this.currentPage - 1; i <= this.currentPage + 1; i++) {
            pages.push(i)
          }
          pages.push('...')
          pages.push(total)
        }
      }

      return pages
    }
  },
  watch: {
    // 监听筛选条件变化，重置到第一页
    filterName() {
      this.currentPage = 1
    },
    filterStudentId() {
      this.currentPage = 1
    },
    filterStatus() {
      this.currentPage = 1
    }
  },
  mounted() {
    // 从路由参数获取比赛ID
    this.competitionId = this.$route.params.id
    if (this.competitionId) {
      this.loadCompetition()
      this.loadRegistrations()
      this.connectWebSocket()
    }
  },
  beforeDestroy() {
    this.disconnectWebSocket()
  },
  methods: {
    async loadCompetition() {
      try {
        // 从路由查询参数或localStorage获取比赛信息
        const competitionData = this.$route.query.competition
        if (competitionData) {
          this.competition = JSON.parse(competitionData)
        } else {
          // 如果没有传递比赛信息，返回列表页
          mMessage.error('缺少比赛信息')
          this.$router.push('/admin/registration')
        }
      } catch (error) {
        console.error('加载比赛信息失败:', error)
        mMessage.error('加载比赛信息失败')
        this.$router.push('/admin/registration')
      }
    },

    async loadRegistrations() {
      if (!this.competitionId) return

      try {
        this.loading = true
        const res = await adminGetRegistrations(this.competitionId)
        if (res && res.code === 200) {
          this.registrations = res.data || []
        } else {
          mMessage.error(res?.message || res?.msg || '加载报名列表失败')
        }
      } catch (error) {
        console.error('加载报名列表失败', error)
        mMessage.error('加载报名列表失败')
      } finally {
        this.loading = false
      }
    },

    getStatusCount(status) {
      return this.registrations.filter(r => r.status === status).length
    },

    clearFilters() {
      this.filterName = ''
      this.filterStudentId = ''
      this.filterStatus = ''
    },

    async updateStatus(reg, status) {
      try {
        const res = await adminUpdateRegistrationStatus(reg.id, { status })
        if (res && res.code === 200) {
          mMessage.success('更新成功')
          // 直接更新本地数据
          const index = this.registrations.findIndex(r => r.id === reg.id)
          if (index !== -1) {
            this.$set(this.registrations[index], 'status', status)
          }
        } else {
          mMessage.error(res?.message || res?.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败', error)
        mMessage.error('更新失败')
      }
    },

    viewDetail(reg) {
      alert(`姓名: ${reg.name}\n班级: ${reg.class}\n学院: ${reg.college}\n学号: ${reg.studentId}`)
    },

    getStatusText(status) {
      if (status === 'pending') return '待审核'
      if (status === 'approved') return '已通过'
      if (status === 'rejected') return '已退回'
    },

    // 聊天相关方法
    openChat(registration) {
      this.currentChatUser = registration
      this.showChatModal = true

      // 先加载消息，然后再标记为已读
      this.loadChatMessages().then(() => {
        // 加载完消息后保存查看时间（标记为已读）
        this.saveAdminLastViewTime()
      })

      this.$nextTick(() => {
        this.scrollToBottom()
      })
    },

    closeChat() {
      if (this.currentChatUser) {
        // 关闭聊天窗口时保存查看时间
        this.saveAdminLastViewTime()
      }
      this.showChatModal = false
      this.currentChatUser = null
      this.chatMessages = []
      // 不要停止轮询，需要在后台继续更新未读消息
    },

    async loadChatMessages() {
      if (!this.currentChatUser) return

      try {
        const remark = this.currentChatUser.remark || ''
        if (remark) {
          try {
            this.chatMessages = JSON.parse(remark)
          } catch (e) {
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
      if (!this.newMessage.trim() || !this.currentChatUser) {
        return
      }

      const messageText = this.newMessage.trim()
      this.sendingMessage = true

      try {
        // 添加管理员消息
        const adminMessage = {
          sender: 'admin',
          message: messageText,
          timestamp: new Date().toISOString()
        }
        this.chatMessages.push(adminMessage)

        // 保存到remark字段
        const data = {
          remark: JSON.stringify(this.chatMessages)
        }
        const userId = this.currentChatUser.id

        const res = await adminUpdateRegistrationStatus(userId, data)

        if (res && res.code === 200) {

          this.newMessage = ''
          this.scrollToBottom()

          // 更新本地registrations中的remark
          const index = this.registrations.findIndex(r => r.id === userId)

          if (index !== -1) {
            this.$set(this.registrations[index], 'remark', data.remark)

            // 同时更新 currentChatUser（如果它是指向同一个对象）
            if (this.currentChatUser && this.currentChatUser.id === userId) {
              this.$set(this.currentChatUser, 'remark', data.remark)
            }
          } else {
            console.warn('⚠️ 在registrations数组中找不到用户, userId:', userId)
          }
        } else {
          // 失败，回滚
          console.error('❌ 消息发送失败, 响应:', res)
          this.chatMessages.pop()
          mMessage.error(res?.message || res?.msg || '发送失败')
        }
      } catch (error) {
        console.error('❌ 发送消息异常:', error)
        this.chatMessages.pop()
        mMessage.error('发送失败')
      } finally {
        this.sendingMessage = false
      }
    },

    handleEnterKey() {
      // 如果正在输入中文，不发送
      if (this.isComposing || this.justFinishedComposing) {
        this.justFinishedComposing = false
        return
      }
      this.sendMessage()
    },

    handleCompositionStart() {
      this.isComposing = true
    },

    handleCompositionEnd() {
      this.isComposing = false
      this.compositionEndTime = Date.now()
      this.justFinishedComposing = true
      setTimeout(() => {
        this.justFinishedComposing = false
      }, 100)
    },

    scrollToBottom() {
      const container = this.$refs.chatMessagesContainer
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    },

    async saveAdminLastViewTime() {
      if (!this.currentChatUser || !this.currentChatUser.id) {
        return
      }

      try {
        const timestamp = new Date().toISOString()
        const data = {
          admin_last_view_time: timestamp
        }
        const userId = this.currentChatUser.id


        const res = await adminUpdateRegistrationStatus(userId, data)

        if (res && res.code === 200) {
          // 安全地更新 currentChatUser
          if (this.currentChatUser && this.currentChatUser.id === userId) {
            this.$set(this.currentChatUser, 'adminLastViewTime', timestamp)
          }

          // 更新 registrations 数组中的数据
          const index = this.registrations.findIndex(r => r.id === userId)
          if (index !== -1) {
            this.$set(this.registrations[index], 'adminLastViewTime', timestamp)
          }

          // 同时保存到 localStorage 作为备份
          localStorage.setItem(`lastView_${userId}`, timestamp)
        } else {
          console.warn('保存管理员查看时间失败', res)
        }
      } catch (error) {
        console.error('保存管理员查看时间失败', error)
      }
    },

    getAdminUnreadCount(registration) {
      // 如果正在查看这个用户的聊天，不显示未读数
      if (this.showChatModal && this.currentChatUser && this.currentChatUser.id === registration.id) {
        return 0
      }

      if (!registration || !registration.remark) {
        return 0
      }

      try {
        const messages = JSON.parse(registration.remark)

        // 获取管理员查看时间，优先使用后端返回的，其次使用 localStorage 的
        let adminViewTime = registration.adminLastViewTime

        if (!adminViewTime || adminViewTime === '0' || adminViewTime === null || adminViewTime === '0001-01-01 00:00:00') {
          // 尝试从 localStorage 读取
          const stored = localStorage.getItem(`lastView_${registration.id}`)
          if (stored && stored !== '0' && stored !== 'null') {
            adminViewTime = stored
          } else {
            // 如果都没有，说明是第一次查看，所有用户消息都算未读
            const userMessages = messages.filter(msg => msg.sender === 'user')
            return userMessages.length
          }
        }

        const viewTime = new Date(adminViewTime)

        // 计算未读消息
        const unreadMessages = messages.filter(msg => {
          if (msg.sender !== 'user') return false
          const msgTime = new Date(msg.timestamp)
          return msgTime > viewTime
        })

        return unreadMessages.length
      } catch (e) {
        console.error('计算未读消息失败', registration.id, e)
        return 0
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

    connectWebSocket() {
      // 连接 WebSocket
      wsManager.connect(this.competitionId)

      // 监听连接成功
      wsManager.on('connected', () => {
      })

      // 监听报名更新
      wsManager.on('registrationUpdate', (registration) => {

        // 更新或添加报名记录
        const index = this.registrations.findIndex(r => r.id === registration.id)
        if (index !== -1) {
          // 更新现有记录
          this.$set(this.registrations, index, registration)

          // 如果正在查看该用户的聊天，同步更新
          if (this.currentChatUser && this.currentChatUser.id === registration.id) {
            this.$set(this, 'currentChatUser', registration)

            // 如果聊天窗口打开，重新加载消息
            if (this.showChatModal) {
              this.loadChatMessages().then(() => {
                this.$nextTick(() => {
                  this.scrollToBottom()
                })
              })
            }
          }
        } else {
          // 新增报名记录
          this.registrations.push(registration)
        }
      })

      // 监听新消息通知
      wsManager.on('newMessage', (data) => {

        // 重新加载报名列表以获取最新数据
        this.loadRegistrations()
      })

      // 监听连接错误
      wsManager.on('error', (error) => {
        console.error('❌ WebSocket 错误:', error)
      })

      // 监听断开连接
      wsManager.on('disconnected', (event) => {
        console.warn('🔌 WebSocket 已断开:', event)
      })
    },

    disconnectWebSocket() {
      // 移除所有事件监听器
      wsManager.off('connected')
      wsManager.off('registrationUpdate')
      wsManager.off('newMessage')
      wsManager.off('error')
      wsManager.off('disconnected')

      // 断开连接
      wsManager.disconnect()
    },

    // 导出 Excel
    async exportToExcel() {
      if (!this.competition || !this.competition.fields) {
        mMessage.error('比赛信息不完整')
        return
      }

      try {
        this.exporting = true

        // 动态导入 xlsx 库
        const XLSX = await import('xlsx')

        // 解析字段配置
        const fieldsConfig = typeof this.competition.fields === 'string'
          ? JSON.parse(this.competition.fields)
          : this.competition.fields

        // 定义字段标签映射
        const fieldLabels = {
          name: '姓名',
          class: '班级',
          college: '学院',
          studentId: '学号',
          gender: '性别',
          shirtSize: 'T恤尺码',
          teamName: '队伍名称',
          qq: 'QQ号'
        }

        // 状态映射
        const statusMap = {
          pending: '待审核',
          approved: '已通过',
          rejected: '已退回'
        }

        // 构建表头
        const headers = ['序号', '姓名', '学号', '性别', '班级', '学院', '队伍名称', 'T恤尺码', 'QQ号', '状态', '报名时间', '审核时间']

        // 构建数据行
        const dataRows = this.filteredRegistrations.map((reg, index) => {
          const row = [
            index + 1,
            reg.name || '',
            reg.studentId || '',
            reg.gender || '',
            reg.class || '',
            reg.college || '',
            reg.teamName || '',
            reg.shirtSize || '',
            reg.qq || '',
            statusMap[reg.status] || reg.status,
            this.formatTime(reg.created_at),
            reg.updatedAt ? this.formatTime(reg.updatedAt) : ''
          ]
          return row
        })

        // 组合表头和数据
        const worksheetData = [headers, ...dataRows]

        // 创建工作表
        const worksheet = XLSX.utils.aoa_to_sheet(worksheetData)

        // 设置列宽
        const colWidths = [
          { wch: 6 },  // 序号
          { wch: 12 }, // 姓名
          { wch: 15 }, // 学号
          { wch: 6 },  // 性别
          { wch: 15 }, // 班级
          { wch: 20 }, // 学院
          { wch: 20 }, // 队伍名称
          { wch: 10 }, // T恤尺码
          { wch: 15 }, // QQ号
          { wch: 10 }, // 状态
          { wch: 20 }, // 报名时间
          { wch: 20 }  // 审核时间
        ]
        worksheet['!cols'] = colWidths

        // 创建工作簿
        const workbook = XLSX.utils.book_new()
        XLSX.utils.book_append_sheet(workbook, worksheet, '报名列表')

        // 生成文件名
        const fileName = `${this.competition.name}_报名列表_${new Date().toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/[/:]\s/g, '-')}.xlsx`

        // 下载文件
        XLSX.writeFile(workbook, fileName)

        mMessage.success('导出成功！')
      } catch (error) {
        console.error('导出Excel失败:', error)
        mMessage.error('导出失败：' + (error.message || '未知错误'))
      } finally {
        this.exporting = false
      }
    },

    // 格式化时间
    formatTime(str) {
      if (!str) return ''
      const date = new Date(str)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      })
    }
  }
}
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.registration-list-page {
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
  border-radius: 16px;
  padding: 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
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

.header-actions {
  display: flex;
  gap: 12px;
}

.stats-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.stat-number {
  font-size: 36px;
  font-weight: 700;
  color: #409EFF;
  margin-bottom: 8px;
}

.stat-number.pending {
  color: #E6A23C;
}

.stat-number.approved {
  color: #67C23A;
}

.stat-number.rejected {
  color: #F56C6C;
}

.stat-label {
  color: #718096;
  font-size: 14px;
}

.card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 24px;
}

.card-header {
  margin-bottom: 20px;
}

.card-header h2 {
  font-size: 20px;
  font-weight: 600;
  color: #2d3748;
}

.filter-section {
  margin-bottom: 20px;
  padding: 16px;
  background: #f7fafc;
  border-radius: 8px;
}

.filter-row {
  display: flex;
  gap: 16px;
  align-items: flex-end;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item label {
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  white-space: nowrap;
}

.filter-input,
.filter-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  min-width: 160px;
  transition: all 0.3s;
}

.filter-input:focus,
.filter-select:focus {
  outline: none;
  border-color: #409EFF;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.filter-info {
  font-size: 13px;
  color: #718096;
  padding-top: 8px;
  border-top: 1px solid #e2e8f0;
}

.filter-info strong {
  color: #409EFF;
  font-size: 16px;
}

.registration-table {
  width: 100%;
  border-collapse: collapse;
}

.registration-table thead {
  background: #f7fafc;
}

.registration-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
  border-bottom: 2px solid #e2e8f0;
}

.registration-table td {
  padding: 12px;
  border-bottom: 1px solid #e2e8f0;
  font-size: 14px;
  color: #2d3748;
}

.registration-table tr:hover {
  background: #f7fafc;
}

.badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.badge-pending {
  background: #E6A23C;
  color: white;
}

.badge-approved {
  background: #67C23A;
  color: white;
}

.badge-rejected {
  background: #F56C6C;
  color: white;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
  text-align: center;
  transition: all 0.3s;
  background: #409EFF;
  color: white;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-small {
  padding: 6px 12px;
  font-size: 12px;
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

.btn-chat {
  position: relative;
}

.unread-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background: #f56c6c;
  color: white;
  border-radius: 10px;
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 700;
}

.total-unread-badge {
  display: inline-block;
  background: #f56c6c;
  color: white;
  border-radius: 10px;
  padding: 4px 10px;
  font-size: 14px;
  font-weight: 600;
  margin-left: 12px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

/* 聊天模态框样式 */
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
}

.chat-modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #718096;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.3s;
}

.close-btn:hover {
  background: #f7fafc;
  color: #2d3748;
}

.chat-modal-body {
  flex: 1;
  display: flex;
  flex-direction: column;
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
  padding: 40px 0;
}

.chat-message {
  margin-bottom: 16px;
  display: flex;
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
}

.message-admin .message-bubble {
  background: #409EFF;
  color: white;
}

.message-user .message-bubble {
  background: white;
  color: #2d3748;
  border: 1px solid #e2e8f0;
}

.message-text {
  margin-bottom: 4px;
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  opacity: 0.7;
}

.chat-input-area {
  padding: 16px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  gap: 12px;
  background: white;
}

.chat-input {
  flex: 1;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  resize: none;
  font-family: inherit;
  font-size: 14px;
}

.chat-input:focus {
  outline: none;
  border-color: #409EFF;
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
}

.btn-send:hover:not(:disabled) {
  background: #66b1ff;
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 分页样式 */
.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 0;
  border-top: 1px solid #e2e8f0;
  margin-top: 20px;
}

.pagination-info {
  color: #718096;
  font-size: 14px;
}

.pagination-controls {
  display: flex;
  gap: 8px;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #4a5568;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
  min-width: 36px;
}

.page-btn:hover:not(:disabled) {
  border-color: #409EFF;
  color: #409EFF;
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-btn.active {
  background: #409EFF;
  color: white;
  border-color: #409EFF;
}

.page-ellipsis {
  border: none;
  background: none;
  cursor: default;
}
</style>
