<template>
  <div class="learning-map-filter">
    <el-select size="small" v-model="status" @change="emitChange" :placeholder="$t('m.Map_Status_Filter')">
      <el-option :label="$t('m.Map_All')" value="all"></el-option>
      <el-option :label="$t('m.Map_Unlocked')" value="unlocked"></el-option>
      <el-option :label="$t('m.Map_Unfinished')" value="unfinished"></el-option>
      <el-option :label="$t('m.Map_Completed')" value="completed"></el-option>
    </el-select>

    <el-select size="small" v-model="type" @change="emitChange" :placeholder="$t('m.Map_Type_Filter')">
      <el-option :label="$t('m.Map_All_Types')" value="all"></el-option>
      <el-option :label="$t('m.Map_Only_Knowledge')" value="knowledge"></el-option>
      <el-option :label="$t('m.Map_Only_Problems')" value="problem"></el-option>
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
