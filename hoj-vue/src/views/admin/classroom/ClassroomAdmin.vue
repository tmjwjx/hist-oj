<template>
  <div class="classroom-admin">
    <!-- 列表视图 -->
    <div v-if="!selectedClassroomId" class="classroom-list-view">
      <el-card>
        <div slot="header" class="header">
          <span class="title">{{ $t('m.Classroom_Management') }}</span>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="loadClassrooms">
            {{ $t('m.Refresh') }}
          </el-button>
        </div>

        <el-table :data="classrooms" v-loading="loading" stripe border>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="className" :label="$t('m.Classroom_Name')" min-width="200" />
          <el-table-column prop="classBelong" :label="$t('m.Classroom_Belong')" width="150" />
          <el-table-column :label="$t('m.Teacher_Name')" width="150">
            <template slot-scope="{ row }">
              <UserName :username="row.teacher ? row.teacher.username : row.teacherId || '-'" />
            </template>
          </el-table-column>
          <el-table-column prop="classCode" :label="$t('m.Classroom_Code')" width="120" />
          <el-table-column prop="createTime" :label="$t('m.Create_Time')" width="180">
            <template slot-scope="{ row }">{{ formatTime(row.createTime) }}</template>
          </el-table-column>
          <el-table-column :label="$t('m.Operation')" width="150" fixed="right">
            <template slot-scope="{ row }">
              <el-button size="mini" type="primary" @click="enterClassroom(row)">
                进入班级
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!loading && classrooms.length === 0" :description="$t('m.No_Data')">
          <p class="empty-tip">{{ $t('m.Classroom_Admin_Empty_Tip') }}</p>
        </el-empty>
      </el-card>
    </div>

    <!-- 班级详情视图 - 复用教师端的界面 -->
    <div v-else class="classroom-detail-view">
      <div class="page-header">
        <el-button icon="el-icon-arrow-left" @click="goBackToList">{{ $t('m.Back') }}</el-button>
        <h2>{{ classroomInfo.className || $t('m.Classroom_Detail') }}</h2>
      </div>
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="$t('m.Student_Management')" name="students">
          <Students v-if="activeTab === 'students'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Checkin_Management')" name="checkin">
          <Checkin v-if="activeTab === 'checkin'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Random_Pick')" name="randomPick">
          <RandomPick v-if="activeTab === 'randomPick'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Homework_Management')" name="homework">
          <Homework v-if="activeTab === 'homework'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Material_Library')" name="materials">
          <Materials v-if="activeTab === 'materials'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Discussion')" name="discussion">
          <Discussion v-if="activeTab === 'discussion'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import Students from '@/views/classroom/teacher/Students.vue'
import Checkin from '@/views/classroom/teacher/Checkin.vue'
import RandomPick from '@/views/classroom/teacher/RandomPick.vue'
import Homework from '@/views/classroom/teacher/Homework.vue'
import Materials from '@/views/classroom/teacher/Materials.vue'
import Discussion from '@/views/classroom/teacher/Discussion.vue'
import UserName from '@/components/oj/common/UserName.vue'

export default {
  name: 'ClassroomAdmin',
  components: {
    UserName,
    Students,
    Checkin,
    RandomPick,
    Homework,
    Materials,
    Discussion
  },
  data() {
    return {
      loading: false,
      classrooms: [],
      selectedClassroomId: null,
      activeTab: 'students',
      classroomInfo: {}
    }
  },
  mounted() {
    this.loadClassrooms()
    // 检查 URL 参数,如果指定了班级ID,直接进入详情
    if (this.$route.query.classroomId) {
      this.enterClassroom({ id: this.$route.query.classroomId })
    }
  },
  methods: {
    async loadClassrooms() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getAllClassroomsForAdmin')
        if (res.code === 200) {
          this.classrooms = res.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    enterClassroom(classroom) {
      this.selectedClassroomId = classroom.id
      this.classroomInfo = classroom
      // 更新 URL 参数,方便刷新时保持状态
      this.$router.replace({ query: { classroomId: classroom.id } })
    },
    goBackToList() {
      this.selectedClassroomId = null
      this.classroomInfo = {}
      this.activeTab = 'students'
      this.$router.replace({ query: {} })
    },
    handleTabClick(tab) {
      // 可以在这里保存标签状态到 URL
      console.log('Tab clicked:', tab.name)
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    }
  }
}
</script>

<style scoped>
.classroom-admin {
  padding: 20px;
  background-color: #fff;
  min-height: 100vh;
}

.classroom-list-view {
  width: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: 600;
}

.empty-tip {
  color: #909399;
  font-size: 14px;
  margin-top: 10px;
}

.classroom-detail-view {
  width: 100%;
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
