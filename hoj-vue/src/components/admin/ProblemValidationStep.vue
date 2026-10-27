<template>
  <div class="validation-step">
    <el-alert :title="$t('m.Verif_Step_Title')"
      :description="$t($slots['save-action'] ? 'm.Verif_Save_Workflow_Desc' : 'm.Verif_Step_Desc')"
      type="info" :closable="false" show-icon />
    <div class="validation-mode">
      <el-tag type="warning">{{ $t('m.Verif_Tag_AI') }}</el-tag>
      <el-tag type="primary">{{ $t('m.Verif_Tag_Creator') }}</el-tag>
      <span>{{ $t('m.Verif_Mode_Note') }}</span>
    </div>

    <el-alert v-if="status && status.syncStatus !== 1" class="status-alert"
      :title="status.syncStatus === 2 ? $t('m.Verif_Import_Failed') : $t('m.Verif_Importing')"
      :description="status.syncMessage" :type="status.syncStatus === 2 ? 'error' : 'warning'"
      :closable="false" show-icon>
      <el-button v-if="status.syncStatus === 2" size="mini" @click="retrySync">{{ $t('m.Verif_Retry_Import') }}</el-button>
    </el-alert>

    <el-form label-position="top" class="standard-program-form">
      <el-row :gutter="16">
        <el-col :md="8" :xs="24">
          <el-form-item :label="$t('m.Verif_Program_Language')">
            <el-select v-model="language" @change="scheduleDraft" style="width: 100%">
              <el-option v-for="item in languageOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :md="16" :xs="24" class="history-action">
          <el-button plain icon="el-icon-document" @click="lastPassedVisible = true">{{ $t('m.Verif_Show_Last_Code') }}</el-button>
        </el-col>
      </el-row>
      <el-form-item :label="$t('m.Verif_Standard_Program')">
        <el-input v-model="code" @input="scheduleDraft" type="textarea" :rows="14"
          :placeholder="$t('m.Verif_Code_Placeholder')" />
      </el-form-item>
      <div class="validation-actions">
        <el-button v-if="showAi" type="primary" plain :disabled="saving || submitting" @click="openAi">{{ $t('m.Verif_One_Click_AI') }}</el-button>
        <el-button plain icon="el-icon-time" @click="openHistory">{{ $t('m.ProbAI_History_Dialog_Title') }}</el-button>
        <el-button type="primary" :loading="submitting" :disabled="saving" @click="submitCreator">{{ $t('m.Verif_Submit_Code') }}</el-button>
      </div>
    </el-form>

    <el-alert v-if="status && status.verificationStatus === 1" class="status-alert"
      :title="status.judgeStatusText || 'Running on test 1'"
      :description="status.judgeMessage || $t('m.Verif_Judging_Desc')"
      type="info" :closable="false" show-icon />
    <el-alert v-if="status && status.verificationStatus === 3" class="status-alert"
      :title="status.judgeStatusText || $t('m.Verif_Not_Passed')"
      :description="status.judgeMessage || $t('m.Verif_Modify_And_Resubmit')"
      type="error" :closable="false" show-icon />
    <el-alert v-if="status && status.verificationStatus === 2" class="status-alert"
      :title="status.judgeStatusText || 'Accepted'" :description="$t(pendingSave ? 'm.Prob_Pending_Ready_To_Save' : 'm.Verif_Passed_Desc')"
      type="success" :closable="false" show-icon />

    <div class="step-actions">
      <el-button :disabled="saving || submitting" @click="back">{{ $t('m.Verif_Back_To_Data') }}</el-button>
      <el-button v-if="hasDraft && !$slots['save-action']" type="danger" plain :disabled="saving || submitting" @click="deleteDraft">{{ $t('m.Verif_Delete_Draft') }}</el-button>
      <slot name="save-action">
        <el-button type="success" :disabled="!verified" @click="finish">{{ $t('m.Verif_Finish_Create') }}</el-button>
      </slot>
    </div>

    <ProblemAIValidationDialog v-if="aiVisible" :visible="aiVisible" :pid="pid"
      :language="language" :standard-program="code"
      @verified="handleVerified" @program-generated="useAiGeneratedProgram" @close="aiVisible = false" />
    <LastPassedCodeDialog v-if="lastPassedVisible" :visible="lastPassedVisible" :pid="lastPassedPid || pid"
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
    lastPassedPid: { type: [Number, String], default: null },
    pendingSave: { type: Boolean, default: false },
    languages: { type: Array, default: () => [] },
    showAi: { type: Boolean, default: true },
    saving: { type: Boolean, default: false }
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
    pid(newPid, oldPid) {
      this.stopTimers()
      if (oldPid) this.persistDraft(false, oldPid)
      this.status = null
      this.verified = false
      this.hasDraft = false
      this.aiVisible = false
      this.lastPassedVisible = false
      this.initialize(true).catch(() => {})
    }
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
    async initialize(preserveCode = false) {
      if (!this.pid) return
      this.language = this.language || this.languageOptions[0]
      await Promise.all([this.loadStatus(), this.loadDraft(preserveCode)])
    },
    async loadDraft(preserveCode = false) {
      const pid = this.pid;
      const previousCode = this.code;
      const previousLanguage = this.language;
      try {
        let draft = (await api.admin_getProblemVerificationDraft(pid)).data.data
        if (!draft && !preserveCode && this.lastPassedPid && this.lastPassedPid !== pid) {
          draft = (await api.admin_getProblemVerificationDraft(this.lastPassedPid)).data.data
        }
        if (pid !== this.pid) return
        this.hasDraft = Boolean(draft)
        if (preserveCode) return
        if (draft && this.code === previousCode && this.language === previousLanguage) {
          this.language = draft.language || this.language
          this.code = draft.code || ''
        }
      } catch (e) {}
    },
    async loadStatus() {
      const pid = this.pid
      const res = await api.admin_getProblemVerification(pid)
      if (pid === this.pid) this.applyStatus(res.data.data)
    },
    applyStatus(status) {
      this.status = status
      this.verified = Boolean(this.status && this.status.verified)
      if (this.status && (this.status.verificationStatus === 1 || this.status.syncStatus === 0)) this.startPolling()
      else this.stopPolling()
    },
    validateProgram() {
      if (!this.language) { this.$message.warning(this.$t('m.Verif_Select_Program_Language')); return false }
      if (!this.code.trim()) { this.$message.warning(this.$t('m.Verif_Code_Required')); return false }
      if (!this.status || this.status.syncStatus !== 1) {
        this.$message.warning(this.$t('m.Verif_Data_Importing_Warn'))
        return false
      }
      return true
    },
    openAi() {
      if (!this.language) return this.$message.warning(this.$t('m.Verif_Select_Program_Language'))
      if (!this.status || this.status.syncStatus !== 1) {
        return this.$message.warning(this.$t('m.Verif_Data_Importing_Warn'))
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
      if (this.saving || this.submitting || !this.validateProgram()) return
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
          await this.$confirm(this.$t('m.Verif_Confirm_Overwrite_Step'), this.$t('m.Verif_Use_Last_Code_Title'), {
            confirmButtonText: this.$t('m.Verif_Confirm_Fill'), cancelButtonText: this.$t('m.Cancel'), type: 'warning'
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
      if (this.$slots['save-action'] || !this.code.trim()) return this.$emit('back')
      this.$confirm(this.$t('m.Verif_Confirm_Save_Back'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.Verif_Save_Back'), cancelButtonText: this.$t('m.Verif_No_Save_Back'),
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
    persistDraft(keepalive, pid = this.pid) {
      if (!this.code.trim()) return Promise.resolve()
      const payload = { pid, language: this.language, code: this.code }
      if (!keepalive) return api.admin_saveProblemVerificationDraft(payload)
        .then(() => { if (pid === this.pid) this.hasDraft = true }).catch(() => {})
      const token = localStorage.getItem('token')
      const headers = { 'Content-Type': 'application/json', 'Url-Type': 'admin' }
      if (token) headers.Authorization = token
      return fetch('/api/admin/problem/verification/draft', {
        method: 'POST', headers, body: JSON.stringify(payload), keepalive: true
      }).catch(() => {})
    },
    async deleteDraft() {
      try {
        await this.$confirm(this.$t('m.Verif_Delete_Draft_Confirm_Step'),
          this.$t('m.Verif_Delete_Draft'), { confirmButtonText: this.$t('m.Confirm_Delete'), cancelButtonText: this.$t('m.Cancel'), type: 'error' })
      } catch (e) { return }
      if (this.draftTimer) clearTimeout(this.draftTimer)
      this.draftTimer = null
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
