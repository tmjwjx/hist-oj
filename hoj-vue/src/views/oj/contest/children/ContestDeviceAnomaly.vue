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
      <span>加载中...</span>
    </div>

    <template v-else>
      <el-card class="summary-card" shadow="never">
        <div slot="header" class="card-header">
          <span>设备异常筛查</span>
          <div class="header-actions">
            <el-button size="small" icon="el-icon-refresh" :loading="loading" @click="loadData">刷新</el-button>
            <el-button size="small" icon="el-icon-download" :loading="exporting" @click="exportCsv"
              :disabled="!report || report.suspiciousAccountCount === 0">导出CSV</el-button>
          </div>
        </div>
        <div class="summary-row">
          <div class="stat-item">
            <div class="stat-value">{{ report ? report.totalSubmissions : 0 }}</div>
            <div class="stat-label">比赛内提交总数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ report ? report.unknownDeviceSubmissions : 0 }}</div>
            <div class="stat-label">未采集设备的提交（历史数据）</div>
          </div>
          <div class="stat-item">
            <div class="stat-value suspicious">{{ report ? report.suspiciousAccountCount : 0 }}</div>
            <div class="stat-label">作弊嫌疑账号</div>
          </div>
          <div class="stat-item">
            <div class="stat-value suspicious">{{ report && report.sharedDevices ? report.sharedDevices.length : 0 }}</div>
            <div class="stat-label">被共用的设备</div>
          </div>
        </div>
        <el-alert
          title="判定规则：一个账号使用多台不同设备提交，或一台设备出现多个账号提交。设备ID为每台电脑独立生成的随机ID；比赛开始前产生的提交没有设备ID，仅计入“未采集”不参与判定。标记结果供教师人工复核。"
          type="info"
          :closable="false"
          class="rule-tip"
        />
      </el-card>

      <!-- 作弊团伙（传递连通分组） -->
      <el-card v-if="clusters.length > 0" class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>作弊团伙（{{ clusters.length }} 组，沿共用设备逐层串联）</span>
        </div>
        <div v-for="cluster in clusters" :key="cluster.clusterId" class="cluster-block">
          <div class="cluster-title">
            <el-tag size="small" type="danger">团伙 #{{ cluster.clusterId }}</el-tag>
            <span class="cluster-size">{{ cluster.usernames.length }} 个账号</span>
          </div>
          <div class="cluster-members">
            <el-tag v-for="name in cluster.usernames" :key="name" size="medium" effect="dark"
              type="danger" class="member-tag">{{ name }}</el-tag>
          </div>
          <div v-for="link in cluster.links" :key="link.deviceId" class="cluster-link">
            <i class="el-icon-connection"></i>
            共用设备 <code>{{ shortId(link.deviceId) }}</code>：
            <span class="link-users">{{ link.usernames.join(' ⇄ ') }}</span>
          </div>
          <div v-if="cluster.links.length === 0" class="cluster-link muted">
            该账号单独命中“一账号多设备”规则，未与其它账号共道设备
          </div>
        </div>
        <div class="cluster-tip">关联关系沿设备逐层推进：A 的设备上有 B，B 的另一台设备上有 C，则 A、B、C 同属一个团伙。</div>
      </el-card>

      <!-- 嫌疑账号清单 -->
      <el-card class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>作弊嫌疑账号（{{ suspiciousAccounts.length }}）</span>
        </div>
        <el-table :data="suspiciousAccounts" border stripe row-key="uid">
          <el-table-column type="expand">
            <template slot-scope="scope">
              <div class="expand-body">
                <div class="expand-title">团伙内关联账号（沿共用设备逐层推进）</div>
                <div v-if="scope.row.relatedUsernames && scope.row.relatedUsernames.length > 0"
                  class="related-users">
                  <el-tag v-for="name in scope.row.relatedUsernames" :key="name" size="small"
                    type="danger" class="member-tag">{{ name }}</el-tag>
                </div>
                <div v-else class="muted" style="margin-bottom:10px;">无（该账号单独命中规则）</div>
                <div class="expand-title">该账号的设备使用明细</div>
                <el-table :data="scope.row.devices" size="small" border>
                  <el-table-column label="设备ID" width="140">
                    <template slot-scope="d">{{ shortId(d.row.deviceId) }}</template>
                  </el-table-column>
                  <el-table-column prop="userAgent" label="浏览器标识 (User-Agent)" min-width="220" show-overflow-tooltip>
                    <template slot-scope="d">{{ d.row.userAgent || '未知' }}</template>
                  </el-table-column>
                  <el-table-column prop="submissionCount" label="提交数" width="80" align="center" />
                  <el-table-column label="首次提交" width="160">
                    <template slot-scope="d">{{ fmtTime(d.row.firstSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column label="最后提交" width="160">
                    <template slot-scope="d">{{ fmtTime(d.row.lastSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column label="该设备上的其它账号" min-width="160">
                    <template slot-scope="d">
                      <el-tag v-for="name in d.row.otherUsernames" :key="name" size="mini" type="danger"
                        class="other-user-tag">{{ name }}</el-tag>
                      <span v-if="!d.row.otherUsernames || d.row.otherUsernames.length === 0" class="muted">无</span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="username" label="账号" min-width="110" />
          <el-table-column label="团伙" width="80" align="center">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.clusterId && scope.row.relatedUsernames
                && scope.row.relatedUsernames.length > 0" size="mini" type="danger">
                #{{ scope.row.clusterId }}
              </el-tag>
              <span v-else class="muted">-</span>
            </template>
          </el-table-column>
          <el-table-column label="命中原因" width="190">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.multiDevice" size="small" type="warning">一账号多设备</el-tag>
              <el-tag v-if="scope.row.deviceSharing" size="small" type="danger">使用了共用设备</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="deviceCount" label="设备数" width="80" align="center" />
          <el-table-column prop="submissionCount" label="提交数" width="80" align="center" />
          <el-table-column label="首次提交" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.firstSubmitTime) }}</template>
          </el-table-column>
          <el-table-column label="最后提交" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.lastSubmitTime) }}</template>
          </el-table-column>
        </el-table>
        <div v-if="suspiciousAccounts.length === 0" class="empty-tip">未发现设备异常，暂无嫌疑账号。</div>
      </el-card>

      <!-- 共用设备清单 -->
      <el-card class="section-card" shadow="never">
        <div slot="header" class="card-header">
          <span>被共用的设备（{{ sharedDevices.length }}）</span>
        </div>
        <el-table :data="sharedDevices" border stripe row-key="deviceId">
          <el-table-column type="expand">
            <template slot-scope="scope">
              <div class="expand-body">
                <div class="expand-title">使用过该设备的账号</div>
                <el-table :data="scope.row.users" size="small" border>
                  <el-table-column prop="username" label="账号" min-width="120" />
                  <el-table-column prop="submissionCount" label="在该设备上的提交数" width="160" align="center" />
                  <el-table-column label="首次提交" width="160">
                    <template slot-scope="u">{{ fmtTime(u.row.firstSubmitTime) }}</template>
                  </el-table-column>
                  <el-table-column label="最后提交" width="160">
                    <template slot-scope="u">{{ fmtTime(u.row.lastSubmitTime) }}</template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="设备ID" width="140">
            <template slot-scope="scope">{{ shortId(scope.row.deviceId) }}</template>
          </el-table-column>
          <el-table-column prop="userAgent" label="浏览器标识 (User-Agent)" min-width="220" show-overflow-tooltip>
            <template slot-scope="scope">{{ scope.row.userAgent || '未知' }}</template>
          </el-table-column>
          <el-table-column prop="userCount" label="账号数" width="80" align="center" />
          <el-table-column prop="submissionCount" label="提交数" width="80" align="center" />
          <el-table-column label="首次提交" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.firstSubmitTime) }}</template>
          </el-table-column>
          <el-table-column label="最后提交" width="160">
            <template slot-scope="scope">{{ fmtTime(scope.row.lastSubmitTime) }}</template>
          </el-table-column>
        </el-table>
        <div v-if="sharedDevices.length === 0" class="empty-tip">未发现多账号共用的设备。</div>
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
    getErrorMessage(error, fallback = '请求失败') {
      return error?.message || error?.response?.data?.msg || error?.response?.data?.message || fallback
    },
    async loadData() {
      if (!this.contestId) {
        this.loading = false
        this.errorMessage = '比赛信息尚未加载完成，请稍后重试'
        return
      }
      this.loading = true
      this.errorMessage = ''
      try {
        const response = await api.getDeviceAnomalies(this.contestId)
        this.report = response.data.data
      } catch (error) {
        this.errorMessage = this.getErrorMessage(error, '设备异常数据加载失败')
      } finally {
        this.loading = false
      }
    },
    shortId(deviceId) {
      if (!deviceId) return '未知'
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
        const rows = [['类别', '账号/设备', '团伙', '命中原因', '数量', '明细', '关联账号', '首次提交', '最后提交']]
        this.suspiciousAccounts.forEach(account => {
          const reasons = [
            account.multiDevice ? '一账号多设备' : '',
            account.deviceSharing ? '使用了共用设备' : ''
          ].filter(Boolean).join('+')
          const deviceDetail = (account.devices || [])
            .map(d => `${this.shortId(d.deviceId)}(${d.submissionCount}次${(d.otherUsernames || []).length ? ',同设备:' + d.otherUsernames.join('/') : ''})`)
            .join('; ')
          const related = (account.relatedUsernames || []).join('/')
          rows.push(['嫌疑账号', account.username,
            account.clusterId && related ? '#' + account.clusterId : '',
            reasons, account.deviceCount + '台设备/' + account.submissionCount + '次提交',
            deviceDetail, related,
            this.fmtTime(account.firstSubmitTime), this.fmtTime(account.lastSubmitTime)])
        })
        this.sharedDevices.forEach(device => {
          const userDetail = (device.users || [])
            .map(u => `${u.username}(${u.submissionCount}次)`)
            .join('; ')
          rows.push(['共用设备', this.shortId(device.deviceId), '', '一设备多账号',
            device.userCount + '个账号/' + device.submissionCount + '次提交', userDetail, '',
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
