<template>
  <div class="judge-terminal-page">
    <el-row :gutter="20">
        <!-- 左侧栏 -->
        <el-col :span="10">
          <h3 class="section-title">
            <i class="el-icon-monitor"></i> BingoJ 判题终端
          </h3>

          <!-- 配置区 -->
          <el-card class="config-card" shadow="hover">
            <div slot="header" class="card-header">
              <i class="el-icon-setting"></i> 配置
            </div>
            <el-form :model="form" size="small" label-width="80px">
              <el-alert
                title="系统说明"
                type="success"
                :closable="false"
                style="margin-bottom: 15px; padding: 8px 12px;">
                已自动使用 BingOJ 账号登录，无需手动输入
              </el-alert>
              <el-form-item label="模式">
                <el-radio-group v-model="form.mode" size="small" disabled>
                  <el-radio label="normal">普通模式</el-radio>
                </el-radio-group>
                <el-button type="primary" size="mini" style="margin-left: 10px" @click="fetchProblemInfo">
                  获取题目
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>

          <!-- 代码编辑器 -->
          <el-card class="code-card" shadow="hover">
            <div slot="header" class="card-header">
              <i class="el-icon-edit"></i> 代码编辑器
            </div>
            <el-form :model="form" size="small">
              <el-row :gutter="10">
                <el-col :span="12">
                  <el-input v-model="form.pid" placeholder="题目ID (如 0001)" size="small"></el-input>
                </el-col>
                <el-col :span="12">
                  <el-select v-model="form.language" placeholder="选择语言" size="small" style="width: 100%">
                    <el-option label="C++ 17" value="C++ 17 With O2"></el-option>
                    <el-option label="C" value="C With O2"></el-option>
                    <el-option label="Python3" value="Python3"></el-option>
                    <el-option label="Java" value="Java"></el-option>
                  </el-select>
                </el-col>
              </el-row>
              <el-input
                type="textarea"
                v-model="form.code"
                :rows="12"
                placeholder="// 在此粘贴代码..."
                style="margin-top: 10px; font-family: 'Consolas', monospace; font-size: 13px"
              ></el-input>
              <el-button
                type="primary"
                :loading="isRunning"
                @click="runCode"
                style="width: 100%; margin-top: 10px"
              >
                <i class="el-icon-video-play"></i> {{ isRunning ? '运行中...' : '运行自测并提交' }}
              </el-button>
            </el-form>
          </el-card>

          <!-- 日志区 -->
          <el-card class="log-card" shadow="hover">
            <div slot="header" class="card-header">
              <i class="el-icon-document"></i> 系统日志
            </div>
            <div class="log-area" ref="logArea">
              <div v-for="(log, index) in logs" :key="index" class="log-line">
                <span class="log-time">[{{ log.time }}]</span> {{ log.msg }}
              </div>
            </div>
          </el-card>

          <!-- 历史记录 -->
          <el-card class="history-card" shadow="hover">
            <div slot="header" class="card-header">
              <i class="el-icon-time"></i> 本地记录
            </div>
            <el-table :data="historyList" size="small" stripe style="width: 100%">
              <el-table-column prop="username" label="用户" width="80"></el-table-column>
              <el-table-column label="远程结果" width="100">
                <template slot-scope="scope">
                  <el-tag
                    :type="getResultType(scope.row.result)"
                    size="mini"
                    class="clickable-result-tag"
                    @click.native="showHistoryCaseDetails(scope.row)">
                    {{ scope.row.result }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="本地自测" width="100">
                <template slot-scope="scope">
                  <el-tag :type="getLocalResultType(scope.row.local_info)" size="mini">
                    {{ scope.row.local_info || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="耗时/内存" width="100">
                <template slot-scope="scope">
                  <span style="font-size: 11px; color: #909399">
                    {{ scope.row.time_used || '--' }} / {{ scope.row.memory_used || '--' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="language" label="语言" width="60">
                <template slot-scope="scope">
                  {{ simplifyLanguage(scope.row.language) }}
                </template>
              </el-table-column>
              <el-table-column label="提交时间" width="120">
                <template slot-scope="scope">
                  {{ formatTime(scope.row.submit_time) }}
                </template>
              </el-table-column>
              <el-table-column label="代码" width="60" align="center">
                <template slot-scope="scope">
                  <el-button type="text" size="mini" @click="showCode(scope.row)">
                    <i class="el-icon-view"></i>
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <div style="margin-top: 10px; text-align: center">
              <el-pagination
                @current-change="handleHistoryPageChange"
                :current-page="historyPagination.currentPage"
                :page-size="historyPagination.pageSize"
                :total="historyPagination.total"
                layout="prev, pager, next, total"
                small
              >
              </el-pagination>
            </div>
          </el-card>
        </el-col>

        <!-- 右侧栏 -->
        <el-col :span="14">
          <!-- 题目详情 -->
          <el-card class="problem-card" shadow="hover">
            <div slot="header" class="card-header">
              <span><i class="el-icon-document"></i> 题目详情</span>
              <el-tag v-if="problemInfo.displayId" type="primary" size="small">
                {{ problemInfo.displayId }}
              </el-tag>
            </div>
            <div class="problem-content" v-if="problemInfo.problem">
              <h3 class="problem-title">{{ problemInfo.displayId }} - {{ problemInfo.problem.title }}</h3>
              <!-- 判题模式显示 -->
              <div class="problem-info-bar">
                <el-tag size="small" type="primary">
                  判题模式: {{ getJudgeModeText(problemInfo.problem.judgeMode) }}
                </el-tag>
                <el-tag size="small" type="primary" style="margin-left: 10px">
                  时间限制: {{ problemInfo.problem.timeLimit }}ms
                </el-tag>
                <el-tag size="small" type="primary" style="margin-left: 10px">
                  内存限制: {{ problemInfo.problem.memoryLimit }}MB
                </el-tag>
              </div>
              <div v-if="problemInfo.problem.description" class="problem-section">
                <h4><i class="el-icon-tickets"></i> 描述</h4>
                <div v-html="renderMarkdown(problemInfo.problem.description)"></div>
              </div>
              <div v-if="problemInfo.problem.input" class="problem-section">
                <h4><i class="el-icon-download"></i> 输入</h4>
                <div v-html="renderMarkdown(problemInfo.problem.input)"></div>
              </div>
              <div v-if="problemInfo.problem.output" class="problem-section">
                <h4><i class="el-icon-upload2"></i> 输出</h4>
                <div v-html="renderMarkdown(problemInfo.problem.output)"></div>
              </div>
              <div v-if="problemInfo.problem.hint" class="problem-section">
                <h4><i class="el-icon-info"></i> 提示</h4>
                <div v-html="renderMarkdown(problemInfo.problem.hint)"></div>
              </div>
              <div v-if="examples.length > 0" class="problem-section">
                <h4><i class="el-icon-document-copy"></i> 样例</h4>
                <el-row :gutter="10" v-for="(example, index) in examples" :key="index" style="margin-bottom: 10px">
                  <el-col :span="12">
                    <div class="example-box">
                      <div class="example-title">样例 {{ index + 1 }} 输入:</div>
                      <pre>{{ example.input }}</pre>
                    </div>
                  </el-col>
                  <el-col :span="12">
                    <div class="example-box">
                      <div class="example-title">样例 {{ index + 1 }} 输出:</div>
                      <pre>{{ example.output }}</pre>
                    </div>
                  </el-col>
                </el-row>
              </div>
            </div>
            <div v-else class="empty-state">
              <i class="el-icon-arrow-left"></i> 请先在左侧输入ID并获取题目
            </div>
          </el-card>

          <!-- 结果面板 -->
          <el-card v-if="showResult" class="result-card" shadow="hover">
            <div slot="header" class="card-header">
              <i class="el-icon-data-analysis"></i> 判题结果报告
            </div>

            <!-- 远程结果 Banner -->
            <div :class="['result-banner', remoteBannerClass]">
              <i :class="remoteBannerIcon"></i> {{ remoteBannerText }}
            </div>

            <!-- 远程判题错误信息 -->
            <div v-if="remoteErrorMessage" class="remote-error-section">
              <div class="error-header">
                <i class="el-icon-warning" style="color: #F56C6C;"></i>
                <strong>错误详情:</strong>
              </div>
              <pre class="remote-error-block">{{ remoteErrorMessage }}</pre>
            </div>

            <!-- 样例测试详情 -->
            <div class="sample-section">
              <div class="sample-header">
                <h4>本地样例自测详情</h4>
                <el-tag size="small">{{ sampleSummary }}</el-tag>
              </div>
              <div class="sample-list">
                <div v-for="sample in sampleResults" :key="sample.id" class="sample-item">
                  <div class="sample-item-header" @click="toggleSample(sample.id)">
                    <span>样例 {{ sample.id }}</span>
                    <span>
                      <i :class="sample.is_ok ? 'el-icon-success' : 'el-icon-error'"
                         :style="{color: sample.is_ok ? '#67C23A' : '#F56C6C'}"></i>
                      {{ sample.is_ok ? '通过' : '失败' }}
                    </span>
                  </div>
                  <div v-show="expandedSamples[sample.id]" class="sample-item-detail">
                    <div><strong>输入:</strong><pre class="code-block">{{ sample.input }}</pre></div>
                    <div><strong>预期:</strong><pre class="code-block">{{ sample.expected }}</pre></div>
                    <div><strong>输出:</strong><pre class="code-block">{{ sample.output }}</pre></div>
                    <!-- 错误信息（当发生RE等错误时显示） -->
                    <div v-if="sample.stderr" class="error-section">
                      <strong style="color: #F56C6C;">错误信息:</strong>
                      <pre class="error-block">{{ sample.stderr }}</pre>
                    </div>

                    <!-- ✅ 新增：详细错误信息（来自 go-judge，包含 testlib 错误） -->
                    <div v-if="sample.detailed_stderr && sample.detailed_stderr !== sample.stderr" class="detailed-error-section">
                      <div class="error-header">
                        <i class="el-icon-info" style="color: #409EFF;"></i>
                        <strong style="color: #409EFF;">详细错误 (go-judge):</strong>
                      </div>
                      <pre class="detailed-error-block">{{ sample.detailed_stderr }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 远程测试点详情 -->
            <div class="case-section">
              <div class="sample-header">
                <h4>远程测试点详情</h4>
                <el-tag size="small" type="info">
                  {{ remoteSubmitId ? '提交ID: ' + remoteSubmitId : '提交ID: --' }}
                </el-tag>
              </div>

              <el-alert
                v-if="isCaseDetailsLoading"
                title="正在加载测试点详情..."
                type="info"
                :closable="false"
                style="margin-bottom: 10px">
              </el-alert>

              <el-table
                v-if="remoteCaseDetails.length > 0"
                :data="remoteCaseDetails"
                size="mini"
                stripe
                border
                style="width: 100%">
                <el-table-column label="#" width="60" align="center">
                  <template slot-scope="scope">
                    {{ scope.row.seq || scope.$index + 1 }}
                  </template>
                </el-table-column>
                <el-table-column prop="case_id" label="CaseID" width="90" align="center"></el-table-column>
                <el-table-column label="结果" min-width="120" align="center">
                  <template slot-scope="scope">
                    <el-tag :type="getCaseStatusTagType(scope.row.status)" size="mini">
                      {{ getCaseStatusText(scope.row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="时间" width="110" align="center">
                  <template slot-scope="scope">
                    {{ formatCaseTime(scope.row.time) }}
                  </template>
                </el-table-column>
                <el-table-column label="内存" width="120" align="center">
                  <template slot-scope="scope">
                    {{ formatCaseMemory(scope.row.memory) }}
                  </template>
                </el-table-column>
                <el-table-column label="分数" width="80" align="center">
                  <template slot-scope="scope">
                    {{ scope.row.score === null || scope.row.score === undefined ? '--' : scope.row.score }}
                  </template>
                </el-table-column>
                <el-table-column label="分组" width="80" align="center">
                  <template slot-scope="scope">
                    {{ scope.row.group_num === null || scope.row.group_num === undefined ? '--' : scope.row.group_num }}
                  </template>
                </el-table-column>
              </el-table>

              <el-empty
                v-else-if="!isRunning"
                :image-size="60"
                description="暂无测试点详情">
              </el-empty>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 代码预览对话框 -->
      <el-dialog
        title="提交代码"
        :visible.sync="codeDialogVisible"
        width="60%"
        @opened="addCodeLineNumbers"
      >
        <div class="code-display-wrapper">
          <Highlight
            :key="codeDialogVisible"
            :code="currentCode"
            :language="mapLanguage(currentLanguage)"
            :classroom-mode="true"
          ></Highlight>
        </div>
      </el-dialog>

      <!-- 历史提交测试点详情 -->
      <el-dialog
        title="历史提交测试点详情"
        :visible.sync="historyCaseDialogVisible"
        width="62%">
        <div class="history-case-header">
          <el-tag size="small" type="info">提交ID: {{ historyCaseSubmitId || '--' }}</el-tag>
          <el-tag size="small">{{ historyCaseResult || '--' }}</el-tag>
        </div>

        <el-alert
          v-if="historyCaseLoading"
          title="正在加载测试点详情..."
          type="info"
          :closable="false"
          style="margin-bottom: 12px">
        </el-alert>

        <el-table
          v-if="historyCaseDetails.length > 0"
          :data="historyCaseDetails"
          size="mini"
          stripe
          border
          style="width: 100%">
          <el-table-column label="#" width="60" align="center">
            <template slot-scope="scope">
              {{ scope.row.seq || scope.$index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="case_id" label="CaseID" width="90" align="center"></el-table-column>
          <el-table-column label="结果" min-width="120" align="center">
            <template slot-scope="scope">
              <el-tag :type="getCaseStatusTagType(scope.row.status)" size="mini">
                {{ getCaseStatusText(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="时间" width="110" align="center">
            <template slot-scope="scope">
              {{ formatCaseTime(scope.row.time) }}
            </template>
          </el-table-column>
          <el-table-column label="内存" width="120" align="center">
            <template slot-scope="scope">
              {{ formatCaseMemory(scope.row.memory) }}
            </template>
          </el-table-column>
          <el-table-column label="分数" width="80" align="center">
            <template slot-scope="scope">
              {{ scope.row.score === null || scope.row.score === undefined ? '--' : scope.row.score }}
            </template>
          </el-table-column>
          <el-table-column label="分组" width="80" align="center">
            <template slot-scope="scope">
              {{ scope.row.group_num === null || scope.row.group_num === undefined ? '--' : scope.row.group_num }}
            </template>
          </el-table-column>
        </el-table>

        <el-empty
          v-else-if="!historyCaseLoading"
          :image-size="56"
          description="该提交暂无测试点详情">
        </el-empty>
      </el-dialog>
    </el-row>
  </div>
</template>

<script>
import { getJudgeInfo, getJudgeHistory, getJudgeCaseDetails, runCombinedJudge } from '@/common/judgeTerminal'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'
const Highlight = () => import('@/components/oj/common/Highlight')
import { addCodeBtn } from '@/common/codeblock'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt()
md.use(MarkdownItKatex)

export default {
  name: 'JudgeTerminalPage',
  components: {
    Highlight
  },
  data() {
    return {
      form: {
        username: '',
        password: '',
        mode: 'normal',
        pid: '',
        language: 'C++ 17 With O2',
        code: ''
      },
      problemInfo: {
        displayId: '',
        problem: null,
        history: []
      },
      examples: [],
      logs: [],
      historyList: [],
      historyPagination: {
        currentPage: 1,
        pageSize: 10,
        total: 0
      },
      isRunning: false,
      showResult: false,
      remoteBannerText: '等待开始...',
      remoteBannerClass: 'bg-pending',
      remoteBannerIcon: 'el-icon-loading',
      remoteErrorMessage: '', // 新增：远程判题错误信息
      remoteSubmitId: '',
      remoteCaseDetails: [],
      isCaseDetailsLoading: false,
      sampleResults: [],
      sampleSummary: '等待中',
      expandedSamples: {},
      codeDialogVisible: false,
      historyCaseDialogVisible: false,
      historyCaseLoading: false,
      historyCaseDetails: [],
      historyCaseSubmitId: '',
      historyCaseResult: '',
      currentCode: '',
      currentLanguage: ''
    }
  },
  mounted() {
    // 使用 $nextTick 确保 DOM 完全加载后再自动填充
    this.$nextTick(() => {
      this.autoFillCredentials()
    })
  },
  methods: {
    // 自动从 BingOJ 填充用户凭据
    autoFillCredentials() {
      this.addLog('正在检查 BingOJ 登录状态...')

      try {
        // 尝试从 BingOJ 获取 token
        const token = localStorage.getItem('token')
        if (!token) {
          this.addLog('未检测到 BingOJ 登录信息，请先登录 BingOJ')
          return
        }

        this.addLog('已检测到 BingOJ Token')

        // 尝试从 BingOJ 获取用户信息
        const userInfoStr = localStorage.getItem('userInfo')
        if (!userInfoStr) {
          this.addLog('未检测到 BingOJ 用户信息，请先登录 BingOJ')
          return
        }

        this.addLog('已检测到 BingOJ 用户信息')

        const userInfo = JSON.parse(userInfoStr)
        if (userInfo.username) {
          this.form.username = userInfo.username
          // Token 会在 API 请求时添加到请求体

          this.addLog(`✅ 自动登录成功！`)
          this.addLog(`用户名: ${userInfo.username}`)
          this.addLog(`认证方式: BingOJ Token (通过请求体发送)`)
          this.$message.success(`已自动登录 BingOJ 用户: ${userInfo.username}`)
        } else {
          this.addLog('用户信息中没有 username 字段')
        }
      } catch (error) {
        this.addLog(`自动登录失败: ${error.message}`)
        this.$message.warning('自动登录失败，请手动输入')
      }
    },

    // 保存用户凭据到 localStorage
    saveUserCredentials() {
      try {
        const credentials = {
          username: this.form.username,
          password: this.form.password
        }
        localStorage.setItem('judge_terminal_credentials', JSON.stringify(credentials))
      } catch (error) {
        this.$message.warning('凭据保存失败,请检查浏览器设置')
      }
    },

    // 获取题目信息
    async fetchProblemInfo() {
      if (!this.form.pid) {
        this.$message.warning('请输入题目ID')
        return
      }

      // 检查是否已自动填充用户名（token 会在 API 请求时添加到请求体）
      if (!this.form.username) {
        this.$message.warning('未检测到 BingOJ 登录信息，请先登录 BingOJ')
        this.addLog('错误: 未检测到 BingOJ 登录信息')
        return
      }

      this.addLog('正在获取题目信息...')

      try {
        const res = await getJudgeInfo({
          pid: this.form.pid,
          cid: '0',
          mode: 'normal',
          username: this.form.username,
          password: this.form.password
        })

        if (res.code === 200) {
          this.saveUserCredentials()
          this.problemInfo = res.data
          this.historyPagination.currentPage = 1
          await this.fetchHistory()
          this.extractExamples()
          this.addLog('获取题目成功')
          this.$message.success('获取题目成功')
        } else {
          this.addLog(`错误: ${res.message}`)
          this.$message.error(res.message)
        }
      } catch (error) {
        this.addLog(`网络错误: ${error.message}`)
        this.$message.error('网络错误')
      }
    },

    // 提取样例
    extractExamples() {
      if (!this.problemInfo.problem || !this.problemInfo.problem.examples) {
        this.examples = []
        return
      }

      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      const examples = []
      let match

      while ((match = regex.exec(this.problemInfo.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      this.examples = examples
    },

    // 运行代码
    runCode() {
      if (!this.form.code) {
        this.$message.warning('代码不能为空')
        return
      }

      if (!this.form.pid) {
        this.$message.warning('请先获取题目')
        return
      }

      // 检查是否已自动填充用户名（token 会在 API 请求时添加到请求体）
      if (!this.form.username) {
        this.$message.warning('未检测到 BingOJ 登录信息，请先登录 BingOJ')
        this.addLog('错误: 未检测到 BingOJ 登录信息')
        return
      }

      this.isRunning = true
      this.showResult = true
      this.logs = []
      this.sampleResults = []
      this.remoteErrorMessage = '' // 清空远程错误信息
      this.remoteSubmitId = ''
      this.remoteCaseDetails = []
      this.isCaseDetailsLoading = false
      this.remoteBannerText = '本地测试中...'
      this.remoteBannerClass = 'bg-pending'
      this.remoteBannerIcon = 'el-icon-loading'

      this.$nextTick(() => {
        const resultCard = document.querySelector('.result-card')
        if (resultCard) {
          resultCard.scrollIntoView({ behavior: 'smooth' })
        }
      })

      runCombinedJudge(
        {
          pid: this.form.pid,
          cid: this.form.cid || '0',
          mode: this.form.mode,
          username: this.form.username,
          // Token 会在 API 层添加到请求体
          language: this.form.language,
          code: this.form.code
        },
        this.handleSSEMessage,
        this.handleSSEError,
        this.handleSSEComplete
      )
    },

    // 处理 SSE 消息
    handleSSEMessage(data) {
      switch (data.type) {
        case 'log':
          this.addLog(data.msg)
          if (!this.remoteSubmitId && typeof data.msg === 'string') {
            const matched = data.msg.match(/提交ID:\s*(\d+)/)
            if (matched && matched[1]) {
              this.remoteSubmitId = matched[1]
              this.fetchCaseDetails(this.remoteSubmitId)
            }
          }
          break
        case 'sample_res':
          this.sampleResults.push(data.data)
          this.updateSampleSummary()
          break
        case 'compile_error':
          this.addLog('编译错误')
          this.remoteBannerText = '本地编译错误'
          this.remoteBannerClass = 'bg-error'
          this.remoteBannerIcon = 'el-icon-close'
          break
        case 'remote_status':
          this.updateRemoteStatus(data.data)
          if (data.data && data.data.submit_id) {
            this.remoteSubmitId = data.data.submit_id
          }
          if (this.isFinalRemoteStatus(data.data.status) && this.remoteSubmitId) {
            this.fetchCaseDetails(this.remoteSubmitId)
          }
          break
        case 'remote_submit':
          if (data.data && data.data.submit_id) {
            this.remoteSubmitId = data.data.submit_id
            this.fetchCaseDetails(this.remoteSubmitId)
          }
          break
        case 'case_details':
          if (data.data && Array.isArray(data.data.cases)) {
            this.remoteCaseDetails = data.data.cases
            this.isCaseDetailsLoading = false
          }
          break
      }
    },

    // 处理 SSE 错误
    handleSSEError(error) {
      this.addLog(`错误: ${error.message}`)
      this.isRunning = false
    },

    // 处理 SSE 完成
    handleSSEComplete() {
      this.isRunning = false
      if (this.remoteSubmitId && this.remoteCaseDetails.length === 0) {
        this.fetchCaseDetails(this.remoteSubmitId)
      }
      if (this.problemInfo.displayId) {
        this.fetchProblemInfo()
      }
    },

    // 更新远程状态
    updateRemoteStatus(data) {
      const status = data.status
      this.remoteBannerText = status

      // 提取错误信息
      if (data.errorMessage) {
        this.remoteErrorMessage = data.errorMessage
      } else {
        this.remoteErrorMessage = ''
      }

      if (status === '答案正确') {
        this.remoteBannerClass = 'bg-success'
        this.remoteBannerIcon = 'el-icon-success'
      } else if (['等待中', '判题中', '提交中'].includes(status)) {
        this.remoteBannerClass = 'bg-pending'
        this.remoteBannerIcon = 'el-icon-loading'
      } else {
        this.remoteBannerClass = 'bg-error'
        this.remoteBannerIcon = 'el-icon-error'
      }
    },

    isFinalRemoteStatus(status) {
      return !['等待中', '判题中', '提交中'].includes(status)
    },

    async fetchCaseDetails(submitId) {
      if (!submitId || this.isCaseDetailsLoading || this.remoteCaseDetails.length > 0) {
        return
      }

      this.isCaseDetailsLoading = true
      try {
        const maxAttempts = 5
        for (let i = 1; i <= maxAttempts; i++) {
          const res = await getJudgeCaseDetails({
            submit_id: submitId,
            username: this.form.username,
            password: this.form.password
          })
          if (res.code === 200 && res.data && Array.isArray(res.data.cases) && res.data.cases.length > 0) {
            this.remoteCaseDetails = res.data.cases
            break
          }
          if (i < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, 1000))
          }
        }
      } catch (error) {
        console.error('获取测试点详情失败:', error)
        this.addLog(`获取测试点详情失败: ${error.message || '未知错误'}`)
      } finally {
        this.isCaseDetailsLoading = false
      }
    },

    // 更新样例摘要
    updateSampleSummary() {
      const passCount = this.sampleResults.filter(s => s.is_ok).length
      const totalCount = this.sampleResults.length
      this.sampleSummary = `通过: ${passCount}/${totalCount}`
    },

    // 切换样例展开状态
    toggleSample(id) {
      this.$set(this.expandedSamples, id, !this.expandedSamples[id])
    },

    getCaseStatusText(status) {
      const statusMap = {
        '-10': '未提交',
        '-5': '结果未知',
        '-4': '已取消',
        '-3': '格式错误',
        '-2': '编译错误',
        '-1': '答案错误',
        '0': '答案正确',
        '1': '时间超限',
        '2': '内存超限',
        '3': '运行错误',
        '4': '系统错误',
        '6': '等待中',
        '7': '判题中',
        '8': '部分正确',
        '9': '提交中',
        '10': '提交失败'
      }
      if (status === null || status === undefined) return '--'
      return statusMap[String(status)] || `状态${status}`
    },

    getCaseStatusTagType(status) {
      if (status === 0) return 'success'
      if ([6, 7, 9].includes(status)) return 'warning'
      if (status === null || status === undefined) return 'info'
      return 'danger'
    },

    formatCaseTime(time) {
      if (time === null || time === undefined) return '--'
      return `${time} ms`
    },

    formatCaseMemory(memory) {
      if (memory === null || memory === undefined) return '--'
      return `${memory} KB`
    },

    // 添加日志
    addLog(msg) {
      const time = new Date().toLocaleTimeString('zh-CN', { hour12: false })
      this.logs.push({ time, msg })
      this.$nextTick(() => {
        const logArea = this.$refs.logArea
        if (logArea) {
          logArea.scrollTop = logArea.scrollHeight
        }
      })
    },

    // 渲染 Markdown
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },

    // 获取结果类型
    getResultType(result) {
      if (result === '答案正确') return 'success'
      if (['答案错误', '时间超限', '内存超限', '运行错误', '编译错误', '格式错误'].some(s => result.includes(s))) {
        return 'danger'
      }
      if (['等待中', '判题中', '提交中'].includes(result)) return 'warning'
      return 'info'
    },

    // 获取本地结果类型
    getLocalResultType(localInfo) {
      if (!localInfo || localInfo === '-') return 'info'
      if (localInfo.includes('全部通过') || localInfo.includes('All Pass')) return 'success'
      if (localInfo.includes('通过') || localInfo.includes('Pass')) return 'danger'
      return 'info'
    },

    // 简化语言名称
    simplifyLanguage(language) {
      if (!language) return ''
      if (language.includes('C++')) return 'C++'
      if (language.includes('C With')) return 'C'
      if (language.includes('Python')) return 'Py3'
      return language
    },

    // 格式化时间
    formatTime(time) {
      if (!time) return '--'
      try {
        const date = new Date(time)
        if (isNaN(date.getTime())) return '--'

        const beijingTime = new Date(date.getTime() + (8 * 60 * 60 * 1000))

        const year = beijingTime.getUTCFullYear()
        const month = String(beijingTime.getUTCMonth() + 1).padStart(2, '0')
        const day = String(beijingTime.getUTCDate()).padStart(2, '0')
        const hours = String(beijingTime.getUTCHours()).padStart(2, '0')
        const minutes = String(beijingTime.getUTCMinutes()).padStart(2, '0')
        const seconds = String(beijingTime.getUTCSeconds()).padStart(2, '0')

        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
      } catch (e) {
        console.error('时间格式化错误:', e)
        return '--'
      }
    },

    // 历史记录分页切换
    async handleHistoryPageChange(page) {
      if (!this.form.pid) {
        this.$message.warning('请先获取题目信息')
        return
      }
      this.historyPagination.currentPage = page
      await this.fetchHistory()
    },

    // 获取历史记录
    async fetchHistory() {
      try {
        const res = await getJudgeHistory({
          pid: this.form.pid,
          cid: '0',
          page: this.historyPagination.currentPage,
          pageSize: this.historyPagination.pageSize
        })

        if (res.code === 200) {
          this.historyList = res.data.list || []
          this.historyPagination.total = res.data.total || 0
        }
      } catch (error) {
        console.error('获取历史记录失败:', error)
      }
    },

    // 显示代码
    showCode(row) {
      this.currentCode = row.code
      this.currentLanguage = row.language || ''
      this.codeDialogVisible = true
    },

    // 查看历史提交测试点详情
    async showHistoryCaseDetails(row) {
      if (!row || !row.submit_id) {
        this.$message.warning('该记录缺少提交ID，无法查询测试点详情')
        return
      }

      this.historyCaseDialogVisible = true
      this.historyCaseLoading = true
      this.historyCaseDetails = []
      this.historyCaseSubmitId = row.submit_id
      this.historyCaseResult = row.result || ''

      try {
        const res = await getJudgeCaseDetails({
          submit_id: row.submit_id,
          username: this.form.username,
          password: this.form.password
        })
        if (res.code === 200 && res.data && Array.isArray(res.data.cases)) {
          this.historyCaseDetails = res.data.cases
        }
      } catch (error) {
        this.$message.error('获取历史测试点详情失败')
        console.error('获取历史测试点详情失败:', error)
      } finally {
        this.historyCaseLoading = false
      }
    },

    // 获取判题模式文本
    getJudgeModeText(mode) {
      const modeMap = {
        'default': '默认模式',
        'spj': '特殊判题 (SPJ)',
        'interactive': '交互式',
        'subtask': '子任务'
      }
      return modeMap[mode] || mode || '默认模式'
    },

    // 添加代码行号
    addCodeLineNumbers() {
      this.$nextTick(() => {
        setTimeout(() => {
          addCodeBtn()
        }, 100)
      })
    },

    // 映射编程语言到 highlight.js 支持的语言标识
    mapLanguage(lang) {
      const languageMap = {
        // C语言变体
        'c': 'c',
        'C': 'c',
        'C With O2': 'c',

        // C++变体
        'cpp': 'cpp',
        'C++': 'cpp',
        'c++': 'cpp',
        'C++ 17 With O2': 'cpp',
        'C++ 17': 'cpp',
        'C++ 20 With O2': 'cpp',
        'C++ 20': 'cpp',

        // Java
        'java': 'java',
        'Java': 'java',

        // Python变体
        'python': 'python',
        'Python': 'python',
        'py': 'py',
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
        'CSharp': 'csharp'
      }
      return languageMap[lang] || 'plaintext'
    }
  }
}
</script>

<style>
/* 引入 KaTeX 样式 */
@import '~katex/dist/katex.min.css';
</style>

<style scoped>
.judge-terminal-page {
  padding: 20px;
}

.section-title {
  margin-bottom: 15px;
  color: #409eff;
  font-weight: 600;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: #303133;
}

.card-header i {
  margin-right: 5px;
  color: #409eff;
}

/* 卡片样式 */
.config-card,
.code-card,
.log-card,
.history-card,
.problem-card,
.result-card {
  margin-bottom: 15px;
}

/* 日志区域 */
.log-area {
  background-color: #1e1e1e;
  color: #00ff00;
  font-family: 'Consolas', 'Monaco', monospace;
  padding: 10px;
  border-radius: 4px;
  height: 120px;
  overflow-y: auto;
  font-size: 12px;
}

.log-line {
  margin-bottom: 2px;
  line-height: 1.4;
}

.log-time {
  color: #888;
  margin-right: 5px;
}

.clickable-result-tag {
  cursor: pointer;
}

.history-case-header {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
}

/* 题目内容 */
.problem-content {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
}

.problem-title {
  text-align: center;
  margin-bottom: 20px;
  color: #409eff;
  font-size: 20px;
}

.problem-section {
  margin-bottom: 20px;
}

.problem-section h4 {
  color: #409eff;
  font-size: 16px;
  margin-bottom: 10px;
  font-weight: 600;
}

.problem-section h4 i {
  margin-right: 5px;
}

.example-box {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  background-color: #f5f7fa;
}

.example-title {
  font-weight: 600;
  margin-bottom: 5px;
  color: #606266;
  font-size: 12px;
}

/* 题目信息栏 */
.problem-info-bar {
  padding: 12px 15px;
  background-color: #f0f9ff;
  border-left: 4px solid #409eff;
  border-radius: 4px;
  margin-bottom: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.example-box pre {
  margin: 0;
  padding: 8px;
  background-color: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
  font-size: 14px;
}

/* 结果面板 */
.result-banner {
  padding: 30px;
  text-align: center;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #fff;
  font-size: 24px;
  font-weight: bold;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.result-banner i {
  margin-right: 10px;
  font-size: 28px;
}

.bg-success {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.bg-error {
  background: linear-gradient(135deg, #f56c6c, #f78989);
}

.bg-pending {
  background: linear-gradient(135deg, #909399, #b1b3b8);
}

/* 样例测试 */
.sample-section {
  margin-top: 20px;
}

.case-section {
  margin-top: 20px;
}

.sample-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e4e7ed;
}

.sample-header h4 {
  margin: 0;
  color: #606266;
  font-size: 16px;
  font-weight: 600;
}

.sample-list {
  margin-top: 10px;
}

.sample-item {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 10px;
  background: #fff;
  overflow: hidden;
}

.sample-item-header {
  padding: 10px 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  background: #fafafa;
  transition: background 0.3s;
}

.sample-item-header:hover {
  background: #f5f7fa;
}

.sample-item-detail {
  padding: 15px;
  background: #fafafa;
  border-top: 1px solid #e4e7ed;
  font-size: 13px;
}

.sample-item-detail > div {
  margin-bottom: 10px;
}

.sample-item-detail > div:last-child {
  margin-bottom: 0;
}

.code-block {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 10px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  white-space: pre-wrap;
  word-wrap: break-word;
  margin-top: 5px;
  font-size: 12px;
  line-height: 1.5;
}

/* 错误信息块样式 */
.error-section {
  margin-top: 10px;
  padding: 10px;
  background: #FEF0F0;
  border-left: 4px solid #F56C6C;
  border-radius: 4px;
}

.error-block {
  background: #FFF;
  color: #F56C6C;
  padding: 8px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  margin-top: 5px;
  max-height: 200px;
  overflow: auto;
}

/* 远程判题错误信息样式 */
.remote-error-section {
  margin-bottom: 20px;
  padding: 15px;
  background: #FEF0F0;
  border-left: 4px solid #F56C6C;
  border-radius: 4px;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
  color: #303133;
}

.remote-error-block {
  background: #FFF;
  color: #F56C6C;
  padding: 10px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 代码显示容器 - 与学生端一致 */
.code-display-wrapper {
  max-height: 600px;
  overflow: auto;
  margin-top: 10px;
}

/* 判题终端专用样式：代码显示 - 与学生端一致 */
.judge-terminal-page .code-display-wrapper .markdown-body pre {
  padding: 0 16px 0 40px !important;  /* 左侧40px给行号留空间 */
  position: relative !important;
}

.judge-terminal-page .code-display-wrapper .markdown-body pre code {
  padding: 0px 16px 0px 0px !important;  /* code不添加额外缩进，总缩进保持40px */
  line-height: 26px !important;
}

.judge-terminal-page .code-display-wrapper .markdown-body pre ol.pre-numbering {
  line-height: 26px !important;
  font-size: 1rem !important;
}

.judge-terminal-page .code-display-wrapper .markdown-body pre ol.pre-numbering li {
  line-height: 26px !important;
  margin: 0 !important;
  padding: 0 !important;
}

.judge-terminal-page .code-display-wrapper .markdown-body pre ol.pre-numbering li:before {
  font-size: 1rem !important;
  line-height: 26px !important;
  vertical-align: top !important;
}

/* 滚动条样式 */
.log-area::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.log-area::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

.log-area::-webkit-scrollbar-thumb:hover {
  background: #777;
}

.log-area::-webkit-scrollbar-track {
  background: #2d2d2d;
}
</style>

<style>
/* Highlight 组件的样式会自动应用，无需额外定义 */
</style>
