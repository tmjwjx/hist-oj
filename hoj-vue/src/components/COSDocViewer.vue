<template>
  <div class="cos-doc-viewer" ref="viewer" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- 全屏按钮（悬浮） -->
    <div v-if="!loading && !error" class="fullscreen-btn" @click="toggleFullscreen">
      <i :class="isFullscreen ? 'el-icon-close' : 'el-icon-full-screen'"></i>
      <span>{{ isFullscreen ? '退出全屏' : '全屏' }}</span>
    </div>

    <!-- 视频预览 - 使用TCPlayer -->
    <div v-if="fileType === 'video' && !error" class="video-container">
      <video
        id="tcplayer-container"
        preload="auto"
        playsinline
        webkit-playsinline
        class="preview-video"
      ></video>
    </div>

    <!-- 音频预览 -->
    <div v-else-if="fileType === 'audio' && !error" class="audio-container">
      <audio
        :src="previewUrl"
        controls
        preload="metadata"
        class="preview-audio"
        @loadedmetadata="onMediaLoad"
        @error="onMediaError"
      >
        您的浏览器不支持音频播放
      </audio>
    </div>

    <!-- 图片预览 -->
    <div v-else-if="fileType === 'image' && !error" class="image-container">
      <img
        :src="previewUrl"
        class="preview-image"
        @load="onMediaLoad"
        @error="onMediaError"
      />
    </div>

    <!-- 腾讯云COS文档预览iframe -->
    <iframe
      v-else-if="fileType === 'document' && previewUrl && !error"
      :src="previewUrl"
      class="cos-preview-iframe"
      frameborder="0"
      @load="onIframeLoad"
    ></iframe>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <p>正在加载预览...</p>
      <p class="hint-text">大文件可能需要几秒钟</p>
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
      fileType: null,
      loadStartTime: null,
      isFullscreen: false,
      tcplayer: null
    }
  },
  mounted() {
    this.init()
    // 监听全屏变化事件
    document.addEventListener('fullscreenchange', this.onFullscreenChange)
    document.addEventListener('webkitfullscreenchange', this.onFullscreenChange)
    document.addEventListener('mozfullscreenchange', this.onFullscreenChange)
    document.addEventListener('MSFullscreenChange', this.onFullscreenChange)

    // 动态加载TCPlayer脚本
    this.loadTCPlayerScript()
  },
  beforeDestroy() {
    document.removeEventListener('fullscreenchange', this.onFullscreenChange)
    document.removeEventListener('webkitfullscreenchange', this.onFullscreenChange)
    document.removeEventListener('mozfullscreenchange', this.onFullscreenChange)
    document.removeEventListener('MSFullscreenChange', this.onFullscreenChange)

    // 销毁TCPlayer实例
    if (this.tcplayer) {
      this.tcplayer.dispose()
      this.tcplayer = null
    }

    // 退出全屏
    if (this.isFullscreen) {
      this.exitFullscreen()
    }
  },
  methods: {
    // 动态加载TCPlayer脚本
    loadTCPlayerScript() {
      // 检查是否已加载
      if (window.TCPlayer) {
        return
      }

      // 加载CSS
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = 'https://web.sdk.qcloud.com/player/tcplayer/release/v4.2.1/tcplayer.min.css'
      document.head.appendChild(link)

      // 先加载 hls.js（TCPlayer 依赖）
      const hlsScript = document.createElement('script')
      hlsScript.src = 'https://web.sdk.qcloud.com/player/tcplayer/release/v4.5.0/hls.min.js'
      hlsScript.onload = () => {
        console.log('[COS Viewer] hls.js加载完成')
        // hls.js 加载完成后，再加载 TCPlayer
        const tcplayerScript = document.createElement('script')
        tcplayerScript.src = 'https://web.sdk.qcloud.com/player/tcplayer/release/v4.5.0/tcplayer.v4.5.0.min.js'
        tcplayerScript.onload = () => {
          console.log('[COS Viewer] TCPlayer加载完成')
        }
        document.head.appendChild(tcplayerScript)
      }
      document.head.appendChild(hlsScript)
    },

    async init() {
      try {
        this.loadStartTime = Date.now()
        this.loading = true

        // 检查本地缓存
        const cacheKey = `cos_preview_${this.materialId}`
        const cachedData = localStorage.getItem(cacheKey)

        let previewUrl, fileType, uploaded

        if (cachedData) {
          const data = JSON.parse(cachedData)
          if (data.previewUrl && data.previewUrl.startsWith('http')) {
            previewUrl = data.previewUrl
            fileType = data.fileType
            uploaded = data.uploaded
          } else {
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
            fileType = response.data.data.fileType || this.detectFileType(this.fileName)
            uploaded = response.data.data.uploaded

            // 缓存有效的URL
            if (previewUrl && previewUrl.startsWith('http')) {
              localStorage.setItem(cacheKey, JSON.stringify({
                previewUrl,
                fileType,
                uploaded,
                timestamp: Date.now()
              }))
            }
          } else {
            throw new Error(response.data.message || '获取预览URL失败')
          }
        }

        this.previewUrl = previewUrl
        this.fileType = fileType || this.detectFileType(this.fileName)

        // 根据文件类型处理
        if (this.fileType === 'video') {
          // 视频需要等待TCPlayer加载
          this.initVideoPlayer()
        } else if (this.fileType === 'audio' || this.fileType === 'image') {
          // 音频和图片不需要等待加载完成
          // loading会在 onMediaLoad 中设置为false
        } else {
          // 文档预览会在 iframe load 事件中处理
        }

        if (uploaded) {
          this.$message.success('预览准备完成')
        } else {
          this.$message.success('文件已上传到云端，预览准备完成')
        }
      } catch (err) {
        console.error('[COS Viewer] 加载失败:', err)
        const errorMsg = err.response?.data?.message || err.message || err.toString() || '未知错误'
        this.error = '加载失败：' + errorMsg
        this.loading = false
      }
    },

    // 初始化视频播放器
    initVideoPlayer() {
      const checkTCPlayer = () => {
        if (window.TCPlayer) {
          this.loading = false
          // 创建TCPlayer实例
          this.tcplayer = TCPlayer('tcplayer-container', {
            reportable: false,
            autoplay: false,
            preload: 'auto',
            plugins: {
              ContinuePlay: { // 开启续播功能
                auto: true,
                text: '上次播放至 {time}，继续播放？'
              }
            }
          })
          this.tcplayer.src(this.previewUrl)

          this.$message.success('视频加载完成')
          this.$emit('viewer-loaded')
        } else {
          // TCPlayer还未加载，等待
          setTimeout(checkTCPlayer, 100)
        }
      }

      checkTCPlayer()
    },

    // 检测文件类型
    detectFileType(fileName) {
      const ext = fileName.split('.').pop().toLowerCase()
      const videoTypes = ['mp4', 'webm', 'ogv', 'mov', 'avi', 'mkv', 'flv', 'm4v']
      const audioTypes = ['mp3', 'wav', 'aac', 'ogg', 'm4a', 'flac']
      const imageTypes = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp']
      const documentTypes = ['pdf', 'txt', 'md', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx']

      if (videoTypes.includes(ext)) return 'video'
      if (audioTypes.includes(ext)) return 'audio'
      if (imageTypes.includes(ext)) return 'image'
      if (documentTypes.includes(ext)) return 'document'
      return 'unknown'
    },

    onIframeLoad() {
      this.loading = false
      this.$emit('viewer-loaded')
    },

    onMediaLoad() {
      this.loading = false
      this.$emit('viewer-loaded')
    },

    onMediaError(event) {
      console.error('[COS Viewer] 媒体加载失败:', event)
      this.error = '媒体文件加载失败，请稍后重试'
      this.loading = false
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

      // 清除缓存
      const cacheKey = `cos_preview_${this.materialId}`
      localStorage.removeItem(cacheKey)

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

/* 视频容器 */
.video-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.preview-video {
  width: 100%;
  height: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* 音频容器 */
.audio-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.preview-audio {
  width: 80%;
  max-width: 600px;
  display: block;
}

/* 图片容器 */
.image-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
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
