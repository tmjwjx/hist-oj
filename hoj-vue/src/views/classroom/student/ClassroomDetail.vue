<template>
  <div class="classroom-detail">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h2>{{ classroomInfo.className || $t('m.Classroom_Detail') }}</h2>
      <el-button type="primary" icon="el-icon-user" @click="goToMyInfo" style="margin-left: auto;">
        我的信息
      </el-button>
    </div>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
      <el-tab-pane :label="$t('m.Homework_List')" name="homework">
        <Homework v-if="activeTab === 'homework'" :classroom-id="classroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Checkin')" name="checkin">
        <Checkin v-if="activeTab === 'checkin'" :classroom-id="classroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Random_Pick')" name="randomPick">
        <RandomPick v-if="activeTab === 'randomPick'" :classroom-id="classroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Materials')" name="materials">
        <Materials v-if="activeTab === 'materials'" :classroom-id="classroomId" />
      </el-tab-pane>
      <el-tab-pane :label="$t('m.Discussion')" name="discussion">
        <Discussion v-if="activeTab === 'discussion'" :classroom-id="classroomId" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import Homework from './Homework.vue'
import Checkin from './Checkin.vue'
import RandomPick from './RandomPick.vue'
import Materials from './Materials.vue'
import Discussion from './Discussion.vue'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'StudentClassroomDetail',
  components: {
    Homework,
    Checkin,
    RandomPick,
    Materials,
    Discussion
  },
  mixins: [studentAuth],
  data() {
    return {
      activeTab: 'homework',
      classroomId: null,
      classroomInfo: {}
    }
  },
  mounted() {
    this.classroomId = this.$route.params.classroomId
    if (this.$route.query.tab) {
      this.activeTab = this.$route.query.tab
    }
    this.loadClassroomInfo()
  },
  methods: {
    async loadClassroomInfo() {
      try {
        const res = await this.$store.dispatch('classroom/getClassroomDetail', this.classroomId)
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
      this.$router.push({ name: 'StudentDashboard' })
    },
    goToMyInfo() {
      this.$router.push({
        name: 'StudentMyInfo',
        params: { classroomId: this.classroomId }
      })
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
