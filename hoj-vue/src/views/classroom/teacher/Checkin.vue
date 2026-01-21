<template>
  <div class="checkin-panel classroom-theme">
    <div class="header">
      <h3>{{ $t('m.Checkin_Management') }}</h3>
      <button class="classroom-btn classroom-btn-primary" @click="showCreateDialog = true">
        <i class="el-icon-plus"></i>
        <span>{{ $t('m.Create_Checkin') }}</span>
      </button>
    </div>

    <div class="checkin-timeline classroom-timeline">
      <div v-for="checkin in checkins" :key="checkin.id" class="timeline-item">
        <div class="timeline-time">{{ formatTime(checkin.startTime) }}</div>
        <div class="timeline-content">
          <div class="classroom-card checkin-card-single-row">
            <!-- 左侧：标题和状态标签 -->
            <div class="checkin-left">
              <h4 class="checkin-title">{{ checkin.checkinName || $t('m.Checkin') }}</h4>
              <div class="checkin-tags">
                <span
                  :class="checkin.checkinType === 'qrcode' ? 'classroom-tag classroom-tag-warning' : 'classroom-tag classroom-tag-primary'"
                >
                  <i :class="checkin.checkinType === 'qrcode' ? 'el-icon-full-screen' : 'el-icon-key'"></i>
                  {{ checkin.checkinType === 'qrcode' ? $t('m.Qrcode_Checkin') : $t('m.Checkin_Code_Checkin') }}
                </span>

                <span v-if="checkin.checkinType === 'code'" class="classroom-tag classroom-tag-info">
                  {{ $t('m.Checkin_Code') }}: {{ checkin.checkinCode }}
                </span>

                <span :class="checkin.status === 1 ? 'classroom-tag classroom-tag-success' : 'classroom-tag classroom-tag-secondary'">
                  {{ checkin.status === 1 ? $t('m.In_Progress') : $t('m.Ended') }}
                </span>
              </div>
            </div>

            <!-- 右侧：操作按钮 -->
            <div class="checkin-right">
              <button
                v-if="checkin.checkinType === 'qrcode' && checkin.status === 1"
                class="classroom-btn classroom-btn-success"
                @click="showQrcode(checkin)"
              >
                <i class="el-icon-full-screen"></i>
                <span>{{ $t('m.Show_Qrcode') }}</span>
              </button>

              <button class="classroom-btn classroom-btn-secondary" @click="viewRecords(checkin)">
                <i class="el-icon-view"></i>
                <span>{{ $t('m.View_Records') }}</span>
              </button>

              <button v-if="checkin.status === 1" class="classroom-btn classroom-btn-warning" @click="endCheckin(checkin)">
                <i class="el-icon-video-pause"></i>
                <span>{{ $t('m.End') }}</span>
              </button>

              <button class="classroom-btn classroom-btn-primary" @click="editCheckin(checkin)">
                <i class="el-icon-edit"></i>
                <span>{{ $t('m.Edit') }}</span>
              </button>

              <button class="classroom-btn classroom-btn-danger" @click="deleteCheckin(checkin)">
                <i class="el-icon-delete"></i>
                <span>{{ $t('m.Delete') }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建签到对话框 -->
    <el-dialog :title="$t('m.Create_Checkin')" :visible.sync="showCreateDialog" width="500px" @close="resetCreateForm" custom-class="classroom-dialog">
      <el-form :model="createForm" :rules="rules" ref="createForm" label-width="140px">
        <el-form-item :label="$t('m.Checkin_Name')" prop="checkinName">
          <el-input v-model="createForm.checkinName" :placeholder="$t('m.Please_Enter_Checkin_Name')" class="classroom-input" />
        </el-form-item>

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
          <span style="margin-left: 10px; color: var(--classroom-text-secondary); font-size: 12px;">
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
        <el-button @click="showCreateDialog = false" class="classroom-btn classroom-btn-secondary">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createCheckin" class="classroom-btn classroom-btn-primary">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 编辑签到对话框 -->
    <el-dialog :title="$t('m.Edit_Checkin')" :visible.sync="showEditDialog" width="500px" @close="resetEditForm" custom-class="classroom-dialog">
      <el-form :model="editForm" :rules="rules" ref="editForm" label-width="140px">
        <el-form-item :label="$t('m.Checkin_Name')" prop="checkinName">
          <el-input v-model="editForm.checkinName" :placeholder="$t('m.Please_Enter_Checkin_Name')" class="classroom-input" />
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
        <el-button @click="showEditDialog = false" class="classroom-btn classroom-btn-secondary">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="updateCheckin" class="classroom-btn classroom-btn-primary">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 查看签到记录对话框 -->
    <el-dialog :title="$t('m.Checkin_Records')" :visible.sync="showRecordsDialog" width="900px" custom-class="classroom-dialog">
      <div class="records-header">
        <button class="classroom-btn classroom-btn-success" @click="exportCheckinRecords">
          <i class="el-icon-download"></i>
          <span>导出签到记录</span>
        </button>
      </div>
      <el-table :data="records" stripe class="classroom-table">
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
@import '../classroom-theme.css';

.checkin-panel {
  padding: 24px;
  background: var(--classroom-bg);
  min-height: 100vh;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(74, 144, 226, 0.08);
}

.header h3 {
  font-size: 22px;
  color: var(--classroom-text);
  margin: 0;
  font-weight: 700;
}

/* 时间线样式 */
.checkin-timeline {
  position: relative;
}

.timeline-item {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
  position: relative;
}

.timeline-item::before {
  content: '';
  position: absolute;
  left: 89px;
  top: 40px;
  bottom: -40px;
  width: 2px;
  background: var(--classroom-primary-lighter);
}

.timeline-item:last-child::before {
  display: none;
}

.timeline-time {
  min-width: 80px;
  text-align: right;
  padding-top: 16px;
  color: var(--classroom-text-secondary);
  font-size: 13px;
  font-weight: 500;
}

.timeline-content {
  flex: 1;
}

/* 单排卡片布局 */
.checkin-card-single-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: white;
  border: 1px solid var(--classroom-border);
  border-radius: 12px;
  gap: 20px;
}

.checkin-left {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 16px;
}

.checkin-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--classroom-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.checkin-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.checkin-right {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

/* 已结束标签样式 */
.classroom-tag-secondary {
  background: #E0E0E0;
  color: #757575;
  border: 1px solid #BDBDBD;
}

.records-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}

/* 对话框样式 */
.classroom-dialog .el-dialog__header {
  background: #E3F2FD;
  border-bottom: 2px solid var(--classroom-primary);
}

.classroom-dialog .el-dialog__title {
  color: var(--classroom-text);
  font-weight: 600;
}

/* 响应式 */
@media (max-width: 768px) {
  .timeline-item {
    flex-direction: column;
    gap: 12px;
  }

  .timeline-item::before {
    display: none;
  }

  .timeline-time {
    text-align: left;
    padding-top: 0;
  }

  .checkin-card-single-row {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }

  .checkin-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .checkin-title {
    white-space: normal;
  }

  .checkin-right {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .checkin-right .classroom-btn {
    flex: 1;
    min-width: 120px;
  }
}
</style>
