import axios from 'axios'

const BASE_URL = '/rating-api/api/contest-question'

export default {
  // 创建问题
  createQuestion(data) {
    return axios.post(`${BASE_URL}`, data)
  },

  // 获取问题列表
  getQuestions(contestId, params) {
    return axios.get(`${BASE_URL}/${contestId}/questions`, { params })
  },

  // 获取问题详情
  getQuestionDetail(questionId) {
    return axios.get(`${BASE_URL}/question/${questionId}`)
  },

  // 发送回复
  sendReply(questionId, content) {
    return axios.post(`${BASE_URL}/question/${questionId}/reply`, { content })
  },

  // 更新问题状态
  updateQuestionStatus(questionId, status) {
    return axios.put(`${BASE_URL}/question/${questionId}/status`, { status })
  },

  // 删除问题
  deleteQuestion(questionId) {
    return axios.delete(`${BASE_URL}/question/${questionId}`)
  }
}
