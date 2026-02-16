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
        <el-form :inline="true" size="small">
          <el-form-item label="教师用户名">
            <el-input
              v-model="teacherUsername"
              placeholder="请输入教师用户名"
              style="width: 300px;"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="addTeacher" :loading="addingTeacher">
              添加
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      <span slot="footer">
        <el-button @click="teacherDialogVisible = false">关闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import Students from '@/views/classroom/teacher/Students.vue'
import Checkin from '@/views/classroom/teacher/Checkin.vue'
import RandomPick from '@/views/classroom/teacher/RandomPick.vue'
import Homework from '@/views/classroom/teacher/Homework.vue'
import Materials from '@/views/classroom/teacher/Materials.vue'
import Discussion from '@/views/classroom/teacher/Discussion.vue'
import UserName from '@/components/oj/common/UserName.vue'

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
      teacherUsername: '',
      addingTeacher: false
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
      if (!this.teacherUsername.trim()) {
        this.$message.warning('请输入教师用户名')
        return
      }

      this.addingTeacher = true
      try {
        const res = await this.$store.dispatch('classroom/addClassroomTeacher', {
          classroomId: this.currentClassroom.id,
          username: this.teacherUsername.trim()
        })

        if (res.code === 200) {
          this.$message.success('添加教师成功')
          this.teacherUsername = ''
          await this.loadClassroomTeachers()
        } else {
          this.$message.error(res.msg || '添加教师失败')
        }
      } catch (error) {
        this.$message.error('添加教师失败')
        console.error('添加教师失败:', error)
      } finally {
        this.addingTeacher = false
      }
    },
    async removeTeacher(teacherRelation) {
      const isMainTeacher = !this.currentClassroom.teacher ||
                               this.currentClassroom.teacher?.uuid !== teacherRelation.teacherId

      if (isMainTeacher) {
        // 移除主教师，需要指定新的主教师
        this.$prompt('移除主教师必须指定新的主教师', '请输入新主教师的用户名', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          inputType: 'text',
          inputPlaceholder: '请输入用户名',
          inputErrorMessage: '用户名不能为空'
        }).then(async ({ value }) => {
          if (value) {
            try {
              const res = await this.$store.dispatch('classroom/removeClassroomTeacher', {
                classroomId: this.currentClassroom.id,
                teacherId: teacherRelation.teacherId,
                newTeacherId: value // 新主教师的用户名
              })

              if (res.code === 200) {
                this.$message.success('更换主教师成功')
                await this.loadClassroomTeachers()
              } else {
                this.$message.error(res.msg || '更换主教师失败')
              }
            } catch (error) {
              this.$message.error('更换主教师失败')
              console.error('更换主教师失败:', error)
            }
          }
        }).catch(() => {})
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
            } else {
              this.$message.error(res.msg || '移除教师失败')
            }
          } catch (error) {
            this.$message.error('移除教师失败')
            console.error('移除教师失败:', error)
          }
        }).catch(() => {})
      }
    },
    resetTeacherDialog() {
      this.currentClassroom = null
      this.currentClassroomTeachers = []
      this.teacherUsername = ''
    }
  }
}
</script>

<style scoped>
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
