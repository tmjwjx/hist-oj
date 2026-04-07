<template>
  <div class="question-editor-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <el-button icon="el-icon-arrow-left" @click="goBack">返回</el-button>
        <h3>{{ pageTitle }}</h3>
      </div>
      <div class="header-right">
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitQuestion">{{ submitButtonText }}</el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="14" class="form-column">
        <el-card shadow="never" class="form-card">
          <el-form :model="form" label-width="110px" class="question-form">
            <el-form-item label="题型" required>
              <el-select v-model="form.type" @change="handleTypeChange">
                <el-option label="单选题" value="single_choice"></el-option>
                <el-option label="多选题" value="multiple_choice"></el-option>
                <el-option label="判断题" value="judge"></el-option>
                <el-option label="主观题" value="subjective"></el-option>
                <el-option label="组合题" value="composite"></el-option>
              </el-select>
            </el-form-item>

            <el-form-item label="题目标题" required>
              <el-input v-model="form.title" placeholder="请输入题目标题"></el-input>
            </el-form-item>

            <el-form-item label="题目内容" required>
              <el-input
                type="textarea"
                v-model="form.content"
                :rows="7"
                placeholder="请输入题目内容，支持 Markdown"
              ></el-input>
            </el-form-item>

            <template v-if="form.type === 'single_choice'">
              <el-form-item label="选项" required>
                <div class="options-container">
                  <div v-for="(_, index) in form.choiceOptions" :key="index" class="option-item">
                    <div class="option-select-wrap">
                      <el-radio v-model="form.correctAnswer" :label="index">
                        {{ optionLetters[index] }}
                      </el-radio>
                    </div>
                    <div class="option-input-wrap">
                      <el-input
                        type="textarea"
                        :rows="2"
                        v-model="form.choiceOptions[index]"
                        :placeholder="`${optionLetters[index]}. 选项内容（支持 Markdown）`"
                      ></el-input>
                    </div>
                    <el-button size="mini" @click="openOptionEditor(index)">窗口编辑</el-button>
                  </div>
                </div>
                <div class="form-tip">
                  <i class="el-icon-info"></i>
                  既可直接输入，也可点击“窗口编辑”查看更大编辑区与实时 Markdown 预览
                </div>
              </el-form-item>
            </template>

            <template v-if="form.type === 'multiple_choice'">
              <el-form-item label="选项" required>
                <div class="options-container">
                  <div v-for="(_, index) in form.choiceOptions" :key="index" class="option-item">
                    <div class="option-select-wrap">
                      <el-checkbox v-model="form.correctAnswers[index]">
                        {{ optionLetters[index] }}
                      </el-checkbox>
                    </div>
                    <div class="option-input-wrap">
                      <el-input
                        type="textarea"
                        :rows="2"
                        v-model="form.choiceOptions[index]"
                        :placeholder="`${optionLetters[index]}. 选项内容（支持 Markdown）`"
                      ></el-input>
                    </div>
                    <el-button size="mini" @click="openOptionEditor(index)">窗口编辑</el-button>
                  </div>
                </div>
                <div class="form-tip">
                  <i class="el-icon-info"></i>
                  既可直接输入，也可点击“窗口编辑”查看更大编辑区与实时 Markdown 预览
                </div>
              </el-form-item>
            </template>

            <template v-if="form.type === 'judge'">
              <el-form-item label="正确答案" required>
                <el-radio-group v-model="form.correctAnswer">
                  <el-radio label="true">正确</el-radio>
                  <el-radio label="false">错误</el-radio>
                </el-radio-group>
              </el-form-item>
            </template>

            <template v-if="form.type === 'subjective'">
              <el-form-item label="参考答案">
                <el-input
                  type="textarea"
                  v-model="form.referenceAnswer"
                  :rows="3"
                  placeholder="请输入参考答案（可选）"
                ></el-input>
              </el-form-item>
            </template>

            <template v-if="form.type === 'composite'">
              <el-form-item label="子题配置" required>
                <div class="composite-panel">
                  <div
                    v-for="(subQuestion, subIndex) in form.compositeQuestions"
                    :key="subQuestion.id"
                    class="composite-sub-question"
                  >
                    <div class="composite-sub-header">
                      <span>子题 {{ subIndex + 1 }}</span>
                      <div class="composite-sub-actions">
                        <span class="composite-sub-score-label">分值</span>
                        <el-input-number
                          v-model="subQuestion.score"
                          :min="1"
                          :max="100"
                          size="mini"
                          @change="recalculateCompositeTotalScore"
                        ></el-input-number>
                        <el-button
                          type="text"
                          size="mini"
                          @click="removeCompositeQuestion(subIndex)"
                          :disabled="form.compositeQuestions.length <= 1"
                        >
                          删除
                        </el-button>
                      </div>
                    </div>

                    <div class="composite-content-editor">
                      <el-input
                        type="textarea"
                        :rows="3"
                        v-model="subQuestion.content"
                        :placeholder="`请输入子题 ${subIndex + 1} 题干（支持 Markdown）`"
                      ></el-input>
                      <el-button size="mini" @click="openCompositeContentEditor(subIndex)">窗口编辑</el-button>
                    </div>

                    <div class="composite-options">
                      <div v-for="(_, optionIndex) in subQuestion.choiceOptions" :key="optionIndex" class="option-item">
                        <div class="option-select-wrap">
                          <el-radio v-model="subQuestion.correctAnswer" :label="optionIndex">
                            {{ optionLetters[optionIndex] }}
                          </el-radio>
                        </div>
                        <div class="option-input-wrap">
                          <el-input
                            type="textarea"
                            :rows="2"
                            v-model="subQuestion.choiceOptions[optionIndex]"
                            :placeholder="`${optionLetters[optionIndex]}. 子题选项内容（支持 Markdown）`"
                          ></el-input>
                        </div>
                        <el-button size="mini" @click="openCompositeOptionEditor(subIndex, optionIndex)">窗口编辑</el-button>
                      </div>
                    </div>

                  </div>
                </div>
                <el-button size="mini" type="primary" plain @click="addCompositeQuestion">新增子题</el-button>
              </el-form-item>
            </template>

            <el-form-item label="题目解析">
              <el-input
                type="textarea"
                v-model="form.analysis"
                :rows="3"
                placeholder="请输入题目解析（可选）"
              ></el-input>
            </el-form-item>

            <el-form-item label="题目标签">
              <div class="tags-input-container">
                <div class="tags-list">
                  <el-tag
                    v-for="(tag, index) in form.tags"
                    :key="index"
                    closable
                    @close="removeTag(index)"
                    style="margin-right: 5px; margin-bottom: 5px;"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
                <el-input
                  v-model="tagInput"
                  placeholder="输入标签名称，按回车添加"
                  @keyup.enter.native="addTag"
                />
              </div>
            </el-form-item>

            <el-row :gutter="12" class="compact-form-row">
              <el-col :span="12">
                <el-form-item label="所属课程">
                  <el-select v-model="form.course" placeholder="请选择课程" style="width: 100%">
                    <el-option
                      v-for="course in commonCourses"
                      :key="course"
                      :label="course"
                      :value="course"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="难度">
                  <el-rate v-model="form.difficulty" :max="3" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="12" class="compact-form-row">
              <el-col :span="12">
                <el-form-item label="默认分值">
                  <el-input-number v-model="form.score" :min="1" :max="100" :disabled="form.type === 'composite'"></el-input-number>
                  <div v-if="form.type === 'composite'" class="compact-tip">组合题总分自动等于所有子题分值之和</div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="共享状态">
                  <el-switch v-model="form.isShared" active-text="共享" inactive-text="个人"></el-switch>
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="10" class="preview-column">
        <el-card class="preview-card" shadow="never">
          <div slot="header">
            <i class="el-icon-view"></i> 实时预览
          </div>
          <div class="preview-content">
            <div v-if="form.title" v-html="renderMarkdown(form.title)" class="markdown-body preview-title" v-highlight></div>
            <p v-else class="preview-placeholder">题目标题预览</p>

            <div v-if="form.content" v-html="renderMarkdown(form.content)" class="markdown-body preview-content-text" v-highlight></div>
            <p v-else class="preview-placeholder">题目内容预览</p>

            <div v-if="form.type === 'single_choice'" class="preview-options">
              <div v-for="(option, index) in form.choiceOptions" :key="index" class="preview-option-item">
                <div class="preview-option-head">
                  <el-tag :type="form.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ optionLetters[index] }}
                  </el-tag>
                </div>
                <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body preview-option-content" v-highlight></div>
                <div v-else class="preview-placeholder">选项内容</div>
              </div>
            </div>

            <div v-if="form.type === 'multiple_choice'" class="preview-options">
              <div v-for="(option, index) in form.choiceOptions" :key="index" class="preview-option-item">
                <div class="preview-option-head">
                  <el-tag :type="form.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ optionLetters[index] }}
                  </el-tag>
                </div>
                <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body preview-option-content" v-highlight></div>
                <div v-else class="preview-placeholder">选项内容</div>
              </div>
            </div>

            <div v-if="form.type === 'judge'" class="preview-options">
              <div class="preview-option-item">
                <el-tag :type="form.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                <span>正确</span>
              </div>
              <div class="preview-option-item">
                <el-tag :type="form.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                <span>错误</span>
              </div>
            </div>

            <div v-if="form.type === 'subjective'" class="preview-subjective">
              <el-alert type="info" :closable="false">
                主观题，学生需输入文字答案
              </el-alert>
            </div>

            <div v-if="form.type === 'composite'" class="preview-options">
              <div
                v-for="(subQuestion, subIndex) in form.compositeQuestions"
                :key="subQuestion.id"
                class="preview-option-item composite-preview-item"
              >
                <div class="composite-preview-title">
                  子题 {{ subIndex + 1 }}（{{ subQuestion.score || 0 }} 分）
                </div>
                <div
                  v-if="subQuestion.content"
                  class="markdown-body preview-option-content"
                  v-html="renderMarkdown(subQuestion.content)"
                  v-highlight
                ></div>
                <div v-else class="preview-placeholder">子题题干预览</div>
                <div class="preview-options">
                  <div
                    v-for="(option, optionIndex) in subQuestion.choiceOptions"
                    :key="optionIndex"
                    class="preview-option-item"
                  >
                    <div class="preview-option-head">
                      <el-tag :type="subQuestion.correctAnswer === optionIndex ? 'success' : 'info'" size="small">
                        {{ optionLetters[optionIndex] }}
                      </el-tag>
                    </div>
                    <div
                      v-if="option"
                      class="markdown-body preview-option-content"
                      v-html="renderMarkdown(option)"
                      v-highlight
                    ></div>
                    <div v-else class="preview-placeholder">选项内容</div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="form.analysis" class="preview-analysis">
              <el-divider content-position="left">
                <i class="el-icon-document" style="color: #E6A23C;"></i>
                <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
              </el-divider>
              <div v-html="renderMarkdown(form.analysis)" class="markdown-body preview-analysis-content" v-highlight></div>
            </div>
            <p v-else class="preview-placeholder" style="margin-top: 15px;">题目解析预览</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog
      :title="optionEditorTitle"
      :visible.sync="optionEditor.visible"
      width="1100px"
      append-to-body
    >
      <el-row :gutter="18">
        <el-col :span="12">
          <div class="dialog-title">编辑区</div>
          <el-input
            type="textarea"
            :rows="18"
            v-model="optionEditor.content"
            placeholder="请输入选项内容，支持 Markdown"
            @input="syncOptionEditorContent"
          ></el-input>
        </el-col>
        <el-col :span="12">
          <div class="dialog-title">Markdown 预览</div>
          <div class="option-dialog-preview markdown-body" v-html="renderMarkdown(optionEditor.content)" v-highlight></div>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="closeOptionEditor">关闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})

md.use(katex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'QuestionBankEditor',
  props: {
    scene: {
      type: String,
      default: 'teacher',
      validator(value) {
        return ['teacher', 'admin'].includes(value)
      }
    }
  },
  data() {
    return {
      loading: false,
      saving: false,
      tagInput: '',
      optionLetters: ['A', 'B', 'C', 'D'],
      optionEditor: {
        visible: false,
        index: null,
        subIndex: null,
        mode: '',
        content: ''
      },
      commonCourses: [
        '数据结构',
        '算法设计与分析',
        '计算机网络',
        '操作系统',
        '计算机组成原理',
        '高等数学',
        '线性代数',
        '政治',
        '英语'
      ],
      form: this.getDefaultForm()
    }
  },
  computed: {
    isAdmin() {
      return this.scene === 'admin'
    },
    questionId() {
      const raw = this.$route.params.questionId
      return raw ? String(raw) : ''
    },
    isEdit() {
      return !!this.questionId
    },
    pageTitle() {
      return this.isEdit ? '编辑客观题' : '创建客观题'
    },
    submitButtonText() {
      return this.isEdit ? '保存修改' : '创建题目'
    },
    backRouteName() {
      return this.isAdmin ? 'admin-question-bank' : 'QuestionBank'
    },
    optionEditorTitle() {
      if (!this.optionEditor.mode) {
        return '编辑内容'
      }
      if (this.optionEditor.mode === 'normal_option') {
        const index = this.optionEditor.index
        const letter = this.optionLetters[index] || ''
        return `编辑选项 ${letter}`
      }
      if (this.optionEditor.mode === 'composite_content') {
        const subIndex = this.optionEditor.subIndex
        return `编辑子题 ${subIndex + 1} 题干`
      }
      if (this.optionEditor.mode === 'composite_option') {
        const subIndex = this.optionEditor.subIndex
        const letter = this.optionLetters[this.optionEditor.index] || ''
        return `编辑子题 ${subIndex + 1} 选项 ${letter}`
      }
      return '编辑内容'
    }
  },
  created() {
    if (this.isEdit) {
      this.loadQuestionDetail()
    }
  },
  methods: {
    getDefaultForm() {
      return {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        compositeQuestions: [this.getDefaultCompositeQuestion(1)],
        referenceAnswer: '',
        analysis: '',
        tags: [],
        course: '',
        difficulty: 1,
        score: 2,
        isShared: false
      }
    },
    getDefaultCompositeQuestion(order) {
      const serial = `${Date.now()}_${Math.floor(Math.random() * 100000)}_${order}`
      return {
        id: `sq_${serial}`,
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        score: 1
      }
    },
    recalculateCompositeTotalScore() {
      if (this.form.type !== 'composite') return
      const total = (this.form.compositeQuestions || []).reduce((sum, item) => {
        const score = Number(item && item.score ? item.score : 0)
        return sum + (score > 0 ? score : 0)
      }, 0)
      this.form.score = total > 0 ? total : 1
    },
    goBack(forceRefresh = false) {
      const route = { name: this.backRouteName }
      if (forceRefresh) {
        route.query = { refreshTs: Date.now() }
      }
      this.$router.push(route)
    },
    handleTypeChange() {
      if (this.form.type === 'single_choice') {
        this.form.correctAnswer = 0
        this.form.correctAnswers = [false, false, false, false]
        this.form.score = 2
      } else if (this.form.type === 'multiple_choice') {
        this.form.correctAnswer = 0
        this.form.correctAnswers = [false, false, false, false]
        this.form.score = 5
      } else if (this.form.type === 'judge') {
        this.form.correctAnswer = 'true'
        this.form.correctAnswers = [false, false, false, false]
        this.form.score = 1
      } else if (this.form.type === 'subjective') {
        this.form.correctAnswer = ''
        this.form.correctAnswers = [false, false, false, false]
        this.form.score = 5
      } else if (this.form.type === 'composite') {
        if (!Array.isArray(this.form.compositeQuestions) || this.form.compositeQuestions.length === 0) {
          this.form.compositeQuestions = [this.getDefaultCompositeQuestion(1)]
        }
        this.form.correctAnswer = ''
        this.form.correctAnswers = [false, false, false, false]
        this.recalculateCompositeTotalScore()
      }
      this.form.referenceAnswer = ''
    },
    addTag() {
      const tag = this.tagInput.trim()
      if (tag && !this.form.tags.includes(tag)) {
        this.form.tags.push(tag)
      }
      this.tagInput = ''
    },
    removeTag(index) {
      this.form.tags.splice(index, 1)
    },
    openOptionEditor(index) {
      this.optionEditor.visible = true
      this.optionEditor.mode = 'normal_option'
      this.optionEditor.index = index
      this.optionEditor.subIndex = null
      this.optionEditor.content = this.form.choiceOptions[index] || ''
    },
    openCompositeContentEditor(subIndex) {
      const subQuestion = this.form.compositeQuestions[subIndex]
      if (!subQuestion) return
      this.optionEditor.visible = true
      this.optionEditor.mode = 'composite_content'
      this.optionEditor.index = null
      this.optionEditor.subIndex = subIndex
      this.optionEditor.content = subQuestion.content || ''
    },
    openCompositeOptionEditor(subIndex, optionIndex) {
      const subQuestion = this.form.compositeQuestions[subIndex]
      if (!subQuestion) return
      this.optionEditor.visible = true
      this.optionEditor.mode = 'composite_option'
      this.optionEditor.index = optionIndex
      this.optionEditor.subIndex = subIndex
      this.optionEditor.content = subQuestion.choiceOptions[optionIndex] || ''
    },
    closeOptionEditor() {
      this.optionEditor.visible = false
      this.optionEditor.mode = ''
      this.optionEditor.index = null
      this.optionEditor.subIndex = null
      this.optionEditor.content = ''
    },
    syncOptionEditorContent(value) {
      if (this.optionEditor.mode === 'normal_option') {
        if (this.optionEditor.index === null) return
        this.$set(this.form.choiceOptions, this.optionEditor.index, value)
        return
      }
      if (this.optionEditor.mode === 'composite_content') {
        const subIndex = this.optionEditor.subIndex
        if (subIndex === null || !this.form.compositeQuestions[subIndex]) return
        this.$set(this.form.compositeQuestions[subIndex], 'content', value)
        return
      }
      if (this.optionEditor.mode === 'composite_option') {
        const subIndex = this.optionEditor.subIndex
        const optionIndex = this.optionEditor.index
        if (subIndex === null || optionIndex === null || !this.form.compositeQuestions[subIndex]) return
        this.$set(this.form.compositeQuestions[subIndex].choiceOptions, optionIndex, value)
      }
    },
    addCompositeQuestion() {
      if (!Array.isArray(this.form.compositeQuestions)) {
        this.$set(this.form, 'compositeQuestions', [])
      }
      this.form.compositeQuestions.push(this.getDefaultCompositeQuestion(this.form.compositeQuestions.length + 1))
      this.recalculateCompositeTotalScore()
    },
    removeCompositeQuestion(index) {
      if (!Array.isArray(this.form.compositeQuestions)) return
      if (this.form.compositeQuestions.length <= 1) {
        this.$message.warning('组合题至少保留一个子题')
        return
      }
      this.form.compositeQuestions.splice(index, 1)
      this.recalculateCompositeTotalScore()
    },
    validateForm() {
      if (!this.form.title || !this.form.title.trim()) {
        this.$message.warning('请输入题目标题')
        return false
      }
      if (!this.form.content || !this.form.content.trim()) {
        this.$message.warning('请输入题目内容')
        return false
      }

      if (this.form.type === 'single_choice' || this.form.type === 'multiple_choice') {
        const hasEmptyOption = this.form.choiceOptions.some(opt => !opt || !String(opt).trim())
        if (hasEmptyOption) {
          this.$message.warning('请填写完整的四个选项内容')
          return false
        }
      }

      if (this.form.type === 'multiple_choice') {
        const selectedCount = this.form.correctAnswers.filter(Boolean).length
        if (selectedCount === 0) {
          this.$message.warning('请至少选择一个正确答案')
          return false
        }
      }

      if (this.form.type === 'composite') {
        if (!Array.isArray(this.form.compositeQuestions) || this.form.compositeQuestions.length === 0) {
          this.$message.warning('请至少添加一个子题')
          return false
        }
        for (let i = 0; i < this.form.compositeQuestions.length; i++) {
          const subQuestion = this.form.compositeQuestions[i]
          if (!subQuestion || !String(subQuestion.content || '').trim()) {
            this.$message.warning(`请填写子题 ${i + 1} 的题干`)
            return false
          }
          const invalidOptionIndex = (subQuestion.choiceOptions || []).findIndex(option => !String(option || '').trim())
          if (invalidOptionIndex !== -1) {
            this.$message.warning(`请填写子题 ${i + 1} 的 ${this.optionLetters[invalidOptionIndex]} 选项内容`)
            return false
          }
          const subScore = Number(subQuestion.score || 0)
          if (!Number.isFinite(subScore) || subScore <= 0) {
            this.$message.warning(`子题 ${i + 1} 的分值必须大于0`)
            return false
          }
          if (subQuestion.correctAnswer === null || subQuestion.correctAnswer === undefined || subQuestion.correctAnswer < 0 || subQuestion.correctAnswer > 3) {
            this.$message.warning(`请选择子题 ${i + 1} 的正确答案`)
            return false
          }
        }
      }
      return true
    },
    buildSubmitData() {
      const submitData = {
        type: this.form.type,
        title: this.form.title,
        content: this.form.content,
        analysis: this.form.analysis || '',
        tags: JSON.stringify(this.form.tags || []),
        course: this.form.course || '',
        difficulty: this.form.difficulty || 1,
        score: this.form.score || 2,
        isShared: this.form.isShared ? 1 : 0
      }

      if (this.form.type === 'single_choice') {
        const optionsArray = this.form.choiceOptions.map((opt, idx) => `${this.optionLetters[idx]}. ${opt}`)
        submitData.options = JSON.stringify(optionsArray)
        submitData.answer = this.optionLetters[this.form.correctAnswer]
      } else if (this.form.type === 'multiple_choice') {
        const optionsArray = this.form.choiceOptions.map((opt, idx) => `${this.optionLetters[idx]}. ${opt}`)
        submitData.options = JSON.stringify(optionsArray)
        const selectedAnswers = this.form.correctAnswers
          .map((checked, idx) => (checked ? this.optionLetters[idx] : null))
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.form.type === 'judge') {
        submitData.answer = String(this.form.correctAnswer) === 'true' ? 'true' : 'false'
        submitData.options = null
      } else if (this.form.type === 'subjective') {
        submitData.answer = this.form.referenceAnswer || '需人工评分'
        submitData.options = null
      } else if (this.form.type === 'composite') {
        const compositeQuestions = (this.form.compositeQuestions || []).map((subQuestion, subIndex) => {
          const subID = String(subQuestion.id || `sq_${subIndex + 1}`)
          return {
            id: subID,
            content: subQuestion.content,
            options: (subQuestion.choiceOptions || []).map((option, optionIndex) =>
              `${this.optionLetters[optionIndex]}. ${option}`
            ),
            score: Number(subQuestion.score || 1)
          }
        })

        const answerMap = {}
        compositeQuestions.forEach((subQuestion, subIndex) => {
          const selectedIndex = this.form.compositeQuestions[subIndex].correctAnswer
          answerMap[subQuestion.id] = this.optionLetters[selectedIndex]
        })

        submitData.options = JSON.stringify(compositeQuestions)
        submitData.answer = JSON.stringify(answerMap)
        submitData.score = compositeQuestions.reduce((sum, item) => sum + (Number(item.score) || 0), 0)
      }

      if (!this.isAdmin) {
        const classroomId = this.$route.query.classroomId
        if (!this.isEdit && classroomId) {
          submitData.classroomId = classroomId
        }
      }
      return submitData
    },
    async loadQuestionDetail() {
      this.loading = true
      try {
        const res = await this.$http.get(`/api/classroom/question/${this.questionId}`)
        if (res.data.code !== 200 || !res.data.data) {
          this.$message.error(res.data.message || '加载题目失败')
          this.goBack()
          return
        }
        this.fillFormByQuestion(res.data.data)
      } catch (error) {
        this.$message.error('加载题目失败')
        this.goBack()
      } finally {
        this.loading = false
      }
    },
    fillFormByQuestion(question) {
      const nextForm = this.getDefaultForm()
      nextForm.type = question.type
      nextForm.title = question.title || ''
      nextForm.content = question.content || ''
      nextForm.difficulty = question.difficulty || 1
      nextForm.score = question.score || 2
      nextForm.isShared = question.isShared === 1
      nextForm.analysis = question.analysis || ''
      nextForm.course = question.course || ''

      if (question.tags) {
        try {
          const tags = JSON.parse(question.tags)
          nextForm.tags = Array.isArray(tags) ? tags : []
        } catch (e) {
          nextForm.tags = []
        }
      }

      if (question.type === 'single_choice' || question.type === 'multiple_choice') {
        if (question.options) {
          try {
            const parsedOptions = JSON.parse(question.options)
            nextForm.choiceOptions = parsedOptions.map(opt => String(opt).replace(/^[A-D]\.\s*/, ''))
          } catch (e) {
            nextForm.choiceOptions = ['', '', '', '']
          }
        }
        const answerIndexMap = { A: 0, B: 1, C: 2, D: 3 }
        if (question.type === 'single_choice') {
          nextForm.correctAnswer = answerIndexMap[question.answer] ?? 0
        } else {
          nextForm.correctAnswers = [false, false, false, false]
          try {
            const answers = typeof question.answer === 'string' ? JSON.parse(question.answer) : question.answer
            if (Array.isArray(answers)) {
              answers.forEach(ans => {
                const index = answerIndexMap[ans]
                if (index !== undefined) nextForm.correctAnswers[index] = true
              })
            }
          } catch (e) {
            // ignore parse failure
          }
        }
      } else if (question.type === 'judge') {
        nextForm.correctAnswer = String(question.answer || '').toLowerCase() === 'true' ? 'true' : 'false'
      } else if (question.type === 'subjective') {
        nextForm.referenceAnswer = question.answer || ''
      } else if (question.type === 'composite') {
        const answerIndexMap = { A: 0, B: 1, C: 2, D: 3 }
        let parsedAnswerMap = {}
        try {
          const parsedAnswer = typeof question.answer === 'string' ? JSON.parse(question.answer || '{}') : question.answer
          if (parsedAnswer && typeof parsedAnswer === 'object' && !Array.isArray(parsedAnswer)) {
            parsedAnswerMap = parsedAnswer
          }
        } catch (e) {
          parsedAnswerMap = {}
        }

        try {
          const parsedOptions = typeof question.options === 'string' ? JSON.parse(question.options || '[]') : question.options
          if (Array.isArray(parsedOptions) && parsedOptions.length > 0) {
            nextForm.compositeQuestions = parsedOptions.map((item, index) => {
              const subID = String(item.id || `sq_${index + 1}`)
              const rawOptions = Array.isArray(item.options) ? item.options : []
              const choiceOptions = rawOptions.length === 4
                ? rawOptions.map(opt => String(opt).replace(/^[A-D]\.\s*/, ''))
                : ['', '', '', '']
              const rawAnswer = String(parsedAnswerMap[subID] || '').toUpperCase()
              return {
                id: subID,
                content: item.content || '',
                choiceOptions,
                correctAnswer: answerIndexMap[rawAnswer] ?? 0,
                score: Number(item.score || item.subScore || 1)
              }
            })
          } else {
            nextForm.compositeQuestions = [this.getDefaultCompositeQuestion(1)]
          }
        } catch (e) {
          nextForm.compositeQuestions = [this.getDefaultCompositeQuestion(1)]
        }
        nextForm.score = (nextForm.compositeQuestions || []).reduce((sum, item) => {
          const score = Number(item && item.score ? item.score : 0)
          return sum + (score > 0 ? score : 0)
        }, 0) || 1
      }

      this.form = nextForm
    },
    async submitQuestion() {
      if (!this.validateForm()) return

      this.saving = true
      const submitData = this.buildSubmitData()
      try {
        let res
        if (this.isAdmin) {
          if (this.isEdit) {
            res = await this.$http.put(`/api/classroom/admin/question/${this.questionId}`, submitData)
          } else {
            res = await this.$http.post('/api/classroom/admin/question', submitData)
          }
        } else if (this.isEdit) {
          res = await this.$http.put(`/api/classroom/question/${this.questionId}`, submitData)
        } else {
          res = await this.$http.post('/api/classroom/question', submitData)
        }

        if (res.data.code === 200) {
          this.$message.success(this.isEdit ? '更新成功' : '创建成功')
          this.goBack(true)
        } else {
          this.$message.error(res.data.message || '操作失败')
        }
      } catch (error) {
        this.$message.error('操作失败')
      } finally {
        this.saving = false
      }
    },
    renderMarkdown(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        return content
      }
    }
  }
}
</script>

<style scoped>
.question-editor-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 18px;
  gap: 14px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h3 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-column,
.preview-column {
  max-height: calc(100vh - 160px);
  overflow-y: auto;
}

.form-card {
  min-height: 100%;
}

.question-form .el-form-item {
  margin-bottom: 14px;
}

.compact-form-row .el-form-item {
  margin-bottom: 10px;
}

.compact-tip {
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
}

.options-container {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.composite-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.composite-sub-question {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #fafbfd;
}

.composite-sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  font-weight: 600;
}

.composite-sub-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.composite-sub-score-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

.composite-content-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 10px;
}

.composite-options {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
}

.composite-preview-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.option-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.option-select-wrap {
  min-width: 54px;
  padding-top: 8px;
}

.option-input-wrap {
  flex: 1;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.tags-input-container {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 5px;
  min-height: 80px;
}

.tags-list {
  margin-bottom: 8px;
}

.preview-card {
  min-height: 100%;
  border: 2px solid #e4e7ed;
}

.preview-content {
  padding: 10px;
}

.preview-title {
  color: #303133;
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 15px;
  border-bottom: 2px solid #e4e7ed;
  padding-bottom: 10px;
}

.preview-content-text {
  color: #606266;
  line-height: 1.8;
  margin-bottom: 20px;
  font-size: 14px;
}

.preview-placeholder {
  color: #c0c4cc;
  font-style: italic;
}

.preview-options {
  margin-top: 15px;
}

.preview-option-item {
  display: block;
  padding: 10px;
  margin-bottom: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.preview-option-head {
  margin-bottom: 6px;
}

.preview-option-content {
  width: 100%;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.preview-option-content p {
  margin: 0;
}

.preview-option-content pre {
  margin: 0;
  max-width: 100%;
  overflow-x: auto;
}

.preview-subjective {
  margin-top: 15px;
}

.preview-analysis {
  margin-top: 15px;
  padding: 10px;
  background: #fff9e6;
  border-left: 3px solid #e6a23c;
  border-radius: 4px;
}

.preview-analysis-content {
  margin-top: 10px;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
  line-height: 1.8;
}

.dialog-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.option-dialog-preview {
  min-height: 390px;
  max-height: 460px;
  overflow-y: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
}

@media (max-width: 1200px) {
  .form-column,
  .preview-column {
    max-height: none;
  }
}
</style>

<style>
.question-editor-page .markdown-body pre {
  padding: 0 !important;
}

.question-editor-page .markdown-body pre code {
  margin: 0 !important;
  text-indent: 0 !important;
}

.question-editor-page .markdown-body pre ol.pre-numbering {
  display: none !important;
}
</style>
