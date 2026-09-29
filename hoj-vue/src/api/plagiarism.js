import axios from 'axios'

const BASE_URL = '/api'

function javaResponse(request) {
  return request
    .then(response => {
      const body = response.data
      if (body && body.code === undefined && body.status !== undefined) {
        body.code = body.status
        body.message = body.msg
      }
      // Java 将结果数量放在 data 对象中，兼容旧页面使用的 data/totalCount 结构。
      if (body && body.data && Array.isArray(body.data.results)) {
        body.totalCount = body.data.totalCount
        body.data = body.data.results
      }
      if (body && body.code !== undefined && body.code !== 200) {
        throw new Error(body.message || body.msg || '请求失败')
      }
      return response
    })
    .catch(error => {
      if (error instanceof Error && !error.response && !error.data) {
        throw error
      }
      const body = error?.response?.data || error?.data || {}
      throw new Error(body.msg || body.message || error?.message || '查重请求失败')
    })
}

function silent(config = {}) {
  return { ...config, silentError: true }
}

export default {
  // 获取查重配置
  getPlagiarismConfig(cid) {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/contest/${cid}/config`, silent()))
  },

  // 保存查重配置
  savePlagiarismConfig(cid, configs) {
    return javaResponse(axios.post(`${BASE_URL}/plagiarism/contest/${cid}/config`, { configs }, silent()))
  },

  // 开始查重
  startPlagiarismCheck(cid) {
    return javaResponse(axios.post(`${BASE_URL}/plagiarism/contest/${cid}/check/start`, null, silent()))
  },

  // 获取查重进度
  getPlagiarismProgress(cid) {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/contest/${cid}/check/progress`, silent()))
  },

  // 获取最新查重任务
  getLatestPlagiarismCheck(cid) {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/contest/${cid}/check/latest`, silent()))
  },

  // 获取查重结果
  getPlagiarismResults(checkId, displayId = '') {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/check/${checkId}/results`, silent({
      params: { displayId }
    })))
  },

  // 导出查重结果
  exportPlagiarismResults(checkId) {
    return axios.get(`${BASE_URL}/plagiarism/check/${checkId}/results/export`, {
      responseType: 'blob',
      silentError: true
    })
  },

  // 获取提交详情
  getPlagiarismSubmission(submitId) {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/submission/${submitId}`, silent()))
  },

  // 比赛设备异常筛查（赛中赛后均可查询）
  getDeviceAnomalies(cid) {
    return javaResponse(axios.get(`${BASE_URL}/plagiarism/contest/${cid}/device-anomalies`, silent()))
  }
}
