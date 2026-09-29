<template>
  <div class="personal-adjust">
    <el-form ref="form" :model="form" :rules="rules" label-width="120px" @submit.native.prevent>
      <el-form-item :label="$t('m.Rating_Username_Label')" prop="username">
        <el-input
          v-model="form.username"
          :placeholder="$t('m.Rating_Enter_Adjust_Username')"
          clearable
          @blur="fetchUserInfo"
        >
          <template slot="append">
            <el-button icon="el-icon-search" @click="fetchUserInfo" :loading="loadingUserInfo">
              {{ $t('m.Rating_Query') }}
            </el-button>
          </template>
        </el-input>
      </el-form-item>

      <!-- 用户信息预览 -->
      <el-alert
        v-if="userInfo"
        :title="$t('m.Rating_Current_User') + ': ' + form.username + ' | Rating: ' + displayCurrentRating"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <div slot>
          <el-tag :style="{ color: getRatingColor(userInfo.rating) }" size="medium">
            {{ getRatingName(userInfo.rating) }}
          </el-tag>
        </div>
      </el-alert>

      <el-form-item :label="$t('m.Rating_Rating_Change')" prop="ratingChange">
        <el-input-number
          v-model="form.ratingChange"
          :step="10"
          controls-position="right"
          style="width: 200px"
        />
        <span class="form-tip">
          <el-tag
            :type="form.ratingChange > 0 ? 'success' : 'danger'"
            size="small"
            effect="plain"
          >
            {{ form.ratingChange > 0 ? '+' : '' }}{{ form.ratingChange }}
          </el-tag>
        </span>
      </el-form-item>

      <el-form-item :label="$t('m.Rating_Adjusted_Rating')" v-if="userInfo">
        <div>
          <span :style="{ color: getRatingColor(expectedRating), fontSize: '24px', fontWeight: 'bold' }">
            {{ expectedRating }}
          </span>
          <el-tag :style="{ color: getRatingColor(expectedRating), marginLeft: '10px' }" size="medium">
            {{ getRatingName(expectedRating) }}
          </el-tag>
        </div>
      </el-form-item>

      <el-form-item :label="$t('m.Rating_Operation_Reason')" prop="reason">
        <el-select
          v-model="form.reason"
          :placeholder="$t('m.Rating_Select_Or_Input_Operation_Reason')"
          filterable
          allow-create
          style="width: 100%"
        >
          <el-option :label="$t('m.Rating_Reason_AI_Cheating_In_Contest')" value="比赛中使用AI作弊" />
          <el-option :label="$t('m.Rating_Reason_Account_Violation')" value="账号违规" />
          <el-option :label="$t('m.Rating_Reason_Proxy_Competing')" value="代打作弊" />
          <el-option :label="$t('m.Rating_Reason_Vuln_Reward')" value="发现系统漏洞奖励" />
          <el-option :label="$t('m.Rating_Reason_Code_Contribution_Reward')" value="贡献代码奖励" />
          <el-option :label="$t('m.Rating_Reason_Other')" value="其他原因" />
        </el-select>
      </el-form-item>

      <el-form-item :label="$t('m.Rating_Related_Contest_Optional')">
        <el-select
          v-model="form.relatedContestId"
          filterable
          clearable
          :placeholder="$t('m.Rating_Select_Related_Contest')"
          style="width: 100%"
          :loading="loadingContests"
        >
          <el-option
            v-for="contest in contests"
            :key="contest.id"
            :label="`#${contest.id} ${contest.title}`"
            :value="contest.id"
          />
        </el-select>
        <div class="form-tip">{{ $t('m.Rating_Related_Contest_Tip') }}</div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          <i class="el-icon-edit"></i> {{ $t('m.Rating_Confirm_Adjust') }}
        </el-button>
        <el-button @click="handleReset">
          <i class="el-icon-refresh-left"></i> {{ $t('m.Rating_Reset') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- 调整历史记录 -->
    <el-divider></el-divider>
    <h3>{{ $t('m.Rating_Recent_Adjustments') }}</h3>
    <el-table :data="history" v-loading="loadingHistory" stripe style="width: 100%">
      <el-table-column prop="username" :label="$t('m.Rating_Username_Label')" width="150" />
      <el-table-column :label="$t('m.Rating_Rating_Change')" width="150">
        <template slot-scope="{ row }">
          <el-tag :type="row.rating_change > 0 ? 'success' : 'danger'" size="small">
            {{ row.rating_change > 0 ? '+' : '' }}{{ row.rating_change }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Rating_Change_Detail')" width="200">
        <template slot-scope="{ row }">
          <span style="color: #909399">{{ row.old_rating }} → {{ row.new_rating }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="reason" :label="$t('m.Rating_Operation_Reason')" />
      <el-table-column :label="$t('m.Rating_Related_Contest')" width="140">
        <template slot-scope="{ row }">
          <span v-if="row.relatedContestId">#{{ row.relatedContestId }}</span>
          <span v-else style="color: #909399">-</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Rating_Status')" width="110">
        <template slot-scope="{ row }">
          <el-tag v-if="row.isCanceled" type="info" size="small">{{ $t('m.Rating_Revoked') }}</el-tag>
          <el-tag v-else type="success" size="small">{{ $t('m.Rating_Active') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="operator_uid" :label="$t('m.Rating_Operator')" width="120" />
      <el-table-column :label="$t('m.Rating_Operation_Time')" width="180">
        <template slot-scope="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Rating_Actions')" width="130">
        <template slot-scope="{ row }">
          <el-button
            type="danger"
            size="mini"
            plain
            :disabled="row.isCanceled || cancelingId === row.id"
            :loading="cancelingId === row.id"
            @click="handleCancelAdjust(row)"
          >
            {{ $t('m.Rating_Cancel_Adjustment') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <div style="text-align: right; margin-top: 10px">
      <el-button type="text" icon="el-icon-refresh" @click="fetchHistory">{{ $t('m.Rating_Refresh') }}</el-button>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { getRatingColor, getRatingName } from '@/common/rating-utils'
import ratingApi from '@/common/rating-api'

export default {
  name: 'PersonalAdjust',
  data() {
    return {
      form: {
        username: '',
        ratingChange: 0,
        reason: '',
        relatedContestId: null
      },
      rules: {
        username: [{ required: true, message: this.$t('m.Rating_Enter_Username'), trigger: 'blur' }],
        ratingChange: [
          { required: true, message: this.$t('m.Rating_Enter_Rating_Change'), trigger: 'blur' },
          {
            validator: (rule, value, callback) => {
              if (value === 0) {
                callback(new Error(this.$t('m.Rating_Rating_Change_Not_Zero')))
              } else {
                callback()
              }
            },
            trigger: 'blur'
          }
        ],
        reason: [{ required: true, message: this.$t('m.Rating_Enter_Or_Select_Reason'), trigger: 'change' }]
      },
      userInfo: null,
      loadingUserInfo: false,
      submitting: false,
      history: [],
      loadingHistory: false,
      contests: [],
      loadingContests: false,
      cancelingId: null
    }
  },
  computed: {
    currentRating() {
      if (!this.userInfo || this.userInfo.rating === undefined || this.userInfo.rating === null) {
        return null
      }
      const rating = Number(this.userInfo.rating)
      return Number.isNaN(rating) ? null : rating
    },
    displayCurrentRating() {
      if (this.currentRating === null) {
        return 'N/A'
      }
      return this.currentRating
    },
    expectedRating() {
      if (this.currentRating === null) {
        return 0
      }
      const newRating = this.currentRating + this.form.ratingChange
      return Math.max(0, newRating) // 最低0分
    }
  },
  mounted() {
    this.loadRatedContests()
    this.fetchHistory()
  },
  methods: {
    getRatingColor,
    getRatingName,

    async loadRatedContests() {
      this.loadingContests = true
      try {
        const response = await axios.get('/api/admin/contest/get-contest-list', {
          params: { limit: 100 }
        })
        const allContests = response.data?.data?.records || []
        if (allContests.length === 0) {
          this.contests = []
          return
        }

        const contestIds = allContests
          .map(c => Number(c.id))
          .filter(id => Number.isInteger(id) && id > 0)

        const ratingMap = await ratingApi.getBatchContestInfo(contestIds)
        this.contests = allContests
          .filter(contest => {
            const ratingInfo = ratingMap[contest.id] || ratingMap[String(contest.id)]
            return ratingInfo && ratingInfo.isRating === true
          })
          .map(contest => ({
            id: Number(contest.id),
            title: contest.title || contest.name || this.$t('m.Rating_Untitled_Contest')
          }))
      } catch (error) {
        this.$message.warning(this.$t('m.Rating_Load_Contests_Fallback'))
      } finally {
        this.loadingContests = false
      }
    },

    // 查询用户信息
    async fetchUserInfo() {
      if (!this.form.username) {
        this.$message.warning(this.$t('m.Rating_Enter_Username_First'))
        return
      }

      this.loadingUserInfo = true
      try {
        const res = await axios.get('/api/rating/user/' + this.form.username)
        console.log('API返回的完整响应:', res)
        console.log('响应类型:', typeof res)
        console.log('是否有code字段:', res.code !== undefined)

        // 检查响应格式并提取数据
        if (res && res.data && res.data.data && res.data.data.rating !== undefined) {
          // 嵌套格式: { data: { code: 200, message: "success", data: { rating: 1356, ... } } }
          this.userInfo = res.data.data
        } else if (res && res.data && res.data.rating !== undefined) {
          // 标准格式: { code: 200, message: "success", data: { rating: 1356, ... } }
          this.userInfo = res.data
        } else if (res && res.rating !== undefined) {
          // 已处理格式: { rating: 1356, ... } (被拦截器处理过)
          this.userInfo = res
        } else {
          console.error('无法解析响应，res:', res)
          throw new Error(this.$t('m.Rating_Response_Format_Invalid'))
        }

        console.log('设置的userInfo:', this.userInfo)
        this.$message.success(this.$t('m.Rating_Query_Success'))
      } catch (error) {
        console.error('查询失败详细错误:', error)
        this.$message.error(this.$t('m.Rating_Query_User_Failed') + (error.response?.data?.message || error.message || this.$t('m.Rating_Unknown_Error')))
        this.userInfo = null
      } finally {
        this.loadingUserInfo = false
      }
    },

    // 提交调整
    async handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) {
          return false
        }

        if (!this.userInfo) {
          this.$message.warning(this.$t('m.Rating_Query_User_First'))
          return
        }

        this.submitting = true
        try {
          const payload = {
            username: this.form.username,
            ratingChange: this.form.ratingChange,
            reason: this.form.reason
          }
          if (this.form.relatedContestId) {
            payload.relatedContestId = this.form.relatedContestId
          }

          const adjustResponse = await axios.post('/api/rating/admin/adjust', payload)
          const adjustResult = adjustResponse?.data?.data || adjustResponse?.data || adjustResponse || {}
          const oldRating = adjustResult.oldRating ?? this.currentRating ?? 0
          const newRating = adjustResult.newRating ?? this.expectedRating
          const realChange = adjustResult.ratingChange ?? this.form.ratingChange

          this.$message.success(this.$t('m.Rating_Adjust_Success'))
          this.$notify({
            title: this.$t('m.Rating_Adjust_Success_Title'),
            message: `${this.form.username}: ${oldRating} → ${newRating} (${realChange > 0 ? '+' : ''}${realChange})`,
            type: 'success',
            duration: 5000
          })

          // 刷新用户信息和历史记录
          await this.fetchUserInfo()
          await this.fetchHistory()

          // 重置表单
          this.handleReset()
        } catch (error) {
          this.$message.error(this.$t('m.Rating_Adjust_Failed') + (error.response?.data?.message || error.message || this.$t('m.Rating_Unknown_Error')))
        } finally {
          this.submitting = false
        }
      })
    },

    // 重置表单
    handleReset() {
      this.$refs.form.resetFields()
      this.userInfo = null
    },

    // 查询历史记录
    async fetchHistory() {
      this.loadingHistory = true
      try {
        const response = await axios.get('/api/rating/admin/history', {
          params: { page: 1, limit: 20 }
        })
        this.history = response.data.data.records || []
      } catch (error) {
        console.error('查询历史记录失败:', error)
        this.$message.warning(this.$t('m.Rating_Query_History_Failed'))
      } finally {
        this.loadingHistory = false
      }
    },

    // 取消一条手动调整
    async handleCancelAdjust(row) {
      if (!row || !row.id) return
      if (row.isCanceled) {
        this.$message.info(this.$t('m.Rating_Adjust_Already_Revoked'))
        return
      }

      try {
        await this.$confirm(
          this.$t('m.Rating_Confirm_Cancel_Adjust_Message', {
            username: row.username,
            change: (row.rating_change > 0 ? '+' : '') + row.rating_change,
            reason: row.reason
          }),
          this.$t('m.Rating_Confirm_Cancel_Adjust'),
          {
            confirmButtonText: this.$t('m.Rating_Confirm_Cancel'),
            cancelButtonText: this.$t('m.Rating_Let_Me_Think'),
            type: 'warning'
          }
        )
      } catch (e) {
        return
      }

      this.cancelingId = row.id
      try {
        await axios.post('/api/rating/admin/adjust/cancel', {
          adjustmentId: row.id
        })
        this.$message.success(this.$t('m.Rating_Cancel_Adjust_Success'))
        await this.fetchHistory()
        if (this.form.username && this.form.username === row.username) {
          await this.fetchUserInfo()
        }
      } catch (error) {
        this.$message.error(this.$t('m.Rating_Cancel_Adjust_Failed') + (error.response?.data?.message || error.message || this.$t('m.Rating_Unknown_Error')))
      } finally {
        this.cancelingId = null
      }
    },

    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    }
  }
}
</script>

<style scoped>
.personal-adjust {
  padding: 20px;
}

.form-tip {
  margin-left: 10px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
</style>
