import api from '@/api/classroom'

export default {
  // 在组件创建时检查是否需要跳过权限验证
  // 管理员页面会在 $options 上设置 _skipTeacherAuth 标志
  beforeCreate() {
    // 检查组件选项中是否设置了跳过权限验证的标志
    const skipAuth = this.$options._skipTeacherAuth ||
                     this.$parent?.$options._skipTeacherAuth ||
                     false

    if (skipAuth) {
      // 设置跳过标志，mounted 钩子会检查这个标志
      this._skipTeacherAuthLocal = true
    }
  },

  async mounted() {
    // 检查是否跳过权限验证
    if (this._skipTeacherAuthLocal) {
      // 跳过权限验证，直接返回
      return
    }

    // 等待权限验证完成，第一次调用时 tryCount 为 0
    await this.checkTeacherPermission(0)
  },

  methods: {
    async checkTeacherPermission(retryCount = 0) {
      const maxRetries = 3
      const retryDelay = 200 // 每次重试间隔 200ms

      try {
        // 使用store中的缓存角色数据，避免重复请求
        await this.$store.dispatch('classroom/loadUserRoles')

        const roles = this.$store.state.classroom.userRoles || []

        // 检查是否有教师或管理员角色
        const hasPermission = roles.includes('teacher') || roles.includes('admin') || roles.includes('root') || roles.includes('problem_admin')

        if (hasPermission) {
          // 有权限
          return true
        } else {
          // 没有权限
          // 如果还有重试次数，尝试重试
          if (retryCount < maxRetries) {
            await new Promise(resolve => setTimeout(resolve, retryDelay))
            return this.checkTeacherPermission(retryCount + 1)
          } else {
            // 重试用尽，仍然失败
            this.$message.error('您没有权限访问此页面')
            setTimeout(() => {
              // 跳转到教师 Dashboard 而不是根路径
              this.$router.push({ name: 'TeacherDashboard' })
            }, 1500)
            return false
          }
        }
      } catch (error) {
        // 如果是网络错误或重试失败，也尝试重试
        if (retryCount < maxRetries) {
          await new Promise(resolve => setTimeout(resolve, retryDelay))
          return this.checkTeacherPermission(retryCount + 1)
        } else {
          console.error('权限验证失败', error)
          this.$message.error('权限验证失败')
          setTimeout(() => {
            this.$router.push('/')
          }, 1500)
          return false
        }
      }
    }
  }
}