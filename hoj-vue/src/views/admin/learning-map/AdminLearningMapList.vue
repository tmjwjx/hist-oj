<template>
  <div class="admin-learning-map-list">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">{{ $t('m.MapAdm_Title') }}</span>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="openCreateDialog">{{ $t('m.MapAdm_Create_Map') }}</el-button>
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
        <vxe-table-column field="title" :title="$t('m.MapAdm_Title_Field')" min-width="260" show-overflow></vxe-table-column>
        <vxe-table-column field="description" :title="$t('m.MapAdm_Description')" min-width="280" show-overflow></vxe-table-column>
        <vxe-table-column :title="$t('m.MapAdm_Status')" width="110">
          <template v-slot="{ row }">
            <el-tag size="mini" :type="row.status === 'published' ? 'success' : 'info'">{{ mapStatusLabel(row.status) }}</el-tag>
          </template>
        </vxe-table-column>
        <vxe-table-column :title="$t('m.MapAdm_Operation')" min-width="260">
          <template v-slot="{ row }">
            <el-button type="primary" size="mini" @click="openEditor(row.id)">{{ $t('m.MapAdm_Edit') }}</el-button>
            <el-button type="danger" size="mini" @click="deleteMap(row.id)">{{ $t('m.MapAdm_Delete') }}</el-button>
          </template>
        </vxe-table-column>
      </vxe-table>
    </el-card>

    <el-dialog :title="$t('m.MapAdm_Create_Map')" :visible.sync="createVisible" width="540px">
      <el-form label-width="90px" :model="createForm">
        <el-form-item :label="$t('m.MapAdm_Title_Field')">
          <el-input v-model="createForm.title" :placeholder="$t('m.MapAdm_Title_Placeholder')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.MapAdm_Description')">
          <el-input type="textarea" :rows="4" v-model="createForm.description"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button size="small" @click="createVisible = false">{{ $t('m.MapAdm_Cancel') }}</el-button>
        <el-button size="small" type="primary" @click="createMap">{{ $t('m.MapAdm_Create') }}</el-button>
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
      return status === 'published' ? this.$t('m.MapAdm_Published') : this.$t('m.MapAdm_Hidden')
    },
    async loadMaps() {
      this.loading = true
      try {
        this.maps = await learningMapApi.adminListMaps()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Load_List_Failed'))
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
        this.$message.warning(this.$t('m.MapAdm_Title_Required'))
        return
      }
      try {
        const res = await learningMapApi.adminCreateMap({
          title: this.createForm.title,
          description: this.createForm.description,
          status: 'draft'
        })
        this.$message.success(this.$t('m.MapAdm_Create_Success'))
        this.createVisible = false
        this.openEditor(res.id)
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Create_Failed'))
      }
    },
    openEditor(mapId) {
      this.$router.push({ name: 'admin-learning-map-editor', params: { mapId: String(mapId) } })
    },
    deleteMap(mapId) {
      this.$confirm(this.$t('m.MapAdm_Delete_Confirm'), this.$t('m.MapAdm_Dangerous_Operation'), {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteMap(mapId)
          this.$message.success(this.$t('m.MapAdm_Delete_Success'))
          this.loadMaps()
        } catch (e) {
          this.$message.error(e.message || this.$t('m.MapAdm_Delete_Failed'))
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
