<template>
  <el-card class="objective-problem-card">
    <div slot="header" class="card-header">
      <div class="header-title">
        <i class="fa fa-list-alt"></i>
        <span>客观题列表</span>
        <el-tag size="mini" type="warning">客观题</el-tag>
      </div>
      <div class="header-actions">
        <el-button type="primary" icon="el-icon-plus" size="small" @click="goCreatePage">
          创建客观题
        </el-button>
        <el-button icon="el-icon-refresh" size="small" @click="loadQuestions">
          刷新
        </el-button>
      </div>
    </div>

    <el-row :gutter="12" class="filter-row">
      <el-col :xs="24" :sm="12" :lg="4">
        <el-select v-model="filters.type" clearable placeholder="题型" @change="handleFilterChange">
          <el-option
            v-for="item in questionTypes"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="4">
        <el-select v-model="filters.course" clearable filterable placeholder="所属课程" @change="handleFilterChange">
          <el-option
            v-for="course in commonCourses"
            :key="course"
            :label="course"
            :value="course"
          />
        </el-select>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="4">
        <el-select v-model="filters.isShared" clearable placeholder="开放权限" @change="handleFilterChange">
          <el-option label="个人题库" value="0" />
          <el-option label="共享题库" value="1" />
        </el-select>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="4">
        <el-select v-model="filters.searchField" placeholder="搜索字段">
          <el-option label="题目标题" value="title" />
          <el-option label="题目 ID" value="id" />
          <el-option label="创建者" value="creator" />
        </el-select>
      </el-col>
      <el-col :xs="24" :sm="24" :lg="8">
        <el-input
          v-model.trim="filters.keyword"
          clearable
          :placeholder="searchPlaceholder"
          @clear="handleFilterChange"
          @keyup.enter.native="handleFilterChange"
        >
          <el-button slot="append" icon="el-icon-search" @click="handleFilterChange" />
        </el-input>
      </el-col>
    </el-row>

    <el-table
      v-loading="loading"
      :data="questions"
      border
      stripe
      class="question-table"
    >
      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column label="题目类型" width="150" align="center">
        <template slot-scope="{ row }">
          <div class="type-tags">
            <el-tag size="mini" type="warning">客观题</el-tag>
            <el-tag size="mini" :type="getQuestionTypeColor(row.type)">
              {{ getQuestionTypeName(row.type) }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="题目信息" min-width="280">
        <template slot-scope="{ row }">
          <div class="question-title">{{ row.title || '-' }}</div>
          <div v-if="row.content" class="question-summary">{{ summarize(row.content) }}</div>
        </template>
      </el-table-column>
      <el-table-column label="课程与标签" min-width="180">
        <template slot-scope="{ row }">
          <el-tag v-if="row.course" size="mini" type="warning">{{ row.course }}</el-tag>
          <span v-if="!row.course && parseTags(row.tags).length === 0">-</span>
          <el-tag
            v-for="(tag, index) in parseTags(row.tags)"
            :key="`${row.id}-${index}`"
            size="mini"
            type="info"
            class="metadata-tag"
          >
            {{ tag }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="难度" width="90" align="center">
        <template slot-scope="{ row }">{{ getDifficultyText(row.difficulty) }}</template>
      </el-table-column>
      <el-table-column label="标准答案" min-width="130">
        <template slot-scope="{ row }">
          <span class="answer-text">{{ getAnswerText(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="分值" width="80" align="center">
        <template slot-scope="{ row }">{{ Number(row.score || 0) }}</template>
      </el-table-column>
      <el-table-column label="创建者" min-width="130">
        <template slot-scope="{ row }">{{ getCreatorName(row) }}</template>
      </el-table-column>
      <el-table-column label="开放权限" width="100" align="center">
        <template slot-scope="{ row }">
          <el-tag size="mini" :type="Number(row.isShared) === 1 ? 'success' : 'info'">
            {{ Number(row.isShared) === 1 ? '共享' : '个人' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="150">
        <template slot-scope="{ row }">{{ row.createdAt | localtime }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right" align="center">
        <template slot-scope="{ row }">
          <el-button size="mini" type="primary" icon="el-icon-edit" @click="goEditPage(row)" />
          <el-button size="mini" type="danger" icon="el-icon-delete" @click="handleDelete(row)" />
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="pagination"
      :current-page.sync="pagination.currentPage"
      :page-size.sync="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="pagination.total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />
  </el-card>
</template>

<script>
import classroomApi from '@/api/classroom'

const QUESTION_TYPES = [
  { value: 'single_choice', label: '单选题' },
  { value: 'multiple_choice', label: '多选题' },
  { value: 'judge', label: '判断题' },
  { value: 'fill_blank', label: '填空题' },
  { value: 'subjective', label: '主观题' },
  { value: 'composite', label: '组合题' }
]

export default {
  name: 'ObjectiveProblemList',
  data() {
    return {
      loading: false,
      questions: [],
      questionTypes: QUESTION_TYPES,
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
      filters: {
        type: '',
        course: '',
        isShared: '',
        searchField: 'title',
        keyword: ''
      },
      pagination: {
        currentPage: 1,
        pageSize: 20,
        total: 0
      }
    }
  },
  computed: {
    searchPlaceholder() {
      const placeholders = {
        title: '搜索题目标题',
        id: '输入题目 ID',
        creator: '搜索创建者用户名'
      }
      return placeholders[this.filters.searchField] || '请输入关键词'
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
    buildQueryParams() {
      const params = {
        page: this.pagination.currentPage,
        limit: this.pagination.pageSize,
        type: this.filters.type || undefined,
        course: this.filters.course || undefined,
        isShared: this.filters.isShared
      }
      if (this.filters.keyword) {
        if (this.filters.searchField === 'id') {
          params.questionId = this.filters.keyword
        } else {
          params.searchField = this.filters.searchField
          params.keyword = this.filters.keyword
        }
      }
      return params
    },
    async loadQuestions() {
      this.loading = true
      try {
        const res = await classroomApi.adminGetQuestionBank(this.buildQueryParams())
        const data = res && res.data ? res.data.data : null
        this.questions = data && Array.isArray(data.questions) ? data.questions : []
        this.pagination.total = data ? Number(data.total || 0) : 0
      } catch (error) {
        this.questions = []
        this.pagination.total = 0
        this.$message.error('加载客观题列表失败')
      } finally {
        this.loading = false
      }
    },
    handleFilterChange() {
      this.pagination.currentPage = 1
      this.loadQuestions()
    },
    handleSizeChange(size) {
      this.pagination.pageSize = size
      this.pagination.currentPage = 1
      this.loadQuestions()
    },
    handleCurrentChange(page) {
      this.pagination.currentPage = page
      this.loadQuestions()
    },
    goCreatePage() {
      this.$router.push({
        name: 'admin-create-problem',
        query: { questionType: 'single_choice' }
      })
    },
    goEditPage(row) {
      if (!row || !row.id) return
      this.$router.push({
        name: 'admin-question-bank-edit',
        params: { questionId: String(row.id) }
      })
    },
    handleDelete(row) {
      if (!row || !row.id) return
      this.$confirm(`确定删除客观题「${row.title || row.id}」吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          await classroomApi.adminDeleteQuestion(row.id)
          this.$message.success('删除成功')
          if (this.questions.length === 1 && this.pagination.currentPage > 1) {
            this.pagination.currentPage -= 1
          }
          this.loadQuestions()
        } catch (error) {
          this.$message.error('删除失败')
        }
      }).catch(() => {})
    },
    getQuestionTypeName(type) {
      const current = this.questionTypes.find(item => item.value === type)
      return current ? current.label : (type || '未知题型')
    },
    getQuestionTypeColor(type) {
      const colors = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        fill_blank: 'info',
        subjective: 'danger',
        composite: 'success'
      }
      return colors[type] || 'info'
    },
    getDifficultyText(difficulty) {
      const labels = ['-', '简单', '中等', '困难']
      return labels[Number(difficulty)] || '-'
    },
    getAnswerText(row) {
      if (!row || !row.answer) return row && row.type === 'subjective' ? '人工评分' : '-'
      if (row.type === 'judge') return String(row.answer).toLowerCase() === 'true' ? '正确' : '错误'
      try {
        const parsed = typeof row.answer === 'string' ? JSON.parse(row.answer) : row.answer
        if (Array.isArray(parsed)) return parsed.join(' / ')
        if (parsed && typeof parsed === 'object') {
          return Object.keys(parsed).map(key => parsed[key]).join(' / ')
        }
      } catch (error) {
        // 后端兼容旧数据时可能直接返回字符串。
      }
      return String(row.answer)
    },
    getCreatorName(row) {
      if (row && row.creator && row.creator.username) return row.creator.username
      return row && row.creatorId ? row.creatorId : '-'
    },
    parseTags(tags) {
      if (Array.isArray(tags)) return tags.filter(Boolean)
      if (!tags) return []
      try {
        const parsed = JSON.parse(tags)
        return Array.isArray(parsed) ? parsed.filter(Boolean) : []
      } catch (error) {
        return String(tags).split(',').map(item => item.trim()).filter(Boolean)
      }
    },
    summarize(content) {
      const plainText = String(content || '')
        .replace(/```[\s\S]*?```/g, ' [代码] ')
        .replace(/!\[[^\]]*]\([^)]*\)/g, ' [图片] ')
        .replace(/<[^>]+>/g, ' ')
        .replace(/[#>*_`~\[\]()]/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
      return plainText.length > 100 ? `${plainText.slice(0, 100)}…` : plainText
    }
  }
}
</script>

<style scoped>
.card-header,
.header-title,
.header-actions,
.type-tags {
  display: flex;
  align-items: center;
}

.card-header {
  justify-content: space-between;
  gap: 16px;
}

.header-title {
  gap: 8px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.header-actions,
.type-tags {
  gap: 6px;
}

.header-actions .el-button + .el-button {
  margin-left: 0;
}

.filter-row {
  margin-bottom: 14px;
}

.filter-row .el-col {
  margin-bottom: 10px;
}

.filter-row .el-select,
.filter-row .el-input {
  width: 100%;
}

.question-table {
  width: 100%;
}

.question-title {
  color: #303133;
  font-weight: 600;
  line-height: 1.5;
}

.question-summary {
  margin-top: 5px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}

.answer-text {
  color: #606266;
  word-break: break-word;
}

.metadata-tag {
  margin-left: 5px;
  margin-bottom: 4px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

@media screen and (max-width: 768px) {
  .card-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .header-actions {
    flex-wrap: wrap;
  }

  .pagination {
    overflow-x: auto;
    text-align: left;
  }
}
</style>
