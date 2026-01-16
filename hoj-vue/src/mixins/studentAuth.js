import api from '@/api/classroom'

export default {
  async mounted() {
    // 等待权限验证完成
    await this.checkClassroomPermission()

    // 如果原组件有mounted方法，调用它
    if (this._originalMounted) {
      this._originalMounted()
    }
  },

  methods: {
    async checkClassroomPermission() {
      try {
        // 获取路由中的classroomId
        const classroomId = this.$route.params.classroomId || this.$route.query.classroomId

        // 首先检查用户是否具有学生角色
        const rolesRes = await api.getCurrentUserRoles()
        if (rolesRes.data.code === 200) {
          const roleObjects = rolesRes.data.data || []
          const roles = roleObjects.map(r => r.role)

          // 检查是否有学生角色
          const isStudent = roles.includes('student')

          // 如果用户不是学生（是教师或管理员），拒绝访问学生端页面
          if (!isStudent) {
            this.$message.error('您没有权限访问学生端页面')
            setTimeout(() => {
              this.$router.push('/classroom')
            }, 1500)
            return
          }
        }

        if (!classroomId) {
          // 如果没有classroomId参数，只检查角色就够了
          return
        }

        // 获取用户加入的班级列表
        const res = await api.getStudentClassrooms()

        if (res.data.code === 200) {
          const classrooms = res.data.data || []
          const classroomIds = classrooms.map(c => c.id.toString())

          // 检查用户是否是该班级的学生
          const isMember = classroomIds.includes(classroomId.toString())

          if (!isMember) {
            this.$message.error('您不是该班级的学生，无法访问此页面')
            setTimeout(() => {
              this.$router.push('/classroom')
            }, 1500)
          }
        }
      } catch (error) {
        console.error('班级权限检查失败:', error)
        // 权限检查失败时也拒绝访问
        this.$message.error('权限验证失败')
        setTimeout(() => {
          this.$router.push('/classroom')
        }, 1500)
      }
    }
  }
}
