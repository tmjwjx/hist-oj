<template>
  <div class="violation-container">
    <el-tabs v-model="activeSection" class="violation-tabs">
      <el-tab-pane label="设备异常" name="device" lazy>
        <ContestDeviceAnomaly v-if="sectionRendered.device" />
      </el-tab-pane>
      <el-tab-pane label="代码查重" name="plagiarism" lazy>
        <ContestPlagiarism v-if="sectionRendered.plagiarism" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import ContestDeviceAnomaly from './ContestDeviceAnomaly.vue'
import ContestPlagiarism from './ContestPlagiarism.vue'

export default {
  name: 'ContestViolation',
  components: {
    ContestDeviceAnomaly,
    ContestPlagiarism
  },
  data() {
    return {
      activeSection: 'device',
      sectionRendered: {
        device: true,
        plagiarism: false
      }
    }
  },
  watch: {
    activeSection(section) {
      // lazy 模式下配合 v-if，切走再切回时保留子组件状态
      this.sectionRendered[section] = true
    }
  }
}
</script>

<style scoped>
.violation-container {
  padding-top: 4px;
}

.violation-tabs >>> .el-tabs__header {
  margin-bottom: 16px;
}
</style>
