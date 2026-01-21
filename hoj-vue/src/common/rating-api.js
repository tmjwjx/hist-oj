import axios from 'axios'

// 根据环境自动选择 API 地址
// 开发环境：使用 /rating-api/api（通过 vue.config.js 代理到 hist-oj 的 /api）
// 生产环境：使用 /api（通过 Nginx 代理到 hist-oj）
const getRatingApiBaseURL = () => {
  // 生产环境（已构建的静态文件）
  if (process.env.NODE_ENV === 'production') {
    return '/api'
  }
  // 开发环境（使用代理，需要加上 /api 前缀）
  return '/rating-api/api'
}

const ratingApi = axios.create({
  baseURL: getRatingApiBaseURL(),
  timeout: 10000
})

// 响应拦截器
ratingApi.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code === 200) {
      return res.data
    } else {
      return Promise.reject(new Error(res.message || 'Error'))
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
  }
}
