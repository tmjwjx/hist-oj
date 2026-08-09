<template>
  <div class="programming-question">
    <!-- 题目未显示时的提示 -->
    <el-alert v-if="!canViewHomework" type="info" :closable="false" show-icon>
      <template #title>
        {{ $t('m.Homework_Not_Started_Or_Hidden') || '教师尚未显示作业内容，请等待' }}
      </template>
    </el-alert>

    <!-- 题目显示区域 - 受 canViewHomework 控制 -->
    <el-card v-if="canViewHomework && problemInfo && problemInfo.problem" class="problem-display-card" shadow="hover">
      <div slot="header" class="problem-header">
        <span class="problem-title">{{ problemInfo.displayId }} - {{ problemInfo.problem.title }}</span>
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
      </div>

      <div class="problem-content">
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
    </el-card>

    <!-- 题解链接 - 仅在已提交且允许查看答案时显示 -->
    <el-card v-if="canViewAnswer && isSubmitted && canViewHomework && problemInfo" class="solution-card" shadow="hover">
      <div slot="header">
        <span><i class="el-icon-document"></i> 参考题库题解</span>
      </div>
      <el-alert type="info" :closable="false">
        <template slot="title">
          请在 BingOJ 题库中查看本题的题解和讨论
        </template>
      </el-alert>
      <div style="margin-top: 15px; text-align: center;">
        <el-button type="primary" icon="el-icon-link" @click="goToProblem">
          查看 BingOJ 题库题解
        </el-button>
      </div>
    </el-card>

    <!-- 加载状态 -->
    <el-card v-else-if="canViewHomework && loading" class="problem-display-card" v-loading="true">
      <div style="height: 200px;"></div>
    </el-card>

    <!-- 错误状态 - 仅在加载完成但无数据时显示 -->
    <el-alert v-else-if="canViewHomework && !loading && (!problemInfo || !problemInfo.problem)" type="error" :closable="false">
      <p>无法加载题目信息，请检查 BingOJ 题目 ID 是否正确。</p>
    </el-alert>

    <!-- 代码编辑和提交区域 - 受 canViewHomework 控制 -->
    <el-card v-if="canViewHomework && problemInfo && problemInfo.problem" class="code-editor-card" shadow="hover">
      <div slot="header">
        <span><i class="el-icon-edit"></i> 代码提交</span>
      </div>

      <el-form label-width="100px" size="small">
        <el-form-item label="编程语言">
          <el-select v-model="submitForm.language" style="width: 100%">
            <el-option label="C++ 17" value="C++ 17 With O2"></el-option>
            <el-option label="C" value="C With O2"></el-option>
            <el-option label="Python3" value="Python3"></el-option>
            <el-option label="Java" value="Java"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="代码">
          <el-input
            type="textarea"
            v-model="submitForm.code"
            :rows="15"
            placeholder="在此输入代码..."
            style="font-family: 'Consolas', monospace; font-size: 13px"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="submitting"
            @click="submitCode"
            :disabled="!submitForm.code.trim()"
          >
            <i class="el-icon-upload"></i> {{ submitting ? '提交中...' : '提交代码' }}
          </el-button>
          <el-button v-if="isSubmitted && canViewHomework" @click="submitToHOJDirectly">
            <i class="el-icon-link"></i> 在 BingOJ 平台打开
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 提交历史 -->
      <el-divider v-if="submitHistory.length > 0"></el-divider>
      <div v-if="submitHistory.length > 0">
        <h4 style="margin-bottom: 10px;">提交历史</h4>
        <el-alert v-if="!canViewScore" type="info" :closable="false" style="margin-bottom: 10px;">
          <i class="el-icon-info"></i>
          {{ $t('m.Score_Not_Available') || '教师尚未开放查看成绩，提交结果已隐藏' }}
        </el-alert>
        <el-table :data="submitHistory" size="small" stripe>
          <el-table-column label="提交时间" width="150">
            <template slot-scope="{ row }">
              {{ formatTime(row.submitTime) }}
            </template>
          </el-table-column>
          <el-table-column prop="language" label="语言" width="100">
            <template slot-scope="{ row }">
              {{ simplifyLanguage(row.language) }}
            </template>
          </el-table-column>
          <!-- 判题结果列 - 根据 canViewScore 控制是否显示 -->
          <el-table-column v-if="canViewScore" label="结果" width="120">
            <template slot-scope="{ row }">
              <el-tag
                :type="getResultType(row.result)"
                size="mini"
                effect="dark"
                :class="['judge-result-tag', getResultClass(row.result)]"
              >
                <i :class="getResultIcon(row.result)" class="judge-result-icon"></i>
                {{ formatJudgeResult(row.result) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column v-else label="结果" width="120">
            <template slot-scope="{ row }">
              <el-tag type="info" size="mini">
                {{ $t('m.Hidden') || '已隐藏' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template slot-scope="{ row }">
              <el-button type="text" size="small" @click="viewCode(row)">查看代码</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 代码查看对话框 -->
    <el-dialog title="提交代码" :visible.sync="showCodeDialog" width="60%">
      <div class="code-display-wrapper">
        <ClassroomCodeViewer
          :code="currentCode"
          :language="mapLanguage(currentLanguage)"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getJudgeInfo } from '@/common/judgeTerminal'
import api from '@/common/api'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import axios from 'axios'
const ClassroomCodeViewer = () => import('@/components/classroom/ClassroomCodeViewer')

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(MarkdownItKatex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'ProgrammingQuestion',
  components: {
    ClassroomCodeViewer
  },
  props: {
    problemId: {
      type: String,
      required: true
    },
    homeworkQuestionId: {
      type: [String, Number],
      required: true
    },
    homeworkId: {
      type: [String, Number],
      required: true
    },
    canViewHomework: {
      type: Boolean,
      default: true // 默认可以查看作业内容
    },
    canViewScore: {
      type: Boolean,
      default: true // 默认可以查看成绩
    },
    canViewAnswer: {
      type: Boolean,
      default: false // 默认不可以查看答案
    },
    isSubmitted: {
      type: Boolean,
      default: false // 作业是否已提交
    }
  },
  data() {
    return {
      loading: false,
      problemInfo: null,
      examples: [],
      submitForm: {
        language: 'C++ 17 With O2',
        code: ''
      },
      submitting: false,
      submitHistory: [],
      submitHistoryLoading: false,
      resultPollingTimer: null,
      resultPollingInterval: 2000,
      resultPollingStartedAt: 0,
      resultPollingMaxDuration: 120000,
      showCodeDialog: false,
      currentCode: '',
      currentLanguage: ''
    }
  },
  mounted() {
    this.loadProblemInfo()
    this.initSubmitHistory()
  },
  beforeDestroy() {
    this.stopResultPolling()
  },
  methods: {
    async initSubmitHistory() {
      const hasPending = await this.loadSubmitHistory()
      if (hasPending) {
        this.startResultPolling()
      }
    },
    startResultPolling() {
      if (this.resultPollingTimer) {
        return
      }
      this.resultPollingStartedAt = Date.now()
      this.resultPollingTimer = setInterval(async () => {
        const hasPending = await this.loadSubmitHistory(true)
        const timeout = Date.now() - this.resultPollingStartedAt >= this.resultPollingMaxDuration
        if (!hasPending || timeout) {
          this.stopResultPolling()
        }
      }, this.resultPollingInterval)
    },
    stopResultPolling() {
      if (this.resultPollingTimer) {
        clearInterval(this.resultPollingTimer)
        this.resultPollingTimer = null
      }
    },
    async loadProblemInfo() {
      this.loading = true
      try {
        // 使用和教师端相同的API
        const res = await api.getProblem(this.problemId, '0', undefined)

        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          this.problemInfo = res.data.data
          this.extractExamples()
        } else {
          console.error('学生端题目加载失败, 响应状态:', res?.status, '数据:', res?.data)
          this.$message.error('获取题目失败')
        }
      } catch (error) {
        // 只保留关键错误日志
        if (error.response && error.response.status !== 502) {
          console.error('学生端加载题目异常:', error.message)
        }
        // 检查是否是502错误（网关错误）
        if (error.response && error.response.status === 502) {
          this.$message.error('加载题目失败：BingOJ服务繁忙，请稍后重试')
        } else if (error.response && error.response.status === 404) {
          this.$message.error(`加载题目失败：题目ID ${this.problemId} 不存在`)
        } else if (error.response && error.response.status === 403) {
          this.$message.error('加载题目失败：您没有权限查看此题目')
        } else {
          this.$message.error('加载题目失败：' + (error.message || '请检查网络连接'))
        }
      } finally {
        this.loading = false
      }
    },
    extractExamples() {
      if (!this.problemInfo || !this.problemInfo.problem.examples) {
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
    async submitCode() {
      if (!this.submitForm.code.trim()) {
        this.$message.warning('请输入代码')
        return
      }

      // 获取当前登录用户信息
      const userInfo = this.$store.getters.userInfo
      if (!userInfo || !userInfo.username) {
        this.$message.error('请先登录')
        return
      }

      // 从 localStorage 获取 token
      const token = localStorage.getItem('token')
      if (!token) {
        this.$message.error('未找到登录凭证，请重新登录')
        return
      }

      this.submitting = true
      try {
        // 调用 hist-oj 提交 API,使用当前登录用户的 token
        const response = await axios.post('/judge-api/submit', {
          pid: this.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token, // 使用 token 而不是密码
          password: '',
          language: this.submitForm.language,
          code: this.submitForm.code
        })

        if (response.data && response.data.code === 200) {
          const submitId = response.data.data.submit_id
          // 保存提交记录到作业提交表（失败时自动重试一次）
          let saved = false
          let saveError = null
          for (let attempt = 0; attempt < 2; attempt++) {
            try {
              await this.saveSubmissionRecord(submitId)
              saved = true
              break
            } catch (err) {
              saveError = err
              if (attempt === 0) {
                await new Promise(resolve => setTimeout(resolve, 800))
              }
            }
          }

          if (!saved) {
            const msg = saveError && saveError.message
              ? saveError.message
              : '提交记录保存失败'
            this.$message.error(`代码已提交到判题端，但作业记录保存失败：${msg}`)
            return
          }

          this.$emit('programming-submitted', {
            problemId: this.problemId,
            homeworkQuestionId: this.homeworkQuestionId,
            submitId
          })
          this.$message.success('提交成功')
          const hasPending = await this.loadSubmitHistory(true)
          if (hasPending) {
            this.startResultPolling()
          } else {
            this.stopResultPolling()
          }
        } else {
          this.$message.error(response.data?.message || '提交失败')
        }
      } catch (error) {
        console.error('提交失败:', error)
        this.$message.error('提交失败')
      } finally {
        this.submitting = false
      }
    },
    async saveSubmissionRecord(submitId) {
      // 从 localStorage 获取 token
      const token = localStorage.getItem('token')

      const res = await this.$store.dispatch('classroom/saveProgrammingSubmission', {
        homeworkId: this.homeworkId,
        problemId: this.problemId,
        submitId: submitId,
        code: this.submitForm.code,
        language: this.submitForm.language,
        token: token
      })

      if (!res || res.code !== 200) {
        throw new Error((res && res.message) || '保存提交记录失败')
      }
    },
    async loadSubmitHistory(forceRefresh = false) {
      if (this.submitHistoryLoading) {
        return this.submitHistory.some(item => this.isPendingResult(item.result))
      }
      this.submitHistoryLoading = true
      try {
        const res = await this.$store.dispatch('classroom/getProgrammingSubmissions', {
          homeworkId: this.homeworkId,
          homeworkQuestionId: this.homeworkQuestionId,
          forceRefresh: !!forceRefresh
        })
        if (res.code === 200) {
          this.submitHistory = res.data || []
          if (this.submitHistory.length > 0) {
            const latest = this.submitHistory[0]
            this.$emit('programming-result-updated', {
              problemId: this.problemId,
              homeworkQuestionId: this.homeworkQuestionId,
              result: latest.result || '',
              isPending: this.isPendingResult(latest.result)
            })
          }
          return this.submitHistory.some(item => this.isPendingResult(item.result))
        }
        return false
      } catch (error) {
        console.error('加载提交历史失败:', error)
        return false
      } finally {
        this.submitHistoryLoading = false
      }
    },
    submitToHOJDirectly() {
      window.open(`/problem/${this.problemId}`, '_blank')
    },
    viewCode(row) {
      this.currentCode = row.code
      this.currentLanguage = row.language || ''
      this.showCodeDialog = true
    },
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    getJudgeModeText(mode) {
      const modeMap = {
        'default': '默认模式',
        'spj': '特殊判题 (SPJ)',
        'interactive': '交互式',
        'subtask': '子任务'
      }
      return modeMap[mode] || mode || '默认模式'
    },
    formatTime(time) {
      if (!time) return '--'
      try {
        const date = new Date(time)
        return date.toLocaleString('zh-CN')
      } catch (e) {
        return '--'
      }
    },
    simplifyLanguage(language) {
      if (!language) return ''
      if (language.includes('C++')) return 'C++'
      if (language.includes('C With')) return 'C'
      if (language.includes('Python')) return 'Py3'
      return language
    },
    normalizeJudgeResultCode(result) {
      if (!result || String(result).trim() === '') return 'PENDING'
      const raw = String(result).trim()
      const upper = raw.toUpperCase()

      if (['AC', 'ACCEPTED', '答案正确'].includes(upper) || raw === '答案正确') return 'AC'
      if (['WA', 'WRONG ANSWER', '答案错误'].includes(upper) || raw === '答案错误') return 'WA'
      if (['CE', 'COMPILATION ERROR', '编译错误'].includes(upper) || raw === '编译错误') return 'CE'
      if (['RE', 'RUNTIME ERROR', '运行错误'].includes(upper) || raw === '运行错误') return 'RE'
      if (['TLE', 'TIME LIMIT EXCEEDED', '时间超限'].includes(upper) || raw === '时间超限') return 'TLE'
      if (['MLE', 'MEMORY LIMIT EXCEEDED', '内存超限'].includes(upper) || raw === '内存超限') return 'MLE'
      if (['PE', 'PRESENTATION ERROR', '格式错误'].includes(upper) || raw === '格式错误') return 'PE'
      if (['PAC', 'PARTIAL ACCEPTED'].includes(upper)) return 'PAC'
      if (['SE', 'SYSTEM ERROR'].includes(upper)) return 'SE'
      if (['CA', 'CANCELLED'].includes(upper)) return 'CA'
      if (['SNR', 'SUBMITTED UNKNOWN RESULT'].includes(upper)) return 'SNR'
      if (['PENDING', 'JUDGING', 'SUBMITTING', 'COMPILING', '等待中', '判题中', '提交中', '评测中...'].includes(raw) ||
          ['PENDING', 'JUDGING', 'SUBMITTING', 'COMPILING'].includes(upper)) {
        return 'PENDING'
      }
      if (upper.startsWith('UNKNOWN(')) return 'UNKNOWN'
      return upper
    },
    isPendingResult(result) {
      return this.normalizeJudgeResultCode(result) === 'PENDING'
    },
    formatJudgeResult(result) {
      const code = this.normalizeJudgeResultCode(result)
      if (code === 'PENDING') return '评测中'
      if (['AC', 'WA', 'CE', 'RE', 'TLE', 'MLE', 'PE', 'PAC', 'SE', 'CA', 'SNR'].includes(code)) {
        return code
      }
      const raw = String(result || '').trim()
      return raw || '评测中'
    },
    getResultType(result) {
      const code = this.normalizeJudgeResultCode(result)
      if (code === 'AC') return 'success'
      if (code === 'PAC') return 'primary'
      if (code === 'PENDING') return 'warning'
      if (['SE', 'CA', 'SNR', 'UNKNOWN'].includes(code)) return 'info'
      return 'danger'
    },
    getResultClass(result) {
      const code = this.normalizeJudgeResultCode(result)
      if (code === 'AC') return 'judge-result-ac'
      if (code === 'PENDING') return 'judge-result-pending'
      if (code === 'PAC') return 'judge-result-pac'
      if (['WA', 'CE', 'RE', 'TLE', 'MLE', 'PE'].includes(code)) return 'judge-result-error'
      return 'judge-result-neutral'
    },
    getResultIcon(result) {
      const code = this.normalizeJudgeResultCode(result)
      if (code === 'AC') return 'el-icon-success'
      if (code === 'PENDING') return 'el-icon-loading'
      if (code === 'PAC') return 'el-icon-star-on'
      if (['WA', 'CE', 'RE', 'TLE', 'MLE', 'PE'].includes(code)) return 'el-icon-error'
      return 'el-icon-info'
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
    },
    goToProblem() {
      // 跳转到 HOJ 题库页面
      window.open(`/problem/${this.problemId}`, '_blank')
    }
  }
}
</script>

<style scoped>
.programming-question {
  margin-top: 20px;
}

.problem-display-card,
.code-editor-card {
  margin-bottom: 20px;
}

.problem-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.problem-title {
  font-size: 18px;
  font-weight: bold;
  color: #409EFF;
}

.problem-info-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.problem-content {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
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

.code-display-wrapper {
  max-height: 600px;
  overflow: auto;
  margin-top: 10px;
}

.judge-result-tag {
  min-width: 74px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  letter-spacing: 0.2px;
}

.judge-result-icon {
  margin-right: 4px;
}

.judge-result-tag.judge-result-ac {
  box-shadow: 0 0 0 1px rgba(103, 194, 58, 0.35) inset;
}

.judge-result-tag.judge-result-pending {
  box-shadow: 0 0 0 1px rgba(230, 162, 60, 0.35) inset;
}

.judge-result-tag.judge-result-pac {
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.35) inset;
}

.judge-result-tag.judge-result-error {
  box-shadow: 0 0 0 1px rgba(245, 108, 108, 0.35) inset;
}
</style>

<style>
@import '~katex/dist/katex.min.css';
</style>
