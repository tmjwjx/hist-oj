<template>
  <div class="learning-map-list-page">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">
          <i class="el-icon-s-opportunity"></i>
          算法知识点与题目航海图
        </span>
      </div>

      <el-empty v-if="!loading && maps.length === 0" description="暂无已发布航海图"></el-empty>

      <el-row v-else :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="12" :md="8" v-for="item in maps" :key="item.id" class="map-card-col">
          <el-card shadow="hover" class="map-card">
            <div class="map-card-title">{{ item.title }}</div>
            <div class="map-card-desc">{{ item.description || '暂无描述' }}</div>
            <div class="map-card-footer">
              <el-tag type="success" size="mini">{{ statusLabel(item.status) }}</el-tag>
              <el-button type="primary" size="mini" @click="openMap(item.id)">进入航海图</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script>
import learningMapApi from '@/api/learningMap'

export default {
  name: 'LearningMapList',
  data() {
    return {
      loading: false,
      maps: []
    }
  },
  mounted() {
    this.loadMaps()
  },
  methods: {
    statusLabel(status) {
      return status === 'published' ? '已发布' : '草稿'
    },
    async loadMaps() {
      this.loading = true
      try {
        this.maps = await learningMapApi.getPublishedMaps()
      } catch (e) {
        this.$message.error(e.message || '获取航海图失败')
      } finally {
        this.loading = false
      }
    },
    openMap(mapId) {
      this.$router.push({ name: 'LearningMapPage', params: { mapId: String(mapId) } })
    }
  }
}
</script>

<style scoped>
.learning-map-list-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 2px solid #409eff;
  padding: 20px;
}

.panel-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.panel-title i {
  color: #409eff;
  margin-right: 8px;
}

.map-card-col {
  margin-bottom: 20px;
}

.map-card {
  min-height: 160px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.2s;
  padding: 20px;
}

.map-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.map-card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.map-card-desc {
  min-height: 48px;
  color: #606266;
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 16px;
}

.map-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
