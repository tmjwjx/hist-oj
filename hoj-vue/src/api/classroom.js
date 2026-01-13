import axios from 'axios'

const BASE_URL = '/rating-api/api/classroom'

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

  // ==================== 学生管理 ====================
  getClassroomStudents(classroomId) {
    return axios.get(`${BASE_URL}/${classroomId}/students`)
  },
  removeStudent(data) {
    return axios.delete(`${BASE_URL}/student`, { data })
  },
  updateStudentInfo(data) {
    console.log('API调用: updateStudentInfo, URL:', `${BASE_URL}/student`, 'data:', data)
    return axios.put(`${BASE_URL}/student`, data)
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
  getCheckinRecords(checkinId) {
    return axios.get(`${BASE_URL}/checkin/${checkinId}/records`)
  },
  updateCheckinRecord(data) {
    return axios.put(`${BASE_URL}/checkin/record`, data)
  },
  endCheckin(checkinId) {
    return axios.post(`${BASE_URL}/checkin/${checkinId}/end`)
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

  // ==================== 编程题提交记录 ====================
  saveProgrammingSubmission(data) {
    return axios.post(`${BASE_URL}/programming/submission`, data)
  },
  getProgrammingSubmissions(params) {
    return axios.get(`${BASE_URL}/programming/submissions`, { params })
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
  uploadMaterial(formData) {
    return axios.post(`${BASE_URL}/material/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  getMaterials(folderId) {
    return axios.get(`${BASE_URL}/folder/${folderId}/materials`)
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
  clearMessages(classroomId) {
    return axios.delete(`${BASE_URL}/${classroomId}/messages`)
  }
}
