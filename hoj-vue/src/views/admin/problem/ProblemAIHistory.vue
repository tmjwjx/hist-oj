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
          <button v-for="record in pagedRecords" :key="record.id"
            class="record-item" :class="{ active: selectedRecord && selectedRecord.id === record.id }"
            @click="selectRecord(record)">
            <div class="record-main">
              <span>#{{ record.id }}</span>
              <span class="record-tags">
                <el-tag size="mini" type="info">{{ recordStage(record) }}</el-tag>
                <el-tag size="mini" :type="statusType(record.status)">{{ statusText(record.status) }}</el-tag>
              </span>
            </div>
            <div class="record-meta">{{ beijingTime(record.gmtCreate) || '—' }} · {{ record.durationMs || 0 }}ms</div>
          </button>
          <el-pagination v-if="records.length > recordPageSize" class="record-pagination"
            small layout="prev, pager, next" :current-page="recordPage" :page-size="recordPageSize"
            :total="records.length" @current-change="changeRecordPage" />
        </aside>

        <main v-if="selectedRecord" class="report-panel">
          <div class="report-header">
            <div>
              <h2>{{ recordHeading }} #{{ selectedRecord.id }}</h2>
              <span class="report-time">{{ beijingTime(selectedRecord.gmtCreate) || '—' }}</span>
            </div>
            <div class="report-actions">
              <el-button v-if="canRecheck" size="small" type="primary" plain
                icon="el-icon-refresh" @click="openRecheck">复检此报告</el-button>
              <el-tag class="report-result" :type="reportType" effect="light">{{ reportTitle }}</el-tag>
            </div>
          </div>
          <el-alert v-if="selectedRecord.errorMessage" :title="friendlyError(selectedRecord.errorMessage)"
            type="error" :closable="false" show-icon />

          <section v-if="generatedProgram" class="generated-program">
            <div class="generated-heading">
              <strong>{{ programHeading }}</strong>
              <div class="program-actions">
                <el-tag size="mini" type="success">{{ generatedProgram.language || '自动识别语言' }}</el-tag>
                <el-button size="mini" plain icon="el-icon-document-copy"
                  @click="copyCode(generatedProgram.code, '标准程序')">复制代码</el-button>
              </div>
            </div>
            <p v-if="generatedProgram.algorithm">{{ generatedProgram.algorithm }}</p>
            <el-collapse>
              <el-collapse-item title="查看生成源码" name="generated-source">
                <pre>{{ generatedProgram.code }}</pre>
              </el-collapse-item>
            </el-collapse>
          </section>
          <section v-if="generatedProgram && generatedProgram.validatorCode" class="validator-program">
            <div class="generated-heading">
              <strong>已生成 testlib 输入校验器</strong>
              <div class="program-actions">
                <el-tag size="mini" type="warning">{{ generatedProgram.validatorLanguage || 'C++' }}</el-tag>
                <el-button size="mini" plain icon="el-icon-document-copy"
                  @click="copyCode(generatedProgram.validatorCode, 'testlib 校验器')">复制代码</el-button>
              </div>
            </div>
            <p v-if="generatedProgram.validatorAlgorithm">{{ generatedProgram.validatorAlgorithm }}</p>
            <el-collapse>
              <el-collapse-item title="查看 testlib 校验器源码" name="validator-source">
                <pre>{{ generatedProgram.validatorCode }}</pre>
              </el-collapse-item>
            </el-collapse>
          </section>

          <template v-if="report">
            <el-alert v-if="report.recheck" class="recheck-note" type="info" :closable="false" show-icon
              :title="`复检来源：报告 #${report.recheck.sourceRecordId || '—'}`"
              :description="report.recheck.note || '本次为报告复检，未重新执行全量判题。'" />
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
                <!-- Match the per-language detail tables: show ten rows and
                     keep the remaining points in an internal scrollbar. -->
                <el-table :data="report.testPointResults" border stripe size="mini" max-height="390">
                  <el-table-column prop="index" label="测试点" width="80" />
                  <el-table-column prop="status" label="AI 结论" width="100">
                    <template slot-scope="scope"><el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status || 'WARN' }}</el-tag></template>
                  </el-table-column>
                  <el-table-column prop="judgeStatus" label="正式判题状态" width="180" />
                  <el-table-column label="testlib 校验" width="110">
                    <template slot-scope="scope">
                      <el-tag size="mini" :type="tagType(scope.row.testlibValidator && scope.row.testlibValidator.status === 'PASS' ? 'PASS' : 'FAIL')">
                        {{ scope.row.testlibValidator ? scope.row.testlibValidator.status : '未执行' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="timeMs" label="耗时(ms)" width="100" />
                  <el-table-column label="stderr" min-width="260">
                    <template slot-scope="scope"><pre class="stderr-output">{{ scope.row.stderr || '[stderr 为空]' }}</pre></template>
                  </el-table-column>
                </el-table>
              </el-collapse-item>
              <el-collapse-item v-if="report.multiLanguageResults && report.multiLanguageResults.length"
                title="多语言全测试点判题（C++17 / Java / PyPy3）" name="multilanguage">
                <el-table :data="report.multiLanguageResults" border stripe>
                  <el-table-column prop="language" label="语言" width="150" />
                  <el-table-column label="结果" width="100"><template slot-scope="scope">
                    <el-tag size="mini" :type="tagType(scope.row.status)">{{ scope.row.status }}</el-tag>
                  </template></el-table-column>
                  <el-table-column label="通过测试点" width="120"><template slot-scope="scope">
                    {{ scope.row.passed || 0 }} / {{ scope.row.total || 0 }}
                  </template></el-table-column>
                  <el-table-column prop="message" label="说明" />
                </el-table>
                <div v-for="language in report.multiLanguageResults" :key="language.language" class="multi-language-details">
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

    <el-dialog title="复检这份 AI 验题报告" :visible.sync="recheckVisible"
      width="620px" append-to-body :close-on-click-modal="false">
      <el-alert type="info" :closable="false" show-icon
        title="复检会使用当前题目内容和这份历史报告作为上下文"
        description="管理员修改题面、约束、样例、测试数据或标准程序后，可在此补充重点要求。系统会生成一份新的复检报告，不会覆盖原报告。" />
      <div class="recheck-source">
        <span>复检来源：</span>
        <el-tag size="mini" type="info">报告 #{{ selectedRecord && selectedRecord.id }}</el-tag>
        <span v-if="selectedRecord">{{ beijingTime(selectedRecord.gmtCreate) || '—' }}</span>
      </div>
      <el-form label-position="top" class="recheck-form">
        <el-form-item label="复检要求">
          <el-input v-model="recheckRequirements" type="textarea" :rows="7" maxlength="4000"
            show-word-limit placeholder="例如：重点确认我修改后的输入约束、样例 2 和标准程序的边界处理是否仍有问题。" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="recheckVisible = false">取消</el-button>
        <el-button type="primary" :loading="rechecking" @click="submitRecheck">开始复检</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import api from '@/common/api'
import time from '@/common/time'

export default {
  name: 'ProblemAIHistory',
  data() {
    return {
      loading: false,
      records: [],
      selectedRecord: null,
      recordPage: 1,
      recordPageSize: 8,
      recheckVisible: false,
      rechecking: false,
      recheckRequirements: ''
    }
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await api.admin_getProblemAIRecords(this.pid)
        this.records = (res.data.data || []).filter(record => this.isValidationRecord(record)).slice().reverse()
        const current = this.selectedRecord && this.records.find(item => item.id === this.selectedRecord.id)
        this.selectedRecord = current || this.records[0] || null
        this.recordPage = this.selectedRecord ? Math.floor(this.records.indexOf(this.selectedRecord) / this.recordPageSize) + 1 : 1
      } catch (e) { this.$message.error('历史 AI 验题记录加载失败') }
      finally { this.loading = false }
    },
    selectRecord(record) { this.selectedRecord = record },
    changeRecordPage(page) { this.recordPage = page },
    beijingTime(value) { return time.utcToBeijing(value) },
    openRecheck() {
      if (!this.canRecheck) return
      this.recheckRequirements = '请基于这份历史 AI 验题报告，结合当前管理员修改后的题面、输入输出、约束、样例、测试数据和标准程序，逐项复核是否仍存在问题，并明确给出问题依据、风险等级和修改建议。'
      this.recheckVisible = true
    },
    async submitRecheck() {
      if (!this.canRecheck) return
      const requirements = String(this.recheckRequirements || '').trim()
      if (!requirements) return this.$message.warning('请填写复检要求')
      const sourceId = this.selectedRecord.id
      this.rechecking = true
      try {
        const res = await api.admin_recheckProblemAIReport({ recordId: sourceId, requirements })
        const returned = res && res.data && res.data.data
        const newRecordId = returned && (returned.id || returned.recordId)
        this.recheckVisible = false
        this.$message.success(returned && returned.status === 'running' ? '复检任务已创建，报告生成后会显示在历史记录顶部' : '复检完成，已生成新的报告')
        await this.load()
        if (newRecordId) {
          const record = this.records.find(item => String(item.id) === String(newRecordId))
          if (record) {
            this.selectedRecord = record
            this.recordPage = Math.floor(this.records.indexOf(record) / this.recordPageSize) + 1
          }
        }
      } catch (e) {
        this.$message.error(this.friendlyError(
          (e && e.data && (e.data.msg || e.data.message))
          || (e && e.response && e.response.data && (e.response.data.msg || e.response.data.message))
          || (e && e.message)))
      } finally {
        this.rechecking = false
      }
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
    isValidationRecord(record) {
      return record && (record.question === 'AI 一键验题' || record.question === 'AI 生成标准程序' || record.question === 'AI 验题复检')
    },
    isGenerationRecord(record) { return record && record.question === 'AI 生成标准程序' },
    isRecheckRecord(record) { return record && record.question === 'AI 验题复检' },
    recordStage(record) {
      if (this.isGenerationRecord(record)) return '生成标准程序'
      return this.isRecheckRecord(record) ? '复检报告' : '综合验题'
    },
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
  },
  computed: {
    pid() { return this.$route.params.problemId },
    pagedRecords() {
      const start = (this.recordPage - 1) * this.recordPageSize
      return this.records.slice(start, start + this.recordPageSize)
    },
    report() { return this.isGenerationRecord(this.selectedRecord) ? null : this.reportData(this.selectedRecord) },
    generatedProgram() {
      if (this.isGenerationRecord(this.selectedRecord)) return this.reportData(this.selectedRecord)
      return this.report && this.report.standardProgram ? this.report.standardProgram : null
    },
    recordHeading() {
      if (this.isGenerationRecord(this.selectedRecord)) return '标准程序生成记录'
      return this.isRecheckRecord(this.selectedRecord) ? '复检报告' : '验题报告'
    },
    programHeading() {
      if (this.isGenerationRecord(this.selectedRecord)) return '已生成标准程序'
      return this.isRecheckRecord(this.selectedRecord) ? '本次复检使用的当前标准程序' : '综合验题使用的标准程序'
    },
    canRecheck() {
      return Boolean(this.selectedRecord && !this.isGenerationRecord(this.selectedRecord)
        && this.selectedRecord.status !== 'running' && this.selectedRecord.status !== 'failed')
    },
    reportType() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return 'danger'
      if (this.selectedRecord.status === 'running') return 'warning'
      if (this.isGenerationRecord(this.selectedRecord)) return 'success'
      return this.hasBlockingProblems(this.report) ? 'danger' : this.hasWarnings(this.report) ? 'warning' : 'success'
    },
    reportTitle() {
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return '执行失败'
      if (this.selectedRecord.status === 'running') return '处理中'
      if (this.isGenerationRecord(this.selectedRecord)) return '生成完成'
      if (this.hasBlockingProblems(this.report)) return '发现问题'
      return this.hasWarnings(this.report) ? '验题通过（有提示）' : '验题通过'
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
.record-pagination { margin: 12px 0 4px; text-align: center; white-space: nowrap; }
.report-header { margin-bottom: 16px; }
.report-actions { display: inline-flex; align-items: center; gap: 10px; }
.recheck-source { display: flex; align-items: center; gap: 8px; margin: 16px 0 12px; color: #606266; font-size: 13px; }
.recheck-form { margin-top: 14px; }
.report-header h2 { display: inline; margin: 0 12px 0 0; color: #303133; font-size: 20px; }
.report-result { font-size: 14px; }
.summary-card { margin: 16px 0; padding: 16px 18px; border-left: 4px solid; border-radius: 4px; }
.recheck-note { margin: 0 0 16px; }
.summary-card.success { border-color: #67c23a; background: #f0f9eb; color: #529b2e; }
.summary-card.warning { border-color: #e6a23c; background: #fdf6ec; color: #b88230; }
.summary-card.danger { border-color: #f56c6c; background: #fef0f0; color: #c45656; }
.summary-label { margin-bottom: 6px; font-weight: 600; }
.summary-text { white-space: pre-wrap; line-height: 1.6; }
.generated-program { margin: 16px 0; padding: 14px 16px; border: 1px solid #b3e19d; border-radius: 4px; background: #f0f9eb; }
.generated-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.program-actions { display: inline-flex; align-items: center; gap: 8px; }
.generated-program p { color: #606266; white-space: pre-wrap; line-height: 1.5; }
.generated-program pre { max-height: 420px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.validator-program { margin: 16px 0; padding: 14px 16px; border: 1px solid #f5dab1; border-radius: 4px; background: #fdf6ec; }
.validator-program p { color: #606266; white-space: pre-wrap; line-height: 1.5; }
.validator-program pre { max-height: 420px; margin: 0; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.report-row { display: flex; gap: 10px; padding: 10px 0; border-bottom: 1px solid #ebeef5; }
.report-row p { margin: 4px 0; white-space: pre-wrap; line-height: 1.5; }
.report-row small { color: #909399; }
.stderr-output { max-height: 100px; margin: 0; overflow: auto; white-space: pre-wrap; word-break: break-word; }
.multi-language-details { margin-top: 14px; }
.multi-language-details > strong { display: block; margin-bottom: 6px; color: #606266; }
.report-empty { align-self: center; justify-self: center; }
@media (max-width: 900px) { .history-layout { grid-template-columns: 1fr; } .record-panel { padding-right: 0; padding-bottom: 16px; border-right: 0; border-bottom: 1px solid #ebeef5; } }
@media (max-width: 600px) { .ai-history-page { padding: 10px; } .page-header { align-items: flex-start; flex-direction: column; } .page-actions { width: 100%; } .page-actions .el-button { flex: 1; } }
</style>
