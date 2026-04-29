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
  max-width: 1280px;
  margin: 0 auto;
  padding: 10px 8px 14px;
  font-family: 'Trebuchet MS', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.map-card-col {
  margin-bottom: 16px;
}
.map-card {
  min-height: 180px;
  border-radius: 16px;
  border: 1px solid #d8e7f7;
  background: linear-gradient(145deg, #ffffff 0%, #f4fbff 62%, #fff9ea 100%);
  box-shadow: 0 10px 20px rgba(32, 90, 147, 0.12);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.map-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 14px 26px rgba(32, 90, 147, 0.18);
}
.map-card-title {
  font-size: 19px;
  font-weight: 700;
  color: #194772;
  margin-bottom: 12px;
}
.map-card-desc {
  min-height: 56px;
  color: #4e6480;
  line-height: 1.5;
  margin-bottom: 18px;
}
.map-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
