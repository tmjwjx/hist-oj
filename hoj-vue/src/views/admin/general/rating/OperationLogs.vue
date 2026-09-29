<template>
  <div class="operation-logs">
    <!-- 筛选栏 -->
    <el-form inline>
      <el-form-item :label="$t('m.Rating_Operation_Type')">
        <el-select v-model="filter.type" :placeholder="$t('m.Rating_All')" clearable style="width: 150px">
          <el-option :label="$t('m.Rating_All')" value="" />
          <el-option :label="$t('m.Rating_Personal_Adjust')" value="personal_adjust" />
          <el-option :label="$t('m.Rating_Skip_User')" value="skip_user" />
          <el-option :label="$t('m.Rating_Cancel_Skip_Op')" value="cancel_skip" />
          <el-option :label="$t('m.Rating_Rating_Recalculate')" value="recalculate" />
        </el-select>
      </el-form-item>

      <el-form-item :label="$t('m.Rating_Time_Range')">
        <el-select v-model="filter.timeRange" :placeholder="$t('m.Rating_Last_7_Days')" style="width: 150px">
          <el-option :label="$t('m.Rating_Last_7_Days')" value="7d" />
          <el-option :label="$t('m.Rating_Last_30_Days')" value="30d" />
          <el-option :label="$t('m.Rating_All')" value="all" />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" icon="el-icon-search" @click="fetchLogs">{{ $t('m.Rating_Query') }}</el-button>
        <el-button icon="el-icon-refresh" @click="handleRefresh">{{ $t('m.Rating_Refresh') }}</el-button>
      </el-form-item>
    </el-form>

    <!-- 日志表格 -->
    <el-table :data="logs" v-loading="loading" stripe>
      <el-table-column prop="operationType" :label="$t('m.Rating_Operation_Type')" width="150">
        <template slot-scope="{ row }">
          <el-tag :type="getOperationTypeTag(row.operationType)" size="small">
            {{ getOperationTypeName(row.operationType) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="operationDetail" :label="$t('m.Rating_Operation_Content')" min-width="200">
        <template slot-scope="{ row }">
          {{ formatOperationDetail(row) }}
        </template>
      </el-table-column>

      <el-table-column prop="operatorUsername" :label="$t('m.Rating_Operator')" width="120" />

      <el-table-column :label="$t('m.Rating_Operation_Time')" width="180">
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
        this.$message.error(this.$t('m.Rating_Query_Logs_Failed') + (error.response?.data?.message || error.message))
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
          return this.$t('m.Rating_Personal_Adjust')
        case 'skip_user':
          return this.$t('m.Rating_Skip_User')
        case 'cancel_skip':
          return this.$t('m.Rating_Cancel_Skip_Op')
        case 'recalculate':
          return this.$t('m.Rating_Rating_Recalculate')
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
            return this.$t('m.Rating_Log_Adjust', { username: detail.username, change: detail.ratingChange })
          case 'skip_user':
            return this.$t('m.Rating_Log_Skip', { contestId: detail.contestId, users: detail.usernames?.join(', ') })
          case 'cancel_skip':
            return this.$t('m.Rating_Log_Cancel_Skip', { contestId: detail.contestId, uids: detail.uids?.join(', ') })
          case 'recalculate':
            return this.$t('m.Rating_Log_Recalc', { contestId: detail.contestId })
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
    }
  }
}
</script>

<style scoped>
.operation-logs {
  padding: 20px;
}
</style>
