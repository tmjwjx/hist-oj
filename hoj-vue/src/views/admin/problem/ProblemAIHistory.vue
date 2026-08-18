<template>
  <div class="ai-history-page">
    <el-card>
      <div slot="header" class="page-header">
        <div>
          <div class="panel-title home-title">历史 AI 验题</div>
          <div class="page-subtitle">题目 #{{ pid }} 的全部 AI 验题结果</div>
        </div>
        <div class="page-actions">
          <el-button icon="el-icon-refresh" :loading="loading" @click="load">刷新记录</el-button>
          <el-button type="primary" plain @click="backToValidation">返回题目验题</el-button>
        </div>
      </div>

      <div v-loading="loading" class="history-layout">
        <aside class="record-panel">
          <div class="section-heading">验题记录 <span>({{ records.length }})</span></div>
          <el-empty v-if="!loading && !records.length" description="暂无历史 AI 验题记录" />
          <button v-for="record in records" :key="record.id"
            class="record-item" :class="{ active: selectedRecord && selectedRecord.id === record.id }"
            @click="selectRecord(record)">
            <div class="record-main">
              <span>#{{ record.id }}</span>
              <span class="record-tags">
                <el-tag size="mini" type="info">{{ recordStage(record) }}</el-tag>
                <el-tag size="mini" :type="statusType(record.status)">{{ statusText(record.status) }}</el-tag>
              </span>
            </div>
            <div class="record-meta">{{ record.gmtCreate || '—' }} · {{ record.durationMs || 0 }}ms</div>
          </button>
        </aside>

        <main v-if="selectedRecord" class="report-panel">
          <div class="report-header">
            <div>
              <h2>{{ recordHeading }} #{{ selectedRecord.id }}</h2>
              <span class="report-time">{{ selectedRecord.gmtCreate || '—' }}</span>
            </div>
            <el-tag class="report-result" :type="reportType" effect="light">{{ reportTitle }}</el-tag>
          </div>
          <el-alert v-if="selectedRecord.errorMessage" :title="friendlyError(selectedRecord.errorMessage)"
            type="error" :closable="false" show-icon />

          <section v-if="generatedProgram" class="generated-program">
            <div class="generated-heading">
              <strong>已生成标准程序</strong>
              <el-tag size="mini" type="success">{{ generatedProgram.language || '自动识别语言' }}</el-tag>
            </div>
            <p v-if="generatedProgram.algorithm">{{ generatedProgram.algorithm }}</p>
            <el-collapse>
              <el-collapse-item title="查看生成源码" name="generated-source">
                <pre>{{ generatedProgram.code }}</pre>
              </el-collapse-item>
            </el-collapse>
          </section>

          <template v-if="report">
            <div class="summary-card" :class="reportType">
              <div class="summary-label">最终结论</div>
              <div class="summary-text">{{ report.summary || '未返回摘要' }}</div>
            </div>

            <el-collapse class="report-sections" :value="['steps', 'tests', 'issues']">
              <el-collapse-item title="验题过程" name="steps">
                <div v-for="(step, index) in report.steps || []" :key="'step' + index" class="report-row">
                  <el-tag size="mini" :type="tagType(step.status)">{{ step.status || 'WARN' }}</el-tag>
                  <div><strong>{{ step.name }}</strong><p>{{ step.detail }}</p></div>
                </div>
                <el-empty v-if="!report.steps || !report.steps.length" description="暂无过程记录" />
              </el-collapse-item>
              <el-collapse-item v-if="report.testPointResults && report.testPointResults.length"
                :title="`测试点与 stderr（${report.testPointResults.length} 个）`" name="tests">
                <el-table :data="report.testPointResults" border stripe>
                  <el-table-column prop="index" label="测试点" width="80" />
                  <el-table-column prop="status" label="AI 结论" width="100">
                    <template slot-scope="scope"><el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status || 'WARN' }}</el-tag></template>
                  </el-table-column>
                  <el-table-column prop="judgeStatus" label="正式判题状态" width="180" />
                  <el-table-column prop="timeMs" label="耗时(ms)" width="100" />
                  <el-table-column label="stderr" min-width="260">
                    <template slot-scope="scope"><pre class="stderr-output">{{ scope.row.stderr || '[stderr 为空]' }}</pre></template>
                  </el-table-column>
                </el-table>
              </el-collapse-item>
              <el-collapse-item v-if="report.issues && report.issues.length" title="问题清单" name="issues">
                <div v-for="(issue, index) in report.issues" :key="'issue' + index" class="report-row">
                  <el-tag size="mini" :type="issueType(issue.severity)">{{ issue.severity }}</el-tag>
                  <div><strong>{{ issue.location }}</strong><p>{{ issue.detail }}</p><small>{{ issue.suggestion }}</small></div>
                </div>
              </el-collapse-item>
            </el-collapse>
          </template>
        </main>
        <el-empty v-else-if="!loading" class="report-empty" description="请选择一条验题记录" />
      </div>
    </el-card>
  </div>
</template>

<script>
import api from '@/common/api'

export default {
  name: 'ProblemAIHistory',
  data() { return { loading: false, records: [], selectedRecord: null } },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await api.admin_getProblemAIRecords(this.pid)
        this.records = (res.data.data || []).filter(record => this.isValidationRecord(record)).slice().reverse()
        const current = this.selectedRecord && this.records.find(item => item.id === this.selectedRecord.id)
        this.selectedRecord = current || this.records[0] || null
      } catch (e) { this.$message.error('历史 AI 验题记录加载失败') }
      finally { this.loading = false }
    },
    selectRecord(record) { this.selectedRecord = record },
    isValidationRecord(record) {
      return record && (record.question === 'AI 一键验题' || record.question === 'AI 生成标准程序')
    },
    isGenerationRecord(record) { return record && record.question === 'AI 生成标准程序' },
    recordStage(record) { return this.isGenerationRecord(record) ? '生成标准程序' : '综合验题' },
    friendlyError(message) {
      const text = String(message || '')
      if (/\b524\b/.test(text)) return 'AI 上游网关生成超时，系统已自动重试；请稍后再试或联系管理员检查 AI 直连地址'
      if (/\b50[234]\b/.test(text)) return 'AI 上游服务暂时不可用，系统已自动重试，请稍后再试'
      if (/timed?\s*out|timeout|超时/i.test(text)) return 'AI 服务连接或生成超时，请稍后重试'
      return text || 'AI 任务执行失败，请稍后重试'
    },
    backToValidation() {
      this.$router.push({
        name: 'admin-edit-problem',
        params: { problemId: this.pid },
        query: { step: '2' }
      })
    },
    reportData(record) {
      if (!record || !record.response) return null
      try { return typeof record.response === 'string' ? JSON.parse(record.response) : record.response }
      catch (e) { return { overall: 'WARN', summary: record.response, steps: [], issues: [] } }
    },
    statusType(status) { return status === 'success' ? 'success' : status === 'failed' ? 'danger' : 'warning' },
    statusText(status) { return status === 'success' ? '成功' : status === 'failed' ? '失败' : '处理中' },
    tagType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    issueType(level) { return level === 'ERROR' ? 'danger' : level === 'INFO' ? 'info' : 'warning' },
    hasProblems(result) {
      if (!result || result.overall !== 'PASS') return true
      return [...(result.issues || []), ...(result.steps || []), ...(result.sampleResults || []), ...(result.testPointResults || [])]
        .some(item => item.severity !== 'INFO' && item.status !== 'PASS')
    }
  },
  computed: {
    pid() { return this.$route.params.problemId },
    report() { return this.isGenerationRecord(this.selectedRecord) ? null : this.reportData(this.selectedRecord) },
    generatedProgram() { return this.isGenerationRecord(this.selectedRecord) ? this.reportData(this.selectedRecord) : null },
    recordHeading() { return this.isGenerationRecord(this.selectedRecord) ? '标准程序生成记录' : '验题报告' },
    reportType() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return 'danger'
      if (this.selectedRecord.status === 'running') return 'warning'
      return this.isGenerationRecord(this.selectedRecord) || !this.hasProblems(this.report) ? 'success' : 'danger'
    },
    reportTitle() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return '执行失败'
      if (this.selectedRecord.status === 'running') return '处理中'
      if (this.isGenerationRecord(this.selectedRecord)) return '生成完成'
      return this.hasProblems(this.report) ? '发现问题' : '验题通过'
    }
  }
}
</script>

<style scoped>
.ai-history-page { padding: 20px; background: #f5f7fa; min-height: calc(100vh - 60px); }
.page-header, .page-actions, .record-main, .report-header { display: flex; align-items: center; }
.page-header, .report-header { justify-content: space-between; gap: 20px; }
.page-subtitle, .record-meta, .report-time { color: #909399; font-size: 13px; }
.page-subtitle { margin-top: 6px; }
.history-layout { display: grid; grid-template-columns: 280px minmax(0, 1fr); gap: 20px; min-height: 620px; }
.record-panel { padding-right: 16px; border-right: 1px solid #ebeef5; }
.section-heading { margin: 8px 0 12px; color: #303133; font-weight: 600; }
.section-heading span { color: #909399; font-weight: 400; }
.record-item { display: block; width: 100%; margin-bottom: 8px; padding: 12px; border: 1px solid #ebeef5; border-radius: 4px; background: #fff; text-align: left; cursor: pointer; }
.record-item:hover, .record-item.active { border-color: #409eff; background: #ecf5ff; }
.record-main { justify-content: space-between; color: #303133; font-weight: 600; }
.record-tags { display: inline-flex; gap: 5px; }
.record-meta { margin-top: 8px; }
.report-header { margin-bottom: 16px; }
.report-header h2 { display: inline; margin: 0 12px 0 0; color: #303133; font-size: 20px; }
.report-result { font-size: 14px; }
.summary-card { margin: 16px 0; padding: 16px 18px; border-left: 4px solid; border-radius: 4px; }
.summary-card.success { border-color: #67c23a; background: #f0f9eb; color: #529b2e; }
.summary-card.danger { border-color: #f56c6c; background: #fef0f0; color: #c45656; }
.summary-label { margin-bottom: 6px; font-weight: 600; }
.summary-text { white-space: pre-wrap; line-height: 1.6; }
.generated-program { margin: 16px 0; padding: 14px 16px; border: 1px solid #b3e19d; border-radius: 4px; background: #f0f9eb; }
.generated-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.generated-program p { color: #606266; white-space: pre-wrap; line-height: 1.5; }
.generated-program pre { max-height: 420px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.report-row { display: flex; gap: 10px; padding: 10px 0; border-bottom: 1px solid #ebeef5; }
.report-row p { margin: 4px 0; white-space: pre-wrap; line-height: 1.5; }
.report-row small { color: #909399; }
.stderr-output { max-height: 100px; margin: 0; overflow: auto; white-space: pre-wrap; word-break: break-word; }
.report-empty { align-self: center; justify-self: center; }
@media (max-width: 900px) { .history-layout { grid-template-columns: 1fr; } .record-panel { padding-right: 0; padding-bottom: 16px; border-right: 0; border-bottom: 1px solid #ebeef5; } }
@media (max-width: 600px) { .ai-history-page { padding: 10px; } .page-header { align-items: flex-start; flex-direction: column; } .page-actions { width: 100%; } .page-actions .el-button { flex: 1; } }
</style>
