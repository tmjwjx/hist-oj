<template>
  <div class="learning-map-page" v-loading="loading.full">
    <el-card class="toolbar-card">
      <div class="toolbar-top">
        <div class="breadcrumb-row">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>算法地图</el-breadcrumb-item>
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.id">{{ item.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="tool-row">
          <LearningMapSearch
            :results="searchResults"
            :loading="loading.search"
            @search="onSearch"
            @select="onSearchSelect"
          />
          <LearningMapFilter v-model="filters" />
          <el-button size="small" icon="el-icon-s-grid" @click="toggleDetailSize">
            {{ detailExpanded ? '恢复标准面板' : '放大学习面板' }}
          </el-button>
          <el-button
            size="small"
            icon="el-icon-reading"
            :disabled="!selectedNode || selectedNode.type !== 'knowledge'"
            @click="openKnowledgeFullscreen()"
          >全屏学习</el-button>
          <el-button size="small" icon="el-icon-refresh" @click="loadFull">刷新进度</el-button>
        </div>
      </div>

      <div class="summary-row">
        <el-tag size="mini" type="info">总节点 {{ summary.total || 0 }}</el-tag>
        <el-tag size="mini">已解锁 {{ unlockedCount }}</el-tag>
        <el-tag size="mini" type="success">已完成 {{ completedCount }}</el-tag>
        <el-tag size="mini" type="warning">进行中 {{ summary.inProgress || 0 }}</el-tag>
        <el-progress
          :percentage="summary.completionRate || 0"
          :stroke-width="14"
          style="width: 240px;"
        ></el-progress>
        <div class="next-box" v-if="nextRecommended">
          下一步推荐：
          <el-link type="primary" @click="selectNode(nextRecommended)">{{ nextRecommended.title }}</el-link>
        </div>
      </div>
    </el-card>

    <div class="map-layout" :class="{ 'detail-expanded': detailExpanded }">
      <div class="canvas-wrap">
        <LearningMapCanvas
          ref="mapCanvas"
          :nodes="nodes"
          :edges="edges"
          :progress-map="progressMap"
          :selected-node-id="selectedNode ? selectedNode.id : null"
          :filter-status="filters.status"
          :filter-type="filters.type"
          @node-click="selectNode"
        />
      </div>
      <div class="detail-wrap">
        <LearningMapDetailPanel
          ref="detailPanel"
          :node="selectedNode"
          :progress="selectedProgress"
          :node-title-map="nodeTitleMap"
          @start-node="startNode"
          @complete-node="completeNode"
          @go-problem="goProblem"
          @open-knowledge-fullscreen="openKnowledgeFullscreen"
        />
      </div>
    </div>

    <el-dialog
      :visible.sync="knowledgeDialogVisible"
      width="92vw"
      top="4vh"
      custom-class="knowledge-dialog"
      :append-to-body="true"
    >
      <div slot="title" class="knowledge-dialog-title">
        <i class="el-icon-reading"></i>
        <span>{{ selectedNode ? selectedNode.title : '学习内容' }}</span>
      </div>
      <div class="knowledge-dialog-body">
        <div class="knowledge-dialog-actions">
          <el-button
            size="mini"
            type="warning"
            :disabled="!selectedProgress || selectedProgress.status === 'locked'"
            @click="startNode(selectedNode)"
          >标记学习中</el-button>
          <el-button
            size="mini"
            type="success"
            :disabled="!selectedProgress || selectedProgress.status === 'locked'"
            @click="completeNode(selectedNode)"
          >标记已学完</el-button>
        </div>

        <div class="knowledge-resources" v-if="fullscreenResources.length > 0">
          <div class="knowledge-resources-title">资料附件</div>
          <div class="knowledge-resources-list">
            <el-link
              v-for="item in fullscreenResources"
              :key="item.url"
              :underline="false"
              type="primary"
              class="knowledge-resource-item"
              @click.native="openResource(item.url)"
            >
              <i :class="item.icon"></i>
              <span>{{ item.name }}</span>
            </el-link>
          </div>
        </div>

        <div class="knowledge-md-wrap">
          <Markdown :content="selectedKnowledgeContent || '暂无学习内容'" :is-avoid-xss="true" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import learningMapApi from '@/api/learningMap'
import Markdown from '@/components/oj/common/Markdown'
import LearningMapCanvas from './components/LearningMapCanvas'
import LearningMapDetailPanel from './components/LearningMapDetailPanel'
import LearningMapSearch from './components/LearningMapSearch'
import LearningMapFilter from './components/LearningMapFilter'

export default {
  name: 'LearningMapPage',
  components: {
    Markdown,
    LearningMapCanvas,
    LearningMapDetailPanel,
    LearningMapSearch,
    LearningMapFilter
  },
  data() {
    return {
      loading: {
        full: false,
        search: false
      },
      mapInfo: null,
      nodes: [],
      edges: [],
      progressList: [],
      progressMap: {},
      summary: {},
      nextRecommended: null,
      selectedNode: null,
      searchResults: [],
      searchTimer: null,
      detailExpanded: false,
      knowledgeDialogVisible: false,
      filters: {
        status: 'all',
        type: 'all'
      }
    }
  },
  computed: {
    mapId() {
      return this.$route.params.mapId
    },
    selectedProgress() {
      if (!this.selectedNode) {
        return null
      }
      return this.progressMap[this.selectedNode.id] || null
    },
    unlockedCount() {
      const s = this.summary
      return (s.total || 0) - (s.locked || 0)
    },
    completedCount() {
      const s = this.summary
      return (s.completed || 0) + (s.mastered || 0)
    },
    nodeTitleMap() {
      const map = {}
      this.nodes.forEach(n => {
        map[n.id] = n.title
      })
      return map
    },
    selectedKnowledgeContent() {
      if (!this.selectedNode || this.selectedNode.type !== 'knowledge') {
        return ''
      }
      return this.selectedNode.knowledgeContent || ''
    },
    fullscreenResources() {
      return this.extractResources(this.selectedKnowledgeContent)
    },
    breadcrumbs() {
      if (!this.selectedNode) {
        return []
      }
      const incoming = {}
      this.edges
        .filter(e => e.type === 'prerequisite')
        .forEach(e => {
          if (!incoming[e.targetNodeId]) incoming[e.targetNodeId] = []
          incoming[e.targetNodeId].push(e.sourceNodeId)
        })
      const visited = new Set()
      const dfs = (nodeId) => {
        if (visited.has(nodeId)) return []
        visited.add(nodeId)
        const pres = incoming[nodeId] || []
        if (pres.length === 0) {
          return [nodeId]
        }
        for (const pre of pres) {
          const sub = dfs(pre)
          if (sub.length > 0) {
            return [...sub, nodeId]
          }
        }
        return [nodeId]
      }
      const pathIds = dfs(this.selectedNode.id)
      return pathIds
        .map(id => this.nodes.find(n => n.id === id))
        .filter(Boolean)
    }
  },
  mounted() {
    this.loadFull()
  },
  methods: {
    async loadFull() {
      this.loading.full = true
      try {
        const data = await learningMapApi.getMapFull(this.mapId)
        this.mapInfo = data.map
        this.nodes = data.nodes || []
        this.edges = data.edges || []
        this.progressList = data.progress || []
        this.summary = data.summary || {}
        this.nextRecommended = data.nextRecommended || null

        const map = {}
        this.progressList.forEach(p => {
          map[p.nodeId] = p
        })
        this.progressMap = map

        if (!this.selectedNode && this.nodes.length > 0) {
          const first = this.nextRecommended || this.nodes[0]
          this.selectNode(first)
          this.$nextTick(() => {
            if (this.$refs.mapCanvas) {
              this.$refs.mapCanvas.focusNode(first.id, false)
            }
          })
        } else if (this.selectedNode) {
          const latest = this.nodes.find(n => n.id === this.selectedNode.id)
          this.selectedNode = latest || null
        }
      } catch (e) {
        this.$message.error(e.message || '加载航海图失败')
      } finally {
        this.loading.full = false
      }
    },
    selectNode(node) {
      const target = this.nodes.find(n => n.id === node.id)
      if (!target) return
      this.selectedNode = target
      this.$nextTick(() => {
        if (this.$refs.mapCanvas) {
          this.$refs.mapCanvas.focusNode(target.id)
        }
      })
    },
    onSearch(keyword) {
      clearTimeout(this.searchTimer)
      if (!keyword) {
        this.searchResults = []
        return
      }
      this.searchTimer = setTimeout(() => {
        this.search(keyword)
      }, 180)
    },
    async search(keyword) {
      this.loading.search = true
      try {
        this.searchResults = await learningMapApi.searchNodes(this.mapId, keyword)
      } catch (e) {
        this.searchResults = []
      } finally {
        this.loading.search = false
      }
    },
    onSearchSelect(item) {
      const node = this.nodes.find(n => n.id === item.id)
      if (!node) return
      this.selectedNode = node
      this.$nextTick(() => {
        if (this.$refs.mapCanvas) {
          this.$refs.mapCanvas.focusNode(node.id)
        }
      })
    },
    toggleDetailSize() {
      this.detailExpanded = !this.detailExpanded
    },
    openKnowledgeFullscreen(node) {
      if (node && node.id) {
        const target = this.nodes.find(n => n.id === node.id)
        this.selectedNode = target || node
      }
      if (!this.selectedNode || this.selectedNode.type !== 'knowledge') {
        return
      }
      this.knowledgeDialogVisible = true
    },
    extractResources(content) {
      if (!content) return []
      const reg = /\[([^\]]+)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g
      const list = []
      const seen = new Set()
      let match
      while ((match = reg.exec(content)) !== null) {
        const name = (match[1] || '').trim()
        const url = (match[2] || '').trim()
        if (!url || seen.has(url)) continue
        const ext = this.getExt(url)
        if (!['ppt', 'pptx', 'doc', 'docx', 'pdf', 'xls', 'xlsx', 'zip', 'rar'].includes(ext)) {
          continue
        }
        seen.add(url)
        list.push({
          name: name || this.filenameFromUrl(url),
          url,
          ext,
          icon: this.extIcon(ext)
        })
      }
      return list
    },
    getExt(url) {
      const pure = (url.split('?')[0] || '').toLowerCase()
      const idx = pure.lastIndexOf('.')
      if (idx < 0) return ''
      return pure.slice(idx + 1)
    },
    filenameFromUrl(url) {
      const pure = url.split('?')[0] || url
      const parts = pure.split('/')
      return parts[parts.length - 1] || url
    },
    extIcon(ext) {
      if (ext === 'pdf') return 'el-icon-document'
      if (ext === 'ppt' || ext === 'pptx') return 'el-icon-document'
      if (ext === 'doc' || ext === 'docx') return 'el-icon-document'
      if (ext === 'xls' || ext === 'xlsx') return 'el-icon-document'
      return 'el-icon-folder-opened'
    },
    openResource(url) {
      window.open(url, '_blank')
    },
    async startNode(node) {
      if (!node || !node.id) return
      try {
        await learningMapApi.startNode(this.mapId, node.id)
        this.$message.success('已标记为学习中')
        await this.loadFull()
      } catch (e) {
        this.$message.error(e.message || '操作失败')
      }
    },
    async completeNode(node) {
      if (!node || !node.id) return
      try {
        await learningMapApi.completeNode(this.mapId, node.id)
        this.$message.success('已标记完成')
        await this.loadFull()
      } catch (e) {
        this.$message.error(e.message || '操作失败')
      }
    },
    goProblem(node) {
      const displayId = node.problemDisplayId || (node.problemInfo && node.problemInfo.problemDisplayId)
      if (!displayId) {
        this.$message.warning('该题目节点未绑定题号')
        return
      }
      this.$router.push({ name: 'ProblemDetails', params: { problemID: displayId } })
    }
  }
}
</script>

<style scoped>
.learning-map-page {
  max-width: 1660px;
  margin: 0 auto;
  position: relative;
  padding: 8px 8px 14px;
  font-family: 'Trebuchet MS', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
.toolbar-card {
  margin-bottom: 12px;
  border: 0;
  border-radius: 16px;
  background: linear-gradient(135deg, #f4fbff 0%, #eef8ff 52%, #fff8e8 100%);
  box-shadow: 0 10px 22px rgba(58, 113, 173, 0.12);
}
.toolbar-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.tool-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
.summary-row {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
.next-box {
  margin-left: auto;
  color: #0f3d66;
  font-size: 13px;
  font-weight: 600;
}
.map-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 12px;
}
.map-layout.detail-expanded {
  grid-template-columns: minmax(0, 1fr) 620px;
}
.canvas-wrap {
  min-width: 0;
}
.detail-wrap {
  border: 1px solid #d8e5f6;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(180deg, #ffffff 0%, #f7fbff 100%);
  box-shadow: 0 8px 20px rgba(45, 108, 176, 0.12);
  min-height: 520px;
}
@media (max-width: 1200px) {
  .map-layout {
    grid-template-columns: 1fr;
  }
  .detail-wrap {
    min-height: 360px;
  }
  .next-box {
    margin-left: 0;
    width: 100%;
  }
}
</style>

<style>
.knowledge-dialog {
  border-radius: 16px;
  overflow: hidden;
}
.knowledge-dialog .el-dialog__body {
  padding-top: 10px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}
.knowledge-dialog-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: #1f3b60;
}
.knowledge-dialog-body {
  max-height: 82vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.knowledge-dialog-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.knowledge-resources {
  border: 1px dashed #a7caf8;
  border-radius: 10px;
  padding: 10px;
  background: #f5f9ff;
}
.knowledge-resources-title {
  font-size: 13px;
  font-weight: 600;
  color: #1d4c84;
  margin-bottom: 6px;
}
.knowledge-resources-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.knowledge-resource-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 999px;
  padding: 5px 10px;
  background: #e8f3ff;
}
.knowledge-md-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border-radius: 12px;
  border: 1px solid #dbe8f8;
  padding: 10px 12px;
  background: #fff;
}
</style>
