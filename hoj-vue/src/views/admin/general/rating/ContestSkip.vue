<template>
  <div class="contest-skip">
    <!-- 选择比赛 -->
    <el-form label-width="120px">
      <el-form-item label="选择比赛">
        <el-select
          v-model="selectedContestId"
          placeholder="请选择比赛"
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
          刷新比赛列表
        </el-button>
      </el-form-item>
    </el-form>

    <!-- 比赛信息卡片 -->
    <el-card v-if="currentContest" shadow="hover" style="margin-bottom: 20px">
      <div slot="header">
        <span>📊 比赛信息</span>
      </div>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="比赛标题">{{ currentContest.title }}</el-descriptions-item>
        <el-descriptions-item label="比赛ID">{{ currentContest.id }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatDate(currentContest.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="currentContest.status === -1" type="info">未开始</el-tag>
          <el-tag v-else-if="currentContest.status === 0" type="warning">进行中</el-tag>
          <el-tag v-else type="success">已结束</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Rating比赛">
          <el-tag :type="currentContest.isRating ? 'success' : 'info'">
            {{ currentContest.isRating ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Skip用户数">
          <el-tag type="danger">{{ skipUsers.length }} 人</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Skip用户列表 -->
    <el-card v-if="selectedContestId" shadow="hover" style="margin-bottom: 20px">
      <div slot="header">
        <span>👥 Skip用户列表 (共{{ skipUsers.length }}人，待应用{{ pendingSkipUsersCount }}人)</span>
        <el-button
          type="primary"
          icon="el-icon-plus"
          size="small"
          style="float: right"
          @click="showBatchSkipDialog = true"
        >
          批量Skip用户
        </el-button>
      </div>

      <el-table :data="skipUsers" stripe style="width: 100%">
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="uid" label="UID" width="200" />
        <el-table-column prop="reason" label="Skip原因" />
        <el-table-column label="状态" width="100">
          <template slot-scope="{ row }">
            <el-tag v-if="row.isApplied" type="success">已应用</el-tag>
            <el-tag v-else type="warning">待应用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operatorUsername" label="操作人" width="120" />
        <el-table-column label="操作时间" width="180">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template slot-scope="{ row }">
            <el-button type="danger" size="mini" @click="handleCancelSkip(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="skipUsers.length === 0" description="暂无Skip用户" />
    </el-card>

    <!-- Rating重算 -->
    <el-card v-if="selectedContestId" shadow="hover">
      <div slot="header">
        <span>🔄 比赛Rating重算</span>
      </div>

      <!-- 显示重算状态提示 -->
      <el-alert
        v-if="pendingSkipUsersCount > 0"
        title="检测到本比赛有待应用的Skip用户"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>当前状态：</strong>有 {{ pendingSkipUsersCount }} 名Skip用户等待应用</p>
          <p style="margin-top: 10px"><strong>点击"开始重算"后将执行：</strong></p>
          <ol style="margin: 10px 0 0 20px; padding: 0">
            <li>重新计算本比赛Rating（Skip用户不计入）</li>
            <li>将Skip用户标记为"已应用"状态</li>
            <li>级联重算后续比赛</li>
          </ol>
        </template>
      </el-alert>
      <el-alert
        v-else-if="needRecalculate"
        title="⚠️ 检测到Rating数据需要重算"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>当前状态：</strong>Skip数据已被修改，但Rating尚未重新计算</p>
          <p style="margin-top: 10px">
            <strong>上次计算时间：</strong>{{ formatDateTime(currentContest?.calculatedAt) }}
          </p>
          <p style="margin-top: 5px">
            <strong>Skip最后修改：</strong>{{ formatDateTime(currentContest?.skipDataChangedAt) }}
          </p>
          <p style="margin-top: 10px; color: #E6A23C;">
            <strong>建议：</strong>点击"开始重算"按钮应用最新的Skip配置
          </p>
        </template>
      </el-alert>
      <el-alert
        v-else
        title="手动触发重算"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template slot="default">
          <p><strong>当前状态：</strong>Rating数据是最新的</p>
          <p style="margin-top: 10px"><strong>点击"开始重算"后将执行：</strong></p>
          <ol style="margin: 10px 0 0 20px; padding: 0">
            <li>重新计算本比赛Rating（包括所有参赛用户）</li>
            <li>级联重算后续比赛</li>
          </ol>
        </template>
      </el-alert>

      <el-button
        type="primary"
        :loading="recalculating"
        :disabled="recalculateLock"
        @click="handleRecalculate"
      >
        <i class="el-icon-refresh"></i> 开始重算
      </el-button>
      <el-tag v-if="recalculateLock" type="danger" style="margin-left: 10px">正在进行重算</el-tag>
    </el-card>

    <!-- 批量Skip弹窗 -->
    <el-dialog
      title="批量Skip用户"
      :visible.sync="showBatchSkipDialog"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-alert
        title="💡 批量处理提示"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <ul style="margin: 10px 0 0 20px; padding: 0">
          <li>支持一次跳过多个用户，只需重算一次</li>
          <li>重复用户会自动过滤，不会重复skip</li>
          <li>建议一次性输入所有需要skip的用户名</li>
        </ul>
      </el-alert>

      <el-form :model="skipForm" label-width="100px">
        <el-form-item label="用户名列表">
          <el-input
            type="textarea"
            v-model="skipForm.usernames"
            :rows="8"
            placeholder="输入格式说明：每行输入一个用户名。例如第一行输入zhangsan，第二行输入lisi。也可以直接从Excel复制用户名列粘贴到这里。"
          />
          <div class="form-tip">
            <i class="el-icon-info"></i>
            已输入 <strong>{{ usernameLines.length }}</strong> 个用户名（每行一个）
            <el-tag v-if="duplicateCount > 0" type="warning" size="small">
              检测到 {{ duplicateCount }} 个重复
            </el-tag>
          </div>
        </el-form-item>

        <el-form-item label="Skip原因">
          <el-select v-model="skipForm.reason" filterable allow-create style="width: 100%">
            <el-option label="代码抄袭" value="代码抄袭" />
            <el-option label="代打作弊" value="代打作弊" />
            <el-option label="AI作弊" value="AI作弊" />
            <el-option label="账号共享" value="账号共享" />
          </el-select>
        </el-form-item>

        <el-form-item label="重算设置">
          <el-checkbox v-model="skipForm.autoRecalc">
            Skip后自动重算Rating
          </el-checkbox>
          <div class="form-tip" v-if="skipForm.autoRecalc">
            <el-alert
              title="⚠️ 重要提示"
              type="warning"
              :closable="false"
              style="margin-top: 10px"
            >
              <div style="line-height: 1.8">
                <p><strong>重算将影响以下比赛：</strong></p>
                <ul style="margin: 5px 0 0 20px; padding: 0">
                  <li>从本比赛开始，<strong>级联重算所有后续比赛</strong></li>
                  <li>每个被重算的比赛都会重新计算所有参赛者的Rating</li>
                  <li>Skip的用户不计入Rating计算</li>
                  <li>预计耗时：{{ estimatedTime }}</li>
                </ul>
              </div>
            </el-alert>
          </div>
          <div class="form-tip" v-else style="margin-top: 10px">
            <el-alert
              title="💡 工作流程说明"
              type="info"
              :closable="false"
            >
              <div style="line-height: 1.8">
                <p><strong>当前设置：</strong>不勾选自动重算</p>
                <p style="margin-top: 10px"><strong>操作流程：</strong></p>
                <ol style="margin: 5px 0 0 20px; padding: 0">
                  <li>添加Skip用户（状态为"待应用"）</li>
                  <li>继续添加其他需要Skip的用户</li>
                  <li>所有用户添加完后，点击"开始重算"按钮</li>
                  <li>重算完成后，Skip用户状态变为"已应用"</li>
                </ol>
              </div>
            </el-alert>
          </div>
        </el-form-item>
      </el-form>

      <div slot="footer">
        <el-button @click="showBatchSkipDialog = false">取消</el-button>
        <el-button type="primary" :loading="skipSubmitting" @click="handleBatchSkip">
          确认Skip
        </el-button>
      </div>
    </el-dialog>

    <!-- 重算进度弹窗 -->
    <el-dialog
      title="Rating重算中..."
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
              计算中...
            </el-tag>
          </div>
        </div>

        <div class="time-info">
          状态: {{ progressStatus === 'success' ? '已完成' : progressStatus === 'exception' ? '失败' : '计算中...' }}
        </div>
      </div>

      <div slot="footer">
        <el-button @click="showProgressDialog = false" :disabled="progressStatus === ''">
          {{ progressStatus === '' ? '重算中...' : '关闭' }}
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
      return '约30秒-2分钟'
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
          throw new Error('响应解析失败: ' + responseText)
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
        this.$message.error('加载比赛列表失败: ' + error.message)
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
          this.$message.error('加载Skip用户失败: ' + error.message)
        }
      }
    },

    // 批量Skip
    async handleBatchSkip() {
      if (this.usernameLines.length === 0) {
        this.$message.warning('请输入用户名')
        return
      }

      if (!this.skipForm.reason) {
        this.$message.warning('请选择或输入Skip原因')
        return
      }

      // 显示确认提示
      const autoRecalcText = this.skipForm.autoRecalc
        ? '<b style="color: #67C23A;">✓ 自动重算</b>：Skip后自动触发重算'
        : '<b style="color: #E6A23C;">⚠️ 需手动重算</b>：Skip后需要手动点击"开始重算"按钮'

      try {
        await this.$msgbox({
          title: '确认批量Skip',
          message: `确定要Skip以下 <b>${this.usernameLines.length}</b> 个用户吗？<br><br>
            用户列表：<b>${this.usernameLines.join(', ')}</b><br><br>
            Skip原因：<b>${this.skipForm.reason}</b><br><br>
            ${autoRecalcText}<br><br>
            <b style="color: #E6A23C;">⚠️ 重要提示：</b><br>
            1. 这些用户将被标记为作弊，不参与本次比赛的Rating计算<br>
            2. 如果开启自动重算，将<b>重新计算本场比赛及后续所有比赛</b>的Rating<br>
            3. 被Skip的用户在本次比赛中Rating变化将显示为红色的"SKIP"<br>
            4. 这是一个<b>不可逆操作</b>，请谨慎操作`,
          dangerouslyUseHTMLString: true,
          confirmButtonText: '确定Skip',
          cancelButtonText: '我再想想',
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
          message += `✓ 成功添加: ${result.successUsers.length}人（状态：待应用）\n\n`
        }
        if (result.duplicatedUsers.length > 0) {
          message += `⊗ 重复: ${result.duplicatedUsers.join(', ')}\n`
        }
        if (result.failedUsers.length > 0) {
          message += `✗ 失败: ${result.failedUsers.join(', ')}`
        }

        // 根据结果决定消息类型
        let messageType = 'success'
        if (result.failedUsers.length > 0 && result.successUsers.length === 0) {
          messageType = 'error'
        } else if (result.failedUsers.length > 0) {
          messageType = 'warning'
        }

        this.$alert(message, 'Skip结果', { type: messageType })
        this.showBatchSkipDialog = false
        this.skipForm.usernames = ''

        // 刷新Skip用户列表（立即显示，状态为"待应用"）
        await this.loadSkipUsers()
        // 刷新比赛信息（获取最新的 skipDataChangedAt）
        await this.refreshContestInfo()

        // 如果没有勾选自动重算，提示用户需要手动重算
        if (result.successUsers.length > 0 && !this.skipForm.autoRecalc) {
          this.$message({
            message: 'Skip用户已添加，请点击"开始重算"按钮应用更改',
            type: 'info',
            duration: 5000
          })
        }

        // 如果需要自动重算，显示进度
        if (result.taskId) {
          this.startPollingProgress(result.taskId)
        }
      } catch (error) {
        this.$message.error('Skip失败: ' + error.message)
      } finally {
        this.skipSubmitting = false
      }
    },

    // 取消Skip
    async handleCancelSkip(row) {
      this.$msgbox({
        title: '确认取消Skip',
        message: `确定要取消用户 <b>${row.username}</b> 的Skip标记吗？<br><br>
          <b style="color: #E6A23C;">⚠️ 重要提示：</b><br>
          1. 取消Skip后，该用户将从Skip列表中移除<br>
          2. 需要手动点击<b>"开始重算"</b>按钮来重新计算Rating<br>
          3. 重算将从本场比赛开始，影响所有后续比赛<br>
          4. 建议完成所有Skip修改后，再统一点击"开始重算"<br>
          5. 这是一个<b>不可逆操作</b>，请谨慎操作`,
        dangerouslyUseHTMLString: true,
        confirmButtonText: '确定取消',
        cancelButtonText: '我再想想',
        type: 'warning',
        distinguishCancelAndClose: true
      }).then(async () => {
        try {
          this.$message.info('正在取消Skip...')
          const result = await ratingApi.cancelSkip(this.selectedContestId, [row.uid])

          // 取消成功
          this.$message.success({
            message: '取消Skip成功！请点击"开始重算"按钮应用更改',
            duration: 5000
          })
          await this.loadSkipUsers()
          // 刷新比赛信息（获取最新的 skipDataChangedAt）
          await this.refreshContestInfo()
        } catch (error) {
          const errorMsg = error.message || '未知错误'
          this.$message.error('取消Skip失败: ' + errorMsg)
        }
      }).catch(() => {
        // 用户取消
      })
    },

    // 触发重算
    async handleRecalculate() {
      this.$msgbox({
        title: '确认重算Rating',
        message: `确定要重新计算比赛的Rating吗？<br><br>
          <b style="color: #E6A23C;">⚠️ 重要提示：</b><br>
          1. 系统<b>将从本场比赛开始</b>，重新计算所有后续比赛的Rating<br>
          2. 所有参赛用户的Rating都会重新计算，分数会有变化<br>
          3. 重算过程可能需要较长时间（取决于后续比赛数量）<br>
          4. 重算期间相关比赛无法进行Rating操作<br>
          5. 这是一个<b>不可逆操作</b>，请谨慎操作`,
        dangerouslyUseHTMLString: true,
        confirmButtonText: '确定重算',
        cancelButtonText: '我再想想',
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
            this.$message.success('Skip 标记已同步')
          } catch (syncError) {
            console.warn('同步 Skip 标记失败（继续重算）:', syncError)
          }

          // 创建重算任务
          const res = await ratingApi.recalculateFromContest(this.selectedContestId)
          console.log('重算任务创建响应:', res)

          if (!res || !res.taskId) {
            this.$message.error('重算任务创建失败: 未返回任务ID')
            console.error('无效的响应:', res)
            return
          }

          this.$message.success('重算任务已创建')
          this.recalculateLock = true

          // 开始轮询进度
          this.startPollingProgress(res.taskId)
        } catch (error) {
          console.error('创建重算任务失败:', error)
          this.$message.error('创建重算任务失败: ' + error.message)
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
              this.$message.success('重算完成！')
              // 刷新Skip用户列表
              await this.loadSkipUsers()
              // 刷新比赛信息（获取最新的 calculatedAt）
              await this.refreshContestInfo()
            } else {
              this.$message.error('重算失败: ' + (result.errorMessage || '未知错误'))
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
          this.$message.error('查询进度失败: ' + error.message)
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
      if (!dateStr) return '未计算'
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
