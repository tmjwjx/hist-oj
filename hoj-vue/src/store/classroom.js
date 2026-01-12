import api from '@/api/classroom'

const state = {
  currentClassroom: null,
  classroomList: [],
  myClassrooms: [],
  students: [],
  checkins: [],
  questions: [],
  homeworks: [],
  folders: [],
  materials: [],
  messages: [],
  pickedStudent: null,
  userRoles: [] // 存储用户的班级角色
}

const getters = {
  isTeacher: (state) => {
    return state.userRoles && state.userRoles.includes('teacher')
  },
  isStudent: (state) => {
    return state.userRoles && state.userRoles.includes('student')
  }
}

const mutations = {
  SET_CURRENT_CLASSROOM(state, classroom) {
    state.currentClassroom = classroom
  },
  SET_CLASSROOM_LIST(state, list) {
    state.classroomList = list
  },
  SET_MY_CLASSROOMS(state, list) {
    state.myClassrooms = list
  },
  SET_STUDENTS(state, list) {
    state.students = list
  },
  SET_CHECKINS(state, list) {
    state.checkins = list
  },
  SET_QUESTIONS(state, list) {
    state.questions = list
  },
  SET_HOMEWORKS(state, list) {
    state.homeworks = list
  },
  SET_FOLDERS(state, list) {
    state.folders = list
  },
  SET_MATERIALS(state, list) {
    state.materials = list
  },
  SET_MESSAGES(state, list) {
    state.messages = list
  },
  ADD_MESSAGE(state, message) {
    state.messages.push(message)
  },
  SET_PICKED_STUDENT(state, student) {
    state.pickedStudent = student
  },
  SET_USER_ROLES(state, roles) {
    state.userRoles = roles || []
  }
}

const actions = {
  // 加载用户角色
  async loadUserRoles({ commit }) {
    try {
      const res = await api.getCurrentUserRoles()

      let roleNames = []

      // 处理 axios 响应格式: res.data 是 API 返回的数据
      if (res && res.data) {
        const apiData = res.data

        // 标准格式: { code: 200, data: [...] }
        if (apiData.code === 200 && apiData.data) {
          const roles = apiData.data
          roleNames = Array.isArray(roles)
            ? roles.map(item => item.role || item)
            : []
        }
        // 直接格式: { code: 200, data: [...] } 但 data 已经是角色数组
        else if (Array.isArray(apiData.data)) {
          roleNames = apiData.data.map(item => item.role || item)
        }
      }

      commit('SET_USER_ROLES', roleNames)
    } catch (error) {
      commit('SET_USER_ROLES', [])
    }
  },

  // 班级管理
  async createClassroom({ commit }, data) {
    const res = await api.createClassroom(data)
    return res.data // 返回 API 响应数据 { code, message, data }
  },
  async getAllClassroomsForAdmin({ commit }) {
    const res = await api.getAllClassroomsForAdmin()
    return res.data
  },
  async deleteClassroom({ commit }, classroomId) {
    const res = await api.deleteClassroom(classroomId)
    return res.data
  },
  async getClassroomList({ commit }) {
    const res = await api.getClassroomList()
    if (res.data.code === 200) {
      commit('SET_CLASSROOM_LIST', res.data.data)
    }
    return res.data
  },
  async getClassroomDetail({ commit }, classroomId) {
    const res = await api.getClassroomDetail(classroomId)
    if (res.data.code === 200) {
      commit('SET_CURRENT_CLASSROOM', res.data.data)
    }
    return res.data
  },
  async joinClassroom({ commit }, data) {
    const res = await api.joinClassroom(data)
    return res.data
  },
  async getMyClassrooms({ commit }) {
    const res = await api.getStudentClassrooms()
    if (res.data.code === 200) {
      commit('SET_MY_CLASSROOMS', res.data.data)
    }
    return res.data
  },

  // 学生管理
  async getClassroomStudents({ commit }, classroomId) {
    const res = await api.getClassroomStudents(classroomId)
    if (res.data.code === 200) {
      commit('SET_STUDENTS', res.data.data)
    }
    return res.data
  },
  async removeStudent({ commit }, data) {
    const res = await api.removeStudent(data)
    return res.data
  },
  async updateStudentInfo({ commit }, data) {
    const res = await api.updateStudentInfo(data)
    return res.data
  },

  // 签到功能
  async createCheckin({ commit }, data) {
    const res = await api.createCheckin(data)
    return res.data
  },
  async studentCheckin({ commit }, data) {
    const res = await api.studentCheckin(data)
    return res.data
  },
  async getCheckinList({ commit }, classroomId) {
    const res = await api.getCheckinList(classroomId)
    if (res.data.code === 200) {
      commit('SET_CHECKINS', res.data.data)
    }
    return res.data
  },
  async getCheckinRecords({ commit }, checkinId) {
    const res = await api.getCheckinRecords(checkinId)
    return res.data
  },
  async updateCheckinRecord({ commit }, data) {
    const res = await api.updateCheckinRecord(data)
    return res.data
  },
  async endCheckin({ commit }, checkinId) {
    const res = await api.endCheckin(checkinId)
    return res.data
  },

  // 题库功能
  async createQuestion({ commit }, data) {
    const res = await api.createQuestion(data)
    return res.data
  },
  async getQuestionBank({ commit }, params) {
    const res = await api.getQuestionBank(params)
    if (res.data.code === 200) {
      commit('SET_QUESTIONS', res.data.data.questions || res.data.data)
    }
    return res.data
  },
  async updateQuestion({ commit }, { questionId, data }) {
    const res = await api.updateQuestion(questionId, data)
    return res.data
  },
  async deleteQuestion({ commit }, questionId) {
    const res = await api.deleteQuestion(questionId)
    return res.data
  },

  // 作业功能
  async createHomework({ commit }, data) {
    const res = await api.createHomework(data)
    return res.data
  },
  async updateHomework({ commit }, data) {
    const res = await api.updateHomework(data.id, data)
    return res.data
  },
  async getHomeworkList({ commit }, classroomId) {
    const res = await api.getHomeworkList(classroomId)
    if (res.data.code === 200) {
      commit('SET_HOMEWORKS', res.data.data)
    }
    return res.data
  },
  async getHomeworkDetail({ commit }, homeworkId) {
    const res = await api.getHomeworkDetail(homeworkId)
    return res.data
  },
  async getStudentHomeworkDetail({ commit }, homeworkId) {
    const res = await api.getStudentHomeworkDetail(homeworkId)
    return res.data
  },
  async saveHomeworkDraft({ commit }, data) {
    const res = await api.saveHomeworkDraft(data)
    return res.data
  },
  async submitHomework({ commit }, data) {
    const res = await api.submitHomework(data)
    return res.data
  },
  async getHomeworkSubmissions({ commit }, homeworkId) {
    const res = await api.getHomeworkSubmissions(homeworkId)
    return res.data
  },
  async deleteHomework({ commit }, homeworkId) {
    const res = await api.deleteHomework(homeworkId)
    return res.data
  },
  async gradeHomework({ commit }, data) {
    const res = await api.gradeHomework(data)
    return res.data
  },
  async recalculateScore({ commit }, data) {
    const res = await api.recalculateScore(data)
    return res.data
  },

  // 资料库功能
  async createFolder({ commit }, data) {
    const res = await api.createFolder(data)
    return res.data
  },
  async getFolders({ commit }, { classroomId, params }) {
    const res = await api.getFolders(classroomId, params)
    if (res.data.code === 200) {
      commit('SET_FOLDERS', res.data.data)
    }
    return res.data
  },
  async deleteFolder({ commit }, folderId) {
    const res = await api.deleteFolder(folderId)
    return res.data
  },
  async uploadMaterial({ commit }, formData) {
    const res = await api.uploadMaterial(formData)
    return res.data
  },
  async getMaterials({ commit }, folderId) {
    const res = await api.getMaterials(folderId)
    if (res.data.code === 200) {
      commit('SET_MATERIALS', res.data.data)
    }
    return res.data
  },
  async deleteMaterial({ commit }, materialId) {
    const res = await api.deleteMaterial(materialId)
    return res.data
  },

  // 随机选人
  async randomPick({ commit }, classroomId) {
    const res = await api.randomPick(classroomId)
    if (res.data.code === 200) {
      commit('SET_PICKED_STUDENT', res.data.data)
    }
    return res.data
  },
  async getPickHistory({ commit }, { classroomId, params }) {
    const res = await api.getPickHistory(classroomId, params)
    return res.data
  },

  // 班级消息
  async sendMessage({ commit }, data) {
    const res = await api.sendMessage(data)
    return res.data
  },
  async getMessages({ commit }, { classroomId, params }) {
    const res = await api.getMessages(classroomId, params)
    if (res.data.code === 200) {
      commit('SET_MESSAGES', res.data.data)
    }
    return res.data
  },
  async recallMessage({ commit }, messageId) {
    const res = await api.recallMessage(messageId)
    return res.data
  },
  async clearMessages({ commit }, classroomId) {
    const res = await api.clearMessages(classroomId)
    return res.data
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
