<template>
  <div class="materials-panel">
    <div class="header">
      <h3>{{ $t('m.Material_Library') }}</h3>
      <div class="actions">
        <el-button icon="el-icon-folder-add" @click="showCreateFolderDialog = true">
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
          <el-button type="primary" icon="el-icon-upload" :loading="uploading">
            {{ uploading ? `${$t('m.Uploading')} (${uploadProgress}%)` : $t('m.Upload_File') }}
          </el-button>
        </el-upload>
      </div>
    </div>

    <el-breadcrumb separator="/">
      <el-breadcrumb-item><a href="javascript:;" @click="goToRoot">{{ $t('m.Root_Directory') }}</a></el-breadcrumb-item>
      <el-breadcrumb-item v-for="(folder, index) in folderPath" :key="folder.id">
        <a href="javascript:;" v-if="index < folderPath.length - 1" @click="navigateToFolder(index)">{{ folder.folderName }}</a>
        <span v-else>{{ folder.folderName }}</span>
      </el-breadcrumb-item>
    </el-breadcrumb>

    <el-row :gutter="20" class="content">
      <el-col :span="8" v-for="folder in folders" :key="folder.id">
        <el-card class="folder-card" @click.native="openFolder(folder)">
          <div class="folder-content">
            <i class="el-icon-folder-opened"></i>
            <span>{{ folder.folderName }}</span>
          </div>
          <div class="folder-actions">
            <el-button size="small" @click.stop="editFolder(folder)">
              重命名
            </el-button>
            <el-button size="small" type="danger" icon="el-icon-delete" @click.stop="deleteFolder(folder)">
              删除
            </el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8" v-for="material in materials" :key="material.id">
        <el-card class="material-card">
          <div class="material-icon">
            <i :class="getFileIcon(material.fileType)"></i>
          </div>
          <div class="material-info">
            <span class="name">{{ material.fileName }}</span>
            <span class="size">{{ formatFileSize(material.fileSize) }}</span>
          </div>
          <div class="material-actions">
            <el-button size="small" @click.stop="downloadMaterial(material)">
              {{ $t('m.Download') }}
            </el-button>
            <el-button size="small" type="danger" @click.stop="deleteMaterial(material)">
              {{ $t('m.Delete') }}
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 创建文件夹对话框 -->
    <el-dialog :title="$t('m.Create_Folder')" :visible.sync="showCreateFolderDialog" width="400px">
      <el-input v-model="newFolderName" :placeholder="$t('m.Enter_Folder_Name')" />
      <span slot="footer">
        <el-button @click="showCreateFolderDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createFolder">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 重命名文件夹对话框 -->
    <el-dialog title="重命名文件夹" :visible.sync="showRenameFolderDialog" width="400px">
      <el-input v-model="editFolderName" placeholder="请输入新的文件夹名称" />
      <span slot="footer">
        <el-button @click="showRenameFolderDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="confirmRenameFolder">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>
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
      folderPath: [],
      currentFolderId: 0,
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
      isInitialLoad: true, // 标记是否是真正的首次加载
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 500,
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
        await Promise.all([
          this.$store.dispatch('classroom/getFolders', {
            classroomId: this.classroomId,
            params: { parentId: this.currentFolderId }
          }),
          this.$store.dispatch('classroom/getMaterials', this.currentFolderId || 'root')
        ])
        const newFolders = this.$store.state.classroom.folders || []
        const newMaterials = this.$store.state.classroom.materials || []

        // 智能更新:只在数量变化时才更新数组
        if (this.isInitialLoad) {
          this.folders = newFolders
          this.materials = newMaterials
        } else {
          // 轮询时:只检查数量变化
          if (newFolders.length !== this.folders.length) {
            this.folders = newFolders
          }
          if (newMaterials.length !== this.materials.length) {
            this.materials = newMaterials
          }
          // 数量相同时不更新,避免闪烁
        }
      } catch (error) {
        if (this.isInitialLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (this.isInitialLoad) {
          this.loading = false
          this.isInitialLoad = false // 标记首次加载完成
        }
      }
    },
    async createFolder() {
      if (!this.newFolderName) {
        this.$message.warning(this.$t('m.Enter_Folder_Name'))
        return
      }
      try {
        const data = {
          classroomId: Number(this.classroomId),  // 确保是数字类型
          folderName: this.newFolderName
        }
        // 只有当currentFolderId不为0时才传递parentId
        if (this.currentFolderId && this.currentFolderId !== 0) {
          data.parentId = Number(this.currentFolderId)  // 确保是数字类型
        }
        const res = await this.$store.dispatch('classroom/createFolder', data)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Create_Success'))
          this.showCreateFolderDialog = false
          this.newFolderName = ''
          this.isInitialLoad = true // 创建后强制刷新
          this.loadContent()
        } else {
          this.$message.error(res.message || this.$t('m.Create_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Create_Failed'))
      }
    },
    openFolder(folder) {
      this.folderPath.push(folder)
      this.currentFolderId = folder.id
      this.isInitialLoad = true // 切换文件夹时重置标记
      this.loadContent()
    },
    goToRoot() {
      this.folderPath = []
      this.currentFolderId = 0
      this.isInitialLoad = true // 返回根目录时重置标记
      this.loadContent()
    },
    navigateToFolder(index) {
      // 跳转到指定层级的文件夹
      this.folderPath = this.folderPath.slice(0, index + 1)
      this.currentFolderId = this.folderPath[index].id
      this.isInitialLoad = true
      this.loadContent()
    },
    handleUploadSuccess(response) {
      this.uploading = false
      this.uploadProgress = 0
      // 清除上传组件的内部文件列表,防止在按钮旁显示
      this.$nextTick(() => {
        if (this.$refs.upload) {
          this.$refs.upload.clearFiles()
        }
      })
      if (response.code === 200) {
        this.$message.success(this.$t('m.Upload_Success'))
        // 直接将新上传的文件添加到列表中,避免重新加载
        // 检查文件是否属于当前文件夹(注意:folderId可能是0或数字,需要类型转换)
        if (response.data) {
          const materialFolderId = response.data.folderId
          const currentFolderId = Number(this.currentFolderId)

          // 只有当文件的 folderId 与当前文件夹 ID 匹配时才添加
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
      // 清除上传组件的内部文件列表,防止在按钮旁显示
      this.$nextTick(() => {
        if (this.$refs.upload) {
          this.$refs.upload.clearFiles()
        }
      })
      let errorMessage = this.$t('m.Upload_Failed')

      // Parse error to provide specific feedback
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
      // Check file size (100MB limit)
      const maxSize = 100 * 1024 * 1024
      if (file.size > maxSize) {
        this.$message.error(this.$t('m.File_Size_Limit') || '文件大小不能超过100MB')
        return false
      }

      // Check file type (basic validation)
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

      // Allow files with no type or common office/document types
      if (file.type && !allowedTypes.includes(file.type) && !file.name.match(/\.(pdf|doc|docx|ppt|pptx|xlsx|xls|txt|mp4|jpg|jpeg|png|gif)$/i)) {
        this.$message.warning(this.$t('m.File_Type_Warning') || '文件类型可能不支持，建议上传常见文档格式')
      }

      return true
    },
    downloadMaterial(material) {
      // 确保文件路径是完整的 URL
      let filePath = material.filePath
      if (filePath.startsWith('/')) {
        // 如果是相对路径，添加域名
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
            this.isInitialLoad = true // 删除后强制刷新
            this.loadContent()
          } else {
            this.$message.error(res.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
        }
      })
    },
    async deleteFolder(folder) {
      this.$confirm('确认删除文件夹及其所有内容吗?此操作不可恢复!', '警告', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteFolder', folder.id)
          if (res.code === 200) {
            this.$message.success('删除成功')
            this.isInitialLoad = true // 删除后强制刷新
            this.loadContent()
          } else {
            this.$message.error(res.message || '删除失败')
          }
        } catch (error) {
          this.$message.error('删除失败')
        }
      })
    },
    editFolder(folder) {
      this.currentEditFolder = folder
      this.editFolderName = folder.folderName
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
          this.isInitialLoad = true // 重命名后强制刷新
          this.loadContent()
        } else {
          this.$message.error(res.message || '重命名失败')
        }
      } catch (error) {
        this.$message.error('重命名失败')
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
.materials-panel {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  font-size: 20px;
  color: #409EFF;
}

.actions {
  display: flex;
  gap: 10px;
}

.content {
  margin-top: 20px;
}

.folder-card,
.material-card {
  margin-bottom: 15px;
  cursor: pointer;
}

.folder-card {
  text-align: center;
  padding: 20px;
}

.folder-content {
  margin-bottom: 10px;
}

.folder-card i {
  font-size: 48px;
  color: #E6A23C;
}

.folder-card span {
  display: block;
  margin-top: 10px;
  font-size: 16px;
}

.folder-actions {
  margin-top: 10px;
}

.material-card {
  display: flex;
  align-items: center;
  padding: 15px;
}

.material-icon i {
  font-size: 36px;
  color: #409EFF;
  margin-right: 15px;
}

.material-info {
  flex: 1;
}

.material-info .name {
  display: block;
  font-size: 14px;
  font-weight: bold;
}

.material-info .size {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.material-actions {
  display: flex;
  gap: 5px;
}

.el-breadcrumb {
  margin-bottom: 20px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.el-breadcrumb a {
  color: #409EFF;
  cursor: pointer;
  text-decoration: none;
}

.el-breadcrumb a:hover {
  text-decoration: underline;
}
</style>
