<template>
  <div class="checkin-records-panel classroom-theme">
    <div class="header">
      <div class="header-left">
        <h3>{{ checkinInfo.checkinName || '签到记录' }}</h3>
        <p class="subtitle">签到时间：{{ formatTime(checkinInfo.startTime) }}</p>
      </div>
      <button class="classroom-btn classroom-btn-secondary" @click="goBack">
        <i class="el-icon-back"></i>
        <span>{{ $t('m.Back') }}</span>
      </button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ filteredRecords.length }}</div>
        <div class="classroom-stat-label">总人数</div>
      </div>
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ statusCount.present }}</div>
        <div class="classroom-stat-label">已签到</div>
      </div>
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ statusCount.absent }}</div>
        <div class="classroom-stat-label">缺勤</div>
      </div>
      <div class="classroom-stat-card">
        <div class="classroom-stat-value">{{ statusCount.leave }}</div>
        <div class="classroom-stat-label">请假</div>
      </div>
    </div>

    <!-- 搜索和操作栏 -->
    <el-card class="search-card classroom-card">
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索学生姓名或学号"
          prefix-icon="el-icon-search"
          class="search-input"
          clearable
          @clear="handleSearch"
          @keyup.enter.native="handleSearch"
        />
        <el-select v-model="statusFilter" placeholder="筛选状态" clearable class="filter-select" @change="handleSearch">
          <el-option label="全部" value="" />
          <el-option label="签到" value="present" />
          <el-option label="缺勤" value="absent" />
          <el-option label="病假" value="sick_leave" />
          <el-option label="事假" value="personal_leave" />
        </el-select>
        <button class="classroom-btn classroom-btn-primary" @click="handleSearch">
          <i class="el-icon-search"></i>
          <span>搜索</span>
        </button>
        <button class="classroom-btn classroom-btn-success export-btn" @click="exportCheckinRecords">
          <i class="el-icon-download"></i>
          <span>导出签到记录</span>
        </button>
      </div>
    </el-card>

    <!-- 签到记录表格 -->
    <el-card class="table-card classroom-card">
      <el-table :data="filteredRecords" stripe v-loading="loading" class="classroom-table">
        <el-table-column :label="$t('m.Username')">
          <template slot-scope="{ row }">
            <UserName :username="row.student?.username" />
          </template>
        </el-table-column>
        <el-table-column label="学号" width="150">
          <template slot-scope="{ row }">
            {{ row.classStudent?.studentNumber || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Real_Name')" width="150">
          <template slot-scope="{ row }">
            {{ row.classStudent?.realName || (row.student?.realName || '-') }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Status')" width="180">
          <template slot-scope="{ row }">
            <el-select v-model="row.status" size="small" @change="updateRecord(row)" class="status-select">
              <el-option label="签到" value="present">
                <span class="classroom-tag classroom-tag-success">签到</span>
              </el-option>
              <el-option label="缺勤" value="absent">
                <span class="classroom-tag classroom-tag-warning">缺勤</span>
              </el-option>
              <el-option label="病假" value="sick_leave">
                <span class="classroom-tag classroom-tag-info">病假</span>
              </el-option>
              <el-option label="事假" value="personal_leave">
                <span class="classroom-tag classroom-tag-primary">事假</span>
              </el-option>
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Checkin_Time')" width="180">
          <template slot-scope="{ row }">
            <span v-if="row.checkinTime" class="time-text">
              <i class="el-icon-time"></i>
              {{ formatTime(row.checkinTime) }}
            </span>
            <span v-else class="time-empty">-</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import moment from 'moment'
import UserName from '@/components/oj/common/UserName.vue'
import teacherAuth from '@/mixins/teacherAuth'

export default {
  name: 'CheckinRecords',
  components: {
    UserName
  },
  mixins: [teacherAuth],
  data() {
    return {
      loading: false,
      checkinInfo: {},
      records: [],
      searchKeyword: '',
      statusFilter: ''
    }
  },
  computed: {
    classroomId() {
      return this.$route.params.classroomId
    },
    checkinId() {
      return this.$route.params.checkinId
    },
    filteredRecords() {
      let filtered = this.records

      // 按关键词搜索
      if (this.searchKeyword) {
        const keyword = this.searchKeyword.toLowerCase()
        filtered = filtered.filter(row => {
          const studentNumber = (row.classStudent?.studentNumber || '').toLowerCase()
          const realName = (row.classStudent?.realName || row.student?.realName || '').toLowerCase()
          const username = (row.student?.username || '').toLowerCase()
          return studentNumber.includes(keyword) ||
                 realName.includes(keyword) ||
                 username.includes(keyword)
        })
      }

      // 按状态筛选
      if (this.statusFilter) {
        filtered = filtered.filter(row => row.status === this.statusFilter)
      }

      return filtered
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        // 并行加载签到记录和班级学生列表
        const [recordsRes, studentsRes] = await Promise.all([
          this.$store.dispatch('classroom/getCheckinRecords', this.checkinId),
          this.$http.get(`/rating-api/api/classroom/${this.classroomId}/students`)
        ])

        if (recordsRes.code === 200) {
          const checkinRecords = recordsRes.data || []
          // 获取签到基本信息
          if (checkinRecords.length > 0) {
            this.checkinInfo = {
              checkinName: checkinRecords[0].checkinName,
              startTime: checkinRecords[0].startTime
            }
          }

          // 获取班级学生列表
          let allStudents = []
          if (studentsRes.data.code === 200) {
            allStudents = studentsRes.data.data || []
          }

          // 创建签到记录映射
          const recordsMap = new Map()
          checkinRecords.forEach(record => {
            recordsMap.set(record.uid, record)
          })

          // 合并数据：包含所有学生（已签到和未签到）
          this.records = allStudents.map(student => {
            const record = recordsMap.get(student.uuid || student.uid)

            return {
              ...student,
              id: record ? record.id : null,
              checkinId: this.checkinId,
              uid: student.uuid || student.uid,
              status: record ? record.status : 'absent',
              checkinTime: record ? record.checkinTime : null,
              student: student.user || {},
              classStudent: {
                studentNumber: student.studentNo || '-',
                realName: student.realName || '-'
              }
            }
          })
        }
      } catch (error) {
        console.error('加载签到记录失败:', error)
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },

    handleSearch() {
      // 搜索逻辑由 computed 属性 filteredRecords 自动处理
    },

    async updateRecord(record) {
      try {
        if (!record.id) {
          // 创建新记录
          const statusText = {
            'present': '签到',
            'absent': '缺勤',
            'sick_leave': '病假',
            'personal_leave': '事假'
          }

          const createRes = await this.$http.post(`/rating-api/api/classroom/checkin/${record.checkinId}/record`, {
            uid: record.uid,
            status: record.status
          })

          if (createRes.data.code === 200) {
            record.id = createRes.data.data.id
            this.$message.success(`已为未签到学生创建记录，状态设为：${statusText[record.status]}`)
          } else {
            this.$message.error(createRes.data.message || '创建记录失败')
            // 恢复原状态
            this.$nextTick(() => {
              record.status = 'absent'
            })
            return
          }
        } else {
          // 更新已有记录
          const res = await this.$http.put(`/rating-api/api/classroom/checkin/record`, {
            recordId: record.id,
            status: record.status
          })

          if (res.data.code === 200) {
            this.$message.success(this.$t('m.Update_Success'))
          } else {
            this.$message.error(res.data.message || this.$t('m.Update_Failed'))
          }
        }
      } catch (error) {
        console.error('更新记录失败:', error)
        this.$message.error(this.$t('m.Update_Failed'))
        // 恢复原状态
        this.loadData()
      }
    },

    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },

    goBack() {
      this.$router.back()
    },

    exportCheckinRecords() {
      const statusMap = {
        present: '签到',
        absent: '缺勤',
        sick_leave: '病假',
        personal_leave: '事假'
      }

      let csvContent = '\uFEFF'
      csvContent += `签到记录_${this.checkinInfo.checkinName || '签到'}_${this.formatTime(this.checkinInfo.startTime)}\n\n`
      csvContent += '用户名,学号,真实姓名,状态,签到时间\n'

      this.filteredRecords.forEach(record => {
        const username = record.student?.username || ''
        const studentNumber = record.classStudent?.studentNumber || '-'
        const realName = record.classStudent?.realName || record.student?.realName || '-'
        const status = statusMap[record.status] || record.status
        const checkinTime = record.checkinTime ? this.formatTime(record.checkinTime) : '-'

        const row = [
          username,
          studentNumber,
          realName,
          status,
          checkinTime
        ].map(field => `"${field}"`).join(',')
        csvContent += row + '\n'
      })

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', `签到记录_${this.checkinInfo.checkinName || '签到'}_${new Date().getTime()}.csv`)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    }
  }
}
</script>

<style scoped>
@import '../classroom-theme.css';

.checkin-records-panel {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header-left h3 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--classroom-text);
}

.subtitle {
  font-size: 14px;
  color: var(--classroom-text-secondary);
  margin: 0;
}

.header .classroom-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
  white-space: nowrap;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.search-card {
  margin-bottom: 20px;
}

.search-card ::v-deep .el-card__body {
  padding: 20px;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.search-input {
  width: 300px;
}

.filter-select {
  width: 150px;
}

.search-bar .classroom-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
}

.export-btn {
  margin-left: auto;
}

.table-card {
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.table-card ::v-deep .el-card__body {
  padding: 20px;
}

/* 表格样式优化 */
.checkin-records-panel ::v-deep .el-table {
  border-radius: 8px;
  overflow: hidden;
  font-size: 14px;
}

.checkin-records-panel ::v-deep .el-table th {
  background: #E3F2FD;
  color: var(--classroom-text);
  font-weight: 600;
  font-size: 14px;
  padding: 16px 12px;
  border-bottom: 2px solid var(--classroom-primary);
  text-align: left;
}

.checkin-records-panel ::v-deep .el-table td {
  padding: 14px 12px;
  font-size: 14px;
  color: var(--classroom-text);
  border-bottom: 1px solid var(--classroom-border);
}

.checkin-records-panel ::v-deep .el-table--striped .el-table__body tr.el-table__row--striped td {
  background: var(--classroom-hover);
}

.checkin-records-panel ::v-deep .el-table__body tr:hover > td {
  background: var(--classroom-hover) !important;
}

/* 状态选择器样式 */
.status-select ::v-deep .el-input__inner {
  border-radius: 8px;
  border: 1px solid var(--classroom-border);
}

.status-select ::v-deep .el-input__inner {
  height: 32px;
  line-height: 32px;
  font-size: 14px;
}

/* 时间文本样式 */
.time-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--classroom-text);
  font-size: 14px;
}

.time-text i {
  color: var(--classroom-primary);
  font-size: 15px;
}

.time-empty {
  color: #909399;
  font-size: 14px;
}

/* 响应式 */
@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .header .classroom-btn {
    width: 100%;
    justify-content: center;
  }

  .stats-row {
    grid-template-columns: 1fr;
  }

  .search-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .search-input,
  .filter-select {
    width: 100% !important;
  }

  .search-bar .classroom-btn {
    width: 100%;
    justify-content: center;
  }

  .export-btn {
    margin-left: 0;
  }
}
</style>
