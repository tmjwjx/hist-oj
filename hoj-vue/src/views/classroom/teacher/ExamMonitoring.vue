<template>
  <div class="exam-monitoring">
    <div class="header">
      <h3>考试监控</h3>
      <div>
        <el-button type="danger" @click="forceSubmitAll" :loading="forcingAll" :disabled="!canForceSubmitAll">
          <i class="el-icon-download"></i>
          <span>强制收卷</span>
        </el-button>
        <el-button v-if="!hideBackButton" @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-number">{{ stats.totalStudents || 0 }}</div>
          <div class="stat-label">总学生数</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card stat-warning">
          <div class="stat-number">{{ stats.notStartedCount || 0 }}</div>
          <div class="stat-label">未开始</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card stat-primary">
          <div class="stat-number">{{ stats.inProgressCount || 0 }}</div>
          <div class="stat-label">答题中</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card stat-success">
          <div class="stat-number">{{ stats.submittedCount || 0 }}</div>
          <div class="stat-label">已提交</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card stat-danger">
          <div class="stat-number">{{ violationCount }}</div>
          <div class="stat-label">有违规</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-number">{{ submissionRate }}%</div>
          <div class="stat-label">提交率</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 学生列表 -->
    <el-card class="student-list-card">
      <div slot="header" class="card-header">
        <span>学生列表</span>
        <el-tag type="info" size="small">每3秒自动刷新</el-tag>
      </div>

      <el-table :data="studentList" stripe v-loading="loading" row-key="uid">
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column label="状态" width="120">
          <template slot-scope="{ row }">
            <el-tag v-if="row.status === 'not_started'" type="info" size="small">未开始</el-tag>
            <el-tag v-else-if="row.status === 'in_progress'" type="primary" size="small">答题中</el-tag>
            <el-tag v-else-if="row.status === 'submitted'" type="success" size="small">已提交</el-tag>
            <el-tag v-else-if="row.status === 'forced_submit'" type="warning" size="small">强制收卷</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="已用时间" width="100">
          <template slot-scope="{ row }">
            <span v-if="row.elapsedMinutes >= 0">{{ row.elapsedMinutes }}分钟</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="剩余时间" width="100">
          <template slot-scope="{ row }">
            <span v-if="row.remainingSeconds >= 0" :class="getTimeClass(row.remainingSeconds)">
              {{ formatTime(row.remainingSeconds) }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="违规记录" width="300">
          <template slot-scope="{ row }">
            <div v-if="row.violations && row.violations.length > 0" class="violations">
              <el-tag
                v-for="v in getSortedViolations(row.violations)"
                :key="v.type"
                :type="getViolationTagType(v.type)"
                size="mini"
                style="margin-right: 5px; margin-bottom: 3px;"
              >
                {{ getViolationTypeText(v.type) }} ×{{ v.count }}
              </el-tag>
            </div>
            <span v-else class="no-violation">无违规</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template slot-scope="{ row }">
            <el-button
              v-if="row.status === 'in_progress'"
              type="danger"
              size="small"
              @click="forceSubmit(row)"
              :loading="forcingUid === row.uid"
            >
              强制交卷
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 强制交卷确认对话框 -->
    <el-dialog
      title="强制交卷确认"
      :visible.sync="showForceSubmitDialog"
      width="450px"
    >
      <div v-if="currentStudent">
        <p>确定要强制学生 <strong>{{ currentStudent.name }}</strong> 交卷吗？</p>
        <el-input
          v-model="forceSubmitReason"
          placeholder="请输入强制交卷原因（可选）"
          type="textarea"
          :rows="3"
        />
      </div>
      <span slot="footer">
        <el-button @click="showForceSubmitDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmForceSubmit" :loading="forcing">确定</el-button>
      </span>
    </el-dialog>

    <!-- 强制收卷确认对话框 -->
    <el-dialog
      title="强制收卷确认"
      :visible.sync="showForceSubmitAllDialog"
      width="450px"
    >
      <div>
        <p><i class="el-icon-warning" style="color: #E6A23C; font-size: 24px;"></i></p>
        <p>确定要强制所有未提交的学生交卷吗？</p>
        <p class="hint">此操作将立即收卷，所有未提交的学生将以当前草稿状态作为最终答案。</p>
      </div>
      <span slot="footer">
        <el-button @click="showForceSubmitAllDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmForceSubmitAll" :loading="forcingAll">确定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import realtimeSync from '@/mixins/realtimeSync'
import teacherAuth from '@/mixins/teacherAuth'

export default {
  name: 'ExamMonitoring',
  mixins: [realtimeSync, teacherAuth],
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
      stats: {
        totalStudents: 0,
        startedCount: 0,
        submittedCount: 0,
        inProgressCount: 0,
        notStartedCount: 0
      },
      studentList: [],
      // 强制交卷
      showForceSubmitDialog: false,
      showForceSubmitAllDialog: false,
      currentStudent: null,
      forceSubmitReason: '',
      forcing: false,
      forcingUid: null,
      forcingAll: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadMonitoringData',
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
    // 确保从 route params 获取正确的 ID
    routeClassroomId() {
      return this.$route.params.classroomId
    },
    routeHomeworkId() {
      return parseInt(this.$route.params.homeworkId)
    },
    violationCount() {
      return this.studentList.filter(s => s.violationCount > 0).length
    },
    submissionRate() {
      if (this.stats.totalStudents === 0) return 0
      return Math.round((this.stats.submittedCount / this.stats.totalStudents) * 100)
    },
    canForceSubmitAll() {
      return this.stats.inProgressCount > 0
    }
  },
  mounted() {
    // 由 realtimeSync mixin 自动启动同步
  },
  methods: {
    async loadMonitoringData() {
      // 使用 computed 中的 homeworkId（优先从 query 获取，再从 params 获取）
      const homeworkId = this.homeworkId

      // 检查 homeworkId 是否有效
      if (!homeworkId) {
        console.error('homeworkId is undefined, route params:', this.$route.params)
        this.$message.error('作业ID缺失')
        return
      }

      if (this.loading) return

      const isFirstLoad = !this.stats || this.stats.totalStudents === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getExamMonitoring', homeworkId)
        if (res.code === 200) {
          // 更新统计数据
          this.stats = {
            totalStudents: res.data.totalStudents,
            startedCount: res.data.startedCount,
            submittedCount: res.data.submittedCount,
            inProgressCount: res.data.inProgressCount,
            notStartedCount: res.data.notStartedCount
          }

          // 深度对比：只在数据真的变化时才更新studentList
          const newStudentList = res.data.studentList || []
          const currentStudentListString = JSON.stringify(this.studentList)
          const newStudentListString = JSON.stringify(newStudentList)

          if (currentStudentListString !== newStudentListString) {
            // 数据真的变化了，才更新
            this.studentList = newStudentList
          }
        } else {
          this.$message.error('加载监控数据失败: ' + (res.msg || '未知错误'))
        }
      } catch (error) {
        console.error('加载监控数据失败:', error)
        if (isFirstLoad) {
          this.$message.error('加载监控数据失败')
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },

    // 格式化时间（秒 -> MM:SS）
    formatTime(seconds) {
      if (!seconds && seconds !== 0) return '-'
      const mins = Math.floor(seconds / 60)
      const secs = seconds % 60
      return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    },

    // 获取时间样式类
    getTimeClass(seconds) {
      if (seconds <= 300) return 'time-critical' // 5分钟内
      if (seconds <= 600) return 'time-warning' // 10分钟内
      return ''
    },

    // 获取违规类型标签颜色
    getViolationTagType(type) {
      const typeMap = {
        'tab_switch': 'warning',
        'fullscreen_exit': 'danger',
        'copy_attempt': 'danger',
        'paste_attempt': 'danger',
        'context_menu': 'warning',
        'devtools_attempt': 'danger',
        'window_blur': 'warning'
      }
      return typeMap[type] || 'info'
    },

    // 获取违规类型文本
    getViolationTypeText(type) {
      const typeMap = {
        'tab_switch': '切换标签页',
        'fullscreen_exit': '退出全屏',
        'copy_attempt': '尝试复制',
        'paste_attempt': '尝试粘贴',
        'context_menu': '右键菜单',
        'devtools_attempt': '开发者工具',
        'forced_submit': '强制收卷',
        'window_blur': '切换窗口'
      }
      return typeMap[type] || type
    },

    // 获取排序后的违规记录（固定顺序，避免跳动）
    getSortedViolations(violations) {
      // 定义违规类型的固定显示顺序
      const order = [
        'forced_submit',      // 强制收卷（最重要，放最前面）
        'fullscreen_exit',    // 退出全屏
        'devtools_attempt',   // 开发者工具
        'window_blur',        // 切换窗口
        'tab_switch',         // 切换标签页
        'copy_attempt',       // 尝试复制
        'paste_attempt',      // 尝试粘贴
        'context_menu'        // 右键菜单
      ]

      // 创建一个 Map 来快速查找顺序索引
      const orderMap = {}
      order.forEach((type, index) => {
        orderMap[type] = index
      })

      // 按照固定顺序排序
      return [...violations].sort((a, b) => {
        const orderA = orderMap[a.type] !== undefined ? orderMap[a.type] : 999
        const orderB = orderMap[b.type] !== undefined ? orderMap[b.type] : 999
        return orderA - orderB
      })
    },

    // 强制单个学生交卷
    forceSubmit(student) {
      this.currentStudent = student
      this.forceSubmitReason = ''
      this.showForceSubmitDialog = true
    },

    async confirmForceSubmit() {
      this.forcing = true
      this.forcingUid = this.currentStudent.uid

      try {
        const res = await this.$store.dispatch('classroom/forceSubmit', {
          homeworkId: this.routeHomeworkId,
          uid: this.currentStudent.uid,
          reason: this.forceSubmitReason
        })

        // 检查响应code
        if (res && res.code === 200) {
          this.$message.success('强制交卷成功')
          this.showForceSubmitDialog = false
          // 立即刷新监控数据
          await this.loadMonitoringData()
        } else {
          // 处理错误响应
          const errorMsg = res && res.msg ? res.msg : '强制交卷失败'
          this.$message.error(errorMsg)
          console.error('强制交卷失败:', res)
        }
      } catch (error) {
        console.error('强制交卷异常:', error)
        // 尝试从error中获取详细信息
        let errorMsg = '强制交卷失败'
        if (error.response && error.response.data && error.response.data.msg) {
          errorMsg = error.response.data.msg
        } else if (error.message) {
          errorMsg = '强制交卷失败: ' + error.message
        }
        this.$message.error(errorMsg)
      } finally {
        this.forcing = false
        this.forcingUid = null
      }
    },

    // 批量强制收卷
    forceSubmitAll() {
      if (!this.canForceSubmitAll) {
        this.$message.warning('没有正在答题的学生')
        return
      }
      this.showForceSubmitAllDialog = true
    },

    async confirmForceSubmitAll() {
      this.forcingAll = true

      try {
        const res = await this.$store.dispatch('classroom/forceSubmitAll', this.routeHomeworkId)

        // 检查响应code
        if (res && res.code === 200) {
          const count = res.data && res.data.forcedCount ? res.data.forcedCount : 0
          this.$message.success(`强制收卷成功，已收卷 ${count} 名学生`)
          this.showForceSubmitAllDialog = false
          // 立即刷新监控数据
          await this.loadMonitoringData()
        } else {
          // 处理错误响应
          const errorMsg = res && res.msg ? res.msg : '强制收卷失败'
          this.$message.error(errorMsg)
          console.error('强制收卷失败:', res)
        }
      } catch (error) {
        console.error('强制收卷异常:', error)
        // 尝试从error中获取详细信息
        let errorMsg = '强制收卷失败'
        if (error.response && error.response.data && error.response.data.msg) {
          errorMsg = error.response.data.msg
        } else if (error.message) {
          errorMsg = '强制收卷失败: ' + error.message
        }
        this.$message.error(errorMsg)
      } finally {
        this.forcingAll = false
      }
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
    }
  }
}
</script>

<style scoped>
.exam-monitoring {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

/* 统计卡片 */
.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  padding: 10px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.stat-warning .stat-number {
  color: #E6A23C;
}

.stat-primary .stat-number {
  color: #409EFF;
}

.stat-success .stat-number {
  color: #67C23A;
}

.stat-danger .stat-number {
  color: #F56C6C;
}

/* 学生列表 */
.student-list-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.violations {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
}

.no-violation {
  color: #909399;
  font-size: 12px;
}

/* 时间样式 */
.time-warning {
  color: #E6A23C;
  font-weight: bold;
}

.time-critical {
  color: #F56C6C;
  font-weight: bold;
}

.hint {
  color: #909399;
  font-size: 12px;
}
</style>
