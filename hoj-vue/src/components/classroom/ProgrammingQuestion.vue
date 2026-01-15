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
              <el-tag :type="getResultType(row.result)" size="mini">
                {{ row.result || '提交中' }}
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
      <pre class="code-preview">{{ currentCode }}</pre>
    </el-dialog>
  </div>
</template>

<script>
import { getJudgeInfo } from '@/common/judgeTerminal'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import axios from 'axios'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(MarkdownItKatex)

export default {
  name: 'ProgrammingQuestion',
  props: {
    problemId: {
      type: String,
      required: true
    },
    questionId: {
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
      showCodeDialog: false,
      currentCode: ''
    }
  },
  mounted() {
    this.loadProblemInfo()
    this.loadSubmitHistory()
  },
  methods: {
    async loadProblemInfo() {
      this.loading = true
      try {
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

        const res = await getJudgeInfo({
          pid: this.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token, // 使用 token 而不是密码
          password: ''
        })

        if (res.code === 200 && res.data) {
          this.problemInfo = res.data
          this.extractExamples()
        } else {
          this.$message.error(res.message || '获取题目失败')
        }
      } catch (error) {
        console.error('加载题目失败:', error)
        this.$message.error('加载题目失败')
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
          // 保存提交记录到作业提交表
          await this.saveSubmissionRecord(submitId)
          this.$message.success('提交成功')
          this.loadSubmitHistory()
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
      try {
        // 从 localStorage 获取 token
        const token = localStorage.getItem('token')

        await this.$store.dispatch('classroom/saveProgrammingSubmission', {
          homeworkId: this.homeworkId,
          problemId: this.problemId, // 修复：使用 problemId 而不是 questionId
          submitId: submitId,
          code: this.submitForm.code,
          language: this.submitForm.language,
          token: token
        })
      } catch (error) {
        console.error('保存提交记录失败:', error)
      }
    },
    async loadSubmitHistory() {
      try {
        const res = await this.$store.dispatch('classroom/getProgrammingSubmissions', {
          homeworkId: this.homeworkId,
          questionId: this.questionId
        })
        if (res.code === 200) {
          this.submitHistory = res.data || []
        }
      } catch (error) {
        console.error('加载提交历史失败:', error)
      }
    },
    submitToHOJDirectly() {
      window.open(`/problem/${this.problemId}`, '_blank')
    },
    viewCode(row) {
      this.currentCode = row.code
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
    getResultType(result) {
      if (result === '答案正确') return 'success'
      if (['答案错误', '时间超限', '内存超限', '运行错误', '编译错误', '格式错误'].some(s => result.includes(s))) {
        return 'danger'
      }
      if (['等待中', '判题中', '提交中'].includes(result)) return 'warning'
      return 'info'
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

.code-preview {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 500px;
  overflow-y: auto;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
}
</style>

<style>
@import '~katex/dist/katex.min.css';
</style>
