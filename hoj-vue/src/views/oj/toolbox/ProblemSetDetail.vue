<template>
  <div class="problem-set-detail-container">
    <el-card v-loading="loading" class="detail-card">
      <div slot="header" class="detail-header">
        <div class="header-left">
          <el-button icon="el-icon-arrow-left" @click="goBack">返回</el-button>
          <span class="header-title">{{ isCreate ? '创建题目集' : '编辑题目集' }}</span>
        </div>
        <div class="header-right">
          <el-button @click="toggleView" :type="viewMode === 'preview' ? 'primary' : ''">
            <i :class="viewMode === 'edit' ? 'fa fa-eye' : 'fa fa-edit'"></i>
            {{ viewMode === 'edit' ? '预览' : '编辑' }}
          </el-button>
          <el-button @click="saveSet" :loading="saving">保存</el-button>
          <el-button v-if="!isCreate" type="success" @click="generatePDF" :loading="generating">
            <i class="fa fa-download"></i> 导出 PDF
          </el-button>
        </div>
      </div>

      <!-- 编辑模式 -->
      <div v-show="viewMode === 'edit'">
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="题集名称" prop="title">
                <el-input v-model="form.title" placeholder="请输入题集名称"></el-input>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="作者" prop="author">
                <el-input v-model="form.author" placeholder="请输入作者"></el-input>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="比赛日期" prop="contest_date">
                <el-date-picker
                  v-model="form.contest_date"
                  type="date"
                  placeholder="选择日期"
                  value-format="yyyy-MM-dd"
                  style="width: 100%;"
                >
                </el-date-picker>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>

        <el-divider>题目列表</el-divider>

        <!-- 图片管理（仅在非创建模式且已保存后显示） -->
        <div v-if="!isCreate && setId" class="images-section">
          <el-divider>图片管理</el-divider>
          <el-button type="primary" icon="el-icon-upload" @click="showImageUploadDialog = true" style="margin-bottom: 15px;">
            上传图片
          </el-button>
          <el-button icon="el-icon-refresh" @click="loadImages" style="margin-bottom: 15px;">
            刷新列表
          </el-button>

          <el-empty v-if="images.length === 0" description="暂无图片，点击上传按钮添加"></el-empty>

          <div v-else class="images-list">
            <el-card
              v-for="image in images"
              :key="image.id"
              class="image-card"
            >
              <div class="image-item">
                <div class="image-info">
                  <i class="fa fa-image"></i>
                  <span class="image-filename">{{ image.filename }}</span>
                  <span class="image-size">({{ formatFileSize(image.file_size) }})</span>
                </div>
                <el-button
                  type="danger"
                  size="mini"
                  icon="el-icon-delete"
                  @click="deleteImage(image.id)"
                >
                  删除
                </el-button>
              </div>
              <div class="image-tip">
                在题面中使用: <code>\includegraphics[width=9cm]{{ image.filename }}</code>
              </div>
            </el-card>
          </div>
        </div>

        <div class="problems-section">
          <el-button type="primary" icon="el-icon-plus" @click="addProblem" style="margin-bottom: 15px;">
            添加题目
          </el-button>

          <el-empty v-if="form.problems.length === 0" description="暂无题目，点击上方按钮添加"></el-empty>

          <el-collapse v-else v-model="activeProblems" accordion>
            <el-collapse-item
              v-for="(problem, index) in form.problems"
              :key="problem.id || 'temp_' + index"
              :name="index"
            >
              <template slot="title">
                <div class="problem-title">
                  <span>{{ problem.problem_letter || String.fromCharCode(65 + index) }}. {{ problem.title || '未命名题目' }}</span>
                  <div class="problem-actions">
                    <el-button
                      type="primary"
                      size="mini"
                      icon="el-icon-top"
                      @click.stop="moveProblemUp(index)"
                      :disabled="index === 0"
                      title="上移"
                    >
                      上移
                    </el-button>
                    <el-button
                      type="primary"
                      size="mini"
                      icon="el-icon-bottom"
                      @click.stop="moveProblemDown(index)"
                      :disabled="index === form.problems.length - 1"
                      title="下移"
                    >
                      下移
                    </el-button>
                    <el-button
                      type="danger"
                      size="mini"
                      icon="el-icon-delete"
                      @click.stop="removeProblem(index)"
                    >
                      删除
                    </el-button>
                  </div>
                </div>
              </template>

              <el-form :model="problem" label-width="100px" class="problem-form">
                <el-row :gutter="20">
                  <el-col :span="8">
                    <el-form-item label="题目字母">
                      <el-input v-model="problem.problem_letter" placeholder="A-Z"></el-input>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="题目标题">
                      <el-input v-model="problem.title" placeholder="请输入题目标题"></el-input>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="时间限制(ms)">
                      <el-input-number v-model="problem.time_limit" :min="1" :max="10000" style="width: 100%;"></el-input-number>
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-form-item label="题目描述">
                  <el-input
                    type="textarea"
                    v-model="problem.description"
                    :rows="4"
                    placeholder="请输入题目描述（支持 LaTeX 公式，如 $x^2$）"
                  ></el-input>
                </el-form-item>

                <el-form-item label="输入格式">
                  <el-input
                    type="textarea"
                    v-model="problem.input_format"
                    :rows="3"
                    placeholder="请输入输入格式"
                  ></el-input>
                </el-form-item>

                <el-form-item label="输出格式">
                  <el-input
                    type="textarea"
                    v-model="problem.output_format"
                    :rows="3"
                    placeholder="请输入输出格式"
                  ></el-input>
                </el-form-item>

                <el-form-item label="注意事项">
                  <el-input
                    type="textarea"
                    v-model="problem.note"
                    :rows="3"
                    placeholder="请输入注意事项（可选）"
                  ></el-input>
                </el-form-item>

                <el-divider>样例</el-divider>

                <div class="examples-section">
                  <el-button size="small" icon="el-icon-plus" @click="addExample(problem)">
                    添加样例
                  </el-button>

                  <el-empty v-if="!problem.examples || problem.examples.length === 0" description="暂无样例" :image-size="80"></el-empty>

                  <div v-else class="examples-list">
                    <el-card
                      v-for="(example, exIndex) in problem.examples"
                      :key="example.id || 'temp_ex_' + index + '_' + exIndex"
                      class="example-card"
                    >
                      <div slot="header" class="example-header">
                        <span>样例 {{ example.example_no || exIndex + 1 }}</span>
                        <el-button
                          type="danger"
                          size="mini"
                          icon="el-icon-delete"
                          @click="removeExample(problem, exIndex)"
                        >
                          删除
                        </el-button>
                      </div>

                      <el-form :model="example" label-width="80px" size="small" v-if="example">
                        <el-form-item label="样例描述">
                          <el-input
                            v-model="example.description"
                            placeholder="样例描述（可选）"
                          ></el-input>
                        </el-form-item>

                        <el-form-item label="样例输入">
                          <el-input
                            type="textarea"
                            v-model="example.input"
                            :rows="4"
                            placeholder="请输入样例输入"
                          ></el-input>
                        </el-form-item>

                        <el-form-item label="样例输出">
                          <el-input
                            type="textarea"
                            v-model="example.output"
                            :rows="4"
                            placeholder="请输入样例输出"
                          ></el-input>
                        </el-form-item>
                      </el-form>
                    </el-card>
                  </div>
                </div>
              </el-form>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>

      <!-- 预览模式 -->
      <div v-show="viewMode === 'preview'" class="preview-container">
        <div class="preview-content">
          <!-- 封面页 -->
          <div class="preview-cover-page">
            <h1 class="preview-title">{{ form.title || '题目集标题' }}</h1>
            <h2 class="preview-subtitle">Problem Set</h2>
            <p class="preview-author">{{ form.author || '作者' }}</p>
            <p class="preview-date">{{ formatDate(form.contest_date) }}</p>
            <div class="preview-problem-list">
              <div v-for="(problem, index) in form.problems" :key="index" class="preview-problem-item">
                {{ problem.problem_letter || String.fromCharCode(65 + index) }} &nbsp;&nbsp;&nbsp;&nbsp; {{ problem.title || '未命名' }}
              </div>
            </div>
          </div>

          <!-- 题目预览 -->
          <div v-for="(problem, index) in form.problems" :key="index" class="preview-problem-page">
            <!-- 题目标题：两行显示 -->
            <div class="problem-header">
              <div class="problem-title-line">Problem {{ problem.problem_letter || String.fromCharCode(65 + index) }}</div>
              <div class="problem-title-line">{{ problem.title }}</div>
              <p class="time-limit">Time limit: {{ problem.time_limit || 1000 }} ms</p>
            </div>

            <div class="problem-section">
              <h3>Description</h3>
              <div class="problem-content" v-html="renderMath(problem.description || '')"></div>
            </div>

            <div class="problem-section">
              <h3>Input</h3>
              <div class="problem-content" v-html="renderMath(problem.input_format || '')"></div>
            </div>

            <div class="problem-section">
              <h3>Output</h3>
              <div class="problem-content" v-html="renderMath(problem.output_format || '')"></div>
            </div>

            <div v-if="problem.examples && problem.examples.length > 0" class="problem-examples">
              <div v-for="(example, exIndex) in problem.examples" :key="'example-' + exIndex">
                <!-- 样例：标题在框外上方，内容在单个框内，中间用竖线分割 -->
                <div class="example-container-title">
                  <span class="example-title-left">Sample Input {{ example.example_no || exIndex + 1 }}</span>
                  <span class="example-title-right">Sample Output {{ example.example_no || exIndex + 1 }}</span>
                </div>
                <div class="example-box-single">
                  <div class="example-half-single">
                    <pre class="example-content-single">{{ example.input || '' }}</pre>
                  </div>
                  <div class="example-divider"></div>
                  <div class="example-half-single">
                    <pre class="example-content-single">{{ example.output || '' }}</pre>
                  </div>
                </div>
                <div v-if="example && example.description" class="example-description">
                  <strong>Explanation:</strong> {{ example.description }}
                </div>
              </div>
            </div>

            <div v-if="problem.note" class="problem-section">
              <h3>Note</h3>
              <div class="problem-content" v-html="renderMath(problem.note || '')"></div>
            </div>

            <!-- 页码 -->
            <div class="preview-page-number">{{ index + 1 }}</div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 图片上传对话框 -->
    <el-dialog
      title="上传图片"
      :visible.sync="showImageUploadDialog"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-upload
        ref="uploadRef"
        :action="uploadUrl"
        :headers="uploadHeaders"
        :on-success="onUploadSuccess"
        :on-error="onUploadError"
        :before-upload="beforeUpload"
        :file-list="uploadFileList"
        :auto-upload="false"
        drag
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <div class="el-upload__tip" slot="tip">
          只能上传 jpg/png/gif/bmp/pdf 文件，且不超过 10MB
        </div>
      </el-upload>
      <span slot="footer" class="dialog-footer">
        <el-button @click="showImageUploadDialog = false">取 消</el-button>
        <el-button type="primary" @click="submitUpload" :loading="uploading">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { problemSetApi } from '@/api/problemSet'
import katex from 'katex'
import 'katex/dist/katex.min.css'

export default {
  name: 'ProblemSetDetail',
  data() {
    return {
      loading: false,
      saving: false,
      generating: false,
      isCreate: false,
      setId: null,
      viewMode: 'edit', // 'edit' or 'preview'
      activeProblems: [],
      form: {
        title: '',
        author: '',
        contest_date: '',
        problems: []
      },
      rules: {
        title: [
          { required: true, message: '请输入题集名称', trigger: 'blur' }
        ]
      },
      // 图片相关
      images: [],
      showImageUploadDialog: false,
      uploading: false,
      uploadFileList: [],
      uploadUrl: '',
      uploadHeaders: {}
    }
  },
  mounted() {
    // 检查当前路由名称，判断是创建还是编辑
    if (this.$route.name === 'ProblemSetCreate') {
      this.isCreate = true
      this.setId = null
    } else {
      this.setId = this.$route.params.id
      if (!this.setId) {
        this.$message.error('缺少题目集ID')
        this.goBack()
        return
      }
      this.loadProblemSet()
      this.loadImages()
      this.uploadUrl = `${process.env.VUE_APP_BASE_API || 'http://localhost:8888'}/api/problem-set/${this.setId}/images`
      this.uploadHeaders = {
        'Authorization': localStorage.getItem('token')
      }
    }
  },
  methods: {
    toggleView() {
      this.viewMode = this.viewMode === 'edit' ? 'preview' : 'edit'
    },

    renderMath(text) {
      if (!text) return ''
      // 确保 text 是字符串
      const safeText = String(text)
      // 简单的 LaTeX 公式渲染，支持 $...$ 和 $$...$$
      let result = safeText
      // 渲染行内公式 $...$
      result = result.replace(/\$([^\$]+)\$/g, (match, formula) => {
        try {
          return katex.renderToString(formula, { throwOnError: false, displayMode: false })
        } catch (e) {
          return match
        }
      })
      // 渲染块级公式 $$...$$
      result = result.replace(/\$\$([^\$]+)\$\$/g, (match, formula) => {
        try {
          return katex.renderToString(formula, { throwOnError: false, displayMode: true })
        } catch (e) {
          return match
        }
      })
      // 处理换行
      result = result.replace(/\n/g, '<br>')
      return result
    },

    formatDate(dateStr) {
      if (!dateStr) return new Date().toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      })
      const date = new Date(dateStr)
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      })
    },

    async loadProblemSet() {
      if (!this.setId) {
        this.$message.error('缺少题目集ID')
        this.goBack()
        return
      }
      this.loading = true
      try {
        const res = await problemSetApi.getProblemSet(this.setId)
        console.log('API Response:', res)
        if (res.code === 200) {
          const data = res.data
          console.log('ProblemSet Data:', data)
          console.log('Problems:', data.problems)

          // 深拷贝 problems 数组以确保 Vue 响应式
          const problems = (data.problems || []).map(p => ({
            id: p.id,
            set_id: p.set_id,
            problem_letter: p.problem_letter || '',
            title: p.title || '',
            time_limit: p.time_limit || 1000,
            description: p.description || '',
            input_format: p.input_format || '',
            output_format: p.output_format || '',
            note: p.note || '',
            sort_order: p.sort_order || 0,
            created_at: p.created_at,
            updated_at: p.updated_at,
            examples: (p.examples || []).map(e => ({
              id: e.id,
              problem_id: e.problem_id,
              example_no: e.example_no,
              description: e.description || '',
              input: e.input || '',
              output: e.output || '',
              created_at: e.created_at
            }))
          }))

          this.form = {
            title: data.title || '',
            author: data.author || '',
            contest_date: data.contest_date ? data.contest_date.substring(0, 10) : '',
            problems: problems
          }

          console.log('Form after load:', this.form)
          console.log('Problems count:', this.form.problems.length)
          console.log('First problem:', this.form.problems[0])
        }
      } catch (error) {
        console.error('Load error:', error)
        this.$message.error('加载题目集失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
      }
    },

    async saveSet() {
      // 确保 formRef 存在
      if (!this.$refs.formRef) {
        this.$message.error('表单未初始化')
        return
      }

      this.$refs.formRef.validate(async (valid) => {
        if (!valid) return

        this.saving = true
        try {
          // 验证题目信息
          for (let i = 0; i < this.form.problems.length; i++) {
            const problem = this.form.problems[i]
            if (!problem || !problem.title) {
              this.$message.error(`第 ${i + 1} 道题的标题不能为空`)
              this.saving = false
              return
            }
            if (problem.examples && problem.examples.length > 0) {
              for (let j = 0; j < problem.examples.length; j++) {
                const example = problem.examples[j]
                if (!example || !example.input || !example.output) {
                  this.$message.error(`第 ${i + 1} 道题的第 ${j + 1} 个样例的输入输出不能为空`)
                  this.saving = false
                  return
                }
              }
            }
          }

          const data = {
            title: this.form.title,
            author: this.form.author,
            contest_date: this.form.contest_date
          }

          let res
          if (this.isCreate) {
            // 创建新题目集
            res = await problemSetApi.createProblemSet(data)
            if (res.code === 200) {
              this.setId = res.data.id
              this.isCreate = false
              this.$message.success('创建成功')
              // 保存题目
              await this.saveProblems()
            }
          } else {
            // 更新题目集
            if (!this.setId) {
              this.$message.error('缺少题目集ID')
              this.saving = false
              return
            }
            res = await problemSetApi.updateProblemSet(this.setId, data)
            if (res.code === 200) {
              // 也需要保存题目修改
              await this.saveProblems()
              this.$message.success('保存成功')
            }
          }
        } catch (error) {
          this.$message.error('保存失败: ' + (error.message || '未知错误'))
        } finally {
          this.saving = false
        }
      })
    },

    async saveProblems() {
      if (!this.setId) return

      for (const problem of this.form.problems) {
        if (problem.id) {
          // 更新现有题目
          await problemSetApi.updateProblem(this.setId, problem.id, problem)
          // 保存样例
          if (problem.examples) {
            for (const example of problem.examples) {
              if (example.id) {
                await problemSetApi.updateExample(this.setId, problem.id, example.id, example)
              } else {
                await problemSetApi.createExample(this.setId, problem.id, example)
              }
            }
          }
        } else {
          // 创建新题目
          const problemData = {
            title: problem.title,
            problem_letter: problem.problem_letter,
            time_limit: problem.time_limit || 1000,
            description: problem.description,
            input_format: problem.input_format,
            output_format: problem.output_format,
            note: problem.note,
            sort_order: problem.sort_order || 0
          }
          const res = await problemSetApi.createProblem(this.setId, problemData)
          if (res.code === 200) {
            problem.id = res.data.id
            // 保存样例
            if (problem.examples && problem.examples.length > 0) {
              for (const example of problem.examples) {
                await problemSetApi.createExample(this.setId, problem.id, example)
              }
            }
          }
        }
      }
    },

    addProblem() {
      this.form.problems.push({
        problem_letter: String.fromCharCode(65 + this.form.problems.length),
        title: '',
        time_limit: 1000,
        description: '',
        input_format: '',
        output_format: '',
        note: '',
        examples: []
      })
      this.activeProblems = [this.form.problems.length - 1]
    },

    async moveProblemUp(index) {
      if (index === 0) {
        this.$message.warning('已经是第一题了')
        return
      }

      const problem = this.form.problems[index]
      const prevProblem = this.form.problems[index - 1]

      // 如果题目已保存到数据库，调用 API
      if (this.setId && problem.id) {
        try {
          await problemSetApi.moveProblemUp(this.setId, problem.id)
          this.$message.success('上移成功')
          // 重新加载题目集
          await this.loadProblemSet()
          return
        } catch (error) {
          this.$message.error('上移失败')
          return
        }
      }

      // 本地移动
      const temp = this.form.problems[index]
      this.$set(this.form.problems, index, this.form.problems[index - 1])
      this.$set(this.form.problems, index - 1, temp)
      this.$forceUpdate()
    },

    async moveProblemDown(index) {
      if (index === this.form.problems.length - 1) {
        this.$message.warning('已经是最后一题了')
        return
      }

      const problem = this.form.problems[index]
      const nextProblem = this.form.problems[index + 1]

      // 如果题目已保存到数据库，调用 API
      if (this.setId && problem.id) {
        try {
          await problemSetApi.moveProblemDown(this.setId, problem.id)
          this.$message.success('下移成功')
          // 重新加载题目集
          await this.loadProblemSet()
          return
        } catch (error) {
          this.$message.error('下移失败')
          return
        }
      }

      // 本地移动
      const temp = this.form.problems[index]
      this.$set(this.form.problems, index, this.form.problems[index + 1])
      this.$set(this.form.problems, index + 1, temp)
      this.$forceUpdate()
    },

    removeProblem(index) {
      if (!this.form.problems || !this.form.problems[index]) {
        this.$message.warning('题目不存在')
        return
      }
      if (!this.setId && this.form.problems[index].id) {
        this.$message.warning('请先保存题目集')
        return
      }
      this.$confirm('确定要删除这道题目吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        if (this.form.problems[index] && this.form.problems[index].id) {
          problemSetApi.deleteProblem(this.setId, this.form.problems[index].id)
        }
        this.form.problems.splice(index, 1)
        // 重新计算题目字母
        this.form.problems.forEach((p, i) => {
          if (p && !p.problem_letter) {
            p.problem_letter = String.fromCharCode(65 + i)
          }
        })
      }).catch(() => {})
    },

    addExample(problem) {
      // 使用展开运算符确保 Vue 响应式更新
      const currentExamples = problem.examples || []
      const newExample = {
        example_no: currentExamples.length + 1,
        description: '',
        input: '',
        output: ''
      }
      this.$set(problem, 'examples', [...currentExamples, newExample])
    },

    removeExample(problem, index) {
      // 确保 examples 存在
      if (!problem.examples || !problem.examples[index]) {
        this.$message.warning('样例不存在')
        return
      }
      if (!this.setId && problem.id && problem.examples[index].id) {
        this.$message.warning('请先保存题目集')
        return
      }
      if (problem.examples[index].id) {
        problemSetApi.deleteExample(this.setId, problem.id, problem.examples[index].id)
      }
      problem.examples.splice(index, 1)
      // 重新计算样例编号
      if (problem.examples) {
        problem.examples.forEach((e, i) => {
          if (e) {
            e.example_no = i + 1
          }
        })
      }
    },

    async generatePDF() {
      if (this.form.problems.length === 0) {
        this.$message.warning('请至少添加一道题目')
        return
      }

      this.generating = true
      try {
        // 使用新的API：从前端数据直接生成PDF（所见即所得，与预览保持一致）
        const pdfData = {
          title: this.form.title,
          author: this.form.author,
          contest_date: this.form.contest_date,
          problems: this.form.problems.map((p, index) => ({
            problem_letter: p.problem_letter || String.fromCharCode(65 + index),
            title: p.title,
            time_limit: parseInt(p.time_limit) || 1000,  // 确保是数字类型
            description: p.description || '',
            input_format: p.input_format || '',
            output_format: p.output_format || '',
            note: p.note || '',
            sort_order: p.sort_order !== undefined ? p.sort_order : index,
            examples: (p.examples || []).map(ex => ({
              example_no: parseInt(ex.example_no) || 1,
              description: ex.description || '',
              input: ex.input || '',
              output: ex.output || ''
            }))
          }))
        }

        console.log('[生成PDF] 发送数据:', pdfData)

        const res = await problemSetApi.generatePDFFromData(pdfData)

        // 创建下载链接
        const blob = new Blob([res], { type: 'application/pdf' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `${this.form.title || '题目集'}.pdf`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)

        this.$message.success('PDF 生成成功（内容与预览一致）')
      } catch (error) {
        console.error('生成PDF失败:', error)
        this.$message.error('生成 PDF 失败')
      } finally {
        this.generating = false
      }
    },

    goBack() {
      this.$router.push('/toolbox/problem-set')
    },

    // 加载图片列表
    async loadImages() {
      if (!this.setId) return
      try {
        const res = await problemSetApi.listImages(this.setId)
        if (res.code === 200) {
          this.images = res.data || []
        }
      } catch (error) {
        console.error('加载图片列表失败:', error)
      }
    },

    // 格式化文件大小
    formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
    },

    // 上传前验证
    beforeUpload(file) {
      const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/bmp', 'application/pdf']
      const isAllowed = allowedTypes.includes(file.type)
      const isLt10M = file.size / 1024 / 1024 < 10

      if (!isAllowed) {
        this.$message.error('只能上传 JPG/PNG/GIF/BMP/PDF 格式的文件!')
        return false
      }
      if (!isLt10M) {
        this.$message.error('上传文件大小不能超过 10MB!')
        return false
      }
      return true
    },

    // 提交上传
    submitUpload() {
      this.$refs.uploadRef.submit()
    },

    // 上传成功
    onUploadSuccess(response, file, fileList) {
      this.uploading = false
      if (response.code === 200) {
        this.$message.success('上传成功')
        this.showImageUploadDialog = false
        this.uploadFileList = []
        this.loadImages()
      } else {
        this.$message.error(response.msg || '上传失败')
      }
    },

    // 上传失败
    onUploadError(error, file, fileList) {
      this.uploading = false
      this.$message.error('上传失败: ' + (error.message || '未知错误'))
    },

    // 删除图片
    deleteImage(imageId) {
      this.$confirm('确定要删除这张图片吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          await problemSetApi.deleteImage(this.setId, imageId)
          this.$message.success('删除成功')
          this.loadImages()
        } catch (error) {
          this.$message.error('删除失败: ' + (error.message || '未知错误'))
        }
      }).catch(() => {})
    }
  },
  watch: {
    // 当 setId 变化时加载图片
    setId(newVal) {
      if (newVal) {
        this.loadImages()
        this.uploadUrl = `${process.env.VUE_APP_BASE_API}/api/problem-set/${newVal}/images`
        this.uploadHeaders = {
          'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
      }
    }
  }
}
</script>

<style scoped>
.problem-set-detail-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.detail-card {
  min-height: 600px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.header-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  gap: 10px;
}

.problems-section {
  margin-top: 20px;
}

/* 图片管理样式 */
.images-section {
  margin-top: 20px;
}

.images-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 15px;
  margin-top: 15px;
}

.image-card {
  border-radius: 8px;
}

.image-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.image-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.image-info i {
  font-size: 24px;
  color: #409EFF;
}

.image-filename {
  font-weight: 600;
  color: #303133;
}

.image-size {
  color: #909399;
  font-size: 12px;
}

.image-tip {
  margin-top: 10px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 12px;
  color: #606266;
}

.image-tip code {
  background: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  color: #E6A23C;
  font-family: 'Courier New', monospace;
}

.problem-title {
  display: flex;
  align-items: center;
  width: 100%;
  font-weight: 600;
  font-size: 16px;
}

.problem-actions {
  display: flex;
  gap: 5px;
  margin-left: auto;
}


.problem-form {
  padding: 20px;
}

.examples-section {
  margin-top: 15px;
}

.examples-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
  gap: 15px;
  margin-top: 15px;
}

.example-card {
  border-radius: 8px;
}

.example-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

/* 预览样式 */
.preview-container {
  background: #fff;
  padding: 40px;
  min-height: 800px;
}

.preview-content {
  max-width: 800px;
  margin: 0 auto;
  font-family: 'Times New Roman', serif;
}

.preview-cover-page {
  text-align: center;
  padding: 100px 0;
  page-break-after: always;
}

.preview-title {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 10px;
}

.preview-subtitle {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 40px;
}

.preview-author {
  font-size: 20px;
  margin-bottom: 20px;
}

.preview-date {
  font-size: 18px;
  margin-bottom: 80px;
}

.preview-problem-list {
  display: inline-block;
  text-align: left;
  margin-top: 40px;
  font-size: 18px;
  line-height: 2;
}

.preview-problem-item {
  font-weight: bold;
}

/* 题目标题：两行显示，居中 */
.problem-header {
  text-align: center;
  margin-bottom: 30px;
}

.problem-title-line {
  font-size: 28px;
  font-weight: bold;
  margin: 5px 0;
}

.problem-title-line:first-child {
  font-size: 32px;
  margin-bottom: 8px;
}

.time-limit {
  font-size: 16px;
  margin-top: 10px;
}

.problem-section {
  margin-bottom: 25px;
}

.problem-section h3 {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 10px;
}

.problem-content {
  line-height: 1.6;
  white-space: pre-wrap;
}

.problem-examples {
  margin: 25px 0;
}

/* 题目页分页样式 */
.preview-problem-page {
  position: relative;
  min-height: calc(100vh - 200px);
  padding: 60px 40px;
  margin-bottom: 40px;
  page-break-after: always;
  break-after: page;
}

/* 页码样式 */
.preview-page-number {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 14px;
  color: #000;
}

/* 样例标题行：左右对齐 */
.example-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.example-col {
  flex: 1;
}

.example-title {
  font-weight: bold;
  font-size: 16px;
}

.example-title-right {
  font-weight: bold;
  font-size: 16px;
  text-align: right;
}

/* 样例内容行：单层边框，中间竖线 */
.example-content-row {
  display: flex;
  border: 1px solid #000;
  border-bottom: none;
  margin-bottom: 20px;
}

.example-content-box {
  flex: 1;
  padding: 10px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  white-space: pre-wrap;
  background: #fff;
}

.example-content-box:first-child {
  border-right: 1px solid #000;
}

/* 样例布局：标题在框外上方，内容在单个框内，中间用竖线分割 */
.example-container-title {
  display: flex;
  margin-bottom: 8px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  font-weight: bold;
}

.example-title-left {
  flex: 0 0 48%;
  text-align: left;
}

.example-title-right {
  flex: 0 0 48%;
  text-align: left;
  margin-left: 4%;
}

.example-box-single {
  display: flex;
  border: 1px solid #000;
  background: #fff;
  position: relative;
}

.example-half-single {
  flex: 1;
  padding: 6px 10px;
  min-width: 0; /* 防止内容溢出 */
  box-sizing: border-box;
}

/* 使用绝对定位的分隔线，确保从上到下完全连接，无空白 */
.example-divider {
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 1px;
  background-color: #000;
  transform: translateX(-50%);
  z-index: 1;
}

.example-content-single {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-wrap: break-word;
  overflow-wrap: break-word;
  line-height: 1.4;
  position: relative;
  z-index: 0;
}

.example-description {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.6;
}

/* KaTeX 公式样式 */
:deep(.katex) {
  font-size: 1.1em;
}

:deep(.katex-display) {
  margin: 1em 0;
}

@media screen and (max-width: 768px) {
  .problem-set-detail-container {
    padding: 10px;
  }

  .detail-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }

  .examples-list {
    grid-template-columns: 1fr;
  }

  .preview-container {
    padding: 20px;
  }

  /* 移动端：样例框改为上下布局 */
  .example-box-single {
    flex-direction: column;
  }

  .example-half-single:first-child {
    border-right: none;
    border-bottom: 1px solid #000;
  }
}
</style>
