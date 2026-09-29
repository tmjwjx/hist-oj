<template>
  <div class="contest-skip">
    <!-- 选择比赛 -->
    <el-form label-width="120px">
      <el-form-item :label="$t('m.Rating_Select_Contest')">
        <el-select
          v-model="selectedContestId"
          :placeholder="$t('m.Rating_Please_Select_Contest')"
          filterable
          @change="handleContestChange"
          style="width: 400px"
        >
          <el-option
            v-for="contest in contests"
            :key="contest.id"
            :label="`${contest.title} (${formatDate(contest.startTime)})`"
            :value="contest.id"
          />
        </el-select>
        <el-button type="primary" icon="el-icon-refresh" @click="loadContests" style="margin-left: 10px">
          {{ $t('m.Rating_Refresh_Contest_List') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- 比赛信息卡片 -->
    <el-card v-if="currentContest" shadow="hover" style="margin-bottom: 20px">
      <div slot="header">
        <span>{{ $t('m.Rating_Contest_Info') }}</span>
      </div>
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('m.Rating_Contest_Title')">{{ currentContest.title }}</el-descriptions-item>
        <el-descriptions-item :label="$t('m.Rating_Contest_ID')">{{ currentContest.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('m.Rating_Start_Time')">{{ formatDate(currentContest.startTime) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('m.Rating_Status')">
          <el-tag v-if="currentContest.status === -1" type="info">{{ $t('m.Rating_Not_Started') }}</el-tag>
          <el-tag v-else-if="currentContest.status === 0" type="warning">{{ $t('m.Rating_In_Progress') }}</el-tag>
          <el-tag v-else type="success">{{ $t('m.Rating_Finished') }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('m.Rating_Is_Rating_Contest')">
          <el-tag :type="currentContest.isRating ? 'success' : 'info'">
            {{ currentContest.isRating ? $t('m.Rating_Yes') : $t('m.Rating_No') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('m.Rating_Skip_User_Count')">
          <el-tag type="danger">{{ skipUsers.length }} {{ $t('m.Rating_People_Unit') }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Skip用户列表 -->
    <el-card v-if="selectedContestId" shadow="hover" style="margin-bottom: 20px">
      <div slot="header">
        <span>{{ $t('m.Rating_Skip_User_List_Header', { total: skipUsers.length, pending: pendingSkipUsersCount }) }}</span>
        <el-button
          type="primary"
          icon="el-icon-plus"
          size="small"
          style="float: right"
          @click="showBatchSkipDialog = true"
        >
          {{ $t('m.Rating_Batch_Skip_Users') }}
        </el-button>
      </div>

      <el-table :data="skipUsers" stripe style="width: 100%">
        <el-table-column prop="username" :label="$t('m.Rating_Username_Label')" width="150" />
        <el-table-column prop="uid" label="UID" width="200" />
        <el-table-column prop="reason" :label="$t('m.Rating_Skip_Reason')" />
        <el-table-column :label="$t('m.Rating_Status')" width="100">
          <template slot-scope="{ row }">
            <el-tag v-if="row.isApplied" type="success">{{ $t('m.Rating_Applied') }}</el-tag>
            <el-tag v-else type="warning">{{ $t('m.Rating_Pending_Apply') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operatorUsername" :label="$t('m.Rating_Operator')" width="120" />
        <el-table-column :label="$t('m.Rating_Operation_Time')" width="180">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('m.Rating_Actions')" width="100">
          <template slot-scope="{ row }">
            <el-button type="danger" size="mini" @click="handleCancelSkip(row)">{{ $t('m.Rating_Cancel') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="skipUsers.length === 0" :description="$t('m.Rating_No_Skip_Users')" />
    </el-card>

    <!-- Rating重算 -->
    <el-card v-if="selectedContestId" shadow="hover">
      <div slot="header">
        <span>{{ $t('m.Rating_Contest_Rating_Recalc') }}</span>
      </div>

      <!-- 显示重算状态提示 -->
      <el-alert
        v-if="pendingSkipUsersCount > 0"
        :title="$t('m.Rating_Pending_Skip_Detected')"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>{{ $t('m.Rating_Current_Status') }}</strong>{{ $t('m.Rating_Pending_Waiting_Apply', { count: pendingSkipUsersCount }) }}</p>
          <p style="margin-top: 10px"><strong>{{ $t('m.Rating_Start_Recalc_Will_Execute') }}</strong></p>
          <ol style="margin: 10px 0 0 20px; padding: 0">
            <li>{{ $t('m.Rating_Recalc_Step_1') }}</li>
            <li>{{ $t('m.Rating_Recalc_Step_2') }}</li>
            <li>{{ $t('m.Rating_Cascade_Recalculate_Following') }}</li>
          </ol>
        </template>
      </el-alert>
      <el-alert
        v-else-if="needRecalculate"
        :title="$t('m.Rating_Rating_Data_Needs_Recalc')"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>{{ $t('m.Rating_Current_Status') }}</strong>{{ $t('m.Rating_Skip_Data_Changed') }}</p>
          <p style="margin-top: 10px">
            <strong>{{ $t('m.Rating_Last_Calculated_At') }}</strong>{{ formatDateTime(currentContest?.calculatedAt) }}
          </p>
          <p style="margin-top: 5px">
            <strong>{{ $t('m.Rating_Skip_Last_Modified') }}</strong>{{ formatDateTime(currentContest?.skipDataChangedAt) }}
          </p>
          <p style="margin-top: 10px; color: #E6A23C;">
            <strong>{{ $t('m.Rating_Suggestion') }}</strong>{{ $t('m.Rating_Apply_Latest_Skip') }}
          </p>
        </template>
      </el-alert>
      <el-alert
        v-else
        :title="$t('m.Rating_Manual_Recalc')"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>{{ $t('m.Rating_Current_Status') }}</strong>{{ $t('m.Rating_Rating_Up_To_Date') }}</p>
          <p style="margin-top: 10px"><strong>{{ $t('m.Rating_Start_Recalc_Will_Execute') }}</strong></p>
          <ol style="margin: 10px 0 0 20px; padding: 0">
            <li>{{ $t('m.Rating_Recalc_Step_1_All') }}</li>
            <li>{{ $t('m.Rating_Cascade_Recalculate_Following') }}</li>
          </ol>
        </template>
      </el-alert>

      <el-button
        type="primary"
        :loading="recalculating"
        :disabled="recalculateLock"
        @click="handleRecalculate"
      >
        <i class="el-icon-refresh"></i> {{ $t('m.Rating_Start_Recalculate') }}
      </el-button>
      <el-tag v-if="recalculateLock" type="danger" style="margin-left: 10px">{{ $t('m.Rating_Recalculating') }}</el-tag>
    </el-card>

    <!-- 批量Skip弹窗 -->
    <el-dialog
      :title="$t('m.Rating_Batch_Skip_Users')"
      :visible.sync="showBatchSkipDialog"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-alert
        :title="$t('m.Rating_Batch_Tip')"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <ul style="margin: 10px 0 0 20px; padding: 0">
          <li>{{ $t('m.Rating_Batch_Tip_1') }}</li>
          <li>{{ $t('m.Rating_Batch_Tip_2') }}</li>
          <li>{{ $t('m.Rating_Batch_Tip_3') }}</li>
        </ul>
      </el-alert>

      <el-form :model="skipForm" label-width="100px">
        <el-form-item :label="$t('m.Rating_Username_List')">
          <el-input
            type="textarea"
            v-model="skipForm.usernames"
            :rows="8"
            :placeholder="$t('m.Rating_Username_Format_Tip')"
          />
          <div class="form-tip">
            <i class="el-icon-info"></i>
            {{ $t('m.Rating_Entered_Prefix') }} <strong>{{ usernameLines.length }}</strong> {{ $t('m.Rating_Entered_Suffix') }}
            <el-tag v-if="duplicateCount > 0" type="warning" size="small">
              {{ $t('m.Rating_Duplicates_Detected', { count: duplicateCount }) }}
            </el-tag>
          </div>
        </el-form-item>

        <el-form-item :label="$t('m.Rating_Skip_Reason')">
          <el-select v-model="skipForm.reason" filterable allow-create style="width: 100%">
            <el-option :label="$t('m.Rating_Reason_Code_Plagiarism')" value="代码抄袭" />
            <el-option :label="$t('m.Rating_Reason_Proxy_Competing')" value="代打作弊" />
            <el-option :label="$t('m.Rating_Reason_AI_Cheating')" value="AI作弊" />
            <el-option :label="$t('m.Rating_Reason_Account_Sharing')" value="账号共享" />
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('m.Rating_Recalc_Settings')">
          <el-checkbox v-model="skipForm.autoRecalc">
            {{ $t('m.Rating_Auto_Recalc_After_Skip') }}
          </el-checkbox>
          <div class="form-tip" v-if="skipForm.autoRecalc">
            <el-alert
              :title="$t('m.Rating_Important_Tip')"
              type="warning"
              :closable="false"
              style="margin-top: 10px"
            >
              <div style="line-height: 1.8">
                <p><strong>{{ $t('m.Rating_Recalc_Will_Affect') }}</strong></p>
                <ul style="margin: 5px 0 0 20px; padding: 0">
                  <li>{{ $t('m.Rating_From_This_Contest') }}<strong>{{ $t('m.Rating_Cascade_Recalc_All_Following') }}</strong></li>
                  <li>{{ $t('m.Rating_Recalc_Affect_2') }}</li>
                  <li>{{ $t('m.Rating_Recalc_Affect_3') }}</li>
                  <li>{{ $t('m.Rating_Estimated_Time', { time: estimatedTime }) }}</li>
                </ul>
              </div>
            </el-alert>
          </div>
          <div class="form-tip" v-else style="margin-top: 10px">
            <el-alert
              :title="$t('m.Rating_Workflow_Expl')"
              type="info"
              :closable="false"
            >
              <div style="line-height: 1.8">
                <p><strong>{{ $t('m.Rating_Current_Settings') }}</strong>{{ $t('m.Rating_Auto_Recalc_Unchecked') }}</p>
                <p style="margin-top: 10px"><strong>{{ $t('m.Rating_Operation_Steps') }}</strong></p>
                <ol style="margin: 5px 0 0 20px; padding: 0">
                  <li>{{ $t('m.Rating_Workflow_Step_1') }}</li>
                  <li>{{ $t('m.Rating_Workflow_Step_2') }}</li>
                  <li>{{ $t('m.Rating_Workflow_Step_3') }}</li>
                  <li>{{ $t('m.Rating_Workflow_Step_4') }}</li>
                </ol>
              </div>
            </el-alert>
          </div>
        </el-form-item>
      </el-form>

      <div slot="footer">
        <el-button @click="showBatchSkipDialog = false">{{ $t('m.Rating_Cancel') }}</el-button>
        <el-button type="primary" :loading="skipSubmitting" @click="handleBatchSkip">
          {{ $t('m.Rating_Confirm_Skip') }}
        </el-button>
      </div>
    </el-dialog>

    <!-- 重算进度弹窗 -->
    <el-dialog
      :title="$t('m.Rating_Recalculating_Title')"
      :visible.sync="showProgressDialog"
      width="500px"
      :close-on-click-modal="false"
    >
      <div class="progress-container">
        <el-progress :percentage="progress" :status="progressStatus" />

        <div class="contest-list">
          <div v-for="contest in progressContests" :key="contest.id" class="contest-item">
            <el-icon :class="getStatusClass(contest.status)">
              <component :is="getStatusIcon(contest.status)" />
            </el-icon>
            <span>{{ contest.title }}</span>
            <el-tag v-if="contest.status === 'calculating'" size="mini" type="warning">
              {{ $t('m.Rating_Calculating') }}
            </el-tag>
          </div>
        </div>

        <div class="time-info">
          {{ $t('m.Rating_Status_Label') }}{{ progressStatus === 'success' ? $t('m.Rating_Completed') : progressStatus === 'exception' ? $t('m.Rating_Failed') : $t('m.Rating_Calculating') }}
        </div>
      </div>

      <div slot="footer">
        <el-button @click="showProgressDialog = false" :disabled="progressStatus === ''">
          {{ progressStatus === '' ? $t('m.Rating_Recalculating_Dots') : $t('m.Rating_Close') }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import ratingApi from '@/common/rating-api'

export default {
  name: 'ContestSkip',
  data() {
    return {
      contests: [],
      selectedContestId: null,
      currentContest: null,
      skipUsers: [],
      showBatchSkipDialog: false,
      showProgressDialog: false,
      skipSubmitting: false,
      recalculating: false,
      recalculateLock: false,
      skipForm: {
        usernames: '',
        reason: '代码抄袭',
        autoRecalc: false
      },
      progress: 0,
      progressStatus: '',
      progressContests: [],
      progressTimer: null
    }
  },
  computed: {
    usernameLines() {
      return this.skipForm.usernames
        .split('\n')
        .map(line => line.trim())
        .filter(line => line)
    },
    duplicateCount() {
      const unique = new Set(this.usernameLines)
      return this.usernameLines.length - unique.size
    },
    estimatedTime() {
      return this.$t('m.Rating_Estimated_Time_Value')
    },
    pendingSkipUsersCount() {
      return this.skipUsers.filter(u => !u.isApplied).length
    },
    needRecalculate() {
      // 判断是否需要重算：skip数据修改时间 > rating计算时间
      if (!this.currentContest) return false
      const calculatedAt = this.currentContest.calculatedAt
      const skipChangedAt = this.currentContest.skipDataChangedAt
      if (!calculatedAt || !skipChangedAt) return false
      return new Date(skipChangedAt) > new Date(calculatedAt)
    }
  },
  mounted() {
    this.loadContests()
  },
  methods: {
    // 加载比赛列表
    async loadContests() {
      try {
        // 获取最近的比赛
        const response = await this.$axios.get('/api/admin/contest/get-contest-list', {
          params: { limit: 50 }
        })

        console.log('=== 比赛列表响应 ===')
        console.log('response.data:', response.data)
        console.log('allContests count:', response.data?.data?.records?.length)

        const allContests = response.data?.data?.records || []

        if (allContests.length === 0) {
          this.contests = []
          return
        }

        // 获取比赛的rating状态 - 使用原生fetch避免axios拦截器问题
        const contestIds = allContests.map(c => Number(c.id))
        console.log('=== 准备发送请求 ===')
        console.log('allContests数量:', allContests.length)
        console.log('contestIds数组:', contestIds)
        console.log('contestIds数组类型:', Array.isArray(contestIds))
        console.log('第一个元素:', contestIds[0], '类型:', typeof contestIds[0])

        const requestBody = { contestIds }
        console.log('requestBody对象:', requestBody)
        console.log('requestBody.contestIds类型:', Array.isArray(requestBody.contestIds))

        const requestBodyStr = JSON.stringify(requestBody)
        console.log('请求体 JSON字符串:')
        console.log(requestBodyStr)
        console.log('JSON字符串长度:', requestBodyStr.length)

        // 验证JSON格式
        let parsedBack
        try {
          parsedBack = JSON.parse(requestBodyStr)
          console.log('解析回的对象:', parsedBack)
          console.log('parsedBack.contestIds类型:', Array.isArray(parsedBack.contestIds))
        } catch (e) {
          console.error('JSON格式验证失败:', e)
        }

        // 使用原生fetch API
        console.log('开始发送fetch请求...')
        const fetchResponse = await fetch('/api/rating/contest/batch', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: requestBodyStr
        })
        console.log('fetch响应状态:', fetchResponse.status, fetchResponse.statusText)

        let ratingResponse
        let responseText
        try {
          responseText = await fetchResponse.text()
          console.log('原始响应文本:', responseText)
          ratingResponse = JSON.parse(responseText)
        } catch (e) {
          console.error('解析响应失败:', e)
          console.log('响应状态:', fetchResponse.status)
          console.log('响应头:', Object.fromEntries(fetchResponse.headers.entries()))
          throw new Error(this.$t('m.Rating_Parse_Response_Failed') + responseText)
        }
        console.log('Rating状态响应:', ratingResponse)

        // 构建isRating映射
        const ratingMap = ratingResponse.data || {}
        console.log('ratingMap:', ratingMap)

        // 过滤出rated比赛，并合并rating信息
        this.contests = allContests
          .filter(contest => {
            const ratingInfo = ratingMap[contest.id] || ratingMap[String(contest.id)]
            return ratingInfo && ratingInfo.isRating === true
          })
          .map(contest => {
            const ratingInfo = ratingMap[contest.id] || ratingMap[String(contest.id)]
            return {
              ...contest,
              isRating: ratingInfo.isRating
            }
          })

        console.log('最终过滤结果:', this.contests.length, '个rated比赛')

        // 如果URL中有contestId，自动加载该比赛的数据
        const contestIdFromUrl = this.$route.query.contestId
        if (contestIdFromUrl && this.contests.some(c => c.id === parseInt(contestIdFromUrl))) {
          console.log('检测到URL中的contestId:', contestIdFromUrl)
          this.selectedContestId = parseInt(contestIdFromUrl)
          // 手动触发handleContestChange
          await this.handleContestChange()
        }
      } catch (error) {
        console.error('加载比赛列表失败:', error)
        this.$message.error(this.$t('m.Rating_Load_Contests_Failed') + error.message)
      }
    },

    // 比赛选择变化
    async handleContestChange() {
      console.log('=== handleContestChange 被调用 ===')
      console.log('selectedContestId:', this.selectedContestId)

      // 更新URL参数（不保留历史记录）
      if (this.selectedContestId) {
        this.$router.replace({
          query: { contestId: this.selectedContestId }
        }).catch(err => {
          // 忽略路由导航重复的错误
          if (err.name !== 'NavigationDuplicated') {
            console.error('更新URL参数失败:', err)
          }
        })
      } else {
        this.$router.replace({ query: {} }).catch(err => {
          if (err.name !== 'NavigationDuplicated') {
            console.error('清除URL参数失败:', err)
          }
        })
      }

      this.currentContest = this.contests.find(c => c.id === this.selectedContestId)
      console.log('currentContest:', this.currentContest)

      // 获取完整的比赛信息（包括 calculatedAt 和 skipDataChangedAt）
      if (this.selectedContestId) {
        try {
          const contestInfo = await ratingApi.getContestInfo(this.selectedContestId)
          console.log('比赛详细信息:', contestInfo)
          // 合并信息到 currentContest
          if (this.currentContest && contestInfo) {
            this.currentContest = {
              ...this.currentContest,
              ...contestInfo
            }
          }
        } catch (error) {
          console.error('获取比赛详细信息失败:', error)
        }
      }

      await this.loadSkipUsers()
    },

    // 加载Skip用户
    async loadSkipUsers() {
      console.log('=== loadSkipUsers 开始执行 ===')
      console.log('selectedContestId:', this.selectedContestId)

      if (!this.selectedContestId) {
        console.log('selectedContestId 为空，跳过')
        return
      }

      try {
        console.log('开始调用 API...')
        const res = await ratingApi.getContestSkipUsers(this.selectedContestId)
        console.log('=== Skip用户API响应 ===')
        console.log('原始响应:', res)
        console.log('响应类型:', typeof res)
        console.log('是否为数组:', Array.isArray(res))
        console.log('第一条数据:', res && res[0])
        if (res && res[0]) {
          console.log('isApplied:', res[0].isApplied, '类型:', typeof res[0].isApplied)
          console.log('operatorUsername:', res[0].operatorUsername)
          console.log('createdAt:', res[0].createdAt)
        }
        this.skipUsers = res || []
        console.log('✓ Skip用户列表已更新，总数:', this.skipUsers.length)
        console.log('待应用用户数:', this.skipUsers.filter(u => !u.isApplied).length)

        // 检查是否有已应用的 Skip 用户，如果有则自动同步标记
        const appliedUsers = this.skipUsers.filter(u => u.isApplied)
        if (appliedUsers.length > 0) {
          console.log(`检测到 ${appliedUsers.length} 个已应用的 Skip 用户，正在同步标记...`)
          try {
            await ratingApi.syncContestSkipFlag(this.selectedContestId)
            console.log('✓ Skip 标记同步成功')
            // 静默同步，不显示消息（避免打扰用户）
          } catch (syncError) {
            console.warn('⚠️ 同步 Skip 标记失败:', syncError)
            // 同步失败不影响后续操作，静默处理
          }
        }
      } catch (error) {
        console.error('❌ 加载Skip用户失败:', error)
        // 如果是取消请求，不显示错误
        if (error.message !== 'cancel') {
          this.$message.error(this.$t('m.Rating_Load_Skip_Users_Failed') + error.message)
        }
      }
    },

    // 批量Skip
    async handleBatchSkip() {
      if (this.usernameLines.length === 0) {
        this.$message.warning(this.$t('m.Rating_Enter_Username'))
        return
      }

      if (!this.skipForm.reason) {
        this.$message.warning(this.$t('m.Rating_Select_Or_Input_Reason'))
        return
      }

      // 显示确认提示
      const autoRecalcText = this.skipForm.autoRecalc
        ? this.$t('m.Rating_Auto_Recalc_Text')
        : this.$t('m.Rating_Manual_Recalc_Text')

      try {
        await this.$msgbox({
          title: this.$t('m.Rating_Confirm_Batch_Skip'),
          message: this.$t('m.Rating_Confirm_Batch_Skip_Message', {
            count: this.usernameLines.length,
            users: this.usernameLines.join(', '),
            reason: this.skipForm.reason,
            autoRecalcText: autoRecalcText
          }),
          dangerouslyUseHTMLString: true,
          confirmButtonText: this.$t('m.Rating_Confirm_Skip'),
          cancelButtonText: this.$t('m.Rating_Let_Me_Think'),
          type: 'warning',
          distinguishCancelAndClose: true
        })
      } catch {
        // 用户取消
        return
      }

      this.skipSubmitting = true
      try {
        const result = await ratingApi.batchSkipUsers(
          this.selectedContestId,
          this.usernameLines,
          this.skipForm.reason,
          this.skipForm.autoRecalc
        )

        // 显示结果
        let message = ''
        if (result.successUsers.length > 0) {
          message += this.$t('m.Rating_Add_Success_Count', { count: result.successUsers.length })
        }
        if (result.duplicatedUsers.length > 0) {
          message += this.$t('m.Rating_Duplicated_List', { users: result.duplicatedUsers.join(', ') })
        }
        if (result.failedUsers.length > 0) {
          message += this.$t('m.Rating_Failed_List', { users: result.failedUsers.join(', ') })
        }

        // 根据结果决定消息类型
        let messageType = 'success'
        if (result.failedUsers.length > 0 && result.successUsers.length === 0) {
          messageType = 'error'
        } else if (result.failedUsers.length > 0) {
          messageType = 'warning'
        }

        this.$alert(message, this.$t('m.Rating_Skip_Result'), { type: messageType })
        this.showBatchSkipDialog = false
        this.skipForm.usernames = ''

        // 刷新Skip用户列表（立即显示，状态为"待应用"）
        await this.loadSkipUsers()
        // 刷新比赛信息（获取最新的 skipDataChangedAt）
        await this.refreshContestInfo()

        // 如果没有勾选自动重算，提示用户需要手动重算
        if (result.successUsers.length > 0 && !this.skipForm.autoRecalc) {
          this.$message({
            message: this.$t('m.Rating_Skip_Added_Click_Recalc'),
            type: 'info',
            duration: 5000
          })
        }

        // 如果需要自动重算，显示进度
        if (result.taskId) {
          this.startPollingProgress(result.taskId)
        }
      } catch (error) {
        this.$message.error(this.$t('m.Rating_Skip_Failed') + error.message)
      } finally {
        this.skipSubmitting = false
      }
    },

    // 取消Skip
    async handleCancelSkip(row) {
      this.$msgbox({
        title: this.$t('m.Rating_Confirm_Cancel_Skip'),
        message: this.$t('m.Rating_Confirm_Cancel_Skip_Message', { username: row.username }),
        dangerouslyUseHTMLString: true,
        confirmButtonText: this.$t('m.Rating_Confirm_Cancel'),
        cancelButtonText: this.$t('m.Rating_Let_Me_Think'),
        type: 'warning',
        distinguishCancelAndClose: true
      }).then(async () => {
        try {
          this.$message.info(this.$t('m.Rating_Cancelling_Skip'))
          const result = await ratingApi.cancelSkip(this.selectedContestId, [row.uid])

          // 取消成功
          this.$message.success({
            message: this.$t('m.Rating_Cancel_Skip_Success'),
            duration: 5000
          })
          await this.loadSkipUsers()
          // 刷新比赛信息（获取最新的 skipDataChangedAt）
          await this.refreshContestInfo()
        } catch (error) {
          const errorMsg = error.message || this.$t('m.Rating_Unknown_Error')
          this.$message.error(this.$t('m.Rating_Cancel_Skip_Failed') + errorMsg)
        }
      }).catch(() => {
        // 用户取消
      })
    },

    // 触发重算
    async handleRecalculate() {
      this.$msgbox({
        title: this.$t('m.Rating_Confirm_Recalc'),
        message: this.$t('m.Rating_Confirm_Recalc_Message'),
        dangerouslyUseHTMLString: true,
        confirmButtonText: this.$t('m.Rating_Confirm_Recalc_Button'),
        cancelButtonText: this.$t('m.Rating_Let_Me_Think'),
        type: 'warning',
        distinguishCancelAndClose: true
      }).then(async () => {
        this.recalculating = true
        try {
          // 先同步 Skip 标记到 rating_history 表（立即生效）
          try {
            console.log('正在同步 Skip 标记...')
            await ratingApi.syncContestSkipFlag(this.selectedContestId)
            console.log('Skip 标记同步成功')
            this.$message.success(this.$t('m.Rating_Skip_Synced'))
          } catch (syncError) {
            console.warn('同步 Skip 标记失败（继续重算）:', syncError)
          }

          // 创建重算任务
          const res = await ratingApi.recalculateFromContest(this.selectedContestId)
          console.log('重算任务创建响应:', res)

          if (!res || !res.taskId) {
            this.$message.error(this.$t('m.Rating_Task_Create_Failed_No_ID'))
            console.error('无效的响应:', res)
            return
          }

          this.$message.success(this.$t('m.Rating_Task_Created'))
          this.recalculateLock = true

          // 开始轮询进度
          this.startPollingProgress(res.taskId)
        } catch (error) {
          console.error('创建重算任务失败:', error)
          this.$message.error(this.$t('m.Rating_Task_Create_Failed') + error.message)
        } finally {
          this.recalculating = false
        }
      })
    },

    // 开始轮询进度
    startPollingProgress(taskId) {
      console.log('开始轮询进度, taskId:', taskId)
      this.showProgressDialog = true
      this.progressStatus = ''
      this.progress = 0

      let isFirstQuery = true

      this.progressTimer = setInterval(async () => {
        try {
          const result = await ratingApi.getRecalculateProgress(taskId)
          console.log('进度查询结果:', result)
          this.updateProgress(result)

          // 检查任务状态
          if (result.status === 'completed' || result.status === 'failed') {
            clearInterval(this.progressTimer)

            // 设置进度为100%
            this.progress = 100
            this.progressStatus = result.status === 'completed' ? 'success' : 'exception'
            this.recalculateLock = false

            // 显示完成消息
            if (result.status === 'completed') {
              this.$message.success(this.$t('m.Rating_Recalc_Complete'))
              // 刷新Skip用户列表
              await this.loadSkipUsers()
              // 刷新比赛信息（获取最新的 calculatedAt）
              await this.refreshContestInfo()
            } else {
              this.$message.error(this.$t('m.Rating_Recalc_Failed') + (result.errorMessage || this.$t('m.Rating_Unknown_Error')))
            }

            // 如果是第一次查询就完成了（快速完成），延迟关闭进度条
            if (isFirstQuery) {
              console.log('任务快速完成，延迟关闭进度条')
              setTimeout(() => {
                console.log('关闭进度条')
              }, 3000) // 3秒后自动关闭（但用户可以手动关闭）
            }

            return
          }

          isFirstQuery = false
        } catch (error) {
          console.error('查询进度失败:', error)
          clearInterval(this.progressTimer)
          this.progressStatus = 'exception'
          this.recalculateLock = false
          this.$message.error(this.$t('m.Rating_Query_Progress_Failed') + error.message)
        }
      }, 1000)
    },

    // 更新进度
    updateProgress(result) {
      if (result.totalContests > 0) {
        this.progress = Math.floor((result.processedContests / result.totalContests) * 100)
      }
      this.progressContests = result.contests || []
    },

    // 获取状态图标
    getStatusIcon(status) {
      switch (status) {
        case 'completed':
          return 'el-icon-check'
        case 'calculating':
          return 'el-icon-loading'
        default:
          return 'el-icon-time'
      }
    },

    // 获取状态样式
    getStatusClass(status) {
      switch (status) {
        case 'completed':
          return 'status-success'
        case 'calculating':
          return 'status-running'
        default:
          return 'status-pending'
      }
    },

    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    },

    // 格式化日期时间（用于显示计算时间和修改时间）
    formatDateTime(dateStr) {
      if (!dateStr) return this.$t('m.Rating_Not_Calculated')
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    },

    // 刷新比赛信息（获取最新的 calculatedAt 和 skipDataChangedAt）
    async refreshContestInfo() {
      if (!this.selectedContestId || !this.currentContest) return

      try {
        const contestInfo = await ratingApi.getContestInfo(this.selectedContestId)
        console.log('刷新后的比赛信息:', contestInfo)
        // 合并信息到 currentContest
        if (contestInfo) {
          this.currentContest = {
            ...this.currentContest,
            ...contestInfo
          }
        }
      } catch (error) {
        console.error('刷新比赛信息失败:', error)
      }
    }
  },
  beforeDestroy() {
    if (this.progressTimer) {
      clearInterval(this.progressTimer)
    }
  }
}
</script>

<style scoped>
.contest-skip {
  padding: 20px;
}

.form-tip {
  margin-top: 5px;
  color: #909399;
  font-size: 12px;
}

.progress-container {
  padding: 20px 0;
}

.contest-list {
  margin-top: 20px;
  max-height: 300px;
  overflow-y: auto;
}

.contest-item {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #ebeef5;
}

.contest-item:last-child {
  border-bottom: none;
}

.contest-item .el-icon {
  margin-right: 10px;
}

.contest-item span {
  flex: 1;
}

.time-info {
  margin-top: 20px;
  text-align: center;
  color: #909399;
}

.status-success {
  color: #67c23a;
}

.status-running {
  color: #e6a23c;
}

.status-pending {
  color: #909399;
}
</style>
