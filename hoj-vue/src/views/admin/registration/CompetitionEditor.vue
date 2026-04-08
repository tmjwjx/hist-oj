<template>
  <div class="competition-editor-page">
    <div class="container">
      <!-- 头部 -->
      <div class="header">
        <div>
          <h1>{{ isEdit ? '编辑比赛' : '创建新比赛' }}</h1>
          <p>填写比赛信息</p>
        </div>
        <div class="header-actions">
          <router-link to="/admin/registration" class="btn">返回列表</router-link>
        </div>
      </div>

      <!-- 表单 -->
      <form @submit.prevent="saveCompetition">
        <!-- ① 顶部基本信息区 -->
        <div class="basic-info-section">
          <div class="section-header">
            <h2>📌 基本信息</h2>
          </div>

          <div class="info-grid">
            <div class="form-group">
              <label>比赛名称 *</label>
              <input v-model="formData.name" type="text" required placeholder="请输入比赛名称">
            </div>

            <div class="form-row-inline">
              <div class="form-group">
                <label>开始时间 *</label>
                <input v-model="formData.startTime" type="datetime-local" required>
              </div>
              <div class="form-group">
                <label>结束时间 *</label>
                <input v-model="formData.endTime" type="datetime-local" required>
              </div>
            </div>

            <div class="form-group-inline">
              <div class="logo-section">
                <label>比赛Logo</label>
                <div class="logo-upload">
                  <div v-if="formData.logoUrl" class="logo-preview">
                    <img :src="formData.logoUrl" alt="Logo预览">
                    <button type="button" class="remove-btn" @click="removeLogo">删除</button>
                  </div>
                  <div v-else class="upload-placeholder" @click="triggerFileInput">
                    <span class="upload-icon">📤</span>
                    <span>点击上传Logo</span>
                    <span class="upload-hint">支持 JPG、PNG、GIF，最大5MB</span>
                  </div>
                  <input
                    ref="fileInput"
                    type="file"
                    accept="image/png,image/jpeg,image/gif"
                    style="display: none"
                    @change="handleLogoUpload"
                  >
                </div>
              </div>

              <div class="visibility-section">
                <label>比赛可见性</label>
                <label class="checkbox-item-inline">
                  <input type="checkbox" v-model="formData.visible">
                  <span>立即显示（可见）</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- ② 编辑器和预览区（对称布局） -->
        <div class="editor-preview-container">
          <!-- 编辑器区域 -->
          <div class="editor-section">
            <div class="panel-header">
              <h3>📝 Markdown 编辑器</h3>
            </div>
            <div class="editor-wrapper">
              <div class="editor-toolbar">
                <button type="button" class="toolbar-btn" @click="insertMarkdown('**', '**')" title="粗体">
                  <strong>B</strong>
                </button>
                <button type="button" class="toolbar-btn" @click="insertMarkdown('*', '*')" title="斜体">
                  <em>I</em>
                </button>
                <button type="button" class="toolbar-btn" @click="insertMarkdown('### ', '')" title="标题">
                  H
                </button>
                <button type="button" class="toolbar-btn" @click="insertMarkdown('- ', '')" title="列表">
                  •
                </button>
                <button type="button" class="toolbar-btn" @click="triggerMarkdownImageUpload" title="插入图片">
                  🖼️ 图片
                </button>
                <input
                  ref="markdownFileInput"
                  type="file"
                  accept="image/png,image/jpeg,image/gif"
                  style="display: none"
                  @change="handleMarkdownImageUpload"
                >
              </div>
              <textarea
                ref="textarea"
                v-model="formData.description"
                class="markdown-editor"
                placeholder="支持Markdown格式，例如：&#10;## 标题&#10;**粗体文本**&#10;*斜体文本*&#10;- 列表项&#10;&#10;[图片链接](图片URL)&#10;```&#10;代码块&#10;```"
              ></textarea>
            </div>
          </div>

          <!-- 预览区域 -->
          <div class="preview-section">
            <div class="panel-header">
              <h3>👁️ 实时预览</h3>
              <p class="preview-hint">💡 点击预览中的图片可以调整大小</p>
            </div>
            <div class="preview-wrapper">
              <div
                class="preview-content markdown-body"
                v-html="renderedMarkdown"
                @click="handlePreviewClick"
              ></div>
            </div>
          </div>
        </div>

        <!-- ③ 底部配置区 -->
        <div class="fields-config-section">
          <div class="section-header">
            <h2>⚙️ 报名字段配置</h2>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-item" v-for="(label, field) in fieldOptions" :key="field">
              <input type="checkbox" v-model="formData.fields[field]">
              {{ label }}
            </label>
          </div>

          <!-- 操作按钮 -->
          <div class="form-actions">
            <router-link to="/admin/registration" class="btn">取消</router-link>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '保存中...' : (isEdit ? '更新比赛' : '创建比赛') }}
            </button>
          </div>
        </div>
      </form>
    </div>

    <!-- 图片大小调整对话框 -->
    <div v-if="showImageSizeModal" class="modal-overlay" @click.self="closeImageSizeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>🖼️ 调整图片大小</h3>
          <button type="button" class="close-btn" @click="closeImageSizeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="info-notice">
            💡 调整后的图片大小将在点击"应用"后更新到预览，记得点击"更新比赛"保存更改！
          </div>
          <div class="image-preview-container">
            <img :src="currentImageUrl" :style="{ width: currentImageWidth + 'px' }" alt="预览">
          </div>

          <div class="size-controls">
            <div class="control-group">
              <label>宽度（像素）：</label>
              <div class="slider-container">
                <input
                  type="range"
                  v-model.number="currentImageWidth"
                  min="100"
                  max="1200"
                  step="10"
                  class="size-slider"
                >
                <input
                  type="number"
                  v-model.number="currentImageWidth"
                  min="100"
                  max="1200"
                  class="size-input"
                >
              </div>
            </div>

            <div class="preset-buttons">
              <button type="button" class="preset-btn" @click="currentImageWidth = 300">小 (300px)</button>
              <button type="button" class="preset-btn" @click="currentImageWidth = 500">中 (500px)</button>
              <button type="button" class="preset-btn" @click="currentImageWidth = 800">大 (800px)</button>
              <button type="button" class="preset-btn" @click="currentImageWidth = 0">原始大小</button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="closeImageSizeModal">取消</button>
          <button type="button" class="btn btn-primary" @click="applyImageSize">应用</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import mMessage from '@/common/message'
import MarkdownIt from 'markdown-it'
import katex from 'katex'
import '@iktakahiro/markdown-it-katex'

// 配置 markdown-it，支持 LaTeX 数学公式
const md = new MarkdownIt({
  html: true,          // 启用 HTML 标签（用于调整图片大小）
  linkify: true,       // 自动转换 URL
  typographer: true,   // 启用一些语言中立的替换和引号美化
})

// 使用 katex 插件
md.use(require('@iktakahiro/markdown-it-katex'), {
  throwOnError: false,
  errorColor: '#cc0000'
})

// 渲染 Markdown（支持 LaTeX 数学公式）
function renderMarkdown(text) {
  if (!text) return ''

  try {
    const html = md.render(text)
    return html
  } catch (error) {
    console.error('Markdown 渲染错误:', error)
    return `<p style="color: #f56c6c;">渲染错误: ${error.message}</p>`
  }
}

export default {
  name: 'CompetitionEditor',
  data() {
    return {
      isEdit: false,
      competitionId: null,
      submitting: false,
      pendingImageWidth: '', // 待插入的图片宽度
      showImageSizeModal: false, // 图片大小调整对话框
      currentImageUrl: '', // 当前编辑的图片URL
      currentImageWidth: 500, // 当前图片宽度
      currentImageStartPos: -1, // 当前图片在markdown中的起始位置
      currentImageEndPos: -1, // 当前图片在markdown中的结束位置
      formData: {
        name: '',
        startTime: '',
        endTime: '',
        logoUrl: '',
        description: '',
        fields: {
          name: true,
          class: true,
          college: true,
          studentId: true,
          gender: true,
          shirtSize: true,
          teamName: true,
          qq: true
        },
        visible: true
      },
      fieldOptions: {
        name: '姓名',
        class: '班级',
        college: '学院',
        studentId: '学号',
        gender: '性别',
        shirtSize: '衣服尺码',
        teamName: '队伍名称',
        qq: 'QQ号'
      },
      renderedMarkdown: '',
      previewUpdateTimer: null // 防抖定时器
    }
  },
  watch: {
    'formData.description': {
      handler() {
        // 使用防抖，避免频繁更新预览
        if (this.previewUpdateTimer) {
          clearTimeout(this.previewUpdateTimer)
        }
        this.previewUpdateTimer = setTimeout(() => {
          this.updatePreview()
        }, 300) // 300ms 防抖延迟
      },
      immediate: true
    }
  },
  mounted() {
    // 检查是否是编辑模式
    const competitionId = this.$route.params.id
    if (competitionId && competitionId !== 'create') {
      this.isEdit = true
      this.competitionId = competitionId
      this.loadCompetition(competitionId)
    }

    // 初始化预览
    this.updatePreview()
  },
  methods: {
    async loadCompetition(id) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`/api/registration/competitions/${id}`, {
          headers: {
            'Authorization': token,
            'Url-Type': 'general'
          }
        })

        const result = await response.json()
        if (result.code === 200 || result.status === 200) {
          const comp = result.data || result
          this.formData = {
            name: comp.name || '',
            startTime: this.formatDateTimeLocal(comp.startTime),
            endTime: this.formatDateTimeLocal(comp.endTime),
            logoUrl: comp.logoUrl || '',
            description: comp.description || '',
            fields: comp.fields ? (typeof comp.fields === 'string' ? JSON.parse(comp.fields) : comp.fields) : this.formData.fields,
            visible: comp.visible !== undefined ? comp.visible : true
          }
          // 立即更新预览（不使用防抖）
          if (this.previewUpdateTimer) {
            clearTimeout(this.previewUpdateTimer)
            this.previewUpdateTimer = null
          }
          this.updatePreview()
          mMessage.success('比赛数据加载成功')
        } else {
          mMessage.error(result.message || result.msg || '加载比赛失败')
        }
      } catch (error) {
        console.error('加载比赛失败:', error)
        mMessage.error('加载比赛失败')
      }
    },

    validateForm() {
      // 验证开始时间
      if (!this.formData.startTime) {
        mMessage.error('请选择开始时间')
        return false
      }

      // 验证结束时间
      if (!this.formData.endTime) {
        mMessage.error('请选择结束时间')
        return false
      }

      // 将时间字符串转换为 Date 对象
      const startTime = new Date(this.formData.startTime)
      const endTime = new Date(this.formData.endTime)
      const now = new Date()

      // 验证开始时间不能晚于结束时间
      if (startTime >= endTime) {
        mMessage.error('开始时间必须早于结束时间')
        return false
      }

      // 验证结束时间不能早于当前时间（可选）
      if (endTime < now) {
        mMessage.error('结束时间不能早于当前时间')
        return false
      }

      // 验证比赛时长至少 10 分钟
      const duration = endTime - startTime
      const minDuration = 10 * 60 * 1000 // 10分钟
      if (duration < minDuration) {
        mMessage.error('比赛时长至少需要 10 分钟')
        return false
      }

      // 验证比赛名称
      if (!this.formData.name || this.formData.name.trim() === '') {
        mMessage.error('请输入比赛名称')
        return false
      }

      return true
    },

    async saveCompetition() {
      // 验证表单
      if (!this.validateForm()) {
        return
      }

      this.submitting = true
      try {
        const token = localStorage.getItem('token')

        // 准备数据
        const data = {
          name: this.formData.name,
          startTime: new Date(this.formData.startTime).toISOString(),
          endTime: new Date(this.formData.endTime).toISOString(),
          logoUrl: this.formData.logoUrl,
          description: this.formData.description,
          fields: JSON.stringify(this.formData.fields),
          visible: this.formData.visible
        }

        let response
        if (this.isEdit) {
          // 更新比赛
          response = await fetch(`/api/registration/admin/competitions/${this.competitionId}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': token,
              'Url-Type': 'general'
            },
            body: JSON.stringify(data)
          })
        } else {
          // 创建比赛
          response = await fetch('/api/registration/admin/competitions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': token,
              'Url-Type': 'general'
            },
            body: JSON.stringify(data)
          })
        }

        const result = await response.json()
        if (result.code === 200) {
          mMessage.success(this.isEdit ? '比赛更新成功' : '比赛创建成功')
          this.$router.push('/admin/registration')
        } else {
          mMessage.error(result.message || '操作失败')
        }
      } catch (error) {
        console.error('保存比赛失败:', error)
        mMessage.error('保存失败，请重试')
      } finally {
        this.submitting = false
      }
    },

    triggerFileInput() {
      this.$refs.fileInput.click()
    },

    async handleLogoUpload(event) {
      const file = event.target.files[0]
      if (!file) return

      // 验证文件类型
      const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/gif']
      if (!allowedTypes.includes(file.type)) {
        mMessage.error('只支持 PNG、JPG、GIF 格式的图片')
        return
      }

      // 验证文件大小（5MB）
      if (file.size > 5 * 1024 * 1024) {
        mMessage.error('图片大小不能超过 5MB')
        return
      }

      // 上传文件
      try {
        const formData = new FormData()
        formData.append('image', file)

        const token = localStorage.getItem('token')
        const response = await fetch('/api/registration/admin/upload/logo', {
          method: 'POST',
          headers: {
            'Authorization': token,
            'Url-Type': 'general'
          },
          body: formData
        })

        const result = await response.json()
        if (result.status === 200) {
          this.formData.logoUrl = result.data.url
          mMessage.success('Logo上传成功')
        } else {
          mMessage.error(result.message || '上传失败')
        }
      } catch (error) {
        console.error('上传失败:', error)
        mMessage.error('上传失败，请重试')
      }

      // 清空input，允许重复上传同一文件
      event.target.value = ''
    },

    removeLogo() {
      this.formData.logoUrl = ''
    },

    insertMarkdown(before, after) {
      const textarea = this.$refs.textarea || document.querySelector('.markdown-editor')
      if (!textarea) return

      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const text = this.formData.description
      const selectedText = text.substring(start, end)

      const newText = text.substring(0, start) + before + selectedText + after + text.substring(end)
      this.formData.description = newText

      // 重新设置光标位置
      this.$nextTick(() => {
        textarea.focus()
        textarea.setSelectionRange(start + before.length, start + before.length + selectedText.length)
      })

      this.updatePreview()
    },

    triggerMarkdownImageUpload() {
      this.pendingImageWidth = ''
      this.$refs.markdownFileInput.click()
    },

    async handleMarkdownImageUpload(event) {
      const file = event.target.files[0]
      if (!file) return

      // 验证文件
      const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/gif']
      if (!allowedTypes.includes(file.type)) {
        mMessage.error('只支持 PNG、JPG、GIF 格式的图片')
        return
      }

      if (file.size > 5 * 1024 * 1024) {
        mMessage.error('图片大小不能超过 5MB')
        return
      }

      // 上传
      try {
        const formData = new FormData()
        formData.append('image', file)

        const token = localStorage.getItem('token')
        const response = await fetch('/api/registration/admin/upload/logo', {
          method: 'POST',
          headers: {
            'Authorization': token,
            'Url-Type': 'general'
          },
          body: formData
        })

        const result = await response.json()
        if (result.code === 200) {
          // 使用完整的图片URL以确保在markdown中正确显示
          const imageUrl = result.data.url.startsWith('http')
            ? result.data.url
            : `${window.location.origin}${result.data.url}`

          // 根据选择的宽度插入不同的格式
          if (this.pendingImageWidth) {
            // 使用 HTML 标签指定宽度
            this.insertMarkdown(`<img src="${imageUrl}" width="${this.pendingImageWidth}" alt="图片描述">`, '')
          } else {
            // 使用标准 Markdown 语法（原始大小）
            this.insertMarkdown(`![图片描述](${imageUrl})`, '')
          }

          mMessage.success('图片上传成功')
        } else {
          mMessage.error(result.message || '上传失败')
        }
      } catch (error) {
        console.error('上传失败:', error)
        mMessage.error('上传失败，请重试')
      }

      event.target.value = ''
      this.pendingImageWidth = '' // 重置宽度
    },

    updatePreview() {
      try {
        const content = this.formData.description || ''

        if (!content.trim()) {
          const emptyHTML = '<p style="color: #999; font-style: italic;">请在左侧输入比赛说明，支持Markdown格式...</p>'
          // 只在内容真正改变时才更新
          if (this.renderedMarkdown !== emptyHTML) {
            this.renderedMarkdown = emptyHTML
          }
          return
        }

        // 使用自定义 markdown 渲染器
        const newHTML = renderMarkdown(content)
        // 只在内容真正改变时才更新 DOM，避免图片重新加载
        if (this.renderedMarkdown !== newHTML) {
          this.renderedMarkdown = newHTML
        }
      } catch (error) {
        console.error('Markdown 渲染错误:', error)
        const errorHTML = `<p style="color: #f56c6c;">Markdown渲染错误: ${error.message}</p>`
        if (this.renderedMarkdown !== errorHTML) {
          this.renderedMarkdown = errorHTML
        }
      }
    },

    formatDateTimeLocal(isoString) {
      if (!isoString) return ''
      const date = new Date(isoString)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day}T${hours}:${minutes}`
    },

    formatTime(str) {
      if (!str) return ''
      const date = new Date(str)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
        timeZone: 'Asia/Shanghai'
      })
    },

    // 处理预览区域点击事件
    handlePreviewClick(event) {
      // 检查是否点击了图片
      if (event.target.tagName === 'IMG') {
        const img = event.target
        const imageUrl = img.src

        // 在 markdown 中查找对应的图片
        this.findImageInMarkdown(imageUrl)
      }
    },

    // 在 markdown 中查找图片
    findImageInMarkdown(imageUrl) {
      const desc = this.formData.description || ''

      // 提取图片路径（去除域名和协议）
      let imagePath = imageUrl
      if (imageUrl.startsWith(window.location.origin)) {
        imagePath = imageUrl.substring(window.location.origin.length)
      }

      // 查找标准 Markdown 图片语法: ![alt](url)
      const mdImageRegex = new RegExp(`!\\[[^\\]]*\\]\\(([^)]*${this.escapeRegex(imagePath)}[^)]*)\\)`, 'g')
      const mdMatch = mdImageRegex.exec(desc)

      // 查找 HTML img 标签
      const htmlImgRegex = new RegExp(`<img[^>]*src=["']([^"']*${this.escapeRegex(imagePath)}[^"']*)[^>]*>`, 'gi')
      const htmlMatch = htmlImgRegex.exec(desc)

      if (mdMatch) {
        // 找到 Markdown 格式的图片
        this.currentImageUrl = imageUrl
        this.currentImageWidth = 0 // 原始大小
        this.currentImageStartPos = mdMatch.index
        this.currentImageEndPos = mdMatch.index + mdMatch[0].length
        this.currentImageType = 'markdown'
        this.showImageSizeModal = true
      } else if (htmlMatch) {
        // 找到 HTML 格式的图片
        const widthMatch = htmlMatch[0].match(/width=["'](\d+)["']/)
        this.currentImageUrl = imageUrl
        this.currentImageWidth = widthMatch ? parseInt(widthMatch[1]) : 0
        this.currentImageStartPos = htmlMatch.index
        this.currentImageEndPos = htmlMatch.index + htmlMatch[0].length
        this.currentImageType = 'html'
        this.showImageSizeModal = true
      } else {
        mMessage.warning('无法找到对应的图片代码')
      }
    },

    // 转义正则表达式特殊字符
    escapeRegex(string) {
      return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    },

    // 关闭图片大小调整对话框
    closeImageSizeModal() {
      this.showImageSizeModal = false
      this.currentImageUrl = ''
      this.currentImageWidth = 500
      this.currentImageStartPos = -1
      this.currentImageEndPos = -1
    },

    // 应用图片大小
    applyImageSize() {
      if (this.currentImageStartPos === -1 || this.currentImageEndPos === -1) {
        mMessage.error('无法找到图片位置')
        return
      }

      const desc = this.formData.description || ''
      const before = desc.substring(0, this.currentImageStartPos)
      const after = desc.substring(this.currentImageEndPos)

      let newImageCode = ''

      if (this.currentImageWidth > 0) {
        // 使用 HTML 标签指定宽度
        newImageCode = `<img src="${this.currentImageUrl}" width="${this.currentImageWidth}" alt="图片描述">`
      } else {
        // 使用标准 Markdown 语法（原始大小）
        newImageCode = `![图片描述](${this.currentImageUrl})`
      }

      this.formData.description = before + newImageCode + after
      this.closeImageSizeModal()
      // 立即更新预览（不使用防抖）
      if (this.previewUpdateTimer) {
        clearTimeout(this.previewUpdateTimer)
        this.previewUpdateTimer = null
      }
      this.updatePreview()
      mMessage.success('图片大小已更新')
    }
  },
  beforeDestroy() {
    // 清理防抖定时器
    if (this.previewUpdateTimer) {
      clearTimeout(this.previewUpdateTimer)
      this.previewUpdateTimer = null
    }
  }
}
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.competition-editor-page {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f7fa;
  min-height: 100vh;
  padding: 20px;
}

.container {
  max-width: 1600px;
  margin: 0 auto;
}

/* 头部 */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #409EFF;
  margin-bottom: 6px;
}

.header p {
  color: #718096;
  font-size: 14px;
}

/* ===== ① 顶部基本信息区 ===== */
.basic-info-section {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.section-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #409EFF;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-group input[type="text"],
.form-group input[type="datetime-local"] {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.form-group input:focus {
  outline: none;
  border-color: #409EFF;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.form-row-inline {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-group-inline {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
  align-items: start;
}

/* Logo 上传 */
.logo-section {
  display: flex;
  flex-direction: column;
}

.logo-upload {
  margin-top: 8px;
}

.logo-preview {
  position: relative;
  width: 200px;
  height: 200px;
  border-radius: 8px;
  overflow: hidden;
  border: 2px dashed #e2e8f0;
  background: #f7fafc;
}

.logo-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.remove-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.95);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.2s;
}

.remove-btn:hover {
  background: #fff;
  transform: scale(1.05);
}

.upload-placeholder {
  width: 200px;
  height: 200px;
  border: 2px dashed #cbd5e0;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #f7fafc;
}

.upload-placeholder:hover {
  border-color: #409EFF;
  background: #edf4fc;
}

.upload-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.upload-hint {
  font-size: 12px;
  color: #718096;
  margin-top: 4px;
}

/* 可见性 */
.visibility-section {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  height: 100%;
}

.checkbox-item-inline {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e2e8f0;
  margin-top: 28px;
}

.checkbox-item-inline:hover {
  background: #edf4fc;
  border-color: #409EFF;
}

.checkbox-item-inline input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-item-inline span {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

/* ===== ② 编辑器和预览对称区 ===== */
.editor-preview-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-bottom: 24px;
  align-items: stretch;
}

/* 编辑器和预览的公共样式 */
.editor-section,
.preview-section {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 16px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
}

.preview-hint {
  font-size: 12px;
  color: #718096;
  margin: 0;
  font-style: italic;
}

/* 编辑器区域 */
.editor-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 500px;
}

.editor-toolbar {
  padding: 10px;
  background: #f7fafc;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.toolbar-btn {
  padding: 6px 12px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.toolbar-btn:hover {
  background: #409EFF;
  color: white;
  border-color: #409EFF;
}

.markdown-editor {
  flex: 1;
  width: 100%;
  padding: 16px;
  border: none;
  resize: none;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.7;
  color: #2d3748;
  background: white;
  min-height: 500px;
}

.markdown-editor:focus {
  outline: none;
}

.markdown-editor::placeholder {
  color: #a0aec0;
}

/* 预览区域 */
.preview-wrapper {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  min-height: 500px;
  background: white;
}

.preview-content {
  font-size: 14px;
  line-height: 1.7;
  color: #2d3748;
}

/* ===== ③ 底部字段配置区 ===== */
.fields-config-section {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.checkbox-item:hover {
  background: #edf4fc;
  border-color: #409EFF;
}

.checkbox-item input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-item span {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

/* 操作按钮 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.btn {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
  text-align: center;
  transition: all 0.3s;
  background: #e2e8f0;
  color: #4a5568;
}

.btn:hover {
  background: #cbd5e0;
  transform: translateY(-1px);
}

.btn-primary {
  background: #409EFF;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #66b1ff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* ===== Markdown 样式 ===== */
.markdown-body h1,
.markdown-body h2,
.markdown-body h3 {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  color: #2d3748;
}

.markdown-body h1 {
  font-size: 24px;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.markdown-body h2 {
  font-size: 20px;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 6px;
}

.markdown-body h3 {
  font-size: 18px;
}

.markdown-body p {
  margin-bottom: 12px;
}

.markdown-body ul,
.markdown-body ol {
  padding-left: 24px;
  margin-bottom: 12px;
}

.markdown-body li {
  margin-bottom: 6px;
}

.markdown-body code {
  background: #f7fafc;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: #e53e3e;
}

.markdown-body pre {
  background: #2d3748;
  color: #f7fafc;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 16px 0;
}

.markdown-body pre code {
  background: transparent;
  padding: 0;
  color: #f7fafc;
}

.markdown-body img {
  max-width: 100%;
  height: auto;
  margin: 12px 0;
  cursor: pointer;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.markdown-body img:hover {
  transform: none;
}

.markdown-body img[width] {
  max-width: 100%;
  width: auto !important;
}

.markdown-body blockquote {
  border-left: 4px solid #409EFF;
  padding-left: 16px;
  margin: 12px 0;
  color: #718096;
  font-style: italic;
}

.markdown-body a {
  color: #409EFF;
  text-decoration: none;
}

.markdown-body a:hover {
  text-decoration: underline;
}

.markdown-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
  overflow-x: auto;
}

.markdown-body th,
.markdown-body td {
  border: 1px solid #e2e8f0;
  padding: 8px 12px;
  text-align: left;
}

.markdown-body th {
  background: #f7fafc;
  font-weight: 600;
}

/* ===== 响应式设计 ===== */

/* 平板横屏：仍然并排，但调整间距 */
@media (max-width: 1200px) {
  .container {
    max-width: 100%;
  }

  .form-group-inline {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .visibility-section {
    margin-top: 0;
  }

  .checkbox-item-inline {
    margin-top: 0;
  }
}

/* 平板竖屏：上下排列 */
@media (max-width: 992px) {
  .editor-preview-container {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .editor-section {
    order: 1;
  }

  .preview-section {
    order: 2;
  }

  .form-row-inline {
    grid-template-columns: 1fr;
  }

  .checkbox-group {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  }
}

/* 手机端：进一步优化 */
@media (max-width: 768px) {
  .competition-editor-page {
    padding: 12px;
  }

  .header {
    padding: 20px;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header h1 {
    font-size: 24px;
  }

  .basic-info-section,
  .fields-config-section {
    padding: 20px;
  }

  .checkbox-group {
    grid-template-columns: 1fr;
  }

  .editor-wrapper,
  .preview-wrapper {
    min-height: 400px;
  }

  .markdown-editor {
    min-height: 400px;
    font-size: 13px;
  }

  .form-actions {
    flex-direction: column-reverse;
  }

  .form-actions .btn {
    width: 100%;
  }
}

/* 小屏手机 */
@media (max-width: 480px) {
  .header h1 {
    font-size: 20px;
  }

  .upload-placeholder {
    width: 100%;
    height: 160px;
  }

  .logo-preview {
    width: 100%;
    height: 160px;
  }

  .panel-header {
    padding: 12px 16px;
  }

  .panel-header h3 {
    font-size: 14px;
  }

  .preview-wrapper {
    padding: 16px;
  }

  .markdown-editor {
    padding: 12px;
  }
}
</style>

<!-- 非 scoped 样式，用于 v-html 渲染的 markdown 内容 -->
<style>
.markdown-body {
  line-height: 1.8;
  color: #2d3748;
  word-wrap: break-word;
}

.markdown-body > *:first-child {
  margin-top: 0 !important;
}

.markdown-body > *:last-child {
  margin-bottom: 0 !important;
}

.markdown-body p {
  margin-bottom: 12px;
}

.markdown-body h1,
.markdown-body h2,
.markdown-body h3,
.markdown-body h4,
.markdown-body h5,
.markdown-body h6 {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-body h1 {
  font-size: 28px;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.markdown-body h2 {
  font-size: 24px;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 6px;
}

.markdown-body h3 {
  font-size: 20px;
}

.markdown-body h4 {
  font-size: 18px;
}

.markdown-body h5 {
  font-size: 16px;
}

.markdown-body h6 {
  font-size: 14px;
  color: #718096;
}

.markdown-body ul,
.markdown-body ol {
  padding-left: 24px;
  margin-bottom: 12px;
}

.markdown-body li {
  margin-bottom: 4px;
}

.markdown-body li > p {
  margin-top: 8px;
}

.markdown-body code {
  background: #f7fafc;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 13px;
  color: #e53e3e;
}

.markdown-body pre {
  background: #2d3748;
  color: #f7fafc;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 16px 0;
}

.markdown-body pre code {
  background: transparent;
  padding: 0;
  color: #f7fafc;
  font-size: 13px;
}

.markdown-body blockquote {
  border-left: 4px solid #409EFF;
  padding-left: 16px;
  margin: 16px 0;
  color: #718096;
  font-style: italic;
}

.markdown-body blockquote > *:first-child {
  margin-top: 0;
}

.markdown-body blockquote > *:last-child {
  margin-bottom: 0;
}

.markdown-body strong {
  font-weight: 600;
  color: #1a202c;
}

.markdown-body em {
  font-style: italic;
  color: #4a5568;
}

.markdown-body a {
  color: #409EFF;
  text-decoration: none;
}

.markdown-body a:hover {
  text-decoration: underline;
}

.markdown-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
  display: block;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.markdown-body table thead {
  background: #f7fafc;
}

.markdown-body th,
.markdown-body td {
  border: 1px solid #e2e8f0;
  padding: 8px 12px;
  text-align: left;
}

.markdown-body th {
  background: #f7fafc;
  font-weight: 600;
}

.markdown-body tr:nth-child(even) {
  background: #f7fafc;
}

.markdown-body hr {
  border: none;
  border-top: 2px solid #e2e8f0;
  margin: 24px 0;
}

.markdown-body img {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 12px 0;
  border: 0;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

/* ===== 模态对话框样式 ===== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.2s ease-in-out;
  padding: 20px;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #2d3748;
}

.modal-header .close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  color: #718096;
}

.modal-header .close-btn:hover {
  background: #f7fafc;
  color: #2d3748;
}

.modal-body {
  padding: 20px;
}

.info-notice {
  background: #e6f7ff;
  border-left: 4px solid #409EFF;
  padding: 12px 16px;
  margin-bottom: 20px;
  border-radius: 6px;
  font-size: 13px;
  color: #2c5282;
  line-height: 1.5;
}

.modal-footer {
  padding: 16px 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.image-preview-container {
  text-align: center;
  margin-bottom: 20px;
  padding: 20px;
  background: #f7fafc;
  border-radius: 8px;
  overflow: auto;
}

.image-preview-container img {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: width 0.3s ease;
}

.size-controls {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.control-group label {
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.slider-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.size-slider {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: #e2e8f0;
  outline: none;
  -webkit-appearance: none;
}

.size-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #409EFF;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.size-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #409EFF;
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.size-input {
  width: 80px;
  padding: 6px 8px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  text-align: center;
}

.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-btn {
  padding: 8px 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
  flex: 1;
  min-width: 80px;
}

.preset-btn:hover {
  background: #409EFF;
  color: white;
  border-color: #409EFF;
}

/* 模态框响应式 */
@media (max-width: 768px) {
  .modal-content {
    max-width: 100%;
    margin: 0;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 16px;
  }

  .preset-buttons {
    flex-direction: column;
  }

  .preset-btn {
    min-width: 100%;
  }
}
</style>
