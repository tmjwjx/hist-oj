<template>
  <div class="materials-panel">
    <div class="header">
      <h3>{{ $t('m.Material_Library') }}</h3>
      <div class="actions">
        <el-button icon="el-icon-folder-add" @click="showCreateFolderDialog = true">
          {{ $t('m.Create_Folder') }}
        </el-button>
        <el-upload
          :action="uploadUrl"
          :headers="uploadHeaders"
          :data="{ folderId: currentFolderId }"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :on-progress="handleUploadProgress"
          :before-upload="beforeUpload"
          :show-file-list="false"
        >
          <el-button type="primary" icon="el-icon-upload" :loading="uploading">
            {{ uploading ? `${$t('m.Uploading')} (${uploadProgress}%)` : $t('m.Upload_File') }}
          </el-button>
        </el-upload>
      </div>
    </div>

    <el-breadcrumb separator="/">
      <el-breadcrumb-item>{{ $t('m.Root_Directory') }}</el-breadcrumb-item>
      <el-breadcrumb-item v-for="folder in folderPath" :key="folder.id">
        {{ folder.folderName }}
      </el-breadcrumb-item>
    </el-breadcrumb>

    <el-row :gutter="20" class="content">
      <el-col :span="8" v-for="folder in folders" :key="folder.id">
        <el-card class="folder-card" @click.native="openFolder(folder)">
          <i class="el-icon-folder-opened"></i>
          <span>{{ folder.folderName }}</span>
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
  </div>
</template>

<script>
export default {
  name: 'Materials',
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
      uploadUrl: '/rating-api/api/classroom/material/upload',
      uploadHeaders: {
        Authorization: localStorage.getItem('token') || ''
      },
      uploading: false,
      uploadProgress: 0
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
      this.loading = true
      try {
        await Promise.all([
          this.$store.dispatch('classroom/getFolders', {
            classroomId: this.classroomId,
            params: { parentId: this.currentFolderId }
          }),
          this.$store.dispatch('classroom/getMaterials', this.currentFolderId || 'root')
        ])
        this.folders = this.$store.state.classroom.folders || []
        this.materials = this.$store.state.classroom.materials || []
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    async createFolder() {
      if (!this.newFolderName) {
        this.$message.warning(this.$t('m.Enter_Folder_Name'))
        return
      }
      try {
        const data = {
          classroomId: this.classroomId,
          folderName: this.newFolderName
        }
        // 只有当currentFolderId不为0时才传递parentId
        if (this.currentFolderId && this.currentFolderId !== 0) {
          data.parentId = this.currentFolderId
        }
        const res = await this.$store.dispatch('classroom/createFolder', data)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Create_Success'))
          this.showCreateFolderDialog = false
          this.newFolderName = ''
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
      this.loadContent()
    },
    handleUploadSuccess(response) {
      this.uploading = false
      this.uploadProgress = 0
      if (response.code === 200) {
        this.$message.success(this.$t('m.Upload_Success'))
        this.loadContent()
      } else {
        this.$message.error(response.message || this.$t('m.Upload_Failed'))
      }
    },
    handleUploadError(error) {
      this.uploading = false
      this.uploadProgress = 0
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

.folder-card i {
  font-size: 48px;
  color: #E6A23C;
}

.folder-card span {
  display: block;
  margin-top: 10px;
  font-size: 16px;
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
</style>
