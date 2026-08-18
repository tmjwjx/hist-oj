<template>
  <el-dialog title="历史 AI 验题记录" :visible.sync="visibleProxy" width="900px" append-to-body>
    <el-table v-loading="loading" :data="records" size="mini" highlight-current-row
      @current-change="selectRecord" empty-text="暂无历史 AI 验题记录">
      <el-table-column label="记录" width="90"><template slot-scope="scope">#{{ scope.row.id }}</template></el-table-column>
      <el-table-column label="状态" width="100"><template slot-scope="scope">
        <el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
      </template></el-table-column>
      <el-table-column prop="durationMs" label="耗时(ms)" width="100" />
      <el-table-column prop="gmtCreate" label="验题时间" min-width="170" />
      <el-table-column label="操作" width="100"><template slot-scope="scope">
        <el-button type="text" @click="selectRecord(scope.row)">查看结果</el-button>
      </template></el-table-column>
    </el-table>

    <section v-if="selectedRecord" class="history-report">
      <div class="report-heading">
        <span>记录 #{{ selectedRecord.id }} 的验题报告</span>
        <el-tag :type="reportType">{{ reportTitle }}</el-tag>
      </div>
      <el-alert v-if="selectedRecord.errorMessage" :title="selectedRecord.errorMessage"
        type="error" :closable="false" show-icon />
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
      if (!this.selectedRecord || !this.selectedRecord.response) return null
      try { return typeof this.selectedRecord.response === 'string' ? JSON.parse(this.selectedRecord.response) : this.selectedRecord.response }
      catch (e) { return { overall: 'WARN', summary: this.selectedRecord.response, steps: [], issues: [] } }
    },
    reportType() { return this.hasProblems(this.report) ? 'danger' : 'success' },
    reportTitle() { return this.hasProblems(this.report) ? '发现问题' : '验题通过' }
  },
  watch: { visible(value) { if (value) this.load() } },
  mounted() { if (this.visible) this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await api.admin_getProblemAIRecords(this.pid)
        this.records = (res.data.data || []).filter(item => item.question === 'AI 一键验题').slice().reverse()
        this.selectedRecord = this.records[0] || null
      } catch (e) { this.$message.error('历史 AI 验题记录加载失败') }
      finally { this.loading = false }
    },
    selectRecord(record) { if (record) this.selectedRecord = record },
    statusType(status) { return status === 'success' ? 'success' : status === 'failed' ? 'danger' : 'warning' },
    statusText(status) { return status === 'success' ? '成功' : status === 'failed' ? '失败' : '处理中' },
    tagType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    issueType(level) { return level === 'ERROR' ? 'danger' : level === 'INFO' ? 'info' : 'warning' },
    hasProblems(result) {
      if (!result || result.overall !== 'PASS') return true
      return [...(result.issues || []), ...(result.steps || []), ...(result.sampleResults || []), ...(result.testPointResults || [])]
        .some(item => item.severity !== 'INFO' && item.status !== 'PASS')
    }
  }
}
</script>

<style scoped>
.history-report { margin-top: 16px; padding-top: 14px; border-top: 1px solid #ebeef5; }
.report-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; font-weight: 600; }
.report-summary { margin: 10px 0; color: #606266; white-space: pre-wrap; line-height: 1.5; }
.report-row { display: flex; gap: 10px; padding: 9px 0; border-bottom: 1px solid #ebeef5; }
.report-row p { margin: 4px 0; white-space: pre-wrap; }
.report-row small { color: #909399; }
.stderr-output { max-height: 80px; margin: 0; overflow: auto; white-space: pre-wrap; }
</style>
