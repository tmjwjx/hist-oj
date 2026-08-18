<template>
  <div class="practice-list-page">
    <el-card shadow="never" class="search-card">
      <div class="search-bar">
        <el-input
          v-model="keyword"
          placeholder="搜索套卷"
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
        <el-empty v-if="papers.length === 0" description="暂无可练习套卷"></el-empty>
        <div v-else class="paper-list">
          <article
            v-for="paper in papers"
            :key="paper.id"
            class="paper-row answer-info compact-answer-info"
          >
            <div class="paper-main" @click="goDetail(paper.id)">
              <div class="paper-topline">
                <span class="paper-origin">套卷练习 · 试卷</span>
                <span class="paper-author">作者 {{ paper.creator ? paper.creator.username : '-' }}</span>
              </div>
              <h3 class="paper-title">{{ paper.title }}</h3>
              <p class="paper-desc">{{ paper.description || '暂无简介' }}</p>
              <div class="paper-meta">
                <span>题目 {{ paper.questionCount || 0 }}</span>
                <span class="meta-dot">·</span>
                <span>总分 {{ paper.totalScore || 0 }}</span>
                <span class="meta-dot">·</span>
                <span>ID {{ paper.id }}</span>
              </div>
            </div>
            <div class="paper-actions">
              <el-button type="primary" size="small" @click="goDetail(paper.id)">进入套卷</el-button>
            </div>
          </article>
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
          this.$message.error((res.data && res.data.message) || '加载套卷练习列表失败')
        }
      } catch (error) {
        this.$message.error('加载套卷练习列表失败')
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
  margin: 0 auto;
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.search-card {
  margin-bottom: 20px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.search-bar {
  max-width: 100%;
}

.list-card {
  min-height: 400px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.paper-list {
  display: flex;
  flex-direction: column;
}

.paper-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
  transition: background-color 0.2s ease;
}

.paper-row:last-child {
  border-bottom: none;
}

.paper-row:hover {
  background: #f5f7fa;
}

.paper-main {
  flex: 1;
  min-width: 0;
  cursor: pointer;
}

.paper-topline {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 12px;
  color: #909399;
}

.paper-origin {
  color: #67c23a;
  font-weight: 500;
}

.paper-author {
  color: #909399;
}

.paper-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  line-height: 1.4;
  font-weight: 600;
  color: #303133;
}

.paper-title:hover {
  color: #409eff;
}

.paper-desc {
  margin: 0 0 8px 0;
  color: #606266;
  font-size: 13px;
  line-height: 1.6;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.paper-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 12px;
}

.meta-dot {
  color: #c0c4cc;
}

.paper-actions {
  padding-top: 8px;
}

.pagination-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .practice-list-page {
    padding: 10px;
  }

  .paper-row {
    flex-direction: column;
    gap: 12px;
    padding: 12px;
  }

  .paper-title {
    font-size: 15px;
  }

  .paper-actions {
    padding-top: 0;
    width: 100%;
  }

  .paper-actions .el-button {
    width: 100%;
  }
}
</style>
