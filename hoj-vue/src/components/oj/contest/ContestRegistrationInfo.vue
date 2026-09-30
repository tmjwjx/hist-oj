<template>
  <el-card class="registration-info-card" shadow="never">
    <div slot="header" class="registration-info-header">
      <span><i class="el-icon-document-checked"></i> {{ $t('m.My_Registration_Info') }}</span>
      <el-button type="text" :loading="loading" @click="$emit('refresh')">{{ $t('m.Refresh') }}</el-button>
    </div>
    <div class="registration-info-grid">
      <div class="registration-info-item">
        <span class="registration-info-label">{{ $t('m.Reg_OJ_Username') }}</span>
        <span>{{ registration.username || registration.uid }}</span>
      </div>
      <div class="registration-info-item">
        <span class="registration-info-label">{{ $t('m.Reg_Time') }}</span>
        <span>{{ registration.gmtCreate | localtime }}</span>
      </div>
      <div
        v-for="field in enabledFields"
        :key="field.value"
        class="registration-info-item"
      >
        <span class="registration-info-label">{{ $t('m.' + field.labelKey) }}</span>
        <span>{{ registration[field.value] || '—' }}</span>
      </div>
    </div>
    <div v-if="enabledFields.length === 0" class="registration-info-empty">
      {{ $t('m.Reg_No_Extra_Fields') }}
    </div>
  </el-card>
</template>

<script>
// 模块级映射表只存 i18n 键名，渲染时通过 $t('m.' + labelKey) 翻译
const FIELDS = [
  { value: "name", labelKey: "Reg_Field_Name" },
  { value: "class", labelKey: "Reg_Field_Class" },
  { value: "college", labelKey: "Reg_Field_College" },
  { value: "studentId", labelKey: "Reg_Field_Student_ID" },
  { value: "gender", labelKey: "Reg_Field_Gender" },
  { value: "qq", labelKey: "Reg_Field_QQ" },
  { value: "phone", labelKey: "Reg_Field_Phone" },
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
