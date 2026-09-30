<template>
  <div class="my-info-container">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h3>{{ $t('m.SMy_Title') }}</h3>
    </div>

    <el-card v-loading="loading" class="info-card">
      <el-form :model="form" :rules="rules" ref="infoForm" label-width="120px">
        <el-form-item :label="$t('m.SMy_Real_Name')" prop="realName">
          <el-input v-model="form.realName" :placeholder="$t('m.SMy_Real_Name_Placeholder')" maxlength="50"></el-input>
        </el-form-item>

        <el-form-item :label="$t('m.SMy_Gender')" prop="gender">
          <el-radio-group v-model="form.gender">
            <el-radio label="男">{{ $t('m.Male') }}</el-radio>
            <el-radio label="女">{{ $t('m.Female') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="$t('m.SMy_Class')" prop="studentClass">
          <el-input v-model="form.studentClass" :placeholder="$t('m.SMy_Class_Placeholder')" maxlength="100"></el-input>
        </el-form-item>

        <el-form-item :label="$t('m.SMy_Student_No')" prop="studentNo">
          <el-input v-model="form.studentNo" :placeholder="$t('m.SMy_Student_No_Placeholder')" maxlength="50"></el-input>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ submitting ? $t('m.SMy_Saving') : $t('m.SMy_Save_Changes') }}
          </el-button>
          <el-button @click="resetForm">{{ $t('m.Reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 说明卡片 -->
    <el-card class="tip-card">
      <div slot="header">
        <span><i class="el-icon-info"></i> {{ $t('m.SMy_Notes') }}</span>
      </div>
      <ul class="tip-list">
        <li>{{ $t('m.SMy_Note_Visible_In_Class') }}</li>
        <li>{{ $t('m.SMy_Note_Used_For') }}</li>
        <li>{{ $t('m.SMy_Note_Accurate') }}</li>
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
          { required: true, message: this.$t('m.SMy_Real_Name_Placeholder'), trigger: 'blur' },
          { min: 2, max: 50, message: this.$t('m.SMy_Length_Range'), trigger: 'blur' }
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
        this.$message.error(this.$t('m.SMy_Class_Id_Required'))
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
          this.$message.error(res.data?.msg || this.$t('m.Load_Failed'))
        }
      } catch (error) {
        console.error('加载班级信息失败:', error)
        this.$message.error(this.$t('m.SMy_Load_Failed_Reason', { reason: error.message || this.$t('m.Unknown') }))
      } finally {
        this.loading = false
      }
    },
    submitForm() {
      this.$refs.infoForm.validate(async (valid) => {
        if (!valid) {
          this.$message.warning(this.$t('m.SMy_Fill_All_Required'))
          return
        }

        this.submitting = true
        try {
          const res = await api.updateClassroomStudentInfo(this.classroomId, this.form)
          if (res.data && res.data.code === 200) {
            this.$message.success(this.$t('m.Save_Success'))
          } else {
            this.$message.error(res.data?.msg || this.$t('m.Save_Failed'))
          }
        } catch (error) {
          console.error('保存班级信息失败:', error)
          this.$message.error(this.$t('m.SMy_Save_Failed_Reason', { reason: error.message || this.$t('m.Unknown') }))
        } finally {
          this.submitting = false
        }
      })
    },
    resetForm() {
      this.$confirm(this.$t('m.SMy_Reset_Confirm'), this.$t('m.SMy_Tip'), {
        confirmButtonText: this.$t('m.OK'),
        cancelButtonText: this.$t('m.Cancel'),
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
