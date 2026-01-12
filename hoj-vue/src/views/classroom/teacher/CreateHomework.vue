<template>
  <div class="create-homework">
    <div class="header">
      <h2>{{ isEditMode ? $t('m.Edit_Homework') : $t('m.Create_Homework') }}</h2>
      <div class="actions">
        <el-button @click="goBack">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="saveHomework" :loading="submitting">
          {{ isEditMode ? $t('m.Save') : $t('m.Create') }}
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
          </el-select>
          <el-button type="primary" icon="el-icon-search" style="margin-left: 10px" @click="loadQuestionBank">
            {{ $t('m.Search') }}
          </el-button>
        </div>

        <div class="question-list" v-loading="questionsLoading">
          <el-table
            :data="questionBank"
            @selection-change="handleSelectionChange"
            stripe
            style="width: 100%"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="title" :label="$t('m.Question_Title')" min-width="200" />
            <el-table-column prop="type" :label="$t('m.Question_Type')" width="120">
              <template slot-scope="{ row }">
                {{ getQuestionTypeText(row.type) }}
              </template>
            </el-table-column>
            <el-table-column prop="difficulty" :label="$t('m.Difficulty')" width="100">
              <template slot-scope="{ row }">
                <el-rate v-model="row.difficulty" disabled />
              </template>
            </el-table-column>
            <el-table-column prop="score" :label="$t('m.Score')" width="80" />
            <el-table-column :label="$t('m.Operation')" width="120">
              <template slot-scope="{ row }">
                <el-button type="text" @click="viewQuestion(row)">
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
        <div slot="header">{{ $t('m.Selected_Questions') }} ({{ selectedQuestions.length }})</div>
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

    <!-- 题目详情对话框 -->
    <el-dialog :title="$t('m.Question_Detail')" :visible.sync="showDetailDialog" width="800px">
      <div v-if="currentQuestion" class="question-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            {{ getQuestionTypeText(currentQuestion.type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Question_Title')">
            {{ currentQuestion.title }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Content')">
            <div v-html="formatContent(currentQuestion.content)"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Options')" v-if="currentQuestion.options">
            <div v-html="formatOptions(currentQuestion.options)"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Answer')">
            {{ currentQuestion.answer }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate v-model="currentQuestion.difficulty" disabled />
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
      currentQuestion: null,
      form: {
        title: '',
        description: '',
        startTime: null,
        endTime: null,
        showHomework: false,
        showScore: true
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
      // 合并已选题目，避免重复
      const newSelection = selection.filter(s => !this.selectedQuestions.find(sq => sq.id === s.id))
      this.selectedQuestions = [...this.selectedQuestions, ...newSelection]
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
      this.currentQuestion = question
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
    formatContent(content) {
      if (!content) return '-'
      return content.replace(/\n/g, '<br>')
    },
    formatOptions(options) {
      if (!options) return '-'
      try {
        const opts = JSON.parse(options)
        return opts.join('<br>')
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
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', this.editId)
        if (res.code === 200 && res.data) {
          const homework = res.data
          // 填充表单数据
          this.form = {
            title: homework.title,
            description: homework.description,
            startTime: new Date(homework.startTime),
            endTime: new Date(homework.endTime),
            showHomework: homework.showHomework === 1,
            showScore: homework.showScore === 1
          }
          // 填充已选题目
          if (homework.questions && homework.questions.length > 0) {
            this.selectedQuestions = homework.questions.map(item => ({
              id: item.question.id,
              title: item.question.title,
              type: item.question.type,
              difficulty: item.question.difficulty,
              score: item.score
            }))
          }
        }
      } catch (error) {
        this.$message.error('加载作业数据失败')
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
          questions: this.selectedQuestions.map(q => ({
            questionId: q.id,
            score: Number(q.score) || 10
          }))
        }

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

          if (res.code === 200) {
            this.$message.success(this.isEditMode ? this.$t('m.Update_Success') : this.$t('m.Create_Success'))
            this.goBack()
          } else {
            this.$message.error(res.message || (this.isEditMode ? '保存失败' : '创建失败'))
          }
        } catch (error) {
          this.$message.error(this.isEditMode ? '保存失败' : '创建失败')
        } finally {
          this.submitting = false
        }
      })
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
</style>
