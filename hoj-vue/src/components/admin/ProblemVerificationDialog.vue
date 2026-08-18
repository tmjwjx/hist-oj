<template>
  <el-dialog
    title="测试数据同步与标准程序验题"
    :visible.sync="visibleProxy"
    width="720px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
  >
    <div v-if="status" class="verification-body">
      <el-button v-if="showAi" size="small" plain type="primary" class="ai-button" @click="openAi">一键 AI 生成程序并验题</el-button>
      <el-alert
        :title="modeText"
        :description="syncDescription"
        :type="status.syncStatus === 1 ? 'success' : 'warning'"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.syncStatus === 2"
        class="verification-alert"
        title="测试数据同步失败"
        description="请确认判题服务器在线后重试，同步成功后才能提交标准程序。"
        type="error"
        show-icon
        :closable="false"
      />
      <el-form v-else label-width="90px" class="verification-form">
        <el-form-item label="提交语言">
          <el-select v-model="form.language" @change="formChanged" placeholder="选择语言" style="width: 220px">
            <el-option v-for="language in languageOptions" :key="language" :label="language" :value="language" />
          </el-select>
          <el-button class="last-code-button" plain @click="lastPassedVisible = true">查看上次通过代码</el-button>
        </el-form-item>
        <el-form-item label="标准程序">
            <el-input v-model="form.code" @input="formChanged" type="textarea" :rows="12" placeholder="提交能够通过全部测试数据的标准程序" />
        </el-form-item>
      </el-form>
      <el-alert
        v-if="status.sampleVerified"
        class="verification-alert"
        title="样例已通过正式判题"
        description="样例不是单独调用测试接口，而是随普通、SPJ 或交互题的完整判题流程一起验证。"
        type="success"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 1"
        class="verification-alert"
        :title="status.judgeStatusText || 'Running on test 1'"
        :description="status.judgeMessage || '标准程序正在正式判题，页面会自动刷新测试点进度。'"
        type="info"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 3"
        class="verification-alert"
        :title="status.judgeStatusText || '标准程序未通过'"
        :description="status.judgeMessage || '请修改程序后重新提交'"
        type="error"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 2"
        class="verification-alert"
        :title="status.judgeStatusText || 'Accepted'"
        description="题目已经可以用于比赛。"
        type="success"
        show-icon
        :closable="false"
      />
    </div>
    <div v-else class="verification-loading">正在读取验题状态…</div>
    <span slot="footer">
      <el-button v-if="hasDraft" type="danger" plain @click="deleteDraft">删除草稿及测试点</el-button>
      <el-button v-if="status && status.syncStatus === 2" @click="retrySync" :loading="loading">重新同步</el-button>
      <el-button @click="cancel">返回编辑</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="!canSubmit"
        @click="submit"
      >提交标准程序</el-button>
    </span>
    <ProblemAIValidationDialog
      v-if="aiVisible"
      :visible="aiVisible"
      :pid="pid"
      :language="form.language"
      :standard-program="form.code"
      :auto-run="true"
      @program-generated="useAiGeneratedProgram"
      @close="aiVisible = false"
    />
    <LastPassedCodeDialog
      v-if="lastPassedVisible"
      :visible="lastPassedVisible"
      :pid="pid"
      @use="useLastPassedCode"
      @close="lastPassedVisible = false"
    />
  </el-dialog>
</template>

<script>
import api from '@/common/api'
import ProblemAIValidationDialog from './ProblemAIValidationDialog.vue'
import LastPassedCodeDialog from './LastPassedCodeDialog.vue'

export default {
  name: 'ProblemVerificationDialog',
  components: { ProblemAIValidationDialog, LastPassedCodeDialog },
  props: {
    visible: Boolean,
    pid: { type: [Number, String], required: true },
    languages: { type: Array, default: () => ['C++'] },
    showAi: { type: Boolean, default: true },
    initialLanguage: { type: String, default: '' },
    initialCode: { type: String, default: '' }
  },
  data() {
    return {
      status: null,
      form: { language: '', code: '' },
      loading: false,
      submitting: false,
      timer: null,
      draftTimer: null,
      draftLoaded: false,
      hasDraft: false,
      explicitExit: false,
      aiVisible: false,
      lastPassedVisible: false
    }
  },
  computed: {
    visibleProxy: {
      get() { return this.visible },
      set(value) { if (!value) this.cancel() }
    },
    languageOptions() {
      return Array.from(new Set([this.form.language, ...(this.languages.length ? this.languages : ['C++'])].filter(Boolean)))
    },
    canSubmit() {
      return this.status && this.status.canSubmit && !this.submitting
    },
    modeText() {
      if (!this.status) return '验题状态'
      if (this.status.judgeMode === 'spj') return '特殊判题题：标准程序会使用已保存的 SPJ 进行校验'
      if (this.status.judgeMode === 'interactive') return '交互题：标准程序会使用已保存的交互程序进行校验'
      return '普通题：标准程序必须通过全部标准输出测试点'
    },
    syncDescription() {
      if (!this.status) return '正在读取测试数据状态'
      if (this.status.syncStatus === 1) return '测试数据同步成功，可提交标准程序进行正式判题。'
      return this.status.syncMessage || '正在读取测试数据状态'
    }
  },
  watch: {
    visible(value) {
      if (value) {
        this.explicitExit = false
        this.draftLoaded = false
        this.load()
      }
      else this.stopPolling()
    }
  },
  mounted() {
    window.addEventListener('beforeunload', this.handleBeforeUnload)
    if (this.visible) this.load()
  },
  beforeDestroy() {
    window.removeEventListener('beforeunload', this.handleBeforeUnload)
    if (!this.explicitExit) this.persistDraft(true)
    this.stopPolling()
    this.stopDraftTimer()
  },
  methods: {
    async load() {
      if (!this.pid) return
      this.loading = true
      try {
        const res = await api.admin_getProblemVerification(this.pid)
        this.status = res.data.data
        if (!this.draftLoaded) {
          const draft = (await api.admin_getProblemVerificationDraft(this.pid)).data.data
          this.hasDraft = Boolean(draft)
          if (draft) {
            this.form.language = draft.language || this.form.language
            this.form.code = draft.code || this.form.code
          } else {
            this.form.language = this.initialLanguage || this.form.language
            this.form.code = this.initialCode || this.form.code
          }
          this.draftLoaded = true
        }
        if (!this.form.language) this.form.language = this.languageOptions[0]
        if (this.status && this.status.verificationStatus === 1) this.startPolling()
        else this.stopPolling()
        if (this.status && this.status.verificationStatus === 2) {
          this.explicitExit = true
          this.$emit('passed')
        }
      } finally {
        this.loading = false
      }
    },
    startPolling() {
      if (!this.timer) this.timer = setInterval(this.load, 2000)
    },
    stopPolling() {
      if (this.timer) clearInterval(this.timer)
      this.timer = null
    },
    async retrySync() {
      this.loading = true
      try {
        await api.admin_retryProblemVerificationSync(this.pid)
        await this.load()
      } finally {
        this.loading = false
      }
    },
    async submit() {
      if (!this.form.code.trim()) return this.$message.warning('请填写标准程序')
      this.submitting = true
      try {
        await api.admin_submitProblemVerification({ pid: this.pid, ...this.form })
        await api.admin_clearProblemVerificationDraft(this.pid)
        this.hasDraft = false
        await this.load()
      } finally {
        this.submitting = false
      }
    },
    openAi() {
      if (!this.form.language) return this.$message.warning('请先选择标准程序语言')
      if (!this.status || this.status.syncStatus !== 1) return this.$message.warning('测试数据尚未同步完成')
      this.aiVisible = true
    },
    useAiGeneratedProgram(program) {
      if (!program || !program.code) return
      this.form.language = program.language || this.form.language
      this.form.code = program.code
      this.formChanged()
    },
    async useLastPassedCode(record) {
      if (this.form.code.trim() && this.form.code !== record.code) {
        try {
          await this.$confirm('填入历史代码会覆盖当前编辑内容，是否继续？', '使用上次通过代码', {
            confirmButtonText: '确认填入', cancelButtonText: '取消', type: 'warning'
          })
        } catch (e) { return }
      }
      this.form.language = record.language
      this.form.code = record.code
      this.formChanged()
      this.$message.success('已按历史提交语言填入上次通过代码')
    },
    cancel() {
      this.stopDraftTimer()
      if (!this.form.code.trim()) {
        this.explicitExit = true
        this.stopPolling()
        this.$emit('cancel')
        return
      }
      this.$confirm('退出前是否保存标准程序草稿？', '提示', {
        confirmButtonText: '保存并退出',
        cancelButtonText: '不保存退出',
        distinguishCancelAndClose: true,
        type: 'warning'
      }).then(async () => {
        await this.persistDraft(false)
        this.explicitExit = true
        this.stopPolling()
        this.$emit('cancel')
      }).catch(async (action) => {
        if (action !== 'cancel') return
        await api.admin_clearProblemVerificationDraft(this.pid)
        this.hasDraft = false
        this.explicitExit = true
        this.stopPolling()
        this.$emit('cancel')
      })
    },
    async deleteDraft() {
      try {
        await this.$confirm(
          '删除草稿会同时删除该题本地及所有判题服务器上的测试点，之后必须重新上传测试数据。是否继续？',
          '删除草稿及测试点',
          { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'error' }
        )
      } catch (e) {
        return
      }
      this.stopDraftTimer()
      await api.admin_deleteProblemVerificationDraft(this.pid)
      this.form.code = ''
      this.hasDraft = false
      this.$message.success('草稿和对应测试点已删除')
      await this.load()
    },
    persistDraft(keepalive) {
      if (!this.pid || !this.form.code.trim()) return Promise.resolve()
      const payload = JSON.stringify({
        pid: this.pid,
        language: this.form.language,
        code: this.form.code
      })
      if (keepalive) {
        const token = localStorage.getItem('token')
        const headers = { 'Content-Type': 'application/json', 'Url-Type': 'admin' }
        if (token) headers.Authorization = token
        return fetch('/api/admin/problem/verification/draft', {
          method: 'POST', headers, body: payload, keepalive: true
        }).catch(() => {})
      }
      return api.admin_saveProblemVerificationDraft({
        pid: this.pid, language: this.form.language, code: this.form.code
      }).then(() => { this.hasDraft = true }).catch(() => {})
    },
    scheduleDraft() {
      if (this.draftTimer) clearTimeout(this.draftTimer)
      this.draftTimer = setTimeout(() => this.persistDraft(false), 700)
    },
    formChanged() {
      this.$emit('change', { ...this.form })
      this.scheduleDraft()
    },
    stopDraftTimer() {
      if (this.draftTimer) clearTimeout(this.draftTimer)
      this.draftTimer = null
    },
    handleBeforeUnload() {
      if (!this.explicitExit) this.persistDraft(true)
    }
  }
}
</script>

<style scoped>
.verification-alert { margin-top: 14px; }
.verification-form { margin-top: 18px; }
.verification-loading { min-height: 180px; padding-top: 80px; text-align: center; color: #909399; }
.last-code-button { margin-left: 10px; }
</style>
