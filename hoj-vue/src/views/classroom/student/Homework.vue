<template>
  <div class="student-homework">
    <el-table :data="homeworks" v-loading="loading" stripe>
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

export default {
  name: 'Homework',
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      homeworks: []
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
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getHomeworkList', this.classroomId)
        if (res.code === 200) {
          this.homeworks = res.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
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
  padding: 20px;
}
</style>
