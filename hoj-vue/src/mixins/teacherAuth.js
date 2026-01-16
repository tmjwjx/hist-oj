import api from '@/api/classroom'

export default {
  async mounted() {
    // 等待权限验证完成
    await this.checkTeacherPermission()

    // 如果原组件有mounted方法，调用它
    if (this._originalMounted) {
      this._originalMounted()
    }
  },

  methods: {
    async checkTeacherPermission() {
      try {
        // 获取当前用户的角色信息
        const res = await api.getCurrentUserRoles()
        if (res.data.code === 200) {
          // res.data.data 是 ClassroomUserRole 对象数组，需要提取 role 字段
          const roleObjects = res.data.data || []
          const roles = roleObjects.map(r => r.role)

          // 检查是否有教师或管理员角色
          const hasPermission = roles.includes('teacher') || roles.includes('admin') || roles.includes('root')

          if (!hasPermission) {
            this.$message.error('您没有权限访问此页面')
            setTimeout(() => {
              this.$router.push('/')
            }, 1500)
          }
        }
      } catch (error) {
        console.error('权限检查失败:', error)
        // 权限检查失败时也拒绝访问
        this.$message.error('权限验证失败')
        setTimeout(() => {
          this.$router.push('/')
        }, 1500)
      }
    }
  }
}
