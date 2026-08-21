<template>
  <el-dialog title="AI 一键验题" :visible.sync="visibleProxy"
             width="min(960px, calc(100vw - 32px))" append-to-body>
    <el-alert
      title="AI 会先独立生成标准程序，再送入正式评测机跑完全部测试点，最后检查题面、样例、数据、stderr 和判题模式。"
      type="info" :closable="false" show-icon
    />
    <div class="validation-kind"><el-tag type="warning">评测类型：AI 验题</el-tag></div>

    <div v-if="running" class="validation-running">
      <div class="running-title"><i class="el-icon-loading" /> AI 正在执行全量验题</div>
      <el-alert v-if="judgeProgress" :title="judgeProgress" :description="judgeMessage"
                type="info" :closable="false" show-icon class="judge-progress" />
      <div class="validation-steps-scroll">
        <el-steps :active="activeStage" finish-status="success" process-status="process" simple>
          <el-step v-for="stage in runningStages" :key="stage" :title="stage" />
        </el-steps>
      </div>
    </div>

    <section v-if="generatedProgram" class="generated-program">
      <div class="program-heading">
        <div>
          <el-tag size="small" type="success">AI 生成标准程序</el-tag>
          <span class="program-language">{{ generatedProgram.language }}</span>
        </div>
        <div class="program-actions">
          <el-button size="mini" plain icon="el-icon-document-copy"
            @click="copyCode(generatedProgram.code, '标准程序')">复制代码</el-button>
          <el-button size="mini" type="primary" plain @click="useGeneratedProgram">填入创建者代码框</el-button>
        </div>
      </div>
      <p v-if="generatedProgram.algorithm" class="program-algorithm">{{ generatedProgram.algorithm }}</p>
      <el-alert v-if="generatedProgram.warnings" :title="generatedProgram.warnings"
        type="warning" :closable="false" show-icon />
      <el-collapse class="program-source">
        <el-collapse-item title="查看 AI 生成的完整源码" name="source">
          <pre>{{ generatedProgram.code }}</pre>
        </el-collapse-item>
      </el-collapse>
    </section>

    <section v-if="generatedProgram && generatedProgram.validatorCode" class="validator-program">
      <div class="program-heading">
        <div>
          <el-tag size="small" type="warning">AI 生成 testlib 输入校验器</el-tag>
          <span class="program-language">{{ generatedProgram.validatorLanguage || 'C++' }}</span>
        </div>
        <el-button size="mini" plain icon="el-icon-document-copy"
          @click="copyCode(generatedProgram.validatorCode, 'testlib 校验器')">复制代码</el-button>
      </div>
      <p v-if="generatedProgram.validatorAlgorithm" class="program-algorithm">{{ generatedProgram.validatorAlgorithm }}</p>
      <el-alert v-if="generatedProgram.validatorWarnings" :title="generatedProgram.validatorWarnings"
        type="warning" :closable="false" show-icon />
      <el-collapse class="program-source">
        <el-collapse-item title="查看 testlib 校验器完整源码" name="validator-source">
          <pre>{{ generatedProgram.validatorCode }}</pre>
        </el-collapse-item>
      </el-collapse>
    </section>

    <template v-if="result">
      <section class="result-summary" :class="resultType">
        <div class="summary-main">
          <i :class="resultType === 'error' ? 'el-icon-error' : resultType === 'warning' ? 'el-icon-warning' : 'el-icon-success'" />
          <div>
            <div class="summary-label">最终结论</div>
            <strong>{{ resultTitle }}</strong>
            <p>{{ result.summary }}</p>
          </div>
        </div>
        <div class="summary-metrics">
          <span>测试点 <b>{{ testPointCount }}</b></span>
          <span>问题 <b>{{ issueCount }}</b></span>
          <span>耗时 <b>{{ resultDurationMs || '--' }}ms</b></span>
        </div>
      </section>

      <el-alert v-if="result.finalRecommendation" class="recommendation"
                title="最终建议" :description="result.finalRecommendation" :type="resultType" :closable="false" />

      <el-collapse v-model="detailPanels" class="report-details">
        <el-collapse-item title="验题过程" name="steps">
          <div v-for="(step, index) in result.steps || []" :key="index" class="result-row">
            <el-tag size="mini" :type="tagType(step.status)">{{ step.status || 'WARN' }}</el-tag>
            <div><strong>{{ step.name }}</strong><p>{{ step.detail }}</p></div>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.sampleResults && result.sampleResults.length" title="样例推演" name="samples">
          <div v-for="sample in result.sampleResults" :key="sample.index" class="result-row">
            <el-tag size="mini" :type="tagType(sample.status)">样例 {{ sample.index }}</el-tag>
            <p>{{ sample.detail }}</p>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.testPointResults && result.testPointResults.length"
                          :title="`全测试点判题与 stderr（${result.testPointResults.length} 个）`" name="testpoints">
          <!-- Keep the report compact: ten rows are visible, the table body
               scrolls independently when a problem has many test points. -->
          <el-table :data="result.testPointResults" border size="mini" max-height="390">
            <el-table-column type="expand">
              <template slot-scope="scope">
                <p><strong>AI 检查过程：</strong>{{ scope.row.detail }}</p>
                <p><strong>是否执行：</strong>{{ scope.row.executed ? '是' : '否' }}</p>
                <p><strong>testlib 校验：</strong>{{ (scope.row.testlibValidator && scope.row.testlibValidator.statusText) || '未执行' }}
                  <span v-if="scope.row.testlibValidator && scope.row.testlibValidator.stderr">；stderr：{{ scope.row.testlibValidator.stderr }}</span>
                </p>
                <strong>stderr：</strong><pre class="stderr-output">{{ scope.row.stderr || '[stderr 为空]' }}</pre>
              </template>
            </el-table-column>
            <el-table-column prop="index" label="测试点" width="74" />
            <el-table-column label="AI 结论" width="92"><template slot-scope="scope"><el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag></template></el-table-column>
            <el-table-column prop="judgeStatus" label="正式判题状态" min-width="150" />
            <el-table-column prop="timeMs" label="时间(ms)" width="90" />
            <el-table-column prop="memoryKb" label="内存(KB)" width="100" />
          </el-table>
        </el-collapse-item>

        <el-collapse-item v-if="result.multiLanguageResults && result.multiLanguageResults.length"
                          title="多语言全测试点判题（C++17 / Java / PyPy3）" name="multilanguage">
          <el-table :data="result.multiLanguageResults" border size="mini">
            <el-table-column prop="language" label="语言" width="150" />
            <el-table-column label="结果" width="100">
              <template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="通过测试点" width="120">
              <template slot-scope="scope">{{ scope.row.passed || 0 }} / {{ scope.row.total || 0 }}</template>
            </el-table-column>
            <el-table-column prop="message" label="说明" />
          </el-table>
          <div v-for="language in result.multiLanguageResults" :key="language.language" class="multi-language-details">
            <strong>{{ language.language }} 测试点明细</strong>
            <el-table :data="language.testPoints || []" border size="mini" max-height="390">
              <el-table-column prop="index" label="测试点" width="80" />
              <el-table-column label="状态" width="100"><template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag>
              </template></el-table-column>
              <el-table-column prop="judgeStatus" label="判题状态" width="150" />
              <el-table-column prop="timeMs" label="耗时(ms)" width="100" />
              <el-table-column prop="stderr" label="stderr" />
            </el-table>
          </div>
        </el-collapse-item>

        <el-collapse-item v-if="result.issues && result.issues.length" title="问题清单" name="issues">
          <div v-for="(issue, index) in result.issues" :key="index" class="issue-row">
            <el-tag size="mini" :type="issueType(issue.severity)">{{ issue.severity }}</el-tag>
            <div><strong>{{ issue.location }}</strong><p>{{ issue.detail }}</p><small>{{ issue.suggestion }}</small></div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </template>

    <el-alert v-if="error" class="result-alert" :title="error" type="error" :closable="false" show-icon />

    <span slot="footer">
      <el-button @click="visibleProxy = false">关闭</el-button>
      <el-button type="primary" :loading="running" :disabled="!canRun" @click="run">
        {{ result ? '重新 AI 验题' : '开始 AI 验题' }}
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
      runningStages: ['AI 生成标准程序', '生成 testlib 校验器', '正式全点判题', '题面与约束', '样例推演', '测试点与 testlib', '最终结论']
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
      if (this.resultType === 'error') return 'AI 验题发现问题'
      return this.resultType === 'warning' ? 'AI 验题通过（有提示）' : 'AI 验题通过'
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
      if (!this.canRun) return this.$message.warning('请先选择标准程序语言')
      this.running = true
      this.error = ''
      this.result = null
      this.generatedProgram = null
      this.activeStage = 0
      this.judgeProgress = '正在生成标准程序'
      this.judgeMessage = 'AI 正在生成可提交的完整标准程序，生成完成后才会进入正式全测试点判题'
      try {
        const generated = (await api.admin_generateProblemAIStandardProgram({
          pid: this.pid, language: this.language
        })).data.data
        this.generatedProgram = Object.assign({}, generated, { aiGenerated: true })
        this.validationRecordId = generated.validationRecordId || null
        this.$emit('program-generated', this.generatedProgram)
        this.activeStage = Math.max(this.activeStage, 1)
        this.judgeProgress = '后台任务已创建'
        this.judgeMessage = '标准程序、正式全测试点判题和 AI 综合分析已交由后台执行，关闭页面不会中断任务'
        const res = this.validationRecordId
          ? await this.waitForValidationRecord(this.validationRecordId)
          : await this.runLegacyValidation(this.generatedProgram)
        this.activeStage = this.runningStages.length - 1
        this.judgeProgress = '验题完成'
        this.judgeMessage = 'AI 验题报告已生成，可查看最终结论和详细过程'
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
      this.judgeProgress = '正在正式判题'
      this.judgeMessage = 'AI 标准程序已生成，正在进行正式全测试点判题'
      const verification = await this.runOfficialJudge(program)
      const officialPassed = verification && verification.verificationStatus === 2
      if (officialPassed) this.$emit('verified')
      this.activeStage = 2
      this.judgeProgress = officialPassed ? '正在分析题面与测试结果' : '正式判题未通过，正在分析失败详情'
      this.judgeMessage = officialPassed
        ? '正在检查题面、约束、样例推演、全部测试点结果及 stderr'
        : '正在保留失败测试点、判题状态和 stderr，并生成完整 AI 验题报告'
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
        this.judgeProgress = '后台任务处理中'
        this.judgeMessage = '后台正在等待正式判题完成并生成 AI 综合验题报告'
        await new Promise(resolve => setTimeout(resolve, 1000))
      }
      throw new Error('后台 AI 验题等待超时，请到历史 AI 验题页面查看任务状态')
    },
    useGeneratedProgram() {
      this.$emit('program-generated', this.generatedProgram)
      this.$message.success('AI 标准程序已填入创建者代码框')
    },
    async copyCode(code, label) {
      const text = String(code || '')
      if (!text) return this.$message.warning('暂无可复制的代码')
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
        this.$message.success(`${label || '代码'}已复制`)
      } catch (e) { this.$message.error('复制失败，请手动选择代码') }
    },
    async waitForJudge(status) {
      for (let count = 0; count < 1200; count += 1) {
        this.judgeProgress = status.judgeStatusText || '正在等待评测机开始'
        this.judgeMessage = status.judgeMessage || '正在等待判题机返回当前测试点状态'
        if (status.verificationStatus !== 1) return status
        await new Promise(resolve => setTimeout(resolve, 1000))
        status = (await api.admin_getProblemVerification(this.pid)).data.data
      }
      throw new Error('正式判题等待超时，请检查判题机状态后重试')
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
      if (status === 524) return 'AI 上游网关生成超时（HTTP 524），请稍后重试或检查管理员配置的 AI 直连地址'
      if (status === 502 || status === 503 || status === 504) {
        return `AI 上游服务暂时不可用（HTTP ${status}），请稍后重试`
      }
      if (status === 408 || /timed?\s*out|timeout|超时/i.test((error && error.message) || '')) {
        return 'AI 服务连接或生成超时，请稍后重试'
      }
      return backendMessage || responseMessage || (error && error.message)
        || 'AI 验题失败，请检查判题机和管理员 AI 配置'
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
