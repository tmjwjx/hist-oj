<template>
  <el-dialog title="上次通过的标准程序" :visible.sync="visibleProxy" width="760px" append-to-body>
    <div v-loading="loading">
      <el-alert v-if="loaded && !record.code" title="该题暂无通过的标准程序" type="info" :closable="false" show-icon />
      <template v-else-if="record.code">
        <div class="code-meta">
          <el-tag type="success">{{ record.language }}</el-tag>
          <span>提交 #{{ record.submitId }}</span>
        </div>
        <pre class="code-preview">{{ record.code }}</pre>
      </template>
    </div>
    <span slot="footer">
      <el-button @click="visibleProxy = false">关闭</el-button>
      <el-button type="primary" :disabled="!record.code" @click="useCode">填入标准程序</el-button>
    </span>
  </el-dialog>
</template>

<script>
import api from '@/common/api'

export default {
  name: 'LastPassedCodeDialog',
  props: { visible: Boolean, pid: { type: [Number, String], required: true } },
  data() { return { record: {}, loading: false, loaded: false } },
  computed: {
    visibleProxy: {
      get() { return this.visible },
      set(value) { if (!value) this.$emit('close') }
    }
  },
  watch: { visible(value) { if (value) this.load() } },
  mounted() { if (this.visible) this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await api.admin_getLastPassedVerificationCode(this.pid)
        this.record = res.data.data || {}
        this.loaded = true
      } finally { this.loading = false }
    },
    useCode() {
      this.$emit('use', { code: this.record.code, language: this.record.language })
      this.visibleProxy = false
    }
  }
}
</script>

<style scoped>
.code-meta { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; color: #909399; }
.code-preview { max-height: 480px; overflow: auto; padding: 14px; background: #f6f8fa; border-radius: 4px; white-space: pre-wrap; word-break: break-word; }
</style>
