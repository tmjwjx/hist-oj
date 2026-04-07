<template>
  <div class="homework-analysis">
    <div class="header">
      <h3>学情分析</h3>
      <div>
        <el-button v-if="!hideBackButton" @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <el-card v-loading="loading">
      <!-- 总体统计 -->
      <div class="overview-section">
        <h4>总体情况</h4>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card clickable" @click="showAllStudents">
              <div class="stat-label">班级总人数</div>
              <div class="stat-value">{{ analysisData.totalStudentCount || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card success clickable" @click="showSubmittedStudents">
              <div class="stat-label">已提交人数</div>
              <div class="stat-value">{{ analysisData.submittedCount || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card warning clickable" @click="showUnsubmittedStudents">
              <div class="stat-label">未提交人数</div>
              <div class="stat-value">{{ analysisData.unsubmittedCount || 0 }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card info">
              <div class="stat-label">提交率</div>
              <div class="stat-value">{{ submissionRate }}%</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <el-divider></el-divider>

      <!-- 题目分析 -->
      <div class="questions-section">
        <h4>题目分析</h4>
        <el-collapse v-model="activeQuestions" accordion>
          <el-collapse-item
            v-for="(question, index) in analysisData.questionAnalysis || []"
            :key="question.homeworkQuestionId"
            :title="`题目 ${question.questionOrder}: ${question.title}`"
            :name="index"
          >
            <template slot="title">
              <div class="question-title">
                <span class="question-order">题目 {{ question.questionOrder }}</span>
                <el-tag :type="getQuestionTypeTag(question.type)" size="small" style="margin: 0 10px;">
                  {{ getQuestionTypeText(question.type) }}
                </el-tag>
                <span class="question-title-text">{{ question.title }}</span>
                <span class="question-score">({{ question.score }}分)</span>
              </div>
            </template>

            <div class="question-analysis-content">
              <!-- 题目统计 -->
              <el-row :gutter="20" class="question-stats">
                <el-col :span="4">
                  <div class="mini-stat">
                    <span class="mini-stat-label">提交人数：</span>
                    <span class="mini-stat-value">{{ question.submittedCount }}</span>
                  </div>
                </el-col>
                <el-col :span="4">
                  <div class="mini-stat">
                    <span class="mini-stat-label">未提交人数：</span>
                    <span class="mini-stat-value">{{ question.unsubmittedCount || 0 }}</span>
                  </div>
                </el-col>
                <el-col :span="4">
                  <div class="mini-stat">
                    <span class="mini-stat-label">平均得分：</span>
                    <span class="mini-stat-value">{{ question.avgScore.toFixed(2) }}</span>
                  </div>
                </el-col>
                <el-col :span="4">
                  <div class="mini-stat">
                    <span class="mini-stat-label">得分率：</span>
                    <span class="mini-stat-value">{{ ((question.avgScore / question.score) * 100).toFixed(1) }}%</span>
                  </div>
                </el-col>
                <el-col :span="8">
                  <el-button-group>
                    <el-button
                      type="primary"
                      size="small"
                      icon="el-icon-user"
                      @click="showQuestionSubmittedStudents(question)"
                      :disabled="!question.submittedBy || question.submittedBy.length === 0"
                    >
                      已提交 ({{ (question.submittedBy || []).length }}人)
                    </el-button>
                    <el-button
                      type="warning"
                      size="small"
                      icon="el-icon-user"
                      @click="showQuestionUnsubmittedStudents(question)"
                      :disabled="!question.unsubmittedBy || question.unsubmittedBy.length === 0"
                    >
                      未提交 ({{ (question.unsubmittedBy || []).length }}人)
                    </el-button>
                  </el-button-group>
                </el-col>
              </el-row>

              <!-- 题目分布扇形图（所有题目类型） -->
              <div v-if="question.options && question.options.length > 0" class="options-analysis">
                <h5>{{ getDistributionTitle(question.type) }}（点击查看学生）</h5>
                <el-row :gutter="20">
                  <el-col :span="12">
                    <!-- 饼图 -->
                    <div class="pie-chart-container">
                      <svg viewBox="0 0 800 300" class="pie-chart">
                        <g v-for="(slice, index) in generatePieSlices(question.options)" :key="index">
                          <!-- 完整的圆（360度） -->
                          <circle
                            v-if="slice.hasData && slice.isFullCircle"
                            cx="400"
                            cy="150"
                            r="70"
                            :fill="slice.color"
                            :stroke="'#fff'"
                            :stroke-width="'2'"
                            class="pie-slice"
                            @click="showOptionStudents(question.options[index])"
                            :style="{ cursor: question.options[index].selectedCount > 0 ? 'pointer' : 'default' }"
                          />
                          <!-- 扇形（非360度） -->
                          <path
                            v-if="slice.hasData && !slice.isFullCircle && slice.path"
                            :d="slice.path"
                            :fill="slice.color"
                            :stroke="'#fff'"
                            :stroke-width="'2'"
                            class="pie-slice"
                            @click="showOptionStudents(question.options[index])"
                            :style="{ cursor: question.options[index].selectedCount > 0 ? 'pointer' : 'default' }"
                          />
                          <!-- 引导线和外部标签（只显示有数据的选项） -->
                          <g v-if="slice.hasData && slice.percentage > 0">
                            <!-- 引导线：从饼图边缘到水平线端点 -->
                            <line
                              :x1="slice.lineStartX"
                              :y1="slice.lineStartY"
                              :x2="slice.labelX"
                              :y2="slice.labelY"
                              :fill="'none'"
                              :stroke="slice.color"
                              :stroke-width="'1.5'"
                            />
                            <!-- 外部标签 -->
                            <text
                              :x="slice.textX"
                              :y="slice.textY"
                              :text-anchor="slice.textAnchor"
                              dominant-baseline="middle"
                              :fill="slice.color"
                              :stroke="'white'"
                              :stroke-width="'3'"
                              :paint-order="'stroke fill'"
                              :font-size="'14'"
                              :font-weight="'bold'"
                              class="pie-label-outside"
                              @click="showOptionStudents(question.options[index])"
                              :style="{ cursor: question.options[index].selectedCount > 0 ? 'pointer' : 'default', pointerEvents: 'all' }"
                            >
                              {{ getSliceLabel(question.options[index], question.type) }} {{ slice.percentage.toFixed(1) }}%
                            </text>
                          </g>
                        </g>
                      </svg>
                    </div>
                    <!-- 正确答案显示（仅选择题和判断题） -->
                    <div v-if="['single_choice', 'multiple_choice', 'judge'].includes(question.type)" class="correct-answer-display">
                      <strong>正确答案：{{ getCorrectAnswerText(question) }}</strong>
                    </div>
                  </el-col>
                  <el-col :span="12">
                    <!-- 图例 -->
                    <div class="options-legend">
                      <div
                        v-for="(option, index) in question.options"
                        :key="index"
                        class="legend-item"
                        @click="showOptionStudents(option)"
                        :style="{ cursor: option.selectedCount > 0 ? 'pointer' : 'default' }"
                      >
                        <span class="legend-color" :style="{ backgroundColor: getOptionColor(index) }"></span>
                        <span class="legend-label">
                          {{ getLegendLabel(option, question.type) }}
                        </span>
                        <span class="legend-count">{{ option.selectedCount }}人</span>
                        <span class="legend-percentage">({{ option.percentage.toFixed(1) }}%)</span>
                      </div>
                    </div>
                  </el-col>
                </el-row>
              </div>

              <!-- 无提交数据时显示 -->
              <div v-else class="submitted-students">
                <h5>暂无提交数据</h5>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>

    <!-- 选项学生列表对话框 -->
    <el-dialog
      :title="`选择该选项的学生 (${optionDialogStudents.length}人)`"
      :visible.sync="optionDialogVisible"
      width="600px"
    >
      <el-table :data="optionDialogStudents" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
        <el-table-column prop="score" label="得分" width="100">
          <template slot-scope="{ row }">
            <el-tag type="info" size="small">{{ row.score }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 全部学生对话框 -->
    <el-dialog
      title="班级全部学生"
      :visible.sync="allStudentsDialogVisible"
      width="600px"
    >
      <el-table :data="allStudentsList" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 已提交学生对话框 -->
    <el-dialog
      :title="`已提交学生 (${submittedStudentsList.length}人)`"
      :visible.sync="submittedStudentsDialogVisible"
      width="600px"
    >
      <el-table :data="submittedStudentsList" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 未提交学生对话框 -->
    <el-dialog
      :title="`未提交学生 (${unsubmittedStudentsList.length}人)`"
      :visible.sync="unsubmittedStudentsDialogVisible"
      width="600px"
    >
      <el-table :data="unsubmittedStudentsList" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 题目作答学生对话框 -->
    <el-dialog
      :title="`题目作答学生 (${questionSubmittedStudents.length}人)`"
      :visible.sync="questionSubmittedDialogVisible"
      width="700px"
    >
      <el-table :data="questionSubmittedStudents" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名" width="150">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
        <el-table-column prop="score" label="得分" width="100">
          <template slot-scope="{ row }">
            <el-tag type="info" size="small">{{ row.score }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="answer" label="作答情况">
          <template slot-scope="{ row }">
            <span v-if="row.answer && row.answer.length > 0">
              {{ row.answer.join(', ') }}
            </span>
            <span v-else class="text-muted">{{ $t('m.No_Answer') }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 题目未提交学生对话框 -->
    <el-dialog
      :title="`题目未提交学生 (${questionUnsubmittedStudents.length}人)`"
      :visible.sync="questionUnsubmittedDialogVisible"
      width="600px"
    >
      <el-table :data="questionUnsubmittedStudents" stripe max-height="400">
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column label="系统用户名">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import teacherAuth from '@/mixins/teacherAuth'
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'

export default {
  name: 'HomeworkAnalysis',
  components: {
    UserName
  },
  mixins: [teacherAuth, realtimeSync],
  props: {
    classroomId: [String, Number],
    homeworkId: [String, Number],
    hideBackButton: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      loading: false,
      analysisData: {},
      activeQuestions: [0],
      optionDialogVisible: false,
      optionDialogStudents: [],
      // 学生列表对话框
      allStudentsDialogVisible: false,
      submittedStudentsDialogVisible: false,
      unsubmittedStudentsDialogVisible: false,
      // 题目作答学生对话框
      questionSubmittedDialogVisible: false,
      questionSubmittedStudents: [],
      // 题目未提交学生对话框
      questionUnsubmittedDialogVisible: false,
      questionUnsubmittedStudents: [],
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 5000,
        syncFunction: 'loadAnalysisData',
        immediate: true
      }
    }
  },
  computed: {
    // 统一的 homeworkId：优先从 props 获取（管理员路由），否则从 route 获取（教师路由）
    homeworkId() {
      // 优先从 props 获取
      if (this.$options.propsData && this.$options.propsData.homeworkId !== undefined) {
        return this.$options.propsData.homeworkId
      }
      // 否则从 route 获取
      return this.$route.query.homeworkId || this.$route.params.homeworkId
    },
    classroomId() {
      // 优先从 props 获取
      if (this.$options.propsData && this.$options.propsData.classroomId !== undefined) {
        return this.$options.propsData.classroomId
      }
      // 否则从 route 获取
      return this.$route.query.classroomId || this.$route.params.classroomId
    },
    submissionRate() {
        if (!this.analysisData.totalStudentCount || this.analysisData.totalStudentCount === 0) {
        return 0
      }
      return ((this.analysisData.submittedCount || 0) / this.analysisData.totalStudentCount * 100).toFixed(1)
    },
    allStudentsList() {
      return [...(this.analysisData.submittedStudents || []), ...(this.analysisData.unsubmittedStudents || [])]
    },
    submittedStudentsList() {
      return this.analysisData.submittedStudents || []
    },
    unsubmittedStudentsList() {
      return this.analysisData.unsubmittedStudents || []
    }
  },
  methods: {
    async loadAnalysisData() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = !this.analysisData || Object.keys(this.analysisData).length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const homeworkId = this.homeworkId  // 使用 computed 中的 homeworkId
        const res = await this.$store.dispatch('classroom/getHomeworkAnalysis', homeworkId)

        if (res.code === 200) {
          const newData = res.data

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.analysisData)
          const newDataString = JSON.stringify(newData)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.analysisData = newData
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error('加载学情分析失败')
          console.error('加载学情分析失败:', error)
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    showOptionStudents(option) {
      this.optionDialogStudents = option.selectedBy || []
      this.optionDialogVisible = true
    },
    showAllStudents() {
      this.allStudentsDialogVisible = true
    },
    showSubmittedStudents() {
      this.submittedStudentsDialogVisible = true
    },
    showUnsubmittedStudents() {
      this.unsubmittedStudentsDialogVisible = true
    },
    showQuestionSubmittedStudents(question) {
      this.questionSubmittedStudents = question.submittedBy || []
      this.questionSubmittedDialogVisible = true
    },
    showQuestionUnsubmittedStudents(question) {
      this.questionUnsubmittedStudents = question.unsubmittedBy || []
      this.questionUnsubmittedDialogVisible = true
    },
    formatAnswer(answer) {
      if (!answer) return ''
      // 将 option_A 格式转换为 A
      if (answer.startsWith('option_')) {
        return answer.replace('option_', '')
      }
      return answer
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        composite: '组合题',
        subjective: '主观题',
        programming: '编程题'
      }
      return map[type] || type
    },
    getQuestionTypeTag(type) {
      const map = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        composite: 'danger',
        subjective: 'info',
        programming: 'danger'
      }
      return map[type] || ''
    },
    getProgressColor(percentage) {
      if (percentage >= 50) return '#F56C6C'
      if (percentage >= 30) return '#E6A23C'
      return '#409EFF'
    },
    getDistributionTitle(type) {
      // 根据题目类型返回不同的标题
      const choiceTypes = ['single_choice', 'multiple_choice', 'judge']
      if (choiceTypes.includes(type)) {
        return '选项分布'
      }
      // 主观题和编程题显示分数段分布
      return '分数段分布'
    },
    getScoreTag(score, fullScore) {
      const rate = score / fullScore
      if (rate >= 0.8) return 'success'
      if (rate >= 0.6) return 'warning'
      return 'danger'
    },
    goBack() {
      // 返回到作业详情页
      const homeworkId = this.$route.params.homeworkId
      const classroomId = this.$route.params.classroomId
      // 必须添加 tab=homework 参数，否则 ClassroomDetail 会使用默认的 students tab
      this.$router.push({
        name: 'TeacherHomeworkDetail',
        params: { classroomId, homeworkId },
        query: { tab: 'homework' }
      })
    },
    generatePieSlices(options) {
      // 计算总提交数
      const total = options.reduce((sum, opt) => sum + opt.selectedCount, 0)

      const slices = []
      const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399', '#EEBE77']

      // 如果没有任何选中，返回空数组
      if (total === 0) {
        return options.map((option, index) => ({
          path: '',
          color: colors[index % colors.length],
          percentage: 0,
          hasData: false,
          isFullCircle: false
        }))
      }

      // SVG配置 - 中心点和半径
      const centerX = 400  // 画布中心X（画布总宽度800）
      const centerY = 150  // 画布中心Y（画布总高度300）
      const radius = 70    // 饼图半径
      const labelRadius = 90  // 标签线起始半径

      // 为有选中的选项生成扇形
      let currentAngle = 0
      options.forEach((option, index) => {
        const percentage = (option.selectedCount / total) * 100
        const angle = (option.selectedCount / total) * 360

        // 跳过没有选中的选项
        if (angle === 0) {
          slices[index] = {
            path: '',
            color: colors[index % colors.length],
            percentage: 0,
            hasData: false,
            isFullCircle: false
          }
          return
        }

        // 判断是否为完整圆（360度）
        const isFullCircle = Math.abs(angle - 360) < 0.001

        let path = ''
        if (!isFullCircle) {
          // 计算扇形的路径
          const startAngle = currentAngle
          const endAngle = currentAngle + angle

          const startX = centerX + radius * Math.cos((startAngle - 90) * Math.PI / 180)
          const startY = centerY + radius * Math.sin((startAngle - 90) * Math.PI / 180)
          const endX = centerX + radius * Math.cos((endAngle - 90) * Math.PI / 180)
          const endY = centerY + radius * Math.sin((endAngle - 90) * Math.PI / 180)

          const largeArcFlag = angle > 180 ? 1 : 0
          path = `M ${centerX} ${centerY} L ${startX} ${startY} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${endX} ${endY} Z`
        }

        // 计算扇形中角度
        const midAngle = currentAngle + angle / 2

        // 计算引导线
        // 从饼图边缘延伸出去
        const edgeX = centerX + radius * Math.cos((midAngle - 90) * Math.PI / 180)
        const edgeY = centerY + radius * Math.sin((midAngle - 90) * Math.PI / 180)

        // 延伸到标签半径
        const labelStartX = centerX + labelRadius * Math.cos((midAngle - 90) * Math.PI / 180)
        const labelStartY = centerY + labelRadius * Math.sin((midAngle - 90) * Math.PI / 180)

        // 判断左右侧
        const isRightSide = midAngle <= 180  // 0-180度在右侧，180-360度在左侧

        // 水平延伸线的长度
        const horizontalLength = 80

        // 引导线终点坐标
        let lineEndX, lineEndY
        if (isRightSide) {
          // 右侧：向右延伸
          lineEndX = labelStartX + horizontalLength
          lineEndY = labelStartY
        } else {
          // 左侧：向左延伸
          lineEndX = labelStartX - horizontalLength
          lineEndY = labelStartY
        }

        // 文字位置和锚点
        let labelX, labelY, textAnchor
        if (isRightSide) {
          // 右侧文字在引导线右侧
          labelX = lineEndX + 10
          labelY = lineEndY + 5
          textAnchor = 'start'
        } else {
          // 左侧文字在引导线左侧
          labelX = lineEndX - 10
          labelY = lineEndY + 5
          textAnchor = 'end'
        }

        slices[index] = {
          path,
          color: colors[index % colors.length],
          percentage,
          hasData: true,
          isFullCircle,
          lineStartX: edgeX,
          lineStartY: edgeY,
          lineEndX: labelStartX,
          lineEndY: labelStartY,
          labelX: lineEndX,
          labelY: lineEndY,
          textX: labelX,
          textY: labelY,
          textAnchor: textAnchor
        }

        currentAngle += angle
      })

      return slices
    },
    getOptionColor(index) {
      const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399', '#EEBE77']
      return colors[index % colors.length]
    },
    getOptionShortLabel(content, type) {
      // 移除"选项："前缀（如果存在）
      let actualContent = content
      if (content && content.startsWith('选项：')) {
        actualContent = content.substring(3)
      }

      // 判断题显示简短标签
      if (type === 'judge') {
        if (actualContent === '正确') return '对'
        if (actualContent === '错误') return '错'
      }

      // 单选和多选题，如果内容是"选项A"这样的格式，提取第一个字符
      if (actualContent && actualContent.length > 2 && actualContent.startsWith('选项')) {
        return actualContent.charAt(2)
      }

      // 其他情况返回前两个字符
      return actualContent ? actualContent.substring(0, 2) : ''
    },
    getSliceLabel(option, type) {
      // 编程题和主观题：使用 label 字段（分数段分布）
      if (type === 'programming' || type === 'subjective') {
        return option.label || ''
      }

      // 兼容旧版本：如果是字符串，说明是选择题的 content
      if (typeof option === 'string') {
        const content = option
        // 判断题显示完整标签：选项：正确/错误
        if (type === 'judge') {
          if (content === '正确') return '选项：正确'
          if (content === '错误') return '选项：错误'
        }
        // 其他题型使用简短标签
        return this.getOptionShortLabel(content, type)
      }

      // 新版本：如果是对象，使用 content 或 label 字段
      const content = option.content || option.label || ''
      // 判断题显示完整标签：选项：正确/错误
      if (type === 'judge') {
        if (content === '正确') return '选项：正确'
        if (content === '错误') return '选项：错误'
      }
      // 其他题型使用简短标签
      return this.getOptionShortLabel(content, type)
    },
    getLegendLabel(option, type) {
      // 编程题和主观题：直接显示 label（分数段）
      if (type === 'programming' || type === 'subjective') {
        return option.label || ''
      }

      // 判断题：显示"选项：正确/错误"
      if (type === 'judge') {
        return `选项：${option.content || option.label || ''}`
      }

      // 其他题型：显示 content
      return option.content || option.label || ''
    },
    getCorrectAnswerText(question) {
      // 如果没有答案，显示提示
      if (!question.answer || question.answer === '' || question.answer === '[]') {
        return '教师未设置参考答案'
      }

      // 判断题
      if (question.type === 'judge') {
        const normalizedAnswer = String(question.answer || '').toLowerCase().trim()
        if (normalizedAnswer === 'true') {
          return '正确'
        } else if (normalizedAnswer === 'false') {
          return '错误'
        }
        return '教师未设置参考答案'
      }

      // 单选题和多选题
      if (question.type === 'single' || question.type === 'multiple') {
        try {
          // 尝试解析 JSON 数组
          const answers = JSON.parse(question.answer)
          if (Array.isArray(answers) && answers.length > 0) {
            return answers.join(',')
          }
        } catch (e) {
          // 如果不是 JSON 数组，直接返回
          return question.answer
        }
      }

      // 其他题型（填空、编程等）
      return question.answer || '教师未设置参考答案'
    }
  }
}
</script>

<style scoped>
.homework-analysis {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.overview-section {
  margin-bottom: 20px;
}

.stat-card {
  background: #ffffff;
  border: 2px solid #667eea;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
  color: #303133;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  transition: all 0.3s;
}

.stat-card.clickable {
  cursor: pointer;
}

.stat-card.clickable:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.15);
}

.stat-card.success {
  background: #ffffff;
  border-color: #67c23a;
}

.stat-card.warning {
  background: #ffffff;
  border-color: #e6a23c;
}

.stat-card.info {
  background: #ffffff;
  border-color: #409eff;
}

.stat-label {
  font-size: 14px;
  margin-bottom: 10px;
  color: #909399;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.students-section {
  margin-bottom: 20px;
}

.questions-section {
  margin-bottom: 20px;
}

.question-title {
  display: flex;
  align-items: center;
  width: 100%;
}

.question-order {
  font-weight: bold;
  margin-right: 10px;
}

.question-title-text {
  flex: 1;
  margin-right: 10px;
}

.question-score {
  color: #909399;
  font-size: 14px;
}

.question-analysis-content {
  padding: 20px;
}

.question-stats {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #EBEEF5;
}

.mini-stat {
  text-align: center;
  padding: 10px;
  background: #F5F7FA;
  border-radius: 4px;
}

.mini-stat-label {
  font-size: 14px;
  color: #909399;
  margin-right: 10px;
}

.mini-stat-value {
  font-size: 18px;
  font-weight: bold;
  color: #303133;
}

.options-analysis {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #EBEEF5;
}

.options-analysis h5 {
  margin-bottom: 15px;
  color: #606266;
}

.correct-answer-display {
  margin-top: 20px;
  padding: 12px 16px;
  background-color: #f0f9ff;
  border-left: 4px solid #409EFF;
  border-radius: 4px;
  color: #409EFF;
  font-size: 14px;
}

.submitted-students h5 {
  margin-bottom: 15px;
  color: #606266;
}

h4 {
  color: #303133;
  margin-bottom: 15px;
}

h5 {
  font-size: 16px;
  font-weight: 500;
}

.pie-chart-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  min-height: 300px;
}

.pie-chart {
  width: 100%;
  height: 100%;
  max-width: 800px;
  max-height: 300px;
}

.pie-slice {
  transition: all 0.3s;
}

.pie-slice:hover {
  filter: brightness(1.1);
  transform: scale(1.05);
  transform-origin: center;
}

.pie-label {
  pointer-events: none;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.options-legend {
  margin-top: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  padding: 10px;
  margin-bottom: 8px;
  background: #F5F7FA;
  border-radius: 4px;
  transition: all 0.3s;
}

.legend-item:hover {
  background: #E4E7ED;
  transform: translateX(5px);
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  margin-right: 10px;
  flex-shrink: 0;
}

.legend-label {
  font-weight: bold;
  margin-right: 10px;
  min-width: 30px;
}

.legend-hint {
  color: #909399;
  font-size: 12px;
  margin-right: 10px;
  font-style: italic;
}

.legend-content {
  color: #909399;
  font-size: 14px;
  margin-right: 10px;
}

.legend-count {
  margin-right: 10px;
  color: #606266;
}

.legend-percentage {
  color: #909399;
  font-size: 14px;
}

.text-muted {
  color: #909399;
}
</style>
