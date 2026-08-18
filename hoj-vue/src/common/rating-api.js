import axios from 'axios'

const ratingApi = axios.create({
  // Rating 已迁入 Java 主后端，不再经过 Go 扩展层代理。
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Cache-Control': 'no-cache',
    'Pragma': 'no-cache'
  }
})

// 请求拦截器 - 添加认证token
ratingApi.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = token
    }
    // 添加操作人UID（用于AdminAuthMiddleware）
    const userInfo = localStorage.getItem('userInfo')
    if (userInfo) {
      try {
        const user = JSON.parse(userInfo)
        if (user && user.uuid) {
          config.headers['X-Operator-UID'] = user.uuid
        }
      } catch (e) {
        console.error('解析userInfo失败:', e)
      }
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
ratingApi.interceptors.response.use(
  response => {
    const res = response.data
    const status = res && res.code !== undefined ? res.code : res && res.status
    if (status === 200) {
      return res.data
    } else {
      return Promise.reject(new Error((res && (res.message || res.msg)) || 'Error'))
    }
  },
  error => {
    return Promise.reject(error)
  }
)

export default {
  // 获取用户 Rating
  getUserRating(uid) {
    return ratingApi.get(`/rating/user/${uid}`)
  },

  // 获取 Rating 历史
  getRatingHistory(uid, page = 1, limit = 20) {
    return ratingApi.get(`/rating/history/${uid}`, {
      params: { page, limit }
    })
  },

  // 获取 Rating 颜色
  getRatingColor(rating) {
    return ratingApi.get(`/rating/color/${rating}`)
  },

  // 获取比赛参赛者 Rating
  getContestParticipantsRating(contestId) {
    return ratingApi.get(`/rating/contest/${contestId}`)
  },

  // 批量获取用户 Rating
  getBatchUserRating(uids) {
    return ratingApi.post('/rating/batch', { uids })
  },

  // 获取比赛信息（包括是否为 Rating 比赛）
  getContestInfo(contestId) {
    return ratingApi.get(`/rating/contest/info/${contestId}`)
  },

  // 批量获取比赛信息（包括是否为 Rating 比赛）
  getBatchContestInfo(contestIds) {
    return ratingApi.post('/rating/contest/batch', { contestIds })
  },

  // 设置比赛 Rating 类型（管理员）
  setContestRatingType(contestId, isRating) {
    return ratingApi.post('/rating/contest/set-rating-type', { contestId, isRating })
  },

  // 手动触发 Rating 计算（管理员）
  calculateRating(contestId) {
    return ratingApi.post(`/rating/calculate/${contestId}`)
  },

  // 获取 Rating 排名列表
  getRatingRank(page = 1, limit = 30, keyword = '') {
    return ratingApi.get('/rating/rank', {
      params: { page, limit, keyword }
    })
  },

  // ============ Skip用户管理（新增）============

  // 批量Skip用户
  batchSkipUsers(contestId, usernames, reason, autoRecalc = false) {
    return ratingApi.post('/rating/admin/contest/skip-users', {
      contestId,
      usernames,
      reason,
      autoRecalc
    })
  },

  // 获取比赛的Skip用户列表
  getContestSkipUsers(contestId) {
    // 添加多个缓存破坏参数
    return ratingApi.get(`/rating/admin/contest/${contestId}/skip-users`, {
      params: {
        _t: Date.now(),
        _rand: Math.random().toString(36).substring(7)
      },
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache'
      }
    })
  },

  // 取消Skip
  cancelSkip(contestId, uids) {
    return ratingApi.delete('/rating/admin/contest/skip-users', {
      data: { contestId, uids }
    })
  },

  // 触发重算
  recalculateFromContest(contestId) {
    return ratingApi.post(`/rating/admin/contest/${contestId}/recalculate`)
  },

  // 同步比赛的 Skip 标记到 rating_history 表
  syncContestSkipFlag(contestId) {
    return ratingApi.post(`/rating/admin/contest/${contestId}/sync-skip`)
  },

  // 获取重算进度
  getRecalculateProgress(taskId) {
    return ratingApi.get(`/rating/admin/recalculate-progress/${taskId}`)
  },

  // 获取操作日志
  getOperationLogs(page = 1, limit = 20, type = '', timeRange = '7d') {
    return ratingApi.get('/rating/admin/operation-logs', {
      params: { page, limit, type, timeRange }
    })
  }
}
