<template>
  <div class="homework-panel classroom-theme">
    <div class="header">
      <h3>{{ $t('m.Homework_Management') }}</h3>
      <button class="classroom-btn classroom-btn-primary" @click="goToCreate">
        <i class="el-icon-plus"></i>
        <span>{{ $t('m.Create_Homework') }}</span>
      </button>
    </div>

    <div v-if="!loading">
      <div v-if="homeworks.length === 0" class="classroom-empty">
        <i class="el-icon-document classroom-empty-icon"></i>
        <div class="classroom-empty-text">{{ $t('m.No_Homework_Available') }}</div>
        <div class="classroom-empty-hint">点击上方按钮创建您的第一个作业吧！</div>
      </div>

      <div v-else class="homework-list">
        <div v-for="hw in homeworks" :key="hw.id" class="homework-card classroom-card classroom-fade-in">
          <div class="homework-card-header">
            <h4 class="homework-title">{{ hw.title }}</h4>
            <div class="homework-actions">
              <button class="classroom-btn classroom-btn-primary" @click="viewHomework(hw)">
                <i class="el-icon-view"></i>
                <span>查看详情</span>
              </button>
              <button class="classroom-btn classroom-btn-danger" @click="deleteHomework(hw)">
                <i class="el-icon-delete"></i>
                <span>{{ $t('m.Delete') }}</span>
              </button>
            </div>
          </div>
          <div class="homework-body">
            <div class="homework-info">
              <div class="info-row">
                <i class="el-icon-time info-icon"></i>
                <span class="info-label">{{ $t('m.Start_Time') }}:</span>
                <span class="info-value">{{ formatTime(hw.startTime) }}</span>
              </div>
              <div class="info-row">
                <i class="el-icon-alarm-clock info-icon"></i>
                <span class="info-label">{{ $t('m.End_Time') }}:</span>
                <span class="info-value">{{ formatTime(hw.endTime) }}</span>
              </div>
            </div>
            <div class="homework-status">
              <span
                :class="hw.status === 2 ? 'classroom-tag classroom-tag-warning' : hw.status === 3 ? 'classroom-tag' : 'classroom-tag classroom-tag-success'"
              >
                {{ hw.status === 1 ? $t('m.Not_Started') : hw.status === 2 ? $t('m.In_Progress') : $t('m.Ended') }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <el-dialog
      title="删除作业"
      :visible.sync="showDeleteDialog"
      width="450px"
      custom-class="classroom-dialog"
    >
      <div class="delete-confirm-content" v-if="currentHomework">
        <i class="el-icon-warning"></i>
        <p>确定要删除作业 "{{ currentHomework.title }}" 吗？</p>
      </div>
      <span slot="footer">
        <el-button @click="showDeleteDialog = false" class="classroom-btn classroom-btn-secondary">{{ $t('m.Cancel') }}</el-button>
        <el-button type="danger" @click="confirmDelete" :loading="deleting" class="classroom-btn classroom-btn-danger">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'Homework',
  mixins: [realtimeSync, teacherAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      homeworks: [],
      showDeleteDialog: false,
      deleting: false,
      currentHomework: null,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadHomeworks',
        immediate: true
      }
    }
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadHomeworks()
        }
      }
    }
  },
  mounted() {
    // mounted 时也会通过 watch 触发加载
  },
  methods: {
    async loadHomeworks() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.homeworks.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getHomeworkList', this.classroomId)
        if (res.code === 200) {
          const newHomeworks = res.data || []

          // 检查数量是否变化
          if (newHomeworks.length !== this.homeworks.length) {
            this.homeworks = newHomeworks
            return
          }

          // 检查每个作业的 ID 是否都相同(避免深度对比整个对象)
          const currentIds = this.homeworks.map(h => h.id).sort().join(',')
          const newIds = newHomeworks.map(h => h.id).sort().join(',')

          if (currentIds !== newIds) {
            // ID 列表不同，说明有作业增删，需要更新
            this.homeworks = newHomeworks
          }
          // 如果 ID 列表相同，不更新数据，避免闪烁
        }
      } catch (error) {
        // 只在首次加载失败时提示错误
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    goToCreate() {
      this.$router.push({
        path: `/classroom/teacher/homework/create/${this.classroomId}`
      })
    },
    viewHomework(homework) {
      this.$router.push({
        name: 'TeacherHomeworkDetail',
        params: {
          classroomId: this.classroomId,
          homeworkId: homework.id
        }
      })
    },
    deleteHomework(homework) {
      this.currentHomework = homework
      this.showDeleteDialog = true
    },
    async confirmDelete() {
      if (!this.currentHomework) return

      this.deleting = true
      try {
        const res = await this.$store.dispatch('classroom/deleteHomework', this.currentHomework.id)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Delete_Success'))
          this.showDeleteDialog = false
          this.currentHomework = null
          // 重新加载作业列表
          await this.loadHomeworks()
        } else {
          this.$message.error(res.message || this.$t('m.Delete_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Delete_Failed'))
      } finally {
        this.deleting = false
      }
    },
    formatTime(time) {
      return time ? moment(time).format('YYYY-MM-DD HH:mm') : '-'
    }
  }
}
</script>

<style scoped>
@import '../classroom-theme.css';

.homework-panel {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header h3 {
  font-size: 22px;
  color: var(--classroom-text);
  margin: 0;
  font-weight: 700;
}

.homework-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 24px;
}

.homework-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
  transition: all 0.3s ease;
  border: 1px solid var(--classroom-border);
}

.homework-card:hover {
  box-shadow: 0 8px 24px rgba(74, 144, 226, 0.15);
  transform: translateY(-4px);
}

.homework-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20px;
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
}

.homework-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--classroom-text);
  flex: 1;
  padding-right: 16px;
}

.homework-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.homework-actions .classroom-btn {
  padding: 8px 16px;
  font-size: 13px;
}

.homework-body {
  padding: 20px;
}

.homework-info {
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-icon {
  font-size: 18px;
  color: var(--classroom-primary);
}

.info-label {
  font-weight: 500;
  color: var(--classroom-text-secondary);
  min-width: 80px;
}

.info-value {
  color: var(--classroom-text);
  font-weight: 500;
}

.homework-status {
  display: flex;
  justify-content: flex-start;
}

.delete-confirm-content {
  text-align: center;
  padding: 20px;
}

.delete-confirm-content i {
  font-size: 48px;
  color: var(--classroom-warning);
  margin-bottom: 16px;
  display: block;
}

.delete-confirm-content p {
  font-size: 16px;
  color: var(--classroom-text);
  margin: 0;
}

/* 对话框样式 */
.classroom-dialog .el-dialog__header {
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
}

.classroom-dialog .el-dialog__title {
  color: var(--classroom-text);
  font-weight: 600;
}

/* 响应式 */
@media (max-width: 768px) {
  .homework-list {
    grid-template-columns: 1fr;
  }

  .homework-card-header {
    flex-direction: column;
    gap: 16px;
  }

  .homework-actions {
    width: 100%;
    flex-direction: column;
  }

  .homework-actions .classroom-btn {
    width: 100%;
  }

  .header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .header .classroom-btn {
    width: 100%;
  }
}
</style>
