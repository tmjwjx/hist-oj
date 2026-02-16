<template>
  <div class="classroom-detail">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h2>{{ classroomInfo.className || $t('m.Classroom_Detail') }}</h2>
    </div>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
      <el-tab-pane :label="$t('m.Student_Management')" name="students">
        <Students v-if="activeTab === 'students'" :classroom-id="localClassroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Checkin_Management')" name="checkin">
        <Checkin v-if="activeTab === 'checkin'" :classroom-id="localClassroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Random_Pick')" name="randomPick">
        <RandomPick v-if="activeTab === 'randomPick'" :classroom-id="localClassroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Homework_Management')" name="homework">
        <Homework v-if="activeTab === 'homework'" :classroom-id="localClassroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Material_Library')" name="materials">
        <Materials v-if="activeTab === 'materials'" :classroom-id="localClassroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Discussion')" name="discussion">
        <Discussion v-if="activeTab === 'discussion'" :classroom-id="localClassroomId" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import Students from './Students.vue'
import Checkin from './Checkin.vue'
import RandomPick from './RandomPick.vue'
import Homework from './Homework.vue'
import Materials from './Materials.vue'
import Discussion from './Discussion.vue'

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'TeacherClassroomDetail',
  mixins: [teacherAuth],
  components: {
    Students,
    Checkin,
    RandomPick,
    Homework,
    Materials,
    Discussion
  },
  props: {
    // 接受 classroomId prop，优先使用（从 ClassroomAdmin 传递）
    classroomId: [String, Number]
  },
  data() {
    return {
      activeTab: 'students',
      localClassroomId: null, // 本地 classroomId（用于 watch）
      classroomInfo: {}
    }
  },
  mounted() {
    // 优先使用 prop，如果没有才从路由参数获取
    this.localClassroomId = this.classroomId || this.$route.params.classroomId

    // 从路由参数初始化 activeTab
    if (this.$route.query.tab) {
      this.activeTab = this.$route.query.tab
    }

    // 使用 localClassroomId 加载数据
    this.loadClassroomInfo()
  },
  watch: {
    // 监听路由参数变化，动态更新 activeTab
    '$route.query.tab'(newTab) {
      if (newTab && newTab !== this.activeTab) {
        this.activeTab = newTab
      }
    },
    // 监听 prop classroomId 变化
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.localClassroomId = newVal
          this.loadClassroomInfo()
        }
      }
    }
  },
  methods: {
    async loadClassroomInfo() {
      try {
        const res = await this.$store.dispatch('classroom/getClassroomDetail', this.localClassroomId)
        if (res.code === 200) {
          this.classroomInfo = res.data || {}
        }
      } catch (error) {
        console.error('加载班级信息失败', error)
      }
    },
    handleTabClick(tab) {
      // 只有当 tab 真正改变时才更新路由
      if (this.$route.query.tab !== tab.name) {
        this.$router.replace({ query: { tab: tab.name } }).catch(err => {
          // 忽略导航重复错误
          if (err.name !== 'NavigationDuplicated') {
            console.error('导航错误:', err)
          }
        })
      }
    },
    goBack() {
      this.$router.push({ name: 'TeacherDashboard' })
    }
  }
}
</script>

<style scoped>
.classroom-detail {
  padding: 20px;
  background-color: #fff;
  min-height: 100vh;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}
</style>
