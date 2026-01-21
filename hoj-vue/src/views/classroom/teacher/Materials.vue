<template>
  <div class="materials-panel classroom-theme">
    <div class="header">
      <h3>{{ $t('m.Material_Library') }}</h3>
      <div class="actions">
        <el-button icon="el-icon-folder-add" @click="showCreateFolderDialog = true" size="medium">
          {{ $t('m.Create_Folder') }}
        </el-button>
        <el-upload
          ref="upload"
          :action="uploadUrl"
          :headers="uploadHeaders"
          :data="uploadData"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :on-progress="handleUploadProgress"
          :before-upload="beforeUpload"
          :show-file-list="false"
          :auto-upload="true"
          :limit="1"
        >
          <el-button type="primary" icon="el-icon-upload" :loading="uploading" size="medium">
            {{ uploading ? `${$t('m.Uploading')} (${uploadProgress}%)` : $t('m.Upload_File') }}
          </el-button>
        </el-upload>
      </div>
    </div>

    <div class="content-layout">
      <!-- 左侧树形导航 -->
      <div class="tree-sidebar">
        <div class="tree-header">
          <span class="tree-title">文件夹结构</span>
          <el-tooltip content="刷新" placement="top">
            <el-button
              type="text"
              icon="el-icon-refresh"
              @click="refreshTree"
              size="small"
            />
          </el-tooltip>
        </div>
        <el-tree
          ref="folderTree"
          :data="folderTree"
          :props="treeProps"
          :highlight-current="true"
          node-key="id"
          :current-node-key="currentFolderId"
          :expand-on-click-node="false"
          @node-click="handleNodeClick"
          @node-contextmenu="handleNodeContextMenu"
          class="folder-tree"
        >
          <span class="custom-tree-node" slot-scope="{ node, data }">
            <span class="node-label">
              <i :class="data.id === 0 ? 'el-icon-folder' : 'el-icon-folder-opened'"></i>
              {{ node.label }}
            </span>
            <span class="node-actions">
              <el-tooltip content="新建子文件夹" placement="top" v-if="data.id !== 0 || node.level < 3">
                <el-button
                  type="text"
                  size="mini"
                  icon="el-icon-folder-add"
                  @click.stop="createSubFolder(data)"
                />
              </el-tooltip>
              <el-tooltip content="上传文件到此处" placement="top">
                <el-button
                  type="text"
                  size="mini"
                  icon="el-icon-upload"
                  @click.stop="uploadToFolder(data)"
                />
              </el-tooltip>
              <el-tooltip content="重命名" placement="top" v-if="data.id !== 0">
                <el-button
                  type="text"
                  size="mini"
                  icon="el-icon-edit"
                  @click.stop="renameFolder(data)"
                />
              </el-tooltip>
              <el-tooltip content="删除" placement="top" v-if="data.id !== 0">
                <el-button
                  type="text"
                  size="mini"
                  icon="el-icon-delete"
                  class="delete-btn"
                  @click.stop="deleteFolder(data)"
                />
              </el-tooltip>
            </span>
          </span>
        </el-tree>
      </div>

      <!-- 右侧内容区域 -->
      <div class="content-area">
        <!-- 面包屑导航 -->
        <div class="breadcrumb-bar">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>
              <a href="javascript:;" @click="navigateToFolder(0)">
                <i class="el-icon-folder-opened"></i>
                {{ $t('m.Root_Directory') }}
              </a>
            </el-breadcrumb-item>
            <el-breadcrumb-item v-for="(folder, index) in currentPath" :key="folder.id">
              <a href="javascript:;" @click="navigateToFolder(folder.id)">
                <i class="el-icon-folder"></i>
                {{ folder.folderName }}
              </a>
            </el-breadcrumb-item>
          </el-breadcrumb>
          <div class="breadcrumb-info">
            <span v-if="currentFolder">
              <el-tag size="small" type="info">{{ folders.length }} 个文件夹</el-tag>
              <el-tag size="small" type="success" style="margin-left: 8px;">{{ materials.length }} 个文件</el-tag>
            </span>
          </div>
        </div>

        <!-- 文件夹和文件列表 -->
        <div v-loading="loading" class="file-grid">
          <!-- 空状态 -->
          <div v-if="!loading && folders.length === 0 && materials.length === 0" class="empty-state">
            <i class="el-icon-folder-opened"></i>
            <p>此文件夹为空</p>
            <p class="hint">点击上方按钮创建文件夹或上传文件</p>
          </div>

          <!-- 文件夹卡片 -->
          <div
            v-for="folder in folders"
            :key="'folder-' + folder.id"
            class="folder-card"
            @click="openFolder(folder)"
            @dblclick="openFolder(folder)"
          >
            <div class="folder-icon">
              <i class="el-icon-folder-opened"></i>
            </div>
            <div class="folder-info">
              <div class="folder-name" :title="folder.folderName">{{ folder.folderName }}</div>
              <div class="folder-meta">文件夹</div>
            </div>
          </div>

          <!-- 文件卡片 -->
          <div
            v-for="material in materials"
            :key="'file-' + material.id"
            class="file-card"
          >
            <div class="file-icon" :class="'file-type-' + getFileTypeClass(material.fileType)">
              <i :class="getFileIcon(material.fileType)"></i>
            </div>
            <div class="file-info">
              <div class="file-name" :title="material.fileName">{{ material.fileName }}</div>
              <div class="file-meta">{{ formatFileSize(material.fileSize) }}</div>
            </div>
            <div class="file-actions">
              <el-tooltip content="下载" placement="top">
                <el-button
                  type="text"
                  icon="el-icon-download"
                  size="small"
                  @click.stop="downloadMaterial(material)"
                />
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button
                  type="text"
                  icon="el-icon-delete"
                  size="small"
                  class="delete-btn"
                  @click.stop="deleteMaterial(material)"
                />
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建文件夹对话框 -->
    <el-dialog :title="$t('m.Create_Folder')" :visible.sync="showCreateFolderDialog" width="400px" custom-class="classroom-dialog">
      <el-input
        v-model="newFolderName"
        :placeholder="$t('m.Enter_Folder_Name')"
        prefix-icon="el-icon-folder"
        size="medium"
      />
      <span slot="footer">
        <el-button @click="showCreateFolderDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createFolder">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 重命名文件夹对话框 -->
    <el-dialog title="重命名文件夹" :visible.sync="showRenameFolderDialog" width="400px" custom-class="classroom-dialog">
      <el-input
        v-model="editFolderName"
        placeholder="请输入新的文件夹名称"
        prefix-icon="el-icon-edit"
        size="medium"
      />
      <span slot="footer">
        <el-button @click="showRenameFolderDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="confirmRenameFolder">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 右键菜单 -->
    <div
      v-show="contextMenuVisible"
      :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
      class="context-menu-wrapper"
    >
      <el-menu
        ref="contextMenu"
        class="context-menu"
        @select="handleContextMenuSelect"
      >
        <el-menu-item index="createSubFolder">
          <i class="el-icon-folder-add"></i>
          新建子文件夹
        </el-menu-item>
        <el-menu-item index="uploadFile">
          <i class="el-icon-upload"></i>
          上传文件到此
        </el-menu-item>
        <el-menu-item index="rename" v-if="contextMenuFolderId !== 0">
          <i class="el-icon-edit"></i>
          重命名
        </el-menu-item>
        <el-menu-item index="delete" v-if="contextMenuFolderId !== 0" class="danger-item">
          <i class="el-icon-delete"></i>
          删除
        </el-menu-item>
      </el-menu>
    </div>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'
import teacherAuth from '@/mixins/teacherAuth'

export default {
  name: 'Materials',
  mixins: [realtimeSync, teacherAuth],
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
      currentPath: [],
      currentFolderId: 0,
      currentFolder: null,
      showCreateFolderDialog: false,
      newFolderName: '',
      showRenameFolderDialog: false,
      editFolderName: '',
      currentEditFolder: null,
      uploadUrl: '/rating-api/api/classroom/material/upload',
      uploadHeaders: {
        Authorization: localStorage.getItem('token') || ''
      },
      uploading: false,
      uploadProgress: 0,
      isInitialLoad: true,
      treeProps: {
        label: 'folderName',
        children: 'children'
      },
      // 右键菜单
      contextMenuVisible: false,
      contextMenuFolderId: null,
      contextMenuPosition: { x: 0, y: 0 },
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
    uploadData() {
      return {
        folderId: this.currentFolderId || 0
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
      if (this.loading) return

      if (this.isInitialLoad) {
        this.loading = true
      }

      try {
        const [foldersRes, materialsRes] = await Promise.all([
          this.$store.dispatch('classroom/getFolders', {
            classroomId: this.classroomId,
            params: { parentId: this.currentFolderId }
          }),
          this.$store.dispatch('classroom/getMaterials', this.currentFolderId || 'root')
        ])

        const newFolders = foldersRes.data || []
        const newMaterials = materialsRes.data || []

        this.folders = newFolders
        this.materials = newMaterials

        // 更新当前路径信息
        this.updateCurrentPath()
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
        // 如果没有指定 parentId，从根目录开始
        if (parentId !== null) {
          params.parentId = parentId
        }

        console.log('fetchAllFoldersRecursive - parentId:', parentId, 'params:', params)

        const res = await this.$store.dispatch('classroom/getFolders', {
          classroomId: this.classroomId,
          params
        })

        if (res.code === 200) {
          const currentFolders = res.data || []
          console.log('fetchAllFoldersRecursive - parentId:', parentId, '返回:', currentFolders.map(f => ({ id: f.id, name: f.folderName, parentId: f.parentId })))
          let allFolders = [...currentFolders]

          // 并行获取所有子文件夹
          const childPromises = currentFolders.map(folder =>
            this.fetchAllFoldersRecursive(folder.id)
          )

          const childResults = await Promise.all(childPromises)
          childResults.forEach(childFolders => {
            allFolders = allFolders.concat(childFolders)
          })

          console.log('fetchAllFoldersRecursive - parentId:', parentId, '总共:', allFolders.length)
          return allFolders
        }

        return []
      } catch (error) {
        console.error('获取文件夹失败:', error)
        return []
      }
    },

    buildFolderTree(folders) {
      console.log('buildFolderTree - 输入的文件夹列表:', folders.map(f => ({ id: f.id, name: f.folderName, parentId: f.parentId })))

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

      console.log('buildFolderTree - map中的所有ID:', Object.keys(map))

      // 构建父子关系
      folders.forEach(folder => {
        const node = map[folder.id]
        const parentId = folder.parentId || 0
        console.log(`buildFolderTree - 处理文件夹 ${folder.id} (${folder.folderName}), parentId: ${parentId}`)

        if (map[parentId]) {
          map[parentId].children.push(node)
          console.log(`buildFolderTree - 将文件夹 ${folder.id} 添加到父节点 ${parentId}`)
        } else {
          // 如果找不到父节点，说明该文件夹的父文件夹没有被获取到（可能是权限问题或已删除）
          // 不再将这种孤立文件夹添加到根节点下，而是跳过它们
          console.warn(`buildFolderTree - 警告: 文件夹 ${folder.id} (${folder.folderName}) 的父节点 ${parentId} 不存在，跳过该文件夹`)
        }
      })

      console.log('buildFolderTree - 根节点下的子节点数量:', rootNode.children.length)
      console.log('buildFolderTree - 根节点下的子节点IDs:', rootNode.children.map(n => n.id))

      tree.push(rootNode)
      return tree
    },

    updateCurrentPath() {
      // 根据currentFolderId构建面包屑路径
      this.currentPath = []
      if (!this.currentFolderId || this.currentFolderId === 0) return

      // 从树中查找路径
      const path = this.findPathInTree(this.folderTree, this.currentFolderId)
      if (path) {
        this.currentPath = path
      }
    },

    findPathInTree(tree, targetId, currentPath = []) {
      for (const node of tree) {
        const path = [...currentPath, node]
        if (node.id === targetId) {
          return path.filter(n => n.id !== 0) // 排除根节点
        }
        if (node.children && node.children.length > 0) {
          const result = this.findPathInTree(node.children, targetId, path)
          if (result) return result
        }
      }
      return null
    },

    handleNodeClick(data, node) {
      this.navigateToFolder(data.id)
    },

    handleNodeContextMenu(event, data) {
      event.preventDefault()
      event.stopPropagation()

      this.contextMenuFolderId = data.id
      this.contextMenuPosition = {
        x: event.clientX,
        y: event.clientY
      }
      this.contextMenuVisible = true

      // 点击其他地方关闭菜单
      this.$nextTick(() => {
        document.addEventListener('click', this.hideContextMenu)
      })
    },

    hideContextMenu() {
      this.contextMenuVisible = false
      document.removeEventListener('click', this.hideContextMenu)
    },

    handleContextMenuSelect(index) {
      const folder = this.findFolderById(this.contextMenuFolderId)

      switch (index) {
        case 'createSubFolder':
          if (folder) {
            this.createSubFolder(folder)
          }
          break
        case 'uploadFile':
          if (folder) {
            this.uploadToFolder(folder)
          }
          break
        case 'rename':
          if (folder) {
            this.renameFolder(folder)
          }
          break
        case 'delete':
          if (folder) {
            this.deleteFolder(folder)
          }
          break
      }

      this.hideContextMenu()
    },

    findFolderById(id, tree = this.folderTree) {
      for (const node of tree) {
        if (node.id === id) {
          return node
        }
        if (node.children && node.children.length > 0) {
          const found = this.findFolderById(id, node.children)
          if (found) return found
        }
      }
      return null
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

    openFolder(folder) {
      this.navigateToFolder(folder.id)
    },

    refreshTree() {
      this.loadFolderTree()
      this.loadContent()
    },

    async createFolder() {
      if (!this.newFolderName) {
        this.$message.warning(this.$t('m.Enter_Folder_Name'))
        return
      }
      try {
        const data = {
          classroomId: Number(this.classroomId),
          folderName: this.newFolderName
        }
        if (this.currentFolderId && this.currentFolderId !== 0) {
          data.parentId = Number(this.currentFolderId)
        }
        const res = await this.$store.dispatch('classroom/createFolder', data)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Create_Success'))
          this.showCreateFolderDialog = false
          this.newFolderName = ''
          this.isInitialLoad = true
          this.loadContent()
          this.loadFolderTree()
        } else {
          this.$message.error(res.message || this.$t('m.Create_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Create_Failed'))
      }
    },

    createSubFolder(folderData) {
      this.navigateToFolder(folderData.id)
      this.showCreateFolderDialog = true
    },

    uploadToFolder(folderData) {
      this.navigateToFolder(folderData.id)
      this.$nextTick(() => {
        // 触发文件选择
        const uploadInput = document.querySelector('.el-upload input')
        if (uploadInput) {
          uploadInput.click()
        }
      })
    },

    renameFolder(folderData) {
      this.currentEditFolder = folderData
      this.editFolderName = folderData.folderName
      this.showRenameFolderDialog = true
    },

    async confirmRenameFolder() {
      if (!this.editFolderName || !this.editFolderName.trim()) {
        this.$message.warning('请输入文件夹名称')
        return
      }
      try {
        const res = await this.$store.dispatch('classroom/updateFolder', {
          id: this.currentEditFolder.id,
          folderName: this.editFolderName
        })
        if (res.code === 200) {
          this.$message.success('重命名成功')
          this.showRenameFolderDialog = false
          this.currentEditFolder = null
          this.editFolderName = ''
          this.isInitialLoad = true
          this.loadContent()
          this.loadFolderTree()
        } else {
          this.$message.error(res.message || '重命名失败')
        }
      } catch (error) {
        this.$message.error('重命名失败')
      }
    },

    async deleteFolder(folder) {
      this.$confirm('确认删除文件夹及其所有内容吗？此操作不可恢复！', '警告', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteFolder', folder.id)
          if (res.code === 200) {
            this.$message.success('删除成功')
            this.isInitialLoad = true
            this.loadContent()
            this.loadFolderTree()
          } else {
            this.$message.error(res.message || '删除失败')
          }
        } catch (error) {
          this.$message.error('删除失败')
        }
      })
    },

    handleUploadSuccess(response) {
      this.uploading = false
      this.uploadProgress = 0
      this.$nextTick(() => {
        if (this.$refs.upload) {
          this.$refs.upload.clearFiles()
        }
      })
      if (response.code === 200) {
        this.$message.success(this.$t('m.Upload_Success'))
        if (response.data) {
          const materialFolderId = response.data.folderId
          const currentFolderId = Number(this.currentFolderId)
          if (materialFolderId === currentFolderId) {
            this.materials.push(response.data)
          }
        }
      } else {
        this.$message.error(response.message || this.$t('m.Upload_Failed'))
      }
    },

    handleUploadError(error) {
      this.uploading = false
      this.uploadProgress = 0
      this.$nextTick(() => {
        if (this.$refs.upload) {
          this.$refs.upload.clearFiles()
        }
      })
      let errorMessage = this.$t('m.Upload_Failed')

      if (error.message) {
        if (error.message.includes('Network Error') || error.message.includes('timeout')) {
          errorMessage = this.$t('m.Network_Error') || '网络错误，请检查网络连接'
        } else if (error.message.includes('413')) {
          errorMessage = this.$t('m.File_Too_Large') || '文件大小超出限制'
        } else if (error.message.includes('415')) {
          errorMessage = this.$t('m.File_Type_Not_Supported') || '不支持的文件类型'
        } else if (error.message.includes('500')) {
          errorMessage = this.$t('m.Server_Error') || '服务器错误，请稍后重试'
        } else if (error.message.includes('401')) {
          errorMessage = this.$t('m.Unauthorized') || '未授权，请重新登录'
        }
      }

      this.$message.error(errorMessage)
    },

    handleUploadProgress(event) {
      this.uploading = true
      this.uploadProgress = Math.floor(event.percent)
    },

    beforeUpload(file) {
      const maxSize = 100 * 1024 * 1024
      if (file.size > maxSize) {
        this.$message.error(this.$t('m.File_Size_Limit') || '文件大小不能超过100MB')
        return false
      }

      const allowedTypes = [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/vnd.ms-powerpoint',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        'application/vnd.ms-excel',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'text/plain',
        'video/mp4',
        'image/jpeg',
        'image/png',
        'image/gif'
      ]

      if (file.type && !allowedTypes.includes(file.type) && !file.name.match(/\.(pdf|doc|docx|ppt|pptx|xlsx|xls|txt|mp4|jpg|jpeg|png|gif)$/i)) {
        this.$message.warning(this.$t('m.File_Type_Warning') || '文件类型可能不支持，建议上传常见文档格式')
      }

      return true
    },

    downloadMaterial(material) {
      let filePath = material.filePath
      if (filePath.startsWith('/')) {
        filePath = window.location.origin + filePath
      }
      window.open(filePath, '_blank')
    },

    async deleteMaterial(material) {
      this.$confirm(this.$t('m.Confirm_Delete_Material'), this.$t('m.Warning'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteMaterial', material.id)
          if (res.code === 200) {
            this.$message.success(this.$t('m.Delete_Success'))
            this.isInitialLoad = true
            this.loadContent()
          } else {
            this.$message.error(res.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
        }
      })
    },

    getFileIcon(type) {
      const icons = {
        pdf: 'el-icon-document',
        word: 'el-icon-document',
        ppt: 'el-icon-document',
        txt: 'el-icon-tickets',
        mp4: 'el-icon-video-play',
        image: 'el-icon-picture'
      }
      return icons[type] || 'el-icon-document'
    },

    getFileTypeClass(type) {
      const typeMap = {
        pdf: 'pdf',
        word: 'word',
        ppt: 'ppt',
        txt: 'txt',
        mp4: 'video',
        image: 'image'
      }
      return typeMap[type] || 'default'
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
@import '../classroom-theme.css';

.materials-panel {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
  max-width: 1600px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header h3 {
  font-size: 22px;
  color: var(--classroom-text);
  margin: 0;
  font-weight: 700;
}

.actions {
  display: flex;
  gap: 10px;
}

/* 主布局 */
.content-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 24px;
  height: calc(100vh - 160px);
}

/* 左侧树形导航 */
.tree-sidebar {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--classroom-border);
  background: #f8f9fa;
}

.tree-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--classroom-text);
}

.folder-tree {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
}

.node-label {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}

.node-label i {
  font-size: 16px;
  color: #E6A23C;
}

.node-actions {
  display: flex;
  gap: 2px;
  padding-left: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.custom-tree-node:hover .node-actions {
  opacity: 1;
}

.node-actions .el-button {
  padding: 4px;
}

.node-actions .delete-btn {
  color: #F56C6C;
}

.node-actions .delete-btn:hover {
  color: #F56C6C;
  background-color: #FEF0F0;
}

/* 右键菜单样式 */
.context-menu-wrapper {
  position: fixed;
  z-index: 9999;
}

.context-menu {
  border: 1px solid #e0e6ed;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  border-radius: 8px;
  padding: 4px 0;
  min-width: 180px;
  background: #fff;
}

.context-menu .el-menu-item {
  padding: 10px 20px;
  font-size: 14px;
  color: #2c3e50;
  line-height: 1.5;
}

.context-menu .el-menu-item i {
  margin-right: 10px;
  font-size: 16px;
}

.context-menu .el-menu-item:hover {
  background-color: #f8f9fa;
}

.context-menu .danger-item {
  color: #F56C6C;
}

.context-menu .danger-item:hover {
  background-color: #FEF0F0;
  color: #F56C6C;
}

/* 右侧内容区域 */
.content-area {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 面包屑导航 */
.breadcrumb-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--classroom-border);
  background: #f8f9fa;
}

.breadcrumb-bar .el-breadcrumb {
  margin-bottom: 0;
  background: transparent;
  padding: 0;
}

.breadcrumb-bar .el-breadcrumb a {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.breadcrumb-bar .el-breadcrumb i {
  font-size: 16px;
  color: #E6A23C;
}

.breadcrumb-info {
  display: flex;
  gap: 8px;
}

/* 文件网格 */
.file-grid {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 20px;
  align-content: start;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: #909399;
}

.empty-state i {
  font-size: 64px;
  margin-bottom: 16px;
  color: #C0C4CC;
}

.empty-state p {
  margin: 4px 0;
  font-size: 14px;
}

.empty-state .hint {
  font-size: 12px;
  color: #C0C4CC;
}

/* 文件夹卡片 */
.folder-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  background: #fff;
  border: 2px solid var(--classroom-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.folder-card:hover {
  border-color: var(--classroom-primary);
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.15);
  transform: translateY(-4px);
}

.folder-icon {
  font-size: 56px;
  color: #E6A23C;
  margin-bottom: 12px;
}

.folder-info {
  text-align: center;
  width: 100%;
}

.folder-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--classroom-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.folder-meta {
  font-size: 12px;
  color: #909399;
}

/* 文件卡片 */
.file-card {
  display: flex;
  flex-direction: column;
  padding: 16px;
  background: #fff;
  border: 2px solid var(--classroom-border);
  border-radius: 12px;
  transition: all 0.3s ease;
  position: relative;
}

.file-card:hover {
  border-color: var(--classroom-primary);
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.15);
}

.file-icon {
  font-size: 48px;
  margin-bottom: 12px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.file-icon i {
  font-size: 48px;
}

.file-icon.file-type-pdf {
  color: #F56C6C;
}

.file-icon.file-type-word {
  color: #409EFF;
}

.file-icon.file-type-ppt {
  color: #E6A23C;
}

.file-icon.file-type-txt {
  color: #909399;
}

.file-icon.file-type-video {
  color: #67C23A;
}

.file-icon.file-type-image {
  color: #909399;
}

.file-icon.file-type-default {
  color: #C0C4CC;
}

.file-info {
  text-align: center;
  margin-bottom: 8px;
}

.file-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--classroom-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.file-meta {
  font-size: 12px;
  color: #909399;
}

.file-actions {
  display: none;
  justify-content: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--classroom-border);
}

.file-card:hover .file-actions {
  display: flex;
}

.file-actions .delete-btn {
  color: #F56C6C;
}

.file-actions .delete-btn:hover {
  color: #F56C6C;
  background-color: #FEF0F0;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 280px 1fr;
  }

  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }
}

@media (max-width: 768px) {
  .materials-panel {
    padding: 16px;
  }

  .content-layout {
    grid-template-columns: 1fr;
    height: auto;
  }

  .tree-sidebar {
    max-height: 300px;
    margin-bottom: 20px;
  }

  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  }

  .breadcrumb-bar {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .header {
    flex-direction: column;
    gap: 16px;
  }

  .actions {
    width: 100%;
    flex-direction: column;
  }

  .actions .el-button {
    width: 100%;
  }
}
</style>
