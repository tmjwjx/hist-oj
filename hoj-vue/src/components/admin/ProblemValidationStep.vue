<template>
  <div class="validation-step">
    <el-alert title="第三步：验证题目"
      description="AI 验题会独立生成标准程序并正式跑完全部测试点；创建者验题使用下方代码，二者分别记录。"
      type="info" :closable="false" show-icon />
    <div class="validation-mode">
      <el-tag type="warning">AI 验题</el-tag>
      <el-tag type="primary">题目创建验证</el-tag>
      <span>两类评测会分别记录，不计入用户题库提交统计。</span>
    </div>

    <el-alert v-if="status && status.syncStatus !== 1" class="status-alert"
      :title="status.syncStatus === 2 ? '测试数据导入失败' : '测试数据正在导入'"
      :description="status.syncMessage" :type="status.syncStatus === 2 ? 'error' : 'warning'"
      :closable="false" show-icon>
      <el-button v-if="status.syncStatus === 2" size="mini" @click="retrySync">重新导入</el-button>
    </el-alert>

    <el-form label-position="top" class="standard-program-form">
      <el-row :gutter="16">
        <el-col :md="8" :xs="24">
          <el-form-item label="标准程序语言">
            <el-select v-model="language" @change="scheduleDraft" style="width: 100%">
              <el-option v-for="item in languageOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :md="16" :xs="24" class="history-action">
          <el-button plain icon="el-icon-document" @click="lastPassedVisible = true">显示上次通过代码</el-button>
        </el-col>
      </el-row>
      <el-form-item label="标准程序">
        <el-input v-model="code" @input="scheduleDraft" type="textarea" :rows="14"
          placeholder="填写用于 AI 验题和创建者正式验题的标准程序" />
      </el-form-item>
      <div class="validation-actions">
        <el-button v-if="showAi" type="primary" plain @click="openAi">一键 AI 生成程序并验题</el-button>
        <el-button plain icon="el-icon-time" @click="openHistory">历史 AI 验题记录</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCreator">提交代码验题</el-button>
      </div>
    </el-form>

    <el-alert v-if="status && status.verificationStatus === 1" class="status-alert"
      :title="status.judgeStatusText || 'Running on test 1'"
      :description="status.judgeMessage || '标准程序正在正式判题，页面会自动刷新测试点进度。'"
      type="info" :closable="false" show-icon />
    <el-alert v-if="status && status.verificationStatus === 3" class="status-alert"
      :title="status.judgeStatusText || '标准程序未通过'"
      :description="status.judgeMessage || '请修改程序后重新提交。'"
      type="error" :closable="false" show-icon />
    <el-alert v-if="status && status.verificationStatus === 2" class="status-alert"
      :title="status.judgeStatusText || 'Accepted'" description="标准程序和样例已通过正式判题。"
      type="success" :closable="false" show-icon />

    <div class="step-actions">
      <el-button @click="back">返回测试数据</el-button>
      <el-button v-if="hasDraft" type="danger" plain @click="deleteDraft">删除草稿及测试点</el-button>
      <el-button type="success" :disabled="!verified" @click="finish">完成题目创建</el-button>
    </div>

    <ProblemAIValidationDialog v-if="aiVisible" :visible="aiVisible" :pid="pid"
      :language="language" :standard-program="code" :auto-run="true"
      @verified="handleVerified" @program-generated="useAiGeneratedProgram" @close="aiVisible = false" />
    <LastPassedCodeDialog v-if="lastPassedVisible" :visible="lastPassedVisible" :pid="pid"
      @use="useLastPassedCode" @close="lastPassedVisible = false" />
  </div>
</template>

<script>
import ProblemAIValidationDialog from './ProblemAIValidationDialog.vue'
import LastPassedCodeDialog from './LastPassedCodeDialog.vue'
import api from '@/common/api'

export default {
  name: 'ProblemValidationStep',
  components: { ProblemAIValidationDialog, LastPassedCodeDialog },
  props: {
    pid: { type: [Number, String], required: true },
    languages: { type: Array, default: () => [] },
    showAi: { type: Boolean, default: true }
  },
  data() {
    return {
      language: '', code: '', status: null, verified: false, hasDraft: false,
      aiVisible: false, lastPassedVisible: false, submitting: false,
      draftTimer: null, pollTimer: null, completed: false, skipDraft: false
    }
  },
  computed: {
    languageOptions() {
      const values = [this.language, ...this.languages].filter(Boolean)
      return Array.from(new Set(values.length ? values : ['C++']))
    }
  },
  watch: {
    languages() { if (!this.language) this.language = this.languageOptions[0] },
    pid() { this.initialize() }
  },
  async mounted() {
    window.addEventListener('beforeunload', this.handleBeforeUnload)
    await this.initialize()
  },
  beforeDestroy() {
    window.removeEventListener('beforeunload', this.handleBeforeUnload)
    this.stopTimers()
    if (!this.completed && !this.skipDraft) this.persistDraft(true)
  },
  methods: {
    async initialize() {
      if (!this.pid) return
      this.language = this.language || this.languageOptions[0]
      await Promise.all([this.loadStatus(), this.loadDraft()])
    },
    async loadDraft() {
      try {
        const draft = (await api.admin_getProblemVerificationDraft(this.pid)).data.data
        this.hasDraft = Boolean(draft)
        if (draft) {
          this.language = draft.language || this.language
          this.code = draft.code || ''
        }
      } catch (e) {}
    },
    async loadStatus() {
      const res = await api.admin_getProblemVerification(this.pid)
      this.status = res.data.data
      this.verified = Boolean(this.status && this.status.verified)
      if (this.status && this.status.verificationStatus === 1) this.startPolling()
      else this.stopPolling()
    },
    validateProgram() {
      if (!this.language) { this.$message.warning('请选择标准程序语言'); return false }
      if (!this.code.trim()) { this.$message.warning('请先填写标准程序'); return false }
      if (!this.status || this.status.syncStatus !== 1) {
        this.$message.warning('测试数据尚未导入完成，请稍后重试')
        return false
      }
      return true
    },
    openAi() {
      if (!this.language) return this.$message.warning('请选择标准程序语言')
      if (!this.status || this.status.syncStatus !== 1) {
        return this.$message.warning('测试数据尚未导入完成，请稍后重试')
      }
      this.aiVisible = true
    },
    useAiGeneratedProgram(program) {
      if (!program || !program.code) return
      this.language = program.language || this.language
      this.code = program.code
      this.scheduleDraft()
    },
    openHistory() {
      this.$router.push({ name: 'admin-problem-ai-history', params: { problemId: this.pid } })
    },
    async submitCreator() {
      if (!this.validateProgram()) return
      this.submitting = true
      try {
        await api.admin_submitProblemVerification({ pid: this.pid, language: this.language, code: this.code,
          verificationType: 'creator_validation' })
        await api.admin_clearProblemVerificationDraft(this.pid).catch(() => {})
        this.hasDraft = false
        await this.loadStatus()
      } finally { this.submitting = false }
    },
    handleVerified() {
      this.verified = true
      this.loadStatus().catch(() => {})
    },
    async retrySync() {
      await api.admin_retryProblemVerificationSync(this.pid)
      await this.loadStatus()
    },
    startPolling() {
      if (!this.pollTimer) this.pollTimer = setInterval(() => this.loadStatus().catch(() => {}), 1500)
    },
    stopPolling() {
      if (this.pollTimer) clearInterval(this.pollTimer)
      this.pollTimer = null
    },
    async useLastPassedCode(record) {
      if (this.code.trim() && this.code !== record.code) {
        try {
          await this.$confirm('填入历史代码会覆盖当前内容，是否继续？', '使用上次通过代码', {
            confirmButtonText: '确认填入', cancelButtonText: '取消', type: 'warning'
          })
        } catch (e) { return }
      }
      this.language = record.language || this.language
      this.code = record.code || ''
      this.scheduleDraft()
    },
    async finish() {
      this.completed = true
      this.stopTimers()
      await api.admin_clearProblemVerificationDraft(this.pid).catch(() => {})
      this.$emit('passed')
    },
    back() {
      if (!this.code.trim()) return this.$emit('back')
      this.$confirm('返回测试数据前是否保存标准程序草稿？', '提示', {
        confirmButtonText: '保存并返回', cancelButtonText: '不保存返回',
        distinguishCancelAndClose: true, type: 'warning'
      }).then(async () => {
        await this.persistDraft(false); this.skipDraft = true; this.$emit('back')
      }).catch(async action => {
        if (action !== 'cancel') return
        await api.admin_clearProblemVerificationDraft(this.pid).catch(() => {})
        this.skipDraft = true; this.$emit('back')
      })
    },
    scheduleDraft() {
      if (this.draftTimer) clearTimeout(this.draftTimer)
      this.draftTimer = setTimeout(() => this.persistDraft(false), 700)
    },
    persistDraft(keepalive) {
      if (!this.code.trim()) return Promise.resolve()
      const payload = { pid: this.pid, language: this.language, code: this.code }
      if (!keepalive) return api.admin_saveProblemVerificationDraft(payload)
        .then(() => { this.hasDraft = true }).catch(() => {})
      const token = localStorage.getItem('token')
      const headers = { 'Content-Type': 'application/json', 'Url-Type': 'admin' }
      if (token) headers.Authorization = token
      return fetch('/api/admin/problem/verification/draft', {
        method: 'POST', headers, body: JSON.stringify(payload), keepalive: true
      }).catch(() => {})
    },
    async deleteDraft() {
      try {
        await this.$confirm('删除草稿会同时删除该题目的测试点，之后必须重新导入测试数据。是否继续？',
          '删除草稿及测试点', { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'error' })
      } catch (e) { return }
      await api.admin_deleteProblemVerificationDraft(this.pid)
      this.code = ''; this.hasDraft = false
      await this.loadStatus()
    },
    handleBeforeUnload() { this.persistDraft(true) },
    stopTimers() {
      if (this.draftTimer) clearTimeout(this.draftTimer)
      this.draftTimer = null
      this.stopPolling()
    }
  }
}
</script>

<style scoped>
.standard-program-form { margin-top: 18px; }
.validation-mode { display: flex; align-items: center; gap: 8px; margin-top: 12px; color: #909399; }
.history-action { padding-top: 40px; }
.validation-actions { display: flex; justify-content: flex-end; gap: 10px; flex-wrap: wrap; }
.status-alert { margin-top: 14px; }
.step-actions { margin-top: 24px; display: flex; justify-content: space-between; gap: 10px; }
@media (max-width: 768px) { .history-action { padding-top: 0; margin-bottom: 12px; } }
</style>
