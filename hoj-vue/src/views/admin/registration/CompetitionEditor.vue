<template>
  <div class="competition-editor-page">
    <div class="container">
      <!-- 头部 -->
      <div class="header">
        <div>
          <h1>{{ isEdit ? '编辑比赛' : '创建新比赛' }}</h1>
          <p>填写比赛信息</p>
        </div>
        <div class="header-actions">
          <router-link to="/admin/registration" class="btn">返回列表</router-link>
        </div>
      </div>

      <!-- 表单 -->
      <form @submit.prevent="saveCompetition">
        <!-- ① 顶部基本信息区 -->
        <div class="basic-info-section">
          <div class="section-header">
            <h2>📌 基本信息</h2>
          </div>

          <div class="info-grid">
            <div class="form-group">
              <label>比赛名称 *</label>
              <input v-model="formData.name" type="text" required placeholder="请输入比赛名称">
            </div>

            <div class="form-row-inline">
              <div class="form-group">
                <label>开始时间 *</label>
                <input v-model="formData.startTime" type="datetime-local" required>
              </div>
              <div class="form-group">
                <label>结束时间 *</label>
                <input v-model="formData.endTime" type="datetime-local" required>
              </div>
            </div>

            <div class="form-group-inline">
              <div class="logo-section">
                <label>比赛Logo</label>
                <div class="logo-upload">
                  <div v-if="formData.logoUrl" class="logo-preview">
                    <img :src="formData.logoUrl" alt="Logo预览">
                    <button type="button" class="remove-btn" @click="removeLogo">删除</button>
                  </div>
                  <div v-else class="upload-placeholder" @click="triggerFileInput">
                    <span class="upload-icon">📤</span>
                    <span>点击上传Logo</span>
                    <span class="upload-hint">支持 JPG、PNG、GIF，最大5MB</span>
                  </div>
                  <input
                    ref="fileInput"
                    type="file"
                    accept="image/png,image/jpeg,image/gif"
                    style="display: none"
                    @change="handleLogoUpload"
                  >
                </div>
              </div>

              <div class="visibility-section">
                <label>比赛可见性</label>
                <label class="checkbox-item-inline">
                  <input type="checkbox" v-model="formData.visible">
                  <span>立即显示（可见）</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- ② Markdown 编辑区 -->
        <div class="markdown-section">
          <div class="section-header">
            <h2>比赛说明</h2>
          </div>
          <Editor :value.sync="formData.description" :allow-upload="true" class="competition-markdown-editor" />
        </div>

        <!-- ③ 底部配置区 -->
        <div class="fields-config-section">
          <div class="section-header">
            <h2>⚙️ 报名字段配置</h2>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-item" v-for="(label, field) in fieldOptions" :key="field">
              <input type="checkbox" v-model="formData.fields[field]">
              {{ label }}
            </label>
          </div>

          <!-- 操作按钮 -->
          <div class="form-actions">
            <router-link to="/admin/registration" class="btn">取消</router-link>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '保存中...' : (isEdit ? '更新比赛' : '创建比赛') }}
            </button>
          </div>
        </div>
      </form>
    </div>

  </div>
</template>

<script>
import mMessage from '@/common/message'
import Editor from '@/components/admin/Editor'

export default {
  name: 'CompetitionEditor',
  components: {
    Editor
  },
  data() {
    return {
      isEdit: false,
      competitionId: null,
      submitting: false,
      formData: {
        name: '',
        startTime: '',
        endTime: '',
        logoUrl: '',
        description: '',
        fields: {
          name: true,
          class: true,
          college: true,
          studentId: true,
          gender: true,
          shirtSize: true,
          teamName: true,
          qq: true
        },
        visible: true
      },
      fieldOptions: {
        name: '姓名',
        class: '班级',
        college: '学院',
        studentId: '学号',
        gender: '性别',
        shirtSize: '衣服尺码',
        teamName: '队伍名称',
        qq: 'QQ号'
      }
    }
  },
  mounted() {
    // 检查是否是编辑模式
    const competitionId = this.$route.params.id
    if (competitionId && competitionId !== 'create') {
      this.isEdit = true
      this.competitionId = competitionId
      this.loadCompetition(competitionId)
    }
  },
  methods: {
    async loadCompetition(id) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`/api/registration/competitions/${id}`, {
          headers: {
            'Authorization': token,
            'Url-Type': 'general'
          }
        })

        const result = await response.json()
        if (result.code === 200 || result.status === 200) {
          const comp = result.data || result
          this.formData = {
            name: comp.name || '',
            startTime: this.formatDateTimeLocal(comp.startTime),
            endTime: this.formatDateTimeLocal(comp.endTime),
            logoUrl: comp.logoUrl || '',
            description: comp.description || '',
            fields: comp.fields ? (typeof comp.fields === 'string' ? JSON.parse(comp.fields) : comp.fields) : this.formData.fields,
            visible: comp.visible !== undefined ? comp.visible : true
          }
          mMessage.success('比赛数据加载成功')
        } else {
          mMessage.error(result.message || result.msg || '加载比赛失败')
        }
      } catch (error) {
        console.error('加载比赛失败:', error)
        mMessage.error('加载比赛失败')
      }
    },

    validateForm() {
      // 验证开始时间
      if (!this.formData.startTime) {
        mMessage.error('请选择开始时间')
        return false
      }

      // 验证结束时间
      if (!this.formData.endTime) {
        mMessage.error('请选择结束时间')
        return false
      }

      // 将时间字符串转换为 Date 对象
      const startTime = new Date(this.formData.startTime)
      const endTime = new Date(this.formData.endTime)
      const now = new Date()

      // 验证开始时间不能晚于结束时间
      if (startTime >= endTime) {
        mMessage.error('开始时间必须早于结束时间')
        return false
      }

      // 验证结束时间不能早于当前时间（可选）
      if (endTime < now) {
        mMessage.error('结束时间不能早于当前时间')
        return false
      }

      // 验证比赛时长至少 10 分钟
      const duration = endTime - startTime
      const minDuration = 10 * 60 * 1000 // 10分钟
      if (duration < minDuration) {
        mMessage.error('比赛时长至少需要 10 分钟')
        return false
      }

      // 验证比赛名称
      if (!this.formData.name || this.formData.name.trim() === '') {
        mMessage.error('请输入比赛名称')
        return false
      }

      return true
    },

    async saveCompetition() {
      // 验证表单
      if (!this.validateForm()) {
        return
      }

      this.submitting = true
      try {
        const token = localStorage.getItem('token')

        // 准备数据
        const data = {
          name: this.formData.name,
          startTime: new Date(this.formData.startTime).toISOString(),
          endTime: new Date(this.formData.endTime).toISOString(),
          logoUrl: this.formData.logoUrl,
          description: this.formData.description,
          fields: JSON.stringify(this.formData.fields),
          visible: this.formData.visible
        }

        let response
        if (this.isEdit) {
          // 更新比赛
          response = await fetch(`/api/registration/admin/competitions/${this.competitionId}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': token,
              'Url-Type': 'general'
            },
            body: JSON.stringify(data)
          })
        } else {
          // 创建比赛
          response = await fetch('/api/registration/admin/competitions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': token,
              'Url-Type': 'general'
            },
            body: JSON.stringify(data)
          })
        }

        const result = await response.json()
        if (result.code === 200) {
          mMessage.success(this.isEdit ? '比赛更新成功' : '比赛创建成功')
          this.$router.push('/admin/registration')
        } else {
          mMessage.error(result.message || '操作失败')
        }
      } catch (error) {
        console.error('保存比赛失败:', error)
        mMessage.error('保存失败，请重试')
      } finally {
        this.submitting = false
      }
    },

    triggerFileInput() {
      this.$refs.fileInput.click()
    },

    async handleLogoUpload(event) {
      const file = event.target.files[0]
      if (!file) return

      // 验证文件类型
      const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/gif']
      if (!allowedTypes.includes(file.type)) {
        mMessage.error('只支持 PNG、JPG、GIF 格式的图片')
        return
      }

      // 验证文件大小（5MB）
      if (file.size > 5 * 1024 * 1024) {
        mMessage.error('图片大小不能超过 5MB')
        return
      }

      // 上传文件
      try {
        const formData = new FormData()
        formData.append('image', file)

        const token = localStorage.getItem('token')
        const response = await fetch('/api/registration/admin/upload/logo', {
          method: 'POST',
          headers: {
            'Authorization': token,
            'Url-Type': 'general'
          },
          body: formData
        })

        const result = await response.json()
        if (result.status === 200) {
          this.formData.logoUrl = result.data.url
          mMessage.success('Logo上传成功')
        } else {
          mMessage.error(result.message || '上传失败')
        }
      } catch (error) {
        console.error('上传失败:', error)
        mMessage.error('上传失败，请重试')
      }

      // 清空input，允许重复上传同一文件
      event.target.value = ''
    },

    removeLogo() {
      this.formData.logoUrl = ''
    },

    formatDateTimeLocal(isoString) {
      if (!isoString) return ''
      const date = new Date(isoString)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day}T${hours}:${minutes}`
    },

    formatTime(str) {
      if (!str) return ''
      const date = new Date(str)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
        timeZone: 'Asia/Shanghai'
      })
    }
  }
}
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.competition-editor-page {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f7fa;
  min-height: 100vh;
  padding: 20px;
}

.container {
  max-width: 1600px;
  margin: 0 auto;
}

/* 头部 */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #409EFF;
  margin-bottom: 6px;
}

.header p {
  color: #718096;
  font-size: 14px;
}

/* ===== ① 顶部基本信息区 ===== */
.basic-info-section {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.section-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #409EFF;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-group input[type="text"],
.form-group input[type="datetime-local"] {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.form-group input:focus {
  outline: none;
  border-color: #409EFF;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.form-row-inline {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-group-inline {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
  align-items: start;
}

/* Logo 上传 */
.logo-section {
  display: flex;
  flex-direction: column;
}

.logo-upload {
  margin-top: 8px;
}

.logo-preview {
  position: relative;
  width: 200px;
  height: 200px;
  border-radius: 8px;
  overflow: hidden;
  border: 2px dashed #e2e8f0;
  background: #f7fafc;
}

.logo-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.remove-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.95);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.2s;
}

.remove-btn:hover {
  background: #fff;
  transform: scale(1.05);
}

.upload-placeholder {
  width: 200px;
  height: 200px;
  border: 2px dashed #cbd5e0;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #f7fafc;
}

.upload-placeholder:hover {
  border-color: #409EFF;
  background: #edf4fc;
}

.upload-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.upload-hint {
  font-size: 12px;
  color: #718096;
  margin-top: 4px;
}

/* 可见性 */
.visibility-section {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  height: 100%;
}

.checkbox-item-inline {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e2e8f0;
  margin-top: 28px;
}

.checkbox-item-inline:hover {
  background: #edf4fc;
  border-color: #409EFF;
}

.checkbox-item-inline input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-item-inline span {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

.markdown-section {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.competition-markdown-editor /deep/ .v-note-wrapper {
  min-height: 520px;
}

/* ===== ③ 底部字段配置区 ===== */
.fields-config-section {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.checkbox-item:hover {
  background: #edf4fc;
  border-color: #409EFF;
}

.checkbox-item input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-item span {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

/* 操作按钮 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.btn {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
  text-align: center;
  transition: all 0.3s;
  background: #e2e8f0;
  color: #4a5568;
}

.btn:hover {
  background: #cbd5e0;
  transform: translateY(-1px);
}

.btn-primary {
  background: #409EFF;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #66b1ff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* ===== 响应式设计 ===== */

/* 平板横屏：仍然并排，但调整间距 */
@media (max-width: 1200px) {
  .container {
    max-width: 100%;
  }

  .form-group-inline {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .visibility-section {
    margin-top: 0;
  }

  .checkbox-item-inline {
    margin-top: 0;
  }
}

@media (max-width: 992px) {
  .form-row-inline {
    grid-template-columns: 1fr;
  }

  .checkbox-group {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  }
}

/* 手机端：进一步优化 */
@media (max-width: 768px) {
  .competition-editor-page {
    padding: 12px;
  }

  .header {
    padding: 20px;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header h1 {
    font-size: 24px;
  }

  .basic-info-section,
  .markdown-section,
  .fields-config-section {
    padding: 20px;
  }

  .checkbox-group {
    grid-template-columns: 1fr;
  }

  .competition-markdown-editor /deep/ .v-note-wrapper {
    min-height: 400px;
  }

  .form-actions {
    flex-direction: column-reverse;
  }

  .form-actions .btn {
    width: 100%;
  }
}

/* 小屏手机 */
@media (max-width: 480px) {
  .header h1 {
    font-size: 20px;
  }

  .upload-placeholder {
    width: 100%;
    height: 160px;
  }

  .logo-preview {
    width: 100%;
    height: 160px;
  }

}
</style>
