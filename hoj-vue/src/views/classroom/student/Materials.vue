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
                    v-if="canPreview(getFileType(material.fileName))"
                    size="mini"
                    type="text"
                    icon="el-icon-view"
                  >
                    预览
                  </el-button>
                  <el-button
                    size="mini"
                    type="text"
                    icon="el-icon-download"
                    @click.stop="downloadMaterial(material)"
                  >
                    下载
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
            <el-button size="small" icon="el-icon-download" @click="downloadCurrentFile">
              下载
            </el-button>
            <el-button size="small" icon="el-icon-full-screen" @click="toggleFullscreen">
              全屏
            </el-button>
          </div>
        </div>

        <div class="preview-content" v-loading="previewLoading" element-loading-text="加载中...">
          <!-- PDF 预览 -->
          <iframe
            v-if="getFileType(selectedMaterial.fileName) === 'pdf' && !previewLoading"
            :src="previewUrl"
            class="preview-iframe"
          ></iframe>

          <!-- 视频预览 -->
          <video
            v-if="getFileType(selectedMaterial.fileName) === 'video' && !previewLoading"
            :src="previewUrl"
            controls
            class="preview-video"
          ></video>

          <!-- 音频预览 -->
          <div v-if="getFileType(selectedMaterial.fileName) === 'audio' && !previewLoading" class="preview-audio">
            <audio :src="previewUrl" controls style="width: 100%; max-width: 600px;"></audio>
            <div class="audio-icon">
              <i class="el-icon-headset"></i>
              <p>音频文件</p>
            </div>
          </div>

          <!-- 图片预览 -->
          <div v-if="getFileType(selectedMaterial.fileName) === 'image' && !previewLoading" class="preview-image">
            <img :src="previewUrl" alt="预览图片" />
          </div>

          <!-- TXT 预览 -->
          <iframe
            v-if="getFileType(selectedMaterial.fileName) === 'txt' && !previewLoading"
            :src="previewUrl"
            class="preview-iframe"
          ></iframe>

          <!-- 不支持预览 -->
          <div v-if="!canPreview(getFileType(selectedMaterial.fileName))" class="preview-unsupported">
            <i :class="getFileIcon(getFileType(selectedMaterial.fileName))"></i>
            <p>该文件类型不支持在线预览</p>
            <el-button type="primary" icon="el-icon-download" @click="downloadCurrentFile">
              点击下载
            </el-button>
          </div>
        </div>
      </div>

      <!-- 未选中文件时的提示 -->
      <div class="preview-panel empty" v-else>
        <div class="empty-hint">
          <i class="el-icon-document"></i>
          <p>点击左侧文件进行预览</p>
          <p class="hint-text">支持 PDF、视频、图片、音频、文本文件在线预览</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'Materials',
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
    }
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadContent()
          this.loadFolderTree()
        }
      }
    }
  },
  mounted() {
    // mounted 时也会通过 watch 触发加载
  },
  methods: {
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
        const materialsRes = await this.$store.dispatch('classroom/getMaterials', this.currentFolderId || 'root')

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
      let filePath = material.filePath
      if (filePath.startsWith('/')) {
        filePath = window.location.origin + filePath
      }
      window.open(filePath, '_blank')
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
      const previewableTypes = ['pdf', 'txt', 'image', 'video', 'audio']
      return previewableTypes.includes(fileType)
    },

    // 选择文件进行预览
    selectMaterial(material) {
      this.selectedMaterial = material
      this.previewLoading = false
    },

    // 下载当前选中的文件
    downloadCurrentFile() {
      if (this.previewUrl) {
        window.open(this.previewUrl, '_blank')
      }
    },

    // 全屏切换
    toggleFullscreen() {
      const element = document.querySelector('.preview-content')
      if (element) {
        if (document.fullscreenElement) {
          document.exitFullscreen()
        } else {
          element.requestFullscreen().catch(err => {
            console.error('无法进入全屏模式:', err)
            this.$message.warning('您的浏览器不支持全屏预览')
          })
        }
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
    }
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
}

/* PDF预览 */
.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: white;
  display: block;
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
