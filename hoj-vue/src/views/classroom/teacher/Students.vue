<template>
  <div class="students-panel classroom-theme">
    <div class="header">
      <div class="header-left">
        <h3>{{ $t('m.Student_List') }}</h3>
        <p class="subtitle">管理班级中的学生信息</p>
      </div>
      <div class="header-actions">
        <button class="classroom-btn classroom-btn-primary" @click="showAddDialog = true">
          <i class="el-icon-plus"></i>
          <span>添加学生</span>
        </button>
        <button class="classroom-btn classroom-btn-success" @click="exportToExcel">
          <i class="el-icon-download"></i>
          <span>导出Excel</span>
        </button>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="stats-row">
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ students.length }}</div>
        <div class="classroom-stat-label">总学生数</div>
      </div>
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ maleCount }}</div>
        <div class="classroom-stat-label">男生</div>
      </div>
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ femaleCount }}</div>
        <div class="classroom-stat-label">女生</div>
      </div>
    </div>

    <!-- 学生列表表格 -->
    <el-card class="table-card classroom-card">
      <el-table :data="students" stripe class="classroom-table">
        <el-table-column :label="$t('m.Username')">
          <template slot-scope="{ row }">
            <UserName :username="row.user?.username" />
          </template>
        </el-table-column>
        <el-table-column prop="realName" :label="$t('m.Real_Name')" />
        <el-table-column prop="gender" :label="$t('m.Gender')" width="80">
          <template slot-scope="{ row }">
            <span v-if="row.gender" class="classroom-tag" :class="row.gender === '男' ? 'classroom-tag-primary' : 'classroom-tag-warning'">
              {{ row.gender }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="studentClass" :label="$t('m.Student_Class')" />
        <el-table-column prop="studentNo" :label="$t('m.Student_No')" />
        <el-table-column :label="$t('m.Operation')" width="220">
          <template slot-scope="{ row }">
            <div class="action-buttons">
              <button class="classroom-btn classroom-btn-primary" size="small" @click="handleEdit(row)">
                <i class="el-icon-edit"></i>
                <span>编辑信息</span>
              </button>
              <button class="classroom-btn classroom-btn-danger" size="small" @click="handleRemove(row)">
                <i class="el-icon-delete"></i>
                <span>{{ $t('m.Remove') }}</span>
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑学生信息对话框 -->
    <el-dialog :title="$t('m.Edit_Student_Info')" :visible.sync="showEditDialog" width="500px" custom-class="classroom-dialog">
      <el-form :model="editForm" :rules="rules" ref="editForm" label-width="120px" class="edit-form">
        <el-form-item :label="$t('m.Real_Name')" prop="realName">
          <el-input v-model="editForm.realName" class="classroom-input" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item :label="$t('m.Gender')" prop="gender">
          <el-radio-group v-model="editForm.gender" class="gender-radio-group">
            <el-radio label="男">{{ $t('m.Male') }}</el-radio>
            <el-radio label="女">{{ $t('m.Female') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('m.Student_Class')" prop="studentClass">
          <el-input v-model="editForm.studentClass" class="classroom-input" placeholder="请输入班级" />
        </el-form-item>
        <el-form-item :label="$t('m.Student_No')" prop="studentNo">
          <el-input v-model="editForm.studentNo" class="classroom-input" placeholder="请输入学号" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <button class="classroom-btn classroom-btn-secondary" @click="showEditDialog = false">{{ $t('m.Cancel') }}</button>
        <button class="classroom-btn classroom-btn-primary" @click="saveEdit">{{ $t('m.Save') }}</button>
      </div>
    </el-dialog>

    <!-- 添加学生对话框 -->
    <el-dialog title="添加学生" :visible.sync="showAddDialog" width="600px" custom-class="classroom-dialog">
      <div class="add-student-dialog">
        <!-- 搜索学生 -->
        <div class="search-section">
          <el-input
            v-model="searchKeyword"
            placeholder="请输入用户名、真实姓名或昵称搜索"
            prefix-icon="el-icon-search"
            clearable
            @keyup.enter.native="handleSearchStudents"
            class="search-input"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleSearchStudents">搜索</el-button>
          </el-input>
        </div>

        <!-- 搜索结果 -->
        <div v-if="searchResults.length > 0" class="search-results">
          <div class="results-header">搜索结果（请点击选择学生）</div>
          <div class="results-list">
            <div
              v-for="user in searchResults"
              :key="user.uid"
              class="user-item"
              :class="{ 'selected': selectedUser && selectedUser.uid === user.uid }"
              @click="selectUser(user)"
            >
              <div class="user-info">
                <div class="user-name">{{ user.username }}</div>
                <div class="user-details">
                  <span v-if="user.realname">{{ user.realname }}</span>
                  <span v-if="user.nickname" class="nickname">({{ user.nickname }})</span>
                </div>
                <div v-if="user.email" class="user-email">{{ user.email }}</div>
              </div>
              <i v-if="selectedUser && selectedUser.uid === user.uid" class="el-icon-check selected-icon"></i>
            </div>
          </div>
        </div>

        <!-- 学生信息表单 -->
        <div v-if="selectedUser" class="student-form">
          <el-divider>填写学生信息</el-divider>
          <el-form :model="addForm" :rules="addRules" ref="addForm" label-width="100px">
            <el-form-item label="已选用户">
              <div class="selected-user-display">
                <strong>{{ selectedUser.username }}</strong>
                <span v-if="selectedUser.realname">（{{ selectedUser.realname }}）</span>
              </div>
            </el-form-item>
            <el-form-item label="真实姓名" prop="realName">
              <el-input v-model="addForm.realName" placeholder="请输入真实姓名" />
            </el-form-item>
            <el-form-item label="性别" prop="gender">
              <el-radio-group v-model="addForm.gender">
                <el-radio label="男">男</el-radio>
                <el-radio label="女">女</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="班级" prop="studentClass">
              <el-input v-model="addForm.studentClass" placeholder="请输入班级" />
            </el-form-item>
            <el-form-item label="学号" prop="studentNo">
              <el-input v-model="addForm.studentNo" placeholder="请输入学号" />
            </el-form-item>
          </el-form>
        </div>

        <!-- 提示信息 -->
        <el-alert
          v-if="!selectedUser && searchResults.length === 0"
          title="请先搜索并选择要添加的学生"
          type="info"
          :closable="false"
          show-icon
        />
      </div>
      <div slot="footer" class="dialog-footer">
        <button class="classroom-btn classroom-btn-secondary" @click="closeAddDialog">取消</button>
        <button
          class="classroom-btn classroom-btn-primary"
          :disabled="!selectedUser"
          @click="handleAddStudent"
        >
          添加
        </button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'
import teacherAuth from '@/mixins/teacherAuth'

export default {
  name: 'Students',
  components: {
    UserName
  },
  mixins: [realtimeSync, teacherAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      students: [],
      showEditDialog: false,
      showAddDialog: false,
      editForm: {
        classroomId: null,
        uid: null,
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      },
      searchKeyword: '',
      searchResults: [],
      selectedUser: null,
      addForm: {
        uid: '',
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      },
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadStudents',
        immediate: true
      },
      rules: {
        realName: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }]
      },
      addRules: {
        realName: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }]
      }
    }
  },
  computed: {
    maleCount() {
      return this.students.filter(s => s.gender === '男').length
    },
    femaleCount() {
      return this.students.filter(s => s.gender === '女').length
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
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.students.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getClassroomStudents', this.classroomId)
        if (res.code === 200) {
          const newStudents = res.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.students)
          const newDataString = JSON.stringify(newStudents)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.students = newStudents
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
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
            uid: student.uid
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
    },
    async handleSearchStudents() {
      if (!this.searchKeyword || this.searchKeyword.trim() === '') {
        this.$message.warning('请输入搜索关键词')
        return
      }

      try {
        const res = await this.$store.dispatch('classroom/searchStudentsToAdd', {
          classroomId: this.classroomId,
          keyword: this.searchKeyword.trim()
        })

        if (res.code === 200) {
          this.searchResults = res.data || []
          if (this.searchResults.length === 0) {
            this.$message.info('未找到匹配的用户')
          }
        } else {
          this.$message.error(res.message || '搜索失败')
        }
      } catch (error) {
        console.error('搜索学生失败:', error)
        this.$message.error('搜索失败')
      }
    },
    selectUser(user) {
      this.selectedUser = user
      this.addForm.uid = user.uid
      // 自动填充真实姓名（如果有）
      if (user.realname) {
        this.addForm.realName = user.realname
      }
    },
    async handleAddStudent() {
      if (!this.selectedUser) {
        this.$message.warning('请先选择要添加的学生')
        return
      }

      this.$refs.addForm.validate(async (valid) => {
        if (valid) {
          try {
            const data = {
              uid: this.addForm.uid,
              realName: this.addForm.realName,
              gender: this.addForm.gender,
              studentClass: this.addForm.studentClass,
              studentNo: this.addForm.studentNo
            }

            const res = await this.$store.dispatch('classroom/addClassroomStudent', {
              classroomId: this.classroomId,
              data
            })

            if (res.code === 200) {
              this.$message.success('添加学生成功')
              this.closeAddDialog()
              this.loadStudents()
            } else {
              this.$message.error(res.message || '添加失败')
            }
          } catch (error) {
            console.error('添加学生失败:', error)
            this.$message.error('添加失败')
          }
        }
      })
    },
    closeAddDialog() {
      this.showAddDialog = false
      this.searchKeyword = ''
      this.searchResults = []
      this.selectedUser = null
      this.addForm = {
        uid: '',
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      }
      if (this.$refs.addForm) {
        this.$refs.addForm.resetFields()
      }
    }
  }
}
</script>

<style scoped>
@import '../classroom-theme.css';

.students-panel {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header-left h3 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--classroom-text);
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

.header .classroom-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
  white-space: nowrap;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.table-card {
  margin-bottom: 20px;
}

.classroom-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
  overflow: hidden;
}

.classroom-card ::v-deep .el-card__body {
  padding: 20px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.action-buttons .classroom-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  font-size: 13px;
  margin: 0;
}

/* 对话框样式 */
.classroom-dialog ::v-deep .el-dialog__header {
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
  padding: 20px 24px;
}

.classroom-dialog ::v-deep .el-dialog__title {
  color: var(--classroom-text);
  font-weight: 600;
  font-size: 18px;
}

.classroom-dialog ::v-deep .el-dialog__body {
  padding: 24px;
}

.edit-form .el-form-item {
  margin-bottom: 20px;
}

.edit-form .el-form-item:last-child {
  margin-bottom: 0;
}

.edit-form ::v-deep .el-form-item__label {
  font-size: 14px;
  font-weight: 500;
  color: var(--classroom-text);
  line-height: 36px;
}

.edit-form ::v-deep .el-input__inner {
  height: 36px;
  line-height: 36px;
  font-size: 14px;
}

.gender-radio-group {
  display: flex;
  gap: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  border-top: 1px solid var(--classroom-border);
}

/* 表格样式 */
.students-panel ::v-deep .el-table {
  font-size: 14px;
  border-radius: 8px;
  overflow: hidden;
}

.students-panel ::v-deep .el-table th {
  font-size: 14px;
  font-weight: 600;
  background: #E3F2FD;
  color: var(--classroom-text);
  border-bottom: 2px solid var(--classroom-primary);
  padding: 16px 12px;
  text-align: left;
}

.students-panel ::v-deep .el-table td {
  font-size: 14px;
  border-bottom: 1px solid var(--classroom-border);
  padding: 14px 12px;
}

.students-panel ::v-deep .el-table--striped .el-table__body tr.el-table__row--striped td {
  background: var(--classroom-hover);
}

.students-panel ::v-deep .el-table__body tr:hover > td {
  background: var(--classroom-hover) !important;
}

/* 添加学生对话框样式 */
.add-student-dialog {
  padding: 0;
}

.search-section {
  margin-bottom: 20px;
}

.search-section .search-input {
  width: 100%;
}

.search-results {
  margin-bottom: 20px;
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--classroom-border);
  border-radius: 8px;
  padding: 12px;
}

.results-header {
  font-size: 14px;
  font-weight: 600;
  color: var(--classroom-text);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--classroom-border);
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--classroom-border);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

.user-item:hover {
  background: var(--classroom-hover);
  border-color: var(--classroom-primary);
}

.user-item.selected {
  background: #E3F2FD;
  border-color: var(--classroom-primary);
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--classroom-text);
  margin-bottom: 4px;
}

.user-details {
  font-size: 13px;
  color: var(--classroom-text-secondary);
  margin-bottom: 2px;
}

.user-details .nickname {
  color: #999;
  margin-left: 4px;
}

.user-email {
  font-size: 12px;
  color: var(--classroom-text-secondary);
}

.selected-icon {
  font-size: 20px;
  color: var(--classroom-primary);
}

.student-form {
  margin-top: 20px;
}

.selected-user-display {
  font-size: 14px;
  color: var(--classroom-text);
  padding: 8px 12px;
  background: #F5F5F5;
  border-radius: 4px;
}

/* 响应式 */
@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 16px;
  }

  .header .classroom-btn {
    width: 100%;
  }

  .stats-row {
    grid-template-columns: 1fr;
  }

  .action-buttons {
    flex-direction: column;
  }

  .action-buttons .classroom-btn {
    width: 100%;
  }
}
</style>
