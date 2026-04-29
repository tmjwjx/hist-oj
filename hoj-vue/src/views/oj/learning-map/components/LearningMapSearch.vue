<template>
  <div class="learning-map-search">
    <el-input
      size="small"
      clearable
      placeholder="搜索知识点或题目"
      v-model="keyword"
      @input="onInput"
      @keyup.enter.native="emitSearch"
    >
      <el-button slot="append" icon="el-icon-search" @click="emitSearch"></el-button>
    </el-input>

    <div class="result-box" v-if="showResult">
      <el-scrollbar style="max-height: 230px;">
        <div
          class="result-item"
          v-for="item in results"
          :key="item.id"
          @click="$emit('select', item)"
        >
          <span class="type" :class="item.type">{{ nodeTypeLabel(item.type) }}</span>
          <span class="title">{{ item.title }}</span>
        </div>
        <div class="empty" v-if="!loading && results.length === 0">无匹配结果</div>
      </el-scrollbar>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LearningMapSearch',
  props: {
    results: {
      type: Array,
      default: () => []
    },
    loading: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      keyword: ''
    }
  },
  computed: {
    showResult() {
      return this.keyword.trim() !== ''
    }
  },
  methods: {
    nodeTypeLabel(type) {
      return type === 'knowledge' ? '知识点' : '题目'
    },
    onInput() {
      this.emitSearch()
    },
    emitSearch() {
      this.$emit('search', this.keyword.trim())
    }
  }
}
</script>

<style scoped>
.learning-map-search {
  position: relative;
  width: 320px;
  max-width: 100%;
}
.result-box {
  position: absolute;
  left: 0;
  right: 0;
  top: 36px;
  z-index: 30;
  background: #fff;
  border: 1px solid #dbe2ef;
  border-radius: 8px;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.16);
  overflow: hidden;
}
.result-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  cursor: pointer;
}
.result-item:hover {
  background: #f3f8ff;
}
.type {
  display: inline-block;
  min-width: 72px;
  font-size: 11px;
  border-radius: 10px;
  padding: 2px 8px;
  text-align: center;
  color: #fff;
}
.type.knowledge {
  background: #10b981;
}
.type.problem {
  background: #f59e0b;
}
.title {
  font-size: 13px;
  color: #1f2937;
}
.empty {
  padding: 10px;
  text-align: center;
  color: #94a3b8;
  font-size: 12px;
}
</style>
