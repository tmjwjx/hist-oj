<template>
  <div class="student-materials classroom-theme">
    <div class="header">
      <h3>{{ $t('m.Material_Library') }}</h3>
    </div>

    <!-- 左右分屏布局 -->
    <div class="content-layout-split">
      <!-- 左侧：文件列表面板 -->
      <div class="file-list-panel">
        <!-- 树形导航 -->
        <div class="tree-sidebar">
          <div class="tree-header">
            <span class="tree-title">文件夹</span>
          </div>
          <el-tree
            ref="folderTree"
            :data="folderTree"
            :props="treeProps"
            :highlight-current="true"
            node-key="id"
            :current-node-key="currentFolderId"
            :expand-on-click-node="false"
            :default-expand-all="true"
            @node-click="handleNodeClick"
            class="folder-tree"
          >
            <span class="custom-tree-node" slot-scope="{ node, data }">
              <span class="node-label">
                <i :class="data.id === 0 ? 'el-icon-folder' : 'el-icon-folder-opened'"></i>
                {{ node.label }}
              </span>
            </span>
          </el-tree>
        </div>

        <!-- 文件列表 -->
        <div class="files-list">
          <div v-loading="loading">
            <!-- 文件列表 -->
            <div v-if="materials.length > 0" class="section">
              <div
                v-for="material in materials"
                :key="'material-' + material.id"
                class="file-item material-item"
                :class="{ active: selectedMaterial && selectedMaterial.id === material.id }"
                @click="selectMaterial(material)"
              >
                <i :class="getFileIcon(getFileType(material.fileName))"></i>
                <div class="file-info">
                  <span class="file-name" :title="material.fileName">{{ material.fileName }}</span>
                  <span class="file-size">{{ formatFileSize(material.fileSize) }}</span>
                </div>
                <div class="file-actions">
                  <el-button
                    v-if="canPreview(getFileType(material.fileName)) && hasPreviewPermission(material)"
                    size="mini"
                    type="text"
                    icon="el-icon-view"
                    @click.stop="selectMaterial(material)"
                  >
                    预览
                  </el-button>
                  <el-button
                    v-if="hasDownloadPermission(material)"
                    size="mini"
                    type="text"
                    icon="el-icon-download"
                    @click.stop="downloadMaterial(material)"
                  >
                    下载
                  </el-button>
                  <el-button
                    v-if="!hasPreviewPermission(material) && !hasDownloadPermission(material)"
                    size="mini"
                    type="text"
                    disabled
                    icon="el-icon-lock"
                  >
                    无权限
                  </el-button>
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <el-empty
              v-if="!loading && materials.length === 0"
              :description="$t('m.No_Files_Yet')"
              :image-size="80"
            />
          </div>
        </div>
      </div>

      <!-- 右侧：预览面板 -->
      <div class="preview-panel" v-if="selectedMaterial">
        <div class="preview-header">
          <div class="preview-info">
            <div class="current-path">
              <i class="el-icon-folder-opened"></i>
              <span>{{ getCurrentFolderPath() }}</span>
            </div>
            <span class="file-title" :title="selectedMaterial.fileName">
              <i :class="getFileIcon(getFileType(selectedMaterial.fileName))"></i>
              {{ selectedMaterial.fileName }}
            </span>
          </div>
          <div class="preview-actions">
            <el-button
              v-if="hasDownloadPermission(selectedMaterial)"
              size="small"
              icon="el-icon-download"
              @click="downloadCurrentFile"
            >
              下载
            </el-button>
          </div>
        </div>

        <div class="preview-content" v-loading="previewLoading" element-loading-text="加载中...">
          <!-- 统一使用腾讯云COS文档预览 -->
          <COSDocViewer
            v-if="selectedMaterial && canPreview(getFileType(selectedMaterial.fileName))"
            :materialId="selectedMaterial.id"
            :fileName="selectedMaterial.fileName"
            :allowDownload="hasDownloadPermission(selectedMaterial)"
            @viewer-loaded="handleCosViewerLoaded"
          />

          <!-- 不支持预览的文件 -->
          <div v-if="selectedMaterial && !canPreview(getFileType(selectedMaterial.fileName))" class="preview-unsupported">
            <i :class="getFileIcon(getFileType(selectedMaterial.fileName))"></i>
            <p>该文件类型不支持在线预览</p>
            <el-button
              v-if="hasDownloadPermission(selectedMaterial)"
              type="primary"
              icon="el-icon-download"
              @click="downloadCurrentFile"
            >
              下载文件
            </el-button>
          </div>
        </div>
      </div>

      <!-- 未选中文件时的提示 -->
      <div class="preview-panel empty" v-else>
        <div class="empty-hint">
          <i class="el-icon-document"></i>
          <p>点击左侧文件进行预览</p>
          <p class="hint-text">支持 PDF、PPT/Word/Excel、TXT 在线预览</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'

import studentAuth from '@/mixins/studentAuth'

import COSDocViewer from '@/components/COSDocViewer.vue'

export default {
  name: 'Materials',
  components: {
    COSDocViewer
  },
  mixins: [realtimeSync, studentAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      folders: [],
      materials: [],
      folderTree: [],
      allFoldersCache: [], // 缓存所有文件夹
      currentFolderId: 0,
      currentFolder: null,
      isInitialLoad: true,
      treeProps: {
        label: 'folderName',
        children: 'children'
      },
      // 预览相关
      selectedMaterial: null,
      previewLoading: false,
      cachedCosViewers: new Set(), // 已缓存的COS预览（避免重复上传）
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadContent',
        immediate: true
      }
    }
  },
  computed: {
    previewUrl() {
      if (!this.selectedMaterial || !this.selectedMaterial.filePath) return ''
      let url = this.selectedMaterial.filePath
      if (url.startsWith('/')) {
        url = window.location.origin + url
      }

      return url
    },
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal, oldVal) {
        // 当班级切换时，重置为根目录
        if (oldVal !== undefined && newVal !== oldVal) {
          console.log('[Materials] 班级切换，重置到根目录')
          this.currentFolderId = 0
          this.selectedMaterial = null
          this.materials = []
          this.folders = []
        }

        if (newVal) {
          this.loadContent()
          this.loadFolderTree()
        }
      }
    },
    selectedMaterial: {
      immediate: false,
      handler(newVal) {
        if (newVal) {
          console.log('[Materials] 选中文件:', newVal.fileName, newVal.id)
        }
      }
    }
  },
  async mounted() {
    this.disableDownloads()
  },
  beforeDestroy() {
    this.enableDownloads()
  },
  methods: {
    // 处理COS查看器加载完成
    handleCosViewerLoaded() {
      console.log('[Materials] 腾讯云COS Viewer加载完成')
    },

    async loadContent() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading
      if (this.isInitialLoad) {
        this.loading = true
      }

      try {
        // 获取当前文件夹下的子文件夹
        const foldersRes = await this.$store.dispatch('classroom/getFolders', {
          classroomId: this.classroomId,
          params: { parentId: this.currentFolderId }
        })

        // 获取当前文件夹下的文件
        const materialsRes = await this.$store.dispatch('classroom/getMaterials', {
          classroomId: this.classroomId,
          folderId: this.currentFolderId || 'root'
        })

        if (foldersRes.code === 200) {
          this.folders = foldersRes.data || []
        }

        if (materialsRes.code === 200) {
          this.materials = materialsRes.data || []
        }
      } catch (error) {
        if (this.isInitialLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (this.isInitialLoad) {
          this.loading = false
          this.isInitialLoad = false
        }
      }
    },

    async loadFolderTree() {
      try {
        // 递归获取所有文件夹构建树形结构
        const allFolders = await this.fetchAllFoldersRecursive()
        this.allFoldersCache = allFolders
        this.folderTree = this.buildFolderTree(allFolders)
      } catch (error) {
        console.error('加载文件夹树失败:', error)
        // 如果加载失败，至少显示根节点
        this.folderTree = [{
          id: 0,
          folderName: this.$t('m.Root_Directory'),
          children: []
        }]
      }
    },

    async fetchAllFoldersRecursive(parentId = null) {
      try {
        const params = {}
        if (parentId !== null) {
          params.parentId = parentId
        }

        const res = await this.$store.dispatch('classroom/getFolders', {
          classroomId: this.classroomId,
          params
        })

        if (res.code === 200) {
          const currentFolders = res.data || []
          let allFolders = [...currentFolders]

          // 并行获取所有子文件夹
          const childPromises = currentFolders.map(folder =>
            this.fetchAllFoldersRecursive(folder.id)
          )

          const childResults = await Promise.all(childPromises)
          childResults.forEach(childFolders => {
            allFolders = allFolders.concat(childFolders)
          })

          return allFolders
        }

        return []
      } catch (error) {
        console.error('获取文件夹失败:', error)
        return []
      }
    },

    buildFolderTree(folders) {
      // 构建树形结构
      const map = {}
      const tree = []

      // 先添加根节点
      const rootNode = {
        id: 0,
        folderName: this.$t('m.Root_Directory'),
        children: []
      }
      map[0] = rootNode

      // 创建所有节点的映射
      folders.forEach(folder => {
        map[folder.id] = {
          id: folder.id,
          folderName: folder.folderName,
          parentId: folder.parentId,
          children: []
        }
      })

      // 构建父子关系
      folders.forEach(folder => {
        const node = map[folder.id]
        const parentId = folder.parentId || 0
        if (map[parentId]) {
          map[parentId].children.push(node)
        } else {
          // 如果找不到父节点，说明该文件夹的父文件夹没有被获取到（可能是权限问题或已删除）
          // 不再将这种孤立文件夹添加到根节点下，而是跳过它们
          console.warn(`buildFolderTree - 警告: 文件夹 ${folder.id} (${folder.folderName}) 的父节点 ${parentId} 不存在，跳过该文件夹`)
        }
      })

      tree.push(rootNode)
      return tree
    },

    handleNodeClick(data, node) {
      this.navigateToFolder(data.id)
    },

    navigateToFolder(folderId) {
      this.currentFolderId = folderId
      this.isInitialLoad = true
      this.loadContent()

      // 高亮树节点
      this.$nextTick(() => {
        if (this.$refs.folderTree) {
          this.$refs.folderTree.setCurrentKey(folderId)
        }
      })
    },

    getCurrentFolderName() {
      if (this.currentFolderId === 0) {
        return this.$t('m.Root_Directory')
      }

      const folder = this.allFoldersCache.find(f => f.id === this.currentFolderId)
      return folder ? folder.folderName : this.$t('m.Root_Directory')
    },

    // 获取当前文件夹的完整路径
    getCurrentFolderPath() {
      if (this.currentFolderId === 0) {
        return this.$t('m.Root_Directory')
      }

      const path = []
      let currentId = this.currentFolderId

      while (currentId !== 0) {
        const folder = this.allFoldersCache.find(f => f.id === currentId)
        if (!folder) break

        path.unshift(folder.folderName)
        currentId = folder.parentId || 0
      }

      path.unshift(this.$t('m.Root_Directory'))
      return path.join(' / ')
    },

    openFolder(folder) {
      this.navigateToFolder(folder.id)
    },

    downloadMaterial(material) {
      // 使用带权限验证的下载API
      const downloadUrl = `/rating-api/api/classroom/material/${material.id}/download`
      this.downloadFile(downloadUrl, material.fileName)
    },

    // 下载文件（携带token）
    downloadFile(url, filename) {
      this.$axios({
        method: 'get',
        url: url,
        responseType: 'blob'
      }).then(response => {
        // 创建blob链接，使用正确的MIME类型
        const blob = new Blob([response.data], {
          type: response.headers['content-type'] || 'application/octet-stream'
        })
        const link = document.createElement('a')
        link.href = window.URL.createObjectURL(blob)
        // 强制下载，不预览
        link.download = filename
        link.style.display = 'none'
        document.body.appendChild(link)
        link.click()
        // 延迟清理，确保下载开始
        setTimeout(() => {
          document.body.removeChild(link)
          window.URL.revokeObjectURL(link.href)
        }, 100)
      }).catch(error => {
        console.error('下载失败:', error)
        if (error.response && error.response.status === 401) {
          this.$message.error('请先登录')
        } else if (error.response && error.response.status === 403) {
          this.$message.error('您没有下载该资料的权限')
        } else {
          this.$message.error('下载失败：' + (error.response?.data?.message || error.message))
        }
      })
    },

    // 根据文件名获取文件类型
    getFileType(fileName) {
      if (!fileName) return 'unknown'
      const ext = fileName.split('.').pop().toLowerCase()

      const imageTypes = ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp', 'bmp']
      const videoTypes = ['mp4', 'webm', 'ogv', 'mov']
      const audioTypes = ['mp3', 'wav', 'aac', 'ogg', 'm4a']
      const documentTypes = ['pdf']
      const textTypes = ['txt', 'md']
      const officeTypes = ['doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx']

      if (imageTypes.includes(ext)) return 'image'
      if (videoTypes.includes(ext)) return 'video'
      if (audioTypes.includes(ext)) return 'audio'
      if (documentTypes.includes(ext)) return 'pdf'
      if (textTypes.includes(ext)) return 'txt'
      if (officeTypes.includes(ext)) {
        if (['doc', 'docx'].includes(ext)) return 'word'
        if (['xls', 'xlsx'].includes(ext)) return 'excel'
        if (['ppt', 'pptx'].includes(ext)) return 'ppt'
      }

      return 'unknown'
    },

    // 判断是否可预览
    canPreview(fileType) {
      // 视频不支持在线预览
      const previewableTypes = ['pdf', 'txt', 'image', 'audio', 'ppt', 'word', 'excel']
      return previewableTypes.includes(fileType)
    },

    // 选择文件进行预览
    async selectMaterial(material) {
      if (!this.hasPreviewPermission(material)) {
        this.$message.warning('您没有预览权限，请联系教师')
        return
      }
      this.selectedMaterial = material
    },


    // 下载当前选中的文件
    downloadCurrentFile() {
      if (this.selectedMaterial) {
        // 使用带权限验证的下载API
        const downloadUrl = `/rating-api/api/classroom/material/${this.selectedMaterial.id}/download`
        this.downloadFile(downloadUrl, this.selectedMaterial.fileName)
      }
    },

    // 全屏切换

    // 在新窗口打开PDF
    openInNewWindow() {
      if (this.selectedMaterial && this.selectedMaterial.filePath) {
        window.open(this.previewUrl, '_blank')
      }
    },

    getFileIcon(type) {
      const icons = {
        pdf: 'el-icon-document',
        word: 'el-icon-document',
        ppt: 'el-icon-ppt',
        txt: 'el-icon-tickets',
        mp4: 'el-icon-video-play'
      }
      return icons[type] || 'el-icon-document'
    },

    formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
    },

    // 检查是否有预览权限
    hasPreviewPermission(material) {
      if (!material || !material.permission) return false
      return material.permission.canPreview === 1
    },

    // 检查是否有下载权限
    hasDownloadPermission(material) {
      if (!material || !material.permission) return false
      return material.permission.canDownload === 1
    },

    // 禁用页面下载功能（安全措施）
    disableDownloads() {
      // 禁用右键菜单
      this.contextMenuHandler = (e) => {
        e.preventDefault()
        return false
      }
      document.addEventListener('contextmenu', this.contextMenuHandler)

      // 禁用常见快捷键（保存、打印、源代码等）
      this.keyDownHandler = (e) => {
        // Ctrl+S (保存)
        if (e.ctrlKey && e.key === 's') {
          e.preventDefault()
          return false
        }
        // Ctrl+P (打印)
        if (e.ctrlKey && e.key === 'p') {
          e.preventDefault()
          return false
        }
        // F12 (开发者工具)
        if (e.key === 'F12') {
          e.preventDefault()
          return false
        }
        // Ctrl+U (查看源代码)
        if (e.ctrlKey && e.key === 'u') {
          e.preventDefault()
          return false
        }
        // Ctrl+Shift+I (开发者工具)
        if (e.ctrlKey && e.shiftKey && e.key === 'I') {
          e.preventDefault()
          return false
        }
        // Ctrl+Shift+C (检查元素)
        if (e.ctrlKey && e.shiftKey && e.key === 'C') {
          e.preventDefault()
          return false
        }
        // Ctrl+Shift+S (另存为)
        if (e.ctrlKey && e.shiftKey && e.key === 'S') {
          e.preventDefault()
          return false
        }
      }
      document.addEventListener('keydown', this.keyDownHandler, true)
    },

    // 恢复页面功能
    enableDownloads() {
      if (this.contextMenuHandler) {
        document.removeEventListener('contextmenu', this.contextMenuHandler)
      }
      if (this.keyDownHandler) {
        document.removeEventListener('keydown', this.keyDownHandler, true)
      }
    },

  }
}
</script>

<style scoped>
.student-materials {
  padding: 8px;
  min-height: 100vh;
  background: var(--classroom-bg, #f5f7fa);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 12px 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header h3 {
  font-size: 16px;
  color: var(--classroom-text);
  margin: 0;
  font-weight: 700;
}

/* 左右分屏布局 */
.content-layout-split {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 16px;
  height: calc(100vh - 100px);
}

/* 左侧列表面板 */
.file-list-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
}

/* 树形导航 */
.tree-sidebar {
  padding: 12px;
  border-bottom: 1px solid #E4E7ED;
  background: #F5F7FA;
  max-height: 200px;
  overflow-y: auto;
}

.tree-header {
  padding: 8px 12px;
  margin-bottom: 8px;
  background: white;
  border-radius: 6px;
}

.tree-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.folder-tree {
  background: white;
  padding: 8px;
  border-radius: 6px;
}

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  font-size: 13px;
}

.node-label {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-label i {
  font-size: 14px;
  color: #E6A23C;
}

/* 文件列表 */
.files-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.files-list .content-header {
  padding: 12px 16px;
  border-bottom: 1px solid #E4E7ED;
  background: #F5F7FA;
}

.current-folder-name {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.files-list > div {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.section {
  margin-bottom: 16px;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #F5F7FA;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 6px;
}

.file-item:hover {
  background: #E6F0FF;
  transform: translateX(2px);
}

.file-item.active {
  background: #E6F0FF;
  border-left: 3px solid #409EFF;
}

.file-item i {
  font-size: 18px;
  color: #409EFF;
  flex-shrink: 0;
}

.file-item.folder-item i {
  color: #E6A23C;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
}

.file-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

/* 右侧预览面板 */
.preview-panel {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-panel.empty {
  justify-content: center;
  align-items: center;
  background: #F5F7FA;
}

.empty-hint {
  text-align: center;
  color: #909399;
}

.empty-hint i {
  font-size: 64px;
  color: #C0C4CC;
  margin-bottom: 16px;
}

.empty-hint p {
  margin: 8px 0;
  font-size: 14px;
}

.hint-text {
  font-size: 12px !important;
  color: #C0C4CC !important;
}

.office-preview-error {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #909399;
  padding: 40px;
  text-align: center;
}

.office-preview-error i {
  font-size: 64px;
  color: #E6A23C;
  margin-bottom: 16px;
}

.office-preview-error p {
  margin: 8px 0;
  font-size: 14px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px 20px;
  border-bottom: 1px solid #E4E7ED;
  background: #F5F7FA;
}

.preview-info {
  flex: 1;
  min-width: 0;
}

.current-path {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 6px;
}

.current-path i {
  font-size: 13px;
  color: #E6A23C;
}

.current-path span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.file-title i {
  font-size: 20px;
  color: #409EFF;
  flex-shrink: 0;
}

.preview-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.preview-content {
  flex: 1;
  background: #f5f5f5;
  position: relative;
  overflow: auto;
  min-height: 600px; /* 确保预览内容有最小高度 */
}

/* PDF预览 - PDF.js渲染 */
.pdf-viewer-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #525659;
}

.pdf-toolbar {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  background: #373737;
  border-bottom: 1px solid #4C4C4C;
  color: #E5E5E5;
  flex-shrink: 0;
}

.pdf-toolbar .el-button-group {
  margin-right: 10px;
}

.pdf-canvas-wrapper {
  flex: 1;
  overflow: auto;
  background: #666;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 20px;
}

.pdf-canvas-wrapper canvas {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  background: white;
  max-width: 100%;
}

.pdf-loading {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: white;
  font-size: 14px;
}

.pdf-loading i {
  font-size: 32px;
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Office文件预览 */
.office-preview {
  background: #f5f5f5;
}

/* Office 预览容器 */
.office-preview-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 600px; /* 确保Office Online有足够的显示空间 */
}

.ppt-hybrid-preview {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 600px;
}

/* Office 容器 */
.office-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  flex: 1;
}

/* Office Viewer iframe */
.office-viewer-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

/* Office 下载按钮遮罩层 - 覆盖右下角下载按钮 */
.office-download-mask {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 60px;
  height: 60px;
  z-index: 9999;
  pointer-events: auto; /* 只在遮罩区域阻止点击 */
  background: transparent;
  cursor: default;
}

/* Office预览加载状态 */
.office-loading {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #409EFF;
}

.office-loading i {
  font-size: 48px;
  animation: rotating 2s linear infinite;
  margin-bottom: 16px;
}

.office-loading p {
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

/* 视频预览 */
.preview-video {
  width: 100%;
  max-height: 100%;
  display: block;
}

/* 音频预览 */
.preview-audio {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.audio-icon {
  text-align: center;
  margin-top: 20px;
}

.audio-icon i {
  font-size: 80px;
  opacity: 0.8;
}

.audio-icon p {
  margin-top: 16px;
  font-size: 16px;
}

/* 图片预览 */
.preview-image {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background: #f5f5f5;
}

.preview-image img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* 不支持预览 */
.preview-unsupported {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #909399;
}

.preview-unsupported i {
  font-size: 64px;
  color: #C0C4CC;
  margin-bottom: 16px;
}

.preview-unsupported p {
  margin-bottom: 16px;
  font-size: 14px;
}

/* 响应式 */
@media (max-width: 1024px) {
  .content-layout-split {
    grid-template-columns: 320px 1fr;
  }
}

@media (max-width: 768px) {
  .content-layout-split {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
    height: auto;
  }

  .file-list-panel {
    max-height: 400px;
  }

  .preview-panel {
    min-height: 500px;
  }
}
</style>
