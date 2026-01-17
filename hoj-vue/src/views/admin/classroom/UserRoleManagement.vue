<template>
  <div class="user-role-management-container">
    <el-card>
      <div slot="header" class="header">
        <span class="title">{{ $t('m.User_Role_Management') }}</span>
        <el-button type="primary" size="small" icon="el-icon-refresh" @click="loadUsers">
          {{ $t('m.Refresh') }}
        </el-button>
      </div>

      <!-- 搜索栏 -->
      <div class="filter-row">
        <el-input
          v-model="keyword"
          :placeholder="$t('m.Enter_keyword')"
          prefix-icon="el-icon-search"
          style="width: 300px; margin-right: 10px;"
          @keyup.enter.native="handleSearch"
          clearable
        >
        </el-input>
        <el-button type="primary" @click="handleSearch">{{ $t('m.Search') }}</el-button>
      </div>

      <el-table
        :data="userList"
        v-loading="loading"
        stripe
        border
        style="margin-top: 20px;"
      >
        <el-table-column prop="uid" label="ID" width="80" />
        <el-table-column :label="$t('m.Username')" min-width="150">
          <template slot-scope="{ row }">
            <UserName :username="row.username" />
            <el-tag v-if="row.titleName" :color="row.titleColor" size="small" style="margin-left: 5px;">
              {{ row.titleName }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="realname" :label="$t('m.RealName')" min-width="120" />
        <el-table-column prop="email" :label="$t('m.Email')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="$t('m.User_Role_Management')" min-width="200">
          <template slot-scope="{ row }">
            <el-radio-group v-model="row.role" @change="handleRoleChange(row)" size="small">
              <el-radio label="">{{ $t('m.None') }}</el-radio>
              <el-radio label="teacher">{{ $t('m.Teacher') }}</el-radio>
              <el-radio label="student">{{ $t('m.Student') }}</el-radio>
            </el-radio-group>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Status')" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
              {{ row.status === 0 ? $t('m.Normal') : $t('m.Disable') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Operation')" width="120" fixed="right">
          <template slot-scope="{ row }">
            <el-button
              v-if="hasRoleChanged(row)"
              size="mini"
              type="primary"
              @click="saveUserRoles(row)"
              :loading="row.saving"
            >
              {{ $t('m.Save') }}
            </el-button>
            <el-tag v-else type="info" size="small">{{ $t('m.No_Change') }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="page"
        layout="prev, pager, next, sizes, total"
        :total="total"
        :page-size="pageSize"
        :current-page.sync="currentPage"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      >
      </el-pagination>
    </el-card>

    <!-- 说明提示 -->
    <el-card style="margin-top: 20px;">
      <div slot="header">
        <span>{{ $t('m.Admin_Enable_Classroom') }}</span>
      </div>
      <el-alert
        :title="$t('m.How_To_Use_Role_Management')"
        type="info"
        :closable="false"
        show-icon
      >
        <template>
          <div style="margin-top: 10px; line-height: 1.8;">
            <p><strong>1. {{ $t('m.Assign_Roles') }}:</strong> {{ $t('m.Assign_Roles_Desc') }}</p>
            <p><strong>2. {{ $t('m.Teacher_Role') }}:</strong> {{ $t('m.Teacher_Role_Desc') }}</p>
            <p><strong>3. {{ $t('m.Student_Role') }}:</strong> {{ $t('m.Student_Role_Desc') }}</p>
            <p><strong>4. {{ $t('m.Save_Changes') }}:</strong> {{ $t('m.Save_Changes_Desc') }}</p>
          </div>
        </template>
      </el-alert>
    </el-card>
  </div>
</template>

<script>
import api from '@/common/api'
import classroomApi from '@/api/classroom'
import UserName from '@/components/oj/common/UserName.vue'

export default {
  name: 'UserRoleManagement',
  components: {
    UserName
  },
  data() {
    return {
      loading: false,
      userList: [],
      total: 0,
      currentPage: 1,
      pageSize: 20,
      keyword: '',
      originalRoles: new Map() // 存储用户的原始角色，用于比较是否有变更
    }
  },
  mounted() {
    this.loadUsers()
  },
  methods: {
    async loadUsers() {
      this.loading = true
      try {
        const res = await api.admin_getUserList(this.currentPage, this.pageSize, this.keyword, false)
        if (res.data.data) {
          const users = res.data.data.records

          // 为每个用户获取其在 classroom_user_role 表中的角色
          const usersWithRoles = await Promise.all(
            users.map(async (user) => {
              try {
                const roleRes = await classroomApi.getUserRoles(user.uid)
                let role = ''

                console.log('getUserRoles response for', user.uid, ':', roleRes.data)

                // 处理标准格式: { code: 200, data: [...] }
                if (roleRes.data && roleRes.data.code === 200 && roleRes.data.data) {
                  const roles = roleRes.data.data || []
                  const roleNames = roles.map(item => item.role)
                  // 转换为单个角色（取第一个）
                  role = roleNames.length > 0 ? roleNames[0] : ''
                  console.log('Extracted role for', user.uid, ':', role)
                }

                return {
                  ...user,
                  role: role, // 单个角色：''、'teacher' 或 'student'
                  originalRole: role,
                  saving: false
                }
              } catch (error) {
                console.error('获取用户角色失败:', user.uid, error)
                return {
                  ...user,
                  role: '',
                  originalRole: '',
                  saving: false
                }
              }
            })
          )

          this.userList = usersWithRoles
          this.total = res.data.data.total
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.currentPage = 1
      this.loadUsers()
    },
    handlePageChange(page) {
      this.currentPage = page
      this.loadUsers()
    },
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.loadUsers()
    },
    hasRoleChanged(user) {
      return user.role !== user.originalRole
    },
    async saveUserRoles(user) {
      user.saving = true
      try {
        // 将单个角色转换为数组（空字符串表示不分配任何角色）
        const roles = user.role ? [user.role] : []

        console.log('Saving roles for user:', user.uid, 'Role:', user.role, 'Roles:', roles)

        // 调用 classroom API 更新用户角色
        const response = await classroomApi.updateUserRoles(user.uid, roles)

        if (response.data.code === 200) {
          user.originalRole = user.role
          this.$message.success(this.$t('m.Role_Save_Success'))
        } else {
          throw new Error(response.data.msg || '保存失败')
        }
      } catch (error) {
        console.error('保存用户角色失败:', error)
        this.$message.error(error.response?.data?.msg || this.$t('m.Role_Save_Failed'))
        // 恢复原始角色
        user.role = user.originalRole
      } finally {
        user.saving = false
      }
    },
    handleRoleChange(user) {
      // 角色变化时的处理，可以在这里添加实时验证
      console.log('Role changed for user:', user.username, 'New role:', user.role)
    }
  }
}
</script>

<style scoped>
.user-role-management-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: 600;
}

.filter-row {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.page {
  margin-top: 20px;
  text-align: right;
}

.el-checkbox-group {
  display: flex;
  gap: 10px;
}
</style>
