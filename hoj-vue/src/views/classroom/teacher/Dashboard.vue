<template>
  <div class="teacher-dashboard classroom-theme">
    <div class="teacher-dashboard-inner">
      <div class="header">
        <div class="header-left">
          <h1>教师工作台</h1>
          <p class="subtitle">管理您的班级</p>
        </div>
        <div class="header-actions">
          <button class="classroom-btn classroom-btn-success" @click="goToQuestionBank">
            <i class="el-icon-document"></i>
            <span>题库管理</span>
          </button>
          <button class="classroom-btn classroom-btn-info" @click="goToExamPaper">
            <i class="el-icon-document-copy"></i>
            <span>试卷库</span>
          </button>
          <button class="classroom-btn classroom-btn-primary" @click="showCreateDialog = true">
            <i class="el-icon-plus"></i>
            <span>创建班级</span>
          </button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="safeClassrooms.length === 0 && !loading" class="classroom-empty">
        <i class="el-icon-school classroom-empty-icon"></i>
        <div class="classroom-empty-text">还没有班级</div>
        <div class="classroom-empty-hint">点击上方按钮创建您的第一个班级吧！</div>
      </div>

      <!-- 班级卡片列表 -->
      <div v-else class="classroom-list">
        <div v-for="classroom in safeClassrooms" :key="classroom.id" class="classroom-card classroom-fade-in">
          <div class="classroom-card-header">
            <div class="card-header-content" @click="viewClassroom(classroom)">
              <div class="class-name">{{ classroom.className }}</div>
              <span class="classroom-tag classroom-tag-success">{{ classroom.classBelong }}</span>
            </div>
          </div>
          <div class="card-body" @click="viewClassroom(classroom)">
            <div class="info-item">
              <i class="el-icon-key info-icon"></i>
              <span class="info-label">班级代码</span>
              <span class="info-value">{{ classroom.classCode }}</span>
              <button
                class="classroom-btn classroom-btn-secondary copy-btn"
                @click.stop="copyClassCode(classroom.classCode)"
              >
                <i class="el-icon-document-copy"></i>
                <span>复制</span>
              </button>
            </div>
            <div class="info-item">
              <i class="el-icon-user info-icon"></i>
              <span class="info-label">教师</span>
              <span class="info-value">{{ getTeacherNames(classroom) }}</span>
            </div>
          </div>
          <div class="card-footer">
            <button class="classroom-btn classroom-btn-primary" @click="viewClassroom(classroom)">
              <i class="el-icon-setting"></i>
              <span>管理班级</span>
            </button>
            <button class="classroom-btn classroom-btn-danger" @click="handleDelete(classroom)">
              <i class="el-icon-delete"></i>
              <span>删除班级</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 创建班级对话框 -->
      <el-dialog title="创建班级" :visible.sync="showCreateDialog" width="500px" custom-class="classroom-dialog">
        <el-form :model="createForm" :rules="rules" ref="createForm" label-width="100px">
          <el-form-item label="班级名称" prop="className">
            <el-input v-model="createForm.className" placeholder="请输入班级名称" class="classroom-input" />
          </el-form-item>
          <el-form-item label="班级所属" prop="classBelong">
            <el-input v-model="createForm.classBelong" placeholder="请输入班级所属" class="classroom-input" />
          </el-form-item>
        </el-form>
        <span slot="footer">
          <el-button @click="showCreateDialog = false" class="classroom-btn classroom-btn-secondary">取消</el-button>
          <el-button type="primary" @click="createClassroom" :loading="submitting" class="classroom-btn classroom-btn-primary">确认创建</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'TeacherDashboard',
  mixins: [realtimeSync, teacherAuth],
  data() {
    return {
      loading: false,
      submitting: false,
      showCreateDialog: false,
      classrooms: [],
      createForm: {
        className: '',
        classBelong: ''
      },
      rules: {
        className: [{ required: true, message: '请输入班级名称', trigger: 'blur' }],
        classBelong: [{ required: true, message: '请输入班级所属', trigger: 'blur' }]
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
    // 正常加载班级列表
    this.loadClassrooms()

    // 检查是否从返回按钮带着参数跳转过来
    // 注意：需要在 loadClassrooms 完成后检查，因为需要 classrooms 数据
    this.$watch('classrooms', (classrooms) => {
      if (this.$route.query.classroomId && this.$route.query.tab && classrooms.length > 0) {
        const classroomId = Number(this.$route.query.classroomId)
        const tab = this.$route.query.tab
        const homeworkId = this.$route.query.homeworkId

        // 查找对应的班级
        const classroom = classrooms.find(c => c.id === classroomId)
        if (classroom) {
          // 延迟执行，确保组件已渲染
          this.$nextTick(() => {
            this.navigateToClassroom(classroom, tab, homeworkId)
          })
        }
      }
    }, { immediate: true })
  },
  methods: {
    async loadClassrooms() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.classrooms.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getClassroomList')
        if (res && res.code === 200 && Array.isArray(res.data)) {
          const newClassrooms = res.data

          // 检查数量是否变化
          if (newClassrooms.length !== this.classrooms.length) {
            this.classrooms = newClassrooms
            return
          }

          // 检查每个班级的 ID 是否都相同(避免深度对比整个对象)
          const currentIds = this.classrooms.map(c => c.id).sort().join(',')
          const newIds = newClassrooms.map(c => c.id).sort().join(',')

          if (currentIds !== newIds) {
            // ID 列表不同，说明有班级增删，需要更新
            this.classrooms = newClassrooms
          }
          // 如果 ID 列表相同，不更新数据，避免闪烁
        } else {
          if (isFirstLoad) {
            this.classrooms = []
          }
        }
      } catch (error) {
        console.error('加载班级列表失败:', error)
        if (isFirstLoad) {
          this.$message.error('加载失败')
          this.classrooms = []
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    async createClassroom() {
      this.$refs.createForm.validate(async (valid) => {
        if (valid) {
          this.submitting = true
          try {
            const res = await this.$store.dispatch('classroom/createClassroom', this.createForm)
            if (res.code === 200) {
              this.$message.success('创建成功')
              this.showCreateDialog = false
              this.createForm = { className: '', classBelong: '' }
              this.loadClassrooms()
            } else {
              this.$message.error(res.message || '创建失败')
            }
          } catch (error) {
            console.error('创建班级失败:', error)
            this.$message.error('创建失败')
          } finally {
            this.submitting = false
          }
        }
      })
    },
    viewClassroom(classroom, tab = 'students') {
      this.navigateToClassroom(classroom, tab)
    },
    navigateToClassroom(classroom, tab = 'students', homeworkId = null) {
      const query = { tab }
      if (homeworkId) {
        query.homeworkId = homeworkId
      }
      this.$router.push({
        name: 'TeacherClassroomDetail',
        params: { classroomId: classroom.id },
        query
      })
    },
    copyClassCode(classCode) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(classCode).then(() => {
          this.$message.success('班级代码已复制')
        }).catch(() => {
          this.fallbackCopy(classCode)
        })
      } else {
        this.fallbackCopy(classCode)
      }
    },
    fallbackCopy(classCode) {
      const textArea = document.createElement('textarea')
      textArea.value = classCode
      textArea.style.position = 'fixed'
      textArea.style.left = '-9999px'
      document.body.appendChild(textArea)
      textArea.select()
      try {
        document.execCommand('copy')
        this.$message.success('班级代码已复制')
      } catch (err) {
        this.$message.error('复制失败')
      }
      document.body.removeChild(textArea)
    },
    handleDelete(classroom) {
      this.$confirm('确认删除该班级吗？', '警告', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteClassroom', classroom.id)
          if (res.code === 200) {
            this.$message.success('删除成功')
            this.loadClassrooms()
          } else {
            this.$message.error(res.msg || '删除失败')
          }
        } catch (error) {
          console.error('删除班级失败:', error)
          this.$message.error('删除失败')
        }
      })
    },
    goToQuestionBank() {
      this.$router.push({ name: 'QuestionBank' })
    },
    goToExamPaper() {
      this.$router.push({ name: 'ExamPaper' })
    },
    getTeacherNames(classroom) {
      const teachers = []

      // 添加主教师
      if (classroom.teacher) {
        teachers.push(classroom.teacher.realname || classroom.teacher.username || classroom.teacher.nickname || '-')
      }

      // 添加其他教师
      if (classroom.teachers && classroom.teachers.length > 0) {
        classroom.teachers.forEach(t => {
          if (t.teacher) {
            // 避免重复添加主教师
            const isDuplicate = classroom.teacher && t.teacher.uuid === classroom.teacher.uuid
            if (!isDuplicate) {
              teachers.push(t.teacher.realname || t.teacher.username || t.teacher.nickname)
            }
          }
        })
      }

      return teachers.length > 0 ? teachers.join('、') : '-'
    }
  },
  computed: {
    safeClassrooms() {
      if (!Array.isArray(this.classrooms)) return []
      return this.classrooms.filter(c => c && typeof c === 'object' && c.className)
    }
  }
}
</script>

<style scoped>
@import '../classroom-theme.css';

.teacher-dashboard {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
  margin: 0;
  --workspace-surface-bg: var(--classroom-card-bg);
}

.teacher-dashboard-inner {
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  padding: 24px;
  background: var(--workspace-surface-bg);
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

.header-actions {
  display: flex;
  gap: 12px;
}

/* 班级卡片列表样式 */
.classroom-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 24px;
}

.classroom-card {
  background: var(--workspace-surface-bg);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
}

.classroom-card:hover {
  box-shadow: 0 8px 24px rgba(74, 144, 226, 0.15);
  transform: translateY(-4px);
}

.classroom-card-header {
  background: var(--workspace-surface-bg);
  padding: 20px;
  border-bottom: 1px solid var(--classroom-border);
}

.card-header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  color: var(--classroom-text-secondary);
  font-size: 14px;
}

.info-icon {
  font-size: 18px;
  color: var(--classroom-primary);
}

.info-label {
  font-weight: 500;
  min-width: 70px;
}

.info-value {
  flex: 1;
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: var(--classroom-primary);
}

.copy-btn {
  padding: 6px 14px;
  font-size: 13px;
}

.card-footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  background: var(--workspace-surface-bg);
  border-top: 1px solid var(--classroom-border);
}

.card-footer .classroom-btn {
  flex: 1;
}

/* 对话框样式 */
.classroom-dialog .el-dialog__header {
  background: var(--workspace-surface-bg);
  border-bottom: 1px solid var(--classroom-border);
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

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .header-actions .classroom-btn {
    width: 100%;
  }
}
</style>
