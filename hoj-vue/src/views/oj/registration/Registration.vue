<template>
  <div class="registration-container">
    <!-- 重定向到报名列表 -->
  </div>
</template>

<script>
export default {
  name: 'Registration',
  mounted() {
    // 重定向到新的报名系统列表页面
    this.$router.replace({ name: 'RegistrationList' })
  }
};
</script>

<style scoped>
.registration-container {
  height: 100vh;
}
</style>
