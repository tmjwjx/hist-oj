<template>
  <el-dialog title="AI 一键验题" :visible.sync="visibleProxy" width="760px" append-to-body>
    <el-alert
      title="AI 会先独立生成标准程序，再送入正式评测机跑完全部测试点，最后检查题面、样例、数据、stderr 和判题模式。"
      type="info" :closable="false" show-icon
    />
    <div class="validation-kind"><el-tag type="warning">评测类型：AI 验题</el-tag></div>

    <div v-if="running" class="validation-running">
      <div class="running-title"><i class="el-icon-loading" /> AI 正在执行全量验题</div>
      <el-alert v-if="judgeProgress" :title="judgeProgress" :description="judgeMessage"
                type="info" :closable="false" show-icon class="judge-progress" />
      <el-steps :active="activeStage" finish-status="success" process-status="process" simple>
        <el-step v-for="stage in runningStages" :key="stage" :title="stage" />
      </el-steps>
    </div>

    <section v-if="generatedProgram" class="generated-program">
      <div class="program-heading">
        <div>
          <el-tag size="small" type="success">AI 生成标准程序</el-tag>
          <span class="program-language">{{ generatedProgram.language }}</span>
        </div>
        <el-button size="mini" type="primary" plain @click="useGeneratedProgram">填入创建者代码框</el-button>
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

    <template v-if="result">
      <section class="result-summary" :class="resultType">
        <div class="summary-main">
          <i :class="resultType === 'success' ? 'el-icon-success' : 'el-icon-error'" />
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
          <el-table :data="result.testPointResults" border size="mini" max-height="300">
            <el-table-column type="expand">
              <template slot-scope="scope">
                <p><strong>AI 检查过程：</strong>{{ scope.row.detail }}</p>
                <p><strong>是否执行：</strong>{{ scope.row.executed ? '是' : '否' }}</p>
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
      runningStages: ['AI 生成标准程序', '正式全点判题', '题面与约束', '样例推演', '逐测试点与 stderr', '最终结论']
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
    resultType() { return this.hasProblems(this.result) ? 'error' : 'success' },
    resultTitle() { return this.resultType === 'success' ? 'AI 验题通过' : 'AI 验题发现问题' },
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
        this.$emit('program-generated', this.generatedProgram)
        this.activeStage = Math.max(this.activeStage, 1)
        this.judgeProgress = '正在正式判题'
        this.judgeMessage = 'AI 标准程序已生成，正在进行正式全测试点判题'
        const verification = await this.runOfficialJudge(this.generatedProgram)
        const officialPassed = verification && verification.verificationStatus === 2
        if (officialPassed) this.$emit('verified')
        this.activeStage = 2
        this.judgeProgress = officialPassed ? '正在分析题面与测试结果' : '正式判题未通过，正在分析失败详情'
        this.judgeMessage = officialPassed
          ? '正在检查题面、约束、样例推演、全部测试点结果及 stderr'
          : '正在保留失败测试点、判题状态和 stderr，并生成完整 AI 验题报告'
        const res = await api.admin_problemAIValidate({
          pid: this.pid, language: this.generatedProgram.language,
          standardProgram: this.generatedProgram.code, aiGenerated: true,
          algorithmSummary: this.generatedProgram.algorithm
        })
        this.activeStage = this.runningStages.length - 1
        this.judgeProgress = '验题完成'
        this.judgeMessage = 'AI 验题报告已生成，可查看最终结论和详细过程'
        this.showRecord(res.data.data)
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
      this.detailPanels = this.hasProblems(this.result) ? ['issues'] : []
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
    useGeneratedProgram() {
      this.$emit('program-generated', this.generatedProgram)
      this.$message.success('AI 标准程序已填入创建者代码框')
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
      return (error && error.data && (error.data.msg || error.data.message))
        || (error && error.message)
        || 'AI 验题失败，请检查判题机和管理员 AI 配置'
    },
    hasProblems(result) {
      if (!result || result.overall !== 'PASS') return true
      const issueProblem = (result.issues || []).some(item => item.severity !== 'INFO')
      const stepProblem = (result.steps || []).some(item => item.status !== 'PASS')
      const sampleProblem = (result.sampleResults || []).some(item => item.status !== 'PASS')
      const testPointProblem = (result.testPointResults || []).some(item => item.status !== 'PASS')
      return issueProblem || stepProblem || sampleProblem || testPointProblem
    },
    tagType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    issueType(level) { return level === 'ERROR' ? 'danger' : level === 'INFO' ? 'info' : 'warning' }
  }
}
</script>

<style scoped>
.validation-running, .result-alert { margin-top: 16px; }
.validation-kind { margin-top: 12px; }
.generated-program {
  margin-top: 14px;
  padding: 14px 16px 8px;
  border: 1px solid #b3e19d;
  border-radius: 6px;
  background: #f0f9eb;
}
.program-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.program-language { margin-left: 10px; color: #606266; font-weight: 600; }
.program-algorithm { margin: 10px 0; color: #606266; line-height: 1.55; white-space: pre-wrap; }
.program-source { margin-top: 8px; }
.program-source pre { max-height: 320px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.judge-progress { margin-bottom: 12px; }
.running-title { margin-bottom: 12px; color: #409eff; }
.result-summary {
  margin-top: 16px;
  padding: 16px 18px;
  border: 1px solid;
  border-radius: 6px;
}
.result-summary.success { color: #2f8a25; background: #f0f9eb; border-color: #b3e19d; }
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
@media (max-width: 640px) {
  .summary-metrics { margin-left: 0; gap: 10px; }
}
</style>
