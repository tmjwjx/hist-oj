<template>
  <div class="students-panel">
    <div class="header">
      <h3>{{ $t('m.Student_List') }}</h3>
      <el-button type="success" icon="el-icon-download" @click="exportToExcel">
        导出Excel
      </el-button>
    </div>

    <el-table :data="students" v-loading="loading" stripe>
      <el-table-column prop="user.username" :label="$t('m.Username')" />
      <el-table-column prop="realName" :label="$t('m.Real_Name')" />
      <el-table-column prop="gender" :label="$t('m.Gender')" width="80" />
      <el-table-column prop="studentClass" :label="$t('m.Student_Class')" />
      <el-table-column prop="studentNo" :label="$t('m.Student_No')" />
      <el-table-column :label="$t('m.Operation')" width="200">
        <template slot-scope="{ row }">
          <el-button size="small" @click="handleEdit(row)">
            编辑信息
          </el-button>
          <el-button size="small" type="danger" @click="handleRemove(row)">
            {{ $t('m.Remove') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 编辑学生信息对话框 -->
    <el-dialog :title="$t('m.Edit_Student_Info')" :visible.sync="showEditDialog" width="500px">
      <el-form :model="editForm" :rules="rules" ref="editForm" label-width="120px">
        <el-form-item :label="$t('m.Real_Name')" prop="realName">
          <el-input v-model="editForm.realName" />
        </el-form-item>
        <el-form-item :label="$t('m.Gender')" prop="gender">
          <el-radio-group v-model="editForm.gender">
            <el-radio label="男">{{ $t('m.Male') }}</el-radio>
            <el-radio label="女">{{ $t('m.Female') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('m.Student_Class')" prop="studentClass">
          <el-input v-model="editForm.studentClass" />
        </el-form-item>
        <el-form-item :label="$t('m.Student_No')" prop="studentNo">
          <el-input v-model="editForm.studentNo" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showEditDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="saveEdit">{{ $t('m.Save') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'Students',
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      students: [],
      showEditDialog: false,
      editForm: {
        classroomId: null,
        uid: null,
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      },
      rules: {
        realName: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }]
      }
    }
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadStudents()
        }
      }
    }
  },
  mounted() {
    // mounted 时也会通过 watch 触发加载
  },
  methods: {
    async loadStudents() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getClassroomStudents', this.classroomId)
        if (res.code === 200) {
          this.students = res.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    handleEdit(student) {
      this.editForm = {
        classroomId: this.classroomId,
        uid: student.uid, // ClassroomStudent 的 uid 字段
        realName: student.realName,
        gender: student.gender || '',
        studentClass: student.studentClass || '',
        studentNo: student.studentNo || ''
      }
      this.showEditDialog = true
    },
    async saveEdit() {
      this.$refs.editForm.validate(async (valid) => {
        if (valid) {
          try {
            // 确保数据类型正确：classroomId 需要是数字
            // 只发送有值的字段，避免发送空字符串导致验证失败
            const data = {
              classroomId: Number(this.editForm.classroomId),
              uid: String(this.editForm.uid)
            }

            // 只有非空字符串才添加到请求中
            if (this.editForm.realName) {
              data.realName = String(this.editForm.realName)
            }
            if (this.editForm.gender) {
              data.gender = String(this.editForm.gender)
            }
            if (this.editForm.studentClass) {
              data.studentClass = String(this.editForm.studentClass)
            }
            if (this.editForm.studentNo) {
              data.studentNo = String(this.editForm.studentNo)
            }

            const res = await this.$store.dispatch('classroom/updateStudentInfo', data)
            if (res.code === 200) {
              this.$message.success(this.$t('m.Save_Success'))
              this.showEditDialog = false
              this.loadStudents()
            } else {
              this.$message.error(res.message || this.$t('m.Save_Failed'))
            }
          } catch (error) {
            this.$message.error(this.$t('m.Save_Failed'))
          }
        }
      })
    },
    handleRemove(student) {
      this.$confirm(this.$t('m.Confirm_Remove_Student'), this.$t('m.Warning'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/removeStudent', {
            classroomId: this.classroomId,
            userId: student.userId
          })
          if (res.code === 200) {
            this.$message.success(this.$t('m.Remove_Success'))
            this.loadStudents()
          } else {
            this.$message.error(res.message || this.$t('m.Remove_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Remove_Failed'))
        }
      })
    },
    exportToExcel() {
      // 创建Excel内容
      let csvContent = '\uFEFF' // UTF-8 BOM
      csvContent += '用户名,真实姓名,性别,班级,学号\n'

      this.students.forEach(student => {
        const row = [
          student.user?.username || '',
          student.realName || '',
          student.gender || '',
          student.studentClass || '',
          student.studentNo || ''
        ].map(field => `"${field}"`).join(',')
        csvContent += row + '\n'
      })

      // 创建Blob并下载
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', `学生信息_${new Date().getTime()}.csv`)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    }
  }
}
</script>

<style scoped>
.students-panel {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  margin: 0;
  font-size: 20px;
  color: #409EFF;
}

/* 增加学生列表表格字体大小 */
.students-panel ::v-deep .el-table {
  font-size: 15px;
}

.students-panel ::v-deep .el-table th {
  font-size: 15px;
  font-weight: 600;
}

.students-panel ::v-deep .el-table td {
  font-size: 15px;
}
</style>
