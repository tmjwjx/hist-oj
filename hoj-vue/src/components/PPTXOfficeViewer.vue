<template>
  <div class="pptx-office-viewer" ref="viewer">
    <!-- Office Online 工具栏遮罩（隐藏下载按钮） -->
    <div v-if="!error && !loading && !allowDownload" class="office-toolbar-mask"></div>

    <!-- Office Online iframe -->
    <iframe
      v-if="!error && previewUrl"
      :src="officeViewerUrl"
      class="office-iframe"
      frameborder="0"
      @load="onIframeLoad"
      allowfullscreen
    ></iframe>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <p>正在加载Office Online预览...</p>
      <p class="hint-text">首次加载大文件可能需要30-60秒，请耐心等待</p>
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

    <!-- Office Online错误提示（用户手动触发） -->
    <div v-if="!error && !loading" class="office-error-hint">
      <el-alert
        v-if="showOfficeErrorHint"
        title="Office Online预览说明"
        type="info"
        :closable="true"
        @close="showOfficeErrorHint = false"
        style="margin-bottom: 10px;"
      >
        <template slot="default">
          <p style="margin: 5px 0; font-weight: bold;">如果页面显示 "An error occurred" 或其他Office Online错误：</p>
          <p style="margin: 5px 0;">这表示微软Office Online服务无法打开此文件，可能原因：</p>
          <ul style="margin: 5px 0; padding-left: 20px; font-size: 13px;">
            <li>文件使用了不兼容的高级功能（如复杂动画、VBA宏等）</li>
            <li>文件内容过于复杂（大量多媒体、图表等）</li>
            <li>文件格式损坏或不标准</li>
            <li>Office Online服务暂时不可用</li>
          </ul>
          <p style="margin: 5px 0; font-weight: bold;">建议解决方案：</p>
          <ul style="margin: 5px 0; padding-left: 20px; font-size: 13px;">
            <li>尝试用PowerPoint重新保存文件，去除不兼容的功能</li>
            <li>简化动画和多媒体内容</li>
            <li>下载文件后使用本地PowerPoint打开</li>
          </ul>
        </template>
      </el-alert>
      <el-button
        type="warning"
        size="small"
        icon="el-icon-info"
        @click="showOfficeErrorHint = !showOfficeErrorHint"
        plain
      >
        预览显示错误？
      </el-button>
    </div>

    <!-- 保护层（防止右键和添加水印） -->
    <div
      class="protection-overlay"
      @contextmenu.prevent
      @click.right.prevent
      @mousedown.right.prevent
      @mouseup.right.prevent
      v-if="!error && !loading"
    >
      <!-- 水印 -->
      <div class="watermark-container" v-if="showWatermark">
        <div
          v-for="i in 20"
          :key="i"
          class="watermark"
          :style="getWatermarkStyle(i)"
        >
          {{ watermarkText }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PPTXOfficeViewer',
  props: {
    previewUrl: {
      type: String,
      required: true
    },
    fileName: {
      type: String,
      default: 'presentation.pptx'
    },
    userName: {
      type: String,
      default: '用户'
    },
    showWatermark: {
      type: Boolean,
      default: true
    },
    allowDownload: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      loading: true,
      error: null,
      loadStartTime: null,
      showOfficeErrorHint: false
    }
  },
  computed: {
    officeViewerUrl() {
      // 使用微软Office Online服务
      // wdStart=0: 从第1页开始
      // wdEmbed=0: 嵌入模式
      return `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(this.previewUrl)}&wdStart=0&wdEmbed=0`
    },
    watermarkText() {
      return `${this.userName} - 仅用于学习`
    }
  },
  mounted() {
    console.log('[OfficeViewer] 组件已挂载', {
      fileName: this.fileName,
      previewUrl: this.previewUrl
    })

    console.log('[OfficeViewer] Office Online URL:', this.officeViewerUrl)

    // 设置保护措施
    this.preventDownload()

    // 记录开始时间
    this.loadStartTime = Date.now()

    // 不设置超时，让Office Online无限制加载
    console.log('[OfficeViewer] 开始加载，无超时限制')
  },
  beforeDestroy() {
    this.cleanup()
  },
  methods: {
    onIframeLoad() {
      const loadTime = this.loadStartTime ? ((Date.now() - this.loadStartTime) / 1000).toFixed(1) : 'N/A'
      console.log('[OfficeViewer] Office Online iframe 加载完成', {
        fileName: this.fileName,
        loadTime: loadTime + '秒'
      })

      this.loading = false

      // 提示用户：如果看到Office Online错误
      const loadTimeSeconds = parseFloat(loadTime)
      if (loadTimeSeconds < 10) {
        // 加载很快可能意味着错误页面
        console.log('[OfficeViewer] 加载完成，如果页面显示错误，点击"预览显示错误？"查看帮助')
      }

      // 发射加载完成事件，用于缓存
      this.$emit('viewer-loaded')
    },

    preventDownload() {
      // 1. 禁用右键菜单
      document.addEventListener('contextmenu', this.handleContextMenu, false)

      // 2. 禁用常用快捷键
      document.addEventListener('keydown', this.handleKeyDown, false)

      // 3. 禁用拖拽
      document.addEventListener('dragstart', this.handleDragStart, false)

      console.log('[OfficeViewer] 保护措施已启用')
    },

    handleContextMenu(e) {
      e.preventDefault()
      return false
    },

    handleKeyDown(e) {
      // Ctrl+S (保存)
      if (e.ctrlKey && (e.key === 's' || e.key === 'S')) {
        e.preventDefault()
        this.$message?.warning('禁止保存文件')
        return false
      }

      // Ctrl+P (打印)
      if (e.ctrlKey && (e.key === 'p' || e.key === 'P')) {
        e.preventDefault()
        this.$message?.warning('禁止打印')
        return false
      }

      // F12 / Ctrl+Shift+I (开发者工具) - 只提醒，不禁用
      if (e.key === 'F12' || (e.ctrlKey && e.shiftKey && (e.key === 'I' || e.key === 'i'))) {
        // 不禁用开发者工具，因为用户可能需要调试
        // 但记录日志
        console.warn('[OfficeViewer] 检测到开发者工具快捷键')
      }

      // PrintScreen - 无法完全阻止，但可以检测
      if (e.key === 'PrintScreen') {
        this.$message?.warning('请遵守版权，禁止截屏传播')
      }
    },

    handleDragStart(e) {
      e.preventDefault()
      return false
    },

    getWatermarkStyle(index) {
      // 生成多个水印，分布在不同的位置和角度
      const positions = [
        { top: '10%', left: '10%', transform: 'rotate(-30deg)' },
        { top: '10%', right: '10%', transform: 'rotate(30deg)' },
        { top: '50%', left: '10%', transform: 'rotate(-20deg)' },
        { top: '50%', right: '10%', transform: 'rotate(20deg)' },
        { bottom: '10%', left: '10%', transform: 'rotate(-35deg)' },
        { bottom: '10%', right: '10%', transform: 'rotate(35deg)' },
        { top: '30%', left: '50%', transform: 'rotate(-25deg)' },
        { top: '70%', left: '50%', transform: 'rotate(25deg)' },
      ]

      const pos = positions[index % positions.length]
      return {
        ...pos,
        opacity: 0.08,
        fontSize: '24px'
      }
    },

    retry() {
      this.error = null
      this.loading = true
      this.$forceUpdate()
    },

    downloadFile() {
      if (this.previewUrl) {
        window.open(this.previewUrl, '_blank')
      }
    },

    getLoadProgress() {
      if (!this.loadStartTime) return 0
      const elapsed = Date.now() - this.loadStartTime
      const progress = Math.min((elapsed / 60000) * 100, 95) // 60秒到95%
      return Math.round(progress)
    },

    cleanup() {
      document.removeEventListener('contextmenu', this.handleContextMenu)
      document.removeEventListener('keydown', this.handleKeyDown)
      document.removeEventListener('dragstart', this.handleDragStart)
    }
  }
}
</script>

<style scoped>
.pptx-office-viewer {
  width: 100%;
  height: 100%;
  position: relative;
  background: #f5f5f5;
  overflow: hidden;
  min-height: 600px; /* 确保最小高度 */
}

.office-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
  min-height: 600px; /* 确保iframe的最小高度 */
}

/* Office Online 工具栏遮罩 */
.office-toolbar-mask {
  position: absolute;
  top: 0;
  right: 0;
  width: 200px; /* 覆盖右侧工具栏区域（包含下载按钮） */
  height: 50px; /* Office Online工具栏高度 */
  background: transparent;
  z-index: 1000;
  pointer-events: auto; /* 允许接收点击事件，阻止访问iframe */
  cursor: default;
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

.loading-container p {
  margin: 8px 0;
  font-size: 16px;
  font-weight: 500;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.hint-text {
  margin-top: 8px;
  font-size: 14px;
  color: #909399;
}

.error-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #F56C6C;
  padding: 40px;
  text-align: center;
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
  overflow: hidden;
}

.watermark-container {
  position: relative;
  width: 100%;
  height: 100%;
}

.watermark {
  position: absolute;
  color: rgba(0, 0, 0, 0.08);
  font-size: 24px;
  font-weight: bold;
  white-space: nowrap;
  user-select: none;
  pointer-events: none;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.office-error-hint {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 99999;
  text-align: right;
}

.office-error-hint .el-alert {
  max-width: 450px;
  text-align: left;
  font-size: 13px;
}

.office-error-hint .el-button {
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid #e6a23c;
  box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.2);
  font-weight: 500;
  padding: 10px 16px;
}

.office-error-hint .el-button:hover {
  background: #fdf6ec;
  box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.3);
}
</style>
