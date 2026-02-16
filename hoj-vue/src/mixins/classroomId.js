/**
 * Classroom ID Mixin
 * 统一处理 classroomId 的获取逻辑
 * 优先使用 prop，如果没有 prop 才从路由参数获取
 */
export default {
  computed: {
    // 计算属性，优先使用 prop，没有才从路由读取
    computedClassroomId() {
      // 优先使用 prop
      if (this.classroomId !== undefined) {
        return this.classroomId
      }
      // 其次尝试从路由参数获取
      if (this.$route && this.$route.params && this.$route.params.classroomId) {
        return this.$route.params.classroomId
      }
      // 最后尝试从路由 query 获取
      if (this.$route && this.$route.query && this.$route.query.classroomId) {
        return this.$route.query.classroomId
      }
      return null
    }
  }
}
