<template>
  <div>
    <el-card>
      <div slot="header">
        <div class="header-container">
          <span class="panel-title home-title">{{ trainingTitle }} - 参与者列表</span>
          <el-button
            size="small"
            @click="goBack"
            icon="el-icon-back"
          >返回训练列表
          </el-button>
        </div>
        <div class="filter-row">
          <vxe-input
            v-model="searchKeyword"
            placeholder="搜索用户名、真实姓名、昵称"
            type="search"
            size="medium"
            @search-click="handleSearch"
            @keyup.enter.native="handleSearch"
            style="width: 300px"
          ></vxe-input>
          <el-select
            v-model="statusFilter"
            placeholder="筛选状态"
            clearable
            size="medium"
            @change="handleSearch"
            style="width: 150px"
          >
            <el-option label="全部" value=""></el-option>
            <el-option label="未开始" value="not_started"></el-option>
            <el-option label="进行中" value="in_progress"></el-option>
            <el-option label="已完成" value="completed"></el-option>
          </el-select>
        </div>
      </div>

      <vxe-table
        :loading="loading"
        ref="xTable"
        :data="pagedParticipants"
        auto-resize
        stripe
        align="center"
      >
        <vxe-table-column field="id" width="80" title="ID"></vxe-table-column>
        <vxe-table-column field="user.username" width="150" title="用户名">
          <template v-slot="{ row }">
            <span v-if="row.user" :style="{ color: getRatingColor(row.user.rating || 0) }">
              {{ row.user.username }}
            </span>
            <span v-else>-</span>
          </template>
        </vxe-table-column>
        <vxe-table-column title="真实姓名" width="150">
          <template v-slot="{ row }">
            <span v-if="row.user">
              {{ row.user.realname || row.user.nickname || row.user.username || '-' }}
            </span>
            <span v-else>-</span>
          </template>
        </vxe-table-column>
        <vxe-table-column title="昵称" width="150">
          <template v-slot="{ row }">
            <span v-if="row.user">{{ row.user.nickname || '-' }}</span>
            <span v-else>-</span>
          </template>
        </vxe-table-column>
        <vxe-table-column title="状态" width="120">
          <template v-slot="{ row }">
            <el-tag :type="getParticipantStatusType(row)" size="small">
              {{ getParticipantStatusText(row) }}
            </el-tag>
          </template>
        </vxe-table-column>
        <vxe-table-column title="完成进度" width="200">
          <template v-slot="{ row }">
            <el-tooltip
              effect="dark"
              :content="`${getParticipantSolvedCount(row)}/${getParticipantTotalCount(row)}`"
              placement="top"
            >
              <el-progress
                :text-inside="true"
                :stroke-width="20"
                :percentage="getParticipantProgress(row)"
              ></el-progress>
            </el-tooltip>
          </template>
        </vxe-table-column>
        <vxe-table-column title="完成情况" width="150">
          <template v-slot="{ row }">
            <span>{{ getParticipantSolvedCount(row) }} / {{ getParticipantTotalCount(row) }}</span>
          </template>
        </vxe-table-column>
        <vxe-table-column prop="joinTime" title="加入时间" width="180">
          <template v-slot="{ row }">
            {{ row.joinTime ? $options.filters.localtime(row.joinTime) : '-' }}
          </template>
        </vxe-table-column>
        <template v-slot:empty>
          <div v-if="!loading">
            <span v-if="participantsList.length === 0">暂无参与者</span>
            <span v-else>未找到匹配的参与者</span>
          </div>
        </template>
      </vxe-table>

      <div class="panel-options">
        <el-pagination
          class="page"
          layout="total, prev, pager, next"
          :current-page="currentPage"
          :page-size="pageSize"
          :total="filteredParticipants.length"
          @current-change="currentChange"
        >
        </el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import api from '@/common/api';
import myMessage from '@/common/message';
import { getRatingColor } from '@/common/rating-utils';

export default {
  name: 'TrainingParticipants',
  data() {
    return {
      trainingId: null,
      trainingTitle: '',
      participantsList: [],
      filteredParticipants: [],
      loading: false,
      searchKeyword: '',
      statusFilter: '',
      currentPage: 1,
      pageSize: 20,
    };
  },
  computed: {
    pagedParticipants() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredParticipants.slice(start, end);
    },
  },
  mounted() {
    // 设置页面标题,避免显示国际化key
    document.title = '训练参与者列表';
    this.trainingId = this.$route.params.trainingId;
    this.trainingTitle = this.$route.query.title || `训练 ${this.trainingId}`;
    this.getParticipants();
  },
  methods: {
    getParticipants() {
      this.loading = true;
      const { getTrainingParticipants } = require('@/api/training');
      getTrainingParticipants(this.trainingId).then(
        (res) => {
          this.loading = false;
          if (res.data.code === 200) {
            this.participantsList = res.data.data || [];
            this.filteredParticipants = [...this.participantsList];
          } else {
            myMessage.error(res.data.message || '获取参与者列表失败');
          }
        },
        (err) => {
          this.loading = false;
          myMessage.error('获取参与者列表失败');
        }
      );
    },
    handleSearch() {
      this.currentPage = 1;
      this.filterParticipants();
    },
    filterParticipants() {
      let filtered = [...this.participantsList];

      // 关键词搜索
      if (this.searchKeyword) {
        const keyword = this.searchKeyword.toLowerCase();
        filtered = filtered.filter((participant) => {
          if (!participant.user) return false;
          const username = (participant.user.username || '').toLowerCase();
          const realname = (participant.user.realname || '').toLowerCase();
          const nickname = (participant.user.nickname || '').toLowerCase();
          return (
            username.includes(keyword) ||
            realname.includes(keyword) ||
            nickname.includes(keyword)
          );
        });
      }

      // 状态筛选
      if (this.statusFilter) {
        filtered = filtered.filter((participant) => {
          const status = this.getParticipantActualStatus(participant);
          return status === this.statusFilter;
        });
      }

      this.filteredParticipants = filtered;
    },
    currentChange(page) {
      this.currentPage = page;
    },
    goBack() {
      this.$router.push({ name: 'admin-training-list' });
    },
    getParticipantSolvedCount(participant) {
      if (participant.progress && participant.progress.solvedCount !== undefined) {
        return participant.progress.solvedCount;
      }
      return 0;
    },
    getParticipantTotalCount(participant) {
      if (participant.progress && participant.progress.totalCount !== undefined) {
        return participant.progress.totalCount;
      }
      return 0;
    },
    getParticipantProgress(participant) {
      const solved = this.getParticipantSolvedCount(participant);
      const total = this.getParticipantTotalCount(participant);
      if (!total || total === 0) {
        return 0;
      }
      return Number(((solved / total) * 100).toFixed(2));
    },
    getParticipantActualStatus(participant) {
      const solved = this.getParticipantSolvedCount(participant);
      const total = this.getParticipantTotalCount(participant);

      if (!total || total === 0) {
        return 'not_started';
      } else if (solved >= total) {
        return 'completed';
      } else if (solved > 0) {
        return 'in_progress';
      } else {
        return 'not_started';
      }
    },
    getParticipantStatusText(participant) {
      const status = this.getParticipantActualStatus(participant);
      const statusMap = {
        not_started: '未开始',
        in_progress: '进行中',
        completed: '已完成',
      };
      return statusMap[status] || '未知';
    },
    getParticipantStatusType(participant) {
      const status = this.getParticipantActualStatus(participant);
      const typeMap = {
        not_started: 'info',
        in_progress: 'warning',
        completed: 'success',
      };
      return typeMap[status] || 'info';
    },
    getRatingColor,
  },
};
</script>

<style scoped>
.header-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.filter-row {
  display: flex;
  gap: 15px;
  align-items: center;
}

.panel-options {
  margin-top: 20px;
  text-align: center;
}

.page {
  display: inline-block;
}
</style>
