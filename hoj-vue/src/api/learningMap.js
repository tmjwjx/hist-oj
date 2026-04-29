import axios from 'axios'

const getBaseURL = () => '/rating-api/api'

const learningMapRequest = axios.create({
  baseURL: getBaseURL(),
  timeout: 15000,
  headers: {
    'Cache-Control': 'no-cache',
    Pragma: 'no-cache'
  }
})

learningMapRequest.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = token
    }
    return config
  },
  error => Promise.reject(error)
)

learningMapRequest.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code === 200) {
      return res.data
    }
    return Promise.reject(new Error(res.message || 'Error'))
  },
  error => Promise.reject(error)
)

export default {
  // 用户端
  getPublishedMaps() {
    return learningMapRequest.get('/learning-maps')
  },
  getMapGraph(mapId) {
    return learningMapRequest.get(`/learning-maps/${mapId}`)
  },
  getMapProgress(mapId) {
    return learningMapRequest.get(`/learning-maps/${mapId}/progress`)
  },
  getMapFull(mapId) {
    return learningMapRequest.get(`/learning-maps/${mapId}/full`)
  },
  startNode(mapId, nodeId) {
    return learningMapRequest.post(`/learning-maps/${mapId}/nodes/${nodeId}/start`)
  },
  completeNode(mapId, nodeId) {
    return learningMapRequest.post(`/learning-maps/${mapId}/nodes/${nodeId}/complete`)
  },
  searchNodes(mapId, keyword) {
    return learningMapRequest.get(`/learning-maps/${mapId}/search`, {
      params: { q: keyword }
    })
  },
  recommendNext(mapId) {
    return learningMapRequest.get(`/learning-maps/${mapId}/recommend-next`)
  },

  // 管理员端
  adminListMaps() {
    return learningMapRequest.get('/admin/learning-maps')
  },
  adminGetMap(mapId) {
    return learningMapRequest.get(`/admin/learning-maps/${mapId}`)
  },
  adminCreateMap(data) {
    return learningMapRequest.post('/admin/learning-maps', data)
  },
  adminUpdateMap(mapId, data) {
    return learningMapRequest.put(`/admin/learning-maps/${mapId}`, data)
  },
  adminDeleteMap(mapId) {
    return learningMapRequest.delete(`/admin/learning-maps/${mapId}`)
  },
  adminCreateNode(mapId, data) {
    return learningMapRequest.post(`/admin/learning-maps/${mapId}/nodes`, data)
  },
  adminUpdateNode(mapId, nodeId, data) {
    return learningMapRequest.put(`/admin/learning-maps/${mapId}/nodes/${nodeId}`, data)
  },
  adminDeleteNode(mapId, nodeId) {
    return learningMapRequest.delete(`/admin/learning-maps/${mapId}/nodes/${nodeId}`)
  },
  adminCreateEdge(mapId, data) {
    return learningMapRequest.post(`/admin/learning-maps/${mapId}/edges`, data)
  },
  adminUpdateEdge(mapId, edgeId, data) {
    return learningMapRequest.put(`/admin/learning-maps/${mapId}/edges/${edgeId}`, data)
  },
  adminDeleteEdge(mapId, edgeId) {
    return learningMapRequest.delete(`/admin/learning-maps/${mapId}/edges/${edgeId}`)
  },
  adminPublishMap(mapId) {
    return learningMapRequest.post(`/admin/learning-maps/${mapId}/publish`)
  },
  adminValidateMap(mapId) {
    return learningMapRequest.get(`/admin/learning-maps/${mapId}/validate`)
  },
  adminSearchProblems(keyword) {
    return learningMapRequest.get('/admin/problems/search', { params: { q: keyword } })
  },
  adminGetProblem(identifier) {
    return learningMapRequest.get(`/admin/problems/${identifier}`)
  }
}
