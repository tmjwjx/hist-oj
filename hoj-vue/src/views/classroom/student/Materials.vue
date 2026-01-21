<template>
  <div class="student-materials classroom-theme">
    <div class="header">
      <h3>{{ $t('m.Material_Library') }}</h3>
    </div>

    <div class="content-layout">
      <!-- 左侧树形导航 -->
      <div class="tree-sidebar">
        <div class="tree-header">
          <span class="tree-title">文件夹结构</span>
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

      <!-- 右侧内容区域 -->
      <div class="content-area">
        <div class="content-header">
          <span class="current-folder-name">{{ getCurrentFolderName() }}</span>
        </div>

        <div class="files-container" v-loading="loading">
          <!-- 文件夹列表 -->
          <div v-if="folders.length > 0" class="section">
            <div class="section-title">文件夹</div>
            <div class="files-grid">
              <div
                v-for="folder in folders"
                :key="folder.id"
                class="file-item folder-item"
                @click="openFolder(folder)"
              >
                <i class="el-icon-folder folder-icon"></i>
                <span class="file-name">{{ folder.folderName }}</span>
              </div>
            </div>
          </div>

          <!-- 文件列表 -->
          <div v-if="materials.length > 0" class="section">
            <div class="section-title">文件</div>
            <div class="files-grid">
              <div
                v-for="material in materials"
                :key="material.id"
                class="file-item material-item"
              >
                <i :class="getFileIcon(material.fileType)" class="file-icon"></i>
                <div class="file-info">
                  <span class="file-name">{{ material.fileName }}</span>
                  <span class="file-size">{{ formatFileSize(material.fileSize) }}</span>
                </div>
                <el-button
                  size="mini"
                  type="primary"
                  icon="el-icon-download"
                  @click="downloadMaterial(material)"
                >
                  下载
                </el-button>
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <el-empty
            v-if="!loading && folders.length === 0 && materials.length === 0"
            :description="$t('m.No_Files_Yet')"
          />
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
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadContent',
        immediate: true
      }
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

.content-layout {
  display: flex;
  gap: 16px;
  height: calc(100vh - 100px);
}

/* 左侧树形导航 */
.tree-sidebar {
  width: 240px;
  flex-shrink: 0;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.tree-header {
  padding: 12px 16px;
  border-bottom: 1px solid #E4E7ED;
  background: #F5F7FA;
}

.tree-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.folder-tree {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
}

.node-label {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-label i {
  font-size: 16px;
  color: #E6A23C;
}

/* 右侧内容区域 */
.content-area {
  flex: 1;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-header {
  padding: 16px 20px;
  border-bottom: 1px solid #E4E7ED;
  background: #F5F7FA;
}

.current-folder-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.files-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: visible;
  padding: 16px;
}

.section {
  margin-bottom: 24px;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #909399;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.files-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #F5F7FA;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.file-item:hover {
  background: #E6F0FF;
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.15);
}

.folder-item {
  cursor: pointer;
}

.folder-icon {
  font-size: 32px;
  color: #E6A23C;
}

.file-icon {
  font-size: 28px;
  color: #409EFF;
}

.file-name {
  flex: 1;
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-size {
  font-size: 12px;
  color: #909399;
}

.material-item {
  cursor: default;
}

.material-item .file-info {
  flex: 1;
}

/* 响应式 */
@media (max-width: 768px) {
  .content-layout {
    flex-direction: column;
  }

  .tree-sidebar {
    width: 100%;
    max-height: 200px;
  }

  .files-grid {
    grid-template-columns: 1fr;
  }
}
</style>
