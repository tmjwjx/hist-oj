<template>
  <div class="problem-set-container">
    <el-card class="header-card">
      <div class="header-content">
        <div class="title-section">
          <i class="fa fa-file-pdf-o" style="color: #67C23A; font-size: 32px;"></i>
          <h2>XCPC 题目集 PDF 生成器</h2>
        </div>
        <el-button type="primary" icon="el-icon-plus" @click="createNewSet">
          创建新题目集
        </el-button>
      </div>
    </el-card>

    <el-card v-loading="loading" class="content-card">
      <el-empty v-if="!loading && problemSets.length === 0" description="暂无题目集，点击上方按钮创建"></el-empty>

      <div v-else class="problem-set-grid">
        <el-card
          v-for="set in problemSets"
          :key="set.id"
          class="problem-set-card"
          shadow="hover"
        >
          <div slot="header" class="card-header">
            <span class="set-title">{{ set.title }}</span>
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, set)">
              <i class="el-icon-more" style="cursor: pointer;"></i>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="pdf">生成 PDF</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>

          <div class="card-body">
            <div class="info-row">
              <i class="el-icon-user"></i>
              <span>{{ set.author || '未知作者' }}</span>
            </div>
            <div class="info-row">
              <i class="el-icon-date"></i>
              <span>{{ formatDate(set.contest_date) }}</span>
            </div>
            <div class="info-row">
              <i class="el-icon-document"></i>
              <span>{{ set.problems?.length || 0 }} 道题目</span>
            </div>
            <div class="info-row">
              <i class="el-icon-time"></i>
              <span>{{ formatDate(set.created_at) }}</span>
            </div>
          </div>

          <div class="card-footer">
            <el-button size="small" @click="editSet(set)">编辑</el-button>
            <el-button size="small" type="success" @click="generatePDF(set)">
              <i class="fa fa-download"></i> PDF
            </el-button>
          </div>
        </el-card>
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog
      title="确认删除"
      :visible.sync="deleteDialogVisible"
      width="400px"
    >
      <p>确定要删除题目集 "{{ currentSet?.title }}" 吗？此操作不可恢复。</p>
      <span slot="footer">
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete">确定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { problemSetApi } from '@/api/problemSet'

export default {
  name: 'ProblemSetList',
  data() {
    return {
      loading: false,
      problemSets: [],
      deleteDialogVisible: false,
      currentSet: null
    }
  },
  mounted() {
    this.loadProblemSets()
  },
  methods: {
    async loadProblemSets() {
      this.loading = true
      try {
        const res = await problemSetApi.getProblemSets()
        if (res.code === 200) {
          this.problemSets = res.data || []
        }
      } catch (error) {
        this.$message.error('加载题目集失败')
      } finally {
        this.loading = false
      }
    },

    createNewSet() {
      this.$router.push('/toolbox/problem-set/create')
    },

    editSet(set) {
      this.$router.push(`/toolbox/problem-set/${set.id}`)
    },

    handleCommand(command, set) {
      if (command === 'edit') {
        this.editSet(set)
      } else if (command === 'pdf') {
        this.generatePDF(set)
      } else if (command === 'delete') {
        this.showDeleteDialog(set)
      }
    },

    showDeleteDialog(set) {
      this.currentSet = set
      this.deleteDialogVisible = true
    },

    async confirmDelete() {
      if (!this.currentSet) return

      try {
        const res = await problemSetApi.deleteProblemSet(this.currentSet.id)
        if (res.code === 200) {
          this.$message.success('删除成功')
          this.loadProblemSets()
        }
      } catch (error) {
        this.$message.error('删除失败')
      } finally {
        this.deleteDialogVisible = false
        this.currentSet = null
      }
    },

    async generatePDF(set) {
      const loading = this.$loading({
        lock: true,
        text: '正在生成 PDF...',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      })

      try {
        const res = await problemSetApi.generatePDF(set.id)

        // 创建下载链接
        const blob = new Blob([res], { type: 'application/pdf' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `${set.title}_${set.id}.pdf`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)

        this.$message.success('PDF 生成成功')
      } catch (error) {
        this.$message.error('生成 PDF 失败')
      } finally {
        loading.close()
      }
    },

    formatDate(dateStr) {
      if (!dateStr) return '未设置'
      const date = new Date(dateStr)
      return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    }
  }
}
</script>

<style scoped>
.problem-set-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.header-card {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 15px;
}

.title-section h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.content-card {
  min-height: 400px;
}

.problem-set-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.problem-set-card {
  border-radius: 8px;
  transition: all 0.3s;
}

.problem-set-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 20px rgba(64, 158, 255, 0.3);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.set-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-body {
  margin: 15px 0;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  color: #606266;
  font-size: 14px;
}

.info-row i {
  color: #909399;
}

.card-footer {
  display: flex;
  gap: 10px;
  border-top: 1px solid #ebeef5;
  padding-top: 15px;
}

.card-footer .el-button {
  flex: 1;
}

@media screen and (max-width: 768px) {
  .problem-set-container {
    padding: 10px;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }

  .problem-set-grid {
    grid-template-columns: 1fr;
  }
}
</style>
