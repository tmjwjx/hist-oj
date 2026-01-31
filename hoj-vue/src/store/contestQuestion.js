import api from '@/api/contestQuestion'

const state = {
  questions: [],
  currentQuestion: null,
  total: 0
}

const getters = {
  hasQuestions: (state) => state.questions.length > 0
}

const mutations = {
  SET_QUESTIONS(state, { list, total }) {
    state.questions = list || []
    state.total = total || 0
  },
  SET_CURRENT_QUESTION(state, question) {
    state.currentQuestion = question
  },
  ADD_QUESTION(state, question) {
    state.questions.unshift(question)
  },
  UPDATE_QUESTION(state, updatedQuestion) {
    const index = state.questions.findIndex(q => q.id === updatedQuestion.id)
    if (index !== -1) {
      state.questions.splice(index, 1, updatedQuestion)
    }
  },
  REMOVE_QUESTION(state, questionId) {
    state.questions = state.questions.filter(q => q.id !== questionId)
  },
  ADD_REPLY(state, { questionId, reply }) {
    const question = state.questions.find(q => q.id === questionId)
    if (question) {
      if (!question.replies) {
        question.replies = []
      }
      question.replies.push(reply)
    }
    if (state.currentQuestion && state.currentQuestion.id === questionId) {
      if (!state.currentQuestion.replies) {
        state.currentQuestion.replies = []
      }
      state.currentQuestion.replies.push(reply)
    }
  }
}

const actions = {
  // 创建问题
  async createQuestion({ commit }, data) {
    return api.createQuestion(data)
  },

  // 获取问题列表
  async getQuestions({ commit }, { contestId, params }) {
    return api.getQuestions(contestId, params)
  },

  // 获取问题详情
  async getQuestionDetail({ commit }, questionId) {
    return api.getQuestionDetail(questionId)
  },

  // 发送回复
  async sendReply({ commit }, { questionId, content }) {
    return api.sendReply(questionId, content)
  },

  // 更新问题状态
  async updateQuestionStatus({ commit }, { questionId, status }) {
    return api.updateQuestionStatus(questionId, status)
  },

  // 删除问题
  async deleteQuestion({ commit }, questionId) {
    return api.deleteQuestion(questionId)
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
