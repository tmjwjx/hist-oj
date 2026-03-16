import axios from 'axios'

// 创建专门用于报名系统API的axios实例
const registrationRequest = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器（添加token）
registrationRequest.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = token
    }
    // 报名系统API都是普通用户接口，设置 Url-Type 为 general
    config.headers['Url-Type'] = 'general'
    return config
  },
  error => {
    console.error('请求拦截器错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器（统一处理错误）
registrationRequest.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('Registration API Error:', error)
    return Promise.reject(error)
  }
)

/**
 * 获取比赛列表
 * @param {Object} params - 查询参数
 * @param {boolean} params.show_hidden - 是否显示隐藏的比赛（管理员）
 */
export function getCompetitions(params) {
  return registrationRequest({
    url: '/registration/competitions',
    method: 'get',
    params
  })
}

/**
 * 获取比赛详情
 * @param {number} id - 比赛ID
 */
export function getCompetition(id) {
  return registrationRequest({
    url: `/registration/competitions/${id}`,
    method: 'get'
  })
}

/**
 * 提交报名
 * @param {Object} data - 报名数据
 */
export function createRegistration(data) {
  return registrationRequest({
    url: '/registration/registrations',
    method: 'post',
    data
  })
}

/**
 * 获取我在某个比赛的报名信息
 * @param {number} competitionId - 比赛ID
 * @param {string} userUuid - 用户UUID（可选，默认使用当前登录用户）
 */
export function getMyRegistration(competitionId, userUuid) {
  return registrationRequest({
    url: `/registration/competitions/${competitionId}/my-registration`,
    method: 'get',
    params: { user_uuid: userUuid }
  })
}

/**
 * 更新报名信息（用户修改自己的报名）
 * @param {number} id - 报名记录ID
 * @param {Object} data - 更新数据
 */
export function updateRegistration(id, data) {
  return registrationRequest({
    url: `/registration/registrations/${id}`,
    method: 'put',
    data
  })
}

// ==================== 管理员接口 ====================

/**
 * 管理员获取所有比赛（包括隐藏的）
 * @param {Object} params - 查询参数
 */
export function adminGetCompetitions(params) {
  return registrationRequest({
    url: '/registration/admin/competitions',
    method: 'get',
    params
  })
}

/**
 * 管理员创建比赛
 * @param {Object} data - 比赛数据
 */
export function adminCreateCompetition(data) {
  return registrationRequest({
    url: '/registration/admin/competitions',
    method: 'post',
    data
  })
}

/**
 * 管理员更新比赛
 * @param {number} id - 比赛ID
 * @param {Object} data - 更新数据
 */
export function adminUpdateCompetition(id, data) {
  return registrationRequest({
    url: `/registration/admin/competitions/${id}`,
    method: 'put',
    data
  })
}

/**
 * 管理员删除比赛
 * @param {number} id - 比赛ID
 */
export function adminDeleteCompetition(id) {
  return registrationRequest({
    url: `/registration/admin/competitions/${id}`,
    method: 'delete'
  })
}

/**
 * 管理员更新比赛可见性
 * @param {number} id - 比赛ID
 * @param {Object} data - { visible: boolean }
 */
export function adminUpdateVisibility(id, data) {
  return registrationRequest({
    url: `/registration/admin/competitions/${id}/visibility`,
    method: 'put',
    data
  })
}

/**
 * 管理员获取比赛的所有报名
 * @param {number} competitionId - 比赛ID
 */
export function adminGetRegistrations(competitionId) {
  return registrationRequest({
    url: `/registration/admin/competitions/${competitionId}/registrations`,
    method: 'get'
  })
}

/**
 * 管理员更新报名状态（审核）
 * @param {number} id - 报名记录ID
 * @param {Object} data - { status: 'pending' | 'approved' | 'rejected', remark?: string }
 */
export function adminUpdateRegistrationStatus(id, data) {
  return registrationRequest({
    url: `/registration/admin/registrations/${id}/status`,
    method: 'put',
    data
  })
}

/**
 * 管理员上传比赛Logo
 * @param {FormData} formData - 包含 image 字段的表单数据
 */
export function adminUploadLogo(formData) {
  return registrationRequest({
    url: '/registration/admin/upload/logo',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * HOJ用户自动登录（兼容接口）
 * @param {Object} params - { token: string, username: string }
 */
export function hojAutoLogin(params) {
  return registrationRequest({
    url: '/registration/user/hoj-auto-login',
    method: 'get',
    params
  })
}
