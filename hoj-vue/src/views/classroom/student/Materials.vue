<template>
  <div class="student-materials">
    <!-- 面包屑导航 -->
    <el-breadcrumb separator="/" class="breadcrumb">
      <el-breadcrumb-item @click.native="goToRoot">
        <a href="javascript:;">{{ $t('m.Root_Directory') }}</a>
      </el-breadcrumb-item>
      <el-breadcrumb-item v-if="currentFolderId">
        {{ currentFolderName }}
      </el-breadcrumb-item>
    </el-breadcrumb>

    <!-- 文件夹和文件列表 -->
    <div v-loading="loading" class="content-area">
      <el-row :gutter="20">
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
            <el-button size="small" type="primary" @click="downloadMaterial(material)">
              {{ $t('m.Download') }}
            </el-button>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="!loading && folders.length === 0 && materials.length === 0" :description="$t('m.No_Files_Yet')" />
    </div>
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
      currentFolderId: 0, // 0表示根目录
      currentFolderName: ''
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
        // 获取当前文件夹下的子文件夹
        const foldersRes = await this.$store.dispatch('classroom/getFolders', {
          classroomId: this.classroomId,
          params: { parentId: this.currentFolderId }
        })

        // 获取当前文件夹下的文件 - 直接传folderId值，不要包在对象里
        const materialsRes = await this.$store.dispatch('classroom/getMaterials', this.currentFolderId || 'root')

        if (foldersRes.code === 200) {
          this.folders = foldersRes.data || []
        }
        if (materialsRes.code === 200) {
          this.materials = materialsRes.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    openFolder(folder) {
      this.currentFolderId = folder.id
      this.currentFolderName = folder.folderName
      this.loadContent()
    },
    goToRoot() {
      this.currentFolderId = 0
      this.currentFolderName = ''
      this.loadContent()
    },
    downloadMaterial(material) {
      // 使用 window.open 触发下载，对新标签页中的文件进行下载
      // 对于无法预览的文件（如 PDF、图片等），浏览器会自动下载
      // 对于可预览的文件，用户可以手动右键保存
      // 确保文件路径是完整的 URL
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
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.breadcrumb a {
  color: #409EFF;
  cursor: pointer;
}

.content-area {
  min-height: 200px;
}

.folder-card,
.material-card {
  margin-bottom: 15px;
  text-align: center;
  cursor: pointer;
  transition: transform 0.2s;
}

.folder-card:hover,
.material-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
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
  padding: 15px;
}

.material-icon i {
  font-size: 36px;
  color: #409EFF;
}

.material-info {
  margin: 10px 0;
}

.material-info .name {
  display: block;
  font-weight: bold;
}

.material-info .size {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}
</style>
