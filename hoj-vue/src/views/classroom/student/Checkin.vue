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
              <!-- 已签到：显示签到状态 -->
              <el-tag v-if="hasCheckined(checkin.id)" :type="getCheckinStatusType(checkin.id)">
                {{ getCheckinStatusText(checkin.id) }}
              </el-tag>

              <!-- 未签到：根据时间和状态显示不同内容 -->
              <template v-else>
                <!-- 签到进行中：显示未签到状态和签到按钮 -->
                <template v-if="isCheckinActive(checkin)">
                  <el-tag type="info" size="small" style="margin-right: 8px;">
                    {{ $t('m.Not_Checkin') }}
                  </el-tag>

                  <!-- 二维码签到类型 -->
                  <el-button
                    v-if="checkin.checkinType === 'qrcode'"
                    type="success"
                    size="small"
                    @click="handleQrcodeCheckin(checkin)"
                    :loading="checkin.submitting"
                    icon="el-icon-full-screen"
                  >
                    {{ $t('m.Scan_Qrcode') }}
                  </el-button>

                  <!-- 签到码类型 -->
                  <el-button
                    v-else
                    type="primary"
                    size="small"
                    @click="handleCheckin(checkin)"
                    :loading="checkin.submitting"
                  >
                    {{ $t('m.Checkin') }}
                  </el-button>
                </template>

                <!-- 签到未开始：显示未开始 -->
                <el-tag v-else-if="!isCheckinStarted(checkin)" type="info" size="small">
                  {{ $t('m.Not_Started') }}
                </el-tag>

                <!-- 签到已结束且未签到：显示缺勤 -->
                <el-tag v-else type="danger" size="small">
                  {{ $t('m.Absent') }}
                </el-tag>
              </template>
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

    <!-- 二维码扫描对话框 -->
    <el-dialog
      :title="$t('m.Scan_Qrcode')"
      :visible.sync="showQrcodeDialog"
      width="500px"
      @close="resetQrcodeForm"
      :close-on-click-modal="false"
    >
      <!-- 扫描器区域 -->
      <div v-if="!qrcodeForm.scannedToken" class="qrcode-scanner">
        <!-- 相机错误提示 -->
        <div v-if="qrcodeForm.cameraError" class="camera-error">
          <i class="el-icon-warning" style="font-size: 48px; color: #F56C6C;"></i>
          <p style="margin-top: 10px;">{{ qrcodeForm.cameraError }}</p>
          <p style="font-size: 12px; color: #909399; margin-top: 5px;">
            请确保：1. 允许浏览器访问摄像头 2. 使用HTTPS或localhost 3. 没有其他应用占用摄像头
          </p>
        </div>

        <!-- 视频预览和扫描区域 -->
        <div v-else class="scanner-wrapper">
          <video
            ref="video"
            class="camera-video"
            autoplay
            playsinline
            muted
          ></video>
          <canvas ref="canvas" style="display: none;"></canvas>

          <!-- 扫描框覆盖层 -->
          <div class="scan-overlay">
            <div class="scan-frame">
              <div class="scan-corner top-left"></div>
              <div class="scan-corner top-right"></div>
              <div class="scan-corner bottom-left"></div>
              <div class="scan-corner bottom-right"></div>
              <p class="scan-hint">将二维码放入框内即可自动扫描</p>
            </div>
          </div>

          <!-- 加载提示 -->
          <div v-if="qrcodeForm.initializing" class="camera-loading-overlay">
            <i class="el-icon-loading"></i>
            <span>{{ $t('m.Loading_Camera') }}</span>
          </div>
        </div>
      </div>

      <!-- 扫描成功 -->
      <div v-else class="scan-success">
        <i class="el-icon-success" style="font-size: 48px; color: #67C23A;"></i>
        <p style="margin-top: 10px;">{{ $t('m.Scan_Success') }}</p>
        <el-button @click="resetScan" size="small">{{ $t('m.Rescan') }}</el-button>
      </div>

      <span slot="footer">
        <el-button @click="showQrcodeDialog = false">{{ $t('m.Cancel') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'
import studentAuth from '@/mixins/studentAuth'
import jsQR from 'jsqr'

// 确保 jsQR 被打包，不被 tree-shaking 移除
// eslint-disable-next-line no-unused-vars
const _ensureJsqrIncluded = jsQR

export default {
  name: 'Checkin',
  components: {},
  mixins: [realtimeSync, studentAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      checkins: [],
      checkinRecords: {}, // 存储每个签到表的记录 { checkinId: [records] }
      classStudents: [], // 班级学生列表
      showCheckinDialog: false,
      showQrcodeDialog: false,
      checkinForm: {
        currentCheckin: null,
        inputCode: '',
        submitting: false
      },
      qrcodeForm: {
        currentCheckin: null,
        inputToken: '',
        submitting: false,
        scannedToken: '',
        cameraError: '',
        initializing: false,
        scanning: false
      },
      // 标记是否正在加载记录，用于避免闪烁
      loadingRecords: false,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadCheckins',
        immediate: true
      }
    }
  },
  computed: {
    // 计算属性：检查学生是否已签到（用于模板显示）
    hasCheckined() {
      return function(checkinId) {
        // 优先使用后端返回的状态
        const checkin = this.checkins.find(c => c.id === checkinId)
        if (checkin && checkin.userCheckinStatus) {
          return true
        }

        // 降级到本地查询
        const records = this.checkinRecords[checkinId] || []
        const userId = this.$store.getters.userInfo?.uid
        const myRecord = records.find(r => r.uid === userId)
        return !!myRecord
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
    // 确保 jsQR 被打包（防止 tree-shaking 移除）
    // eslint-disable-next-line no-unused-vars
    const _jsqr = jsQR
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
        // 并行加载签到列表和签到记录
        const [checkinsResult] = await Promise.all([
          // 学生端使用安全API，不返回敏感信息
          this.$store.dispatch('classroom/getCheckinListForStudent', this.classroomId)
        ])

        if (checkinsResult.code === 200) {
          const newCheckins = (checkinsResult.data || []).map(c => ({ ...c, submitting: false }))
          this.checkins = newCheckins

          // 立即并行加载所有签到记录
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
    async loadAllRecords(forceRefresh = false) {
      // 学生端不需要加载全班学生列表
      // await this.loadClassStudents()  // 注释掉，学生端不需要

      // 使用 Promise.all 并行加载所有签到的记录
      const recordPromises = this.checkins.map(async (checkin) => {
        try {
          const records = await this.$store.dispatch('classroom/getCheckinRecords', checkin.id)
          return {
            checkinId: checkin.id,
            data: records
          }
        } catch (error) {
          console.error(`加载签到记录失败: ${checkin.id}`, error)
          return {
            checkinId: checkin.id,
            data: { code: 500, data: [] }
          }
        }
      })

      // 等待所有请求完成
      const results = await Promise.all(recordPromises)

      // 批量更新签到记录
      const userInfo = this.$store.getters.userInfo
      const userId = userInfo?.uid

      for (const result of results) {
        if (result.data.code === 200) {
          const newRecords = result.data.data || []
          const oldRecords = this.checkinRecords[result.checkinId] || []

          const oldMyRecord = oldRecords.find(r => r.uid === userId)
          const newMyRecord = newRecords.find(r => r.uid === userId)

          // 强制刷新或状态变化时才更新
          if (forceRefresh || !oldMyRecord || !newMyRecord ||
              !oldMyRecord !== !newMyRecord || oldMyRecord.status !== newMyRecord.status) {
            this.$set(this.checkinRecords, result.checkinId, newRecords)
          }
        }
      }

      // 合并学生列表和签到记录
      this.mergeStudentsAndRecords()
    },

    // 获取班级学生列表
    async loadClassStudents() {
      if (this.classStudents.length > 0) {
        return
      }

      try {
        const res = await this.$http.get(`/api/classroom/${this.classroomId}/students`)
        if (res.data.code === 200) {
          this.classStudents = res.data.data || []
        }
      } catch (error) {
        console.error('加载学生列表失败:', error)
        this.classStudents = []
      }
    },

    // 合并学生列表和签到记录
    // 注意：学生端只显示已签到的记录，不显示缺勤列表
    mergeStudentsAndRecords() {
      // 学生端不需要合并全班学生列表
      // 只保留真实的签到记录即可
      // 不做任何处理，避免将未签到的学生标记为缺勤
    },
    // 获取当前用户在该签到表的状态
    getCheckinStatus(checkinId) {
      // 优先使用后端返回的状态
      const checkin = this.checkins.find(c => c.id === checkinId)
      if (checkin && checkin.userCheckinStatus) {
        return checkin.userCheckinStatus
      }

      // 降级到本地查询
      const records = this.checkinRecords[checkinId] || []
      const userId = this.$store.getters.userInfo?.uid
      const myRecord = records.find(r => r.uid === userId)
      return myRecord ? myRecord.status : null
    },

    // 判断签到是否处于活跃状态（可以签到）
    isCheckinActive(checkin) {
      const now = new Date()
      const startTime = new Date(checkin.startTime)

      // 检查是否已开始
      if (now < startTime) {
        return false
      }

      // 检查是否已结束
      if (checkin.endTime) {
        const endTime = new Date(checkin.endTime)
        if (now > endTime) {
          return false
        }
      }

      // 检查签到状态（status === 1 表示进行中）
      return checkin.status === 1
    },

    // 判断签到是否已开始
    isCheckinStarted(checkin) {
      const now = new Date()
      const startTime = new Date(checkin.startTime)
      return now >= startTime
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
          // 重新加载记录（强制刷新，不检查是否变化）
          await this.loadAllRecords(true)
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

    // 处理二维码签到
    async handleQrcodeCheckin(checkin) {
      this.qrcodeForm.currentCheckin = checkin
      this.qrcodeForm.inputToken = ''
      this.qrcodeForm.scannedToken = ''
      this.qrcodeForm.cameraError = ''
      this.showQrcodeDialog = true

      // 对话框打开后启动相机
      this.$nextTick(() => {
        this.startCamera()
      })
    },

    // 新增：确认二维码签到
    async confirmQrcodeCheckin() {
      if (!this.qrcodeForm.inputToken || this.qrcodeForm.inputToken.trim() === '') {
        this.$message.warning('请输入Token')
        return
      }

      this.qrcodeForm.submitting = true
      try {
        const res = await this.$store.dispatch('classroom/submitQrcodeCheckin', {
          token: this.qrcodeForm.inputToken.trim(),
          checkinId: this.qrcodeForm.currentCheckin.id
        })

        if (res.code === 200) {
          this.$message.success(this.$t('m.Checkin_Success'))
          this.showQrcodeDialog = false
          // 重新加载记录
          await this.loadAllRecords()
        } else {
          this.$message.error(res.message || this.$t('m.Checkin_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Checkin_Failed'))
      } finally {
        this.qrcodeForm.submitting = false
      }
    },

    // 重置二维码表单
    resetQrcodeForm() {
      this.stopCamera()
      this.qrcodeForm = {
        currentCheckin: null,
        inputToken: '',
        submitting: false,
        scannedToken: '',
        cameraError: '',
        initializing: false,
        scanning: false
      }
    },

    // ==================== 相机和扫描功能 ====================

    // 启动相机
    async startCamera() {
      this.qrcodeForm.initializing = true
      this.qrcodeForm.cameraError = ''

      try {
        // 检查浏览器支持
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
          throw new Error('当前浏览器不支持摄像头访问')
        }

        // 获取摄像头流（优先使用后置摄像头）
        const stream = await navigator.mediaDevices.getUserMedia({
          video: {
            facingMode: 'environment', // 优先使用后置摄像头
            width: { ideal: 1280 },
            height: { ideal: 720 }
          }
        })

        const video = this.$refs.video
        if (!video) {
          throw new Error('视频组件未就绪')
        }

        video.srcObject = stream
        await video.play()

        this.qrcodeForm.scanning = true
        this.$message.success('相机已启动，请将二维码对准扫描框')

        // 开始扫描循环
        this.scanLoop()

      } catch (error) {
        console.error('启动相机失败:', error)
        let errorMsg = '相机启动失败'

        if (error.name === 'NotAllowedError') {
          errorMsg = '请允许浏览器访问摄像头'
        } else if (error.name === 'NotFoundError') {
          errorMsg = '未检测到摄像头设备'
        } else if (error.name === 'NotSupportedError') {
          errorMsg = '当前浏览器不支持摄像头访问，请使用 Chrome 或 Edge'
        } else if (error.name === 'NotReadableError') {
          errorMsg = '摄像头被其他应用占用或权限不足'
        } else if (error.message) {
          errorMsg = error.message
        }

        this.qrcodeForm.cameraError = errorMsg
        this.$message.error(errorMsg)
      } finally {
        this.qrcodeForm.initializing = false
      }
    },

    // 扫描循环
    async scanLoop() {
      if (!this.qrcodeForm.scanning || this.qrcodeForm.scannedToken) {
        return
      }

      try {
        const video = this.$refs.video
        const canvas = this.$refs.canvas

        if (!video || !canvas || video.readyState !== video.HAVE_ENOUGH_DATA) {
          requestAnimationFrame(() => this.scanLoop())
          return
        }

        // 将视频帧绘制到 canvas
        canvas.width = video.videoWidth
        canvas.height = video.videoHeight
        const ctx = canvas.getContext('2d', { willReadFrequently: true })
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height)

        // 使用 jsQR 识别二维码
        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
        const code = jsQR(imageData.data, imageData.width, imageData.height)

        if (code) {
          const token = this.extractTokenFromUrl(code.data)

          if (token) {
            // 先停止摄像头
            this.stopCamera()

            // 设置 token 并提交签到
            this.qrcodeForm.scannedToken = token
            this.qrcodeForm.inputToken = token

            // 显示扫描成功提示
            this.$message.success(this.$t('m.Scan_Success'))

            // 直接提交签到
            this.confirmQrcodeCheckin().then(() => {
              // 签到成功后，延迟1.5秒再关闭对话框
              setTimeout(() => {
                this.showQrcodeDialog = false
                // 重新加载签到列表（强制刷新）
                this.loadCheckins()
                // 强制刷新签到记录
                this.loadAllRecords(true)
              }, 1500)
            }).catch((error) => {
              // 失败也延迟关闭对话框
              console.error('签到提交失败:', error)
              setTimeout(() => {
                this.showQrcodeDialog = false
              }, 2000)
            })

            return
          } else {
            this.$message.warning('二维码格式不正确，请扫描正确的签到二维码')
          }
        }

        // 继续下一帧扫描
        requestAnimationFrame(() => this.scanLoop())

      } catch (error) {
        console.error('扫描循环错误:', error)
        // 出错后继续尝试
        setTimeout(() => this.scanLoop(), 1000)
      }
    },

    // 停止相机
    stopCamera() {
      this.qrcodeForm.scanning = false

      const video = this.$refs.video
      if (video && video.srcObject) {
        const tracks = video.srcObject.getTracks()
        tracks.forEach(track => track.stop())
        video.srcObject = null
      }
    },

    // 从URL中提取token
    extractTokenFromUrl(url) {
      // 如果内容为空
      if (!url || url.trim() === '') {
        this.$message.error('二维码内容为空，请扫描正确的签到二维码')
        return null
      }

      // 方法1：直接使用正则表达式从任何格式中提取token
      // token格式通常是: 数字_数字_字母数字 (例如: 16_1768705780133_3e83f2c8)
      const tokenPattern = /token=([^&\s]+)/
      const match = url.match(tokenPattern)

      if (match && match[1]) {
        return match[1]
      }

      // 方法2：尝试构造完整URL再解析
      try {
        // 如果是相对路径，添加当前域名
        let fullUrl = url
        if (!url.startsWith('http://') && !url.startsWith('https://')) {
          // 获取当前页面的协议和主机
          const protocol = window.location.protocol
          const host = window.location.host
          fullUrl = `${protocol}//${host}${url.startsWith('/') ? '' : '/'}${url}`
        }

        const urlObj = new URL(fullUrl)
        const token = urlObj.searchParams.get('token')
        if (token) {
          return token
        }

        this.$message.warning('二维码格式不正确，未找到token参数')
      } catch (e) {
        // 方法3：如果是直接扫描的token字符串（包含下划线）
        // Token通常包含下划线，例如：16_1768705780133_3e83f2c8
        if (url && url.includes('_')) {
          // 提取token部分（去除可能的前缀）
          const parts = url.split('_')
          if (parts.length >= 2) {
            // 找到最后一个包含token=的部分
            for (let i = parts.length - 1; i >= 0; i--) {
              const part = parts[i]
              // 检查是否是字母数字混合（token的最后部分）
              if (/^[a-zA-Z0-9]+$/.test(part)) {
                const token = parts.slice(Math.max(0, i - 2), i + 1).join('_')
                return token
              }
            }
          }

          return url
        } else {
          this.$message.error('二维码格式不正确，请扫描正确的签到二维码')
        }
      }

      return null
    },

    // 重新扫描
    resetScan() {
      this.qrcodeForm.scannedToken = ''
      this.qrcodeForm.cameraError = ''
      this.startCamera()
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

/* 二维码扫描器样式 */
.qrcode-scanner {
  text-align: center;
  padding: 20px 0;
}

.scanner-wrapper {
  position: relative;
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}

.camera-video {
  width: 100%;
  height: auto;
  display: block;
}

.scan-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.scan-frame {
  position: relative;
  width: 250px;
  height: 250px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 12px;
}

.scan-corner {
  position: absolute;
  width: 30px;
  height: 30px;
  border-color: #409EFF;
  border-style: solid;
}

.scan-corner.top-left {
  top: -2px;
  left: -2px;
  border-width: 4px 0 0 4px;
  border-radius: 12px 0 0 0;
}

.scan-corner.top-right {
  top: -2px;
  right: -2px;
  border-width: 4px 4px 0 0;
  border-radius: 0 12px 0 0;
}

.scan-corner.bottom-left {
  bottom: -2px;
  left: -2px;
  border-width: 0 0 4px 4px;
  border-radius: 0 0 0 12px;
}

.scan-corner.bottom-right {
  bottom: -2px;
  right: -2px;
  border-width: 0 4px 4px 0;
  border-radius: 0 0 12px 0;
}

.scan-hint {
  position: absolute;
  bottom: -40px;
  left: 50%;
  transform: translateX(-50%);
  color: white;
  font-size: 14px;
  white-space: nowrap;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.camera-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  color: white;
}

.camera-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #F56C6C;
  text-align: center;
}

.scan-success {
  text-align: center;
  padding: 40px 20px;
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


<style scoped>
.student-checkin {
  padding: 20px;
}

/* 二维码扫描器样式 */
.qrcode-scanner {
  text-align: center;
  padding: 20px 0;
}

.scanner-wrapper {
  position: relative;
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}

.camera-video {
  width: 100%;
  height: auto;
  display: block;
}

.scan-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.scan-frame {
  position: relative;
  width: 250px;
  height: 250px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 12px;
}

.scan-corner {
  position: absolute;
  width: 30px;
  height: 30px;
  border-color: #409EFF;
  border-style: solid;
}

.scan-corner.top-left {
  top: -2px;
  left: -2px;
  border-width: 4px 0 0 4px;
  border-radius: 12px 0 0 0;
}

.scan-corner.top-right {
  top: -2px;
  right: -2px;
  border-width: 4px 4px 0 0;
  border-radius: 0 12px 0 0;
}

.scan-corner.bottom-left {
  bottom: -2px;
  left: -2px;
  border-width: 0 0 4px 4px;
  border-radius: 0 0 0 12px;
}

.scan-corner.bottom-right {
  bottom: -2px;
  right: -2px;
  border-width: 0 4px 4px 0;
  border-radius: 0 0 12px 0;
}

.scan-hint {
  position: absolute;
  bottom: -40px;
  left: 50%;
  transform: translateX(-50%);
  color: white;
  font-size: 14px;
  white-space: nowrap;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.camera-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  color: white;
}

.camera-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #F56C6C;
  text-align: center;
}

.scan-success {
  text-align: center;
  padding: 40px 20px;
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
