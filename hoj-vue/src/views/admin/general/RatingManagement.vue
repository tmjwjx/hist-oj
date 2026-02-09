<template>
  <div class="rating-management">
    <el-card shadow="never">
      <div slot="header" class="card-header">
        <span class="title">Rating 管理中心</span>
        <el-tag type="warning" size="small">超级管理员功能</el-tag>
      </div>

      <!-- Tab标签页 -->
      <el-tabs v-model="activeTab" type="border-card" @tab-click="handleTabClick">
        <!-- Tab 1: 个人调整 -->
        <el-tab-pane label="👤 个人调整" name="personal">
          <personal-adjust />
        </el-tab-pane>

        <!-- Tab 2: 比赛Skip管理 -->
        <el-tab-pane label="🏆 比赛Skip管理" name="skip">
          <contest-skip />
        </el-tab-pane>

        <!-- Tab 3: 操作日志 -->
        <el-tab-pane label="📋 操作日志" name="logs">
          <operation-logs />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script>
import PersonalAdjust from './rating/PersonalAdjust.vue'
import ContestSkip from './rating/ContestSkip.vue'
import OperationLogs from './rating/OperationLogs.vue'

export default {
  name: 'RatingManagement',
  components: {
    PersonalAdjust,
    ContestSkip,
    OperationLogs
  },
  data() {
    return {
      activeTab: 'personal'
    }
  },
  mounted() {
    // 从URL参数中恢复tab状态
    const tabFromUrl = this.$route.query.tab
    const contestIdFromUrl = this.$route.query.contestId

    // 如果有contestId但没有tab，说明是在skip tab中刷新的
    if (contestIdFromUrl && !tabFromUrl) {
      this.activeTab = 'skip'
    } else if (tabFromUrl && ['personal', 'skip', 'logs'].includes(tabFromUrl)) {
      this.activeTab = tabFromUrl
    }
  },
  methods: {
    handleTabClick(tab) {
      // 当tab切换时，更新URL参数（保留其他参数，如contestId）
      const query = { ...this.$route.query, tab: tab.name }
      // 如果切换回personal或logs，清除contestId
      if (tab.name !== 'skip') {
        delete query.contestId
      }
      this.$router.replace({ query }).catch(err => {
        // 忽略路由导航重复的错误
        if (err.name !== 'NavigationDuplicated') {
          console.error('更新URL参数失败:', err)
        }
      })
    }
  }
}
</script>

<style scoped>
.rating-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header .title {
  font-size: 18px;
  font-weight: bold;
}
</style>
