<template>
  <el-dialog :title="$t('m.ProbAI_Validate_Dialog_Title')" :visible.sync="visibleProxy"
             width="min(960px, calc(100vw - 32px))" append-to-body>
    <el-alert
      :title="$t('m.ProbAI_Validate_Alert_Desc')"
      type="info" :closable="false" show-icon
    />
    <div class="validation-kind"><el-tag type="warning">{{ $t('m.ProbAI_Kind_Tag') }}</el-tag></div>

    <div v-if="running" class="validation-running">
      <div class="running-title"><i class="el-icon-loading" /> {{ $t('m.ProbAI_Running_Title') }}</div>
      <el-alert v-if="judgeProgress" :title="judgeProgress" :description="judgeMessage"
                type="info" :closable="false" show-icon class="judge-progress" />
      <div class="validation-steps-scroll">
        <el-steps :active="activeStage" finish-status="success" process-status="process" simple>
          <el-step v-for="stage in runningStages" :key="stage" :title="$t('m.' + stage)" />
        </el-steps>
      </div>
    </div>

    <section v-if="generatedProgram" class="generated-program">
      <div class="program-heading">
        <div>
          <el-tag size="small" type="success">{{ $t('m.ProbAI_Generated_Program_Tag') }}</el-tag>
          <span class="program-language">{{ generatedProgram.language }}</span>
        </div>
        <div class="program-actions">
          <el-button size="mini" plain icon="el-icon-document-copy"
            @click="copyCode(generatedProgram.code, $t('m.ProbAI_Standard_Program'))">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
          <el-button size="mini" type="primary" plain @click="useGeneratedProgram">{{ $t('m.ProbAI_Use_Generated_Program') }}</el-button>
        </div>
      </div>
      <p v-if="generatedProgram.algorithm" class="program-algorithm">{{ generatedProgram.algorithm }}</p>
      <el-alert v-if="generatedProgram.warnings" :title="generatedProgram.warnings"
        type="warning" :closable="false" show-icon />
      <el-collapse class="program-source">
        <el-collapse-item :title="$t('m.ProbAI_View_AI_Source')" name="source">
          <pre>{{ generatedProgram.code }}</pre>
        </el-collapse-item>
      </el-collapse>
    </section>

    <section v-if="generatedProgram && generatedProgram.validatorCode" class="validator-program">
      <div class="program-heading">
        <div>
          <el-tag size="small" type="warning">{{ $t('m.ProbAI_Testlib_Tag') }}</el-tag>
          <span class="program-language">{{ generatedProgram.validatorLanguage || 'C++' }}</span>
        </div>
        <el-button size="mini" plain icon="el-icon-document-copy"
          @click="copyCode(generatedProgram.validatorCode, $t('m.ProbAI_Testlib_Validator'))">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
      </div>
      <p v-if="generatedProgram.validatorAlgorithm" class="program-algorithm">{{ generatedProgram.validatorAlgorithm }}</p>
      <el-alert v-if="generatedProgram.validatorWarnings" :title="generatedProgram.validatorWarnings"
        type="warning" :closable="false" show-icon />
      <el-collapse class="program-source">
        <el-collapse-item :title="$t('m.ProbAI_View_Testlib_Full_Source')" name="validator-source">
          <pre>{{ generatedProgram.validatorCode }}</pre>
        </el-collapse-item>
      </el-collapse>
    </section>

    <template v-if="result">
      <section class="result-summary" :class="resultType">
        <div class="summary-main">
          <i :class="resultType === 'error' ? 'el-icon-error' : resultType === 'warning' ? 'el-icon-warning' : 'el-icon-success'" />
          <div>
            <div class="summary-label">{{ $t('m.ProbAI_Final_Conclusion') }}</div>
            <strong>{{ resultTitle }}</strong>
            <p>{{ result.summary }}</p>
          </div>
        </div>
        <div class="summary-metrics">
          <span>{{ $t('m.ProbAI_Test_Point') }} <b>{{ testPointCount }}</b></span>
          <span>{{ $t('m.ProbAI_Issues') }} <b>{{ issueCount }}</b></span>
          <span>{{ $t('m.ProbAI_Duration') }} <b>{{ resultDurationMs || '--' }}ms</b></span>
        </div>
      </section>

      <el-alert v-if="result.finalRecommendation" class="recommendation"
                :title="$t('m.ProbAI_Final_Recommendation')" :description="result.finalRecommendation" :type="resultType" :closable="false" />

      <el-collapse v-model="detailPanels" class="report-details">
        <el-collapse-item :title="$t('m.ProbAI_Validation_Process')" name="steps">
          <div v-for="(step, index) in result.steps || []" :key="index" class="result-row">
            <el-tag size="mini" :type="tagType(step.status)">{{ step.status || 'WARN' }}</el-tag>
            <div><strong>{{ step.name }}</strong><p>{{ step.detail }}</p></div>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.sampleResults && result.sampleResults.length" :title="$t('m.ProbAI_Sample_Simulation')" name="samples">
          <div v-for="sample in result.sampleResults" :key="sample.index" class="result-row">
            <el-tag size="mini" :type="tagType(sample.status)">{{ $t('m.ProbAI_Sample') }} {{ sample.index }}</el-tag>
            <p>{{ sample.detail }}</p>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.testPointResults && result.testPointResults.length"
                          :title="$t('m.ProbAI_All_TestPoints_And_Stderr', { count: result.testPointResults.length })" name="testpoints">
          <!-- Keep the report compact: ten rows are visible, the table body
               scrolls independently when a problem has many test points. -->
          <el-table :data="result.testPointResults" border size="mini" max-height="390">
            <el-table-column type="expand">
              <template slot-scope="scope">
                <p><strong>{{ $t('m.ProbAI_AI_Check_Process') }}</strong>{{ scope.row.detail }}</p>
                <p><strong>{{ $t('m.ProbAI_Executed_Label') }}</strong>{{ scope.row.executed ? $t('m.Yes') : $t('m.No') }}</p>
                <p><strong>{{ $t('m.ProbAI_Testlib_Check_Label') }}</strong>{{ (scope.row.testlibValidator && scope.row.testlibValidator.statusText) || $t('m.ProbAI_Not_Executed') }}
                  <span v-if="scope.row.testlibValidator && scope.row.testlibValidator.stderr">{{ $t('m.ProbAI_Stderr_Label') }}{{ scope.row.testlibValidator.stderr }}</span>
                </p>
                <strong>{{ $t('m.ProbAI_Stderr_Colon') }}</strong><pre class="stderr-output">{{ scope.row.stderr || $t('m.ProbAI_Stderr_Empty') }}</pre>
              </template>
            </el-table-column>
            <el-table-column prop="index" :label="$t('m.ProbAI_Test_Point')" width="74" />
            <el-table-column :label="$t('m.ProbAI_AI_Conclusion')" width="92"><template slot-scope="scope"><el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag></template></el-table-column>
            <el-table-column prop="judgeStatus" :label="$t('m.ProbAI_Formal_Judge_Status')" min-width="150" />
            <el-table-column prop="timeMs" :label="$t('m.ProbAI_Time_MS')" width="90" />
            <el-table-column prop="memoryKb" :label="$t('m.ProbAI_Memory_KB')" width="100" />
          </el-table>
        </el-collapse-item>

        <el-collapse-item v-if="result.multiLanguageResults && result.multiLanguageResults.length"
                          :title="$t('m.ProbAI_Multi_Language_Judging')" name="multilanguage">
          <el-table :data="result.multiLanguageResults" border size="mini">
            <el-table-column prop="language" :label="$t('m.Language')" width="150" />
            <el-table-column :label="$t('m.ProbAI_Result')" width="100">
              <template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.ProbAI_Passed_Test_Points')" width="120">
              <template slot-scope="scope">{{ scope.row.passed || 0 }} / {{ scope.row.total || 0 }}</template>
            </el-table-column>
            <el-table-column prop="message" :label="$t('m.ProbAI_Message')" />
          </el-table>
          <div v-for="language in result.multiLanguageResults" :key="language.language" class="multi-language-details">
            <strong>{{ $t('m.ProbAI_TestPoint_Details', { lang: language.language }) }}</strong>
            <el-table :data="language.testPoints || []" border size="mini" max-height="390">
              <el-table-column prop="index" :label="$t('m.ProbAI_Test_Point')" width="80" />
              <el-table-column :label="$t('m.Status')" width="100"><template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag>
              </template></el-table-column>
              <el-table-column prop="judgeStatus" :label="$t('m.ProbAI_Judge_Status')" width="150" />
              <el-table-column prop="timeMs" :label="$t('m.ProbAI_Time_Cost')" width="100" />
              <el-table-column prop="stderr" label="stderr" />
            </el-table>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.issues && result.issues.length" :title="$t('m.ProbAI_Issue_List')" name="issues">
          <div v-for="(issue, index) in result.issues" :key="index" class="issue-row">
            <el-tag size="mini" :type="issueType(issue.severity)">{{ issue.severity }}</el-tag>
            <div><strong>{{ issue.location }}</strong><p>{{ issue.detail }}</p><small>{{ issue.suggestion }}</small></div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </template>

    <el-alert v-if="error" class="result-alert" :title="error" type="error" :closable="false" show-icon />

    <span slot="footer">
      <el-button @click="visibleProxy = false">{{ $t('m.Close') }}</el-button>
      <el-button type="primary" :loading="running" :disabled="!canRun" @click="run">
        {{ result ? $t('m.ProbAI_Revalidate') : $t('m.ProbAI_Start_Validation') }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import api from '@/common/api'

export default {
  name: 'ProblemAIValidationDialog',
  props: {
    visible: Boolean,
    pid: { type: [Number, String], required: true },
    language: { type: String, default: '' },
    standardProgram: { type: String, default: '' },
    autoRun: { type: Boolean, default: false }
  },
  data() {
    return {
      result: null, running: false, error: '', autoStarted: false,
      detailPanels: [], resultDurationMs: 0,
      judgeProgress: '', judgeMessage: '',
      activeStage: 0,
      generatedProgram: null,
      validationRecordId: null,
      // 模块级数据存 i18n 键名字符串，模板中通过 $t('m.' + stage) 翻译
      runningStages: ['ProbAI_Generated_Program_Tag', 'ProbAI_Stage_Testlib', 'ProbAI_Stage_Official_Judge', 'ProbAI_Stage_Statement', 'ProbAI_Sample_Simulation', 'ProbAI_Stage_TestPoints_Testlib', 'ProbAI_Final_Conclusion']
    }
  },
  computed: {
    visibleProxy: {
      get() { return this.visible },
      set(value) {
        if (!value) {
          this.$emit('close')
        }
      }
    },
    canRun() { return Boolean(this.language) && !this.running },
    resultType() {
      if (this.hasBlockingProblems(this.result)) return 'error'
      return this.hasWarnings(this.result) ? 'warning' : 'success'
    },
    resultTitle() {
      if (this.resultType === 'error') return this.$t('m.ProbAI_Result_Found_Issues')
      return this.resultType === 'warning' ? this.$t('m.ProbAI_Result_Pass_Warnings') : this.$t('m.ProbAI_Result_Passed')
    },
    testPointCount() { return (this.result && this.result.testPointResults && this.result.testPointResults.length) || 0 },
    issueCount() { return (this.result && this.result.issues && this.result.issues.length) || 0 }
  },
  watch: { visible(value) { if (value) this.initialize() } },
  mounted() { if (this.visible) this.initialize() },
  methods: {
    async initialize() {
      if (this.autoRun && !this.autoStarted) {
        this.autoStarted = true
        await this.run()
      }
    },
    async run() {
      if (!this.canRun) return this.$message.warning(this.$t('m.ProbAI_Select_Language_First'))
      this.running = true
      this.error = ''
      this.result = null
      this.generatedProgram = null
      this.activeStage = 0
      this.judgeProgress = this.$t('m.ProbAI_Progress_Generating')
      this.judgeMessage = this.$t('m.ProbAI_Progress_Generating_Desc')
      try {
        const generated = (await api.admin_generateProblemAIStandardProgram({
          pid: this.pid, language: this.language
        })).data.data
        this.generatedProgram = Object.assign({}, generated, { aiGenerated: true })
        this.validationRecordId = generated.validationRecordId || null
        this.$emit('program-generated', this.generatedProgram)
        this.activeStage = Math.max(this.activeStage, 1)
        this.judgeProgress = this.$t('m.ProbAI_Progress_Backend_Created')
        this.judgeMessage = this.$t('m.ProbAI_Progress_Backend_Desc')
        const res = this.validationRecordId
          ? await this.waitForValidationRecord(this.validationRecordId)
          : await this.runLegacyValidation(this.generatedProgram)
        this.activeStage = this.runningStages.length - 1
        this.judgeProgress = this.$t('m.ProbAI_Progress_Done')
        this.judgeMessage = this.$t('m.ProbAI_Progress_Done_Desc')
        const record = res && res.data ? res.data.data : res
        this.showRecord(record)
        const parsed = record && this.parseResult(record.response)
        if (record && record.status === 'success' && parsed && !this.hasBlockingProblems(parsed)) this.$emit('verified')
      } catch (e) {
        this.error = this.errorMessage(e)
      } finally {
        this.running = false
      }
    },
    showRecord(record) {
      this.error = record && record.errorMessage ? record.errorMessage : ''
      this.result = this.parseResult(record && record.response)
      if (this.result && this.result.standardProgram) this.generatedProgram = this.result.standardProgram
      this.resultDurationMs = (record && record.durationMs) || 0
      this.detailPanels = this.hasBlockingProblems(this.result) ? ['issues'] : []
    },
    async runOfficialJudge(program) {
      let status = (await api.admin_getProblemVerification(this.pid)).data.data
      if (status.verificationStatus === 1) await this.waitForJudge(status)
      status = (await api.admin_submitProblemVerification({
        pid: this.pid, language: program.language, code: program.code,
        verificationType: 'ai_validation'
      })).data.data
      return this.waitForJudge(status)
    },
    async runLegacyValidation(program) {
      this.judgeProgress = this.$t('m.ProbAI_Progress_Official_Judging')
      this.judgeMessage = this.$t('m.ProbAI_Progress_Official_Desc')
      const verification = await this.runOfficialJudge(program)
      const officialPassed = verification && verification.verificationStatus === 2
      if (officialPassed) this.$emit('verified')
      this.activeStage = 2
      this.judgeProgress = officialPassed ? this.$t('m.ProbAI_Progress_Analyzing') : this.$t('m.ProbAI_Progress_Analyzing_Failed')
      this.judgeMessage = officialPassed
        ? this.$t('m.ProbAI_Progress_Checking_Desc')
        : this.$t('m.ProbAI_Progress_Keeping_Desc')
      return api.admin_problemAIValidate({
        pid: this.pid, language: program.language,
        standardProgram: program.code, aiGenerated: true,
        algorithmSummary: program.algorithm,
        validatorLanguage: program.validatorLanguage,
        validatorCode: program.validatorCode
      })
    },
    async waitForValidationRecord(recordId) {
      for (let count = 0; count < 1200; count += 1) {
        const record = (await api.admin_getProblemAIRecord(recordId)).data.data
        if (record.status !== 'running') return record
        this.judgeProgress = this.$t('m.ProbAI_Progress_Backend_Running')
        this.judgeMessage = this.$t('m.ProbAI_Progress_Backend_Waiting')
        await new Promise(resolve => setTimeout(resolve, 1000))
      }
      throw new Error(this.$t('m.ProbAI_Err_Backend_Timeout'))
    },
    useGeneratedProgram() {
      this.$emit('program-generated', this.generatedProgram)
      this.$message.success(this.$t('m.ProbAI_Program_Filled'))
    },
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
    async waitForJudge(status) {
      for (let count = 0; count < 1200; count += 1) {
        this.judgeProgress = status.judgeStatusText || this.$t('m.ProbAI_Progress_Waiting_Judge')
        this.judgeMessage = status.judgeMessage || this.$t('m.ProbAI_Progress_Waiting_Judge_Desc')
        if (status.verificationStatus !== 1) return status
        await new Promise(resolve => setTimeout(resolve, 1000))
        status = (await api.admin_getProblemVerification(this.pid)).data.data
      }
      throw new Error(this.$t('m.ProbAI_Err_Judge_Timeout'))
    },
    parseResult(response) {
      if (!response) return null
      try { return typeof response === 'string' ? JSON.parse(response) : response }
      catch (e) { return { overall: 'WARN', summary: response, steps: [], issues: [], sampleResults: [] } }
    },
    errorMessage(error) {
      const response = error && error.response
      const status = Number((response && response.status) || (error && error.status) || 0)
      const responseData = response && response.data
      const backendMessage = error && error.data && (error.data.msg || error.data.message)
      const responseMessage = responseData && (responseData.msg || responseData.message)
      if (status === 524) return this.$t('m.ProbAI_Err_524')
      if (status === 502 || status === 503 || status === 504) {
        return this.$t('m.ProbAI_Err_Upstream', { status: status })
      }
      if (status === 408 || /timed?\s*out|timeout|超时/i.test((error && error.message) || '')) {
        return this.$t('m.ProbAI_Err_Connection_Timeout')
      }
      return backendMessage || responseMessage || (error && error.message)
        || this.$t('m.ProbAI_Err_Validation_Failed')
    },
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
    },
    tagType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    issueType(level) { return level === 'ERROR' ? 'danger' : level === 'INFO' ? 'info' : 'warning' }
  }
}
</script>

<style scoped>
.validation-running, .result-alert { margin-top: 16px; }
.validation-running {
  padding: 18px 20px 14px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #f5f7fa;
}
.validation-kind { margin-top: 12px; }
.generated-program {
  margin-top: 14px;
  padding: 14px 16px 8px;
  border: 1px solid #b3e19d;
  border-radius: 6px;
  background: #f0f9eb;
}
.validator-program {
  margin-top: 14px;
  padding: 14px 16px 8px;
  border: 1px solid #f5dab1;
  border-radius: 6px;
  background: #fdf6ec;
}
.program-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.program-actions { display: inline-flex; align-items: center; gap: 8px; }
.program-language { margin-left: 10px; color: #606266; font-weight: 600; }
.program-algorithm { margin: 10px 0; color: #606266; line-height: 1.55; white-space: pre-wrap; }
.program-source { margin-top: 8px; }
.program-source pre { max-height: 320px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.judge-progress { margin-bottom: 12px; }
.running-title { margin-bottom: 12px; color: #409eff; }
.validation-steps-scroll { width: 100%; overflow-x: auto; overflow-y: hidden; padding-bottom: 4px; }
.validation-steps-scroll /deep/ .el-steps--simple { min-width: 860px; padding: 16px 12px; }
.validation-steps-scroll /deep/ .el-step.is-simple .el-step__title {
  min-width: 0;
  max-width: 122px;
  padding: 0 4px;
  font-size: 14px;
  overflow-wrap: normal;
  word-break: normal;
  white-space: normal;
  text-align: center;
  line-height: 1.35;
}
.validation-steps-scroll /deep/ .el-step.is-simple .el-step__arrow { margin: 0 4px; }
.result-summary {
  margin-top: 16px;
  padding: 16px 18px;
  border: 1px solid;
  border-radius: 6px;
}
.result-summary.success { color: #2f8a25; background: #f0f9eb; border-color: #b3e19d; }
.result-summary.warning { color: #9a6700; background: #fdf6ec; border-color: #f5dab1; }
.result-summary.error { color: #c45656; background: #fef0f0; border-color: #fbc4c4; }
.summary-main { display: flex; align-items: flex-start; gap: 12px; }
.summary-main > i { font-size: 28px; margin-top: 2px; }
.summary-label { font-size: 12px; opacity: .78; margin-bottom: 2px; }
.summary-main strong { font-size: 18px; }
.summary-main p { margin: 6px 0 0; line-height: 1.5; white-space: pre-wrap; }
.summary-metrics { display: flex; gap: 20px; margin: 14px 0 0 40px; font-size: 12px; opacity: .9; }
.summary-metrics b { font-size: 14px; margin-left: 3px; }
.report-details { margin-top: 14px; }
.report-details /deep/ .el-collapse-item__header { height: 40px; line-height: 40px; font-weight: 600; }
.report-details /deep/ .el-collapse-item__content { padding: 0 8px 8px; }
.result-row, .issue-row { display: flex; gap: 10px; padding: 10px 0; border-bottom: 1px solid #ebeef5; }
.result-row p, .issue-row p { margin: 4px 0; white-space: pre-wrap; }
.issue-row small { color: #909399; }
.recommendation { margin-top: 16px; }
.stderr-output { max-height: 180px; overflow: auto; padding: 10px; background: #f5f7fa; white-space: pre-wrap; }
.multi-language-details { margin-top: 14px; }
.multi-language-details > strong { display: block; margin-bottom: 6px; color: #606266; }
@media (max-width: 640px) {
  .validation-running { padding: 14px 10px 10px; }
  .validation-steps-scroll /deep/ .el-steps--simple { min-width: 820px; }
  .summary-metrics { margin-left: 0; gap: 10px; }
}
</style>
