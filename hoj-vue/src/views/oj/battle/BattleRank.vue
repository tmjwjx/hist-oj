<template>
  <div class="battle-rank-container">
    <el-card class="rank-card">
      <div slot="header" class="card-header">
        <div class="header-left">
          <el-button type="primary" size="small" @click="backToHome" class="back-button">
            <i class="fa fa-arrow-left"></i> 返回
          </el-button>
          <div class="header-title">
            <i class="fa fa-ranking-star"></i>
            <span>对战排行榜</span>
          </div>
        </div>
      </div>

      <div class="rank-content">
        <!-- 搜索框 -->
        <div class="search-box">
          <el-input
            v-model="searchUsername"
            placeholder="输入用户名搜索"
            clearable
            @clear="handleSearch"
            @keyup.enter.native="handleSearch"
            prefix-icon="el-icon-search"
            class="search-input"
          >
            <el-button slot="append" @click="handleSearch" icon="el-icon-search">搜索</el-button>
          </el-input>
          <el-button v-if="searchUsername" @click="handleReset" icon="el-icon-refresh" style="margin-left: 10px;">重置</el-button>
        </div>

        <el-table
          :data="rankList"
          style="width: 100%"
          v-loading="loading"
          :row-class-name="tableRowClassName"
        >
          <el-table-column label="排名" width="100" align="center">
            <template slot-scope="scope">
              <!-- 显示数字排名，前三名使用特殊样式 -->
              <span :class="getRankClass(getRankDisplay(scope.$index))">
                {{ getRankDisplay(scope.$index) }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="用户" width="200" align="center">
            <template slot-scope="scope">
              <div class="user-cell">
                <i class="fa fa-user-circle fa-2x user-avatar"></i>
                <span class="username" :style="{ color: getRatingColor(scope.row.rating) }">
                  {{ scope.row.username }}
                </span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="totalBattles" label="总场次" align="center" sortable></el-table-column>

          <el-table-column prop="winCount" label="胜场" align="center" sortable>
            <template slot-scope="scope">
              <span style="color: #67C23A; font-weight: bold;">{{ scope.row.winCount }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="loseCount" label="负场" align="center" sortable>
            <template slot-scope="scope">
              <span style="color: #F56C6C;">{{ scope.row.loseCount }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="winRate" label="胜率" align="center" sortable>
            <template slot-scope="scope">
              <el-progress
                :percentage="parseFloat(scope.row.winRate.toFixed(2))"
                :color="getProgressColor(scope.row.winRate)"
              ></el-progress>
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
import { getBattleRank } from '@/api/battle';
import { getRatingColor } from '@/common/rating-utils';

export default {
  name: 'BattleRank',
  data() {
    return {
      rankList: [],
      total: 0,
      limit: 50,
      currentPage: 1,
      loading: false,
      searchUsername: '' // 搜索用户名
    };
  },
  mounted() {
    this.loadRank();
  },
  methods: {
    getRatingColor(rating) {
      return getRatingColor(rating);
    },
    async loadRank() {
      this.loading = true;
      try {
        const params = {
          limit: this.limit,
          currentPage: this.currentPage
        };

        // 添加搜索条件
        if (this.searchUsername) {
          params.username = this.searchUsername;
        }

        const res = await getBattleRank(params);
        console.log('Battle rank response:', res.data);
        if (res.data.code === 0) {  // 业务状态码：0=成功
          this.rankList = res.data.data.rankList;
          this.total = res.data.data.total;
          console.log('Rank list:', this.rankList);

          // 详细的调试输出
          if (this.rankList.length > 0) {
            console.log('=== 排行榜详细信息 ===');
            this.rankList.forEach((user, index) => {
              console.log(`#${index + 1} ${user.username}:`, {
                totalBattles: user.totalBattles,
                winCount: user.winCount,
                loseCount: user.loseCount,
                winRate: user.winRate
              });
            });
            console.log('==================');
          }
        } else {
          this.$message.error(res.data.msg || '加载排行榜失败');
        }
      } catch (error) {
        console.error('Load rank error:', error);
        this.$message.error('加载排行榜失败');
      } finally {
        this.loading = false;
      }
    },

    handlePageChange(page) {
      this.currentPage = page;
      this.loadRank();
    },

    handleSearch() {
      this.currentPage = 1;
      this.loadRank();
    },

    handleReset() {
      this.searchUsername = '';
      this.currentPage = 1;
      this.loadRank();
    },

    getRankDisplay(index) {
      // 优先使用后端返回的 rank 字段
      if (this.rankList[index] && this.rankList[index].rank !== undefined && this.rankList[index].rank !== null) {
        return this.rankList[index].rank;
      }
      // 如果后端没有返回rank，则计算全局排名：(当前页 - 1) * 每页数量 + 当前索引 + 1
      return (this.currentPage - 1) * this.limit + index + 1;
    },

    tableRowClassName({ rowIndex }) {
      const userId = this.$store.getters.userInfo?.uuid;
      if (this.rankList[rowIndex].userId === userId) {
        return 'my-row';
      }
      return '';
    },

    getProgressColor(winRate) {
      if (winRate >= 70) return '#67C23A';
      if (winRate >= 50) return '#409EFF';
      if (winRate >= 30) return '#E6A23C';
      return '#F56C6C';
    },

    getRankClass(rank) {
      if (rank === 1) return 'rank-badge rank-1';
      if (rank === 2) return 'rank-badge rank-2';
      if (rank === 3) return 'rank-badge rank-3';
      return 'rank-number';
    },

    backToHome() {
      this.$router.push({ name: 'BattleHome' });
    }
  }
};
</script>

<style scoped>
.battle-rank-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.rank-card {
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
  color: #e6a23c;
  font-size: 20px;
}

.rank-content {
  padding: 24px 20px;
  background: #fff;
}

.search-box {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}

.search-input {
  max-width: 400px;
}

/* 美化表格 */
::v-deep .el-table {
  background: transparent;
  border-radius: 4px;
  overflow: hidden;
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

.rank-badge {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  min-width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 14px;
  font-weight: 600;
}

.rank-1 {
  background: #ffd700;
  color: #8b6914;
}

.rank-2 {
  background: #c0c0c0;
  color: #4a4a4a;
}

.rank-3 {
  background: #cd7f32;
  color: #5c3a1a;
}

.rank-number {
  font-size: 14px;
  font-weight: 500;
  color: #606266;
}

.user-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.user-avatar {
  color: #c0c4cc;
}

.username {
  font-weight: 500;
  font-size: 14px;
}

/* 美化进度条 */
::v-deep .el-progress-bar__outer {
  background: #e4e7ed;
  border-radius: 4px;
  height: 10px !important;
}

::v-deep .el-progress-bar__inner {
  border-radius: 4px;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

::v-deep .el-pagination.is-background .el-pager li:not(.disabled).active {
  background: #409eff;
}

/* 高亮自己的行 */
::v-deep .el-table .my-row {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

::v-deep .el-table .my-row:hover > td {
  background: #ecf5ff !important;
}

@media screen and (max-width: 768px) {
  .battle-rank-container {
    padding: 10px;
  }

  .rank-content {
    padding: 15px;
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
