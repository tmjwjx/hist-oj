<template>
  <div class="battle-records-admin">
    <el-card class="search-card" shadow="never">
      <div slot="header" class="card-header">
        <span class="title">对战记录查询</span>
      </div>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="用户名">
          <el-input
            v-model="searchForm.username"
            placeholder="输入用户名搜索"
            clearable
            @clear="handleSearch"
          ></el-input>
        </el-form-item>
        <el-form-item label="房间号">
          <el-input
            v-model="searchForm.roomId"
            placeholder="输入房间号"
            clearable
            @input="searchForm.roomId = searchForm.roomId.toUpperCase()"
            @clear="handleSearch"
            style="text-transform: uppercase;"
          ></el-input>
        </el-form-item>
        <el-form-item label="对战结果">
          <el-select v-model="searchForm.result" placeholder="全部" clearable @change="handleSearch">
            <el-option label="胜利" value="win"></el-option>
            <el-option label="失败" value="lose"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch" icon="el-icon-search">搜索</el-button>
          <el-button @click="handleReset" icon="el-icon-refresh">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 记录列表 -->
    <el-card class="records-card" shadow="never">
      <el-table
        :data="recordList"
        style="width: 100%"
        v-loading="loading"
        stripe
      >
        <el-table-column label="用户" width="120" align="center">
          <template slot-scope="scope">
            <div class="user-cell">
              <span class="username" :style="{ color: getUserRatingColor(scope.row) }">
                {{ scope.row.username }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="对手" width="120" align="center">
          <template slot-scope="scope">
            <div class="opponent-cell">
              <span class="opponent-username" :style="{ color: getOpponentRatingColor(scope.row) }">
                {{ scope.row.opponentUsername }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="problemTitle" label="题目" align="center" min-width="150"></el-table-column>

        <el-table-column label="结果" width="80" align="center">
          <template slot-scope="scope">
            <el-tag :type="scope.row.isWinner ? 'success' : 'danger'" size="small">
              {{ scope.row.isWinner ? '胜' : '负' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="结束原因" width="100" align="center">
          <template slot-scope="scope">
            {{ getEndReasonText(scope.row) }}
          </template>
        </el-table-column>

        <el-table-column label="房间号" width="100" align="center">
          <template slot-scope="scope">
            <el-tag size="mini" type="info">{{ scope.row.roomId }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="对战时长" width="80" align="center">
          <template slot-scope="scope">
            {{ formatTime(scope.row.battleTime) }}
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
        layout="total, prev, pager, next, jumper"
        :total="total"
        :page-size="limit"
        :current-page="currentPage"
        @current-change="handlePageChange"
      ></el-pagination>
    </el-card>
  </div>
</template>

<script>
import { getAllBattleRecords } from '@/api/battle';
import { getRatingColor } from '@/common/rating-utils';

export default {
  name: 'BattleRecordsAdmin',
  data() {
    return {
      searchForm: {
        username: '',
        roomId: '',
        result: ''
      },
      recordList: [],
      total: 0,
      limit: 20,
      currentPage: 1,
      loading: false
    };
  },
  mounted() {
    this.loadRecords();
  },
  methods: {
    getRatingColor(rating) {
      return getRatingColor(rating);
    },

    getUserRatingColor(row) {
      // 如果有用户的rating字段，使用它；NULL或undefined时使用默认1200分
      const rating = row.userRating !== null && row.userRating !== undefined
        ? row.userRating
        : 1200; // 默认初始分数
      return this.getRatingColor(rating);
    },

    getOpponentRatingColor(row) {
      // 如果有对手的rating字段，使用它；NULL或undefined时使用默认1200分
      const rating = row.opponentRating !== null && row.opponentRating !== undefined
        ? row.opponentRating
        : 1200; // 默认初始分数
      return this.getRatingColor(rating);
    },

    async loadRecords() {
      this.loading = true;
      try {
        const params = {
          limit: this.limit,
          currentPage: this.currentPage
        };

        // 添加搜索条件
        if (this.searchForm.username) {
          params.username = this.searchForm.username;
        }
        if (this.searchForm.roomId) {
          params.roomId = this.searchForm.roomId;
        }
        if (this.searchForm.result) {
          params.isWinner = this.searchForm.result === 'win' ? 'true' : 'false';
        }

        const res = await getAllBattleRecords(params);
        if (res.data.code === 0) {
          this.recordList = res.data.data.records;
          this.total = res.data.data.total;
        } else {
          this.$message.error(res.data.msg || '加载记录失败');
        }
      } catch (error) {
        this.$message.error('加载记录失败');
      } finally {
        this.loading = false;
      }
    },

    handleSearch() {
      this.currentPage = 1;
      this.loadRecords();
    },

    handleReset() {
      this.searchForm = {
        username: '',
        roomId: '',
        result: ''
      };
      this.currentPage = 1;
      this.loadRecords();
    },

    handlePageChange(page) {
      this.currentPage = page;
      this.loadRecords();
    },

    getEndReasonText(row) {
      if (row.endReason === 'ac') {
        return 'AC解决';
      }
      if (row.endReason === 'giveup') {
        return '放弃';
      }
      if (row.endReason === 'timeout') {
        return '超时';
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
.battle-records-admin {
  padding: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-form {
  margin-bottom: 0;
}

.records-card {
  border-radius: 8px;
}

.user-cell,
.opponent-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.username,
.opponent-username {
  font-weight: 600;
  font-size: 14px;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

::v-deep .el-table th {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

::v-deep .el-table__row:hover {
  background-color: #f5f7fa;
}
</style>
