<template>
  <div class="admin-learning-map-list">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">算法航海图管理</span>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="openCreateDialog">新建航海图</el-button>
      </div>

      <vxe-table
        :data="maps"
        border
        stripe
        auto-resize
        align="center"
        :loading="loading"
      >
        <vxe-table-column field="id" title="ID" width="90"></vxe-table-column>
        <vxe-table-column field="title" title="标题" min-width="260" show-overflow></vxe-table-column>
        <vxe-table-column field="description" title="描述" min-width="280" show-overflow></vxe-table-column>
        <vxe-table-column title="状态" width="110">
          <template v-slot="{ row }">
            <el-tag size="mini" :type="row.status === 'published' ? 'success' : 'info'">{{ mapStatusLabel(row.status) }}</el-tag>
          </template>
        </vxe-table-column>
        <vxe-table-column title="操作" min-width="260">
          <template v-slot="{ row }">
            <el-button type="primary" size="mini" @click="openEditor(row.id)">编辑</el-button>
            <el-button type="success" size="mini" @click="publishMap(row.id)">发布</el-button>
            <el-button size="mini" @click="previewMap(row.id)">预览</el-button>
            <el-button type="danger" size="mini" @click="deleteMap(row.id)">删除</el-button>
          </template>
        </vxe-table-column>
      </vxe-table>
    </el-card>

    <el-dialog title="新建航海图" :visible.sync="createVisible" width="540px">
      <el-form label-width="90px" :model="createForm">
        <el-form-item label="标题">
          <el-input v-model="createForm.title" placeholder="例如：算法基础路线"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input type="textarea" :rows="4" v-model="createForm.description"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button size="small" @click="createVisible = false">取消</el-button>
        <el-button size="small" type="primary" @click="createMap">创建</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import learningMapApi from '@/api/learningMap'

export default {
  name: 'AdminLearningMapList',
  data() {
    return {
      loading: false,
      maps: [],
      createVisible: false,
      createForm: {
        title: '',
        description: ''
      }
    }
  },
  mounted() {
    this.loadMaps()
  },
  methods: {
    mapStatusLabel(status) {
      return status === 'published' ? '已发布' : '草稿'
    },
    async loadMaps() {
      this.loading = true
      try {
        this.maps = await learningMapApi.adminListMaps()
      } catch (e) {
        this.$message.error(e.message || '获取航海图列表失败')
      } finally {
        this.loading = false
      }
    },
    openCreateDialog() {
      this.createForm = { title: '', description: '' }
      this.createVisible = true
    },
    async createMap() {
      if (!this.createForm.title.trim()) {
        this.$message.warning('请填写标题')
        return
      }
      try {
        const res = await learningMapApi.adminCreateMap({
          title: this.createForm.title,
          description: this.createForm.description,
          status: 'draft'
        })
        this.$message.success('创建成功')
        this.createVisible = false
        this.openEditor(res.id)
      } catch (e) {
        this.$message.error(e.message || '创建失败')
      }
    },
    openEditor(mapId) {
      this.$router.push({ name: 'admin-learning-map-editor', params: { mapId: String(mapId) } })
    },
    previewMap(mapId) {
      const url = this.$router.resolve({ name: 'LearningMapPage', params: { mapId: String(mapId) } })
      window.open(url.href, '_blank')
    },
    publishMap(mapId) {
      this.$confirm('发布前将进行依赖与题目绑定校验，确认继续？', '发布确认', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminPublishMap(mapId)
          this.$message.success('发布成功')
          this.loadMaps()
        } catch (e) {
          this.$message.error(e.message || '发布失败')
        }
      })
    },
    deleteMap(mapId) {
      this.$confirm('删除后不可恢复，确认删除该航海图？', '危险操作', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteMap(mapId)
          this.$message.success('删除成功')
          this.loadMaps()
        } catch (e) {
          this.$message.error(e.message || '删除失败')
        }
      })
    }
  }
}
</script>

<style scoped>
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
