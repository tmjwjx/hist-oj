<template>
  <section class="judge-program-review">
    <h3>{{ $t('m.ProbAI_Judge_Program_Review') }}</h3>
    <el-empty v-if="!review" :description="$t('m.ProbAI_Judge_Program_Empty')" />
    <template v-else>
      <div class="review-heading">
        <el-tag :type="statusType(review.status)">{{ statusText(review.status) }}</el-tag>
        <span>{{ review.mode === 'interactive' ? $t('m.ProbAI_Judge_Program_Interactive') : $t('m.ProbAI_Judge_Program_Spj') }}</span>
        <span v-if="review.language">{{ review.language }}</span>
      </div>
      <p v-if="review.summary" class="review-text">{{ review.summary }}</p>
      <el-alert :title="$t('m.ProbAI_Judge_Program_Limit')" type="info" :closable="false" show-icon />
      <p v-if="review.sourceSha256" class="source-hash">SHA-256: {{ review.sourceSha256 }}</p>

      <h4>{{ $t('m.ProbAI_Judge_Program_Checks') }}</h4>
      <el-table v-if="checks.length" :data="checks" border size="mini">
        <el-table-column prop="id" :label="$t('m.ProbAI_Judge_Check_Item')" min-width="130" />
        <el-table-column :label="$t('m.ProbAI_Result')" width="120">
          <template slot-scope="scope"><el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="detail" :label="$t('m.ProbAI_Message')" min-width="240" />
      </el-table>
      <el-empty v-else :description="$t('m.ProbAI_Not_Executed')" />

      <h4>{{ $t('m.ProbAI_Judge_Program_Probes') }}</h4>
      <p class="review-text">{{ $t('m.ProbAI_Judge_Probe_Hint') }}</p>
      <el-table v-if="probes.length" :data="probes" border size="mini">
        <el-table-column type="expand">
          <template slot-scope="scope">
            <div class="probe-detail">
              <p v-if="scope.row.reason" class="review-text">{{ scope.row.reason }}</p>
              <el-alert v-if="scope.row.errorMessage" :title="scope.row.errorMessage" type="warning" :closable="false" show-icon />
              <div class="probe-metrics">
                <span v-if="scope.row.submitId">{{ $t('m.ProbAI_Submission') }} #{{ scope.row.submitId }}</span>
                <span>{{ $t('m.ProbAI_Failed_Cases') }}：{{ count(scope.row.failedCases) }}</span>
                <span>{{ $t('m.ProbAI_Pending_Cases') }}：{{ count(scope.row.pendingCases) }}</span>
              </div>
              <el-collapse v-if="scope.row.code">
                <el-collapse-item :title="$t('m.ProbAI_View_AI_Source')" name="source">
                  <el-button size="mini" plain icon="el-icon-document-copy" @click="$emit('copy', scope.row.code, scope.row.id)">{{ $t('m.ProbAI_Copy_Code') }}</el-button>
                  <pre class="probe-code">{{ scope.row.code }}</pre>
                </el-collapse-item>
              </el-collapse>
              <el-table v-if="scope.row.caseResults && scope.row.caseResults.length" :data="scope.row.caseResults" border size="mini" max-height="390">
                <el-table-column prop="index" :label="$t('m.ProbAI_Test_Point')" width="80" />
                <el-table-column :label="$t('m.ProbAI_Judge_Status')" min-width="150">
                  <template slot-scope="point">{{ caseStatusText(point.row) }}</template>
                </el-table-column>
                <el-table-column prop="timeMs" :label="$t('m.ProbAI_Time_MS')" width="95" />
                <el-table-column prop="memoryKb" :label="$t('m.ProbAI_Memory_KB')" width="105" />
                <el-table-column label="stderr" min-width="220"><template slot-scope="point"><pre class="stderr-output">{{ point.row.stderr || '—' }}</pre></template></el-table-column>
              </el-table>
              <p v-else>{{ $t('m.ProbAI_Not_Executed') }}</p>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" :label="$t('m.ProbAI_Judge_Check_Item')" min-width="130" />
        <el-table-column prop="category" :label="$t('m.ProbAI_Judge_Probe_Category')" min-width="130" />
        <el-table-column :label="$t('m.ProbAI_Judge_Probe_Expectation')" min-width="120">
          <template slot-scope="scope">{{ expectationText(scope.row.expectation) }}</template>
        </el-table-column>
        <el-table-column :label="$t('m.ProbAI_Result')" width="120">
          <template slot-scope="scope"><el-tag size="mini" :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column :label="$t('m.ProbAI_Judge_Probe_Accepted')" min-width="140">
          <template slot-scope="scope">{{ count(scope.row.passedCases) }} / {{ count(scope.row.totalCases) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-else :description="$t('m.ProbAI_Not_Executed')" />
    </template>
  </section>
</template>

<script>
import { JUDGE_STATUS } from '@/common/constants'

export default {
  name: 'ProblemAIJudgeProgramReview',
  props: { review: { type: Object, default: null } },
  computed: {
    checks() { return Array.isArray(this.review && this.review.checks) ? this.review.checks : [] },
    probes() { return Array.isArray(this.review && this.review.probes) ? this.review.probes : [] }
  },
  methods: {
    count(value) { return value === null || value === undefined ? '—' : value },
    statusType(status) { return status === 'PASS' ? 'success' : status === 'FAIL' ? 'danger' : 'warning' },
    statusText(status) {
      const keys = {
        PASS: 'ProbAI_Judge_Check_Passed', FAIL: 'ProbAI_Status_Found_Issues',
        WARN: 'ProbAI_Judge_Check_Warning', INCONCLUSIVE: 'ProbAI_Judge_Check_Inconclusive'
      }
      return this.$t('m.' + (keys[status] || 'ProbAI_Not_Executed'))
    },
    expectationText(expectation) {
      if (expectation === 'ACCEPT') return this.$t('m.ProbAI_Judge_Probe_Accept')
      if (expectation === 'REJECT') return this.$t('m.ProbAI_Judge_Probe_Reject')
      return '—'
    },
    caseStatusText(point) {
      if (point.status === 'NOT_RUN') return this.$t('m.ProbAI_Not_Executed')
      return point.status || (JUDGE_STATUS[point.judgeStatus] && JUDGE_STATUS[point.judgeStatus].name) || this.$t('m.ProbAI_Not_Executed')
    }
  }
}
</script>

<style scoped>
.judge-program-review { margin: 16px 0; min-width: 0; }
.judge-program-review h3, .judge-program-review h4 { margin: 16px 0 10px; }
.review-heading, .probe-metrics { display: flex; align-items: center; flex-wrap: wrap; gap: 12px; margin: 12px 0; }
.review-text { white-space: pre-wrap; line-height: 1.6; }
.source-hash { overflow-wrap: anywhere; color: #606266; font-size: 12px; }
.probe-detail { padding: 0 12px; }
.probe-code { max-height: 420px; overflow: auto; padding: 12px; background: #1f2329; color: #e6edf3; white-space: pre; }
.stderr-output { max-height: 140px; margin: 0; overflow: auto; white-space: pre-wrap; word-break: break-word; }
</style>
