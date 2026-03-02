<template>
  <div class="pptx-cos-viewer" ref="viewer">
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

    <!-- 保护层（防止右键） -->
    <div
      class="protection-overlay"
      @contextmenu.prevent
      v-if="!error && !loading"
    ></div>
  </div>
</template>

<script>
export default {
  name: 'PPTXCOSViewer',
  props: {
    materialId: {
      type: [String, Number],
      required: true
    },
    fileName: {
      type: String,
      default: 'presentation.pptx'
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
      uploadMessage: null
    }
  },
  mounted() {
    console.log('[COS Viewer] 组件已挂载', {
      materialId: this.materialId,
      fileName: this.fileName
    })
    this.init()
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
          // 使用缓存的URL（不触发后端请求，不产生CI费用）
          const data = JSON.parse(cachedData)
          previewUrl = data.previewUrl
          uploaded = data.uploaded
          console.log('[COS Viewer] 使用缓存的预览URL')
        } else {
          // 第一次预览，请求后端（会触发CI处理）
          const response = await this.$axios.get(
            `/rating-api/api/classroom/material/${this.materialId}/cos-preview-url`
          )

          if (response.data.code === 200) {
            previewUrl = response.data.data.previewUrl
            uploaded = response.data.data.uploaded

            // 缓存预览URL（永久缓存）
            localStorage.setItem(cacheKey, JSON.stringify({
              previewUrl,
              uploaded,
              timestamp: Date.now()
            }))

            console.log('[COS Viewer] 预览URL已获取并缓存', {
              url: previewUrl.substring(0, 100) + '...'
            })
          } else {
            throw new Error(response.data.message || '获取预览URL失败')
          }
        }

        this.previewUrl = previewUrl

        if (uploaded) {
          this.$message.success('预览准备完成')
        } else {
          this.$message.success('文件已上传到云端，预览准备完成')
        }
      } catch (err) {
        console.error('[COS Viewer] 加载失败:', err)
        this.error = '加载失败：' + (err.response?.data?.message || err.message)
        this.loading = false
      }
    },

    onIframeLoad() {
      const loadTime = this.loadStartTime ? ((Date.now() - this.loadStartTime) / 1000).toFixed(1) : 'N/A'
      console.log('[COS Viewer] iframe 加载完成', {
        fileName: this.fileName,
        loadTime: loadTime + '秒'
      })
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
    }
  }
}
</script>

<style scoped>
.pptx-cos-viewer {
  width: 100%;
  height: 100%;
  position: relative;
  background: #f5f5f5;
  overflow: hidden;
  min-height: 600px;
}

.cos-preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
  min-height: 600px;
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

.protection-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 100;
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
