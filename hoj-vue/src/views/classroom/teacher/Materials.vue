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

    <div class="content-layout-split">
      <!-- 左侧：树形导航和文件列表面板 -->
      <div class="file-list-panel">
        <!-- 树形导航 -->
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
            :default-expand-all="true"
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

        <!-- 文件列表 -->
        <div class="files-list">
          <div v-loading="loading">
            <!-- 文件夹和文件列表 -->
            <div v-if="folders.length > 0 || materials.length > 0" class="section">
              <!-- 文件夹卡片 -->
              <div
                v-for="folder in folders"
                :key="'folder-' + folder.id"
                class="file-item folder-item"
                @click="openFolder(folder)"
              >
                <i class="el-icon-folder-opened"></i>
                <div class="file-info">
                  <span class="file-name" :title="folder.folderName">{{ folder.folderName }}</span>
                  <span class="file-size">文件夹</span>
                </div>
              </div>

              <!-- 文件卡片 -->
              <div
                v-for="material in materials"
                :key="'file-' + material.id"
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
                  <el-tooltip content="预览" placement="top" v-if="canPreview(getFileType(material.fileName))">
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-view"
                      @click.stop="selectMaterial(material)"
                    >
                      预览
                    </el-button>
                  </el-tooltip>
                  <el-tooltip content="权限设置" placement="top">
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-setting"
                      @click.stop="openPermissionDialog(material)"
                    >
                      权限
                    </el-button>
                  </el-tooltip>
                  <el-tooltip v-if="isAdmin" content="下载" placement="top">
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-download"
                      @click.stop="downloadMaterial(material)"
                    >
                      下载
                    </el-button>
                  </el-tooltip>
                  <el-tooltip content="删除" placement="top">
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-delete"
                      class="delete-btn"
                      @click.stop="deleteMaterial(material)"
                    >
                      删除
                    </el-button>
                  </el-tooltip>
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-if="!loading && folders.length === 0 && materials.length === 0" class="empty-state">
              <i class="el-icon-folder-opened"></i>
              <p>此文件夹为空</p>
              <p class="hint">点击上方按钮创建文件夹或上传文件</p>
            </div>
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
              v-if="isAdmin"
              size="small"
              icon="el-icon-download"
              @click="downloadCurrentFile"
            >
              下载
            </el-button>
            <el-button size="small" icon="el-icon-close" @click="closePreview">
              关闭
            </el-button>
          </div>
        </div>

        <div class="preview-content" v-loading="previewLoading" element-loading-text="加载中...">
          <!-- 统一使用腾讯云COS文档预览 -->
          <COSDocViewer
            v-if="selectedMaterial && canPreview(getFileType(selectedMaterial.fileName))"
            :materialId="selectedMaterial.id"
            :fileName="selectedMaterial.fileName"
            :allowDownload="isAdmin"
            @viewer-loaded="handleCosViewerLoaded"
          />

          <!-- 不支持预览的文件 -->
          <div v-if="selectedMaterial && !canPreview(getFileType(selectedMaterial.fileName))" class="preview-unsupported">
            <i :class="getFileIcon(getFileType(selectedMaterial.fileName))"></i>
            <p>该文件类型不支持在线预览</p>
            <p class="hint-text">请通过管理员权限下载文件</p>
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

    <!-- 权限设置对话框 -->
    <el-dialog
      title="资料权限设置"
      :visible.sync="showPermissionDialog"
      width="800px"
      custom-class="classroom-dialog permission-dialog"
      @open="loadPermissions"
    >
      <div v-loading="permissionLoading">
        <div class="permission-header">
          <div class="material-info">
            <i :class="getFileIcon(getFileType(currentPermissionMaterial?.fileName))"></i>
            <span>{{ currentPermissionMaterial?.fileName }}</span>
          </div>
          <div class="batch-actions">
            <el-button size="small" @click="batchSetPermission(true, false)">
              <i class="el-icon-view"></i>
              全部可预览
            </el-button>
            <el-button size="small" @click="batchSetPermission(false, true)">
              <i class="el-icon-download"></i>
              全部可下载
            </el-button>
            <el-button size="small" type="primary" @click="batchSetPermission(true, true)">
              <i class="el-icon-check"></i>
              全部开启
            </el-button>
            <el-button size="small" @click="batchSetPermission(false, false)">
              <i class="el-icon-close"></i>
              全部关闭
            </el-button>
          </div>
        </div>

        <!-- 空状态提示 -->
        <el-empty
          v-if="!permissionLoading && studentPermissions.length === 0"
          description="该班级暂无学生"
          :image-size="80"
        >
          <p class="empty-hint">请先添加学生到班级，然后再设置权限</p>
        </el-empty>

        <!-- 学生列表 -->
        <el-table
          v-if="studentPermissions.length > 0"
          :data="studentPermissions"
          style="width: 100%; margin-top: 16px;"
          max-height="400"
        >
          <el-table-column prop="realName" label="学生姓名" width="150">
            <template slot-scope="scope">
              <div class="student-info">
                <span class="student-name">{{ scope.row.realName }}</span>
                <span class="student-username">{{ scope.row.username }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="可预览" width="100" align="center">
            <template slot-scope="scope">
              <el-checkbox v-model="scope.row.canPreview"></el-checkbox>
            </template>
          </el-table-column>
          <el-table-column label="可下载" width="100" align="center">
            <template slot-scope="scope">
              <el-checkbox v-model="scope.row.canDownload"></el-checkbox>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="120" align="center">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.canPreview || scope.row.canDownload ? 'success' : 'info'"
                size="small"
              >
                {{ scope.row.canPreview || scope.row.canDownload ? '有权限' : '无权限' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <span slot="footer">
        <el-button @click="showPermissionDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="savePermissions" :loading="permissionLoading">
          保存设置
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'
import teacherAuth from '@/mixins/teacherAuth'
import COSDocViewer from '@/components/COSDocViewer.vue'

export default {
  name: 'Materials',
  components: {
    COSDocViewer
  },
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
      // 预览相关
      selectedMaterial: null,
      previewLoading: false,
      userRoles: [], // 用户角色列表
      // 权限管理相关
      showPermissionDialog: false,
      currentPermissionMaterial: null,
      permissionLoading: false,
      studentPermissions: [],
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
        classroomId: this.classroomId,
        folderId: this.currentFolderId || 0
      }
    },
    previewUrl() {
      if (!this.selectedMaterial || !this.selectedMaterial.filePath) return ''
      let url = this.selectedMaterial.filePath
      if (url.startsWith('/')) {
        url = window.location.origin + url
      }

      // 对于PDF文件，添加参数隐藏工具栏（禁用下载按钮）
      if (this.getFileType(this.selectedMaterial.fileName) === 'pdf') {
        url += '#toolbar=0&navpanes=0&scrollbar=0'
      }

      return url
    },
    // 判断是否为管理员（管理员可以下载）
    isAdmin() {
      return this.userRoles.includes('admin') || this.userRoles.includes('root') || this.userRoles.includes('problem_admin')
    }
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
    // 获取用户角色
    await this.fetchUserRoles()
    this.disableDownloads()
  },
  beforeDestroy() {
    this.enableDownloads()
  },
  methods: {
    // 获取用户角色
    async fetchUserRoles() {
      try {
        const res = await this.$axios.get('/api/get-user-auth-info')
        // 后端返回格式：{status: 200, data: {roles: ["root"], permissions: null}}
        if (res.data.status === 200) {
          this.userRoles = res.data.data?.roles || []
          console.log('[Materials] 用户角色:', this.userRoles)
          console.log('[Materials] 是否管理员:', this.isAdmin)
        }
      } catch (error) {
        console.error('[Materials] 获取用户角色失败:', error)
      }
    },

    // 处理COS查看器加载完成
    handleCosViewerLoaded() {
      console.log('[Materials] 腾讯云COS Viewer加载完成')
    },

    // 处理PPT查看器加载完成（用于缓存）
    handlePptViewerLoaded(materialId) {
      if (materialId && !this.cachedPptViewers.has(materialId)) {
        const material = this.materials.find(m => m.id === materialId)
        console.log('[Materials] PPT查看器加载完成，添加到缓存:', material?.fileName || materialId)
        this.cachedPptViewers.add(materialId)
      }
    },

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
          this.$store.dispatch('classroom/getMaterials', {
            classroomId: this.classroomId,
            folderId: this.currentFolderId || 'root'
          })
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
        }
        // 如果找不到父节点，跳过该孤立文件夹
      })

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
      this.selectedMaterial = material
      const fileType = this.getFileType(material.fileName)

      // 非PDF文件直接关闭loading
      if (fileType !== 'pdf') {
        this.previewLoading = false
      }
    },


    // 关闭预览
    closePreview() {
      this.selectedMaterial = null
      this.previewLoading = false
    },

    // 下载当前选中的文件
    downloadCurrentFile() {
      if (this.selectedMaterial) {
        // 使用带权限验证的下载API
        const downloadUrl = `/rating-api/api/classroom/material/${this.selectedMaterial.id}/download`
        this.downloadFile(downloadUrl, this.selectedMaterial.fileName)
      }
    },

    // 打开权限设置对话框
    openPermissionDialog(material) {
      this.currentPermissionMaterial = material
      this.showPermissionDialog = true
    },

    // 加载学生权限列表
    async loadPermissions() {
      if (!this.currentPermissionMaterial) return

      this.permissionLoading = true
      try {
        const response = await this.$axios.get(`/rating-api/api/classroom/material/${this.currentPermissionMaterial.id}/permissions`)
        if (response.data.code === 200) {
          this.studentPermissions = response.data.data || []
        } else {
          this.$message.error(response.data.message || '加载权限失败')
        }
      } catch (error) {
        console.error('加载权限失败:', error)
        this.$message.error('加载权限失败')
      } finally {
        this.permissionLoading = false
      }
    },

    // 批量设置权限
    batchSetPermission(canPreview, canDownload) {
      this.studentPermissions.forEach(student => {
        student.canPreview = canPreview
        student.canDownload = canDownload
      })
      this.$message.success('已批量设置，请点击保存按钮保存更改')
    },

    // 保存权限设置
    async savePermissions() {
      this.permissionLoading = true
      try {
        const data = {
          materialId: this.currentPermissionMaterial.id,
          permissions: this.studentPermissions.map(student => ({
            uid: student.uid,
            canPreview: student.canPreview,
            canDownload: student.canDownload
          }))
        }

        const response = await this.$axios.post('/rating-api/api/classroom/material/permissions', data)
        if (response.data.code === 200) {
          this.$message.success('权限设置保存成功')
          this.showPermissionDialog = false
        } else {
          this.$message.error(response.data.message || '保存失败')
        }
      } catch (error) {
        console.error('保存权限失败:', error)
        this.$message.error('保存权限失败')
      } finally {
        this.permissionLoading = false
      }
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

/* 主布局 - 左右分屏 */
.content-layout-split {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 16px;
  height: calc(100vh - 140px);
}

/* 左侧文件列表面板 */
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
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  justify-content: space-between;
  font-size: 13px;
  padding-right: 4px;
}

.node-label {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}

.node-label i {
  font-size: 14px;
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

/* 文件列表 */
.files-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
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

/* PDF预览 */
.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: white;
  display: block;
}

/* PDF 容器 */
.pdf-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #525252;
}

/* PDF.js 工具栏 */
.pdf-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #323232;
  color: white;
  border-bottom: 1px solid #444;
  flex-shrink: 0;
}

.pdf-toolbar-left,
.pdf-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pdf-page-info {
  padding: 0 12px;
  font-size: 13px;
  color: #fff;
  min-width: 60px;
  text-align: center;
}

/* Canvas 容器 */
.pdf-canvas-container {
  flex: 1;
  overflow: auto;
  display: flex;
  justify-content: center;
  padding: 20px;
  background: #525252;
}

.pdf-canvas-container canvas {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  background: white;
  max-width: 100%;
  height: auto;
}

/* PDF 工具栏遮罩层 (旧版，已弃用) */
.pdf-toolbar-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 48px;
  z-index: 10;
  pointer-events: auto;
  /* 透明背景，但阻止点击 */
  background: transparent;
  cursor: default;
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

/* 权限对话框样式 */
.permission-dialog .permission-header {
  padding-bottom: 16px;
  border-bottom: 1px solid #E4E7ED;
  margin-bottom: 16px;
}

.permission-dialog .material-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}

.permission-dialog .material-info i {
  font-size: 18px;
  color: #409EFF;
}

.permission-dialog .batch-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.permission-dialog .student-info {
  display: flex;
  flex-direction: column;
}

.permission-dialog .student-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.permission-dialog .student-username {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

@media (max-width: 768px) {
  .permission-dialog .batch-actions {
    flex-direction: column;
  }

  .permission-dialog .batch-actions .el-button {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .materials-panel {
    padding: 16px;
  }

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

/* Office文件预览样式 */
.office-preview-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.office-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  flex: 1;
}

.office-viewer-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

.office-download-mask {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 60px;
  height: 60px;
  z-index: 9999;
  pointer-events: auto;
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
</style>
