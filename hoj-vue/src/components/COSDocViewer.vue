<template>
  <div class="cos-doc-viewer" ref="viewer" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- 全屏按钮（悬浮） -->
    <div v-if="!loading && !error" class="fullscreen-btn" @click="toggleFullscreen">
      <i :class="isFullscreen ? 'el-icon-close' : 'el-icon-full-screen'"></i>
      <span>{{ isFullscreen ? '退出全屏' : '全屏' }}</span>
    </div>

    <!-- 腾讯云COS文档预览iframe -->
    <iframe
      v-if="previewUrl && !error"
      :src="previewUrl"
      class="cos-preview-iframe"
      frameborder="0"
      @load="onIframeLoad"
    ></iframe>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <p>正在加载预览...</p>
      <p class="hint-text">文件较大时可能需要10-30秒</p>
      <el-progress
        v-if="loadStartTime"
        :percentage="getLoadProgress()"
        :stroke-width="6"
        :show-text="false"
        style="width: 200px; margin-top: 12px;"
      ></el-progress>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-container">
      <i class="el-icon-warning-outline"></i>
      <p class="error-message">{{ error }}</p>
      <el-button type="primary" @click="retry" size="small">
        重试
      </el-button>
      <el-button
        v-if="allowDownload"
        type="text"
        @click="downloadFile"
        size="small"
      >
        下载文件
      </el-button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'COSDocViewer',
  props: {
    materialId: {
      type: [String, Number],
      required: true
    },
    fileName: {
      type: String,
      default: 'document'
    },
    allowDownload: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      loading: true,
      error: null,
      previewUrl: null,
      loadStartTime: null,
      isFullscreen: false
    }
  },
  mounted() {
    this.init()
    // 监听全屏变化事件
    document.addEventListener('fullscreenchange', this.onFullscreenChange)
    document.addEventListener('webkitfullscreenchange', this.onFullscreenChange)
    document.addEventListener('mozfullscreenchange', this.onFullscreenChange)
    document.addEventListener('MSFullscreenChange', this.onFullscreenChange)
  },
  beforeDestroy() {
    document.removeEventListener('fullscreenchange', this.onFullscreenChange)
    document.removeEventListener('webkitfullscreenchange', this.onFullscreenChange)
    document.removeEventListener('mozfullscreenchange', this.onFullscreenChange)
    document.removeEventListener('MSFullscreenChange', this.onFullscreenChange)
    // 退出全屏
    if (this.isFullscreen) {
      this.exitFullscreen()
    }
  },
  methods: {
    async init() {
      try {
        this.loadStartTime = Date.now()
        this.loading = true

        // 检查本地缓存
        const cacheKey = `cos_preview_${this.materialId}`
        const cachedData = localStorage.getItem(cacheKey)

        let previewUrl
        let uploaded = false

        if (cachedData) {
          // 验证缓存数据是否有效
          const data = JSON.parse(cachedData)
          // 如果缓存的URL无效（旧版本bug），清除缓存并重新请求
          if (data.previewUrl && data.previewUrl.startsWith('http')) {
            previewUrl = data.previewUrl
            uploaded = data.uploaded
          } else {
            // 清除无效的缓存
            localStorage.removeItem(cacheKey)
          }
        }

        // 如果没有有效缓存，请求后端
        if (!previewUrl) {
          const response = await this.$axios.get(
            `/rating-api/api/classroom/material/${this.materialId}/cos-preview-url`
          )

          if (response.data.code === 200) {
            previewUrl = response.data.data.previewUrl
            uploaded = response.data.data.uploaded

            // 只缓存有效的URL
            if (previewUrl && previewUrl.startsWith('http')) {
              localStorage.setItem(cacheKey, JSON.stringify({
                previewUrl,
                uploaded,
                timestamp: Date.now()
              }))
            }
          } else {
            throw new Error(response.data.message || '获取预览URL失败')
          }
        }

        this.previewUrl = previewUrl

        // 验证previewUrl是否有效
        if (!previewUrl || typeof previewUrl !== 'string' || !previewUrl.startsWith('http')) {
          console.error('[COS Viewer] 预览URL无效:', previewUrl)
          // 清除无效的缓存
          localStorage.removeItem(cacheKey)
          throw new Error('预览URL格式错误，请刷新页面重试')
        }

        if (uploaded) {
          this.$message.success('预览准备完成')
        } else {
          this.$message.success('文件已上传到云端，预览准备完成')
        }
      } catch (err) {
        console.error('[COS Viewer] 加载失败:', err.message)
        this.error = '加载失败：' + (err.response?.data?.message || err.message)
        this.loading = false
      }
    },

    onIframeLoad() {
      this.loading = false
      this.$emit('viewer-loaded')
    },

    getLoadProgress() {
      if (!this.loadStartTime) return 0
      const elapsed = Date.now() - this.loadStartTime
      const progress = Math.min((elapsed / 30000) * 100, 95)
      return Math.round(progress)
    },

    retry() {
      this.error = null
      this.loading = true
      this.loadStartTime = Date.now()
      this.init()
    },

    downloadFile() {
      const downloadUrl = `/rating-api/api/classroom/material/${this.materialId}/download`
      const link = document.createElement('a')
      link.href = downloadUrl
      link.download = this.fileName
      link.click()
    },

    toggleFullscreen() {
      if (this.isFullscreen) {
        this.exitFullscreen()
      } else {
        this.enterFullscreen()
      }
    },

    enterFullscreen() {
      const element = this.$refs.viewer
      if (element.requestFullscreen) {
        element.requestFullscreen()
      } else if (element.webkitRequestFullscreen) {
        element.webkitRequestFullscreen()
      } else if (element.mozRequestFullScreen) {
        element.mozRequestFullScreen()
      } else if (element.msRequestFullscreen) {
        element.msRequestFullscreen()
      }
    },

    exitFullscreen() {
      if (document.exitFullscreen) {
        document.exitFullscreen()
      } else if (document.webkitExitFullscreen) {
        document.webkitExitFullscreen()
      } else if (document.mozCancelFullScreen) {
        document.mozCancelFullScreen()
      } else if (document.msExitFullscreen) {
        document.msExitFullscreen()
      }
    },

    onFullscreenChange() {
      this.isFullscreen = !!(
        document.fullscreenElement ||
        document.webkitFullscreenElement ||
        document.mozFullScreenElement ||
        document.msFullscreenElement
      )
    }
  }
}
</script>

<style scoped>
.cos-doc-viewer {
  width: 100%;
  height: 100%;
  position: relative;
  background: #f5f5f5;
  overflow: hidden;
  min-height: 600px;
}

.cos-doc-viewer.is-fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  width: 100vw;
  height: 100vh;
  min-height: 100vh;
}

.cos-preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
  min-height: 600px;
}

.fullscreen-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  transition: all 0.3s;
}

.fullscreen-btn:hover {
  background: rgba(0, 0, 0, 0.9);
  transform: scale(1.05);
}

.loading-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  min-height: 400px;
  color: #409EFF;
  padding: 40px;
}

.loading-container i {
  font-size: 48px;
  margin-bottom: 16px;
  animation: rotating 2s linear infinite;
}

.hint-text {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.error-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  min-height: 400px;
  color: #F56C6C;
  padding: 40px;
}

.error-container i {
  font-size: 64px;
  margin-bottom: 16px;
}

.error-message {
  margin: 8px 0 16px 0;
  font-size: 14px;
  color: #606266;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
