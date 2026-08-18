<template>
  <div class="operation-logs">
    <!-- 筛选栏 -->
    <el-form inline>
      <el-form-item label="操作类型">
        <el-select v-model="filter.type" placeholder="全部" clearable style="width: 150px">
          <el-option label="全部" value="" />
          <el-option label="个人调整" value="personal_adjust" />
          <el-option label="Skip用户" value="skip_user" />
          <el-option label="取消Skip" value="cancel_skip" />
          <el-option label="Rating重算" value="recalculate" />
        </el-select>
      </el-form-item>

      <el-form-item label="时间范围">
        <el-select v-model="filter.timeRange" placeholder="最近7天" style="width: 150px">
          <el-option label="最近7天" value="7d" />
          <el-option label="最近30天" value="30d" />
          <el-option label="全部" value="all" />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" icon="el-icon-search" @click="fetchLogs">查询</el-button>
        <el-button icon="el-icon-refresh" @click="handleRefresh">刷新</el-button>
        <el-button type="warning" icon="el-icon-edit" @click="handleFixUsername" :loading="fixing">修复用户名</el-button>
      </el-form-item>
    </el-form>

    <!-- 日志表格 -->
    <el-table :data="logs" v-loading="loading" stripe>
      <el-table-column prop="operationType" label="操作类型" width="150">
        <template slot-scope="{ row }">
          <el-tag :type="getOperationTypeTag(row.operationType)" size="small">
            {{ getOperationTypeName(row.operationType) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="operationDetail" label="操作内容" min-width="200">
        <template slot-scope="{ row }">
          {{ formatOperationDetail(row) }}
        </template>
      </el-table-column>

      <el-table-column prop="operatorUsername" label="操作人" width="120" />

      <el-table-column label="操作时间" width="180">
        <template slot-scope="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      :current-page="pagination.page"
      :page-sizes="[10, 20, 50, 100]"
      :page-size="pagination.limit"
      layout="total, sizes, prev, pager, next, jumper"
      :total="pagination.total"
      style="margin-top: 20px; text-align: right"
    />
  </div>
</template>

<script>
import ratingApi from '@/common/rating-api'

export default {
  name: 'OperationLogs',
  data() {
    return {
      logs: [],
      loading: false,
      fixing: false,
      filter: {
        type: '',
        timeRange: '7d'
      },
      pagination: {
        page: 1,
        limit: 20,
        total: 0
      }
    }
  },
  mounted() {
    this.fetchLogs()
  },
  methods: {
    // 获取日志列表
    async fetchLogs() {
      this.loading = true
      try {
        const data = await ratingApi.getOperationLogs(
          this.pagination.page,
          this.pagination.limit,
          this.filter.type,
          this.filter.timeRange
        )
        this.logs = data.records || []
        this.pagination.total = data.total || 0
      } catch (error) {
        console.error('查询日志失败:', error)
        this.$message.error('查询日志失败: ' + (error.response?.data?.message || error.message))
      } finally {
        this.loading = false
      }
    },

    // 刷新
    handleRefresh() {
      this.pagination.page = 1
      this.fetchLogs()
    },

    // 分页大小变化
    handleSizeChange(val) {
      this.pagination.limit = val
      this.fetchLogs()
    },

    // 当前页变化
    handleCurrentChange(val) {
      this.pagination.page = val
      this.fetchLogs()
    },

    // 获取操作类型标签颜色
    getOperationTypeTag(type) {
      switch (type) {
        case 'personal_adjust':
          return 'success'
        case 'skip_user':
          return 'warning'
        case 'cancel_skip':
          return 'info'
        case 'recalculate':
          return 'danger'
        default:
          return ''
      }
    },

    // 获取操作类型名称
    getOperationTypeName(type) {
      switch (type) {
        case 'personal_adjust':
          return '个人调整'
        case 'skip_user':
          return 'Skip用户'
        case 'cancel_skip':
          return '取消Skip'
        case 'recalculate':
          return 'Rating重算'
        default:
          return type
      }
    },

    // 格式化操作详情
    formatOperationDetail(row) {
      try {
        const detail = typeof row.operationDetail === 'string'
          ? JSON.parse(row.operationDetail)
          : row.operationDetail

        switch (row.operationType) {
          case 'personal_adjust':
            return `调整用户 ${detail.username} rating ${detail.ratingChange}`
          case 'skip_user':
            return `比赛 ${detail.contestId} skip用户: ${detail.usernames?.join(', ')}`
          case 'cancel_skip':
            return `比赛 ${detail.contestId} 取消skip: ${detail.uids?.join(', ')}`
          case 'recalculate':
            return `从比赛 ${detail.contestId} 开始重算rating`
          default:
            return JSON.stringify(detail)
        }
      } catch (e) {
        return row.operationDetail || '-'
      }
    },

    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    },

    // 修复缺失的操作人用户名
    async handleFixUsername() {
      this.$confirm('此操作将自动补充缺失的操作人用户名，是否继续？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.fixing = true
        this.axios.post('/api/rating/admin/fix-logs-username')
          .then(response => {
            const data = response.data
            if (data.code === 200 || data.status === 200) {
              this.$message.success('修复成功！')
              console.log('修复统计:', data.data && data.data.statistics)
              // 刷新日志列表
              this.fetchLogs()
            } else {
              this.$message.error('修复失败: ' + (data.message || data.msg))
            }
          })
          .catch(error => {
            console.error('修复失败:', error)
            this.$message.error('修复失败: ' + (error.response?.data?.message || error.message))
          })
          .finally(() => {
            this.fixing = false
          })
      }).catch(() => {
        // 用户取消
      })
    }
  }
}
</script>

<style scoped>
.operation-logs {
  padding: 20px;
}
</style>
