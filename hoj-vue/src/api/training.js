import axios from 'axios'

const BASE_URL = '/rating-api/api'

/**
 * 参加训练
 * @param {Number} trainingId
 * @returns
 */
export function joinTraining(trainingId) {
  return axios.post(`${BASE_URL}/training/${trainingId}/join`)
}

/**
 * 获取训练参与者列表
 * @param {Number} trainingId
 * @returns
 */
export function getTrainingParticipants(trainingId) {
  return axios.get(`${BASE_URL}/training/${trainingId}/participants`)
}

/**
 * 获取我的训练记录
 * @param {Number} trainingId
 * @returns
 */
export function getMyTrainingRecord(trainingId) {
  return axios.get(`${BASE_URL}/training/${trainingId}/my-record`)
}
