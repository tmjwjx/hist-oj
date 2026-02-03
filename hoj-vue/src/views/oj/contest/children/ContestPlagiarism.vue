<template>
  <div class="plagiarism-container">
    <!-- 加载中 -->
    <div v-if="loading" class="loading-container">
      <i class="el-icon-loading"></i>
      <span>加载中...</span>
    </div>

    <!-- 比赛未结束提示 -->
    <el-alert
      v-else-if="!isContestEnded"
      title="比赛未结束"
      type="warning"
      description="比赛未结束，不允许进行查重。请等待比赛结束后再操作。"
      show-icon
      :closable="false"
      style="margin-bottom: 20px"
    />

    <!-- 比赛已结束，显示查重功能 -->
    <div v-else>
      <!-- 步骤1: 设置查重率 -->
      <el-card class="config-card" v-if="currentStep === 'config'">
        <div slot="header" class="card-header">
          <span>步骤1: 设置查重率阈值</span>
          <el-button type="primary" size="small" @click="saveConfig" :loading="saving" :disabled="isCheckRunning">
            保存配置
          </el-button>
        </div>

        <el-alert
          title="查重说明"
          type="info"
          description="查重仅对比比赛期间通过评测（AC）的代码，非AC提交不会被比较。"
          show-icon
          :closable="false"
          style="margin-bottom: 20px"
        />

        <el-table :data="problems" style="width: 100%">
          <el-table-column prop="displayId" label="题号" width="100" />
          <el-table-column prop="displayTitle" label="题目名称" />
          <el-table-column label="查重率阈值 (%)" width="250">
            <template slot-scope="scope">
              <el-input-number
                v-model="scope.row.threshold"
                :min="0"
                :max="100"
                :step="5"
                size="small"
              />
              <span style="margin-left: 10px; color: #909399">
                超过此值将被记录
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 步骤2: 开始查重 -->
      <el-card class="check-card" v-if="currentStep === 'check' || currentStep === 'result'">
        <div slot="header" class="card-header">
          <span>步骤2: 执行查重</span>
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
                {{ checkStatus && (checkStatus.status === 'completed' || checkStatus.status === 'failed') ? '重新查重' : '开始查重' }}
              </el-button>
              <!-- 查重进行中时，禁用按钮并显示状态 -->
              <el-button
                v-if="checkStatus && checkStatus.status === 'running'"
                type="info"
                size="small"
                :loading="true"
                disabled
              >
                查重中 {{ checkStatus.progress }}%
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
                重新查重
              </el-button>
              <el-button
                type="success"
                size="small"
                @click="resetConfig"
              >
                重新设置
              </el-button>
            </template>
          </div>
        </div>

        <!-- 查重进度 -->
        <div v-if="checkStatus">
          <el-row :gutter="20" style="margin-bottom: 20px">
            <el-col :span="6">
              <statistic-card title="状态" :value="getStatusText(checkStatus.status)" />
            </el-col>
            <el-col :span="6">
              <statistic-card title="总提交数" :value="checkStatus.totalSubmissions" />
            </el-col>
            <el-col :span="6">
              <statistic-card title="已对比" :value="`${checkStatus.checkedPairs}/${checkStatus.totalPairs}`" />
            </el-col>
            <el-col :span="6">
              <statistic-card title="进度" :value="`${checkStatus.progress}%`" />
            </el-col>
          </el-row>

          <el-progress
            :percentage="checkStatus.progress"
            :status="checkStatus.status === 'completed' ? 'success' : ''"
            :stroke-width="20"
          />

          <el-alert
            v-if="checkStatus.status === 'running'"
            title="查重进行中，请稍候..."
            type="info"
            :closable="false"
            style="margin-top: 20px"
          />
        </div>
      </el-card>

      <!-- 步骤3: 查看结果 -->
      <el-card class="result-card" v-if="currentStep === 'result' && checkStatus && checkStatus.status === 'completed'">
        <div slot="header" class="card-header">
          <span>步骤3: 查重结果</span>
          <div>
            <el-input
              v-model="displayIdFilter"
              placeholder="输入题号筛选（如：A、B、C）"
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
              {{ exporting ? '正在导出...' : '导出Excel' }}
            </el-button>
          </div>
        </div>

        <!-- 提示：只显示超过阈值的结果 -->
        <el-alert
          title="数据说明"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        >
          <template slot="default">
            <div>当前页面仅显示<strong>超过阈值</strong>的查重结果（共 {{ totalCount }} 条）。</div>
            <div style="margin-top: 8px">
              如需查看<strong>所有查重结果</strong>（包括未超过阈值的），请点击右上角的「导出Excel」按钮下载完整数据。
            </div>
          </template>
        </el-alert>

        <!-- 超过 500 条提示 -->
        <el-alert
          v-if="totalCount > 500"
          title="数据量提示"
          type="warning"
          :closable="false"
          style="margin-bottom: 20px"
        >
          当前筛选结果有 {{ totalCount }} 条记录，超过 500 条。建议导出 Excel 查看完整数据。
        </el-alert>

        <!-- 结果统计 -->
        <el-row :gutter="20" style="margin-bottom: 20px">
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ totalCount }}</div>
                <div class="stat-label">超过阈值结果数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover" type="danger">
              <div class="stat-item">
                <div class="stat-value danger">{{ suspiciousUsers }} 人</div>
                <div class="stat-label">涉及可疑用户</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ results.length }}</div>
                <div class="stat-label">当前显示</div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 结果列表 -->
        <el-table :data="pagedResults" style="width: 100%" stripe v-loading="loading">
          <el-table-column prop="displayId" label="题号" width="80" />
          <el-table-column prop="problemTitle" label="题目名称" width="200" />
          <el-table-column label="用户对比" width="250">
            <template slot-scope="scope">
              <el-tag type="info" size="small">{{ scope.row.username1 }}</el-tag>
              <span style="margin: 0 10px">vs</span>
              <el-tag type="info" size="small">{{ scope.row.username2 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="提交ID" width="180">
            <template slot-scope="scope">
              <div style="font-size: 12px">
                <div>A: {{ scope.row.contestRecordId1 || scope.row.submitId1 }}</div>
                <div>B: {{ scope.row.contestRecordId2 || scope.row.submitId2 }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="相似度" width="180">
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
          <el-table-column label="编程语言" width="120">
            <template slot-scope="scope">
              {{ scope.row.language }}
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="150">
            <template slot-scope="scope">
              <el-button
                size="mini"
                type="primary"
                icon="el-icon-view"
                :loading="loadingCode"
                :disabled="loadingCode"
                @click="viewCode(scope.row)"
              >
                对比代码
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
      title="代码对比"
      :visible.sync="codeDialogVisible"
      width="90%"
      :close-on-click-modal="false"
    >
      <!-- 提交信息 -->
      <el-row :gutter="20" style="margin-bottom: 20px">
        <el-col :span="12">
          <el-card shadow="hover">
            <div slot="header">
              <span style="font-weight: bold">用户A: {{ codeData.user1?.username }}</span>
            </div>
            <div class="submission-info">
              <p><strong>题号：</strong>{{ codeData.displayId }}</p>
              <p><strong>题目：</strong>{{ codeData.problemTitle }}</p>
              <p><strong>提交时间：</strong>{{ codeData.submitTime1 }}</p>
              <p><strong>语言：</strong>{{ codeData.language }}</p>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover">
            <div slot="header">
              <span style="font-weight: bold">用户B: {{ codeData.user2?.username }}</span>
            </div>
            <div class="submission-info">
              <p><strong>题号：</strong>{{ codeData.displayId }}</p>
              <p><strong>题目：</strong>{{ codeData.problemTitle }}</p>
              <p><strong>提交时间：</strong>{{ codeData.submitTime2 }}</p>
              <p><strong>语言：</strong>{{ codeData.language }}</p>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 代码对比区域 -->
      <div class="code-compare-container">
        <div class="code-panel">
          <div class="code-panel-header">
            <span>{{ codeData.user1?.username }} 的代码</span>
          </div>
          <div class="code-viewer" v-if="codeData.code1">
            <pre><code :key="`code1-${codeDialogVisible}`" ref="codeBlock1" :class="`language-${mapLanguage(codeData.language1)}`">{{ codeData.code1 }}</code></pre>
          </div>
        </div>
        <div class="code-panel">
          <div class="code-panel-header">
            <span>{{ codeData.user2?.username }} 的代码</span>
          </div>
          <div class="code-viewer" v-if="codeData.code2">
            <pre><code :key="`code2-${codeDialogVisible}`" ref="codeBlock2" :class="`language-${mapLanguage(codeData.language2)}`">{{ codeData.code2 }}</code></pre>
          </div>
        </div>
      </div>

      <div slot="footer" class="dialog-footer">
        <el-button @click="codeDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyCode(1)">复制A代码</el-button>
        <el-button type="primary" @click="copyCode(2)">复制B代码</el-button>
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
      totalCount: 0, // 超过阈值的结果总数
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
      loadingCode: false // 查看代码的加载状态
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
  mounted() {
    // 如果 contest 已经加载，直接加载数据
    if (this.isContestLoaded) {
      this.loadData()
    }
  },
  beforeDestroy() {
    if (this.progressTimer) {
      clearInterval(this.progressTimer)
    }
  },
  methods: {
    ...mapActions(['getContestProblems']),
    async loadData() {
      // 如果比赛未结束，设置 loading 为 false 并返回
      if (!this.isContestEnded) {
        this.loading = false
        return
      }

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
        // 如果是 404 错误，说明权限不足，静默处理（不显示错误消息）
        // 其他错误也静默处理，避免干扰用户体验
        // 可以选择性地显示提示
        // this.$message.warning('您没有查重权限')
        console.error('加载数据失败:', error)
      } finally {
        this.loading = false
      }
    },

    async saveConfig() {
      this.saving = true
      try {
        const configs = this.problems.map(p => ({
          cpid: p.cpid,
          threshold: p.threshold
        }))

        await api.savePlagiarismConfig(this.contest.id, configs)
        this.$message.success('配置保存成功')

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
        this.$message.error('配置保存失败: ' + error.message)
      } finally {
        this.saving = false
      }
    },

    async startCheck() {
      // 防止重复点击
      if (this.starting) return

      // 如果已经有正在运行的查重任务，提示用户
      if (this.checkStatus && this.checkStatus.status === 'running') {
        this.$message.warning('查重任务正在运行中，请稍候...')
        return
      }

      this.starting = true
      try {
        const res = await api.startPlagiarismCheck(this.contest.id)
        this.$message.success('查重任务已启动')

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

        // 立即轮询一次，避免等待2秒才显示进度
        await this.pollProgress()

        // 启动定时轮询
        this.startPolling()
      } catch (error) {
        this.$message.error('启动查重失败: ' + (error.response?.data?.message || error.message))
      } finally {
        this.starting = false
      }
    },

    async restartCheck() {
      // 防止重复点击
      if (this.starting) return

      this.$confirm('重新查重将删除当前任务和结果，确定要继续吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        // 停止当前的轮询
        if (this.progressTimer) {
          clearInterval(this.progressTimer)
          this.progressTimer = null
        }

        this.starting = true
        try {
          const res = await api.startPlagiarismCheck(this.contest.id)
          this.$message.success('查重任务已重新启动')

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

          // 立即轮询一次获取真实进度
          await this.pollProgress()

          // 启动定时轮询
          this.startPolling()
        } catch (error) {
          this.$message.error('重新查重失败: ' + (error.response?.data?.message || error.message))
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

      // 立即执行一次
      this.pollProgress()

      // 启动定时轮询
      this.progressTimer = setInterval(() => {
        this.pollProgress()
      }, 2000) // 每2秒轮询一次
    },

    async pollProgress() {
      try {
        const res = await api.getPlagiarismProgress(this.contest.id)

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
          this.$message.success('查重完成')
        } else if (this.checkStatus.status === 'failed') {
          // 清除定时器
          if (this.progressTimer) {
            clearInterval(this.progressTimer)
            this.progressTimer = null
          }
          this.$message.error('查重失败: ' + (this.checkStatus.errorMessage || '未知错误'))
        }
        // 如果是 running 状态，继续轮询（由定时器处理）
      } catch (error) {
        console.error('获取进度失败:', error)
        // 如果 checkStatus 不存在或状态不明确，不清除定时器，继续尝试获取进度
        if (!this.checkStatus || this.checkStatus.status !== 'running') {
          console.warn('checkStatus 状态异常，停止轮询')
          if (this.progressTimer) {
            clearInterval(this.progressTimer)
            this.progressTimer = null
          }
        }
      }
    },

    async loadResults() {
      if (!this.checkStatus) return

      this.loading = true
      try {
        const res = await api.getPlagiarismResults(this.checkStatus.id, this.displayIdFilter)
        this.results = res.data.data || []
        this.totalCount = res.data.totalCount || 0
        // 重置到第一页
        this.currentPage = 1
      } catch (error) {
        console.error('加载结果失败:', error)
        this.$message.error('加载查重结果失败: ' + error.message)
      } finally {
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
        this.$message.error('获取代码失败: ' + error.message)
      } finally {
        this.loadingCode = false
      }
    },

    copyCode(userNum) {
      const code = userNum === 1 ? this.codeData.code1 : this.codeData.code2
      const user = userNum === 1 ? this.codeData.user1?.username : this.codeData.user2?.username

      navigator.clipboard.writeText(code).then(() => {
        this.$message.success(`${user} 的代码已复制到剪贴板`)
      })
    },

    async exportExcel() {
      // 设置导出状态
      this.exporting = true

      // 显示提示
      const loadingMessage = this.$message({
        message: '正在生成Excel文件，数据量大时可能需要较长时间，请耐心等待...',
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
        link.download = `查重结果_${this.contest.title}_${Date.now()}.csv`
        link.click()

        // 关闭加载提示
        loadingMessage.close()

        this.$message.success('导出成功！')
      } catch (error) {
        // 关闭加载提示
        loadingMessage.close()

        this.$message.error('导出失败: ' + (error.message || '未知错误'))
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
        pending: '待开始',
        running: '进行中',
        completed: '已完成',
        failed: '失败'
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
    contest: {
      handler(newVal) {
        // 当 contest 数据加载完成后
        if (newVal && newVal.id !== undefined) {
          // 不要在这里设置 loading = false，让 loadData() 自己管理
          this.loadData()
        }
      },
      deep: true,
      immediate: true  // 立即执行一次，确保 contest 已存在时也能触发
    }
  }
}
</script>

<style scoped>
.plagiarism-container {
  padding: 20px;
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
