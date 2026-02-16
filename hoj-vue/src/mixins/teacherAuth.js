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
        // 获取当前用户的角色信息
        const res = await api.getCurrentUserRoles()

        if (res.data.code === 200) {
          // res.data.data 是 ClassroomUserRole 对象数组，需要提取 role 字段
          const roleObjects = res.data.data || []
          const roles = roleObjects.map(r => r.role)

          // 检查是否有教师或管理员角色
          const hasPermission = roles.includes('teacher') || roles.includes('admin') || roles.includes('root') || roles.includes('problem_admin')

          if (hasPermission) {
            // 有权限，检查是否是重试调用
            if (tryCount > 0) {
              console.log(`[teacherAuth] 第 ${tryCount} 次重试，权限验证通过`)
            } else {
              // 第一次调用
              console.log('[teacherAuth] 权限检查通过')
            }
            return true
          }
        } else {
          // 没有权限
          // 如果还有重试次数，尝试重试
          if (tryCount < maxRetries) {
            console.warn(`[teacherAuth] 权限验证失败，${retryDelay}ms 后第 ${tryCount + 1} 次重试`)
            await new Promise(resolve => setTimeout(resolve, retryDelay))
            return this.checkTeacherPermission(tryCount + 1)
          } else {
            // 重试用尽，仍然失败
            console.error('[teacherAuth] 权限验证失败，已重试 ' + maxRetries + ' 次')
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
        if (tryCount < maxRetries) {
          console.error('[teacherAuth] 权限检查异常:', error)
          console.warn(`[teacherAuth] ${retryDelay}ms 后第 ${tryCount + 1} 次重试`)
          await new Promise(resolve => setTimeout(resolve, retryDelay))
          return this.checkTeacherPermission(tryCount + 1)
        } else {
          console.error('[teacherAuth] 权限验证异常，已重试 ' + maxRetries + ' 次')
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