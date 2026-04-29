<template>
  <div class="learning-map-canvas" ref="stage" @wheel.prevent="onWheel" @mousedown="onStageMouseDown" @touchstart="onStageTouchStart">
    <div class="map-grid"></div>

    <div class="world" :style="worldStyle">
      <svg class="edge-layer" :width="worldSize.width" :height="worldSize.height">
        <line
          v-for="edge in filteredEdges"
          :key="edge.id"
          :x1="getNodePosition(edge.sourceNodeId).x"
          :y1="getNodePosition(edge.sourceNodeId).y"
          :x2="getNodePosition(edge.targetNodeId).x"
          :y2="getNodePosition(edge.targetNodeId).y"
          :class="['edge-line', edge.type === 'related' ? 'edge-related' : 'edge-prerequisite']"
        />
      </svg>

      <div class="node-layer">
        <div
          v-for="node in filteredNodes"
          :key="node.id"
          class="map-node"
          :class="[
            `type-${node.type}`,
            `status-${getNodeStatus(node.id)}`,
            selectedNodeId === node.id ? 'is-selected' : '',
            draggingNodeId === node.id ? 'is-dragging' : ''
          ]"
          :style="nodeStyle(node)"
          :title="buildNodeTooltip(node)"
          @click.stop="handleNodeClick(node)"
          @mousedown.stop="onNodeMouseDown(node, $event)"
          @touchstart.stop="onNodeTouchStart(node, $event)"
        >
          <div class="node-icon">
            <i v-if="getNodeStatus(node.id) === 'locked'" class="el-icon-lock"></i>
            <i v-else-if="getNodeStatus(node.id) === 'completed'" class="el-icon-circle-check"></i>
            <i v-else-if="getNodeStatus(node.id) === 'mastered'" class="el-icon-medal-1"></i>
            <i v-else-if="node.type === 'problem'" class="el-icon-document"></i>
            <i v-else class="el-icon-reading"></i>
          </div>
          <div class="node-title">{{ node.title }}</div>
        </div>
      </div>
    </div>

    <div class="canvas-tools">
      <el-button size="mini" icon="el-icon-plus" @click="zoom(1.12)"></el-button>
      <el-button size="mini" icon="el-icon-minus" @click="zoom(0.88)"></el-button>
      <el-button size="mini" icon="el-icon-refresh-left" @click="goBackView">返回视角</el-button>
      <el-button size="mini" icon="el-icon-full-screen" @click="fitToContent">适配视图</el-button>
    </div>

    <div class="mini-map" v-if="filteredNodes.length > 0">
      <svg :width="miniSize.width" :height="miniSize.height">
        <rect x="0" y="0" :width="miniSize.width" :height="miniSize.height" rx="8" ry="8" class="mini-bg" />

        <line
          v-for="edge in filteredEdges"
          :key="`mini-${edge.id}`"
          :x1="miniX(getNodePosition(edge.sourceNodeId).x)"
          :y1="miniY(getNodePosition(edge.sourceNodeId).y)"
          :x2="miniX(getNodePosition(edge.targetNodeId).x)"
          :y2="miniY(getNodePosition(edge.targetNodeId).y)"
          class="mini-edge"
        />

        <circle
          v-for="node in filteredNodes"
          :key="`mini-node-${node.id}`"
          :cx="miniX(getNodePosition(node.id).x)"
          :cy="miniY(getNodePosition(node.id).y)"
          r="3"
          :class="['mini-node', `mini-${node.type}`]"
        />

        <rect
          :x="miniViewport.x"
          :y="miniViewport.y"
          :width="miniViewport.width"
          :height="miniViewport.height"
          class="mini-viewport"
        />
      </svg>
    </div>
  </div>
</template>

<script>
const MIN_SCALE = 0.28
const MAX_SCALE = 2.8
const NODE_TYPE_LABEL = {
  knowledge: '知识点',
  problem: '题目'
}

const STATUS_LABEL = {
  locked: '未解锁',
  available: '可学习',
  in_progress: '进行中',
  completed: '已完成',
  mastered: '已精通'
}

export default {
  name: 'LearningMapCanvas',
  props: {
    nodes: {
      type: Array,
      default: () => []
    },
    edges: {
      type: Array,
      default: () => []
    },
    progressMap: {
      type: Object,
      default: () => ({})
    },
    selectedNodeId: {
      type: Number,
      default: null
    },
    filterStatus: {
      type: String,
      default: 'all'
    },
    filterType: {
      type: String,
      default: 'all'
    },
    editable: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      viewport: {
        x: 120,
        y: 120,
        scale: 0.85
      },
      stageSize: {
        width: 1280,
        height: 760
      },
      worldSize: {
        width: 2600,
        height: 1800
      },
      miniSize: {
        width: 220,
        height: 140
      },
      panning: false,
      dragStart: null,
      draggingNodeId: null,
      draggingStartWorld: null,
      localNodePositions: {},
      viewHistory: [],
      lastTouchDistance: 0
    }
  },
  computed: {
    filteredNodes() {
      return this.nodes.filter(node => {
        if (this.filterType !== 'all' && node.type !== this.filterType) {
          return false
        }
        const status = this.getNodeStatus(node.id)
        switch (this.filterStatus) {
          case 'unlocked':
            return status !== 'locked'
          case 'unfinished':
            return status === 'available' || status === 'in_progress'
          case 'completed':
            return status === 'completed' || status === 'mastered'
          default:
            return true
        }
      })
    },
    visibleNodeSet() {
      const set = {}
      this.filteredNodes.forEach(n => {
        set[n.id] = true
      })
      return set
    },
    filteredEdges() {
      return this.edges.filter(edge => this.visibleNodeSet[edge.sourceNodeId] && this.visibleNodeSet[edge.targetNodeId])
    },
    worldStyle() {
      return {
        width: `${this.worldSize.width}px`,
        height: `${this.worldSize.height}px`,
        transform: `translate(${this.viewport.x}px, ${this.viewport.y}px) scale(${this.viewport.scale})`
      }
    },
    miniBounds() {
      if (this.filteredNodes.length === 0) {
        return { minX: 0, maxX: 1, minY: 0, maxY: 1 }
      }
      const xs = this.filteredNodes.map(n => this.getNodePosition(n.id).x)
      const ys = this.filteredNodes.map(n => this.getNodePosition(n.id).y)
      const minX = Math.min(...xs) - 120
      const maxX = Math.max(...xs) + 120
      const minY = Math.min(...ys) - 120
      const maxY = Math.max(...ys) + 120
      return { minX, maxX, minY, maxY }
    },
    miniViewport() {
      const rangeX = Math.max(this.miniBounds.maxX - this.miniBounds.minX, 1)
      const rangeY = Math.max(this.miniBounds.maxY - this.miniBounds.minY, 1)
      const worldLeft = (-this.viewport.x) / this.viewport.scale
      const worldTop = (-this.viewport.y) / this.viewport.scale
      const worldW = this.stageSize.width / this.viewport.scale
      const worldH = this.stageSize.height / this.viewport.scale
      const x = ((worldLeft - this.miniBounds.minX) / rangeX) * this.miniSize.width
      const y = ((worldTop - this.miniBounds.minY) / rangeY) * this.miniSize.height
      const width = (worldW / rangeX) * this.miniSize.width
      const height = (worldH / rangeY) * this.miniSize.height
      return {
        x: Math.max(0, Math.min(this.miniSize.width - width, x)),
        y: Math.max(0, Math.min(this.miniSize.height - height, y)),
        width: Math.max(8, Math.min(this.miniSize.width, width)),
        height: Math.max(8, Math.min(this.miniSize.height, height))
      }
    }
  },
  watch: {
    nodes: {
      deep: true,
      immediate: true,
      handler() {
        this.recalculateWorldSize()
      }
    }
  },
  mounted() {
    this.updateStageSize()
    window.addEventListener('resize', this.updateStageSize)
    window.addEventListener('mousemove', this.onPointerMove)
    window.addEventListener('mouseup', this.onPointerUp)
    window.addEventListener('touchmove', this.onTouchMove, { passive: false })
    window.addEventListener('touchend', this.onPointerUp)
    this.$nextTick(() => {
      this.fitToContent()
    })
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.updateStageSize)
    window.removeEventListener('mousemove', this.onPointerMove)
    window.removeEventListener('mouseup', this.onPointerUp)
    window.removeEventListener('touchmove', this.onTouchMove)
    window.removeEventListener('touchend', this.onPointerUp)
  },
  methods: {
    getNodeTypeLabel(type) {
      return NODE_TYPE_LABEL[type] || type
    },
    getStatusLabel(status) {
      return STATUS_LABEL[status] || status
    },
    getNodeStatus(nodeId) {
      if (!this.progressMap || !this.progressMap[nodeId]) {
        return 'locked'
      }
      return this.progressMap[nodeId].status || 'locked'
    },
    buildNodeTooltip(node) {
      const status = this.getNodeStatus(node.id)
      return `${node.title}\n类型：${this.getNodeTypeLabel(node.type)}\n状态：${this.getStatusLabel(status)}`
    },
    nodeStyle(node) {
      const pos = this.getNodePosition(node.id)
      return {
        left: `${pos.x}px`,
        top: `${pos.y}px`
      }
    },
    getNodePosition(nodeId) {
      const local = this.localNodePositions[nodeId]
      if (local) {
        return local
      }
      const raw = this.nodes.find(n => n.id === nodeId)
      if (!raw) {
        return { x: 0, y: 0 }
      }
      return { x: Number(raw.x) || 0, y: Number(raw.y) || 0 }
    },
    setNodePosition(nodeId, x, y) {
      this.$set(this.localNodePositions, nodeId, { x, y })
    },
    updateStageSize() {
      if (!this.$refs.stage) return
      const rect = this.$refs.stage.getBoundingClientRect()
      this.stageSize.width = Math.max(320, Math.floor(rect.width))
      this.stageSize.height = Math.max(260, Math.floor(rect.height))
    },
    recalculateWorldSize() {
      if (!this.nodes.length) {
        this.worldSize = { width: 2600, height: 1800 }
        return
      }
      const xs = this.nodes.map(n => Number(n.x) || 0)
      const ys = this.nodes.map(n => Number(n.y) || 0)
      const maxX = Math.max(...xs)
      const maxY = Math.max(...ys)
      const minX = Math.min(...xs)
      const minY = Math.min(...ys)
      const width = Math.max(2200, Math.ceil(maxX - minX + 520))
      const height = Math.max(1600, Math.ceil(maxY - minY + 520))
      this.worldSize = { width, height }
    },
    miniX(x) {
      const rangeX = Math.max(this.miniBounds.maxX - this.miniBounds.minX, 1)
      return ((x - this.miniBounds.minX) / rangeX) * this.miniSize.width
    },
    miniY(y) {
      const rangeY = Math.max(this.miniBounds.maxY - this.miniBounds.minY, 1)
      return ((y - this.miniBounds.minY) / rangeY) * this.miniSize.height
    },
    onWheel(e) {
      const delta = e.deltaY < 0 ? 1.1 : 0.9
      this.zoomAt(delta, e.clientX, e.clientY)
    },
    zoom(factor) {
      const cx = this.$refs.stage.getBoundingClientRect().left + this.stageSize.width / 2
      const cy = this.$refs.stage.getBoundingClientRect().top + this.stageSize.height / 2
      this.zoomAt(factor, cx, cy)
    },
    zoomAt(factor, clientX, clientY) {
      const rect = this.$refs.stage.getBoundingClientRect()
      const pointX = clientX - rect.left
      const pointY = clientY - rect.top
      const nextScale = Math.max(MIN_SCALE, Math.min(MAX_SCALE, this.viewport.scale * factor))
      const worldX = (pointX - this.viewport.x) / this.viewport.scale
      const worldY = (pointY - this.viewport.y) / this.viewport.scale
      this.viewport.scale = nextScale
      this.viewport.x = pointX - worldX * nextScale
      this.viewport.y = pointY - worldY * nextScale
      this.emitViewport()
    },
    onStageMouseDown(e) {
      if (this.draggingNodeId) return
      if (e.button !== 0) return
      this.panning = true
      this.dragStart = {
        x: e.clientX,
        y: e.clientY,
        vx: this.viewport.x,
        vy: this.viewport.y
      }
    },
    onStageTouchStart(e) {
      if (e.touches.length === 2) {
        this.lastTouchDistance = this.touchDistance(e.touches)
        return
      }
      if (e.touches.length === 1 && !this.draggingNodeId) {
        const t = e.touches[0]
        this.panning = true
        this.dragStart = {
          x: t.clientX,
          y: t.clientY,
          vx: this.viewport.x,
          vy: this.viewport.y
        }
      }
    },
    onNodeMouseDown(node, e) {
      if (!this.editable || e.button !== 0) {
        return
      }
      const world = this.clientToWorld(e.clientX, e.clientY)
      const pos = this.getNodePosition(node.id)
      this.draggingNodeId = node.id
      this.draggingStartWorld = {
        offsetX: world.x - pos.x,
        offsetY: world.y - pos.y
      }
    },
    onNodeTouchStart(node, e) {
      if (!this.editable || !e.touches || e.touches.length !== 1) {
        return
      }
      const t = e.touches[0]
      const world = this.clientToWorld(t.clientX, t.clientY)
      const pos = this.getNodePosition(node.id)
      this.draggingNodeId = node.id
      this.draggingStartWorld = {
        offsetX: world.x - pos.x,
        offsetY: world.y - pos.y
      }
    },
    onPointerMove(e) {
      if (this.draggingNodeId) {
        const clientX = e.touches ? e.touches[0].clientX : e.clientX
        const clientY = e.touches ? e.touches[0].clientY : e.clientY
        const world = this.clientToWorld(clientX, clientY)
        const x = world.x - this.draggingStartWorld.offsetX
        const y = world.y - this.draggingStartWorld.offsetY
        this.setNodePosition(this.draggingNodeId, x, y)
        return
      }
      if (!this.panning || !this.dragStart) {
        return
      }
      const clientX = e.touches ? e.touches[0].clientX : e.clientX
      const clientY = e.touches ? e.touches[0].clientY : e.clientY
      this.viewport.x = this.dragStart.vx + (clientX - this.dragStart.x)
      this.viewport.y = this.dragStart.vy + (clientY - this.dragStart.y)
      this.emitViewport()
    },
    onTouchMove(e) {
      if (e.touches.length === 2) {
        e.preventDefault()
        const distance = this.touchDistance(e.touches)
        if (!this.lastTouchDistance) {
          this.lastTouchDistance = distance
          return
        }
        const factor = distance / this.lastTouchDistance
        const cx = (e.touches[0].clientX + e.touches[1].clientX) / 2
        const cy = (e.touches[0].clientY + e.touches[1].clientY) / 2
        this.zoomAt(factor, cx, cy)
        this.lastTouchDistance = distance
        return
      }
      if (this.draggingNodeId || this.panning) {
        e.preventDefault()
      }
      this.onPointerMove(e)
    },
    onPointerUp() {
      this.panning = false
      this.dragStart = null
      this.lastTouchDistance = 0
      if (this.draggingNodeId) {
        const nodeId = this.draggingNodeId
        const pos = this.localNodePositions[nodeId]
        this.draggingNodeId = null
        this.draggingStartWorld = null
        if (pos) {
          this.$emit('node-position-change', { nodeId, x: pos.x, y: pos.y })
        }
      }
    },
    touchDistance(touches) {
      const dx = touches[0].clientX - touches[1].clientX
      const dy = touches[0].clientY - touches[1].clientY
      return Math.sqrt(dx * dx + dy * dy)
    },
    clientToWorld(clientX, clientY) {
      const rect = this.$refs.stage.getBoundingClientRect()
      const x = (clientX - rect.left - this.viewport.x) / this.viewport.scale
      const y = (clientY - rect.top - this.viewport.y) / this.viewport.scale
      return { x, y }
    },
    emitViewport() {
      this.$emit('viewport-change', { ...this.viewport })
    },
    handleNodeClick(node) {
      this.$emit('node-click', node)
    },
    focusNode(nodeId, withHistory = true) {
      const node = this.filteredNodes.find(n => n.id === nodeId) || this.nodes.find(n => n.id === nodeId)
      if (!node) {
        return
      }
      if (withHistory) {
        this.viewHistory.push({ ...this.viewport })
      }
      const targetScale = Math.max(1, this.viewport.scale)
      const targetX = this.stageSize.width / 2 - this.getNodePosition(node.id).x * targetScale
      const targetY = this.stageSize.height / 2 - this.getNodePosition(node.id).y * targetScale
      this.animateViewport({ x: targetX, y: targetY, scale: targetScale })
    },
    goBackView() {
      if (this.viewHistory.length === 0) {
        return
      }
      const prev = this.viewHistory.pop()
      this.animateViewport(prev)
    },
    fitToContent() {
      if (!this.filteredNodes.length) {
        return
      }
      const xs = this.filteredNodes.map(n => this.getNodePosition(n.id).x)
      const ys = this.filteredNodes.map(n => this.getNodePosition(n.id).y)
      const minX = Math.min(...xs) - 180
      const maxX = Math.max(...xs) + 180
      const minY = Math.min(...ys) - 140
      const maxY = Math.max(...ys) + 140
      const width = Math.max(300, maxX - minX)
      const height = Math.max(260, maxY - minY)
      const scale = Math.max(MIN_SCALE, Math.min(MAX_SCALE, Math.min(this.stageSize.width / width, this.stageSize.height / height)))
      const x = this.stageSize.width / 2 - (minX + width / 2) * scale
      const y = this.stageSize.height / 2 - (minY + height / 2) * scale
      this.animateViewport({ x, y, scale })
    },
    animateViewport(target) {
      const start = { ...this.viewport }
      const duration = 360
      const begin = performance.now()
      const step = now => {
        const p = Math.min(1, (now - begin) / duration)
        const ease = 1 - Math.pow(1 - p, 3)
        this.viewport.x = start.x + (target.x - start.x) * ease
        this.viewport.y = start.y + (target.y - start.y) * ease
        this.viewport.scale = start.scale + (target.scale - start.scale) * ease
        this.emitViewport()
        if (p < 1) {
          requestAnimationFrame(step)
        }
      }
      requestAnimationFrame(step)
    }
  }
}
</script>

<style scoped>
.learning-map-canvas {
  position: relative;
  width: 100%;
  height: 72vh;
  min-height: 540px;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid #cfe1f8;
  background:
    radial-gradient(circle at 14% 18%, rgba(255, 241, 189, 0.65), transparent 38%),
    radial-gradient(circle at 85% 24%, rgba(199, 241, 255, 0.65), transparent 42%),
    linear-gradient(180deg, #ecf6ff 0%, #e7f3ff 52%, #fff5da 100%);
  touch-action: none;
}
.map-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(to right, rgba(74, 144, 226, 0.1) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(74, 144, 226, 0.1) 1px, transparent 1px);
  background-size: 28px 28px;
  pointer-events: none;
}
.world {
  position: absolute;
  left: 0;
  top: 0;
  transform-origin: 0 0;
}
.edge-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}
.edge-line {
  stroke-width: 2;
}
.edge-prerequisite {
  stroke: #2a6ed4;
}
.edge-related {
  stroke: #94a3b8;
  stroke-dasharray: 6 4;
}
.node-layer {
  position: absolute;
  inset: 0;
}
.map-node {
  position: absolute;
  width: 160px;
  min-height: 64px;
  transform: translate(-50%, -50%);
  border-radius: 16px;
  border: 2px solid #ceddf2;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 6px 14px rgba(27, 39, 94, 0.12);
  padding: 8px 10px;
  cursor: pointer;
  transition: box-shadow 0.18s ease, transform 0.18s ease, border-color 0.18s ease;
  user-select: none;
}
.map-node:hover {
  transform: translate(-50%, -52%);
  box-shadow: 0 10px 18px rgba(36, 78, 160, 0.2);
}
.map-node.is-selected {
  box-shadow: 0 0 0 3px #2d8cf0, 0 10px 22px rgba(45, 140, 240, 0.24);
}
.map-node.is-dragging {
  cursor: grabbing;
}
.map-node .node-icon {
  font-size: 16px;
  margin-bottom: 5px;
}
.map-node .node-title {
  font-size: 13px;
  line-height: 1.35;
  font-weight: 600;
  word-break: break-word;
}
.map-node.type-knowledge {
  border-color: #7fd8b1;
  background: linear-gradient(180deg, #ffffff 0%, #f2fff7 100%);
}
.map-node.type-problem {
  border-color: #ffd08a;
  background: linear-gradient(180deg, #ffffff 0%, #fff8ea 100%);
}
.map-node.status-available,
.map-node.status-in_progress {
  animation: floatNode 2.8s ease-in-out infinite;
}
.map-node.status-locked {
  filter: grayscale(1);
  opacity: 0.65;
}
.map-node.status-completed {
  border-color: #42b983;
  background: #f0fbf5;
}
.map-node.status-mastered {
  border-color: #f6ad55;
  animation: masteredGlow 2.2s ease-in-out infinite;
}
@keyframes masteredGlow {
  0% {
    box-shadow: 0 0 0 0 rgba(246, 173, 85, 0.2);
  }
  50% {
    box-shadow: 0 0 14px 3px rgba(246, 173, 85, 0.35);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(246, 173, 85, 0.2);
  }
}
@keyframes floatNode {
  0% {
    transform: translate(-50%, -50%);
  }
  50% {
    transform: translate(-50%, -53%);
  }
  100% {
    transform: translate(-50%, -50%);
  }
}
.canvas-tools {
  position: absolute;
  right: 12px;
  top: 12px;
  display: flex;
  gap: 6px;
  z-index: 15;
}
.mini-map {
  position: absolute;
  right: 12px;
  bottom: 12px;
  z-index: 12;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
}
.mini-bg {
  fill: rgba(15, 23, 42, 0.78);
}
.mini-edge {
  stroke: rgba(206, 216, 255, 0.52);
  stroke-width: 1;
}
.mini-node {
  fill: #dbeafe;
}
.mini-knowledge {
  fill: #34d399;
}
.mini-problem {
  fill: #f59e0b;
}
.mini-viewport {
  fill: rgba(125, 211, 252, 0.2);
  stroke: #7dd3fc;
  stroke-width: 1.2;
}
@media (max-width: 900px) {
  .learning-map-canvas {
    min-height: 470px;
    height: 64vh;
  }
  .map-node {
    width: 132px;
  }
  .canvas-tools {
    right: 8px;
    top: 8px;
    flex-wrap: wrap;
    max-width: 220px;
    justify-content: flex-end;
  }
  .mini-map {
    right: 8px;
    bottom: 8px;
  }
}
</style>
