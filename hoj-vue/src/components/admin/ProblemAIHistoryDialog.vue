<template>
  <el-dialog :title="$t('m.ProbAI_History_Dialog_Title')" :visible.sync="visibleProxy" width="900px" append-to-body>
    <el-table v-loading="loading" :data="records" size="mini" highlight-current-row
      @current-change="selectRecord" :empty-text="$t('m.ProbAI_Empty_Records')">
      <el-table-column :label="$t('m.ProbAI_Record')" width="90"><template slot-scope="scope">#{{ scope.row.id }}</template></el-table-column>
      <el-table-column :label="$t('m.ProbAI_Stage')" width="125"><template slot-scope="scope">
        <el-tag size="mini" type="info">{{ recordStage(scope.row) }}</el-tag>
      </template></el-table-column>
      <el-table-column :label="$t('m.Status')" width="100"><template slot-scope="scope">
        <el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
      </template></el-table-column>
      <el-table-column prop="durationMs" :label="$t('m.ProbAI_Time_Cost')" width="100" />
      <el-table-column :label="$t('m.ProbAI_Record_Time')" min-width="170"><template slot-scope="scope">
        {{ beijingTime(scope.row.gmtCreate) }}
      </template></el-table-column>
      <el-table-column :label="$t('m.Operation')" width="100"><template slot-scope="scope">
        <el-button type="text" @click="selectRecord(scope.row)">{{ $t('m.ProbAI_View_Result') }}</el-button>
      </template></el-table-column>
    </el-table>

    <section v-if="selectedRecord" class="history-report">
      <div class="report-heading">
        <span>{{ isGenerationRecord(selectedRecord) ? $t('m.ProbAI_Generate_Record_Heading') : $t('m.ProbAI_Validation_Report') }} #{{ selectedRecord.id }}
          <small class="report-time">{{ beijingTime(selectedRecord.gmtCreate) }}</small>
        </span>
        <el-tag :type="reportType">{{ reportTitle }}</el-tag>
      </div>
      <el-alert v-if="selectedRecord.errorMessage" :title="friendlyError(selectedRecord.errorMessage)"
        type="error" :closable="false" show-icon />
      <section v-if="generatedProgram" class="generated-program">
        <div class="generated-heading">
          <strong>{{ $t('m.ProbAI_Program_Generated') }}</strong>
          <div class="program-actions">
            <el-tag size="mini" type="success">{{ generatedProgram.language || $t('m.ProbAI_Auto_Language') }}</el-tag>
            <el-button size="mini" plain icon="el-icon-document-copy"
              @click="copyCode(generatedProgram.code, $t('m.ProbAI_Standard_Program'))">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
          </div>
        </div>
        <p v-if="generatedProgram.algorithm">{{ generatedProgram.algorithm }}</p>
        <pre>{{ generatedProgram.code }}</pre>
      </section>
      <section v-if="generatedProgram && generatedProgram.validatorCode" class="validator-program">
        <div class="generated-heading">
          <strong>{{ $t('m.ProbAI_Testlib_Validator_Generated') }}</strong>
          <div class="program-actions">
            <el-tag size="mini" type="warning">{{ generatedProgram.validatorLanguage || 'C++' }}</el-tag>
            <el-button size="mini" plain icon="el-icon-document-copy"
              @click="copyCode(generatedProgram.validatorCode, $t('m.ProbAI_Testlib_Validator'))">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
          </div>
        </div>
        <p v-if="generatedProgram.validatorAlgorithm">{{ generatedProgram.validatorAlgorithm }}</p>
        <pre>{{ generatedProgram.validatorCode }}</pre>
      </section>
      <template v-if="report">
        <p class="report-summary">{{ report.summary || $t('m.ProbAI_No_Summary') }}</p>
        <el-collapse>
          <el-collapse-item :title="$t('m.ProbAI_Validation_Process')" name="steps">
            <div v-for="(step, index) in report.steps || []" :key="index" class="report-row">
              <el-tag size="mini" :type="tagType(step.status)">{{ step.status || 'WARN' }}</el-tag>
              <div><strong>{{ step.name }}</strong><p>{{ step.detail }}</p></div>
            </div>
          </el-collapse-item>
          <el-collapse-item v-if="report.testPointResults && report.testPointResults.length"
            :title="$t('m.ProbAI_TestPoints_And_Stderr', { count: report.testPointResults.length })" name="tests">
            <el-table :data="report.testPointResults" border size="mini" max-height="260">
              <el-table-column prop="index" :label="$t('m.ProbAI_Test_Point')" width="70" />
              <el-table-column prop="status" :label="$t('m.ProbAI_AI_Conclusion')" width="90" />
              <el-table-column prop="judgeStatus" :label="$t('m.ProbAI_Formal_Judge_Status')" min-width="140" />
              <el-table-column :label="$t('m.ProbAI_Testlib_Check')" width="110"><template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.testlibValidator && scope.row.testlibValidator.status === 'PASS' ? 'PASS' : 'FAIL')">
                  {{ scope.row.testlibValidator ? scope.row.testlibValidator.status : $t('m.ProbAI_Not_Executed') }}
                </el-tag>
              </template></el-table-column>
              <el-table-column label="stderr" min-width="180"><template slot-scope="scope">
                <pre class="stderr-output">{{ scope.row.stderr || $t('m.ProbAI_Stderr_Empty') }}</pre>
              </template></el-table-column>
            </el-table>
          </el-collapse-item>
          <el-collapse-item v-if="report.issues && report.issues.length" :title="$t('m.ProbAI_Issue_List')" name="issues">
            <div v-for="(issue, index) in report.issues" :key="index" class="report-row">
              <el-tag size="mini" :type="issueType(issue.severity)">{{ issue.severity }}</el-tag>
              <div><strong>{{ issue.location }}</strong><p>{{ issue.detail }}</p><small>{{ issue.suggestion }}</small></div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </template>
    </section>
  </el-dialog>
</template>

<script>
import api from '@/common/api'
import time from '@/common/time'

export default {
  name: 'ProblemAIHistoryDialog',
  props: { visible: Boolean, pid: { type: [Number, String], required: true } },
  data() { return { loading: false, records: [], selectedRecord: null } },
  computed: {
    visibleProxy: {
      get() { return this.visible },
      set(value) { if (!value) this.$emit('close') }
    },
    report() {
      if (this.isGenerationRecord(this.selectedRecord)) return null
      if (!this.selectedRecord || !this.selectedRecord.response) return null
      try { return typeof this.selectedRecord.response === 'string' ? JSON.parse(this.selectedRecord.response) : this.selectedRecord.response }
      catch (e) { return { overall: 'WARN', summary: this.selectedRecord.response, steps: [], issues: [] } }
    },
    generatedProgram() {
      if (!this.isGenerationRecord(this.selectedRecord) || !this.selectedRecord.response) return null
      try { return typeof this.selectedRecord.response === 'string' ? JSON.parse(this.selectedRecord.response) : this.selectedRecord.response }
      catch (e) { return null }
    },
    reportType() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return 'danger'
      if (this.selectedRecord.status === 'running') return 'warning'
      if (this.isGenerationRecord(this.selectedRecord)) return 'success'
      return this.hasBlockingProblems(this.report) ? 'danger' : this.hasWarnings(this.report) ? 'warning' : 'success'
    },
    reportTitle() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return this.$t('m.ProbAI_Status_Failed')
      if (this.selectedRecord.status === 'running') return this.$t('m.ProbAI_Processing')
      if (this.isGenerationRecord(this.selectedRecord)) return this.$t('m.ProbAI_Status_Generated')
      if (this.hasBlockingProblems(this.report)) return this.$t('m.ProbAI_Status_Found_Issues')
      return this.hasWarnings(this.report) ? this.$t('m.ProbAI_Status_Pass_Warnings') : this.$t('m.ProbAI_Status_Passed')
    }
  },
  watch: { visible(value) { if (value) this.load() } },
  mounted() { if (this.visible) this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await api.admin_getProblemAIRecords(this.pid)
        this.records = (res.data.data || []).filter(this.isValidationRecord).slice().reverse()
        this.selectedRecord = this.records[0] || null
      } catch (e) { this.$message.error(this.$t('m.ProbAI_Load_Failed')) }
      finally { this.loading = false }
    },
    selectRecord(record) { if (record) this.selectedRecord = record },
    beijingTime(value) { return time.utcToBeijing(value) },
    async copyCode(code, label) {
      const text = String(code || '')
      if (!text) return this.$message.warning(this.$t('m.ProbAI_No_Code_To_Copy'))
      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(text)
        } else {
          const textarea = document.createElement('textarea')
          textarea.value = text
          textarea.setAttribute('readonly', '')
          textarea.style.position = 'fixed'
          textarea.style.opacity = '0'
          document.body.appendChild(textarea)
          textarea.select()
          const copied = document.execCommand('copy')
          document.body.removeChild(textarea)
          if (!copied) throw new Error('copy failed')
        }
        this.$message.success((label || this.$t('m.ProbAI_Code')) + this.$t('m.ProbAI_Copied'))
      } catch (e) { this.$message.error(this.$t('m.ProbAI_Copy_Failed')) }
    },
    isValidationRecord(record) {
      return record && (record.question === 'AI 一键验题' || record.question === 'AI 生成标准程序')
    },
    isGenerationRecord(record) { return record && record.question === 'AI 生成标准程序' },
    recordStage(record) { return this.isGenerationRecord(record) ? this.$t('m.ProbAI_Stage_Generate') : this.$t('m.ProbAI_Stage_Comprehensive') },
    friendlyError(message) {
      const text = String(message || '')
      if (/\b524\b/.test(text)) return this.$t('m.ProbAI_Err_Gateway_Timeout')
      if (/\b50[234]\b/.test(text)) return this.$t('m.ProbAI_Err_Upstream_Unavailable')
      if (/timed?\s*out|timeout|超时/i.test(text)) return this.$t('m.ProbAI_Err_Connection_Timeout')
      return text || this.$t('m.ProbAI_Err_Task_Failed')
    },
    statusType(status) { return status === 'success' ? 'success' : status === 'failed' ? 'danger' : 'warning' },
    statusText(status) { return status === 'success' ? this.$t('m.Success') : status === 'failed' ? this.$t('m.ProbAI_Failed') : this.$t('m.ProbAI_Processing') },
    tagType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    issueType(level) { return level === 'ERROR' ? 'danger' : level === 'INFO' ? 'info' : 'warning' },
    hasBlockingProblems(result) {
      if (!result) return true
      if (result.overall === 'FAIL') return true
      if ((result.issues || []).some(item => item.severity === 'ERROR')) return true
      return [...(result.steps || []), ...(result.sampleResults || []), ...(result.testPointResults || [])]
        .some(item => item.status === 'FAIL')
    },
    hasWarnings(result) {
      if (!result) return false
      if (result.overall === 'WARN') return true
      if ((result.issues || []).some(item => item.severity === 'WARNING')) return true
      return [...(result.steps || []), ...(result.sampleResults || []), ...(result.testPointResults || [])]
        .some(item => item.status === 'WARN')
    }
  }
}
</script>

<style scoped>
.history-report { margin-top: 16px; padding-top: 14px; border-top: 1px solid #ebeef5; }
.report-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; font-weight: 600; }
.program-actions { display: inline-flex; align-items: center; gap: 8px; }
.report-summary { margin: 10px 0; color: #606266; white-space: pre-wrap; line-height: 1.5; }
.report-row { display: flex; gap: 10px; padding: 9px 0; border-bottom: 1px solid #ebeef5; }
.report-row p { margin: 4px 0; white-space: pre-wrap; }
.report-row small { color: #909399; }
.stderr-output { max-height: 80px; margin: 0; overflow: auto; white-space: pre-wrap; }
.validator-program { margin-top: 12px; padding: 14px 16px; border: 1px solid #f5dab1; border-radius: 4px; background: #fdf6ec; }
.validator-program p { color: #606266; white-space: pre-wrap; line-height: 1.5; }
.validator-program pre { max-height: 320px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
</style>
