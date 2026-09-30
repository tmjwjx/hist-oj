<template>
  <el-dialog
    :title="$t('m.Verif_Dialog_Title')"
    :visible.sync="visibleProxy"
    width="720px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
  >
    <div v-if="status" class="verification-body">
      <el-button v-if="showAi" size="small" plain type="primary" class="ai-button" @click="openAi">{{ $t('m.Verif_One_Click_AI') }}</el-button>
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
        :title="$t('m.Verif_Sync_Failed_Title')"
        :description="$t('m.Verif_Sync_Failed_Desc')"
        type="error"
        show-icon
        :closable="false"
      />
      <el-form v-else label-width="90px" class="verification-form">
        <el-form-item :label="$t('m.Verif_Language')">
          <el-select v-model="form.language" @change="formChanged" :placeholder="$t('m.Verif_Select_Language')" style="width: 220px">
            <el-option v-for="language in languageOptions" :key="language" :label="language" :value="language" />
          </el-select>
          <el-button class="last-code-button" plain @click="lastPassedVisible = true">{{ $t('m.Verif_View_Last_Code') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('m.Verif_Standard_Program')">
            <el-input v-model="form.code" @input="formChanged" type="textarea" :rows="12" :placeholder="$t('m.Verif_Program_Placeholder')" />
        </el-form-item>
      </el-form>
      <el-alert
        v-if="status.sampleVerified"
        class="verification-alert"
        :title="$t('m.Verif_Sample_Passed_Title')"
        :description="$t('m.Verif_Sample_Passed_Desc')"
        type="success"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 1"
        class="verification-alert"
        :title="status.judgeStatusText || 'Running on test 1'"
        :description="status.judgeMessage || $t('m.Verif_Judging_Desc')"
        type="info"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 3"
        class="verification-alert"
        :title="status.judgeStatusText || $t('m.Verif_Not_Passed')"
        :description="status.judgeMessage || $t('m.Verif_Modify_And_Resubmit')"
        type="error"
        show-icon
        :closable="false"
      />
      <el-alert
        v-if="status.verificationStatus === 2"
        class="verification-alert"
        :title="status.judgeStatusText || 'Accepted'"
        :description="$t('m.Verif_Accepted_Desc')"
        type="success"
        show-icon
        :closable="false"
      />
    </div>
    <div v-else class="verification-loading">{{ $t('m.Verif_Loading_Status') }}</div>
    <span slot="footer">
      <el-button v-if="hasDraft" type="danger" plain @click="deleteDraft">{{ $t('m.Verif_Delete_Draft') }}</el-button>
      <el-button v-if="status && status.syncStatus === 2" @click="retrySync" :loading="loading">{{ $t('m.Verif_Retry_Sync') }}</el-button>
      <el-button @click="cancel">{{ $t('m.Verif_Back_To_Edit') }}</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="!canSubmit"
        @click="submit"
      >{{ $t('m.Verif_Submit_Program') }}</el-button>
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
      if (!this.status) return this.$t('m.Verif_Status')
      if (this.status.judgeMode === 'spj') return this.$t('m.Verif_Mode_Spj')
      if (this.status.judgeMode === 'interactive') return this.$t('m.Verif_Mode_Interactive')
      return this.$t('m.Verif_Mode_Default')
    },
    syncDescription() {
      if (!this.status) return this.$t('m.Verif_Reading_Data_Status')
      if (this.status.syncStatus === 1) return this.$t('m.Verif_Sync_Success_Desc')
      return this.status.syncMessage || this.$t('m.Verif_Reading_Data_Status')
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
      if (!this.form.code.trim()) return this.$message.warning(this.$t('m.Verif_Program_Required'))
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
      if (!this.form.language) return this.$message.warning(this.$t('m.ProbAI_Select_Language_First'))
      if (!this.status || this.status.syncStatus !== 1) return this.$message.warning(this.$t('m.Verif_Data_Not_Synced'))
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
          await this.$confirm(this.$t('m.Verif_Confirm_Overwrite'), this.$t('m.Verif_Use_Last_Code_Title'), {
            confirmButtonText: this.$t('m.Verif_Confirm_Fill'), cancelButtonText: this.$t('m.Cancel'), type: 'warning'
          })
        } catch (e) { return }
      }
      this.form.language = record.language
      this.form.code = record.code
      this.formChanged()
      this.$message.success(this.$t('m.Verif_Filled_Last_Code'))
    },
    cancel() {
      this.stopDraftTimer()
      if (!this.form.code.trim()) {
        this.explicitExit = true
        this.stopPolling()
        this.$emit('cancel')
        return
      }
      this.$confirm(this.$t('m.Verif_Confirm_Save_Draft'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.Verif_Save_Exit'),
        cancelButtonText: this.$t('m.Verif_No_Save_Exit'),
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
          this.$t('m.Verif_Delete_Draft_Confirm'),
          this.$t('m.Verif_Delete_Draft'),
          { confirmButtonText: this.$t('m.Confirm_Delete'), cancelButtonText: this.$t('m.Cancel'), type: 'error' }
        )
      } catch (e) {
        return
      }
      this.stopDraftTimer()
      await api.admin_deleteProblemVerificationDraft(this.pid)
      this.form.code = ''
      this.hasDraft = false
      this.$message.success(this.$t('m.Verif_Draft_Deleted'))
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
