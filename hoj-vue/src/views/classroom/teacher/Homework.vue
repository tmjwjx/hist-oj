<template>
  <div class="homework-panel">
    <div class="header">
      <h3>{{ $t('m.Homework_Management') }}</h3>
      <el-button type="primary" icon="el-icon-plus" @click="goToCreate">
        {{ $t('m.Create_Homework') }}
      </el-button>
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁，改用 v-if -->
    <div v-if="!loading">
      <el-empty v-if="homeworks.length === 0" :description="$t('m.No_Homework_Available')" />

      <div v-else class="homework-list">
        <el-card v-for="hw in homeworks" :key="hw.id" class="homework-card">
          <div slot="header">
            <span>{{ hw.title }}</span>
            <div style="float: right">
              <el-button size="small" @click="viewHomework(hw)">
                {{ $t('m.View_Detail') }}
              </el-button>
              <el-button size="small" type="danger" @click="deleteHomework(hw)" style="margin-left: 10px">
                {{ $t('m.Delete') }}
              </el-button>
            </div>
          </div>
          <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(hw.startTime) }}</p>
          <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(hw.endTime) }}</p>
          <el-tag :type="hw.status === 2 ? 'warning' : hw.status === 3 ? 'info' : 'success'">
            {{ hw.status === 1 ? $t('m.Not_Started') : hw.status === 2 ? $t('m.In_Progress') : $t('m.Ended') }}
          </el-tag>
        </el-card>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <el-dialog
      :title="$t('m.Confirm_Delete')"
      :visible.sync="showDeleteDialog"
      width="400px"
    >
      <p>{{ $t('m.Confirm_Delete_Homework') }}</p>
      <span slot="footer">
        <el-button @click="showDeleteDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="danger" @click="confirmDelete" :loading="deleting">
          {{ $t('m.Confirm') }}
        </el-button>
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
.homework-panel {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  font-size: 20px;
  color: #409EFF;
  margin: 0;
}

.homework-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.homework-card {
  /* 移除 transition 避免轮询时闪烁 */
}

.homework-card:hover {
  /* 移除 transform 避免轮询时闪烁 */
}
</style>
