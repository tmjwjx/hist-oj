<template>
  <div>
    <el-card>
      <div slot="header">
        <span class="panel-title">
          <i class="el-icon-trophy"></i> Rating 排名
        </span>
      </div>

      <!-- 搜索框 -->
      <div class="search-container">
        <el-input
          v-model="keyword"
          placeholder="搜索用户名或昵称"
          prefix-icon="el-icon-search"
          clearable
          @keyup.enter.native="handleSearch"
          style="width: 300px; margin-right: 10px;"
        ></el-input>
        <el-button type="primary" @click="handleSearch" icon="el-icon-search">
          搜索
        </el-button>
      </div>

      <!-- 排名表格 -->
      <vxe-table
        border="inner"
        stripe
        auto-resize
        align="center"
        :data="rankList"
        :loading="loading"
        style="margin-top: 20px;"
      >
        <!-- 排名列 -->
        <vxe-table-column
          field="rank"
          title="排名"
          min-width="80"
          align="center"
        >
          <template v-slot="{ rowIndex }">
            <span :class="getRankClass(rowIndex)">{{ (currentPage - 1) * limit + rowIndex + 1 }}</span>
          </template>
        </vxe-table-column>

        <!-- 用户名列 -->
        <vxe-table-column
          field="username"
          title="用户"
          min-width="200"
          align="left"
        >
          <template v-slot="{ row }">
            <div style="display: flex; align-items: center;">
              <avatar
                :username="row.username"
                :inline="true"
                :size="30"
                color="#FFF"
                :src="row.avatar"
                style="margin-right: 10px;"
              ></avatar>
              <a
                @click="goUserHome(row.username, row.uid)"
                :style="{ color: row.color, fontWeight: 'bold', cursor: 'pointer' }"
              >
                {{ row.username }}
              </a>
              <span v-if="row.nickname" style="margin-left: 8px; color: #999;">
                ({{ row.nickname }})
              </span>
            </div>
          </template>
        </vxe-table-column>

        <!-- Rating 列 -->
        <vxe-table-column
          field="rating"
          title="Rating"
          min-width="120"
          align="center"
        >
          <template v-slot="{ row }">
            <span :style="{ color: row.color, fontWeight: 'bold', fontSize: '16px' }">
              {{ row.rating }}
            </span>
          </template>
        </vxe-table-column>

        <!-- 等级列 -->
        <vxe-table-column
          field="level"
          title="等级"
          min-width="150"
          align="center"
        >
          <template v-slot="{ row }">
            <el-tag
              :color="row.color"
              effect="dark"
              style="color: white; font-weight: bold;"
            >
              {{ row.level }}
            </el-tag>
          </template>
        </vxe-table-column>

        <!-- 学校列 -->
        <vxe-table-column
          field="school"
          title="学校"
          min-width="150"
          align="center"
          show-overflow
        >
          <template v-slot="{ row }">
            <span>{{ row.school || '-' }}</span>
          </template>
        </vxe-table-column>
      </vxe-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          @current-change="handlePageChange"
          :current-page="currentPage"
          :page-size="limit"
          :total="total"
          layout="prev, pager, next, total"
          background
        ></el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import ratingApi from '@/common/rating-api'
import Avatar from 'vue-avatar'

export default {
  name: 'RatingRank',
  components: {
    Avatar
  },
  data() {
    return {
      rankList: [],
      loading: false,
      currentPage: 1,
      limit: 30,
      total: 0,
      keyword: ''
    }
  },
  mounted() {
    this.fetchRankList()
  },
  methods: {
    async fetchRankList() {
      this.loading = true
      try {
        const result = await ratingApi.getRatingRank(this.currentPage, this.limit, this.keyword)
        this.rankList = result.records || []
        this.total = result.total || 0
      } catch (error) {
        console.error('获取 Rating 排名失败:', error)
        this.$message.error('获取排名数据失败')
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.currentPage = 1
      this.fetchRankList()
    },
    handlePageChange(page) {
      this.currentPage = page
      this.fetchRankList()
    },
    getRankClass(rowIndex) {
      const rank = (this.currentPage - 1) * this.limit + rowIndex + 1
      if (rank === 1) return 'rank-tag no1'
      if (rank === 2) return 'rank-tag no2'
      if (rank === 3) return 'rank-tag no3'
      return 'rank-tag'
    },
    goUserHome(username, uid) {
      this.$router.push({
        path: '/user-home',
        query: { uid, username }
      })
    }
  }
}
</script>

<style scoped>
.panel-title {
  font-size: 21px;
  font-weight: 500;
}

.search-container {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

/* 排名标签样式 */
span.rank-tag {
  font: 16px/22px FZZCYSK;
  min-width: 14px;
  height: 22px;
  padding: 0 4px;
  text-align: center;
  color: #fff;
  background: #000;
  background: rgba(0, 0, 0, 0.6);
  display: inline-block;
}

span.rank-tag.no1 {
  background: #bf2c24;
  font-weight: bold;
}

span.rank-tag.no2 {
  background: #e67225;
  font-weight: bold;
}

span.rank-tag.no3 {
  background: #e6bf25;
  font-weight: bold;
}
</style>
