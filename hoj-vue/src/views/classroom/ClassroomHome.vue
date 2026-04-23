<template>
  <div class="classroom-home">
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'ClassroomHome'
}
</script>

<style scoped>
.classroom-home {
  min-height: 100vh;
  background-color: #f5f9fc;
}
</style>
