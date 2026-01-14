<template>
  <div class="classroom-index">
    <div class="header">
      <h1>{{ $t('m.Classroom') }}</h1>
    </div>

    <!-- 有角色的情况 -->
    <div class="content" v-if="hasRole">
      <el-card v-if="isTeacher" shadow="hover" @click.native="goToTeacherDashboard" class="card">
        <i class="el-icon-s-custom"></i>
        <h2>{{ $t('m.Teacher_Workspace') }}</h2>
        <p>{{ $t('m.Manage_your_classrooms') }}</p>
      </el-card>
      <el-card v-if="isStudent" shadow="hover" @click.native="goToStudentDashboard" class="card">
        <i class="el-icon-reading"></i>
        <h2>{{ $t('m.My_Classrooms') }}</h2>
        <p>{{ $t('m.View_joined_classrooms') }}</p>
      </el-card>
    </div>

    <!-- 无角色的情况 -->
    <el-empty v-else class="empty-state" :description="$t('m.No_Classroom_Role_Description')">
      <template #image>
        <i class="el-icon-info" style="font-size: 100px; color: #909399;"></i>
      </template>
      <div class="role-info">
        <p>{{ $t('m.Classroom_Not_Enabled') }}</p>
        <el-alert
          v-if="isAdminRole"
          :title="$t('m.Admin_Enable_Classroom')"
          type="info"
          :closable="false"
          show-icon
          style="margin-top: 20px; max-width: 600px;"
        >
          <template>
            <div>{{ $t('m.Admin_Enable_Classroom_Tip1') }}</div>
            <div>{{ $t('m.Admin_Enable_Classroom_Tip2') }}</div>
          </template>
        </el-alert>
        <el-alert
          v-else
          :title="$t('m.Contact_Admin_For_Classroom')"
          type="warning"
          :closable="false"
          show-icon
          style="margin-top: 20px; max-width: 600px;"
        />
      </div>
    </el-empty>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'ClassroomIndex',
  data() {
    return {
      debug: false,
      pollingTimer: null
    }
  },
  computed: {
    ...mapGetters(['isTeacher', 'isStudent', 'isAdminRole']),
    userRoles() {
      return this.$store.state.classroom.userRoles
    },
    hasRole() {
      return this.isTeacher || this.isStudent
    }
  },
  mounted() {
    // 每次进入页面都重新加载角色
    this.$store.dispatch('classroom/loadUserRoles')

    // 启动轮询检测用户角色
    this.startPolling()
  },
  beforeDestroy() {
    // 组件销毁前清除定时器
    this.stopPolling()
  },
  methods: {
    goToTeacherDashboard() {
      this.$router.push({ name: 'TeacherDashboard' })
    },
    goToStudentDashboard() {
      this.$router.push({ name: 'StudentDashboard' })
    },
    // 启动轮询
    startPolling() {
      this.pollingTimer = setInterval(() => {
        this.$store.dispatch('classroom/loadUserRoles')
      }, 500)
    },
    // 停止轮询
    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }
    }
  }
}
</script>

<style scoped>
.classroom-index {
  padding: 20px;
}

.header {
  text-align: center;
  margin-bottom: 40px;
}

.header h1 {
  font-size: 36px;
  color: #409EFF;
}

.content {
  display: flex;
  justify-content: center;
  gap: 30px;
  flex-wrap: wrap;
}

.card {
  width: 300px;
  height: 200px;
  text-align: center;
  cursor: pointer;
  /* 移除 transition 避免轮询时闪烁 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.card:hover {
  /* 移除 transform 避免轮询时闪烁 */
  box-shadow: 0 4px 20px rgba(64, 158, 255, 0.3);
}

.card i {
  font-size: 60px;
  color: #409EFF;
  margin-bottom: 15px;
}

.card h2 {
  font-size: 24px;
  margin: 10px 0;
}

.card p {
  color: #909399;
}

.empty-state {
  margin-top: 60px;
}

.role-info {
  text-align: center;
}

.role-info p {
  font-size: 16px;
  color: #606266;
  margin-bottom: 20px;
}
</style>
