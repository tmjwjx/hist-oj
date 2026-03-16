<template>
  <div class="registration-container">
    <!-- 初始化加载中 -->
    <div v-if="isInitializing" class="login-container">
      <div class="login-card" style="text-align: center;">
        <div style="font-size: 48px; margin-bottom: 20px;">⏳</div>
        <h2 style="color: #4facfe; margin-bottom: 12px;">正在加载...</h2>
        <p style="color: #718096;">正在从 BingOJ 主系统获取登录信息</p>
      </div>
    </div>

    <!-- 未登录提示 -->
    <div v-else-if="!isLoggedIn" class="login-container">
      <div class="login-card">
        <div class="login-logo">
          <span class="login-logo-icon">🎯</span>
          <h2 class="login-title">BingOJ</h2>
          <p class="login-subtitle">比赛报名系统</p>
        </div>

        <div style="text-align: center; padding: 20px;">
          <div style="font-size: 64px; margin-bottom: 20px;">🔐</div>
          <h3 style="color: #4facfe; margin-bottom: 16px; font-size: 24px;">需要先登录 BingOJ 主系统</h3>
          <p style="color: #718096; margin-bottom: 24px; line-height: 1.6; font-size: 16px;">
            报名系统已集成 BingOJ 账号系统，请先登录主系统
          </p>
          <a href="/" class="btn btn-primary" style="display: inline-block; text-decoration: none; text-align: center; font-size: 18px; padding: 18px 40px;">
            前往 BingOJ 主系统登录
          </a>
          <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
            <p style="color: #909399; font-size: 14px; margin-bottom: 8px;">👆 点击上方按钮登录 BingOJ</p>
            <p style="color: #909399; font-size: 14px;">登录后返回此页面即可自动进入报名系统</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 主界面 -->
    <div v-else class="container">
      <div class="header">
        <div>
          <h1>📝 BingOJ 比赛报名系统</h1>
          <p>欢迎，{{ currentUser.username }}</p>
        </div>
        <div>
          <router-link to="/" class="btn-home">🏠 返回首页</router-link>
        </div>
      </div>

      <!-- 比赛列表 -->
      <div class="card">
        <h2>📋 可报名的比赛</h2>
        <div v-if="competitions.length > 0" class="competition-grid">
          <div
            v-for="comp in competitions"
            :key="comp.id"
            class="competition-card"
            @click="goToCompetition(comp)">
            <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 12px;">
              <div v-if="comp.logoUrl" style="flex-shrink: 0;">
                <img :src="getLogoUrl(comp.logoUrl)" :alt="comp.name" style="width: 64px; height: 64px; object-fit: contain; border-radius: 8px;">
              </div>
              <div style="flex: 1; min-width: 0;">
                <h3 style="font-size: 18px; color: #2d3748; margin-bottom: 10px; position: relative; display: inline-block;">
                  {{ comp.name }}
                  <span v-if="getUnreadCount(comp.id) > 0"
                        style="position: absolute; top: -8px; right: -20px; background: #ef4444; color: white; border-radius: 10px; padding: 2px 8px; font-size: 12px; font-weight: 600; min-width: 20px; text-align: center;">
                    {{ getUnreadCount(comp.id) > 99 ? '99+' : getUnreadCount(comp.id) }}
                  </span>
                </h3>
              </div>
            </div>
            <p style="color: #718096; font-size: 14px;">📅 {{ formatTime(comp.startTime) }} - {{ formatTime(comp.endTime) }}</p>
            <div style="margin-top: 10px; display: flex; gap: 8px; flex-wrap: wrap;">
              <span class="badge" :class="getStatusClass(comp.id)">
                {{ getStatusText(comp.id) }}
              </span>
              <span class="badge" :style="{ background: getRegistrationStatus(comp).color, color: 'white' }">
                {{ getRegistrationStatus(comp).text }}
              </span>
            </div>
          </div>
        </div>
        <div v-else style="text-align: center; padding: 40px; color: #718096;">
          <p>暂无可报名的比赛</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getCompetitions, getMyRegistration } from '@/api/registration'
import { JUDGE_STATUS } from '@/common/constants'
import mMessage from '@/common/message'

export default {
  name: 'RegistrationList',
  data() {
    return {
      isInitializing: true,
      isLoggedIn: false,
      currentUser: null,
      competitions: [],
      userRegistrations: {},
      unreadCounts: {},
      pollTimer: null
    }
  },
  methods: {
    // 从 BingOJ 主系统获取用户信息（自动登录）
    async autoLoginFromBingOJ() {
      try {
        const token = localStorage.getItem('token')
        if (!token) {
          return false
        }

        const userInfoStr = localStorage.getItem('userInfo')
        if (!userInfoStr) {
          return false
        }

        const hojUserInfo = JSON.parse(userInfoStr)
        const uuid = hojUserInfo.uid || hojUserInfo.uuid || hojUserInfo.UUID || hojUserInfo.userId || hojUserInfo.user_id
        const username = hojUserInfo.username || hojUserInfo.Username || hojUserInfo.uname || hojUserInfo.nickname

        if (!uuid || !username) {
          return false
        }

        this.currentUser = {
          uuid: uuid,
          username: username
        }
        this.isLoggedIn = true

        localStorage.setItem('registrationUser', JSON.stringify(this.currentUser))

        await this.loadCompetitions()
        this.startPolling()

        return true
      } catch (error) {
        return false
      }
    },

    goToCompetition(comp) {
      this.$router.push({ name: 'CompetitionDetail', params: { id: comp.id } })
    },

    async loadCompetitions() {
      try {
        const res = await getCompetitions()
        if (res.code === 200) {
          this.competitions = res.data
          for (const comp of this.competitions) {
            await this.loadUserRegistration(comp.id)
          }
        }
      } catch (error) {
        console.error('加载失败', error)
      }
    },

    async loadUserRegistration(competitionId) {
      try {
        const res = await getMyRegistration(competitionId, this.currentUser.uuid)
        if (res.code === 200) {
          this.userRegistrations[competitionId] = res.data

          if (res.data && res.data.remark) {
            try {
              const messages = JSON.parse(res.data.remark)

              let lastViewTime = res.data.lastViewTime || '0'

              if (!lastViewTime || lastViewTime === '0') {
                lastViewTime = localStorage.getItem(`lastView_${competitionId}`) || '0'
              }

              if (!lastViewTime || lastViewTime === '0') {
                lastViewTime = new Date().toISOString()
                localStorage.setItem(`lastView_${competitionId}`, lastViewTime)
              }

              const unreadCount = messages.filter(msg => {
                return msg.sender === 'admin' &&
                       new Date(msg.timestamp) > new Date(lastViewTime)
              }).length
              this.unreadCounts[competitionId] = unreadCount
            } catch (e) {
              this.unreadCounts[competitionId] = 0
            }
          } else {
            this.unreadCounts[competitionId] = 0
          }
        } else {
          this.userRegistrations[competitionId] = null
          this.unreadCounts[competitionId] = 0
        }
      } catch (e) {
        this.userRegistrations[competitionId] = null
        this.unreadCounts[competitionId] = 0
      }
    },

    getStatusClass(competitionId) {
      const reg = this.userRegistrations[competitionId]
      if (!reg) return 'none'
      return reg.status
    },

    getStatusText(competitionId) {
      const reg = this.userRegistrations[competitionId]
      if (!reg) return '未报名'
      if (reg.status === 'pending') return '待审核'
      if (reg.status === 'approved') return '报名成功'
      if (reg.status === 'rejected') return '报名失败'
    },

    getUnreadCount(competitionId) {
      return this.unreadCounts[competitionId] || 0
    },

    getRegistrationStatus(comp) {
      if (!comp.startTime || !comp.endTime) {
        return { text: '未知', color: '#9ca3af' }
      }

      const now = new Date()
      const start = new Date(comp.startTime)
      const end = new Date(comp.endTime)

      if (now < start) {
        return { text: '未开始报名', color: '#f59e0b' }
      } else if (now > end) {
        return { text: '报名结束', color: '#ef4444' }
      } else {
        return { text: '报名中', color: '#10b981' }
      }
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

    getLogoUrl(url) {
      if (!url) return ''
      if (url.startsWith('http://') || url.startsWith('https://')) {
        return url
      }
      return url
    },

    startPolling() {
      this.pollTimer = setInterval(() => {
        if (this.isLoggedIn && this.currentUser) {
          this.loadCompetitions()
        }
      }, 5000)
    },

    stopPolling() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
        this.pollTimer = null
      }
    }
  },
  async mounted() {
    const autoLoginSuccess = await this.autoLoginFromBingOJ()
    this.isInitializing = false

    if (autoLoginSuccess) {
      return
    }

    const savedUser = localStorage.getItem('registrationUser')
    if (savedUser) {
      try {
        this.currentUser = JSON.parse(savedUser)
        this.isLoggedIn = true
        await this.loadCompetitions()
        this.startPolling()
      } catch (e) {
        localStorage.removeItem('registrationUser')
      }
    }
  },
  beforeDestroy() {
    this.stopPolling()
  }
}
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.registration-container {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #ffffff;
  min-height: 100vh;
  padding: 20px;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 80vh;
}

.login-card {
  background: white;
  border-radius: 24px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  width: 100%;
  max-width: 440px;
  animation: slideUp 0.5s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-logo {
  text-align: center;
  margin-bottom: 24px;
}

.login-logo-icon {
  font-size: 64px;
  margin-bottom: 12px;
  display: block;
}

.login-title {
  font-size: 28px;
  font-weight: 800;
  text-align: center;
  margin-bottom: 8px;
  
  
  
  
}

.login-subtitle {
  text-align: center;
  color: #718096;
  font-size: 14px;
  margin-bottom: 32px;
}

.btn {
  width: 100%;
  padding: 16px;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s;
  margin-top: 8px;
  background: #409EFF;
  color: white;
}

.btn-primary {
  background: #409EFF;
  color: white;
}

.btn-primary:hover {
  background: #66b1ff;
  transform: translateY(-2px);
}

.btn-primary:active {
  transform: translateY(0);
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
  color: #2d3748;
}

.btn-home {
  background: #4299e1;
  color: white;
  padding: 10px 20px;
  width: auto;
  margin-right: 10px;
  text-decoration: none;
  display: inline-block;
  border-radius: 12px;
  font-weight: 600;
  transition: all 0.3s;
  cursor: pointer;
}

.btn-home:hover {
  background: #3182ce;
  transform: translateY(-2px);
}

.card {
  background: white;
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  margin-bottom: 24px;
}

.card h2 {
  font-size: 20px;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 20px;
}

.competition-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.competition-card {
  background: #f7fafc;
  padding: 24px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.competition-card:hover {
  border-color: #4facfe;
  transform: translateY(-4px);
}

.badge {
  padding: 6px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  display: inline-block;
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

.badge.none {
  background: #909399;
  color: white;
}
</style>
