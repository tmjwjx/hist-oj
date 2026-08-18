<template>
  <div class="registration-config">
    <div class="registration-title">比赛报名</div>

    <!-- 第一行：开关项 -->
    <div class="registration-switches">
      <el-form-item label="开启报名信息填写">
        <el-switch v-model="contest.openRegistration"></el-switch>
      </el-form-item>
      <el-form-item v-if="contest.openRegistration" label="替换比赛内名称">
        <el-switch v-model="contest.useRegistrationName"></el-switch>
      </el-form-item>
    </div>

    <!-- 第二行：选择器 -->
    <div class="registration-selects" v-if="contest.openRegistration">
      <el-form-item label="报名字段" required>
        <el-select v-model="contest.registrationFields" multiple collapse-tags filterable>
          <el-option v-for="field in fields" :key="field.value" :label="field.label" :value="field.value" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="contest.useRegistrationName" label="名称组合字段" required>
        <el-select v-model="contest.registrationNameFields" multiple collapse-tags>
          <el-option v-for="field in enabledFields" :key="field.value" :label="field.label" :value="field.value" />
        </el-select>
      </el-form-item>
    </div>

    <div v-if="contest.openRegistration && contest.useRegistrationName" class="config-help">
      名称按勾选顺序组合，例如：姓名 + 学号
    </div>
  </div>
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
  name: "ContestRegistrationConfig",
  props: {
    contest: { type: Object, required: true },
  },
  data() {
    return { fields: FIELDS };
  },
  computed: {
    enabledFields() {
      const enabled = this.contest.registrationFields || [];
      return this.fields.filter((field) => enabled.includes(field.value));
    },
  },
  watch: {
    "contest.registrationFields"(fields) {
      const enabled = fields || [];
      this.contest.registrationNameFields = (this.contest.registrationNameFields || [])
        .filter((field) => enabled.includes(field));
    },
    "contest.openRegistration"(open) {
      if (!open) this.contest.useRegistrationName = false;
    },
  },
};
</script>

<style scoped>
.registration-config {
  margin: 24px 0;
  padding: 20px;
  background: #f9fafb;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.registration-title {
  margin-bottom: 16px;
  color: #303133;
  font-weight: 600;
  font-size: 16px;
  line-height: 1.5;
  padding-bottom: 12px;
  border-bottom: 2px solid #409eff;
}

/* 第一行：开关项 */
.registration-switches {
  display: flex;
  gap: 40px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.registration-switches .el-form-item {
  margin-bottom: 0;
  flex: 0 0 auto;
}

/* 第二行：选择器 */
.registration-selects {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 12px;
}

.registration-selects .el-form-item {
  margin-bottom: 0;
}

.registration-selects .el-select {
  width: 100%;
}

.config-help {
  margin-top: 12px;
  padding: 8px 12px;
  background: #ecf5ff;
  border: 1px solid #d9ecff;
  border-radius: 4px;
  color: #409eff;
  font-size: 13px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .registration-config {
    padding: 16px;
  }

  .registration-switches {
    flex-direction: column;
    gap: 16px;
  }

  .registration-selects {
    grid-template-columns: 1fr;
  }
}
</style>
