<template>
  <div class="classroom-admin-container">
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
            {{ row.teacher ? row.teacher.username : row.teacherId || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="classCode" :label="$t('m.Classroom_Code')" width="120" />
        <el-table-column prop="createTime" :label="$t('m.Create_Time')" width="180">
          <template slot-scope="{ row }">{{ formatTime(row.createTime) }}</template>
        </el-table-column>
        <el-table-column :label="$t('m.Operation')" width="150" fixed="right">
          <template slot-scope="{ row }">
            <el-button size="mini" type="primary" @click="viewDetail(row)">
              {{ $t('m.View_Detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && classrooms.length === 0" :description="$t('m.No_Data')">
        <p class="empty-tip">{{ $t('m.Classroom_Admin_Empty_Tip') }}</p>
      </el-empty>
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'

export default {
  name: 'ClassroomAdmin',
  data() {
    return {
      loading: false,
      classrooms: []
    }
  },
  mounted() {
    this.loadClassrooms()
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
    viewDetail(classroom) {
      // 跳转到教师端班级详情页面
      this.$router.push({
        name: 'TeacherClassroomDetail',
        params: { classroomId: classroom.id },
        query: { tab: 'students' }
      })
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    }
  }
}
</script>

<style scoped>
.classroom-admin-container {
  padding: 20px;
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
</style>
