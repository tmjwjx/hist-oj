<template>
  <div class="admin-homework-analysis">
    <div class="page-header">
      <el-button icon="el-icon-back" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h2>作业学情分析</h2>
    </div>

    <!-- 复用教师端的分析组件，通过 props 传递参数 -->
    <TeacherHomeworkAnalysis
      :classroomId="classroomId"
      :homeworkId="homeworkId"
      :hideBackButton="true"
    />
  </div>
</template>

<script>
import TeacherHomeworkAnalysis from '@/views/classroom/teacher/HomeworkAnalysis.vue'

export default {
  name: 'AdminHomeworkAnalysis',
  components: {
    TeacherHomeworkAnalysis
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
.admin-homework-analysis {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}
</style>
