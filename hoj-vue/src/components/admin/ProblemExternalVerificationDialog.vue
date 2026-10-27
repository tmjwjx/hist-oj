<template>
  <el-dialog :title="(initialTab === 'submissions' ? '外部提交 · ' : '验题链接 · ') + title" :visible="true"
    :width="initialTab === 'submissions' ? '90%' : 'min(640px, 90%)'" :close-on-click-modal="false" @close="$emit('close')">
    <ExternalSubmissionList v-if="initialTab === 'submissions'" :pid="pid" />
    <div v-else v-loading="loading" class="link-settings">
          <el-alert title="开启后，持有链接的人无需登录即可查看本题并提交代码。" description="外部提交单独记录，不计入用户题库提交统计。关闭链接后，将无法继续访问或提交。" type="info" :closable="false" show-icon />
          <el-alert v-if="error" :title="error" type="error" :closable="false" class="link-row" />
          <div v-if="link" class="link-row">
            <el-switch v-if="link.token" :value="link.enabled" :disabled="saving" active-text="允许外部验题" inactive-text="已关闭" @change="setEnabled" />
            <el-button v-else type="primary" :loading="saving" @click="setEnabled(true)">生成并开启验题链接</el-button>
          </div>
          <div v-if="link && link.token" class="link-row">
            <el-input :value="url" readonly aria-label="验题链接">
              <el-button slot="append" icon="el-icon-document-copy" @click="copyLink">复制链接</el-button>
            </el-input>
            <p class="link-note">{{ link.enabled ? '可将此链接发给验题人员；请仅分享给可信任的人。' : '链接已关闭；重新开启后仍使用此链接。' }}</p>
            <a v-if="link.enabled" :href="url" target="_blank" rel="noopener noreferrer">打开验题页面 <i class="el-icon-top-right" /></a>
          </div>
          <el-button v-if="!link && !loading" size="small" @click="loadLink">重试</el-button>
    </div>
  </el-dialog>
</template>

<script>
import api from '@/common/external-verification';
import ExternalSubmissionList from '@/components/external-verification/SubmissionList.vue';

export default {
  name: 'ProblemExternalVerificationDialog',
  components: { ExternalSubmissionList },
  props: {
    pid: { type: [Number, String], required: true },
    title: { type: String, default: '' },
    initialTab: { type: String, default: 'link' },
  },
  data() { return { link: null, loading: false, saving: false, error: '' }; },
  computed: {
    url() {
      if (!this.link || !this.link.token) return '';
      return new URL(this.$router.resolve({ name: 'ExternalVerification', params: { token: this.link.token } }).href, window.location.origin).href;
    },
  },
  mounted() { if (this.initialTab !== 'submissions') this.loadLink(); },
  methods: {
    async loadLink() {
      this.loading = true; this.error = '';
      try { this.link = await api.getLink(this.pid); }
      catch (error) { this.error = error.message || '加载验题链接失败'; }
      finally { this.loading = false; }
    },
    async setEnabled(enabled) {
      if (this.saving) return;
      this.saving = true; this.error = '';
      try {
        this.link = await api.setLink(this.pid, enabled);
        this.$message.success(enabled ? '验题链接已开启' : '验题链接已关闭');
      } catch (error) { this.error = error.message || '保存失败，请重试'; }
      finally { this.saving = false; }
    },
    async copyLink() {
      try { await this.$copyText(this.url); this.$message.success('验题链接已复制'); }
      catch (error) { this.$message.error('复制失败，请手动复制链接'); }
    },
  },
};
</script>

<style scoped>
.link-settings { max-width: 900px; min-height: 180px; padding: 8px 0 20px; }
.link-row { margin-top: 24px; }
.link-note { color: #8492a6; margin: 12px 0; }
</style>
