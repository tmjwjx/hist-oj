import axios from 'axios'

const BASE_URL = '/api'

function javaResponse(request) {
  return request.then(response => {
    if (response.data && response.data.code === undefined) {
      response.data.code = response.data.status
      response.data.message = response.data.msg
    }
    return response
  })
}

/**
 * 参加训练
 * @param {Number} trainingId
 * @returns
 */
export function joinTraining(trainingId) {
  return javaResponse(axios.post(`${BASE_URL}/training/${trainingId}/join`))
}

/**
 * 获取训练参与者列表
 * @param {Number} trainingId
 * @returns
 */
export function getTrainingParticipants(trainingId) {
  return javaResponse(axios.get(`${BASE_URL}/training/${trainingId}/participants`))
}

/**
 * 获取我的训练记录
 * @param {Number} trainingId
 * @returns
 */
export function getMyTrainingRecord(trainingId) {
  return javaResponse(axios.get(`${BASE_URL}/training/${trainingId}/my-record`, {
    // 兼容尚未更新后端的节点：未参加训练的 404 不应触发全局错误提示。
    silentError: true
  }))
}
