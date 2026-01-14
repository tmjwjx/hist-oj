<template>
  <div class="student-checkin">
    <el-card class="checkin-list-card">
      <div slot="header">
        <span>{{ $t('m.Checkin_List') }}</span>
      </div>
      <!-- 移除 v-loading 避免轮询时闪烁 -->
      <div>
        <el-empty v-if="checkins.length === 0" :description="$t('m.No_Checkin_Available')" />
        <div v-else class="checkin-items">
          <el-card
            v-for="checkin in checkins"
            :key="checkin.id"
            class="checkin-item"
            shadow="hover"
          >
            <div class="checkin-header">
              <div class="checkin-info">
                <h4>{{ checkin.checkinName || $t('m.Checkin') }}</h4>
                <el-tag :type="checkin.status === 1 ? 'success' : 'info'" size="small">
                  {{ checkin.status === 1 ? $t('m.In_Progress') : $t('m.Ended') }}
                </el-tag>
              </div>
              <div class="checkin-time">
                <i class="el-icon-time"></i>
                {{ formatTime(checkin.startTime) }} - {{ checkin.endTime ? formatTime(checkin.endTime) : $t('m.Unlimited') }}
              </div>
            </div>

            <!-- 检查学生是否已签到 -->
            <div class="checkin-status">
              <el-tag v-if="getCheckinStatus(checkin.id)" :type="getCheckinStatusType(checkin.id)">
                {{ getCheckinStatusText(checkin.id) }}
              </el-tag>
              <el-button
                v-else-if="checkin.status === 1"
                type="primary"
                size="small"
                @click="handleCheckin(checkin)"
                :loading="checkin.submitting"
              >
                {{ $t('m.Checkin') }}
              </el-button>
              <el-tag v-else type="info" size="small">
                {{ $t('m.Checkin_Ended') }}
              </el-tag>
            </div>
          </el-card>
        </div>
      </div>
    </el-card>

    <!-- 签到码输入对话框 -->
    <el-dialog
      :title="$t('m.Enter_Checkin_Code')"
      :visible.sync="showCheckinDialog"
      width="400px"
      @close="resetCheckinForm"
    >
      <el-form @submit.native.prevent="confirmCheckin">
        <el-form-item :label="$t('m.Checkin_Code')">
          <el-input
            v-model="checkinForm.inputCode"
            :placeholder="$t('m.Please_Enter_Checkin_Code')"
            maxlength="6"
            show-word-limit
            @keyup.enter.native="confirmCheckin"
            ref="checkinCodeInput"
          />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showCheckinDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="confirmCheckin" :loading="checkinForm.submitting">
          {{ $t('m.Confirm_Checkin') }}
        </el-button>
      </span>
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
      checkinRecords: {}, // 存储每个签到表的记录 { checkinId: [records] }
      showCheckinDialog: false,
      checkinForm: {
        currentCheckin: null,
        inputCode: '',
        submitting: false
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
          const newCheckins = (res.data || []).map(c => ({ ...c, submitting: false }))

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.checkins)
          const newDataString = JSON.stringify(newCheckins)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.checkins = newCheckins
          }

          // 加载每个签到的记录
          await this.loadAllRecords()
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
    async loadAllRecords() {
      for (const checkin of this.checkins) {
        try {
          const records = await this.$store.dispatch('classroom/getCheckinRecords', checkin.id)
          if (records.code === 200) {
            this.$set(this.checkinRecords, checkin.id, records.data || [])
          }
        } catch (error) {
          console.error(`加载签到记录失败: ${checkin.id}`, error)
        }
      }
    },
    // 获取当前用户在该签到表的状态
    getCheckinStatus(checkinId) {
      const records = this.checkinRecords[checkinId] || []
      // 从 Vuex store 获取当前用户信息
      const userId = this.$store.getters.userInfo?.uuid || this.$store.getters.userId
      const myRecord = records.find(r => r.userId === userId || r.uid === userId)
      return myRecord ? myRecord.status : null
    },
    getCheckinStatusType(checkinId) {
      const status = this.getCheckinStatus(checkinId)
      const map = {
        present: 'success',
        absent: 'danger',
        sick_leave: 'warning',
        personal_leave: 'info'
      }
      return map[status] || ''
    },
    getCheckinStatusText(checkinId) {
      const status = this.getCheckinStatus(checkinId)
      const map = {
        present: this.$t('m.Present'),
        absent: this.$t('m.Absent'),
        sick_leave: this.$t('m.Sick_Leave'),
        personal_leave: this.$t('m.Personal_Leave')
      }
      return map[status] || status
    },
    async handleCheckin(checkin) {
      // 打开签到码输入对话框
      this.checkinForm.currentCheckin = checkin
      this.checkinForm.inputCode = ''
      this.showCheckinDialog = true

      // 对话框打开后自动聚焦输入框
      this.$nextTick(() => {
        if (this.$refs.checkinCodeInput) {
          this.$refs.checkinCodeInput.focus()
        }
      })
    },
    async confirmCheckin() {
      if (!this.checkinForm.inputCode || this.checkinForm.inputCode.trim() === '') {
        this.$message.warning(this.$t('m.Please_Enter_Checkin_Code'))
        return
      }

      this.checkinForm.submitting = true
      try {
        const res = await this.$store.dispatch('classroom/studentCheckin', {
          checkinId: this.checkinForm.currentCheckin.id,
          checkinCode: this.checkinForm.inputCode.trim()
        })
        if (res.code === 200) {
          this.$message.success(this.$t('m.Checkin_Success'))
          this.showCheckinDialog = false
          // 重新加载记录
          await this.loadAllRecords()
        } else {
          this.$message.error(res.message || this.$t('m.Checkin_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Checkin_Failed'))
      } finally {
        this.checkinForm.submitting = false
      }
    },
    resetCheckinForm() {
      this.checkinForm = {
        currentCheckin: null,
        inputCode: '',
        submitting: false
      }
    },
    formatTime(time) {
      return time ? moment(time).format('YYYY-MM-DD HH:mm') : '-'
    }
  }
}
</script>

<style scoped>
.student-checkin {
  padding: 20px;
}

.checkin-list-card {
  margin-bottom: 20px;
}

.checkin-items {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.checkin-item {
  border-radius: 8px;
  /* 移除 transition 避免轮询时闪烁 */
}

.checkin-item:hover {
  /* 移除 transform 避免轮询时闪烁 */
}

.checkin-header {
  margin-bottom: 15px;
}

.checkin-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.checkin-info h4 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.checkin-time {
  font-size: 13px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.checkin-status {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  min-height: 32px;
}
</style>
