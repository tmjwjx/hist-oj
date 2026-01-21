<template>
  <div class="student-homework">
    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-table :data="homeworks" stripe>
      <el-table-column prop="title" :label="$t('m.Homework_Title')" />
      <el-table-column :label="$t('m.Start_Time')">
        <template slot-scope="{ row }">{{ formatTime(row.startTime) }}</template>
      </el-table-column>
      <el-table-column :label="$t('m.End_Time')">
        <template slot-scope="{ row }">{{ formatTime(row.endTime) }}</template>
      </el-table-column>
      <el-table-column :label="$t('m.Status')" width="100">
        <template slot-scope="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Submission_Status')" width="120">
        <template slot-scope="{ row }">
          <el-tag :type="row.isCompleted ? 'success' : 'info'" size="small">
            {{ row.isCompleted ? $t('m.Completed') : $t('m.Not_Completed') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Operation')" width="150">
        <template slot-scope="{ row }">
          <!-- 进行中且未完成：显示"做作业"按钮 -->
          <el-button
            v-if="row.status === 2 && !row.isCompleted"
            size="small"
            type="primary"
            @click="doHomework(row)"
          >
            {{ $t('m.Do_Homework') }}
          </el-button>
          <!-- 已完成或作业已结束：显示"查看结果"按钮（需要教师允许查看） -->
          <el-button
            v-if="row.isCompleted || row.status === 3"
            size="small"
            @click="viewResult(row)"
          >
            {{ $t('m.View_Result') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'Homework',
  mixins: [realtimeSync, studentAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      homeworks: [],
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

          // 检查每个作业的 ID 是否都相同
          const currentIds = this.homeworks.map(h => h.id).sort().join(',')
          const newIds = newHomeworks.map(h => h.id).sort().join(',')

          if (currentIds !== newIds) {
            this.homeworks = newHomeworks
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    doHomework(homework) {
      this.$router.push({
        name: 'StudentHomeworkDetail',
        params: { homeworkId: homework.id }
      })
    },
    viewResult(homework) {
      this.$router.push({
        name: 'StudentHomeworkDetail',
        params: { homeworkId: homework.id }
      })
    },
    getStatusType(status) {
      const map = { 1: 'info', 2: 'warning', 3: 'success' }
      return map[status] || ''
    },
    getStatusText(status) {
      const map = {
        1: this.$t('m.Not_Started'),
        2: this.$t('m.In_Progress'),
        3: this.$t('m.Ended')
      }
      return map[status] || '-'
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    }
  }
}
</script>

<style scoped>
.student-homework {
  padding: 8px;
  min-height: 100vh;
  background: var(--classroom-bg, #f5f7fa);
}
</style>
