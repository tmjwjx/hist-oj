<template>
  <div class="detail-panel" v-if="node">
    <div class="detail-header">
      <div class="detail-title-wrap">
        <div class="detail-title">{{ node.title }}</div>
        <div class="detail-sub">{{ node.type === 'knowledge' ? '知识点' : '题目节点' }}</div>
      </div>
      <el-tag size="mini" :type="statusTagType">{{ statusLabel }}</el-tag>
    </div>

    <div class="detail-meta">
      <el-tag size="mini" effect="plain">难度：{{ difficultyLabel }}</el-tag>
      <el-tag v-for="tag in node.tagsList || []" :key="tag" size="mini" type="info" effect="plain">{{ tag }}</el-tag>
    </div>

    <div class="locked-box" v-if="progress && progress.status === 'locked'">
      <div class="locked-title"><i class="el-icon-lock"></i> 未解锁</div>
      <div class="locked-text">完成以下前置节点后可解锁：</div>
      <ul>
        <li v-for="id in progress.missingPrerequisiteIds || []" :key="id">
          {{ nodeTitleMap[id] || `节点 ${id}` }}
        </li>
      </ul>
    </div>

    <template v-if="node.type === 'knowledge'">
      <div class="action-row">
        <el-button
          size="mini"
          type="warning"
          :disabled="!progress || progress.status === 'locked'"
          @click="$emit('start-node', node)"
        >标记学习中</el-button>
        <el-button
          size="mini"
          type="success"
          :disabled="!progress || progress.status === 'locked'"
          @click="$emit('complete-node', node)"
        >标记已学完</el-button>
        <el-button
          size="mini"
          plain
          icon="el-icon-full-screen"
          :disabled="!progress || progress.status === 'locked'"
          @click="$emit('open-knowledge-fullscreen', node)"
        >全屏学习</el-button>
      </div>
      <div class="resource-box" v-if="knowledgeResources.length > 0">
        <div class="resource-title">资料附件</div>
        <div class="resource-list">
          <el-link
            v-for="item in knowledgeResources"
            :key="item.url"
            class="resource-item"
            :underline="false"
            type="primary"
            @click.native="openResource(item.url)"
          >
            <i :class="item.icon"></i>
            <span>{{ item.name }}</span>
          </el-link>
        </div>
      </div>
      <div class="content-title">学习内容</div>
      <Markdown :content="node.knowledgeContent || '暂无学习内容'" :is-avoid-xss="true" />
    </template>

    <template v-else>
      <div class="problem-card">
        <div class="problem-id">题号：{{ node.problemDisplayId || (node.problemInfo && node.problemInfo.problemDisplayId) || '-' }}</div>
        <div class="problem-title">{{ (node.problemInfo && node.problemInfo.title) || node.title }}</div>
        <div class="problem-tags">
          <el-tag
            v-for="tag in (node.problemInfo && node.problemInfo.tags) || []"
            :key="tag"
            size="mini"
            effect="plain"
          >{{ tag }}</el-tag>
        </div>
        <div class="problem-remark-box" v-if="nodeRemark">
          <div class="problem-remark-title">
            <i class="el-icon-document"></i>
            <span>备注</span>
          </div>
          <div class="problem-remark-content">
            <Markdown :content="nodeRemark" :is-avoid-xss="true" />
          </div>
        </div>
        <el-button
          size="mini"
          type="primary"
          :disabled="progress && progress.status === 'locked'"
          @click="$emit('go-problem', node)"
        >进入主 OJ 题目</el-button>
      </div>
    </template>
  </div>

  <div class="detail-empty" v-else>
    <i class="el-icon-position"></i>
    <span>点击航海点查看详情</span>
  </div>
</template>

<script>
import Markdown from '@/components/oj/common/Markdown'

const STATUS_LABEL = {
  locked: '未解锁',
  available: '可学习',
  in_progress: '进行中',
  completed: '已完成',
  mastered: '已精通'
}

const DIFFICULTY_LABEL = {
  beginner: '入门',
  easy: '简单',
  medium: '中等',
  hard: '困难',
  expert: '专家'
}

export default {
  name: 'LearningMapDetailPanel',
  components: { Markdown },
  props: {
    node: {
      type: Object,
      default: null
    },
    progress: {
      type: Object,
      default: null
    },
    nodeTitleMap: {
      type: Object,
      default: () => ({})
    }
  },
  computed: {
    knowledgeResources() {
      if (!this.node || this.node.type !== 'knowledge') {
        return []
      }
      return this.extractResources(this.node.knowledgeContent || '')
    },
    nodeRemark() {
      if (!this.node || this.node.type !== 'problem') {
        return ''
      }
      if (this.node.remark) {
        return String(this.node.remark).trim()
      }
      const metadata = this.parseNodeMetadata(this.node.metadata)
      return typeof metadata.remark === 'string' ? metadata.remark.trim() : ''
    },
    difficultyLabel() {
      const key = this.node && this.node.difficulty
      return DIFFICULTY_LABEL[key] || '入门'
    },
    statusLabel() {
      const key = this.progress && this.progress.status
      return STATUS_LABEL[key] || '未知'
    },
    statusTagType() {
      const key = this.progress && this.progress.status
      if (key === 'completed' || key === 'mastered') return 'success'
      if (key === 'available' || key === 'in_progress') return 'warning'
      return 'info'
    }
  },
  methods: {
    parseNodeMetadata(raw) {
      if (!raw || !String(raw).trim()) return {}
      try {
        const parsed = JSON.parse(raw)
        return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
      } catch (e) {
        return {}
      }
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
    }
  }
}
</script>

<style scoped>
.detail-panel {
  height: 100%;
  overflow-y: auto;
  padding: 14px;
  background: #fff;
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.detail-title {
  font-size: 18px;
  font-weight: 700;
}
.detail-sub {
  margin-top: 4px;
  color: #6b7280;
  font-size: 12px;
}
.detail-meta {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.locked-box {
  border: 1px solid #dbeafe;
  background: #eff6ff;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 12px;
}
.locked-title {
  font-weight: 600;
  color: #1d4ed8;
  margin-bottom: 4px;
}
.locked-text {
  font-size: 12px;
  color: #475569;
  margin-bottom: 4px;
}
.locked-box ul {
  margin: 0;
  padding-left: 18px;
}
.action-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.resource-box {
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 10px;
  border: 1px dashed #9ec5ff;
  background: #f8fbff;
}
.resource-title {
  font-size: 13px;
  color: #274c77;
  margin-bottom: 6px;
  font-weight: 600;
}
.resource-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.resource-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 999px;
  background: #e8f3ff;
}
.content-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}
.problem-card {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px;
  background: #fafafa;
}
.problem-id {
  color: #4b5563;
  font-size: 12px;
  margin-bottom: 6px;
}
.problem-title {
  font-weight: 600;
  margin-bottom: 8px;
}
.problem-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}
.problem-remark-box {
  border: 1px dashed #f0c36a;
  border-radius: 8px;
  padding: 9px 10px;
  margin-bottom: 10px;
  background: #fffaf0;
}
.problem-remark-title {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 600;
  color: #8a5a12;
  margin-bottom: 6px;
}
.problem-remark-content >>> .markdown-body {
  font-size: 13px;
  line-height: 1.7;
  color: #374151;
  word-break: break-word;
}
.problem-remark-content >>> .markdown-body p {
  margin: 0 0 6px;
}
.problem-remark-content >>> .markdown-body p:last-child {
  margin-bottom: 0;
}
.problem-remark-content >>> .markdown-body a {
  word-break: break-all;
}
.detail-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #94a3b8;
  font-size: 14px;
  background: #f8fafc;
}
</style>
