<template>
  <div class="my-records-container">
    <el-card class="records-card" shadow="never">
      <div slot="header" class="card-header">
        <div class="header-left">
          <el-button type="primary" size="small" @click="backToHome" class="back-button">
            <i class="fa fa-arrow-left"></i> 返回
          </el-button>
          <div class="header-title">
            <i class="fa fa-history"></i>
            <span>我的对战记录</span>
          </div>
        </div>
        <el-button type="success" size="small" @click="loadRecords" :loading="loading" class="refresh-button">
          <i class="fa fa-refresh"></i> 刷新
        </el-button>
      </div>

      <div class="records-content">
        <!-- 统计卡片 -->
        <el-row :gutter="20" class="stats-row">
          <el-col :span="6">
            <el-card class="stat-card total" shadow="never">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.totalBattles }}</div>
                <div class="stat-label">全部场次</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card win" shadow="never">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.winCount }}</div>
                <div class="stat-label">全部胜场</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card lose" shadow="never">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.loseCount }}</div>
                <div class="stat-label">负场</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card rate" shadow="never">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.winRate }}%</div>
                <div class="stat-label">全部获胜率</div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 记录列表 -->
        <el-table
          :data="recordList"
          style="width: 100%; margin-top: 20px;"
          v-loading="loading"
        >
          <el-table-column label="对手" width="150" align="center">
            <template slot-scope="scope">
              <div class="opponent-cell">
                <img v-if="scope.row.opponentAvatar" :src="scope.row.opponentAvatar" class="opponent-avatar" />
                <i v-else class="fa fa-user-circle opponent-avatar-placeholder"></i>
                <span class="opponent-username" :style="{ color: getOpponentRatingColor(scope.row) }">
                  {{ scope.row.opponentUsername }}
                </span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="problemTitle" label="题目" align="center"></el-table-column>

          <el-table-column label="结果" width="100" align="center">
            <template slot-scope="scope">
              <el-tag :type="getResultType(scope.row)" size="medium">
                {{ getResultText(scope.row) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="结束原因" width="150" align="center">
            <template slot-scope="scope">
              {{ getEndReasonText(scope.row) }}
            </template>
          </el-table-column>

          <el-table-column label="对战时长" width="100" align="center">
            <template slot-scope="scope">
              {{ formatTime(scope.row.battleTime) }}
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100" align="center">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.isExcluded" type="warning" size="small">
                不计入
              </el-tag>
              <el-tag v-else type="info" size="small">
                已计入
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="时间" width="180" align="center">
            <template slot-scope="scope">
              {{ formatDate(scope.row.gmtCreate) }}
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          class="pagination"
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="limit"
          :current-page="currentPage"
          @current-change="handlePageChange"
        ></el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import { getMyRecords } from '@/api/battle';
import { getRatingColor } from '@/common/rating-utils';

export default {
  name: 'BattleMyRecords',
  data() {
    return {
      recordList: [],
      stats: {
        totalBattles: 0,
        winCount: 0,
        loseCount: 0,
        winRate: 0
      },
      total: 0,
      limit: 10,
      currentPage: 1,
      loading: false,
      // 用于存储全部统计数据
      allTimeStats: {
        totalBattles: 0,
        winCount: 0,
        loseCount: 0,
        winRate: 0
      }
    };
  },
  mounted() {
    this.loadRecords();
    // 添加页面可见性监听，当页面重新可见时刷新数据
    document.addEventListener('visibilitychange', this.handleVisibilityChange);
  },
  beforeDestroy() {
    // 移除事件监听
    document.removeEventListener('visibilitychange', this.handleVisibilityChange);
  },
  activated() {
    // 如果使用了 keep-alive，页面激活时刷新数据
    this.loadRecords();
  },
  methods: {
    handleVisibilityChange() {
      // 当页面从隐藏变为可见时，自动刷新数据
      if (!document.hidden) {
        this.loadRecords();
      }
    },

    getRatingColor(rating) {
      return getRatingColor(rating);
    },

    getOpponentRatingColor(row) {
      // 如果有对手的rating字段，使用它；NULL或undefined时使用默认0分
      const rating = row.opponentRating !== null && row.opponentRating !== undefined
        ? row.opponentRating
        : 0; // 默认初始分数
      return this.getRatingColor(rating);
    },

    async loadRecords() {
      this.loading = true;
      try {
        // 加载当前页数据
        const res = await getMyRecords({
          limit: this.limit,
          currentPage: this.currentPage
        });
        if (res.data.code === 0) {  // 业务状态码：0=成功
          this.recordList = res.data.data.records;
          this.total = res.data.data.total;

          // 调试：打印第一条记录的 opponentRating
          if (this.recordList.length > 0) {
            console.log('[BattleMyRecords] 第一条记录:', this.recordList[0]);
            console.log('[BattleMyRecords] opponentRating:', this.recordList[0].opponentRating);
          }

          // 计算当前页统计
          this.calculateStats();

          // 加载全部数据用于全部统计
          await this.loadAllTimeStats();
        } else {
          this.$message.error(res.data.msg || '加载记录失败');
        }
      } catch (error) {
        this.$message.error('加载记录失败');
      } finally {
        this.loading = false;
      }
    },

    async loadAllTimeStats() {
      try {
        // 获取所有记录（不限制limit）
        const res = await getMyRecords({
          limit: this.total, // 使用总数作为limit
          currentPage: 1
        });
        if (res.data.code === 0 && res.data.data.records) {
          const allRecords = res.data.data.records;
          // 排除不计入的记录
          const validRecords = allRecords.filter(r => !r.isExcluded);
          const total = validRecords.length;
          const wins = validRecords.filter(r => r.isWinner).length;
          const losses = total - wins;

          this.allTimeStats = {
            totalBattles: total,
            winCount: wins,
            loseCount: losses,
            winRate: total > 0 ? ((wins / total) * 100).toFixed(1) : 0
          };
        }
      } catch (error) {
        // 静默处理错误，使用当前页统计作为后备
        console.error('加载全部统计失败:', error);
      }
    },

    calculateStats() {
      // 计算当前页统计（用于其他可能需要的地方）
      // 排除不计入的记录
      const validRecords = this.recordList.filter(r => !r.isExcluded);
      const total = validRecords.length;
      const wins = validRecords.filter(r => r.isWinner).length;
      const losses = total - wins;

      this.stats = {
        totalBattles: total,
        winCount: wins,
        loseCount: losses,
        winRate: total > 0 ? ((wins / total) * 100).toFixed(1) : 0
      };
    },

    handlePageChange(page) {
      this.currentPage = page;
      this.loadRecords();
    },

    backToHome() {
      this.$router.push({ name: 'BattleHome' });
    },

    getResultText(row) {
      // 只显示胜利或失败，不显示具体原因
      return row.isWinner ? '胜利' : '失败';
    },

    getResultType(row) {
      // 根据是否获胜返回标签类型
      return row.isWinner ? 'success' : 'danger';
    },

    getEndReasonText(row) {
      // 显示更详细的结束原因
      if (row.endReason === 'ac') {
        return row.isWinner ? '你已解决' : '对方已解决';
      }
      if (row.endReason === 'giveup') {
        return row.isWinner ? '对方放弃' : '你放弃';
      }
      if (row.endReason === 'timeout') {
        return row.isWinner ? '对方超时' : '你超时';
      }
      return row.endReason || '未知';
    },

    formatTime(seconds) {
      if (!seconds) return '-';
      const minutes = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${minutes}:${secs.toString().padStart(2, '0')}`;
    },

    formatDate(dateStr) {
      if (!dateStr) return '-';
      const date = new Date(dateStr);
      return date.toLocaleString('zh-CN');
    }
  }
};
</script>

<style scoped>
.my-records-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.records-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 2px solid #409eff;
  padding: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.refresh-button {
  background: #67C23A;
  border-color: #67C23A;
  color: white;
  transition: all 0.2s;
}

.refresh-button:hover {
  background: #5daf34;
  border-color: #5daf34;
}

.refresh-button i {
  margin-right: 4px;
}

.back-button {
  background: #fff;
  border: 1px solid #dcdfe6;
  color: #606266;
  transition: all 0.2s;
}

.back-button:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background: #ecf5ff;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.header-title i {
  color: #409eff;
  font-size: 20px;
}

.records-content {
  padding: 24px 20px;
  background: #fff;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  border-radius: 4px;
  transition: all 0.2s;
  border: 1px solid #e4e7ed;
}

.stat-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.stat-card.win {
  background: #f0f9ff;
}

.stat-card.lose {
  background: #fef0f0;
}

.stat-card.rate {
  background: #fdf6ec;
}

.stat-card.total {
  background: #f0f9ff;
}

.stat-item {
  padding: 16px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

/* 美化表格 */
::v-deep .el-table {
  background: transparent;
  border-radius: 4px;
  overflow: hidden;
  margin-top: 16px;
  border: 1px solid #e4e7ed;
}

::v-deep .el-table th {
  background: #f5f7fa !important;
  color: #606266 !important;
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid #e4e7ed;
  padding: 12px 0;
}

::v-deep .el-table td {
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 0;
}

::v-deep .el-table__row {
  transition: background-color 0.2s;
  background: white;
}

::v-deep .el-table__row:hover {
  background: #f5f7fa !important;
}

::v-deep .el-table--enable-row-hover .el-table__body tr:hover > td {
  background: transparent;
}

/* 美化标签 */
::v-deep .el-tag {
  border-radius: 2px;
  padding: 4px 8px;
  font-weight: 500;
  font-size: 12px;
  border: 1px solid;
}

::v-deep .el-tag--success {
  background: #f0f9ff;
  color: #67c23a;
  border-color: #c2e7b0;
}

::v-deep .el-tag--danger {
  background: #fef0f0;
  color: #f56c6c;
  border-color: #fbc4c4;
}

::v-deep .el-tag--warning {
  background: #fdf6ec;
  color: #e6a23c;
  border-color: #f5dab1;
}

::v-deep .el-tag--info {
  background: #f4f4f5;
  color: #909399;
  border-color: #d3d4d6;
}

::v-deep .el-tag.el-tag--small {
  padding: 2px 6px;
  font-size: 12px;
}

::v-deep .el-tag.el-tag--medium {
  padding: 6px 12px;
  font-size: 13px;
}

.opponent-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.opponent-username {
  font-weight: 500;
  font-size: 14px;
}

.opponent-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #dcdfe6;
}

.opponent-avatar-placeholder {
  font-size: 32px;
  color: #c0c4cc;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

::v-deep .el-pagination.is-background .el-pager li:not(.disabled).active {
  background: #409eff;
}

@media screen and (max-width: 768px) {
  .my-records-container {
    padding: 10px;
  }

  .records-content {
    padding: 15px;
  }

  .stats-row .el-col {
    margin-bottom: 12px;
  }

  .stat-value {
    font-size: 24px;
  }

  .stat-label {
    font-size: 12px;
  }

  .header-title {
    font-size: 16px;
  }

  ::v-deep .el-table th {
    font-size: 12px;
    padding: 10px 0;
  }

  ::v-deep .el-table td {
    padding: 8px 0;
  }
}
</style>
