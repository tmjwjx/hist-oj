<template>
  <div class="admin-learning-map-editor" v-loading="loading.full">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">航海图编辑器 - {{ mapInfo.title || `#${mapId}` }}</span>
        <div>
          <el-button size="small" @click="goList">返回列表</el-button>
          <el-button size="small" @click="previewMap">预览用户端</el-button>
          <el-button size="small" type="warning" @click="validateMap">发布校验</el-button>
          <el-button size="small" type="success" @click="publishMap">发布</el-button>
        </div>
      </div>

      <el-form :inline="true" label-width="80px" class="map-meta-form">
        <el-form-item label="标题">
          <el-input v-model="mapInfo.title" style="width: 300px"></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="mapInfo.status" style="width: 140px">
            <el-option label="草稿" value="draft"></el-option>
            <el-option label="已发布" value="published"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" @click="saveMapMeta">保存草稿</el-button>
        </el-form-item>
      </el-form>
      <el-form label-width="80px">
        <el-form-item label="描述">
          <el-input type="textarea" :rows="2" v-model="mapInfo.description"></el-input>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" style="margin-top: 12px;">
      <el-col :md="16" :sm="24">
        <el-card>
          <div slot="header" class="header-row">
            <span>地图画布（拖拽节点调整坐标）</span>
            <div>
              <el-button size="mini" type="primary" @click="openCreateNode('knowledge')">新增知识点</el-button>
              <el-button size="mini" type="warning" @click="openCreateNode('problem')">新增题目节点</el-button>
            </div>
          </div>
          <LearningMapCanvas
            ref="canvas"
            editable
            :nodes="nodes"
            :edges="edges"
            :progress-map="canvasProgressMap"
            :selected-node-id="selectedNode ? selectedNode.id : null"
            @node-click="onNodeClick"
            @node-position-change="persistNodePosition"
          />
        </el-card>
      </el-col>

      <el-col :md="8" :sm="24">
        <el-card>
          <div slot="header" class="header-row">
            <span>节点与连线管理</span>
            <el-button size="mini" icon="el-icon-refresh" @click="loadMap">刷新</el-button>
          </div>

          <div class="panel-section">
            <div class="section-title">节点列表</div>
            <el-scrollbar style="max-height: 280px;">
              <div
                v-for="node in nodes"
                :key="node.id"
                class="node-item"
                :class="selectedNode && selectedNode.id === node.id ? 'active' : ''"
              >
                <div class="node-item-main" @click="focusNode(node)">
                  <span class="node-type" :class="node.type">{{ nodeTypeLabel(node.type) }}</span>
                  <span class="node-title">{{ node.title }}</span>
                </div>
                <div class="node-actions">
                  <el-button size="mini" type="text" @click="openEditNode(node)">编辑</el-button>
                  <el-button size="mini" type="text" style="color:#f56c6c" @click="removeNode(node)">删</el-button>
                </div>
              </div>
            </el-scrollbar>
          </div>

          <div class="panel-section">
            <div class="section-title">新增连线</div>
            <el-form label-width="70px" size="mini">
              <el-form-item label="起点">
                <el-select v-model="edgeForm.sourceNodeId" filterable style="width: 100%;">
                  <el-option v-for="n in nodes" :key="`s-${n.id}`" :label="n.title" :value="n.id"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="终点">
                <el-select v-model="edgeForm.targetNodeId" filterable style="width: 100%;">
                  <el-option v-for="n in nodes" :key="`t-${n.id}`" :label="n.title" :value="n.id"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="类型">
                <el-select v-model="edgeForm.type" style="width: 100%;">
                  <el-option label="前置依赖（强）" value="prerequisite"></el-option>
                  <el-option label="关联（弱）" value="related"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" size="mini" style="width:100%" @click="createEdge">创建连线</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="panel-section">
            <div class="section-title">连线列表</div>
            <el-scrollbar style="max-height: 250px;">
              <div v-for="edge in edges" :key="edge.id" class="edge-item">
                <div class="edge-text">
                  {{ nodeTitleMap[edge.sourceNodeId] || edge.sourceNodeId }}
                  <i class="el-icon-right"></i>
                  {{ nodeTitleMap[edge.targetNodeId] || edge.targetNodeId }}
                  <el-tag size="mini" type="info">{{ edgeTypeLabel(edge.type) }}</el-tag>
                </div>
                <el-button size="mini" type="text" style="color:#f56c6c" @click="removeEdge(edge)">删</el-button>
              </div>
            </el-scrollbar>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog :title="nodeDialogTitle" :visible.sync="nodeDialogVisible" width="96vw" top="2vh" custom-class="learning-node-editor-dialog">
      <el-form label-width="110px" :model="nodeForm">
        <el-row :gutter="12">
          <el-col :md="12" :sm="24">
            <el-form-item label="节点类型">
              <el-select v-model="nodeForm.type" :disabled="isEditingNode" style="width:100%">
                <el-option label="知识点" value="knowledge"></el-option>
                <el-option label="题目" value="problem"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :md="12" :sm="24">
            <el-form-item label="标题">
              <el-input v-model="nodeForm.title"></el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="简介">
          <el-input type="textarea" :rows="2" v-model="nodeForm.description"></el-input>
        </el-form-item>

        <el-row :gutter="12">
          <el-col :md="8" :sm="24">
            <el-form-item label="难度">
              <el-select v-model="nodeForm.difficulty" style="width:100%">
                <el-option label="入门" value="beginner"></el-option>
                <el-option label="简单" value="easy"></el-option>
                <el-option label="中等" value="medium"></el-option>
                <el-option label="困难" value="hard"></el-option>
                <el-option label="专家" value="expert"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item label="标签（逗号分隔）">
              <el-input v-model="nodeForm.tagsCsv" placeholder="例如：图论,最短路"></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item label="发布状态">
              <el-switch v-model="nodeForm.published"></el-switch>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :md="8" :sm="24">
            <el-form-item label="X">
              <el-input-number v-model="nodeForm.x" :step="20" :precision="0" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item label="Y">
              <el-input-number v-model="nodeForm.y" :step="20" :precision="0" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item label="层级">
              <el-input-number v-model="nodeForm.level" :step="1" :precision="0" :min="0" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="区域">
          <el-input v-model="nodeForm.region"></el-input>
        </el-form-item>

        <template v-if="nodeForm.type === 'problem'">
          <el-divider>题目绑定</el-divider>
          <el-row :gutter="10">
            <el-col :md="14" :sm="24">
              <el-input v-model="problemKeyword" placeholder="搜索主 OJ 题号或标题"></el-input>
            </el-col>
            <el-col :md="5" :sm="12">
              <el-button style="width:100%" @click="searchProblems">搜索题目</el-button>
            </el-col>
            <el-col :md="5" :sm="12">
              <el-button type="primary" style="width:100%" @click="verifyProblem">校验并填充</el-button>
            </el-col>
          </el-row>

          <div class="problem-list" v-if="problemCandidates.length">
            <div class="problem-option" v-for="p in problemCandidates" :key="p.id" @click="pickProblem(p)">
              <div>
                <span class="pid">{{ p.problemDisplayId }}</span>
                <span>{{ p.title }}</span>
              </div>
              <el-tag size="mini" type="warning">难度 {{ difficultyLabel(p.difficulty) }}</el-tag>
            </div>
          </div>

          <el-row :gutter="12" style="margin-top: 10px;">
            <el-col :md="12" :sm="24">
              <el-form-item label="主站题目ID">
                <el-input-number v-model="nodeForm.problemId" :min="1" :step="1" style="width:100%"></el-input-number>
              </el-form-item>
            </el-col>
            <el-col :md="12" :sm="24">
              <el-form-item label="展示题号">
                <el-input v-model="nodeForm.problemDisplayId"></el-input>
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <template v-else>
          <el-divider>知识点学习内容（Markdown 格式）</el-divider>
          <div class="editor-tips">支持长篇 Markdown；可通过编辑器工具栏上传并插入 PPT/Word/PDF 附件链接。</div>
          <Editor :value.sync="nodeForm.knowledgeContent" />
        </template>

        <el-form-item label="扩展信息（JSON 格式）">
          <el-input type="textarea" :rows="3" v-model="nodeForm.metadataJson" placeholder='例如：{"estimatedMinutes":30}'></el-input>
        </el-form-item>
      </el-form>

      <div slot="footer">
        <el-button size="small" @click="nodeDialogVisible = false">取消</el-button>
        <el-button size="small" type="primary" @click="saveNode">保存节点</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import learningMapApi from '@/api/learningMap'
import LearningMapCanvas from '@/views/oj/learning-map/components/LearningMapCanvas'
import Editor from '@/components/admin/Editor'

const NODE_TYPE_LABEL = {
  knowledge: '知识点',
  problem: '题目'
}

const EDGE_TYPE_LABEL = {
  prerequisite: '前置依赖',
  related: '关联'
}

const DIFFICULTY_LABEL = {
  beginner: '入门',
  easy: '简单',
  medium: '中等',
  hard: '困难',
  expert: '专家'
}

export default {
  name: 'AdminLearningMapEditor',
  components: {
    LearningMapCanvas,
    Editor
  },
  data() {
    return {
      loading: {
        full: false
      },
      mapInfo: {
        title: '',
        description: '',
        status: 'draft'
      },
      nodes: [],
      edges: [],
      selectedNode: null,
      edgeForm: {
        sourceNodeId: null,
        targetNodeId: null,
        type: 'prerequisite'
      },
      nodeDialogVisible: false,
      isEditingNode: false,
      editingNodeId: null,
      nodeForm: this.getDefaultNodeForm('knowledge'),
      problemKeyword: '',
      problemCandidates: []
    }
  },
  computed: {
    mapId() {
      return this.$route.params.mapId
    },
    nodeTitleMap() {
      const map = {}
      this.nodes.forEach(n => {
        map[n.id] = n.title
      })
      return map
    },
    canvasProgressMap() {
      const map = {}
      this.nodes.forEach(n => {
        map[n.id] = { status: n.published ? 'available' : 'locked' }
      })
      return map
    },
    nodeDialogTitle() {
      return this.isEditingNode ? `编辑节点 #${this.editingNodeId}` : '新建节点'
    }
  },
  mounted() {
    this.loadMap()
  },
  methods: {
    getDefaultNodeForm(type) {
      return {
        type,
        title: '',
        description: '',
        difficulty: 'beginner',
        tagsCsv: '',
        x: 400,
        y: 300,
        level: 1,
        region: '',
        published: true,
        knowledgeContent: '',
        problemId: null,
        problemDisplayId: '',
        metadataJson: '{}'
      }
    },
    parseMetadata(json) {
      if (!json || !json.trim()) return {}
      try {
        return JSON.parse(json)
      } catch (e) {
        throw new Error('扩展信息不是合法 JSON')
      }
    },
    nodeTypeLabel(type) {
      return NODE_TYPE_LABEL[type] || type
    },
    edgeTypeLabel(type) {
      return EDGE_TYPE_LABEL[type] || type
    },
    difficultyLabel(value) {
      if (typeof value === 'number') {
        const d = Number(value)
        if (d <= 0) return DIFFICULTY_LABEL.beginner
        if (d === 1) return DIFFICULTY_LABEL.easy
        if (d === 2) return DIFFICULTY_LABEL.medium
        if (d === 3) return DIFFICULTY_LABEL.hard
        return DIFFICULTY_LABEL.expert
      }
      return DIFFICULTY_LABEL[value] || value || DIFFICULTY_LABEL.beginner
    },
    toNodePayload(form) {
      return {
        type: form.type,
        title: form.title,
        description: form.description,
        difficulty: form.difficulty,
        tags: form.tagsCsv
          .split(',')
          .map(t => t.trim())
          .filter(Boolean),
        x: Number(form.x || 0),
        y: Number(form.y || 0),
        level: Number(form.level || 0),
        region: form.region,
        published: form.published,
        knowledgeContent: form.type === 'knowledge' ? form.knowledgeContent : '',
        problemId: form.type === 'problem' && form.problemId ? Number(form.problemId) : null,
        problemDisplayId: form.type === 'problem' ? (form.problemDisplayId || '').trim() : '',
        metadata: this.parseMetadata(form.metadataJson)
      }
    },
    fillNodeForm(node) {
      this.nodeForm = {
        type: node.type,
        title: node.title || '',
        description: node.description || '',
        difficulty: node.difficulty || 'beginner',
        tagsCsv: (node.tagsList || []).join(','),
        x: Number(node.x || 0),
        y: Number(node.y || 0),
        level: Number(node.level || 0),
        region: node.region || '',
        published: !!node.published,
        knowledgeContent: node.knowledgeContent || '',
        problemId: node.problemId || null,
        problemDisplayId: node.problemDisplayId || '',
        metadataJson: node.metadata || '{}'
      }
    },
    async loadMap() {
      this.loading.full = true
      try {
        const data = await learningMapApi.adminGetMap(this.mapId)
        this.mapInfo = { ...data.map }
        this.nodes = data.nodes || []
        this.edges = data.edges || []
      } catch (e) {
        this.$message.error(e.message || '加载航海图失败')
      } finally {
        this.loading.full = false
      }
    },
    goList() {
      this.$router.push({ name: 'admin-learning-map-list' })
    },
    async saveMapMeta() {
      try {
        await learningMapApi.adminUpdateMap(this.mapId, {
          title: this.mapInfo.title,
          description: this.mapInfo.description,
          status: this.mapInfo.status || 'draft'
        })
        this.$message.success('已保存')
      } catch (e) {
        this.$message.error(e.message || '保存失败')
      }
    },
    previewMap() {
      const url = this.$router.resolve({ name: 'LearningMapPage', params: { mapId: String(this.mapId) } })
      window.open(url.href, '_blank')
    },
    async validateMap() {
      try {
        const res = await learningMapApi.adminValidateMap(this.mapId)
        if (res.valid) {
          this.$message.success('校验通过，可以发布')
        } else {
          const message = (res.errors || []).map(e => `- ${e.message}`).join('\n')
          this.$alert(`<pre style="white-space:pre-wrap">${message}</pre>`, '发布校验失败', {
            dangerouslyUseHTMLString: true,
            type: 'warning'
          })
        }
      } catch (e) {
        this.$message.error(e.message || '校验失败')
      }
    },
    publishMap() {
      this.$confirm('确认发布当前航海图？发布前会执行完整校验。', '发布确认', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminPublishMap(this.mapId)
          this.mapInfo.status = 'published'
          this.$message.success('发布成功')
        } catch (e) {
          this.$message.error(e.message || '发布失败')
        }
      })
    },
    openCreateNode(type) {
      this.isEditingNode = false
      this.editingNodeId = null
      const base = this.getDefaultNodeForm(type)
      base.x = 420 + Math.round(Math.random() * 280)
      base.y = 260 + Math.round(Math.random() * 240)
      this.nodeForm = base
      this.problemCandidates = []
      this.problemKeyword = ''
      this.nodeDialogVisible = true
    },
    openEditNode(node) {
      this.isEditingNode = true
      this.editingNodeId = node.id
      this.fillNodeForm(node)
      this.problemCandidates = []
      this.problemKeyword = ''
      this.nodeDialogVisible = true
    },
    async saveNode() {
      if (!this.nodeForm.title.trim()) {
        this.$message.warning('节点标题不能为空')
        return
      }
      try {
        const payload = this.toNodePayload(this.nodeForm)
        if (this.isEditingNode) {
          await learningMapApi.adminUpdateNode(this.mapId, this.editingNodeId, payload)
          this.$message.success('节点已更新')
        } else {
          await learningMapApi.adminCreateNode(this.mapId, payload)
          this.$message.success('节点已创建')
        }
        this.nodeDialogVisible = false
        await this.loadMap()
      } catch (e) {
        this.$message.error(e.message || '保存节点失败')
      }
    },
    removeNode(node) {
      this.$confirm(`确认删除节点「${node.title}」及关联连线？`, '删除节点', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteNode(this.mapId, node.id)
          if (this.selectedNode && this.selectedNode.id === node.id) {
            this.selectedNode = null
          }
          this.$message.success('节点已删除')
          await this.loadMap()
        } catch (e) {
          this.$message.error(e.message || '删除节点失败')
        }
      })
    },
    onNodeClick(node) {
      this.selectedNode = node
    },
    focusNode(node) {
      this.selectedNode = node
      this.$nextTick(() => {
        if (this.$refs.canvas) {
          this.$refs.canvas.focusNode(node.id)
        }
      })
    },
    async persistNodePosition({ nodeId, x, y }) {
      const node = this.nodes.find(n => n.id === nodeId)
      if (!node) return
      node.x = x
      node.y = y
      try {
        const payload = {
          type: node.type,
          title: node.title,
          description: node.description || '',
          difficulty: node.difficulty || 'beginner',
          tags: node.tagsList || [],
          x,
          y,
          level: node.level || 0,
          region: node.region || '',
          published: !!node.published,
          knowledgeContent: node.knowledgeContent || '',
          problemId: node.problemId || null,
          problemDisplayId: node.problemDisplayId || '',
          metadata: (() => {
            try {
              return node.metadata ? JSON.parse(node.metadata) : {}
            } catch (e) {
              return {}
            }
          })()
        }
        await learningMapApi.adminUpdateNode(this.mapId, nodeId, payload)
      } catch (e) {
        this.$message.error('节点坐标保存失败')
      }
    },
    async createEdge() {
      if (!this.edgeForm.sourceNodeId || !this.edgeForm.targetNodeId) {
        this.$message.warning('请选择起点和终点')
        return
      }
      try {
        await learningMapApi.adminCreateEdge(this.mapId, {
          sourceNodeId: this.edgeForm.sourceNodeId,
          targetNodeId: this.edgeForm.targetNodeId,
          type: this.edgeForm.type
        })
        this.$message.success('连线创建成功')
        this.edgeForm = { sourceNodeId: null, targetNodeId: null, type: 'prerequisite' }
        await this.loadMap()
      } catch (e) {
        this.$message.error(e.message || '创建连线失败')
      }
    },
    removeEdge(edge) {
      this.$confirm('确认删除该连线？', '删除连线', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteEdge(this.mapId, edge.id)
          this.$message.success('连线已删除')
          await this.loadMap()
        } catch (e) {
          this.$message.error(e.message || '删除连线失败')
        }
      })
    },
    async searchProblems() {
      if (!this.problemKeyword.trim()) {
        this.problemCandidates = []
        return
      }
      try {
        this.problemCandidates = await learningMapApi.adminSearchProblems(this.problemKeyword.trim())
      } catch (e) {
        this.problemCandidates = []
        this.$message.error('搜索题目失败')
      }
    },
    pickProblem(problem) {
      this.nodeForm.problemId = problem.id
      this.nodeForm.problemDisplayId = problem.problemDisplayId
      if (!this.nodeForm.title || this.nodeForm.title === '') {
        this.nodeForm.title = problem.title
      }
      if (problem.tags && problem.tags.length) {
        this.nodeForm.tagsCsv = problem.tags.join(',')
      }
      if (typeof problem.difficulty === 'number') {
        const d = Number(problem.difficulty)
        if (d <= 0) this.nodeForm.difficulty = 'beginner'
        else if (d === 1) this.nodeForm.difficulty = 'easy'
        else if (d === 2) this.nodeForm.difficulty = 'medium'
        else if (d === 3) this.nodeForm.difficulty = 'hard'
        else this.nodeForm.difficulty = 'expert'
      }
    },
    async verifyProblem() {
      const id = String(this.nodeForm.problemDisplayId || this.nodeForm.problemId || '').trim()
      if (!id) {
        this.$message.warning('请先填写主站题目ID或展示题号')
        return
      }
      try {
        const problem = await learningMapApi.adminGetProblem(id)
        this.nodeForm.problemId = problem.id
        this.nodeForm.problemDisplayId = problem.problemDisplayId
        if (!this.nodeForm.title) {
          this.nodeForm.title = problem.title
        }
        if (problem.tags && problem.tags.length) {
          this.nodeForm.tagsCsv = problem.tags.join(',')
        }
        this.$message.success('题目校验通过')
      } catch (e) {
        this.$message.error(e.message || '题目不存在或不可用')
      }
    }
  }
}
</script>

<style scoped>
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.map-meta-form {
  margin-bottom: 8px;
}
.panel-section {
  margin-top: 12px;
  border-top: 1px dashed #e5e7eb;
  padding-top: 10px;
}
.section-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #475569;
}
.node-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  border-radius: 6px;
  margin-bottom: 4px;
  border: 1px solid #eef2f7;
}
.node-item.active {
  background: #ecf5ff;
  border-color: #b3d8ff;
}
.node-item-main {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  min-width: 0;
}
.node-type {
  display: inline-block;
  min-width: 74px;
  text-align: center;
  border-radius: 10px;
  font-size: 11px;
  color: #fff;
  padding: 2px 6px;
}
.node-type.knowledge {
  background: #10b981;
}
.node-type.problem {
  background: #f59e0b;
}
.node-title {
  font-size: 12px;
  color: #1f2937;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.edge-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  border: 1px solid #eef2f7;
  border-radius: 6px;
  margin-bottom: 4px;
}
.edge-text {
  font-size: 12px;
  color: #334155;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.problem-list {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-top: 8px;
  max-height: 180px;
  overflow-y: auto;
}
.problem-option {
  padding: 8px 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
}
.problem-option:hover {
  background: #f5f9ff;
}
.pid {
  display: inline-block;
  font-weight: 600;
  color: #2563eb;
  margin-right: 8px;
}
.editor-tips {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: #46658a;
  background: #f2f8ff;
  border: 1px dashed #b9d8ff;
  border-radius: 8px;
  padding: 8px 10px;
}
</style>

<style>
.learning-node-editor-dialog {
  max-height: 92vh;
}
.learning-node-editor-dialog .el-dialog__body {
  max-height: calc(92vh - 120px);
  overflow: auto;
}
.learning-node-editor-dialog .v-note-wrapper {
  min-height: 58vh;
}
</style>
