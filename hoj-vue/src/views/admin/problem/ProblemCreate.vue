<template>
  <div class="problem-create-page">
    <el-card shadow="never" class="type-selector-card">
      <div class="type-selector-row">
        <div>
          <div class="selector-title">{{ $t('m.ProbCreate_Type_Title') }}</div>
          <div class="selector-description">{{ selectedTypeDescription }}</div>
        </div>
        <el-select v-model="selectedType" class="type-selector" @change="handleTypeChange">
          <el-option :label="$t('m.ProbCreate_Programming')" value="programming" />
          <el-option
            v-for="item in objectiveTypes"
            v-if="isSuperAdmin"
            :key="item.value"
            :label="$t('m.' + item.labelKey)"
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

// 模块级映射表：仅存 i18n 键名字符串，展示时通过 this.$t('m.' + key) 翻译
const OBJECTIVE_TYPES = [
  { value: 'single_choice', labelKey: 'Practice_Single_Choice', descKey: 'ProbCreate_Desc_Single_Choice' },
  { value: 'multiple_choice', labelKey: 'Practice_Multiple_Choice', descKey: 'ProbCreate_Desc_Multiple_Choice' },
  { value: 'judge', labelKey: 'Practice_Judge', descKey: 'ProbCreate_Desc_Judge' },
  { value: 'fill_blank', labelKey: 'Practice_Fill_Blank', descKey: 'ProbCreate_Desc_Fill_Blank' },
  { value: 'subjective', labelKey: 'Practice_Subjective', descKey: 'ProbCreate_Desc_Subjective' },
  { value: 'composite', labelKey: 'Practice_Composite', descKey: 'ProbCreate_Desc_Composite' }
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
        return this.$t('m.ProbCreate_Desc_Programming')
      }
      const current = this.objectiveTypes.find(item => item.value === this.selectedType)
      return current ? this.$t('m.' + current.descKey) : this.$t('m.ProbCreate_Desc_Objective')
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
