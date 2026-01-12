<template>
  <div class="classroom-detail">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h2>{{ classroomInfo.className || $t('m.Classroom_Detail') }}</h2>
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

export default {
  name: 'StudentClassroomDetail',
  components: {
    Homework,
    Checkin,
    RandomPick,
    Materials,
    Discussion
  },
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
      this.$router.replace({ query: { tab: tab.name } })
    },
    goBack() {
      this.$router.push({ name: 'StudentDashboard' })
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
