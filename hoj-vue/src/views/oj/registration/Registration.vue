<template>
  <div class="registration-container"></div>
</template>

<script>
export default {
  name: 'Registration',
  mounted() {
    // 直接跳转到报名系统页面
    window.location.href = '/registration/registration.html';
  }
};
</script>

<style scoped>
.registration-container {
  height: 100vh;
}
</style>
