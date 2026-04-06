<template>
  <div class="practice-list-page">
    <el-card shadow="never" class="search-card">
      <div class="search-bar">
        <el-input
          v-model="keyword"
          placeholder="搜索公开试卷"
          clearable
          @clear="handleSearch"
          @keyup.enter.native="handleSearch"
        >
          <el-button slot="append" icon="el-icon-search" @click="handleSearch"></el-button>
        </el-input>
      </div>
    </el-card>

    <el-card shadow="never" class="list-card">
      <div v-loading="loading">
        <el-empty v-if="papers.length === 0" description="暂无公开练习试卷"></el-empty>
        <div v-else class="paper-grid">
          <el-card v-for="paper in papers" :key="paper.id" shadow="hover" class="paper-item">
            <div class="paper-header">
              <span class="paper-title">{{ paper.title }}</span>
              <el-tag type="danger" size="mini">公开练习</el-tag>
            </div>
            <p class="paper-desc">{{ paper.description || '暂无简介' }}</p>
            <div class="paper-meta">
              <span>题目: {{ paper.questionCount || 0 }}</span>
              <span>总分: {{ paper.totalScore || 0 }}</span>
              <span>作者: {{ paper.creator ? paper.creator.username : '-' }}</span>
            </div>
            <div class="paper-actions">
              <el-button type="primary" size="small" @click="goDetail(paper.id)">进入练习</el-button>
            </div>
          </el-card>
        </div>
      </div>

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="total, prev, pager, next, jumper"
          :total="pagination.total"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          @current-change="handlePageChange"
        ></el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import classroomApi from '@/api/classroom'

export default {
  name: 'PracticeList',
  data() {
    return {
      loading: false,
      keyword: '',
      papers: [],
      pagination: {
        page: 1,
        limit: 12,
        total: 0
      }
    }
  },
  created() {
    this.loadPapers()
  },
  methods: {
    async loadPapers() {
      this.loading = true
      try {
        const res = await classroomApi.getPublicExamPaperList({
          page: this.pagination.page,
          limit: this.pagination.limit,
          keyword: this.keyword || undefined
        })
        if (res.data && res.data.code === 200) {
          const data = res.data.data || {}
          this.papers = data.papers || []
          this.pagination.total = data.total || 0
        } else {
          this.$message.error((res.data && res.data.message) || '加载练习列表失败')
        }
      } catch (error) {
        this.$message.error('加载练习列表失败')
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.pagination.page = 1
      this.loadPapers()
    },
    handlePageChange(page) {
      this.pagination.page = page
      this.loadPapers()
    },
    goDetail(paperId) {
      this.$router.push({ name: 'PracticeDetail', params: { paperId: String(paperId) } })
    }
  }
}
</script>

<style scoped>
.practice-list-page {
  max-width: 1200px;
  margin: 20px auto;
}

.search-card {
  margin-bottom: 16px;
}

.search-bar {
  max-width: 420px;
}

.list-card {
  min-height: 360px;
}

.paper-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.paper-item {
  border-radius: 10px;
}

.paper-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.paper-title {
  font-weight: 600;
  color: #303133;
}

.paper-desc {
  min-height: 42px;
  color: #606266;
  line-height: 1.6;
}

.paper-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #909399;
  font-size: 13px;
}

.paper-actions {
  margin-top: 14px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
