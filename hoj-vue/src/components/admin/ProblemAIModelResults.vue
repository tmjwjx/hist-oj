<template>
  <section v-if="models.length || report.difficulty" class="model-results">
    <div v-if="report.difficulty" class="difficulty-summary">
      <strong>{{ $t('m.ProbAI_CF_Difficulty') }}：{{ rating(report.difficulty) }}</strong>
      <span v-if="report.difficulty.confidence">{{ $t('m.ProbAI_Confidence') }}：{{ report.difficulty.confidence }}</span>
      <p v-if="report.difficulty.reason">{{ report.difficulty.reason }}</p>
      <small>{{ $t('m.ProbAI_Difficulty_Hint') }}</small>
    </div>
    <template v-if="models.length">
      <h4>{{ $t('m.ProbAI_Model_Results') }}</h4>
      <el-table :data="models" border size="mini" row-key="modelId">
        <el-table-column type="expand">
          <template slot-scope="scope">
            <div class="model-detail">
              <p v-if="scope.row.algorithm">{{ scope.row.algorithm }}</p>
              <el-alert v-if="scope.row.errorMessage" :title="scope.row.errorMessage" type="error" :closable="false" show-icon />
              <el-alert v-if="scope.row.warnings" :title="warningText(scope.row.warnings)" type="warning" :closable="false" show-icon />
              <p v-if="scope.row.difficulty">
                <strong>{{ $t('m.ProbAI_CF_Difficulty') }}：{{ rating(scope.row.difficulty) }}</strong>
                <span v-if="scope.row.difficulty.confidence"> · {{ $t('m.ProbAI_Confidence') }}：{{ scope.row.difficulty.confidence }}</span>
              </p>
              <p v-if="scope.row.difficulty && scope.row.difficulty.reason">{{ scope.row.difficulty.reason }}</p>
              <div class="model-metrics">
                <span v-if="scope.row.submitId">{{ $t('m.ProbAI_Submission') }} #{{ scope.row.submitId }}</span>
                <span v-if="scope.row.judgeStatus !== null && scope.row.judgeStatus !== undefined">{{ judgeStatusText(scope.row.judgeStatus) }}</span>
                <span>{{ $t('m.ProbAI_Failed_Cases') }}：{{ scope.row.failedCases || 0 }}</span>
                <span>{{ $t('m.ProbAI_Pending_Cases') }}：{{ scope.row.pendingCases || 0 }}</span>
              </div>
              <el-collapse v-if="scope.row.code">
                <el-collapse-item :title="$t('m.ProbAI_View_AI_Source')" name="source">
                  <div class="code-actions">
                    <el-tag size="mini">{{ scope.row.language || 'C++' }}</el-tag>
                    <el-button size="mini" plain icon="el-icon-document-copy" @click="$emit('copy', scope.row.code, scope.row.name || scope.row.model)">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
                  </div>
                  <pre class="model-code">{{ scope.row.code }}</pre>
                </el-collapse-item>
              </el-collapse>
              <el-table v-if="scope.row.caseResults && scope.row.caseResults.length" :data="scope.row.caseResults" border size="mini" max-height="390">
                <el-table-column prop="index" :label="$t('m.ProbAI_Test_Point')" width="80" />
                <el-table-column :label="$t('m.ProbAI_Judge_Status')" min-width="150">
                  <template slot-scope="point">{{ point.row.status === 'NOT_RUN' ? $t('m.ProbAI_Not_Executed') : point.row.status || judgeStatusText(point.row.judgeStatus) }}</template>
                </el-table-column>
                <el-table-column prop="timeMs" :label="$t('m.ProbAI_Time_MS')" width="95" />
                <el-table-column prop="memoryKb" :label="$t('m.ProbAI_Memory_KB')" width="105" />
                <el-table-column label="stderr" min-width="180"><template slot-scope="point"><pre class="stderr-output">{{ point.row.stderr || '—' }}</pre></template></el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Sys_Model')" min-width="180">
          <template slot-scope="scope">
            <strong>{{ scope.row.name || scope.row.model }}</strong>
            <el-tag v-if="scope.row.primary" size="mini">{{ $t('m.ProbAI_Primary_Model') }}</el-tag>
            <div class="model-name">{{ scope.row.model }}</div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.ProbAI_Result')" width="100"><template slot-scope="scope"><el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag></template></el-table-column>
        <el-table-column :label="$t('m.ProbAI_Passed_Test_Points')" width="120"><template slot-scope="scope">{{ scope.row.passedCases || 0 }} / {{ scope.row.totalCases || 0 }}</template></el-table-column>
        <el-table-column :label="$t('m.ProbAI_CF_Difficulty')" min-width="160"><template slot-scope="scope">{{ rating(scope.row.difficulty) }}</template></el-table-column>
      </el-table>
    </template>
  </section>
</template>

<script>
import { JUDGE_STATUS } from '@/common/constants'

export default {
  name: 'ProblemAIModelResults',
  props: { report: { type: Object, required: true } },
  computed: { models() { return this.report.modelResults || [] } },
  methods: {
    rating(difficulty) {
      if (!difficulty || !difficulty.rating) return this.$t('m.ProbAI_Difficulty_Unavailable')
      const range = difficulty.rangeLow && difficulty.rangeHigh ? ` (${difficulty.rangeLow}–${difficulty.rangeHigh})` : ''
      return `${difficulty.rating}${range}`
    },
    warningText(warnings) { return Array.isArray(warnings) ? warnings.join('\n') : String(warnings) },
    judgeStatusText(status) { return status !== null && status !== undefined && JUDGE_STATUS[status] ? JUDGE_STATUS[status].name : '—' },
    statusType(status) { return status === 'SUCCESS' ? 'success' : ['FAILED', 'ERROR'].includes(status) ? 'danger' : 'warning' },
    statusText(status) {
      if (status === 'SUCCESS') return this.$t('m.Success')
      if (status === 'ERROR') return this.$t('m.ProbAI_Status_Failed')
      if (status === 'FAILED') return this.$t('m.ProbAI_Failed')
      return this.$t('m.ProbAI_Processing')
    }
  }
}
</script>

<style scoped>
.model-results { margin: 16px 0; }
.model-results h4 { margin: 0 0 10px; }
.difficulty-summary { margin-bottom: 16px; padding: 14px; border: 1px solid #d9ecff; background: #ecf5ff; border-radius: 4px; line-height: 1.6; }
.difficulty-summary > span { margin-left: 16px; }
.difficulty-summary p, .model-detail p { white-space: pre-wrap; line-height: 1.6; }
.difficulty-summary small, .model-name { color: #606266; }
.model-detail { padding: 0 12px; }
.model-detail .el-alert { margin: 8px 0; }
.model-metrics, .code-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 16px; margin: 10px 0; }
.model-code { max-height: 420px; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.stderr-output { max-height: 100px; margin: 0; overflow: auto; white-space: pre-wrap; word-break: break-word; }
</style>
