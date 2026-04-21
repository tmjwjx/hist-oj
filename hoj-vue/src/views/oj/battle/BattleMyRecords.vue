<template>
  <div class="my-records-container">
    <el-card class="records-card">
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
            <el-card class="stat-card total" shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.totalBattles }}</div>
                <div class="stat-label">全部场次</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card win" shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.winCount }}</div>
                <div class="stat-label">全部胜场</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card lose" shadow="hover">
              <div class="stat-item">
                <div class="stat-value">{{ allTimeStats.loseCount }}</div>
                <div class="stat-label">负场</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card rate" shadow="hover">
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
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.records-card {
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  animation: fadeInUp 0.6s ease;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #e3f2fd;
  padding: 25px 30px;
  border-radius: 16px 16px 0 0;
  position: relative;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  flex: 1;
}

.refresh-button {
  background: linear-gradient(135deg, #67C23A 0%, #5daf34 100%);
  border: none;
  color: white;
  font-weight: 600;
  transition: all 0.3s ease;
}

.refresh-button:hover {
  background: linear-gradient(135deg, #5daf34 0%, #4a9628 100%);
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.4);
}

.refresh-button i {
  margin-right: 5px;
}

.back-button {
  background: #ffffff;
  border: 2px solid #bbdefb;
  color: #1976d2;
  font-weight: 600;
  transition: all 0.3s ease;
}

.back-button:hover {
  background: #f5f5f5;
  border-color: #90caf9;
  transform: translateX(-3px);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: #1565c0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: #1565c0;
}

.header-title i {
  color: #1976d2;
  font-size: 28px;
}

.records-content {
  padding: 30px;
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  border-radius: 0 0 16px 16px;
}

.stats-row {
  margin-bottom: 30px;
}

.stat-card {
  text-align: center;
  border-radius: 12px;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0;
  transition: opacity 0.3s;
  background: linear-gradient(135deg, rgba(79, 172, 254, 0.05) 0%, rgba(0, 242, 254, 0.05) 100%);
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-card:hover {
  transform: translateY(-8px) scale(1.02);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
  border-color: #4facfe;
}

.stat-card.win {
  background: linear-gradient(135deg, #f0f9ff 0%, #e1f3ff 100%);
}

.stat-card.lose {
  background: linear-gradient(135deg, #fef0f0 0%, #fde2e2 100%);
}

.stat-card.rate {
  background: linear-gradient(135deg, #fff9e6 0%, #fff3cd 100%);
}

.stat-card.total {
  background: linear-gradient(135deg, #f0fff4 0%, #dcfce7 100%);
}

.stat-item {
  padding: 20px 10px;
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 36px;
  font-weight: 800;
  color: #303133;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  transition: all 0.3s ease;
}

.stat-card:hover .stat-value {
  transform: scale(1.1);
}

.stat-label {
  font-size: 15px;
  color: #909399;
  font-weight: 500;
}

/* 美化表格 */
::v-deep .el-table {
  background: transparent;
  border-radius: 12px;
  overflow: hidden;
  margin-top: 20px;
}

::v-deep .el-table__header-wrapper {
  border-radius: 12px 12px 0 0;
}

::v-deep .el-table th {
  background: #e3f2fd !important;
  color: #1976d2 !important;
  font-weight: 600;
  font-size: 15px;
  border: none;
  padding: 18px 0;
}

::v-deep .el-table td {
  border: none;
  padding: 15px 0;
}

::v-deep .el-table__row {
  transition: all 0.3s ease;
  background: white;
}

::v-deep .el-table__row:hover {
  background: linear-gradient(135deg, #f0f9ff 0%, #e1f3ff 100%) !important;
  transform: scale(1.01);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

::v-deep .el-table--enable-row-hover .el-table__body tr:hover > td {
  background: transparent;
}

/* 美化标签 */
::v-deep .el-tag {
  border-radius: 20px;
  padding: 10px 20px;
  font-weight: 700;
  font-size: 15px;
  border: none;
  letter-spacing: 1px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

::v-deep .el-tag--success {
  background: linear-gradient(135deg, #67C23A 0%, #5daf34 100%);
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.4);
  color: white !important;
}

::v-deep .el-tag--danger {
  background: linear-gradient(135deg, #F56C6C 0%, #f15454 100%);
  box-shadow: 0 4px 12px rgba(245, 108, 108, 0.4);
  color: white !important;
}

::v-deep .el-tag--warning {
  background: linear-gradient(135deg, #E6A23C 0%, #d9972a 100%);
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.4);
  color: white !important;
}

::v-deep .el-tag--info {
  background: linear-gradient(135deg, #909399 0%, #7a7d82 100%);
  box-shadow: 0 4px 12px rgba(144, 147, 153, 0.3);
  color: white !important;
}

::v-deep .el-tag.el-tag--small {
  padding: 6px 12px;
  font-size: 12px;
}

.opponent-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.opponent-username {
  font-weight: 600;
  font-size: 14px;
  transition: all 0.3s ease;
}

::v-deep .el-table__row:hover .opponent-username {
  transform: scale(1.05);
}

.opponent-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #e0e0e0;
  transition: all 0.3s ease;
}

::v-deep .el-table__row:hover .opponent-avatar {
  transform: scale(1.2) rotate(5deg);
  border-color: #4facfe;
  box-shadow: 0 4px 8px rgba(79, 172, 254, 0.3);
}

.opponent-avatar-placeholder {
  font-size: 40px;
  color: #c0c4cc;
  transition: all 0.3s ease;
}

::v-deep .el-table__row:hover .opponent-avatar-placeholder {
  transform: scale(1.2);
  color: #4facfe;
}

.pagination {
  margin-top: 30px;
  text-align: center;
}

::v-deep .el-pagination.is-background .el-pager li:not(.disabled).active {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  border-radius: 8px;
}

::v-deep .el-pagination.is-background .el-pager li {
  border-radius: 8px;
  transition: all 0.3s ease;
}

::v-deep .el-pagination.is-background .el-pager li:hover {
  transform: scale(1.1);
}

@media screen and (max-width: 768px) {
  .my-records-container {
    padding: 10px;
  }

  .records-content {
    padding: 15px;
  }

  .stats-row .el-col {
    margin-bottom: 15px;
  }

  .stat-value {
    font-size: 28px;
  }

  .stat-label {
    font-size: 13px;
  }

  .header-title {
    font-size: 18px;
  }

  ::v-deep .el-table th {
    font-size: 13px;
    padding: 12px 0;
  }

  ::v-deep .el-table td {
    padding: 10px 0;
  }
}
</style>
