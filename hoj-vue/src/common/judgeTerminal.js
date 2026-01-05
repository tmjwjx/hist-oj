import axios from 'axios'

/**
 * 获取题目信息
 * @param {Object} data - 请求参数
 * @param {string} data.pid - 题目ID
 * @param {string} data.cid - 比赛ID（可选）
 * @param {string} data.mode - 模式（normal/contest）
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码
 */
export function getJudgeInfo(data) {
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
  return axios.post('/judge-api/get-history', data).then(res => res.data)
}

/**
 * 本地测试并提交代码（SSE 流式接口）
 * @param {Object} data - 请求参数
 * @param {string} data.pid - 题目ID
 * @param {string} data.cid - 比赛ID（可选）
 * @param {string} data.mode - 模式（normal/contest）
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码
 * @param {string} data.language - 编程语言
 * @param {string} data.code - 代码
 * @param {Function} onMessage - 消息回调函数
 * @param {Function} onError - 错误回调函数
 * @param {Function} onComplete - 完成回调函数
 */
export function runCombinedJudge(data, onMessage, onError, onComplete) {
  // 使用 EventSource 接收 SSE 流
  const url = `/judge-api/run-combined`

  // 由于 EventSource 不支持 POST，我们需要使用 fetch
  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data)
  }).then(response => {
    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    function read() {
      reader.read().then(({ done, value }) => {
        if (done) {
          if (onComplete) onComplete()
          return
        }

        // 解析 SSE 数据
        const text = decoder.decode(value)
        const lines = text.split('\n\n')

        lines.forEach(line => {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.substring(6))
              if (onMessage) onMessage(data)
            } catch (e) {
              console.error('解析 SSE 消息失败:', e)
            }
          }
        })

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
