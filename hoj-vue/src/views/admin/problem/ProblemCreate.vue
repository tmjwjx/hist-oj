<template>
  <div class="problem-create-page">
    <el-card shadow="never" class="type-selector-card">
      <div class="type-selector-row">
        <div>
          <div class="selector-title">题目类型</div>
          <div class="selector-description">{{ selectedTypeDescription }}</div>
        </div>
        <el-select v-model="selectedType" class="type-selector" @change="handleTypeChange">
          <el-option label="编程题" value="programming" />
          <el-option
            v-for="item in objectiveTypes"
            v-if="isSuperAdmin"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>
    </el-card>

    <Problem v-if="selectedType === 'programming'" />
    <QuestionBankEditor
      v-else
      scene="admin"
      :initial-type="selectedType"
      @type-change="handleEditorTypeChange"
    />
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Problem from '@/views/admin/problem/Problem.vue'
import QuestionBankEditor from '@/views/common/QuestionBankEditor.vue'

const OBJECTIVE_TYPES = [
  { value: 'single_choice', label: '单选题', description: '从多个选项中选择一个正确答案' },
  { value: 'multiple_choice', label: '多选题', description: '从多个选项中选择一个或多个正确答案' },
  { value: 'judge', label: '判断题', description: '判断题目陈述是否正确' },
  { value: 'fill_blank', label: '填空题', description: '填写文本答案，支持配置多个标准答案' },
  { value: 'subjective', label: '主观题', description: '填写文字答案，由教师人工评分' },
  { value: 'composite', label: '组合题', description: '在一道大题下配置多道子题及独立分值' }
]

export default {
  name: 'ProblemCreate',
  components: {
    Problem,
    QuestionBankEditor
  },
  data() {
    const requestedType = String(this.$route.query.questionType || '')
    const canUseRequestedType = OBJECTIVE_TYPES.some(item => item.value === requestedType)
    return {
      objectiveTypes: OBJECTIVE_TYPES,
      selectedType: canUseRequestedType ? requestedType : 'programming'
    }
  },
  computed: {
    ...mapGetters(['isSuperAdmin']),
    selectedTypeDescription() {
      if (this.selectedType === 'programming') {
        return '默认创建由评测数据自动判定结果的编程题'
      }
      const current = this.objectiveTypes.find(item => item.value === this.selectedType)
      return current ? current.description : '创建客观题'
    }
  },
  created() {
    if (!this.isSuperAdmin && this.selectedType !== 'programming') {
      this.selectedType = 'programming'
    }
  },
  methods: {
    handleTypeChange() {
      const query = { ...this.$route.query }
      if (this.selectedType === 'programming') {
        delete query.questionType
      } else {
        query.questionType = this.selectedType
      }
      this.$router.replace({ name: 'admin-create-problem', query }).catch(() => {})
    },
    handleEditorTypeChange(type) {
      if (type === this.selectedType) return
      this.selectedType = type
      this.handleTypeChange()
    }
  }
}
</script>

<style scoped>
.type-selector-card {
  margin-bottom: 14px;
}

.type-selector-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.selector-title {
  margin-bottom: 6px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.selector-description {
  color: #909399;
  line-height: 1.5;
}

.type-selector {
  width: 220px;
  flex: none;
}

@media screen and (max-width: 768px) {
  .type-selector-row {
    align-items: stretch;
    flex-direction: column;
    gap: 12px;
  }

  .type-selector {
    width: 100%;
  }
}
</style>
