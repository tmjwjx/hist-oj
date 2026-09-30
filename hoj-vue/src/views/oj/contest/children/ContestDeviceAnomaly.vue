<template>
  <div class="device-anomaly-container">
    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      closable
      class="page-error"
      @close="errorMessage = ''"
    />

    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <span>{{ $t('m.Plag_Loading') }}</span>
    </div>

    <template v-else>
      <el-card class="summary-card" shadow="never">
        <div slot="header" class="card-header">
          <span>{{ $t('m.DevAnom_Title') }}</span>
          <div class="header-actions">
            <el-button size="small" icon="el-icon-refresh" :loading="loading" @click="loadData">{{ $t('m.Refresh') }}</el-button>
            <el-button size="small" icon="el-icon-download" :loading="exporting" @click="exportCsv"
              :disabled="!report || report.suspiciousAccountCount === 0">{{ $t('m.DevAnom_Export_CSV') }}</el-button>
          </div>
        </div>
        <div class="summary-row">
          <div class="stat-item">
            <div class="stat-value">{{ report ? report.totalSubmissions : 0 }}</div>
            <div class="stat-label">{{ $t('m.DevAnom_Total_Submissions') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ report ? report.unknownDeviceSubmissions : 0 }}</div>
            <div class="stat-label">{{ $t('m.DevAnom_Unknown_Device_Submissions') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value suspicious">{{ report ? report.suspiciousAccountCount : 0 }}</div>
            <div class="stat-label">{{ $t('m.DevAnom_Suspicious_Accounts') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value suspicious">{{ report && report.sharedDevices ? report.sharedDevices.length : 0 }}</div>
            <div class="stat-label">{{ $t('m.DevAnom_Shared_Devices') }}</div>
          </div>
        </div>
        <el-alert
          :title="$t('m.DevAnom_Rule_Tip')"
          type="info"
          :closable="false"
          class="rule-tip"
        />
      </el-card>

      <!-- 作弊团伙（传递连通分组） -->
      <el-card v-if="clusters.length > 0" class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>{{ $t('m.DevAnom_Clusters_Header', { count: clusters.length }) }}</span>
        </div>
        <div v-for="cluster in clusters" :key="cluster.clusterId" class="cluster-block">
          <div class="cluster-title">
            <el-tag size="small" type="danger">{{ $t('m.DevAnom_Cluster_Tag', { id: cluster.clusterId }) }}</el-tag>
            <span class="cluster-size">{{ $t('m.DevAnom_Account_Count', { count: cluster.usernames.length }) }}</span>
          </div>
          <div class="cluster-members">
            <el-tag v-for="name in cluster.usernames" :key="name" size="medium" effect="dark"
              type="danger" class="member-tag">{{ name }}</el-tag>
          </div>
          <div v-for="link in cluster.links" :key="link.deviceId" class="cluster-link">
            <i class="el-icon-connection"></i>
            {{ $t('m.DevAnom_Shared_Device') }} <code>{{ shortId(link.deviceId) }}</code>：
            <span class="link-users">{{ link.usernames.join(' ⇄ ') }}</span>
          </div>
          <div v-if="cluster.links.length === 0" class="cluster-link muted">
            {{ $t('m.DevAnom_Single_Rule_Hit') }}
          </div>
        </div>
        <div class="cluster-tip">{{ $t('m.DevAnom_Cluster_Tip') }}</div>
      </el-card>

      <!-- 嫌疑账号清单 -->
      <el-card class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>{{ $t('m.DevAnom_Suspicious_Header', { count: suspiciousAccounts.length }) }}</span>
        </div>
        <el-table :data="suspiciousAccounts" border stripe row-key="uid">
          <el-table-column type="expand">
            <template slot-scope="scope">
              <div class="expand-body">
                <div class="expand-title">{{ $t('m.DevAnom_Related_Accounts') }}</div>
                <div v-if="scope.row.relatedUsernames && scope.row.relatedUsernames.length > 0"
                  class="related-users">
                  <el-tag v-for="name in scope.row.relatedUsernames" :key="name" size="small"
                    type="danger" class="member-tag">{{ name }}</el-tag>
                </div>
                <div v-else class="muted" style="margin-bottom:10px;">{{ $t('m.DevAnom_No_Related') }}</div>
                <div class="expand-title">{{ $t('m.DevAnom_Device_Detail_Title') }}</div>
                <el-table :data="scope.row.devices" size="small" border>
                  <el-table-column :label="$t('m.DevAnom_Device_ID')" width="140">
                    <template slot-scope="d">{{ shortId(d.row.deviceId) }}</template>
                  </el-table-column>
                  <el-table-column prop="userAgent" :label="$t('m.DevAnom_User_Agent')" min-width="220" show-overflow-tooltip>
                    <template slot-scope="d">{{ d.row.userAgent || $t('m.Unknown') }}</template>
                  </el-table-column>
                  <el-table-column prop="submissionCount" :label="$t('m.DevAnom_Submission_Count')" width="80" align="center" />
                  <el-table-column :label="$t('m.DevAnom_First_Submission')" width="160">
                    <template slot-scope="d">{{ fmtTime(d.row.firstSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('m.DevAnom_Last_Submission')" width="160">
                    <template slot-scope="d">{{ fmtTime(d.row.lastSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('m.DevAnom_Other_Accounts')" min-width="160">
                    <template slot-scope="d">
                      <el-tag v-for="name in d.row.otherUsernames" :key="name" size="mini" type="danger"
                        class="other-user-tag">{{ name }}</el-tag>
                      <span v-if="!d.row.otherUsernames || d.row.otherUsernames.length === 0" class="muted">{{ $t('m.DevAnom_None') }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="username" :label="$t('m.DevAnom_Account')" min-width="110" />
          <el-table-column :label="$t('m.DevAnom_Cluster')" width="80" align="center">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.clusterId && scope.row.relatedUsernames
                && scope.row.relatedUsernames.length > 0" size="mini" type="danger">
                #{{ scope.row.clusterId }}
              </el-tag>
              <span v-else class="muted">-</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.DevAnom_Reason')" width="190">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.multiDevice" size="small" type="warning">{{ $t('m.DevAnom_Reason_Multi_Device') }}</el-tag>
              <el-tag v-if="scope.row.deviceSharing" size="small" type="danger">{{ $t('m.DevAnom_Reason_Shared_Device') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="deviceCount" :label="$t('m.DevAnom_Device_Count')" width="80" align="center" />
          <el-table-column prop="submissionCount" :label="$t('m.DevAnom_Submission_Count')" width="80" align="center" />
          <el-table-column :label="$t('m.DevAnom_First_Submission')" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.firstSubmitTime) }}</template>
          </el-table-column>
          <el-table-column :label="$t('m.DevAnom_Last_Submission')" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.lastSubmitTime) }}</template>
          </el-table-column>
        </el-table>
        <div v-if="suspiciousAccounts.length === 0" class="empty-tip">{{ $t('m.DevAnom_No_Suspicious') }}</div>
      </el-card>

      <!-- 共用设备清单 -->
      <el-card class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>{{ $t('m.DevAnom_Shared_Header', { count: sharedDevices.length }) }}</span>
        </div>
        <el-table :data="sharedDevices" border stripe row-key="deviceId">
          <el-table-column type="expand">
            <template slot-scope="scope">
              <div class="expand-body">
                <div class="expand-title">{{ $t('m.DevAnom_Device_Users_Title') }}</div>
                <el-table :data="scope.row.users" size="small" border>
                  <el-table-column prop="username" :label="$t('m.DevAnom_Account')" min-width="120" />
                  <el-table-column prop="submissionCount" :label="$t('m.DevAnom_Submissions_On_Device')" width="160" align="center" />
                  <el-table-column :label="$t('m.DevAnom_First_Submission')" width="160">
                    <template slot-scope="u">{{ fmtTime(u.row.firstSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('m.DevAnom_Last_Submission')" width="160">
                    <template slot-scope="u">{{ fmtTime(u.row.lastSubmitTime) }}</template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.DevAnom_Device_ID')" width="140">
            <template slot-scope="scope">{{ shortId(scope.row.deviceId) }}</template>
          </el-table-column>
          <el-table-column prop="userAgent" :label="$t('m.DevAnom_User_Agent')" min-width="220" show-overflow-tooltip>
            <template slot-scope="scope">{{ scope.row.userAgent || $t('m.Unknown') }}</template>
          </el-table-column>
          <el-table-column prop="userCount" :label="$t('m.DevAnom_Account_Count_Col')" width="80" align="center" />
          <el-table-column prop="submissionCount" :label="$t('m.DevAnom_Submission_Count')" width="80" align="center" />
          <el-table-column :label="$t('m.DevAnom_First_Submission')" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.firstSubmitTime) }}</template>
          </el-table-column>
          <el-table-column :label="$t('m.DevAnom_Last_Submission')" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.lastSubmitTime) }}</template>
          </el-table-column>
        </el-table>
        <div v-if="sharedDevices.length === 0" class="empty-tip">{{ $t('m.DevAnom_No_Shared_Devices') }}</div>
      </el-card>
    </template>
  </div>
</template>

<script>
import api from '@/api/plagiarism'
import { mapState } from 'vuex'

export default {
  name: 'ContestDeviceAnomaly',
  data() {
    return {
      loading: false,
      exporting: false,
      errorMessage: '',
      report: null
    }
  },
  computed: {
    ...mapState({
      contest: (state) => state.contest.contest,
    }),
    contestId() {
      return this.contest && this.contest.id
    },
    suspiciousAccounts() {
      return (this.report && this.report.suspiciousAccounts) || []
    },
    sharedDevices() {
      return (this.report && this.report.sharedDevices) || []
    },
    clusters() {
      return (this.report && this.report.clusters) || []
    }
  },
  watch: {
    // 深链直达本页时，比赛信息可能晚于组件挂载，ID就绪后自动加载
    contestId: {
      immediate: true,
      handler(val) {
        if (val && !this.report && !this.loading) {
          this.loadData()
        }
      }
    }
  },
  methods: {
    getErrorMessage(error, fallback = this.$t('m.Plag_Request_Failed')) {
      return error?.message || error?.response?.data?.msg || error?.response?.data?.message || fallback
    },
    async loadData() {
      if (!this.contestId) {
        this.loading = false
        this.errorMessage = this.$t('m.DevAnom_Contest_Not_Loaded')
        return
      }
      this.loading = true
      this.errorMessage = ''
      try {
        const response = await api.getDeviceAnomalies(this.contestId)
        this.report = response.data.data
      } catch (error) {
        this.errorMessage = this.getErrorMessage(error, this.$t('m.DevAnom_Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    shortId(deviceId) {
      if (!deviceId) return this.$t('m.Unknown')
      return deviceId.length > 8 ? deviceId.slice(0, 8) : deviceId
    },
    fmtTime(time) {
      if (!time) return '-'
      const d = new Date(time)
      if (Number.isNaN(d.getTime())) return String(time)
      const pad = (n) => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    },
    exportCsv() {
      if (!this.report) return
      this.exporting = true
      try {
        const rows = [[this.$t('m.DevAnom_CSV_Category'), this.$t('m.DevAnom_CSV_Account_Device'), this.$t('m.DevAnom_CSV_Cluster'), this.$t('m.DevAnom_CSV_Reason'), this.$t('m.DevAnom_CSV_Count'), this.$t('m.DevAnom_CSV_Detail'), this.$t('m.DevAnom_CSV_Related'), this.$t('m.DevAnom_First_Submission'), this.$t('m.DevAnom_Last_Submission')]]
        this.suspiciousAccounts.forEach(account => {
          const reasons = [
            account.multiDevice ? this.$t('m.DevAnom_Reason_Multi_Device') : '',
            account.deviceSharing ? this.$t('m.DevAnom_Reason_Shared_Device') : ''
          ].filter(Boolean).join('+')
          const deviceDetail = (account.devices || [])
            .map(d => `${this.shortId(d.deviceId)}(${this.$t('m.DevAnom_CSV_Times', { count: d.submissionCount })}${(d.otherUsernames || []).length ? this.$t('m.DevAnom_CSV_Same_Device') + d.otherUsernames.join('/') : ''})`)
            .join('; ')
          const related = (account.relatedUsernames || []).join('/')
          rows.push([this.$t('m.DevAnom_CSV_Suspicious_Account'), account.username,
            account.clusterId && related ? '#' + account.clusterId : '',
            reasons, this.$t('m.DevAnom_CSV_Device_Summary', { devices: account.deviceCount, submissions: account.submissionCount }),
            deviceDetail, related,
            this.fmtTime(account.firstSubmitTime), this.fmtTime(account.lastSubmitTime)])
        })
        this.sharedDevices.forEach(device => {
          const userDetail = (device.users || [])
            .map(u => this.$t('m.DevAnom_CSV_User_Detail', { name: u.username, count: u.submissionCount }))
            .join('; ')
          rows.push([this.$t('m.DevAnom_Shared_Device'), this.shortId(device.deviceId), '', this.$t('m.DevAnom_Reason_Multi_Account'),
            this.$t('m.DevAnom_CSV_Account_Summary', { accounts: device.userCount, submissions: device.submissionCount }), userDetail, '',
            this.fmtTime(device.firstSubmitTime), this.fmtTime(device.lastSubmitTime)])
        })
        const csv = '\uFEFF' + rows
          .map(row => row.map(cell => `"${String(cell == null ? '' : cell).replace(/"/g, '""')}"`).join(','))
          .join('\r\n')
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = `device_anomalies_contest_${this.contestId}.csv`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(link.href)
      } finally {
        this.exporting = false
      }
    }
  }
}
</script>

<style scoped>
.device-anomaly-container {
  padding: 4px 0;
}

.page-error {
  margin-bottom: 16px;
}

.loading-container {
  text-align: center;
  padding: 60px 0;
  color: #909399;
}

.loading-container span {
  margin-left: 8px;
}

.summary-card,
.section-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.summary-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

.stat-item {
  flex: 1;
  min-width: 150px;
  text-align: center;
  padding: 12px 0;
  background: #f5f7fa;
  border-radius: 6px;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  color: #303133;
}

.stat-value.suspicious {
  color: #e6a23c;
}

.stat-label {
  margin-top: 4px;
  font-size: 13px;
  color: #909399;
}

.rule-tip {
  margin-top: 4px;
}

.expand-body {
  padding: 8px 16px 12px;
}

.cluster-block {
  border: 1px solid #ebeef5;
  border-left: 3px solid #f56c6c;
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 12px;
  background: #fffafa;
}

.cluster-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.cluster-size {
  color: #606266;
  font-size: 13px;
}

.cluster-members {
  margin-bottom: 8px;
}

.member-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

.cluster-link {
  font-size: 13px;
  color: #606266;
  margin-bottom: 4px;
}

.cluster-link code {
  background: #f5f7fa;
  padding: 1px 6px;
  border-radius: 4px;
  color: #e6a23c;
}

.link-users {
  font-weight: 600;
}

.cluster-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.related-users {
  margin-bottom: 10px;
}

.expand-title {
  font-weight: 600;
  margin-bottom: 8px;
  color: #606266;
}

.other-user-tag {
  margin-right: 6px;
}

.muted {
  color: #c0c4cc;
}

.empty-tip {
  text-align: center;
  color: #909399;
  padding: 12px 0 4px;
}
</style>
