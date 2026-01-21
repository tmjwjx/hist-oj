<template>
  <div class="random-pick-container classroom-theme">
    <!-- 顶部操作区 -->
    <el-card class="action-card classroom-card">
      <div class="action-section">
        <div class="action-info">
          <h3>{{ $t('m.Random_Pick') }}</h3>
          <p class="student-count">{{ $t('m.Total_Students') }}: {{ students.length }}</p>
        </div>
        <button
          class="classroom-btn classroom-btn-primary"
          @click="handleRandomPick"
          :disabled="students.length === 0"
        >
          <i class="el-icon-microphone"></i>
          <span>{{ picking ? $t('m.Picking') : $t('m.Start_Pick') }}</span>
        </button>
      </div>
    </el-card>

    <!-- 选中学生展示区 -->
    <el-card v-if="pickedStudent" class="result-card classroom-card classroom-fade-in">
      <div slot="header" class="result-header">
        <span>{{ $t('m.Picked_Student') }}</span>
        <span class="classroom-tag classroom-tag-success">{{ $t('m.Just_Picked') }}</span>
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
    <el-card v-else class="placeholder-card classroom-card">
      <div class="placeholder-content">
        <i class="el-icon-microphone"></i>
        <p>{{ students.length > 0 ? $t('m.Click_To_Start_Pick') : $t('m.No_Students_In_Classroom') }}</p>
      </div>
    </el-card>

    <!-- 历史记录 -->
    <el-card class="history-card classroom-card">
      <div slot="header" class="history-header">
        <span>{{ $t('m.Pick_History') }}</span>
        <button class="classroom-btn classroom-btn-secondary" @click="loadHistory">
          <i class="el-icon-refresh"></i>
          <span>{{ $t('m.Refresh') }}</span>
        </button>
      </div>
      <!-- 移除 v-loading 避免轮询时闪烁 -->
      <el-table :data="history" stripe class="classroom-table">
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

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'RandomPick',
  mixins: [realtimeSync, teacherAuth],
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
@import '../classroom-theme.css';

.random-pick-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
}

.action-card {
  background: #4A90E2;
  border: none;
}

.action-card ::v-deep .el-card__body {
  padding: 32px;
}

.action-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.action-info h3 {
  color: #fff;
  margin: 0 0 12px 0;
  font-size: 28px;
  font-weight: 700;
}

.student-count {
  color: rgba(255, 255, 255, 0.95);
  font-size: 15px;
  margin: 0;
}

.action-section .classroom-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  color: var(--classroom-primary);
  border: none;
  font-weight: 600;
  padding: 14px 36px;
  font-size: 16px;
  white-space: nowrap;
  flex-shrink: 0;
}

.action-section .classroom-btn:hover {
  background: #f0f0f0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.action-section .classroom-btn:disabled {
  background: rgba(255, 255, 255, 0.5);
  cursor: not-allowed;
}

.result-card {
  border: 2px solid var(--classroom-success);
}

.result-card ::v-deep .el-card__header {
  padding: 20px 24px;
  background: #E8F5E9;
}

.result-card ::v-deep .el-card__body {
  padding: 24px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
  color: var(--classroom-success);
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
  font-size: 36px;
  color: #4A90E2;
  margin: 0 0 24px 0;
  font-weight: 700;
}

.detail-row {
  display: flex;
  margin-bottom: 16px;
  font-size: 15px;
  align-items: center;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row .label {
  color: var(--classroom-text-secondary);
  min-width: 100px;
  font-weight: 500;
  padding-right: 12px;
}

.detail-row .value {
  color: var(--classroom-text);
  font-weight: 600;
}

.pick-time {
  margin-top: 24px;
  color: var(--classroom-text-secondary);
  font-size: 14px;
  padding: 10px 16px;
  background: var(--classroom-bg);
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.pick-time i {
  font-size: 16px;
  color: var(--classroom-primary);
}

.placeholder-card {
  text-align: center;
  padding: 80px 20px;
}

.placeholder-content i {
  font-size: 100px;
  color: var(--classroom-primary-lighter);
  margin-bottom: 24px;
  display: block;
}

.placeholder-content p {
  font-size: 16px;
  color: var(--classroom-text-secondary);
  margin: 0;
}

.history-card {
  margin-top: 0;
}

.history-card ::v-deep .el-card__header {
  padding: 20px 24px;
}

.history-card ::v-deep .el-card__body {
  padding: 20px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  color: var(--classroom-text);
  gap: 16px;
}

.history-header .classroom-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 14px;
  white-space: nowrap;
}

/* 表格样式 */
.random-pick-container ::v-deep .el-table {
  border-radius: 8px;
  overflow: hidden;
  font-size: 14px;
  background: white;
}

.random-pick-container ::v-deep .el-table th {
  background: #E3F2FD;
  color: var(--classroom-text);
  font-weight: 600;
  border-bottom: 2px solid var(--classroom-primary);
  padding: 16px 12px;
  text-align: left;
}

.random-pick-container ::v-deep .el-table td {
  border-bottom: 1px solid var(--classroom-border);
  padding: 14px 12px;
  font-size: 14px;
  transition: background-color 0.2s ease;
}

.random-pick-container ::v-deep .el-table--striped .el-table__body tr.el-table__row--striped td {
  background: #fafafa;
}

.random-pick-container ::v-deep .el-table__body tr:hover > td {
  background: #E3F2FD !important;
}

.random-pick-container ::v-deep .el-table__body tr.el-table__row--striped:hover > td {
  background: #E3F2FD !important;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .action-section {
    flex-direction: column;
    gap: 20px;
    text-align: center;
  }

  .action-section .classroom-btn {
    width: 100%;
    justify-content: center;
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
    min-width: auto;
    width: auto;
  }

  .history-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .history-header .classroom-btn {
    width: 100%;
  }
}
</style>
