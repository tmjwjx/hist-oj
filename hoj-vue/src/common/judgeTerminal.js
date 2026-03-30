import axios from 'axios'

/**
 * 获取题目信息
 * @param {Object} data - 请求参数
 * @param {string} data.pid - 题目ID
 * @param {string} data.cid - 比赛ID（可选）
 * @param {string} data.mode - 模式（normal/contest）
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码（可选，如果不提供则使用 token）
 */
export function getJudgeInfo(data) {
  // 如果没有提供密码，尝试使用 BingOJ token
  const token = localStorage.getItem('token')

  if (token && !data.password) {
    // 移除 password 字段
    const { password, ...dataWithoutPassword } = data
    data = dataWithoutPassword
    // 添加 token 字段到请求体
    data.token = token
  }

  return axios.post('/judge-api/get-info', data).then(res => res.data)
}

/**
 * 获取判题历史记录（分页）
 * @param {Object} data - 请求参数
 * @param {string} data.pid - 题目ID
 * @param {string} data.cid - 比赛ID（可选）
 * @param {number} data.page - 页码（从1开始）
 * @param {number} data.pageSize - 每页条数
 */
export function getJudgeHistory(data) {
  const token = localStorage.getItem('token')

  if (token && !data.password) {
    // 移除 password 字段
    const { password, ...dataWithoutPassword } = data
    data = dataWithoutPassword
    // 添加 token 字段到请求体
    data.token = token
  }

  return axios.post('/judge-api/get-history', data).then(res => res.data)
}

/**
 * 获取远程判题测试点详情
 * @param {Object} data - 请求参数
 * @param {string} data.submit_id - 提交ID
 */
export function getJudgeCaseDetails(data) {
  const token = localStorage.getItem('token')

  if (token && !data.password) {
    const { password, ...dataWithoutPassword } = data
    data = dataWithoutPassword
    data.token = token
  }

  return axios.post('/judge-api/get-case-details', data).then(res => res.data)
}

/**
 * 本地测试并提交代码（SSE 流式接口）
 * @param {Object} data - 请求参数
 * @param {string} data.pid - 题目ID
 * @param {string} data.cid - 比赛ID（可选）
 * @param {string} data.mode - 模式（normal/contest）
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码（可选，如果不提供则使用 token）
 * @param {string} data.language - 编程语言
 * @param {string} data.code - 代码
 * @param {Function} onMessage - 消息回调函数
 * @param {Function} onError - 错误回调函数
 * @param {Function} onComplete - 完成回调函数
 */
export function runCombinedJudge(data, onMessage, onError, onComplete) {
  // 如果没有提供密码，尝试使用 BingOJ token
  const token = localStorage.getItem('token')

  if (token && !data.password) {
    // 移除 password 字段
    const { password, ...dataWithoutPassword } = data
    data = dataWithoutPassword
    // 添加 token 字段到请求体
    data.token = token
  }

  // 由于 EventSource 不支持 POST，我们需要使用 fetch
  const url = `/judge-api/run-combined`

  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data)
  }).then(response => {
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    function handleEvent(rawEvent) {
      if (!rawEvent) return
      const lines = rawEvent.split('\n')
      const dataLines = lines
        .filter(line => line.startsWith('data:'))
        .map(line => line.replace(/^data:\s?/, ''))

      if (dataLines.length === 0) return

      const payload = dataLines.join('\n')
      try {
        const parsed = JSON.parse(payload)
        if (onMessage) onMessage(parsed)
      } catch (e) {
        console.error('解析 SSE 消息失败:', e, payload)
      }
    }

    function read() {
      reader.read().then(({ done, value }) => {
        if (done) {
          // 处理最后一段未以 \n\n 结尾的数据
          if (buffer.trim()) {
            handleEvent(buffer)
          }
          if (onComplete) onComplete()
          return
        }

        // 带缓冲的 SSE 解析，避免 JSON 被分片截断
        buffer += decoder.decode(value, { stream: true })
        const events = buffer.split('\n\n')
        buffer = events.pop() || ''
        events.forEach(handleEvent)

        // 继续读取
        read()
      }).catch(err => {
        if (onError) onError(err)
      })
    }

    read()
  }).catch(err => {
    if (onError) onError(err)
  })
}
