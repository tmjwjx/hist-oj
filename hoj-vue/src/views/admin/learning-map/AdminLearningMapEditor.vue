<template>
  <div class="admin-learning-map-editor" v-loading="loading.full">
    <el-card>
      <div slot="header" class="header-row">
        <span class="panel-title home-title">{{ $t('m.MapAdm_Editor') }} - {{ mapInfo.title || `#${mapId}` }}</span>
        <div>
          <el-button size="small" @click="goList">{{ $t('m.MapAdm_Back_To_List') }}</el-button>
          <el-button
            size="small"
            :type="mapInfo.status === 'published' ? 'warning' : 'success'"
            @click="toggleMapStatus"
          >
            {{ mapInfo.status === 'published' ? $t('m.MapAdm_Hide') : $t('m.MapAdm_Publish') }}
          </el-button>
        </div>
      </div>

      <el-form label-width="80px" class="map-meta-form">
        <el-form-item :label="$t('m.MapAdm_Title_Field')">
          <el-input
            v-model="mapInfo.title"
            maxlength="120"
            show-word-limit
            :placeholder="$t('m.MapAdm_Title_Input_Placeholder')"
            @blur="syncMapMeta"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.MapAdm_Description')">
          <el-input
            type="textarea"
            :rows="2"
            v-model="mapInfo.description"
            :placeholder="$t('m.MapAdm_Description_Input_Placeholder')"
            @blur="syncMapMeta"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.MapAdm_Status')">
          <el-tag size="mini" :type="mapInfo.status === 'published' ? 'success' : 'info'">
            {{ mapInfo.status === 'published' ? $t('m.MapAdm_Published') : $t('m.MapAdm_Hidden') }}
          </el-tag>
          <span class="meta-saving" v-if="savingMeta">{{ $t('m.MapAdm_Saving') }}</span>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" style="margin-top: 12px;">
      <el-col :md="16" :sm="24">
        <el-card>
          <div slot="header" class="header-row">
            <span>{{ $t('m.MapAdm_Map_Canvas') }}</span>
            <div>
              <el-button size="mini" type="primary" @click="openCreateNode('knowledge')">{{ $t('m.MapAdm_Add_Knowledge') }}</el-button>
              <el-button size="mini" type="warning" @click="openCreateNode('problem')">{{ $t('m.MapAdm_Add_Problem_Node') }}</el-button>
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
            <span>{{ $t('m.MapAdm_Node_Edge_Management') }}</span>
            <el-button size="mini" icon="el-icon-refresh" @click="loadMap">{{ $t('m.MapAdm_Refresh') }}</el-button>
          </div>

          <div class="panel-section">
            <div class="section-title">{{ $t('m.MapAdm_Node_List') }}</div>
            <el-scrollbar class="admin-list-scroll node-list-scroll">
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
                  <el-button size="mini" type="text" @click="openEditNode(node)">{{ $t('m.MapAdm_Edit') }}</el-button>
                  <el-button size="mini" type="text" style="color:#f56c6c" @click="removeNode(node)">{{ $t('m.MapAdm_Delete_Short') }}</el-button>
                </div>
              </div>
            </el-scrollbar>
          </div>

          <div class="panel-section">
            <div class="section-title">{{ $t('m.MapAdm_Add_Edge') }}</div>
            <el-form label-width="70px" size="mini">
              <el-form-item :label="$t('m.MapAdm_Start_Point')">
                <el-select v-model="edgeForm.sourceNodeId" filterable style="width: 100%;">
                  <el-option v-for="n in nodes" :key="`s-${n.id}`" :label="n.title" :value="n.id"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item :label="$t('m.MapAdm_End_Point')">
                <el-select v-model="edgeForm.targetNodeId" filterable style="width: 100%;">
                  <el-option v-for="n in nodes" :key="`t-${n.id}`" :label="n.title" :value="n.id"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item :label="$t('m.MapAdm_Type')">
                <el-select v-model="edgeForm.type" style="width: 100%;">
                  <el-option :label="$t('m.MapAdm_Prerequisite_Strong')" value="prerequisite"></el-option>
                  <el-option :label="$t('m.MapAdm_Related_Weak')" value="related"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" size="mini" style="width:100%" @click="createEdge">{{ $t('m.MapAdm_Create_Edge') }}</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="panel-section">
            <div class="section-title">{{ $t('m.MapAdm_Edge_List') }}</div>
            <el-scrollbar class="admin-list-scroll edge-list-scroll">
              <div v-for="edge in edges" :key="edge.id" class="edge-item">
                <div class="edge-text">
                  {{ nodeTitleMap[edge.sourceNodeId] || edge.sourceNodeId }}
                  <i class="el-icon-right"></i>
                  {{ nodeTitleMap[edge.targetNodeId] || edge.targetNodeId }}
                  <el-tag size="mini" type="info">{{ edgeTypeLabel(edge.type) }}</el-tag>
                </div>
                <el-button size="mini" type="text" style="color:#f56c6c" @click="removeEdge(edge)">{{ $t('m.MapAdm_Delete_Short') }}</el-button>
              </div>
            </el-scrollbar>
          </div>

          <div class="panel-section">
            <div class="section-title">{{ $t('m.MapAdm_Access_Control') }}</div>
            <div class="permission-mode-row">
              <el-tag size="mini" :type="permissionMode === 'all_open' ? 'success' : 'warning'">
                {{ accessModeLabel(permissionMode) }}
              </el-tag>
              <div class="permission-mode-actions">
                <el-button size="mini" type="success" @click="setAccessMode('all_open')">{{ $t('m.MapAdm_Open_All') }}</el-button>
                <el-button size="mini" type="warning" @click="setAccessMode('all_closed')">{{ $t('m.MapAdm_Close_All') }}</el-button>
              </div>
            </div>

            <el-input
              v-model="permissionKeyword"
              size="mini"
              clearable
              :placeholder="$t('m.MapAdm_Search_User_Placeholder')"
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
                  <el-button size="mini" type="success" @click="setUserPermission(u, true)">{{ $t('m.MapAdm_Enable') }}</el-button>
                  <el-button size="mini" type="danger" plain @click="setUserPermission(u, false)">{{ $t('m.MapAdm_Disable') }}</el-button>
                </div>
              </div>
            </div>

            <el-scrollbar class="admin-list-scroll permission-list-scroll">
              <div v-if="mapPermissions.length === 0" class="permission-empty">{{ $t('m.MapAdm_No_Permission_Users') }}</div>
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
                  <el-button size="mini" type="text" style="color:#909399" @click="removeUserPermission(item)">{{ $t('m.MapAdm_Remove') }}</el-button>
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
            <el-form-item :label="$t('m.MapAdm_Node_Type')">
              <el-select v-model="nodeForm.type" :disabled="isEditingNode" style="width:100%">
                <el-option :label="$t('m.MapAdm_Knowledge')" value="knowledge"></el-option>
                <el-option :label="$t('m.MapAdm_Problem')" value="problem"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :md="12" :sm="24">
            <el-form-item :label="$t('m.MapAdm_Title_Field')">
              <el-input v-model="nodeForm.title"></el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="$t('m.MapAdm_Brief')">
          <el-input type="textarea" :rows="2" v-model="nodeForm.description"></el-input>
        </el-form-item>

        <el-row :gutter="12">
          <el-col :md="8" :sm="24">
            <el-form-item :label="$t('m.MapAdm_Difficulty')">
              <el-select v-model="nodeForm.difficulty" style="width:100%">
                <el-option :label="$t('m.MapAdm_Diff_Beginner')" value="beginner"></el-option>
                <el-option :label="$t('m.MapAdm_Diff_Easy')" value="easy"></el-option>
                <el-option :label="$t('m.MapAdm_Diff_Medium')" value="medium"></el-option>
                <el-option :label="$t('m.MapAdm_Diff_Hard')" value="hard"></el-option>
                <el-option :label="$t('m.MapAdm_Diff_Expert')" value="expert"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item :label="$t('m.MapAdm_Tags_Label')">
              <el-input v-model="nodeForm.tagsCsv" :placeholder="$t('m.MapAdm_Tags_Placeholder')"></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-form-item :label="$t('m.MapAdm_Publish_Status')">
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
            <el-form-item :label="$t('m.MapAdm_Level')">
              <el-input-number v-model="nodeForm.level" :step="1" :precision="0" :min="0" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
        </el-row>
        <div class="field-hint">{{ $t('m.MapAdm_Level_Hint') }}</div>

        <el-form-item :label="$t('m.MapAdm_Region')">
          <el-input v-model="nodeForm.region"></el-input>
        </el-form-item>
        <div class="field-hint">{{ $t('m.MapAdm_Region_Hint') }}</div>

        <template v-if="nodeForm.type === 'problem'">
          <el-divider>{{ $t('m.MapAdm_Problem_Binding') }}</el-divider>
          <el-row :gutter="10">
            <el-col :md="14" :sm="24">
              <el-input v-model="problemKeyword" :placeholder="$t('m.MapAdm_Search_Problem_Placeholder')"></el-input>
            </el-col>
            <el-col :md="5" :sm="12">
              <el-button style="width:100%" @click="searchProblems">{{ $t('m.MapAdm_Search_Problem') }}</el-button>
            </el-col>
            <el-col :md="5" :sm="12">
              <el-button type="primary" style="width:100%" @click="verifyProblem">{{ $t('m.MapAdm_Verify_Fill') }}</el-button>
            </el-col>
          </el-row>

          <div class="problem-list" v-if="problemCandidates.length">
            <div class="problem-option" v-for="p in problemCandidates" :key="p.id" @click="pickProblem(p)">
              <div>
                <span class="pid">{{ p.problemDisplayId }}</span>
                <span>{{ p.title }}</span>
              </div>
              <el-tag size="mini" type="warning">{{ $t('m.MapAdm_Difficulty') }} {{ difficultyLabel(p.difficulty) }}</el-tag>
            </div>
          </div>

          <el-row :gutter="12" style="margin-top: 10px;">
            <el-col :md="12" :sm="24">
              <el-form-item :label="$t('m.MapAdm_Main_Problem_Id')">
                <el-input-number v-model="nodeForm.problemId" :min="1" :step="1" style="width:100%"></el-input-number>
              </el-form-item>
            </el-col>
            <el-col :md="12" :sm="24">
              <el-form-item :label="$t('m.MapAdm_Display_Problem_Id')">
                <el-input v-model="nodeForm.problemDisplayId"></el-input>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('m.MapAdm_Remark')">
            <Editor :value.sync="nodeForm.problemRemark" />
          </el-form-item>
          <div class="field-hint">{{ $t('m.MapAdm_Remark_Hint') }}</div>
        </template>

        <template v-else>
          <el-divider>{{ $t('m.MapAdm_Knowledge_Content') }}</el-divider>
          <div class="editor-tips">{{ $t('m.MapAdm_Editor_Tips') }}</div>
          <Editor :value.sync="nodeForm.knowledgeContent" />
        </template>

        <el-form-item :label="$t('m.MapAdm_Metadata_Label')">
          <el-input type="textarea" :rows="3" v-model="nodeForm.metadataJson" :placeholder="$t('m.MapAdm_Metadata_Prefix') + '{"estimatedMinutes":30}'"></el-input>
        </el-form-item>
        <div class="field-hint">{{ $t('m.MapAdm_Metadata_Hint') }}</div>
      </el-form>

      <div slot="footer">
        <el-button size="small" @click="nodeDialogVisible = false">{{ $t('m.MapAdm_Cancel') }}</el-button>
        <el-button size="small" type="primary" @click="saveNode">{{ $t('m.MapAdm_Save_Node') }}</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import learningMapApi from '@/api/learningMap'
import LearningMapCanvas from '@/views/oj/learning-map/components/LearningMapCanvas'
import Editor from '@/components/admin/Editor'

// 模块级映射表只存 i18n 键名，渲染时通过 tKey() 在组件上下文中翻译
const NODE_TYPE_LABEL = {
  knowledge: 'MapAdm_Knowledge',
  problem: 'MapAdm_Problem'
}

const EDGE_TYPE_LABEL = {
  prerequisite: 'MapAdm_Prerequisite',
  related: 'MapAdm_Related'
}

const DIFFICULTY_LABEL = {
  beginner: 'MapAdm_Diff_Beginner',
  easy: 'MapAdm_Diff_Easy',
  medium: 'MapAdm_Diff_Medium',
  hard: 'MapAdm_Diff_Hard',
  expert: 'MapAdm_Diff_Expert'
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
      return this.isEditingNode ? this.$t('m.MapAdm_Edit_Node', { id: this.editingNodeId }) : this.$t('m.MapAdm_Create_Node')
    }
  },
  mounted() {
    this.loadMap()
  },
  methods: {
    // 将模块级映射表中的 i18n 键名翻译为当前语言的文案
    tKey(key) {
      return key ? this.$t('m.' + key) : ''
    },
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
        problemRemark: '',
        metadataJson: '{}'
      }
    },
    parseMetadata(json) {
      if (!json || !json.trim()) return {}
      try {
        const parsed = JSON.parse(json)
        if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
          throw new Error(this.$t('m.MapAdm_Metadata_Must_Object'))
        }
        return parsed
      } catch (e) {
        if (e.message === this.$t('m.MapAdm_Metadata_Must_Object')) {
          throw e
        }
        throw new Error(this.$t('m.MapAdm_Metadata_Invalid_Json'))
      }
    },
    parseMetadataQuietly(json) {
      if (!json || !String(json).trim()) return {}
      try {
        const parsed = JSON.parse(json)
        return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
      } catch (e) {
        return {}
      }
    },
    nodeTypeLabel(type) {
      return NODE_TYPE_LABEL[type] ? this.tKey(NODE_TYPE_LABEL[type]) : type
    },
    edgeTypeLabel(type) {
      return EDGE_TYPE_LABEL[type] ? this.tKey(EDGE_TYPE_LABEL[type]) : type
    },
    difficultyLabel(value) {
      if (typeof value === 'number') {
        const d = Number(value)
        if (d <= 0) return this.tKey(DIFFICULTY_LABEL.beginner)
        if (d === 1) return this.tKey(DIFFICULTY_LABEL.easy)
        if (d === 2) return this.tKey(DIFFICULTY_LABEL.medium)
        if (d === 3) return this.tKey(DIFFICULTY_LABEL.hard)
        return this.tKey(DIFFICULTY_LABEL.expert)
      }
      return this.tKey(DIFFICULTY_LABEL[value]) || value || this.tKey(DIFFICULTY_LABEL.beginner)
    },
    toNodePayload(form) {
      const metadata = this.parseMetadata(form.metadataJson)
      if (form.type === 'problem') {
        const remark = (form.problemRemark || '').trim()
        if (remark) {
          metadata.remark = remark
        } else {
          delete metadata.remark
        }
      }
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
        metadata
      }
    },
    fillNodeForm(node) {
      const metadata = this.parseMetadataQuietly(node.metadata)
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
        problemRemark: typeof metadata.remark === 'string' ? metadata.remark : (node.remark || ''),
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
        this.$message.error(e.message || this.$t('m.MapAdm_Load_Failed'))
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
          this.$message.warning(this.$t('m.MapAdm_Title_Empty'))
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
          this.$message.error(e.message || this.$t('m.MapAdm_Save_Failed'))
        }
        return false
      } finally {
        this.savingMeta = false
      }
    },
    toggleMapStatus() {
      if (this.mapInfo.status === 'published') {
        this.$confirm(this.$t('m.MapAdm_Hide_Confirm'), this.$t('m.MapAdm_Hide_Confirm_Title'), {
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
            this.$message.success(this.$t('m.MapAdm_Hidden'))
          } catch (e) {
            this.$message.error(e.message || this.$t('m.MapAdm_Hide_Failed'))
          }
        })
        return
      }

      this.$confirm(this.$t('m.MapAdm_Publish_Confirm'), this.$t('m.MapAdm_Publish_Confirm_Title'), {
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
          this.$message.success(this.$t('m.MapAdm_Publish_Success'))
        } catch (e) {
          this.$message.error(e.message || this.$t('m.MapAdm_Publish_Failed'))
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
        this.$message.warning(this.$t('m.MapAdm_Node_Title_Empty'))
        return
      }
      try {
        const payload = this.toNodePayload(this.nodeForm)
        if (this.isEditingNode) {
          await learningMapApi.adminUpdateNode(this.mapId, this.editingNodeId, payload)
          this.$message.success(this.$t('m.MapAdm_Node_Updated'))
        } else {
          await learningMapApi.adminCreateNode(this.mapId, payload)
          this.$message.success(this.$t('m.MapAdm_Node_Created'))
        }
        this.nodeDialogVisible = false
        await this.loadMap()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Save_Node_Failed'))
      }
    },
    removeNode(node) {
      this.$confirm(this.$t('m.MapAdm_Delete_Node_Confirm', { title: node.title }), this.$t('m.MapAdm_Delete_Node_Title'), {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteNode(this.mapId, node.id)
          if (this.selectedNode && this.selectedNode.id === node.id) {
            this.selectedNode = null
          }
          this.$message.success(this.$t('m.MapAdm_Node_Deleted'))
          await this.loadMap()
        } catch (e) {
          this.$message.error(e.message || this.$t('m.MapAdm_Delete_Node_Failed'))
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
        this.$message.error(this.$t('m.MapAdm_Node_Position_Save_Failed'))
      }
    },
    async createEdge() {
      if (!this.edgeForm.sourceNodeId || !this.edgeForm.targetNodeId) {
        this.$message.warning(this.$t('m.MapAdm_Select_Start_End'))
        return
      }
      try {
        await learningMapApi.adminCreateEdge(this.mapId, {
          sourceNodeId: this.edgeForm.sourceNodeId,
          targetNodeId: this.edgeForm.targetNodeId,
          type: this.edgeForm.type
        })
        this.$message.success(this.$t('m.MapAdm_Edge_Create_Success'))
        this.edgeForm = { sourceNodeId: null, targetNodeId: null, type: 'prerequisite' }
        await this.loadMap()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Create_Edge_Failed'))
      }
    },
    removeEdge(edge) {
      this.$confirm(this.$t('m.MapAdm_Delete_Edge_Confirm'), this.$t('m.MapAdm_Delete_Edge_Title'), {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteEdge(this.mapId, edge.id)
          this.$message.success(this.$t('m.MapAdm_Edge_Deleted'))
          await this.loadMap()
        } catch (e) {
          this.$message.error(e.message || this.$t('m.MapAdm_Delete_Edge_Failed'))
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
        this.$message.error(this.$t('m.MapAdm_Search_Problem_Failed'))
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
        this.$message.warning(this.$t('m.MapAdm_Problem_Id_Required'))
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
        this.$message.success(this.$t('m.MapAdm_Problem_Verify_Success'))
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Problem_Not_Available'))
      }
    },
    accessModeLabel(mode) {
      return mode === 'all_closed' ? this.$t('m.MapAdm_Access_Mode_Closed') : this.$t('m.MapAdm_Access_Mode_Open')
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
        this.$message.success(mode === 'all_open' ? this.$t('m.MapAdm_Set_Open_Success') : this.$t('m.MapAdm_Set_Close_Success'))
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Set_Failed'))
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
        this.$message.error(e.message || this.$t('m.MapAdm_Search_User_Failed'))
      }
    },
    async setUserPermission(user, enabled) {
      if (!user || !user.userId) return
      try {
        await learningMapApi.adminSetMapUserPermission(this.mapId, user.userId, enabled)
        this.$message.success(enabled ? this.$t('m.MapAdm_Permission_Enabled') : this.$t('m.MapAdm_Permission_Disabled'))
        await this.loadPermissions()
      } catch (e) {
        this.$message.error(e.message || this.$t('m.MapAdm_Set_User_Permission_Failed'))
      }
    },
    async toggleUserPermission(item, enabled) {
      const original = item.enabled
      item.enabled = enabled
      try {
        await learningMapApi.adminSetMapUserPermission(this.mapId, item.userId, enabled)
        this.$message.success(this.$t('m.MapAdm_Permission_Updated'))
      } catch (e) {
        item.enabled = original
        this.$message.error(e.message || this.$t('m.MapAdm_Update_Permission_Failed'))
      }
    },
    removeUserPermission(item) {
      this.$confirm(this.$t('m.MapAdm_Remove_User_Permission_Confirm', { uid: item.userId }), this.$t('m.MapAdm_Remove_Permission_Title'), {
        type: 'warning'
      }).then(async () => {
        try {
          await learningMapApi.adminDeleteMapUserPermission(this.mapId, item.userId)
          this.$message.success(this.$t('m.MapAdm_Removed'))
          await this.loadPermissions()
        } catch (e) {
          this.$message.error(e.message || this.$t('m.MapAdm_Remove_Failed'))
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
.admin-list-scroll >>> .el-scrollbar__wrap {
  overflow-x: hidden;
  overflow-y: auto;
}
.node-list-scroll >>> .el-scrollbar__wrap {
  max-height: 280px;
}
.edge-list-scroll >>> .el-scrollbar__wrap {
  max-height: 250px;
}
.permission-list-scroll {
  margin-top: 8px;
}
.permission-list-scroll >>> .el-scrollbar__wrap {
  max-height: 200px;
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
