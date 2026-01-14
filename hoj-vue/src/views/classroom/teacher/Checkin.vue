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
              <el-tag :type="checkin.status === 1 ? 'success' : 'info'">
                {{ checkin.status === 1 ? $t('m.In_Progress') : $t('m.Ended') }}
              </el-tag>
              <el-button size="small" @click="viewRecords(checkin)">
                {{ $t('m.View_Records') }}
              </el-button>
              <el-button v-if="checkin.status === 1" size="small" type="warning" @click="endCheckin(checkin)">
                {{ $t('m.End') }}
              </el-button>
            </div>
          </div>
          <p class="code">{{ $t('m.Checkin_Code') }}: {{ checkin.checkinCode }}</p>
        </el-card>
      </el-timeline-item>
    </el-timeline>

    <!-- 创建签到对话框 -->
    <el-dialog :title="$t('m.Create_Checkin')" :visible.sync="showCreateDialog" width="500px">
      <el-form :model="createForm" :rules="rules" ref="createForm" label-width="120px">
        <el-form-item :label="$t('m.Checkin_Name')" prop="checkinName">
          <el-input v-model="createForm.checkinName" />
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

    <!-- 查看签到记录对话框 -->
    <el-dialog :title="$t('m.Checkin_Records')" :visible.sync="showRecordsDialog" width="800px">
      <div style="margin-bottom: 15px;">
        <el-button type="success" icon="el-icon-download" @click="exportCheckinRecords">
          导出签到记录
        </el-button>
      </div>
      <!-- 移除 v-loading 避免轮询时闪烁 -->
      <el-table :data="records" stripe>
        <el-table-column prop="student.username" :label="$t('m.Username')" />
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
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'

export default {
  name: 'Checkin',
  mixins: [realtimeSync],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      checkins: [],
      showCreateDialog: false,
      showRecordsDialog: false,
      records: [],
      recordsLoading: false,
      createForm: {
        classroomId: null,
        checkinName: '',
        startTime: null,
        endTime: null
      },
      rules: {
        startTime: [{ required: true, message: this.$t('m.Required'), trigger: 'change' }]
      },
      // 实时同步配置
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
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.checkins.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getCheckinList', this.classroomId)
        if (res.code === 200) {
          const newCheckins = res.data || []

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.checkins)
          const newDataString = JSON.stringify(newCheckins)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
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
            startTime: moment(this.createForm.startTime).format()
          }
          // 只有当 endTime 有值时才添加
          if (this.createForm.endTime) {
            data.endTime = moment(this.createForm.endTime).format()
          }
          try {
            const res = await this.$store.dispatch('classroom/createCheckin', data)
            if (res.code === 200) {
              this.$message.success(this.$t('m.Create_Success'))
              this.showCreateDialog = false
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
      this.recordsLoading = true
      this.showRecordsDialog = true
      try {
        const res = await this.$store.dispatch('classroom/getCheckinRecords', checkin.id)
        if (res.code === 200) {
          this.records = res.data || []
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.recordsLoading = false
      }
    },
    async updateRecord(record) {
      try {
        const res = await this.$store.dispatch('classroom/updateCheckinRecord', {
          recordId: record.id,
          status: record.status
        })
        if (res.code === 200) {
          this.$message.success(this.$t('m.Update_Success'))
        } else {
          this.$message.error(res.message || this.$t('m.Update_Failed'))
        }
      } catch (error) {
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
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    exportCheckinRecords() {
      if (!this.currentCheckin) return

      // 状态映射
      const statusMap = {
        present: '签到',
        absent: '缺勤',
        sick_leave: '病假',
        personal_leave: '事假'
      }

      // 创建CSV内容
      let csvContent = '\uFEFF' // UTF-8 BOM
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

      // 创建Blob并下载
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
  gap: 10px;
  align-items: center;
}

.code {
  color: #409EFF;
  font-weight: bold;
  margin: 10px 0 0 0;
}
</style>
