<template>
  <div class="teacher-dashboard">
    <div class="header">
      <div class="header-left">
        <h1>教师工作台</h1>
        <p class="subtitle">管理您的班级</p>
      </div>
      <div class="header-actions">
        <el-button type="success" icon="el-icon-document" @click="goToQuestionBank">
          题库
        </el-button>
        <el-button type="primary" icon="el-icon-plus" @click="showCreateDialog = true">
          创建班级
        </el-button>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="safeClassrooms.length === 0 && !loading" class="empty-state">
      <i class="el-icon-school empty-icon"></i>
      <h3>还没有班级</h3>
      <p>点击上方按钮创建您的第一个班级吧！</p>
    </div>

    <!-- 班级列表 -->
    <el-table
      v-else
      :data="safeClassrooms"
      v-loading="loading"
      stripe
      class="classroom-table"
    >
      <el-table-column prop="className" label="班级名称" />
      <el-table-column prop="classBelong" label="班级所属" />
      <el-table-column label="班级代码" width="250">
        <template slot-scope="{ row }">
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-family: monospace; font-size: 14px;">{{ row.classCode }}</span>
            <el-button
              type="text"
              icon="el-icon-document-copy"
              @click="copyClassCode(row.classCode)"
              style="padding: 0;"
            >
            </el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250">
        <template slot-scope="{ row }">
          <el-button size="small" @click="viewClassroom(row)">
            管理
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建班级对话框 -->
    <el-dialog title="创建班级" :visible.sync="showCreateDialog" width="500px">
      <el-form :model="createForm" :rules="rules" ref="createForm" label-width="100px">
        <el-form-item label="班级名称" prop="className">
          <el-input v-model="createForm.className" placeholder="请输入班级名称" />
        </el-form-item>
        <el-form-item label="班级所属" prop="classBelong">
          <el-input v-model="createForm.classBelong" placeholder="请输入班级所属" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createClassroom" :loading="submitting">确认</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'TeacherDashboard',
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
      }
    }
  },
  mounted() {
    this.loadClassrooms()
  },
  methods: {
    async loadClassrooms() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getClassroomList')
        if (res && res.code === 200 && Array.isArray(res.data)) {
          this.classrooms = res.data
        } else {
          this.classrooms = []
        }
      } catch (error) {
        console.error('加载班级列表失败:', error)
        this.$message.error('加载失败')
        this.classrooms = []
      } finally {
        this.loading = false
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
    viewClassroom(classroom) {
      this.$router.push({
        name: 'TeacherClassroomDetail',
        params: { classroomId: classroom.id },
        query: { tab: 'homework' }
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
          if (res.status === 200) {
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
.teacher-dashboard {
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

.header-actions {
  display: flex;
  gap: 10px;
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

/* 表格样式 */
.classroom-table {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

/* 增加班级列表表格字体大小 */
.teacher-dashboard ::v-deep .el-table {
  font-size: 15px;
}

.teacher-dashboard ::v-deep .el-table th {
  font-size: 15px;
  font-weight: 600;
}

.teacher-dashboard ::v-deep .el-table td {
  font-size: 15px;
}
</style>
