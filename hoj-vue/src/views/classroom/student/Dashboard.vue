<template>
  <div class="student-dashboard classroom-theme">
    <div class="header">
      <div class="header-left">
        <h1>{{ $t('m.My_Classrooms') }}</h1>
        <p class="subtitle">查看和管理您的课程学习</p>
      </div>
      <button class="classroom-btn classroom-btn-primary" @click="showJoinDialog = true">
        <i class="el-icon-plus"></i>
        <span>{{ $t('m.Join_Classroom') }}</span>
      </button>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && (!classrooms || classrooms.length === 0)" class="classroom-empty">
      <i class="el-icon-reading classroom-empty-icon"></i>
      <div class="classroom-empty-text">{{ $t('m.No_Classroom_Yet') || '还没有加入班级' }}</div>
      <div class="classroom-empty-hint">{{ $t('m.Join_Classroom_Tip') || '点击上方按钮加入您的第一个班级吧！' }}</div>
    </div>

    <!-- 班级卡片列表 -->
    <div v-else class="classroom-list">
      <div
        v-for="classroom in classrooms"
        :key="classroom.id"
        class="classroom-card classroom-fade-in"
        @click="viewClassroom(classroom)"
      >
        <div class="classroom-card-header">
          <div class="class-name">{{ classroom.className }}</div>
          <span class="classroom-tag classroom-tag-success">{{ classroom.classBelong }}</span>
        </div>
        <div class="card-body">
          <div class="info-item">
            <i class="el-icon-user info-icon"></i>
            <span class="info-label">{{ $t('m.Teacher') }}</span>
            <span class="info-value">{{ getTeacherName(classroom) }}</span>
          </div>
          <div class="info-item">
            <i class="el-icon-key info-icon"></i>
            <span class="info-label">{{ $t('m.Class_Code') }}</span>
            <span class="info-value">{{ classroom.classCode }}</span>
          </div>
        </div>
        <div class="card-footer">
          <span class="enter-hint">
            <i class="el-icon-arrow-right"></i>
            <span>点击进入班级</span>
          </span>
        </div>
      </div>
    </div>

    <!-- 加入班级对话框 -->
    <el-dialog :title="$t('m.Join_Classroom')" :visible.sync="showJoinDialog" width="500px" custom-class="classroom-dialog">
      <el-form :model="joinForm" :rules="rules" ref="joinForm" label-width="120px">
        <el-form-item :label="$t('m.Class_Code')" prop="classCode">
          <el-input v-model="joinForm.classCode" :placeholder="$t('m.Enter_Class_Code')" class="classroom-input" />
        </el-form-item>
        <el-form-item :label="$t('m.Real_Name')" prop="realName">
          <el-input v-model="joinForm.realName" class="classroom-input" />
        </el-form-item>
        <el-form-item :label="$t('m.Gender')" prop="gender">
          <el-radio-group v-model="joinForm.gender">
            <el-radio label="男">{{ $t('m.Male') }}</el-radio>
            <el-radio label="女">{{ $t('m.Female') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('m.Student_Class')" prop="studentClass">
          <el-input v-model="joinForm.studentClass" class="classroom-input" />
        </el-form-item>
        <el-form-item :label="$t('m.Student_No')" prop="studentNo">
          <el-input v-model="joinForm.studentNo" class="classroom-input" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showJoinDialog = false" class="classroom-btn classroom-btn-secondary">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="joinClassroom" class="classroom-btn classroom-btn-primary">{{ $t('m.Confirm_Join') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'StudentDashboard',
  mixins: [realtimeSync, studentAuth],
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
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadMyClassrooms',
        immediate: true
      }
    }
  },
  computed: {
    rules() {
      return {
        classCode: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }],
        realName: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }],
        gender: [{ required: true, message: this.$t('m.Required'), trigger: 'change' }],
        studentClass: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }],
        studentNo: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }]
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
@import '../classroom-theme.css';

.student-dashboard {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header-left h1 {
  font-size: 28px;
  color: var(--classroom-text);
  margin: 0 0 8px 0;
  font-weight: 700;
}

.subtitle {
  font-size: 14px;
  color: var(--classroom-text-secondary);
  margin: 0;
}

/* 班级卡片列表 */
.classroom-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 24px;
}

.classroom-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
  border: 1px solid var(--classroom-border);
}

.classroom-card:hover {
  box-shadow: 0 8px 24px rgba(74, 144, 226, 0.15);
  transform: translateY(-4px);
}

.classroom-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
}

.class-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--classroom-text);
}

.card-body {
  padding: 20px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-icon {
  font-size: 18px;
  color: var(--classroom-primary);
}

.info-label {
  font-weight: 500;
  color: var(--classroom-text-secondary);
  min-width: 80px;
}

.info-value {
  flex: 1;
  color: var(--classroom-text);
  font-weight: 500;
}

.card-footer {
  text-align: center;
  padding: 16px;
  background: var(--classroom-bg);
  border-top: 1px solid var(--classroom-border);
}

.enter-hint {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--classroom-primary);
  font-size: 14px;
  font-weight: 500;
}

.enter-hint i {
  font-size: 16px;
  transition: transform 0.3s ease;
}

.classroom-card:hover .enter-hint i {
  transform: translateX(4px);
}

/* 对话框样式 */
.classroom-dialog .el-dialog__header {
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
}

.classroom-dialog .el-dialog__title {
  color: var(--classroom-text);
  font-weight: 600;
}

/* 响应式 */
@media (max-width: 768px) {
  .classroom-list {
    grid-template-columns: 1fr;
  }

  .header {
    flex-direction: column;
    gap: 16px;
  }

  .header .classroom-btn {
    width: 100%;
  }

  .info-item {
    flex-wrap: wrap;
  }

  .info-label {
    min-width: auto;
  }
}
</style>
