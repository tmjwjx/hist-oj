<template>
  <div class="learning-map-page" v-loading="loading.full">
    <el-card class="toolbar-card">
      <div class="toolbar-top">
        <div class="breadcrumb-row">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>{{ $t('m.NavBar_Learning_Map') }}</el-breadcrumb-item>
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
            {{ detailExpanded ? $t('m.Map_Restore_Panel') : $t('m.Map_Expand_Panel') }}
          </el-button>
          <el-button
            size="small"
            icon="el-icon-reading"
            :disabled="!selectedNode || selectedNode.type !== 'knowledge'"
            @click="openKnowledgeFullscreen()"
          >{{ $t('m.Map_Fullscreen') }}</el-button>
          <el-button size="small" icon="el-icon-refresh" @click="loadFull">{{ $t('m.Map_Refresh_Progress') }}</el-button>
        </div>
      </div>

      <div class="summary-row">
        <el-tag size="mini" type="info">{{ $t('m.Map_Total_Nodes') }} {{ summary.total || 0 }}</el-tag>
        <el-tag size="mini">{{ $t('m.Map_Unlocked') }} {{ unlockedCount }}</el-tag>
        <el-tag size="mini" type="success">{{ $t('m.Map_Completed') }} {{ completedCount }}</el-tag>
        <el-tag size="mini" type="warning">{{ $t('m.Map_In_Progress') }} {{ summary.inProgress || 0 }}</el-tag>
        <el-progress
          :percentage="summary.completionRate || 0"
          :stroke-width="14"
          style="width: 240px;"
        ></el-progress>
        <div class="next-box" v-if="nextRecommended">
          {{ $t('m.Map_Next_Recommend') }}
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
        <span>{{ selectedNode ? selectedNode.title : $t('m.Map_Learning_Content') }}</span>
      </div>
      <div class="knowledge-dialog-body">
        <div class="knowledge-dialog-actions">
          <el-button
            size="mini"
            type="warning"
            :disabled="!selectedProgress || selectedProgress.status === 'locked'"
            @click="startNode(selectedNode)"
          >{{ $t('m.Map_Mark_Learning') }}</el-button>
          <el-button
            size="mini"
            type="success"
            :disabled="!selectedProgress || selectedProgress.status === 'locked'"
            @click="completeNode(selectedNode)"
          >{{ $t('m.Map_Mark_Completed') }}</el-button>
        </div>

        <div class="knowledge-resources" v-if="fullscreenResources.length > 0">
          <div class="knowledge-resources-title">{{ $t('m.Map_Resources') }}</div>
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
          <Markdown :content="selectedKnowledgeContent || $t('m.Map_No_Content')" :is-avoid-xss="true" />
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
        this.$message.error(e.message || this.$t('m.Map_Load_Failed'))
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
        this.$message.success(this.$t('m.Map_Marked_Learning'))
        await this.loadFull()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.Battle_Operate_Failed'))
      }
    },
    async completeNode(node) {
      if (!node || !node.id) return
      try {
        await learningMapApi.completeNode(this.mapId, node.id)
        this.$message.success(this.$t('m.Map_Marked_Completed'))
        await this.loadFull()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.Battle_Operate_Failed'))
      }
    },
    goProblem(node) {
      const displayId = node.problemDisplayId || (node.problemInfo && node.problemInfo.problemDisplayId)
      if (!displayId) {
        this.$message.warning(this.$t('m.Map_Node_No_Problem'))
        return
      }
      this.$router.push({ name: 'ProblemDetails', params: { problemID: displayId } })
    }
  }
}
</script>

<style scoped>
.learning-map-page {
  max-width: 1600px;
  margin: 0 auto;
  position: relative;
  padding: 20px;
  min-height: calc(100vh - 60px);
}
.toolbar-card {
  margin-bottom: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
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
  margin-top: 12px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
}
.next-box {
  margin-left: auto;
  color: #606266;
  font-size: 13px;
  font-weight: 500;
}
.map-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
}
.map-layout.detail-expanded {
  grid-template-columns: minmax(0, 1fr) 620px;
}
.canvas-wrap {
  min-width: 0;
}
.detail-wrap {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
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
  border-radius: 4px;
  overflow: hidden;
}
.knowledge-dialog .el-dialog__body {
  padding-top: 12px;
  background: #fff;
}
.knowledge-dialog-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
}
.knowledge-dialog-body {
  max-height: 82vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.knowledge-dialog-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.knowledge-resources {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 12px;
  background: #f5f7fa;
}
.knowledge-resources-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
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
  border-radius: 4px;
  padding: 5px 10px;
  background: #ecf5ff;
  border: 1px solid #d9ecff;
}
.knowledge-md-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  padding: 12px;
  background: #fff;
}
</style>
