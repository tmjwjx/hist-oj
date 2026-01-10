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
          if (this.rankList.length > 0) {
            console.log('First item rank:', this.rankList[0].rank);
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
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.rank-card {
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

.header-title i {
  color: #FFD700;
  font-size: 28px;
  text-shadow: 0 2px 10px rgba(255, 215, 0, 0.5);
  animation: glow 2s ease-in-out infinite;
}

@keyframes glow {
  0%, 100% {
    filter: brightness(1);
    text-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
  }
  50% {
    filter: brightness(1.2);
    text-shadow: 0 0 20px rgba(255, 215, 0, 0.8);
  }
}

.rank-content {
  padding: 30px;
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
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
  border-radius: 12px;
  overflow: hidden;
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

.rank-badge {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  min-width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 18px;
  font-weight: bold;
  animation: bounceIn 0.6s ease;
}

.rank-1 {
  background: linear-gradient(135deg, #FFE5B4 0%, #FFDAB9 100%);
  color: #D2691E;
  box-shadow: 0 2px 8px rgba(255, 215, 0, 0.3);
}

.rank-2 {
  background: linear-gradient(135deg, #E8E8E8 0%, #D3D3D3 100%);
  color: #696969;
  box-shadow: 0 2px 8px rgba(192, 192, 192, 0.3);
}

.rank-3 {
  background: linear-gradient(135deg, #F5DEB3 0%, #DEB887 100%);
  color: #CD853F;
  box-shadow: 0 2px 8px rgba(205, 127, 50, 0.3);
}

@keyframes bounceIn {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.rank-number {
  font-size: 18px;
  font-weight: 600;
  color: #909399;
}

.user-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.user-avatar {
  color: #c0c4cc;
  transition: all 0.3s ease;
}

.username {
  font-weight: 600;
  font-size: 16px;
  transition: all 0.3s ease;
}

::v-deep .el-table__row:hover .user-avatar {
  transform: scale(1.2);
  color: #667eea;
}

::v-deep .el-table__row:hover .username {
  transform: scale(1.05);
}

/* 美化进度条 */
::v-deep .el-progress-bar__outer {
  background: linear-gradient(90deg, #e0e0e0 0%, #f5f5f5 100%);
  border-radius: 10px;
  height: 12px !important;
}

::v-deep .el-progress-bar__inner {
  border-radius: 10px;
  transition: all 0.3s ease;
}

::v-deep .el-table__row:hover .el-progress-bar__inner {
  transform: scaleY(1.2);
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

/* 高亮自己的行 */
::v-deep .el-table .my-row {
  background: linear-gradient(135deg, #f0f9ff 0%, #e1f3ff 100%);
  box-shadow: inset 3px 0 0 #4facfe;
}

::v-deep .el-table .my-row:hover > td {
  background: linear-gradient(135deg, #e1f3ff 0%, #d1e9ff 100%) !important;
  box-shadow: inset 3px 0 0 #4facfe, 0 4px 12px rgba(79, 172, 254, 0.2);
}

@media screen and (max-width: 768px) {
  .battle-rank-container {
    padding: 10px;
  }

  .rank-content {
    padding: 15px;
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
