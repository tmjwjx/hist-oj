import api from '@/common/api'

// 课堂模块只关心 HOJ 原生题目详情，统一在这里适配公共响应格式。
export async function getClassroomProblem(request) {
  const problemId = typeof request === 'object' ? request.pid : request
  const response = await api.getProblem(problemId, '0', undefined)
  const body = response && response.data ? response.data : {}

  return {
    code: body.status || body.code || response.status,
    message: body.message,
    data: body.data
  }
}
