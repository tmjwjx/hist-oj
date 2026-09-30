<template>
  <div class="registration-form-card">
    <p class="registration-tip">{{ registrationTip }}</p>
    <el-form label-width="90px" @submit.native.prevent>
      <el-form-item v-if="passwordRequired" :label="$t('m.Contest_Password')" required>
        <el-input v-model="form.password" type="password" show-password></el-input>
      </el-form-item>
      <el-form-item
        v-for="field in enabledFields"
        :key="field.value"
        :label="$t('m.' + field.labelKey)"
        required
      >
        <el-select v-if="field.value === 'gender'" v-model="form.gender" style="width: 100%">
          <el-option :label="$t('m.Male')" value="男"></el-option>
          <el-option :label="$t('m.Female')" value="女"></el-option>
        </el-select>
        <el-input v-else v-model="form[field.value]" :maxlength="field.maxLength"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">{{ $t('m.Reg_Submit') }}</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
// 模块级映射表只存 i18n 键名，渲染时通过 $t('m.' + labelKey) 翻译
const FIELDS = [
  { value: "name", labelKey: "Reg_Field_Name", maxLength: 100 },
  { value: "class", labelKey: "Reg_Field_Class", maxLength: 100 },
  { value: "college", labelKey: "Reg_Field_College", maxLength: 100 },
  { value: "studentId", labelKey: "Reg_Field_Student_ID", maxLength: 50 },
  { value: "gender", labelKey: "Reg_Field_Gender", maxLength: 10 },
  { value: "qq", labelKey: "Reg_Field_QQ", maxLength: 20 },
  { value: "phone", labelKey: "Reg_Field_Phone", maxLength: 30 },
];

export default {
  name: "ContestRegistrationForm",
  props: {
    contest: { type: Object, required: true },
    loading: { type: Boolean, default: false },
  },
  data() {
    return {
      form: this.emptyForm(),
    };
  },
  computed: {
    passwordRequired() {
      return this.contest.auth !== 0;
    },
    enabledFields() {
      let values = [];
      try {
        values = JSON.parse(this.contest.registrationFields || "[]");
      } catch (e) {
        values = [];
      }
      return FIELDS.filter((field) => values.includes(field.value));
    },
    registrationTip() {
      if (this.enabledFields.length === 0 && !this.passwordRequired) {
        return this.$t("m.Reg_Tip_Simple");
      }
      return this.$t("m.Reg_Tip_Fill");
    },
  },
  methods: {
    emptyForm() {
      return { password: "", name: "", class: "", college: "", studentId: "", gender: "", qq: "", phone: "" };
    },
    submit() {
      if (this.passwordRequired && !this.form.password) {
        this.$message.warning(this.$t("m.Reg_Enter_Password"));
        return;
      }
      const missing = this.enabledFields.find((field) => !String(this.form[field.value] || "").trim());
      if (missing) {
        this.$message.warning(this.$t("m.Reg_Field_Required", { field: this.$t("m." + missing.labelKey) }));
        return;
      }
      this.$emit("submit", { ...this.form });
    },
  },
};
</script>

<style scoped>
.registration-form-card {
  margin-bottom: 15px;
}
.registration-tip {
  color: #606266;
  text-align: center;
}
</style>
