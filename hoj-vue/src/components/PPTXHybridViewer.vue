<template>
  <div class="pptx-hybrid-viewer" v-loading="loading" element-loading-text="正在加载PPT...">
    <!-- PPTX内容 -->
    <PPTXViewer
      v-if="!error && fileUrl"
      :fileUrl="fileUrl"
      :fileName="fileName"
      :key="fileUrl"
      @error="handleViewerError"
    />

    <!-- 错误提示 -->
    <div v-if="error" class="preview-unsupported">
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
import PPTXViewer from './PPTXViewer.vue'

export default {
  name: 'PPTXHybridViewer',
  components: {
    PPTXViewer
  },
  props: {
    fileUrl: {
      type: String,
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
      loading: false,
      error: null
    }
  },

  mounted() {
    console.log('[PPTXHybridViewer] 组件已挂载', {
      fileUrl: this.fileUrl ? this.fileUrl.substring(0, 80) + '...' : 'empty',
      fileName: this.fileName
    })
  },
  methods: {
    handleViewerError(errorMessage) {
      this.error = errorMessage || 'PPT文件加载失败，请稍后重试'
      this.loading = false
    },

    retry() {
      this.error = null
      this.loading = true
      // 强制重新渲染PPTXViewer
      this.$forceUpdate()
    },

    downloadFile() {
      if (this.fileUrl) {
        window.open(this.fileUrl, '_blank')
      }
    }
  }
}
</script>

<style scoped>
.pptx-hybrid-viewer {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
  position: relative;
}

.preview-unsupported {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #909399;
  padding: 40px;
  text-align: center;
}

.preview-unsupported i {
  font-size: 64px;
  color: #C0C4CC;
  margin-bottom: 16px;
}

.error-message {
  margin: 8px 0 16px 0;
  font-size: 14px;
  color: #606266;
}
</style>
