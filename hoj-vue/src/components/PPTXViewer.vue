<template>
  <div class="pptx-viewer" v-loading="loading" element-loading-text="正在解析PPT...">
    <!-- 错误提示 -->
    <div v-if="error" class="error-container">
      <i class="el-icon-warning-outline"></i>
      <p class="error-message">{{ error }}</p>
      <el-button v-if="error.includes('网络')" type="text" @click="retry">重试</el-button>
    </div>

    <!-- 加载进度 -->
    <div v-if="loading && loadProgress > 0" class="loading-progress">
      <el-progress :percentage="loadProgress" :stroke-width="8"></el-progress>
      <p class="progress-text">正在解析文件...</p>
    </div>

    <!-- 幻灯片容器 -->
    <div v-else-if="slides.length > 0" class="slides-container">
      <!-- 当前幻灯片 -->
      <div class="slide-wrapper">
        <div
          class="slide-content"
          :style="getSlideStyle()"
          :class="{ 'slide-animate': isAnimating }"
        >
          <div
            class="slide-content-inner"
            v-html="currentSlideHtml"
          ></div>
        </div>
      </div>

      <!-- 导航控制 -->
      <div class="navigation" v-if="slides.length > 1">
        <el-button
          icon="el-icon-arrow-left"
          :disabled="currentSlide === 0"
          @click="prevSlide"
          circle
          size="small"
        ></el-button>
        <span class="page-number">{{ currentSlide + 1 }} / {{ slides.length }}</span>
        <el-button
          icon="el-icon-arrow-right"
          :disabled="currentSlide === slides.length - 1"
          @click="nextSlide"
          circle
          size="small"
        ></el-button>
        <!-- 幻灯片缩略图导航 -->
        <el-dropdown trigger="click" @command="jumpToSlide">
          <el-button icon="el-icon-menu" circle size="small"></el-button>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item
              v-for="(slide, index) in slides"
              :key="index"
              :command="index"
              :class="{ 'is-active': currentSlide === index }"
            >
              {{ index + 1 }}. {{ getSlideTitle(slide, index) }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
        <!-- 自动播放按钮 -->
        <el-button
          :icon="isPlaying ? 'el-icon-video-pause' : 'el-icon-video-play'"
          @click="toggleAutoPlay"
          circle
          size="small"
        ></el-button>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <i class="el-icon-document"></i>
      <p>正在解析PPT文件...</p>
    </div>
  </div>
</template>

<script>
// 动态加载JSZip
export default {
  name: 'PPTXViewer',
  props: {
    fileUrl: {
      type: String,
      required: true
    },
    fileName: {
      type: String,
      default: 'presentation.pptx'
    }
  },
  data() {
    return {
      loading: true,
      error: null,
      slides: [],
      currentSlide: 0,
      JSZip: null,
      loadProgress: 0,
      fileSizeMB: 0,
      loadTimeout: null,
      aborted: false,
      isAnimating: false,
      isPlaying: false,
      autoPlayTimer: null
    }
  },
  computed: {
    // 当前幻灯片的HTML内容
    currentSlideHtml() {
      if (!this.slides[this.currentSlide]) return ''
      return this.slides[this.currentSlide].html
    }
  },
  mounted() {
    this.init()
  },
  beforeDestroy() {
    this.cleanup()
  },
  watch: {
    fileUrl() {
      this.reset()
      this.init()
    }
  },
  methods: {
    reset() {
      this.loading = true
      this.error = null
      this.slides = []
      this.currentSlide = 0
      this.loadProgress = 0
      this.fileSizeMB = 0
      this.aborted = false
      this.isAnimating = false
      this.stopAutoPlay()
      if (this.loadTimeout) {
        clearTimeout(this.loadTimeout)
        this.loadTimeout = null
      }
    },

    cleanup() {
      this.aborted = true
      this.stopAutoPlay()
      if (this.loadTimeout) {
        clearTimeout(this.loadTimeout)
        this.loadTimeout = null
      }
      this.slides = []
    },

    async init() {
      try {
        await this.loadJSZip()
        await this.loadPPTX()
      } catch (err) {
        console.error('[PPTXViewer] 加载失败，详细错误:', {
          message: err.message,
          stack: err.stack,
          fileUrl: this.fileUrl ? this.fileUrl.substring(0, 100) + '...' : 'empty'
        })
        this.error = this.getErrorMessage(err)
        this.loading = false
        this.$emit('error', this.error)
      }
    },

    getErrorMessage(err) {
      const errorMsg = err.message || ''

      // HTTP状态码错误
      if (errorMsg.includes('文件下载失败: 403')) {
        return '没有权限访问此文件，请联系教师'
      }
      if (errorMsg.includes('文件下载失败: 401')) {
        return '登录已过期，请重新登录'
      }
      if (errorMsg.includes('文件下载失败: 404')) {
        return '文件不存在'
      }
      if (errorMsg.includes('文件下载失败:')) {
        // 提取状态码
        const statusMatch = errorMsg.match(/文件下载失败: (\d+)/)
        const status = statusMatch ? statusMatch[1] : '未知'
        return `服务器错误 (${status})，请稍后重试`
      }

      // Mixed Content 错误（特殊处理）
      if (errorMsg === 'Failed to fetch' && this.fileUrl && this.fileUrl.startsWith('http://')) {
        return '文件地址使用了不安全的HTTP协议，无法在HTTPS页面中加载'
      }

      // 网络错误
      const errorMap = {
        'Failed to fetch': '文件加载失败，请检查网络连接',
        'NetworkError': '网络错误，请检查网络连接',
        'timeout': '加载超时，请稍后重试',
        'not a pptx': '不支持的文件格式',
        'corrupted': '文件已损坏，无法预览'
      }

      for (const [key, message] of Object.entries(errorMap)) {
        if (errorMsg.includes(key)) {
          return message
        }
      }

      // 未知错误
      console.warn('[PPTXViewer] 未知错误类型:', errorMsg)
      return '加载失败，请下载后查看'
    },

    retry() {
      this.reset()
      this.init()
    },

    loadJSZip() {
      return new Promise((resolve, reject) => {
        if (window.JSZip) {
          this.JSZip = window.JSZip
          resolve()
          return
        }

        const script = document.createElement('script')
        script.src = '/lib/jszip.min.js'
        script.onload = () => {
          if (window.JSZip) {
            this.JSZip = window.JSZip
            resolve()
          } else {
            reject(new Error('JSZip加载失败'))
          }
        }
        script.onerror = () => reject(new Error('JSZip加载失败'))
        document.head.appendChild(script)
      })
    },

    async loadPPTX() {
      this.loadTimeout = setTimeout(() => {
        this.aborted = true
        throw new Error('timeout')
      }, 30000)

      try {
        this.loadProgress = 10

        // 修复 Mixed Content 错误：自动将 HTTP 转换为 HTTPS
        let fetchUrl = this.fileUrl
        if (fetchUrl.startsWith('http://') && window.location.protocol === 'https:') {
          fetchUrl = fetchUrl.replace('http://', 'https://')
          console.log('[PPTXViewer] 检测到Mixed Content问题，自动将HTTP转换为HTTPS')
        }

        const headers = {}
        const token = localStorage.getItem('token')
        if (token) {
          // 尝试多种token格式
          headers['Authorization'] = token.startsWith('Bearer ') ? token : `Bearer ${token}`
          console.log('[PPTXViewer] 使用token认证，token前缀:', token.substring(0, 20) + '...')
        } else {
          console.warn('[PPTXViewer] 未找到token，尝试匿名访问')
        }

        console.log('[PPTXViewer] 开始下载文件:', {
          url: fetchUrl.substring(0, 80) + '...',
          originalUrl: this.fileUrl.substring(0, 80) + '...',
          hasToken: !!token,
          headers: Object.keys(headers)
        })

        const response = await fetch(fetchUrl, { headers })
        if (!response.ok) {
          console.error('[PPTXViewer] 文件下载失败:', {
            status: response.status,
            statusText: response.statusText,
            url: fetchUrl.substring(0, 100) + '...'
          })
          throw new Error(`文件下载失败: ${response.status}`)
        }

        const contentLength = response.headers.get('content-length')
        if (contentLength) {
          this.fileSizeMB = (contentLength / (1024 * 1024)).toFixed(1)
        }

        this.loadProgress = 40

        const arrayBuffer = await response.arrayBuffer()

        if (!this.isPPTXFile(arrayBuffer)) {
          throw new Error('not a pptx file')
        }

        this.loadProgress = 60

        const zip = await this.JSZip.loadAsync(arrayBuffer)

        this.loadProgress = 80

        await this.parsePresentation(zip)

        this.loadProgress = 100

        if (this.aborted) {
          throw new Error('加载已取消')
        }

        this.loading = false
      } catch (err) {
        console.error('PPTX解析失败:', err)
        throw err
      } finally {
        if (this.loadTimeout) {
          clearTimeout(this.loadTimeout)
          this.loadTimeout = null
        }
      }
    },

    isPPTXFile(arrayBuffer) {
      const header = new Uint8Array(arrayBuffer.slice(0, 4))
      return header[0] === 0x50 && header[1] === 0x4B &&
             (header[2] === 0x03 || header[2] === 0x05)
    },

    async parsePresentation(zip) {
      try {
        const contentTypeFile = zip.file('[Content_Types].xml')
        if (!contentTypeFile) {
          throw new Error('corrupted')
        }

        const contentType = await contentTypeFile.async('string')

        const slideMatches = contentType.match(/\/ppt\/slides\/slide\d+\.xml/g) || []
        const slideNumbers = slideMatches.map(match => {
          const num = match.match(/slide(\d+)\.xml/)[1]
          return parseInt(num)
        }).sort((a, b) => a - b)

        if (slideNumbers.length === 0) {
          throw new Error('PPT文件中没有找到幻灯片')
        }

        // 提取所有图片
        const imageMap = await this.extractImages(zip)

        const slides = []
        for (const slideNum of slideNumbers) {
          if (this.aborted) {
            throw new Error('加载已取消')
          }

          const slidePath = `ppt/slides/slide${slideNum}.xml`
          const slideFile = zip.file(slidePath)

          if (!slideFile) {
            console.warn(`幻灯片 ${slideNum} 不存在，跳过`)
            continue
          }

          // 读取关系文件以获取图片引用
          const relsPath = `ppt/slides/_rels/slide${slideNum}.xml.rels`
          const relsFile = zip.file(relsPath)
          let slideRels = {}
          if (relsFile) {
            const relsXml = await relsFile.async('string')
            slideRels = this.parseRelationships(relsXml)
          }

          const slideXml = await slideFile.async('string')
          const html = this.convertSlideToHTML(slideXml, slideRels, imageMap)
          const title = this.extractSlideTitle(slideXml)
          slides.push({ html, title, file: slidePath })
        }

        if (slides.length === 0) {
          throw new Error('无法解析PPT幻灯片')
        }

        this.slides = slides
      } catch (err) {
        console.error('演示文稿解析失败:', err)
        throw err
      }
    },

    // 提取所有图片
    async extractImages(zip) {
      const imageMap = {}

      // 查找所有媒体文件
      const mediaFolder = zip.folder('ppt/media')
      if (mediaFolder) {
        const files = Object.keys(mediaFolder.files)
        for (const filename of files) {
          if (mediaFolder.files[filename].dir) continue

          try {
            const file = mediaFolder.files[filename]
            const extension = filename.split('.').pop().toLowerCase()
            const mimeType = this.getImageMimeType(extension)

            if (mimeType) {
              const data = await file.async('base64')
              imageMap[filename] = `data:${mimeType};base64,${data}`
            }
          } catch (err) {
            console.warn('图片提取失败:', filename, err)
          }
        }
      }

      return imageMap
    },

    // 获取图片MIME类型
    getImageMimeType(extension) {
      const mimeTypes = {
        'png': 'image/png',
        'jpg': 'image/jpeg',
        'jpeg': 'image/jpeg',
        'gif': 'image/gif',
        'bmp': 'image/bmp',
        'svg': 'image/svg+xml'
      }
      return mimeTypes[extension] || null
    },

    // 解析关系文件
    parseRelationships(relsXml) {
      const rels = {}
      const matches = relsXml.match(/<Relationship[^>]*>/g) || []

      matches.forEach(match => {
        const idMatch = match.match(/Id="([^"]+)"/)
        const targetMatch = match.match(/Target="([^"]+)"/)

        if (idMatch && targetMatch) {
          rels[idMatch[1]] = targetMatch[1]
        }
      })

      return rels
    },

    extractSlideTitle(slideXml) {
      const titleMatch = slideXml.match(/<a:t[^>]*>([^<]+)<\/a:t>/)
      if (titleMatch) {
        const title = titleMatch[1].trim()
        return title.substring(0, 30) + (title.length > 30 ? '...' : '')
      }
      return '幻灯片'
    },

    getSlideTitle(slide, index) {
      return slide.title || `幻灯片 ${index + 1}`
    },

    convertSlideToHTML(slideXml, slideRels, imageMap) {
      let html = '<div class="slide-content-inner">'

      try {
        // 提取图片
        const picMatches = slideXml.match(/<pic:pic[^>]*>.*?<\/pic:pic>/gs) || []

        picMatches.forEach(picXml => {
          const blipMatch = picXml.match(/<a:blip[^>]*r:embed="([^"]+)"/)
          if (blipMatch) {
            const relId = blipMatch[1]
            const imageTarget = slideRels[relId]

            if (imageTarget) {
              // 提取图片文件名
              const imageName = imageTarget.split('/').pop()
              const imageUrl = imageMap[imageName]

              if (imageUrl) {
                // 提取图片位置和大小
                const offMatch = picXml.match(/<a:off[^>]*x="(\d+)"[^>]*y="(\d+)"/)
                const extMatch = picXml.match(/<a:ext[^>]*cx="(\d+)"[^>]*cy="(\d+)"/)

                let style = ''
                if (offMatch && extMatch) {
                  const x = Math.round(parseInt(offMatch[1]) / 9525)
                  const y = Math.round(parseInt(offMatch[2]) / 9525)
                  const w = Math.round(parseInt(extMatch[1]) / 9525)
                  const h = Math.round(parseInt(extMatch[2]) / 9525)
                  style = `position: absolute; left: ${x}px; top: ${y}px; width: ${w}px; height: ${h}px;`
                }

                html += `<img src="${imageUrl}" class="slide-image" style="${style}" alt="slide image" />`
              }
            }
          }
        })

        // 提取段落文本
        const pMatches = slideXml.match(/<a:p[^>]*>.*?<\/a:p>/gs) || []

        if (pMatches.length > 0) {
          html += '<div class="slide-text-container">'

          pMatches.forEach((paragraph, idx) => {
            // 检查段落是否包含图片
            if (paragraph.includes('<a:blip')) return

            const textInParagraph = paragraph.match(/<a:t[^>]*>([^<]+)<\/a:t>/g) || []
            const textContent = textInParagraph
              .map(t => t.replace(/<a:t[^>]*>|<\/a:t>/g, ''))
              .join('')

            if (textContent.trim()) {
              const styleClass = this.getParagraphStyleClass(idx, pMatches.length)

              // 提取文本样式
              const paragraphStyle = this.extractParagraphStyle(paragraph)

              html += `<p class="${styleClass}" style="${paragraphStyle}">${textContent}</p>`
            }
          })

          html += '</div>'
        }

        html += '</div>'
      } catch (err) {
        console.error('转换幻灯片失败:', err)
        html += '<div class="slide-text-container"><p class="slide-paragraph-error">解析失败</p></div></div>'
      }

      return html
    },

    // 提取段落样式
    extractParagraphStyle(paragraph) {
      const styles = []

      // 提取字体颜色
      const colorMatch = paragraph.match(/<a:solidFill>.*?<a:srgbClr[^>]*val="([A-F0-9]+)"/)
      if (colorMatch) {
        styles.push(`color: #${colorMatch[1]}`)
      }

      // 提取字体大小
      const sizeMatch = paragraph.match(/<a:latin[^>]*typeface="([^"]+)"/)
      if (sizeMatch) {
        styles.push(`font-family: "${sizeMatch[1]}", sans-serif`)
      }

      return styles.join('; ')
    },

    getParagraphStyleClass(index, total) {
      if (index === 0 && total > 1) {
        return 'slide-paragraph slide-paragraph-title'
      }
      if (index === total - 1 && total > 3) {
        return 'slide-paragraph slide-paragraph-small'
      }
      return 'slide-paragraph slide-paragraph-body'
    },

    getSlideStyle() {
      return {
        width: '100%',
        height: '100%',
        backgroundColor: '#ffffff',
        position: 'relative',
        overflow: 'hidden'
      }
    },

    prevSlide() {
      if (this.currentSlide > 0) {
        this.triggerAnimation(() => {
          this.currentSlide--
        })
      }
    },

    nextSlide() {
      if (this.currentSlide < this.slides.length - 1) {
        this.triggerAnimation(() => {
          this.currentSlide++
        })
      } else if (this.isPlaying) {
        // 自动播放时循环到第一页
        this.triggerAnimation(() => {
          this.currentSlide = 0
        })
      }
    },

    jumpToSlide(index) {
      if (index >= 0 && index < this.slides.length) {
        this.triggerAnimation(() => {
          this.currentSlide = index
        })
      }
    },

    // 触发过渡动画
    triggerAnimation(callback) {
      this.isAnimating = true
      setTimeout(() => {
        callback()
        setTimeout(() => {
          this.isAnimating = false
        }, 50)
      }, 300)
    },

    // 自动播放
    toggleAutoPlay() {
      if (this.isPlaying) {
        this.stopAutoPlay()
      } else {
        this.startAutoPlay()
      }
    },

    startAutoPlay() {
      this.isPlaying = true
      this.autoPlayTimer = setInterval(() => {
        if (this.currentSlide < this.slides.length - 1) {
          this.nextSlide()
        } else {
          this.currentSlide = 0
        }
      }, 3000) // 每3秒切换一页
    },

    stopAutoPlay() {
      this.isPlaying = false
      if (this.autoPlayTimer) {
        clearInterval(this.autoPlayTimer)
        this.autoPlayTimer = null
      }
    }
  }
}
</script>

<style scoped>
.pptx-viewer {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
  position: relative;
}

.error-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #f56c6c;
  padding: 20px;
  text-align: center;
}

.error-container i {
  font-size: 48px;
  margin-bottom: 16px;
}

.error-message {
  margin: 8px 0;
  font-size: 14px;
}

.loading-progress {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  padding: 40px;
}

.progress-text {
  margin-top: 16px;
  font-size: 14px;
  color: #606266;
}

.slides-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.slide-wrapper {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  overflow: auto;
}

.slide-content {
  background: white;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  max-width: 100%;
  max-height: 100%;
  position: relative;
  border-radius: 4px;
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.slide-content.slide-animate {
  opacity: 0.5;
  transform: scale(0.98);
}

.slide-content-inner {
  padding: 40px;
  min-width: 800px;
  min-height: 600px;
  display: flex;
  flex-direction: column;
  position: relative;
}

.slide-text-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
  z-index: 1;
}

.slide-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.slide-paragraph {
  margin: 0;
  padding: 8px 16px;
  line-height: 1.6;
  transition: all 0.3s ease;
}

.slide-paragraph-title {
  font-size: 36px;
  font-weight: bold;
  text-align: center;
  margin-bottom: 32px;
  color: #303133;
  padding: 20px 16px;
}

.slide-paragraph-body {
  font-size: 22px;
  color: #606266;
  padding: 12px 16px;
}

.slide-paragraph-small {
  font-size: 18px;
  color: #909399;
  padding: 8px 16px;
}

.slide-paragraph-empty {
  font-size: 18px;
  color: #c0c4cc;
  text-align: center;
  font-style: italic;
}

.slide-paragraph-error {
  font-size: 18px;
  color: #f56c6c;
  text-align: center;
}

.navigation {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.95);
  border-top: 1px solid #e4e7ed;
  backdrop-filter: blur(10px);
}

.page-number {
  font-size: 14px;
  color: #606266;
  min-width: 80px;
  text-align: center;
  font-weight: 500;
}

.empty-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #909399;
}

.empty-state i {
  font-size: 64px;
  margin-bottom: 16px;
  color: #c0c4cc;
}

.is-active {
  color: #409eff;
  font-weight: bold;
}
</style>
