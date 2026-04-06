<template>
  <div class="question-bank-admin-container">
    <el-card class="question-bank-card">
      <div slot="header" class="card-header">
        <div class="header-left">
          <i class="fa fa-list-alt"></i>
          <span>客观题题库管理</span>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="el-icon-plus" @click="goCreatePage" size="small">创建客观题</el-button>
          <el-button icon="el-icon-refresh" @click="loadQuestions" size="small">刷新</el-button>
        </div>
      </div>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-row :gutter="15" style="margin-bottom: 10px;">
          <el-col :span="6">
            <el-select v-model="filters.type" placeholder="题型筛选" clearable @change="handleFilterChange" style="width: 100%;">
              <el-option label="全部题型" value=""></el-option>
              <el-option label="单选题" value="single_choice"></el-option>
              <el-option label="多选题" value="multiple_choice"></el-option>
              <el-option label="判断题" value="judge"></el-option>
              <el-option label="主观题" value="subjective"></el-option>
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.course" placeholder="课程筛选" clearable @change="handleFilterChange" style="width: 100%;">
              <el-option label="全部课程" value=""></el-option>
              <el-option
                v-for="course in commonCourses"
                :key="course"
                :label="course"
                :value="course"
              />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.isShared" placeholder="共享状态" clearable @change="handleFilterChange" style="width: 100%;">
              <el-option label="全部" value=""></el-option>
              <el-option label="个人题库" value="0"></el-option>
              <el-option label="共享题库" value="1"></el-option>
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-input
              v-model="filters.tag"
              placeholder="标签筛选"
              clearable
              @clear="handleFilterChange"
              @keyup.enter.native="handleFilterChange"
            >
              <el-button slot="append" icon="el-icon-price-tag" @click="handleFilterChange"></el-button>
            </el-input>
          </el-col>
        </el-row>
        <el-row :gutter="15">
          <el-col :span="4">
            <el-select v-model="filters.searchField" placeholder="搜索字段" @change="handleFilterChange" style="width: 100%;">
              <el-option label="题目标题" value="title"></el-option>
              <el-option label="题目ID" value="id"></el-option>
              <el-option label="创建者" value="creator"></el-option>
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-input
              v-model="filters.keyword"
              :placeholder="getSearchPlaceholder()"
              clearable
              @clear="handleFilterChange"
              @keyup.enter.native="handleFilterChange"
            >
              <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
            </el-input>
          </el-col>
        </el-row>
      </div>

      <!-- 题目列表 -->
      <el-table
        :data="questions"
        v-loading="loading"
        stripe
        border
        style="width: 100%; margin-top: 20px"
      >
        <el-table-column type="expand">
          <template slot-scope="{ row }">
            <div class="question-detail">
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="detail-section">
                    <h4>题目内容</h4>
                    <div class="markdown-body" v-html="renderMarkdown(row.title)" v-highlight></div>
                    <div class="markdown-body" v-html="renderMarkdown(row.content)" v-highlight></div>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="detail-section">
                    <h4>选项与答案</h4>
                    <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'">
                      <div v-if="row.options">
                        <div v-for="(opt, idx) in parseOptions(row.options)" :key="idx" class="option-item detail-option-item">
                          <div class="detail-option-head">
                            <el-tag :type="isCorrectAnswer(row.answer, idx) ? 'success' : 'info'" size="small">
                              {{ ['A', 'B', 'C', 'D'][idx] }}
                            </el-tag>
                          </div>
                          <div class="detail-option-content markdown-body" v-html="renderMarkdown(opt)" v-highlight></div>
                        </div>
                      </div>
                      <div class="answer-info">
                        <strong>正确答案：</strong>
                        <el-tag type="success">{{ formatAnswer(row) }}</el-tag>
                      </div>
                    </div>
                    <div v-else-if="row.type === 'judge'">
                      <div class="answer-info">
                        <strong>正确答案：</strong>
                        <el-tag :type="isJudgeTrue(row.answer) ? 'success' : 'warning'">{{ isJudgeTrue(row.answer) ? '正确' : '错误' }}</el-tag>
                      </div>
                    </div>
                    <div v-else-if="row.type === 'subjective'">
                      <div class="answer-info">
                        <strong>参考答案：</strong>
                        <div class="markdown-body" v-html="renderMarkdown(row.answer)" v-highlight></div>
                      </div>
                    </div>

                    <!-- 题目解析 -->
                    <div v-if="row.analysis" class="analysis-info" style="margin-top: 15px;">
                      <h4 style="color: #409EFF; margin-bottom: 8px;">题目解析</h4>
                      <div class="markdown-body" v-html="renderMarkdown(row.analysis)" v-highlight></div>
                    </div>

                    <!-- 题目标签 -->
                    <div v-if="row.tags" class="tags-info" style="margin-top: 10px;">
                      <strong>标签：</strong>
                      <el-tag
                        v-for="(tag, index) in parseTags(row.tags)"
                        :key="index"
                        size="small"
                        style="margin-right: 5px;"
                      >
                        {{ tag }}
                      </el-tag>
                    </div>

                    <!-- 所属课程 -->
                    <div v-if="row.course" class="course-info" style="margin-top: 10px;">
                      <strong>所属课程：</strong>
                      <el-tag type="warning" size="small">{{ row.course }}</el-tag>
                    </div>
                  </div>
                </el-col>
              </el-row>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="title" label="题目标题" min-width="200">
          <template slot-scope="{ row }">
            <div v-html="renderMarkdown(row.title)" class="markdown-body" v-highlight></div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="题型" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="getQuestionTypeColor(row.type)">
              {{ getQuestionTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creatorId" label="创建者" width="120">
          <template slot-scope="{ row }">
            {{ row.creator ? row.creator.username : row.creatorId }}
          </template>
        </el-table-column>
        <el-table-column prop="difficulty" label="难度" width="100">
          <template slot-scope="{ row }">
            <el-rate :value="getDifficultyStars(row.difficulty)" disabled />
          </template>
        </el-table-column>
        <el-table-column prop="course" label="所属课程" width="120">
          <template slot-scope="{ row }">
            <el-tag v-if="row.course" type="warning" size="small">{{ row.course }}</el-tag>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" width="180">
          <template slot-scope="{ row }">
            <el-tag
              v-for="(tag, idx) in parseTags(row.tags)"
              :key="idx"
              size="mini"
              type="info"
              style="margin-right: 3px;"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!row.tags || row.tags === '[]'" style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="score" label="默认分值" width="90"></el-table-column>
        <el-table-column prop="isShared" label="共享状态" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="row.isShared ? 'success' : 'info'">
              {{ row.isShared ? '共享' : '个人' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160">
          <template slot-scope="{ row }">
            {{ row.createTime | localtime }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template slot-scope="{ row }">
            <el-button size="mini" type="primary" @click="goEditPage(row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        class="pagination"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        :current-page="pagination.currentPage"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pagination.pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :total="pagination.total"
      >
      </el-pagination>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      :title="editForm.id ? '编辑题目' : '创建题目'"
      :visible.sync="showEditDialog"
      width="1480px"
      :close-on-click-modal="false"
      class="question-edit-dialog"
    >
      <el-row :gutter="20">
        <el-col :span="14" class="question-form-column">
          <el-form :model="editForm" ref="editForm" label-width="110px" class="question-form">
            <el-form-item label="题型" prop="type">
              <el-select v-model="editForm.type" @change="handleTypeChange">
                <el-option label="单选题" value="single_choice"></el-option>
                <el-option label="多选题" value="multiple_choice"></el-option>
                <el-option label="判断题" value="judge"></el-option>
                <el-option label="主观题" value="subjective"></el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="题目标题" prop="title" required>
              <el-input v-model="editForm.title" placeholder="请输入题目标题"></el-input>
            </el-form-item>
            <el-form-item label="题目内容" prop="content" required>
              <el-input
                type="textarea"
                v-model="editForm.content"
                :rows="7"
                placeholder="请输入题目内容"
              ></el-input>
            </el-form-item>

            <!-- 单选题 -->
            <template v-if="editForm.type === 'single_choice'">
              <el-form-item label="选项" required>
                <div class="options-container">
                  <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                    <el-radio v-model="editForm.correctAnswer" :label="index">
                      {{ ['A', 'B', 'C', 'D'][index] }}
                    </el-radio>
                    <el-input
                      v-model="editForm.choiceOptions[index]"
                      :placeholder="`${['A', 'B', 'C', 'D'][index]}. 选项内容`"
                    ></el-input>
                  </div>
                </div>
              </el-form-item>
            </template>

            <!-- 多选题 -->
            <template v-if="editForm.type === 'multiple_choice'">
              <el-form-item label="选项" required>
                <div class="options-container">
                  <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                    <el-checkbox v-model="editForm.correctAnswers[index]">
                      {{ ['A', 'B', 'C', 'D'][index] }}
                    </el-checkbox>
                    <el-input
                      v-model="editForm.choiceOptions[index]"
                      :placeholder="`${['A', 'B', 'C', 'D'][index]}. 选项内容`"
                    ></el-input>
                  </div>
                </div>
              </el-form-item>
            </template>

            <!-- 判断题 -->
            <template v-if="editForm.type === 'judge'">
              <el-form-item label="正确答案" required>
                <el-radio-group v-model="editForm.correctAnswer">
                  <el-radio label="true">正确</el-radio>
                  <el-radio label="false">错误</el-radio>
                </el-radio-group>
              </el-form-item>
            </template>

            <!-- 主观题 -->
            <template v-if="editForm.type === 'subjective'">
              <el-form-item label="参考答案">
                <el-input
                  type="textarea"
                  v-model="editForm.referenceAnswer"
                  :rows="3"
                  placeholder="请输入参考答案"
                ></el-input>
              </el-form-item>
            </template>

            <!-- 题目解析 -->
            <el-form-item label="题目解析">
              <el-input
                type="textarea"
                v-model="editForm.analysis"
                :rows="2"
                placeholder="请输入题目解析（可选）"
              ></el-input>
            </el-form-item>

            <!-- 题目标签 -->
            <el-form-item label="题目标签">
              <div class="tags-input-container">
                <div class="tags-list">
                  <el-tag
                    v-for="(tag, index) in editForm.tags"
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
                  style="width: 100%;"
                />
              </div>
              <div style="margin-top: 5px; color: #909399; font-size: 12px;">
                <i class="el-icon-info"></i>
                输入标签名称后按回车添加，可添加多个标签
              </div>
            </el-form-item>

            <el-row :gutter="12" class="compact-form-row">
              <el-col :span="12">
                <el-form-item label="所属课程">
                  <el-select
                    v-model="editForm.course"
                    placeholder="请选择课程"
                    style="width: 100%"
                  >
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
                <el-form-item label="难度" prop="difficulty">
                  <el-rate v-model="editForm.difficulty" :max="3"></el-rate>
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12" class="compact-form-row">
              <el-col :span="12">
                <el-form-item label="默认分值" prop="score">
                  <el-input-number v-model="editForm.score" :min="1" :max="100"></el-input-number>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="共享状态" prop="isShared">
                  <el-switch v-model="editForm.isShared" active-text="共享" inactive-text="个人"></el-switch>
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-col>
        <el-col :span="10" class="preview-column">
          <el-card class="preview-card">
            <div slot="header">
              <i class="el-icon-view"></i> 实时预览
            </div>
            <div class="preview-content">
              <div v-if="editForm.title" v-html="renderMarkdown(editForm.title)" class="markdown-body preview-title" v-highlight></div>
              <p v-else class="preview-placeholder">题目标题预览</p>

              <div v-if="editForm.content" v-html="renderMarkdown(editForm.content)" class="markdown-body preview-content-text" v-highlight></div>
              <p v-else class="preview-placeholder">题目内容预览</p>

              <!-- 选项预览 -->
              <div v-if="editForm.type === 'single_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <div v-if="editForm.type === 'multiple_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <div v-if="editForm.type === 'judge'" class="preview-options">
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                  <span>正确</span>
                </div>
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                  <span>错误</span>
                </div>
              </div>

              <div v-if="editForm.type === 'subjective'" class="preview-subjective">
                <el-alert type="info" :closable="false">
                  <i class="el-icon-edit"></i> 主观题，学生需要输入文字答案
                </el-alert>
              </div>

              <!-- 题目解析预览 -->
              <div v-if="editForm.analysis" class="preview-analysis">
                <el-divider content-position="left">
                  <i class="el-icon-document" style="color: #E6A23C;"></i>
                  <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
                </el-divider>
                <div v-html="renderMarkdown(editForm.analysis)" class="markdown-body preview-analysis-content" v-highlight></div>
              </div>
              <p v-else class="preview-placeholder" style="margin-top: 15px;">题目解析预览</p>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="saveQuestion">确定</el-button>
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
md.use(katex)

export default {
  name: 'QuestionBankAdmin',
  data() {
    return {
      loading: false,
      questions: [],
      filters: {
        type: '',
        isShared: '',
        searchField: 'title', // 默认搜索题目标题
        keyword: '',
        course: '',
        tag: ''
      },
      pagination: {
        currentPage: 1,
        pageSize: 20,
        total: 0
      },
      showEditDialog: false,
      tagInput: '', // 标签输入
      editForm: {
        id: null,
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: '', // 题目解析
        tags: [], // 题目标签
        course: '', // 题目所属课程
        difficulty: 1,
        score: 2,
        isShared: false
      },
      // 常用标签
      commonTags: [
        '基础概念',
        '逻辑推理',
        '计算题',
        '应用题',
        '综合分析',
        '易错题',
        '重点',
        '难点'
      ],
      // 常用课程
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
      ]
    }
  },
  mounted() {
    this.loadQuestions()
  },
  watch: {
    '$route.query.refreshTs'() {
      this.loadQuestions()
    }
  },
  methods: {
    async loadQuestions() {
      this.loading = true
      try {
        const res = await this.$http.get('/api/classroom/admin/questions', {
          params: {
            page: this.pagination.currentPage,
            limit: this.pagination.pageSize,
            type: this.filters.type || undefined,
            isShared: this.filters.isShared,
            searchField: this.filters.searchField,
            keyword: this.filters.keyword || undefined,
            course: this.filters.course || undefined,
            tag: this.filters.tag || undefined
          }
        })
        if (res.data.code === 200) {
          this.questions = res.data.data.questions || []
          this.pagination.total = res.data.data.total || 0
        } else {
          this.$message.error(res.data.message || '加载失败')
        }
      } catch (error) {
        this.$message.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    getSearchPlaceholder() {
      const placeholders = {
        title: '搜索题目标题',
        id: '输入题目ID',
        creator: '输入创建者用户名'
      }
      return placeholders[this.filters.searchField] || '请输入搜索关键词'
    },
    handleFilterChange() {
      this.pagination.currentPage = 1
      this.loadQuestions()
    },
    handleSizeChange(val) {
      this.pagination.pageSize = val
      this.loadQuestions()
    },
    handleCurrentChange(val) {
      this.pagination.currentPage = val
      this.loadQuestions()
    },
    goCreatePage() {
      this.$router.push({ name: 'admin-question-bank-create' })
    },
    goEditPage(row) {
      if (!row || !row.id) {
        this.$message.warning('题目ID无效')
        return
      }
      this.$router.push({ name: 'admin-question-bank-edit', params: { questionId: String(row.id) } })
    },
    handleCreate() {
      this.editForm = {
        id: null,
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: '',
        tags: [],
        course: '',
        difficulty: 1,
        score: 2,
        isShared: false
      }
      this.tagInput = ''
      this.showEditDialog = true
    },
    handleEdit(row) {
      this.editForm = {
        id: row.id,
        type: row.type,
        title: row.title,
        content: row.content || '',
        difficulty: row.difficulty,
        score: row.score,
        isShared: row.isShared === 1,
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: row.analysis || '', // 题目解析
        tags: [], // 题目标签
        course: row.course || '' // 题目所属课程
      }

      // 解析标签（从JSON字符串转为数组）
      if (row.tags) {
        try {
          this.editForm.tags = JSON.parse(row.tags)
        } catch (e) {
          this.editForm.tags = []
        }
      }

      // 解析选项和答案
      if (row.type === 'single_choice' || row.type === 'multiple_choice') {
        if (row.options) {
          try {
            const optionsArray = JSON.parse(row.options)
            this.editForm.choiceOptions = optionsArray.map(opt => {
              return opt.replace(/^[A-D]\.\s*/, '')
            })
          } catch (e) {
            console.error('解析选项失败', e)
          }
        }

        if (row.type === 'single_choice') {
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswer = answerMap[row.answer] || 0
        } else {
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswers = [false, false, false, false]
          try {
            const answers = JSON.parse(row.answer)
            if (Array.isArray(answers)) {
              answers.forEach(ans => {
                const idx = answerMap[ans]
                if (idx !== undefined) {
                  this.editForm.correctAnswers[idx] = true
                }
              })
            }
          } catch (e) {
            const answers = row.answer.split(',')
            answers.forEach(ans => {
              const idx = answerMap[ans.trim()]
              if (idx !== undefined) {
                this.editForm.correctAnswers[idx] = true
              }
            })
          }
        }
      } else if (row.type === 'judge') {
        this.editForm.correctAnswer = this.isJudgeTrue(row.answer) ? 'true' : 'false'
      } else if (row.type === 'subjective') {
        this.editForm.referenceAnswer = row.answer || ''
      }

      this.showEditDialog = true
    },
    async handleDelete(row) {
      this.$confirm('确定要删除这道题目吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$http.delete(`/api/classroom/admin/question/${row.id}`)
          if (res.data.code === 200) {
            this.$message.success('删除成功')
            this.loadQuestions()
          } else {
            this.$message.error(res.data.message || '删除失败')
          }
        } catch (error) {
          this.$message.error('删除失败')
        }
      })
    },
    handleTypeChange() {
      if (this.editForm.type === 'single_choice') {
        this.editForm.correctAnswer = 0
        this.editForm.correctAnswers = [false, false, false, false]
        this.editForm.score = 2
      } else if (this.editForm.type === 'multiple_choice') {
        this.editForm.correctAnswer = 0
        this.editForm.correctAnswers = [false, false, false, false]
        this.editForm.score = 5
      } else if (this.editForm.type === 'judge') {
        this.editForm.score = 1
        this.editForm.correctAnswer = 'true'
        this.editForm.correctAnswers = [false, false, false, false]
      } else if (this.editForm.type === 'subjective') {
        this.editForm.score = 5
        this.editForm.correctAnswer = ''
        this.editForm.correctAnswers = [false, false, false, false]
      }
      this.editForm.referenceAnswer = ''
    },
    // 添加标签
    addTag() {
      const tag = this.tagInput.trim()
      if (tag && !this.editForm.tags.includes(tag)) {
        this.editForm.tags.push(tag)
      }
      this.tagInput = '' // 清空输入
    },
    // 删除标签
    removeTag(index) {
      this.editForm.tags.splice(index, 1)
    },
    async saveQuestion() {
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
        analysis: this.editForm.analysis || '', // 题目解析
        tags: JSON.stringify(this.editForm.tags || []), // 题目标签（JSON格式）
        course: this.editForm.course || '', // 题目所属课程
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
        const selectedAnswers = this.editForm.correctAnswers
          .map((selected, idx) => selected ? ['A', 'B', 'C', 'D'][idx] : null)
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.editForm.type === 'judge') {
        const answerValue = String(this.editForm.correctAnswer)
        submitData.answer = answerValue === 'true' ? 'true' : 'false'
        submitData.options = null
      } else if (this.editForm.type === 'subjective') {
        submitData.answer = this.editForm.referenceAnswer || '需人工评分'
        submitData.options = null
      }

      try {
        let res
        if (this.editForm.id) {
          // 更新
          res = await this.$http.put(`/api/classroom/admin/question/${this.editForm.id}`, submitData)
        } else {
          // 创建
          res = await this.$http.post('/api/classroom/admin/question', submitData)
        }

        if (res.data.code === 200) {
          this.$message.success(this.editForm.id ? '更新成功' : '创建成功')
          this.showEditDialog = false
          this.loadQuestions()
        } else {
          this.$message.error(res.data.message || '操作失败')
        }
      } catch (error) {
        this.$message.error('操作失败')
      }
    },
    parseOptions(optionsStr) {
      try {
        return JSON.parse(optionsStr)
      } catch (e) {
        return []
      }
    },
    parseTags(tagsStr) {
      try {
        return JSON.parse(tagsStr)
      } catch (e) {
        return []
      }
    },
    isCorrectAnswer(answer, index) {
      const label = ['A', 'B', 'C', 'D'][index]
      if (answer.startsWith('[')) {
        try {
          const answers = JSON.parse(answer)
          return answers.includes(label)
        } catch (e) {
          return false
        }
      }
      return answer === label
    },
    formatAnswer(row) {
      if (row.type === 'single_choice') {
        return row.answer
      } else if (row.type === 'multiple_choice') {
        try {
          return JSON.parse(row.answer).join(', ')
        } catch (e) {
          return row.answer
        }
      } else if (row.type === 'judge') {
        return this.isJudgeTrue(row.answer) ? '正确' : '错误'
      } else if (row.type === 'subjective') {
        return '需人工评分'
      }
      return row.answer
    },
    isJudgeTrue(answer) {
      const raw = String(answer || '').trim()
      const lowered = raw.toLowerCase()
      return lowered === 'true'
    },
    getQuestionTypeName(type) {
      const map = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        subjective: '主观题'
      }
      return map[type] || type
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
    getDifficultyStars(difficulty) {
      const difficultyMap = {
        1: 1,
        2: 3,
        3: 5
      }
      return difficultyMap[difficulty] || 3
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
.question-bank-admin-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
}

.filter-bar {
  margin-bottom: 20px;
}

.question-form-column,
.preview-column {
  max-height: 72vh;
  overflow-y: auto;
}

.question-form-column {
  padding-right: 6px;
}

.preview-column {
  padding-left: 6px;
}

.question-form .el-form-item {
  margin-bottom: 14px;
}

.compact-form-row .el-form-item {
  margin-bottom: 10px;
}

.question-detail {
  padding: 20px;
  background-color: #f5f7fa;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  margin-bottom: 10px;
  color: #303133;
  font-size: 16px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  padding: 8px;
  background: white;
  border-radius: 4px;
}

.detail-option-item {
  display: block;
}

.detail-option-head {
  margin-bottom: 6px;
}

.detail-option-content {
  width: 100%;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.detail-option-content p {
  margin: 0;
}

.detail-option-content pre {
  margin: 0;
  max-width: 100%;
  overflow-x: auto;
}

.answer-info {
  margin-top: 15px;
  padding: 10px;
  background: #f0f9ff;
  border-left: 3px solid #409eff;
  border-radius: 4px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.options-container {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 14px;
}

.options-container .option-item {
  margin-bottom: 0;
  align-items: flex-start;
}

.preview-card {
  min-height: 100%;
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

.preview-subjective {
  margin-top: 15px;
}

.markdown-body {
  word-wrap: break-word;
  word-break: break-word;
}

/* 标签输入容器样式 */
.tags-input-container {
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 5px;
  min-height: 80px;
}

.tags-list {
  margin-bottom: 8px;
}

/* 题目解析预览样式 */
.preview-analysis {
  margin-top: 15px;
  padding: 10px;
  background: #fff9e6;
  border-left: 3px solid #E6A23C;
  border-radius: 4px;
}

.preview-analysis-content {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
}

@media (max-width: 1280px) {
  .options-container {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
.markdown-body h1 {
  font-size: 1.86em;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.3em;
  margin-bottom: 16px;
}

.markdown-body h2 {
  font-size: 1.45em;
  border-bottom: 1px solid #eee;
  background: #cce5ff;
  padding: 8px 10px;
  margin-bottom: 16px;
}

.markdown-body h3 {
  font-size: 1.3em;
  border-left: 4px solid #03a9f4;
  padding-left: 6px;
  margin-bottom: 16px;
}

.markdown-body p {
  margin-bottom: 16px;
}

.markdown-body code {
  background: #f8f8f9;
  padding: 2px 6px;
  border-radius: 3px;
}

.markdown-body pre {
  padding: 5px 10px;
  background: #f8f8f9;
  border: 1px dashed #e9eaec;
  border-radius: 3px;
  margin: 15px 0;
}
</style>
