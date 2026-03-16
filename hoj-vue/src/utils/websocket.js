/**
 * WebSocket 客户端管理器
 * 用于实时接收报名系统的更新
 */

class WebSocketManager {
  constructor() {
    this.ws = null
    this.reconnectTimer = null
    this.reconnectDelay = 3000 // 3秒后重连
    this.listeners = new Map() // 事件监听器
    this.competitionId = null
    this.isConnected = false
    this.isManualClose = false
  }

  /**
   * 连接 WebSocket
   * @param {number|string} competitionId - 比赛ID
   */
  connect(competitionId) {
    if (this.ws && this.isConnected) {
      return
    }

    this.competitionId = competitionId
    this.isManualClose = false

    // 构建WebSocket URL，通过查询参数传递token
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const token = localStorage.getItem('token') || ''
    const wsUrl = `${protocol}//${host}/api/hist-oj/registration/ws/${competitionId}?token=${encodeURIComponent(token)}`


    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.isConnected = true
        this.clearReconnectTimer()
        this.emit('connected')
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (error) {
          console.error('解析 WebSocket 消息失败:', error, event.data)
        }
      }

      this.ws.onerror = (error) => {
        console.error('❌ WebSocket 错误:', error)
        this.emit('error', error)
      }

      this.ws.onclose = (event) => {
        console.log('🔌 WebSocket 连接关闭', event.code, event.reason)
        this.isConnected = false
        this.emit('disconnected', event)

        // 如果不是手动关闭，则自动重连
        if (!this.isManualClose) {
          this.scheduleReconnect()
        }
      }
    } catch (error) {
      console.error('创建 WebSocket 连接失败:', error)
      this.scheduleReconnect()
    }
  }

  /**
   * 断开连接
   */
  disconnect() {
    this.isManualClose = true
    this.clearReconnectTimer()

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }

    this.isConnected = false
  }

  /**
   * 安排重连
   */
  scheduleReconnect() {
    if (this.reconnectTimer) return

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      if (!this.isManualClose && this.competitionId) {
        this.connect(this.competitionId)
      }
    }, this.reconnectDelay)
  }

  /**
   * 清除重连定时器
   */
  clearReconnectTimer() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  /**
   * 处理收到的消息
   */
  handleMessage(message) {
    const { type, data, timestamp } = message

    switch (type) {
      case 'registration_update':
        this.emit('registrationUpdate', data)
        break

      case 'new_message':
        this.emit('newMessage', data)
        break

      case 'competition_update':
        this.emit('competitionUpdate', data)
        break

      default:
        console.warn('未知的消息类型:', type)
        this.emit('unknown', message)
    }
  }

  /**
   * 添加事件监听器
   * @param {string} event - 事件名称
   * @param {Function} callback - 回调函数
   */
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event).push(callback)
  }

  /**
   * 移除事件监听器
   * @param {string} event - 事件名称
   * @param {Function} callback - 回调函数
   */
  off(event, callback) {
    if (!this.listeners.has(event)) return

    const callbacks = this.listeners.get(event)
    const index = callbacks.indexOf(callback)
    if (index > -1) {
      callbacks.splice(index, 1)
    }
  }

  /**
   * 触发事件
   * @param {string} event - 事件名称
   * @param {*} data - 事件数据
   */
  emit(event, data) {
    if (!this.listeners.has(event)) return

    const callbacks = this.listeners.get(event)
    callbacks.forEach(callback => {
      try {
        callback(data)
      } catch (error) {
        console.error(`事件回调执行错误 [${event}]:`, error)
      }
    })
  }

  /**
   * 获取连接状态
   */
  getState() {
    return {
      isConnected: this.isConnected,
      competitionId: this.competitionId,
      listenerCount: this.listeners.size
    }
  }
}

// 导出单例
export const wsManager = new WebSocketManager()

// 也支持创建新实例
export default WebSocketManager
