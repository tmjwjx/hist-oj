import axios from 'axios'

const BASE_URL = '/api/contest-question'

function javaResponse(request) {
  return request.then(response => {
    if (response.data && response.data.code === undefined) {
      response.data.code = response.data.status
    }
    return response
  })
}

export default {
  // 创建问题
  createQuestion(data) {
    return javaResponse(axios.post(`${BASE_URL}`, data))
  },

  // 获取问题列表
  getQuestions(contestId, params) {
    return javaResponse(axios.get(`${BASE_URL}/${contestId}/questions`, { params }))
  },

  // 获取问题详情
  getQuestionDetail(questionId) {
    return javaResponse(axios.get(`${BASE_URL}/question/${questionId}`))
  },

  // 发送回复
  sendReply(questionId, content) {
    return javaResponse(axios.post(`${BASE_URL}/question/${questionId}/reply`, { content }))
  },

  // 更新问题状态
  updateQuestionStatus(questionId, status) {
    return javaResponse(axios.put(`${BASE_URL}/question/${questionId}/status`, { status }))
  },

  // 删除问题
  deleteQuestion(questionId) {
    return javaResponse(axios.delete(`${BASE_URL}/question/${questionId}`))
  }
}
