<template>
  <div class="classroom-code-viewer markdown-body">
    <div class="classroom-code-box">
      <ol class="code-line-numbers" aria-hidden="true">
        <li v-for="line in lineNumbers" :key="line"></li>
      </ol>
      <pre v-highlight="code" class="classroom-code-pre"><code :class="language"></code></pre>
      <button type="button" class="code-copy" title="copy" @click="copyCode">
        <i class="el-icon-document-copy"></i> COPY
      </button>
    </div>
  </div>
</template>

<script>
import 'highlight.js/styles/atom-one-dark.css'

export default {
  name: 'ClassroomCodeViewer',
  props: {
    code: {
      type: String,
      default: ''
    },
    language: {
      type: String,
      default: 'plaintext'
    }
  },
  computed: {
    lineNumbers() {
      return String(this.code || '').split('\n').map((_, index) => index + 1)
    }
  },
  methods: {
    async copyCode() {
      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(this.code)
        } else {
          this.copyCodeWithTextarea()
        }
        this.$message.success('复制成功')
      } catch (e) {
        if (this.copyCodeWithTextarea()) {
          this.$message.success('复制成功')
        } else {
          this.$message.error('复制失败')
        }
      }
    },
    copyCodeWithTextarea() {
      const textarea = document.createElement('textarea')
      textarea.value = this.code
      textarea.setAttribute('readonly', 'readonly')
      textarea.style.position = 'fixed'
      textarea.style.left = '-9999px'
      document.body.appendChild(textarea)
      textarea.select()
      const copied = document.execCommand('Copy')
      document.body.removeChild(textarea)
      return copied
    }
  }
}
</script>

<style scoped>
.classroom-code-box {
  position: relative;
  overflow: auto;
  background: #f8f8f9;
  border: 1px dashed #e9eaec;
  border-radius: 3px;
}

.classroom-code-viewer .classroom-code-pre {
  margin: 0 !important;
  padding: 0 10px 0 40px !important;
  overflow: visible !important;
  white-space: pre !important;
  background: transparent !important;
  border: 0 !important;
  border-radius: 0 !important;
}

.classroom-code-viewer .classroom-code-pre code,
.classroom-code-viewer .classroom-code-pre code.hljs {
  display: block;
  margin: 0 !important;
  padding: 0 16px 0 0 !important;
  text-indent: 0 !important;
  white-space: pre !important;
  line-height: 26px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.code-line-numbers {
  position: absolute;
  top: 0;
  left: 0;
  width: 40px;
  min-height: 100%;
  margin: 0;
  padding: 0;
  list-style: none;
  counter-reset: code-line;
  background: #f1f1f1;
  color: #777;
  font-size: 1rem;
  line-height: 26px;
}

.code-line-numbers li {
  height: 26px;
  line-height: 26px;
  margin: 0 !important;
  padding: 0 !important;
  counter-increment: code-line;
}

.code-line-numbers li::before {
  content: counter(code-line);
  display: inline-block;
  width: 40px;
  text-align: center;
  vertical-align: top;
}

.code-copy {
  position: absolute;
  top: 5px;
  right: 5px;
  display: none;
  padding: 5px;
  border: 0;
  border-radius: 3px;
  background-color: #2196f3;
  color: #fff;
  font-size: 11px;
  cursor: pointer;
}

.classroom-code-box:hover .code-copy {
  display: block;
}
</style>
