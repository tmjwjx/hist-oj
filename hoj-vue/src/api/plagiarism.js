import axios from 'axios'

const BASE_URL = '/plagiarism-api/api'

export default {
  // 获取查重配置
  getPlagiarismConfig(cid) {
    return axios.get(`${BASE_URL}/plagiarism/contest/${cid}/config`)
  },

  // 保存查重配置
  savePlagiarismConfig(cid, configs) {
    return axios.post(`${BASE_URL}/plagiarism/contest/${cid}/config`, { configs })
  },

  // 开始查重
  startPlagiarismCheck(cid) {
    return axios.post(`${BASE_URL}/plagiarism/contest/${cid}/check/start`)
  },

  // 获取查重进度
  getPlagiarismProgress(cid) {
    return axios.get(`${BASE_URL}/plagiarism/contest/${cid}/check/progress`)
  },

  // 获取最新查重任务
  getLatestPlagiarismCheck(cid) {
    return axios.get(`${BASE_URL}/plagiarism/contest/${cid}/check/latest`)
  },

  // 获取查重结果
  getPlagiarismResults(checkId, displayId = '') {
    return axios.get(`${BASE_URL}/plagiarism/check/${checkId}/results`, {
      params: { displayId }
    })
  },

  // 导出查重结果
  exportPlagiarismResults(checkId) {
    return axios.get(`${BASE_URL}/plagiarism/check/${checkId}/results/export`, {
      responseType: 'blob'
    })
  },

  // 获取提交详情
  getPlagiarismSubmission(submitId) {
    return axios.get(`${BASE_URL}/plagiarism/submission/${submitId}`)
  }
}
