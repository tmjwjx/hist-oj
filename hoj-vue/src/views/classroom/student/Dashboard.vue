<template>
  <div class="student-dashboard">
    <div class="header">
      <div class="header-left">
        <h1>{{ $t('m.My_Classrooms') }}</h1>
        <p class="subtitle">查看和管理您的课程学习</p>
      </div>
      <el-button type="primary" icon="el-icon-plus" @click="showJoinDialog = true">
        {{ $t('m.Join_Classroom') }}
      </el-button>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && (!classrooms || classrooms.length === 0)" class="empty-state">
      <i class="el-icon-reading empty-icon"></i>
      <h3>{{ $t('m.No_Classroom_Yet') || '还没有加入班级' }}</h3>
      <p>{{ $t('m.Join_Classroom_Tip') || '点击上方按钮加入您的第一个班级吧！' }}</p>
    </div>

    <!-- 班级卡片列表 -->
    <el-row v-else :gutter="20" class="classroom-list">
      <el-col :span="8" v-for="classroom in classrooms" :key="classroom.id">
        <el-card class="classroom-card" @click.native="viewClassroom(classroom)">
          <div class="card-header">
            <div class="class-name">{{ classroom.className }}</div>
            <el-tag size="small" type="success">{{ classroom.classBelong }}</el-tag>
          </div>
          <div class="card-content">
            <div class="info-item">
              <i class="el-icon-user"></i>
              <span>{{ $t('m.Teacher') }}: {{ getTeacherName(classroom) }}</span>
            </div>
            <div class="info-item">
              <i class="el-icon-key"></i>
              <span>{{ $t('m.Class_Code') }}: {{ classroom.classCode }}</span>
            </div>
          </div>
          <div class="card-footer">
            <el-tag size="mini" type="info">
              <i class="el-icon-arrow-right"></i> 点击进入
            </el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 加入班级对话框 -->
    <el-dialog :title="$t('m.Join_Classroom')" :visible.sync="showJoinDialog" width="500px">
      <el-form :model="joinForm" :rules="rules" ref="joinForm" label-width="120px">
        <el-form-item :label="$t('m.Class_Code')" prop="classCode">
          <el-input v-model="joinForm.classCode" :placeholder="$t('m.Enter_Class_Code')" />
        </el-form-item>
        <el-form-item :label="$t('m.Real_Name')" prop="realName">
          <el-input v-model="joinForm.realName" />
        </el-form-item>
        <el-form-item :label="$t('m.Gender')" prop="gender">
          <el-radio-group v-model="joinForm.gender">
            <el-radio label="男">{{ $t('m.Male') }}</el-radio>
            <el-radio label="女">{{ $t('m.Female') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('m.Student_Class')" prop="studentClass">
          <el-input v-model="joinForm.studentClass" />
        </el-form-item>
        <el-form-item :label="$t('m.Student_No')" prop="studentNo">
          <el-input v-model="joinForm.studentNo" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showJoinDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="joinClassroom">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'

export default {
  name: 'StudentDashboard',
  mixins: [realtimeSync],
  data() {
    return {
      loading: false,
      classrooms: [], // 确保初始化为空数组
      showJoinDialog: false,
      joinForm: {
        classCode: '',
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      },
      rules: {
        classCode: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }],
        realName: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }]
      },
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadMyClassrooms',
        immediate: true
      }
    }
  },
  mounted() {
    this.loadMyClassrooms()
  },
  methods: {
    async loadMyClassrooms() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.classrooms.length === 0 && !this._hasLoadedOnce
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getMyClassrooms')

        if (res) {
          if (res.code === 200) {
            const newClassrooms = (res.data || []).filter(c => c && typeof c === 'object' && c.id)

            // 深度对比：使用 JSON.stringify 检查数据是否真的变化
            const currentDataString = JSON.stringify(this.classrooms)
            const newDataString = JSON.stringify(newClassrooms)

            if (currentDataString !== newDataString) {
              // 数据真的变化了，才更新
              this.classrooms = newClassrooms
            }
          } else {
            // 只在首次加载失败时设置为空数组
            if (!this._hasLoadedOnce) {
              this.classrooms = []
            }
          }
        } else {
          // 只在首次加载失败时设置为空数组
          if (!this._hasLoadedOnce) {
            this.classrooms = []
          }
        }
      } catch (error) {
        // 只在首次加载失败时设置为空数组并显示错误
        if (!this._hasLoadedOnce) {
          this.classrooms = []
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
          this._hasLoadedOnce = true
        }
      }
    },
    async joinClassroom() {
      this.$refs.joinForm.validate(async (valid) => {
        if (valid) {
          try {
            const res = await this.$store.dispatch('classroom/joinClassroom', this.joinForm)
            if (res.code === 200) {
              this.$message.success(this.$t('m.Join_Success'))
              this.showJoinDialog = false
              this.joinForm = {
                classCode: '',
                realName: '',
                gender: '',
                studentClass: '',
                studentNo: ''
              }
              this.loadMyClassrooms()
            } else {
              this.$message.error(res.message || this.$t('m.Join_Failed'))
            }
          } catch (error) {
            this.$message.error(this.$t('m.Join_Failed'))
          }
        }
      })
    },
    viewClassroom(classroom) {
      this.$router.push({
        name: 'StudentClassroomDetail',
        params: { classroomId: classroom.id },
        query: { tab: 'homework' }
      })
    },
    getTeacherName(classroom) {
      if (classroom.teacher) {
        // 优先显示用户名，如果没有则显示昵称
        return classroom.teacher.username || classroom.teacher.nickname || '-'
      }
      return '-'
    }
  }
}
</script>

<style scoped>
.student-dashboard {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 30px;
}

.header-left h1 {
  font-size: 32px;
  color: #303133;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* 空状态样式 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
  background: #fff;
  border-radius: 8px;
}

.empty-icon {
  font-size: 120px;
  color: #DCDFE6;
  margin-bottom: 20px;
}

.empty-state h3 {
  font-size: 20px;
  color: #606266;
  margin: 0 0 10px 0;
}

.empty-state p {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* 班级卡片列表 */
.classroom-list {
  margin-top: 20px;
}

.classroom-card {
  margin-bottom: 20px;
  cursor: pointer;
  /* 移除 transition 避免轮询时闪烁 */
  border-radius: 8px;
  overflow: hidden;
}

.classroom-card:hover {
  /* 移除 transform 避免轮询时闪烁 */
  box-shadow: 0 8px 30px rgba(64, 158, 255, 0.3);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 12px;
  border-bottom: 1px solid #EBEEF5;
}

.class-name {
  font-size: 18px;
  font-weight: bold;
  color: #303133;
}

.card-content {
  margin-bottom: 15px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
  color: #606266;
}

.info-item i {
  font-size: 16px;
  color: #909399;
}

.info-item:last-child {
  margin-bottom: 0;
}

.card-footer {
  text-align: center;
  padding-top: 12px;
  border-top: 1px solid #EBEEF5;
}
</style>
