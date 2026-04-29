import axios from 'axios'

const BASE_URL = '/api/classroom'

export default {
  // ==================== 权限管理 ====================
  getCurrentUserRoles() {
    return axios.get(`${BASE_URL}/user/roles`)
  },
  updateUserRoles(uid, roles) {
    return axios.put(`${BASE_URL}/admin/user/${uid}/roles`, { roles })
  },
  grantUserRole(data) {
    return axios.post(`${BASE_URL}/admin/role/grant`, data)
  },
  revokeUserRole(data) {
    return axios.delete(`${BASE_URL}/admin/role/revoke`, { data })
  },
  getUserRoles(userId) {
    return axios.get(`${BASE_URL}/admin/roles/${userId}`)
  },
  getAllClassroomsForAdmin() {
    return axios.get(`${BASE_URL}/admin/classrooms`)
  },
  // 搜索用户（用于角色管理）- 允许所有管理员调用
  searchUsersForRoleManagement(keyword, currentPage = 1, limit = 20) {
    return axios.get(`${BASE_URL}/admin/users/search`, {
      params: { keyword, currentPage, limit }
    })
  },

  // ==================== 班级管理 ====================
  createClassroom(data) {
    return axios.post(`${BASE_URL}/create`, data)
  },
  deleteClassroom(classroomId) {
    return axios.delete(`${BASE_URL}/${classroomId}`)
  },
  getClassroomList() {
    return axios.get(`${BASE_URL}/list`)
  },
  getClassroomDetail(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}`)
  },
  joinClassroom(data) {
    return axios.post(`${BASE_URL}/join`, data)
  },
  getStudentClassrooms() {
    return axios.get(`${BASE_URL}/my-classrooms`)
  },
  getTeacherClassrooms() {
    return axios.get(`${BASE_URL}/teacher-classrooms`)
  },

  // ==================== 学生管理 ====================
  getClassroomStudents(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/students`)
  },
  addClassroomStudent(classroomId, data) {
    return axios.post(`${BASE_URL}/${classroomId}/students`, data)
  },
  searchStudentsToAdd(classroomId, keyword) {
    return axios.get(`${BASE_URL}/${classroomId}/students/search`, {
      params: { keyword }
    })
  },
  removeStudent(data) {
    return axios.delete(`${BASE_URL}/student`, { data })
  },
  updateStudentInfo(data) {
    console.log('API调用: updateStudentInfo, URL:', `${BASE_URL}/student`, 'data:', data)
    return axios.put(`${BASE_URL}/student`, data)
  },

  // ==================== 学生班级个人信息管理 ====================
  getClassroomStudentInfo(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/my-info`)
  },
  updateClassroomStudentInfo(classroomId, data) {
    return axios.put(`${BASE_URL}/${classroomId}/my-info`, data)
  },

  // ==================== 签到功能 ====================
  createCheckin(data) {
    return axios.post(`${BASE_URL}/checkin`, data)
  },
  studentCheckin(data) {
    return axios.post(`${BASE_URL}/checkin/submit`, data)
  },
  getCheckinList(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/checkins`)
  },
  getCheckinListForStudent(classroomId) {
    // 学生端安全API - 不返回敏感信息（checkinCode, qrcodeToken等）
    return axios.get(`${BASE_URL}/${classroomId}/checkins/student`)
  },
  getCheckinRecords(checkinId) {
    return axios.get(`${BASE_URL}/checkin/${checkinId}/records`)
  },
  createCheckinRecord(checkinId, data) {
    return axios.post(`${BASE_URL}/checkin/${checkinId}/record`, data)
  },
  updateCheckinRecord(data) {
    return axios.put(`${BASE_URL}/checkin/record`, data)
  },
  endCheckin(checkinId) {
    return axios.post(`${BASE_URL}/checkin/${checkinId}/end`)
  },

  // ==================== 二维码签到功能 ====================
  getQrcodeToken(checkinId) {
    return axios.get(`${BASE_URL}/checkin/${checkinId}/qrcode`)
  },
  refreshQrcode(checkinId) {
    return axios.post(`${BASE_URL}/checkin/${checkinId}/qrcode/refresh`)
  },
  submitQrcodeCheckin(data) {
    return axios.post(`${BASE_URL}/checkin/qrcode/submit`, data)
  },

  // ==================== 题库功能 ====================
  createQuestion(data) {
    return axios.post(`${BASE_URL}/question`, data)
  },
  getQuestionBank(params) {
    return axios.get(`${BASE_URL}/questions`, { params })
  },
  updateQuestion(questionId, data) {
    return axios.put(`${BASE_URL}/question/${questionId}`, data)
  },
  deleteQuestion(questionId) {
    return axios.delete(`${BASE_URL}/question/${questionId}`)
  },
  getQuestionDetail(questionId) {
    return axios.get(`${BASE_URL}/question/${questionId}`)
  },

  // ==================== 作业功能 ====================
  createHomework(data) {
    return axios.post(`${BASE_URL}/homework`, data)
  },
  updateHomework(homeworkId, data) {
    return axios.put(`${BASE_URL}/homework/${homeworkId}`, data)
  },
  getHomeworkList(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/homeworks`)
  },
  getHomeworkDetail(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}`)
  },
  saveHomeworkDraft(data) {
    return axios.post(`${BASE_URL}/homework/draft`, data)
  },
  submitHomework(data) {
    return axios.post(`${BASE_URL}/homework/submit`, data)
  },
  getHomeworkSubmissions(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/submissions`)
  },
  getStudentHomeworkStatus(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/status`)
  },
  getStudentHomeworkDetail(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/my-detail`)
  },
  deleteHomework(homeworkId) {
    return axios.delete(`${BASE_URL}/homework/${homeworkId}`)
  },
  gradeHomework(data) {
    return axios.post(`${BASE_URL}/homework/grade`, data)
  },
  recalculateScore(data) {
    return axios.post(`${BASE_URL}/homework/recalculate`, data)
  },
  getHomeworkAnalysis(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/analysis`)
  },
  getHomeworkRanking(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/ranking`)
  },

  // ==================== 编程题提交记录 ====================
  saveProgrammingSubmission(data) {
    return axios.post(`${BASE_URL}/programming/submission`, data)
  },
  getProgrammingSubmissions(params) {
    const query = { ...(params || {}) }
    if (!query.homeworkQuestionId && query.questionId) {
      query.homeworkQuestionId = query.questionId
      delete query.questionId
    }

    // 为编程题提交记录查询设置更长的超时时间（30秒）
    // 因为这个查询可能涉及判题系统，需要更长时间
    return axios.get(`${BASE_URL}/programming/submissions`, {
      params: query,
      timeout: 30000 // 30秒超时
    })
  },

  // ==================== 资料库功能 ====================
  createFolder(data) {
    return axios.post(`${BASE_URL}/folder`, data)
  },
  getFolders(classroomId, params) {
    return axios.get(`${BASE_URL}/${classroomId}/folders`, { params })
  },
  deleteFolder(folderId) {
    return axios.delete(`${BASE_URL}/folder/${folderId}`)
  },
  updateFolder(data) {
    return axios.put(`${BASE_URL}/folder`, data)
  },
  uploadMaterial(formData) {
    return axios.post(`${BASE_URL}/material/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  getMaterials(classroomId, folderId) {
    return axios.get(`${BASE_URL}/${classroomId}/folder/${folderId}/materials`)
  },
  deleteMaterial(materialId) {
    return axios.delete(`${BASE_URL}/material/${materialId}`)
  },
  copyMaterialToClassroom(data) {
    return axios.post(`${BASE_URL}/material/copy`, data)
  },

  // ==================== 随机选人 ====================
  randomPick(classroomId) {
    return axios.post(`${BASE_URL}/${classroomId}/random-pick`)
  },
  getPickHistory(classroomId, params) {
    return axios.get(`${BASE_URL}/${classroomId}/pick-history`, { params })
  },

  // ==================== 班级消息 ====================
  sendMessage(data) {
    return axios.post(`${BASE_URL}/message`, data)
  },
  getMessages(classroomId, params) {
    return axios.get(`${BASE_URL}/${classroomId}/messages`, { params })
  },
  recallMessage(messageId) {
    return axios.delete(`${BASE_URL}/message/${messageId}`)
  },

  // ==================== 考试模式相关 ====================
  // 开始考试
  startExam(homeworkId, data) {
    return axios.post(`${BASE_URL}/homework/${homeworkId}/start-exam`, data)
  },
  // 获取考试状态
  getExamStatus(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/exam-status`)
  },
  // 记录违规行为
  logViolation(data) {
    return axios.post(`${BASE_URL}/homework/violation`, data)
  },
  // 获取考试监控数据
  getExamMonitoring(homeworkId) {
    return axios.get(`${BASE_URL}/homework/${homeworkId}/exam-monitoring`)
  },
  // 强制单个学生交卷
  forceSubmit({ homeworkId, uid, reason }) {
    return axios.post(`${BASE_URL}/homework/force-submit`, { homeworkId, uid, reason })
  },
  // 批量强制收卷
  forceSubmitAll(homeworkId) {
    return axios.post(`${BASE_URL}/homework/${homeworkId}/force-submit-all`)
  },

  // ==================== 班级教师管理 ====================
  // 获取班级教师列表
  getClassroomTeachers(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/teachers`)
  },
  // 添加班级教师
  addClassroomTeacher(data) {
    return axios.post(`${BASE_URL}/admin/teacher/add`, data)
  },
  // 移除班级教师
  removeClassroomTeacher(data) {
    return axios.delete(`${BASE_URL}/admin/teacher/remove`, { data })
  },

  // ==================== 试卷库功能 ====================
  // 创建试卷
  createExamPaper(data) {
    return axios.post(`${BASE_URL}/exam-paper`, data)
  },
  // 获取试卷列表
  getExamPaperList(params) {
    return axios.get(`${BASE_URL}/exam-papers`, { params })
  },
  // 获取试卷详情
  getExamPaperDetail(paperId) {
    return axios.get(`${BASE_URL}/exam-paper/${paperId}`)
  },
  // 获取主界面公开练习试卷列表（无需认证）
  getPublicExamPaperList(params) {
    return axios.get(`${BASE_URL}/public/exam-papers`, { params })
  },
  // 获取主界面公开练习试卷详情（无需认证）
  getPublicExamPaperDetail(paperId) {
    return axios.get(`${BASE_URL}/public/exam-paper/${paperId}`)
  },
  // 获取主界面公开练习某题标准答案（按题加载）
  getPublicExamPaperQuestionAnswer(paperId, questionId) {
    return axios.get(`${BASE_URL}/public/exam-paper/${paperId}/question/${questionId}/answer`)
  },
  // 更新试卷
  updateExamPaper(paperId, data) {
    return axios.put(`${BASE_URL}/exam-paper/${paperId}`, data)
  },
  // 删除试卷
  deleteExamPaper(paperId) {
    return axios.delete(`${BASE_URL}/exam-paper/${paperId}`)
  },
  // 导入试卷到作业
  importExamPaper(data) {
    return axios.post(`${BASE_URL}/exam-paper/import`, data)
  },

  // ==================== 管理员专用API ====================
  // 管理员获取所有试卷列表
  adminGetExamPaperList(params) {
    return axios.get(`${BASE_URL}/admin/exam-papers`, { params })
  },
  // 管理员更新试卷
  adminUpdateExamPaper(paperId, data) {
    return axios.put(`${BASE_URL}/admin/exam-paper/${paperId}`, data)
  },
  // 管理员删除试卷
  adminDeleteExamPaper(paperId) {
    return axios.delete(`${BASE_URL}/admin/exam-paper/${paperId}`)
  },
  // 管理员获取试卷详情
  adminGetExamPaperDetail(paperId) {
    return axios.get(`${BASE_URL}/exam-paper/${paperId}`)
  },
  // 管理员获取题库（所有题目，包括私有）
  adminGetQuestionBank(params) {
    return axios.get(`${BASE_URL}/admin/question-bank`, { params })
  },

  // ==================== 权限申请管理 ====================
  // 申请班级角色（用户）
  createRoleApplication(data) {
    return axios.post(`${BASE_URL}/role/apply`, data)
  },
  // 取消角色申请（用户）
  cancelRoleApplication(applicationId) {
    return axios.delete(`${BASE_URL}/role/application/${applicationId}`)
  },
  // 获取当前用户的角色申请列表（用户）
  getMyRoleApplications(status = '0') {
    return axios.get(`${BASE_URL}/role/my_applications`, { params: { status } })
  },
  // 获取角色申请列表（管理员）
  getRoleApplications(params) {
    return axios.get(`${BASE_URL}/admin/role/applications`, { params })
  },
  // 审批角色申请（管理员）
  reviewRoleApplication(data) {
    return axios.post(`${BASE_URL}/admin/role/review`, data)
  }
}
