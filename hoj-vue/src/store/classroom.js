import api from '@/api/classroom'

const state = {
  currentClassroom: null,
  classroomList: [],
  myClassrooms: [],
  teacherClassrooms: [], // 缓存教师班级列表
  students: [],
  checkins: [],
  questions: [],
  homeworks: [],
  folders: [],
  materials: [],
  selectedQuestions: [], // 临时保存从题库选中的题目
  questionBankSyncActive: false, // 标记是否从题库浏览页返回并需要同步客观题
  messages: [],
  pickedStudent: null,
  userRoles: [], // 存储用户的班级角色
  // 请求状态追踪，避免重复请求
  loadingStates: {
    teacherClassrooms: false,
    studentClassrooms: false,
    userRoles: false,
    roleApplications: false
  },
  // 缓存时间戳
  cacheTimestamps: {
    teacherClassrooms: 0,
    studentClassrooms: 0,
    userRoles: 0
  },
  // 编程题提交记录缓存（减少频繁轮询对服务器的压力）
  programmingSubmissionsCache: {},
  homeworkDetailCache: {}, // 作业详情缓存
  // 缓存TTL配置（毫秒）
  cacheTTL: {
    programmingSubmissions: 15000, // 编程题提交记录缓存15秒（增加以减少API调用）
    homeworkDetail: 10000 // 作业详情缓存10秒
  },
  // 权限申请相关状态
  roleApplications: [],
  error: null
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
  },
  SET_TEACHER_CLASSROOMS(state, list) {
    state.teacherClassrooms = list || []
    state.cacheTimestamps.teacherClassrooms = Date.now()
    state.loadingStates.teacherClassrooms = false
  },
  SET_LOADING_STATE(state, { key, value }) {
    state.loadingStates[key] = value
  },
  CLEAR_CACHE(state, key) {
    if (key) {
      state.cacheTimestamps[key] = 0
    } else {
      // 清除所有缓存
      state.cacheTimestamps.teacherClassrooms = 0
      state.cacheTimestamps.studentClassrooms = 0
      state.cacheTimestamps.userRoles = 0
    }
  },
  SET_PROGRAMMING_SUBMISSIONS_CACHE(state, { cacheKey, data, timestamp }) {
    state.programmingSubmissionsCache[cacheKey] = {
      data,
      timestamp
    }
  },
  SET_HOMEWORK_DETAIL_CACHE(state, { homeworkId, data, timestamp }) {
    state.homeworkDetailCache[homeworkId] = {
      data,
      timestamp
    }
  },
  SET_ROLE_APPLICATIONS(state, applications) {
    state.roleApplications = applications || []
  },
  SET_SELECTED_QUESTIONS(state, questions) {
    state.selectedQuestions = questions || []
  },
  SET_QUESTION_BANK_SYNC_ACTIVE(state, active) {
    state.questionBankSyncActive = !!active
  }
}

const actions = {
  // 加载用户角色（带缓存和请求去重）
  async loadUserRoles({ commit, state }, forceRefresh = false) {
    // 检查是否正在加载
    if (state.loadingStates.userRoles) {
      return // 如果正在加载，直接返回等待结果
    }

    // 检查缓存（5分钟有效期）
    const cacheAge = Date.now() - state.cacheTimestamps.userRoles
    if (!forceRefresh && state.userRoles.length > 0 && cacheAge < 5 * 60 * 1000) {
      return // 使用缓存
    }

    commit('SET_LOADING_STATE', { key: 'userRoles', value: true })

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
      state.cacheTimestamps.userRoles = Date.now()
    } catch (error) {
      console.error('[classroom/loadUserRoles] 加载用户角色失败', error)
      commit('SET_USER_ROLES', [])
    } finally {
      commit('SET_LOADING_STATE', { key: 'userRoles', value: false })
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
  async getTeacherClassrooms({ commit, state }, forceRefresh = false) {
    // 检查是否正在加载
    if (state.loadingStates.teacherClassrooms) {
      // 如果正在加载，返回缓存数据（如果有）
      return { code: 200, data: state.teacherClassrooms }
    }

    // 检查缓存（2分钟有效期）
    const cacheAge = Date.now() - state.cacheTimestamps.teacherClassrooms
    if (!forceRefresh && state.teacherClassrooms.length > 0 && cacheAge < 2 * 60 * 1000) {
      return { code: 200, data: state.teacherClassrooms }
    }

    commit('SET_LOADING_STATE', { key: 'teacherClassrooms', value: true })

    try {
      const res = await api.getTeacherClassrooms()
      if (res && res.data && res.data.code === 200) {
        commit('SET_TEACHER_CLASSROOMS', res.data.data)
        return res.data
      }
      return res?.data || { code: 500, data: [] }
    } catch (error) {
      console.error('获取教师班级列表失败', error)
      // 发生错误时，如果有缓存就返回缓存
      if (state.teacherClassrooms.length > 0) {
        return { code: 200, data: state.teacherClassrooms }
      }
      return { code: 500, data: [] }
    } finally {
      commit('SET_LOADING_STATE', { key: 'teacherClassrooms', value: false })
    }
  },
  async getStudentClassrooms({ commit, state }, forceRefresh = false) {
    // 检查是否正在加载
    if (state.loadingStates.studentClassrooms) {
      return { code: 200, data: state.myClassrooms }
    }

    // 检查缓存（2分钟有效期）
    const cacheAge = Date.now() - state.cacheTimestamps.studentClassrooms
    if (!forceRefresh && state.myClassrooms.length > 0 && cacheAge < 2 * 60 * 1000) {
      return { code: 200, data: state.myClassrooms }
    }

    commit('SET_LOADING_STATE', { key: 'studentClassrooms', value: true })

    try {
      const res = await api.getStudentClassrooms()
      if (res && res.data && res.data.code === 200) {
        commit('SET_MY_CLASSROOMS', res.data.data)
        state.cacheTimestamps.studentClassrooms = Date.now()
        return res.data
      }
      return res?.data || { code: 500, data: [] }
    } catch (error) {
      console.error('获取学生班级列表失败', error)
      // 发生错误时，如果有缓存就返回缓存
      if (state.myClassrooms.length > 0) {
        return { code: 200, data: state.myClassrooms }
      }
      return { code: 500, data: [] }
    } finally {
      commit('SET_LOADING_STATE', { key: 'studentClassrooms', value: false })
    }
  },

  // 学生管理
  async getClassroomStudents({ commit }, classroomId) {
    const res = await api.getClassroomStudents(classroomId)
    if (res.data.code === 200) {
      commit('SET_STUDENTS', res.data.data)
    }
    return res.data
  },
  async addClassroomStudent({ commit }, { classroomId, data }) {
    const res = await api.addClassroomStudent(classroomId, data)
    return res.data
  },
  async searchStudentsToAdd({ commit }, { classroomId, keyword }) {
    const res = await api.searchStudentsToAdd(classroomId, keyword)
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
  async getHomeworkDetail({ commit, state }, homeworkId, forceRefresh = false) {
    // 如果强制刷新，跳过缓存
    if (!forceRefresh) {
      // 检查缓存
      const cached = state.homeworkDetailCache[homeworkId]
      const now = Date.now()

      if (cached && (now - cached.timestamp < state.cacheTTL.homeworkDetail)) {
        return cached.data
      }
    }

    // 调用API
    const res = await api.getHomeworkDetail(homeworkId)

    // 更新缓存
    const now = Date.now()
    commit('SET_HOMEWORK_DETAIL_CACHE', {
      homeworkId,
      data: res.data,
      timestamp: now
    })

    return res.data
  },
  async getStudentHomeworkDetail({ commit, state }, homeworkId) {
    // 检查缓存
    const cached = state.homeworkDetailCache[`student-${homeworkId}`]
    const now = Date.now()

    if (cached && (now - cached.timestamp < state.cacheTTL.homeworkDetail)) {
      return cached.data
    }

    // 调用API
    const res = await api.getStudentHomeworkDetail(homeworkId)

    // 更新缓存
    commit('SET_HOMEWORK_DETAIL_CACHE', {
      homeworkId: `student-${homeworkId}`,
      data: res.data,
      timestamp: now
    })

    return res.data
  },
  async saveHomeworkDraft({ commit }, data) {
    const res = await api.saveHomeworkDraft(data)
    return res.data
  },
  async submitHomework({ commit, state }, data) {
    const res = await api.submitHomework(data)

    // 提交成功后，清除相关缓存，确保能立即获取最新数据
    if (res.data && res.data.code === 200) {
      const homeworkId = data.homeworkId
      // 清除作业详情缓存
      delete state.homeworkDetailCache[homeworkId]
      delete state.homeworkDetailCache[`student-${homeworkId}`]
    }

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
  async saveProgrammingSubmission({ commit, state }, data) {
    const res = await api.saveProgrammingSubmission(data)

    // 清除相关缓存，以便下次获取最新数据
    if (res.data && res.data.code === 200) {
      const homeworkPrefix = `${data.homeworkId}-`
      Object.keys(state.programmingSubmissionsCache).forEach((cacheKey) => {
        if (cacheKey.startsWith(homeworkPrefix)) {
          delete state.programmingSubmissionsCache[cacheKey]
        }
      })
    }

    return res.data
  },
  async getProgrammingSubmissions({ commit, state }, params) {
    const requestParams = { ...(params || {}) }
    const forceRefresh = !!requestParams.forceRefresh
    delete requestParams.forceRefresh

    const homeworkQuestionId = requestParams.homeworkQuestionId || requestParams.questionId
    if (!homeworkQuestionId) {
      return { code: 400, message: 'homeworkQuestionId不能为空', data: [] }
    }
    requestParams.homeworkQuestionId = homeworkQuestionId
    delete requestParams.questionId

    // 生成缓存键
    const cacheKey = `${requestParams.homeworkId}-${requestParams.homeworkQuestionId}`
    const cached = state.programmingSubmissionsCache[cacheKey]
    const now = Date.now()

    // 如果有缓存且未过期，直接返回缓存数据
    if (!forceRefresh && cached && (now - cached.timestamp < state.cacheTTL.programmingSubmissions)) {
      return cached.data
    }

    // 没有缓存或缓存已过期，调用API
    const res = await api.getProgrammingSubmissions(requestParams)

    // 更新缓存
    commit('SET_PROGRAMMING_SUBMISSIONS_CACHE', {
      cacheKey,
      data: res.data,
      timestamp: now
    })

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
  },

  // ==================== 试卷库管理 ====================
  // 获取试卷列表
  async getExamPaperList({ commit }, params) {
    const res = await api.getExamPaperList(params)
    return res.data
  },
  // 导入试卷到作业
  async importExamPaper({ commit }, data) {
    const res = await api.importExamPaper(data)
    return res.data
  },

  // ==================== 权限申请管理 ====================
  // 申请班级角色
  async createRoleApplication({ commit }, data) {
    try {
      const res = await api.createRoleApplication(data)
      commit('SET_ERROR', null)
      return res.data
    } catch (error) {
      commit('SET_ERROR', error.message || '申请权限失败')
      throw error
    }
  },

  // 获取当前用户的角色申请列表
  async getMyRoleApplications({ commit }, status = '0') {
    try {
      const res = await api.getMyRoleApplications(status)
      // 后端返回格式：{ code: 200, data: { applications: [...], count: 1 } }
      const applications = res.data?.data?.applications || []
      commit('SET_ROLE_APPLICATIONS', applications)
      commit('SET_ERROR', null)
      return res.data
    } catch (error) {
      // API调用失败时也要设置为空数组，确保用户能看到申请按钮
      commit('SET_ROLE_APPLICATIONS', [])
      commit('SET_ERROR', error.message || '获取申请列表失败')
      throw error
    }
  },

  // 取消角色申请（用户）
  async cancelRoleApplication({ commit }, applicationId) {
    try {
      const res = await api.cancelRoleApplication(applicationId)
      commit('SET_ERROR', null)
      return res.data
    } catch (error) {
      commit('SET_ERROR', error.message || '取消申请失败')
      throw error
    }
  },

  // 获取角色申请列表（管理员）
  async getRoleApplications({ commit }, params) {
    try {
      const res = await api.getRoleApplications(params)
      commit('SET_ERROR', null)
      return res.data
    } catch (error) {
      commit('SET_ERROR', error.message || '获取申请列表失败')
      throw error
    }
  },

  // 审批角色申请（管理员）
  async reviewRoleApplication({ commit }, data) {
    try {
      const res = await api.reviewRoleApplication(data)
      commit('SET_ERROR', null)
      return res.data
    } catch (error) {
      commit('SET_ERROR', error.message || '审批申请失败')
      throw error
    }
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
