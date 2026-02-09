import axios from 'axios'
import { Message } from 'element-ui'

const api = axios.create({
  baseURL: process.env.VUE_APP_BASE_API || '/',
})

// 请求拦截器，添加 token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = token
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器，统一处理错误
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response) {
      const { status, data } = error.response
      Message.error(data.message || '请求失败')
    } else {
      Message.error('网络错误')
    }
    return Promise.reject(error)
  }
)

// 题目集 API
export const problemSetApi = {
  // 获取题目集列表
  getProblemSets() {
    return api.get('/api/problem-set')
  },

  // 获取题目集详情
  getProblemSet(id) {
    return api.get(`/api/problem-set/${id}`)
  },

  // 创建题目集
  createProblemSet(data) {
    return api.post('/api/problem-set', data)
  },

  // 更新题目集
  updateProblemSet(id, data) {
    return api.put(`/api/problem-set/${id}`, data)
  },

  // 删除题目集
  deleteProblemSet(id) {
    return api.delete(`/api/problem-set/${id}`)
  },

  // 添加题目
  createProblem(setId, data) {
    return api.post(`/api/problem-set/${setId}/problem`, data)
  },

  // 更新题目
  updateProblem(setId, problemId, data) {
    return api.put(`/api/problem-set/${setId}/problem/${problemId}`, data)
  },

  // 删除题目
  deleteProblem(setId, problemId) {
    return api.delete(`/api/problem-set/${setId}/problem/${problemId}`)
  },

  // 添加样例
  createExample(setId, problemId, data) {
    return api.post(`/api/problem-set/${setId}/problem/${problemId}/example`, data)
  },

  // 更新样例
  updateExample(setId, problemId, exampleId, data) {
    return api.put(`/api/problem-set/${setId}/problem/${problemId}/example/${exampleId}`, data)
  },

  // 删除样例
  deleteExample(setId, problemId, exampleId) {
    return api.delete(`/api/problem-set/${setId}/problem/${problemId}/example/${exampleId}`)
  },

  // 生成并下载 PDF（从数据库读取）
  generatePDF(id) {
    return api.get(`/api/problem-set/${id}/pdf`, {
      responseType: 'blob'
    })
  },

  // 从前端数据生成 PDF（所见即所得，与预览保持一致）
  generatePDFFromData(data) {
    return api.post('/api/problem-set/pdf/from-data', data, {
      responseType: 'blob'
    })
  },

  // 批量重新排序题目
  reorderProblems(setId, problemIds) {
    return api.post(`/api/problem-set/${setId}/reorder`, { problem_ids: problemIds })
  },

  // 上移题目
  moveProblemUp(setId, problemId) {
    return api.post(`/api/problem-set/${setId}/problem/${problemId}/move-up`)
  },

  // 下移题目
  moveProblemDown(setId, problemId) {
    return api.post(`/api/problem-set/${setId}/problem/${problemId}/move-down`)
  },

  // 图片管理
  // 上传图片
  uploadImage(setId, file) {
    const formData = new FormData()
    formData.append('file', file)
    return api.post(`/api/problem-set/${setId}/images`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 列出图片
  listImages(setId) {
    return api.get(`/api/problem-set/${setId}/images`)
  },

  // 删除图片
  deleteImage(setId, imageId) {
    return api.delete(`/api/problem-set/${setId}/images/${imageId}`)
  }
}

export default api
