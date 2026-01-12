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
      <el-table :data="history" v-loading="loadingHistory" stripe>
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

export default {
  name: 'RandomPickStudent',
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
      pollingTimer: null
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
    this.loadHistory()
    // 启动轮询，每5秒刷新一次
    this.startPolling()
  },
  beforeDestroy() {
    this.stopPolling()
  },
  methods: {
    async loadHistory() {
      this.loadingHistory = true
      try {
        const res = await this.$store.dispatch('classroom/getPickHistory', {
          classroomId: this.classroomId,
          limit: 20
        })
        if (res.code === 200) {
          this.history = res.data || []
          // 最新的记录是第一条
          if (this.history.length > 0) {
            this.latestPick = this.history[0]
          }
        }
      } catch (error) {
        console.error('加载历史记录失败:', error)
      } finally {
        this.loadingHistory = false
      }
    },
    startPolling() {
      // 每5秒轮询一次
      this.pollingTimer = setInterval(() => {
        this.loadHistory()
      }, 5000)
    },
    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
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
