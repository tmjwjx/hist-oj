<template>
  <div class="registration-form-card">
    <p class="registration-tip">{{ registrationTip }}</p>
    <el-form label-width="90px" @submit.native.prevent>
      <el-form-item v-if="passwordRequired" label="比赛密码" required>
        <el-input v-model="form.password" type="password" show-password></el-input>
      </el-form-item>
      <el-form-item
        v-for="field in enabledFields"
        :key="field.value"
        :label="field.label"
        required
      >
        <el-select v-if="field.value === 'gender'" v-model="form.gender" style="width: 100%">
          <el-option label="男" value="男"></el-option>
          <el-option label="女" value="女"></el-option>
        </el-select>
        <el-input v-else v-model="form[field.value]" :maxlength="field.maxLength"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">提交报名</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
const FIELDS = [
  { value: "name", label: "姓名", maxLength: 100 },
  { value: "class", label: "班级", maxLength: 100 },
  { value: "college", label: "学院", maxLength: 100 },
  { value: "studentId", label: "学号", maxLength: 50 },
  { value: "gender", label: "性别", maxLength: 10 },
  { value: "qq", label: "QQ", maxLength: 20 },
  { value: "phone", label: "电话号码", maxLength: 30 },
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
        return "点击提交即可完成报名。";
      }
      return "请填写报名所需信息，提交后完成报名。";
    },
  },
  methods: {
    emptyForm() {
      return { password: "", name: "", class: "", college: "", studentId: "", gender: "", qq: "", phone: "" };
    },
    submit() {
      if (this.passwordRequired && !this.form.password) {
        this.$message.warning("请输入比赛密码");
        return;
      }
      const missing = this.enabledFields.find((field) => !String(this.form[field.value] || "").trim());
      if (missing) {
        this.$message.warning(`请填写${missing.label}`);
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
