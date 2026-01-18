<template>
  <el-dialog
    :title="$t('m.Qrcode_Checkin')"
    :visible="dialogVisible"
    @update:visible="val => dialogVisible = val"
    width="500px"
    @close="handleClose"
    :close-on-click-modal="false"
  >
    <div class="qrcode-container" v-loading="loading">
      <!-- 二维码显示区域 -->
      <div class="qrcode-wrapper" v-if="qrcodeData.qrcodeToken">
        <div class="qrcode-box" :class="{ 'expired': isExpired }">
          <img
            v-if="qrcodeImageUrl"
            :src="qrcodeImageUrl"
            alt="签到二维码"
            class="qrcode-image"
          />
          <div v-else class="qrcode-placeholder">
            <i class="el-icon-loading"></i>
            <span>生成中...</span>
          </div>
        </div>
      </div>

      <!-- 倒计时显示 -->
      <div class="countdown-wrapper">
        <el-alert
          :type="isExpired ? 'error' : countdown <= 5 ? 'warning' : 'success'"
          :closable="false"
          show-icon
        >
          <template slot="title">
            <div class="countdown-content">
              <i class="el-icon-time"></i>
              <span v-if="isExpired">{{ $t('m.Qrcode_Expired') }}</span>
              <span v-else>{{ $t('m.Auto_Refresh_In') }}: {{ countdown }}s</span>
            </div>
          </template>
        </el-alert>
      </div>

      <!-- 签到信息 -->
      <div class="checkin-info">
        <p><strong>{{ $t('m.Checkin_Name') }}:</strong> {{ checkinName }}</p>
        <p><strong>{{ $t('m.Qrcode_Refresh_Interval') }}:</strong> {{ refreshInterval }}s</p>
      </div>

      <!-- 已签到人数（可选） -->
      <div class="stats" v-if="checkedInCount !== null">
        <el-tag type="success" size="large">
          <i class="el-icon-user"></i>
          {{ $t('m.Checked_In_Count') }}: {{ checkedInCount }}
        </el-tag>
      </div>
    </div>

    <span slot="footer" class="dialog-footer">
      <el-button @click="handleClose">{{ $t('m.Close') }}</el-button>
      <el-button
        type="primary"
        @click="handleManualRefresh"
        :loading="refreshing"
        :icon="refreshing ? 'el-icon-loading' : 'el-icon-refresh'"
      >
        {{ $t('m.Refresh_Now') }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import VueQrcode from '@chenfengyuan/vue-qrcode'
import QRCode from 'qrcode'

export default {
  name: 'QrcodeDisplay',
  components: {
    VueQrcode
  },
  props: {
    checkinId: {
      type: [String, Number],
      required: true
    },
    checkinName: {
      type: String,
      default: ''
    },
    refreshInterval: {
      type: Number,
      default: 15
    }
  },
  data() {
    return {
      dialogVisible: false,
      visible: false,
      loading: false,
      refreshing: false,
      qrcodeData: {
        qrcodeToken: '',
        qrcodeUrl: '',
        expiresAt: '',
        refreshIn: 0
      },
      qrcodeImageUrl: '',
      countdown: 0,
      isExpired: false,
      timer: null,
      checkedInCount: null
    }
  },
  watch: {
    countdown(newVal) {
      if (newVal <= 0 && !this.isExpired) {
        // 倒计时结束，自动刷新
        this.autoRefresh()
      }
    }
  },
  methods: {
    async open() {
      this.dialogVisible = true
      this.visible = true
      await this.loadQrcode()
      this.startCountdown()
    },

    async loadQrcode() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getQrcodeToken', this.checkinId)

        if (res.code === 200) {
          this.qrcodeData = res.data || {}
          // 生成二维码图片URL
          await this.generateQrcodeImage()
          this.countdown = this.qrcodeData.refreshIn || this.refreshInterval
          this.isExpired = false
        } else {
          this.$message.error(res.message || this.$t('m.Load_Failed'))
        }
      } catch (error) {
        console.error('加载二维码失败:', error)
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
      }
    },

    async generateQrcodeImage() {
      try {
        const canvas = document.createElement('canvas')
        const dataToEncode = this.qrcodeData.qrcodeUrl || this.qrcodeData.qrcodeToken

        await QRCode.toCanvas(canvas, dataToEncode, {
          width: 280,
          margin: 2,
          color: {
            dark: '#000000',
            light: '#FFFFFF'
          }
        })
        this.qrcodeImageUrl = canvas.toDataURL('image/png')
      } catch (error) {
        console.error('生成二维码图片失败:', error)
        // 降级方案：使用在线API
        const encodedUrl = encodeURIComponent(this.qrcodeData.qrcodeUrl || this.qrcodeData.qrcodeToken)
        this.qrcodeImageUrl = `https://api.qrserver.com/v1/create-qr-code/?size=280x280&data=${encodedUrl}`
      }
    },

    startCountdown() {
      // 清除旧定时器
      if (this.timer) {
        clearInterval(this.timer)
      }

      // 启动倒计时
      this.timer = setInterval(() => {
        if (this.countdown > 0) {
          this.countdown--
        }
      }, 1000)
    },

    async autoRefresh() {
      // 自动刷新
      await this.refreshQrcode()
    },

    async handleManualRefresh() {
      await this.refreshQrcode()
    },

    async refreshQrcode() {
      this.refreshing = true
      try {
        const res = await this.$store.dispatch('classroom/refreshQrcode', this.checkinId)
        if (res.code === 200) {
          this.qrcodeData = res.data || {}
          await this.generateQrcodeImage()
          this.countdown = this.qrcodeData.refreshIn || this.refreshInterval
          this.isExpired = false
          this.$message.success(this.$t('m.Refresh_Success'))
        } else {
          this.$message.error(res.message || this.$t('m.Refresh_Failed'))
        }
      } catch (error) {
        console.error('刷新二维码失败:', error)
        this.$message.error(this.$t('m.Refresh_Failed'))
      } finally {
        this.refreshing = false
      }
    },

    handleClose() {
      this.dialogVisible = false
      this.visible = false
      if (this.timer) {
        clearInterval(this.timer)
        this.timer = null
      }
      this.$emit('close')
    }
  },

  beforeDestroy() {
    if (this.timer) {
      clearInterval(this.timer)
    }
  }
}
</script>

<style scoped>
.qrcode-container {
  padding: 20px 0;
}

.qrcode-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.qrcode-box {
  padding: 15px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.qrcode-box.expired {
  opacity: 0.5;
  filter: grayscale(100%);
}

.qrcode-image {
  display: block;
  width: 280px;
  height: 280px;
}

.qrcode-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 280px;
  height: 280px;
  color: #909399;
}

.qrcode-placeholder i {
  font-size: 48px;
  margin-bottom: 10px;
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.countdown-wrapper {
  margin-bottom: 20px;
}

.countdown-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
}

.checkin-info {
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 15px;
}

.checkin-info p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.checkin-info strong {
  color: #303133;
  margin-right: 8px;
}

.stats {
  text-align: center;
  padding: 10px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
