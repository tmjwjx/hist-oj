<template>
  <div class="learning-map-filter">
    <el-select size="small" v-model="status" @change="emitChange" placeholder="状态筛选">
      <el-option label="全部" value="all"></el-option>
      <el-option label="已解锁" value="unlocked"></el-option>
      <el-option label="未完成" value="unfinished"></el-option>
      <el-option label="已完成" value="completed"></el-option>
    </el-select>

    <el-select size="small" v-model="type" @change="emitChange" placeholder="类型筛选">
      <el-option label="全部类型" value="all"></el-option>
      <el-option label="只看知识点" value="knowledge"></el-option>
      <el-option label="只看题目" value="problem"></el-option>
    </el-select>
  </div>
</template>

<script>
export default {
  name: 'LearningMapFilter',
  props: {
    value: {
      type: Object,
      default: () => ({ status: 'all', type: 'all' })
    }
  },
  data() {
    return {
      status: this.value.status || 'all',
      type: this.value.type || 'all'
    }
  },
  watch: {
    value: {
      deep: true,
      handler(v) {
        this.status = v.status || 'all'
        this.type = v.type || 'all'
      }
    }
  },
  methods: {
    emitChange() {
      this.$emit('input', { status: this.status, type: this.type })
      this.$emit('change', { status: this.status, type: this.type })
    }
  }
}
</script>

<style scoped>
.learning-map-filter {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
