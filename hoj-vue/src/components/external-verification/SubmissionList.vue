<template>
  <section class="external-submissions">
    <div class="list-header">
      <span>{{ pid ? '外部提交' : '我的外部提交' }} <small v-if="!pid">仅显示当前浏览器在本链接下的提交</small></span>
      <el-button size="small" icon="el-icon-refresh" :loading="loading" @click="refresh()">刷新</el-button>
    </div>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <el-table :data="records" v-loading="loading" stripe empty-text="暂无外部提交" @row-click="showDetail">
      <el-table-column prop="submitId" label="提交编号" min-width="100" />
      <el-table-column prop="nickname" label="昵称" min-width="100">
        <template slot-scope="scope">{{ scope.row.nickname || '匿名访客' }}</template>
      </el-table-column>
      <el-table-column prop="language" label="语言" min-width="110" />
      <el-table-column label="结果" min-width="170">
        <template slot-scope="scope"><el-tag size="small" :type="statusType(scope.row)">{{ statusText(scope.row) }}</el-tag></template>
      </el-table-column>
      <el-table-column label="用时" min-width="85"><template slot-scope="scope">{{ measurement(scope.row.time, 'ms') }}</template></el-table-column>
      <el-table-column label="内存" min-width="90"><template slot-scope="scope">{{ measurement(scope.row.memory, 'KB') }}</template></el-table-column>
      <el-table-column prop="score" label="得分" width="70" />
      <el-table-column label="提交时间" min-width="170"><template slot-scope="scope">{{ scope.row.submitTime | localtime }}</template></el-table-column>
      <el-table-column label="操作" width="80"><template slot-scope="scope"><el-button type="text" @click.stop="showDetail(scope.row)">详情</el-button></template></el-table-column>
    </el-table>
    <el-pagination class="pagination" layout="total, prev, pager, next" :page-size="limit" :total="total" :current-page="currentPage" @current-change="changePage" />
    <el-dialog title="外部提交详情" :visible.sync="detailVisible" width="90%" append-to-body @closed="detail = null">
      <div v-loading="detailLoading">
        <el-alert v-if="detailError" :title="detailError" type="error" :closable="false" />
        <template v-if="detail">
          <div class="detail-meta">
            <strong>#{{ detail.submitId }}</strong>
            <el-tag :type="statusType(detail)">{{ statusText(detail) }}</el-tag>
            <span>{{ detail.nickname || '匿名访客' }} · {{ detail.language }}</span>
            <span>{{ measurement(detail.time, 'ms') }} / {{ measurement(detail.memory, 'KB') }}</span>
            <span v-if="detail.score != null">得分：{{ detail.score }}</span>
          </div>
          <template v-if="detail.errorMessage"><h3>评测信息</h3><pre>{{ detail.errorMessage }}</pre></template>
          <h3>提交代码</h3>
          <pre v-if="detail.code" :key="detail.submitId" v-highlight="detail.code" class="submission-code"><code :class="codeLanguage">{{ detail.code }}</code></pre>
          <template v-if="detail.caseResults && detail.caseResults.length">
            <h3>测试点结果</h3>
            <el-table :data="detail.caseResults" stripe max-height="400">
              <el-table-column label="测试点" width="90"><template slot-scope="scope">{{ scope.row.seq || scope.$index + 1 }}</template></el-table-column>
              <el-table-column label="结果" min-width="160"><template slot-scope="scope"><el-tag size="small" :type="statusType(scope.row)">{{ statusText(scope.row) }}</el-tag></template></el-table-column>
              <el-table-column prop="time" label="用时 (ms)" />
              <el-table-column prop="memory" label="内存 (KB)" />
              <el-table-column prop="score" label="得分" />
            </el-table>
          </template>
        </template>
      </div>
    </el-dialog>
  </section>
</template>

<script>
import api from '@/common/external-verification';
import { JUDGE_STATUS } from '@/common/constants';
import hljs from 'highlight.js';

export default {
  name: 'ExternalSubmissionList',
  props: {
    pid: { type: [Number, String], default: null },
    token: { type: String, default: '' },
    visitorToken: { type: String, default: '' },
  },
  data() {
    return { records: [], total: 0, currentPage: 1, limit: 10, loading: false, error: '', timer: null,
      detail: null, detailVisible: false, detailLoading: false, detailError: '', detailRequest: 0, active: true, refreshQueued: false };
  },
  computed: {
    codeLanguage() {
      const name = String(this.detail && this.detail.language || '').toLowerCase();
      if (name.includes('++')) return 'cpp';
      if (/python|pypy/.test(name)) return 'python';
      if (/javascript|node/.test(name)) return 'javascript';
      if (name.includes('c#')) return 'csharp';
      const language = name.split(/[\s(]/)[0];
      return hljs.getLanguage(language) ? language : 'plaintext';
    },
  },
  mounted() { this.refresh(); },
  beforeDestroy() { this.active = false; clearTimeout(this.timer); },
  methods: {
    context() { return { pid: this.pid, token: this.token, visitorToken: this.visitorToken }; },
    running(row) { return row && [5, 6, 7, 9].includes(Number(row.status)); },
    statusText(row) {
      if (row.statusText) return row.statusText;
      if (Number(row.status) === 7 && row.currentTest) return 'Running on test ' + row.currentTest;
      return (JUDGE_STATUS[row.status] || {}).name || String(row.status == null ? '等待结果' : row.status);
    },
    statusType(row) {
      const type = (JUDGE_STATUS[row.status] || {}).type;
      return type === 'error' ? 'danger' : type == null ? 'info' : type;
    },
    measurement(value, unit) { return value == null ? '—' : value + ' ' + unit; },
    fail(error) {
      if (!this.pid && [401, 403, 404, 410].includes(Number(error.status))) {
        this.records = []; this.detail = null; this.detailVisible = false;
        this.$emit('access-error', error);
      }
    },
    async refresh(reset = false) {
      if (!this.active) return;
      if (this.loading) { if (reset) this.refreshQueued = true; return; }
      clearTimeout(this.timer);
      if (reset) this.currentPage = 1;
      this.loading = true; this.error = '';
      try {
        const result = await api.getSubmissions({ ...this.context(), currentPage: this.currentPage, limit: this.limit });
        if (!this.active) return;
        this.records = result.records || []; this.total = result.total || 0;
        if (this.detailVisible && this.detail && this.running(this.detail)) await this.showDetail(this.detail, true);
        if (!this.active) return;
        const pending = this.records.some(this.running) || (this.detailVisible && this.running(this.detail));
        // Keep anonymous pages aware of a link being disabled even after judging finishes.
        if (pending || !this.pid) this.timer = setTimeout(() => this.refresh(), pending ? 2000 : 15000);
      } catch (error) {
        if (!this.active) return;
        this.error = error.message || '加载提交失败，请刷新重试'; this.fail(error);
      } finally {
        this.loading = false;
        if (this.active && this.refreshQueued) {
          this.refreshQueued = false; this.$nextTick(() => this.refresh(true));
        }
      }
    },
    changePage(page) { if (this.loading) return; this.currentPage = page; this.refresh(); },
    async showDetail(row, silent = false) {
      const request = ++this.detailRequest;
      this.detailVisible = true; this.detailError = '';
      if (!silent) { this.detail = null; this.detailLoading = true; }
      try {
        const detail = await api.getSubmission({ ...this.context(), submitId: row.submitId });
        if (this.active && request === this.detailRequest && this.detailVisible) {
          this.detail = detail;
          if (!this.loading && this.running(detail)) {
            clearTimeout(this.timer); this.timer = setTimeout(() => this.refresh(), 2000);
          }
        }
      } catch (error) {
        if (!this.active || request !== this.detailRequest) return;
        this.detailError = error.message || '加载详情失败'; this.fail(error);
      } finally { if (request === this.detailRequest) this.detailLoading = false; }
    },
  },
};
</script>

<style scoped>
.list-header, .detail-meta { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 16px; }
.list-header { justify-content: space-between; font-size: 16px; }
.list-header small { display: inline-block; color: #8492a6; font-size: 12px; }
.pagination { margin-top: 18px; text-align: right; }
pre { background: #f5f7fa; padding: 16px; border-radius: 6px; white-space: pre-wrap; overflow-wrap: anywhere; max-height: 480px; overflow: auto; }
.submission-code { white-space: pre; }
.submission-code code { background: transparent; padding: 0; font-family: Menlo, Monaco, Consolas, "Courier New", monospace; }
h3 { font-size: 15px; margin: 22px 0 12px; }
</style>
