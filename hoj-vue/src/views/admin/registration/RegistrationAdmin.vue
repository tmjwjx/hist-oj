<template>
  <div class="registration-admin-container"></div>
</template>

<script>
export default {
  name: 'RegistrationAdmin',
  mounted() {
    // 直接跳转到报名系统管理页面
    window.location.href = '/registration/admin.html';
  }
};
</script>

<style scoped>
.registration-admin-container {
  height: 100vh;
}
</style>
