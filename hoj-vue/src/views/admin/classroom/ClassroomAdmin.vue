<template>
  <div class="classroom-admin">
    <!-- 列表视图 -->
    <div v-if="!selectedClassroomId" class="classroom-list-view">
      <el-card>
        <div slot="header" class="header">
          <span class="title">{{ $t('m.Classroom_Management') }}</span>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="loadClassrooms">
            {{ $t('m.Refresh') }}
          </el-button>
        </div>

        <el-table :data="classrooms" v-loading="loading" stripe border>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="className" :label="$t('m.Classroom_Name')" min-width="200" />
          <el-table-column prop="classBelong" :label="$t('m.Classroom_Belong')" width="150" />
          <el-table-column :label="$t('m.Teacher_Name')" width="200">
            <template slot-scope="{ row }">
              <!-- 显示所有教师（teachers 数组） -->
              <div v-if="row.teachers && row.teachers.length > 0" class="teacher-list">
                <span v-for="(teacherRel, index) in row.teachers" :key="index">
                  <UserName :username="teacherRel.teacher ? teacherRel.teacher.username : teacherRel.teacher.teacherId" />
                  <span v-if="index < row.teachers.length - 1">, </span>
                </span>
              </div>
              <!-- 兼容旧数据：只显示主教师（teacher 对象） -->
              <UserName v-else :username="row.teacher ? row.teacher.username : row.teacherId || '-'" />
            </template>
          </el-table-column>
          <el-table-column prop="classCode" :label="$t('m.Classroom_Code')" width="120" />
          <el-table-column prop="createTime" :label="$t('m.Create_Time')" width="180">
            <template slot-scope="{ row }">{{ formatTime(row.createTime) }}</template>
          </el-table-column>
          <el-table-column :label="$t('m.Operation')" width="220" fixed="right">
            <template slot-scope="{ row }">
              <el-button size="mini" type="success" @click="manageTeachers(row)" style="margin-right: 5px;">
                管理教师
              </el-button>
              <el-button size="mini" type="primary" @click="enterClassroom(row)">
                进入班级
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!loading && classrooms.length === 0" :description="$t('m.No_Data')">
          <p class="empty-tip">{{ $t('m.Classroom_Admin_Empty_Tip') }}</p>
        </el-empty>
      </el-card>
    </div>

    <!-- 班级详情视图 - 复用教师端的界面 -->
    <div v-else class="classroom-detail-view">
      <div class="page-header">
        <el-button icon="el-icon-arrow-left" @click="goBackToList">{{ $t('m.Back') }}</el-button>
        <h2>{{ classroomInfo.className || $t('m.Classroom_Detail') }}</h2>
      </div>
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="$t('m.Student_Management')" name="students">
          <Students v-if="activeTab === 'students'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Checkin_Management')" name="checkin">
          <Checkin v-if="activeTab === 'checkin'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Random_Pick')" name="randomPick">
          <RandomPick v-if="activeTab === 'randomPick'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Homework_Management')" name="homework">
          <Homework v-if="activeTab === 'homework'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Material_Library')" name="materials">
          <Materials v-if="activeTab === 'materials'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
        <el-tab-pane :label="$t('m.Discussion')" name="discussion">
          <Discussion v-if="activeTab === 'discussion'" :classroom-id="selectedClassroomId" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 教师管理对话框 -->
    <el-dialog
      title="班级教师管理"
      :visible.sync="teacherDialogVisible"
      width="600px"
      @close="resetTeacherDialog"
    >
      <div v-if="currentClassroom">
        <p style="margin-bottom: 15px;">
          班级：<strong>{{ currentClassroom.className }}</strong>
        </p>

        <!-- 已添加的教师列表 -->
        <div style="margin-bottom: 20px;">
          <h4 style="margin-bottom: 10px;">已添加的教师</h4>
          <el-table :data="currentClassroomTeachers" size="small" stripe>
            <el-table-column label="教师" min-width="150">
              <template slot-scope="{ row }">
                <UserName v-if="row.teacher" :username="row.teacher.username" />
                <span v-else>{{ row.teacherId }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template slot-scope="{ row }">
                <el-button
                  size="mini"
                  type="danger"
                  @click="removeTeacher(row)"
                >
                  移除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 添加教师表单 -->
        <el-divider></el-divider>
        <h4 style="margin-bottom: 10px;">添加教师</h4>
        <el-form :inline="true" size="small" @submit.native.prevent>
          <el-form-item label="教师用户名">
            <el-select
              v-model="selectedTeacherUid"
              filterable
              remote
              clearable
              reserve-keyword
              :remote-method="searchTeacherOptions"
              :loading="searchingTeachers"
              placeholder="输入用户名 / 姓名 / 昵称搜索"
              :loading-text="searchingTeachers ? '搜索中...' : '加载中'"
              style="width: 300px;"
            >
              <el-option
                v-for="user in teacherOptions"
                :key="user.uuid"
                :label="user.username + (user.nickname ? '（' + user.nickname + '）' : '')"
                :value="user.uuid"
              >
                <div class="teacher-option">
                  <span>{{ user.username }}<span v-if="user.nickname" class="teacher-option-nickname">（{{ user.nickname }}）</span></span>
                  <el-tag size="mini" :type="classroomRoleTagType(user.role)">
                    {{ classroomRoleText(user.role) }}
                  </el-tag>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              @click="addTeacher"
              :loading="addingTeacher"
              :disabled="!selectedTeacherUser || isKnownNonTeacher"
            >
              添加
            </el-button>
          </el-form-item>
        </el-form>
        <div class="add-teacher-hint">
          <template v-if="!teacherOptions.length">输入关键词搜索用户，仅拥有<strong>教师角色</strong>的用户可被添加</template>
          <template v-else-if="!selectedTeacherUser">请在下拉框中选择一个用户</template>
          <template v-else-if="selectedTeacherUser.role !== 'teacher'">
            <span style="color: #F56C6C;">该用户当前班级角色：{{ classroomRoleText(selectedTeacherUser.role) }}，需先在「角色管理」中授予教师角色后才能添加</span>
          </template>
          <template v-else>
            <span style="color: #67C23A;">将添加 {{ selectedTeacherUser.username }} 为班级教师</span>
          </template>
        </div>
      </div>
      <span slot="footer">
        <el-button @click="teacherDialogVisible = false">关闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import Vue from 'vue'
import moment from 'moment'
import Students from '@/views/classroom/teacher/Students.vue'
import Checkin from '@/views/classroom/teacher/Checkin.vue'
import RandomPick from '@/views/classroom/teacher/RandomPick.vue'
import Homework from '@/views/classroom/teacher/Homework.vue'
import Materials from '@/views/classroom/teacher/Materials.vue'
import Discussion from '@/views/classroom/teacher/Discussion.vue'
import UserName from '@/components/oj/common/UserName.vue'

// 创建一个教师选择器组件
const TeacherSelectDialog = Vue.extend({
  props: {
    teachers: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      selectedTeacher: ''
    }
  },
  methods: {
    getValue() {
      return this.selectedTeacher
    },
    reset() {
      this.selectedTeacher = ''
    }
  },
  render(h) {
    const self = this
    return h('div', { style: 'padding: 10px 0;' }, [
      h('p', { style: 'margin-bottom: 15px;' }, '请选择新的主教师：'),
      h('el-select', {
        props: {
          value: this.selectedTeacher,
          placeholder: '请选择教师',
          style: 'width: 100%;',
          clearable: true
        },
        on: {
          input: (val) => {
            self.selectedTeacher = val
            // 向父组件发出事件
            self.$emit('input', val)
          }
        }
      }, this.teachers.map(teacher => {
        return h('el-option', {
          props: {
            value: teacher.teacherId,
            label: teacher.teacher ? teacher.teacher.username : teacher.teacherId
          }
        })
      }))
    ])
  }
})

export default {
  name: 'ClassroomAdmin',
  components: {
    UserName,
    Students,
    Checkin,
    RandomPick,
    Homework,
    Materials,
    Discussion
  },
  data() {
    return {
      loading: false,
      classrooms: [],
      selectedClassroomId: null,
      activeTab: 'students',
      classroomInfo: {},
      // 教师管理
      teacherDialogVisible: false,
      currentClassroom: null,
      currentClassroomTeachers: [],
      addingTeacher: false,
      // 添加教师：远程搜索 + 角色校验
      teacherOptions: [],
      searchingTeachers: false,
      selectedTeacherUid: '',
      selectedNewTeacher: '', // 选中的新主教师ID
      tempTeacherSelect: null // 临时用于对话框中的教师选择器
    }
  },
  computed: {
    // 获取班级列表中的当前班级对象，用于更新显示
    currentClassroomInList() {
      return this.classrooms.find(c => c.id === this.currentClassroom?.id)
    },
    // 当前在下拉框中选中的用户对象
    selectedTeacherUser() {
      return this.teacherOptions.find(u => u.uuid === this.selectedTeacherUid) || null
    },
    // 后端返回了 role 字段且明确不是教师 → 禁用添加；字段缺失（旧后端）时不误锁
    isKnownNonTeacher() {
      const u = this.selectedTeacherUser
      return !!u && u.role !== undefined && u.role !== 'teacher'
    }
  },
  mounted() {
    this.loadClassrooms()
    // 检查 URL 参数,如果指定了班级ID,直接进入详情
    if (this.$route.query.classroomId) {
      // 检查是否同时指定了作业ID或其他资源ID
      if (this.$route.query.homeworkId) {
        // 有作业ID，设置为作业标签页
        this.selectedClassroomId = this.$route.query.classroomId
        this.activeTab = 'homework'
        // 不触发 enterClassroom，直接设置 classroomInfo
        // 作业详情会通过 Homework 组件内部的 computed 处理
      } else {
        this.enterClassroom({ id: this.$route.query.classroomId })
      }
    }
  },
  methods: {
    async loadClassrooms() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getAllClassroomsForAdmin')
        if (res.code === 200) {
          this.classrooms = res.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },
    enterClassroom(classroom) {
      this.selectedClassroomId = classroom.id
      this.classroomInfo = classroom
      // 更新 URL 参数,方便刷新时保持状态
      // 使用 name 保持路由上下文，避免跳转到根路由
      this.$router.replace({
        name: 'admin-classroom',
        query: { classroomId: classroom.id }
      })
    },
    goBackToList() {
      this.selectedClassroomId = null
      this.classroomInfo = {}
      this.activeTab = 'students'
      // 使用 name 保持路由上下文
      this.$router.replace({
        name: 'admin-classroom',
        query: {}
      })
    },
    handleTabClick(tab) {
      // 可以在这里保存标签状态到 URL
      console.log('Tab clicked:', tab.name)
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    },
    // 教师管理相关方法
    async manageTeachers(classroom) {
      this.currentClassroom = classroom
      this.teacherDialogVisible = true
      await this.loadClassroomTeachers()
    },
    async loadClassroomTeachers() {
      if (!this.currentClassroom) return

      try {
        const res = await this.$store.dispatch('classroom/getClassroomTeachers', this.currentClassroom.id)
        if (res.code === 200) {
          this.currentClassroomTeachers = res.data || []
        }
      } catch (error) {
        this.$message.error('加载教师列表失败')
        console.error('加载教师列表失败:', error)
      }
    },
    async addTeacher() {
      const user = this.selectedTeacherUser
      if (!user) {
        this.$message.warning('请先搜索并选择用户')
        return
      }
      // 前端兜底校验：已知角色且非教师时拦截（后端同样会校验）
      if (user.role !== undefined && user.role !== 'teacher') {
        this.$message.warning('该用户没有教师角色，无法添加为班级教师')
        return
      }

      this.addingTeacher = true
      try {
        const res = await this.$store.dispatch('classroom/addClassroomTeacher', {
          classroomId: this.currentClassroom.id,
          teacherId: user.uuid
        })

        if (res.code === 200) {
          this.$message.success('添加教师成功')
          this.selectedTeacherUid = ''
          this.teacherOptions = []
          await this.loadClassroomTeachers()
          // 同时更新列表中的班级信息
          await this.loadClassrooms()
        }
        // 注意：错误时的消息提示已由 axios 拦截器处理，无需重复显示
      } catch (error) {
        // axios 拦截器已处理错误提示，这里只记录日志
        console.error('添加教师失败:', error)
      } finally {
        this.addingTeacher = false
      }
    },
    // 远程搜索用户（后端返回附带班级角色）
    async searchTeacherOptions(keyword) {
      if (!keyword || !keyword.trim()) {
        this.teacherOptions = []
        return
      }
      this.searchingTeachers = true
      try {
        const res = await this.$store.dispatch('classroom/searchTeachersForAdmin', keyword.trim())
        if (res.code === 200) {
          this.teacherOptions = res.data || []
        }
      } catch (error) {
        console.error('搜索用户失败:', error)
      } finally {
        this.searchingTeachers = false
      }
    },
    classroomRoleText(role) {
      if (role === undefined) return '未知角色'
      if (!role) return '无角色'
      if (role.includes('teacher') && role.includes('student')) return '教师/学生'
      if (role.includes('teacher')) return '教师'
      if (role.includes('student')) return '学生'
      return role
    },
    classroomRoleTagType(role) {
      if (role !== undefined && role.includes('teacher')) return 'success'
      if (role !== undefined && role.includes('student')) return 'warning'
      return 'info'
    },
    async removeTeacher(teacherRelation) {
      // 判断要删除的教师是否是主教师
      const isMainTeacher = this.currentClassroom.teacher?.uuid === teacherRelation.teacherId

      if (isMainTeacher) {
        // 移除主教师，需要从其他教师中选择新的主教师
        // 过滤出除了当前要删除的主教师之外的其他教师
        const otherTeachers = this.currentClassroomTeachers.filter(
          t => t.teacherId !== teacherRelation.teacherId
        )

        if (otherTeachers.length === 0) {
          this.$message.warning('没有其他教师可以担任主教师，请先添加其他教师')
          return
        }

        // 重置临时选择
        this.tempTeacherSelect = ''

        const h = this.$createElement

        // 使用 h 函数创建一个包含 el-select 的组件
        this.$msgbox({
          title: '移除主教师',
          message: h('div', { style: 'padding: 10px 0;' }, [
            h('p', { style: 'margin-bottom: 15px;' }, '请选择新的主教师：'),
            h(TeacherSelectDialog, {
              props: {
                teachers: otherTeachers
              },
              ref: 'teacherSelectDialog',
              on: {
                input: (val) => {
                  this.tempTeacherSelect = val
                }
              }
            })
          ]),
          showCancelButton: true,
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          beforeClose: (action, instance, done) => {
            if (action === 'confirm') {
              // 通过 ref 获取组件实例
              const dialog = this.$refs.teacherSelectDialog
              const selectedValue = dialog ? dialog.getValue() : this.tempTeacherSelect
              if (!selectedValue) {
                this.$message.warning('请选择新的主教师')
                return
              }
              this.transferMainTeacher(teacherRelation.teacherId, selectedValue).then(() => {
                this.tempTeacherSelect = null
                done()
              })
            } else {
              this.tempTeacherSelect = null
              done()
            }
          }
        }).catch(() => {
          this.tempTeacherSelect = null
        })
      } else {
        // 移除普通教师
        this.$confirm('确认移除该教师吗？', '警告', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(async () => {
          try {
            const res = await this.$store.dispatch('classroom/removeClassroomTeacher', {
              classroomId: this.currentClassroom.id,
              teacherId: teacherRelation.teacherId
            })

            if (res.code === 200) {
              this.$message.success('移除教师成功')
              await this.loadClassroomTeachers()
            }
            // 注意：错误时的消息提示已由 axios 拦截器处理，无需重复显示
          } catch (error) {
            // axios 拦截器已处理错误提示，这里只记录日志
            console.error('移除教师失败:', error)
          }
        }).catch(() => {})
      }
    },
    async transferMainTeacher(oldTeacherId, newTeacherId) {
      try {
        const res = await this.$store.dispatch('classroom/removeClassroomTeacher', {
          classroomId: this.currentClassroom.id,
          teacherId: oldTeacherId,
          newTeacherId: newTeacherId
        })

        if (res.code === 200) {
          this.$message.success('更换主教师成功')
          await this.loadClassroomTeachers()
        }
        // 注意：错误时的消息提示已由 axios 拦截器处理，无需重复显示
      } catch (error) {
        // axios 拦截器已处理错误提示，这里只记录日志
        console.error('更换主教师失败:', error)
      }
    },
    resetTeacherDialog() {
      this.currentClassroom = null
      this.currentClassroomTeachers = []
      this.teacherOptions = []
      this.searchingTeachers = false
      this.selectedTeacherUid = ''
    }
  }
}
</script>

<style scoped>
/* 教师搜索下拉：左侧用户名、右侧角色标签 */
.teacher-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.teacher-option-nickname {
  color: #909399;
}

.add-teacher-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
}

.classroom-admin {
  padding: 20px;
  background-color: #fff;
  min-height: 100vh;
}

.classroom-list-view {
  width: 100%;
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

.empty-tip {
  color: #909399;
  font-size: 14px;
  margin-top: 10px;
}

.classroom-detail-view {
  width: 100%;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}
</style>
