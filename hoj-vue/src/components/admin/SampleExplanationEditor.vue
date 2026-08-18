<template>
  <div class="sample-explanation-editor">
    <div class="editor-toolbar">
      <span class="markdown-hint">
        <i class="el-icon-info"></i>
        支持 Markdown、代码块、表格与公式
      </span>
      <el-radio-group v-model="mode" size="mini">
        <el-radio-button label="edit">编辑</el-radio-button>
        <el-radio-button label="preview">预览</el-radio-button>
      </el-radio-group>
    </div>

    <el-input
      v-if="mode === 'edit'"
      v-model="content"
      type="textarea"
      :autosize="{ minRows: 4, maxRows: 12 }"
      placeholder="说明样例输入如何得到样例输出，可使用 Markdown，可留空"
    />
    <div v-else class="markdown-preview">
      <Markdown
        v-if="content.trim()"
        :content="content"
        :isAvoidXss="true"
      />
      <span v-else class="empty-preview">暂无样例解释</span>
    </div>
  </div>
</template>

<script>
import Markdown from "@/components/oj/common/Markdown.vue";

export default {
  name: "SampleExplanationEditor",
  components: { Markdown },
  props: {
    value: {
      type: String,
      default: "",
    },
  },
  data() {
    return { mode: "edit" };
  },
  computed: {
    content: {
      get() {
        return this.value || "";
      },
      set(value) {
        this.$emit("input", value);
      },
    },
  },
};
</script>

<style scoped>
.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.markdown-hint,
.empty-preview {
  color: #909399;
  font-size: 13px;
}

.markdown-preview {
  min-height: 92px;
  padding: 12px 14px;
  overflow: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: #fafafa;
}

.markdown-preview /deep/ .markdown-body {
  margin: 0;
}

@media (max-width: 640px) {
  .editor-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
