/**
 * 实时同步混入
 * 为页面提供实时数据刷新功能
 *
 * 使用方式:
 * 1. 在组件中引入: import realtimeSync from '@/mixins/realtimeSync'
 * 2. 混入: mixins: [realtimeSync]
 * 3. 配置同步选项:
 *    - realtimeSyncConfig: {
 *        enabled: true,  // 是否启用实时同步
 *        interval: 3000,  // 同步间隔（毫秒），默认3秒
 *        syncFunction: 'loadData',  // 同步调用的方法名
 *        immediate: true,  // 是否立即执行一次
 *        dataKey: 'messages',  // 要对比的数据键名（可选）
 *        mergeStrategy: 'append'  // 合并策略: 'replace' | 'append' | 'prepend'（可选）
 *      }
 * 4. 在需要停止同步时调用: this.stopRealtimeSync()
 * 5. 在需要手动触发同步时调用: this.triggerRealtimeSync()
 */

export default {
  data() {
    return {
      realtimeSyncTimer: null,
      realtimeSyncEnabled: false,
      realtimeSyncInterval: 3000, // 默认3秒，减少闪烁
      realtimeSyncLastTime: 0,
      realtimeSyncMinInterval: 1000, // 最小同步间隔1秒
      realtimeSyncPreviousData: null // 缓存上一次的数据
    }
  },

  mounted() {
    if (this.realtimeSyncConfig && this.realtimeSyncConfig.enabled) {
      this.startRealtimeSync()
    }
  },

  beforeDestroy() {
    this.stopRealtimeSync()
  },

  methods: {
    /**
     * 启动实时同步
     */
    startRealtimeSync() {
      if (!this.realtimeSyncConfig) {
        return
      }

      const config = this.realtimeSyncConfig
      this.realtimeSyncInterval = config.interval || 3000
      this.realtimeSyncEnabled = true

      // 如果配置了立即执行，先执行一次
      if (config.immediate !== false) {
        this.$nextTick(() => {
          this.triggerRealtimeSync()
        })
      }

      // 启动定时器
      this.startSyncTimer()
    },

    /**
     * 停止实时同步
     */
    stopRealtimeSync() {
      if (this.realtimeSyncTimer) {
        clearInterval(this.realtimeSyncTimer)
        this.realtimeSyncTimer = null
        this.realtimeSyncEnabled = false
      }
    },

    /**
     * 启动同步定时器
     */
    startSyncTimer() {
      // 清除旧的定时器
      if (this.realtimeSyncTimer) {
        clearInterval(this.realtimeSyncTimer)
      }

      // 创建新的定时器
      this.realtimeSyncTimer = setInterval(() => {
        this.triggerRealtimeSync()
      }, this.realtimeSyncInterval)
    },

    /**
     * 触发同步
     * 会检查最小间隔，防止过于频繁调用
     */
    async triggerRealtimeSync() {
      if (!this.realtimeSyncEnabled) {
        return
      }

      const now = Date.now()
      const timeSinceLastSync = now - this.realtimeSyncLastTime

      // 检查是否满足最小间隔
      if (timeSinceLastSync < this.realtimeSyncMinInterval) {
        return
      }

      this.realtimeSyncLastTime = now

      try {
        const syncFunction = this.realtimeSyncConfig.syncFunction
        if (syncFunction && typeof this[syncFunction] === 'function') {
          await this[syncFunction]()
        }
      } catch (error) {
        // 静默处理错误
      }
    },

    /**
     * 更新同步间隔
     * @param {number} interval - 新的同步间隔（毫秒）
     */
    updateSyncInterval(interval) {
      if (interval < this.realtimeSyncMinInterval) {
        interval = this.realtimeSyncMinInterval
      }

      this.realtimeSyncInterval = interval

      if (this.realtimeSyncEnabled) {
        this.startSyncTimer()
      }
    },

    /**
     * 检查数据是否有变化
     * @param {Array|Object} newData - 新数据
     * @param {Array|Object} oldData - 旧数据
     * @returns {boolean} - 是否有变化
     */
    hasDataChanged(newData, oldData) {
      // 如果没有旧数据，认为有变化
      if (!oldData) return true

      // 数组类型对比
      if (Array.isArray(newData) && Array.isArray(oldData)) {
        if (newData.length !== oldData.length) return true

        // 深度对比数组元素（使用 JSON 序列化简化对比）
        return JSON.stringify(newData) !== JSON.stringify(oldData)
      }

      // 对象类型对比
      if (typeof newData === 'object' && typeof oldData === 'object') {
        return JSON.stringify(newData) !== JSON.stringify(oldData)
      }

      // 基本类型对比
      return newData !== oldData
    },

    /**
     * 智能更新数据，避免不必要的重新渲染
     * @param {string} dataKey - 数据键名
     * @param {Array|Object} newData - 新数据
     * @param {string} strategy - 合并策略: 'replace' | 'append' | 'prepend'
     */
    smartUpdateData(dataKey, newData, strategy = 'replace') {
      if (!dataKey || !this.hasOwnProperty(dataKey)) {
        // 如果没有指定 dataKey 或者属性不存在，直接赋值
        return false
      }

      const currentData = this[dataKey]
      const hasChanged = this.hasDataChanged(newData, currentData)

      if (!hasChanged) {
        // 数据没有变化，跳过更新
        return false
      }

      // 根据策略更新数据
      if (strategy === 'append' && Array.isArray(newData) && Array.isArray(currentData)) {
        // 追加新数据：只添加不在当前数据中的项
        const existingIds = new Set(currentData.map(item => item.id || item.uuid))
        const newItems = newData.filter(item => !existingIds.has(item.id || item.uuid))
        if (newItems.length > 0) {
          this[dataKey] = [...currentData, ...newItems]
        }
      } else if (strategy === 'prepend' && Array.isArray(newData) && Array.isArray(currentData)) {
        // 前置新数据
        const existingIds = new Set(currentData.map(item => item.id || item.uuid))
        const newItems = newData.filter(item => !existingIds.has(item.id || item.uuid))
        if (newItems.length > 0) {
          this[dataKey] = [...newItems, ...currentData]
        }
      } else {
        // 默认策略：直接替换
        this[dataKey] = newData
      }

      return true
    },

    /**
     * 暂停同步（不停止定时器，只是跳过执行）
     */
    pauseSync() {
      this.realtimeSyncEnabled = false
    },

    /**
     * 恢复同步
     */
    resumeSync() {
      if (!this.realtimeSyncTimer) {
        return
      }

      this.realtimeSyncEnabled = true
    }
  }
}
