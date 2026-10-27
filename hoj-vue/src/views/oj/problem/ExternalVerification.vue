<template>
  <main class="external-verification" v-loading="loading">
    <header class="page-header"><span class="eyebrow">外部验题</span><span class="anonymous-note">无需登录 · 提交单独记录</span></header>
    <el-card v-if="!problem" class="unavailable">
      <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
      <p v-else>正在加载题目…</p>
      <el-button v-if="error" class="retry-button" @click="load">重新加载</el-button>
    </el-card>
    <el-tabs v-if="problem" v-model="activeTab" type="border-card" class="problem-tabs">
      <el-tab-pane name="statement">
        <span slot="label"><i class="fa fa-list-alt"></i> 题目描述</span>
          <article class="statement">
            <div class="problem-heading">
              <h1>{{ problem.title }}</h1>
              <div class="problem-meta"><span>{{ problem.problemId }}</span><el-tag size="small" effect="plain">{{ Number(problem.type) === 1 ? 'OI' : 'ACM' }}</el-tag></div>
              <div class="limits"><div>时间限制：C/C++ {{ problem.timeLimit }}MS，其他语言 {{ problem.timeLimit * 2 }}MS</div><div>内存限制：C/C++ {{ problem.memoryLimit }}MB，其他语言 {{ problem.memoryLimit * 2 }}MB</div><div>判题模式：{{ judgeModeText }}</div></div>
              <div v-if="problem.isFileIO" class="file-io-note"><el-tag type="warning" size="small">文件输入输出</el-tag><span>输入文件：{{ problem.ioReadFileName }}</span><span>输出文件：{{ problem.ioWriteFileName }}</span></div>
            </div>
            <div class="problem-content">
              <section><h2>题目描述</h2><Markdown :content="problem.description || ''" :is-avoid-xss="true" /></section>
              <section><h2>输入格式</h2><Markdown :content="problem.input || ''" :is-avoid-xss="true" /></section>
              <section><h2>输出格式</h2><Markdown :content="problem.output || ''" :is-avoid-xss="true" /></section>
              <section v-for="(sample, index) in samples" :key="index">
                <div class="sample-grid"><div><h2>样例输入 {{ index + 1 }} <el-button type="text" icon="el-icon-document-copy" aria-label="复制样例输入" @click="copySample(sample.input)" /></h2><pre>{{ sample.input }}</pre></div><div><h2>样例输出 {{ index + 1 }} <el-button type="text" icon="el-icon-document-copy" aria-label="复制样例输出" @click="copySample(sample.output)" /></h2><pre>{{ sample.output }}</pre></div></div>
                <div v-if="sample.explanation" class="sample-explanation"><h3>样例解释 {{ index + 1 }}</h3><Markdown :content="sample.explanation" :is-avoid-xss="true" /></div>
              </section>
              <section v-if="problem.hint"><h2>提示</h2><Markdown :content="problem.hint" :is-avoid-xss="true" /></section>
            </div>
          </article>
      </el-tab-pane>
      <el-tab-pane name="submit">
        <span slot="label"><i class="el-icon-edit-outline"></i> 提交代码</span>
        <div class="problem-submit-page is-embedded">
          <el-card shadow="never" class="submit-card">
            <el-alert v-if="identityError" :title="identityError" type="warning" :closable="false" show-icon />
            <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
            <el-form label-position="left" label-width="92px" class="submit-form" @submit.native.prevent="submit">
              <el-form-item label="题目"><el-input :value="problemLabel" class="form-control" disabled /></el-form-item>
              <el-form-item label="语言" required><el-select v-model="language" class="form-control" placeholder="请选择语言"><el-option v-for="item in languages" :key="item.name" :label="item.name" :value="item.name" /></el-select></el-form-item>
              <el-form-item label="昵称（选填）"><el-input v-model="nickname" class="form-control" maxlength="40" show-word-limit placeholder="便于管理员识别您的提交" /></el-form-item>
              <el-form-item label="源代码" required>
                <div class="source-toolbar">
                  <el-button size="small" icon="el-icon-upload2" @click="$refs.sourceFileInput.click()">选择文件</el-button>
                  <span v-if="sourceFileName" class="source-file-name">已加载文件：{{ sourceFileName }}</span>
                  <input ref="sourceFileInput" type="file" class="source-file-input" @change="readSourceFile" />
                </div>
                <el-input v-model="code" type="textarea" :autosize="{ minRows: 22, maxRows: 36 }" resize="vertical" class="source-input" placeholder="请在此粘贴源代码，或选择源代码文件" @keydown.native="handleSourceKeydown" />
                <div class="source-hint">支持直接粘贴代码或上传源代码文件，代码长度不能超过 65535 字符。当前 {{ code.length }} 字符。</div>
              </el-form-item>
              <el-form-item class="submit-actions">
                <el-button type="primary" icon="el-icon-upload" :loading="submitting" :disabled="!visitorToken || !languages.length" @click="submit">{{ submitting ? '正在提交' : '提交代码' }}</el-button>
                <el-button @click="activeTab = 'statement'">取消</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </div>
      </el-tab-pane>
      <!-- Keep mounted while hidden so link revocation is checked from every tab. -->
      <el-tab-pane name="submissions">
        <span slot="label"><i class="el-icon-time"></i> 我的提交</span>
        <div class="history-content"><ExternalSubmissionList v-if="visitorToken" ref="submissions" :key="token" :token="token" :visitor-token="visitorToken" @access-error="accessError" /><el-alert v-else :title="identityError" type="warning" :closable="false" /></div>
      </el-tab-pane>
    </el-tabs>
  </main>
</template>

<script>
import api from '@/common/external-verification';
import utils from '@/common/utils';
import Markdown from '@/components/oj/common/Markdown.vue';
import ExternalSubmissionList from '@/components/external-verification/SubmissionList.vue';

export default {
  name: 'ExternalVerification',
  components: { Markdown, ExternalSubmissionList },
  data() {
    return { problem: null, languages: [], language: '', nickname: '', code: '', visitorToken: '',
      identityError: '', error: '', loading: false, submitting: false, requestId: 0, active: true,
      activeTab: 'statement', sourceFileName: '' };
  },
  computed: {
    token() { return this.$route.params.token || ''; },
    samples() {
      if (!this.problem) return [];
      return Array.isArray(this.problem.examples) ? this.problem.examples : utils.stringToExamples(this.problem.examples);
    },
    problemLabel() { return this.problem ? this.problem.problemId + ' - ' + this.problem.title : ''; },
    judgeModeText() { return ({ default: '默认模式', spj: '特殊判题', interactive: '交互判题' })[this.problem && this.problem.judgeMode] || '默认模式'; },
  },
  watch: { token() { this.problem = null; this.visitorToken = ''; this.code = ''; this.nickname = ''; this.sourceFileName = ''; this.activeTab = 'statement'; this.load(); } },
  mounted() { this.load(); },
  beforeDestroy() { this.active = false; this.requestId++; },
  methods: {
    async copySample(value) {
      try { await this.$copyText(value); this.$message.success('样例已复制'); }
      catch (error) { this.$message.error('复制失败，请手动复制'); }
    },
    readSourceFile(event) {
      const input = event.target;
      const file = input.files && input.files[0];
      if (!file) return;
      if (file.size > 262140) {
        input.value = ''; this.$message.warning('代码不能超过 65535 字符'); return;
      }
      const token = this.token;
      const reader = new FileReader();
      reader.onload = () => {
        input.value = '';
        if (!this.active || token !== this.token) return;
        const text = String(reader.result || '');
        if (text.length > 65535) this.$message.warning('代码不能超过 65535 字符');
        else { this.code = text; this.sourceFileName = file.name; }
      };
      reader.onerror = () => { input.value = ''; this.$message.error('文件读取失败，原代码已保留'); };
      reader.readAsText(file, 'UTF-8');
    },
    handleSourceKeydown(event) {
      if (event.key !== 'Tab') return;
      event.preventDefault();
      const textarea = event.target;
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      this.code = this.code.slice(0, start) + '    ' + this.code.slice(end);
      this.$nextTick(() => { textarea.selectionStart = textarea.selectionEnd = start + 4; });
    },
    initializeVisitor() {
      this.identityError = '';
      const key = 'hoj-external-visitor:' + this.token;
      let token;
      try { token = localStorage.getItem(key); } catch (error) { /* Keep a session identity if storage is unavailable. */ }
      if (!/^[a-f0-9]{64}$/.test(token || '')) {
        if (!window.crypto || !window.crypto.getRandomValues) {
          this.identityError = '当前浏览器不支持安全访客身份，请使用现代浏览器打开链接。';
          return;
        }
        const bytes = new Uint8Array(32);
        window.crypto.getRandomValues(bytes);
        token = Array.from(bytes, value => value.toString(16).padStart(2, '0')).join('');
        try { localStorage.setItem(key, token); }
        catch (error) { this.identityError = '浏览器无法保存访客身份；本次仍可提交，刷新页面后可能无法查看本次提交记录。'; }
      }
      this.visitorToken = token;
    },
    async load() {
      const requestId = ++this.requestId;
      this.loading = true; this.error = '';
      if (!/^[a-f0-9]{64}$/.test(this.token)) {
        this.problem = null; this.error = '验题链接无效，请联系题目管理员。'; this.loading = false; return;
      }
      if (!this.visitorToken) this.initializeVisitor();
      try {
        const data = await api.getProblem(this.token);
        if (!this.active || requestId !== this.requestId) return;
        this.languages = data.languages || [];
        this.language = this.languages.some(item => item.name === this.language) ? this.language : (this.languages[0] || {}).name || '';
        this.problem = data.problem;
      } catch (error) {
        if (this.active && requestId === this.requestId) { this.problem = null; this.error = error.message || '题目加载失败，请重试'; }
      } finally { if (requestId === this.requestId) this.loading = false; }
    },
    accessError(error) {
      this.problem = null; this.languages = []; this.error = error.message || '验题链接已关闭或失效，请联系题目管理员。';
    },
    async submit() {
      if (this.submitting || !this.problem || !this.visitorToken) return;
      if (!this.language) return this.$message.warning('请选择语言');
      if (!this.code.trim()) return this.$message.warning('请输入代码');
      if (this.code.length > 65535) return this.$message.warning('代码不能超过 65535 字符');
      if (this.nickname.length > 40) return this.$message.warning('昵称不能超过 40 字符');
      this.submitting = true; this.error = '';
      const token = this.token;
      try {
        await api.submit(token, { visitorToken: this.visitorToken, nickname: this.nickname.trim(), language: this.language, code: this.code });
        if (!this.active || token !== this.token) return;
        this.$message.success('提交成功，正在评测');
        this.activeTab = 'submissions';
        await this.$nextTick();
        if (this.$refs.submissions) await this.$refs.submissions.refresh(true);
      } catch (error) {
        if (!this.active || token !== this.token) return;
        if ([401, 403, 404, 410].includes(Number(error.status))) this.accessError(error);
        else this.error = error.message || '提交失败，请稍后重试';
      } finally { this.submitting = false; }
    },
  },
};
</script>

<style scoped>
.external-verification { margin: 20px auto; padding: 0 18px 28px; font-size: 14px; }
.page-header { display: flex; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 16px; }
.eyebrow { font-size: 20px; font-weight: 600; color: #303133; }
.anonymous-note { color: #8492a6; }
.problem-tabs /deep/ > .el-tabs__content { padding: 0; }
.problem-heading { padding: 24px 28px 0; }
h1 { margin: 0 0 22px; color: #303133; font-size: 30px; font-weight: 600; line-height: 1.35; text-align: center; }
.problem-meta, .file-io-note { display: flex; justify-content: center; align-items: center; flex-wrap: wrap; gap: 12px 22px; }
.problem-meta { color: #909399; }
.limits { margin: 24px 0 18px; color: #495060; font-size: 15px; line-height: 1.9; text-align: center; }
.problem-content { margin: 28px auto 0; padding: 0 28px 40px; font-size: 16px; line-height: 1.8; }
h2 { margin: 25px 0 8px; font-size: 16px; font-weight: 600; color: #3091f2; }
h3 { font-size: 15px; margin: 0 0 4px; }
.sample-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
.sample-grid > div { min-width: 0; display: flex; flex-direction: column; }
.sample-grid pre { flex: 1; }
pre { background: #f5f7fa; border: 1px solid #ebeef5; border-radius: 4px; padding: 12px; white-space: pre-wrap; overflow-wrap: anywhere; margin: 0; font-size: 14px; line-height: 1.6; }
.sample-explanation { margin: 14px 0 18px; padding: 10px 14px; border-left: 3px solid #67c23a; background: #f0f9eb; }
.sample-explanation h3 { color: #67c23a; }
.problem-submit-page.is-embedded { margin: 0; padding: 24px 28px 32px; }
.submit-card { border: 0; border-radius: 0; }
.submit-card /deep/ .el-alert { margin-bottom: 14px; }
.submit-form { max-width: 1040px; margin: 12px auto 0; }
.form-control { width: min(100%, 520px); }
.source-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; margin-bottom: 12px; }
.source-file-input { display: none; }
.source-file-name { color: #67c23a; font-size: 13px; }
.source-input /deep/ textarea { min-height: 440px !important; padding: 14px 16px; color: #202124; background: #fff; font-family: Menlo, Monaco, Consolas, "Courier New", monospace; font-size: 14px; line-height: 1.6; tab-size: 4; }
.source-hint { margin-top: 8px; color: #909399; font-size: 12px; line-height: 1.5; }
.submit-actions { margin-top: 24px; }
.history-content { padding: 24px 28px 32px; }
.retry-button { margin-top: 20px; }
@media (min-width: 992px) { .external-verification { width: 82%; max-width: 2000px; } }
@media (max-width: 768px) {
  .external-verification { margin-top: 10px; padding: 0 8px 18px; }
  .problem-submit-page.is-embedded, .history-content { padding: 14px 8px 20px; }
  .submit-card /deep/ .el-card__body { padding: 12px 8px; }
  .submit-form { margin-top: 4px; }
  .source-input /deep/ textarea { min-height: 360px !important; }
  .problem-heading { padding: 22px 18px 0; }
  .problem-content { padding: 0 18px 24px; }
  h1 { font-size: 24px; }
}
@media (max-width: 540px) {
  .page-header { align-items: flex-start; flex-direction: column; gap: 4px; }
  .sample-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
