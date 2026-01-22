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
        :data="mergedRecordList"
        style="width: 100%"
        v-loading="loading"
        stripe
        :row-class-name="getRowClassName"
      >
        <el-table-column label="对局" width="250" align="center">
          <template slot-scope="scope">
            <div class="match-cell">
              <div class="player">
                <span class="player-name" :style="{ color: getUserRatingColor(scope.row.record1) }">
                  {{ scope.row.record1.username }}
                </span>
                <el-tag :type="scope.row.record1.isWinner ? 'success' : 'danger'" size="mini">
                  {{ scope.row.record1.isWinner ? '胜' : '负' }}
                </el-tag>
              </div>
              <div class="vs-divider">VS</div>
              <div class="player">
                <span class="player-name" :style="{ color: getOpponentRatingColor(scope.row.record1) }">
                  {{ scope.row.record1.opponentUsername }}
                </span>
                <el-tag :type="scope.row.record2 ? (scope.row.record2.isWinner ? 'success' : 'danger') : (scope.row.record1.isWinner ? 'danger' : 'success')" size="mini">
                  {{ scope.row.record2 ? (scope.row.record2.isWinner ? '胜' : '负') : (scope.row.record1.isWinner ? '负' : '胜') }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="record1.problemTitle" label="题目" align="center" min-width="150"></el-table-column>

        <el-table-column label="结束原因" width="100" align="center">
          <template slot-scope="scope">
            {{ getEndReasonText(scope.row.record1) }}
          </template>
        </el-table-column>

        <el-table-column label="房间号" width="100" align="center">
          <template slot-scope="scope">
            <el-tag size="mini" type="info">{{ scope.row.record1.roomId }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="对战时长" width="80" align="center">
          <template slot-scope="scope">
            {{ formatTime(scope.row.record1.battleTime) }}
          </template>
        </el-table-column>

        <el-table-column label="时间" width="180" align="center">
          <template slot-scope="scope">
            {{ formatDate(scope.row.record1.gmtCreate) }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.isExcluded" type="info" size="small">不计入</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button
              type="text"
              size="small"
              @click="toggleExclude(scope.row)"
              :icon="scope.row.isExcluded ? 'el-icon-check' : 'el-icon-close'"
            >
              {{ scope.row.isExcluded ? '计入' : '不计入' }}
            </el-button>
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
import { getAllBattleRecords, excludeRecord } from '@/api/battle';
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
      mergedRecordList: [], // 合并后的记录列表
      total: 0,
      limit: 20,
      currentPage: 1,
      loading: false
    };
  },
  watch: {
    // 监听 recordList 变化，自动合并
    recordList: {
      handler(newList) {
        this.mergeRecords();
      },
      deep: true,
      immediate: true
    }
  },
  mounted() {
    this.loadRecords();
  },
  methods: {
    // 合并记录：将同一房间的两条记录合并为一条
    mergeRecords() {
      const merged = [];
      const usedIds = new Set();

      for (let i = 0; i < this.recordList.length; i++) {
        const record = this.recordList[i];

        // 如果这条记录已经被合并过，跳过
        if (usedIds.has(record.id)) {
          continue;
        }

        // 查找配对记录（同一房间，用户和对手互换）
        const pairedRecord = this.recordList.find(r =>
          r.id !== record.id &&
          !usedIds.has(r.id) &&
          r.roomId === record.roomId &&
          ((r.userId === record.userId && r.opponentId === record.opponentId) ||
           (r.userId === record.opponentId && r.opponentId === record.userId))
        );

        // 创建合并记录
        const mergedRecord = {
          record1: record,
          record2: pairedRecord || null,
          isExcluded: record.isExcluded,
          roomId: record.roomId
        };

        merged.push(mergedRecord);
        usedIds.add(record.id);

        // 如果找到配对记录，也标记为已使用
        if (pairedRecord) {
          usedIds.add(pairedRecord.id);
        }
      }

      this.mergedRecordList = merged;
    },

    getRowClassName({ row, rowIndex }) {
      // 合并后不需要特殊样式，每行就是一个完整的对局
      return '';
    },
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
        // 请求翻倍的数据量,因为合并后记录数会减半
        const params = {
          limit: this.limit * 2,
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
          // total 除以 2 向上取整,因为每两条记录合并为一条
          this.total = Math.ceil(res.data.data.total / 2);
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
    },

    async toggleExclude(mergedRow) {
      const action = mergedRow.isExcluded ? '计入' : '不计入';
      const newExcludedState = !mergedRow.isExcluded;

      try {
        await this.$confirm(
          `确定要将该场对决${action}吗?`,
          '提示',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        );

        // 只调用一次 API，后端会自动更新同一房间的配对记录
        await excludeRecord({
          recordId: mergedRow.record1.id,
          isExcluded: newExcludedState
        });

        this.$message.success(`已${action}本场对决`);

        // 更新本地状态
        mergedRow.record1.isExcluded = newExcludedState;
        if (mergedRow.record2) {
          mergedRow.record2.isExcluded = newExcludedState;
        }
        mergedRow.isExcluded = newExcludedState;

      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('操作失败');
        }
      }
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

/* 对局单元格样式 */
.match-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
}

.player {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}

.player-name {
  font-weight: 600;
  font-size: 14px;
}

.vs-divider {
  font-size: 12px;
  font-weight: 700;
  color: #909399;
  letter-spacing: 2px;
}

/* 同一房间的配对记录样式 */
::v-deep .el-table .paired-row {
  background-color: #fef9e6 !important;
}

::v-deep .el-table .paired-row:hover {
  background-color: #fdf0d7 !important;
}

/* 不计入的记录样式 */
::v-deep .el-table .excluded-row {
  opacity: 0.6;
  background-color: #f5f5f5 !important;
}
</style>
