<template>
  <div class="judge-terminal">
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
              title="凭据说明"
              type="info"
              :closable="false"
              style="margin-bottom: 15px; padding: 8px 12px;">
              首次输入的用户名和密码将在验证成功后自动保存，下次访问时自动填充
            </el-alert>
            <el-row :gutter="10">
              <el-col :span="12">
                <el-form-item label="用户名">
                  <el-input v-model="form.username" placeholder="请输入用户名" clearable></el-input>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="密码">
                  <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password clearable></el-input>
                </el-form-item>
              </el-col>
            </el-row>
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
                <el-tag :type="getResultType(scope.row.result)" size="mini">
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
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 代码预览对话框 -->
    <el-dialog title="提交代码预览" :visible.sync="codeDialogVisible" width="60%">
      <pre class="code-preview">{{ currentCode }}</pre>
    </el-dialog>
  </div>
</template>

<script>
import { getJudgeInfo, getJudgeHistory, runCombinedJudge } from '@/common/judgeTerminal'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt()
md.use(MarkdownItKatex)

export default {
  name: 'JudgeTerminal',
  data() {
    return {
      form: {
        username: '',
        password: '',
        mode: 'normal', // 固定为普通模式
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
      sampleResults: [],
      sampleSummary: '等待中',
      expandedSamples: {},
      codeDialogVisible: false,
      currentCode: ''
    }
  },
  mounted() {
    // 从 localStorage 读取用户凭据
    this.loadUserCredentials()
  },
  methods: {
    // 从 localStorage 加载用户凭据
    loadUserCredentials() {
      try {
        const savedCredentials = localStorage.getItem('judge_terminal_credentials')
        if (savedCredentials) {
          const credentials = JSON.parse(savedCredentials)
          this.form.username = credentials.username || ''
          this.form.password = credentials.password || ''
          if (this.form.username && this.form.password) {
            console.log('已从本地加载用户凭据')
          }
        }
      } catch (error) {
        console.error('读取凭据失败:', error)
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
        console.log('凭据已保存到本地')
      } catch (error) {
        console.error('保存凭据失败:', error)
        this.$message.warning('凭据保存失败，请检查浏览器设置')
      }
    },
    // 获取题目信息
    async fetchProblemInfo() {
      if (!this.form.pid) {
        this.$message.warning('请输入题目ID')
        return
      }

      // 验证用户名和密码
      if (!this.form.username || !this.form.password) {
        this.$message.warning('请输入用户名和密码')
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
          // 保存凭据到本地
          this.saveUserCredentials()

          this.problemInfo = res.data
          this.historyPagination.currentPage = 1 // 重置到第一页
          await this.fetchHistory() // 获取分页历史记录
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

      // 验证用户名和密码
      if (!this.form.username || !this.form.password) {
        this.$message.warning('请输入用户名和密码')
        return
      }

      // 保存凭据
      this.saveUserCredentials()

      this.isRunning = true
      this.showResult = true
      this.logs = []
      this.sampleResults = []
      this.remoteBannerText = '本地测试中...'
      this.remoteBannerClass = 'bg-pending'
      this.remoteBannerIcon = 'el-icon-loading'

      // 滚动到结果面板
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
          password: this.form.password,
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
          this.updateRemoteStatus(data.data.status)
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
      // 刷新历史记录
      if (this.problemInfo.displayId) {
        this.fetchProblemInfo()
      }
    },

    // 更新远程状态
    updateRemoteStatus(status) {
      this.remoteBannerText = status

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

    // 格式化时间（北京时间）
    formatTime(time) {
      if (!time) return '--'
      try {
        // 处理 ISO 8601 格式的时间字符串
        const date = new Date(time)
        if (isNaN(date.getTime())) return '--'

        // 转换为北京时间（UTC+8）
        const beijingTime = new Date(date.getTime() + (8 * 60 * 60 * 1000))

        // 格式化为 YYYY-MM-DD HH:mm:ss
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
      this.codeDialogVisible = true
    }
  }
}
</script>

<style>
/* 引入 KaTeX 样式 */
@import '~katex/dist/katex.min.css';
</style>

<style scoped>
.judge-terminal {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: calc(100vh - 100px);
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

/* 代码预览 */
.code-preview {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 600px;
  overflow-y: auto;
  font-size: 13px;
  line-height: 1.6;
}

/* 滚动条样式 */
.log-area::-webkit-scrollbar,
.code-preview::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.log-area::-webkit-scrollbar-thumb,
.code-preview::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

.log-area::-webkit-scrollbar-thumb:hover,
.code-preview::-webkit-scrollbar-thumb:hover {
  background: #777;
}

.log-area::-webkit-scrollbar-track,
.code-preview::-webkit-scrollbar-track {
  background: #2d2d2d;
}
</style>
