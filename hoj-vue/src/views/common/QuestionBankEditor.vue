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
      <el-col :span="24" class="form-column">
        <el-card shadow="never" class="form-card">
          <el-form :model="form" label-width="110px" class="question-form">
            <el-form-item label="题型" required>
              <el-select v-model="form.type" @change="handleTypeChange">
                <el-option label="单选题" value="single_choice"></el-option>
                <el-option label="多选题" value="multiple_choice"></el-option>
                <el-option label="判断题" value="judge"></el-option>
                <el-option label="填空题" value="fill_blank"></el-option>
                <el-option label="主观题" value="subjective"></el-option>
                <el-option label="组合题" value="composite"></el-option>
              </el-select>
            </el-form-item>

            <el-form-item label="题目标题" required>
              <el-input v-model="form.title" placeholder="请输入题目标题"></el-input>
            </el-form-item>

            <el-form-item label="题目内容" required>
              <Editor :value.sync="form.content" :allow-upload="true" class="question-content-markdown-editor" />
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
                  既可直接输入，也可点击“窗口编辑”使用完整 Markdown 编辑器
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
                  既可直接输入，也可点击“窗口编辑”使用完整 Markdown 编辑器
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

            <template v-if="form.type === 'fill_blank'">
              <el-form-item label="正确答案" required>
                <div class="fill-blank-answer-list">
                  <div
                    v-for="(answer, index) in form.fillBlankAnswers"
                    :key="`fill-blank-${index}`"
                    class="fill-blank-answer-item"
                  >
                    <el-input
                      v-model="form.fillBlankAnswers[index]"
                      placeholder="请输入一个可判对的答案"
                    ></el-input>
                    <el-button
                      type="primary"
                      icon="el-icon-plus"
                      circle
                      plain
                      title="新增答案"
                      @click="addFillBlankAnswer"
                    ></el-button>
                    <el-button
                      v-if="form.fillBlankAnswers.length > 1"
                      type="danger"
                      icon="el-icon-minus"
                      circle
                      plain
                      title="删除答案"
                      @click="removeFillBlankAnswer(index)"
                    ></el-button>
                  </div>
                </div>
                <div class="form-tip">
                  <i class="el-icon-info"></i>
                  教师和管理员可设置多个标准答案，学生答案命中任意一个即判对。
                </div>
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
              <Editor :value.sync="form.analysis" :allow-upload="true" class="question-analysis-markdown-editor" />
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

    </el-row>

    <el-dialog
      :title="optionEditorTitle"
      :visible.sync="optionEditor.visible"
      width="900px"
      append-to-body
    >
      <Editor
        :value="optionEditor.content"
        :allow-upload="true"
        class="question-dialog-markdown-editor"
        @update:value="syncOptionEditorContent"
      />
      <span slot="footer">
        <el-button @click="closeOptionEditor">关闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import Editor from '@/components/admin/Editor'

const QUESTION_IMAGE_UPLOAD_PREFIX = '/uploads/classroom/questions/'
const QUESTION_IMAGE_MARKDOWN_REGEX = /!\[[^\]]*]\(([^)]+)\)/g
const QUESTION_IMAGE_HTML_REGEX = /<img[^>]*src\s*=\s*["']([^"']+)["'][^>]*>/gi

export default {
  name: 'QuestionBankEditor',
  components: {
    Editor
  },
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
        'GESP 考级课',
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
      originalQuestionImageUrls: [],
      sessionUploadedQuestionImageUrls: [],
      hasSavedQuestion: false,
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
    this.originalQuestionImageUrls = []
    this.sessionUploadedQuestionImageUrls = []
    this.hasSavedQuestion = false
    if (this.isEdit) {
      this.loadQuestionDetail()
    }
  },
  beforeDestroy() {
    if (!this.hasSavedQuestion) {
      this.cleanupSessionUploadedQuestionImagesOnExit()
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
        fillBlankAnswers: [''],
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
    normalizeQuestionImageUrl(rawUrl) {
      let value = String(rawUrl || '').trim()
      if (!value) return ''

      if (value.startsWith('<') && value.endsWith('>')) {
        value = value.slice(1, -1).trim()
      }
      if (!value) return ''

      try {
        if (value.startsWith('//')) {
          value = new URL(`https:${value}`).pathname
        } else if (/^https?:\/\//i.test(value)) {
          value = new URL(value).pathname
        }
      } catch (e) {
        // ignore parse failure and continue with raw value
      }

      if (typeof window !== 'undefined' && value.startsWith(window.location.origin)) {
        value = value.substring(window.location.origin.length)
      }

      value = value.split('?')[0]
      value = value.split('#')[0]
      if (!value.startsWith(QUESTION_IMAGE_UPLOAD_PREFIX)) return ''

      const filename = value.substring(QUESTION_IMAGE_UPLOAD_PREFIX.length).trim()
      if (!filename) return ''
      if (filename.includes('/') || filename.includes('\\') || filename.includes('..')) return ''
      const lowerFilename = filename.toLowerCase()
      if (!lowerFilename.endsWith('.png') && !lowerFilename.endsWith('.jpg') && !lowerFilename.endsWith('.jpeg')) return ''
      return `${QUESTION_IMAGE_UPLOAD_PREFIX}${filename}`
    },
    uniqueQuestionImageUrls(urls) {
      const result = []
      const seen = new Set()
      ;(urls || []).forEach((item) => {
        const normalized = this.normalizeQuestionImageUrl(item)
        if (!normalized || seen.has(normalized)) return
        seen.add(normalized)
        result.push(normalized)
      })
      return result
    },
    extractQuestionImageUrlsFromText(content) {
      const text = String(content || '')
      if (!text.trim()) return []

      const urls = []
      const markdownRegex = new RegExp(QUESTION_IMAGE_MARKDOWN_REGEX.source, 'g')
      let markdownMatch = markdownRegex.exec(text)
      while (markdownMatch) {
        let candidate = (markdownMatch[1] || '').trim()
        if (candidate) {
          const titleMatch = candidate.match(/^(\S+)\s+["'][^"']*["']$/)
          if (titleMatch && titleMatch[1]) {
            candidate = titleMatch[1]
          }
          const normalized = this.normalizeQuestionImageUrl(candidate)
          if (normalized) urls.push(normalized)
        }
        markdownMatch = markdownRegex.exec(text)
      }

      const htmlRegex = new RegExp(QUESTION_IMAGE_HTML_REGEX.source, 'gi')
      let htmlMatch = htmlRegex.exec(text)
      while (htmlMatch) {
        const normalized = this.normalizeQuestionImageUrl(htmlMatch[1])
        if (normalized) urls.push(normalized)
        htmlMatch = htmlRegex.exec(text)
      }

      return this.uniqueQuestionImageUrls(urls)
    },
    collectQuestionImageUrlsFromForm(targetForm = this.form) {
      if (!targetForm) return []

      const texts = []
      texts.push(targetForm.title || '')
      texts.push(targetForm.content || '')
      texts.push(targetForm.analysis || '')

      const normalOptions = Array.isArray(targetForm.choiceOptions) ? targetForm.choiceOptions : []
      normalOptions.forEach((option) => texts.push(option || ''))

      const compositeQuestions = Array.isArray(targetForm.compositeQuestions) ? targetForm.compositeQuestions : []
      compositeQuestions.forEach((subQuestion) => {
        if (!subQuestion) return
        texts.push(subQuestion.content || '')
        const subOptions = Array.isArray(subQuestion.choiceOptions) ? subQuestion.choiceOptions : []
        subOptions.forEach((option) => texts.push(option || ''))
      })

      const all = []
      texts.forEach((text) => {
        all.push(...this.extractQuestionImageUrlsFromText(text))
      })
      return this.uniqueQuestionImageUrls(all)
    },
    getSessionOrphanQuestionImageUrls(currentImageUrls = null) {
      const activeUrls = this.uniqueQuestionImageUrls(
        Array.isArray(currentImageUrls) ? currentImageUrls : this.collectQuestionImageUrlsFromForm()
      )
      const activeSet = new Set(activeUrls)
      return this.uniqueQuestionImageUrls(this.sessionUploadedQuestionImageUrls).filter((url) => !activeSet.has(url))
    },
    async requestDeleteQuestionImages(urls, silent = false) {
      const targetUrls = this.uniqueQuestionImageUrls(urls)
      if (!targetUrls.length) return true

      try {
        const res = await this.$http.post('/api/classroom/question/delete-image', { urls: targetUrls })
        if (res.data.code !== 200) {
          if (!silent) this.$message.warning(res.data.message || '图片回收失败')
          return false
        }
        return true
      } catch (error) {
        if (!silent) this.$message.warning('图片回收失败')
        return false
      }
    },
    async cleanupUnusedSessionUploadedQuestionImages() {
      const orphanUrls = this.getSessionOrphanQuestionImageUrls()
      if (!orphanUrls.length) return

      const cleaned = await this.requestDeleteQuestionImages(orphanUrls, true)
      if (cleaned) {
        const orphanSet = new Set(orphanUrls)
        this.sessionUploadedQuestionImageUrls = this.uniqueQuestionImageUrls(
          this.sessionUploadedQuestionImageUrls.filter((item) => !orphanSet.has(this.normalizeQuestionImageUrl(item)))
        )
      }
    },
    cleanupSessionUploadedQuestionImagesOnExit() {
      const targetUrls = this.uniqueQuestionImageUrls(this.sessionUploadedQuestionImageUrls)
      if (!targetUrls.length) return

      this.$http.post('/api/classroom/question/delete-image', { urls: targetUrls }).catch(() => {})
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
      } else if (this.form.type === 'fill_blank') {
        this.form.correctAnswer = ''
        this.form.correctAnswers = [false, false, false, false]
        this.form.fillBlankAnswers = ['']
        this.form.score = 2
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
      if (this.form.type !== 'fill_blank') {
        this.form.fillBlankAnswers = ['']
      }
    },
    normalizeFillBlankStorageValue(raw) {
      const trimmed = String(raw || '').trim()
      if (!trimmed) return ''
      return trimmed.split(/\s+/).join(' ')
    },
    getNormalizedFillBlankAnswers(rawList, allowEmpty = false) {
      const source = Array.isArray(rawList) ? rawList : []
      const normalized = []
      const seen = new Set()
      source.forEach((item) => {
        const value = this.normalizeFillBlankStorageValue(item)
        if (!value || seen.has(value)) return
        seen.add(value)
        normalized.push(value)
      })
      if (!allowEmpty && normalized.length === 0) {
        return []
      }
      return normalized
    },
    addFillBlankAnswer() {
      if (!Array.isArray(this.form.fillBlankAnswers)) {
        this.$set(this.form, 'fillBlankAnswers', [''])
        return
      }
      this.form.fillBlankAnswers.push('')
    },
    removeFillBlankAnswer(index) {
      if (!Array.isArray(this.form.fillBlankAnswers)) return
      if (this.form.fillBlankAnswers.length <= 1) {
        this.$message.warning('填空题至少保留一个答案输入框')
        return
      }
      this.form.fillBlankAnswers.splice(index, 1)
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
      this.optionEditor.content = value
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
        return
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

      if (this.form.type === 'fill_blank') {
        const answers = this.getNormalizedFillBlankAnswers(this.form.fillBlankAnswers)
        if (answers.length === 0) {
          this.$message.warning('填空题至少需要填写一个有效答案')
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
      } else if (this.form.type === 'fill_blank') {
        const normalizedAnswers = this.getNormalizedFillBlankAnswers(this.form.fillBlankAnswers)
        submitData.answer = JSON.stringify(normalizedAnswers)
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
      } else if (question.type === 'fill_blank') {
        let answers = []
        try {
          const parsed = typeof question.answer === 'string' ? JSON.parse(question.answer) : question.answer
          if (Array.isArray(parsed)) {
            answers = parsed
          } else if (question.answer) {
            answers = [question.answer]
          }
        } catch (e) {
          if (question.answer) {
            answers = [question.answer]
          }
        }
        const normalizedAnswers = this.getNormalizedFillBlankAnswers(answers, true)
        nextForm.fillBlankAnswers = normalizedAnswers.length > 0 ? normalizedAnswers : ['']
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
      this.originalQuestionImageUrls = this.collectQuestionImageUrlsFromForm(nextForm)
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
          await this.cleanupUnusedSessionUploadedQuestionImages()
          this.originalQuestionImageUrls = this.collectQuestionImageUrlsFromForm()
          this.hasSavedQuestion = true
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

.form-column {
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

.fill-blank-answer-list {
  display: grid;
  gap: 10px;
}

.fill-blank-answer-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.fill-blank-answer-item .el-input {
  flex: 1;
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

.question-content-markdown-editor /deep/ .v-note-wrapper {
  min-height: 360px;
}

.question-analysis-markdown-editor /deep/ .v-note-wrapper {
  min-height: 300px;
}

.question-dialog-markdown-editor /deep/ .v-note-wrapper {
  min-height: 420px;
}

@media (max-width: 1200px) {
  .form-column {
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
