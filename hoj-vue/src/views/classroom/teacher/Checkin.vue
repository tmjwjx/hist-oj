<template>
  <div class="checkin-panel">
    <div class="header">
      <h3>{{ $t('m.Checkin_Management') }}</h3>
      <el-button type="primary" icon="el-icon-plus" @click="showCreateDialog = true">
        {{ $t('m.Create_Checkin') }}
      </el-button>
    </div>

    <el-timeline>
      <el-timeline-item v-for="checkin in checkins" :key="checkin.id" :timestamp="formatTime(checkin.startTime)">
        <el-card>
          <div class="checkin-header">
            <h4>{{ checkin.checkinName || $t('m.Checkin') }}</h4>
            <div class="actions">
              <!-- 签到类型标签 -->
              <el-tag :type="checkin.checkinType === 'qrcode' ? 'warning' : 'primary'" size="small">
                <i :class="checkin.checkinType === 'qrcode' ? 'el-icon-full-screen' : 'el-icon-key'"></i>
                {{ checkin.checkinType === 'qrcode' ? $t('m.Qrcode_Checkin') : $t('m.Checkin_Code_Checkin') }}
              </el-tag>

              <!-- 显示二维码按钮（仅二维码类型且进行中） -->
              <el-button
                v-if="checkin.checkinType === 'qrcode' && checkin.status === 1"
                size="small"
                type="success"
                @click="showQrcode(checkin)"
                icon="el-icon-full-screen"
              >
                {{ $t('m.Show_Qrcode') }}
              </el-button>

              <!-- 显示签到码（仅签到码类型） -->
              <el-tag v-if="checkin.checkinType === 'code'" type="info" size="small">
                {{ $t('m.Checkin_Code') }}: {{ checkin.checkinCode }}
              </el-tag>

              <el-tag :type="checkin.status === 1 ? 'success' : 'info'" size="small">
                {{ checkin.status === 1 ? $t('m.In_Progress') : $t('m.Ended') }}
              </el-tag>
              <el-button size="small" @click="viewRecords(checkin)">
                {{ $t('m.View_Records') }}
              </el-button>
              <el-button v-if="checkin.status === 1" size="small" type="warning" @click="endCheckin(checkin)">
                {{ $t('m.End') }}
              </el-button>
              <el-button size="small" type="primary" @click="editCheckin(checkin)">
                {{ $t('m.Edit') }}
              </el-button>
              <el-button size="small" type="danger" @click="deleteCheckin(checkin)">
                {{ $t('m.Delete') }}
              </el-button>
            </div>
          </div>
        </el-card>
      </el-timeline-item>
    </el-timeline>

    <!-- 创建签到对话框 -->
    <el-dialog :title="$t('m.Create_Checkin')" :visible.sync="showCreateDialog" width="500px" @close="resetCreateForm">
      <el-form :model="createForm" :rules="rules" ref="createForm" label-width="140px">
        <el-form-item :label="$t('m.Checkin_Name')" prop="checkinName">
          <el-input v-model="createForm.checkinName" :placeholder="$t('m.Please_Enter_Checkin_Name')" />
        </el-form-item>

        <!-- 签到类型选择 -->
        <el-form-item :label="$t('m.Checkin_Type')" prop="checkinType">
          <el-radio-group v-model="createForm.checkinType">
            <el-radio label="code">
              <i class="el-icon-key"></i> {{ $t('m.Checkin_Code_Type') }}
            </el-radio>
            <el-radio label="qrcode">
              <i class="el-icon-full-screen"></i> {{ $t('m.Qrcode_Checkin_Type') }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 二维码刷新间隔（仅二维码类型显示） -->
        <el-form-item
          v-if="createForm.checkinType === 'qrcode'"
          :label="$t('m.Qrcode_Refresh_Interval')"
          prop="qrcodeRefreshInterval"
        >
          <el-input-number
            v-model="createForm.qrcodeRefreshInterval"
            :min="10"
            :max="60"
            :step="5"
          />
          <span style="margin-left: 10px; color: #909399; font-size: 12px;">
            {{ $t('m.Qrcode_Refresh_Tip') }}
          </span>
        </el-form-item>

        <el-form-item :label="$t('m.Start_Time')" prop="startTime">
          <el-date-picker
            v-model="createForm.startTime"
            type="datetime"
            :placeholder="$t('m.Select_Start_Time')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('m.End_Time')" prop="endTime">
          <el-date-picker
            v-model="createForm.endTime"
            type="datetime"
            :placeholder="$t('m.Select_End_Time')"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createCheckin">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 编辑签到对话框 -->
    <el-dialog :title="$t('m.Edit_Checkin')" :visible.sync="showEditDialog" width="500px" @close="resetEditForm">
      <el-form :model="editForm" :rules="rules" ref="editForm" label-width="140px">
        <el-form-item :label="$t('m.Checkin_Name')" prop="checkinName">
          <el-input v-model="editForm.checkinName" :placeholder="$t('m.Please_Enter_Checkin_Name')" />
        </el-form-item>

        <el-form-item :label="$t('m.Start_Time')" prop="startTime">
          <el-date-picker
            v-model="editForm.startTime"
            type="datetime"
            :placeholder="$t('m.Select_Start_Time')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('m.End_Time')" prop="endTime">
          <el-date-picker
            v-model="editForm.endTime"
            type="datetime"
            :placeholder="$t('m.Select_End_Time')"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showEditDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="updateCheckin">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 查看签到记录对话框 -->
    <el-dialog :title="$t('m.Checkin_Records')" :visible.sync="showRecordsDialog" width="800px">
      <div style="margin-bottom: 15px;">
        <el-button type="success" icon="el-icon-download" @click="exportCheckinRecords">
          导出签到记录
        </el-button>
      </div>
      <el-table :data="records" stripe>
        <el-table-column :label="$t('m.Username')">
          <template slot-scope="{ row }">
            <UserName :username="row.student?.username" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Real_Name')">
          <template slot-scope="{ row }">
            {{ row.classStudent ? row.classStudent.realName : (row.student ? row.student.realName : '-') }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Status')" width="120">
          <template slot-scope="{ row }">
            <el-select v-model="row.status" size="small" @change="updateRecord(row)">
              <el-option label="签到" value="present" />
              <el-option label="缺勤" value="absent" />
              <el-option label="病假" value="sick_leave" />
              <el-option label="事假" value="personal_leave" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Checkin_Time')">
          <template slot-scope="{ row }">
            {{ row.checkinTime ? formatTime(row.checkinTime) : '-' }}
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 二维码显示组件 -->
    <QrcodeDisplay
      ref="qrcodeDisplay"
      :checkin-id="currentCheckin.id"
      :checkin-name="currentCheckin.checkinName"
      :refresh-interval="currentCheckin.qrcodeRefreshInterval || 15"
      @close="handleQrcodeClose"
    />
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'
import UserName from '@/components/oj/common/UserName.vue'
import QrcodeDisplay from '@/components/classroom/QrcodeDisplay.vue'
import teacherAuth from '@/mixins/teacherAuth'

export default {
  name: 'Checkin',
  components: {
    UserName,
    QrcodeDisplay
  },
  mixins: [realtimeSync, teacherAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      checkins: [],
      showCreateDialog: false,
      showEditDialog: false,
      showRecordsDialog: false,
      records: [],
      recordsLoading: false,
      currentCheckin: {},
      createForm: {
        classroomId: null,
        checkinName: '',
        checkinType: 'code',
        qrcodeRefreshInterval: 15,
        startTime: null,
        endTime: null
      },
      editForm: {
        id: null,
        checkinName: '',
        startTime: null,
        endTime: null
      },
      rules: {
        checkinName: [{ required: true, message: this.$t('m.Required'), trigger: 'blur' }],
        startTime: [{ required: true, message: this.$t('m.Required'), trigger: 'change' }],
        checkinType: [{ required: true, message: this.$t('m.Required'), trigger: 'change' }]
      },
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadCheckins',
        immediate: true
      }
    }
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadCheckins()
        }
      }
    }
  },
  mounted() {
    // mounted 时也会通过 watch 触发加载
  },
  methods: {
    async loadCheckins() {
      if (this.loading) return

      const isFirstLoad = this.checkins.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getCheckinList', this.classroomId)
        if (res.code === 200) {
          const newCheckins = res.data || []

          const currentDataString = JSON.stringify(this.checkins)
          const newDataString = JSON.stringify(newCheckins)

          if (currentDataString !== newDataString) {
            this.checkins = newCheckins
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },

    async createCheckin() {
      this.$refs.createForm.validate(async (valid) => {
        if (valid) {
          const data = {
            classroomId: Number(this.classroomId),
            checkinName: this.createForm.checkinName,
            checkinType: this.createForm.checkinType,
            startTime: moment(this.createForm.startTime).format()
          }

          if (this.createForm.endTime) {
            data.endTime = moment(this.createForm.endTime).format()
          }

          if (this.createForm.checkinType === 'qrcode') {
            data.qrcodeRefreshInterval = this.createForm.qrcodeRefreshInterval
          }

          try {
            const res = await this.$store.dispatch('classroom/createCheckin', data)
            if (res.code === 200) {
              this.$message.success(this.$t('m.Create_Success'))
              this.showCreateDialog = false

              // 如果是二维码类型，直接显示二维码
              if (this.createForm.checkinType === 'qrcode' && res.data) {
                this.currentCheckin = res.data
                this.$nextTick(() => {
                  this.$refs.qrcodeDisplay.open()
                })
              }

              this.loadCheckins()
            } else {
              this.$message.error(res.message || this.$t('m.Create_Failed'))
            }
          } catch (error) {
            this.$message.error(this.$t('m.Create_Failed'))
          }
        }
      })
    },

    async viewRecords(checkin) {
      this.currentCheckin = checkin
      this.recordsLoading = true
      this.showRecordsDialog = true
      try {
        // 并行加载签到记录和班级学生列表
        const [recordsRes, studentsRes] = await Promise.all([
          this.$store.dispatch('classroom/getCheckinRecords', checkin.id),
          this.$http.get(`/api/classroom/${this.classroomId}/students`)
        ])

        if (recordsRes.code === 200) {
          const checkinRecords = recordsRes.data || []

          // 获取班级学生列表
          let allStudents = []
          if (studentsRes.data.code === 200) {
            allStudents = studentsRes.data.data || []
          }

          // 创建签到记录映射
          const recordsMap = new Map()
          checkinRecords.forEach(record => {
            recordsMap.set(record.uid, record)
          })

          // 合并数据：包含所有学生（已签到和未签到）
          this.records = allStudents.map(student => {
            const record = recordsMap.get(student.uuid || student.uid)
            return {
              ...student,
              id: record ? record.id : null,  // 保留记录ID用于更新
              checkinId: checkin.id,
              uid: student.uuid || student.uid,
              status: record ? record.status : 'absent', // 未签到默认为缺勤
              checkinTime: record ? record.checkinTime : null,
              student: student.user || {},  // 修复：使用student.user而不是student.userInfo
              classStudent: student
            }
          })
        }
      } catch (error) {
        console.error('加载签到记录失败:', error)
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.recordsLoading = false
      }
    },

    async updateRecord(record) {
      try {
        // 如果没有记录ID，说明是未签到的学生，需要先创建签到记录
        if (!record.id) {
          // 状态映射中文提示
          const statusText = {
            'present': '出席',
            'absent': '缺勤',
            'sick_leave': '病假',
            'personal_leave': '事假'
          }

          const createRes = await this.$http.post(`/api/classroom/checkin/${record.checkinId}/record`, {
            uid: record.uid,
            status: record.status
          })

          if (createRes.data.code === 200) {
            record.id = createRes.data.data.id
            this.$message.success(`已为未签到学生创建记录，状态设为：${statusText[record.status]}`)
          } else {
            this.$message.error(createRes.data.message || '创建记录失败')
            return
          }
        } else {
          // 已有记录，直接更新
          const res = await this.$store.dispatch('classroom/updateCheckinRecord', {
            recordId: record.id,
            status: record.status
          })
          if (res.code === 200) {
            this.$message.success(this.$t('m.Update_Success'))
          } else {
            this.$message.error(res.message || this.$t('m.Update_Failed'))
          }
        }
      } catch (error) {
        console.error('更新记录失败:', error)
        this.$message.error(this.$t('m.Update_Failed'))
      }
    },

    async endCheckin(checkin) {
      try {
        const res = await this.$store.dispatch('classroom/endCheckin', checkin.id)
        if (res.code === 200) {
          this.$message.success(this.$t('m.End_Success'))
          this.loadCheckins()
        } else {
          this.$message.error(res.message || this.$t('m.End_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.End_Failed'))
      }
    },

    // 编辑签到
    editCheckin(checkin) {
      this.editForm = {
        id: checkin.id,
        checkinName: checkin.checkinName,
        startTime: new Date(checkin.startTime),
        endTime: checkin.endTime ? new Date(checkin.endTime) : null
      }
      this.showEditDialog = true
    },

    // 更新签到
    async updateCheckin() {
      this.$refs.editForm.validate(async (valid) => {
        if (valid) {
          const data = {
            checkinId: this.editForm.id,
            checkinName: this.editForm.checkinName,
            startTime: moment(this.editForm.startTime).format()
          }

          if (this.editForm.endTime) {
            data.endTime = moment(this.editForm.endTime).format()
          }

          try {
            const res = await this.$http.put(`/api/classroom/checkin/${this.editForm.id}`, data)
            if (res.data.code === 200) {
              this.$message.success(this.$t('m.Update_Success'))
              this.showEditDialog = false
              this.loadCheckins()
            } else {
              this.$message.error(res.data.message || this.$t('m.Update_Failed'))
            }
          } catch (error) {
            this.$message.error(this.$t('m.Update_Failed'))
          }
        }
      })
    },

    // 重置编辑表单
    resetEditForm() {
      this.editForm = {
        id: null,
        checkinName: '',
        startTime: null,
        endTime: null
      }
      if (this.$refs.editForm) {
        this.$refs.editForm.clearValidate()
      }
    },

    // 删除签到
    async deleteCheckin(checkin) {
      this.$confirm(
        `${this.$t('m.Delete_Checkin_Confirm')}: ${checkin.checkinName}?`,
        this.$t('m.Confirm_Delete'),
        {
          confirmButtonText: this.$t('m.Confirm'),
          cancelButtonText: this.$t('m.Cancel'),
          type: 'warning'
        }
      ).then(async () => {
        try {
          const res = await this.$http.delete(`/api/classroom/checkin/${checkin.id}`)
          if (res.data.code === 200) {
            this.$message.success(this.$t('m.Delete_Success'))
            this.loadCheckins()
          } else {
            this.$message.error(res.data.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
        }
      }).catch(() => {
        // 用户取消删除
      })
    },

    showQrcode(checkin) {
      this.currentCheckin = checkin
      this.$nextTick(() => {
        this.$refs.qrcodeDisplay.open()
      })
    },

    handleQrcodeClose() {
      // 二维码关闭后的处理
    },

    resetCreateForm() {
      this.createForm = {
        classroomId: null,
        checkinName: '',
        checkinType: 'code',
        qrcodeRefreshInterval: 15,
        startTime: null,
        endTime: null
      }
      if (this.$refs.createForm) {
        this.$refs.createForm.clearValidate()
      }
    },

    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    },

    exportCheckinRecords() {
      if (!this.currentCheckin) return

      const statusMap = {
        present: '签到',
        absent: '缺勤',
        sick_leave: '病假',
        personal_leave: '事假'
      }

      let csvContent = '\uFEFF'
      csvContent += `签到记录_${this.currentCheckin.checkinName || '签到'}_${this.formatTime(this.currentCheckin.startTime)}\n\n`
      csvContent += '用户名,真实姓名,状态,签到时间\n'

      this.records.forEach(record => {
        const realName = record.classStudent?.realName || record.student?.realName || '-'
        const status = statusMap[record.status] || record.status
        const checkinTime = record.checkinTime ? this.formatTime(record.checkinTime) : '-'

        const row = [
          record.student?.username || '',
          realName,
          status,
          checkinTime
        ].map(field => `"${field}"`).join(',')
        csvContent += row + '\n'
      })

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', `签到记录_${this.currentCheckin.checkinName || '签到'}_${new Date().getTime()}.csv`)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    }
  }
}
</script>

<style scoped>
.checkin-panel {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h3 {
  font-size: 20px;
  color: #409EFF;
}

.checkin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.checkin-header h4 {
  margin: 0;
}

.actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
</style>
