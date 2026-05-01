<template>
  <div class="admin-learning-map-editor" v-loading="loading.full">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">航海图编辑器 - {{ mapInfo.title || `#${mapId}` }}</span>
        <div>
          <el-button size="small" @click="goList">返回列表</el-button>
          <el-button
            size="small"
            :type="mapInfo.status === 'published' ? 'warning' : 'success'"
            @click="toggleMapStatus"
          >
            {{ mapInfo.status === 'published' ? '隐藏' : '发布' }}
          </el-button>
        </div>
      </div>

      <el-form label-width="80px" class="map-meta-form">
        <el-form-item label="标题">
          <el-input
            v-model="mapInfo.title"
            maxlength="120"
            show-word-limit
            placeholder="请输入航海图标题"
            @blur="syncMapMeta"
          ></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            type="textarea"
            :rows="2"
            v-model="mapInfo.description"
            placeholder="请输入航海图描述"
            @blur="syncMapMeta"
          ></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-tag size="mini" :type="mapInfo.status === 'published' ? 'success' : 'info'">
            {{ mapInfo.status === 'published' ? '已发布' : '已隐藏' }}
          </el-tag>
          <span class="meta-saving" v-if="savingMeta">正在保存...</span>
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

          <div class="panel-section">
            <div class="section-title">使用权限控制</div>
            <div class="permission-mode-row">
              <el-tag size="mini" :type="permissionMode === 'all_open' ? 'success' : 'warning'">
                {{ accessModeLabel(permissionMode) }}
              </el-tag>
              <div class="permission-mode-actions">
                <el-button size="mini" type="success" @click="setAccessMode('all_open')">全部开启</el-button>
                <el-button size="mini" type="warning" @click="setAccessMode('all_closed')">全部关闭</el-button>
              </div>
            </div>

            <el-input
              v-model="permissionKeyword"
              size="mini"
              clearable
              placeholder="搜索用户：uid/用户名/昵称"
              @keyup.enter.native="searchPermissionUsers"
            >
              <el-button slot="append" icon="el-icon-search" @click="searchPermissionUsers"></el-button>
            </el-input>

            <div class="permission-candidates" v-if="permissionCandidates.length">
              <div class="permission-candidate-item" v-for="u in permissionCandidates" :key="u.userId">
                <div class="permission-user-main">
                  <div class="permission-user-name">{{ userDisplayName(u) }}</div>
                  <div class="permission-user-id">{{ u.userId }}</div>
                </div>
                <div class="permission-actions">
                  <el-button size="mini" type="success" @click="setUserPermission(u, true)">开通</el-button>
                  <el-button size="mini" type="danger" plain @click="setUserPermission(u, false)">关闭</el-button>
                </div>
              </div>
            </div>

            <el-scrollbar style="max-height: 200px; margin-top: 8px;">
              <div v-if="mapPermissions.length === 0" class="permission-empty">暂无单独配置用户</div>
              <div class="permission-item" v-for="item in mapPermissions" :key="item.userId">
                <div class="permission-user-main">
                  <div class="permission-user-name">{{ userDisplayName(item) }}</div>
                  <div class="permission-user-id">{{ item.userId }}</div>
                </div>
                <div class="permission-actions">
                  <el-switch
                    :value="item.enabled"
                    @change="toggleUserPermission(item, $event)"
                    active-color="#13ce66"
                    inactive-color="#ff4949"
                  ></el-switch>
                  <el-button size="mini" type="text" style="color:#909399" @click="removeUserPermission(item)">移除</el-button>
                </div>
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
        <div class="field-hint">层级：用于推荐学习顺序排序（数字越小越靠前），不直接参与解锁判断。</div>

        <el-form-item label="区域">
          <el-input v-model="nodeForm.region"></el-input>
        </el-form-item>
        <div class="field-hint">区域：用于给节点做分区归类，例如「数据结构区」「字符串区」，便于筛选和可视化布局。</div>

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
        <div class="field-hint">扩展信息：用于存储个性化字段（如预计学习时长、视频链接、讲义地址等），不影响基础解锁逻辑。</div>
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
        status: 'draft',
        accessMode: 'all_open'
      },
      savingMeta: false,
      lastSavedMeta: '',
      nodes: [],
      edges: [],
      permissionMode: 'all_open',
      mapPermissions: [],
      permissionKeyword: '',
      permissionCandidates: [],
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
        this.lastSavedMeta = JSON.stringify({
          title: this.mapInfo.title || '',
          description: this.mapInfo.description || '',
          status: this.mapInfo.status || 'draft',
          accessMode: this.mapInfo.accessMode || 'all_open'
        })
        this.permissionMode = this.mapInfo.accessMode || 'all_open'
        this.nodes = data.nodes || []
        this.edges = data.edges || []
        await this.loadPermissions()
      } catch (e) {
        this.$message.error(e.message || '加载航海图失败')
      } finally {
        this.loading.full = false
      }
    },
    goList() {
      this.$router.push({ name: 'admin-learning-map-list' })
    },
    async syncMapMeta(options = {}) {
      const { force = false, silent = true } = options
      const title = (this.mapInfo.title || '').trim()
      if (!title) {
        if (!silent) {
          this.$message.warning('标题不能为空')
        }
        return false
      }
      const payload = {
        title,
        description: (this.mapInfo.description || '').trim(),
        status: this.mapInfo.status || 'draft',
        accessMode: this.permissionMode || 'all_open'
      }
      const snapshot = JSON.stringify(payload)
      if (!force && snapshot === this.lastSavedMeta) {
        return true
      }
      if (this.savingMeta) {
        return false
      }
      this.savingMeta = true
      try {
        const updated = await learningMapApi.adminUpdateMap(this.mapId, payload)
        this.mapInfo = { ...this.mapInfo, ...updated }
        this.lastSavedMeta = JSON.stringify({
          title: this.mapInfo.title || '',
          description: this.mapInfo.description || '',
          status: this.mapInfo.status || 'draft',
          accessMode: this.mapInfo.accessMode || 'all_open'
        })
        return true
      } catch (e) {
        if (!silent) {
          this.$message.error(e.message || '保存失败')
        }
        return false
      } finally {
        this.savingMeta = false
      }
    },
    toggleMapStatus() {
      if (this.mapInfo.status === 'published') {
        this.$confirm('确认将当前航海图设置为隐藏？隐藏后普通用户将无法访问。', '隐藏确认', {
          type: 'warning'
        }).then(async () => {
          try {
            await learningMapApi.adminUpdateMap(this.mapId, {
              title: this.mapInfo.title,
              description: this.mapInfo.description || '',
              status: 'draft',
              accessMode: this.permissionMode || 'all_open'
            })
            this.mapInfo.status = 'draft'
            this.$message.success('已隐藏')
          } catch (e) {
            this.$message.error(e.message || '隐藏失败')
          }
        })
        return
      }

      this.$confirm('确认发布当前航海图？发布前会执行完整校验。', '发布确认', {
        type: 'warning'
      }).then(async () => {
        try {
          const ok = await this.syncMapMeta({ force: true, silent: false })
          if (!ok) {
            return
          }
          await learningMapApi.adminPublishMap(this.mapId)
          this.mapInfo.status = 'published'
          this.lastSavedMeta = JSON.stringify({
            title: this.mapInfo.title || '',
            description: this.mapInfo.description || '',
            status: this.mapInfo.status || 'published',
            accessMode: this.mapInfo.accessMode || 'all_open'
          })
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
    },
    accessModeLabel(mode) {
      return mode === 'all_closed' ? '默认：全部关闭，仅授权用户可见' : '默认：全部开启，按用户可单独关闭'
    },
    userDisplayName(user) {
      return user.nickname || user.realname || user.username || user.userId
    },
    async loadPermissions() {
      const cfg = await learningMapApi.adminGetMapPermissions(this.mapId)
      this.permissionMode = cfg.accessMode || 'all_open'
      this.mapPermissions = cfg.permissions || []
    },
    async setAccessMode(mode) {
      try {
        await learningMapApi.adminSetMapAccessMode(this.mapId, mode)
        this.permissionMode = mode
        this.mapInfo.accessMode = mode
        this.$message.success(mode === 'all_open' ? '已设置为全部开启' : '已设置为全部关闭')
      } catch (e) {
        this.$message.error(e.message || '设置失败')
      }
    },
    async searchPermissionUsers() {
      const keyword = (this.permissionKeyword || '').trim()
      if (!keyword) {
        this.permissionCandidates = []
        return
      }
      try {
        this.permissionCandidates = await learningMapApi.adminSearchMapPermissionUsers(keyword)
      } catch (e) {
        this.permissionCandidates = []
        this.$message.error(e.message || '搜索用户失败')
      }
    },
    async setUserPermission(user, enabled) {
      if (!user || !user.userId) return
      try {
        await learningMapApi.adminSetMapUserPermission(this.mapId, user.userId, enabled)
        this.$message.success(enabled ? '已开通权限' : '已关闭权限')
        await this.loadPermissions()
      } catch (e) {
        this.$message.error(e.message || '设置用户权限失败')
      }
    },
    async toggleUserPermission(item, enabled) {
      const original = item.enabled
      item.enabled = enabled
      try {
        await learningMapApi.adminSetMapUserPermission(this.mapId, item.userId, enabled)
        this.$message.success('权限已更新')
      } catch (e) {
        item.enabled = original
        this.$message.error(e.message || '更新权限失败')
      }
    },
    removeUserPermission(item) {
      this.$confirm(`确认移除用户 ${item.userId} 的单独权限配置？`, '移除权限', {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteMapUserPermission(this.mapId, item.userId)
          this.$message.success('已移除')
          await this.loadPermissions()
        } catch (e) {
          this.$message.error(e.message || '移除失败')
        }
      })
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
  margin-top: 4px;
}
.map-meta-form >>> .el-form-item {
  margin-bottom: 10px;
}
.meta-saving {
  margin-left: 8px;
  font-size: 12px;
  color: #64748b;
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
.permission-mode-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.permission-mode-actions {
  display: flex;
  gap: 6px;
}
.permission-candidates {
  margin-top: 8px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}
.permission-candidate-item,
.permission-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  border-bottom: 1px solid #eef2f7;
}
.permission-candidate-item:last-child,
.permission-item:last-child {
  border-bottom: 0;
}
.permission-user-main {
  min-width: 0;
}
.permission-user-name {
  font-size: 12px;
  color: #1f2937;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.permission-user-id {
  font-size: 11px;
  color: #64748b;
}
.permission-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.permission-empty {
  font-size: 12px;
  color: #94a3b8;
  padding: 10px 6px;
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
.field-hint {
  margin: -6px 0 8px 112px;
  font-size: 12px;
  color: #64748b;
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
