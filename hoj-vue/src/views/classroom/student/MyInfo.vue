<template>
  <div class="my-info-container">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h3>我的班级信息</h3>
    </div>

    <el-card v-loading="loading" class="info-card">
      <el-form :model="form" :rules="rules" ref="infoForm" label-width="120px">
        <el-form-item label="真实姓名" prop="realName">
          <el-input v-model="form.realName" placeholder="请输入真实姓名" maxlength="50"></el-input>
        </el-form-item>

        <el-form-item label="性别" prop="gender">
          <el-radio-group v-model="form.gender">
            <el-radio label="男">男</el-radio>
            <el-radio label="女">女</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="班级" prop="studentClass">
          <el-input v-model="form.studentClass" placeholder="例如：高一1班" maxlength="100"></el-input>
        </el-form-item>

        <el-form-item label="学号" prop="studentNo">
          <el-input v-model="form.studentNo" placeholder="请输入学号" maxlength="50"></el-input>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ submitting ? '保存中...' : '保存修改' }}
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 说明卡片 -->
    <el-card class="tip-card">
      <div slot="header">
        <span><i class="el-icon-info"></i> 说明</span>
      </div>
      <ul class="tip-list">
        <li>以上信息仅在当前班级中可见，不会影响您在其他班级的信息</li>
        <li>真实姓名和学号等信息将用于作业提交和成绩记录</li>
        <li>请确保信息真实准确，便于教师识别和管理</li>
      </ul>
    </el-card>
  </div>
</template>

<script>
import api from '@/api/classroom'

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'MyInfo',
  mixins: [studentAuth],
  data() {
    return {
      loading: false,
      submitting: false,
      classroomId: null,
      form: {
        realName: '',
        gender: '',
        studentClass: '',
        studentNo: ''
      },
      rules: {
        realName: [
          { required: true, message: '请输入真实姓名', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ]
      }
    }
  },
  mounted() {
    this.classroomId = this.$route.params.classroomId
    this.loadMyInfo()
  },
  methods: {
    async loadMyInfo() {
      if (!this.classroomId) {
        this.$message.error('班级ID不能为空')
        return
      }

      this.loading = true
      try {
        const res = await api.getClassroomStudentInfo(this.classroomId)
        if (res.data && res.data.code === 200) {
          const info = res.data.data
          this.form = {
            realName: info.realName || '',
            gender: info.gender || '',
            studentClass: info.studentClass || '',
            studentNo: info.studentNo || ''
          }
        } else {
          this.$message.error(res.data?.msg || '加载失败')
        }
      } catch (error) {
        console.error('加载班级信息失败:', error)
        this.$message.error('加载失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
      }
    },
    submitForm() {
      this.$refs.infoForm.validate(async (valid) => {
        if (!valid) {
          this.$message.warning('请填写所有必填项')
          return
        }

        this.submitting = true
        try {
          const res = await api.updateClassroomStudentInfo(this.classroomId, this.form)
          if (res.data && res.data.code === 200) {
            this.$message.success('保存成功')
          } else {
            this.$message.error(res.data?.msg || '保存失败')
          }
        } catch (error) {
          console.error('保存班级信息失败:', error)
          this.$message.error('保存失败: ' + (error.message || '未知错误'))
        } finally {
          this.submitting = false
        }
      })
    },
    resetForm() {
      this.$confirm('确定要重置表单吗？所有未保存的修改将丢失。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loadMyInfo()
      }).catch(() => {})
    },
    goBack() {
      this.$router.back()
    }
  }
}
</script>

<style scoped>
.my-info-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h3 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.info-card {
  margin-bottom: 20px;
}

.tip-card {
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
}

.tip-list {
  margin: 0;
  padding-left: 20px;
  color: #606266;
  line-height: 1.8;
}

.tip-list li {
  margin-bottom: 8px;
}
</style>
