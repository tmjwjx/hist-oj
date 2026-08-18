<template>
  <el-card class="registration-info-card" shadow="never">
    <div slot="header" class="registration-info-header">
      <span><i class="el-icon-document-checked"></i> 我的报名信息</span>
      <el-button type="text" :loading="loading" @click="$emit('refresh')">刷新</el-button>
    </div>
    <div class="registration-info-grid">
      <div class="registration-info-item">
        <span class="registration-info-label">OJ名称</span>
        <span>{{ registration.username || registration.uid }}</span>
      </div>
      <div
        v-for="field in enabledFields"
        :key="field.value"
        class="registration-info-item"
      >
        <span class="registration-info-label">{{ field.label }}</span>
        <span>{{ registration[field.value] || '—' }}</span>
      </div>
    </div>
    <div v-if="enabledFields.length === 0" class="registration-info-empty">
      本比赛未要求填写额外报名信息。
    </div>
  </el-card>
</template>

<script>
const FIELDS = [
  { value: "name", label: "姓名" },
  { value: "class", label: "班级" },
  { value: "college", label: "学院" },
  { value: "studentId", label: "学号" },
  { value: "gender", label: "性别" },
  { value: "qq", label: "QQ" },
  { value: "phone", label: "电话号码" },
];

export default {
  name: "ContestRegistrationInfo",
  props: {
    contest: { type: Object, required: true },
    registration: { type: Object, required: true },
    loading: { type: Boolean, default: false },
  },
  computed: {
    enabledFields() {
      const configured = this.contest.registrationFields;
      let values = Array.isArray(configured) ? configured : [];
      if (typeof configured === "string") {
        try {
          values = JSON.parse(configured || "[]");
        } catch (e) {
          values = [];
        }
      }
      return FIELDS.filter((field) => values.includes(field.value));
    },
  },
};
</script>

<style scoped>
.registration-info-card {
  margin-bottom: 15px;
}
.registration-info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.registration-info-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px 24px;
}
.registration-info-item {
  display: flex;
  gap: 8px;
  min-width: 0;
  line-height: 26px;
  word-break: break-word;
}
.registration-info-label {
  color: #909399;
  flex: 0 0 auto;
}
.registration-info-empty {
  color: #909399;
}
@media (max-width: 768px) {
  .registration-info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
