<template>
  <div class="random-pick-container">
    <!-- 顶部操作区 -->
    <el-card class="action-card">
      <div class="action-section">
        <div class="action-info">
          <h3>{{ $t('m.Random_Pick') }}</h3>
          <p class="student-count">{{ $t('m.Total_Students') }}: {{ students.length }}</p>
        </div>
        <el-button
          type="primary"
          size="large"
          icon="el-icon-microphone"
          @click="handleRandomPick"
          :loading="picking"
          :disabled="students.length === 0"
        >
          {{ picking ? $t('m.Picking') : $t('m.Start_Pick') }}
        </el-button>
      </div>
    </el-card>

    <!-- 选中学生展示区 -->
    <el-card v-if="pickedStudent" class="result-card">
      <div slot="header" class="result-header">
        <span>{{ $t('m.Picked_Student') }}</span>
        <el-tag type="success">{{ $t('m.Just_Picked') }}</el-tag>
      </div>
      <div class="student-display">
        <div class="student-avatar">
          <el-avatar :size="120" :src="pickedStudent.pickedUser?.avatar">
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
            {{ formatTime(pickedStudent.pickTime) }}
          </div>
        </div>
      </div>
    </el-card>

    <!-- 未选中时的占位符 -->
    <el-card v-else class="placeholder-card">
      <div class="placeholder-content">
        <i class="el-icon-microphone"></i>
        <p>{{ students.length > 0 ? $t('m.Click_To_Start_Pick') : $t('m.No_Students_In_Classroom') }}</p>
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

export default {
  name: 'RandomPick',
  mixins: [realtimeSync],
  props: {
    classroomId: {
      type: [String, Number],
      required: true
    }
  },
  data() {
    return {
      students: [],
      pickedStudent: null,
      picking: false,
      history: [],
      loadingHistory: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadStudents',
        immediate: true
      }
    }
  },
  computed: {
    // 优先显示学生信息中的真名，如果没有则显示用户昵称
    displayName() {
      if (this.pickedStudent?.studentInfo?.realName) {
        return this.pickedStudent.studentInfo.realName
      }
      return this.pickedStudent?.pickedUser?.nickname
    },
    // 学生信息对象，优先使用studentInfo
    studentInfo() {
      if (this.pickedStudent?.studentInfo) {
        return this.pickedStudent.studentInfo
      }
      return {}
    }
  },
  mounted() {
    this.loadStudents()
    this.loadHistory()
  },
  methods: {
    async loadStudents() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.students.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getClassroomStudents', this.classroomId)
        if (res.code === 200) {
          const newStudents = res.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.students)
          const newDataString = JSON.stringify(newStudents)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.students = newStudents
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          console.error('加载学生列表失败:', error)
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    async loadHistory() {
      this.loadingHistory = true
      try {
        const res = await this.$store.dispatch('classroom/getPickHistory', {
          classroomId: this.classroomId,
          limit: 20
        })
        if (res.code === 200) {
          this.history = res.data || []
        }
      } catch (error) {
        console.error('加载历史记录失败:', error)
      } finally {
        this.loadingHistory = false
      }
    },
    async handleRandomPick() {
      if (this.students.length === 0) {
        this.$message.warning(this.$t('m.No_Students_In_Classroom'))
        return
      }

      this.picking = true
      try {
        const res = await this.$store.dispatch('classroom/randomPick', this.classroomId)
        if (res.code === 200) {
          this.pickedStudent = res.data
          // 获取显示名称：优先使用学生真名，其次使用用户昵称
          const displayName = res.data.studentInfo?.realName || res.data.pickedUser?.nickname || '-'
          this.$notify({
            title: this.$t('m.Random_Pick'),
            message: this.$t('m.Picked_Student') + ': ' + displayName,
            type: 'success',
            duration: 5000
          })
          // 刷新历史记录
          this.loadHistory()
        } else {
          this.$message.error(res.message || this.$t('m.Pick_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Pick_Failed'))
      } finally {
        this.picking = false
      }
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    },
    getStudentDisplayName(row) {
      // 优先显示学生真名，其次显示昵称
      if (row.studentInfo?.realName) {
        return row.studentInfo.realName
      }
      return row.pickedUser?.nickname || '-'
    }
  }
}
</script>

<style scoped>
.random-pick-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.action-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.action-card ::v-deep .el-card__body {
  padding: 30px;
}

.action-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-info h3 {
  color: #fff;
  margin: 0 0 10px 0;
  font-size: 24px;
}

.student-count {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  margin: 0;
}

.action-section .el-button {
  background: #fff;
  color: #667eea;
  border: none;
  font-weight: 600;
  padding: 15px 40px;
  font-size: 16px;
}

.action-section .el-button:hover {
  background: #f0f0f0;
}

.action-section .el-button:disabled {
  background: rgba(255, 255, 255, 0.5);
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
  .action-section {
    flex-direction: column;
    gap: 20px;
  }

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
