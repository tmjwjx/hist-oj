<template>
  <div class="admin-exam-monitoring">
    <div class="page-header">
      <el-button icon="el-icon-back" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h2>作业考试监控</h2>
    </div>

    <!-- 复用教师端的监控组件，通过 props 传递参数 -->
    <TeacherExamMonitoring
      :classroomId="classroomId"
      :homeworkId="homeworkId"
      :hideBackButton="true"
    />
  </div>
</template>

<script>
import TeacherExamMonitoring from '@/views/classroom/teacher/ExamMonitoring.vue'

export default {
  name: 'AdminExamMonitoring',
  components: {
    TeacherExamMonitoring
  },
  computed: {
    classroomId() {
      return this.$route.query.classroomId
    },
    homeworkId() {
      return this.$route.query.homeworkId
    }
  },
  methods: {
    goBack() {
      // 返回到管理员作业详情页，而不是班级列表
      const classroomId = this.$route.query.classroomId
      const homeworkId = this.$route.query.homeworkId

      this.$router.push({
        name: 'admin-classroom',
        query: {
          classroomId: classroomId,
          homeworkId: homeworkId,
          activeTab: 'homework'
        }
      })
    }
  }
}
</script>

<style scoped>
.admin-exam-monitoring {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}
</style>
