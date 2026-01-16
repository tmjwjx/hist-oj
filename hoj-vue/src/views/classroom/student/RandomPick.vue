<template>
  <div class="random-pick-student">
    <!-- 选中学生展示区 -->
    <el-card v-if="latestPick" class="result-card">
      <div slot="header" class="result-header">
        <span>{{ $t('m.Latest_Pick_Result') }}</span>
        <el-tag type="success">{{ $t('m.Just_Picked') }}</el-tag>
      </div>
      <div class="student-display">
        <div class="student-avatar">
          <el-avatar :size="120" :src="latestPick.pickedUser?.avatar">
            <i class="el-icon-user-solid"></i>
          </el-avatar>
        </div>
        <div class="student-details">
          <h2 class="student-name">{{ displayName || '-' }}</h2>
          <div class="detail-row">
            <span class="label">{{ $t('m.Real_Name') }}:</span>
            <span class="value">{{ studentInfo.realName || '-' }}</span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('m.Student_No') }}:</span>
            <span class="value">{{ studentInfo.studentNo || '-' }}</span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('m.Student_Class') }}:</span>
            <span class="value">{{ studentInfo.studentClass || '-' }}</span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('m.Gender') }}:</span>
            <span class="value">{{ studentInfo.gender || '-' }}</span>
          </div>
          <div class="pick-time">
            <i class="el-icon-time"></i>
            {{ formatTime(latestPick.pickTime) }}
          </div>
        </div>
      </div>
    </el-card>

    <!-- 未选中时的占位符 -->
    <el-card v-else class="placeholder-card">
      <div class="placeholder-content">
        <i class="el-icon-microphone"></i>
        <p>{{ $t('m.No_Pick_Yet') }}</p>
      </div>
    </el-card>

    <!-- 历史记录 -->
    <el-card class="history-card">
      <div slot="header" class="history-header">
        <span>{{ $t('m.Pick_History') }}</span>
        <el-button
          type="text"
          icon="el-icon-refresh"
          @click="loadHistory"
          :loading="loadingHistory"
        >
          {{ $t('m.Refresh') }}
        </el-button>
      </div>
      <!-- 移除 v-loading 避免轮询时闪烁 -->
      <el-table :data="history" stripe>
        <el-table-column type="index" :label="$t('m.Index')" width="60" />
        <el-table-column :label="$t('m.Student_Name')" min-width="120">
          <template slot-scope="{ row }">
            {{ getStudentDisplayName(row) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Student_No')" width="120">
          <template slot-scope="{ row }">
            {{ row.studentInfo?.studentNo || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Student_Class')" width="120">
          <template slot-scope="{ row }">
            {{ row.studentInfo?.studentClass || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Pick_Time')" width="180">
          <template slot-scope="{ row }">
            {{ formatTime(row.pickTime) }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loadingHistory && history.length === 0" :description="$t('m.No_History_Yet')" />
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'RandomPickStudent',
  mixins: [realtimeSync, studentAuth],
  props: {
    classroomId: {
      type: [String, Number],
      required: true
    }
  },
  data() {
    return {
      latestPick: null,
      history: [],
      loadingHistory: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadHistory',
        immediate: true
      }
    }
  },
  computed: {
    displayName() {
      if (this.latestPick?.studentInfo?.realName) {
        return this.latestPick.studentInfo.realName
      }
      return this.latestPick?.pickedUser?.nickname
    },
    studentInfo() {
      if (this.latestPick?.studentInfo) {
        return this.latestPick.studentInfo
      }
      return {}
    }
  },
  mounted() {
    // realtimeSync mixin 会自动启动轮询
  },
  methods: {
    async loadHistory() {
      // 避免重复请求
      if (this.loadingHistory) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.history.length === 0
      if (isFirstLoad) {
        this.loadingHistory = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getPickHistory', {
          classroomId: this.classroomId,
          limit: 20
        })
        if (res.code === 200) {
          const newHistory = res.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.history)
          const newDataString = JSON.stringify(newHistory)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.history = newHistory
            // 最新的记录是第一条
            if (this.history.length > 0) {
              this.latestPick = this.history[0]
            }
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          console.error('加载历史记录失败:', error)
        }
      } finally {
        if (isFirstLoad) {
          this.loadingHistory = false
        }
      }
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    },
    getStudentDisplayName(row) {
      if (row.studentInfo?.realName) {
        return row.studentInfo.realName
      }
      return row.pickedUser?.nickname || '-'
    }
  }
}
</script>

<style scoped>
.random-pick-student {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.result-card {
  border: 2px solid #67c23a;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
  color: #67c23a;
}

.student-display {
  display: flex;
  gap: 40px;
  align-items: center;
  padding: 20px 0;
}

.student-avatar {
  flex-shrink: 0;
}

.student-details {
  flex: 1;
}

.student-name {
  font-size: 32px;
  color: #303133;
  margin: 0 0 20px 0;
  font-weight: 600;
}

.detail-row {
  display: flex;
  margin-bottom: 12px;
  font-size: 16px;
}

.detail-row .label {
  color: #909399;
  width: 100px;
  font-weight: 500;
}

.detail-row .value {
  color: #303133;
  font-weight: 600;
}

.pick-time {
  margin-top: 20px;
  color: #909399;
  font-size: 14px;
}

.pick-time i {
  margin-right: 5px;
}

.placeholder-card {
  text-align: center;
  padding: 60px 20px;
}

.placeholder-content i {
  font-size: 80px;
  color: #c0c4cc;
  margin-bottom: 20px;
}

.placeholder-content p {
  font-size: 16px;
  color: #909399;
  margin: 0;
}

.history-card {
  margin-top: 20px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .student-display {
    flex-direction: column;
    text-align: center;
    gap: 20px;
  }

  .detail-row {
    justify-content: center;
  }

  .detail-row .label {
    width: auto;
  }
}
</style>
