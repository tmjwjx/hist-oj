<template>
  <div class="plagiarism-container">
    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      closable
      class="page-error"
      @close="errorMessage = ''"
    />

    <!-- 加载中 -->
    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <span>{{ $t('m.Plag_Loading') }}</span>
    </div>

    <!-- 比赛未结束提示 -->
    <el-alert
      v-else-if="!isContestEnded"
      :title="$t('m.Plag_Contest_Not_Ended_Title')"
      type="warning"
      :description="$t('m.Plag_Contest_Not_Ended_Desc')"
      show-icon
      :closable="false"
      style="margin-bottom: 20px"
    />

    <!-- 比赛已结束，显示查重功能 -->
    <div v-else>
      <!-- 步骤1: 设置查重率 -->
      <el-card class="config-card" v-if="currentStep === 'config'">
        <div slot="header" class="card-header">
          <span>{{ $t('m.Plag_Step1_Title') }}</span>
          <el-button type="primary" size="small" @click="saveConfig" :loading="saving" :disabled="isCheckRunning">
            {{ $t('m.Plag_Save_Config') }}
          </el-button>
        </div>

        <el-alert
          :title="$t('m.Plag_Desc_Title')"
          type="info"
          :description="$t('m.Plag_Desc')"
          show-icon
          :closable="false"
          style="margin-bottom: 20px"
        />

        <el-table :data="problems" style="width: 100%">
          <el-table-column prop="displayId" :label="$t('m.Plag_Problem_Number')" width="100" />
          <el-table-column prop="displayTitle" :label="$t('m.Plag_Problem_Name')" />
          <el-table-column :label="$t('m.Plag_Threshold_Label')" width="250">
            <template slot-scope="scope">
              <el-input-number
                v-model="scope.row.threshold"
                :min="0"
                :max="100"
                :step="5"
                size="small"
              />
              <span style="margin-left: 10px; color: #909399">
                {{ $t('m.Plag_Threshold_Tip') }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 步骤2: 开始查重 -->
      <el-card class="check-card" v-if="currentStep === 'check' || currentStep === 'result'">
        <div slot="header" class="card-header">
          <span>{{ $t('m.Plag_Step2_Title') }}</span>
          <div>
            <!-- check 步骤的按钮 -->
            <template v-if="currentStep === 'check'">
              <!-- 没有查重任务 或 任务已完成/失败时，显示"开始查重"按钮 -->
              <el-button
                v-if="!checkStatus || checkStatus.status === 'pending' || checkStatus.status === 'completed' || checkStatus.status === 'failed'"
                type="primary"
                size="small"
                @click="startCheck"
                :loading="starting"
                :disabled="checkStatus && checkStatus.status === 'running' && starting"
              >
                {{ checkStatus && (checkStatus.status === 'completed' || checkStatus.status === 'failed') ? $t('m.Plag_Recheck') : $t('m.Plag_Start_Check') }}
              </el-button>
              <!-- 查重进行中时，禁用按钮并显示状态 -->
              <el-button
                v-if="checkStatus && checkStatus.status === 'running'"
                type="info"
                size="small"
                :loading="true"
                disabled
              >
                {{ $t('m.Plag_Checking', { progress: checkStatus.progress }) }}
              </el-button>
            </template>
            <!-- result 步骤的按钮 -->
            <template v-if="currentStep === 'result'">
              <el-button
                type="primary"
                size="small"
                @click="restartCheck"
                :loading="starting"
              >
                {{ $t('m.Plag_Recheck') }}
              </el-button>
              <el-button
                type="success"
                size="small"
                @click="resetConfig"
              >
                {{ $t('m.Plag_Reset_Config') }}
              </el-button>
            </template>
          </div>
        </div>

        <!-- 查重进度 -->
        <div v-if="checkStatus">
          <el-row :gutter="20" style="margin-bottom: 20px">
            <el-col :span="6">
              <statistic-card :title="$t('m.Status')" :value="getStatusText(checkStatus.status)" />
            </el-col>
            <el-col :span="6">
              <statistic-card :title="$t('m.Plag_Total_Submissions')" :value="checkStatus.totalSubmissions" />
            </el-col>
            <el-col :span="6">
              <statistic-card :title="$t('m.Plag_Checked_Pairs')" :value="`${checkStatus.checkedPairs}/${checkStatus.totalPairs}`" />
            </el-col>
            <el-col :span="6">
              <statistic-card :title="$t('m.Progress')" :value="`${checkStatus.progress}%`" />
            </el-col>
          </el-row>

          <el-progress
            :percentage="checkStatus.progress"
            :status="checkStatus.status === 'completed' ? 'success' : ''"
            :stroke-width="20"
          />

          <el-alert
            v-if="checkStatus.status === 'running'"
            :title="$t('m.Plag_Running_Tip')"
            type="info"
            :closable="false"
            style="margin-top: 20px"
          />
        </div>
      </el-card>

      <!-- 步骤3: 查看结果 -->
      <el-card class="result-card" v-if="currentStep === 'result' && checkStatus && checkStatus.status === 'completed'">
        <div slot="header" class="card-header">
          <span>{{ $t('m.Plag_Step3_Title') }}</span>
          <div>
            <el-input
              v-model="displayIdFilter"
              :placeholder="$t('m.Plag_Filter_Placeholder')"
              style="width: 200px; margin-right: 10px"
              clearable
              @clear="handleFilterChange"
              @keyup.enter.native="handleFilterChange"
            >
              <el-button slot="append" icon="el-icon-search" @click="handleFilterChange" />
            </el-input>
            <el-button
              type="success"
              size="small"
              icon="el-icon-download"
              @click="exportExcel"
              :loading="exporting"
              :disabled="exporting"
            >
              {{ exporting ? $t('m.Plag_Exporting') : $t('m.Export_Excel') }}
            </el-button>
          </div>
        </div>

        <!-- 提示：只显示超过阈值的结果 -->
        <el-alert
          :title="$t('m.Plag_Data_Note_Title')"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        >
          <template slot="default">
            <div>{{ $t('m.Plag_Note1_Before') }}<strong>{{ $t('m.Plag_Note1_Highlight') }}</strong>{{ $t('m.Plag_Note1_After', { total: totalCount }) }}</div>
            <div style="margin-top: 8px">
              {{ $t('m.Plag_Note2_Before') }}<strong>{{ $t('m.Plag_Note2_Highlight') }}</strong>{{ $t('m.Plag_Note2_After') }}
            </div>
          </template>
        </el-alert>

        <!-- 超过 500 条提示 -->
        <el-alert
          v-if="totalCount > 500"
          :title="$t('m.Plag_Too_Many_Title')"
          type="warning"
          :closable="false"
          style="margin-bottom: 20px"
        >
          {{ $t('m.Plag_Too_Many_Desc', { total: totalCount }) }}
        </el-alert>

        <!-- 结果统计 -->
        <el-row :gutter="20" style="margin-bottom: 20px">
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ totalCount }}</div>
            <div class="stat-label">{{ $t('m.Plag_Bidirectional_Count') }}</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover" type="danger">
              <div class="stat-item">
                <div class="stat-value danger">{{ $t('m.Plag_People', { count: suspiciousUsers }) }}</div>
                <div class="stat-label">{{ $t('m.Plag_Suspicious_User_Count') }}</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ results.length }}</div>
                <div class="stat-label">{{ $t('m.Plag_Currently_Shown') }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 结果列表 -->
        <el-table :data="pagedResults" style="width: 100%" stripe v-loading="loading">
          <el-table-column prop="displayId" :label="$t('m.Plag_Problem_Number')" width="80" />
          <el-table-column prop="problemTitle" :label="$t('m.Plag_Problem_Name')" width="200" />
          <el-table-column :label="$t('m.Plag_User_Compare')" width="250">
            <template slot-scope="scope">
              <el-tag type="info" size="small">{{ scope.row.username1 }}</el-tag>
              <span style="margin: 0 10px">vs</span>
              <el-tag type="info" size="small">{{ scope.row.username2 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Plag_Submission_ID')" width="180">
            <template slot-scope="scope">
              <div style="font-size: 12px">
                <div>A: {{ scope.row.contestRecordId1 || scope.row.submitId1 }}</div>
                <div>B: {{ scope.row.contestRecordId2 || scope.row.submitId2 }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Plag_Similarity')" width="180">
            <template slot-scope="scope">
              <el-tag
                :type="getSimilarityType(scope.row.maxSimilarity)"
                size="medium"
              >
                {{ scope.row.similarity1to2 }}% / {{ scope.row.similarity2to1 }}%
              </el-tag>
              <el-progress
                :percentage="scope.row.maxSimilarity"
                :color="getSimilarityColor(scope.row.maxSimilarity)"
                :show-text="false"
                :stroke-width="4"
                style="margin-top: 5px"
              />
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Plag_Language')" width="120">
            <template slot-scope="scope">
              {{ scope.row.language }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Operation')" fixed="right" width="150">
            <template slot-scope="scope">
              <el-button
                size="mini"
                type="primary"
                icon="el-icon-view"
                :loading="loadingCode"
                :disabled="loadingCode"
                @click="viewCode(scope.row)"
              >
                {{ $t('m.Plag_Compare_Code') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          v-if="results.length > 0"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="currentPage"
          :page-sizes="[10, 20, 50, 100]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="results.length"
          style="margin-top: 20px; text-align: right"
        />
      </el-card>
    </div>

    <!-- 代码对比对话框 -->
    <el-dialog
      :title="$t('m.Plag_Code_Compare_Title')"
      :visible.sync="codeDialogVisible"
      width="90%"
      :close-on-click-modal="false"
    >
      <!-- 提交信息 -->
      <el-row :gutter="20" style="margin-bottom: 20px">
        <el-col :span="12">
          <el-card shadow="hover">
            <div slot="header">
              <span style="font-weight: bold">{{ $t('m.Plag_User_A', { name: codeData.user1?.username }) }}</span>
            </div>
            <div class="submission-info">
              <p><strong>{{ $t('m.Plag_Problem_Number_Label') }}</strong>{{ codeData.displayId }}</p>
              <p><strong>{{ $t('m.Plag_Problem_Label') }}</strong>{{ codeData.problemTitle }}</p>
              <p><strong>{{ $t('m.Plag_Submit_Time_Label') }}</strong>{{ codeData.submitTime1 }}</p>
              <p><strong>{{ $t('m.Plag_Language_Label') }}</strong>{{ codeData.language }}</p>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover">
            <div slot="header">
              <span style="font-weight: bold">{{ $t('m.Plag_User_B', { name: codeData.user2?.username }) }}</span>
            </div>
            <div class="submission-info">
              <p><strong>{{ $t('m.Plag_Problem_Number_Label') }}</strong>{{ codeData.displayId }}</p>
              <p><strong>{{ $t('m.Plag_Problem_Label') }}</strong>{{ codeData.problemTitle }}</p>
              <p><strong>{{ $t('m.Plag_Submit_Time_Label') }}</strong>{{ codeData.submitTime2 }}</p>
              <p><strong>{{ $t('m.Plag_Language_Label') }}</strong>{{ codeData.language }}</p>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 代码对比区域 -->
      <div class="code-compare-container">
        <div class="code-panel">
          <div class="code-panel-header">
            <span>{{ $t('m.Plag_Code_Of', { name: codeData.user1?.username }) }}</span>
          </div>
          <div class="code-viewer" v-if="codeData.code1">
            <pre><code :key="`code1-${codeDialogVisible}`" ref="codeBlock1" :class="`language-${mapLanguage(codeData.language1)}`">{{ codeData.code1 }}</code></pre>
          </div>
        </div>
        <div class="code-panel">
          <div class="code-panel-header">
            <span>{{ $t('m.Plag_Code_Of', { name: codeData.user2?.username }) }}</span>
          </div>
          <div class="code-viewer" v-if="codeData.code2">
            <pre><code :key="`code2-${codeDialogVisible}`" ref="codeBlock2" :class="`language-${mapLanguage(codeData.language2)}`">{{ codeData.code2 }}</code></pre>
          </div>
        </div>
      </div>

      <div slot="footer" class="dialog-footer">
        <el-button @click="codeDialogVisible = false">{{ $t('m.Close') }}</el-button>
        <el-button type="primary" @click="copyCode(1)">{{ $t('m.Plag_Copy_Code_A') }}</el-button>
        <el-button type="primary" @click="copyCode(2)">{{ $t('m.Plag_Copy_Code_B') }}</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import api from '@/api/plagiarism'
import { mapGetters, mapState, mapActions } from 'vuex'
import { CONTEST_STATUS } from '@/common/constants'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// StatisticCard 简单统计卡片组件
const StatisticCard = {
  props: ['title', 'value'],
  template: `
    <el-card shadow="hover">
      <div class="stat-item">
        <div class="stat-value">{{ value }}</div>
        <div class="stat-label">{{ title }}</div>
      </div>
    </el-card>
  `
}

export default {
  name: 'ContestPlagiarism',
  components: {
    StatisticCard
  },
  data() {
    return {
      loading: true, // 加载状态
      currentStep: 'config', // config, check, result
      problems: [],
      configs: [],
      saving: false,
      starting: false,
      exporting: false, // 导出状态
      checkStatus: null,
      results: [],
      totalCount: 0, // 双向均达到阈值的结果总数
      displayIdFilter: '', // 题号筛选
      progressTimer: null,
      codeDialogVisible: false,
      codeData: {
        code1: '',
        code2: '',
        language1: '',
        language2: '',
        user1: null,
        user2: null,
        displayId: '',
        problemTitle: '',
        submitTime1: '',
        submitTime2: '',
        language: ''
      },
      currentPage: 1,
      pageSize: 20,
      loadingCode: false, // 查看代码的加载状态
      errorMessage: '',
      dataLoading: false,
      loadingResults: false,
      progressRequestPending: false,
      progressErrorCount: 0
    }
  },
  computed: {
    ...mapGetters(['isSuperAdmin', 'isContestAdmin', 'userInfo']),
    ...mapState({
      contest: (state) => state.contest.contest,
      contestProblems: (state) => state.contest.contestProblems,
    }),
    // 判断 contest 数据是否已加载
    isContestLoaded() {
      return this.contest && this.contest.id !== undefined
    },
    contestId() {
      return this.contest && this.contest.id
    },
    isContestEnded() {
      return this.contest && Number(this.contest.status) === CONTEST_STATUS.ENDED
    },
    suspiciousUsers() {
      const users = new Set()
      ;(this.results || []).forEach(r => {
        users.add(r.username1)
        users.add(r.username2)
      })
      return users.size
    },
    // 分页后的结果（只渲染当前页的数据）
    pagedResults() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return (this.results || []).slice(start, end)
    },
    isCheckRunning() {
      return this.checkStatus && (this.checkStatus.status === 'running' || this.checkStatus.status === 'pending')
    }
  },
  beforeDestroy() {
    if (this.progressTimer) {
      clearInterval(this.progressTimer)
    }
  },
  methods: {
    ...mapActions(['getContestProblems']),
    getErrorMessage(error, fallback = this.$t('m.Plag_Request_Failed')) {
      return error?.message || error?.response?.data?.msg || error?.response?.data?.message || fallback
    },
    showError(prefix, error, fallback) {
      const detail = this.getErrorMessage(error, fallback)
      this.errorMessage = prefix ? `${prefix}：${detail}` : detail
    },
    clearError() {
      this.errorMessage = ''
    },
    async loadData() {
      if (this.dataLoading) return
      // 如果比赛未结束，设置 loading 为 false 并返回
      if (!this.isContestEnded) {
        this.loading = false
        return
      }

      this.dataLoading = true
      this.clearError()
      try {
        // 确保 contestProblems 已加载
        if (!this.contestProblems || this.contestProblems.length === 0) {
          await this.getContestProblems()
        }

        // 从 Vuex store 获取题目并设置默认阈值
        this.problems = (this.contestProblems || []).map(p => ({
          ...p,
          cpid: p.id,  // 添加 cpid 字段，指向 contest_problem 表的主键 id
          threshold: 50 // 默认阈值50%
        }))

        // 获取查重配置
        const configRes = await api.getPlagiarismConfig(this.contest.id)
        if (configRes.data?.data && configRes.data.data.length > 0) {
          // 应用已有配置
          configRes.data.data.forEach(cfg => {
            const problem = this.problems.find(p => p.cpid === cfg.cpid)
            if (problem) {
              problem.threshold = cfg.threshold
            }
          })
          this.currentStep = 'check'
        }

        // 获取最新查重任务（优先获取，以正确恢复状态）
        const checkRes = await api.getLatestPlagiarismCheck(this.contest.id)
        if (checkRes.data?.data) {
          this.checkStatus = checkRes.data.data

          // 根据查重任务状态决定显示哪个步骤
          if (this.checkStatus.status === 'completed') {
            this.currentStep = 'result'
            await this.loadResults()
          } else if (this.checkStatus.status === 'running' || this.checkStatus.status === 'pending') {
            // running 或 pending 状态，显示 check 步骤并开始轮询
            this.currentStep = 'check'
            // 无论是 running 还是 pending，都启动轮询
            this.startPolling()
          } else {
            // failed 或其他状态，显示 config 步骤
            this.currentStep = 'config'
          }
        } else {
          // 没有查重任务，显示 config 步骤
          this.currentStep = 'config'
        }

        // 如果配置为空，尝试加载默认配置（只在config步骤时）
        if (this.currentStep === 'config' && (!this.problems || this.problems.length === 0)) {
          // 确保 contestProblems 已加载
          if (!this.contestProblems || this.contestProblems.length === 0) {
            await this.getContestProblems()
          }

          // 从 Vuex store 获取题目并设置默认阈值
          this.problems = (this.contestProblems || []).map(p => ({
            ...p,
            cpid: p.id,  // 添加 cpid 字段，指向 contest_problem 表的主键 id
            threshold: 50 // 默认阈值50%
          }))

          // 获取查重配置
          const configRes = await api.getPlagiarismConfig(this.contest.id)
          if (configRes.data?.data && configRes.data.data.length > 0) {
            // 应用已有配置
            configRes.data.data.forEach(cfg => {
              const problem = this.problems.find(p => p.cpid === cfg.cpid)
              if (problem) {
                problem.threshold = cfg.threshold
              }
            })
          }
        }
      } catch (error) {
        console.error('加载数据失败:', error)
        this.showError(this.$t('m.Plag_Load_Data_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        this.dataLoading = false
        this.loading = false
      }
    },

    async saveConfig() {
      this.saving = true
      this.clearError()
      try {
        const configs = this.problems.map(p => ({
          cpid: p.cpid,
          threshold: p.threshold
        }))

        await api.savePlagiarismConfig(this.contest.id, configs)
        this.$message.success(this.$t('m.Plag_Config_Saved'))

        // 保存配置后，检查是否有查重任务
        const checkRes = await api.getLatestPlagiarismCheck(this.contest.id)
        if (checkRes.data.data) {
          this.checkStatus = checkRes.data.data
          // 如果有查重任务（无论什么状态），都显示 check 或 result 步骤
          if (this.checkStatus.status === 'completed') {
            this.currentStep = 'result'
            await this.loadResults()
          } else {
            // 如果是 running 或 pending，显示 check 步骤并开始轮询
            this.currentStep = 'check'
            // 无论是 running 还是 pending，都启动轮询以监控状态变化
            this.startPolling()
          }
        } else {
          // 没有查重任务，显示 check 步骤
          this.currentStep = 'check'
        }
      } catch (error) {
        this.showError(this.$t('m.Plag_Config_Save_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        this.saving = false
      }
    },

    async startCheck() {
      // 防止重复点击
      if (this.starting) return

      // 如果已经有正在运行的查重任务，提示用户
      if (this.checkStatus && this.checkStatus.status === 'running') {
        this.$message.warning(this.$t('m.Plag_Check_Running'))
        return
      }

      this.starting = true
      this.clearError()
      try {
        const res = await api.startPlagiarismCheck(this.contest.id)
        this.$message.success(this.$t('m.Plag_Check_Started'))

        // 立即更新 checkStatus，并乐观地设置为 running 状态，确保UI能显示进度
        if (res.data.data) {
          this.checkStatus = {
            ...res.data.data,
            status: 'running',  // 强制设置为 running，确保 UI 正确显示
            progress: res.data.data.progress || 0
          }
        } else {
          // 如果 API 没有返回数据，创建一个新的 running 状态
          this.checkStatus = {
            status: 'running',
            progress: 0
          }
        }

        // 切换到 check 步骤并开始轮询
        this.currentStep = 'check'

        // 启动定时轮询
        this.startPolling()
      } catch (error) {
        this.showError(this.$t('m.Plag_Start_Check_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        this.starting = false
      }
    },

    async restartCheck() {
      // 防止重复点击
      if (this.starting) return

      this.$confirm(this.$t('m.Plag_Recheck_Confirm'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.OK'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        // 停止当前的轮询
        if (this.progressTimer) {
          clearInterval(this.progressTimer)
          this.progressTimer = null
        }

        this.starting = true
        this.clearError()
        try {
          const res = await api.startPlagiarismCheck(this.contest.id)
          this.$message.success(this.$t('m.Plag_Check_Restarted'))

          // 清空旧结果
          this.results = []

          // 立即更新 checkStatus，并乐观地设置为 running 状态
          if (res.data.data) {
            this.checkStatus = {
              ...res.data.data,
              status: 'running',  // 强制设置为 running，确保 UI 正确显示
              progress: 0
            }
          } else {
            // 如果 API 没有返回数据，创建一个新的 running 状态
            this.checkStatus = {
              status: 'running',
              progress: 0
            }
          }

          // 切换到 check 步骤
          this.currentStep = 'check'

          // 启动定时轮询
          this.startPolling()
        } catch (error) {
          this.showError(this.$t('m.Plag_Recheck_Failed'), error, this.$t('m.Plag_Retry_Later'))
        } finally {
          this.starting = false
        }
      }).catch(() => {
        // 用户取消
      })
    },

    startPolling() {
      // 避免重复启动定时器
      if (this.progressTimer) {
        clearInterval(this.progressTimer)
      }
      this.progressErrorCount = 0

      // 立即执行一次
      this.pollProgress()

      // 启动定时轮询
      this.progressTimer = setInterval(() => {
        this.pollProgress()
      }, 2000) // 每2秒轮询一次
    },

    async pollProgress() {
      if (this.progressRequestPending) return
      this.progressRequestPending = true
      try {
        const res = await api.getPlagiarismProgress(this.contest.id)
        this.progressErrorCount = 0

        if (!res.data.data) {
          console.warn('未获取到查重进度数据')
          return
        }

        // 更新 checkStatus，保留现有的状态数据
        this.checkStatus = {
          ...this.checkStatus,
          ...res.data.data
        }

        if (this.checkStatus.status === 'completed') {
          // 清除定时器
          if (this.progressTimer) {
            clearInterval(this.progressTimer)
            this.progressTimer = null
          }
          this.currentStep = 'result'
          await this.loadResults()
          this.$message.success(this.$t('m.Plag_Check_Completed'))
        } else if (this.checkStatus.status === 'failed') {
          // 清除定时器
          if (this.progressTimer) {
            clearInterval(this.progressTimer)
            this.progressTimer = null
          }
          this.errorMessage = this.$t('m.Plag_Check_Failed') + (this.checkStatus.errorMessage || this.$t('m.Plag_Unknown_Error'))
        }
        // 如果是 running 状态，继续轮询（由定时器处理）
      } catch (error) {
        console.error('获取进度失败:', error)
        this.progressErrorCount += 1
        if (this.progressErrorCount === 1) {
          this.showError(this.$t('m.Plag_Progress_Failed'), error, this.$t('m.Plag_Auto_Retry'))
        }
        // 连续失败三次再停止，避免一次瞬时网络错误中断任务状态更新。
        if (this.progressErrorCount >= 3) {
          if (this.progressTimer) {
            clearInterval(this.progressTimer)
            this.progressTimer = null
          }
        }
      } finally {
        this.progressRequestPending = false
      }
    },

    async loadResults() {
      if (!this.checkStatus || this.loadingResults) return

      this.loadingResults = true
      this.loading = true
      this.clearError()
      try {
        const res = await api.getPlagiarismResults(this.checkStatus.id, this.displayIdFilter)
        this.results = res.data.data || []
        this.totalCount = res.data.totalCount || 0
        // 重置到第一页
        this.currentPage = 1
      } catch (error) {
        console.error('加载结果失败:', error)
        this.showError(this.$t('m.Plag_Load_Results_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        this.loadingResults = false
        this.loading = false
      }
    },

    handleFilterChange() {
      // 切换过滤器时重新加载结果
      this.loadResults()
    },

    async viewCode(result) {
      if (this.loadingCode) return // 防止重复点击

      this.loadingCode = true
      try {
        // 同时获取两份代码
        const [res1, res2] = await Promise.all([
          api.getPlagiarismSubmission(result.submitId1),
          api.getPlagiarismSubmission(result.submitId2)
        ])

        const submission1 = res1.data.data
        const submission2 = res2.data.data

        // 格式化提交时间
        const formatTime = (time) => {
          if (!time) return ''
          return new Date(time).toLocaleString('zh-CN')
        }

        // 设置代码数据
        this.codeData = {
          code1: submission1.code,
          code2: submission2.code,
          language1: submission1.language.toLowerCase(),
          language2: submission2.language.toLowerCase(),
          user1: { username: result.username1 },
          user2: { username: result.username2 },
          displayId: result.displayId,
          problemTitle: result.problemTitle,
          submitTime1: formatTime(submission1.submitTime),
          submitTime2: formatTime(submission2.submitTime),
          language: submission1.language
        }

        this.codeDialogVisible = true

        // 等待 DOM 更新后进行代码高亮
        this.$nextTick(() => {
          if (this.$refs.codeBlock1) {
            hljs.highlightElement(this.$refs.codeBlock1)
          }
          if (this.$refs.codeBlock2) {
            hljs.highlightElement(this.$refs.codeBlock2)
          }
        })
      } catch (error) {
        this.showError(this.$t('m.Plag_Get_Code_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        this.loadingCode = false
      }
    },

    copyCode(userNum) {
      const code = userNum === 1 ? this.codeData.code1 : this.codeData.code2
      const user = userNum === 1 ? this.codeData.user1?.username : this.codeData.user2?.username

      navigator.clipboard.writeText(code).then(() => {
        this.$message.success(this.$t('m.Plag_Code_Copied', { name: user }))
      })
    },

    async exportExcel() {
      // 设置导出状态
      this.exporting = true

      // 显示提示
      const loadingMessage = this.$message({
        message: this.$t('m.Plag_Generating_Excel'),
        type: 'info',
        duration: 0,
        showClose: false
      })

      try {
        const res = await api.exportPlagiarismResults(this.checkStatus.id)
        const blob = new Blob([res.data], {
          type: 'text/csv;charset=utf-8'
        })
        const link = document.createElement('a')
        link.href = window.URL.createObjectURL(blob)
        link.download = `${this.$t('m.Plag_Result_File')}_${this.contest.title}_${Date.now()}.csv`
        link.click()

        // 关闭加载提示
        loadingMessage.close()

        this.$message.success(this.$t('m.Plag_Export_Success'))
      } catch (error) {
        // 关闭加载提示
        loadingMessage.close()

        this.showError(this.$t('m.Plag_Export_Failed'), error, this.$t('m.Plag_Retry_Later'))
      } finally {
        // 恢复按钮状态
        this.exporting = false
      }
    },

    resetConfig() {
      // 停止轮询
      if (this.progressTimer) {
        clearInterval(this.progressTimer)
        this.progressTimer = null
      }

      this.currentStep = 'config'
      this.checkStatus = null
      this.results = []
      this.starting = false
      this.exporting = false // 重置导出状态
      this.clearError()
    },

    handleSizeChange(val) {
      this.pageSize = val
      this.currentPage = 1 // 重置到第一页
    },

    handleCurrentChange(val) {
      this.currentPage = val
      // 滚动到顶部
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },

    getStatusText(status) {
      const statusMap = {
        pending: this.$t('m.Plag_Status_Pending'),
        running: this.$t('m.Plag_Status_Running'),
        completed: this.$t('m.Plag_Status_Completed'),
        failed: this.$t('m.Plag_Status_Failed')
      }
      return statusMap[status] || status
    },

    getSimilarityType(similarity) {
      if (similarity >= 80) return 'danger'
      if (similarity >= 50) return 'warning'
      return 'success'
    },

    getSimilarityColor(similarity) {
      if (similarity >= 80) return '#F56C6C'
      if (similarity >= 50) return '#E6A23C'
      return '#67C23A'
    },

    mapLanguage(lang) {
      // 映射编程语言到 highlight.js 支持的语言标识
      const languageMap = {
        // C语言变体
        'c': 'c',
        'C': 'c',
        'C With O2': 'c',
        'c with o2': 'c',

        // C++变体
        'cpp': 'cpp',
        'C++': 'cpp',
        'c++': 'cpp',
        'C++ 17 With O2': 'cpp',
        'c++ 17 with o2': 'cpp',
        'C++ 17': 'cpp',
        'c++ 17': 'cpp',
        'C++ 20 With O2': 'cpp',
        'c++ 20 with o2': 'cpp',
        'C++ 20': 'cpp',
        'c++ 20': 'cpp',

        // Java
        'java': 'java',
        'Java': 'java',

        // Python变体
        'python': 'python',
        'Python': 'python',
        'py': 'python',
        'python3': 'python',
        'Python3': 'python',
        'python2': 'python',
        'Python2': 'python',
        'pypy3': 'python',
        'PyPy3': 'python',
        'pypy2': 'python',
        'PyPy2': 'python',

        // Go
        'go': 'go',
        'golang': 'go',
        'Go': 'go',

        // Rust
        'rust': 'rust',
        'Rust': 'rust',

        // JavaScript变体
        'javascript': 'javascript',
        'js': 'javascript',
        'JavaScript': 'javascript',
        'javascript node': 'javascript',
        'JavaScript Node': 'javascript',
        'javascript v8': 'javascript',
        'JavaScript V8': 'javascript',

        // TypeScript
        'typescript': 'typescript',
        'ts': 'typescript',
        'TypeScript': 'typescript',

        // PHP
        'php': 'php',
        'PHP': 'php',

        // Ruby
        'ruby': 'ruby',
        'Ruby': 'ruby',

        // Kotlin
        'kotlin': 'kotlin',
        'Kotlin': 'kotlin',

        // Scala
        'scala': 'scala',
        'Scala': 'scala',

        // C#
        'csharp': 'csharp',
        'c#': 'csharp',
        'C#': 'csharp',

        // Shell
        'shell': 'bash',
        'bash': 'bash',

        // SQL
        'sql': 'sql',

        // Web
        'html': 'html',
        'css': 'css',
        'xml': 'xml',
        'json': 'json',
        'yaml': 'yaml',
        'yml': 'yaml',

        // Markdown
        'markdown': 'markdown',
        'md': 'markdown'
      }
      return languageMap[lang] || 'plaintext'  // 默认使用纯文本
    }
  },
  watch: {
    contestId: {
      handler(newVal, oldVal) {
        if (newVal && newVal !== oldVal) {
          this.loadData()
        }
      },
      immediate: true
    }
  }
}
</script>

<style scoped>
.plagiarism-container {
  padding: 20px;
}

.page-error {
  margin-bottom: 20px;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 100px 0;
  font-size: 16px;
  color: #909399;
}

.loading-container i {
  margin-right: 10px;
  font-size: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-card,
.check-card,
.result-card {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
}

.stat-value.danger {
  color: #F56C6C;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.code-viewer {
  max-height: 600px;
  overflow: auto;
  background: #282c34;  /* atom-one-dark 主题背景色 */
  padding: 20px;
  border-radius: 0;
  text-align: left;
}

.code-viewer pre {
  margin: 0;
  white-space: pre;
  word-wrap: normal;
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  background: transparent;
  color: #abb2bf;
}

.code-viewer code {
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  display: block;
  white-space: pre;
  background: transparent !important;
  padding: 0 !important;
}

/* 代码对比样式 */
.code-compare-container {
  display: flex;
  gap: 20px;
  height: 600px;
}

.code-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.code-panel-header {
  background: #f5f7fa;
  padding: 12px 16px;
  border-bottom: 1px solid #dcdfe6;
  font-weight: bold;
  color: #303133;
}

.code-panel .code-viewer {
  flex: 1;
  overflow: auto;
  background: #282c34;  /* 与 atom-one-dark 主题一致 */
  margin: 0;
  border: none;
  border-radius: 0;
}

.submission-info p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.submission-info p strong {
  color: #303133;
  margin-right: 8px;
}
</style>

<style>
/* Highlight.js 代码高亮全局样式 */
.code-viewer .hljs {
  display: block;
  overflow-x: auto;
  padding: 0;
  background: #282c34;
  color: #abb2bf;
}

.code-viewer .hljs-comment,
.code-viewer .hljs-quote {
  color: #5c6370;
  font-style: italic;
}

.code-viewer .hljs-keyword,
.code-viewer .hljs-selector-tag,
.code-viewer .hljs-subst {
  color: #c678dd;
}

.code-viewer .hljs-number,
.code-viewer .hljs-literal,
.code-viewer .hljs-variable,
.code-viewer .hljs-template-variable,
.code-viewer .hljs-tag .hljs-attr {
  color: #d19a66;
}

.code-viewer .hljs-string,
.code-viewer .hljs-doctag {
  color: #98c379;
}

.code-viewer .hljs-title,
.code-viewer .hljs-section,
.code-viewer .hljs-selector-id {
  color: #61afef;
}

.code-viewer .hljs-type,
.code-viewer .hljs-class .hljs-title {
  color: #e5c07b;
}

.code-viewer .hljs-tag,
.code-viewer .hljs-name,
.code-viewer .hljs-attribute {
  color: #e06c75;
  font-weight: normal;
}

.code-viewer .hljs-regexp,
.code-viewer .hljs-link {
  color: #56b6c2;
}

.code-viewer .hljs-symbol,
.code-viewer .hljs-bullet {
  color: #61afef;
}

.code-viewer .hljs-built_in,
.code-viewer .hljs-builtin-name {
  color: #e6c07b;
}

.code-viewer .hljs-meta {
  color: #61afef;
}

.code-viewer .hljs-deletion {
  background: #f8756f;
}

.code-viewer .hljs-addition {
  background: #98c379;
}

.code-viewer .hljs-emphasis {
  font-style: italic;
}

.code-viewer .hljs-strong {
  font-weight: bold;
}
</style>
