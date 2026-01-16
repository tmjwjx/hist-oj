<template>
  <div class="question-bank-panel">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h3>{{ $t('m.Question_Bank') }}</h3>
    </div>
    <div class="action-bar">
      <el-button type="primary" icon="el-icon-plus" @click="showCreateDialog = true">
        {{ $t('m.Create_Question') }}
      </el-button>
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-table :data="questions" stripe>
      <el-table-column prop="title" :label="$t('m.Question_Title')">
        <template slot-scope="{ row }">
          <div v-html="renderMarkdown(row.title)" class="markdown-body"></div>
        </template>
      </el-table-column>
      <el-table-column prop="type" :label="$t('m.Question_Type')" width="100">
        <template slot-scope="{ row }">
          <el-tag :type="getQuestionTypeColor(row.type)">
            {{ getQuestionTypeName(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="difficulty" :label="$t('m.Difficulty')" width="100">
        <template slot-scope="{ row }">
          <el-rate :value="getDifficultyStars(row.difficulty)" disabled />
        </template>
      </el-table-column>
      <el-table-column prop="isShared" :label="$t('m.Shared')" width="80">
        <template slot-scope="{ row }">
          <el-tag :type="row.isShared ? 'success' : 'info'">
            {{ row.isShared ? $t('m.Yes') : $t('m.No') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Operation')" width="200">
        <template slot-scope="{ row }">
          <el-button size="small" @click="handleEdit(row)">{{ $t('m.Edit') }}</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">
            {{ $t('m.Delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="$t('m.Create_Question')" :visible.sync="showCreateDialog" width="1200px">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form :model="createForm" ref="createForm" label-width="120px">
            <el-form-item :label="$t('m.Question_Type')" prop="type">
              <el-select v-model="createForm.type" @change="handleTypeChange">
                <el-option :label="$t('m.Single_Choice')" value="single_choice" />
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
                <el-option :label="$t('m.Judge')" value="judge" />
                <el-option :label="$t('m.Subjective')" value="subjective" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('m.Question_Title')" prop="title">
              <el-input v-model="createForm.title" />
            </el-form-item>
            <el-form-item :label="$t('m.Content')" prop="content">
              <el-input type="textarea" v-model="createForm.content" :rows="4" />
            </el-form-item>

        <!-- 单选题：固定4个选项，单选 -->
        <template v-if="createForm.type === 'single_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="option-item">
                <el-radio v-model="createForm.correctAnswer" :label="index" class="option-radio">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-radio>
                <el-input v-model="createForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
            <div class="answer-tip">
              <i class="el-icon-info"></i>
              {{ $t('m.Select_Correct_Answer_Tip') }}
            </div>
          </el-form-item>
        </template>

        <!-- 多选题：固定4个选项，多选 -->
        <template v-if="createForm.type === 'multiple_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="option-item">
                <el-checkbox v-model="createForm.correctAnswers[index]" :label="index" class="option-checkbox">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-checkbox>
                <el-input v-model="createForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
            <div class="answer-tip">
              <i class="el-icon-info"></i>
              {{ $t('m.Select_Correct_Answers_Tip') }}
            </div>
          </el-form-item>
        </template>

        <!-- 判断题：选择正确/错误 -->
        <template v-if="createForm.type === 'judge'">
          <el-form-item :label="$t('m.Correct_Answer')" required>
            <el-radio-group v-model="createForm.correctAnswer">
              <el-radio label="true">{{ $t('m.True') }}</el-radio>
              <el-radio label="false">{{ $t('m.False') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <!-- 主观题：需要人工打分 -->
        <template v-if="createForm.type === 'subjective'">
          <el-form-item :label="$t('m.Reference_Answer')">
            <el-input type="textarea" v-model="createForm.referenceAnswer" :rows="4" :placeholder="$t('m.Reference_Answer_Tip')" />
          </el-form-item>
        </template>

        <el-form-item :label="$t('m.Difficulty')" prop="difficulty">
          <el-rate v-model="createForm.difficulty" :max="3" />
        </el-form-item>
        <el-form-item :label="$t('m.Score')" prop="score">
          <el-input-number v-model="createForm.score" :min="1" :max="100" />
        </el-form-item>
        <el-form-item :label="$t('m.Share_To_Question_Pool')">
          <el-switch v-model="createForm.isShared" />
          <div class="form-tip">
            <i class="el-icon-info"></i>
            {{ $t('m.Share_To_Question_Pool_Tip') }}
          </div>
        </el-form-item>
      </el-form>
        </el-col>
        <el-col :span="12">
          <el-card class="preview-card markdown-preview">
            <div slot="header">
              <i class="el-icon-view"></i> 实时预览
            </div>
            <div class="preview-content">
              <div v-if="createForm.title" v-html="renderMarkdown(createForm.title)" class="markdown-body preview-title"></div>
              <p v-else class="preview-placeholder">题目标题预览</p>

              <div v-if="createForm.content" v-html="renderMarkdown(createForm.content)" class="markdown-body preview-content-text"></div>
              <p v-else class="preview-placeholder">题目内容预览</p>

              <!-- 单选题选项预览 -->
              <!-- 单选题选项预览 -->
              <div v-if="createForm.type === 'single_choice'" class="preview-options">
                <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body"></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 多选题选项预览 -->
              <div v-if="createForm.type === 'multiple_choice'" class="preview-options">
                <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="createForm.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body"></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 判断题预览 -->
              <div v-if="createForm.type === 'judge'" class="preview-options">
                <div class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                  <span>{{ $t('m.True') }}</span>
                </div>
                <div class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                  <span>{{ $t('m.False') }}</span>
                </div>
              </div>

              <!-- 主观题预览 -->
              <div v-if="createForm.type === 'subjective'" class="preview-subjective">
                <el-alert type="info" :closable="false">
                  <i class="el-icon-edit"></i> 主观题，学生需要输入文字答案
                </el-alert>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createQuestion">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 编辑题目对话框 -->
    <el-dialog :title="$t('m.Edit_Question')" :visible.sync="showEditDialog" width="1200px">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form :model="editForm" ref="editForm" label-width="120px">
            <el-form-item :label="$t('m.Question_Type')" prop="type">
              <el-select v-model="editForm.type" @change="handleEditTypeChange">
                <el-option :label="$t('m.Single_Choice')" value="single_choice" />
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
                <el-option :label="$t('m.Judge')" value="judge" />
                <el-option :label="$t('m.Subjective')" value="subjective" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('m.Question_Title')" prop="title">
              <el-input v-model="editForm.title" />
            </el-form-item>
            <el-form-item :label="$t('m.Content')" prop="content">
              <el-input type="textarea" v-model="editForm.content" :rows="4" />
            </el-form-item>

        <!-- 单选题 -->
        <template v-if="editForm.type === 'single_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                <el-radio v-model="editForm.correctAnswer" :label="index" class="option-radio">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-radio>
                <el-input v-model="editForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- 多选题 -->
        <template v-if="editForm.type === 'multiple_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                <el-checkbox v-model="editForm.correctAnswers[index]" :label="index" class="option-checkbox">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-checkbox>
                <el-input v-model="editForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- 判断题 -->
        <template v-if="editForm.type === 'judge'">
          <el-form-item :label="$t('m.Correct_Answer')" required>
            <el-radio-group v-model="editForm.correctAnswer">
              <el-radio label="true">{{ $t('m.True') }}</el-radio>
              <el-radio label="false">{{ $t('m.False') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <!-- 主观题 -->
        <template v-if="editForm.type === 'subjective'">
          <el-form-item :label="$t('m.Reference_Answer')">
            <el-input type="textarea" v-model="editForm.referenceAnswer" :rows="4" :placeholder="$t('m.Reference_Answer_Tip')" />
          </el-form-item>
        </template>

        <el-form-item :label="$t('m.Difficulty')" prop="difficulty">
          <el-rate v-model="editForm.difficulty" :max="3" />
        </el-form-item>
        <el-form-item :label="$t('m.Score')" prop="score">
          <el-input-number v-model="editForm.score" :min="1" :max="100" />
        </el-form-item>
        <el-form-item :label="$t('m.Share_To_Question_Pool')">
          <el-switch v-model="editForm.isShared" />
          <div class="form-tip">
            <i class="el-icon-info"></i>
            {{ $t('m.Share_To_Question_Pool_Tip') }}
          </div>
        </el-form-item>
      </el-form>
        </el-col>
        <el-col :span="12">
          <el-card class="preview-card markdown-preview">
            <div slot="header">
              <i class="el-icon-view"></i> 实时预览
            </div>
            <div class="preview-content">
              <div v-if="editForm.title" v-html="renderMarkdown(editForm.title)" class="markdown-body preview-title"></div>
              <p v-else class="preview-placeholder">题目标题预览</p>

              <div v-if="editForm.content" v-html="renderMarkdown(editForm.content)" class="markdown-body preview-content-text"></div>
              <p v-else class="preview-placeholder">题目内容预览</p>

              <!-- 单选题选项预览 -->
              <div v-if="editForm.type === 'single_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body"></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 多选题选项预览 -->
              <div v-if="editForm.type === 'multiple_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body"></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 判断题预览 -->
              <div v-if="editForm.type === 'judge'" class="preview-options">
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                  <span>{{ $t('m.True') }}</span>
                </div>
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                  <span>{{ $t('m.False') }}</span>
                </div>
              </div>

              <!-- 主观题预览 -->
              <div v-if="editForm.type === 'subjective'" class="preview-subjective">
                <el-alert type="info" :closable="false">
                  <i class="el-icon-edit"></i> 主观题，学生需要输入文字答案
                </el-alert>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="showEditDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="updateQuestion">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import realtimeSync from '@/mixins/realtimeSync'

// 配置 markdown-it 支持 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex)

export default {
  name: 'QuestionBank',
  mixins: [realtimeSync],
  props: {
    classroomId: {
      type: [String, Number],
      required: false
    }
  },
  data() {
    return {
      loading: false,
      questions: [],
      showCreateDialog: false,
      showEditDialog: false,
      currentEditId: null,
      createForm: {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''], // 4个选项的内容
        correctAnswer: 0, // 单选题正确答案（0-3）
        correctAnswers: [false, false, false, false], // 多选题正确答案（数组）
        referenceAnswer: '', // 主观题参考答案
        difficulty: 1,
        score: 2, // 单选题默认2分
        isShared: false
      },
      editForm: {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        difficulty: 1,
        score: 2, // 单选题默认2分
        isShared: false
      },
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadQuestions',
        immediate: true
      }
    }
  },
  mounted() {
    // 由 realtimeSync mixin 自动启动同步
  },
  methods: {
    handleTypeChange() {
      this.handleFormTypeChange(this.createForm)
    },
    handleEditTypeChange() {
      this.handleFormTypeChange(this.editForm)
    },
    handleFormTypeChange(form) {
      // 切换题目类型时重置答案相关字段和默认分数
      if (form.type === 'single_choice') {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
        form.score = 2 // 单选题默认2分
      } else if (form.type === 'multiple_choice') {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
        form.score = 5 // 多选题默认5分
      } else if (form.type === 'judge') {
        form.score = 1 // 判断题默认1分
        form.correctAnswer = 'true'
        form.correctAnswers = [false, false, false, false]
      } else if (form.type === 'subjective') {
        form.score = 5 // 主观题默认5分
        form.correctAnswer = ''
        form.correctAnswers = [false, false, false, false]
      } else {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
      }
      form.referenceAnswer = ''
    },
    async loadQuestions() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.questions.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getQuestionBank', {
          classroomId: this.classroomId,
          page: 1,
          limit: 50
        })
        if (res.code === 200) {
          const newQuestions = res.data.questions || res.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.questions)
          const newDataString = JSON.stringify(newQuestions)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.questions = newQuestions
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    async createQuestion() {
      // 验证必填字段
      if (!this.createForm.title || !this.createForm.title.trim()) {
        this.$message.warning('请输入题目标题')
        return
      }
      if (!this.createForm.content || !this.createForm.content.trim()) {
        this.$message.warning('请输入题目内容')
        return
      }

      // 构建提交数据
      const submitData = {
        classroomId: this.classroomId,
        type: this.createForm.type,
        title: this.createForm.title,
        content: this.createForm.content,
        difficulty: this.createForm.difficulty || 1,
        score: this.createForm.score || 2,
        isShared: this.createForm.isShared ? 1 : 0
      }

      // 根据题型设置答案格式
      if (this.createForm.type === 'single_choice') {
        // 单选题
        const optionsArray = this.createForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        submitData.answer = ['A', 'B', 'C', 'D'][this.createForm.correctAnswer]
      } else if (this.createForm.type === 'multiple_choice') {
        // 多选题
        const optionsArray = this.createForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        // 多选题答案：保存为 JSON 数组格式，如 ["A","B","C"]
        const selectedAnswers = this.createForm.correctAnswers
          .map((selected, idx) => selected ? ['A', 'B', 'C', 'D'][idx] : null)
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.createForm.type === 'judge') {
        // 判断题 - 确保 correctAnswer 是字符串类型
        const answerValue = String(this.createForm.correctAnswer)
        submitData.answer = answerValue === 'true' ? '正确' : '错误'
        submitData.options = null // 判断题不需要选项
      } else if (this.createForm.type === 'subjective') {
        // 主观题
        submitData.answer = this.createForm.referenceAnswer || '需人工评分'
        submitData.options = null // 主观题不需要选项
      }

      try {
        const res = await this.$store.dispatch('classroom/createQuestion', submitData)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Create_Success'))
          this.showCreateDialog = false
          this.resetForm()
          this.loadQuestions()
        } else {
          this.$message.error(res.message || this.$t('m.Create_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Create_Failed'))
      }
    },
    resetForm() {
      this.createForm = {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        difficulty: 1,
        score: 10,
        isShared: false
      }
    },
    handleEdit(question) {
      this.currentEditId = question.id
      // 解析题目数据并填充到编辑表单
      this.editForm = {
        type: question.type,
        title: question.title,
        content: question.content || '',
        difficulty: question.difficulty,
        score: question.score,
        isShared: question.isShared === 1,
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: ''
      }

      // 解析选项和答案
      if (question.type === 'single_choice' || question.type === 'multiple_choice') {
        // 解析选项 JSON
        if (question.options) {
          try {
            const optionsArray = JSON.parse(question.options)
            this.editForm.choiceOptions = optionsArray.map(opt => {
              // 去掉 "A. " 这样的前缀
              return opt.replace(/^[A-D]\.\s*/, '')
            })
          } catch (e) {
            console.error('解析选项失败', e)
          }
        }

        // 解析答案
        if (question.type === 'single_choice') {
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswer = answerMap[question.answer] || 0
        } else {
          // 多选题
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswers = [false, false, false, false]

          // 尝试解析 JSON 数组格式
          try {
            const answers = JSON.parse(question.answer)
            if (Array.isArray(answers)) {
              answers.forEach(ans => {
                const idx = answerMap[ans]
                if (idx !== undefined) {
                  this.editForm.correctAnswers[idx] = true
                }
              })
            }
          } catch (e) {
            // 兼容旧的逗号分隔格式
            const answers = question.answer.split(',')
            answers.forEach(ans => {
              const idx = answerMap[ans.trim()]
              if (idx !== undefined) {
                this.editForm.correctAnswers[idx] = true
              }
            })
          }
        }
      } else if (question.type === 'judge') {
        this.editForm.correctAnswer = question.answer === '正确' ? 'true' : 'false'
      } else if (question.type === 'subjective') {
        this.editForm.referenceAnswer = question.answer || ''
      }

      this.showEditDialog = true
    },
    async updateQuestion() {
      // 验证必填字段
      if (!this.editForm.title || !this.editForm.title.trim()) {
        this.$message.warning('请输入题目标题')
        return
      }
      if (!this.editForm.content || !this.editForm.content.trim()) {
        this.$message.warning('请输入题目内容')
        return
      }

      const submitData = {
        type: this.editForm.type,
        title: this.editForm.title,
        content: this.editForm.content,
        difficulty: this.editForm.difficulty || 1,
        score: this.editForm.score || 2,
        isShared: this.editForm.isShared ? 1 : 0
      }

      // 根据题型设置答案格式
      if (this.editForm.type === 'single_choice') {
        const optionsArray = this.editForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        submitData.answer = ['A', 'B', 'C', 'D'][this.editForm.correctAnswer]
      } else if (this.editForm.type === 'multiple_choice') {
        const optionsArray = this.editForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        // 多选题答案：保存为 JSON 数组格式，如 ["A","B","C"]
        const selectedAnswers = this.editForm.correctAnswers
          .map((selected, idx) => selected ? ['A', 'B', 'C', 'D'][idx] : null)
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.editForm.type === 'judge') {
        // 判断题 - 确保 correctAnswer 是字符串类型
        const answerValue = String(this.editForm.correctAnswer)
        submitData.answer = answerValue === 'true' ? '正确' : '错误'
        submitData.options = null // 判断题不需要选项
      } else if (this.editForm.type === 'subjective') {
        submitData.answer = this.editForm.referenceAnswer || '需人工评分'
        submitData.options = null // 主观题不需要选项
      }

      try {
        const res = await this.$store.dispatch('classroom/updateQuestion', {
          questionId: this.currentEditId,
          data: submitData
        })
        if (res.code === 200) {
          this.$message.success(this.$t('m.Update_Success'))
          this.showEditDialog = false
          this.loadQuestions()
        } else {
          this.$message.error(res.message || this.$t('m.Update_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Update_Failed'))
      }
    },
    async handleDelete(question) {
      this.$confirm(this.$t('m.Confirm_Delete_Question'), this.$t('m.Warning'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteQuestion', question.id)
          if (res.code === 200) {
            this.$message.success(this.$t('m.Delete_Success'))
            this.loadQuestions()
          } else {
            this.$message.error(res.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
        }
      })
    },
    getQuestionTypeName(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        subjective: this.$t('m.Subjective')
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
    getQuestionTypeColor(type) {
      const map = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        subjective: 'info'
      }
      return map[type] || ''
    },
    goBack() {
      this.$router.push({ name: 'TeacherDashboard' })
    },
    renderMarkdown(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return content
      }
    }
  }
}
</script>

<style scoped>
.question-bank-panel {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h3 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.action-bar {
  margin-bottom: 20px;
}

.options-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.option-radio {
  margin-right: 8px;
  font-weight: bold;
  font-size: 14px;
  min-width: 20px;
}

.option-checkbox {
  margin-right: 8px;
  font-weight: bold;
  font-size: 14px;
  min-width: 20px;
}

.answer-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.answer-tip i {
  font-size: 14px;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.form-tip i {
  font-size: 14px;
}

/* 预览卡片样式 */
.preview-card {
  height: 100%;
  border: 2px solid #E4E7ED;
}

.preview-content {
  padding: 10px;
}

.preview-title {
  color: #303133;
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 15px;
  border-bottom: 2px solid #E4E7ED;
  padding-bottom: 10px;
}

.preview-content-text {
  color: #606266;
  line-height: 1.8;
  margin-bottom: 20px;
  font-size: 14px;
}

.preview-placeholder {
  color: #C0C4CC;
  font-style: italic;
}

.preview-options {
  margin-top: 15px;
}

.preview-option-item {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  margin-bottom: 10px;
  background: #F5F7FA;
  border-radius: 4px;
  gap: 8px;
}

.preview-option-item .el-tag {
  flex-shrink: 0;
  margin-right: 0;
  padding: 0 6px !important;
  height: 20px !important;
  line-height: 20px !important;
  font-size: 12px !important;
  width: 24px !important;
  min-width: 24px !important;
  max-width: 24px !important;
  text-align: center;
  display: inline-block;
  box-sizing: border-box;
}

.preview-option-item .markdown-body {
  flex: 1;
  line-height: 1.6;
  word-wrap: break-word;
  word-break: break-word;
}

.preview-option-item span {
  flex: 1;
  line-height: 1.6;
}

.preview-subjective {
  margin-top: 15px;
}
</style>

<!-- 非scoped样式，确保markdown-body样式生效 -->
<style>
/* 预览卡片中的 Markdown 样式 - 高优先级 */
.markdown-preview .markdown-body,
.el-card.preview-card .markdown-body,
.preview-card .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.markdown-preview .markdown-body h1,
.markdown-preview .markdown-body h2,
.markdown-preview .markdown-body h3,
.markdown-preview .markdown-body h4,
.markdown-preview .markdown-body h5,
.markdown-preview .markdown-body h6,
.el-card.preview-card .markdown-body h1,
.el-card.preview-card .markdown-body h2,
.el-card.preview-card .markdown-body h3,
.el-card.preview-card .markdown-body h4,
.el-card.preview-card .markdown-body h5,
.el-card.preview-card .markdown-body h6,
.preview-card .markdown-body h1,
.preview-card .markdown-body h2,
.preview-card .markdown-body h3,
.preview-card .markdown-body h4,
.preview-card .markdown-body h5,
.preview-card .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.markdown-preview .markdown-body h1,
.el-card.preview-card .markdown-body h1,
.preview-card .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.markdown-preview .markdown-body h2,
.el-card.preview-card .markdown-body h2,
.preview-card .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.markdown-preview .markdown-body h3,
.el-card.preview-card .markdown-body h3,
.preview-card .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.markdown-preview .markdown-body h3:before,
.el-card.preview-card .markdown-body h3:before,
.preview-card .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.markdown-preview .markdown-body p,
.el-card.preview-card .markdown-body p,
.preview-card .markdown-body p {
  margin-bottom: 16px !important;
}

.markdown-preview .markdown-body strong,
.el-card.preview-card .markdown-body strong,
.preview-card .markdown-body strong {
  font-weight: bold !important;
}

.markdown-preview .markdown-body em,
.el-card.preview-card .markdown-body em,
.preview-card .markdown-body em {
  font-style: italic !important;
}

.markdown-preview .markdown-body code,
.el-card.preview-card .markdown-body code,
.preview-card .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.markdown-preview .markdown-body pre,
.el-card.preview-card .markdown-body pre,
.preview-card .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 通用样式 - 兼容没有 .el-card 前缀的情况 */
.preview-card .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.preview-card .markdown-body h1,
.preview-card .markdown-body h2,
.preview-card .markdown-body h3,
.preview-card .markdown-body h4,
.preview-card .markdown-body h5,
.preview-card .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.preview-card .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.preview-card .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.preview-card .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.preview-card .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.preview-card .markdown-body p {
  margin-bottom: 16px !important;
}

.preview-card .markdown-body strong {
  font-weight: bold !important;
}

.preview-card .markdown-body em {
  font-style: italic !important;
}

.preview-card .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.preview-card .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}
</style>
