<template>
  <div class="markdown-body" :class="classroomMode ? 'classroom-code-display' : ''">
    <pre
      v-highlight="code"
      :style="preStyle"
    ><code :class="language" :style="codeStyle"></code></pre>
  </div>
</template>

<script>
export default {
  name: 'highlight',
  data() {
    return {
      styleObject: {
        'border-left': '3px solid ' + this.borderColor,
      },
    };
  },
  props: {
    language: {
      type: String,
    },
    code: {
      required: true,
      type: String,
    },
    borderColor: {
      type: String,
      default: '#19be6b',
    },
    classroomMode: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    preStyle() {
      const base = {
        'border-left': '3px solid ' + this.borderColor,
      };
      return base;
    },
    codeStyle() {
      return {};
    },
  },
  watch: {
    borderColor(newval, oldval) {
      if (newval != oldval) {
        this.styleObject['border-left'] = '3px solid ' + newval;
      }
    },
  },
};
</script>

<style>
</style>
