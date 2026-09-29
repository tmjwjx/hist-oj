<template>
  <div class="user-role-management-container">
    <el-card>
      <div slot="header" class="header">
        <span class="title">{{ $t('m.User_Role_Management') }}</span>
        <el-button type="primary" size="small" icon="el-icon-refresh" @click="refreshCurrentTab">
          {{ $t('m.Refresh') }}
        </el-button>
      </div>

      <!-- Tab 切换 -->
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <!-- Tab 1: 用户角色管理 -->
        <el-tab-pane :label="$t('m.User_Role_Management')" name="roleManagement">
          <!-- 搜索栏 -->
          <div class="filter-row">
            <el-input
              v-model="keyword"
              :placeholder="$t('m.Role_Search_Username_Placeholder')"
              prefix-icon="el-icon-search"
              style="width: 300px; margin-right: 10px;"
              @keyup.enter.native="handleSearch"
              clearable
            >
            </el-input>
            <el-button type="primary" @click="handleSearch">{{ $t('m.Search') }}</el-button>
          </div>

          <!-- 未搜索时的提示 -->
          <div v-if="!hasSearched" class="search-hint">
            <el-alert
              :title="$t('m.Role_Input_Username_Search')"
              type="info"
              :closable="false"
              show-icon
            >
              <template>
                <div style="margin-top: 10px;">
                  <p>{{ $t('m.Role_Search_Hint_1') }}</p>
                  <p>{{ $t('m.Role_Search_Hint_2') }}</p>
                </div>
              </template>
            </el-alert>
          </div>

          <!-- 搜索结果表格 -->
          <el-table
            v-else
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
                  <el-radio label="teacher" :disabled="!isSuperAdmin">{{ $t('m.Teacher') }}</el-radio>
                  <el-radio label="student">{{ $t('m.Student') }}</el-radio>
                </el-radio-group>
                <el-tooltip v-if="!isSuperAdmin && row.role === 'teacher'" :content="$t('m.Role_Only_Super_Admin_Teacher')" placement="top">
                  <i class="el-icon-info" style="color: #E6A23C; margin-left: 5px;"></i>
                </el-tooltip>
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
            v-if="hasSearched"
            class="page"
            layout="prev, pager, next, sizes, total"
            :total="total"
            :page-size="pageSize"
            :current-page.sync="currentPage"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          >
          </el-pagination>
        </el-tab-pane>

        <!-- Tab 2: 权限申请管理 -->
        <el-tab-pane :label="$t('m.Role_Applications')" name="applicationManagement">
          <!-- 筛选器 -->
          <div class="filter-row">
            <el-select v-model="applicationFilter.role" :placeholder="$t('m.Apply_Role')" style="width: 150px; margin-right: 10px;" clearable @change="loadApplications">
              <el-option value="all" :label="$t('m.Role_All')" />
              <el-option value="teacher" :label="$t('m.Teacher')" />
              <el-option value="student" :label="$t('m.Student')" />
            </el-select>
            <el-select v-model="applicationFilter.status" :placeholder="$t('m.Status')" style="width: 150px; margin-right: 10px;" clearable @change="loadApplications">
              <el-option value="all" :label="$t('m.Role_All')" />
              <el-option value="0" :label="$t('m.Pending_Approval')" />
              <el-option value="1" :label="$t('m.Approved')" />
              <el-option value="2" :label="$t('m.Rejected')" />
            </el-select>
            <el-button type="primary" icon="el-icon-refresh" @click="loadApplications">{{ $t('m.Refresh') }}</el-button>
          </div>

          <!-- 批量操作按钮 -->
          <div v-if="applications.length > 0" class="batch-actions">
            <el-button size="small" type="success" :disabled="selectedApplications.length === 0" @click="handleBatchAction('approve')">
              <i class="el-icon-check"></i>
              {{ $t('m.Batch_Approve') }}
            </el-button>
            <el-button size="small" type="danger" :disabled="selectedApplications.length === 0" @click="handleBatchAction('reject')">
              <i class="el-icon-close"></i>
              {{ $t('m.Batch_Reject') }}
            </el-button>
            <span v-if="selectedApplications.length > 0" style="margin-left: 10px; color: #909399;">
              {{ $t('m.Role_Selected_Count', { count: selectedApplications.length }) }}
            </span>
          </div>

          <!-- 申请列表表格 -->
          <el-table
            :data="applications"
            v-loading="loadingApplications"
            stripe
            border
            style="margin-top: 20px;"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" :selectable="checkSelectable" />
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column :label="$t('m.Username')" min-width="120">
              <template slot-scope="{ row }">
                <UserName :username="row.username" />
              </template>
            </el-table-column>
            <el-table-column prop="realname" :label="$t('m.RealName')" min-width="100" />
            <el-table-column :label="$t('m.Apply_Role')" width="100">
              <template slot-scope="{ row }">
                <el-tag v-if="row.role === 'teacher'" type="primary" size="small">{{ $t('m.Teacher') }}</el-tag>
                <el-tag v-else-if="row.role === 'student'" type="success" size="small">{{ $t('m.Student') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.Status')" width="100">
              <template slot-scope="{ row }">
                <el-tag v-if="row.status === 0" type="warning" size="small">{{ $t('m.Pending_Approval') }}</el-tag>
                <el-tag v-else-if="row.status === 1" type="success" size="small">{{ $t('m.Approved') }}</el-tag>
                <el-tag v-else-if="row.status === 2" type="danger" size="small">{{ $t('m.Rejected') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.Application_Reason')" min-width="200" show-overflow-tooltip>
              <template slot-scope="{ row }">
                {{ row.reason || '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.Reviewer')" width="100">
              <template slot-scope="{ row }">
                {{ row.reviewerName || '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.Review_Time')" width="160">
              <template slot-scope="{ row }">
                {{ row.reviewTime ? formatTime(row.reviewTime) : '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('m.Operation')" width="180" fixed="right">
              <template slot-scope="{ row }">
                <div v-if="row.status === 0" class="action-buttons">
                  <el-button
                    size="mini"
                    type="success"
                    @click="handleSingleAction(row, 'approve')"
                  >
                    {{ $t('m.Approve') }}
                  </el-button>
                  <el-button
                    size="mini"
                    type="danger"
                    @click="handleSingleAction(row, 'reject')"
                  >
                    {{ $t('m.Reject') }}
                  </el-button>
                </div>
                <el-tag v-else type="info" size="small">{{ row.statusName }}</el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-if="applicationTotal > 0"
            class="page"
            layout="prev, pager, next, sizes, total"
            :total="applicationTotal"
            :page-size="applicationPageSize"
            :current-page.sync="applicationCurrentPage"
            @current-change="handleApplicationPageChange"
            @size-change="handleApplicationSizeChange"
          >
          </el-pagination>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 审批对话框 -->
    <el-dialog
      :title="reviewDialogTitle"
      :visible.sync="reviewDialogVisible"
      width="500px"
    >
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item :label="$t('m.Action')">
          <el-tag v-if="reviewForm.action === 'approve'" type="success">{{ $t('m.Approve') }}</el-tag>
          <el-tag v-else-if="reviewForm.action === 'reject'" type="danger">{{ $t('m.Reject') }}</el-tag>
        </el-form-item>
        <el-form-item :label="$t('m.Count')">
          <strong>{{ reviewForm.applicationIds.length }}</strong> {{ $t('m.Count_Applications') }}
        </el-form-item>
        <el-form-item :label="$t('m.Review_Note')">
          <el-input
            v-model="reviewForm.reviewNote"
            type="textarea"
            :rows="3"
            :placeholder="$t('m.Role_Review_Note_Placeholder')"
          />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="reviewDialogVisible = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="confirmReview" :loading="reviewing">
          {{ $t('m.Confirm') }}
        </el-button>
      </div>
    </el-dialog>

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
            <p><strong>5. {{ $t('m.Role_Permission_Applications') }}</strong> {{ $t('m.Role_Permission_Applications_Desc') }}</p>
            <el-divider></el-divider>
            <p style="color: #E6A23C;">
              <i class="el-icon-warning"></i>
              <strong>{{ $t('m.Role_Permission_Notes') }}</strong>
              <template v-if="isSuperAdmin">
                {{ $t('m.Role_Super_Admin_Note') }}
              </template>
              <template v-else>
                {{ $t('m.Role_Non_Super_Admin_Note') }}
              </template>
            </p>
          </div>
        </template>
      </el-alert>
    </el-card>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
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
      // Tab
      activeTab: 'roleManagement',
      
      // 用户角色管理
      loading: false,
      userList: [],
      total: 0,
      currentPage: 1,
      pageSize: 20,
      keyword: '',
      originalRoles: new Map(),
      hasSearched: false,

      // 权限申请管理
      loadingApplications: false,
      applications: [],
      applicationTotal: 0,
      applicationCurrentPage: 1,
      applicationPageSize: 20,
      applicationFilter: {
        role: 'all',
        status: 'all'
      },
      selectedApplications: [],

      // 审批对话框
      reviewDialogVisible: false,
      reviewing: false,
      reviewForm: {
        applicationIds: [],
        action: '',
        reviewNote: ''
      }
    }
  },
  computed: {
    ...mapGetters(['isSuperAdmin', 'isProblemAdmin', 'isAdminRole']),
    reviewDialogTitle() {
      if (this.reviewForm.action === 'approve') {
        return this.$t('m.Confirm_Batch_Approve').replace('{count}', this.reviewForm.applicationIds.length)
      } else if (this.reviewForm.action === 'reject') {
        return this.$t('m.Confirm_Batch_Reject').replace('{count}', this.reviewForm.applicationIds.length)
      }
      return ''
    }
  },
  mounted() {
    // 默认不自动加载用户，需要手动搜索
  },
  methods: {
    // ==================== Tab 切换 ====================
    handleTabClick(tab) {
      if (tab.name === 'applicationManagement') {
        this.loadApplications()
      }
    },
    refreshCurrentTab() {
      if (this.activeTab === 'roleManagement') {
        if (this.hasSearched) {
          this.loadUsers()
        }
      } else {
        this.loadApplications()
      }
    },

    // ==================== 用户角色管理 ====================
    async loadUsers() {
      this.loading = true
      try {
        const res = await classroomApi.searchUsersForRoleManagement(this.keyword, this.currentPage, this.pageSize)
        if (res.data.data) {
          const users = res.data.data.records

          const usersWithRoles = await Promise.all(
            users.map(async (user) => {
              try {
                const roleRes = await classroomApi.getUserRoles(user.uid)
                let role = ''

                if (roleRes.data && roleRes.data.code === 200 && roleRes.data.data) {
                  const roles = roleRes.data.data || []
                  const roleNames = roles.map(item => item.role || item)
                  role = roleNames.length > 0 ? roleNames[0] : ''
                }

                return {
                  ...user,
                  role: role,
                  originalRole: role,
                  saving: false
                }
              } catch (error) {
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
      if (!this.keyword || this.keyword.trim() === '') {
        this.$message.warning(this.$t('m.Role_Search_Username_Placeholder'))
        return
      }
      this.currentPage = 1
      this.hasSearched = true
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
        const roles = user.role ? [user.role] : []

        const response = await classroomApi.updateUserRoles(user.uid, roles)

        if (response.data.code === 200) {
          user.originalRole = user.role
          this.$message.success(this.$t('m.Role_Save_Success'))
        } else {
          throw new Error(response.data.msg || this.$t('m.Role_Save_Failed'))
        }
      } catch (error) {
        this.$message.error(error.response?.data?.msg || this.$t('m.Role_Save_Failed'))
        user.role = user.originalRole
      } finally {
        user.saving = false
      }
    },
    handleRoleChange(user) {
      if (user.role === 'teacher' && !this.isSuperAdmin) {
        this.$message.warning(this.$t('m.Role_Only_Super_Admin_Can_Set'))
        user.role = user.originalRole
        return
      }
    },

    // ==================== 权限申请管理 ====================
    async loadApplications() {
      this.loadingApplications = true
      try {
        const params = {
          role: this.applicationFilter.role,
          status: this.applicationFilter.status,
          currentPage: this.applicationCurrentPage,
          limit: this.applicationPageSize
        }

        const res = await this.$store.dispatch('classroom/getRoleApplications', params)

        if (res.code === 200) {
          this.applications = res.data.records || []
          this.applicationTotal = res.data.total || 0
          this.selectedApplications = []
        } else {
          this.$message.error(res.message || this.$t('m.Load_Failed'))
        }
      } catch (error) {
        console.error('加载申请列表失败:', error)
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loadingApplications = false
      }
    },
    handleApplicationPageChange(page) {
      this.applicationCurrentPage = page
      this.loadApplications()
    },
    handleApplicationSizeChange(size) {
      this.applicationPageSize = size
      this.applicationCurrentPage = 1
      this.loadApplications()
    },
    handleSelectionChange(selection) {
      this.selectedApplications = selection
    },
    checkSelectable(row) {
      return row.status === 0 // 只能选择待审批的申请
    },
    async handleSingleAction(row, action) {
      this.reviewForm.applicationIds = [row.id]
      this.reviewForm.action = action
      this.reviewForm.reviewNote = ''
      this.reviewDialogVisible = true
    },
    handleBatchAction(action) {
      if (this.selectedApplications.length === 0) {
        this.$message.warning(this.$t('m.Role_Select_Application_First'))
        return
      }
      this.reviewForm.applicationIds = this.selectedApplications.map(app => app.id)
      this.reviewForm.action = action
      this.reviewForm.reviewNote = ''
      this.reviewDialogVisible = true
    },
    async confirmReview() {
      this.reviewing = true
      try {
        const res = await this.$store.dispatch('classroom/reviewRoleApplication', {
          applicationIds: this.reviewForm.applicationIds,
          action: this.reviewForm.action,
          reviewNote: this.reviewForm.reviewNote
        })

        if (res.code === 200) {
          const message = this.$t('m.Batch_Operation_Success')

          this.$message.success(message + this.$t('m.Role_Batch_Processed_Count', { count: res.data.successCount }))
          this.reviewDialogVisible = false
          this.loadApplications()
        } else {
          this.$message.error(res.message || this.$t('m.Batch_Operation_Failed'))
        }
      } catch (error) {
        console.error('审批失败:', error)
        this.$message.error(this.$t('m.Batch_Operation_Failed'))
      } finally {
        this.reviewing = false
      }
    },
    formatTime(time) {
      if (!time) return '-'
      const date = new Date(time)
      return date.toLocaleString('zh-CN')
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

.search-hint {
  margin-top: 20px;
}

.page {
  margin-top: 20px;
  text-align: right;
}

.batch-actions {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.dialog-footer {
  text-align: right;
}

/* 操作按钮容器样式 */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
}

.action-buttons .el-button {
  margin: 0;
  min-width: 60px;
}
</style>
