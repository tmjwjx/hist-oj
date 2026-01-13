<template>
  <div class="create-homework">
    <div class="header">
      <h2>{{ isEditMode ? '编辑作业' : '创建作业' }}</h2>
      <div class="actions">
        <el-button @click="goBack">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="saveHomework" :loading="submitting">
          {{ isEditMode ? '保存' : '创建' }}
        </el-button>
      </div>
    </div>

    <el-form :model="form" :rules="rules" ref="homeworkForm" label-width="120px" class="homework-form">
      <el-card class="form-section">
        <div slot="header">{{ $t('m.Basic_Info') }}</div>
        <el-form-item :label="$t('m.Homework_Title')" prop="title">
          <el-input v-model="form.title" :placeholder="$t('m.Please_Enter_Homework_Title')" />
        </el-form-item>
        <el-form-item :label="$t('m.Description')" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            :placeholder="$t('m.Homework_Description_Tip')"
          />
        </el-form-item>
        <el-form-item :label="$t('m.Start_Time')" prop="startTime">
          <el-date-picker
            v-model="form.startTime"
            type="datetime"
            :placeholder="$t('m.Please_Select_Start_Time')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('m.End_Time')" prop="endTime">
          <el-date-picker
            v-model="form.endTime"
            type="datetime"
            :placeholder="$t('m.Please_Select_End_Time')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('m.Display_Settings')">
          <el-checkbox v-model="form.showHomework">{{ $t('m.Show_Homework_After_Complete') }}</el-checkbox>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">
            {{ $t('m.Show_Homework_Tip') }}
          </span>
          <br><br>
          <el-checkbox v-model="form.showScore">{{ $t('m.Show_Score_After_Complete') }}</el-checkbox>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">
            {{ $t('m.Show_Score_Tip') }}
          </span>
          <br><br>
          <el-checkbox v-model="form.showAnswer">允许学生提交后查看答案</el-checkbox>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">
            学生提交作业后，可以查看每道题的正确答案（主观题显示参考答案，编程题显示题库链接）
          </span>
        </el-form-item>
      </el-card>

      <el-card class="form-section">
        <div slot="header">
          <span>{{ $t('m.Select_Questions') }}</span>
          <span style="margin-left: 10px; color: #909399; font-size: 12px;">
            ({{ $t('m.Selected') }}: {{ selectedQuestions.length }})
          </span>
        </div>

        <div class="filter-bar">
          <el-input
            v-model="searchKeyword"
            :placeholder="$t('m.Search_Questions')"
            prefix-icon="el-icon-search"
            style="width: 300px"
            clearable
            @clear="loadQuestionBank"
            @keyup.enter.native="loadQuestionBank"
          />
          <el-select v-model="filterType" :placeholder="$t('m.Question_Type')" clearable style="width: 150px; margin-left: 10px" @change="loadQuestionBank">
            <el-option :label="$t('m.All')" value="" />
            <el-option :label="$t('m.Single_Choice')" value="single_choice" />
            <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
            <el-option :label="$t('m.Judge')" value="judge" />
            <el-option :label="$t('m.Subjective')" value="subjective" />
            <el-option :label="$t('m.Programming')" value="programming" />
          </el-select>
          <el-button type="primary" icon="el-icon-search" style="margin-left: 10px" @click="loadQuestionBank">
            {{ $t('m.Search') }}
          </el-button>
          <el-button type="success" icon="el-icon-plus" style="margin-left: 10px" @click="showAddProgrammingDialog = true">
            添加 HOJ 编程题
          </el-button>
        </div>

        <div class="question-list" v-loading="questionsLoading">
          <el-table
            :data="questionBank"
            stripe
            style="width: 100%"
          >
            <el-table-column prop="title" :label="$t('m.Question_Title')" min-width="200" />
            <el-table-column prop="type" :label="$t('m.Question_Type')" width="120">
              <template slot-scope="{ row }">
                {{ getQuestionTypeText(row.type) }}
              </template>
            </el-table-column>
            <el-table-column prop="difficulty" :label="$t('m.Difficulty')" width="100">
              <template slot-scope="{ row }">
                <el-rate :value="getDifficultyStars(row.difficulty)" disabled />
              </template>
            </el-table-column>
            <el-table-column prop="score" :label="$t('m.Score')" width="80" />
            <el-table-column :label="$t('m.Operation')" width="150">
              <template slot-scope="{ row }">
                <el-button
                  type="primary"
                  size="small"
                  icon="el-icon-plus"
                  @click="addQuestion(row)"
                  :disabled="isQuestionSelected(row)"
                >
                  {{ isQuestionSelected(row) ? '已添加' : '添加' }}
                </el-button>
                <el-button type="text" @click="viewQuestion(row)" style="margin-left: 5px;">
                  {{ $t('m.View_Detail') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-if="total > 0"
            :current-page="currentPage"
            :page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            style="margin-top: 20px; text-align: right"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>

      <!-- 已选题目列表 -->
      <el-card class="form-section" v-if="selectedQuestions.length > 0">
        <div slot="header">
          <span>已选题目 ({{ selectedQuestions.length }})</span>
          <el-button
            type="primary"
            size="small"
            icon="el-icon-view"
            style="float: right; margin-top: -5px;"
            @click="showFullPreview"
          >
            全卷预览
          </el-button>
        </div>
        <el-table :data="selectedQuestions" stripe>
          <el-table-column :label="$t('m.Order_Number')" width="80">
            <template slot-scope="{ $index }">
              {{ $index + 1 }}
            </template>
          </el-table-column>
          <el-table-column prop="title" :label="$t('m.Question_Title')" min-width="200" />
          <el-table-column prop="type" :label="$t('m.Question_Type')" width="120">
            <template slot-scope="{ row }">
              {{ getQuestionTypeText(row.type) }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Score')" width="150">
            <template slot-scope="{ row }">
              <el-input-number
                v-model="row.score"
                :min="1"
                :max="100"
                size="small"
                style="width: 120px"
              />
            </template>
          </el-table-column>
          <el-table-column :label="$t('m.Operation')" width="180">
            <template slot-scope="{ row, $index }">
              <el-button
                type="text"
                size="small"
                :disabled="$index === 0"
                @click="moveUp($index)"
                icon="el-icon-arrow-up"
              >
                {{ $t('m.Move_Up') }}
              </el-button>
              <el-button
                type="text"
                size="small"
                :disabled="$index === selectedQuestions.length - 1"
                @click="moveDown($index)"
                icon="el-icon-arrow-down"
              >
                {{ $t('m.Move_Down') }}
              </el-button>
              <el-button type="text" @click="removeQuestion(row)" style="color: #F56C6C;">
                {{ $t('m.Remove') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-form>

    <!-- 添加 HOJ 编程题对话框 -->
    <el-dialog title="添加 HOJ 编程题" :visible.sync="showAddProgrammingDialog" width="900px">
      <el-form :model="programmingForm" label-width="120px">
        <el-form-item label="HOJ 题目 ID" required>
          <el-input v-model="programmingForm.problemId" placeholder="请输入 HOJ 题目 ID（如 0001）" style="width: 300px;" />
          <el-button
            type="primary"
            icon="el-icon-search"
            style="margin-left: 10px;"
            @click="fetchProgrammingProblemInfo"
            :loading="fetchingProblem"
          >
            获取题目信息
          </el-button>
          <div style="margin-top: 8px; color: #909399; font-size: 12px;">
            <i class="el-icon-info"></i>
            输入 HOJ 题库中的题目 ID，点击"获取题目信息"预览题目内容
          </div>
        </el-form-item>

        <!-- 题目预览区域 -->
        <div v-if="programmingProblemPreview" class="problem-preview">
          <el-divider content-position="left">题目预览</el-divider>
          <el-card>
            <h3>{{ programmingProblemPreview.problem.title }}</h3>
            <div class="problem-meta">
              <el-tag size="small">题目ID: {{ programmingProblemPreview.problem.problemId }}</el-tag>
              <el-tag size="small" type="info">时间限制: {{ programmingProblemPreview.problem.timeLimit }}ms</el-tag>
              <el-tag size="small" type="warning">内存限制: {{ programmingProblemPreview.problem.memoryLimit }}MB</el-tag>
              <el-tag size="small" type="success">判题模式: {{ getJudgeModeText(programmingProblemPreview.problem.judgeMode) }}</el-tag>
            </div>
            <div class="problem-content">
              <div class="content-section">
                <h4>题目描述</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.description)"></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.input">
                <h4>输入格式</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.input)"></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.output">
                <h4>输出格式</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.output)"></div>
              </div>
              <div class="content-section" v-if="programmingExamples.length > 0">
                <h4>样例</h4>
                <div v-for="(example, index) in programmingExamples" :key="index" class="example-item">
                  <el-alert :title="`样例 ${index + 1}`" type="info" :closable="false">
                    <div slot="default">
                      <p><strong>输入：</strong></p>
                      <pre>{{ example.input }}</pre>
                      <p><strong>输出：</strong></p>
                      <pre>{{ example.output }}</pre>
                    </div>
                  </el-alert>
                </div>
              </div>
            </div>
          </el-card>
        </div>

        <el-form-item label="分值" required>
          <el-input-number v-model="programmingForm.score" :min="1" :max="100" />
        </el-form-item>
      </el-form>

      <span slot="footer">
        <el-button @click="showAddProgrammingDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="confirmAddProgrammingQuestion"
          :disabled="!programmingProblemPreview"
        >
          确定添加
        </el-button>
      </span>
    </el-dialog>

    <!-- 全卷预览对话框 -->
    <el-dialog
      title="全卷预览（学生视角）"
      :visible.sync="showFullPreviewDialog"
      width="900px"
      top="5vh"
      :append-to-body="false"
      :close-on-click-modal="false"
      @opened="handlePreviewOpened"
    >
      <div class="full-preview-container" v-if="selectedQuestions.length > 0">
        <!-- 作业基本信息 -->
        <el-card class="preview-info" shadow="never">
          <h3>{{ form.title || '作业标题' }}</h3>
          <p><strong>描述：</strong>{{ form.description || '无' }}</p>
          <p><strong>开始时间：</strong>{{ formatTime(form.startTime) }}</p>
          <p><strong>结束时间：</strong>{{ formatTime(form.endTime) }}</p>
          <p><strong>总分：</strong>{{ getTotalScore() }} 分</p>
          <p><strong>题目数量：</strong>{{ selectedQuestions.length }} 题</p>
        </el-card>

        <!-- 题目列表（和学生看到的一样） -->
        <div class="questions-preview">
          <el-divider content-position="left">题目内容</el-divider>
          <div v-for="(item, index) in selectedQuestions" :key="item.id || item.problemId" class="question-item">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <span class="question-type">({{ getQuestionTypeText(item.type) }})</span>
              <span class="question-score">{{ item.score }}分</span>
            </div>
            <div class="question-title markdown-body" v-html="renderMarkdown(item.title)"></div>
            <div class="question-content markdown-body" v-html="renderMarkdown(item.content)"></div>

            <!-- 单选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'single_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
                </div>
              </div>
            </div>

            <!-- 多选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'multiple_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
                </div>
              </div>
            </div>

            <!-- 判断题（预览模式，不可作答） -->
            <div v-if="item.type === 'judge'" class="question-options">
              <div class="option-preview">
                <span class="option-letter">✓</span>
                <span class="option-text">对</span>
              </div>
              <div class="option-preview">
                <span class="option-letter">✗</span>
                <span class="option-text">错</span>
              </div>
            </div>

            <!-- 主观题 -->
            <div v-if="item.type === 'subjective'" class="subjective-preview">
              <el-alert type="info" :closable="false">
                <i class="el-icon-edit"></i> 主观题，学生需要在此处输入文字答案
              </el-alert>
            </div>

            <!-- 编程题 -->
            <div v-if="item.type === 'programming'" class="programming-preview">
              <template v-if="getProgrammingProblemInfo(item.problemId)">
                <div class="problem-preview">
                  <el-divider content-position="left">编程题详情</el-divider>
                  <el-card>
                    <h3>{{ getProgrammingProblemInfo(item.problemId).problem.title }}</h3>
                    <div class="problem-meta">
                      <el-tag size="small">题目ID: {{ getProgrammingProblemInfo(item.problemId).problem.problemId }}</el-tag>
                      <el-tag size="small" type="info">时间限制: {{ getProgrammingProblemInfo(item.problemId).problem.timeLimit }}ms</el-tag>
                      <el-tag size="small" type="warning">内存限制: {{ getProgrammingProblemInfo(item.problemId).problem.memoryLimit }}MB</el-tag>
                      <el-tag size="small" type="success">判题模式: {{ getJudgeModeText(getProgrammingProblemInfo(item.problemId).problem.judgeMode) }}</el-tag>
                    </div>
                    <div class="problem-content">
                      <div class="content-section">
                        <h4>题目描述</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.description)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.input">
                        <h4>输入格式</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.input)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.output">
                        <h4>输出格式</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.output)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingExamplesForProblem(item.problemId).length > 0">
                        <h4>样例</h4>
                        <div v-for="(example, idx) in getProgrammingExamplesForProblem(item.problemId)" :key="idx" class="example-item">
                          <el-alert :title="`样例 ${idx + 1}`" type="info" :closable="false">
                            <div slot="default">
                              <p><strong>输入：</strong></p>
                              <pre>{{ example.input }}</pre>
                              <p><strong>输出：</strong></p>
                              <pre>{{ example.output }}</pre>
                            </div>
                          </el-alert>
                        </div>
                      </div>
                    </div>
                  </el-card>
                </div>
              </template>

              <el-alert v-else type="warning" :closable="false">
                <p>学生将看到完整的编程题目界面，包括题目描述、样例、代码编辑器和提交按钮</p>
              </el-alert>

              <el-alert v-if="!item.problemId" type="warning" :closable="false">
                此编程题未关联 HOJ 题目 ID
              </el-alert>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 题目详情对话框 -->
    <el-dialog :title="$t('m.Question_Detail')" :visible.sync="showDetailDialog" width="800px">
      <div v-if="currentQuestion" class="question-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            {{ getQuestionTypeText(currentQuestion.type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Question_Title')">
            <div v-html="renderMarkdown(currentQuestion.title)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Content')">
            <div v-html="renderMarkdown(currentQuestion.content)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Options')" v-if="currentQuestion.options">
            <div v-html="renderOptions(currentQuestion.options)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Answer')">
            {{ currentQuestion.answer }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate :value="currentQuestion.difficulty" :max="3" disabled />
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Score')">
            {{ currentQuestion.score }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import { getJudgeInfo } from '@/common/judgeTerminal'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  // 配置不要把单独的 ^ 当作上标
  typographer: false
})
// 配置 KaTeX，只在 $...$ 或 $$...$$ 中渲染数学公式
md.use(MarkdownItKatex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  // 禁用自动识别上标/下标，避免 ^9 被当作上标渲染
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'CreateHomework',
  data() {
    return {
      submitting: false,
      questionsLoading: false,
      searchKeyword: '',
      filterType: '',
      currentPage: 1,
      pageSize: 10,
      total: 0,
      questionBank: [],
      selectedQuestions: [],
      showDetailDialog: false,
      showFullPreviewDialog: false, // 全卷预览对话框
      currentQuestion: null,
      // HOJ 编程题相关
      showAddProgrammingDialog: false,
      programmingForm: {
        problemId: '',
        score: 10
      },
      programmingProblemPreview: null,
      programmingExamples: [],
      fetchingProblem: false,
      programmingProblemsCache: {}, // 缓存编程题信息，用于预览
      form: {
        title: '',
        description: '',
        startTime: null,
        endTime: null,
        showHomework: false,
        showScore: true,
        showAnswer: false
      },
      rules: {
        title: [{ required: true, message: this.$t('m.Please_Enter_Homework_Title'), trigger: 'blur' }],
        startTime: [{ required: true, message: this.$t('m.Please_Select_Start_Time'), trigger: 'change' }],
        endTime: [{ required: true, message: this.$t('m.Please_Select_End_Time'), trigger: 'change' }]
      }
    }
  },
  computed: {
    classroomId() {
      return this.$route.params.classroomId
    },
    isEditMode() {
      return !!this.$route.query.editId
    },
    editId() {
      return this.$route.query.editId
    }
  },
  mounted() {
    this.loadQuestionBank()
    // 如果是编辑模式，加载作业数据
    if (this.isEditMode) {
      this.loadHomeworkData()
    }
  },
  methods: {
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const params = {
          classroomId: this.classroomId,
          page: this.currentPage,
          limit: this.pageSize
        }
        if (this.searchKeyword) {
          params.keyword = this.searchKeyword
        }
        if (this.filterType) {
          params.type = this.filterType
        }

        const res = await this.$store.dispatch('classroom/getQuestionBank', params)
        if (res.code === 200) {
          const questions = res.data.questions || res.data || []
          this.questionBank = questions.map(q => ({
            ...q,
            difficulty: parseInt(q.difficulty) || 2, // 确保是数字类型
            score: q.score || 10 // 默认分值
          }))
          this.total = res.data.total || questions.length
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.questionsLoading = false
      }
    },
    handleSelectionChange(selection) {
      // 不再使用勾选框，此方法保留但不做任何操作
    },
    isQuestionSelected(question) {
      // 检查题目是否已添加（通过id或problemId）
      return this.selectedQuestions.some(sq =>
        (sq.id === question.id) ||
        (sq.problemId && sq.problemId === question.problemId)
      )
    },
    addQuestion(question) {
      // 检查是否已添加
      if (this.isQuestionSelected(question)) {
        this.$message.warning('该题目已添加')
        return
      }

      // 添加题目（深拷贝，避免修改原数据）
      const newQuestion = JSON.parse(JSON.stringify(question))

      // 设置默认分数（如果没有分数或分数为0）
      if (!newQuestion.score || newQuestion.score === 0) {
        // 根据题型设置默认分数
        const defaultScores = {
          single_choice: 2,      // 单选题默认2分
          multiple_choice: 5,    // 多选题默认5分
          judge: 1,              // 判断题默认1分
          subjective: 5,         // 主观题默认5分
          programming: 20        // 编程题默认20分
        }
        newQuestion.score = defaultScores[newQuestion.type] || 10
      }

      this.selectedQuestions.push(newQuestion)

      this.$message.success('添加成功')
    },
    removeQuestion(question) {
      const index = this.selectedQuestions.findIndex(q => q.id === question.id)
      if (index > -1) {
        this.selectedQuestions.splice(index, 1)
      }
    },
    moveUp(index) {
      if (index > 0) {
        const temp = this.selectedQuestions[index - 1]
        this.$set(this.selectedQuestions, index - 1, this.selectedQuestions[index])
        this.$set(this.selectedQuestions, index, temp)
      }
    },
    moveDown(index) {
      if (index < this.selectedQuestions.length - 1) {
        const temp = this.selectedQuestions[index + 1]
        this.$set(this.selectedQuestions, index + 1, this.selectedQuestions[index])
        this.$set(this.selectedQuestions, index, temp)
      }
    },
    viewQuestion(question) {
      // 深拷贝题目对象，确保不影响原数据
      this.currentQuestion = {
        ...question,
        difficulty: parseInt(question.difficulty) || 2 // 确保是数字类型
      }
      this.showDetailDialog = true
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        subjective: this.$t('m.Subjective'),
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    // 将数据库中的难度值（1-3）转换为 el-rate 的星星数（1-5）
    getDifficultyStars(difficulty) {
      // 数据库: 1=简单, 2=中等, 3=困难
      // 显示: 映射为 1星, 3星, 5星
      const difficultyMap = {
        1: 1, // 简单 -> 1星
        2: 3, // 中等 -> 3星
        3: 5  // 困难 -> 5星
      }
      return difficultyMap[difficulty] || 3 // 默认3星
    },
    renderOptions(options) {
      if (!options) return '-'
      try {
        const opts = JSON.parse(options)
        return opts.map(opt => {
          // Render each option with markdown
          return this.renderMarkdown(opt)
        }).join('<br>')
      } catch (e) {
        return options.replace(/\n/g, '<br>')
      }
    },
    handlePageChange(page) {
      this.currentPage = page
      this.loadQuestionBank()
    },
    async loadHomeworkData() {
      try {
        console.log('开始加载作业数据, editId:', this.editId)
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', this.editId)
        console.log('getHomeworkDetail 响应:', res)
        if (res.code === 200 && res.data) {
          const homework = res.data
          console.log('作业数据:', homework)
          console.log('题目列表:', homework.questions)

          // 先加载所有编程题的详情
          const programmingProblemIds = homework.questions
            .filter(q => q.problemId)
            .map(q => q.problemId)

          if (programmingProblemIds.length > 0) {
            console.log('预加载编程题数据:', programmingProblemIds)
            await this.loadProgrammingProblemByIds(programmingProblemIds)
          }

          // 填充表单数据
          this.form = {
            title: homework.title,
            description: homework.description,
            startTime: new Date(homework.startTime),
            endTime: new Date(homework.endTime),
            showHomework: homework.showHomework === 1,
            showScore: homework.showScore === 1,
            showAnswer: homework.showAnswer === 1
          }
          // 填充已选题目
          if (homework.questions && homework.questions.length > 0) {
            this.selectedQuestions = homework.questions.map((item, index) => {
              console.log(`处理题目 ${index + 1}:`, item)
              // 编程题：使用 problemId
              if (item.problemId) {
                console.log('  -> 这是编程题')
                // 从缓存获取编程题详情（现在已经预加载过了）
                const cached = this.programmingProblemsCache[item.problemId]
                console.log('  -> 缓存数据:', cached)
                const score = item.score || item.question?.score || 20
                if (cached && cached.problem) {
                  console.log('  -> 使用缓存数据')
                  return {
                    id: item.id,
                    problemId: item.problemId,
                    title: cached.problem.title,
                    type: 'programming',
                    difficulty: 5,
                    score: score,
                    content: cached.problem.description || '',
                    input: cached.problem.input || '',
                    output: cached.problem.output || '',
                    examples: cached.problem.examples || '',
                    hint: cached.problem.hint || '',
                    judgeMode: cached.problem.judgeMode || '',
                    timeLimit: cached.problem.timeLimit || 0,
                    memoryLimit: cached.problem.memoryLimit || 0
                  }
                }
                // 如果缓存中没有，返回基本信息（不影响编辑）
                console.log('  -> 使用基本信息（无缓存）')
                return {
                  id: item.id,
                  problemId: item.problemId,
                  title: `HOJ 编程题 - ${item.problemId}`,
                  type: 'programming',
                  difficulty: 5,
                  score: score
                }
              }
              // 普通题目：使用 question
              if (item.question && item.question.id) {
                console.log('  -> 这是普通题目')
                return {
                  id: item.question.id,
                  title: item.question.title,
                  type: item.question.type,
                  difficulty: parseInt(item.question.difficulty) || 2,
                  score: item.score || item.question.score || 10,
                  content: item.question.content,
                  options: item.question.options,
                  answer: item.question.answer
                }
              }
              // 兜底：如果既没有 problemId 也没有 question，跳过
              console.warn('  -> 跳过无效题目:', item)
              return null
            }).filter(q => q !== null)
            console.log('最终 selectedQuestions:', this.selectedQuestions)
          }
        }
      } catch (error) {
        console.error('加载作业数据失败:', error)
        this.$message.error('加载作业数据失败: ' + (error.message || error))
      }
    },
    async saveHomework() {
      if (!this.$refs.homeworkForm) {
        this.$message.error('表单未初始化')
        return
      }

      this.$refs.homeworkForm.validate(async (valid) => {
        if (!valid) {
          this.$message.warning('请填写所有必填项')
          return
        }
        if (this.selectedQuestions.length === 0) {
          this.$message.warning(this.$t('m.Please_Select_At_Least_One_Question'))
          return
        }

        this.submitting = true
        const data = {
          classroomId: Number(this.classroomId),
          title: this.form.title,
          description: this.form.description || '',
          startTime: moment(this.form.startTime).format(),
          endTime: moment(this.form.endTime).format(),
          showHomework: this.form.showHomework ? 1 : 0,
          showScore: this.form.showScore ? 1 : 0,
          showAnswer: this.form.showAnswer ? 1 : 0,
          questions: this.selectedQuestions.map((q, index) => {
            const result = {
              score: Number(q.score) || 10,
              questionOrder: index + 1,
              questionType: q.type // 添加题目类型，用于后端设置默认分数
            }
            // 编程题使用 problemId,其他题型使用 questionId
            if (q.type === 'programming' && q.problemId) {
              result.problemId = q.problemId
              result.questionId = null
            } else {
              result.questionId = q.id
              result.problemId = null
            }
            return result
          })
        }

        // 调试：打印提交的数据
        console.log('=== 提交作业数据 ===')
        console.log('完整数据:', JSON.stringify(data, null, 2))
        console.log('questions 数量:', data.questions.length)
        data.questions.forEach((q, idx) => {
          console.log(`题目${idx + 1}:`, JSON.stringify(q, null, 2))
        })

        try {
          let res
          if (this.isEditMode) {
            // 编辑模式：调用更新接口
            data.id = Number(this.editId)
            res = await this.$store.dispatch('classroom/updateHomework', data)
          } else {
            // 创建模式：调用创建接口
            res = await this.$store.dispatch('classroom/createHomework', data)
          }

          console.log('=== 后端响应 ===')
          console.log('完整响应:', res)

          if (res.code === 200) {
            this.$message.success(this.isEditMode ? this.$t('m.Update_Success') : this.$t('m.Create_Success'))
            this.goBack()
          } else {
            this.$message.error(res.message || (this.isEditMode ? '保存失败' : '创建失败'))
          }
        } catch (error) {
          console.error('=== 请求异常 ===')
          console.error('错误详情:', error)
          this.$message.error(this.isEditMode ? '保存失败' : '创建失败')
        } finally {
          this.submitting = false
        }
      })
    },
    // 获取 HOJ 题目信息
    async fetchProgrammingProblemInfo() {
      console.log('=== 开始获取 HOJ 题目信息 ===')
      console.log('题目ID:', this.programmingForm.problemId)

      if (!this.programmingForm.problemId) {
        this.$message.warning('请输入 HOJ 题目 ID')
        return
      }

      // 获取当前登录用户信息
      const userInfo = this.$store.getters.userInfo
      console.log('当前用户信息:', userInfo)

      if (!userInfo || !userInfo.username) {
        this.$message.warning('请先登录')
        return
      }

      // 从 localStorage 获取 token
      const token = localStorage.getItem('token')
      console.log('当前 Token:', token ? token.substring(0, 20) + '...' : '无')

      if (!token) {
        this.$message.warning('未找到登录凭证，请重新登录')
        return
      }

      this.fetchingProblem = true
      try {
        const requestData = {
          pid: this.programmingForm.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token, // 使用 token 而不是密码
          password: ''
        }
        console.log('请求HOJ API, 参数:', { ...requestData, token: token.substring(0, 20) + '...' })

        const res = await getJudgeInfo(requestData)
        console.log('HOJ API 响应:', res)

        if (res.code === 200 && res.data) {
          this.programmingProblemPreview = res.data
          this.extractProgrammingExamples()
          this.$message.success('题目获取成功')
        } else {
          this.$message.error(res.message || '获取题目失败')
        }
      } catch (error) {
        console.error('获取题目异常:', error)
        this.$message.error('网络错误，获取题目失败')
      } finally {
        this.fetchingProblem = false
      }
    },
    // 提取样例
    extractProgrammingExamples() {
      if (!this.programmingProblemPreview || !this.programmingProblemPreview.problem.examples) {
        this.programmingExamples = []
        return
      }

      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      const examples = []
      let match

      while ((match = regex.exec(this.programmingProblemPreview.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      this.programmingExamples = examples
    },
    // 确认添加编程题
    confirmAddProgrammingQuestion() {
      if (!this.programmingForm.problemId) {
        this.$message.warning('请输入 HOJ 题目 ID')
        return
      }

      // 必须先预览题目
      if (!this.programmingProblemPreview) {
        this.$message.warning('请先点击"获取题目信息"按钮预览题目')
        return
      }

      // 使用预览获取的真实题目信息
      const tempQuestion = {
        id: `hoj_${this.programmingForm.problemId}`, // 临时ID,使用 hoj_ 前缀
        problemId: this.programmingForm.problemId, // HOJ 题目ID,保持字符串类型
        title: this.programmingProblemPreview.problem.title, // 使用真实标题
        type: 'programming',
        difficulty: 5, // 默认难度
        score: this.programmingForm.score,
        content: `HOJ 题目 ID: ${this.programmingForm.problemId}`
      }

      // 检查是否已经添加过该题目
      const exists = this.selectedQuestions.some(q => q.problemId === this.programmingForm.problemId)
      if (exists) {
        this.$message.warning('该题目已添加')
        return
      }

      // 添加到已选题目
      this.selectedQuestions.push(tempQuestion)
      this.$message.success('添加成功')

      // 关闭对话框并重置
      this.showAddProgrammingDialog = false
      this.resetProgrammingForm()
    },
    // 重置编程题表单
    resetProgrammingForm() {
      this.programmingForm = {
        problemId: '',
        score: 10
      }
      this.programmingProblemPreview = null
      this.programmingExamples = []
    },
    // 渲染 Markdown
    renderMarkdown(text) {
      if (!text) return ''
      try {
        const rendered = md.render(text)
        console.log('渲染 Markdown，输入长度:', text.length, '输出长度:', rendered.length)
        return rendered
      } catch (e) {
        console.error('Markdown 渲染失败:', e)
        return text
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
    // 全卷预览相关方法
    async showFullPreview() {
      if (this.selectedQuestions.length === 0) {
        this.$message.warning('请先添加题目')
        return
      }

      // 检查是否有编程题需要加载
      const hasProgrammingQuestions = this.selectedQuestions.some(q => q.type === 'programming' && q.problemId)
      if (hasProgrammingQuestions) {
        const loading = this.$loading({
          lock: true,
          text: '正在加载编程题数据...',
          spinner: 'el-icon-loading',
          background: 'rgba(0, 0, 0, 0.7)'
        })

        try {
          // 先加载所有编程题的信息，等待加载完成后再显示预览
          await this.loadProgrammingProblemsForPreview()
        } finally {
          loading.close()
        }
      }

      this.showFullPreviewDialog = true
    },
    handlePreviewOpened() {
      // 对话框打开后，确保 KaTeX 样式已加载
      console.log('预览对话框已打开，编程题缓存:', this.programmingProblemsCache)
      // 不需要强制刷新，Vue 的响应式系统会自动处理
    },
    async loadProgrammingProblemsForPreview() {
      // 获取所有编程题的 problemId
      const programmingQuestions = this.selectedQuestions.filter(q => q.type === 'programming' && q.problemId)
      const problemIds = programmingQuestions.map(q => q.problemId)

      await this.loadProgrammingProblemByIds(problemIds)
    },
    // 批量加载编程题数据（用于编辑和预览）
    async loadProgrammingProblemByIds(problemIds) {
      if (!problemIds || problemIds.length === 0) return

      const userInfo = this.$store.getters.userInfo
      const token = localStorage.getItem('token')

      if (!userInfo || !token) {
        console.warn('未登录，无法加载编程题数据')
        return
      }

      console.log('开始批量加载编程题:', problemIds)

      // 过滤掉已缓存的
      const uncachedIds = problemIds.filter(id => !this.programmingProblemsCache[id])
      console.log('需要加载的编程题:', uncachedIds)

      for (const problemId of uncachedIds) {
        try {
          console.log('加载编程题:', problemId)
          const res = await getJudgeInfo({
            pid: problemId,
            cid: '0',
            mode: 'normal',
            username: userInfo.username,
            token: token,
            password: ''
          })

          if (res.code === 200 && res.data) {
            console.log('编程题加载成功:', problemId)
            console.log('题目描述包含数学公式:', res.data.problem.description.includes('$'))
            this.$set(this.programmingProblemsCache, problemId, res.data)
          }
        } catch (error) {
          console.error('加载编程题信息失败:', problemId, error)
        }
      }
    },
    getProgrammingProblemInfo(problemId) {
      return this.programmingProblemsCache[problemId] || null
    },
    getProgrammingExamplesForProblem(problemId) {
      const problemInfo = this.programmingProblemsCache[problemId]
      if (!problemInfo || !problemInfo.problem || !problemInfo.problem.examples) {
        return []
      }

      // 解析样例
      const examples = []
      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      let match

      while ((match = regex.exec(problemInfo.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      return examples
    },
    getTotalScore() {
      return this.selectedQuestions.reduce((sum, q) => sum + (q.score || 0), 0)
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        return options.map((opt, idx) => ({
          letter: String.fromCharCode(65 + idx), // A, B, C, D...
          text: opt
        }))
      } catch (e) {
        return []
      }
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },

    goBack() {
      this.$router.go(-1)
    }
  }
}
</script>

<style scoped>
.create-homework {
  padding: 20px;
  background-color: #fff;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #DCDFE6;
}

.header h2 {
  margin: 0;
  color: #409EFF;
}

.homework-form {
  max-width: 1200px;
}

.form-section {
  margin-bottom: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.question-list {
  min-height: 300px;
}

.question-detail {
  padding: 10px;
}

/* 编程题预览样式 */
.problem-preview {
  margin: 20px 0;
}

.problem-preview h3 {
  margin-top: 0;
  color: #303133;
  font-size: 18px;
}

.problem-meta {
  margin: 15px 0;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.problem-content {
  margin-top: 20px;
}

.content-section {
  margin-bottom: 20px;
}

.content-section h4 {
  color: #409EFF;
  font-size: 16px;
  margin-bottom: 10px;
  border-left: 3px solid #409EFF;
  padding-left: 10px;
}

.content-section pre {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 10px 0;
}

.example-item {
  margin-bottom: 15px;
}

.example-item pre {
  margin: 8px 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 全卷预览样式 */
.full-preview-container {
  max-height: 70vh;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 10px;
}

/* 确保对话框内容可以正常滚动 */
.full-preview-container >>> .el-card {
  margin-bottom: 20px;
  overflow: visible;
}

.full-preview-container >>> .question-item {
  overflow: visible;
}

.preview-info {
  margin-bottom: 20px;
  background-color: #f5f7fa;
}

.preview-info h3 {
  margin: 0 0 15px 0;
  color: #409EFF;
  font-size: 20px;
}

.preview-info p {
  margin: 8px 0;
  color: #606266;
}

.questions-preview {
  margin-top: 20px;
}

.question-item {
  padding: 20px;
  margin-bottom: 20px;
  background-color: #fff;
  border: 1px solid #DCDFE6;
  border-radius: 4px;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 2px solid #E4E7ED;
}

.question-number {
  font-size: 18px;
  font-weight: bold;
  color: #409EFF;
}

.question-type {
  color: #909399;
  font-size: 14px;
}

.question-score {
  margin-left: auto;
  color: #E6A23C;
  font-weight: bold;
}

.question-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 15px;
  color: #303133;
}

.question-content {
  margin: 15px 0;
  color: #606266;
  line-height: 1.8;
  font-size: 14px;
}

.question-options {
  margin-top: 15px;
  padding-left: 20px;
}

.option-item {
  margin-bottom: 10px;
}

.option-preview {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  background-color: #F5F7FA;
  border-radius: 4px;
  cursor: default;
}

.option-preview:hover {
  background-color: #ECF5FF;
}

.option-letter {
  display: inline-block;
  min-width: 30px;
  font-weight: bold;
  color: #409EFF;
  font-size: 15px;
}

.option-text {
  flex: 1;
  color: #606266;
  line-height: 1.6;
}

.subjective-preview,
.programming-preview {
  margin-top: 15px;
  padding: 15px;
  background-color: #FDF6EC;
  border-left: 3px solid #E6A23C;
  border-radius: 4px;
}

.programming-detail {
  margin-top: 15px;
  padding: 15px;
  background-color: #fff;
  border-radius: 4px;
  border: 1px solid #DCDFE6;
}

.programming-meta {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.programming-content h4 {
  color: #409EFF;
  font-size: 16px;
  margin-bottom: 10px;
  border-left: 3px solid #409EFF;
  padding-left: 10px;
}

.programming-content {
  /* 确保长内容不会破坏布局 */
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.programming-content >>> * {
  max-width: 100%;
  overflow-x: auto;
}

/* 特别处理 pre/code 块，防止内容溢出 */
.programming-content >>> pre,
.programming-content >>> code {
  white-space: pre-wrap !important;
  word-break: break-word !important;
  max-width: 100% !important;
  overflow-x: auto !important;
}

/* 确保 Markdown 渲染的数学公式等不会撑破容器 */
.programming-content >>> .katex,
.programming-content >>> .katex-display {
  max-width: 100% !important;
  overflow-x: auto !important;
}

/* 确保 table 不会撑破容器 */
.programming-content >>> table {
  max-width: 100% !important;
  overflow-x: auto !important;
  display: block !important;
}
</style>

<!-- 引入 KaTeX 样式 -->
<style>
@import '~katex/dist/katex.min.css';
</style>
