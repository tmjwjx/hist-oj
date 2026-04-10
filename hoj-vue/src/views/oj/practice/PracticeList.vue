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
        <div v-else class="paper-list">
          <article
            v-for="paper in papers"
            :key="paper.id"
            class="paper-row answer-info compact-answer-info"
          >
            <div class="paper-main" @click="goDetail(paper.id)">
              <div class="paper-topline">
                <span class="paper-origin">公开练习 · 试卷</span>
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
              <el-button type="primary" size="small" @click="goDetail(paper.id)">进入练习</el-button>
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
  max-width: 1080px;
  margin: 20px auto;
}

.search-card {
  margin-bottom: 16px;
  border-radius: 10px;
}

.search-bar {
  max-width: 560px;
}

.list-card {
  min-height: 360px;
  border-radius: 10px;
}

.paper-list {
  display: flex;
  flex-direction: column;
}

.paper-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding: 16px 8px;
  border-bottom: 1px solid #eceff3;
  transition: background-color 0.2s ease;
}

.paper-row:last-child {
  border-bottom: none;
}

.paper-row:hover {
  background: #f8fbff;
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
  margin-bottom: 6px;
  font-size: 12px;
  color: #5f6368;
}

.paper-origin {
  color: #188038;
}

.paper-author {
  color: #5f6368;
}

.paper-title {
  margin: 0;
  font-size: 20px;
  line-height: 1.3;
  font-weight: 600;
  color: #1a73e8;
}

.paper-desc {
  margin: 8px 0 10px;
  color: #3c4043;
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
  color: #5f6368;
  font-size: 13px;
}

.meta-dot {
  color: #9aa0a6;
}

.paper-actions {
  padding-top: 18px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 900px) {
  .paper-row {
    flex-direction: column;
    gap: 10px;
  }

  .paper-title {
    font-size: 18px;
  }

  .paper-actions {
    padding-top: 0;
  }
}
</style>
