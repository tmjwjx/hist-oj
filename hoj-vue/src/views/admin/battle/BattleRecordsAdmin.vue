<template>
  <div class="battle-records-admin">
    <el-card class="search-card" shadow="never">
      <div slot="header" class="card-header">
        <span class="title">{{ $t('m.BatRec_Query_Title') }}</span>
      </div>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item :label="$t('m.BatRec_Username')">
          <el-input
            v-model="searchForm.username"
            :placeholder="$t('m.BatRec_Username_Search_Placeholder')"
            clearable
            @clear="handleSearch"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.BatRec_Room_Number')">
          <el-input
            v-model="searchForm.roomId"
            :placeholder="$t('m.BatRec_Room_Number_Placeholder')"
            clearable
            @input="searchForm.roomId = searchForm.roomId.toUpperCase()"
            @clear="handleSearch"
            style="text-transform: uppercase;"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.BatRec_Battle_Result')">
          <el-select v-model="searchForm.result" :placeholder="$t('m.BatRec_All')" clearable @change="handleSearch">
            <el-option :label="$t('m.BatRec_Win')" value="win"></el-option>
            <el-option :label="$t('m.BatRec_Lose')" value="lose"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch" icon="el-icon-search">{{ $t('m.BatRec_Search') }}</el-button>
          <el-button @click="handleReset" icon="el-icon-refresh">{{ $t('m.BatRec_Reset') }}</el-button>
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
        <el-table-column :label="$t('m.BatRec_Match')" width="250" align="center">
          <template slot-scope="scope">
            <div class="match-cell">
              <div class="player">
                <span class="player-name" :style="{ color: getUserRatingColor(scope.row.record1) }">
                  {{ scope.row.record1.username }}
                </span>
                <el-tag :type="scope.row.record1.isWinner ? 'success' : 'danger'" size="mini">
                  {{ scope.row.record1.isWinner ? $t('m.BatRec_Win_Short') : $t('m.BatRec_Lose_Short') }}
                </el-tag>
              </div>
              <div class="vs-divider">VS</div>
              <div class="player">
                <span class="player-name" :style="{ color: getOpponentRatingColor(scope.row.record1) }">
                  {{ scope.row.record1.opponentUsername }}
                </span>
                <el-tag :type="scope.row.record2 ? (scope.row.record2.isWinner ? 'success' : 'danger') : (scope.row.record1.isWinner ? 'danger' : 'success')" size="mini">
                  {{ scope.row.record2 ? (scope.row.record2.isWinner ? $t('m.BatRec_Win_Short') : $t('m.BatRec_Lose_Short')) : (scope.row.record1.isWinner ? $t('m.BatRec_Lose_Short') : $t('m.BatRec_Win_Short')) }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="record1.problemTitle" :label="$t('m.BatRec_Problem')" align="center" min-width="150"></el-table-column>

        <el-table-column :label="$t('m.BatRec_End_Reason')" width="100" align="center">
          <template slot-scope="scope">
            {{ getEndReasonText(scope.row.record1) }}
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.BatRec_Room_Number')" width="100" align="center">
          <template slot-scope="scope">
            <el-tag size="mini" type="info">{{ scope.row.record1.roomId }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.BatRec_Battle_Duration')" width="80" align="center">
          <template slot-scope="scope">
            {{ formatTime(scope.row.record1.battleTime) }}
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.BatRec_Time')" width="180" align="center">
          <template slot-scope="scope">
            {{ formatDate(scope.row.record1.gmtCreate) }}
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.BatRec_Status')" width="100" align="center">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.isExcluded" type="info" size="small">{{ $t('m.BatRec_Not_Counted') }}</el-tag>
            <el-tag v-else type="success" size="small">{{ $t('m.BatRec_Normal') }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.BatRec_Operation')" width="120" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button
              type="text"
              size="small"
              @click="toggleExclude(scope.row)"
              :icon="scope.row.isExcluded ? 'el-icon-check' : 'el-icon-close'"
            >
              {{ scope.row.isExcluded ? $t('m.BatRec_Count_In') : $t('m.BatRec_Not_Count_Action') }}
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
      // 如果有用户的rating字段，使用它；NULL或undefined时使用默认0分
      const rating = row.userRating !== null && row.userRating !== undefined
        ? row.userRating
        : 0; // 默认初始分数
      return this.getRatingColor(rating);
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
          this.$message.error(res.data.msg || this.$t('m.BatRec_Load_Failed'));
        }
      } catch (error) {
        this.$message.error(this.$t('m.BatRec_Load_Failed'));
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
        return this.$t('m.BatRec_End_Reason_AC');
      }
      if (row.endReason === 'giveup') {
        return this.$t('m.BatRec_End_Reason_Giveup');
      }
      if (row.endReason === 'timeout') {
        return this.$t('m.BatRec_End_Reason_Timeout');
      }
      return row.endReason || this.$t('m.BatRec_Unknown');
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
      const action = mergedRow.isExcluded ? this.$t('m.BatRec_Count_In') : this.$t('m.BatRec_Not_Count_Action');
      const newExcludedState = !mergedRow.isExcluded;

      try {
        await this.$confirm(
          this.$t('m.BatRec_Confirm_Toggle', { action: action }),
          this.$t('m.BatRec_Tip'),
          {
            confirmButtonText: this.$t('m.BatRec_Confirm'),
            cancelButtonText: this.$t('m.BatRec_Cancel'),
            type: 'warning'
          }
        );

        // 只调用一次 API，后端会自动更新同一房间的配对记录
        await excludeRecord({
          recordId: mergedRow.record1.id,
          isExcluded: newExcludedState
        });

        this.$message.success(this.$t('m.BatRec_Toggle_Success', { action: action }));

        // 更新本地状态
        mergedRow.record1.isExcluded = newExcludedState;
        if (mergedRow.record2) {
          mergedRow.record2.isExcluded = newExcludedState;
        }
        mergedRow.isExcluded = newExcludedState;

      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error(this.$t('m.BatRec_Operation_Failed'));
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
