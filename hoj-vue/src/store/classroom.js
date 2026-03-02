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
    // Return consistent format: { code, data, message }
    if (res.data && res.data.code === 200) {
      return { code: 200, data: res.data.data }
    } else {
      return {
        code: res.data?.code || 500,
        message: res.data?.message || '获取班级列表失败',
        data: []
      }
    }
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
  async getTeacherClassrooms({ commit }) {
    const res = await api.getTeacherClassrooms()
    return res.data
  },
  async getStudentClassrooms({ commit }) {
    const res = await api.getStudentClassrooms()
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
  async getCheckinListForStudent({ commit }, classroomId) {
    // 学生端安全API - 不返回敏感信息
    const res = await api.getCheckinListForStudent(classroomId)
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

  // ==================== 二维码签到功能 ====================
  async getQrcodeToken({ commit }, checkinId) {
    const res = await api.getQrcodeToken(checkinId)
    return res.data
  },
  async refreshQrcode({ commit }, checkinId) {
    const res = await api.refreshQrcode(checkinId)
    return res.data
  },
  async submitQrcodeCheckin({ commit }, data) {
    const res = await api.submitQrcodeCheckin(data)
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
  async getHomeworkAnalysis({ commit }, homeworkId) {
    const res = await api.getHomeworkAnalysis(homeworkId)
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
  async updateFolder({ commit }, data) {
    const res = await api.updateFolder(data)
    return res.data
  },
  async uploadMaterial({ commit }, formData) {
    const res = await api.uploadMaterial(formData)
    return res.data
  },
  async getMaterials({ commit }, { classroomId, folderId }) {
    const res = await api.getMaterials(classroomId, folderId)
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

  // 编程题提交记录
  async saveProgrammingSubmission({ commit }, data) {
    const res = await api.saveProgrammingSubmission(data)
    return res.data
  },
  async getProgrammingSubmissions({ commit }, params) {
    const res = await api.getProgrammingSubmissions(params)
    return res.data
  },

  // ==================== 考试模式相关 ====================
  // 开始考试
  async startExam({ commit }, { homeworkId, deviceInfo, browserInfo }) {
    const res = await api.startExam(homeworkId, { deviceInfo, browserInfo })
    return res.data
  },
  // 获取考试状态
  async getExamStatus({ commit }, homeworkId) {
    const res = await api.getExamStatus(homeworkId)
    return res.data
  },
  // 记录违规
  async logViolation({ commit }, { homeworkId, violationType, description }) {
    const res = await api.logViolation({ homeworkId, violationType, description })
    return res.data
  },
  // 获取考试监控数据
  async getExamMonitoring({ commit }, homeworkId) {
    const res = await api.getExamMonitoring(homeworkId)
    return res.data
  },
  // 强制单个学生交卷
  async forceSubmit({ commit }, { homeworkId, uid, reason }) {
    const res = await api.forceSubmit({ homeworkId, uid, reason })
    return res.data
  },
  // 批量强制收卷
  async forceSubmitAll({ commit }, homeworkId) {
    const res = await api.forceSubmitAll(homeworkId)
    return res.data
  },

  // ==================== 班级教师管理 ====================
  // 获取班级教师列表
  async getClassroomTeachers({ commit }, classroomId) {
    const res = await api.getClassroomTeachers(classroomId)
    return res.data
  },
  // 添加班级教师
  async addClassroomTeacher({ commit }, data) {
    const res = await api.addClassroomTeacher(data)
    return res.data
  },
  // 移除班级教师
  async removeClassroomTeacher({ commit }, data) {
    const res = await api.removeClassroomTeacher(data)
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
