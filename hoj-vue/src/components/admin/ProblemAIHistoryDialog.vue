<template>
  <el-dialog title="历史 AI 验题记录" :visible.sync="visibleProxy" width="900px" append-to-body>
    <el-table v-loading="loading" :data="records" size="mini" highlight-current-row
      @current-change="selectRecord" empty-text="暂无历史 AI 验题记录">
      <el-table-column label="记录" width="90"><template slot-scope="scope">#{{ scope.row.id }}</template></el-table-column>
      <el-table-column label="阶段" width="125"><template slot-scope="scope">
        <el-tag size="mini" type="info">{{ recordStage(scope.row) }}</el-tag>
      </template></el-table-column>
      <el-table-column label="状态" width="100"><template slot-scope="scope">
        <el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
      </template></el-table-column>
      <el-table-column prop="durationMs" label="耗时(ms)" width="100" />
      <el-table-column label="验题时间" min-width="170"><template slot-scope="scope">
        {{ beijingTime(scope.row.gmtCreate) }}
      </template></el-table-column>
      <el-table-column label="操作" width="100"><template slot-scope="scope">
        <el-button type="text" @click="selectRecord(scope.row)">查看结果</el-button>
      </template></el-table-column>
    </el-table>

    <section v-if="selectedRecord" class="history-report">
      <div class="report-heading">
        <span>{{ isGenerationRecord(selectedRecord) ? '标准程序生成记录' : '验题报告' }} #{{ selectedRecord.id }}
          <small class="report-time">{{ beijingTime(selectedRecord.gmtCreate) }}</small>
        </span>
        <el-tag :type="reportType">{{ reportTitle }}</el-tag>
      </div>
      <el-alert v-if="selectedRecord.errorMessage" :title="friendlyError(selectedRecord.errorMessage)"
        type="error" :closable="false" show-icon />
      <section v-if="generatedProgram" class="generated-program">
        <div class="generated-heading">
          <strong>已生成标准程序</strong>
          <div class="program-actions">
            <el-tag size="mini" type="success">{{ generatedProgram.language || '自动识别语言' }}</el-tag>
            <el-button size="mini" plain icon="el-icon-document-copy"
              @click="copyCode(generatedProgram.code, '标准程序')">复制代码</el-button>
          </div>
        </div>
        <p v-if="generatedProgram.algorithm">{{ generatedProgram.algorithm }}</p>
        <pre>{{ generatedProgram.code }}</pre>
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
        <pre>{{ generatedProgram.validatorCode }}</pre>
      </section>
      <template v-if="report">
        <p class="report-summary">{{ report.summary || '未返回摘要' }}</p>
        <el-collapse>
          <el-collapse-item title="验题过程" name="steps">
            <div v-for="(step, index) in report.steps || []" :key="index" class="report-row">
              <el-tag size="mini" :type="tagType(step.status)">{{ step.status || 'WARN' }}</el-tag>
              <div><strong>{{ step.name }}</strong><p>{{ step.detail }}</p></div>
            </div>
          </el-collapse-item>
          <el-collapse-item v-if="report.testPointResults && report.testPointResults.length"
            :title="`测试点与 stderr（${report.testPointResults.length} 个）`" name="tests">
            <el-table :data="report.testPointResults" border size="mini" max-height="260">
              <el-table-column prop="index" label="测试点" width="70" />
              <el-table-column prop="status" label="AI 结论" width="90" />
              <el-table-column prop="judgeStatus" label="正式判题状态" min-width="140" />
              <el-table-column label="testlib 校验" width="110"><template slot-scope="scope">
                <el-tag size="mini" :type="tagType(scope.row.testlibValidator && scope.row.testlibValidator.status === 'PASS' ? 'PASS' : 'FAIL')">
                  {{ scope.row.testlibValidator ? scope.row.testlibValidator.status : '未执行' }}
                </el-tag>
              </template></el-table-column>
              <el-table-column label="stderr" min-width="180"><template slot-scope="scope">
                <pre class="stderr-output">{{ scope.row.stderr || '[stderr 为空]' }}</pre>
              </template></el-table-column>
            </el-table>
          </el-collapse-item>
          <el-collapse-item v-if="report.issues && report.issues.length" title="问题清单" name="issues">
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
      if (!this.selectedRecord || this.selectedRecord.status === 'failed') return '执行失败'
      if (this.selectedRecord.status === 'running') return '处理中'
      if (this.isGenerationRecord(this.selectedRecord)) return '生成完成'
      if (this.hasBlockingProblems(this.report)) return '发现问题'
      return this.hasWarnings(this.report) ? '验题通过（有提示）' : '验题通过'
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
      } catch (e) { this.$message.error('历史 AI 验题记录加载失败') }
      finally { this.loading = false }
    },
    selectRecord(record) { if (record) this.selectedRecord = record },
    beijingTime(value) { return time.utcToBeijing(value) },
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
